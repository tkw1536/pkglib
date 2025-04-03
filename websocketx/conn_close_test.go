//spellchecker:words websocketx
package websocketx_test

//spellchecker:words context testing time github gorilla websocket pkglib websocketx uber goleak
import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tkw1536/pkglib/websocketx"
	"go.uber.org/goleak"
)

//spellchecker:words nolint errorlint

// so it is guaranteed that no valid test uses them.
const (
	// do nothing and simply exit the handler code.
	shutdownDoNothing websocketx.StatusCode = iota + 100

	// force close the handler without waiting.
	shutdownForceClose
)

var shutdownTests = []struct {
	Name string // name of the test

	// set any of these to skip the test in the corresponding section
	SkipServerTest  bool
	SkipHandlerTest bool
	SkipClientTest  bool

	// the frame to send to the server
	SendFrame websocketx.CloseFrame

	// expectations when closing from the server side
	WantCloseCalled bool
	WantCode        int
	WantReason      string

	// expectations when closing from the client side
	// if the frame is the zero value, omitted.
	WantCloseCause websocketx.CloseCause
}{
	{
		Name: "normal shutdown without message",

		SendFrame: websocketx.CloseFrame{
			Code: websocketx.StatusNormalClosure,
		},

		WantCloseCalled: true,
		WantCode:        websocket.CloseNormalClosure,
		WantReason:      "",

		WantCloseCause: websocketx.CloseCause{
			Frame: websocketx.CloseFrame{
				Code: websocketx.StatusNormalClosure,
			},
			WasClean: true,
		},
	},

	{
		Name: "normal shutdown with message",

		SendFrame: websocketx.CloseFrame{
			Code:   websocketx.StatusNormalClosure,
			Reason: "hello world",
		},
		WantCloseCalled: true,
		WantCode:        websocket.CloseNormalClosure,
		WantReason:      "hello world",

		WantCloseCause: websocketx.CloseCause{
			Frame: websocketx.CloseFrame{
				Code:   websocketx.StatusNormalClosure,
				Reason: "hello world",
			},
			WasClean: true,
		},
	},

	{
		Name:           "normal shutdown with huge message",
		SkipClientTest: true,

		SendFrame: websocketx.CloseFrame{
			Code:   websocketx.StatusNormalClosure,
			Reason: string(make([]byte, 1000)),
		},
		WantCloseCalled: true,
		WantCode:        websocket.CloseInternalServerErr,
		WantReason:      websocketx.MsgFailedCloseFrame,

		WantCloseCause: websocketx.CloseCause{
			Frame: websocketx.CloseFrame{
				Code:   websocketx.StatusNormalClosure,
				Reason: string(make([]byte, 1000)),
			},
			WasClean: true,
		},
	},

	{
		Name: "abnormal closure",

		SendFrame: websocketx.CloseFrame{
			Code: websocketx.StatusProtocolError,
		},

		WantCloseCalled: true,
		WantCode:        websocket.CloseProtocolError,
		WantReason:      "",

		WantCloseCause: websocketx.CloseCause{
			Frame: websocketx.CloseFrame{
				Code: websocketx.StatusProtocolError,
			},
			WasClean: true,
		},
	},
	{
		Name: "abnormal closure with message",

		SendFrame: websocketx.CloseFrame{
			Code:   websocketx.StatusProtocolError,
			Reason: "some message",
		},

		WantCloseCalled: true,
		WantCode:        websocket.CloseProtocolError,
		WantReason:      "some message",

		WantCloseCause: websocketx.CloseCause{
			Frame: websocketx.CloseFrame{
				Code:   websocketx.StatusProtocolError,
				Reason: "some message",
			},
			WasClean: true,
		},
	},
	{
		Name: "don't close",

		SendFrame: websocketx.CloseFrame{
			Code: shutdownDoNothing,
		},

		WantCloseCalled: true,
		WantCode:        websocket.CloseNormalClosure,
		WantReason:      "",
	},

	{
		Name: "force close",

		SendFrame: websocketx.CloseFrame{
			Code: shutdownForceClose,
		},

		WantCloseCalled: false,

		WantCloseCause: websocketx.CloseCause{
			Frame: websocketx.CloseFrame{
				Code: websocketx.StatusAbnormalClosure,
			},
			WasClean: false,
			Err:      io.ErrUnexpectedEOF,
		},
	},
}

