//spellchecker:words collection
package collection

//spellchecker:words maps slices
import (
	"cmp"
	"iter"
	"maps"
	"slices"
)

// IterSorted returns an iterator that iterates over the map in ascending order of keys.
func IterSorted[K cmp.Ordered, V any](m map[K]V) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		// get the keys of the map and sort them
		keys := make([]K, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		slices.Sort(keys)

		// and do the iteration
		for _, key := range keys {
			if !yield(key, m[key]) {
				return
			}
		}
	}
}

// MapValues creates a new map which has the same keys as m.
// The values of the map are determined by passing the old key and values into f.
func MapValues[K comparable, V1, V2 any](m map[K]V1, f func(K, V1) V2) map[K]V2 {
	if m == nil {
		return nil
	}

	M2 := make(map[K]V2, len(m))
	for k, v := range m {
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
