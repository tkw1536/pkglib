package contextx

import (
	"context"
	"fmt"
	"time"
)

func ExampleAnyways() {
	const short = 100 * time.Millisecond

	// on a non-cancelled context it just behaves like short
	{
		ctx, cancel := Anyways(context.Background(), short)
		defer cancel()

		start := time.Now()
		<-ctx.Done()
		waited := time.Since(start) > short

		fmt.Println("Background() waited more than short:", waited)
	}

	// on a canceled context it delays the cancellation by the timeout
	{
		ctx, cancel := Anyways(Canceled(), short)
		defer cancel()

		start := time.Now()
		<-ctx.Done()
		waited := time.Since(start) > short

		fmt.Println("Canceled() waited more than short:", waited)
	}

	// Output: Background() waited more than short: true
	// Canceled() waited more than short: true
}
