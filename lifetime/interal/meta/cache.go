package meta

import (
	"reflect"
	"sync"

	"github.com/tkw1536/pkglib/reflectx"
)

// spellchecker:words reflectx

// TODO: This needs to be tested!

// Cache holds a set of Data
type Cache[Component any] struct {
	m sync.Mutex

	tpComponent reflect.Type // reflectx.TypeFor[Component]
	cache       map[reflect.Type]Datum[Component]
}

// Size returns the number of components available in this cache
func (mc *Cache[Component]) Size() int {
	if mc == nil {
		return 0
	}

	mc.m.Lock()
	defer mc.m.Unlock()

	return len(mc.cache)
}

// Iterate calls f for every component in this cache
func (mc *Cache[Component]) Iterate(f func(Datum[Component])) {
	if mc == nil || f == nil {
		return
	}

	// safely read the cache
	mc.m.Lock()
	comps := mc.cache
	mc.m.Unlock()

	// iterate over them
	for _, m := range comps {
		f(m)
	}
}

// Get creates or returns a new cache for the meta in this component
func (mc *Cache[Component]) Get(concrete reflect.Type) Datum[Component] {
	if mc == nil {
		panic("metaCache.Get: mc is nil")
	}

	mc.m.Lock()
	defer mc.m.Unlock()

	// if we already have a cache, return it
	if m, ok := mc.cache[concrete]; ok {
		return m
	}

	// create a new cache
	if mc.cache == nil {
		mc.cache = make(map[reflect.Type]Datum[Component])
	}

	// ensure that TypeFor[Component] is set
	if mc.tpComponent == nil {
		mc.tpComponent = reflectx.TypeFor[Component]()
	}

	// initialize a new datum for the given component
	m := newDatum[Component](mc.tpComponent, concrete)

	//store it in the cache
	mc.cache[concrete] = m

	// and return the meta we just created
	return m
}
