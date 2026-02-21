package runtime

import (
	"fmt"
	"strings"
)

// getAttrGenerator handles attribute access on *PyGenerator values.
func (vm *VM) getAttrGenerator(gen *PyGenerator, name string) (Value, error) {
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
	return nil, fmt.Errorf("AttributeError: 'generator' object has no attribute '%s'", name)
}

// getAttrCoroutine handles attribute access on *PyCoroutine values.
func (vm *VM) getAttrCoroutine(coro *PyCoroutine, name string) (Value, error) {
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
	return nil, fmt.Errorf("AttributeError: 'coroutine' object has no attribute '%s'", name)
}

// getAttrException handles attribute access on *PyException values.
func (vm *VM) getAttrException(o *PyException, name string) (Value, error) {
	switch name {
	case "args":
		if o.Args != nil {
			return o.Args, nil
		}
		return &PyTuple{Items: []Value{}}, nil
	case "__class__":
		if o.ExcType != nil {
			return o.ExcType, nil
		}
		return None, nil
	case "__name__":
		return &PyString{Value: o.Type()}, nil
	case "__cause__":
		if o.Cause != nil {
			return o.Cause, nil
		}
		return None, nil
	case "__context__":
		if o.Context != nil {
			return o.Context, nil
		}
		return None, nil
	case "__suppress_context__":
		if o.SuppressContext {
			return True, nil
		}
		return False, nil
	case "__traceback__":
		return None, nil
	case "__notes__":
		if o.Notes != nil {
			return o.Notes, nil
		}
		return None, nil
	case "add_note":
		exc := o
		return &PyBuiltinFunc{
			Name: "add_note",
			Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("add_note() takes exactly one argument (%d given)", len(args))
				}
				note, ok := args[0].(*PyString)
				if !ok {
					return nil, fmt.Errorf("TypeError: note must be a str, not '%s'", vm.typeName(args[0]))
				}
				if exc.Notes == nil {
					exc.Notes = &PyList{Items: []Value{}}
				}
				exc.Notes.Items = append(exc.Notes.Items, note)
				return None, nil
			},
		}, nil
	}
	// Delegate to Instance for ExceptionGroup attributes (message, exceptions, subgroup, etc.)
	if o.Instance != nil {
		return vm.getAttr(o.Instance, name)
	}
	return nil, fmt.Errorf("AttributeError: '%s' object has no attribute '%s'", o.Type(), name)
}

// getAttrProperty handles attribute access on *PyProperty values.
func (vm *VM) getAttrProperty(prop *PyProperty, name string) (Value, error) {
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
	return nil, fmt.Errorf("AttributeError: 'property' object has no attribute '%s'", name)
}

// getAttrSuper handles attribute access on *PySuper values.
func (vm *VM) getAttrSuper(o *PySuper, name string) (Value, error) {
	// super() proxy - look up attribute in MRO starting after ThisClass
	if o.Instance == nil {
		return nil, fmt.Errorf("AttributeError: 'super' object has no attribute '%s'", name)
	}

	// Get the MRO to search
	var mro []*PyClass
	var instance Value = o.Instance
	if inst, ok := o.Instance.(*PyInstance); ok {
		mro = inst.Class.Mro
	} else if cls, ok := o.Instance.(*PyClass); ok {
		// Check if ThisClass is in the metaclass MRO
		useMetaMro := false
		if cls.Metaclass != nil {
			for _, mc := range cls.Metaclass.Mro {
				if mc == o.ThisClass {
					useMetaMro = true
					break
				}
			}
		}
		if useMetaMro {
			mro = cls.Metaclass.Mro
		} else {
			mro = cls.Mro
		}
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
				if bf, ok := cm.Func.(*PyBuiltinFunc); ok {
					boundInst := instance
					return &PyBuiltinFunc{
						Name: bf.Name,
						Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
							allArgs := append([]Value{boundInst}, args...)
							return bf.Fn(allArgs, kwargs)
						},
					}, nil
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

			// Bind builtin function to instance (e.g. type.__call__)
			if bf, ok := val.(*PyBuiltinFunc); ok {
				boundInst := instance
				return &PyBuiltinFunc{
					Name: bf.Name,
					Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
						allArgs := append([]Value{boundInst}, args...)
						return bf.Fn(allArgs, kwargs)
					},
				}, nil
			}

			return val, nil
		}
	}
	return nil, fmt.Errorf("AttributeError: 'super' object has no attribute '%s'", name)
}

// getAttrInstance handles attribute access on *PyInstance values.
func (vm *VM) getAttrInstance(o *PyInstance, name string) (Value, error) {
	// Check for custom __getattribute__ (not inherited from object)
	for _, cls := range o.Class.Mro {
		if cls.Name == "object" {
			break // object.__getattribute__ is the default, skip
		}
		if gaFn, ok := cls.Dict["__getattribute__"]; ok {
			result, err := vm.callMethodValue(gaFn, o, &PyString{Value: name})
			if err != nil {
				// If __getattribute__ raises AttributeError, fall through to __getattr__
				if isAttributeError(err) {
					if gaResult, found, gaErr := vm.callDunder(o, "__getattr__", &PyString{Value: name}); found {
						if gaErr != nil {
							return nil, gaErr
						}
						return gaResult, nil
					}
				}
				return nil, err
			}
			return result, nil
		}
	}

	// Use default attribute lookup, then fall back to __getattr__
	result, err := vm.defaultGetAttribute(o, name)
	if err != nil {
		// Default lookup failed: try __getattr__ as last resort
		if isAttributeError(err) {
			if gaResult, found, gaErr := vm.callDunder(o, "__getattr__", &PyString{Value: name}); found {
				if gaErr != nil {
					return nil, gaErr
				}
				return gaResult, nil
			}
		}
		return nil, err
	}
	return result, nil
}

