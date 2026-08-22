//spellchecker:words lifetime
package lifetime

//spellchecker:words reflect sync pkglib lifetime interal lreflect reflectx
import (
	"fmt"
	"reflect"
	"sync"

	"go.tkw01536.de/pkglib/lifetime/interal/lreflect"
	"go.tkw01536.de/pkglib/reflectx"
)

// Registry allows registering components with a lifetime using [lifetime.Lifetime.Register].
// The order in which components are registered is independent of their dependencies.
//
// The Register function should register each component used within the lifetime.
// It should only consist of calls to [Registry.Register] and [Registry.Place].
// It must not maintain a reference to the registry beyond the function call.
type Registry[Component any, InitParams any] struct {
	m sync.Mutex

	c          reflect.Type // reflectx.TypeOf[Component]
	components map[reflect.Type]func(InitParams) Component
}

// Register registers a concrete component with a registry.
//
// During component initialization, the component is first initialized to its' zero value.
// Then the Init parameter is called with the concrete component.
// During the call to Init the dependencies are not yet initialized or set.
// Init may be nil, in which case it is not called.
//
// Register may only be called from within a call to [lifetime.Lifetime.Register].
// Register may be safely called concurrently.
func (registry *Registry[Component, InitParams]) Register[Concrete any](init func(Concrete, InitParams)) {
	if registry == nil {
		panic("Register: nil context passed (are you inside Lifetime.Register?)")
	}

	registry.m.Lock()
	defer registry.m.Unlock()

	C := reflect.TypeFor[Concrete]()
	if b, _ := lreflect.ImplementsAsStructPointer(registry.c, C); !b {
		panic("Register: Attempt to register " + fmt.Sprint(C) + " as non-struct-pointer component")
	}

	// get the struct type
	S := C.Elem()

	// ensure that we haven't registered the same component yet
	if _, ok := registry.components[C]; ok {
		panic("Register: Duplicate registration of " + fmt.Sprint(S))
	}

	// make sure the map is there
	if registry.components == nil {
		registry.components = make(map[reflect.Type]func(InitParams) Component)
	}

	// Add the init function for the component
	registry.components[C] = func(ip InitParams) Component {
		// ensure that there is only one call going on at the same time
		if !registry.m.TryLock() {
			panic("Register: Concurrent call detected")
		}
		defer registry.m.Unlock()

		defer func() {
			if value := recover(); value != nil {
				panic(fmt.Sprintf("Register init for %s panicked: %v", reflectx.NameOf(S), value))
			}
		}()

		// make the component
		comp := reflect.New(S).Interface().(Concrete)

		// call the init function (if any)
		if init != nil {
			init(comp, ip)
		}

		// and return the component
		return any(comp).(Component)
	}
}

// Place is like [Register], except that the Init function is always nil.
//
// As such, the same restrictions as for Place apply.
// Place may only be called from within a call to [lifetime.Lifetime.Register].
// Place may be safely called concurrently, even with calls to [Register].
func (registry *Registry[Component, InitParams]) Place[Concrete any]() {
	registry.Register[Concrete](nil)
}
