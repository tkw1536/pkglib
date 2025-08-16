//spellchecker:words contextx
package contextx_test

//spellchecker:words testing time synctest pkglib contextx
import (
	"testing"
	"time"

	"testing/synctest"

	"go.tkw01536.de/pkglib/contextx"
)

func TestAnyways(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		t.Helper()

		const short = 100 * time.Millisecond

		// on a non-cancelled context it just behaves like short
		{
			ctx, cancel := contextx.Anyways(t.Context(), short)
			defer cancel()

			start := time.Now()
			<-ctx.Done()
			synctest.Wait()
			waited := time.Since(start) >= short

			if !waited {
				t.Errorf("Background() waited less than short: %v", waited)
			}
		}

		// on a canceled context it delays the cancellation by the timeout
		{
			ctx, cancel := contextx.Anyways(contextx.Canceled(), short)
			defer cancel()
			_ = ctx

			start := time.Now()
			synctest.Wait()
			<-ctx.Done()
			waited := time.Since(start) >= short

			if !waited {
				t.Errorf("Canceled() waited less than short: %v", waited)
			}
		}
	})
}
