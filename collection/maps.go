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
// the values are determined by f.
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
