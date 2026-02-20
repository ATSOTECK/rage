package runtime

import (
	"fmt"
	"math"
	"math/big"
	"math/bits"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// isAttributeError checks if an error is an AttributeError.
func isAttributeError(err error) bool {
	if pyExc, ok := err.(*PyException); ok {
		return pyExc.Type() == "AttributeError"
	}
	return strings.HasPrefix(err.Error(), "AttributeError:")
}

// callMethodValue calls a class-dict method value with an instance as self.
func (vm *VM) callMethodValue(method Value, self Value, args ...Value) (Value, error) {
	allArgs := make([]Value, 1+len(args))
	allArgs[0] = self
	copy(allArgs[1:], args)
	switch fn := method.(type) {
	case *PyFunction:
		return vm.callFunction(fn, allArgs, nil)
	case *PyBuiltinFunc:
		return fn.Fn(allArgs, nil)
	default:
		return vm.call(method, allArgs, nil)
	}
}

// defaultGetAttribute performs the standard attribute lookup for a PyInstance.
// This is the logic behind object.__getattribute__: data descriptors → instance dict → class MRO.
// It does NOT call __getattr__; the caller handles that fallback.
func (vm *VM) defaultGetAttribute(o *PyInstance, name string) (Value, error) {
	// Descriptor protocol: First check class MRO for data descriptors
	for _, cls := range o.Class.Mro {
		if val, ok := cls.Dict[name]; ok {
			if prop, ok := val.(*PyProperty); ok {
				if prop.Fget == nil {
					return nil, fmt.Errorf("property '%s' has no getter", name)
				}
				return vm.call(prop.Fget, []Value{o}, nil)
			}
			if inst, ok := val.(*PyInstance); ok {
				if vm.hasMethod(inst, "__set__") || vm.hasMethod(inst, "__delete__") {
					if getResult, found, err := vm.callDunder(inst, "__get__", o, o.Class); found {
						if err != nil {
							return nil, err
						}
						return getResult, nil
					}
				}
			}
		}
	}

	// Handle __dict__ access on instances
	if name == "__dict__" && o.Dict != nil {
		d := &PyDict{Items: make(map[Value]Value), instanceOwner: o}
		for k, v := range o.Dict {
			d.DictSet(&PyString{Value: k}, v, vm)
		}
		return d, nil
	}

	// Then check instance dict or slots
	if o.Slots != nil {
		if val, ok := o.Slots[name]; ok {
			return val, nil
		}
	} else if val, ok := o.Dict[name]; ok {
		return val, nil
	}

	// Then check class MRO for methods/attributes (non-data descriptors)
	for _, cls := range o.Class.Mro {
		if val, ok := cls.Dict[name]; ok {
			if cm, ok := val.(*PyClassMethod); ok {
				if fn, ok := cm.Func.(*PyFunction); ok {
					return &PyMethod{Func: fn, Instance: o.Class}, nil
				}
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
			if sm, ok := val.(*PyStaticMethod); ok {
				return sm.Func, nil
			}
			if fn, ok := val.(*PyFunction); ok {
				return &PyMethod{Func: fn, Instance: o}, nil
			}
			if bf, ok := val.(*PyBuiltinFunc); ok {
				boundInst := Value(o)
				return &PyBuiltinFunc{
					Name:  bf.Name,
					Bound: true,
					Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
						allArgs := append([]Value{boundInst}, args...)
						return bf.Fn(allArgs, kwargs)
					},
				}, nil
			}
			if inst, ok := val.(*PyInstance); ok {
				if getResult, found, err := vm.callDunder(inst, "__get__", o, o.Class); found {
					if err != nil {
						return nil, err
					}
					return getResult, nil
				}
			}
			return val, nil
		}
	}
	return nil, fmt.Errorf("AttributeError: '%s' object has no attribute '%s'", o.Class.Name, name)
}

// builtinTypeDunders maps builtin type names to the set of dunders they support.
// Used by hasattr() and getAttr() to report dunder availability on builtin types.
var builtinTypeDunders = map[string]map[string]bool{
	"list": {
		"__len__": true, "__iter__": true, "__contains__": true, "__getitem__": true,
		"__setitem__": true, "__delitem__": true, "__add__": true, "__mul__": true,
		"__iadd__": true, "__imul__": true, "__eq__": true, "__ne__": true,
		"__str__": true, "__repr__": true, "__hash__": true, "__class__": true,
		"__reversed__": true,
	},
	"str": {
		"__len__": true, "__iter__": true, "__contains__": true, "__getitem__": true,
		"__add__": true, "__mul__": true, "__eq__": true, "__ne__": true,
		"__lt__": true, "__le__": true, "__gt__": true, "__ge__": true,
		"__str__": true, "__repr__": true, "__hash__": true, "__class__": true,
		"__mod__": true,
	},
	"dict": {
		"__len__": true, "__iter__": true, "__contains__": true, "__getitem__": true,
		"__setitem__": true, "__delitem__": true, "__eq__": true, "__ne__": true,
		"__str__": true, "__repr__": true, "__class__": true,
	},
	"set": {
		"__len__": true, "__iter__": true, "__contains__": true,
		"__eq__": true, "__ne__": true, "__sub__": true, "__and__": true,
		"__or__": true, "__xor__": true,
		"__str__": true, "__repr__": true, "__class__": true,
	},
	"frozenset": {
		"__len__": true, "__iter__": true, "__contains__": true,
		"__eq__": true, "__ne__": true, "__sub__": true, "__and__": true,
		"__or__": true, "__xor__": true, "__hash__": true,
		"__str__": true, "__repr__": true, "__class__": true,
	},
	"tuple": {
		"__len__": true, "__iter__": true, "__contains__": true, "__getitem__": true,
		"__add__": true, "__mul__": true, "__eq__": true, "__ne__": true,
		"__lt__": true, "__le__": true, "__gt__": true, "__ge__": true,
		"__str__": true, "__repr__": true, "__hash__": true, "__class__": true,
	},
	"int": {
		"__add__": true, "__sub__": true, "__mul__": true, "__truediv__": true,
		"__floordiv__": true, "__mod__": true, "__pow__": true,
		"__and__": true, "__or__": true, "__xor__": true,
		"__lshift__": true, "__rshift__": true, "__neg__": true, "__pos__": true,
		"__abs__": true, "__invert__": true,
		"__eq__": true, "__ne__": true, "__lt__": true, "__le__": true,
		"__gt__": true, "__ge__": true,
		"__str__": true, "__repr__": true, "__hash__": true, "__class__": true,
		"__int__": true, "__float__": true, "__bool__": true, "__index__": true,
		"__ceil__": true, "__floor__": true, "__round__": true, "__trunc__": true,
	},
	"float": {
		"__add__": true, "__sub__": true, "__mul__": true, "__truediv__": true,
		"__floordiv__": true, "__mod__": true, "__pow__": true,
		"__neg__": true, "__pos__": true, "__abs__": true,
		"__eq__": true, "__ne__": true, "__lt__": true, "__le__": true,
		"__gt__": true, "__ge__": true,
		"__str__": true, "__repr__": true, "__hash__": true, "__class__": true,
		"__int__": true, "__float__": true, "__bool__": true,
		"__ceil__": true, "__floor__": true, "__round__": true, "__trunc__": true,
	},
	"bool": {
		"__add__": true, "__sub__": true, "__mul__": true, "__truediv__": true,
		"__floordiv__": true, "__mod__": true, "__pow__": true,
		"__and__": true, "__or__": true, "__xor__": true,
		"__eq__": true, "__ne__": true, "__lt__": true, "__le__": true,
		"__gt__": true, "__ge__": true,
		"__str__": true, "__repr__": true, "__hash__": true, "__class__": true,
		"__int__": true, "__float__": true, "__bool__": true,
	},
	"NoneType": {
		"__str__": true, "__repr__": true, "__hash__": true, "__class__": true,
		"__bool__": true, "__eq__": true, "__ne__": true,
	},
	"bytes": {
		"__len__": true, "__iter__": true, "__contains__": true, "__getitem__": true,
		"__add__": true, "__mul__": true, "__eq__": true, "__ne__": true,
		"__str__": true, "__repr__": true, "__hash__": true, "__class__": true,
	},
	"range": {
		"__len__": true, "__iter__": true, "__contains__": true, "__getitem__": true,
		"__eq__": true, "__ne__": true,
		"__str__": true, "__repr__": true, "__hash__": true, "__class__": true,
	},
	"complex": {
		"__add__": true, "__sub__": true, "__mul__": true, "__truediv__": true,
		"__pow__": true, "__neg__": true, "__pos__": true, "__abs__": true,
		"__eq__": true, "__ne__": true,
		"__str__": true, "__repr__": true, "__hash__": true, "__class__": true,
	},
}

// builtinValueTypeName returns the Python type name for a builtin value, or "" if unknown.
func builtinValueTypeName(v Value) string {
	switch v.(type) {
	case *PyList:
		return "list"
	case *PyString:
		return "str"
	case *PyDict:
		return "dict"
	case *PySet:
		return "set"
	case *PyFrozenSet:
		return "frozenset"
	case *PyTuple:
		return "tuple"
	case *PyBool:
		return "bool"
	case *PyInt:
		return "int"
	case *PyFloat:
		return "float"
	case *PyNone:
		return "NoneType"
	case *PyBytes:
		return "bytes"
	case *PyRange:
		return "range"
	case *PyComplex:
		return "complex"
	}
	return ""
}

// builtinHasDunder returns true if the given builtin type name supports the given dunder.
func builtinHasDunder(typeName, dunder string) bool {
	if dunders, ok := builtinTypeDunders[typeName]; ok {
		return dunders[dunder]
	}
	return false
}

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
		return nil, fmt.Errorf("AttributeError: 'generator' object has no attribute '%s'", name)

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
		return nil, fmt.Errorf("AttributeError: 'coroutine' object has no attribute '%s'", name)

	case *PyComplex:
		switch name {
		case "real":
			return &PyFloat{Value: o.Real}, nil
		case "imag":
			return &PyFloat{Value: o.Imag}, nil
		case "conjugate":
			c := o
			return &PyBuiltinFunc{Name: "complex.conjugate", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				return MakeComplex(c.Real, -c.Imag), nil
			}}, nil
		}
		return nil, fmt.Errorf("AttributeError: 'complex' object has no attribute '%s'", name)

	case *PyException:
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
		return nil, fmt.Errorf("AttributeError: '%s' object has no attribute '%s'", vm.typeName(obj), name)
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
		return nil, fmt.Errorf("AttributeError: 'property' object has no attribute '%s'", name)
	case *PySuper:
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

	case *PyInstance:
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
	case *GenericAlias:
		switch name {
		case "__origin__":
			return o.Origin, nil
		case "__args__":
			return &PyTuple{Items: o.Args}, nil
		case "__mro_entries__":
			origin := o.Origin
			return &PyBuiltinFunc{
				Name: "__mro_entries__",
				Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
					// Return the origin class as a tuple for MRO resolution
					if cls, ok := origin.(*PyClass); ok {
						return &PyTuple{Items: []Value{cls}}, nil
					}
					return &PyTuple{Items: []Value{origin}}, nil
				},
			}, nil
		}
		return nil, fmt.Errorf("AttributeError: 'GenericAlias' object has no attribute '%s'", name)
	case *PyDict:
		d := o
		switch name {
		case "get":
			return &PyBuiltinFunc{Name: "dict.get", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) < 1 {
					return nil, fmt.Errorf("get() requires at least 1 argument")
				}
				key := args[0]
				def := Value(None)
				if len(args) > 1 {
					def = args[1]
				}
				val, found := d.DictGet(key, vm)
				if found {
					return val, nil
				}
				return def, nil
			}}, nil
		case "keys":
			return &PyBuiltinFunc{Name: "dict.keys", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				keys := make([]Value, len(d.Keys(vm)))
				copy(keys, d.Keys(vm))
				return &PyList{Items: keys}, nil
			}}, nil
		case "values":
			return &PyBuiltinFunc{Name: "dict.values", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				orderedKeys := d.Keys(vm)
				vals := make([]Value, 0, len(orderedKeys))
				for _, k := range orderedKeys {
					if v, ok := d.DictGet(k, vm); ok {
						vals = append(vals, v)
					}
				}
				return &PyList{Items: vals}, nil
			}}, nil
		case "items":
			return &PyBuiltinFunc{Name: "dict.items", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				orderedKeys := d.Keys(vm)
				items := make([]Value, 0, len(orderedKeys))
				for _, k := range orderedKeys {
					if v, ok := d.DictGet(k, vm); ok {
						items = append(items, &PyTuple{Items: []Value{k, v}})
					}
				}
				return &PyList{Items: items}, nil
			}}, nil
		case "update":
			return &PyBuiltinFunc{Name: "dict.update", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) > 0 {
					switch src := args[0].(type) {
					case *PyDict:
						for _, k := range src.Keys(vm) {
							if v, ok := src.DictGet(k, vm); ok {
								d.DictSet(k, v, vm)
							}
						}
					default:
						items, err := vm.toList(args[0])
						if err != nil {
							return nil, err
						}
						for _, item := range items {
							pair, err := vm.toList(item)
							if err != nil {
								return nil, err
							}
							if len(pair) != 2 {
								return nil, fmt.Errorf("ValueError: dictionary update sequence element has length %d; 2 is required", len(pair))
							}
							d.DictSet(pair[0], pair[1], vm)
						}
					}
				}
				for k, v := range kwargs {
					d.DictSet(&PyString{Value: k}, v, vm)
				}
				return None, nil
			}}, nil
		case "pop":
			return &PyBuiltinFunc{Name: "dict.pop", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) < 1 {
					return nil, fmt.Errorf("pop expected at least 1 argument")
				}
				key := args[0]
				val, found := d.DictGet(key, vm)
				if found {
					d.DictDelete(key, vm)
					return val, nil
				}
				if len(args) > 1 {
					return args[1], nil
				}
				return nil, fmt.Errorf("KeyError: %s", vm.str(key))
			}}, nil
		case "popitem":
			return &PyBuiltinFunc{Name: "dict.popitem", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(d.Items) == 0 {
					return nil, fmt.Errorf("KeyError: 'popitem(): dictionary is empty'")
				}
				keys := d.Keys(vm)
				lastKey := keys[len(keys)-1]
				lastVal, _ := d.DictGet(lastKey, vm)
				d.DictDelete(lastKey, vm)
				return &PyTuple{Items: []Value{lastKey, lastVal}}, nil
			}}, nil
		case "setdefault":
			return &PyBuiltinFunc{Name: "dict.setdefault", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) < 1 {
					return nil, fmt.Errorf("setdefault() takes at least 1 argument")
				}
				key := args[0]
				val, found := d.DictGet(key, vm)
				if found {
					return val, nil
				}
				def := Value(None)
				if len(args) > 1 {
					def = args[1]
				}
				d.DictSet(key, def, vm)
				return def, nil
			}}, nil
		case "clear":
			return &PyBuiltinFunc{Name: "dict.clear", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				d.Items = make(map[Value]Value)
				d.buckets = make(map[uint64][]dictEntry)
				d.orderedKeys = nil
				return None, nil
			}}, nil
		case "copy":
			return &PyBuiltinFunc{Name: "dict.copy", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				cp := &PyDict{Items: make(map[Value]Value), buckets: make(map[uint64][]dictEntry)}
				for _, k := range d.Keys(vm) {
					if v, ok := d.DictGet(k, vm); ok {
						cp.DictSet(k, v, vm)
					}
				}
				return cp, nil
			}}, nil
		case "fromkeys":
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
	case *PySet:
		s := o
		switch name {
		case "add":
			return &PyBuiltinFunc{Name: "set.add", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("add() takes exactly 1 argument")
				}
				if !isHashable(args[0]) {
					return nil, fmt.Errorf("TypeError: unhashable type: '%s'", vm.typeName(args[0]))
				}
				s.SetAdd(args[0], vm)
				return None, nil
			}}, nil
		case "discard":
			return &PyBuiltinFunc{Name: "set.discard", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("discard() takes exactly 1 argument")
				}
				s.SetRemove(args[0], vm)
				return None, nil
			}}, nil
		case "remove":
			return &PyBuiltinFunc{Name: "set.remove", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("remove() takes exactly 1 argument")
				}
				if !s.SetContains(args[0], vm) {
					return nil, fmt.Errorf("KeyError: %s", vm.str(args[0]))
				}
				s.SetRemove(args[0], vm)
				return None, nil
			}}, nil
		case "pop":
			return &PyBuiltinFunc{Name: "set.pop", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(s.Items) == 0 {
					return nil, fmt.Errorf("KeyError: 'pop from an empty set'")
				}
				var item Value
				for k := range s.Items {
					item = k
					break
				}
				s.SetRemove(item, vm)
				return item, nil
			}}, nil
		case "clear":
			return &PyBuiltinFunc{Name: "set.clear", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				s.Items = make(map[Value]struct{})
				s.buckets = make(map[uint64][]setEntry)
				return None, nil
			}}, nil
		case "copy":
			return &PyBuiltinFunc{Name: "set.copy", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				cp := &PySet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
				for k := range s.Items {
					cp.SetAdd(k, vm)
				}
				return cp, nil
			}}, nil
		case "update":
			return &PyBuiltinFunc{Name: "set.update", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				for _, arg := range args {
					items, err := vm.toList(arg)
					if err != nil {
						return nil, err
					}
					for _, item := range items {
						if !isHashable(item) {
							return nil, fmt.Errorf("TypeError: unhashable type: '%s'", vm.typeName(item))
						}
						s.SetAdd(item, vm)
					}
				}
				return None, nil
			}}, nil
		case "union":
			return &PyBuiltinFunc{Name: "set.union", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				result := &PySet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
				for k := range s.Items {
					result.SetAdd(k, vm)
				}
				for _, arg := range args {
					items, err := vm.toList(arg)
					if err != nil {
						return nil, err
					}
					for _, item := range items {
						result.SetAdd(item, vm)
					}
				}
				return result, nil
			}}, nil
		case "intersection":
			return &PyBuiltinFunc{Name: "set.intersection", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				result := &PySet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
				if len(args) == 0 {
					for k := range s.Items {
						result.SetAdd(k, vm)
					}
					return result, nil
				}
				for k := range s.Items {
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
						result.SetAdd(k, vm)
					}
				}
				return result, nil
			}}, nil
		case "difference":
			return &PyBuiltinFunc{Name: "set.difference", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				result := &PySet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
				for k := range s.Items {
					result.SetAdd(k, vm)
				}
				for _, arg := range args {
					items, err := vm.toList(arg)
					if err != nil {
						return nil, err
					}
					for _, item := range items {
						result.SetRemove(item, vm)
					}
				}
				return result, nil
			}}, nil
		case "symmetric_difference":
			return &PyBuiltinFunc{Name: "set.symmetric_difference", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("symmetric_difference() takes exactly 1 argument")
				}
				other, err := vm.toList(args[0])
				if err != nil {
					return nil, err
				}
				result := &PySet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
				for k := range s.Items {
					found := false
					for _, item := range other {
						if vm.equal(k, item) {
							found = true
							break
						}
					}
					if !found {
						result.SetAdd(k, vm)
					}
				}
				for _, item := range other {
					if !s.SetContains(item, vm) {
						result.SetAdd(item, vm)
					}
				}
				return result, nil
			}}, nil
		case "issubset":
			return &PyBuiltinFunc{Name: "set.issubset", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("issubset() takes exactly 1 argument")
				}
				other, err := vm.toList(args[0])
				if err != nil {
					return nil, err
				}
				for k := range s.Items {
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
		case "issuperset":
			return &PyBuiltinFunc{Name: "set.issuperset", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("issuperset() takes exactly 1 argument")
				}
				other, err := vm.toList(args[0])
				if err != nil {
					return nil, err
				}
				for _, item := range other {
					if !s.SetContains(item, vm) {
						return False, nil
					}
				}
				return True, nil
			}}, nil
		case "isdisjoint":
			return &PyBuiltinFunc{Name: "set.isdisjoint", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("isdisjoint() takes exactly 1 argument")
				}
				other, err := vm.toList(args[0])
				if err != nil {
					return nil, err
				}
				for _, item := range other {
					if s.SetContains(item, vm) {
						return False, nil
					}
				}
				return True, nil
			}}, nil
		case "difference_update":
			return &PyBuiltinFunc{Name: "set.difference_update", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				for _, arg := range args {
					items, err := vm.toList(arg)
					if err != nil {
						return nil, err
					}
					for _, item := range items {
						s.SetRemove(item, vm)
					}
				}
				return None, nil
			}}, nil
		case "intersection_update":
			return &PyBuiltinFunc{Name: "set.intersection_update", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) == 0 {
					return None, nil
				}
				// Collect all items to keep: those in self AND in all args
				keep := &PySet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
				for k := range s.Items {
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
						keep.SetAdd(k, vm)
					}
				}
				// Replace contents of s with keep
				s.Items = keep.Items
				s.buckets = keep.buckets
				s.size = keep.size
				return None, nil
			}}, nil
		case "symmetric_difference_update":
			return &PyBuiltinFunc{Name: "set.symmetric_difference_update", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("symmetric_difference_update() takes exactly 1 argument")
				}
				other, err := vm.toList(args[0])
				if err != nil {
					return nil, err
				}
				// Items in self but not in other stay
				// Items in other but not in self get added
				// Items in both get removed
				toRemove := []Value{}
				toAdd := []Value{}
				for k := range s.Items {
					for _, item := range other {
						if vm.equal(k, item) {
							toRemove = append(toRemove, k)
							break
						}
					}
				}
				for _, item := range other {
					if !s.SetContains(item, vm) {
						toAdd = append(toAdd, item)
					}
				}
				for _, item := range toRemove {
					s.SetRemove(item, vm)
				}
				for _, item := range toAdd {
					s.SetAdd(item, vm)
				}
				return None, nil
			}}, nil
		}
	case *PyTuple:
		tpl := o
		switch name {
		case "count":
			return &PyBuiltinFunc{Name: "tuple.count", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("count() takes exactly 1 argument")
				}
				count := 0
				for _, item := range tpl.Items {
					if vm.equal(item, args[0]) {
						count++
					}
				}
				return MakeInt(int64(count)), nil
			}}, nil
		case "index":
			return &PyBuiltinFunc{Name: "tuple.index", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) < 1 {
					return nil, fmt.Errorf("index() takes at least 1 argument")
				}
				start := 0
				end := len(tpl.Items)
				if len(args) > 1 {
					start = int(vm.toInt(args[1]))
				}
				if len(args) > 2 {
					end = int(vm.toInt(args[2]))
				}
				for i := start; i < end; i++ {
					if vm.equal(tpl.Items[i], args[0]) {
						return MakeInt(int64(i)), nil
					}
				}
				return nil, fmt.Errorf("ValueError: tuple.index(x): x not in tuple")
			}}, nil
		}
	case *PyInt:
		v := o
		switch name {
		// Properties
		case "real":
			return v, nil
		case "imag":
			return MakeInt(0), nil
		case "numerator":
			return v, nil
		case "denominator":
			return MakeInt(1), nil

		// Methods
		case "bit_length":
			return &PyBuiltinFunc{Name: "int.bit_length", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if v.BigValue != nil {
					return MakeInt(int64(v.BigValue.BitLen())), nil
				}
				val := v.Value
				if val < 0 {
					val = -val
				}
				if val == 0 {
					return MakeInt(0), nil
				}
				return MakeInt(int64(bits.Len64(uint64(val)))), nil
			}}, nil

		case "bit_count":
			return &PyBuiltinFunc{Name: "int.bit_count", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if v.BigValue != nil {
					// Count bits in absolute value
					abs := new(big.Int).Abs(v.BigValue)
					count := 0
					for _, word := range abs.Bits() {
						count += bits.OnesCount(uint(word))
					}
					return MakeInt(int64(count)), nil
				}
				val := v.Value
				if val < 0 {
					val = -val
				}
				return MakeInt(int64(bits.OnesCount64(uint64(val)))), nil
			}}, nil

		case "conjugate":
			return &PyBuiltinFunc{Name: "int.conjugate", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				return v, nil
			}}, nil

		case "as_integer_ratio":
			return &PyBuiltinFunc{Name: "int.as_integer_ratio", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				return &PyTuple{Items: []Value{v, MakeInt(1)}}, nil
			}}, nil

		case "to_bytes":
			return &PyBuiltinFunc{Name: "int.to_bytes", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				// to_bytes(length, byteorder, *, signed=False)
				length := 0
				byteorder := ""
				signed := false

				if len(args) >= 1 {
					length = int(vm.toInt(args[0]))
				} else if kv, ok := kwargs["length"]; ok {
					length = int(vm.toInt(kv))
				} else {
					return nil, fmt.Errorf("TypeError: to_bytes() missing required argument: 'length'")
				}

				if len(args) >= 2 {
					if s, ok := args[1].(*PyString); ok {
						byteorder = s.Value
					} else {
						return nil, fmt.Errorf("TypeError: to_bytes() argument 'byteorder' must be str")
					}
				} else if kv, ok := kwargs["byteorder"]; ok {
					if s, ok := kv.(*PyString); ok {
						byteorder = s.Value
					} else {
						return nil, fmt.Errorf("TypeError: to_bytes() argument 'byteorder' must be str")
					}
				} else {
					return nil, fmt.Errorf("TypeError: to_bytes() missing required argument: 'byteorder'")
				}

				if byteorder != "big" && byteorder != "little" {
					return nil, fmt.Errorf("ValueError: byteorder must be either 'little' or 'big'")
				}

				if kv, ok := kwargs["signed"]; ok {
					signed = vm.Truthy(kv)
				}

				val := v.Value
				if v.BigValue != nil {
					if v.BigValue.IsInt64() {
						val = v.BigValue.Int64()
					} else {
						return nil, fmt.Errorf("OverflowError: int too big to convert")
					}
				}

				if val < 0 && !signed {
					return nil, fmt.Errorf("OverflowError: can't convert negative int to unsigned")
				}

				var uval uint64
				if val < 0 {
					// Two's complement for negative values
					uval = uint64(val)
				} else {
					uval = uint64(val)
				}

				result := make([]byte, length)

				if signed && val < 0 {
					// Fill with 0xff for negative numbers (sign extension)
					for i := range result {
						result[i] = 0xff
					}
				}

				// Write bytes in big-endian order first
				for i := length - 1; i >= 0; i-- {
					result[i] = byte(uval & 0xff)
					uval >>= 8
				}

				// Check for overflow: if uval still has bits, the number doesn't fit
				if val >= 0 && uval != 0 {
					return nil, fmt.Errorf("OverflowError: int too big to convert")
				}
				if signed && val < 0 {
					// For negative signed, check that sign bit is set in the MSB
					if length > 0 && result[0]&0x80 == 0 {
						return nil, fmt.Errorf("OverflowError: int too big to convert")
					}
				} else if !signed && val >= 0 {
					// Already checked uval == 0 above
				} else if signed && val >= 0 {
					// Check sign bit isn't set (would look negative)
					if length > 0 && result[0]&0x80 != 0 {
						return nil, fmt.Errorf("OverflowError: int too big to convert")
					}
				}

				if byteorder == "little" {
					// Reverse the bytes
					for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
						result[i], result[j] = result[j], result[i]
					}
				}

				return &PyBytes{Value: result}, nil
			}}, nil

		case "from_bytes":
			return &PyBuiltinFunc{Name: "int.from_bytes", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				return intFromBytesImpl(vm, args, kwargs)
			}}, nil

		// Dunder methods
		case "__abs__":
			return &PyBuiltinFunc{Name: "int.__abs__", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if v.BigValue != nil {
					return MakeBigInt(new(big.Int).Abs(v.BigValue)), nil
				}
				val := v.Value
				if val < 0 {
					val = -val
				}
				return MakeInt(val), nil
			}}, nil

		case "__bool__":
			return &PyBuiltinFunc{Name: "int.__bool__", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if v.BigValue != nil {
					if v.BigValue.Sign() == 0 {
						return False, nil
					}
					return True, nil
				}
				if v.Value == 0 {
					return False, nil
				}
				return True, nil
			}}, nil

		case "__ceil__":
			return &PyBuiltinFunc{Name: "int.__ceil__", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				return v, nil
			}}, nil

		case "__floor__":
			return &PyBuiltinFunc{Name: "int.__floor__", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				return v, nil
			}}, nil

		case "__trunc__":
			return &PyBuiltinFunc{Name: "int.__trunc__", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				return v, nil
			}}, nil

		case "__round__":
			return &PyBuiltinFunc{Name: "int.__round__", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				// __round__(ndigits=None)
				if len(args) == 0 {
					return v, nil
				}
				if args[0] == None {
					return v, nil
				}
				ndigits := int(vm.toInt(args[0]))
				if ndigits >= 0 {
					return v, nil
				}
				// Negative ndigits: round to nearest 10^(-ndigits) with banker's rounding
				val := v.Value
				pow := int64(1)
				for i := 0; i < -ndigits; i++ {
					pow *= 10
				}
				remainder := val % pow
				truncated := val - remainder
				half := pow / 2

				if remainder < 0 {
					remainder = -remainder
				}

				if remainder > half {
					if val >= 0 {
						truncated += pow
					} else {
						truncated -= pow
					}
				} else if remainder == half {
					// Banker's rounding: round to even
					quotient := truncated / pow
					if quotient%2 != 0 {
						if val >= 0 {
							truncated += pow
						} else {
							truncated -= pow
						}
					}
				}

				return MakeInt(truncated), nil
			}}, nil

		case "__int__":
			return &PyBuiltinFunc{Name: "int.__int__", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				return v, nil
			}}, nil

		case "__float__":
			return &PyBuiltinFunc{Name: "int.__float__", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if v.BigValue != nil {
					f, _ := v.BigValue.Float64()
					return &PyFloat{Value: f}, nil
				}
				return &PyFloat{Value: float64(v.Value)}, nil
			}}, nil

		case "__index__":
			return &PyBuiltinFunc{Name: "int.__index__", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				return v, nil
			}}, nil
		}
		return nil, fmt.Errorf("AttributeError: 'int' object has no attribute '%s'", name)

	case *PyFloat:
		f := o
		switch name {
		// Properties
		case "real":
			return f, nil
		case "imag":
			return &PyFloat{Value: 0.0}, nil

		// Methods
		case "is_integer":
			return &PyBuiltinFunc{Name: "float.is_integer", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if math.IsInf(f.Value, 0) || math.IsNaN(f.Value) {
					return False, nil
				}
				if f.Value == math.Trunc(f.Value) {
					return True, nil
				}
				return False, nil
			}}, nil

		case "conjugate":
			return &PyBuiltinFunc{Name: "float.conjugate", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				return f, nil
			}}, nil

		case "hex":
			return &PyBuiltinFunc{Name: "float.hex", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				return &PyString{Value: floatHex(f.Value)}, nil
			}}, nil

		case "fromhex":
			return &PyBuiltinFunc{Name: "float.fromhex", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				return floatFromHexImpl(args)
			}}, nil

		case "as_integer_ratio":
			return &PyBuiltinFunc{Name: "float.as_integer_ratio", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if math.IsInf(f.Value, 0) {
					return nil, fmt.Errorf("OverflowError: cannot convert Infinity to integer ratio")
				}
				if math.IsNaN(f.Value) {
					return nil, fmt.Errorf("ValueError: cannot convert NaN to integer ratio")
				}
				num, den := floatAsIntegerRatio(f.Value)
				return &PyTuple{Items: []Value{num, den}}, nil
			}}, nil

		// Dunder methods
		case "__abs__":
			return &PyBuiltinFunc{Name: "float.__abs__", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				return &PyFloat{Value: math.Abs(f.Value)}, nil
			}}, nil

		case "__bool__":
			return &PyBuiltinFunc{Name: "float.__bool__", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if f.Value == 0.0 {
					return False, nil
				}
				return True, nil
			}}, nil

		case "__ceil__":
			return &PyBuiltinFunc{Name: "float.__ceil__", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if math.IsInf(f.Value, 0) || math.IsNaN(f.Value) {
					return nil, fmt.Errorf("OverflowError: cannot convert float %s to integer", formatSpecialFloat(f.Value))
				}
				return MakeInt(int64(math.Ceil(f.Value))), nil
			}}, nil

		case "__floor__":
			return &PyBuiltinFunc{Name: "float.__floor__", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if math.IsInf(f.Value, 0) || math.IsNaN(f.Value) {
					return nil, fmt.Errorf("OverflowError: cannot convert float %s to integer", formatSpecialFloat(f.Value))
				}
				return MakeInt(int64(math.Floor(f.Value))), nil
			}}, nil

		case "__trunc__":
			return &PyBuiltinFunc{Name: "float.__trunc__", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if math.IsInf(f.Value, 0) || math.IsNaN(f.Value) {
					return nil, fmt.Errorf("OverflowError: cannot convert float %s to integer", formatSpecialFloat(f.Value))
				}
				return MakeInt(int64(math.Trunc(f.Value))), nil
			}}, nil

		case "__round__":
			return &PyBuiltinFunc{Name: "float.__round__", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) == 0 {
					// No ndigits: return int with banker's rounding
					if math.IsInf(f.Value, 0) || math.IsNaN(f.Value) {
						return nil, fmt.Errorf("OverflowError: cannot convert float %s to integer", formatSpecialFloat(f.Value))
					}
					return MakeInt(int64(math.RoundToEven(f.Value))), nil
				}
				if args[0] == None {
					if math.IsInf(f.Value, 0) || math.IsNaN(f.Value) {
						return nil, fmt.Errorf("OverflowError: cannot convert float %s to integer", formatSpecialFloat(f.Value))
					}
					return MakeInt(int64(math.RoundToEven(f.Value))), nil
				}
				ndigits := int(vm.toInt(args[0]))
				pow := math.Pow(10, float64(ndigits))
				return &PyFloat{Value: math.RoundToEven(f.Value*pow) / pow}, nil
			}}, nil

		case "__int__":
			return &PyBuiltinFunc{Name: "float.__int__", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if math.IsInf(f.Value, 0) || math.IsNaN(f.Value) {
					return nil, fmt.Errorf("OverflowError: cannot convert float %s to integer", formatSpecialFloat(f.Value))
				}
				return MakeInt(int64(f.Value)), nil
			}}, nil

		case "__float__":
			return &PyBuiltinFunc{Name: "float.__float__", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				return f, nil
			}}, nil
		}
		return nil, fmt.Errorf("AttributeError: 'float' object has no attribute '%s'", name)

	case *PyList:
		lst := o
		switch name {
		case "append":
			return &PyBuiltinFunc{Name: "list.append", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("append() takes exactly 1 argument")
				}
				lst.Items = append(lst.Items, args[0])
				return None, nil
			}}, nil
		case "pop":
			return &PyBuiltinFunc{Name: "list.pop", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(lst.Items) == 0 {
					return nil, fmt.Errorf("IndexError: pop from empty list")
				}
				idx := len(lst.Items) - 1
				if len(args) > 0 {
					idx = int(vm.toInt(args[0]))
					if idx < 0 {
						idx += len(lst.Items)
					}
				}
				if idx < 0 || idx >= len(lst.Items) {
					return nil, fmt.Errorf("IndexError: pop index out of range")
				}
				val := lst.Items[idx]
				lst.Items = append(lst.Items[:idx], lst.Items[idx+1:]...)
				return val, nil
			}}, nil
		case "extend":
			return &PyBuiltinFunc{Name: "list.extend", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("extend() takes exactly 1 argument")
				}
				items, err := vm.toList(args[0])
				if err != nil {
					return nil, err
				}
				lst.Items = append(lst.Items, items...)
				return None, nil
			}}, nil
		case "insert":
			return &PyBuiltinFunc{Name: "list.insert", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) != 2 {
					return nil, fmt.Errorf("insert() takes exactly 2 arguments")
				}
				idx := int(vm.toInt(args[0]))
				if idx < 0 {
					idx += len(lst.Items)
					if idx < 0 {
						idx = 0
					}
				}
				if idx >= len(lst.Items) {
					lst.Items = append(lst.Items, args[1])
				} else {
					lst.Items = append(lst.Items, nil)
					copy(lst.Items[idx+1:], lst.Items[idx:])
					lst.Items[idx] = args[1]
				}
				return None, nil
			}}, nil
		case "remove":
			return &PyBuiltinFunc{Name: "list.remove", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("remove() takes exactly 1 argument")
				}
				for i, item := range lst.Items {
					if vm.equal(item, args[0]) {
						lst.Items = append(lst.Items[:i], lst.Items[i+1:]...)
						return None, nil
					}
				}
				return nil, fmt.Errorf("ValueError: list.remove(x): x not in list")
			}}, nil
		case "clear":
			return &PyBuiltinFunc{Name: "list.clear", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				lst.Items = []Value{}
				return None, nil
			}}, nil
		case "index":
			return &PyBuiltinFunc{Name: "list.index", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) < 1 {
					return nil, fmt.Errorf("index() takes at least 1 argument")
				}
				start := 0
				end := len(lst.Items)
				if len(args) > 1 {
					start = int(vm.toInt(args[1]))
				}
				if len(args) > 2 {
					end = int(vm.toInt(args[2]))
				}
				for i := start; i < end && i < len(lst.Items); i++ {
					if vm.equal(lst.Items[i], args[0]) {
						return MakeInt(int64(i)), nil
					}
				}
				return nil, fmt.Errorf("ValueError: %s is not in list", vm.str(args[0]))
			}}, nil
		case "count":
			return &PyBuiltinFunc{Name: "list.count", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("count() takes exactly 1 argument")
				}
				count := 0
				for _, item := range lst.Items {
					if vm.equal(item, args[0]) {
						count++
					}
				}
				return MakeInt(int64(count)), nil
			}}, nil
		case "reverse":
			return &PyBuiltinFunc{Name: "list.reverse", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				for i, j := 0, len(lst.Items)-1; i < j; i, j = i+1, j-1 {
					lst.Items[i], lst.Items[j] = lst.Items[j], lst.Items[i]
				}
				return None, nil
			}}, nil
		case "sort":
			return &PyBuiltinFunc{Name: "list.sort", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				var keyFn Value
				if k, ok := kwargs["key"]; ok && k != None {
					keyFn = k
				}
				reverse := false
				if r, ok := kwargs["reverse"]; ok {
					reverse = vm.truthy(r)
				}
				var sortErr error
				sort.SliceStable(lst.Items, func(i, j int) bool {
					if sortErr != nil {
						return false
					}
					a, b := lst.Items[i], lst.Items[j]
					if keyFn != nil {
						var err error
						a, err = vm.call(keyFn, []Value{a}, nil)
						if err != nil {
							sortErr = err
							return false
						}
						b, err = vm.call(keyFn, []Value{b}, nil)
						if err != nil {
							sortErr = err
							return false
						}
					}
					// For reverse sort, compare b < a instead of a < b to maintain stability.
					// Using !less would break stability because equal elements (less=false)
					// would return true, swapping their order.
					var cmpA, cmpB Value
					if reverse {
						cmpA, cmpB = b, a
					} else {
						cmpA, cmpB = a, b
					}
					result := vm.compareOp(OpCompareLt, cmpA, cmpB)
					if result == nil {
						// compareOp set vm.currentException (e.g. TypeError for incompatible types)
						if vm.currentException != nil {
							sortErr = vm.currentException
							vm.currentException = nil
						}
						return false
					}
					return vm.truthy(result)
				})
				if sortErr != nil {
					return nil, sortErr
				}
				return None, nil
			}}, nil
		case "copy":
			return &PyBuiltinFunc{Name: "list.copy", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				cp := make([]Value, len(lst.Items))
				copy(cp, lst.Items)
				return &PyList{Items: cp}, nil
			}}, nil
		}
	case *PyString:
		str := o
		switch name {
		case "upper":
			return &PyBuiltinFunc{Name: "str.upper", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				return &PyString{Value: strings.ToUpper(str.Value)}, nil
			}}, nil
		case "lower":
			return &PyBuiltinFunc{Name: "str.lower", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				return &PyString{Value: strings.ToLower(str.Value)}, nil
			}}, nil
		case "split":
			return &PyBuiltinFunc{Name: "str.split", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				var strParts []string
				if len(args) == 0 {
					strParts = strings.Fields(str.Value)
				} else {
					sep := vm.str(args[0])
					maxSplit := -1
					if len(args) > 1 {
						maxSplit = int(vm.toInt(args[1]))
					}
					if maxSplit < 0 {
						strParts = strings.Split(str.Value, sep)
					} else {
						strParts = strings.SplitN(str.Value, sep, maxSplit+1)
					}
				}
				parts := make([]Value, len(strParts))
				for i, s := range strParts {
					parts[i] = &PyString{Value: s}
				}
				return &PyList{Items: parts}, nil
			}}, nil
		case "rsplit":
			return &PyBuiltinFunc{Name: "str.rsplit", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) == 0 {
					strParts := strings.Fields(str.Value)
					parts := make([]Value, len(strParts))
					for i, s := range strParts {
						parts[i] = &PyString{Value: s}
					}
					return &PyList{Items: parts}, nil
				}
				sep := vm.str(args[0])
				maxSplit := -1
				if len(args) > 1 {
					maxSplit = int(vm.toInt(args[1]))
				}
				if maxSplit < 0 {
					strParts := strings.Split(str.Value, sep)
					parts := make([]Value, len(strParts))
					for i, s := range strParts {
						parts[i] = &PyString{Value: s}
					}
					return &PyList{Items: parts}, nil
				}
				// rsplit from right
				s := str.Value
				var result []string
				for maxSplit > 0 {
					idx := strings.LastIndex(s, sep)
					if idx < 0 {
						break
					}
					result = append([]string{s[idx+len(sep):]}, result...)
					s = s[:idx]
					maxSplit--
				}
				result = append([]string{s}, result...)
				parts := make([]Value, len(result))
				for i, p := range result {
					parts[i] = &PyString{Value: p}
				}
				return &PyList{Items: parts}, nil
			}}, nil
		case "join":
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
					s, ok := item.(*PyString)
					if !ok {
						return nil, fmt.Errorf("TypeError: sequence item: expected str instance, %s found", vm.typeName(item))
					}
					parts = append(parts, s.Value)
				}
				return &PyString{Value: strings.Join(parts, str.Value)}, nil
			}}, nil
		case "strip":
			return &PyBuiltinFunc{Name: "str.strip", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) > 0 {
					chars := vm.str(args[0])
					return &PyString{Value: strings.Trim(str.Value, chars)}, nil
				}
				return &PyString{Value: strings.TrimSpace(str.Value)}, nil
			}}, nil
		case "lstrip":
			return &PyBuiltinFunc{Name: "str.lstrip", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) > 0 {
					chars := vm.str(args[0])
					return &PyString{Value: strings.TrimLeft(str.Value, chars)}, nil
				}
				return &PyString{Value: strings.TrimLeftFunc(str.Value, func(r rune) bool {
					return r == ' ' || r == '\t' || r == '\n' || r == '\r'
				})}, nil
			}}, nil
		case "rstrip":
			return &PyBuiltinFunc{Name: "str.rstrip", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) > 0 {
					chars := vm.str(args[0])
					return &PyString{Value: strings.TrimRight(str.Value, chars)}, nil
				}
				return &PyString{Value: strings.TrimRightFunc(str.Value, func(r rune) bool {
					return r == ' ' || r == '\t' || r == '\n' || r == '\r'
				})}, nil
			}}, nil
		case "replace":
			return &PyBuiltinFunc{Name: "str.replace", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) < 2 {
					return nil, fmt.Errorf("replace() takes at least 2 arguments")
				}
				old := vm.str(args[0])
				newStr := vm.str(args[1])
				count := -1
				if len(args) > 2 {
					count = int(vm.toInt(args[2]))
				}
				return &PyString{Value: strings.Replace(str.Value, old, newStr, count)}, nil
			}}, nil
		case "find":
			return &PyBuiltinFunc{Name: "str.find", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) < 1 {
					return nil, fmt.Errorf("find() takes at least 1 argument")
				}
				sub := vm.str(args[0])
				s := str.Value
				start := 0
				if len(args) > 1 {
					start = int(vm.toInt(args[1]))
					if start < 0 {
						start += len([]rune(s))
						if start < 0 {
							start = 0
						}
					}
				}
				if start > len(s) {
					return MakeInt(-1), nil
				}
				idx := strings.Index(s[start:], sub)
				if idx < 0 {
					return MakeInt(-1), nil
				}
				return MakeInt(int64(start + idx)), nil
			}}, nil
		case "rfind":
			return &PyBuiltinFunc{Name: "str.rfind", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) < 1 {
					return nil, fmt.Errorf("rfind() takes at least 1 argument")
				}
				sub := vm.str(args[0])
				idx := strings.LastIndex(str.Value, sub)
				return MakeInt(int64(idx)), nil
			}}, nil
		case "index":
			return &PyBuiltinFunc{Name: "str.index", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) < 1 {
					return nil, fmt.Errorf("index() takes at least 1 argument")
				}
				sub := vm.str(args[0])
				idx := strings.Index(str.Value, sub)
				if idx < 0 {
					return nil, fmt.Errorf("ValueError: substring not found")
				}
				return MakeInt(int64(idx)), nil
			}}, nil
		case "rindex":
			return &PyBuiltinFunc{Name: "str.rindex", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) < 1 {
					return nil, fmt.Errorf("rindex() takes at least 1 argument")
				}
				sub := vm.str(args[0])
				idx := strings.LastIndex(str.Value, sub)
				if idx < 0 {
					return nil, fmt.Errorf("ValueError: substring not found")
				}
				return MakeInt(int64(idx)), nil
			}}, nil
		case "startswith":
			return &PyBuiltinFunc{Name: "str.startswith", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) < 1 {
					return nil, fmt.Errorf("startswith() takes at least 1 argument")
				}
				// Handle tuple of prefixes
				if t, ok := args[0].(*PyTuple); ok {
					for _, item := range t.Items {
						prefix := vm.str(item)
						if strings.HasPrefix(str.Value, prefix) {
							return True, nil
						}
					}
					return False, nil
				}
				prefix := vm.str(args[0])
				if strings.HasPrefix(str.Value, prefix) {
					return True, nil
				}
				return False, nil
			}}, nil
		case "endswith":
			return &PyBuiltinFunc{Name: "str.endswith", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) < 1 {
					return nil, fmt.Errorf("endswith() takes at least 1 argument")
				}
				if t, ok := args[0].(*PyTuple); ok {
					for _, item := range t.Items {
						suffix := vm.str(item)
						if strings.HasSuffix(str.Value, suffix) {
							return True, nil
						}
					}
					return False, nil
				}
				suffix := vm.str(args[0])
				if strings.HasSuffix(str.Value, suffix) {
					return True, nil
				}
				return False, nil
			}}, nil
		case "count":
			return &PyBuiltinFunc{Name: "str.count", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) < 1 {
					return nil, fmt.Errorf("count() takes at least 1 argument")
				}
				sub := vm.str(args[0])
				s := str.Value
				start := 0
				end := len(s)
				if len(args) > 1 {
					start = int(vm.toInt(args[1]))
				}
				if len(args) > 2 {
					end = int(vm.toInt(args[2]))
				}
				if start > len(s) {
					return MakeInt(0), nil
				}
				if end > len(s) {
					end = len(s)
				}
				return MakeInt(int64(strings.Count(s[start:end], sub))), nil
			}}, nil
		case "center":
			return &PyBuiltinFunc{Name: "str.center", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) < 1 {
					return nil, fmt.Errorf("center() takes at least 1 argument")
				}
				width := int(vm.toInt(args[0]))
				fillChar := " "
				if len(args) > 1 {
					fillChar = vm.str(args[1])
				}
				s := str.Value
				if len(s) >= width {
					return &PyString{Value: s}, nil
				}
				total := width - len(s)
				left := total / 2
				right := total - left
				return &PyString{Value: strings.Repeat(fillChar, left) + s + strings.Repeat(fillChar, right)}, nil
			}}, nil
		case "ljust":
			return &PyBuiltinFunc{Name: "str.ljust", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) < 1 {
					return nil, fmt.Errorf("ljust() takes at least 1 argument")
				}
				width := int(vm.toInt(args[0]))
				fillChar := " "
				if len(args) > 1 {
					fillChar = vm.str(args[1])
				}
				s := str.Value
				if len(s) >= width {
					return &PyString{Value: s}, nil
				}
				return &PyString{Value: s + strings.Repeat(fillChar, width-len(s))}, nil
			}}, nil
		case "rjust":
			return &PyBuiltinFunc{Name: "str.rjust", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) < 1 {
					return nil, fmt.Errorf("rjust() takes at least 1 argument")
				}
				width := int(vm.toInt(args[0]))
				fillChar := " "
				if len(args) > 1 {
					fillChar = vm.str(args[1])
				}
				s := str.Value
				if len(s) >= width {
					return &PyString{Value: s}, nil
				}
				return &PyString{Value: strings.Repeat(fillChar, width-len(s)) + s}, nil
			}}, nil
		case "zfill":
			return &PyBuiltinFunc{Name: "str.zfill", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("zfill() takes exactly 1 argument")
				}
				width := int(vm.toInt(args[0]))
				s := str.Value
				if len(s) >= width {
					return &PyString{Value: s}, nil
				}
				if len(s) > 0 && (s[0] == '+' || s[0] == '-') {
					return &PyString{Value: string(s[0]) + strings.Repeat("0", width-len(s)) + s[1:]}, nil
				}
				return &PyString{Value: strings.Repeat("0", width-len(s)) + s}, nil
			}}, nil
		case "expandtabs":
			return &PyBuiltinFunc{Name: "str.expandtabs", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				tabSize := 8
				if len(args) > 0 {
					tabSize = int(vm.toInt(args[0]))
				}
				var result strings.Builder
				col := 0
				for _, ch := range str.Value {
					if ch == '\t' {
						spaces := tabSize - (col % tabSize)
						result.WriteString(strings.Repeat(" ", spaces))
						col += spaces
					} else if ch == '\n' || ch == '\r' {
						result.WriteRune(ch)
						col = 0
					} else {
						result.WriteRune(ch)
						col++
					}
				}
				return &PyString{Value: result.String()}, nil
			}}, nil
		case "partition":
			return &PyBuiltinFunc{Name: "str.partition", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("partition() takes exactly 1 argument")
				}
				sep := vm.str(args[0])
				idx := strings.Index(str.Value, sep)
				if idx < 0 {
					return &PyTuple{Items: []Value{
						&PyString{Value: str.Value},
						&PyString{Value: ""},
						&PyString{Value: ""},
					}}, nil
				}
				return &PyTuple{Items: []Value{
					&PyString{Value: str.Value[:idx]},
					&PyString{Value: sep},
					&PyString{Value: str.Value[idx+len(sep):]},
				}}, nil
			}}, nil
		case "rpartition":
			return &PyBuiltinFunc{Name: "str.rpartition", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("rpartition() takes exactly 1 argument")
				}
				sep := vm.str(args[0])
				idx := strings.LastIndex(str.Value, sep)
				if idx < 0 {
					return &PyTuple{Items: []Value{
						&PyString{Value: ""},
						&PyString{Value: ""},
						&PyString{Value: str.Value},
					}}, nil
				}
				return &PyTuple{Items: []Value{
					&PyString{Value: str.Value[:idx]},
					&PyString{Value: sep},
					&PyString{Value: str.Value[idx+len(sep):]},
				}}, nil
			}}, nil
		case "title":
			return &PyBuiltinFunc{Name: "str.title", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				return &PyString{Value: strings.Title(str.Value)}, nil
			}}, nil
		case "swapcase":
			return &PyBuiltinFunc{Name: "str.swapcase", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				var result strings.Builder
				for _, ch := range str.Value {
					if ch >= 'a' && ch <= 'z' {
						result.WriteRune(ch - 32)
					} else if ch >= 'A' && ch <= 'Z' {
						result.WriteRune(ch + 32)
					} else {
						result.WriteRune(ch)
					}
				}
				return &PyString{Value: result.String()}, nil
			}}, nil
		case "capitalize":
			return &PyBuiltinFunc{Name: "str.capitalize", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				s := str.Value
				if len(s) == 0 {
					return &PyString{Value: ""}, nil
				}
				runes := []rune(s)
				result := strings.ToUpper(string(runes[0:1])) + strings.ToLower(string(runes[1:]))
				return &PyString{Value: result}, nil
			}}, nil
		case "isalpha":
			return &PyBuiltinFunc{Name: "str.isalpha", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				s := str.Value
				if len(s) == 0 {
					return False, nil
				}
				for _, ch := range s {
					if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')) {
						return False, nil
					}
				}
				return True, nil
			}}, nil
		case "isdigit":
			return &PyBuiltinFunc{Name: "str.isdigit", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				s := str.Value
				if len(s) == 0 {
					return False, nil
				}
				for _, ch := range s {
					if ch < '0' || ch > '9' {
						return False, nil
					}
				}
				return True, nil
			}}, nil
		case "isalnum":
			return &PyBuiltinFunc{Name: "str.isalnum", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				s := str.Value
				if len(s) == 0 {
					return False, nil
				}
				for _, ch := range s {
					if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9')) {
						return False, nil
					}
				}
				return True, nil
			}}, nil
		case "isspace":
			return &PyBuiltinFunc{Name: "str.isspace", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				s := str.Value
				if len(s) == 0 {
					return False, nil
				}
				for _, ch := range s {
					if ch != ' ' && ch != '\t' && ch != '\n' && ch != '\r' && ch != '\f' && ch != '\v' {
						return False, nil
					}
				}
				return True, nil
			}}, nil
		case "isupper":
			return &PyBuiltinFunc{Name: "str.isupper", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				s := str.Value
				hasUpper := false
				for _, ch := range s {
					if ch >= 'a' && ch <= 'z' {
						return False, nil
					}
					if ch >= 'A' && ch <= 'Z' {
						hasUpper = true
					}
				}
				if hasUpper {
					return True, nil
				}
				return False, nil
			}}, nil
		case "islower":
			return &PyBuiltinFunc{Name: "str.islower", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				s := str.Value
				hasLower := false
				for _, ch := range s {
					if ch >= 'A' && ch <= 'Z' {
						return False, nil
					}
					if ch >= 'a' && ch <= 'z' {
						hasLower = true
					}
				}
				if hasLower {
					return True, nil
				}
				return False, nil
			}}, nil
		case "format":
			return &PyBuiltinFunc{Name: "str.format", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				return vm.strFormat(str.Value, args, kwargs)
			}}, nil
		case "splitlines":
			return &PyBuiltinFunc{Name: "str.splitlines", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				keepends := false
				if len(args) > 0 {
					keepends = vm.truthy(args[0])
				}
				s := str.Value
				if len(s) == 0 {
					return &PyList{Items: []Value{}}, nil
				}
				var lines []Value
				start := 0
				for i := 0; i < len(s); i++ {
					if s[i] == '\n' || s[i] == '\r' {
						end := i
						if s[i] == '\r' && i+1 < len(s) && s[i+1] == '\n' {
							i++
						}
						if keepends {
							lines = append(lines, &PyString{Value: s[start : i+1]})
						} else {
							lines = append(lines, &PyString{Value: s[start:end]})
						}
						start = i + 1
					}
				}
				if start < len(s) {
					lines = append(lines, &PyString{Value: s[start:]})
				}
				return &PyList{Items: lines}, nil
			}}, nil
		case "encode":
			return &PyBuiltinFunc{Name: "str.encode", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				return &PyBytes{Value: []byte(str.Value)}, nil
			}}, nil

		case "casefold":
			return &PyBuiltinFunc{Name: "str.casefold", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				// casefold is like lower() but more aggressive for caseless matching
				// Go's strings.ToLower handles most Unicode casefolding
				// For full CPython compatibility we'd need special cases (e.g. ß → ss)
				// but strings.ToLower is a good approximation
				result := []rune{}
				for _, r := range str.Value {
					// Handle special casefold cases
					if r == 'ß' || r == 'ẞ' {
						result = append(result, 's', 's')
					} else {
						result = append(result, []rune(strings.ToLower(string(r)))...)
					}
				}
				return &PyString{Value: string(result)}, nil
			}}, nil

		case "isascii":
			return &PyBuiltinFunc{Name: "str.isascii", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				// Empty string returns True (CPython behavior)
				for _, r := range str.Value {
					if r > 127 {
						return False, nil
					}
				}
				return True, nil
			}}, nil

		case "isdecimal":
			return &PyBuiltinFunc{Name: "str.isdecimal", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(str.Value) == 0 {
					return False, nil
				}
				for _, r := range str.Value {
					// Decimal digits are Unicode category Nd with digit value
					if r < '0' || r > '9' {
						// Check for other Unicode decimal digits
						if !unicode.Is(unicode.Nd, r) {
							return False, nil
						}
					}
				}
				return True, nil
			}}, nil

		case "isnumeric":
			return &PyBuiltinFunc{Name: "str.isnumeric", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(str.Value) == 0 {
					return False, nil
				}
				for _, r := range str.Value {
					// Numeric includes digits, fractions, subscripts, superscripts, etc.
					if !unicode.Is(unicode.Nd, r) && !unicode.Is(unicode.Nl, r) && !unicode.Is(unicode.No, r) {
						return False, nil
					}
				}
				return True, nil
			}}, nil

		case "isidentifier":
			return &PyBuiltinFunc{Name: "str.isidentifier", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(str.Value) == 0 {
					return False, nil
				}
				for i, r := range str.Value {
					if i == 0 {
						if r != '_' && !unicode.IsLetter(r) {
							return False, nil
						}
					} else {
						if r != '_' && !unicode.IsLetter(r) && !unicode.IsDigit(r) {
							return False, nil
						}
					}
				}
				return True, nil
			}}, nil

		case "isprintable":
			return &PyBuiltinFunc{Name: "str.isprintable", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				// Empty string returns True (CPython behavior)
				for _, r := range str.Value {
					if !unicode.IsPrint(r) {
						return False, nil
					}
				}
				return True, nil
			}}, nil

		case "istitle":
			return &PyBuiltinFunc{Name: "str.istitle", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(str.Value) == 0 {
					return False, nil
				}
				// A titlecased string has uppercase after uncased chars
				// and lowercase after cased chars
				prevCased := false
				hasCased := false
				for _, r := range str.Value {
					if unicode.IsUpper(r) || unicode.IsTitle(r) {
						if prevCased {
							return False, nil
						}
						prevCased = true
						hasCased = true
					} else if unicode.IsLower(r) {
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

		case "removeprefix":
			return &PyBuiltinFunc{Name: "str.removeprefix", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("TypeError: removeprefix() takes exactly one argument (%d given)", len(args))
				}
				prefix, ok := args[0].(*PyString)
				if !ok {
					return nil, fmt.Errorf("TypeError: removeprefix arg must be str, not '%s'", vm.typeName(args[0]))
				}
				if strings.HasPrefix(str.Value, prefix.Value) {
					return &PyString{Value: str.Value[len(prefix.Value):]}, nil
				}
				return str, nil
			}}, nil

		case "removesuffix":
			return &PyBuiltinFunc{Name: "str.removesuffix", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("TypeError: removesuffix() takes exactly one argument (%d given)", len(args))
				}
				suffix, ok := args[0].(*PyString)
				if !ok {
					return nil, fmt.Errorf("TypeError: removesuffix arg must be str, not '%s'", vm.typeName(args[0]))
				}
				if suffix.Value != "" && strings.HasSuffix(str.Value, suffix.Value) {
					return &PyString{Value: str.Value[:len(str.Value)-len(suffix.Value)]}, nil
				}
				return str, nil
			}}, nil

		case "format_map":
			return &PyBuiltinFunc{Name: "str.format_map", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("TypeError: format_map() takes exactly one argument (%d given)", len(args))
				}
				mapping, ok := args[0].(*PyDict)
				if !ok {
					return nil, fmt.Errorf("TypeError: format_map() argument must be a mapping, not '%s'", vm.typeName(args[0]))
				}
				// Build format args from the mapping
				template := str.Value
				var result strings.Builder
				i := 0
				for i < len(template) {
					if template[i] == '{' {
						if i+1 < len(template) && template[i+1] == '{' {
							result.WriteByte('{')
							i += 2
							continue
						}
						j := i + 1
						for j < len(template) && template[j] != '}' {
							j++
						}
						if j >= len(template) {
							return nil, fmt.Errorf("ValueError: Single '{' encountered in format string")
						}
						field := template[i+1 : j]
						var fieldName, formatSpec string
						if colonIdx := strings.Index(field, ":"); colonIdx >= 0 {
							fieldName = field[:colonIdx]
							formatSpec = field[colonIdx+1:]
						} else {
							fieldName = field
						}
						val, found := mapping.DictGet(&PyString{Value: fieldName}, vm)
						if !found {
							return nil, fmt.Errorf("KeyError: '%s'", fieldName)
						}
						if formatSpec != "" {
							formatted, err := vm.formatValue(val, formatSpec)
							if err != nil {
								return nil, err
							}
							result.WriteString(formatted)
						} else {
							result.WriteString(vm.str(val))
						}
						i = j + 1
					} else if template[i] == '}' {
						if i+1 < len(template) && template[i+1] == '}' {
							result.WriteByte('}')
							i += 2
							continue
						}
						return nil, fmt.Errorf("ValueError: Single '}' encountered in format string")
					} else {
						result.WriteByte(template[i])
						i++
					}
				}
				return &PyString{Value: result.String()}, nil
			}}, nil

		case "maketrans":
			return &PyBuiltinFunc{Name: "str.maketrans", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				return strMaketransImpl(vm, args)
			}}, nil

		case "translate":
			return &PyBuiltinFunc{Name: "str.translate", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("TypeError: translate() takes exactly one argument (%d given)", len(args))
				}
				table, ok := args[0].(*PyDict)
				if !ok {
					return nil, fmt.Errorf("TypeError: translate() argument must be a dict")
				}
				var result strings.Builder
				for _, r := range str.Value {
					key := MakeInt(int64(r))
					if val, found := table.DictGet(key, vm); found {
						if val == None {
							// Delete the character
							continue
						}
						switch v := val.(type) {
						case *PyString:
							result.WriteString(v.Value)
						case *PyInt:
							result.WriteRune(rune(v.Value))
						default:
							result.WriteRune(r)
						}
					} else {
						result.WriteRune(r)
					}
				}
				return &PyString{Value: result.String()}, nil
			}}, nil
		}
	case *PyBytes:
		b := o
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
		}
		return nil, fmt.Errorf("AttributeError: 'bytes' object has no attribute '%s'", name)
	case *PyFunction:
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
	case *PyBuiltinFunc:
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
	}

	return nil, fmt.Errorf("AttributeError: '%s' object has no attribute '%s'", vm.typeName(obj), name)
}

