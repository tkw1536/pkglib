// Package websocket provides a handler for websockets
package websocket

import (
	"context"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/tkw1536/pkglib/lazy"
)

// Server implements a websocket server.
//
// Once a single call to [ServeHTTP] has been called, changes to any fields
// may be ignored.
type Server struct {
	// Context tbd
	Context context.Context

	// Handler is called for incoming client connections.
	// It must not be nil.
	Handler Handler

	// Fallback specifies the handler for generating HTTP responses if the client
	// did not request an upgrade to a websocket connection.
	// If Fallback is nil, then http.NotFound.
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

	upgrader lazy.Lazy[*websocket.Upgrader]
}

// getUpgrader sets defaults on [Options] and returns a (possibly cached)
// [websocket.Upgrader].
func (server *Server) getUpgrader() *websocket.Upgrader {
	return server.upgrader.Get(func() *websocket.Upgrader {
		server.Options.SetDefaults()

		return &websocket.Upgrader{
			HandshakeTimeout:  server.Options.HandshakeTimeout,
			ReadBufferSize:    server.Options.ReadBufferSize,
			WriteBufferSize:   server.Options.WriteBufferSize,
			WriteBufferPool:   server.Options.WriteBufferPool,
			Subprotocols:      server.Options.Subprotocols,
			Error:             server.Error,
			CheckOrigin:       server.CheckOrigin,
			EnableCompression: server.Options.CompressionEnabled(),
		}
	})
}

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
	// upgrade the connection or bail out!
	conn, err := h.getUpgrader().Upgrade(w, r, nil)
	if err != nil {
		return
	}

	// create a new connection
	var socket Connection
	defer socket.reset()

	// setup properties for the connection
	socket.r = r.Clone(r.Context())
	socket.conn = conn
	socket.opts = h.Options

	// and start handling
	socket.serve(h.Context, h.Handler)
}
