package contextx

import "fmt"

func ExampleCanceled() {
	ctx := Canceled()

	select {
	case <-ctx.Done():
		fmt.Println("context was canceled")
	default:
		fmt.Println("context was not canceled")
	}

	// Output: context was canceled
}
