package websockettest_test

import (
	"io"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/tkw1536/pkglib/websocketx/websockettest"
)

func TestNewServer(t *testing.T) {
	// Create a simple echo server
	echo := websockettest.NewServer(websockettest.NewHandler(func(conn *websocket.Conn) {
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
				{Type: websocket.TextMessage, Body: "hello"},
			},
		},

		{
			"lots of messages",
			[]Message{
				{Type: websocket.TextMessage, Body: "hello"},
				{Type: websocket.BinaryMessage, Body: "world"},
				{Type: websocket.TextMessage, Body: "how"},
				{Type: websocket.BinaryMessage, Body: "are"},
				{Type: websocket.TextMessage, Body: "you"},
				{Type: websocket.BinaryMessage, Body: "doing"},
				{Type: websocket.TextMessage, Body: "today"},
			},
		},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			// create a new client
			client, _ := echo.Dial(nil, nil)

			for _, msg := range tt.Messages {
				// write the message to the network
				writer, err := client.NextWriter(msg.Type)
				if err != nil {
					t.Fatal(err)
				}
				_, _ = writer.Write([]byte(msg.Body))
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
