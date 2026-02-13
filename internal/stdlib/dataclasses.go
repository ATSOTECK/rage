package stdlib

import (
	"fmt"
	"strings"

	"github.com/ATSOTECK/rage/internal/runtime"
)

// missingSentinel is the MISSING sentinel value for dataclasses
type missingSentinel struct{}

func (m *missingSentinel) Type() string   { return "dataclasses.MISSING" }
func (m *missingSentinel) String() string { return "MISSING" }

var dcMissing = &missingSentinel{}

// DataclassField represents a field in a dataclass
type DataclassField struct {
	Name           string
	FieldType      runtime.Value // The annotation type
	Default        runtime.Value // dcMissing sentinel if not set
	DefaultFactory runtime.Value // dcMissing sentinel if not set
	Init           bool
	Repr           bool
	Compare        bool
	Hash           *bool // nil = follow eq; true/false = explicit
	KwOnly         bool
}

// InitDataclassesModule registers the dataclasses module.
func InitDataclassesModule() {
	runtime.RegisterModule("dataclasses", func(vm *runtime.VM) *runtime.PyModule {
		mod := runtime.NewModule("dataclasses")

		mod.Dict["MISSING"] = dcMissing

		mod.Dict["FrozenInstanceError"] = &runtime.PyBuiltinFunc{
			Name: "FrozenInstanceError",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				msg := "cannot assign to field"
				if len(args) > 0 {
					msg = fmt.Sprintf("%v", args[0])
				}
				return nil, &runtime.PyException{
					TypeName: "FrozenInstanceError",
					Message:  msg,
				}
			},
		}

		// dataclass decorator
		mod.Dict["dataclass"] = &runtime.PyBuiltinFunc{
			Name: "dataclass",
			Fn:   makeDataclassDecoratorFn(vm),
		}

		// field() function
		mod.Dict["field"] = &runtime.PyBuiltinFunc{
			Name: "field",
			Fn:   makeFieldFn(),
		}

		// fields() function
		mod.Dict["fields"] = &runtime.PyBuiltinFunc{
			Name: "fields",
			Fn:   makeFieldsFn(vm),
		}

		// asdict() function
		mod.Dict["asdict"] = &runtime.PyBuiltinFunc{
			Name: "asdict",
			Fn:   makeAsDictFn(vm),
		}

		// astuple() function
		mod.Dict["astuple"] = &runtime.PyBuiltinFunc{
			Name: "astuple",
			Fn:   makeAsTupleFn(vm),
		}

		return mod
	})
}

// makeFieldFn creates the field() builtin function
func makeFieldFn() func([]runtime.Value, map[string]runtime.Value) (runtime.Value, error) {
	return func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
		f := &DataclassField{
			Default:        dcMissing,
			DefaultFactory: dcMissing,
			Init:           true,
			Repr:           true,
			Compare:        true,
		}

		if kwargs != nil {
			if v, ok := kwargs["default"]; ok {
				f.Default = v
			}
			if v, ok := kwargs["default_factory"]; ok {
				f.DefaultFactory = v
			}
			if v, ok := kwargs["init"]; ok {
				f.Init = runtime.IsTrue(v)
			}
			if v, ok := kwargs["repr"]; ok {
				f.Repr = runtime.IsTrue(v)
			}
			if v, ok := kwargs["compare"]; ok {
				f.Compare = runtime.IsTrue(v)
			}
			if v, ok := kwargs["hash"]; ok {
				if _, isNone := v.(*runtime.PyNone); isNone {
					f.Hash = nil
				} else {
					b := runtime.IsTrue(v)
					f.Hash = &b
				}
			}
			if v, ok := kwargs["kw_only"]; ok {
				f.KwOnly = runtime.IsTrue(v)
			}
		}

		// Validate: can't specify both default and default_factory
		if f.Default != dcMissing && f.DefaultFactory != dcMissing {
			return nil, fmt.Errorf("ValueError: cannot specify both default and default_factory")
		}

		return runtime.NewUserData(f), nil
	}
}

