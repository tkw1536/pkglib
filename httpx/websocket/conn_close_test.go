package websocket_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	gwebsocket "github.com/gorilla/websocket"
	"github.com/tkw1536/pkglib/httpx/websocket"
	"go.uber.org/goleak"
)

const (
	// do nothing and simply exit the handler code
	shutdownDoNothing = -(iota + 1)

	// force close the handler without waiting
	shutdownForceClose
)

var shutdownTests = []struct {
	Name string // name of the test

	SendCode int    // code to send during shutdown (or special action)
	SendText string // text to send during shutdown

	// expectations when closing from the server side
	WantCloseCalled bool
	WantCode        int
	WantText        string

	// expectations when closing from the client side
	// if omitted, skip the test.
	WantCloseCause string
}{
	{
		Name: "normal shutdown without message",

		SendCode: websocket.CloseNormalClosure,
		SendText: "",

		WantCloseCalled: true,
		WantCode:        websocket.CloseNormalClosure,
		WantText:        "",

		WantCloseCause: "websocket: close 1000 (normal)",
	},

	{
		Name: "normal shutdown with message",

		SendCode: websocket.CloseNormalClosure,
		SendText: "hello world",

		WantCloseCalled: true,
		WantCode:        websocket.CloseNormalClosure,
		WantText:        "hello world",

		WantCloseCause: "websocket: close 1000 (normal): hello world",
	},

	{
		Name: "abnormal closure",

		SendCode: websocket.CloseProtocolError,
		SendText: "",

		WantCloseCalled: true,
		WantCode:        websocket.CloseProtocolError,
		WantText:        "",

		WantCloseCause: "websocket: close 1002 (protocol error)",
	},
	{
		Name: "abnormal closure with message",

		SendCode: websocket.CloseProtocolError,
		SendText: "some message",

		WantCloseCalled: true,
		WantCode:        websocket.CloseProtocolError,
		WantText:        "some message",

		WantCloseCause: "websocket: close 1002 (protocol error): some message",
	},
	{
		Name: "don't close",

		SendCode: shutdownDoNothing,

		WantCloseCalled: true,
		WantCode:        websocket.CloseNormalClosure,
		WantText:        "",
	},

	{
		Name: "force close",

		SendCode: shutdownForceClose,

		WantCloseCalled: false,

		WantCloseCause: "read error: websocket: close 1006 (abnormal closure): unexpected EOF",
	},
}

// Tests that calling Shutdown and friends on the server struct
// actually shuts down the server.
func TestServer_ServerShutdown(t *testing.T) {
	defer goleak.VerifyNone(t)

	for _, tt := range shutdownTests {
		if tt.SendCode == shutdownDoNothing {
			continue
		}

		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			testServer(t, nil, func(client *gwebsocket.Conn, server *websocket.Server) {
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

				done := make(chan struct{})
				go func() {
					defer close(done)

					if tt.SendCode == shutdownForceClose {
						server.Close()
						return
					}

					server.ShutdownWith(websocket.CloseError{Code: tt.SendCode, Text: tt.SendText})
				}()

				// read the closing message
				client.ReadMessage()
				<-done

				if gotCalled != tt.WantCloseCalled {
					t.Errorf("wanted close called %v, but got close called %v", tt.WantCloseCalled, gotCalled)
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

// Tests that calling shutdown from within the handler
// does the expected thing.
func TestServer_HandlerShutdown(t *testing.T) {
	defer goleak.VerifyNone(t)

	for _, tt := range shutdownTests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			testServer(t, func(server *websocket.Server) websocket.Handler {
				return func(c *websocket.Connection) {
					if tt.SendCode == shutdownDoNothing {
						return
					}
					if tt.SendCode == shutdownForceClose {
						c.Close()
						return
					}

					c.ShutdownWith(websocket.CloseError{Code: tt.SendCode, Text: tt.SendText})
				}
			}, func(client *gwebsocket.Conn, _ *websocket.Server) {
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
					t.Errorf("wanted close called %v, but got close called %v", tt.WantCloseCalled, gotCalled)
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

// Tests that shutting down from the client side does the right thing.
func TestServer_ClientClose(t *testing.T) {
	defer goleak.VerifyNone(t)

	for _, tt := range shutdownTests {
		// skip tests that aren't supported
		if tt.SendCode == shutdownDoNothing || tt.WantCloseCause == "" {
			continue
		}

		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			var gotCloseCause error
			testServer(t, func(server *websocket.Server) websocket.Handler {
				// record the close cause received
				return func(c *websocket.Connection) {
					ctx := c.Context()
					<-ctx.Done()
					gotCloseCause = context.Cause(ctx)
				}
			}, func(client *gwebsocket.Conn, _ *websocket.Server) {
				if tt.SendCode == shutdownForceClose {
					client.Close()
					return
				}

				// write the close message to the server
				client.WriteControl(websocket.CloseMessage, gwebsocket.FormatCloseMessage(tt.SendCode, tt.SendText), time.Now().Add(time.Second))

				// receive the close message back
				if _, _, err := client.ReadMessage(); err == nil {
					t.Error("client-side did not receive an error when reading message")
				}
			})

			// check that the close cause is as expected
			gotCloseMessage := fmt.Sprint(gotCloseCause)
			if gotCloseMessage != tt.WantCloseCause {
				t.Errorf("server-side got close cause %q, but want %q", gotCloseMessage, tt.WantCloseCause)
			}
		})
	}
}
