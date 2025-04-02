//spellchecker:words lifetime
package lifetime

//spellchecker:words reflect sync github pkglib lifetime interal lreflect reflectx
import (
	"fmt"
	"reflect"
	"sync"

	"github.com/tkw1536/pkglib/lifetime/interal/lreflect"
	"github.com/tkw1536/pkglib/reflectx"
)

// Registry allows registering components with a lifetime using [lifetime.Lifetime.Register].
// The order in which components are registered is independent of their dependencies.
//
// The Register function should register each component used within the lifetime.
// It should only consist of calls to [Register] and [Place].
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
func Register[Concrete any, Component any, InitParams any](context *Registry[Component, InitParams], init func(Concrete, InitParams)) {
	if context == nil {
		panic("Register: nil context passed (are you inside Lifetime.Register?)")
	}

	context.m.Lock()
	defer context.m.Unlock()

	C := reflect.TypeFor[Concrete]()
	if b, _ := lreflect.ImplementsAsStructPointer(context.c, C); !b {
		panic("Register: Attempt to register " + fmt.Sprint(C) + " as non-struct-pointer component")
	}

	// get the struct type
	S := C.Elem()

	// ensure that we haven't registered the same component yet
	if _, ok := context.components[C]; ok {
		panic("Register: Duplicate registration of " + fmt.Sprint(S))
	}

	// make sure the map is there
	if context.components == nil {
		context.components = make(map[reflect.Type]func(InitParams) Component)
	}

	// Add the init function for the component
	context.components[C] = func(ip InitParams) Component {
		// ensure that there is only one call going on at the same time
		if !context.m.TryLock() {
			panic("Register: Concurrent call detected")
		}
		defer context.m.Unlock()

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
func Place[Concrete any, Component any, InitParams any](context *Registry[Component, InitParams]) {
	Register[Concrete](context, nil)
}
