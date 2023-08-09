package collection

import (
	"slices"

	"cmp"
)

// First returns the first value in slice for which test returns true.
// When no such value exists, returns the zero value of T.
//
// To find the index of such an element, use "slices".IndexFunc.
func First[T any](slice []T, test func(T) bool) T {
	for _, v := range slice {
		if test(v) {
			return v
		}
	}
	var v T
	return v
}

// Filter modifies slice in-place to those elements where filter returns true.
// Filter never re-allocates, invalidating the previous value of slice.
// Values in the old slice, but no longer referenced by the new slice are zeroed out.
//
// Filter guarantees that the filter function is called in element order.
// Filter guarantees that the return value is non-nil if and only if slice is non-nil.
//
// To create a new slice instead, use [FilterClone].
func Filter[T any, S ~[]T](slice S, filter func(T) bool) S {
	if slice == nil {
		return nil
	}

	var (
		results  = slice[:0] // the current result slice
		allMatch = true      // did we have all elements (0...index)?
	)
	for index, value := range slice {
		// we did not have all elements
		if !filter(value) {
			allMatch = false
			continue
		}

		if allMatch {
			results = slice[:index+1] // all elements => just re-slice
		} else {
			results = append(results, value) // append the element regularly
		}
	}

	// we need to zero out other entries
	// so that they can be picked up by gc!
	if len(results) < len(slice) {
		var zero T
		for i := len(results); i < len(slice); i++ {
			slice[i] = zero
		}
	}

	return results
}

// FilterClone behaves like [Filter], except that it generates a new slice
// and keeps the original reference slice valid.
func FilterClone[T any, S ~[]T](slice S, filter func(T) bool) (results S) {
	// if there are no elements, make a copy of the original slice
	if len(slice) == 0 {
		if slice == nil {
			return nil
		}
		return []T{}
	}

	// NOTE(twiesing): We could allocate a slice of the original capacity here.
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
	cache := make(map[T]struct{}, len(slice))

	return Filter(slice, func(t T) bool {
		// check if we saw the element before
		// seen => don't include it
		_, seen := cache[t]
		if seen {
			return false
		}

		// not seen => include
		cache[t] = struct{}{}
		return true
	})
}

// AsAny returns a new slice containing the same elements as slice, but as any.
func AsAny[T any](slice []T) []any {
	// NOTE(twiesing): This function is untested because MapSlice is tested.
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
	if len(results) < len(slice) {
		var zero T
		for i := len(results); i < len(slice); i++ {
			slice[i] = zero
		}
	}

	return results
}
