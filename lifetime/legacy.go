//spellchecker:words lifetime
package lifetime

// Register is a legacy alias for [Registry.Register], which was used prior to the introduction of generic methods.
//
// Deprecated: Use [Registry.Register] instead.
//
//go:fix inline
func Register[Concrete any, Component, InitParams any](registry *Registry[Component, InitParams], init func(Concrete, InitParams)) {
	registry.Register(init)
}

// Place is a legacy alias for [Registry.Place], which was used prior to the introduction of generic methods.
//
// Deprecated: Use [Registry.Place] instead.
//
//go:fix inline
func Place[Concrete any, Component, InitParams any](registry *Registry[Component, InitParams]) {
	registry.Place[Concrete]()
}

// Export is a legacy alias for [Lifetime.Export], which was used prior to the introduction of generic methods.
//
// Deprecated: Use [Lifetime.Export] instead.
//
//go:fix inline
func Export[ConcreteComponentType any, Component any, InitParams any](
	lt *Lifetime[Component, InitParams],
	params InitParams,
) ConcreteComponentType {
	return lt.Export[ConcreteComponentType](params)
}

// ExportSlice is a legacy alias for [Lifetime.ExportSlice], which was used prior to the introduction of generic methods.
//
// Deprecated: Use [Lifetime.ExportSlice] instead.
//
//go:fix inline
func ExportSlice[ConcreteComponentType any, Component any, InitParams any](
	lt *Lifetime[Component, InitParams],
	params InitParams,
) []ConcreteComponentType {
	return lt.ExportSlice[ConcreteComponentType](params)
}
