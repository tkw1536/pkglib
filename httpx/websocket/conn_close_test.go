package websocket_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	gwebsocket "github.com/gorilla/websocket"
	"github.com/tkw1536/pkglib/httpx/websocket"
)

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
				if _, _, err := client.ReadMessage(); err == nil {
					t.Error("client-side did not receive an error when reading message")
				}
			})

			// check that the close cause is as expected
			gotCloseMessage := fmt.Sprint(gotCloseCause)
			if gotCloseMessage != tt.WantCloseMessage {
				t.Errorf("server-side got close cause %q, but want %q", gotCloseMessage, tt.WantCloseMessage)
			}
		})
	}
}
