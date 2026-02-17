package stdlib

import (
	"fmt"

	"github.com/ATSOTECK/rage/internal/runtime"
)

// InitOperatorModule registers the operator module.
func InitOperatorModule() {
	runtime.RegisterModule("operator", func(vm *runtime.VM) *runtime.PyModule {
		mod := runtime.NewModule("operator")

		mod.Dict["length_hint"] = &runtime.PyBuiltinFunc{
			Name: "length_hint",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				if len(args) < 1 || len(args) > 2 {
					return nil, fmt.Errorf("TypeError: length_hint() requires 1 to 2 arguments (%d given)", len(args))
				}

				obj := args[0]
				defaultVal := int64(0)
				if len(args) == 2 {
					if d, ok := args[1].(*runtime.PyInt); ok {
						defaultVal = d.Value
					} else {
						return nil, fmt.Errorf("TypeError: '%s' object cannot be interpreted as an integer", vm.TypeNameOf(args[1]))
					}
				}
				if v, ok := kwargs["default"]; ok {
					if d, okd := v.(*runtime.PyInt); okd {
						defaultVal = d.Value
					} else {
						return nil, fmt.Errorf("TypeError: '%s' object cannot be interpreted as an integer", vm.TypeNameOf(v))
					}
				}

				// Try __len__ first
				if inst, ok := obj.(*runtime.PyInstance); ok {
					if result, found, err := vm.CallDunderWithError(inst, "__len__"); found {
						if err != nil {
							return nil, err
						}
						if i, ok := result.(*runtime.PyInt); ok {
							return i, nil
						}
					}
					// Try __length_hint__
					if result, found, err := vm.CallDunderWithError(inst, "__length_hint__"); found {
						if err != nil {
							return nil, err
						}
						if i, ok := result.(*runtime.PyInt); ok {
							if i.Value < 0 {
								return nil, fmt.Errorf("ValueError: __length_hint__() should return >= 0")
							}
							return i, nil
						}
						return nil, fmt.Errorf("TypeError: __length_hint__ must be an integer, not %s", vm.TypeNameOf(result))
					}
					// Neither found, return default
					return runtime.MakeInt(defaultVal), nil
				}

				// For built-in types, try len
				switch v := obj.(type) {
				case *runtime.PyList:
					return runtime.MakeInt(int64(len(v.Items))), nil
				case *runtime.PyTuple:
					return runtime.MakeInt(int64(len(v.Items))), nil
				case *runtime.PyDict:
					return runtime.MakeInt(int64(len(v.Items))), nil
				case *runtime.PyString:
					return runtime.MakeInt(int64(len(v.Value))), nil
				case *runtime.PySet:
					return runtime.MakeInt(int64(len(v.Items))), nil
				case *runtime.PyFrozenSet:
					return runtime.MakeInt(int64(len(v.Items))), nil
				case *runtime.PyBytes:
					return runtime.MakeInt(int64(len(v.Value))), nil
				default:
					return runtime.MakeInt(defaultVal), nil
				}
			},
		}

		// operator.index(a) - call __index__ on a
		mod.Dict["index"] = &runtime.PyBuiltinFunc{
			Name: "index",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("TypeError: index() takes exactly one argument (%d given)", len(args))
				}
				i, err := vm.GetIntIndex(args[0])
				if err != nil {
					return nil, err
				}
				return runtime.MakeInt(i), nil
			},
		}

		return mod
	})
}
