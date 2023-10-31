package lifetime

/*
type Analytics struct {
	Components map[string]*ComponentAnalytics
	Groups     map[string]*GroupAnalytics
}

type ComponentAnalytics struct {
	Type   string   // Type name
	Groups []string // groups this is contained in

	CFields map[string]string // fields with type C for which C implements component
	IFields map[string]string // fields []I where I is an interface that implements component

	DCFields map[string]string // fields of the auto field with type C for which C implements component
	DIFields map[string]string // fields of the auto field []I where I is an interface that implements component

	Methods map[string]string // Method signatures of type
}
type GroupAnalytics struct {
	Type       string   // Type name
	Components []string // Components of this Type

	Methods map[string]string // Method signatures of this interface
}

// anal writes analytics about this context to anal
func (context *InjectorContext[Component]) anal(anal *Analytics, groups []reflect.Type) {
	size := context.metaCache.Size()
	anal.Components = make(map[string]*ComponentAnalytics, size)
	anal.Groups = make(map[string]*GroupAnalytics)

	// collect all the pointers, and setup the anal.Components map!
	tpPointers := make([]reflect.Type, 0, size)
	context.metaCache.Iterate(func(meta meta.Datum[Component]) {
		tp := reflect.PointerTo(meta.Elem)
		tpPointers = append(tpPointers, tp)

		mcount := tp.NumMethod()

		anal.Components[meta.Name] = &ComponentAnalytics{
			Groups:  make([]string, 0),
			Methods: make(map[string]string, mcount),
		}
		for i := 0; i < mcount; i++ {
			method := tp.Method(i)
			anal.Components[meta.Name].Methods[method.Name] = method.Type.String()
		}
	})

	// collect interfaces to analyze
	ifaces := make([]reflect.Type, len(groups))
	copy(ifaces, groups)

	// take all of the components out of the cache
	context.metaCache.Iterate(func(meta meta.Datum[Component]) {
		anal.Components[meta.Name].Type = meta.Name
		anal.Components[meta.Name].CFields = collection.MapValues(meta.CFields, func(key string, tp reflect.Type) string {
			return reflectx.NameOf(tp.Elem())
		})
		anal.Components[meta.Name].DCFields = collection.MapValues(meta.DCFields, func(key string, tp reflect.Type) string {
			return reflectx.NameOf(tp.Elem())
		})

		anal.Components[meta.Name].IFields = collection.MapValues(meta.IFields, func(key string, iface reflect.Type) string {
			ifaces = append(ifaces, iface)
			return reflectx.NameOf(iface)
		})
		anal.Components[meta.Name].DIFields = collection.MapValues(meta.DIFields, func(key string, iface reflect.Type) string {
			ifaces = append(ifaces, iface)
			return reflectx.NameOf(iface)
		})
	})

	// and analyze all interfaces
	for _, iface := range ifaces {
		name := reflectx.NameOf(iface)
		if _, ok := anal.Groups[name]; ok {
			continue
		}

		types := collection.FilterClone(tpPointers, func(tp reflect.Type) bool {
			return tp.AssignableTo(iface)
		})

		anal.Groups[name] = &GroupAnalytics{
			Type: name,
			Components: collection.MapSlice(types, func(tp reflect.Type) string {
				cname := reflectx.NameOf(tp.Elem())
				anal.Components[cname].Groups = append(anal.Components[cname].Groups, name)
				return cname
			}),
		}

		mcount := iface.NumMethod()
		anal.Groups[name].Methods = make(map[string]string, mcount)
		for i := 0; i < mcount; i++ {
			method := iface.Method(i)
			anal.Groups[name].Methods[method.Name] = method.Type.String()
		}
	}

	for _, comp := range anal.Components {
		slices.Sort(comp.Groups)
	}
}

*/