// makeDataclassDecoratorFn creates the @dataclass decorator function
func makeDataclassDecoratorFn(vm *runtime.VM) func([]runtime.Value, map[string]runtime.Value) (runtime.Value, error) {
	return func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
		opts := &dataclassOptions{
			init:   true,
			repr:   true,
			eq:     true,
			order:  false,
			frozen: false,
			kwOnly: false,
		}

		// @dataclass applied directly to a class: @dataclass class Foo: ...
		if len(args) == 1 {
			if cls, ok := args[0].(*runtime.PyClass); ok {
				if kwargs != nil {
					parseDataclassOptions(opts, kwargs)
				}
				return processDataclass(vm, cls, opts)
			}
		}

		// @dataclass() or @dataclass(init=False, ...) — return a decorator
		if kwargs != nil {
			parseDataclassOptions(opts, kwargs)
		}

		decorator := &runtime.PyBuiltinFunc{
			Name: "dataclass",
			Fn: func(innerArgs []runtime.Value, innerKwargs map[string]runtime.Value) (runtime.Value, error) {
				if len(innerArgs) != 1 {
					return nil, fmt.Errorf("TypeError: dataclass decorator expects a class argument")
				}
				cls, ok := innerArgs[0].(*runtime.PyClass)
				if !ok {
					return nil, fmt.Errorf("TypeError: dataclass decorator expects a class, got %T", innerArgs[0])
				}
				return processDataclass(vm, cls, opts)
			},
		}

		return decorator, nil
	}
}

type dataclassOptions struct {
	init   bool
	repr   bool
	eq     bool
	order  bool
	frozen bool
	kwOnly bool
}

func parseDataclassOptions(opts *dataclassOptions, kwargs map[string]runtime.Value) {
	if v, ok := kwargs["init"]; ok {
		opts.init = runtime.IsTrue(v)
	}
	if v, ok := kwargs["repr"]; ok {
		opts.repr = runtime.IsTrue(v)
	}
	if v, ok := kwargs["eq"]; ok {
		opts.eq = runtime.IsTrue(v)
	}
	if v, ok := kwargs["order"]; ok {
		opts.order = runtime.IsTrue(v)
	}
	if v, ok := kwargs["frozen"]; ok {
		opts.frozen = runtime.IsTrue(v)
	}
	if v, ok := kwargs["kw_only"]; ok {
		opts.kwOnly = runtime.IsTrue(v)
	}
}

// processDataclass processes a class and turns it into a dataclass
func processDataclass(vm *runtime.VM, cls *runtime.PyClass, opts *dataclassOptions) (runtime.Value, error) {
	allFields := collectFields(vm, cls, opts)

	// Store __dataclass_fields__ on the class
	fieldsDict := &runtime.PyDict{Items: make(map[runtime.Value]runtime.Value)}
	for _, f := range allFields {
		ud := runtime.NewUserData(f)
		fieldsDict.DictSet(runtime.NewString(f.Name), ud, vm)
	}
	cls.Dict["__dataclass_fields__"] = fieldsDict

	// Store __dataclass_params__ to mark this as a dataclass
	cls.Dict["__dataclass_params__"] = runtime.NewUserData(opts)

	// Generate __init__
	if opts.init {
		if !hasOwnMethod(cls, "__init__") {
			cls.Dict["__init__"] = makeInit(vm, cls, allFields, opts.frozen)
		}
	}

	// Generate __repr__
	if opts.repr {
		if !hasOwnMethod(cls, "__repr__") {
			cls.Dict["__repr__"] = makeRepr(vm, cls, allFields)
		}
	}

	// Generate __eq__
	if opts.eq {
		if !hasOwnMethod(cls, "__eq__") {
			cls.Dict["__eq__"] = makeEq(vm, cls, allFields)
		}
	}

	// Generate ordering methods
	if opts.order {
		cls.Dict["__lt__"] = makeOrder(vm, cls, allFields, "lt")
		cls.Dict["__le__"] = makeOrder(vm, cls, allFields, "le")
		cls.Dict["__gt__"] = makeOrder(vm, cls, allFields, "gt")
		cls.Dict["__ge__"] = makeOrder(vm, cls, allFields, "ge")
	}

	// Handle frozen and hash
	if opts.frozen {
		cls.Dict["__setattr__"] = makeFrozenSetattr()
		cls.Dict["__delattr__"] = makeFrozenDelattr()
		if opts.eq {
			cls.Dict["__hash__"] = makeFrozenHash(vm, allFields)
		}
	} else if opts.eq {
		// eq=True and frozen=False -> unhashable
		cls.Dict["__hash__"] = runtime.None
	}

	return cls, nil
}

