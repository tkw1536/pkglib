//spellchecker:words contextx
package contextx_test

//spellchecker:words context errors github pkglib contextx
import (
	"context"
	"errors"
	"fmt"

	"go.tkw01536.de/pkglib/contextx"
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
