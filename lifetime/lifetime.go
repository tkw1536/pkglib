// Package lifetime provides a dependency injection framework n the form of a lifetime.
package lifetime

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/tkw1536/pkglib/lazy"
	"github.com/tkw1536/pkglib/lifetime/interal/lreflect"
	"github.com/tkw1536/pkglib/lifetime/interal/meta"
	"github.com/tkw1536/pkglib/reflectx"
)

// Lifetime implements a dependency injection framework.
// Each type of component is treated as a singleton for purposes of this lifetime.
//
// Component must be an interface type, that should be implemented by various pointers to structs.
// Components may reference each other, even circularly.
//
// Each type of struct is considered a singleton an initialized only once.
//
// See [Lifetime.All], [Export] and [Export].
//
// The zero value is ready to use.
type Lifetime[Component any, InitParams any] struct {
	// Init is called on every component to be initialized.
	//
	// Init is called after the component has been fully initialized, i.e. all dependencies have been set.
	// Init is called before any component is returned to a user.
	// There will only be one concurrent call to Init at any point.
	//
	// If Init is nil, it is not called.
	Init func(Component, InitParams)

	// Register is called once to register all components to be used by this lifetime.
	Register func(register *RegisterContext[Component, InitParams])

	// registry holds all the initialized components.
	registry lazy.Lazy[*meta.Registry]
}

// RegisterContext is passed to the call to register.
type RegisterContext[Component any, InitParams any] struct {
	m          sync.Mutex
	c          reflect.Type // reflectx.TypeOf[Component]
	components map[reflect.Type]func(InitParams) Component
}

// Register adds a new concrete component during a call to Register.
//
// The component is first initialized using the zero value for the struct type that is being pointed to.
// If Init is not nil, it is then passed to the Init function, along with appropriate InitParams.
//
// Context must be a context that was passed to the Register Member function.
// For each context, Register may not be called concurrently.
//
// Different contexts may safely call Register concurrently.
func Register[Concrete any, Component any, InitParams any](context *RegisterContext[Component, InitParams], Init func(Concrete, InitParams)) {
	if context == nil {
		panic("Register: nil context passed (are you inside Lifetime.Register?)")
	}

	C := reflectx.TypeFor[Concrete]()
	if b, _ := lreflect.ImplementsAsStructPointer(context.c, C); !b {
		panic("Register: Attempt to register " + fmt.Sprint(C) + " as non-struct-pointer component")
	}

	// get the struct type
	S := C.Elem()

	// ensure that we haven't registered the same component yet
	if _, ok := context.components[C]; ok {
		panic("Register: Duplicate registration of " + reflectx.NameOf(S))
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
		if Init != nil {
			Init(comp, ip)
		}

		// and return the component
		return any(comp).(Component)
	}
}

// Place is like Register, except Init is always nil.
func Place[Concrete any, Component any, InitParams any](context *RegisterContext[Component, InitParams]) {
	Register[Concrete](context, nil)
}

func (lt *Lifetime[Component, InitParams]) getRegistry(params InitParams) *meta.Registry {
	return lt.registry.Get(func() *meta.Registry {
		// get the component
		if lt.Register == nil {
			panic("All: lt.Register is nil")
		}

		// call the registration function
		context := &RegisterContext[Component, InitParams]{
			c: reflectx.TypeFor[Component](),
		}
		lt.Register(context)

		// create a new set of components
		components := make([]Component, 0, len(context.components))
		for _, init := range context.components {
			components = append(components, init(params))
		}
		return meta.NewRegistry(components)
	})
}

// All initializes or returns all components stored in this initializes.
// The order of components is undefined, but guaranteed to be consistent within a concrete lifetime.
func (lt *Lifetime[Component, InitParams]) All(Params InitParams) []Component {
	all, err := lt.getRegistry(Params).All(true)
	if err != nil {
		panic(err)
	}
	return all.Interface().([]Component)
}

// ExportComponents exports all components that are a ConcreteComponentType from the lifetime.
//
// All should be the function of the core that initializes all components.
// All should only make calls to [InitComponent].
func ExportSlice[ConcreteComponentType any, Component any, InitParams any](
	lt *Lifetime[Component, InitParams],
	Params InitParams,
) []ConcreteComponentType {
	export, err := lt.getRegistry(Params).ExportClass(reflectx.TypeFor[ConcreteComponentType]())
	if err != nil {
		panic(err)
	}
	return export.Interface().([]ConcreteComponentType)
}

// ExportComponent exports the first component that is a ConcreteComponent from the lifetime.
//
// All should be the function of the core that initializes all components.
// All should only make calls to [InitComponent].
func Export[ConcreteComponentType any, Component any, InitParams any](
	lt *Lifetime[Component, InitParams],
	Params InitParams,
) ConcreteComponentType {
	export, err := lt.getRegistry(Params).Export(reflectx.TypeFor[ConcreteComponentType]())
	if err != nil {
		panic(err)
	}
	return export.Interface().(ConcreteComponentType)
}
