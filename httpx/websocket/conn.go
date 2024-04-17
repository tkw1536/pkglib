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

// Connection represents a connection to a single websocket client.
type Connection struct {
	r    *http.Request   // underlying http request
	conn *websocket.Conn // underlying connection
	opts Options         // opts defined for the connection

	context context.Context // context to cancel the connection
	cancel  context.CancelCauseFunc

	wg sync.WaitGroup // blocks all the ongoing tasks

	// incoming and outgoing tasks
	incoming chan Message
	outgoing chan queuedMessage
}

// serve serves the provided connection
// r is the original request that has been passed
func (conn *Connection) serve(ctx context.Context, handler Handler) {
	// enable compression if requested
	if enabled := conn.opts.CompressionEnabled(); enabled {
		conn.conn.EnableWriteCompression(true)
		_ = conn.conn.SetCompressionLevel(conn.opts.CompressionLevel)
	}

	// create a context for the connection
	if ctx == nil {
		ctx = context.Background()
	}
	conn.context, conn.cancel = context.WithCancelCause(ctx)

	// start receiving and sending messages
	conn.wg.Add(2)
	conn.sendMessages()
	conn.recvMessages()

	// wait for the context to be cancelled, then close the connection
	conn.wg.Add(1)
	go func() {
		defer conn.wg.Done()
		<-conn.context.Done()
		conn.conn.Close()
	}()

	// start the application logic
	conn.wg.Add(1)
	go conn.handle(handler)

	// wait for closing operations
	conn.wg.Wait()
}

func (conn *Connection) handle(handler Handler) {
	defer func() {
		defer conn.wg.Done()

		// when the handler panic()s, simply print the stack!
		// to not cause the server to crash!
		if value := recover(); value != nil {
			debug.PrintStack()

			conn.cancel(errCloseHandlerPanic)
			return
		}

		conn.cancel(errCloseHandlerReturn)
	}()

	handler(conn)
}

// Request returns a clone of the original request used for upgrading the connection.
// It can be used to e.g. check for authentication.
//
// Multiple calls to Request may return the same Request.
func (conn *Connection) Request() *http.Request {
	return conn.r
}

func (conn *Connection) sendMessages() {
	// turn on write compression!
	conn.conn.EnableWriteCompression(true)

	conn.outgoing = make(chan queuedMessage)

	go func() {
		// close connection when done!
		defer func() {
			conn.wg.Done()
			conn.cancel(errors.New(""))
		}()

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
						conn.cancel(WriteError{err: err})
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

// Close constants contain reasons for closing the server
const (
	CloseNormalClosure     = websocket.CloseNormalClosure
	CloseGoingAway         = websocket.CloseGoingAway
	CloseProtocolError     = websocket.CloseProtocolError
	CloseNoStatusReceived  = websocket.CloseNoStatusReceived
	ClosePolicyViolation   = websocket.ClosePolicyViolation
	CloseInternalServerErr = websocket.CloseInternalServerErr
	CloseServiceRestart    = websocket.CloseServiceRestart
)

func (conn *Connection) recvMessages() {
	conn.incoming = make(chan Message)

	// set a read handler
	conn.conn.SetReadLimit(conn.opts.ReadLimit)

	// configure a pong handler
	_ = conn.conn.SetReadDeadline(time.Now().Add(conn.opts.ReadInterval))
	conn.conn.SetPongHandler(func(string) error {
		return conn.conn.SetReadDeadline(time.Now().Add(conn.opts.ReadInterval))
	})

	// handle incoming messages
	go func() {
		// close connection when done!
		defer func() {
			conn.wg.Done()
			conn.cancel(errCloseOther)
		}()

		for {
			messageType, messageBytes, err := conn.conn.ReadMessage()
			if err != nil {
				// record client close error
				if err, ok := err.(*websocket.CloseError); ok {
					conn.cancel(err)
				}

				// otherwise return a read error
				conn.cancel(ReadError{err: err})
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

// Context returns a context that is closed once the connection is closed.
//
// Calling context.Cause(ctx) will return a *CloseError, ReadError or WriteError.
func (conn *Connection) Context() context.Context {
	return conn.context
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
