// Package websocket provides a handler for websockets
package websocket

import (
	"context"
	"net/http"
	"runtime/debug"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Server implements a websocket server.
type Server struct {
	Context context.Context // context which closes all connections
	Options Options         // Options for websocket connections

	Handler  Handler
	Fallback http.Handler

	upgrader websocket.Upgrader // upgrades upgrades connections
}

// Common message types see the gorilla websocket package for details.
const (
	TextMessage   = websocket.TextMessage
	BinaryMessage = websocket.BinaryMessage
	CloseMessage  = websocket.CloseMessage
	PingMessage   = websocket.PingMessage
	PongMessage   = websocket.PongMessage
)

// Handler handles a new incoming websocket connection.
// Handler may not retain a reference to its' argument past the function returning.
type Handler func(*Connection)

func (h *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// if the user did not request a websocket, go to the fallback handler
	if !websocket.IsWebSocketUpgrade(r) {
		h.serveFallback(w, r)
		return
	}

	// else deal with the websocket!
	h.serveWebsocket(w, r)
}

func (h *Server) serveFallback(w http.ResponseWriter, r *http.Request) {
	if h.Fallback == nil {
		http.NotFound(w, r)
		return
	}

	h.Fallback.ServeHTTP(w, r)
}

func (h *Server) serveWebsocket(w http.ResponseWriter, r *http.Request) {
	// clone the incoming request (to avoid having duplicate connection state)
	r2 := r.Clone(r.Context())

	// upgrade the connection or bail out!
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	// create a new connection
	var socket Connection
	defer socket.reset()
	socket.serve(h.Context, h.Options, r2, conn, h.Handler)
}

// Message represents a message sent between client and server.
type Message struct {
	Type  int
	Bytes []byte
}

// Binary checks if this message is a binary message
func (msg Message) Binary() bool {
	return msg.Type == BinaryMessage
}

// Text checks if this message is a text message
func (msg Message) Text() bool {
	return msg.Type == TextMessage
}

type outMessage struct {
	Message
	done chan<- struct{} // done should be closed when finished
}

// Connection represents a connection to a client.
type Connection struct {
	r    *http.Request   // underlying http request
	conn *websocket.Conn // underlying connection
	opts Options

	context context.Context // context to cancel the connection
	cancel  context.CancelFunc

	wg sync.WaitGroup // blocks all the ongoing tasks

	// incoming and outgoing tasks
	incoming chan Message
	outgoing chan outMessage
}

// serve serves the provided connection
// r is the original request that has been passed
func (conn *Connection) serve(ctx context.Context, opts Options, r *http.Request, c *websocket.Conn, handler Handler) {
	// use the connection!
	conn.r = r
	conn.conn = c

	// setup limits
	conn.opts = opts
	conn.opts.SetDefaults()

	// create a context for the connection
	if ctx == nil {
		ctx = context.Background()
	}
	conn.context, conn.cancel = context.WithCancel(ctx)

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
		// when the handler panic()s, simply print the stack!
		// to not cause the server to crash!
		if value := recover(); value != nil {
			debug.PrintStack()
		}

		conn.wg.Done()
		conn.cancel()
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

	conn.outgoing = make(chan outMessage)

	go func() {
		// close connection when done!
		defer func() {
			conn.wg.Done()
			conn.cancel()
		}()

		// setup a timer for pings!
		ticker := time.NewTicker(conn.opts.PingInterval)
		defer ticker.Stop()

		for {
			select {
			// everything is done!
			case <-conn.context.Done():
				return

			// send outgoing messages
			case message := <-conn.outgoing:
				(func() {
					defer close(message.done)

					err := conn.writeRaw(message.Type, message.Bytes)
					if err != nil {
						return
					}
					message.done <- struct{}{}
				})()
			// send a ping message
			case <-ticker.C:
				if err := conn.writeRaw(PingMessage, []byte{}); err != nil {
					return
				}
			}
		}
	}()

}

// writeRaw writes to the underlying socket
func (conn *Connection) writeRaw(messageType int, data []byte) error {
	conn.conn.SetWriteDeadline(time.Now().Add(conn.opts.WriteWait))
	return conn.conn.WriteMessage(messageType, data)
}

// Write queues the provided message for sending.
// The returned channel is closed once the message has been sent.
func (conn *Connection) Write(message Message) <-chan struct{} {
	callback := make(chan struct{}, 1)
	go func() {
		select {

		// write an outgoing message
		case conn.outgoing <- outMessage{
			Message: message,
			done:    callback,
		}:
		// context
		case <-conn.context.Done():
			close(callback)
		}
	}()
	return callback
}

// WriteText is a convenience method to send a TextMessage.
// The returned channel is closed once the message has been sent.
func (sh *Connection) WriteText(text string) <-chan struct{} {
	return sh.Write(Message{
		Type:  TextMessage,
		Bytes: []byte(text),
	})
}

// WriteText is a convenience method to send a BinaryMessage.
// The returned channel is closed once the message has been sent.
func (conn *Connection) WriteBinary(source []byte) <-chan struct{} {
	return conn.Write(Message{
		Type:  BinaryMessage,
		Bytes: source,
	})
}

func (conn *Connection) recvMessages() {
	conn.incoming = make(chan Message)

	// set a read handler
	conn.conn.SetReadLimit(conn.opts.MaxMessageSize)

	// configure a pong handler
	conn.conn.SetReadDeadline(time.Now().Add(conn.opts.PongWait))
	conn.conn.SetPongHandler(func(string) error { conn.conn.SetReadDeadline(time.Now().Add(conn.opts.PongWait)); return nil })

	// handle incoming messages
	go func() {
		// close connection when done!
		defer func() {
			conn.wg.Done()
			conn.cancel()
		}()

		for {
			messageType, messageBytes, err := conn.conn.ReadMessage()
			if err != nil {
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

// Read returns a channel that receives messages.
// The channel is closed if no more messages are available (for instance because the connection closed)
func (conn *Connection) Read() <-chan Message {
	return conn.incoming
}

// Context returns a context that is closed once this connection is closed.
func (conn *Connection) Context() context.Context {
	return conn.context
}

// Close closes the underlying connection.
func (conn *Connection) Close() {
	conn.cancel()
}

// reset resets this connection to be empty
func (h *Connection) reset() {
	h.opts = Options{}
	h.r = nil
	h.conn = nil
	h.incoming = nil
	h.outgoing = nil
	h.context, h.cancel = nil, nil
}

// Options describes limits for [Server].
type Options struct {
	WriteWait      time.Duration // maximum time to wait for writing
	PongWait       time.Duration // time to wait for pong responses
	PingInterval   time.Duration // interval to send pings to the client
	MaxMessageSize int64         // maximal message size in bytes
}

// Defaults for [Options]
const (
	DefaultWriteWait      = 10 * time.Second
	DefaultPongWait       = time.Minute
	DefaultPingInterval   = (DefaultPongWait * 9) / 10
	DefaultMaxMessageSize = 2048 // bytes
)

// SetDefaults sets defaults for options.
// See the appropriate default constants for this package.
func (opt *Options) SetDefaults() {
	if opt.WriteWait == 0 {
		opt.WriteWait = DefaultWriteWait
	}
	if opt.PongWait == 0 {
		opt.PongWait = DefaultPongWait
	}
	if opt.PingInterval <= 0 {
		opt.PingInterval = DefaultPingInterval
	}
	if opt.MaxMessageSize <= 0 {
		opt.MaxMessageSize = DefaultMaxMessageSize
	}
}
