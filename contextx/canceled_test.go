//spellchecker:words contextx
package contextx_test

//spellchecker:words github pkglib contextx
import (
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

	// Output: context was canceled
}
