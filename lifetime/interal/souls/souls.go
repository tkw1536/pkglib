// Package souls implements component storage retrieval using the [Souls] struct.
//
//spellchecker:words souls
package souls

//spellchecker:words errors math rand reflect github pkglib lazy lifetime interal lreflect
import (
	"errors"
	"fmt"
	"math/rand"
	"reflect"

	"go.tkw01536.de/pkglib/lazy"
	"go.tkw01536.de/pkglib/lifetime/interal/lreflect"
)

//spellchecker:words nolint wrapcheck

// Souls holds the set of all (possibly initialized) components.
type Souls struct {
	all        reflect.Value // all holds all components
	allT       reflect.Type  // reflect.TypeOf(all)
	componentT reflect.Type  // allT.Elem()

	// cache for component indexes
	components map[reflect.Type]reflect.Value // map[*Component]Index
	classes    map[reflect.Type]reflect.Value // map[Class]Slice

	// have we been initialized?
	initErr lazy.Lazy[error] // error that occurred during init
}

// New creates a new souls from the given slice of all components.
func New(all any) *Souls {
	return &Souls{
		all: reflect.ValueOf(all),
	}
}

// Init initializes all components.
// If this souls has already been initialized, this call is a noop.
func (r *Souls) Init() error {
	// do an initialization
	return r.initErr.Get(func() error { //nolint:wrapcheck // ignore wrapping from held error
		// set allT and componentT correctly
		{
			if !r.all.IsValid() {
				return errWrongAll
			}
			r.allT = r.all.Type()
			if r.allT.Kind() != reflect.Slice {
				return errWrongAll
			}

			r.componentT = r.allT.Elem()
			if r.componentT.Kind() != reflect.Interface {
				return errWrongAll
			}
		}

		// randomly shuffle all elements, then sort (if needed)
		l := r.all.Len()
		rand.Shuffle(l, reflect.Swapper(r.all.Interface()))
		if err := lreflect.SortSliceByRank(r.all); err != nil {
			return err //nolint:wrapcheck // don't wrap lreflect errors
		}

		// initialize maps for components and classes
		r.components = make(map[reflect.Type]reflect.Value, l)
		r.classes = make(map[reflect.Type]reflect.Value)

		// iterate over all the elements
		for i := range l {
			if err := r.initComponent(i); err != nil {
				return err
			}
		}

		return nil
	})
}

// initComponent initializes the component with the given id.
func (r *Souls) initComponent(index int) error {
	// the underlying element at the given index
	elem := r.all.Index(index).Elem()
	concrete := elem.Type()

	// access the pointed to struct
	elem = elem.Elem()

	// attempt to initialize the given component metadata
	m, err := newSoul(r.componentT, concrete)
	if err != nil {
		return err
	}

	dElem := elem.FieldByName(dependenciesFieldName)

	// assign the component fields
	for field, eType := range m.CFields {
		c, err := r.export(eType)
		if err != nil {
			return initFieldError{Concrete: concrete.Elem(), InDependencies: false, Field: field, Err: err}
		}

		field := elem.FieldByName(field)
		_ = lreflect.UnsafeSetAnyValue(field, c)
	}
	for field, eType := range m.DCFields {
		c, err := r.export(eType)
		if err != nil {
			return initFieldError{Concrete: concrete.Elem(), InDependencies: true, Field: field, Err: err}
		}

		field := dElem.FieldByName(field)
		_ = lreflect.UnsafeSetAnyValue(field, c)
	}

	// assign the interface subtypes
	for field, eType := range m.IFields {
		cs, err := r.exportClass(eType)
		if err != nil {
			return initFieldError{Concrete: concrete.Elem(), InDependencies: false, Field: field, Err: err}
		}

		field := elem.FieldByName(field)
		_ = lreflect.UnsafeSetAnyValue(field, cs)
	}
	for field, eType := range m.DIFields {
		cs, err := r.exportClass(eType)
		if err != nil {
			return initFieldError{Concrete: concrete.Elem(), InDependencies: true, Field: field, Err: err}
		}

		field := dElem.FieldByName(field)
		_ = lreflect.UnsafeSetAnyValue(field, cs)
	}

	return nil
}