// intFromBytesImpl implements int.from_bytes(bytes, byteorder, *, signed=False)
func intFromBytesImpl(vm *VM, args []Value, kwargs map[string]Value) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("TypeError: from_bytes() missing required argument: 'bytes'")
	}

	var data []byte
	switch b := args[0].(type) {
	case *PyBytes:
		data = b.Value
	case *PyList:
		data = make([]byte, len(b.Items))
		for i, item := range b.Items {
			data[i] = byte(vm.toInt(item))
		}
	case *PyTuple:
		data = make([]byte, len(b.Items))
		for i, item := range b.Items {
			data[i] = byte(vm.toInt(item))
		}
	default:
		return nil, fmt.Errorf("TypeError: cannot convert '%s' object to bytes", vm.typeName(args[0]))
	}

	byteorder := ""
	if len(args) >= 2 {
		if s, ok := args[1].(*PyString); ok {
			byteorder = s.Value
		} else {
			return nil, fmt.Errorf("TypeError: from_bytes() argument 'byteorder' must be str")
		}
	} else if kv, ok := kwargs["byteorder"]; ok {
		if s, ok := kv.(*PyString); ok {
			byteorder = s.Value
		} else {
			return nil, fmt.Errorf("TypeError: from_bytes() argument 'byteorder' must be str")
		}
	} else {
		return nil, fmt.Errorf("TypeError: from_bytes() missing required argument: 'byteorder'")
	}

	if byteorder != "big" && byteorder != "little" {
		return nil, fmt.Errorf("ValueError: byteorder must be either 'little' or 'big'")
	}

	signed := false
	if kv, ok := kwargs["signed"]; ok {
		signed = vm.Truthy(kv)
	}

	// Make a copy to avoid mutating
	b := make([]byte, len(data))
	copy(b, data)

	if byteorder == "little" {
		// Reverse to big-endian for processing
		for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
			b[i], b[j] = b[j], b[i]
		}
	}

	// Convert bytes to integer (big-endian)
	var result uint64
	if len(b) <= 8 {
		for _, byteVal := range b {
			result = (result << 8) | uint64(byteVal)
		}
		if signed && len(b) > 0 && b[0]&0x80 != 0 {
			// Sign extend
			for i := len(b); i < 8; i++ {
				result |= 0xff << (uint(i) * 8)
			}
			return MakeInt(int64(result)), nil
		}
		return MakeInt(int64(result)), nil
	}

	// For larger values, use big.Int
	bi := new(big.Int).SetBytes(b)
	if signed && len(b) > 0 && b[0]&0x80 != 0 {
		// Subtract 2^(len*8) for negative
		twoN := new(big.Int).Lsh(big.NewInt(1), uint(len(b)*8))
		bi.Sub(bi, twoN)
	}
	return MakeBigInt(bi), nil
}

