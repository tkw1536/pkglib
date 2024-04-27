package websocketx_test

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tkw1536/pkglib/websocketx"
	"github.com/tkw1536/pkglib/websocketx/websockettest"
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
			var server websocketx.Server
			server.Options.Subprotocols = tt.ServerProto

			// record the protocol that we got
			var gotProto string
			done := make(chan struct{})
			server.Handler = func(c *websocketx.Connection) {
				defer close(done)
				gotProto = c.Subprotocol()
			}

			// create a wss server
			wss := websockettest.NewServer(&server)
			defer wss.Close()

			c, _ := wss.Dial(func(d *websocket.Dialer) {
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

func TestServer_RequireProtocols(t *testing.T) {
	for _, tt := range []struct {
		Name               string
		ServerProtocols    []string
		ClientProtocols    []string
		WantClientProtocol string
	}{
		{
			Name:               "cannot connect with no subprotocols",
			ServerProtocols:    []string{"my-protocol"},
			ClientProtocols:    []string{},
			WantClientProtocol: "",
		},
		{
			Name:               "should connect with known protocol",
			ServerProtocols:    []string{"my-protocol"},
			ClientProtocols:    []string{"my-protocol"},
			WantClientProtocol: "my-protocol",
		},
		{
			Name:               "should connect with one known protocol",
			ServerProtocols:    []string{"my-protocol"},
			ClientProtocols:    []string{"i-don't-known-this", "my-protocol", "other_protocol"},
			WantClientProtocol: "my-protocol",
		},
		{
			Name:               "should connect with first known protocol",
			ServerProtocols:    []string{"a-protocol", "b-protocol"},
			ClientProtocols:    []string{"a-protocol"},
			WantClientProtocol: "a-protocol",
		},
		{
			Name:               "should connect with second known protocol",
			ServerProtocols:    []string{"a-protocol", "b-protocol"},
			ClientProtocols:    []string{"b-protocol"},
			WantClientProtocol: "b-protocol",
		},
		{
			Name:               "should connect with all known protocols",
			ServerProtocols:    []string{"a-protocol", "b-protocol"},
			ClientProtocols:    []string{"c-protocol", "b-protocol", "a-protocol"},
			WantClientProtocol: "a-protocol",
		},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			// setup the subprotocols and require them
			var server websocketx.Server
			server.Options.Subprotocols = tt.ServerProtocols
			server.RequireProtocols()

			// handler just returns the selected protocol
			server.Handler = func(ws *websocketx.Connection) {
				ws.WriteText("selected protocol: " + ws.Subprotocol())
			}

			// create a new test server
			wss := websockettest.NewServer(&server)
			defer wss.Close()

			// if we shouldn't connect test that we didn't
			if tt.WantClientProtocol == "" {
				conn, _, err := websocket.DefaultDialer.Dial(wss.URL, nil)
				if err == websocket.ErrBadHandshake {
					return
				}
				t.Error("connection attempt did not get a bad handshake")
				if err != nil {
					return
				}
				defer conn.Close()

				return
			}

			// create a new client
			client, _ := wss.Dial(func(d *websocket.Dialer) {
				d.Subprotocols = tt.ClientProtocols
			}, nil)
			defer client.Close()

			// read the next text message
			tp, p, err := client.ReadMessage()
			if err != nil {
				t.Error("connection failed with error", err)
				return
			}
			if tp != websocket.TextMessage {
				t.Error("did not receive text message")
				return
			}
			if string(p) != "selected protocol: "+tt.WantClientProtocol {
				t.Errorf("got wrong message from server")
			}

		})
	}
}
func TestServer_timeout(t *testing.T) {
	// expect to read a message before the timeout expires
	// NOTE: This must be smaller than the server timeout
	timeout := 500 * time.Millisecond
	if timeout >= testServerTimeout {
		panic("timeout is too big, pick a smaller one")
	}

	testServer(t, func(server *websocketx.Server) websocketx.Handler {
		server.Options.ReadInterval = timeout
		server.Options.PingInterval = testServerTimeout // don't send pings (which a sane client would respond to)

		// the handler just wait for the connection to close on it's own
		return func(c *websocketx.Connection) {
			<-c.Context().Done()
		}
	}, func(c *websocket.Conn, _ *websocketx.Server) {
		// don't send a message during the timeout
		time.Sleep(timeout)
	})
}

