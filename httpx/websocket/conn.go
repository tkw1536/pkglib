package websocket

import (
	"context"
	"errors"
	"net/http"
	"runtime/debug"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// accept accepts a websocket connection with the specified handler.
// The caller should call [connection.serve] to start serving the connection.
func (handler Handler) accept(r *http.Request, conn *websocket.Conn, opt Options) *Connection {
	context, cancel := context.WithCancelCause(context.Background())

	return &Connection{
		state: CONNECTING,

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
	state  ConnectionState // connection state
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
	context context.Context
	cancel  context.CancelCauseFunc

	// incoming and outgoing messages
	incoming chan Message
	outgoing chan queuedMessage
}

// ConnectionState is the state of the connection
type ConnectionState int

// Different connection states
const (
	CONNECTING ConnectionState = iota
	OPEN
	CLOSING
	CLOSED
)

// serve starts serving data on the given connection.
// Serve returns once the connection has been closed, be it cleanly or non-cleanly.
// calling serve more than once is an error.
func (conn *Connection) serve() {
	// check that serve was only called once
	// and check that we are in the closed state
	{
		conn.stateM.Lock()
		if conn.state != CONNECTING {
			panic("Connection.serve: Called more than once")
		}
		conn.state = OPEN
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
		conn.close(websocket.CloseError{Code: code}, &websocket.CloseError{Code: code, Text: text}, true)
		return nil
	})

	return errors.Join(errs...)
}

func (conn *Connection) handle(handler Handler) {
	go func() {
		defer close(conn.handlerDone)
		defer func() {
			// when the handler panic()s, simply print the stack!
			// to not cause the server to crash!
			if value := recover(); value != nil {
				debug.PrintStack()

				// and produce an abnormal close error
				conn.close(websocket.CloseError{Code: websocket.CloseInternalServerErr}, nil, false)
				return
			}

			// close regularly
			conn.close(websocket.CloseError{Code: websocket.CloseNormalClosure}, nil, false)
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
		ping, err := websocket.NewPreparedMessage(PingMessage, []byte{})
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
						conn.close(websocket.CloseError{Code: websocket.CloseNoStatusReceived}, WriteError{err: err}, false)
						return
					}
					message.done <- struct{}{}
				})()

			// send a ping message
			case <-ticker.C:
				if err := conn.writeRaw(queuedMessage{prep: ping}); err != nil {
					conn.cancel(WriteError{err: err})
					return
				}
			}
		}
	}()

}

func (conn *Connection) writeRaw(message queuedMessage) error {
	if err := conn.conn.SetWriteDeadline(time.Now().Add(conn.opts.WriteInterval)); err != nil {
		return err
	}

	if message.prep != nil {
		return conn.conn.WritePreparedMessage(message.prep)
	}
	return conn.conn.WriteMessage(message.msg.Type, message.msg.Bytes)
}

// Write queues the provided message for sending.
// The returned channel is closed once the message has been sent or the connection is closed.
func (conn *Connection) Write(message Message) <-chan struct{} {
	return conn.write(queuedMessage{msg: message})
}

// WritePrepared queues the provided prepared message for sending.
// The returned channel is closed once the message has been sent or the connection is closed.
func (conn *Connection) WritePrepared(message PreparedMessage) <-chan struct{} {
	return conn.write(queuedMessage{prep: message.m})
}

func (conn *Connection) write(message queuedMessage) <-chan struct{} {
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
	return done
}

// WriteText is a convenience method to send a TextMessage.
// The returned channel is closed once the message has been sent.
func (sh *Connection) WriteText(text string) <-chan struct{} {
	return sh.Write(NewTextMessage(text))
}

// WriteText is a convenience method to send a BinaryMessage.
// The returned channel is closed once the message has been sent.
func (conn *Connection) WriteBinary(source []byte) <-chan struct{} {
	return conn.Write(NewBinaryMessage(source))
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
			if err != nil {
				conn.close(websocket.CloseError{Code: websocket.CloseNoStatusReceived}, ReadError{err: err}, false)
				return
			}

			// try to send a message to the incoming message channel
			select {
			case conn.incoming <- Message{
				Type:  messageType,
				Bytes: messageBytes,
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
// The [context.Cause] of the error will be one of *[CloseError], [ReadError] or [WriteError].
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

// Common message types see the gorilla websocket package for details.
const (
	TextMessage   = websocket.TextMessage
	BinaryMessage = websocket.BinaryMessage
	CloseMessage  = websocket.CloseMessage
	PingMessage   = websocket.PingMessage
	PongMessage   = websocket.PongMessage
)
