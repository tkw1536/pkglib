//spellchecker:words websocketx
package websocketx

//spellchecker:words errors strconv strings time github gorilla websocket
import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

//spellchecker:words nolint errname

const MsgFailedCloseFrame = "error writing close frame"

var failedCloseFrameMessage = websocket.FormatCloseMessage(websocket.CloseInternalServerErr, MsgFailedCloseFrame)

// close closes this connection with the given cause.
// It updates the connection state and sends a closing frame to the remote endpoint.
//
// The frame communicated to the remote endpoint will be the frame found in the cause,
// unless the caller provides an explicit frame instead.
//
// When the close frame fails to be written to the connection, sends an internal server error
// close frame with [MsgFailedCloseFrame] instead.
// Such a failure is typically caused by the close frame exceeding a certain size.
//
// This function also updates the internal connection state.
// If force is true, also calls forceClose.
//
// If the connection is already closed, this function does nothing.
func (conn *Connection) close(cause CloseCause, frame *CloseFrame, force bool) {
	conn.cancel(cause)

	// modifying the state
	conn.stateM.Lock()
	defer conn.stateM.Unlock()

	// if we are already closed, don't bother sending another frame back.
	// because we are closed!
	if conn.state == stateClosed {
		return
	}

	// write the close frame; but ignore any failures
	cf := cause.Frame
	if frame != nil {
		cf = *frame
	}

	err := conn.conn.WriteControl(websocket.CloseMessage, cf.Body(), time.Now().Add(conn.opts.HandshakeTimeout))
	if err != nil {
		// if the close frame failed to encode (probably it's too big) write a generic error instead
		// explicitly ignore the error cause we're in a fallback situation already
		_ = conn.conn.WriteControl(websocket.CloseMessage, failedCloseFrameMessage, time.Now().Add(conn.opts.HandshakeTimeout))
	}

	// do the actual close
	if force {
		_ = conn.forceClose(nil) // ignore any error; we did our best!
		return
	}

	// now in closing state
	conn.state = stateClosing
}

// forceClose kills the underlying network connection.
// and then updates the state of the connection.
//
// # If error is not nil, it also cancels the context with the given context
//
// the caller should hold stateM
func (conn *Connection) forceClose(err error) error {
	if conn.stateM.TryLock() {
		panic("conn.forceClose: stateM is not held")
	}

	// cancel the context if requested
	if err != nil {
		conn.cancel(CloseCause{
			Frame: CloseFrame{
				Code: StatusAbnormalClosure,
			},
			WasClean: false,
			Err:      err,
		})
	}

	errs := make([]error, 3)
	errs[0] = conn.conn.Close()

	// terminate any ongoing reads/writes
	// this may or may not have any effect
	now := time.Now()

	errs[1] = conn.conn.SetReadDeadline(now)
	errs[2] = conn.conn.SetWriteDeadline(now)

	// we're now in closed state
	conn.state = stateClosed

	return errors.Join(errs...)
}

var (
	// ErrConnectionShutdownWith indicates the connection closed because connection.ShutdownWith was called.
	ErrConnectionShutdownWith = errors.New("Connection.ShutdownWith called")

	// ErrConnectionClose indicates that the connection closed because connection.Close was called.
	ErrConnectionClose = errors.New("connection.Close called")
)

// ShutdownWith shuts down this connection with the given code and text for the client.
//
// ShutdownWith automatically formats a close message, sends it, and waits for the close handshake to complete or timeout.
// The timeout used is the normal ReadInterval timeout.
//
// When closeCode is 0, uses CloseNormalClosure.
func (conn *Connection) ShutdownWith(frame CloseFrame) {
	if frame.Code <= 0 {
		frame.Code = websocket.CloseNormalClosure
	}

	// write the connection close
	conn.close(CloseCause{Frame: frame, WasClean: true, Err: ErrConnectionShutdownWith}, nil, false)

	// wait for everything to close
	conn.wg.Wait()
}

// Close closes the connection to the peer without sending a specific close message.
// See [CloseWith] for providing the client with a reason for the closure.
func (conn *Connection) Close() error {
	conn.stateM.Lock()
	defer conn.stateM.Unlock()

	return conn.forceClose(ErrConnectionClose)
}

// CloseFrame represents a closing control frame of a websocket connection.
// It is defined RFC 6455, section 5.5.1.
type CloseFrame struct {
	Code   StatusCode
	Reason string
}

// Message returns the message body used to send this frame to
// a remote endpoint.
func (cf CloseFrame) Body() []byte {
	return websocket.FormatCloseMessage(int(cf.Code), cf.Reason)
}

// IsZero checks if this CloseFrame has the zero value
func (cf CloseFrame) IsZero() bool {
	var zero CloseFrame
	return zero == cf
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

// CloseCause is returned by calling [close.Cause] on the context of a connection.
// It indicates the reason why the server was closed.
//
//nolint:errname
type CloseCause struct {
	// CloseFrame is the close frame that cause the closure.
	// If no close frame was received, contains the [StatusAbnormalClosure] code.
	Frame CloseFrame

	// WasClean indicates if the connection is closed after receiving a close frame
	// from the client, or after having sent a close frame.
	//
	// NOTE: This roughly corresponds to the websocket JavaScript API's CloseEvent's wasClean.
	// However in situations where the server sends a close frame, but never receives a response
	// the WasClean field may still be true.
	// This detail is not considered part of the public API of this package, and may change in the future.
	WasClean bool

	// Err contains the error that caused the closure of this connection.
	// This may be an error returned from the read or write end of the connection,
	// or a server-side error.
	//
	// A special value contained in this field is ErrShuttingDown.
	Err error
}

// CloseCause implements the error interface.
func (cc CloseCause) Error() string {
	var builder strings.Builder

	_, err := fmt.Fprint(&builder, cc.Frame.Code)
	if err != nil {
		goto error
	}

	if cc.Frame.Reason != "" {
		_, err := fmt.Fprintf(&builder, " (reason: %q)", cc.Frame.Reason)
		if err != nil {
			goto error
		}
	}

	if !cc.WasClean {
		_, err := fmt.Fprint(&builder, " (unclean)")
		if err != nil {
			goto error
		}
	}

	if cc.Err != nil {
		_, err := fmt.Fprintf(&builder, ": %s", cc.Err)
		if err != nil {
			goto error
		}
	}

	return builder.String()
error:
	return "(error formatting CloseCause)"
}

// Unwrap implements the underlying Unwrap interface.
func (cc CloseCause) Unwrap() error {
	return cc.Err
}
