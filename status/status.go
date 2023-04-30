// Package status provides Status, LineBuffer and Group
package status

import (
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gosuri/uilive"
	"github.com/tkw1536/pkglib/nobufio"
	"github.com/tkw1536/pkglib/noop"
	"github.com/tkw1536/pkglib/stream"
	"golang.org/x/exp/maps"
)

// Status represents an interactive status display that can write to multiple lines at once.
//
// A Status must be initialized using [New], then started (and stopped again) to write messages.
// Status may not be reused.
//
// A typical usage is as follows:
//
//	st := New(os.Stdout, 10)
//	st.Start()
//	defer st.Stop()
//
//	// ... whatever usage here ...
//	st.Set("line 0", 0)
//
// Using the status to Write messages outside of the Start / Stop process results in no-ops.
//
// In addition to writing to lines directly, a status keeps separate logfiles for each line.
// These are automatically deleted once the Stop method is called, unless a separate call to the Keep() method is made.
//
// Status should only be used on interactive terminals.
// On other [io.Writer]s, a so-called compatibility mode can be used, that writes updates to the terminal line by line.
// See [NewWithCompat].
type Status struct {
	state   uint64 // see state* comments below
	keepLog uint64 // if non-zero, keeps the log files (has to be 64-bit-aligned)

	w      *uilive.Writer // underlying uilive writer
	compat bool           // compatibility mode enabled

	logPath      string                 // temporary path for log files (passed when creating logwriters)
	logWriters   map[int]io.WriteCloser // writers for the backup loggers
	logNamesLock sync.RWMutex           // protects the below
	logNames     map[int]string         // the names of the log files

	counter int32 // the first free message id, increased atomically

	ids  []int       // ordered list of active message ids
	idsI map[int]int // inverse list of active message ids

	messages map[int]string // content of all the messages

	lastFlush time.Time // last time we flushed

	actions chan action // channel that status updates are sent to
	done    chan struct{}
}

// state* describe the livecycle of a Status
const (
	stateInvalid uint64 = iota
	stateNewCalled
	stateStartCalled
	stateStopCalled
)

// lineAction describe the types of actions for lines
type lineAction uint8

const (
	setAction lineAction = iota
	openAction
	closeAction
)

// action describes actions to perform on a [Status]
type action struct {
	action  lineAction // what kind of action to perform
	id      int        // id of line to perform action on
	message string     // content of the line
}

// New creates a new writer with the provided number of status lines.
//
// The ids of the status lines are guaranteed to be 0...(count-1).
// When count is less than 0, it is set to 0.
func New(writer io.Writer, count int) *Status {
	// when a zero writer was passed, we don't need a status.
	// and everything should become a no-op.
	if stream.IsNullWriter(writer) {
		return nil
	}

	if count < 0 {
		count = 0
	}

	st := &Status{
		state: stateNewCalled,

		w:      uilive.New(),
		compat: false,

		counter: int32(count),

		ids:  make([]int, count),
		idsI: make(map[int]int, count),

		messages: make(map[int]string, count),

		actions: make(chan action, count),
		done:    make(chan struct{}),
	}

	// setup new ids
	for i := range st.ids {
		st.ids[i] = i
		st.idsI[i] = i

		// open the logger!
		st.openLogger(i)
	}

	st.w.Out = writer
	return st
}

// NewWithCompat is like [New], but places the Status into a compatibility mode if and only if writer does not represent a terminal.
//
// In compatibility mode, Status automatically prints each line to the output, instead of putting them onto separate lines.
func NewWithCompat(writer io.Writer, count int) (st *Status) {
	st = New(writer, count)
	st.compat = !nobufio.IsTerminal(writer)
	return st
}

// Start instructs this Status to start writing output to the underlying writer.
//
// No other process should write to the underlying writer, while this process is running.
// Instead [Bypass] should be used.
// See also [Stop].
//
// Start may not be called more than once, extra calls may result in a panic.
func (st *Status) Start() {
	// nil check for no-op status
	if st == nil {
		return
	}

	if atomic.LoadUint64(&st.state) == stateInvalid {
		panic("Status: Not created using New")
	}
	if !atomic.CompareAndSwapUint64(&st.state, stateNewCalled, stateStartCalled) {
		panic("Status: Start() called multiple times")
	}

	go st.listen()
}

const minFlushDelay = 50 * time.Millisecond

// flush flushes the output of this Status to the underlying writer.
// see [flushCompat] and [flushNormal]
func (st *Status) flush(force bool, changed int) {
	st.flushLogs(changed)
	if st.compat {
		st.flushCompat(changed)
		return
	}
	st.flushNormal(force)
}