type unregisteredComponentError struct {
	T reflect.Type
}

func (eug unregisteredComponentError) Error() string {
	return fmt.Sprintf("attempt to export un-registered component: %s", eug.T)
}

// export exports a component that is assignable to typ.
func (r *Souls) export(typ reflect.Type) (reflect.Value, error) {
	// if we already have the component type cached, then return it
	if c, ok := r.components[typ]; ok {
		return c, nil
	}

	// get the first assignable element
	c, err := lreflect.FirstAssignableInterfaceElement(r.all, typ)
	if err != nil {
		return reflect.Value{}, err //nolint:wrapcheck // don't wrap lreflect errors
	}

	// if it is nil, don't do anything with it
	if c.IsNil() {
		return reflect.Value{}, unregisteredComponentError{T: typ}
	}

	// store it in the cache and return it
	r.components[typ] = c
	return c, nil
}

// exportClass exports all components assignable to interface T.
func (r *Souls) exportClass(typ reflect.Type) (reflect.Value, error) {
	// if we already have the class cached, then return a copy
	if clz, ok := r.classes[typ]; ok {
		return lreflect.CopySlice(clz), nil
	}

	// get the class
	clz, err := lreflect.FilterSliceInterface(r.all, typ)
	if err != nil {
		return reflect.Value{}, err //nolint:wrapcheck // don't wrap lreflect errors
	}

	// sort the slice by rank
	if err := lreflect.SortSliceByRank(clz); err != nil {
		return reflect.Value{}, err //nolint:wrapcheck // don't wrap lreflect errors
	}

	// store it in the cache and return it
	r.classes[typ] = lreflect.CopySlice(clz)
	return clz, nil
}

// All returns the list of all components.
// cp indicates if a copy should be made.
func (r *Souls) All(cp bool) (reflect.Value, error) {
	// do the initialization
	if err := r.Init(); err != nil {
		return reflect.Value{}, err
	}

	// if we didn't request a copy, return as is
	if !cp {
		return r.all, nil
	}

	// return a copy of the slice
	return lreflect.CopySlice(r.all), nil
}

// Export exports a specific component.
func (r *Souls) Export(typ reflect.Type) (reflect.Value, error) {
	// initialize the registry
	if err := r.Init(); err != nil {
		return reflect.Value{}, err
	}

	// ensure that we have a pointer to a struct
	if ok, err := lreflect.ImplementsAsStructPointer(r.componentT, typ); err != nil || !ok {
		return reflect.Value{}, errNotPointerToStruct
	}

	// and do the export
	return r.export(typ)
}

// ExportClass exports a specific component class.
func (r *Souls) ExportClass(typ reflect.Type) (reflect.Value, error) {
	// initialize the registry
	if err := r.Init(); err != nil {
		return reflect.Value{}, err
	}

	// ensure that T is a valid type that can implement a class
	if typ == nil || typ.Kind() != reflect.Interface || !typ.Implements(r.componentT) {
		return reflect.Value{}, errComponentNotImplemented
	}

	// and export the class
	return r.exportClass(typ)
}

var errWrongAll = errors.New("wrong type for souls.all")

type initFieldError struct {
	Concrete       reflect.Type
	InDependencies bool
	Field          string
	Err            error
}

func (err initFieldError) Error() string {
	var fieldPrefix string
	if err.InDependencies {
		fieldPrefix = "dependencies "
	}
	return fmt.Sprintf("Type %s, %sField %s: %s", err.Concrete, fieldPrefix, err.Field, err.Err)
}

func (err initFieldError) Unwrap() error {
	return err.Err
}