// formatSpecialFloat returns "infinity" or "NaN" for error messages.
func formatSpecialFloat(f float64) string {
	if math.IsInf(f, 0) {
		return "infinity"
	}
	return "NaN"
}

// floatHex returns the hex representation of a float64, matching Python's float.hex().
func floatHex(f float64) string {
	if math.IsInf(f, 1) {
		return "inf"
	}
	if math.IsInf(f, -1) {
		return "-inf"
	}
	if math.IsNaN(f) {
		return "nan"
	}

	sign := ""
	if math.Signbit(f) {
		sign = "-"
		f = -f
	}

	if f == 0 {
		return sign + "0x0.0000000000000p+0"
	}

	fbits := math.Float64bits(f)
	mantissa := fbits & ((1 << 52) - 1)
	biasedExp := int((fbits >> 52) & 0x7ff)

	if biasedExp == 0 {
		// Subnormal
		return fmt.Sprintf("%s0x0.%013xp-1022", sign, mantissa)
	}
	// Normal
	exp := biasedExp - 1023
	return fmt.Sprintf("%s0x1.%013xp%+d", sign, mantissa, exp)
}

// floatFromHexImpl implements float.fromhex(string).
func floatFromHexImpl(args []Value) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("TypeError: float.fromhex() requires a string argument")
	}
	s, ok := args[0].(*PyString)
	if !ok {
		return nil, fmt.Errorf("TypeError: float.fromhex() argument must be a string")
	}
	str := strings.TrimSpace(s.Value)
	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return nil, fmt.Errorf("ValueError: could not convert string to float: '%s'", s.Value)
	}
	return &PyFloat{Value: f}, nil
}

