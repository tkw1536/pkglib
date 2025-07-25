//spellchecker:words contextx
package contextx_test

//spellchecker:words context time github pkglib contextx
import (
	"context"
	"fmt"
	"time"

	"go.tkw01536.de/pkglib/contextx"
)

func ExampleRun() {
	// for this example, we create a "work" function that runs until the cancel function is called.
	var work func() int
	var cancel func()
	{
		done := make(chan struct{})

		work = func() int {
			fmt.Println("start working")
			<-done
			fmt.Println("done working")
			return 42
		}
		cancel = func() {
			fmt.Println("cancel called")
			close(done)
		}
	}

	// create a context that is stopped after 100 milliseconds
	ctx, ctxCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer ctxCancel()

	// and run the function with the context and explicit cancel!
	result, err := contextx.Run(ctx, func(start func()) int {
		start() // allow calling cancel immediately!

		// start the work!
		return work()
	}, cancel)

	fmt.Println(result, err)

	// Output:
	// start working
	// cancel called
	// done working
	// 42 context deadline exceeded
}