// getAttrClass handles attribute access on *PyClass values.
func (vm *VM) getAttrClass(o *PyClass, name string) (Value, error) {
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

			// Check for custom descriptor with __get__ on class access
			// Invokes descriptor.__get__(None, owner_class)
			if inst, ok := val.(*PyInstance); ok {
				if getResult, found, err := vm.callDunder(inst, "__get__", None, o); found {
					if err != nil {
						return nil, err
					}
					return getResult, nil
				}
			}

			return val, nil
		}
	}
	return nil, fmt.Errorf("type object '%s' has no attribute '%s'", o.Name, name)
}

// getAttrFunction handles attribute access on *PyFunction values.
func (vm *VM) getAttrFunction(o *PyFunction, name string) (Value, error) {
	switch name {
	case "__name__":
		return &PyString{Value: o.Name}, nil
	case "__doc__":
		return None, nil
	case "__isabstractmethod__":
		return &PyBool{Value: o.IsAbstract}, nil
	case "__wrapped__":
		// Check if we have __wrapped__ stored in closure
		if o.Closure != nil {
			for _, cell := range o.Closure {
				if cell != nil {
					return cell, nil
				}
			}
		}
		return nil, fmt.Errorf("AttributeError: 'function' object has no attribute '__wrapped__'")
	}
	// Check custom attributes
	if o.Dict != nil {
		if val, ok := o.Dict[name]; ok {
			return val, nil
		}
	}
	return nil, fmt.Errorf("AttributeError: 'function' object has no attribute '%s'", name)
}

// getAttrBuiltinFunc handles attribute access on *PyBuiltinFunc values.
func (vm *VM) getAttrBuiltinFunc(o *PyBuiltinFunc, name string) (Value, error) {
	// Handle class methods on builtin types (e.g., dict.fromkeys)
	if o.Name == "dict" && name == "fromkeys" {
		return &PyBuiltinFunc{Name: "dict.fromkeys", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("fromkeys() requires at least 1 argument")
			}
			keys, err := vm.toList(args[0])
			if err != nil {
				return nil, err
			}
			def := Value(None)
			if len(args) > 1 {
				def = args[1]
			}
			result := &PyDict{Items: make(map[Value]Value), buckets: make(map[uint64][]dictEntry)}
			for _, k := range keys {
				result.DictSet(k, def, vm)
			}
			return result, nil
		}}, nil
	}
	if o.Name == "bytes" && name == "maketrans" {
		return &PyBuiltinFunc{Name: "bytes.maketrans", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			return bytesMaketransImpl(args)
		}}, nil
	}
	if o.Name == "str" && name == "maketrans" {
		return &PyBuiltinFunc{Name: "str.maketrans", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			return strMaketransImpl(vm, args)
		}}, nil
	}
	if o.Name == "float" && name == "fromhex" {
		return &PyBuiltinFunc{Name: "float.fromhex", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			return floatFromHexImpl(args)
		}}, nil
	}
	if o.Name == "int" && name == "from_bytes" {
		return &PyBuiltinFunc{Name: "int.from_bytes", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			return intFromBytesImpl(vm, args, kwargs)
		}}, nil
	}
	// Handle dunder attribute queries on builtin type constructors (e.g., hasattr(list, "__iter__"))
	if builtinHasDunder(o.Name, name) {
		return &PyBuiltinFunc{Name: o.Name + "." + name}, nil
	}
	return nil, fmt.Errorf("AttributeError: '%s' object has no attribute '%s'", vm.typeName(o), name)
}

// getAttrRange handles attribute access on *PyRange values.
func (vm *VM) getAttrRange(r *PyRange, name string) (Value, error) {
	switch name {
	case "start":
		return MakeInt(r.Start), nil
	case "stop":
		return MakeInt(r.Stop), nil
	case "step":
		return MakeInt(r.Step), nil
	case "count":
		return &PyBuiltinFunc{Name: "range.count", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("TypeError: count() takes exactly one argument (%d given)", len(args))
			}
			val, ok := args[0].(*PyInt)
			if !ok {
				return MakeInt(0), nil
			}
			v := val.Value
			if r.Contains(v) {
				return MakeInt(1), nil
			}
			return MakeInt(0), nil
		}}, nil
	case "index":
		return &PyBuiltinFunc{Name: "range.index", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("TypeError: index() takes exactly one argument (%d given)", len(args))
			}
			val, ok := args[0].(*PyInt)
			if !ok {
				return nil, fmt.Errorf("ValueError: %s is not in range", vm.str(args[0]))
			}
			v := val.Value
			if !r.Contains(v) {
				return nil, fmt.Errorf("ValueError: %d is not in range", v)
			}
			if r.Step > 0 {
				return MakeInt((v - r.Start) / r.Step), nil
			}
			return MakeInt((r.Start - v) / (-r.Step)), nil
		}}, nil
	}
	return nil, fmt.Errorf("AttributeError: 'range' object has no attribute '%s'", name)
}

