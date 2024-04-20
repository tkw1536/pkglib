// Package websocket provides a handler for websockets
package websocketx

import (
	"context"
	"errors"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Server provides a websocket server.
type Server struct {
	m        sync.Mutex // protects modifying context below
	initDone bool       // true once we have initialized the server

	context context.Context         // closed upon closing / shutting down server
	cancel  context.CancelCauseFunc // used to cancel the server
	conns   sync.WaitGroup          // holds 2 for every active connection

	upgrader websocket.Upgrader // used to upgrade connections

	// Handler is called for incoming client connections.
	// It must not be nil.
	Handler Handler

	// Fallback specifies the handler for generating HTTP responses if the client
	// did not request an upgrade to a websocket connection.
	// If Fallback is nil, sends an appropriate [http.StatusUpgradeRequired]
	Fallback http.Handler

	// CheckOrigin returns true if the request Origin header is acceptable. If
	// CheckOrigin is nil, then a safe default is used: return false if the
	// Origin request header is present and the origin host is not equal to
	// request Host header.
	//
	// A CheckOrigin function should carefully validate the request origin to
	// prevent cross-site request forgery.
	CheckOrigin func(r *http.Request) bool

	// Error specifies the function for generating HTTP error responses. If Error
	// is nil, then http.Error is used to generate the HTTP response.
	Error func(w http.ResponseWriter, r *http.Request, status int, reason error)

	// Options determine further options for future connections.
	Options Options
}

// Handler handles a connection sent to the server.
// It should exit as soon as possible once the connection's context is closed.
// Handler may not retain a reference to its' argument past the function returning.
type Handler func(*Connection)

// ServeHTTP implements the [http.Handler] interface.
//
// This is typically a websocket upgrade request.
// If a non http-client
func (server *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// if the user did not request a websocket, go to the fallback handler
	if !websocket.IsWebSocketUpgrade(r) {
		server.serveFallback(w, r)
		return
	}

	// else deal with the websocket!
	server.serveWebsocket(w, r)
}

// serveFallback serves a response to the client when no upgrade header was detected.
//
// This makes use of server.Fallback by default.
// If this is nil, sends an appropriate [http.StatusUpgradeRequired] response back instead.
func (server *Server) serveFallback(w http.ResponseWriter, r *http.Request) {
	// no fallback configured
	if server.Fallback != nil {
		server.Fallback.ServeHTTP(w, r)
		return
	}

	// inform the client they should upgrade to websockets
	w.Header().Add("Connection", "Upgrade")
	w.Header().Add("Upgrade", "websocket")
	w.WriteHeader(http.StatusUpgradeRequired)
}

// tryAccept atomically accepts a connection and increases the server.conns by 1.
// tryAccept guarantees that the server is initialized afterwards.
func (server *Server) tryAccept() bool {
	server.m.Lock()
	defer server.m.Unlock()

	server.init()

	// context is already closed
	if server.context.Err() != nil {
		return false
	}

	server.conns.Add(1)
	return true
}

// init initializes the server.
// m must be held while calling this method.
func (server *Server) init() {
	if server.initDone {
		return
	}
	server.initDone = true

	server.Options.SetDefaults()

	server.upgrader = websocket.Upgrader{
		HandshakeTimeout:  server.Options.HandshakeTimeout,
		ReadBufferSize:    server.Options.ReadBufferSize,
		WriteBufferSize:   server.Options.WriteBufferSize,
		WriteBufferPool:   server.Options.WriteBufferPool,
		Subprotocols:      server.Options.Subprotocols,
		Error:             server.Error,
		CheckOrigin:       server.CheckOrigin,
		EnableCompression: server.Options.CompressionEnabled(),
	}

	server.context, server.cancel = context.WithCancelCause(context.Background())
}

func (server *Server) serveWebsocket(w http.ResponseWriter, r *http.Request) {
	// try and accept a new server
	if !server.tryAccept() {
		http.Error(w, "websocket server closed", http.StatusServiceUnavailable)
		return
	}
	defer server.conns.Done()

	// upgrade the connection or bail out!
	wsconn, err := server.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	// Accept the handler
	conn := server.Handler.accept(r, wsconn, server.Options)

	server.conns.Add(1)
	go func() {
		defer server.conns.Done()

		select {
		case <-conn.Context().Done():
		case <-server.context.Done():
			// server is in shutdown process
			cause := context.Cause(server.context)

			// server shutting down =>
			// just wait for it to finish
			if cause == errServerShutdown {
				return
			}

			// server is closing =>
			// close the connection
			if cause == errServerClose {
				conn.Close()
				return
			}

			// server is shutting down with a specific code =>
			// close the server with that specific code
			if err, ok := cause.(*websocket.CloseError); ok {
				conn.ShutdownWith(*err)
				return
			}

			panic("programming error: server.context received unknown cancel cause")
		}
	}()

	conn.serve()
}

var (
	// ErrServerShuttingDown is sent to clients when the server is shutting down
	// and no other error message has been provided.
	ErrServerShuttingDown = websocket.CloseError{Code: websocket.CloseGoingAway, Text: "server shutting down"}

	// codes used to signal specific server shutdown actions
	errServerShutdown = errors.New("shutting down now")
	errServerClose    = errors.New("closing server now")
)

// Shutdown gracefully shuts down the server.
//
// Shutdown first informs the server to stop accepting new connection attempts.
// Then it waits (indefinitely) for all existing connections to stop.
//
// See also [ShutdownWith] and [Close].
func (server *Server) Shutdown() {
	server.close(errServerShutdown)
	server.conns.Wait()
}

// ShutdownWith gracefully shuts down the server by sending each client a CloseError.
//
// ShutdownWith first informs the server to stop accepting new connection attempts.
// Then it closes all existing connections by sending the given error.
// Finally it waits (indefinitely) for all existing connections to stop.
//
// See also [Shutdown] and [Close].
func (server *Server) ShutdownWith(err websocket.CloseError) {
	if err.Code == 0 {
		err = ErrServerShuttingDown
	}

	server.close(&err)
	server.conns.Wait()
}

// Close immediately closes the server by closing all client connections.
// Close does not wait for any existing connections to finish their shutdown.
//
// To force shutdown, and then wait for any active handlers, use a call to Close
// followed by a call to Shutdown.
//
// Use [Shutdown] or [ShutdownWith] instead.
func (server *Server) Close() {
	server.close(errServerClose)
}

// close closes the server with the given cause
func (server *Server) close(cause error) {
	server.m.Lock()
	defer server.m.Unlock()

	server.init()
	server.cancel(cause)
}