// flushCompat flushes the provided updated message, if it is valid.
func (st *Status) flushCompat(changed int) {
	line, ok := st.messages[changed]
	if !ok {
		return
	}
	fmt.Fprintln(st.w.Out, line)
}

// flushLogs flushes to the given logfile
func (st *Status) flushLogs(changed int) {
	line, ok := st.messages[changed]
	if !ok {
		return
	}
	logger, ok := st.logWriters[changed]
	if !ok {
		return
	}
	fmt.Fprintln(logger, line)
}

// flushNormal implements flushing in normal mode.
// Respects [minFlushDelay], unless force is set to true.
func (st *Status) flushNormal(force bool) {

	now := time.Now()
	if !force && now.Sub(st.lastFlush) < minFlushDelay {
		return
	}
	st.lastFlush = now

	// write out each of the lines
	var line io.Writer
	for i, key := range st.ids {
		if i == 0 {
			line = st.w
		} else {
			line = st.w.Newline()
		}

		fmt.Fprintln(line, st.messages[key])
	}

	// flush the output
	st.w.Flush()
}

// Keep instructs this Status to not keep any log files, and returns a map from ids to file names.
func (st *Status) Keep() map[int]string {
	// we keep the log files!
	atomic.StoreUint64(&st.keepLog, 1)

	st.logNamesLock.RLock()
	defer st.logNamesLock.RUnlock()

	// make a copy of the logNames!
	files := make(map[int]string, len(st.logNames))
	maps.Copy(files, st.logNames)
	return files
}

// Stop blocks until all updates to finish processing.
// It then stops writing updates to the underlying writer.
// It then deletes all logfiles, unless a call to Keep() has been made.
//
// Stop must be called after [Start] has been called.
// Start may not be called more than once.
func (st *Status) Stop() {
	// nil check for no-op status
	if st == nil {
		return
	}

	if !atomic.CompareAndSwapUint64(&st.state, stateStartCalled, stateStopCalled) {
		panic("Status: Stop() called out-of-order")
	}

	close(st.actions)
	<-st.done
	st.flush(true, int(atomic.AddInt32(&st.counter, 1))) // force an invalid flush!

	// close the remaining loggers
	for _, id := range st.ids {
		st.closeLogger(id)
	}

	// if we requested for the log files to be deleted, do it!
	if atomic.LoadUint64(&st.keepLog) == 0 {
		st.logNamesLock.Lock()
		defer st.logNamesLock.Unlock()

		for _, name := range st.logNames {
			os.Remove(name)
		}
	}

}

// openLogger opens the logger for the line with the given id
func (st *Status) openLogger(id int) {
	if st == nil {
		return
	}

	file, err := os.CreateTemp(st.logPath, "status-*.log")
	if err != nil {
		return
	}

	st.logNamesLock.Lock()
	defer st.logNamesLock.Unlock()

	if st.logNames == nil {
		st.logNames = make(map[int]string)
	}

	// store the file and name!
	if st.logWriters == nil {
		st.logWriters = make(map[int]io.WriteCloser, 1)
	}
	st.logWriters[id] = file

	if st.logNames == nil {
		st.logNames = make(map[int]string, 1)
	}
	st.logNames[id] = file.Name()
}

// closeLogger closes the logger for the line with the given id
func (st *Status) closeLogger(id int) {
	defer func() { recover() }()

	if st == nil {
		return
	}

	// get and delete the log writer
	handle, ok := st.logWriters[id]
	delete(st.logWriters, id)

	// delete it if ok
	if ok {
		handle.Close()
	}
}

// Set sets the status line with the given id to contain message.
// message should not contain newline characters.
// Set may block until the addition has been processed.
//
// Calling Set on a line which is not active results is a no-op.
//
// Set may safely be called concurrently with other methods.
//
// Set may only be called after [Start] has been called, but before [Stop].
// Other calls are silently ignored, and return an invalid line id.
func (st *Status) Set(id int, message string) {
	if atomic.LoadUint64(&st.state) != stateStartCalled {
		return
	}

	st.actions <- action{
		action:  setAction,
		id:      id,
		message: message,
	}
}