// floatAsIntegerRatio returns (numerator, denominator) for a finite float.
func floatAsIntegerRatio(f float64) (Value, Value) {
	if f == 0 {
		if math.Signbit(f) {
			return MakeInt(0), MakeInt(1)
		}
		return MakeInt(0), MakeInt(1)
	}

	// Use math.Frexp: f = frac * 2^exp, 0.5 <= |frac| < 1.0
	frac, exp := math.Frexp(f)

	// Multiply frac by 2^53 to get an exact integer mantissa
	// (float64 has 53 bits of significand)
	mantissa := int64(frac * (1 << 53))
	exp -= 53

	// Remove trailing zeros from mantissa to reduce the fraction
	for mantissa != 0 && mantissa%2 == 0 {
		mantissa /= 2
		exp++
	}

	if exp >= 0 {
		// numerator = mantissa * 2^exp, denominator = 1
		if exp <= 62 {
			return MakeInt(mantissa * (1 << uint(exp))), MakeInt(1)
		}
		// Large shift — use big.Int
		num := new(big.Int).SetInt64(mantissa)
		num.Lsh(num, uint(exp))
		return MakeBigInt(num), MakeInt(1)
	}

	// numerator = mantissa, denominator = 2^(-exp)
	negExp := uint(-exp)
	if negExp <= 62 {
		return MakeInt(mantissa), MakeInt(1 << negExp)
	}
	den := new(big.Int).Lsh(big.NewInt(1), negExp)
	return MakeInt(mantissa), MakeBigInt(den)
}