// hasOwnMethod checks if a method is defined directly on the class (not just inherited)
func hasOwnMethod(cls *runtime.PyClass, name string) bool {
	val, ok := cls.Dict[name]
	if !ok {
		return false
	}
	// Check if the value is the same object as in a base class (inherited)
	for _, base := range cls.Bases {
		if baseVal, ok := base.Dict[name]; ok {
			if val == baseVal {
				return false // Same pointer = inherited, not overridden
			}
		}
	}
	return true
}

// collectFields gathers fields from base classes and the current class
func collectFields(vm *runtime.VM, cls *runtime.PyClass, opts *dataclassOptions) []*DataclassField {
	var fields []*DataclassField
	seen := make(map[string]bool)

	// Collect from base classes in MRO order (reversed, skip self and object)
	for i := len(cls.Mro) - 1; i >= 1; i-- {
		base := cls.Mro[i]
		if baseFields, ok := base.Dict["__dataclass_fields__"]; ok {
			if baseDict, ok := baseFields.(*runtime.PyDict); ok {
				for _, key := range baseDict.Keys(vm) {
					keyStr := pyStringValue(key)
					if keyStr == "" {
						continue
					}
					val, _ := baseDict.DictGet(key, vm)
					if ud, ok := val.(*runtime.PyUserData); ok {
						if f, ok := ud.Value.(*DataclassField); ok {
							if !seen[f.Name] {
								fieldCopy := *f
								fields = append(fields, &fieldCopy)
								seen[f.Name] = true
							}
						}
					}
				}
			}
		}
	}

	// Process this class's own annotations
	annotations, ok := cls.Dict["__annotations__"]
	if !ok {
		return fields
	}
	annDict, ok := annotations.(*runtime.PyDict)
	if !ok {
		return fields
	}

	for _, key := range annDict.Keys(vm) {
		keyStr := pyStringValue(key)
		if keyStr == "" {
			continue
		}

		annType, _ := annDict.DictGet(key, vm)

		f := &DataclassField{
			Name:           keyStr,
			FieldType:      annType,
			Default:        dcMissing,
			DefaultFactory: dcMissing,
			Init:           true,
			Repr:           true,
			Compare:        true,
		}

		if opts.kwOnly {
			f.KwOnly = true
		}

		// Check if there's a default value or field() in the class dict
		if val, exists := cls.Dict[keyStr]; exists {
			if ud, ok := val.(*runtime.PyUserData); ok {
				if fieldSpec, ok := ud.Value.(*DataclassField); ok {
					f.Default = fieldSpec.Default
					f.DefaultFactory = fieldSpec.DefaultFactory
					f.Init = fieldSpec.Init
					f.Repr = fieldSpec.Repr
					f.Compare = fieldSpec.Compare
					f.Hash = fieldSpec.Hash
					f.KwOnly = fieldSpec.KwOnly
				} else {
					f.Default = val
				}
			} else {
				f.Default = val
			}
			// Remove the default from class dict
			delete(cls.Dict, keyStr)
		}

		if seen[keyStr] {
			for i, existing := range fields {
				if existing.Name == keyStr {
					fields[i] = f
					break
				}
			}
		} else {
			fields = append(fields, f)
			seen[keyStr] = true
		}
	}

	return fields
}