// getAttrBytes handles attribute access on *PyBytes values.
func (vm *VM) getAttrBytes(b *PyBytes, name string) (Value, error) {
	switch name {
	case "find":
		return &PyBuiltinFunc{Name: "bytes.find", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("find() takes at least 1 argument")
			}
			var sub []byte
			switch s := args[0].(type) {
			case *PyBytes:
				sub = s.Value
			case *PyInt:
				if s.Value < 0 || s.Value > 255 {
					return nil, fmt.Errorf("ValueError: byte must be in range(0, 256)")
				}
				sub = []byte{byte(s.Value)}
			default:
				return nil, fmt.Errorf("TypeError: argument should be integer or bytes-like object, not '%s'", vm.typeName(args[0]))
			}
			data := b.Value
			start := 0
			if len(args) > 1 {
				start = int(vm.toInt(args[1]))
				if start < 0 {
					start += len(data)
					if start < 0 {
						start = 0
					}
				}
			}
			end := len(data)
			if len(args) > 2 {
				end = int(vm.toInt(args[2]))
				if end < 0 {
					end += len(data)
				}
				if end > len(data) {
					end = len(data)
				}
			}
			if start > len(data) || start >= end {
				return MakeInt(-1), nil
			}
			searchData := data[start:end]
			idx := bytesIndex(searchData, sub)
			if idx < 0 {
				return MakeInt(-1), nil
			}
			return MakeInt(int64(start + idx)), nil
		}}, nil
	case "index":
		return &PyBuiltinFunc{Name: "bytes.index", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("index() takes at least 1 argument")
			}
			var sub []byte
			switch s := args[0].(type) {
			case *PyBytes:
				sub = s.Value
			case *PyInt:
				if s.Value < 0 || s.Value > 255 {
					return nil, fmt.Errorf("ValueError: byte must be in range(0, 256)")
				}
				sub = []byte{byte(s.Value)}
			default:
				return nil, fmt.Errorf("TypeError: argument should be integer or bytes-like object, not '%s'", vm.typeName(args[0]))
			}
			data := b.Value
			start := 0
			if len(args) > 1 {
				start = int(vm.toInt(args[1]))
				if start < 0 {
					start += len(data)
					if start < 0 {
						start = 0
					}
				}
			}
			end := len(data)
			if len(args) > 2 {
				end = int(vm.toInt(args[2]))
				if end < 0 {
					end += len(data)
				}
				if end > len(data) {
					end = len(data)
				}
			}
			if start > len(data) || start >= end {
				return nil, fmt.Errorf("ValueError: subsection not found")
			}
			searchData := data[start:end]
			idx := bytesIndex(searchData, sub)
			if idx < 0 {
				return nil, fmt.Errorf("ValueError: subsection not found")
			}
			return MakeInt(int64(start + idx)), nil
		}}, nil
	case "count":
		return &PyBuiltinFunc{Name: "bytes.count", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("count() takes at least 1 argument")
			}
			var sub []byte
			switch s := args[0].(type) {
			case *PyBytes:
				sub = s.Value
			case *PyInt:
				if s.Value < 0 || s.Value > 255 {
					return nil, fmt.Errorf("ValueError: byte must be in range(0, 256)")
				}
				sub = []byte{byte(s.Value)}
			default:
				return nil, fmt.Errorf("TypeError: argument should be integer or bytes-like object, not '%s'", vm.typeName(args[0]))
			}
			data := b.Value
			start := 0
			end := len(data)
			if len(args) > 1 {
				start = int(vm.toInt(args[1]))
				if start < 0 {
					start += len(data)
					if start < 0 {
						start = 0
					}
				}
			}
			if len(args) > 2 {
				end = int(vm.toInt(args[2]))
				if end < 0 {
					end += len(data)
				}
				if end > len(data) {
					end = len(data)
				}
			}
			if start > len(data) || start >= end {
				return MakeInt(0), nil
			}
			count := bytesCount(data[start:end], sub)
			return MakeInt(int64(count)), nil
		}}, nil
	case "replace":
		return &PyBuiltinFunc{Name: "bytes.replace", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("replace() takes at least 2 arguments")
			}
			oldB, ok := args[0].(*PyBytes)
			if !ok {
				return nil, fmt.Errorf("TypeError: a bytes-like object is required, not '%s'", vm.typeName(args[0]))
			}
			newB, ok := args[1].(*PyBytes)
			if !ok {
				return nil, fmt.Errorf("TypeError: a bytes-like object is required, not '%s'", vm.typeName(args[1]))
			}
			count := -1
			if len(args) > 2 {
				count = int(vm.toInt(args[2]))
			}
			result := bytesReplace(b.Value, oldB.Value, newB.Value, count)
			return &PyBytes{Value: result}, nil
		}}, nil
	case "split":
		return &PyBuiltinFunc{Name: "bytes.split", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) == 0 || args[0] == None {
				// Split on whitespace
				parts := bytesSplitWhitespace(b.Value)
				items := make([]Value, len(parts))
				for i, p := range parts {
					items[i] = &PyBytes{Value: p}
				}
				return &PyList{Items: items}, nil
			}
			sep, ok := args[0].(*PyBytes)
			if !ok {
				return nil, fmt.Errorf("TypeError: a bytes-like object is required, not '%s'", vm.typeName(args[0]))
			}
			if len(sep.Value) == 0 {
				return nil, fmt.Errorf("ValueError: empty separator")
			}
			maxSplit := -1
			if len(args) > 1 {
				maxSplit = int(vm.toInt(args[1]))
			}
			parts := bytesSplit(b.Value, sep.Value, maxSplit)
			items := make([]Value, len(parts))
			for i, p := range parts {
				items[i] = &PyBytes{Value: p}
			}
			return &PyList{Items: items}, nil
		}}, nil
	case "join":
		return &PyBuiltinFunc{Name: "bytes.join", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("join() takes exactly 1 argument")
			}
			items, err := vm.toList(args[0])
			if err != nil {
				return nil, err
			}
			var parts [][]byte
			for _, item := range items {
				bItem, ok := item.(*PyBytes)
				if !ok {
					return nil, fmt.Errorf("TypeError: sequence item: expected a bytes-like object, %s found", vm.typeName(item))
				}
				parts = append(parts, bItem.Value)
			}
			result := bytesJoin(b.Value, parts)
			return &PyBytes{Value: result}, nil
		}}, nil
	case "strip":
		return &PyBuiltinFunc{Name: "bytes.strip", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) > 0 && args[0] != None {
				chars, ok := args[0].(*PyBytes)
				if !ok {
					return nil, fmt.Errorf("TypeError: a bytes-like object is required, not '%s'", vm.typeName(args[0]))
				}
				return &PyBytes{Value: bytesStripChars(b.Value, chars.Value)}, nil
			}
			return &PyBytes{Value: bytesStripWhitespace(b.Value)}, nil
		}}, nil
	case "lstrip":
		return &PyBuiltinFunc{Name: "bytes.lstrip", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) > 0 && args[0] != None {
				chars, ok := args[0].(*PyBytes)
				if !ok {
					return nil, fmt.Errorf("TypeError: a bytes-like object is required, not '%s'", vm.typeName(args[0]))
				}
				return &PyBytes{Value: bytesLstripChars(b.Value, chars.Value)}, nil
			}
			return &PyBytes{Value: bytesLstripWhitespace(b.Value)}, nil
		}}, nil
	case "rstrip":
		return &PyBuiltinFunc{Name: "bytes.rstrip", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) > 0 && args[0] != None {
				chars, ok := args[0].(*PyBytes)
				if !ok {
					return nil, fmt.Errorf("TypeError: a bytes-like object is required, not '%s'", vm.typeName(args[0]))
				}
				return &PyBytes{Value: bytesRstripChars(b.Value, chars.Value)}, nil
			}
			return &PyBytes{Value: bytesRstripWhitespace(b.Value)}, nil
		}}, nil
	case "upper":
		return &PyBuiltinFunc{Name: "bytes.upper", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			result := make([]byte, len(b.Value))
			for i, c := range b.Value {
				if c >= 'a' && c <= 'z' {
					result[i] = c - 32
				} else {
					result[i] = c
				}
			}
			return &PyBytes{Value: result}, nil
		}}, nil
	case "lower":
		return &PyBuiltinFunc{Name: "bytes.lower", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			result := make([]byte, len(b.Value))
			for i, c := range b.Value {
				if c >= 'A' && c <= 'Z' {
					result[i] = c + 32
				} else {
					result[i] = c
				}
			}
			return &PyBytes{Value: result}, nil
		}}, nil
	case "startswith":
		return &PyBuiltinFunc{Name: "bytes.startswith", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("startswith() takes at least 1 argument")
			}
			// Handle tuple of prefixes
			if t, ok := args[0].(*PyTuple); ok {
				for _, item := range t.Items {
					prefix, ok := item.(*PyBytes)
					if !ok {
						return nil, fmt.Errorf("TypeError: a bytes-like object is required, not '%s'", vm.typeName(item))
					}
					if bytesStartsWith(b.Value, prefix.Value) {
						return True, nil
					}
				}
				return False, nil
			}
			prefix, ok := args[0].(*PyBytes)
			if !ok {
				return nil, fmt.Errorf("TypeError: a bytes-like object is required, not '%s'", vm.typeName(args[0]))
			}
			if bytesStartsWith(b.Value, prefix.Value) {
				return True, nil
			}
			return False, nil
		}}, nil
	case "endswith":
		return &PyBuiltinFunc{Name: "bytes.endswith", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("endswith() takes at least 1 argument")
			}
			if t, ok := args[0].(*PyTuple); ok {
				for _, item := range t.Items {
					suffix, ok := item.(*PyBytes)
					if !ok {
						return nil, fmt.Errorf("TypeError: a bytes-like object is required, not '%s'", vm.typeName(item))
					}
					if bytesEndsWith(b.Value, suffix.Value) {
						return True, nil
					}
				}
				return False, nil
			}
			suffix, ok := args[0].(*PyBytes)
			if !ok {
				return nil, fmt.Errorf("TypeError: a bytes-like object is required, not '%s'", vm.typeName(args[0]))
			}
			if bytesEndsWith(b.Value, suffix.Value) {
				return True, nil
			}
			return False, nil
		}}, nil
	case "hex":
		return &PyBuiltinFunc{Name: "bytes.hex", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			sep := ""
			if len(args) > 0 {
				if s, ok := args[0].(*PyString); ok {
					sep = s.Value
				}
			}
			var result strings.Builder
			for i, c := range b.Value {
				if i > 0 && sep != "" {
					result.WriteString(sep)
				}
				fmt.Fprintf(&result, "%02x", c)
			}
			return &PyString{Value: result.String()}, nil
		}}, nil
	case "decode":
		return &PyBuiltinFunc{Name: "bytes.decode", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			// Default encoding is utf-8
			// We just convert bytes to string directly
			return &PyString{Value: string(b.Value)}, nil
		}}, nil
	case "isalpha":
		return &PyBuiltinFunc{Name: "bytes.isalpha", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(b.Value) == 0 {
				return False, nil
			}
			for _, c := range b.Value {
				if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')) {
					return False, nil
				}
			}
			return True, nil
		}}, nil
	case "isdigit":
		return &PyBuiltinFunc{Name: "bytes.isdigit", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(b.Value) == 0 {
				return False, nil
			}
			for _, c := range b.Value {
				if c < '0' || c > '9' {
					return False, nil
				}
			}
			return True, nil
		}}, nil
	case "isalnum":
		return &PyBuiltinFunc{Name: "bytes.isalnum", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(b.Value) == 0 {
				return False, nil
			}
			for _, c := range b.Value {
				if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
					return False, nil
				}
			}
			return True, nil
		}}, nil
	case "isspace":
		return &PyBuiltinFunc{Name: "bytes.isspace", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(b.Value) == 0 {
				return False, nil
			}
			for _, c := range b.Value {
				if c != ' ' && c != '\t' && c != '\n' && c != '\r' && c != '\f' && c != '\v' {
					return False, nil
				}
			}
			return True, nil
		}}, nil
	case "isupper":
		return &PyBuiltinFunc{Name: "bytes.isupper", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			hasUpper := false
			for _, c := range b.Value {
				if c >= 'a' && c <= 'z' {
					return False, nil
				}
				if c >= 'A' && c <= 'Z' {
					hasUpper = true
				}
			}
			if hasUpper {
				return True, nil
			}
			return False, nil
		}}, nil
	case "islower":
		return &PyBuiltinFunc{Name: "bytes.islower", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			hasLower := false
			for _, c := range b.Value {
				if c >= 'A' && c <= 'Z' {
					return False, nil
				}
				if c >= 'a' && c <= 'z' {
					hasLower = true
				}
			}
			if hasLower {
				return True, nil
			}
			return False, nil
		}}, nil
	case "title":
		return &PyBuiltinFunc{Name: "bytes.title", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			result := make([]byte, len(b.Value))
			newWord := true
			for i, c := range b.Value {
				if c >= 'a' && c <= 'z' {
					if newWord {
						result[i] = c - 32
					} else {
						result[i] = c
					}
					newWord = false
				} else if c >= 'A' && c <= 'Z' {
					if !newWord {
						result[i] = c + 32
					} else {
						result[i] = c
					}
					newWord = false
				} else {
					result[i] = c
					newWord = true
				}
			}
			return &PyBytes{Value: result}, nil
		}}, nil
	case "capitalize":
		return &PyBuiltinFunc{Name: "bytes.capitalize", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(b.Value) == 0 {
				return &PyBytes{Value: []byte{}}, nil
			}
			result := make([]byte, len(b.Value))
			// Capitalize first byte, lowercase rest
			if b.Value[0] >= 'a' && b.Value[0] <= 'z' {
				result[0] = b.Value[0] - 32
			} else {
				result[0] = b.Value[0]
			}
			for i := 1; i < len(b.Value); i++ {
				if b.Value[i] >= 'A' && b.Value[i] <= 'Z' {
					result[i] = b.Value[i] + 32
				} else {
					result[i] = b.Value[i]
				}
			}
			return &PyBytes{Value: result}, nil
		}}, nil
	case "swapcase":
		return &PyBuiltinFunc{Name: "bytes.swapcase", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			result := make([]byte, len(b.Value))
			for i, c := range b.Value {
				if c >= 'a' && c <= 'z' {
					result[i] = c - 32
				} else if c >= 'A' && c <= 'Z' {
					result[i] = c + 32
				} else {
					result[i] = c
				}
			}
			return &PyBytes{Value: result}, nil
		}}, nil
	case "center":
		return &PyBuiltinFunc{Name: "bytes.center", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("center() takes at least 1 argument")
			}
			width := int(vm.toInt(args[0]))
			fillByte := byte(' ')
			if len(args) > 1 {
				if fb, ok := args[1].(*PyBytes); ok && len(fb.Value) == 1 {
					fillByte = fb.Value[0]
				}
			}
			if len(b.Value) >= width {
				cp := make([]byte, len(b.Value))
				copy(cp, b.Value)
				return &PyBytes{Value: cp}, nil
			}
			total := width - len(b.Value)
			left := total / 2
			right := total - left
			result := make([]byte, width)
			for i := 0; i < left; i++ {
				result[i] = fillByte
			}
			copy(result[left:], b.Value)
			for i := 0; i < right; i++ {
				result[left+len(b.Value)+i] = fillByte
			}
			return &PyBytes{Value: result}, nil
		}}, nil
	case "ljust":
		return &PyBuiltinFunc{Name: "bytes.ljust", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("ljust() takes at least 1 argument")
			}
			width := int(vm.toInt(args[0]))
			fillByte := byte(' ')
			if len(args) > 1 {
				if fb, ok := args[1].(*PyBytes); ok && len(fb.Value) == 1 {
					fillByte = fb.Value[0]
				}
			}
			if len(b.Value) >= width {
				cp := make([]byte, len(b.Value))
				copy(cp, b.Value)
				return &PyBytes{Value: cp}, nil
			}
			result := make([]byte, width)
			copy(result, b.Value)
			for i := len(b.Value); i < width; i++ {
				result[i] = fillByte
			}
			return &PyBytes{Value: result}, nil
		}}, nil
	case "rjust":
		return &PyBuiltinFunc{Name: "bytes.rjust", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("rjust() takes at least 1 argument")
			}
			width := int(vm.toInt(args[0]))
			fillByte := byte(' ')
			if len(args) > 1 {
				if fb, ok := args[1].(*PyBytes); ok && len(fb.Value) == 1 {
					fillByte = fb.Value[0]
				}
			}
			if len(b.Value) >= width {
				cp := make([]byte, len(b.Value))
				copy(cp, b.Value)
				return &PyBytes{Value: cp}, nil
			}
			result := make([]byte, width)
			padding := width - len(b.Value)
			for i := 0; i < padding; i++ {
				result[i] = fillByte
			}
			copy(result[padding:], b.Value)
			return &PyBytes{Value: result}, nil
		}}, nil
	case "rfind":
		return &PyBuiltinFunc{Name: "bytes.rfind", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("rfind() takes at least 1 argument")
			}
			var sub []byte
			switch s := args[0].(type) {
			case *PyBytes:
				sub = s.Value
			case *PyInt:
				if s.Value < 0 || s.Value > 255 {
					return nil, fmt.Errorf("ValueError: byte must be in range(0, 256)")
				}
				sub = []byte{byte(s.Value)}
			default:
				return nil, fmt.Errorf("TypeError: argument should be integer or bytes-like object, not '%s'", vm.typeName(args[0]))
			}
			idx := bytesRindex(b.Value, sub)
			return MakeInt(int64(idx)), nil
		}}, nil
	case "rindex":
		return &PyBuiltinFunc{Name: "bytes.rindex", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("rindex() takes at least 1 argument")
			}
			var sub []byte
			switch s := args[0].(type) {
			case *PyBytes:
				sub = s.Value
			case *PyInt:
				if s.Value < 0 || s.Value > 255 {
					return nil, fmt.Errorf("ValueError: byte must be in range(0, 256)")
				}
				sub = []byte{byte(s.Value)}
			default:
				return nil, fmt.Errorf("TypeError: argument should be integer or bytes-like object, not '%s'", vm.typeName(args[0]))
			}
			idx := bytesRindex(b.Value, sub)
			if idx < 0 {
				return nil, fmt.Errorf("ValueError: subsection not found")
			}
			return MakeInt(int64(idx)), nil
		}}, nil
	case "partition":
		return &PyBuiltinFunc{Name: "bytes.partition", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("partition() takes exactly 1 argument")
			}
			sep, ok := args[0].(*PyBytes)
			if !ok {
				return nil, fmt.Errorf("TypeError: a bytes-like object is required, not '%s'", vm.typeName(args[0]))
			}
			idx := bytesIndex(b.Value, sep.Value)
			if idx < 0 {
				return &PyTuple{Items: []Value{
					&PyBytes{Value: append([]byte{}, b.Value...)},
					&PyBytes{Value: []byte{}},
					&PyBytes{Value: []byte{}},
				}}, nil
			}
			return &PyTuple{Items: []Value{
				&PyBytes{Value: append([]byte{}, b.Value[:idx]...)},
				&PyBytes{Value: append([]byte{}, sep.Value...)},
				&PyBytes{Value: append([]byte{}, b.Value[idx+len(sep.Value):]...)},
			}}, nil
		}}, nil
	case "rpartition":
		return &PyBuiltinFunc{Name: "bytes.rpartition", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("rpartition() takes exactly 1 argument")
			}
			sep, ok := args[0].(*PyBytes)
			if !ok {
				return nil, fmt.Errorf("TypeError: a bytes-like object is required, not '%s'", vm.typeName(args[0]))
			}
			idx := bytesRindex(b.Value, sep.Value)
			if idx < 0 {
				return &PyTuple{Items: []Value{
					&PyBytes{Value: []byte{}},
					&PyBytes{Value: []byte{}},
					&PyBytes{Value: append([]byte{}, b.Value...)},
				}}, nil
			}
			return &PyTuple{Items: []Value{
				&PyBytes{Value: append([]byte{}, b.Value[:idx]...)},
				&PyBytes{Value: append([]byte{}, sep.Value...)},
				&PyBytes{Value: append([]byte{}, b.Value[idx+len(sep.Value):]...)},
			}}, nil
		}}, nil
	case "expandtabs":
		return &PyBuiltinFunc{Name: "bytes.expandtabs", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			tabSize := 8
			if len(args) > 0 {
				tabSize = int(vm.toInt(args[0]))
			}
			var result []byte
			col := 0
			for _, c := range b.Value {
				if c == '\t' {
					spaces := tabSize - (col % tabSize)
					for j := 0; j < spaces; j++ {
						result = append(result, ' ')
					}
					col += spaces
				} else if c == '\n' || c == '\r' {
					result = append(result, c)
					col = 0
				} else {
					result = append(result, c)
					col++
				}
			}
			return &PyBytes{Value: result}, nil
		}}, nil
	case "zfill":
		return &PyBuiltinFunc{Name: "bytes.zfill", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("zfill() takes exactly 1 argument")
			}
			width := int(vm.toInt(args[0]))
			if len(b.Value) >= width {
				cp := make([]byte, len(b.Value))
				copy(cp, b.Value)
				return &PyBytes{Value: cp}, nil
			}
			result := make([]byte, width)
			padding := width - len(b.Value)
			startData := 0
			startResult := 0
			if len(b.Value) > 0 && (b.Value[0] == '+' || b.Value[0] == '-') {
				result[0] = b.Value[0]
				startData = 1
				startResult = 1
			}
			for i := startResult; i < startResult+padding; i++ {
				result[i] = '0'
			}
			copy(result[startResult+padding:], b.Value[startData:])
			return &PyBytes{Value: result}, nil
		}}, nil
	case "splitlines":
		return &PyBuiltinFunc{Name: "bytes.splitlines", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			keepends := false
			if len(args) > 0 {
				keepends = vm.truthy(args[0])
			}
			if len(b.Value) == 0 {
				return &PyList{Items: []Value{}}, nil
			}
			var lines []Value
			start := 0
			data := b.Value
			for i := 0; i < len(data); i++ {
				if data[i] == '\n' || data[i] == '\r' {
					end := i
					if data[i] == '\r' && i+1 < len(data) && data[i+1] == '\n' {
						i++
					}
					if keepends {
						lines = append(lines, &PyBytes{Value: append([]byte{}, data[start:i+1]...)})
					} else {
						lines = append(lines, &PyBytes{Value: append([]byte{}, data[start:end]...)})
					}
					start = i + 1
				}
			}
			if start < len(data) {
				lines = append(lines, &PyBytes{Value: append([]byte{}, data[start:]...)})
			}
			return &PyList{Items: lines}, nil
		}}, nil
	case "rsplit":
		return &PyBuiltinFunc{Name: "bytes.rsplit", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) == 0 || args[0] == None {
				// Split on whitespace
				parts := bytesSplitWhitespace(b.Value)
				items := make([]Value, len(parts))
				for i, p := range parts {
					items[i] = &PyBytes{Value: p}
				}
				return &PyList{Items: items}, nil
			}
			sep, ok := args[0].(*PyBytes)
			if !ok {
				return nil, fmt.Errorf("TypeError: a bytes-like object is required, not '%s'", vm.typeName(args[0]))
			}
			if len(sep.Value) == 0 {
				return nil, fmt.Errorf("ValueError: empty separator")
			}
			maxSplit := -1
			if len(args) > 1 {
				maxSplit = int(vm.toInt(args[1]))
			}
			if maxSplit < 0 {
				// Same as regular split with no limit
				parts := bytesSplit(b.Value, sep.Value, -1)
				items := make([]Value, len(parts))
				for i, p := range parts {
					items[i] = &PyBytes{Value: p}
				}
				return &PyList{Items: items}, nil
			}
			// rsplit from right
			data := b.Value
			var resultParts [][]byte
			for maxSplit > 0 {
				idx := bytesRindex(data, sep.Value)
				if idx < 0 {
					break
				}
				resultParts = append([][]byte{append([]byte{}, data[idx+len(sep.Value):]...)}, resultParts...)
				data = data[:idx]
				maxSplit--
			}
			resultParts = append([][]byte{append([]byte{}, data...)}, resultParts...)
			items := make([]Value, len(resultParts))
			for i, p := range resultParts {
				items[i] = &PyBytes{Value: p}
			}
			return &PyList{Items: items}, nil
		}}, nil
	case "removeprefix":
		return &PyBuiltinFunc{Name: "bytes.removeprefix", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("TypeError: removeprefix() takes exactly one argument (%d given)", len(args))
			}
			prefix, ok := args[0].(*PyBytes)
			if !ok {
				return nil, fmt.Errorf("TypeError: removeprefix arg must be bytes, not '%s'", vm.typeName(args[0]))
			}
			if len(prefix.Value) <= len(b.Value) && bytesStartsWith(b.Value, prefix.Value) {
				return &PyBytes{Value: b.Value[len(prefix.Value):]}, nil
			}
			return b, nil
		}}, nil
	case "removesuffix":
		return &PyBuiltinFunc{Name: "bytes.removesuffix", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("TypeError: removesuffix() takes exactly one argument (%d given)", len(args))
			}
			suffix, ok := args[0].(*PyBytes)
			if !ok {
				return nil, fmt.Errorf("TypeError: removesuffix arg must be bytes, not '%s'", vm.typeName(args[0]))
			}
			if len(suffix.Value) > 0 && len(suffix.Value) <= len(b.Value) && bytesEndsWith(b.Value, suffix.Value) {
				return &PyBytes{Value: b.Value[:len(b.Value)-len(suffix.Value)]}, nil
			}
			return b, nil
		}}, nil
	case "isascii":
		return &PyBuiltinFunc{Name: "bytes.isascii", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			for _, c := range b.Value {
				if c > 127 {
					return False, nil
				}
			}
			return True, nil
		}}, nil
	case "istitle":
		return &PyBuiltinFunc{Name: "bytes.istitle", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(b.Value) == 0 {
				return False, nil
			}
			prevCased := false
			hasCased := false
			for _, c := range b.Value {
				isUpper := c >= 'A' && c <= 'Z'
				isLower := c >= 'a' && c <= 'z'
				if isUpper {
					if prevCased {
						return False, nil
					}
					prevCased = true
					hasCased = true
				} else if isLower {
					if !prevCased {
						return False, nil
					}
					prevCased = true
					hasCased = true
				} else {
					prevCased = false
				}
			}
			if !hasCased {
				return False, nil
			}
			return True, nil
		}}, nil
	case "maketrans":
		return &PyBuiltinFunc{Name: "bytes.maketrans", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			return bytesMaketransImpl(args)
		}}, nil
	case "translate":
		return &PyBuiltinFunc{Name: "bytes.translate", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("TypeError: translate() takes at least 1 argument")
			}
			var table []byte
			if args[0] != None {
				t, ok := args[0].(*PyBytes)
				if !ok {
					return nil, fmt.Errorf("TypeError: translate() argument 1 must be bytes or None")
				}
				if len(t.Value) != 256 {
					return nil, fmt.Errorf("ValueError: translation table must be 256 characters long")
				}
				table = t.Value
			}
			// Optional second argument: bytes to delete
			var deleteSet [256]bool
			if len(args) >= 2 {
				del, ok := args[1].(*PyBytes)
				if !ok {
					return nil, fmt.Errorf("TypeError: translate() argument 2 must be bytes")
				}
				for _, c := range del.Value {
					deleteSet[c] = true
				}
			}
			result := make([]byte, 0, len(b.Value))
			for _, c := range b.Value {
				if deleteSet[c] {
					continue
				}
				if table != nil {
					result = append(result, table[c])
				} else {
					result = append(result, c)
				}
			}
			return &PyBytes{Value: result}, nil
		}}, nil
	}
	return nil, fmt.Errorf("AttributeError: 'bytes' object has no attribute '%s'", name)
}

