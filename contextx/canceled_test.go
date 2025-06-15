//spellchecker:words contextx
package contextx_test

//spellchecker:words github pkglib contextx
import (
	"context"
	"errors"
	"fmt"

	"github.com/tkw1536/pkglib/contextx"
)

func ExampleCanceled() {
	ctx := contextx.Canceled()

	select {
	case <-ctx.Done():
		fmt.Println("context was canceled")
	default:
		fmt.Println("context was not canceled")
	}

	isContextXCancelled := errors.Is(context.Cause(ctx), contextx.ErrCanceled)
	fmt.Printf("errors.Is(context.Cause(ctx), contextx.ErrCanceled) == %v\n", isContextXCancelled)

	// Output: context was canceled
	// errors.Is(context.Cause(ctx), contextx.ErrCanceled) == true
}
