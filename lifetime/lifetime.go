// Package lifetime provides a dependency injection framework called [Lifetime].
//
//spellchecker:words lifetime
package lifetime

//spellchecker:words reflect pkglib lazy lifetime interal souls
import (
	"reflect"

	"go.tkw01536.de/pkglib/lazy"
	"go.tkw01536.de/pkglib/lifetime/interal/souls"
)

// Lifetime implements a dependency injection framework.
// Each type of component is treated as a singleton for purposes of a single lifetime.
//
// Component must be an interface type, that should be implemented by various pointers to structs.
// Each type of struct is considered a singleton an initialized only once.
// By default components are initialized with their zero value.
// They may furthermore reference each other (even circularly).
// These references are set automatically.
//
// For this purpose they may make use of a struct called "dependencies".
// Each field in this struct may be a pointer to a different component, or a slice of a specific component subtype.
// Components may also refer to other components using a field with an `inject: "auto"` struct tag.
//
// Components must be registered using the Register function, see [Registry] for details.
// Components must be retrieved using [lifetime.Lifetime.All], [Export] or [ExportSlice].
//
// When using slices of components (e.g. in dependencies or using the All or ExportSlice methods) their order is undefined by default.
// This means that multiple lifetimes (even with the same Component and Init functions) may return components in a different order.
// It is however guaranteed that the same lifetime struct always returns components in the same order.
//
// Order of a specific slice type can be fixed by giving the slice element a method named "Rank${Typ}" with a signature func()T.
// T must be of kind int, uint, float or string.
// Slices of this type will then be sorted ascending by the appropriate "<" operator.
//
// See the examples for concrete details.
type Lifetime[Component any, InitParams any] struct {
	// Init is called on every component once it has been initialized, and all dependency references have been set.
	// There is no guarantee that the Init function has been called on dependent components.
	//
	// Init is called before any component is returned to a user.
	// There will only be one concurrent call to Init at any point.
	//
	// If Init is nil, it is not called.
	Init func(Component, InitParams)

	// Register is called by the Lifetime to register all components.
	// Register will be called at most once, and may not be nil.
	//
	// See [Registry] on how to register components.
	Register func(r *Registry[Component, InitParams])

	souls lazy.Lazy[*souls.Souls]
}

// getSouls retrieves the souls associated with this lifetime.
func (lt *Lifetime[Component, InitParams]) getSouls(params InitParams) *souls.Souls {
	return lt.souls.Get(func() *souls.Souls {
		// get the component
		if lt.Register == nil {
			panic("lt.Register is nil")
		}

		// create a context and call the register function
		context := &Registry[Component, InitParams]{
			c: reflect.TypeFor[Component](),
		}
		lt.Register(context)

		// create a new set of components
		components := make([]Component, 0, len(context.components))
		for _, init := range context.components {
			components = append(components, init(params))
		}

		// get the souls
		souls := souls.New(components)

		// call the init function on the lifetime if needed
		if lt.Init != nil {
			if err := souls.Init(); err != nil {
				panic(err)
			}
			for _, c := range components {
				lt.Init(c, params)
			}
		}

		// and return it
		return souls
	})
}

// All initializes and returns all registered components from the lifetime.
// Params is passed to all Init functions that are being called.
//
// See [Lifetime] regarding order of the exported slice.
//
// Export may be safely called concurrently with other calls retrieving components.
func (lt *Lifetime[Component, InitParams]) All(params InitParams) []Component {
	all, err := lt.getSouls(params).All(true)
	if err != nil {
		panic(err)
	}
	return all.Interface().([]Component)
}

// ExportSlice initializes and returns all components that are a ConcreteComponentType from the lifetime.
//
// ConcreteComponentType must be an interface that implements Component.
// Params is passed to all Init functions that are being called.
//
// See [Lifetime] regarding order of the exported slice.
//
// ExportSlice may be safely called concurrently with other calls retrieving components.
func ExportSlice[ConcreteComponentType any, Component any, InitParams any](
	lt *Lifetime[Component, InitParams],
	params InitParams,
) []ConcreteComponentType {
	export, err := lt.getSouls(params).ExportClass(reflect.TypeFor[ConcreteComponentType]())
	if err != nil {
		panic(err)
	}
	return export.Interface().([]ConcreteComponentType)
}

// Export initializes and returns the component of ConcreteComponentType from the lifetime.
//
// ConcreteComponentType must be an a registered component of the lifetime.
// Params is passed to all Init functions that are being called.
//
// Export may be safely called concurrently with other calls retrieving components.
func Export[ConcreteComponentType any, Component any, InitParams any](
	lt *Lifetime[Component, InitParams],
	params InitParams,
) ConcreteComponentType {
	export, err := lt.getSouls(params).Export(reflect.TypeFor[ConcreteComponentType]())
	if err != nil {
		panic(err)
	}
	return export.Interface().(ConcreteComponentType)
}
