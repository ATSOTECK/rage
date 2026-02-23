package runtime

import (
	"fmt"
	"strings"
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
// This is the logic behind object.__getattribute__: data descriptors -> instance dict -> class MRO.
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
		return vm.getAttrGenerator(o, name)
	case *PyCoroutine:
		return vm.getAttrCoroutine(o, name)
	case *PyComplex:
		return vm.getAttrComplex(o, name)
	case *PyException:
		return vm.getAttrException(o, name)
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
		return vm.getAttrProperty(o, name)
	case *PySuper:
		return vm.getAttrSuper(o, name)
	case *PyInstance:
		return vm.getAttrInstance(o, name)
	case *PyClass:
		return vm.getAttrClass(o, name)
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
	case *UnionType:
		switch name {
		case "__args__":
			return &PyTuple{Items: o.Args}, nil
		case "__class__":
			if cls, ok := vm.builtins["type"]; ok {
				return cls, nil
			}
			return &PyBuiltinFunc{Name: "types.UnionType"}, nil
		}
		return nil, fmt.Errorf("AttributeError: 'types.UnionType' object has no attribute '%s'", name)
	case *PyDict:
		return vm.getAttrDict(o, name)
	case *PyFrozenSet:
		return vm.getAttrFrozenSet(o, name)
	case *PySet:
		return vm.getAttrSet(o, name)
	case *PyTuple:
		return vm.getAttrTuple(o, name)
	case *PyInt:
		return vm.getAttrInt(o, name)
	case *PyFloat:
		return vm.getAttrFloat(o, name)
	case *PyList:
		return vm.getAttrList(o, name)
	case *PyString:
		return vm.getAttrString(o, name)
	case *PyBytes:
		return vm.getAttrBytes(o, name)
	case *PyRange:
		return vm.getAttrRange(o, name)
	case *PyFunction:
		return vm.getAttrFunction(o, name)
	case *PyBuiltinFunc:
		return vm.getAttrBuiltinFunc(o, name)
	}

	return nil, fmt.Errorf("AttributeError: '%s' object has no attribute '%s'", vm.typeName(obj), name)
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