func pyStringValue(v runtime.Value) string {
	if s, ok := v.(*runtime.PyString); ok {
		return s.Value
	}
	return ""
}

// makeInit generates the __init__ method
func makeInit(vm *runtime.VM, cls *runtime.PyClass, fields []*DataclassField, frozen bool) *runtime.PyBuiltinFunc {
	return &runtime.PyBuiltinFunc{
		Name: "__init__",
		Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("TypeError: __init__() missing required argument: 'self'")
			}

			self, ok := args[0].(*runtime.PyInstance)
			if !ok {
				return nil, fmt.Errorf("TypeError: __init__() first argument must be an instance")
			}

			argIdx := 1
			for _, f := range fields {
				if !f.Init {
					var val runtime.Value
					if f.DefaultFactory != dcMissing {
						var err error
						val, err = vm.Call(f.DefaultFactory, nil, nil)
						if err != nil {
							return nil, err
						}
					} else if f.Default != dcMissing {
						val = f.Default
					} else {
						continue
					}
					dcSetField(self, f.Name, val, frozen)
					continue
				}

				var val runtime.Value

				// Check kwargs first
				if kwargs != nil {
					if v, ok := kwargs[f.Name]; ok {
						dcSetField(self, f.Name, v, frozen)
						continue
					}
				}

				// Check positional args
				if argIdx < len(args) {
					val = args[argIdx]
					argIdx++
				} else if f.Default != dcMissing {
					val = f.Default
				} else if f.DefaultFactory != dcMissing {
					var err error
					val, err = vm.Call(f.DefaultFactory, nil, nil)
					if err != nil {
						return nil, err
					}
				} else {
					return nil, fmt.Errorf("TypeError: __init__() missing required argument: '%s'", f.Name)
				}

				dcSetField(self, f.Name, val, frozen)
			}

			// Call __post_init__ if it exists
			if postInit, ok := cls.Dict["__post_init__"]; ok {
				_, err := vm.Call(postInit, []runtime.Value{self}, nil)
				if err != nil {
					return nil, err
				}
			}

			return runtime.None, nil
		},
	}
}

// dcSetField sets a field on an instance, bypassing __setattr__ for frozen dataclasses
func dcSetField(inst *runtime.PyInstance, name string, val runtime.Value, frozen bool) {
	if inst.Slots != nil {
		inst.Slots[name] = val
	} else {
		if inst.Dict == nil {
			inst.Dict = make(map[string]runtime.Value)
		}
		inst.Dict[name] = val
	}
}

// makeRepr generates the __repr__ method
func makeRepr(vm *runtime.VM, cls *runtime.PyClass, fields []*DataclassField) *runtime.PyBuiltinFunc {
	return &runtime.PyBuiltinFunc{
		Name: "__repr__",
		Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("TypeError: __repr__() missing required argument: 'self'")
			}
			self, ok := args[0].(*runtime.PyInstance)
			if !ok {
				return nil, fmt.Errorf("TypeError: __repr__() requires an instance")
			}

			var parts []string
			for _, f := range fields {
				if !f.Repr {
					continue
				}
				val := getFieldValue(self, f.Name)
				parts = append(parts, fmt.Sprintf("%s=%s", f.Name, vm.Repr(val)))
			}

			return runtime.NewString(fmt.Sprintf("%s(%s)", cls.Name, strings.Join(parts, ", "))), nil
		},
	}
}

