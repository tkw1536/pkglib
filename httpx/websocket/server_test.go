package websocket_test

import (
	"testing"
	"time"

	gwebsocket "github.com/gorilla/websocket"
	"github.com/tkw1536/pkglib/httpx/websocket"
	"github.com/tkw1536/pkglib/httpx/websocket/websockettest"
)

func TestServer_subprotocols(t *testing.T) {
	for _, tt := range []struct {
		Name string

		ServerProto []string
		ClientProto []string
		WantProto   string
	}{
		{
			Name: "no protocols",

			ServerProto: nil,
			ClientProto: nil,
			WantProto:   "",
		},

		{
			Name: "server supports known client protocol",

			ServerProto: []string{"a"},
			ClientProto: []string{"a"},
			WantProto:   "a",
		},

		{
			Name: "server knows more protocols than client",

			ServerProto: []string{"a", "b", "c"},
			ClientProto: []string{"c"},
			WantProto:   "c",
		},

		{
			Name: "client and server have no protocols in common",

			ServerProto: []string{"a", "b", "c"},
			ClientProto: []string{"d", "e", "f"},
			WantProto:   "",
		},

		{
			Name: "client and server have no protocols in common",

			ServerProto: []string{"a", "b", "c"},
			ClientProto: []string{"d", "e", "f"},
			WantProto:   "",
		},
		{
			Name: "client and server have multiple protocols in common",

			ServerProto: []string{"a", "b", "c"},
			ClientProto: []string{"c", "b"},
			WantProto:   "b",
		},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			// create a server with the specified websocket protocols
			var server websocket.Server
			server.Options.Subprotocols = tt.ServerProto

			// record the protocol that we got
			var gotProto string
			done := make(chan struct{})
			server.Handler = func(c *websocket.Connection) {
				defer close(done)
				gotProto = c.Subprotocol()
			}

			// create a wss server
			wss := websockettest.NewServer(&server)
			defer wss.Close()

			c, _ := wss.Dial(func(d *gwebsocket.Dialer) {
				d.Subprotocols = tt.ClientProto
			}, nil)

			// close the connection and wait for the record to be done
			c.Close()
			<-done

			if gotProto != tt.WantProto {
				t.Errorf("got protocol %q, but wanted %q", gotProto, tt.WantProto)
			}

		})
	}
}

func TestServer_timeout(t *testing.T) {
	// expect to read a message before the timeout expires
	// NOTE(twiesing): This must be smaller than the server timeout
	timeout := 500 * time.Millisecond
	if timeout >= testServerTimeout {
		panic("timeout is too big, pick a smaller one")
	}

	testServer(t, func(server *websocket.Server) websocket.Handler {
		server.Options.ReadInterval = timeout
		server.Options.PingInterval = testServerTimeout // don't send pings (which a sane client would respond to)

		// the handler just wait for the connection to close on it's own
		return func(c *websocket.Connection) {
			<-c.Context().Done()
		}
	}, func(c *gwebsocket.Conn, _ *websocket.Server) {
		// don't send a message during the timeout
		time.Sleep(timeout)
	})
}

const testServerTimeout = time.Minute

// testServer create a new testing server and initiates a cl ient.
func testServer(t *testing.T, initHandler func(server *websocket.Server) websocket.Handler, doClient func(client *gwebsocket.Conn, server *websocket.Server)) {
	t.Helper()

	// create the server
	var server websocket.Server

	// have the test code setup the handler
	var handler websocket.Handler
	if initHandler != nil {
		handler = initHandler(&server)
	}
	if handler == nil {
		handler = func(c *websocket.Connection) { <-c.Context().Done() }
	}
	if server.Handler != nil {
		panic("initHandler set server.Handler (wrong test code: return it instead)")
	}

	// update the actual handler
	done := make(chan struct{})
	server.Handler = func(c *websocket.Connection) {
		defer close(done)
		handler(c)
	}

	// create a wss server
	wss := websockettest.NewServer(&server)
	defer wss.Close()

	// make a connection, but don't send anything
	client, _ := wss.Dial(nil, nil)
	defer client.Close()

	// call the client code
	doClient(client, &server)

	select {
	case <-done:
	case <-time.After(testServerTimeout):
		t.Error("handler did not close within the given timeout")
	}
}

func TestServer_ReadLimit(t *testing.T) {
	const (
		readLimit       = 10 * 1024
		biggerThanLimit = readLimit + 10
	)

	testServer(t, func(server *websocket.Server) websocket.Handler {
		server.Options.ReadLimit = readLimit

		return func(c *websocket.Connection) {
			select {
			case <-c.Read():
				t.Error("read large package unexpectedly")
			case <-c.Context().Done():
				/* closed connection */
			}
		}
	}, func(client *gwebsocket.Conn, _ *websocket.Server) {
		// simply send a big message
		big := make([]byte, biggerThanLimit)
		_ = client.WriteMessage(gwebsocket.TextMessage, big)
	})
}
