//spellchecker:words traversal
package traversal

//spellchecker:words rangefunc

// RangeFunc is an iterator over sequences of individual values.
// When called as seq(yield), seq calls yield(v) for each value v in the sequence,
// stopping early if yield returns false.
//
// NOTE: This corresponds to the "iter".Seq type from the rangefunc experiment.
// At some point in the future when rangefunc is stable, and generic aliases are implemented,
// this will be replaced with an alias to "iter".Seq.
// See golang issues #61405 and #46477.
type RangeFunc[V any] func(yield func(V) bool)

// Sequence creates a new iterator from the given RangeFunc.
func Sequence[T any](seq RangeFunc[T]) Iterator[T] {
	return New(func(sender Generator[T]) {
		defer sender.Return()
		seq(sender.Yield)
	})
}

// Range turns an iterator into a function compatible with the RangeFunc experiment.
func Range[T any](iter Iterator[T]) RangeFunc[T] {
	return func(yield func(T) bool) {
		defer iter.Close()

		for iter.Next() {
			if !yield(iter.Datum()) {
				break
			}
		}
	}
}