// makeEq generates the __eq__ method
func makeEq(vm *runtime.VM, cls *runtime.PyClass, fields []*DataclassField) *runtime.PyBuiltinFunc {
	return &runtime.PyBuiltinFunc{
		Name: "__eq__",
		Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("TypeError: __eq__() requires 2 arguments")
			}
			self, ok := args[0].(*runtime.PyInstance)
			if !ok {
				return nil, nil // NotImplemented
			}
			other, ok := args[1].(*runtime.PyInstance)
			if !ok {
				return nil, nil // NotImplemented
			}
			if self.Class != other.Class {
				return nil, nil // NotImplemented
			}

			for _, f := range fields {
				if !f.Compare {
					continue
				}
				selfVal := getFieldValue(self, f.Name)
				otherVal := getFieldValue(other, f.Name)
				if !vm.Equal(selfVal, otherVal) {
					return runtime.False, nil
				}
			}

			return runtime.True, nil
		},
	}
}

// makeOrder generates __lt__, __le__, __gt__, __ge__ methods
func makeOrder(vm *runtime.VM, cls *runtime.PyClass, fields []*DataclassField, op string) *runtime.PyBuiltinFunc {
	return &runtime.PyBuiltinFunc{
		Name: "__" + op + "__",
		Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("TypeError: __%s__() requires 2 arguments", op)
			}
			self, ok := args[0].(*runtime.PyInstance)
			if !ok {
				return nil, nil // NotImplemented
			}
			other, ok := args[1].(*runtime.PyInstance)
			if !ok {
				return nil, nil // NotImplemented
			}
			if self.Class != other.Class {
				return nil, nil // NotImplemented
			}

			var selfItems, otherItems []runtime.Value
			for _, f := range fields {
				if !f.Compare {
					continue
				}
				selfItems = append(selfItems, getFieldValue(self, f.Name))
				otherItems = append(otherItems, getFieldValue(other, f.Name))
			}

			selfTuple := &runtime.PyTuple{Items: selfItems}
			otherTuple := &runtime.PyTuple{Items: otherItems}

			var opcode runtime.Opcode
			switch op {
			case "lt":
				opcode = runtime.OpCompareLt
			case "le":
				opcode = runtime.OpCompareLe
			case "gt":
				opcode = runtime.OpCompareGt
			case "ge":
				opcode = runtime.OpCompareGe
			}

			return vm.CompareOp(opcode, selfTuple, otherTuple), nil
		},
	}
}

// makeFrozenSetattr generates __setattr__ that raises FrozenInstanceError
func makeFrozenSetattr() *runtime.PyBuiltinFunc {
	return &runtime.PyBuiltinFunc{
		Name: "__setattr__",
		Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
			return nil, &runtime.PyException{
				TypeName: "FrozenInstanceError",
				Message:  "cannot assign to field",
			}
		},
	}
}

// makeFrozenDelattr generates __delattr__ that raises FrozenInstanceError
func makeFrozenDelattr() *runtime.PyBuiltinFunc {
	return &runtime.PyBuiltinFunc{
		Name: "__delattr__",
		Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
			return nil, &runtime.PyException{
				TypeName: "FrozenInstanceError",
				Message:  "cannot delete field",
			}
		},
	}
}

// makeFrozenHash generates __hash__ for frozen dataclasses
func makeFrozenHash(vm *runtime.VM, fields []*DataclassField) *runtime.PyBuiltinFunc {
	return &runtime.PyBuiltinFunc{
		Name: "__hash__",
		Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("TypeError: __hash__() requires 1 argument")
			}
			self, ok := args[0].(*runtime.PyInstance)
			if !ok {
				return nil, fmt.Errorf("TypeError: __hash__() first argument must be an instance")
			}

			var items []runtime.Value
			for _, f := range fields {
				if !f.Compare {
					continue
				}
				items = append(items, getFieldValue(self, f.Name))
			}

			tuple := &runtime.PyTuple{Items: items}
			h := vm.HashValue(tuple)
			return runtime.NewInt(int64(h)), nil
		},
	}
}

