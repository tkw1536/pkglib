package meta

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/tkw1536/pkglib/lazy"
	"github.com/tkw1536/pkglib/lifetime/interal/lreflect"
)

// Registry holds a set of (possibly initialized) components.
type Registry struct {
	all        reflect.Value // all holds all components
	allT       reflect.Type  // reflect.TypeOf(all)
	componentT reflect.Type  // allT.Elem()

	// cache for component indexes
	components map[reflect.Type]reflect.Value // map[*Component]Instance
	classes    map[reflect.Type]reflect.Value // map[Class]Index

	initErr lazy.Lazy[error] // error that occurred during init
}

func NewRegistry(all any) *Registry {
	return &Registry{
		all: reflect.ValueOf(all),
	}
}

func (r *Registry) doInit() error {
	// do an initialization
	return r.initErr.Get(func() error {
		// set allT and componentT correctly
		{
			if !r.all.IsValid() {
				return errRegistryInternalWrongAll
			}
			r.allT = r.all.Type()
			if r.allT.Kind() != reflect.Slice {
				return errRegistryInternalWrongAll
			}

			r.componentT = r.allT.Elem()
			if r.componentT.Kind() != reflect.Interface {
				return errRegistryInternalWrongAll
			}
		}

		// initialize maps for components and classes
		r.components = make(map[reflect.Type]reflect.Value, r.all.Len())
		r.classes = make(map[reflect.Type]reflect.Value)

		// iterate over all the elements
		l := r.all.Len()
		for i := 0; i < l; i += 1 {
			if err := r.initComponent(i); err != nil {
				return err
			}
		}

		return nil
	})
}

// initComponent initializes the component with the given id
func (r *Registry) initComponent(index int) error {
	// the underlying element at the given index
	elem := r.all.Index(index).Elem()
	concrete := elem.Type()

	// access the pointed to struct
	elem = elem.Elem()

	// attempt to initialize the given component metadata
	m, err := New(r.componentT, concrete)
	if err != nil {
		return err
	}

	dElem := elem.FieldByName(dependenciesFieldName)

	// assign the component fields
	for field, eType := range m.CFields {
		c, err := r.export(eType)
		if err != nil {
			return errInitField{Concrete: concrete.Elem(), InDependencies: false, Field: field, Err: err}
		}

		field := elem.FieldByName(field)
		lreflect.UnsafeSetAnyValue(field, c)
	}
	for field, eType := range m.DCFields {
		c, err := r.export(eType)
		if err != nil {
			return errInitField{Concrete: concrete.Elem(), InDependencies: true, Field: field, Err: err}
		}

		field := dElem.FieldByName(field)
		lreflect.UnsafeSetAnyValue(field, c)
	}

	// assign the interface subtypes
	for field, eType := range m.IFields {
		cs, err := r.exportClass(eType)
		if err != nil {
			return errInitField{Concrete: concrete.Elem(), InDependencies: false, Field: field, Err: err}
		}

		field := elem.FieldByName(field)
		lreflect.UnsafeSetAnyValue(field, cs)
	}
	for field, eType := range m.DIFields {
		cs, err := r.exportClass(eType)
		if err != nil {
			return errInitField{Concrete: concrete.Elem(), InDependencies: true, Field: field, Err: err}
		}

		field := dElem.FieldByName(field)
		lreflect.UnsafeSetAnyValue(field, cs)
	}

	return nil
}

type errUnregisteredComponent struct {
	T reflect.Type
}

func (eug errUnregisteredComponent) Error() string {
	return fmt.Sprintf("attempt to export un-registered component: %s", eug.T)
}

// export exports a component that is assignable to T
func (r *Registry) export(T reflect.Type) (reflect.Value, error) {
	// if we already have the component type cached, then return it
	if c, ok := r.components[T]; ok {
		return c, nil
	}

	// get the first assignable element
	c, err := lreflect.FirstAssignableInterfaceElement(r.all, T)
	if err != nil {
		return reflect.Value{}, err
	}

	// if it is nil, don't do anything with it
	if c.IsNil() {
		return reflect.Value{}, errUnregisteredComponent{T: T}
	}

	// store it in the cache and return it
	r.components[T] = c
	return c, nil
}

// exportClass exports all components assignable to interface T
func (r *Registry) exportClass(T reflect.Type) (reflect.Value, error) {
	// if we already have the class cached, then return a copy
	if clz, ok := r.classes[T]; ok {
		return lreflect.CopySlice(clz), nil
	}

	// get the class
	clz, err := lreflect.FilterSliceInterface(r.all, T)
	if err != nil {
		return reflect.Value{}, err
	}

	// store it in the cache and return it
	r.classes[T] = lreflect.CopySlice(clz)
	return clz, nil
}

// All returns the list of all components
func (r *Registry) All(copy bool) (reflect.Value, error) {
	// do the initialization
	if err := r.doInit(); err != nil {
		return reflect.Value{}, err
	}

	// if we didn't request a copy, return as is
	if !copy {
		return r.all, nil
	}

	// return a copy of the slice
	return lreflect.CopySlice(r.all), nil
}

// Export exports a specific component.
func (r *Registry) Export(T reflect.Type) (reflect.Value, error) {
	// initialize the registry
	if err := r.doInit(); err != nil {
		return reflect.Value{}, err
	}

	// ensure that we have a pointer to a struct
	if ok, err := lreflect.ImplementsAsStructPointer(r.componentT, T); err != nil || !ok {
		return reflect.Value{}, errNotPointerToStruct
	}

	// and do the export
	return r.export(T)
}

// ExportClass exports a specific component class.
func (r *Registry) ExportClass(T reflect.Type) (reflect.Value, error) {
	// initialize the registry
	if err := r.doInit(); err != nil {
		return reflect.Value{}, err
	}

	// ensure that T is a valid type that can implement a class
	if T == nil || T.Kind() != reflect.Interface || !T.Implements(r.componentT) {
		return reflect.Value{}, errComponentNotImplemented
	}

	// and export the class
	return r.exportClass(T)
}

var errRegistryInternalWrongAll = errors.New("Registry: Internal error: Wrong type for registry.all")

type errInitField struct {
	Concrete       reflect.Type
	InDependencies bool
	Field          string
	Err            error
}

func (err errInitField) Error() string {
	var fieldPrefix string
	if err.InDependencies {
		fieldPrefix = "dependencies "
	}
	return fmt.Sprintf("Type %s, %sField %s: %s", err.Concrete, fieldPrefix, err.Field, err.Err)
}

func (err errInitField) Unwrap() error {
	return err.Err
}
