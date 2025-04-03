//spellchecker:words websocketx
package websocketx

//spellchecker:words context errors http runtime debug sync time github gorilla websocket
import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

//spellchecker:words nolint containedctx

// accept accepts a websocket connection with the specified handler.
// The caller should call [connection.serve] to start serving the connection.
func (handler Handler) accept(r *http.Request, conn *websocket.Conn, opt Options) *Connection {
	context, cancel := context.WithCancelCause(context.Background())

	return &Connection{
		state: stateConnecting,

		r:       r.Clone(r.Context()),
		conn:    conn,
		opts:    opt,
		handler: handler,

		context:     context,
		cancel:      cancel,
		handlerDone: make(chan struct{}),
	}
}

// Connection represents a connection to a single websocket client.
type Connection struct {
	state  state // connection state
	stateM sync.Mutex

	// caller provided settings
	r       *http.Request
	conn    *websocket.Conn
	opts    Options
	handler Handler

	// holds all internal serve tasks
	wg          sync.WaitGroup // once the recv and send queues are finished
	handlerDone chan struct{}  // once the handler has returned

	// context that is open as long as user read/writes are permitted
	context context.Context         //nolint:containedctx
	cancel  context.CancelCauseFunc //nolint:containedctx

	// incoming and outgoing messages
	incoming chan Message
	outgoing chan queuedMessage
}

// state is the state of the connection
type state int32

// Different connection states
const (
	stateConnecting state = iota
	stateOpen
	stateClosing
	stateClosed
)

// serve starts serving data on the given connection.
// Serve returns once the connection has been closed, be it cleanly or non-cleanly.
// calling serve more than once is an error.
func (conn *Connection) serve() {
	// check that serve was only called once
	// and check that we are in the closed state
	{
		conn.stateM.Lock()
		if conn.state != stateConnecting {
			panic("Connection.serve: Called more than once")
		}
		conn.state = stateOpen
		conn.stateM.Unlock()
	}

	// setup connection options, then start sending and receiving
	_ = conn.setConnOpts()
	conn.sendMessages()
	conn.recvMessages()

	// start the application logic
	conn.handle(conn.handler)

	// wait for all the operations to finish
	conn.Shutdown()
}

func (conn *Connection) setConnOpts() error {
	var errs []error

	// enable compression for the write end of the connection
	if enabled := conn.opts.CompressionEnabled(); enabled {
		conn.conn.EnableWriteCompression(true)
		err := conn.conn.SetCompressionLevel(conn.opts.CompressionLevel)
		errs = append(errs, err)
	}

	// set a read handler
	conn.conn.SetReadLimit(conn.opts.ReadLimit)

	// when receiving a close frame, switch the state over to the close state
	conn.conn.SetCloseHandler(func(code int, text string) error {
		// setup a valid status code
		// if it is within a valid range
		sc := StatusNoStatusReceived
		if code >= 0 && code < (1<<16) {
			sc = StatusCode(code) // #nosec G115 -- explicit bounds check
		}

		conn.close(CloseCause{
			Frame: CloseFrame{
				Code:   sc,
				Reason: text,
			},
			WasClean: true,
			Err:      nil,
		}, &CloseFrame{Code: sc}, true)
		return nil
	})

	return errors.Join(errs...)
}

func (conn *Connection) handle(handler Handler) {
	go func() {
		defer close(conn.handlerDone)
		defer func() {

			// if we didn't panic, just close the connection
			// with a regular close frame
			value := recover()
			if value == nil {
				conn.ShutdownWith(CloseFrame{Code: StatusNormalClosure})
				return
			}

			// the handler has panic()ed, so we just print the stack!
			debug.PrintStack()

			// and cause a non-clean closure
			conn.close(CloseCause{
				Frame: CloseFrame{
					Code: StatusInternalErr,
				},
				WasClean: true,
				Err:      fmt.Errorf("%v", value), // nolint:err113
			}, nil, false)
		}()

		handler(conn)
	}()
}

// Request returns a clone of the original request used for upgrading the connection.
// It can be used to e.g. check for authentication.
//
// Multiple calls to Request may return the same Request.
func (conn *Connection) Request() *http.Request {
	return conn.r
}

func (conn *Connection) sendMessages() {
	conn.outgoing = make(chan queuedMessage)

	conn.wg.Add(1)
	go func() {
		defer conn.wg.Done()

		// setup a timer for pings!
		ticker := time.NewTicker(conn.opts.PingInterval)
		defer ticker.Stop()

		// prepare a ping message
		ping, err := websocket.NewPreparedMessage(websocket.PingMessage, []byte{})
		if err != nil {
			return
		}

		for {
			select {
			// everything is done!
			case <-conn.context.Done():
				return

			// send outgoing messages
			case message := <-conn.outgoing:
				(func() {
					defer close(message.done)

					err := conn.writeRaw(message)
					if err != nil {
						return
					}
					message.done <- struct{}{}
				})()

			// send a ping message
			case <-ticker.C:
				if err := conn.writeRaw(queuedMessage{prep: ping}); err != nil {
					return
				}
			}
		}
	}()

}

