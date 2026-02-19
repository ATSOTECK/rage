package stdlib

import (
	"fmt"

	"github.com/ATSOTECK/rage/internal/runtime"
)

// getInstanceClass returns the PyClass of a value, or nil if not a PyInstance.
func getInstanceClass(v runtime.Value) *runtime.PyClass {
	if inst, ok := v.(*runtime.PyInstance); ok {
		return inst.Class
	}
	return nil
}

// getValueClassOrBuiltin returns the class representation of a value for ABC checks.
// For PyInstance, returns its Class. For builtin types (int, str, etc.), returns the
// corresponding builtin type value (which may be *PyBuiltinFunc or *PyClass).
func getValueClassOrBuiltin(vm *runtime.VM, v runtime.Value) runtime.Value {
	if inst, ok := v.(*runtime.PyInstance); ok {
		return inst.Class
	}
	// Map builtin values to their type constructors
	var typeName string
	switch v.(type) {
	case *runtime.PyBool:
		typeName = "bool"
	case *runtime.PyInt:
		typeName = "int"
	case *runtime.PyFloat:
		typeName = "float"
	case *runtime.PyString:
		typeName = "str"
	case *runtime.PyList:
		typeName = "list"
	case *runtime.PyTuple:
		typeName = "tuple"
	case *runtime.PyDict:
		typeName = "dict"
	case *runtime.PySet:
		typeName = "set"
	case *runtime.PyFrozenSet:
		typeName = "frozenset"
	case *runtime.PyBytes:
		typeName = "bytes"
	case *runtime.PyNone:
		typeName = "NoneType"
	case *runtime.PyComplex:
		typeName = "complex"
	}
	if typeName != "" {
		return vm.GetBuiltin(typeName)
	}
	return nil
}

// callSubclassHookValue searches cls's MRO for __subclasshook__ and calls it with subclass.
// subclass can be *PyClass or *PyBuiltinFunc (for builtin types).
// Returns True/False if the hook gave a definitive answer, or nil if it returned NotImplemented
// or was not found (meaning the caller should fall through to normal behavior).
func callSubclassHookValue(vm *runtime.VM, cls *runtime.PyClass, subclass runtime.Value) runtime.Value {
	for _, mroCls := range cls.Mro {
		if hook, ok := mroCls.Dict["__subclasshook__"]; ok {
			// __subclasshook__ is a classmethod; call with cls as first arg
			var result runtime.Value
			var err error
			switch fn := hook.(type) {
			case *runtime.PyClassMethod:
				// Unwrap the classmethod
				switch inner := fn.Func.(type) {
				case *runtime.PyFunction:
					result, err = vm.CallFunction(inner, []runtime.Value{cls, subclass}, nil)
				case *runtime.PyBuiltinFunc:
					result, err = inner.Fn([]runtime.Value{cls, subclass}, nil)
				}
			case *runtime.PyFunction:
				result, err = vm.CallFunction(fn, []runtime.Value{cls, subclass}, nil)
			case *runtime.PyBuiltinFunc:
				result, err = fn.Fn([]runtime.Value{cls, subclass}, nil)
			}
			if err != nil {
				return nil // treat errors as NotImplemented
			}
			// Check for NotImplemented
			if _, isNotImpl := result.(*runtime.PyNotImplementedType); isNotImpl {
				return nil // fall through to normal check
			}
			if result == nil {
				return nil
			}
			if vm.Truthy(result) {
				return runtime.True
			}
			return runtime.False
		}
	}
	return nil
}

// callSubclassHook is a convenience wrapper for callSubclassHookValue with *PyClass args.
func callSubclassHook(vm *runtime.VM, cls *runtime.PyClass, subclass *runtime.PyClass) runtime.Value {
	return callSubclassHookValue(vm, cls, subclass)
}