// strMaketransImpl implements str.maketrans(x[, y[, z]]).
func strMaketransImpl(vm *VM, args []Value) (Value, error) {
	result := &PyDict{Items: make(map[Value]Value), buckets: make(map[uint64][]dictEntry)}
	if len(args) == 1 {
		d, ok := args[0].(*PyDict)
		if !ok {
			return nil, fmt.Errorf("TypeError: if you give only one argument to maketrans it must be a dict")
		}
		for _, key := range d.Keys(vm) {
			val, _ := d.DictGet(key, vm)
			var intKey Value
			switch k := key.(type) {
			case *PyString:
				runes := []rune(k.Value)
				if len(runes) != 1 {
					return nil, fmt.Errorf("ValueError: string keys in translate table must be of length 1, found %d", len(runes))
				}
				intKey = MakeInt(int64(runes[0]))
			case *PyInt:
				intKey = k
			default:
				return nil, fmt.Errorf("TypeError: keys in translate table must be strings or integers")
			}
			switch v := val.(type) {
			case *PyString:
				runes := []rune(v.Value)
				if len(runes) == 1 {
					result.DictSet(intKey, MakeInt(int64(runes[0])), vm)
				} else {
					result.DictSet(intKey, val, vm)
				}
			case *PyInt:
				result.DictSet(intKey, val, vm)
			case *PyNone:
				result.DictSet(intKey, None, vm)
			default:
				result.DictSet(intKey, val, vm)
			}
		}
		return result, nil
	}
	if len(args) < 2 {
		return nil, fmt.Errorf("TypeError: maketrans requires 1 or 2-3 string arguments")
	}
	x, ok1 := args[0].(*PyString)
	y, ok2 := args[1].(*PyString)
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("TypeError: maketrans arguments must be strings")
	}
	xRunes := []rune(x.Value)
	yRunes := []rune(y.Value)
	if len(xRunes) != len(yRunes) {
		return nil, fmt.Errorf("ValueError: the first two maketrans arguments must have equal length")
	}
	for i, r := range xRunes {
		result.DictSet(MakeInt(int64(r)), MakeInt(int64(yRunes[i])), vm)
	}
	if len(args) >= 3 {
		z, ok := args[2].(*PyString)
		if !ok {
			return nil, fmt.Errorf("TypeError: maketrans third argument must be a string")
		}
		for _, r := range z.Value {
			result.DictSet(MakeInt(int64(r)), None, vm)
		}
	}
	return result, nil
}