// writeRaw writes an underlying message to the connection.
// If an error occurs, closes the connection.
func (conn *Connection) writeRaw(message queuedMessage) (err error) {
	defer func() {
		if err != nil {
			conn.close(CloseCause{
				Frame: CloseFrame{
					Code: websocket.CloseAbnormalClosure,
				},
				WasClean: false,
				Err:      err,
			}, nil, false)
		}
	}()
	if err := conn.conn.SetWriteDeadline(time.Now().Add(conn.opts.WriteInterval)); err != nil {
		return err
	}

	if message.prep != nil {
		return conn.conn.WritePreparedMessage(message.prep)
	}
	return conn.conn.WriteMessage(message.msg.Type, message.msg.Body)
}

// Write queues the provided message for sending
// and blocks until the given message has been sent.
//
// Write returns a non-nil error if and only if the message failed to send.
// In such a case, Write will internally close the connection
// and return an error of type CancelCause.
//
// Call Write concurrently with other read and write calls is safe.
// When multiple calls to Write are in-progress, all
// messages will be sent, but their order is undefined
// unless the callers explicitly coordinate.
func (conn *Connection) Write(message Message) error {
	return conn.write(queuedMessage{msg: message})
}

// WritePrepared is like Write, but sends a PreparedMessage instead.
func (conn *Connection) WritePrepared(message PreparedMessage) error {
	return conn.write(queuedMessage{prep: message.m})
}

// WriteText is like Write, but is sends a TextMessage with the given text.
func (sh *Connection) WriteText(text string) error {
	return sh.Write(NewTextMessage(text))
}

// WriteBinary is like Write, but it sends a BinaryMessage with the given text.
func (conn *Connection) WriteBinary(source []byte) error {
	return conn.Write(NewBinaryMessage(source))
}

func (conn *Connection) write(message queuedMessage) error {
	done := make(chan struct{}, 1)

	go func() {
		message.done = done
		select {

		// write an outgoing message
		case conn.outgoing <- message:

		// closed
		case <-conn.context.Done():
			close(done)
		}
	}()

	_, ok := <-done
	if !ok {
		return context.Cause(conn.context)
	}

	return nil
}

func (conn *Connection) recvMessages() {
	conn.incoming = make(chan Message)

	// upon receiving a pong, delay the read interval
	conn.conn.SetPongHandler(func(string) error {
		return conn.conn.SetReadDeadline(time.Now().Add(conn.opts.ReadInterval))
	})

	conn.wg.Add(1)
	go func() {
		defer conn.wg.Done()

		for {
			// set a timeout for the next read
			_ = conn.conn.SetReadDeadline(time.Now().Add(conn.opts.ReadInterval))

			messageType, messageBytes, err := conn.conn.ReadMessage()

			// intercept any unexpected CloseErrors
			// this only has an effect if the context has not yet been closed.
			if ce, ok := err.(*websocket.CloseError); ok {
				err = fmt.Errorf("%s", ce.Text) // nolint:err113
			}
			if err != nil {
				conn.close(CloseCause{
					Frame: CloseFrame{
						Code: StatusAbnormalClosure,
					},
					WasClean: false,
					Err:      err,
				}, nil, false)
				return
			}

			// try to send a message to the incoming message channel
			select {
			case conn.incoming <- Message{
				Type: messageType,
				Body: messageBytes,
			}:
			case <-conn.context.Done():
				return
			}
		}
	}()
}

// Read returns a channel that receives Text and Binary Messages from the peer.
// Once the websocket connection state is corrupted or closed, the channel is closed.
//
// Multiple invocations of Read returns the same channel.
func (conn *Connection) Read() <-chan Message {
	return conn.incoming
}

// Subprotocol returns the negotiated protocol for the connection.
// If no subprotocol has been negotiated, returns the empty string.
func (conn *Connection) Subprotocol() string {
	return conn.conn.Subprotocol()
}

// Context returns a context for operations on this context.
// Once it is closed, no more read or write operations are permitted.
//
// The [context.Cause] will return an error of type [CloseCause],
// representing the reason for the closure.
//
// The Context may close before the Handler function has returned.
// To wait for the handler function to return, use Wait() instead.
func (conn *Connection) Context() context.Context {
	return conn.context
}

// Shutdown waits until the connection has been closed and the handler returns.
//
// NOTE: This name exists to correspond to the graceful server shutdown.
func (conn *Connection) Shutdown() {
	conn.wg.Wait()
	<-conn.handlerDone
}

// queuedMessage is a message queued for writing.
// it is either a regular message or a prepared message.
type queuedMessage struct {
	msg  Message
	prep *websocket.PreparedMessage

	done chan<- struct{} // done should be closed when finished
}

//spellchecker:words nosec
