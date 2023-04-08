package collection

import (
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

// IterateSorted iterates over the the map, calling f for each element.
// Iteration is performed in ascending order of the keys.
func IterateSorted[K constraints.Ordered, V any](M map[K]V, f func(k K, v V)) {
	keys := maps.Keys(M)
	slices.Sort(keys)

	for _, key := range keys {
		f(key, M[key])
	}
}

// MapValues creates a new map which has the same keys as M.
// The values of the map are determined by passin the old key and values into f.
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
// It is the map equivalent of the built-in append function for slices.
func Append[K comparable, V any](maps ...map[K]V) (mp map[K]V) {
	// take the first map (if it is non-nil)
	if len(maps) != 0 && maps[0] != nil {
		mp = maps[0]
	} else {
		mp = map[K]V{}
	}

	// add elements from all the other maps
	for i := 1; i < len(maps); i++ {
		for k, v := range maps[i] {
			mp[k] = v
		}
	}

	// and return it
	return mp
}
