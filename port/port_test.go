//spellchecker:word//spellchecker:words testutil
//spellchecker:words port
package port_test

//spellchecker:words context strconv pkglib port
import (
	"context"
	"fmt"
	"net"
	"strconv"

	"go.tkw01536.de/pkglib/port"
)

// ExampleFindFreePort demonstrates how to use FindFreePort.
func ExampleFindFreePort() {
	port, err := port.FindFreePort(context.Background(), "localhost")
	if err != nil {
		panic(err)
	}
	_ = port
	fmt.Println("picked a free port")

	// Output:
	// picked a free port
}

func ExampleWaitForPort() {
	// pick a random choice
	choice, err := port.FindFreePort(context.Background(), "localhost")
	if err != nil {
		panic(err)
	}
	addr := net.JoinHostPort("127.0.0.1", strconv.Itoa(choice))

	waitPortReturned := make(chan struct{})
	// wait for that port do be available
	go func() {
		defer close(waitPortReturned)

		err := port.WaitForPort(context.Background(), addr, 0)
		fmt.Printf("WaitForPort returned: %v\n", err)
	}()

	var lc net.ListenConfig
	listener, err := lc.Listen(context.Background(), "tcp", addr)
	if err != nil {
		panic(err)
	}
	<-waitPortReturned
	if err := listener.Close(); err != nil {
		panic(err)
	}
	fmt.Println("listener closed")

	// Output:
	// WaitForPort returned: <nil>
	// listener closed
}