// Tests that calling Shutdown and friends on the server struct
// actually shuts down the server.
func TestServer_ServerShutdown(t *testing.T) {
	defer goleak.VerifyNone(t)

	for _, tt := range shutdownTests {
		if tt.SendFrame.Code == shutdownDoNothing || tt.SkipServerTest {
			continue
		}

		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			testServer(t, nil, func(client *websocket.Conn, server *websocketx.Server) {
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

					if tt.SendFrame.Code == shutdownForceClose {
						server.Close()
						return
					}

					server.ShutdownWith(tt.SendFrame)
				}()

				// read the closing message
				_, _, _ = client.ReadMessage()
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
				if gotText != tt.WantReason {
					t.Errorf("got text %q, but want text %q", gotText, tt.WantReason)
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
		if tt.SkipHandlerTest {
			continue
		}
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			testServer(t, func(server *websocketx.Server) websocketx.Handler {
				return func(c *websocketx.Connection) {
					if tt.SendFrame.Code == shutdownDoNothing {
						return
					}
					if tt.SendFrame.Code == shutdownForceClose {
						_ = c.Close()
						return
					}

					c.ShutdownWith(tt.SendFrame)
				}
			}, func(client *websocket.Conn, _ *websocketx.Server) {
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
				_, _, _ = client.ReadMessage()

				if gotCalled != tt.WantCloseCalled {
					t.Errorf("wanted close called %v, but got close called %v", tt.WantCloseCalled, gotCalled)
				}

				if !tt.WantCloseCalled {
					return
				}
				if gotCode != tt.WantCode {
					t.Errorf("got code %d, but want code %d", gotCode, tt.WantCode)
				}
				if gotText != tt.WantReason {
					t.Errorf("got text %q, but want text %q", gotText, tt.WantReason)
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
		if tt.SendFrame.Code == shutdownDoNothing || tt.WantCloseCause.Frame.IsZero() || tt.SkipClientTest {
			continue
		}

		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			var gotCancelCause error
			testServer(t, func(server *websocketx.Server) websocketx.Handler {
				// record the close cause received
				return func(c *websocketx.Connection) {
					ctx := c.Context()
					<-ctx.Done()
					gotCancelCause = context.Cause(ctx)
				}
			}, func(client *websocket.Conn, _ *websocketx.Server) {
				if tt.SendFrame.Code == shutdownForceClose {
					_ = client.Close()
					return
				}

				// write the close message to the server
				if err := client.WriteControl(websocket.CloseMessage, tt.SendFrame.Body(), time.Now().Add(time.Second)); err != nil {
					t.Error("error writing close message")
				}

				// receive the close message back
				if _, _, err := client.ReadMessage(); err == nil {
					t.Error("client-side did not receive an error when reading message")
				}
			})

			gotCloseCause, ok := gotCancelCause.(websocketx.CloseCause) //nolint:errorlint
			if !ok {
				t.Errorf("server-side didn't return a close cause")
				return
			}

			if !closeCauseEquals(gotCloseCause, tt.WantCloseCause) {
				t.Errorf("server-side got close cause %q, but want %q", gotCloseCause, tt.WantCloseCause)
			}
		})
	}
}

// closeCauseEquals compares two [CloseCause]s.
func closeCauseEquals(left, right websocketx.CloseCause) bool {
	return left.Frame == right.Frame && left.WasClean == right.WasClean && fmt.Sprint(left.Err) == fmt.Sprint(right.Err)
}
