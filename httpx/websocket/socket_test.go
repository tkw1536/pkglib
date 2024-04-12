package websocket_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/tkw1536/pkglib/httpx/websocket"
	"github.com/tkw1536/pkglib/httpx/websocket/websockettest"

	gwebsocket "github.com/gorilla/websocket"
)

// A simple server that sends data to the client.
func ExampleServer_send() {
	var server websocket.Server

	server.Handler = func(ws *websocket.Connection) {
		<-ws.WriteText("hello")
		<-ws.WriteText("world")
	}

	// The following code below is just for connection to the server.
	// It is just used to make sure that everything works.

	// create a new test server
	wss := websockettest.NewServer(&server)
	defer wss.Close()

	// create a new test client
	client, _ := wss.Dial(nil)
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

	var server websocket.Server

	// prepare a message to send
	msg := websocket.NewTextMessage("i am prepared").MustPrepare()
	server.Handler = func(ws *websocket.Connection) {
		<-ws.WritePrepared(msg)
	}

	// The following code below is just for connection to the server.
	// It is just used to make sure that everything works.

	// create an actual server
	wss := websockettest.NewServer(&server)
	defer wss.Close()

	client, _ := wss.Dial(nil)
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

// How to use subprotocols
func ExampleServer_subprotocols() {

	// setup a server that supports three subprotocols
	// and echoes back the negotiated one.
	var server websocket.Server
	server.Options.Subprotocols = []string{"a", "b", "c"}
	server.Handler = func(ws *websocket.Connection) {
		proto := ws.Subprotocol()
		if proto == "" {
			<-ws.WriteText("No subprotocol selected")
			return
		}
		<-ws.WriteText("Selected subprotocol: " + proto)
	}

	// The following code below is just for connection to the server.
	// It is just used to make sure that everything works.

	// create an actual server
	s := httptest.NewServer(&server)
	defer s.Close()

	// get the websocket url
	url := "ws" + strings.TrimPrefix(s.URL, "http")

	// connect to the dummy server, and print whatever comes back
	connectAndPrint := func(subprotocols []string) {
		var dialer gwebsocket.Dialer

		dialer.Subprotocols = subprotocols
		dialer.Proxy = http.ProxyFromEnvironment
		dialer.HandshakeTimeout = 45 * time.Second

		// Connect to the server
		client, _, err := dialer.Dial(url, nil)
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
	}

	connectAndPrint(nil)                     // no subprotocol
	connectAndPrint([]string{"b", "d"})      // mix between known and unknown
	connectAndPrint([]string{"a"})           // nothing
	connectAndPrint([]string{"e", "f", "g"}) // no supported subprotocol

	// Output: No subprotocol selected
	// Selected subprotocol: b
	// Selected subprotocol: a
	// No subprotocol selected
}

// Demonstrates how panic()ing handlers are handled handler
func ExampleServer_panic() {
	var server websocket.Server

	server.Handler = func(ws *websocket.Connection) {
		<-ws.WriteText("normal message")
		panic("test panic")
	}

	// The following code below is just for connection to the server.
	// It is just used to make sure that everything works.

	// create an actual server
	s := httptest.NewServer(&server)
	defer s.Close()

	// get the websocket url
	url := "ws" + strings.TrimPrefix(s.URL, "http")

	// Connect to the server
	client, _, err := gwebsocket.DefaultDialer.Dial(url, nil)
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
	var server websocket.Server

	done := make(chan struct{})
	server.Handler = func(ws *websocket.Connection) {
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
				<-ws.Write(msg)
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
	client, _ := wss.Dial(nil)
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
			panic("incorrect answer recevied")
		}
	}

	client.Close()
	<-done
	// Output: Handler() returned
}
