package collection

import (
	"cmp"
	"maps"
	"slices"
)

// IterateSorted iterates over the map, calling f for each element.
// Iteration is performed in ascending order of the keys.
//
// If f returns false, iteration is stopped early.
// IterateSorted returns false if the iteration was stopped early, and true otherwise.
func IterateSorted[K cmp.Ordered, V any](M map[K]V, f func(k K, v V) bool) bool {
	// get the keys of the map and sort them
	keys := make([]K, 0, len(M))
	for k := range M {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	// and do the iteration
	for _, key := range keys {
		if !f(key, M[key]) {
			return false
		}
	}
	return true
}

// MapValues creates a new map which has the same keys as M.
// The values of the map are determined by passing the old key and values into f.
func MapValues[K comparable, V1, V2 any](M map[K]V1, f func(K, V1) V2) map[K]V2 {
	if M == nil {
		return nil
	}

	M2 := make(map[K]V2, len(M))
	for k, v := range M {
		M2[k] = f(k, v)
	}
	return M2
}

// Append adds elements from all other maps into the first map.
// If the first map is nil, a new empty map is used instead.
//
// It is the map equivalent of the built-in append for slices.
func Append[K comparable, V any](mps ...map[K]V) (mp map[K]V) {
	// no maps provided => return an empty map
	if len(mps) == 0 {
		return make(map[K]V)
	}

	// use the first map as the return value
	mp = mps[0]

	// ensure that it is non-nil and has enough space for all of the elements.
	// NOTE: There could be duplicates, so we may be over-allocating here.
	if mp == nil {
		var size int
		for _, m := range mps {
			size += len(m)
		}
		mp = make(map[K]V, size)
	}

	// add elements from all the other maps
	for _, iMp := range mps[1:] {
		maps.Copy(mp, iMp)
	}

	// and return it
	return mp
}
