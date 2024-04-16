// websockettest provides a server for testing.
// This package is not intended for production code, and should only be used in tests.
package websockettest

// spellchecker:words websockettest

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	gwebsocket "github.com/gorilla/websocket"
)

// A Server is a websocket test server listening on a system-chosen port on the
// local loopback interface, for use in end-to-end Websocket tests.
type Server struct {
	http *httptest.Server
	URL  string
}

// NewServer starts and returns a new [Server].
// Handler must respond to websocket upgrade requests.
// The caller should call Close when finished, to shut it down.
func NewServer(handler http.Handler) *Server {
	ts := httptest.NewServer(handler)

	return &Server{
		http: ts,
		URL:  "ws" + strings.TrimPrefix(ts.URL, "http"),
	}
}

// NewHandler returns a new http.Handler that upgrades connections and calls handler.
// Upon return from the handler function, the connection is automatically closed.
func NewHandler(handler func(conn *gwebsocket.Conn)) http.Handler {
	var upgrader gwebsocket.Upgrader

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// upgrade the connection
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// handle it!
		handler(conn)
	})
}

// Dial creates a new connection to the server.
// If connecting to the server fails (for instance because it has been shut down), panics.
func (srv *Server) Dial(opts func(*gwebsocket.Dialer), requestHeader http.Header) (*gwebsocket.Conn, *http.Response) {
	wsdialer := &gwebsocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
	}
	if opts != nil {
		opts(wsdialer)
	}

	conn, response, err := wsdialer.Dial(srv.URL, requestHeader)
	if err != nil {
		panic(fmt.Sprintf("websockettest.Server.Dial: Failed to connect: %s", err))
	}
	return conn, response
}

func (srv *Server) Close() {
	srv.http.Close()
}
