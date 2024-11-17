//spellchecker:words traversal
package traversal

//spellchecker:words context
import "context"

// AsChannel returns a channel that receives each element of the underlying iterator.
// The channel and underlying iterator are closed once the iterator has no more values, or the context is cancelled.
// Note that a cancellation of the context might not be registered until the Next method of the iterator has returned.
//
// The caller should ensure that the returned channel is drained, to ensure the iterator is garbage collected.
func AsChannel[T any](it Iterator[T], ctx context.Context) <-chan T {
	out := make(chan T)
	go func(out chan<- T) {
		defer close(out)
		defer it.Close()

		for it.Next() {
			select {
			case out <- it.Datum():
			case <-ctx.Done():
				_ = it.Close() // ignore any error during close
			}
		}
	}(out)
	return out
}

// FromChannel creates a new iterator that receives values from the provided channel.
// The context is cancelled once the receiving end of the iterator requests cancellation or the input channel is exhausted.
//
// NOTE: Even if the receiving iterator end requests cancellation, the input channel will always be drained.
// This ensures that any process sending to the iterator can always continue to send values.
func FromChannel[T any](in <-chan T) (Iterator[T], context.Context) {
	ctx, cancel := context.WithCancel(context.Background())

	it := New(func(generator Generator[T]) {
		defer cancel()

		// yield elements of the generator
		for elem := range in {
			if !generator.Yield(elem) {
				cancel()
				break
			}
		}

		// we have sent everything
		generator.Return()

		// drain the channel
		for range in {
		}
	})

	return it, ctx
}
