package websocket

import (
	"errors"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

// close sends a close frame on the underlying network connection.
// It then sets the connection into the appropriate new closing state.
//
// It also cancels the connection context with the given reason.
// If the reason is nil, it defaults to frame.
//
// If force is true, also calls forceClose.
//
// If the connection is already closed, this function does nothing.
func (conn *Connection) close(frame CloseError, reason error, force bool) {
	// cancel the context with the appropriate reason
	if reason == nil {
		reason = &frame
	}
	conn.cancel(reason)

	// modifying the state
	conn.stateM.Lock()
	defer conn.stateM.Unlock()

	// if we are already closed, don't bother sending another frame back.
	// because we are closed!
	if conn.state == CLOSED {
		return
	}

	// write the close frame; but ignore any failures
	message := websocket.FormatCloseMessage(frame.Code, frame.Text)
	_ = conn.conn.WriteControl(websocket.CloseMessage, message, time.Now().Add(conn.opts.HandshakeTimeout))

	// do the actual close
	if force {
		conn.forceClose()
		return
	}

	// now in closing state
	conn.state = CLOSING
}

// forceClose kills the underlying network connection
// and then updates the state of the connection
//
// the caller should hold stateM
func (conn *Connection) forceClose() error {
	if conn.stateM.TryLock() {
		panic("conn.forceClose: stateM is not held")
	}

	errs := make([]error, 3)
	errs[0] = conn.conn.Close()

	// terminate any ongoing reads/writes
	// this may or may not have any effect
	now := time.Now()

	errs[1] = conn.conn.SetReadDeadline(now)
	errs[2] = conn.conn.SetWriteDeadline(now)

	// we're now in closed state
	conn.state = CLOSED

	return errors.Join(errs...)
}

// ShutdownWith shuts down this connection with the given code and text for the client.
//
// ShutdownWith automatically formats a close message, sends it, and waits for the close handshake to complete or timeout.
// The timeout used is the normal ReadInterval timeout.
//
// When closeCode is 0, uses CloseNormalClosure.
func (conn *Connection) ShutdownWith(frame CloseError) {
	if frame.Code <= 0 {
		frame.Code = CloseNormalClosure
	}

	// write the connection close
	conn.close(frame, nil, false)

	// wait for everything to close
	conn.wg.Wait()
}

// Close closes the connection to the peer without sending a specific close message.
// See [CloseWith] for providing the client with a reason for the closure.
func (conn *Connection) Close() error {
	conn.stateM.Lock()
	defer conn.stateM.Unlock()

	return conn.forceClose()
}

// CloseError is the cancel cause of [Connection.Context] if a closing handshake too place.
type CloseError = websocket.CloseError

// ReadError is the cancel cause of [Connection.Context] if an error occurred during writing.
type ReadError struct {
	err error
}

func (err ReadError) Error() string {
	return fmt.Sprintf("read error: %v", err.err)
}
func (err ReadError) Unwrap() error {
	return err.err
}

// WriteError is the cancel cause of [Connection.Context] if an error occurred during writing.
type WriteError struct {
	err error
}

func (err WriteError) Error() string {
	return fmt.Sprintf("write error: %v", err.err)
}
func (err WriteError) Unwrap() error {
	return err.err
}
