// Package status provides Status, LineBuffer and Group
//
//spellchecker:words status
package status

import "slices"

//spellchecker:words errors maps sync atomic time github gosuri uilive pkglib nobufio noop stream
import (
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gosuri/uilive"
	"github.com/tkw1536/pkglib/nobufio"
	"github.com/tkw1536/pkglib/noop"
	"github.com/tkw1536/pkglib/stream"
)

//spellchecker:words annot compat

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
// In addition to writing to lines directly, a status keeps separate log files for each line.
// These are automatically deleted once the Stop method is called, unless a separate call to the Keep() method is made.
//
// Status should only be used on interactive terminals.
// On other [io.Writer]s, a so-called compatibility mode can be used, that writes updates to the terminal line by line.
// See [NewWithCompat].
type Status struct {
	state   atomic.Uint64 // see state* comments below
	keepLog atomic.Bool   // keep the log files around
	counter atomic.Uint64 // the first free message id, increased atomically

	w      *uilive.Writer // underlying uilive writer
	compat bool           // compatibility mode enabled

	logPath      string                    // temporary path for log files (passed when creating logWriters)
	logWriters   map[uint64]io.WriteCloser // writers for the backup loggers
	logNamesLock sync.RWMutex              // protects the below
	logNames     map[uint64]string         // the names of the log files

	ids  []uint64       // ordered list of active message ids
	idsI map[uint64]int // inverse list of active message ids

	messages map[uint64]string // content of all the messages

	lastFlush time.Time // last time we flushed

	actions chan action // channel that status updates are sent to
	done    chan struct{}
}

// state* describe the lifecycle of a Status.
const (
	stateInvalid uint64 = iota
	stateNewCalled
	stateStartCalled
	stateStopCalled
)

// lineAction describe the types of actions for lines.
type lineAction uint8

const (
	setAction lineAction = iota
	openAction
	closeAction
)

// action describes actions to perform on a [Status].
type action struct {
	action  lineAction // what kind of action to perform
	id      uint64     // id of line to perform action on
	message string     // content of the line
}

// New creates a new writer with the provided number of status lines.
// count must fit into the uint64 type, meaning it has to be non-negative.
//
// The ids of the status lines are guaranteed to be 0...(count-1).
// When count is less than 0, it is set to 0.
func New(writer io.Writer, count int) *Status {
	if int(uint64(count)) /* #nosec G115 -- explicit check if it fits */ != count {
		panic("Status: count does not fit into uint64")
	}

	// when a zero writer was passed, we don't need a status.
	// and everything should become a no-op.
	if stream.IsNullWriter(writer) {
		return nil
	}

	if count < 0 {
		count = 0
	}

	st := &Status{
		w:      uilive.New(),
		compat: false,

		ids:  make([]uint64, count),
		idsI: make(map[uint64]int, count),

		messages: make(map[uint64]string, count),

		actions: make(chan action, count),
		done:    make(chan struct{}),
	}
	st.state.Store(stateNewCalled)
	st.counter.Store(uint64(count)) // #nosec G115 -- count fits into uint64

	// setup new ids
	for index := range st.ids {
		i := uint64(index) // #nosec G115 -- index < count which fits into uint64
		st.ids[index] = i
		st.idsI[i] = index

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

	if st.state.Load() == stateInvalid {
		panic("Status: Not created using New")
	}
	if !st.state.CompareAndSwap(stateNewCalled, stateStartCalled) {
		panic("Status: Start() called multiple times")
	}

	go st.listen()
}

const minFlushDelay = 50 * time.Millisecond

// see [flushCompat] and [flushNormal].
func (st *Status) flush(force bool, changed uint64) error {
	lErr := st.flushLogs(changed)

	var rErr error
	if st.compat {
		rErr = st.flushCompat(changed)
	} else {
		rErr = st.flushNormal(force)
	}

	return errors.Join(lErr, rErr)
}

// flushCompat flushes the provided updated message, if it is valid.
func (st *Status) flushCompat(changed uint64) error {
	line, ok := st.messages[changed]
	if !ok {
		return nil
	}
	_, err := fmt.Fprintln(st.w.Out, line)
	if err != nil {
		return fmt.Errorf("failed to flush writer: %w", err)
	}
	return nil
}

// flushLogs flushes to the given log file.
func (st *Status) flushLogs(changed uint64) error {
	line, ok := st.messages[changed]
	if !ok {
		return nil
	}
	logger, ok := st.logWriters[changed]
	if !ok {
		return nil
	}
	_, err := fmt.Fprintln(logger, line)
	if err != nil {
		return fmt.Errorf("failed to flush logger: %w", err)
	}
	return nil
}

// flushNormal implements flushing in normal mode.
// Respects [minFlushDelay], unless force is set to true.
func (st *Status) flushNormal(force bool) error {
	now := time.Now()
	if !force && now.Sub(st.lastFlush) < minFlushDelay {
		return nil
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

		_, _ = fmt.Fprintln(line, st.messages[key]) // ignore because error is always nil
	}

	// flush the output
	err := st.w.Flush()
	if err != nil {
		return fmt.Errorf("failed to flush writer: %w", err)
	}
	return nil
}

// Keep instructs this Status to not keep any log files, and returns a map from ids to file names.
func (st *Status) Keep() map[uint64]string {
	// we keep the log files!
	st.keepLog.Store(true)

	st.logNamesLock.RLock()
	defer st.logNamesLock.RUnlock()

	// make a copy of the logNames!
	files := make(map[uint64]string, len(st.logNames))
	maps.Copy(files, st.logNames)
	return files
}

// Stop blocks until all updates to finish processing.
// It then stops writing updates to the underlying writer.
// It then deletes all log files, unless a call to Keep() has been made.
//
// Stop must be called after [Start] has been called.
// Start may not be called more than once.
func (st *Status) Stop() {
	// nil check for no-op status
	if st == nil {
		return
	}

	if !st.state.CompareAndSwap(stateStartCalled, stateStopCalled) {
		panic("Status: Stop() called out-of-order")
	}

	close(st.actions)
	<-st.done
	_ = st.flush(true, st.counter.Add(1)) // force an invalid flush!

	// close the remaining loggers
	for _, id := range st.ids {
		_ = st.closeLogger(id) // closing the logger is low priority, so ignore closing errors
	}

	// if we requested for the log files to be deleted, do it!
	if !st.keepLog.Load() {
		st.logNamesLock.Lock()
		defer st.logNamesLock.Unlock()

		for _, name := range st.logNames {
			_ = os.Remove(name) // deleting the logs is low priority, so ignore the error
		}
	}
}

// openLogger opens the logger for the line with the given id.
func (st *Status) openLogger(id uint64) {
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
		st.logNames = make(map[uint64]string)
	}

	// store the file and name!
	if st.logWriters == nil {
		st.logWriters = make(map[uint64]io.WriteCloser, 1)
	}
	st.logWriters[id] = file

	if st.logNames == nil {
		st.logNames = make(map[uint64]string, 1)
	}
	st.logNames[id] = file.Name()
}