func (vm *VM) setAttr(obj Value, name string, val Value) error {
	switch o := obj.(type) {
	case *PyInstance:
		// Check for user-defined __setattr__ (skip object base class)
		objectClass := vm.builtins["object"].(*PyClass)
		for _, cls := range o.Class.Mro {
			if cls == objectClass {
				break
			}
			if method, ok := cls.Dict["__setattr__"]; ok {
				_, err := vm.call(method, []Value{o, &PyString{Value: name}, val}, nil)
				return err
			}
		}
		// Check for data descriptors (property or custom __set__) in class MRO
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
				// Check for custom data descriptor (instance with __set__)
				if inst, ok := clsVal.(*PyInstance); ok {
					if _, found, err := vm.callDunder(inst, "__set__", o, val); found {
						return err
					}
				}
				break // Found in class dict but not a descriptor, fall through to instance assignment
			}
		}
		// Not a descriptor, set on instance dict or slots
		if o.Slots != nil {
			if !isValidSlot(o.Class, name) {
				return fmt.Errorf("AttributeError: '%s' object has no attribute '%s'", o.Class.Name, name)
			}
			o.Slots[name] = val
		} else {
			o.Dict[name] = val
		}
		return nil
	case *PyClass:
		o.Dict[name] = val
		return nil
	case *PyFunction:
		// Allow setting arbitrary attributes on functions (Python supports func.__dict__)
		switch name {
		case "__name__":
			if s, ok := val.(*PyString); ok {
				o.Name = s.Value
			}
			return nil
		case "__isabstractmethod__":
			o.IsAbstract = vm.truthy(val)
			return nil
		default:
			if o.Dict == nil {
				o.Dict = make(map[string]Value)
			}
			o.Dict[name] = val
			return nil
		}
	case *PyException:
		switch name {
		case "__cause__":
			if val == None || val == nil {
				o.Cause = nil
			} else {
				o.Cause = vm.createException(val, nil)
			}
			// Setting __cause__ always sets __suppress_context__ = True (CPython behavior)
			o.SuppressContext = true
			return nil
		case "__context__":
			if val == None || val == nil {
				o.Context = nil
			} else {
				o.Context = vm.createException(val, nil)
			}
			return nil
		case "__suppress_context__":
			o.SuppressContext = vm.truthy(val)
			return nil
		}
		return fmt.Errorf("'%s' object attribute '%s' is read-only", o.Type(), name)
	}
	return fmt.Errorf("'%s' object attribute '%s' is read-only", vm.typeName(obj), name)
}