// getAttrNone handles attribute access on *PyNone values.
// PyNone has no specific attributes.
func (vm *VM) getAttrNone(o *PyNone, name string) (Value, error) {
	return nil, fmt.Errorf("AttributeError: 'NoneType' object has no attribute '%s'", name)
}

// bytesMaketransImpl implements bytes.maketrans(from, to).
func bytesMaketransImpl(args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("TypeError: maketrans() takes exactly 2 arguments (%d given)", len(args))
	}
	from, ok1 := args[0].(*PyBytes)
	to, ok2 := args[1].(*PyBytes)
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("TypeError: maketrans arguments must be bytes")
	}
	if len(from.Value) != len(to.Value) {
		return nil, fmt.Errorf("ValueError: maketrans arguments must have same length")
	}
	table := make([]byte, 256)
	for i := range table {
		table[i] = byte(i)
	}
	for i, f := range from.Value {
		table[f] = to.Value[i]
	}
	return &PyBytes{Value: table}, nil
}

// --- Bytes helper functions ---

// bytesIndex finds the first occurrence of sub in data, returns -1 if not found
func bytesIndex(data, sub []byte) int {
	if len(sub) == 0 {
		return 0
	}
	if len(sub) > len(data) {
		return -1
	}
	for i := 0; i <= len(data)-len(sub); i++ {
		match := true
		for j := 0; j < len(sub); j++ {
			if data[i+j] != sub[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

// bytesRindex finds the last occurrence of sub in data, returns -1 if not found
func bytesRindex(data, sub []byte) int {
	if len(sub) == 0 {
		return len(data)
	}
	if len(sub) > len(data) {
		return -1
	}
	for i := len(data) - len(sub); i >= 0; i-- {
		match := true
		for j := 0; j < len(sub); j++ {
			if data[i+j] != sub[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

// bytesCount counts non-overlapping occurrences of sub in data
func bytesCount(data, sub []byte) int {
	if len(sub) == 0 {
		return len(data) + 1
	}
	count := 0
	for i := 0; i <= len(data)-len(sub); {
		idx := bytesIndex(data[i:], sub)
		if idx < 0 {
			break
		}
		count++
		i += idx + len(sub)
	}
	return count
}

// bytesReplace replaces occurrences of old with new in data
func bytesReplace(data, old, new []byte, count int) []byte {
	if len(old) == 0 {
		// Special case: empty old pattern - insert new between every byte
		var result []byte
		n := count
		if n < 0 {
			n = len(data) + 1
		}
		for i := 0; i < len(data); i++ {
			if n > 0 {
				result = append(result, new...)
				n--
			}
			result = append(result, data[i])
		}
		if n > 0 {
			result = append(result, new...)
		}
		return result
	}
	var result []byte
	i := 0
	replaced := 0
	for i <= len(data)-len(old) {
		if count >= 0 && replaced >= count {
			break
		}
		idx := bytesIndex(data[i:], old)
		if idx < 0 {
			break
		}
		result = append(result, data[i:i+idx]...)
		result = append(result, new...)
		i += idx + len(old)
		replaced++
	}
	result = append(result, data[i:]...)
	return result
}

// bytesSplit splits data on sep with optional max splits
func bytesSplit(data, sep []byte, maxSplit int) [][]byte {
	var parts [][]byte
	i := 0
	splits := 0
	for i <= len(data)-len(sep) {
		if maxSplit >= 0 && splits >= maxSplit {
			break
		}
		idx := bytesIndex(data[i:], sep)
		if idx < 0 {
			break
		}
		parts = append(parts, append([]byte{}, data[i:i+idx]...))
		i += idx + len(sep)
		splits++
	}
	parts = append(parts, append([]byte{}, data[i:]...))
	return parts
}

// bytesSplitWhitespace splits data on whitespace
func bytesSplitWhitespace(data []byte) [][]byte {
	var parts [][]byte
	i := 0
	for i < len(data) {
		// Skip whitespace
		for i < len(data) && isWhitespace(data[i]) {
			i++
		}
		if i >= len(data) {
			break
		}
		// Find end of non-whitespace
		start := i
		for i < len(data) && !isWhitespace(data[i]) {
			i++
		}
		parts = append(parts, append([]byte{}, data[start:i]...))
	}
	return parts
}

func isWhitespace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r' || b == '\f' || b == '\v'
}

// bytesJoin joins byte slices with separator
func bytesJoin(sep []byte, parts [][]byte) []byte {
	if len(parts) == 0 {
		return []byte{}
	}
	// Calculate total length
	totalLen := 0
	for i, p := range parts {
		totalLen += len(p)
		if i > 0 {
			totalLen += len(sep)
		}
	}
	result := make([]byte, 0, totalLen)
	for i, p := range parts {
		if i > 0 {
			result = append(result, sep...)
		}
		result = append(result, p...)
	}
	return result
}

// bytesStripWhitespace strips leading and trailing whitespace
func bytesStripWhitespace(data []byte) []byte {
	start := 0
	for start < len(data) && isWhitespace(data[start]) {
		start++
	}
	end := len(data)
	for end > start && isWhitespace(data[end-1]) {
		end--
	}
	return append([]byte{}, data[start:end]...)
}

// bytesLstripWhitespace strips leading whitespace
func bytesLstripWhitespace(data []byte) []byte {
	start := 0
	for start < len(data) && isWhitespace(data[start]) {
		start++
	}
	return append([]byte{}, data[start:]...)
}

// bytesRstripWhitespace strips trailing whitespace
func bytesRstripWhitespace(data []byte) []byte {
	end := len(data)
	for end > 0 && isWhitespace(data[end-1]) {
		end--
	}
	return append([]byte{}, data[:end]...)
}

// bytesStripChars strips leading and trailing chars in the given set
func bytesStripChars(data, chars []byte) []byte {
	charSet := make(map[byte]bool)
	for _, c := range chars {
		charSet[c] = true
	}
	start := 0
	for start < len(data) && charSet[data[start]] {
		start++
	}
	end := len(data)
	for end > start && charSet[data[end-1]] {
		end--
	}
	return append([]byte{}, data[start:end]...)
}

// bytesLstripChars strips leading chars in the given set
func bytesLstripChars(data, chars []byte) []byte {
	charSet := make(map[byte]bool)
	for _, c := range chars {
		charSet[c] = true
	}
	start := 0
	for start < len(data) && charSet[data[start]] {
		start++
	}
	return append([]byte{}, data[start:]...)
}

// bytesRstripChars strips trailing chars in the given set
func bytesRstripChars(data, chars []byte) []byte {
	charSet := make(map[byte]bool)
	for _, c := range chars {
		charSet[c] = true
	}
	end := len(data)
	for end > 0 && charSet[data[end-1]] {
		end--
	}
	return append([]byte{}, data[:end]...)
}

// bytesStartsWith checks if data starts with prefix
func bytesStartsWith(data, prefix []byte) bool {
	if len(prefix) > len(data) {
		return false
	}
	for i := 0; i < len(prefix); i++ {
		if data[i] != prefix[i] {
			return false
		}
	}
	return true
}

// bytesEndsWith checks if data ends with suffix
func bytesEndsWith(data, suffix []byte) bool {
	if len(suffix) > len(data) {
		return false
	}
	offset := len(data) - len(suffix)
	for i := 0; i < len(suffix); i++ {
		if data[offset+i] != suffix[i] {
			return false
		}
	}
	return true
}
