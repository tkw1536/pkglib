package websocketx

import (
	"errors"
	"fmt"
	"strconv"
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
func (conn *Connection) close(frame websocket.CloseError, reason error, force bool) {
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
func (conn *Connection) ShutdownWith(frame websocket.CloseError) {
	if frame.Code <= 0 {
		frame.Code = websocket.CloseNormalClosure
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

// CloseFrame represents a closing control frame of a websocket connection.
// It is defined RFC 6455, section 5.5.1.
type CloseFrame struct {
	Code   StatusCode
	Reason string
}

// StatusCode is a status code as defined in RFC 6455, Section 7.4.
type StatusCode uint16

// String turns this status code in human-readable form for display.
func (st StatusCode) String() string {
	code := strconv.FormatUint(uint64(st), 10)
	name := st.Name()
	if name == "" {
		return code
	}
	return fmt.Sprintf("%s (%s)", code, name)
}

// Name returns the name of this status code, if it is known.
// If the name is not known, returns the empty string.
func (st StatusCode) Name() string {
	return statusCodeNames[st]
}

// Defined Websocket status codes as per RFC 6455, Section 7.4.1.
// Primarily from [rfc], see also [iana].
//
// [rfc]: https://www.rfc-editor.org/rfc/rfc6455.html#section-7.4.1
// [iana]: https://www.iana.org/assignments/websocket/websocket.xhtml#close-code-number
const (
	// Indicates a normal closure, meaning that the purpose for
	// which the connection was established has been fulfilled.
	StatusNormalClosure StatusCode = 1000

	// Indicates that an endpoint is "going away", such as a server
	// going down or a browser having navigated away from a page.
	StatusGoingAway StatusCode = 1001

	// indicates that an endpoint is terminating the connection due
	// to a protocol error.
	StatusProtocolError StatusCode = 1002

	// indicates that an endpoint is terminating the connection
	// because it has received a type of data it cannot accept (e.g., an
	// endpoint that understands only text data MAY send this if it
	// receives a binary message).
	StatusUnsupportedData StatusCode = 1003

	// A reserved value and MUST NOT be set as a status code in a
	// Close control frame by an endpoint.  It is designated for use in
	// applications expecting a status code to indicate that no status
	// code was actually present.
	StatusNoStatusReceived StatusCode = 1005

	// A reserved value and MUST NOT be set as a status code in a
	// Close control frame by an endpoint.  It is designated for use in
	// applications expecting a status code to indicate that the
	// connection was closed abnormally, e.g., without sending or
	// receiving a Close control frame.
	StatusAbnormalClosure StatusCode = 1006

	// Indicates that an endpoint is terminating the connection
	// because it has received data within a message that was not
	// consistent with the type of the message (e.g., non-UTF-8 [RFC3629]
	// data within a text message).
	StatusInvalidFramePayloadData StatusCode = 1007

	// Indicates that an endpoint is terminating the connection
	// because it has received a message that violates its policy.  This
	// is a generic status code that can be returned when there is no
	// other more suitable status code (e.g., [StatusUnsupportedData] or [StatusCloseMessageTooBig]) or if there
	// is a need to hide specific details about the policy.
	StatusPolicyViolation StatusCode = 1008

	// Indicates that an endpoint is terminating the connection
	// because it has received a message that is too big for it to
	// process.
	StatusMessageTooBig StatusCode = 1009

	// Indicates that an endpoint (client) is terminating the
	// connection because it has expected the server to negotiate one or
	// more extension, but the server didn't return them in the response
	// message of the WebSocket handshake.  The list of extensions that
	// are needed SHOULD appear in the /reason/ part of the Close frame.
	// Note that this status code is not used by the server, because it
	// can fail the WebSocket handshake instead.
	StatusMandatoryExtension StatusCode = 1010

	// Indicates that a remote endpoint is terminating the connection
	// because it encountered an unexpected condition that prevented it from
	// fulfilling the request.
	StatusInternalErr StatusCode = 1011

	// Indicates that the service is restarted. A client may reconnect,
	// and if it choses to do, should reconnect using a randomized delay
	// of 5 - 30s.
	StatusServiceRestart StatusCode = 1012

	// Indicates that the service is experiencing overload. A client
	// should only connect to a different IP (when there are multiple for the
	// target) or reconnect to the same IP upon user action.
	StatusTryAgainLater StatusCode = 1013

	// Indicates that the server was acting as a gateway or proxy and
	// received an invalid response from the upstream server.
	StatusBadGateway StatusCode = 1014

	// Additional status codes registered in the IANA registry.
	StatusUnauthorized StatusCode = 3000
	StatusForbidden    StatusCode = 3003
	StatusTimeout      StatusCode = 3008
)

var statusCodeNames = map[StatusCode]string{
	StatusNormalClosure:           "Normal Closure",
	StatusGoingAway:               "Going Away",
	StatusProtocolError:           "Protocol Error",
	StatusUnsupportedData:         "Unsupported Data",
	StatusNoStatusReceived:        "No Status Received",
	StatusAbnormalClosure:         "Abnormal Closure",
	StatusInvalidFramePayloadData: "Invalid Frame Payload Data",
	StatusPolicyViolation:         "Policy Violation",
	StatusMessageTooBig:           "Message too Big",
	StatusMandatoryExtension:      "Mandatory Extension",
	StatusInternalErr:             "Internal Error",
	StatusServiceRestart:          "Service Restart",
	StatusTryAgainLater:           "Try Again Later",
	StatusBadGateway:              "Bad Gateway",
	StatusUnauthorized:            "Unauthorized",
	StatusForbidden:               "Forbidden",
	StatusTimeout:                 "Timeout",
}

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