// getFieldValue gets a field value from an instance
func getFieldValue(inst *runtime.PyInstance, name string) runtime.Value {
	if inst.Slots != nil {
		if v, ok := inst.Slots[name]; ok {
			return v
		}
	}
	if inst.Dict != nil {
		if v, ok := inst.Dict[name]; ok {
			return v
		}
	}
	return runtime.None
}

// makeFieldsFn creates the fields() function
func makeFieldsFn(vm *runtime.VM) func([]runtime.Value, map[string]runtime.Value) (runtime.Value, error) {
	return func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("TypeError: fields() takes exactly 1 argument (%d given)", len(args))
		}

		var cls *runtime.PyClass
		switch v := args[0].(type) {
		case *runtime.PyClass:
			cls = v
		case *runtime.PyInstance:
			cls = v.Class
		default:
			return nil, fmt.Errorf("TypeError: fields() argument must be a dataclass or instance of one")
		}

		fieldsDict, ok := cls.Dict["__dataclass_fields__"]
		if !ok {
			return nil, fmt.Errorf("TypeError: fields() argument must be a dataclass or instance of one")
		}

		pyDict, ok := fieldsDict.(*runtime.PyDict)
		if !ok {
			return nil, fmt.Errorf("TypeError: fields() argument must be a dataclass or instance of one")
		}

		var result []runtime.Value
		for _, key := range pyDict.Keys(vm) {
			val, _ := pyDict.DictGet(key, vm)
			result = append(result, val)
		}

		return runtime.NewTuple(result), nil
	}
}

// makeAsDictFn creates the asdict() function
func makeAsDictFn(vm *runtime.VM) func([]runtime.Value, map[string]runtime.Value) (runtime.Value, error) {
	return func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("TypeError: asdict() takes exactly 1 argument (%d given)", len(args))
		}

		inst, ok := args[0].(*runtime.PyInstance)
		if !ok {
			return nil, fmt.Errorf("TypeError: asdict() argument must be a dataclass instance")
		}

		fieldsDict, ok := inst.Class.Dict["__dataclass_fields__"]
		if !ok {
			return nil, fmt.Errorf("TypeError: asdict() argument must be a dataclass instance")
		}

		pyDict, ok := fieldsDict.(*runtime.PyDict)
		if !ok {
			return nil, fmt.Errorf("TypeError: asdict() argument must be a dataclass instance")
		}

		result := &runtime.PyDict{Items: make(map[runtime.Value]runtime.Value)}
		for _, key := range pyDict.Keys(vm) {
			keyStr := pyStringValue(key)
			if keyStr == "" {
				continue
			}
			val := getFieldValue(inst, keyStr)
			result.DictSet(key, val, vm)
		}

		return result, nil
	}
}

// makeAsTupleFn creates the astuple() function
func makeAsTupleFn(vm *runtime.VM) func([]runtime.Value, map[string]runtime.Value) (runtime.Value, error) {
	return func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("TypeError: astuple() takes exactly 1 argument (%d given)", len(args))
		}

		inst, ok := args[0].(*runtime.PyInstance)
		if !ok {
			return nil, fmt.Errorf("TypeError: astuple() argument must be a dataclass instance")
		}

		fieldsDict, ok := inst.Class.Dict["__dataclass_fields__"]
		if !ok {
			return nil, fmt.Errorf("TypeError: astuple() argument must be a dataclass instance")
		}

		pyDict, ok := fieldsDict.(*runtime.PyDict)
		if !ok {
			return nil, fmt.Errorf("TypeError: astuple() argument must be a dataclass instance")
		}

		var items []runtime.Value
		for _, key := range pyDict.Keys(vm) {
			keyStr := pyStringValue(key)
			if keyStr == "" {
				continue
			}
			items = append(items, getFieldValue(inst, keyStr))
		}

		return runtime.NewTuple(items), nil
	}
}
