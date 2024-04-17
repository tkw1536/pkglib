package websocket_test

import (
	"context"
	"fmt"
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
	}, func(c *gwebsocket.Conn) {
		// don't send a message during the timeout
		time.Sleep(timeout)
	})
}

const testServerTimeout = time.Minute

// testServer create a new testing server and initiates a cl ient.
func testServer(t *testing.T, initHandler func(server *websocket.Server) websocket.Handler, doClient func(client *gwebsocket.Conn)) {
	t.Helper()

	// create the server
	var server websocket.Server

	// have the test code setup the handler
	handler := initHandler(&server)
	if handler == nil {
		panic("initHandler return nil (wrong test code: return a non-nil handler)")
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
	doClient(client)

	select {
	case <-done:
	case <-time.After(testServerTimeout):
		t.Error("client connection not closed after timeout")
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
	}, func(client *gwebsocket.Conn) {
		// simply send a big message
		big := make([]byte, biggerThanLimit)
		client.WriteMessage(gwebsocket.TextMessage, big)
	})
}

// test that closing from the server side works as expected
func TestServer_ServerClose(t *testing.T) {
	const (
		sendNothing = -(iota + 1)
		sendForceClose
	)

	for _, tt := range []struct {
		Name string

		SendCode int
		SendText string

		WantCloseCalled bool
		WantCode        int
		WantText        string
	}{
		{
			Name: "normal closure without message",

			SendCode: websocket.CloseNormalClosure,
			SendText: "",

			WantCloseCalled: true,
			WantCode:        websocket.CloseNormalClosure,
			WantText:        "",
		},

		{
			Name: "normal closure with message",

			SendCode: websocket.CloseNormalClosure,
			SendText: "hello world",

			WantCloseCalled: true,
			WantCode:        websocket.CloseNormalClosure,
			WantText:        "hello world",
		},

		{
			Name: "abnormal closure",

			SendCode: websocket.CloseProtocolError,
			SendText: "",

			WantCloseCalled: true,
			WantCode:        websocket.CloseProtocolError,
			WantText:        "",
		},

		{
			Name: "don't close",

			SendCode: sendNothing,

			WantCloseCalled: false,
		},

		{
			Name: "force close",

			SendCode: sendForceClose,

			WantCloseCalled: false,
		},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			testServer(t, func(server *websocket.Server) websocket.Handler {
				return func(c *websocket.Connection) {
					if tt.SendCode == sendNothing {
						return
					}
					if tt.SendCode == sendForceClose {
						c.Close()
						return
					}

					c.CloseWith(tt.SendCode, tt.SendText)
				}
			}, func(client *gwebsocket.Conn) {
				var gotCode int
				var gotText string
				var gotCalled bool

				// get the default close handler
				handler := client.CloseHandler()

				client.SetCloseHandler(func(code int, text string) error {
					// store the code we got
					gotCode = code
					gotText = text
					gotCalled = true

					// and invoke the original handler
					return handler(code, text)
				})

				// read the closing message
				client.ReadMessage()

				if gotCalled != tt.WantCloseCalled {
					t.Errorf("wanted close called %v, but got close called %v", gotCalled, tt.WantCloseCalled)
				}

				if !tt.WantCloseCalled {
					return
				}
				if gotCode != tt.WantCode {
					t.Errorf("got code %d, but want code %d", gotCode, tt.WantCode)
				}
				if gotText != tt.WantText {
					t.Errorf("got text %q, but want text %q", gotText, tt.WantText)
				}
			})
		})
	}
}

// test that closing from the server side works as expected
func TestServer_ClientClose(t *testing.T) {
	const (
		sendClose = -(iota + 1)
	)

	for _, tt := range []struct {
		Name string

		SendCode int
		SendText string

		WantCloseMessage string
	}{
		{
			Name: "normal closure without message",

			SendCode: websocket.CloseNormalClosure,
			SendText: "",

			WantCloseMessage: "websocket: close 1000 (normal)",
		},

		{
			Name: "normal closure with message",

			SendCode: websocket.CloseNormalClosure,
			SendText: "some message by the client",

			WantCloseMessage: "websocket: close 1000 (normal): some message by the client",
		},

		{
			Name: "abnormal closure with message",

			SendCode: websocket.CloseProtocolError,
			SendText: "some message by the client",

			WantCloseMessage: "websocket: close 1002 (protocol error): some message by the client",
		},

		{
			Name: "abruptly close",

			SendCode: sendClose,

			WantCloseMessage: "websocket: close 1006 (abnormal closure): unexpected EOF",
		},
	} {
		t.Run(tt.Name, func(t *testing.T) {

			var gotCloseCause error
			testServer(t, func(server *websocket.Server) websocket.Handler {
				// record the close cause received
				return func(c *websocket.Connection) {
					ctx := c.Context()
					<-ctx.Done()
					gotCloseCause = context.Cause(ctx)
				}
			}, func(client *gwebsocket.Conn) {
				if tt.SendCode == sendClose {
					client.Close()
					return
				}

				// write the close message to the server
				client.WriteControl(websocket.CloseMessage, gwebsocket.FormatCloseMessage(tt.SendCode, tt.SendText), time.Now().Add(time.Second))

				// receive the close message back
				client.ReadMessage()
			})

			// check that the close cause is as expected
			gotCloseMessage := fmt.Sprint(gotCloseCause)
			if gotCloseMessage != tt.WantCloseMessage {
				t.Errorf("got close cause %q, but want %q", gotCloseMessage, tt.WantCloseMessage)
			}
		})
	}
}
