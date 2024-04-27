package collection

import (
	"slices"

	"cmp"
)

// First returns the first value in slice for which test returns true.
// When no such value exists, returns the zero value of T.
//
// To find the index of such an element, use [slices.IndexFunc].
func First[T any](slice []T, test func(T) bool) T {
	for _, v := range slice {
		if test(v) {
			return v
		}
	}
	var v T
	return v
}

// KeepFunc modifies slice in-place to those elements where filter returns true.
// KeepFunc never re-allocates, invalidating the previous value of slice.
// Values in the old slice, but no longer referenced by the new slice are zeroed out.
//
// KeepFunc guarantees that the filter function is called in element order.
// KeepFunc guarantees that the return value is non-nil if and only if slice is non-nil.
//
// To create a new slice instead, use [FilterClone].
func KeepFunc[T any, S ~[]T](slice S, filter func(T) bool) S {
	// fast case: no elements in the slice
	if len(slice) == 0 {
		return slice
	}

	// assume that all elements in the slice match.
	var (
		results  = slice // the eventual result slice, assumed to be the original slice
		allMatch = true  // true iff [0:index] matches the filter
	)
	for index, value := range slice {
		// we have a non-matching element, so we need to filter it out
		if !filter(value) {
			if allMatch {
				// no element was filtered before
				// so we need to make sure to only use special ones
				results = results[:index]
				allMatch = false
			}
			continue
		}

		// we had some elements filtered
		// so we need to append individually
		if !allMatch {
			results = append(results, value)
		}
	}

	// zero out remaining elements to re-slice
	clear(slice[len(results):])

	return results
}

// KeepFuncClone behaves like [KeepFunc], except that it generates a new slice
// and keeps the original slice valid.
func KeepFuncClone[T any, S ~[]T](slice S, filter func(T) bool) (results S) {
	// if there are no elements, make a copy of the original slice
	if len(slice) == 0 {
		if slice == nil {
			return nil
		}
		return []T{}
	}

	// NOTE: We could allocate a slice of the original capacity here.
	// But that would waste space for calls where only few operations are required.
	for _, value := range slice {
		if filter(value) {
			results = append(results, value)
		}
	}

	return
}

// MapSlice generates a new slice by passing each element of S through f.
func MapSlice[T1 any, S1 ~[]T1, T2 any](slice S1, f func(T1) T2) []T2 {
	if slice == nil {
		return nil
	}

	results := make([]T2, len(slice))
	for i, v := range slice {
		results[i] = f(v)
	}
	return results
}

// Deduplicate removes duplicates from the given slice in place.
// Elements are not reordered.
func Deduplicate[T comparable, S ~[]T](slice S) S {
	// keep a set of elements seen
	seen := make(map[T]struct{}, len(slice))

	// keep only the first copy of each element
	return KeepFunc(slice, func(t T) bool {
		_, ok := seen[t]
		seen[t] = struct{}{}
		return !ok
	})
}

// AsAny returns a new slice containing the same elements as slice, but as any.
func AsAny[T any](slice []T) []any {
	// NOTE: This function is untested because MapSlice is tested.
	return MapSlice(slice, func(t T) any { return t })
}

// Partition partitions elements of s into sub-slices by passing them through f.
// Order of elements within sub-slices is preserved.
func Partition[T any, S ~[]T, P comparable](slice S, f func(T) P) map[P]S {
	result := make(map[P]S)
	for _, v := range slice {
		key := f(v)
		result[key] = append(result[key], v)
	}
	return result
}

// NonSequential sorts slice, and then removes sequential elements for which test() returns true.
// NonSequential does not re-allocate, but uses the existing slice.
func NonSequential[T cmp.Ordered, S ~[]T](slice S, test func(prev, current T) bool) S {
	// if there is at most one element, then we can return the slice as is
	if len(slice) < 2 {
		return slice
	}

	// sort the slice and prepare a results array
	slices.Sort(slice)
	results := slice[:1]

	// do the main filter parts
	prev := results[0]
	for _, current := range slice[1:] {
		if !test(prev, current) {
			results = append(results, current)
		}
		prev = current
	}

	// we need to zero out other entries
	// so that they can be picked up by the GC
	clear(slice[len(results):])

	return results
}
