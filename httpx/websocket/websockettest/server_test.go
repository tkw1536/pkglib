package websockettest_test

import (
	"io"
	"testing"

	gwebsocket "github.com/gorilla/websocket"
	"github.com/tkw1536/pkglib/httpx/websocket/websockettest"
)

func TestNewServer(t *testing.T) {
	// Create a simple echo server
	echo := websockettest.NewServer(websockettest.NewHandler(func(conn *gwebsocket.Conn) {
		for {
			// read the message from the server
			mt, message, err := conn.ReadMessage()
			if err != nil {
				break
			}

			// and echo it back
			err = conn.WriteMessage(mt, message)
			if err != nil {
				break
			}
		}
	}))
	t.Cleanup(echo.Close)

	type Message = struct {
		Type int
		Body string
	}

	for _, tt := range []struct {
		Name     string
		Messages []Message
	}{
		{
			"one message",
			[]Message{
				{Type: gwebsocket.TextMessage, Body: "hello"},
			},
		},

		{
			"lots of messages",
			[]Message{
				{Type: gwebsocket.TextMessage, Body: "hello"},
				{Type: gwebsocket.BinaryMessage, Body: "world"},
				{Type: gwebsocket.TextMessage, Body: "how"},
				{Type: gwebsocket.BinaryMessage, Body: "are"},
				{Type: gwebsocket.TextMessage, Body: "you"},
				{Type: gwebsocket.BinaryMessage, Body: "doing"},
				{Type: gwebsocket.TextMessage, Body: "today"},
			},
		},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			// create a new client
			client, _ := echo.Dial(nil)

			for _, msg := range tt.Messages {
				// write the message to the network
				writer, err := client.NextWriter(msg.Type)
				if err != nil {
					t.Fatal(err)
				}
				writer.Write([]byte(msg.Body))
				writer.Close()

				// receive the message and check it is of the same type
				typ, reader, err := client.NextReader()
				if err != nil {
					t.Fatal(err)
				}
				if typ != msg.Type {
					t.Errorf("expected type %d, but got %d", msg.Type, typ)
				}

				// check that it is valid
				bytes, err := io.ReadAll(reader)
				if err != nil {
					t.Fatal(err)
				}
				if got := string(bytes); got != msg.Body {
					t.Errorf("expected body %q, but got %q", msg.Body, got)
				}

			}

		})
	}
}