const testServerTimeout = time.Minute

// testServer create a new testing server and initiates a client
func testServer(t *testing.T, initHandler func(server *websocketx.Server) websocketx.Handler, doClient func(client *websocket.Conn, server *websocketx.Server)) {
	t.Helper()

	// create the server
	var server websocketx.Server

	// have the test code setup the handler
	var handler websocketx.Handler
	if initHandler != nil {
		handler = initHandler(&server)
	}
	if handler == nil {
		handler = func(c *websocketx.Connection) { <-c.Context().Done() }
	}
	if server.Handler != nil {
		panic("initHandler set server.Handler (wrong test code: return it instead)")
	}

	// update the actual handler
	done := make(chan struct{})
	server.Handler = func(c *websocketx.Connection) {
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

// TestServer_concurrent checks that writing and sending
// a lot of concurrent messages work, even across goroutines.
//
// This test may take longer on some older systems and
// should be skipped in short mode.
func TestServer_concurrent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long concurrent test")
	}

	// Concurrency determines the number of concurrent messages
	// to read and write to the channel.
	concurrency := 100_000

	// messages received on the server side
	var server_m sync.Mutex
	got_server := make(map[int]struct{}, concurrency)

	// messages received on the client side
	got_client := make(map[int]struct{}, concurrency)

	// collect collects a message and the given error
	collect := func(side string, m map[int]struct{}, l *sync.Mutex, err error, body []byte) {
		if err != nil {
			t.Errorf("%s failed to parse message: %v", side, err)
			return
		}

		// parse it as an int
		id, err := strconv.Atoi(string(body))
		if err != nil {
			t.Errorf("%s failed to decode body (got: %s)", side, string(body))
			return
		}

		// check that we haven't received a message with
		// this id yet and record that we did
		if l != nil {
			l.Lock()
			defer l.Unlock()
		}

		i := int(id)
		if _, ok := m[i]; ok {
			t.Errorf("%s received %d more than once", side, i)
			return
		}
		m[i] = struct{}{}
	}

	check := func(side string, m map[int]struct{}) {
		for i := range m {
			if _, ok := m[i]; !ok {
				t.Errorf("%s did not receive %d", side, i)
			}
		}
	}

	testServer(t, func(server *websocketx.Server) websocketx.Handler {
		return func(c *websocketx.Connection) {
			var wg sync.WaitGroup

			// receive all the ints
			wg.Add(concurrency)
			for range concurrency {
				go func() {
					defer wg.Done()

					// read the body
					msg, ok := <-c.Read()
					if !ok {
						t.Error("server failed to receive message")
						return
					}

					collect("server", got_server, &server_m, nil, msg.Body)
				}()
			}

			// write all the ints
			wg.Add(concurrency)
			for i := range concurrency {
				i := i
				go func() {
					defer wg.Done()
					c.WriteText(strconv.Itoa(i))
				}()
			}

			wg.Wait()

			check("server", got_server)
		}
	}, func(client *websocket.Conn, server *websocketx.Server) {
		var wg sync.WaitGroup
		wg.Add(2)

		// send all the messages
		go func() {
			defer wg.Done()

			for i := range concurrency {
				if err := client.WriteMessage(websocket.TextMessage, []byte(strconv.Itoa(i))); err != nil {
					t.Errorf("client failed to send message: %v", err)
				}
			}
		}()

		// receive the integers
		go func() {
			defer wg.Done()

			for range concurrency {
				_, msg, err := client.ReadMessage()
				collect("client", got_client, nil, err, msg)
			}
		}()

		wg.Wait()

		check("client", got_client)
	})

}

func TestServer_ReadLimit(t *testing.T) {
	const (
		readLimit       = 10 * 1024
		biggerThanLimit = readLimit + 10
	)

	testServer(t, func(server *websocketx.Server) websocketx.Handler {
		server.Options.ReadLimit = readLimit

		return func(c *websocketx.Connection) {
			select {
			case <-c.Read():
				t.Error("read large package unexpectedly")
			case <-c.Context().Done():
				/* closed connection */
			}
		}
	}, func(client *websocket.Conn, _ *websocketx.Server) {
		// simply send a big message
		big := make([]byte, biggerThanLimit)
		_ = client.WriteMessage(websocket.TextMessage, big)
	})
}
