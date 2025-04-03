//spellchecker:words status
package status

//spellchecker:words bytes errors sync time
import (
	"bytes"
	"errors"
	"sync"
	"time"
)

//spellchecker:words annot

// LineBuffer is an [io.Writer] that calls a function Line for every newline-delimited line written to it.
// Do not copy a non-zero LineBuffer.
type LineBuffer struct {
	m      sync.Mutex
	buffer bytes.Buffer
	closed bool

	// Line is called once a complete newline-terminated line has been written to this LineBuffer.
	// It is called for each line, in the correct order.
	// The parameter has a trailing '\r\n' or '\n' trimmed.
	//
	// Write methods block until the Line function has returned.
	// Therefore Line must not trigger another Write into the LineBuffer.
	Line func(line string)

	// FlushPartialLineAfter forces flushing output after a specific amount of time.
	// Set to 0 to disable.
	//
	// If the line has not been completed, the current partial line will be flushed.
	// A completed line will always be written.
	FlushPartialLineAfter time.Duration
	lastFlush             time.Time

	// FlushLineOnClose indicates if Line should be called a final time when calling close.
	// Line will only be called when the last write did not end in a newline character.
	FlushLineOnClose bool

	// CloseLine is called when the Close function of this LineBuffer is called for the first time.
	//
	// CloseLine may be nil, in which case it is not called.
	CloseLine func()

	annot   bool   // is the annotation valid
	annotID uint64 // annotated id (for use from within status)
}

// Write writes b into the internal buffer.
// When this completes one or more lines, calls Line appropriately.
func (lb *LineBuffer) Write(b []byte) (int, error) {
	lb.m.Lock()
	defer lb.m.Unlock()

	if lb.closed {
		return 0, errLineBufferClosed
	}

	defer lb.flush()
	return lb.buffer.Write(b)
}

// WriteByte is like [Write], but takes a single byte.
func (lb *LineBuffer) WriteByte(b byte) error {
	lb.m.Lock()
	defer lb.m.Unlock()

	if lb.closed {
		return errLineBufferClosed
	}

	defer lb.flush()
	return lb.buffer.WriteByte(b)
}

// WriteRune is like [Write], but takes a single rune.
func (lb *LineBuffer) WriteRune(r rune) (int, error) {
	lb.m.Lock()
	defer lb.m.Unlock()

	if lb.closed {
		return 0, errLineBufferClosed
	}

	defer lb.flush()
	return lb.buffer.WriteRune(r)
}

// WriteString is like [Write], but takes a string.
func (lb *LineBuffer) WriteString(s string) (int, error) {
	lb.m.Lock()
	defer lb.m.Unlock()

	if lb.closed {
		return 0, errLineBufferClosed
	}

	defer lb.flush()
	return lb.buffer.WriteString(s)
}

// runeR and runeN represent the bytes corresponding to '\r' and '\n' respectively.
const runeR byte = '\r'
const runeN byte = '\n'

// flush takes any completed lines in the internal buffer and calls the Line function.
func (lb *LineBuffer) flush() {
	// if we're closed, just delete all the lines!
	if lb.closed {
		lb.buffer.Reset()
		return
	}

	now := time.Now()
	for {
		// find the index of any '\n'
		index := bytes.IndexByte(lb.buffer.Bytes(), runeN)
		if index == -1 {
			break
		}

		// take the line, and trim any trailing '\r'
		line := lb.buffer.Next(index + 1)[:index]
		if index > 0 && line[index-1] == runeR {
			line = line[:index-1]
		}

		// call the line function!
		lb.Line(string(line))
		lb.lastFlush = now
	}

	lb.buffer.Grow(0) // trigger an internal re-slice!

	// check if we need to flush a line
	if lb.FlushPartialLineAfter > 0 {
		if now.Sub(lb.lastFlush) > lb.FlushPartialLineAfter {
			lb.forceFlush()
			lb.lastFlush = now
		}
	}
}

// forceFlush forcibly flushes the current line from the buffer.
// There is no guarantee that the line is complete.
// The buffer is not changed.
func (lb *LineBuffer) forceFlush() {
	rest := lb.buffer.Bytes()
	if len(rest) > 0 && rest[len(rest)-1] == runeR {
		rest = rest[:len(rest)-1]
	}
	if len(rest) > 0 {
		lb.Line(string(rest))
	}
}

var errLineBufferClosed = errors.New("`LineBuffer`: close was called")

// Close closes this LineBuffer, ensuring any future calls to [Write] or [Close] and friends return an error.
// When there was an unfinished line, close may cause a final flush of the buffer
// Close may block and wait for any concurrent calls to [Write] and friends active at the time of the Close call to finish.
//
// Writing to this LineBuffer after Close has returned a nil error no longer call the [Line] function.
// Calling Close multiple times returns nil error, and performs no further actions.
func (lb *LineBuffer) Close() error {
	lb.m.Lock()
	defer lb.m.Unlock()

	// mark the buffer as closed, unless
	if lb.closed {
		return nil
	}
	lb.closed = true

	// flush the final line if requested
	if lb.FlushLineOnClose {
		lb.forceFlush()
	}
	lb.buffer.Reset()

	// call the CloseLine function (if any)
	if lb.CloseLine != nil {
		lb.CloseLine()
	}
	return nil
}
