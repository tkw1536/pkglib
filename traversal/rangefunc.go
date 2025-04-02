//spellchecker:words traversal
package traversal

import "iter"

//spellchecker:words rangefunc

// RangeFunc is a deprecated alias for [iter.Seq].
//
// Deprecated: Use [iter.Seq] instead.
type RangeFunc[V any] = iter.Seq[V]

// Sequence creates a new iterator from the given RangeFunc.
func Sequence[T any](seq iter.Seq[T]) Iterator[T] {
	return New(func(sender Generator[T]) {
		defer sender.Return()
		seq(sender.Yield)
	})
}

// Range turns a custom iterator into a native iterator.
func Range[T any](iter Iterator[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		defer func() {
			_ = iter.Close()
		}()

		for iter.Next() {
			if !yield(iter.Datum()) {
				break
			}
		}
	}
}
