package stdlib

import (
	"fmt"

	"github.com/ATSOTECK/rage/internal/runtime"
)

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

		// Create ABC class (convenience base class)
		abcClass := &runtime.PyClass{
			Name:  "ABC",
			Bases: []*runtime.PyClass{objectClass},
			Dict:  make(map[string]runtime.Value),
			IsABC: true,
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
