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
			// TODO: with this enabled the loopclosure checker flags this code.
			// So far there is no way of silencing this; as it is correct as of go 1.22+.
			// t.Parallel()

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

const (
	timeoutShortTime = 100 * time.Millisecond
	timeoutLongTime  = 2 * timeoutShortTime
)

func TestServer_timeout(t *testing.T) {
	// create a server with very short timeouts
	var server websocket.Server
	server.Options.ReadInterval = timeoutShortTime
	server.Options.PingInterval = timeoutLongTime

	// create a handler that waits for the timeout
	// and simply check that the context got closed.
	done := make(chan struct{})
	server.Handler = func(c *websocket.Connection) {
		defer close(done)
		<-c.Context().Done()
	}

	// create a wss server
	wss := websockettest.NewServer(&server)
	defer wss.Close()

	// make a connection, but don't send anything
	c, _ := wss.Dial(nil, nil)
	defer c.Close()

	// wait for the timeout to expire
	time.Sleep(timeoutShortTime)

	// and check that the server didn't do anything
	select {
	case <-done:
	case <-time.After(timeoutLongTime):
		t.Error("server still open after a long time")
	}
}

const (
	readLimit       = 10 * 1024
	biggerThanLimit = readLimit + 10
)

func TestServer_ReadLimit(t *testing.T) {
	// create a server with very short timeouts
	var server websocket.Server
	server.Options.ReadLimit = readLimit

	// create a handler that waits for the timeout
	// and simply check that the context got closed.
	done := make(chan struct{})
	server.Handler = func(c *websocket.Connection) {
		defer close(done)
		select {
		case <-c.Read():
			t.Error("read large package unexpectedly")
		case <-c.Context().Done():
			/* closed connection */
		}
	}

	// create a wss server
	wss := websockettest.NewServer(&server)
	defer wss.Close()

	// make a connection
	c, _ := wss.Dial(nil, nil)
	defer c.Close()

	// try to send a large package
	// and wait for the handler to be done
	c.WriteMessage(gwebsocket.TextMessage, make([]byte, biggerThanLimit))
	<-done
}
