package collection

import (
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/slices"
)

// First returns the first value in slice for which test returns true.
// When no such value exists, returns the zero value of T.
func First[T any](slice []T, test func(T) bool) T {
	for _, v := range slice {
		if test(v) {
			return v
		}
	}
	var v T
	return v
}

// Any returns true if there is some value in slice for which test returns true
// and false otherwise.
func Any[T any](slice []T, test func(T) bool) bool {
	return slices.IndexFunc(slice, test) >= 0
}

// Filter modifies slice in-place to those elements where filter returns true.
// Filter never re-allocates, invalidating the previous value of slice.
// Values in the old slice, but no longer referenced by the new slice are zeroed out.
//
// To create a new slice instead, use [FilterClone].
func Filter[T any, S ~[]T](slice S, filter func(T) bool) S {
	if slice == nil {
		return nil
	}

	results := slice[:0]
	for _, value := range slice {
		if filter(value) {
			results = append(results, value)
		}
	}

	// we need to zero out other entries
	// so that they can be picked up by the GC
	var zero T
	for i := len(results); i < len(slice); i++ {
		slice[i] = zero
	}

	return results
}

// FilterClone behaves like [Filter], except that it generates a new slice
// and keeps the original reference slice valid.
func FilterClone[T any, S ~[]T](slice S, filter func(T) bool) (results S) {
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

// AsAny returns a new slice containing the same elements as slice, but as any.
func AsAny[T any](slice []T) []any {
	// NOTE(twiesing): This function is untested because MapSlice is tested.
	return MapSlice(slice, func(t T) any { return t })
}

// Partition partitions elements of s into sub-slices by passing them through f.
// Order of elements within subslices is preserved.
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
func NonSequential[T constraints.Ordered, S ~[]T](slice S, test func(prev, current T) bool) S {
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
	var zero T
	for i := len(results); i < len(slice); i++ {
		slice[i] = zero
	}

	return results
}
