//spellchecker:words lreflect
package lreflect

//spellchecker:words reflect sort
import (
	"reflect"
	"sort"
)

// SortSliceByRank sorts slice by a magic rank function found on the element type of slice.
// Slice must be a slice of some value.
//
// The rank function has to be called "Rank${ElementType}" and have signature func()T.
// T must be comparable, meaning it is of kind int, uint, float or string.
// If no such function exists on the element type of slice, it is returned unchanged.
//
// The sort performed is guaranteed to be stable, meaning to equally do not change positions.
func SortSliceByRank(slice reflect.Value) error {
	// check that we have some valid value
	if !slice.IsValid() {
		return errInvalidValue("slice")
	}

	// check that we have a slice type
	S := slice.Type()
	if S.Kind() != reflect.Slice {
		return errNoSlice("slice")
	}

	// get the name of the rank method
	E := S.Elem()

	// get the name of the rank method
	name, t, ok := getRankMethod(E)
	if ok {
		// create a new sort interface
		var si sortIf
		si.swapper = reflect.Swapper(slice.Interface())
		si.less = t.LessMethod()

		// compute the rank of each value
		si.values = make([]reflect.Value, slice.Len())
		for i := range si.values {
			si.values[i] = slice.Index(i).MethodByName(name).Call(nil)[0]
		}

		// sort the sort interface
		sort.Stable(&si)
	}
	return nil
}

// rankTyp describes the type of a rank method
type rankTyp string

const (
	rankTypeInvalid rankTyp = ""
	rankTypeInt     rankTyp = "int"
	rankTypeUint    rankTyp = "uint"
	rankTypeFloat   rankTyp = "float"
	rankTypeString  rankTyp = "string"
)

// LessMethod returns a method that compare two reflect values of the given rank method
func (t rankTyp) LessMethod() func(l, r reflect.Value) bool {
	switch t {
	case rankTypeInt:
		return func(l, r reflect.Value) bool { return l.Int() < r.Int() }
	case rankTypeUint:
		return func(l, r reflect.Value) bool { return l.Uint() < r.Uint() }
	case rankTypeFloat:
		return func(l, r reflect.Value) bool { return l.Float() < r.Float() }
	case rankTypeString:
		return func(l, r reflect.Value) bool { return l.Interface().(string) < r.Interface().(string) }
	}
	return nil
}

// getRankMethod returns the rank method of the given type (if any)
func getRankMethod(typ reflect.Type) (string, rankTyp, bool) {
	// get the name of the method
	name := typ.Name()
	if name == "" {
		return "", rankTypeInvalid, false
	}
	name = "Rank" + name

	// check that it exists
	m, ok := typ.MethodByName(name)
	if !ok {
		return "", rankTypeInvalid, false
	}

	// offset for number of parameters if a receiver is included
	rOffset := 0
	if typ.Kind() != reflect.Interface {
		rOffset = 1
	}

	// it must be func()T
	if m.Type.NumIn() != 0+rOffset || m.Type.NumOut() != 1 {
		return "", rankTypeInvalid, false
	}

	// where rTyp is a comparable type
	var rTyp rankTyp
	switch m.Type.Out(0).Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		rTyp = rankTypeInt
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		rTyp = rankTypeUint
	case reflect.Float32, reflect.Float64:
		rTyp = rankTypeFloat
	case reflect.String:
		rTyp = rankTypeString
	default:
		return "", rankTypeInvalid, false
	}

	// and return that we did
	return name, rTyp, true
}

var _ sort.Interface = (*sortIf)(nil)

type sortIf struct {
	swapper func(i, j int)
	values  []reflect.Value
	less    func(l, r reflect.Value) bool
}

func (rs *sortIf) Len() int {
	return len(rs.values)
}
func (rs *sortIf) Swap(i, j int) {
	rs.swapper(i, j)                                        // swap original
	rs.values[i], rs.values[j] = rs.values[j], rs.values[i] // swap values
}
func (rs *sortIf) Less(i, j int) bool {
	return rs.less(rs.values[i], rs.values[j])
}
