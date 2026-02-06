package runtime

import (
	"fmt"
	"strings"
)

// Attribute access

func (vm *VM) getAttr(obj Value, name string) (Value, error) {
	switch o := obj.(type) {
	case *PyGenerator:
		gen := o
		switch name {
		case "__iter__":
			return &PyBuiltinFunc{Name: "generator.__iter__", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				return gen, nil
			}}, nil
		case "__next__":
			return &PyBuiltinFunc{Name: "generator.__next__", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				val, done, err := vm.GeneratorSend(gen, None)
				if err != nil {
					return nil, err
				}
				if done {
					return nil, &PyException{TypeName: "StopIteration", Message: ""}
				}
				return val, nil
			}}, nil
		case "send":
			return &PyBuiltinFunc{Name: "generator.send", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				var sendVal Value = None
				if len(args) > 0 {
					sendVal = args[0]
				}
				val, done, err := vm.GeneratorSend(gen, sendVal)
				if err != nil {
					return nil, err
				}
				if done {
					return nil, &PyException{TypeName: "StopIteration", Message: ""}
				}
				return val, nil
			}}, nil
		case "throw":
			return &PyBuiltinFunc{Name: "generator.throw", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				var excType, excValue Value = &PyString{Value: "Exception"}, None
				if len(args) > 0 {
					excType = args[0]
				}
				if len(args) > 1 {
					excValue = args[1]
				}
				val, done, err := vm.GeneratorThrow(gen, excType, excValue)
				if err != nil {
					return nil, err
				}
				if done {
					return nil, &PyException{TypeName: "StopIteration", Message: ""}
				}
				return val, nil
			}}, nil
		case "close":
			return &PyBuiltinFunc{Name: "generator.close", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				err := vm.GeneratorClose(gen)
				if err != nil {
					return nil, err
				}
				return None, nil
			}}, nil
		}
		return nil, fmt.Errorf("'generator' object has no attribute '%s'", name)

	case *PyCoroutine:
		coro := o
		switch name {
		case "__await__":
			return &PyBuiltinFunc{Name: "coroutine.__await__", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				return coro, nil
			}}, nil
		case "send":
			return &PyBuiltinFunc{Name: "coroutine.send", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				var sendVal Value = None
				if len(args) > 0 {
					sendVal = args[0]
				}
				val, done, err := vm.CoroutineSend(coro, sendVal)
				if err != nil {
					return nil, err
				}
				if done {
					return nil, &PyException{TypeName: "StopIteration", Message: ""}
				}
				return val, nil
			}}, nil
		case "throw":
			return &PyBuiltinFunc{Name: "coroutine.throw", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				var excType, excValue Value = &PyString{Value: "Exception"}, None
				if len(args) > 0 {
					excType = args[0]
				}
				if len(args) > 1 {
					excValue = args[1]
				}
				val, done, err := vm.CoroutineThrow(coro, excType, excValue)
				if err != nil {
					return nil, err
				}
				if done {
					return nil, &PyException{TypeName: "StopIteration", Message: ""}
				}
				return val, nil
			}}, nil
		case "close":
			return &PyBuiltinFunc{Name: "coroutine.close", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				coro.State = GenClosed
				return None, nil
			}}, nil
		}
		return nil, fmt.Errorf("'coroutine' object has no attribute '%s'", name)

	case *PyModule:
		if val, ok := o.Dict[name]; ok {
			return val, nil
		}
		return nil, fmt.Errorf("module '%s' has no attribute '%s'", o.Name, name)
	case *PyUserData:
		// Look up method in metatable by type name
		if o.Metatable != nil {
			// Find __type__ key in metatable (iterate because Value keys use pointers)
			var typeName string
			for k, v := range o.Metatable.Items {
				if ks, ok := k.(*PyString); ok && ks.Value == "__type__" {
					typeName = vm.str(v)
					break
				}
			}
			if typeName != "" {
				mt := typeMetatables[typeName]
				if mt != nil {
					// Check properties first (like Python @property - called automatically)
					if mt.Properties != nil {
						if propGetter, ok := mt.Properties[name]; ok {
							// Call the property getter with userdata as arg 1
							ud := o
							oldStack := vm.frame.Stack
							oldSP := vm.frame.SP
							vm.frame.Stack = make([]Value, 17)
							vm.frame.Stack[0] = ud
							vm.frame.SP = 1
							n := propGetter(vm)
							var result Value = None
							if n > 0 {
								// Result was pushed onto stack after ud
								result = vm.frame.Stack[vm.frame.SP-1]
							}
							vm.frame.Stack = oldStack
							vm.frame.SP = oldSP
							return result, nil
						}
					}
					if method, ok := mt.Methods[name]; ok {
						// Capture the userdata and method in closure
						ud := o
						m := method
						// Return a bound method that includes the userdata as first arg
						return &PyGoFunc{
							Name: name,
							Fn: func(vm *VM) int {
								// Shift stack to insert userdata as first argument
								top := vm.GetTop()
								newStack := make([]Value, top+17) // Extra space for stack operations
								newStack[0] = ud
								for i := 0; i < top; i++ {
									newStack[i+1] = vm.Get(i + 1)
								}
								vm.frame.Stack = newStack
								vm.frame.SP = top + 1
								return m(vm)
							},
						}, nil
					}
				}
			}
		}
		return nil, fmt.Errorf("'%s' object has no attribute '%s'", vm.typeName(obj), name)
	case *PyProperty:
		prop := o
		switch name {
		case "setter":
			return &PyBuiltinFunc{
				Name: "property.setter",
				Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
					if len(args) != 1 {
						return nil, fmt.Errorf("setter() takes exactly 1 argument")
					}
					return &PyProperty{
						Fget: prop.Fget,
						Fset: args[0],
						Fdel: prop.Fdel,
						Doc:  prop.Doc,
					}, nil
				},
			}, nil
		case "deleter":
			return &PyBuiltinFunc{
				Name: "property.deleter",
				Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
					if len(args) != 1 {
						return nil, fmt.Errorf("deleter() takes exactly 1 argument")
					}
					return &PyProperty{
						Fget: prop.Fget,
						Fset: prop.Fset,
						Fdel: args[0],
						Doc:  prop.Doc,
					}, nil
				},
			}, nil
		case "getter":
			return &PyBuiltinFunc{
				Name: "property.getter",
				Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
					if len(args) != 1 {
						return nil, fmt.Errorf("getter() takes exactly 1 argument")
					}
					return &PyProperty{
						Fget: args[0],
						Fset: prop.Fset,
						Fdel: prop.Fdel,
						Doc:  prop.Doc,
					}, nil
				},
			}, nil
		case "fget":
			if prop.Fget != nil {
				return prop.Fget, nil
			}
			return None, nil
		case "fset":
			if prop.Fset != nil {
				return prop.Fset, nil
			}
			return None, nil
		case "fdel":
			if prop.Fdel != nil {
				return prop.Fdel, nil
			}
			return None, nil
		case "__doc__":
			return &PyString{Value: prop.Doc}, nil
		}
		return nil, fmt.Errorf("'property' object has no attribute '%s'", name)
	case *PySuper:
		// super() proxy - look up attribute in MRO starting after ThisClass
		if o.Instance == nil {
			return nil, fmt.Errorf("'super' object has no attribute '%s'", name)
		}

		// Get the MRO to search
		var mro []*PyClass
		var instance Value = o.Instance
		if inst, ok := o.Instance.(*PyInstance); ok {
			mro = inst.Class.Mro
		} else if cls, ok := o.Instance.(*PyClass); ok {
			mro = cls.Mro
			instance = cls
		}

		// Search MRO starting from StartIdx
		for i := o.StartIdx; i < len(mro); i++ {
			cls := mro[i]
			if val, ok := cls.Dict[name]; ok {
				// Handle classmethod - bind class
				if cm, ok := val.(*PyClassMethod); ok {
					if fn, ok := cm.Func.(*PyFunction); ok {
						if inst, ok := o.Instance.(*PyInstance); ok {
							return &PyMethod{Func: fn, Instance: inst.Class}, nil
						}
						return &PyMethod{Func: fn, Instance: instance}, nil
					}
				}

				// Handle staticmethod - return unwrapped function
				if sm, ok := val.(*PyStaticMethod); ok {
					return sm.Func, nil
				}

				// Handle property - call fget with the original instance
				if prop, ok := val.(*PyProperty); ok {
					if prop.Fget == nil {
						return nil, fmt.Errorf("property '%s' has no getter", name)
					}
					return vm.call(prop.Fget, []Value{instance}, nil)
				}

				// Bind method if it's a function - bind to original instance
				if fn, ok := val.(*PyFunction); ok {
					return &PyMethod{Func: fn, Instance: instance}, nil
				}
				return val, nil
			}
		}
		return nil, fmt.Errorf("'super' object has no attribute '%s'", name)

	case *PyInstance:
		// Descriptor protocol: First check class MRO for data descriptors (property with setter)
		// Data descriptors take precedence over instance dict
		for _, cls := range o.Class.Mro {
			if val, ok := cls.Dict[name]; ok {
				if prop, ok := val.(*PyProperty); ok {
					// Property is a data descriptor - call fget
					if prop.Fget == nil {
						return nil, fmt.Errorf("property '%s' has no getter", name)
					}
					return vm.call(prop.Fget, []Value{obj}, nil)
				}
			}
		}

		// Then check instance dict
		if val, ok := o.Dict[name]; ok {
			return val, nil
		}

		// Then check class MRO for methods/attributes (non-data descriptors)
		for _, cls := range o.Class.Mro {
			if val, ok := cls.Dict[name]; ok {
				// Handle classmethod - bind class instead of instance
				if cm, ok := val.(*PyClassMethod); ok {
					if fn, ok := cm.Func.(*PyFunction); ok {
						return &PyMethod{Func: fn, Instance: o.Class}, nil
					}
					// For non-PyFunction callables, return a wrapper
					return &PyBuiltinFunc{
						Name: "bound_classmethod",
						Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
							newArgs := make([]Value, len(args)+1)
							newArgs[0] = o.Class
							copy(newArgs[1:], args)
							return vm.call(cm.Func, newArgs, kwargs)
						},
					}, nil
				}

				// Handle staticmethod - return unwrapped function
				if sm, ok := val.(*PyStaticMethod); ok {
					return sm.Func, nil
				}

				// Bind method if it's a function
				if fn, ok := val.(*PyFunction); ok {
					return &PyMethod{Func: fn, Instance: obj}, nil
				}
				return val, nil
			}
		}
		return nil, fmt.Errorf("'%s' object has no attribute '%s'", o.Class.Name, name)
	case *PyClass:
		// Handle special class attributes
		switch name {
		case "__mro__":
			// Return the MRO as a tuple
			mroItems := make([]Value, len(o.Mro))
			for i, cls := range o.Mro {
				mroItems[i] = cls
			}
			return &PyTuple{Items: mroItems}, nil
		case "__bases__":
			// Return the direct base classes as a tuple
			baseItems := make([]Value, len(o.Bases))
			for i, base := range o.Bases {
				baseItems[i] = base
			}
			return &PyTuple{Items: baseItems}, nil
		case "__name__":
			return &PyString{Value: o.Name}, nil
		case "__dict__":
			// Return a copy of the class dict
			dictCopy := make(map[Value]Value)
			for k, v := range o.Dict {
				dictCopy[&PyString{Value: k}] = v
			}
			return &PyDict{Items: dictCopy}, nil
		}
		// Check class dict and MRO
		for _, cls := range o.Mro {
			if val, ok := cls.Dict[name]; ok {
				// Handle classmethod - bind with class
				if cm, ok := val.(*PyClassMethod); ok {
					if fn, ok := cm.Func.(*PyFunction); ok {
						return &PyMethod{Func: fn, Instance: o}, nil
					}
					// For non-PyFunction callables, return a wrapper
					return &PyBuiltinFunc{
						Name: "bound_classmethod",
						Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
							newArgs := make([]Value, len(args)+1)
							newArgs[0] = o
							copy(newArgs[1:], args)
							return vm.call(cm.Func, newArgs, kwargs)
						},
					}, nil
				}

				// Handle staticmethod - return unwrapped function
				if sm, ok := val.(*PyStaticMethod); ok {
					return sm.Func, nil
				}

				// Handle property on class access - return the property object itself
				// (In Python, accessing a property on the class returns the property object)
				if _, ok := val.(*PyProperty); ok {
					return val, nil
				}

				return val, nil
			}
		}
		return nil, fmt.Errorf("type object '%s' has no attribute '%s'", o.Name, name)
	case *PyDict:
		if name == "get" {
			return &PyBuiltinFunc{Name: "dict.get", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) < 1 {
					return nil, fmt.Errorf("get() requires at least 1 argument")
				}
				key := args[0]
				def := Value(None)
				if len(args) > 1 {
					def = args[1]
				}
				if val, ok := o.Items[key]; ok {
					return val, nil
				}
				return def, nil
			}}, nil
		}
		if name == "keys" {
			return &PyBuiltinFunc{Name: "dict.keys", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				var keys []Value
				for k := range o.Items {
					keys = append(keys, k)
				}
				return &PyList{Items: keys}, nil
			}}, nil
		}
		if name == "values" {
			return &PyBuiltinFunc{Name: "dict.values", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				var vals []Value
				for _, v := range o.Items {
					vals = append(vals, v)
				}
				return &PyList{Items: vals}, nil
			}}, nil
		}
		if name == "items" {
			return &PyBuiltinFunc{Name: "dict.items", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				var items []Value
				for k, v := range o.Items {
					items = append(items, &PyTuple{Items: []Value{k, v}})
				}
				return &PyList{Items: items}, nil
			}}, nil
		}
	case *PyFrozenSet:
		fs := o
		if name == "copy" {
			return &PyBuiltinFunc{Name: "frozenset.copy", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				newFS := &PyFrozenSet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
				for k := range fs.Items {
					newFS.FrozenSetAdd(k, vm)
				}
				return newFS, nil
			}}, nil
		}
		if name == "union" {
			return &PyBuiltinFunc{Name: "frozenset.union", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				result := &PyFrozenSet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
				for k := range fs.Items {
					result.FrozenSetAdd(k, vm)
				}
				for _, arg := range args {
					items, err := vm.toList(arg)
					if err != nil {
						return nil, err
					}
					for _, item := range items {
						if !isHashable(item) {
							return nil, fmt.Errorf("TypeError: unhashable type: '%s'", vm.typeName(item))
						}
						result.FrozenSetAdd(item, vm)
					}
				}
				return result, nil
			}}, nil
		}
		if name == "intersection" {
			return &PyBuiltinFunc{Name: "frozenset.intersection", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				result := &PyFrozenSet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
				if len(args) == 0 {
					for k := range fs.Items {
						result.FrozenSetAdd(k, vm)
					}
					return result, nil
				}
				// Start with items from this frozenset that are in all other sets
				for k := range fs.Items {
					inAll := true
					for _, arg := range args {
						items, err := vm.toList(arg)
						if err != nil {
							return nil, err
						}
						found := false
						for _, item := range items {
							if vm.equal(k, item) {
								found = true
								break
							}
						}
						if !found {
							inAll = false
							break
						}
					}
					if inAll {
						result.FrozenSetAdd(k, vm)
					}
				}
				return result, nil
			}}, nil
		}
		if name == "difference" {
			return &PyBuiltinFunc{Name: "frozenset.difference", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				result := &PyFrozenSet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
				for k := range fs.Items {
					result.FrozenSetAdd(k, vm)
				}
				for _, arg := range args {
					items, err := vm.toList(arg)
					if err != nil {
						return nil, err
					}
					for _, item := range items {
						// Remove if present
						for rk := range result.Items {
							if vm.equal(rk, item) {
								delete(result.Items, rk)
								// Also remove from buckets
								if result.buckets != nil {
									h := hashValue(rk)
									entries := result.buckets[h]
									for i, e := range entries {
										if vm.equal(e.value, rk) {
											result.buckets[h] = append(entries[:i], entries[i+1:]...)
											result.size--
											break
										}
									}
								}
								break
							}
						}
					}
				}
				return result, nil
			}}, nil
		}
		if name == "symmetric_difference" {
			return &PyBuiltinFunc{Name: "frozenset.symmetric_difference", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("symmetric_difference() takes exactly 1 argument")
				}
				other, err := vm.toList(args[0])
				if err != nil {
					return nil, err
				}
				result := &PyFrozenSet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
				// Add items from fs that are not in other
				for k := range fs.Items {
					found := false
					for _, item := range other {
						if vm.equal(k, item) {
							found = true
							break
						}
					}
					if !found {
						result.FrozenSetAdd(k, vm)
					}
				}
				// Add items from other that are not in fs
				for _, item := range other {
					if !isHashable(item) {
						return nil, fmt.Errorf("TypeError: unhashable type: '%s'", vm.typeName(item))
					}
					if !fs.FrozenSetContains(item, vm) {
						result.FrozenSetAdd(item, vm)
					}
				}
				return result, nil
			}}, nil
		}
		if name == "issubset" {
			return &PyBuiltinFunc{Name: "frozenset.issubset", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("issubset() takes exactly 1 argument")
				}
				other, err := vm.toList(args[0])
				if err != nil {
					return nil, err
				}
				for k := range fs.Items {
					found := false
					for _, item := range other {
						if vm.equal(k, item) {
							found = true
							break
						}
					}
					if !found {
						return False, nil
					}
				}
				return True, nil
			}}, nil
		}
		if name == "issuperset" {
			return &PyBuiltinFunc{Name: "frozenset.issuperset", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("issuperset() takes exactly 1 argument")
				}
				other, err := vm.toList(args[0])
				if err != nil {
					return nil, err
				}
				for _, item := range other {
					if !fs.FrozenSetContains(item, vm) {
						return False, nil
					}
				}
				return True, nil
			}}, nil
		}
		if name == "isdisjoint" {
			return &PyBuiltinFunc{Name: "frozenset.isdisjoint", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("isdisjoint() takes exactly 1 argument")
				}
				other, err := vm.toList(args[0])
				if err != nil {
					return nil, err
				}
				for _, item := range other {
					if fs.FrozenSetContains(item, vm) {
						return False, nil
					}
				}
				return True, nil
			}}, nil
		}
	case *PyList:
		if name == "append" {
			return &PyBuiltinFunc{Name: "list.append", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("append() takes exactly 1 argument")
				}
				o.Items = append(o.Items, args[0])
				return None, nil
			}}, nil
		}
		if name == "pop" {
			return &PyBuiltinFunc{Name: "list.pop", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(o.Items) == 0 {
					return nil, fmt.Errorf("pop from empty list")
				}
				idx := len(o.Items) - 1
				if len(args) > 0 {
					idx = int(vm.toInt(args[0]))
				}
				val := o.Items[idx]
				o.Items = append(o.Items[:idx], o.Items[idx+1:]...)
				return val, nil
			}}, nil
		}
		if name == "extend" {
			return &PyBuiltinFunc{Name: "list.extend", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("extend() takes exactly 1 argument")
				}
				items, err := vm.toList(args[0])
				if err != nil {
					return nil, err
				}
				o.Items = append(o.Items, items...)
				return None, nil
			}}, nil
		}
	case *PyString:
		if name == "upper" {
			return &PyBuiltinFunc{Name: "str.upper", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				return &PyString{Value: strings.ToUpper(o.Value)}, nil
			}}, nil
		}
		if name == "lower" {
			return &PyBuiltinFunc{Name: "str.lower", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				return &PyString{Value: strings.ToLower(o.Value)}, nil
			}}, nil
		}
		if name == "split" {
			return &PyBuiltinFunc{Name: "str.split", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				var strParts []string
				if len(args) == 0 {
					// No separator: split on whitespace and remove empty strings
					strParts = strings.Fields(o.Value)
				} else {
					sep := vm.str(args[0])
					strParts = strings.Split(o.Value, sep)
				}
				parts := make([]Value, len(strParts))
				for i, s := range strParts {
					parts[i] = &PyString{Value: s}
				}
				return &PyList{Items: parts}, nil
			}}, nil
		}
		if name == "join" {
			return &PyBuiltinFunc{Name: "str.join", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("join() takes exactly 1 argument")
				}
				items, err := vm.toList(args[0])
				if err != nil {
					return nil, err
				}
				var parts []string
				for _, item := range items {
					parts = append(parts, vm.str(item))
				}
				result := ""
				for i, p := range parts {
					if i > 0 {
						result += o.Value
					}
					result += p
				}
				return &PyString{Value: result}, nil
			}}, nil
		}
		if name == "strip" {
			return &PyBuiltinFunc{Name: "str.strip", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				s := o.Value
				start := 0
				end := len(s)
				for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
					start++
				}
				for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
					end--
				}
				return &PyString{Value: s[start:end]}, nil
			}}, nil
		}
		if name == "replace" {
			return &PyBuiltinFunc{Name: "str.replace", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) < 2 {
					return nil, fmt.Errorf("replace() takes at least 2 arguments")
				}
				old := vm.str(args[0])
				new := vm.str(args[1])
				result := ""
				for i := 0; i < len(o.Value); {
					if i+len(old) <= len(o.Value) && o.Value[i:i+len(old)] == old {
						result += new
						i += len(old)
					} else {
						result += string(o.Value[i])
						i++
					}
				}
				return &PyString{Value: result}, nil
			}}, nil
		}
	case *PyFunction:
		switch name {
		case "__name__":
			return &PyString{Value: o.Name}, nil
		case "__doc__":
			return None, nil
		case "__wrapped__":
			// Check if we have __wrapped__ stored in closure
			if o.Closure != nil {
				for _, cell := range o.Closure {
					if cell != nil {
						return cell, nil
					}
				}
			}
			return nil, fmt.Errorf("'function' object has no attribute '__wrapped__'")
		}
		return nil, fmt.Errorf("'function' object has no attribute '%s'", name)
	}
	return nil, fmt.Errorf("'%s' object has no attribute '%s'", vm.typeName(obj), name)
}

func (vm *VM) setAttr(obj Value, name string, val Value) error {
	switch o := obj.(type) {
	case *PyInstance:
		// Check for property with setter in class MRO
		for _, cls := range o.Class.Mro {
			if clsVal, ok := cls.Dict[name]; ok {
				if prop, ok := clsVal.(*PyProperty); ok {
					if prop.Fset == nil {
						return fmt.Errorf("property '%s' has no setter", name)
					}
					// Call the setter with (instance, value)
					_, err := vm.call(prop.Fset, []Value{obj, val}, nil)
					return err
				}
				break // Found in class dict but not a property, fall through to instance assignment
			}
		}
		// Not a property, set on instance dict
		o.Dict[name] = val
		return nil
	case *PyClass:
		o.Dict[name] = val
		return nil
	}
	return fmt.Errorf("'%s' object attribute '%s' is read-only", vm.typeName(obj), name)
}