// makeRegisterMethod creates a register() builtin for an ABC class.
// It appends the argument to the class's RegisteredSubclasses and returns it (for use as a decorator).
func makeRegisterMethod(owner *runtime.PyClass) *runtime.PyBuiltinFunc {
	return &runtime.PyBuiltinFunc{
		Name: "register",
		Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("TypeError: register() takes exactly 1 argument (%d given)", len(args))
			}
			subcls, ok := args[0].(*runtime.PyClass)
			if !ok {
				return nil, fmt.Errorf("TypeError: register() argument must be a class")
			}
			// Avoid duplicate registrations
			for _, existing := range owner.RegisteredSubclasses {
				if existing == subcls {
					return subcls, nil
				}
			}
			owner.RegisteredSubclasses = append(owner.RegisteredSubclasses, subcls)
			return subcls, nil
		},
	}
}

// InitAbcModule registers the abc (Abstract Base Classes) module.
func InitAbcModule() {
	runtime.RegisterModule("abc", func(vm *runtime.VM) *runtime.PyModule {
		objectClass := vm.GetBuiltin("object").(*runtime.PyClass)
		typeClass := vm.GetBuiltin("type").(*runtime.PyClass)

		// Create ABCMeta class inheriting from type (the metaclass)
		abcMetaClass := &runtime.PyClass{
			Name:  "ABCMeta",
			Bases: []*runtime.PyClass{typeClass},
			Dict:  make(map[string]runtime.Value),
			IsABC: true,
		}
		abcMetaClass.Mro, _ = vm.ComputeC3MRO(abcMetaClass, abcMetaClass.Bases)
		abcMetaClass.Dict["register"] = makeRegisterMethod(abcMetaClass)

		// __instancecheck__ - called by isinstance() when metaclass is ABCMeta
		abcMetaClass.Dict["__instancecheck__"] = &runtime.PyBuiltinFunc{
			Name: "__instancecheck__",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				if len(args) < 2 {
					return nil, fmt.Errorf("TypeError: __instancecheck__ requires 2 arguments")
				}
				cls, ok := args[0].(*runtime.PyClass)
				if !ok {
					return nil, fmt.Errorf("TypeError: __instancecheck__ first arg must be a class")
				}
				instance := args[1]

				// Try __subclasshook__ on cls MRO
				// Support both PyInstance and builtin types
				classVal := getValueClassOrBuiltin(vm, instance)
				if classVal != nil {
					if instClass, ok := classVal.(*runtime.PyClass); ok {
						result := callSubclassHook(vm, cls, instClass)
						if result != nil {
							return result, nil
						}
					} else {
						// Builtin type represented as *PyBuiltinFunc
						result := callSubclassHookValue(vm, cls, classVal)
						if result != nil {
							return result, nil
						}
					}
				}

				// Fall back to normal isinstance check
				if inst, ok := instance.(*runtime.PyInstance); ok {
					if vm.IsInstanceOf(inst, cls) {
						return runtime.True, nil
					}
					// Check registered virtual subclasses on cls and its MRO
					for _, mroCls := range cls.Mro {
						for _, reg := range mroCls.RegisteredSubclasses {
							if vm.IsInstanceOf(inst, reg) {
								return runtime.True, nil
							}
						}
					}
				}
				return runtime.False, nil
			},
		}

		// __subclasscheck__ - called by issubclass() when metaclass is ABCMeta
		abcMetaClass.Dict["__subclasscheck__"] = &runtime.PyBuiltinFunc{
			Name: "__subclasscheck__",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				if len(args) < 2 {
					return nil, fmt.Errorf("TypeError: __subclasscheck__ requires 2 arguments")
				}
				cls, ok := args[0].(*runtime.PyClass)
				if !ok {
					return nil, fmt.Errorf("TypeError: __subclasscheck__ first arg must be a class")
				}
				subclass := args[1]

				// Try __subclasshook__ on cls MRO
				if subPyClass, ok := subclass.(*runtime.PyClass); ok {
					result := callSubclassHook(vm, cls, subPyClass)
					if result != nil {
						return result, nil
					}
				} else {
					// Builtin type represented as *PyBuiltinFunc
					result := callSubclassHookValue(vm, cls, subclass)
					if result != nil {
						return result, nil
					}
				}

				// Fall back to normal MRO-based subclass check (only works with *PyClass)
				if subPyClass, ok := subclass.(*runtime.PyClass); ok {
					for _, mroClass := range subPyClass.Mro {
						if mroClass == cls {
							return runtime.True, nil
						}
					}
				}
				// Check registered virtual subclasses on cls and its MRO
				if subPyClass, ok := subclass.(*runtime.PyClass); ok {
					for _, mroCls := range cls.Mro {
						for _, reg := range mroCls.RegisteredSubclasses {
							for _, mroClass := range subPyClass.Mro {
								if mroClass == reg {
									return runtime.True, nil
								}
							}
						}
					}
				}
				return runtime.False, nil
			},
		}

		// Create ABC class (convenience base class)
		abcClass := &runtime.PyClass{
			Name:      "ABC",
			Bases:     []*runtime.PyClass{objectClass},
			Dict:      make(map[string]runtime.Value),
			IsABC:     true,
			Metaclass: abcMetaClass,
		}
		abcClass.Mro, _ = vm.ComputeC3MRO(abcClass, abcClass.Bases)
		abcClass.Dict["register"] = makeRegisterMethod(abcClass)

		// abstractmethod decorator function
		abstractMethodFn := &runtime.PyBuiltinFunc{
			Name: "abstractmethod",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("TypeError: abstractmethod requires exactly 1 argument, got %d", len(args))
				}
				switch fn := args[0].(type) {
				case *runtime.PyFunction:
					fn.IsAbstract = true
					return fn, nil
				case *runtime.PyProperty:
					// Mark the underlying fget as abstract
					if fget, ok := fn.Fget.(*runtime.PyFunction); ok {
						fget.IsAbstract = true
					}
					return fn, nil
				case *runtime.PyClassMethod:
					if inner, ok := fn.Func.(*runtime.PyFunction); ok {
						inner.IsAbstract = true
					}
					return fn, nil
				case *runtime.PyStaticMethod:
					if inner, ok := fn.Func.(*runtime.PyFunction); ok {
						inner.IsAbstract = true
					}
					return fn, nil
				default:
					return args[0], nil
				}
			},
		}

		// abstractclassmethod (deprecated) - wraps in classmethod and marks abstract
		abstractClassMethodFn := &runtime.PyBuiltinFunc{
			Name: "abstractclassmethod",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("TypeError: abstractclassmethod requires exactly 1 argument, got %d", len(args))
				}
				cm := &runtime.PyClassMethod{Func: args[0]}
				if fn, ok := args[0].(*runtime.PyFunction); ok {
					fn.IsAbstract = true
				}
				return cm, nil
			},
		}

		// abstractstaticmethod (deprecated) - wraps in staticmethod and marks abstract
		abstractStaticMethodFn := &runtime.PyBuiltinFunc{
			Name: "abstractstaticmethod",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("TypeError: abstractstaticmethod requires exactly 1 argument, got %d", len(args))
				}
				sm := &runtime.PyStaticMethod{Func: args[0]}
				if fn, ok := args[0].(*runtime.PyFunction); ok {
					fn.IsAbstract = true
				}
				return sm, nil
			},
		}

		// abstractproperty (deprecated) - creates property and marks abstract
		abstractPropertyFn := &runtime.PyBuiltinFunc{
			Name: "abstractproperty",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				if len(args) < 1 {
					return nil, fmt.Errorf("TypeError: abstractproperty requires at least 1 argument")
				}
				prop := &runtime.PyProperty{Fget: args[0]}
				if len(args) > 1 {
					prop.Fset = args[1]
				}
				if fn, ok := args[0].(*runtime.PyFunction); ok {
					fn.IsAbstract = true
				}
				return prop, nil
			},
		}

		mod := runtime.NewModule("abc")
		mod.Dict["ABCMeta"] = abcMetaClass
		mod.Dict["ABC"] = abcClass
		mod.Dict["abstractmethod"] = abstractMethodFn
		mod.Dict["abstractclassmethod"] = abstractClassMethodFn
		mod.Dict["abstractstaticmethod"] = abstractStaticMethodFn
		mod.Dict["abstractproperty"] = abstractPropertyFn
		return mod
	})
}