// closeLogger closes the logger for the line with the given id.
func (st *Status) closeLogger(id uint64) error {
	defer func() { _ = recover() }() // silently ignore errors

	// get and delete the log writer
	handle, ok := st.logWriters[id]
	delete(st.logWriters, id)

	// delete it if ok
	if ok {
		err := handle.Close()
		if err != nil {
			return fmt.Errorf("failed to close logger for id %d: %w", id, err)
		}
		return nil
	}
	return nil
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
func (st *Status) Set(id uint64, message string) {
	if st.state.Load() != stateStartCalled {
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
func (st *Status) Line(prefix string, id uint64) io.WriteCloser {
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

// NoLine indicates that the given writer does not have an associated line id.
const NoLine = ^uint64(0)

// LineOf returns the id of a line returned by the Line and OpenLine methods.
// If a different writer is passed (or there is no associated id), returns NoLine.
func LineOf(line io.WriteCloser) uint64 {
	lb, ok := line.(*LineBuffer)
	if !ok || !lb.annot {
		return NoLine
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
func (st *Status) Open(message string) (id uint64) {
	// nil check for no-op status
	if st == nil {
		return 0
	}

	// even when not active, generate a new id
	// this guarantees that other calls are no-ops.
	id = st.counter.Add(1)
	if st.state.Load() != stateStartCalled {
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
func (st *Status) Close(id uint64) {
	// nil check for no-op status
	if st == nil {
		return
	}
	if st.state.Load() != stateStartCalled {
		return
	}

	st.actions <- action{
		action: closeAction,
		id:     id,
	}
}

// listen listens for updates.
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
			_ = st.flush(false, msg.id) // no way to report error, so ignore it
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
			_ = st.flush(true, msg.id) // no way to report error
		case closeAction:
			// make sure that the line exists!
			index, ok := st.idsI[msg.id]
			if !ok {
				break
			}

			// close the logger
			_ = st.closeLogger(msg.id) // logging is low priority, so ignore errors

			// update the list of active ids
			// and rebuild the inverse index map
			st.ids = slices.Delete(st.ids, index, index+1)
			for key, value := range st.ids {
				st.idsI[value] = key
			}
			delete(st.idsI, msg.id)

			// flush out the current message!
			_, _ = fmt.Fprintln(st.w.Bypass(), st.messages[msg.id]) // ignore errors
			delete(st.messages, msg.id)

			// and flush all the other lines
			_ = st.flush(true, msg.id) // ignore errors
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

//spellchecker:words nosec