// Line returns an [io.WriteCloser] linked to the status line with the provided id.
// Writing a complete newline-delimited line to it behaves just like [Set] with that line prefixed with prefix would.
// Calling [io.WriteCloser.Close] behaves just like [Close] would.
//
// Line may be called at any time.
// Line should not be called multiple times with the same id.
func (st *Status) Line(prefix string, id int) io.WriteCloser {
	// nil check for no-op status
	if st == nil {
		return stream.Null
	}

	// setup a delay for flushing partial lines after writes.
	// when in compatibility mode, this should be turned off.
	delay := 10 * minFlushDelay
	if st.compat {
		delay = 0
	}
	return &LineBuffer{
		FlushPartialLineAfter: delay,

		Line: func(message string) { st.Set(id, prefix+message) },

		FlushLineOnClose: true,
		CloseLine:        func() { st.Close(id) },

		annot:   true,
		annotID: id,
	}
}

// LineOf returns the id of a line returned by the Line and OpenLine methods.
// If a different writer is passed (or there is no associated id), returns -1.
func LineOf(line io.WriteCloser) int {
	lb, ok := line.(*LineBuffer)
	if !ok || !lb.annot {
		return -1
	}
	return lb.annotID
}

// Open adds a new status line and returns its' id.
// The new status line is initially set to message.
// It may be further updated with calls to [Set], or removed with [Done].
// Open may block until the addition has been processed.
//
// Open may safely be called concurrently with other methods.
//
// Open may only be called after [Start] has been called, but before [Stop].
// Other calls are silently ignored, and return an invalid line id.
func (st *Status) Open(message string) (id int) {
	// nil check for no-op status
	if st == nil {
		return 0
	}

	// even when not active, generate a new id
	// this guarantees that other calls are no-ops.
	id = int(atomic.AddInt32(&st.counter, 1))
	if atomic.LoadUint64(&st.state) != stateStartCalled {
		return
	}

	st.actions <- action{
		action:  openAction,
		id:      id,
		message: message,
	}
	return
}

// OpenLine behaves like a call to [Open] followed by a call to [Line].
//
// OpenLine may only be called after [Start] has been called, but before [Stop].
// Other calls are silently ignored, and return a no-op io.Writer.
//
// To retrieve the id of the newly created line, use [LineOf].
func (st *Status) OpenLine(prefix, data string) io.WriteCloser {
	// nil check for no-op status
	if st == nil {
		return noop.Writer{Writer: io.Discard}
	}
	return st.Line(prefix, st.Open(prefix+data))
}

// Close removes the status line with the provided id from this status.
// The last value of the status line is written to the top of the output.
// Close may block until the removal has been processed.
//
// Calling Close on a line which is not active results is a no-op.
//
// Close may safely be called concurrently with other methods.
//
// Close may only be called after [Start] has been called, but before [Stop].
// Other calls are silently ignored.
func (st *Status) Close(id int) {
	// nil check for no-op status
	if st == nil {
		return
	}
	if atomic.LoadUint64(&st.state) != stateStartCalled {
		return
	}

	st.actions <- action{
		action: closeAction,
		id:     id,
	}
}

// listen listens for updates
func (st *Status) listen() {
	// nil check for no-op status
	if st == nil {
		return
	}

	defer close(st.done)
	for msg := range st.actions {
		switch msg.action {
		case setAction:
			// if the id doesn't exist, do nothing!
			if _, ok := st.idsI[msg.id]; !ok {
				break
			}

			// store the message, and do a normal flush!
			st.messages[msg.id] = msg.message
			st.flush(false, msg.id)
		case openAction:
			// duplicate id, shouldn't occur
			if _, ok := st.idsI[msg.id]; ok {
				break
			}

			// add the item to the ids!
			st.ids = append(st.ids, msg.id)
			st.idsI[msg.id] = len(st.ids) - 1

			// setup the initial message
			st.messages[msg.id] = msg.message

			// open the logger!
			st.openLogger(msg.id)

			// force a flush so that we see it
			st.flush(true, msg.id)
		case closeAction:
			// make sure that the line exists!
			index, ok := st.idsI[msg.id]
			if !ok {
				break
			}

			// close the logger
			st.closeLogger(msg.id)

			// update the list of active ids
			// and rebuild the inverse index map
			st.ids = append(st.ids[:index], st.ids[index+1:]...)
			for key, value := range st.ids {
				st.idsI[value] = key
			}
			delete(st.idsI, msg.id)

			// flush out the current message!
			fmt.Fprintln(st.w.Bypass(), st.messages[msg.id])
			delete(st.messages, msg.id)

			// and flush all the other lines
			st.flush(true, msg.id)
		}
	}
}

// Bypass returns a writer that completely bypasses this Status, and writes directly to the underlying writer.
// [Start] must have been called.
func (st *Status) Bypass() io.Writer {
	// nil check for no-op status
	if st == nil {
		return io.Discard
	}

	return st.w.Bypass()
}