func (vm *VM) delAttr(obj Value, name string) error {
	switch o := obj.(type) {
	case *PyInstance:
		// Check for user-defined __delattr__ (skip object base class)
		objectClass := vm.builtins["object"].(*PyClass)
		for _, cls := range o.Class.Mro {
			if cls == objectClass {
				break
			}
			if method, ok := cls.Dict["__delattr__"]; ok {
				_, err := vm.call(method, []Value{o, &PyString{Value: name}}, nil)
				return err
			}
		}
		// Check for property with deleter or custom descriptor __delete__ in class MRO
		for _, cls := range o.Class.Mro {
			if clsVal, ok := cls.Dict[name]; ok {
				if prop, ok := clsVal.(*PyProperty); ok {
					if prop.Fdel == nil {
						return fmt.Errorf("property '%s' has no deleter", name)
					}
					_, err := vm.call(prop.Fdel, []Value{o}, nil)
					return err
				}
				// Check for custom descriptor with __delete__
				if inst, ok := clsVal.(*PyInstance); ok {
					if _, found, err := vm.callDunder(inst, "__delete__", o); found {
						return err
					}
				}
				break
			}
		}
		if o.Slots != nil {
			val, exists := o.Slots[name]
			if !exists {
				return fmt.Errorf("AttributeError: '%s' object has no attribute '%s'", o.Class.Name, name)
			}
			vm.callDel(val)
			delete(o.Slots, name)
		} else {
			val, exists := o.Dict[name]
			if !exists {
				return fmt.Errorf("AttributeError: '%s' object has no attribute '%s'", o.Class.Name, name)
			}
			vm.callDel(val)
			delete(o.Dict, name)
		}
		return nil
	case *PyModule:
		if _, exists := o.Dict[name]; !exists {
			return fmt.Errorf("AttributeError: module '%s' has no attribute '%s'", o.Name, name)
		}
		delete(o.Dict, name)
		return nil
	case *PyClass:
		if _, exists := o.Dict[name]; !exists {
			return fmt.Errorf("AttributeError: type object '%s' has no attribute '%s'", o.Name, name)
		}
		delete(o.Dict, name)
		return nil
	case *PyDict:
		key := &PyString{Value: name}
		if o.DictDelete(key, vm) {
			return nil
		}
		return fmt.Errorf("AttributeError: 'dict' object has no attribute '%s'", name)
	}
	return fmt.Errorf("AttributeError: '%s' object has no attribute '%s'", vm.typeName(obj), name)
}

