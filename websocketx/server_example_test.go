//spellchecker:words websocketx
package websocketx_test

//spellchecker:words github pkglib websocketx websockettest gorilla websocket
import (
	"fmt"

	"github.com/tkw1536/pkglib/websocketx"
	"github.com/tkw1536/pkglib/websocketx/websockettest"

	"github.com/gorilla/websocket"
)

// A simple server that sends data to the client.
func ExampleServer_send() {
	var server websocketx.Server

	server.Handler = func(ws *websocketx.Connection) {
		ws.WriteText("hello")
		ws.WriteText("world")
	}

	// The following code below is just for connection to the server.
	// It is just used to make sure that everything works.

	// create a new test server
	wss := websockettest.NewServer(&server)
	defer wss.Close()

	// create a new test client
	client, _ := wss.Dial(nil, nil)
	defer client.Close()

	// print text messages
	for {
		tp, p, err := client.ReadMessage()
		if err != nil {
			return
		}

		// ignore non-text-messages
		if tp != websocket.TextMessage {
			continue
		}
		fmt.Println(string(p))
	}

	// Output: hello
	// world
}

func ExampleServer_prepared() {

	var server websocketx.Server

	// prepare a message to send
	msg := websocketx.NewTextMessage("i am prepared").MustPrepare()
	server.Handler = func(ws *websocketx.Connection) {
		ws.WritePrepared(msg)
	}

	// The following code below is just for connection to the server.
	// It is just used to make sure that everything works.

	// create an actual server
	wss := websockettest.NewServer(&server)
	defer wss.Close()

	client, _ := wss.Dial(nil, nil)
	defer client.Close()

	// print text messages
	for {
		tp, p, err := client.ReadMessage()
		if err != nil {
			return
		}

		// ignore non-text-messages
		if tp != websocket.TextMessage {
			continue
		}
		fmt.Println(string(p))
	}

	// Output: i am prepared
}

// Demonstrates how panic()ing handlers are handled handler
func ExampleServer_panic() {
	var server websocketx.Server

	server.Handler = func(ws *websocketx.Connection) {
		ws.WriteText("normal message")
		panic("test panic")
	}

	// The following code below is just for connection to the server.
	// It is just used to make sure that everything works.

	// create an actual server
	wss := websockettest.NewServer(&server)
	defer wss.Close()

	// Connect to the server
	client, _, err := websocket.DefaultDialer.Dial(wss.URL, nil)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	// print text messages
	for {
		tp, p, err := client.ReadMessage()
		if err != nil {
			return
		}

		// ignore non-text-messages
		if tp != websocket.TextMessage {
			continue
		}
		fmt.Println(string(p))
	}

	// Output: normal message
}

// A simple echo server for messages
func ExampleServer_echo() {
	// create a very simple websocket server that just echoes back messages
	var server websocketx.Server

	done := make(chan struct{})
	server.Handler = func(ws *websocketx.Connection) {
		// when finished, print that the handler exited
		// and close the done channel
		defer fmt.Println("Handler() returned")
		defer close(done)

		// read and write messages back forever
		for {
			select {
			case <-ws.Context().Done():
				return
			case msg := <-ws.Read():
				ws.Write(msg)
			}
		}
	}

	// The following code below is just for connection to the server.
	// It is just used to make sure that everything works.

	// create an actual server
	// create a new test server
	wss := websockettest.NewServer(&server)
	defer wss.Close()

	// create a new test client
	client, _ := wss.Dial(nil, nil)
	defer client.Close()

	// send a bunch of example messages
	messageCount := 1000

	// send it a lot of times
	for i := range messageCount {
		// generate an example message to send
		message := fmt.Sprintf("message %d", i)
		var kind int
		if i%2 == 0 {
			kind = websocket.BinaryMessage
		} else {
			kind = websocket.TextMessage
		}

		// write it or die
		if err := client.WriteMessage(kind, []byte(message)); err != nil {
			panic(err)
		}

		// read the message
		tp, p, err := client.ReadMessage()
		if err != nil {
			panic(err)
		}
		if tp != kind {
			panic("incorrect type received")
		}
		if string(p) != message {
			panic("incorrect answer received")
		}
	}

	client.Close()
	<-done
	// Output: Handler() returned
}
