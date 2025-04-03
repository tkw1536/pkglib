// Package websocket provides a handler for websockets
//
//spellchecker:words websocketx
package websocketx

//spellchecker:words context errors http sync github gorilla websocket
import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

//spellchecker:words upgrader websockets nolint containedctx errorlint

// Server implements a websocket server.
type Server struct {
	m        sync.Mutex // protects modifying context below
	initDone bool       // true once we have initialized the server

	//nolint:containedctx
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

	// Check check if additional client requirements are met before establishing
	// a websocket connection and returns a caller-exposed error if
	// this is not the case. If Check is nil, every potential client is
	// allowed to connect.
	//
	// A client that is rejected will receive a [http.StatusForbidden] response
	// with a body of the error returned by this function.
	//
	// A typical use case includes enforcing a specific subprotocol,
	// before even reaching the handler, see [RequireProtocols].
	//
	// Check is called before CheckOrigin.
	Check func(r *http.Request) error

	// CheckOrigin returns true if the request Origin header is acceptable. If
	// CheckOrigin is nil, then a safe default is used: return false if the
	// Origin request header is present and the origin host is not equal to
	// request Host header.
	//
	// A CheckOrigin function should carefully validate the request origin to
	// prevent cross-site request forgery.
	//
	// CheckOrigin is only called if the Check function passes.
	CheckOrigin func(r *http.Request) bool

	// Error specifies the function for generating HTTP error responses. If Error
	// is nil, then http.Error is used to generate the HTTP response.
	Error func(w http.ResponseWriter, r *http.Request, status int, reason error)

	// Options determine further options for future connections.
	Options Options
}

// NoSupportedProtocolError is returned by [RequireProtocols].
// It indicates to a client that none of the required subprotocols are supported.
type NoSupportedProtocolError struct {
	Protocols []string
}

func (err NoSupportedProtocolError) Error() string {
	if len(err.Protocols) == 1 {
		return fmt.Sprintf("client does not support required %s subprotocol", err.Protocols[0])
	} else {
		return fmt.Sprintf("client does not support any of the required subprotocols %v", err.Protocols)
	}
}

// RequireProtocols returns a function which enforces that at least one of the given protocols is used by the given request.
// Empty protocols are ignored; if no non-empty protocols are provided, allows every protocol.
// It is intended to be used with [Server.Check], and returns an error wrapping.
func RequireProtocols(protocols ...string) func(*http.Request) error {
	// create a protocol set
	protocol_set := make(map[string]struct{}, len(protocols))
	the_protocols := make([]string, 0, len(protocols))
	for _, proto := range protocols {
		// skip empty subprotocol
		if proto == "" {
			continue
		}
		// add the protocol to the list if we saw it for the first time
		if _, ok := protocol_set[proto]; !ok {
			the_protocols = append(the_protocols, proto)
		}
		// record it in the set
		protocol_set[proto] = struct{}{}
	}

	// if we don't have any known protocols
	// we don't need to check anything
	if len(protocol_set) == 0 {
		return nil
	}

	// generate an error message to return
	// in case a subprotocol is not supported
	errNoSupportedProtocol := NoSupportedProtocolError{Protocols: the_protocols}

	// check the actual request
	return func(r *http.Request) error {
		for _, proto := range websocket.Subprotocols(r) {
			if _, ok := protocol_set[proto]; ok {
				return nil
			}
		}
		return errNoSupportedProtocol
	}
}

// RequireProtocols enforces that for any client to connect,
// at least one of the subprotocols has to be supported.
//
// This overrides the server.Check function.
func (server *Server) RequireProtocols() {
	server.Check = RequireProtocols(server.Options.Subprotocols...)
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

	// call the check function if it is provided
	if server.Check != nil {
		err := server.Check(r)
		if err != nil {
			http.Error(w, fmt.Sprint(err), http.StatusForbidden)
		}
	}

	// upgrade the connection or bail out!
	websocket_conn, err := server.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	// Accept the handler
	conn := server.Handler.accept(r, websocket_conn, server.Options)

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
			if errors.Is(cause, errServerShutdown) {
				return
			}

			// server is closing =>
			// close the connection
			if errors.Is(cause, errServerClose) {
				_ = conn.Close() // try our best to close the connection, and ignore errors
				return
			}

			// server is shutting down with a specific code =>
			// close the server with that specific code
			if cc, ok := cause.(CloseCause); ok { //nolint:errorlint
				conn.ShutdownWith(cc.Frame)
				return
			}

			panic("programming error: server.context received unknown cancel cause")
		}
	}()

	conn.serve()
}

var (
	errServerShutdown = errors.New("server shutting down")
	errServerClose    = errors.New("server closing")
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

var serverShuttingDown = CloseFrame{
	Code:   StatusGoingAway,
	Reason: "server shutting down",
}

// ShutdownWith gracefully shuts down the server by sending each client the given
// CloseFrame.
// If frame is the zero value, uses a default frame indicating that the server
// is shutting down instead.
//
// ShutdownWith first informs the server to stop accepting new connection attempts.
// Then it closes all existing connections by sending the given error.
// Finally it waits (indefinitely) for all existing connections to stop.
//
// See also [Shutdown] and [Close].
func (server *Server) ShutdownWith(frame CloseFrame) {
	if frame.IsZero() {
		frame = serverShuttingDown
	}

	server.close(CloseCause{Frame: frame})
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