// strFormat implements Python's str.format() method
func (vm *VM) strFormat(template string, args []Value, kwargs map[string]Value) (Value, error) {
	var result strings.Builder
	argIdx := 0

	i := 0
	for i < len(template) {
		if template[i] == '{' {
			if i+1 < len(template) && template[i+1] == '{' {
				result.WriteByte('{')
				i += 2
				continue
			}
			// Find closing brace
			j := i + 1
			for j < len(template) && template[j] != '}' {
				j++
			}
			if j >= len(template) {
				return nil, fmt.Errorf("ValueError: Single '{' encountered in format string")
			}
			field := template[i+1 : j]

			// Parse field name and format spec
			var fieldName, formatSpec string
			if colonIdx := strings.Index(field, ":"); colonIdx >= 0 {
				fieldName = field[:colonIdx]
				formatSpec = field[colonIdx+1:]
			} else {
				fieldName = field
			}

			// Get value
			var val Value
			if fieldName == "" {
				if argIdx >= len(args) {
					return nil, fmt.Errorf("IndexError: Replacement index %d out of range", argIdx)
				}
				val = args[argIdx]
				argIdx++
			} else if idx, err := strconv.Atoi(fieldName); err == nil {
				if idx >= len(args) {
					return nil, fmt.Errorf("IndexError: Replacement index %d out of range", idx)
				}
				val = args[idx]
			} else {
				if v, ok := kwargs[fieldName]; ok {
					val = v
				} else {
					return nil, fmt.Errorf("KeyError: '%s'", fieldName)
				}
			}

			// Apply format spec
			formatted, fmtErr := vm.formatValue(val, formatSpec)
			if fmtErr != nil {
				return nil, fmtErr
			}
			result.WriteString(formatted)
			i = j + 1
		} else if template[i] == '}' {
			if i+1 < len(template) && template[i+1] == '}' {
				result.WriteByte('}')
				i += 2
				continue
			}
			return nil, fmt.Errorf("ValueError: Single '}' encountered in format string")
		} else {
			result.WriteByte(template[i])
			i++
		}
	}
	return &PyString{Value: result.String()}, nil
}

// formatValue formats a value with the given format spec, checking __format__ first.
func (vm *VM) formatValue(val Value, spec string) (string, error) {
	if inst, ok := val.(*PyInstance); ok {
		if result, found, err := vm.callDunder(inst, "__format__", &PyString{Value: spec}); found {
			if err != nil {
				return "", err
			}
			if s, ok := result.(*PyString); ok {
				return s.Value, nil
			}
			return "", fmt.Errorf("TypeError: __format__ must return a str, not %s", vm.typeName(result))
		}
	}
	return vm.applyFormatSpec(val, spec), nil
}

// applyFormatSpec applies a format spec like ">10", "<10", "^10", ".2f", "05d"
func (vm *VM) applyFormatSpec(val Value, spec string) string {
	if len(spec) == 0 {
		return vm.str(val)
	}

	// Parse alignment and fill
	fill := " "
	align := byte(0)
	i := 0

	// Check for fill+align or just align
	if len(spec) > 1 && (spec[1] == '<' || spec[1] == '>' || spec[1] == '^') {
		fill = string(spec[0])
		align = spec[1]
		i = 2
	} else if len(spec) > 0 && (spec[0] == '<' || spec[0] == '>' || spec[0] == '^') {
		align = spec[0]
		i = 1
	}

	// Check for sign
	sign := byte(0)
	if i < len(spec) && (spec[i] == '+' || spec[i] == '-' || spec[i] == ' ') {
		sign = spec[i]
		i++
	}

	// Check for zero-fill
	zeroFill := false
	if i < len(spec) && spec[i] == '0' {
		zeroFill = true
		fill = "0"
		if align == 0 {
			align = '>'
		}
		i++
	}

	// Parse width
	width := 0
	for i < len(spec) && spec[i] >= '0' && spec[i] <= '9' {
		width = width*10 + int(spec[i]-'0')
		i++
	}

	// Parse precision
	precision := -1
	if i < len(spec) && spec[i] == '.' {
		i++
		precision = 0
		for i < len(spec) && spec[i] >= '0' && spec[i] <= '9' {
			precision = precision*10 + int(spec[i]-'0')
			i++
		}
	}

	// Parse type
	typeChar := byte(0)
	if i < len(spec) {
		typeChar = spec[i]
	}

	// Format the value
	var s string
	switch typeChar {
	case 'f', 'F':
		f := vm.toFloat(val)
		if precision < 0 {
			precision = 6
		}
		s = strconv.FormatFloat(f, 'f', precision, 64)
	case 'd':
		n := vm.toInt(val)
		s = strconv.FormatInt(n, 10)
	case 'x':
		n := vm.toInt(val)
		s = strconv.FormatInt(n, 16)
	case 'X':
		n := vm.toInt(val)
		s = strings.ToUpper(strconv.FormatInt(n, 16))
	case 'o':
		n := vm.toInt(val)
		s = strconv.FormatInt(n, 8)
	case 'b':
		n := vm.toInt(val)
		s = strconv.FormatInt(n, 2)
	case 'e', 'E':
		f := vm.toFloat(val)
		if precision < 0 {
			precision = 6
		}
		s = strconv.FormatFloat(f, byte(typeChar), precision, 64)
	case 'g', 'G':
		f := vm.toFloat(val)
		if precision < 0 {
			precision = 6
		}
		s = strconv.FormatFloat(f, byte(typeChar), precision, 64)
	case 's', 0:
		s = vm.str(val)
		if precision >= 0 && len(s) > precision {
			s = s[:precision]
		}
	default:
		s = vm.str(val)
	}

	// Apply sign for numeric types
	if sign == '+' && len(s) > 0 && s[0] != '-' {
		s = "+" + s
	} else if sign == ' ' && len(s) > 0 && s[0] != '-' {
		s = " " + s
	}

	// Apply zero-fill for numeric types
	if zeroFill && width > 0 {
		prefix := ""
		if len(s) > 0 && (s[0] == '-' || s[0] == '+' || s[0] == ' ') {
			prefix = string(s[0])
			s = s[1:]
		}
		for len(s) < width-len(prefix) {
			s = "0" + s
		}
		s = prefix + s
		return s
	}

	// Apply width and alignment
	if width > utf8.RuneCountInString(s) {
		padding := width - utf8.RuneCountInString(s)
		switch align {
		case '<':
			s = s + strings.Repeat(fill, padding)
		case '>':
			s = strings.Repeat(fill, padding) + s
		case '^':
			left := padding / 2
			right := padding - left
			s = strings.Repeat(fill, left) + s + strings.Repeat(fill, right)
		default:
			// Default: right-align for numbers, left-align for strings
			switch val.(type) {
			case *PyInt, *PyFloat:
				s = strings.Repeat(fill, padding) + s
			default:
				s = s + strings.Repeat(fill, padding)
			}
		}
	}

	return s
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
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

// Ensure unused imports are satisfied
var _ = sort.SliceStable
var _ = strconv.Atoi
var _ = utf8.RuneCountInString
