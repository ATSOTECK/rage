package stdlib

import (
	"fmt"
	"strings"

	"github.com/ATSOTECK/rage/internal/runtime"
)

// InitEnumModule registers the enum module.
func InitEnumModule() {
	runtime.RegisterModule("enum", func(vm *runtime.VM) *runtime.PyModule {
		objectClass, ok := vm.GetBuiltin("object").(*runtime.PyClass)
		if !ok {
			return nil
		}
		typeClass, ok := vm.GetBuiltin("type").(*runtime.PyClass)
		if !ok {
			return nil
		}

		// Forward declarations — filled in after enumType is built.
		var enumClass *runtime.PyClass
		var intEnumClass *runtime.PyClass
		var strEnumClass *runtime.PyClass
		var flagClass *runtime.PyClass
		var intFlagClass *runtime.PyClass
		var enumType *runtime.PyClass

		// auto sentinel class
		autoClass := &runtime.PyClass{
			Name:  "auto",
			Bases: []*runtime.PyClass{objectClass},
			Dict:  make(map[string]runtime.Value),
		}
		autoClass.Mro, _ = vm.ComputeC3MRO(autoClass, autoClass.Bases)
		autoClass.Dict["__init__"] = &runtime.PyBuiltinFunc{
			Name: "__init__",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				return runtime.None, nil
			},
		}

		// isAutoInstance checks if a value is an auto() sentinel.
		isAutoInstance := func(v runtime.Value) bool {
			if inst, ok := v.(*runtime.PyInstance); ok {
				return inst.Class == autoClass
			}
			return false
		}

		// Helper: check if a name should be a member (skip dunders, descriptors, callables from base)
		isMemberCandidate := func(name string, val runtime.Value) bool {
			if strings.HasPrefix(name, "__") && strings.HasSuffix(name, "__") {
				return false
			}
			if strings.HasPrefix(name, "_") {
				return false
			}
			switch val.(type) {
			case *runtime.PyFunction, *runtime.PyBuiltinFunc, *runtime.PyClassMethod,
				*runtime.PyStaticMethod, *runtime.PyProperty:
				return false
			}
			return true
		}

		// Helper: get int value from a runtime.Value
		getIntValue := func(v runtime.Value) (int64, bool) {
			switch val := v.(type) {
			case *runtime.PyInt:
				return val.Value, true
			case *runtime.PyBool:
				if val.Value {
					return 1, true
				}
				return 0, true
			}
			return 0, false
		}

		// ── EnumType metaclass ──────────────────────────────────────────────

		enumType = &runtime.PyClass{
			Name:  "EnumType",
			Bases: []*runtime.PyClass{typeClass},
			Dict:  make(map[string]runtime.Value),
		}
		enumType.Mro, _ = vm.ComputeC3MRO(enumType, enumType.Bases)

		// isEnumBase returns true if cls is one of the base enum classes (Enum, IntEnum, etc.)
		isEnumBase := func(cls *runtime.PyClass) bool {
			return cls == enumClass || cls == intEnumClass || cls == strEnumClass ||
				cls == flagClass || cls == intFlagClass
		}

		// hasEnumBase returns true if any base in bases has EnumType as metaclass
		hasEnumBase := func(bases []*runtime.PyClass) bool {
			for _, b := range bases {
				if b.Metaclass == enumType {
					return true
				}
			}
			return false
		}

		// getEnumBase returns the first enum base class
		getEnumBase := func(bases []*runtime.PyClass) *runtime.PyClass {
			for _, b := range bases {
				if b.Metaclass == enumType {
					return b
				}
			}
			return nil
		}

		// isFlagType checks if a class is Flag or IntFlag (or derives from them)
		isFlagType := func(cls *runtime.PyClass) bool {
			for _, mro := range cls.Mro {
				if mro == flagClass || mro == intFlagClass {
					return true
				}
			}
			return false
		}

		// isIntEnumType checks if a class derives from IntEnum or IntFlag
		isIntEnumType := func(cls *runtime.PyClass) bool {
			for _, mro := range cls.Mro {
				if mro == intEnumClass || mro == intFlagClass {
					return true
				}
			}
			return false
		}

		// isStrEnumType checks if a class derives from StrEnum
		isStrEnumType := func(cls *runtime.PyClass) bool {
			for _, mro := range cls.Mro {
				if mro == strEnumClass {
					return true
				}
			}
			return false
		}

		// ── EnumType.__new__ ────────────────────────────────────────────────

		enumType.Dict["__new__"] = &runtime.PyStaticMethod{
			Func: &runtime.PyBuiltinFunc{
				Name: "__new__",
				Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
					if len(args) < 4 {
						return nil, fmt.Errorf("TypeError: EnumType.__new__() takes 4 arguments")
					}
					mcs, ok := args[0].(*runtime.PyClass)
					if !ok {
						return nil, fmt.Errorf("TypeError: EnumType.__new__() argument 1 must be a class")
					}
					nameVal, ok := args[1].(*runtime.PyString)
					if !ok {
						return nil, fmt.Errorf("TypeError: EnumType.__new__() argument 2 must be a string")
					}
					nameStr := nameVal.Value
					basesTuple, ok := args[2].(*runtime.PyTuple)
					if !ok {
						return nil, fmt.Errorf("TypeError: EnumType.__new__() argument 3 must be a tuple")
					}
					nsDict, ok := args[3].(*runtime.PyDict)
					if !ok {
						return nil, fmt.Errorf("TypeError: EnumType.__new__() argument 4 must be a dict")
					}

					// Convert bases
					bases := make([]*runtime.PyClass, len(basesTuple.Items))
					for i, b := range basesTuple.Items {
						bc, ok := b.(*runtime.PyClass)
						if !ok {
							return nil, fmt.Errorf("TypeError: EnumType bases must be classes")
						}
						bases[i] = bc
					}

					// Build class dict from namespace
					classDict := make(map[string]runtime.Value)
					for _, k := range nsDict.Keys(vm) {
						if ks, ok := k.(*runtime.PyString); ok {
							if v, ok := nsDict.DictGet(k, vm); ok {
								classDict[ks.Value] = v
							}
						}
					}

					// If no enum base, this is a base enum class definition — just create the class.
					if !hasEnumBase(bases) {
						cls := &runtime.PyClass{
							Name:      nameStr,
							Bases:     bases,
							Dict:      classDict,
							Metaclass: mcs,
						}
						cls.Mro, _ = vm.ComputeC3MRO(cls, bases)
						return cls, nil
					}

					// ── User-defined enum ───────────────────────────────

					// Create the class shell
					cls := &runtime.PyClass{
						Name:      nameStr,
						Bases:     bases,
						Dict:      classDict,
						Metaclass: mcs,
					}
					cls.Mro, _ = vm.ComputeC3MRO(cls, bases)

					enumBase := getEnumBase(bases)

					// Collect member names in order (from namespace iteration order)
					var memberNames []string
					memberMap := make(map[string]*runtime.PyInstance)     // name → member
					value2member := make(map[string]*runtime.PyInstance)  // str(value) → member (first one wins)
					value2memberKeys := make([]runtime.Value, 0)         // original value keys for lookup

					// Determine _generate_next_value_ function
					var generateNextValue func(name string, start int, count int, lastValues []runtime.Value) runtime.Value

					// Check class dict first, then MRO
					if gnv, ok := classDict["_generate_next_value_"]; ok {
						switch fn := gnv.(type) {
						case *runtime.PyFunction:
							generateNextValue = func(name string, start int, count int, lastValues []runtime.Value) runtime.Value {
								items := make([]runtime.Value, len(lastValues))
								copy(items, lastValues)
								result, err := vm.CallFunction(fn, []runtime.Value{
									runtime.NewString(name),
									runtime.MakeInt(int64(start)),
									runtime.MakeInt(int64(count)),
									&runtime.PyList{Items: items},
								}, nil)
								if err != nil {
									return runtime.MakeInt(int64(count + 1))
								}
								return result
							}
						case *runtime.PyBuiltinFunc:
							generateNextValue = func(name string, start int, count int, lastValues []runtime.Value) runtime.Value {
								items := make([]runtime.Value, len(lastValues))
								copy(items, lastValues)
								result, err := fn.Fn([]runtime.Value{
									runtime.NewString(name),
									runtime.MakeInt(int64(start)),
									runtime.MakeInt(int64(count)),
									&runtime.PyList{Items: items},
								}, nil)
								if err != nil {
									return runtime.MakeInt(int64(count + 1))
								}
								return result
							}
						case *runtime.PyStaticMethod:
							generateNextValue = func(name string, start int, count int, lastValues []runtime.Value) runtime.Value {
								items := make([]runtime.Value, len(lastValues))
								copy(items, lastValues)
								result, err := vm.Call(fn.Func, []runtime.Value{
									runtime.NewString(name),
									runtime.MakeInt(int64(start)),
									runtime.MakeInt(int64(count)),
									&runtime.PyList{Items: items},
								}, nil)
								if err != nil {
									return runtime.MakeInt(int64(count + 1))
								}
								return result
							}
						}
					}

					if generateNextValue == nil {
						// Check enum base's Dict
						for _, mro := range enumBase.Mro {
							if gnv, ok := mro.Dict["_generate_next_value_"]; ok {
								switch fn := gnv.(type) {
								case *runtime.PyBuiltinFunc:
									generateNextValue = func(name string, start int, count int, lastValues []runtime.Value) runtime.Value {
										items := make([]runtime.Value, len(lastValues))
										copy(items, lastValues)
										result, _ := fn.Fn([]runtime.Value{
											runtime.NewString(name),
											runtime.MakeInt(int64(start)),
											runtime.MakeInt(int64(count)),
											&runtime.PyList{Items: items},
										}, nil)
										return result
									}
								case *runtime.PyStaticMethod:
									generateNextValue = func(name string, start int, count int, lastValues []runtime.Value) runtime.Value {
										items := make([]runtime.Value, len(lastValues))
										copy(items, lastValues)
										result, _ := vm.Call(fn.Func, []runtime.Value{
											runtime.NewString(name),
											runtime.MakeInt(int64(start)),
											runtime.MakeInt(int64(count)),
											&runtime.PyList{Items: items},
										}, nil)
										return result
									}
								}
								if generateNextValue != nil {
									break
								}
							}
						}
					}

					if generateNextValue == nil {
						// Default: sequential integers
						generateNextValue = func(name string, start int, count int, lastValues []runtime.Value) runtime.Value {
							return runtime.MakeInt(int64(count + 1))
						}
					}

					// Scan namespace for member candidates
					var lastValues []runtime.Value
					count := 0

					// Iterate the namespace dict to get members (in insertion order)
					for _, k := range nsDict.Keys(vm) {
						ks, ok := k.(*runtime.PyString)
						if !ok {
							continue
						}
						v, ok := nsDict.DictGet(k, vm)
						if !ok {
							continue
						}
						name := ks.Value
						if !isMemberCandidate(name, v) {
							continue
						}

						memberValue := v
						if isAutoInstance(v) {
							memberValue = generateNextValue(name, 1, count, lastValues)
						}

						// Check for alias (duplicate value)
						valKey := vm.Repr(memberValue)
						if existing, ok := value2member[valKey]; ok {
							// Alias: point to the same member instance
							memberMap[name] = existing
							cls.Dict[name] = existing
							// Don't add to memberNames (aliases are not in iteration order)
							continue
						}

						// Create member instance
						member := &runtime.PyInstance{
							Class: cls,
							Dict:  make(map[string]runtime.Value),
						}
						member.Dict["_name_"] = runtime.NewString(name)
						member.Dict["_value_"] = memberValue
						member.Dict["name"] = runtime.NewString(name)
						member.Dict["value"] = memberValue

						memberNames = append(memberNames, name)
						memberMap[name] = member
						value2member[valKey] = member
						value2memberKeys = append(value2memberKeys, memberValue)

						// Store in class dict for attribute access
						cls.Dict[name] = member

						lastValues = append(lastValues, memberValue)
						count++
					}

					// Store metadata on the class
					memberNamesList := make([]runtime.Value, len(memberNames))
					for i, n := range memberNames {
						memberNamesList[i] = runtime.NewString(n)
					}
					cls.Dict["_member_names_"] = &runtime.PyList{Items: memberNamesList}

					// _member_map_ as a PyDict
					mmDict := runtime.NewDict()
					for name, member := range memberMap {
						mmDict.DictSet(runtime.NewString(name), member, vm)
					}
					cls.Dict["_member_map_"] = mmDict

					// _value2member_map_ as a PyDict
					v2mDict := runtime.NewDict()
					for valKey, member := range value2member {
						// Store with repr-key for lookup
						v2mDict.DictSet(runtime.NewString(valKey), member, vm)
					}
					cls.Dict["_value2member_map_"] = v2mDict

					// Store the original value keys for value-based lookup
					cls.Dict["_value2member_keys_"] = &runtime.PyList{Items: value2memberKeys}

					// __members__ property — returns _member_map_ copy
					cls.Dict["__members__"] = mmDict

					// ── Install dunders on the class ────────────────────

					// __iter__: iterate members in definition order
					// getIter calls getAttr(obj, "__iter__") which returns the raw func,
					// then calls it with no args. So __iter__ must work with 0 or 1 args.
					cls.Dict["__iter__"] = &runtime.PyBuiltinFunc{
						Name: "__iter__",
						Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
							namesList := cls.Dict["_member_names_"].(*runtime.PyList)
							items := make([]runtime.Value, len(namesList.Items))
							for i, n := range namesList.Items {
								name := n.(*runtime.PyString).Value
								items[i] = memberMap[name]
							}
							return &runtime.PyIterator{Items: items, Index: 0}, nil
						},
					}

					// __contains__: check if value is a member
					// Called from contains() with args = [cls, item]
					cls.Dict["__contains__"] = &runtime.PyBuiltinFunc{
						Name: "__contains__",
						Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
							// args may be [cls, item] or [item] depending on call path
							var item runtime.Value
							if len(args) >= 2 {
								item = args[1]
							} else if len(args) == 1 {
								item = args[0]
							} else {
								return runtime.False, nil
							}
							// Check if item is a member instance
							if inst, ok := item.(*runtime.PyInstance); ok {
								for _, m := range memberMap {
									if m == inst {
										return runtime.True, nil
									}
								}
							}
							return runtime.False, nil
						},
					}

					// __class_getitem__: Color['RED'] → name-based lookup
					cls.Dict["__class_getitem__"] = &runtime.PyBuiltinFunc{
						Name: "__class_getitem__",
						Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
							if len(args) < 2 {
								return nil, fmt.Errorf("TypeError: __class_getitem__ requires 2 arguments")
							}
							key := args[1]
							if ks, ok := key.(*runtime.PyString); ok {
								if m, ok := memberMap[ks.Value]; ok {
									return m, nil
								}
								return nil, fmt.Errorf("KeyError: '%s'", ks.Value)
							}
							return nil, fmt.Errorf("TypeError: enum key must be a string")
						},
					}

					// __repr__: <ClassName.MEMBER: value>
					cls.Dict["__repr__"] = &runtime.PyBuiltinFunc{
						Name: "__repr__",
						Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
							if len(args) < 1 {
								return runtime.NewString("<?>"), nil
							}
							inst, ok := args[0].(*runtime.PyInstance)
							if !ok {
								return runtime.NewString(fmt.Sprintf("<%s>", cls.Name)), nil
							}
							mName := ""
							mValue := ""
							if n, ok := inst.Dict["_name_"]; ok {
								if ns, ok := n.(*runtime.PyString); ok {
									mName = ns.Value
								}
							}
							if v, ok := inst.Dict["_value_"]; ok {
								mValue = vm.Repr(v)
							}
							return runtime.NewString(fmt.Sprintf("<%s.%s: %s>", cls.Name, mName, mValue)), nil
						},
					}

					// __str__: ClassName.MEMBER
					strFn := &runtime.PyBuiltinFunc{
						Name: "__str__",
						Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
							if len(args) < 1 {
								return runtime.NewString(""), nil
							}
							inst, ok := args[0].(*runtime.PyInstance)
							if !ok {
								return runtime.NewString(cls.Name), nil
							}
							mName := ""
							if n, ok := inst.Dict["_name_"]; ok {
								if ns, ok := n.(*runtime.PyString); ok {
									mName = ns.Value
								}
							}
							return runtime.NewString(fmt.Sprintf("%s.%s", cls.Name, mName)), nil
						},
					}
					cls.Dict["__str__"] = strFn

					// __eq__: identity-based for regular enums
					cls.Dict["__eq__"] = &runtime.PyBuiltinFunc{
						Name: "__eq__",
						Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
							if len(args) < 2 {
								return runtime.False, nil
							}
							if args[0] == args[1] {
								return runtime.True, nil
							}
							// For IntEnum/IntFlag, compare by value to ints
							if isIntEnumType(cls) {
								selfInst, ok1 := args[0].(*runtime.PyInstance)
								if ok1 {
									selfVal := selfInst.Dict["_value_"]
									otherVal := args[1]
									if otherInst, ok := args[1].(*runtime.PyInstance); ok {
										otherVal = otherInst.Dict["_value_"]
									}
									if vm.Equal(selfVal, otherVal) {
										return runtime.True, nil
									}
								}
							}
							// For StrEnum, compare by value to strings
							if isStrEnumType(cls) {
								selfInst, ok1 := args[0].(*runtime.PyInstance)
								if ok1 {
									selfVal := selfInst.Dict["_value_"]
									otherVal := args[1]
									if otherInst, ok := args[1].(*runtime.PyInstance); ok {
										otherVal = otherInst.Dict["_value_"]
									}
									if vm.Equal(selfVal, otherVal) {
										return runtime.True, nil
									}
								}
							}
							return runtime.False, nil
						},
					}

					// __hash__: based on value
					cls.Dict["__hash__"] = &runtime.PyBuiltinFunc{
						Name: "__hash__",
						Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
							if len(args) < 1 {
								return runtime.MakeInt(0), nil
							}
							inst, ok := args[0].(*runtime.PyInstance)
							if !ok {
								return runtime.MakeInt(0), nil
							}
							val := inst.Dict["_value_"]
							h := vm.HashValue(val)
							return runtime.MakeInt(int64(h)), nil
						},
					}

					// __ne__: not equal
					cls.Dict["__ne__"] = &runtime.PyBuiltinFunc{
						Name: "__ne__",
						Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
							if len(args) < 2 {
								return runtime.True, nil
							}
							if args[0] == args[1] {
								return runtime.False, nil
							}
							if isIntEnumType(cls) || isStrEnumType(cls) {
								selfInst, ok1 := args[0].(*runtime.PyInstance)
								if ok1 {
									selfVal := selfInst.Dict["_value_"]
									otherVal := args[1]
									if otherInst, ok := args[1].(*runtime.PyInstance); ok {
										otherVal = otherInst.Dict["_value_"]
									}
									if vm.Equal(selfVal, otherVal) {
										return runtime.False, nil
									}
								}
							}
							return runtime.True, nil
						},
					}

					// ── IntEnum / IntFlag specific dunders ──────────────
					if isIntEnumType(cls) {
						// __int__
						cls.Dict["__int__"] = &runtime.PyBuiltinFunc{
							Name: "__int__",
							Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
								if len(args) < 1 {
									return runtime.MakeInt(0), nil
								}
								inst, ok := args[0].(*runtime.PyInstance)
								if !ok {
									return runtime.MakeInt(0), nil
								}
								return inst.Dict["_value_"], nil
							},
						}

						// comparison operators for IntEnum
						for _, op := range []struct {
							name string
							cmp  func(a, b int64) bool
						}{
							{"__lt__", func(a, b int64) bool { return a < b }},
							{"__le__", func(a, b int64) bool { return a <= b }},
							{"__gt__", func(a, b int64) bool { return a > b }},
							{"__ge__", func(a, b int64) bool { return a >= b }},
						} {
							cmpFn := op.cmp
							opName := op.name
							cls.Dict[opName] = &runtime.PyBuiltinFunc{
								Name: opName,
								Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
									if len(args) < 2 {
										return runtime.NotImplemented, nil
									}
									var selfVal, otherVal int64
									var ok bool
									if inst, ok2 := args[0].(*runtime.PyInstance); ok2 {
										selfVal, ok = getIntValue(inst.Dict["_value_"])
										if !ok {
											return runtime.NotImplemented, nil
										}
									} else {
										return runtime.NotImplemented, nil
									}
									if inst, ok2 := args[1].(*runtime.PyInstance); ok2 {
										otherVal, ok = getIntValue(inst.Dict["_value_"])
									} else {
										otherVal, ok = getIntValue(args[1])
									}
									if !ok {
										return runtime.NotImplemented, nil
									}
									if cmpFn(selfVal, otherVal) {
										return runtime.True, nil
									}
									return runtime.False, nil
								},
							}
						}

						// arithmetic ops for IntEnum
						for _, op := range []struct {
							name string
							fn   func(a, b int64) int64
						}{
							{"__add__", func(a, b int64) int64 { return a + b }},
							{"__sub__", func(a, b int64) int64 { return a - b }},
							{"__mul__", func(a, b int64) int64 { return a * b }},
							{"__mod__", func(a, b int64) int64 { return a % b }},
							{"__floordiv__", func(a, b int64) int64 { return a / b }},
						} {
							arithFn := op.fn
							opName := op.name
							cls.Dict[opName] = &runtime.PyBuiltinFunc{
								Name: opName,
								Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
									if len(args) < 2 {
										return runtime.NotImplemented, nil
									}
									var selfVal, otherVal int64
									var ok bool
									if inst, ok2 := args[0].(*runtime.PyInstance); ok2 {
										selfVal, ok = getIntValue(inst.Dict["_value_"])
										if !ok {
											return runtime.NotImplemented, nil
										}
									} else {
										return runtime.NotImplemented, nil
									}
									if inst, ok2 := args[1].(*runtime.PyInstance); ok2 {
										otherVal, ok = getIntValue(inst.Dict["_value_"])
									} else {
										otherVal, ok = getIntValue(args[1])
									}
									if !ok {
										return runtime.NotImplemented, nil
									}
									if opName == "__mod__" || opName == "__floordiv__" {
										if otherVal == 0 {
											return nil, fmt.Errorf("ZeroDivisionError: integer division or modulo by zero")
										}
									}
									return runtime.MakeInt(arithFn(selfVal, otherVal)), nil
								},
							}
						}
					}

					// ── StrEnum specific dunders ────────────────────────
					if isStrEnumType(cls) {
						// __str__ for StrEnum returns the value (a string), not ClassName.MEMBER
						cls.Dict["__str__"] = &runtime.PyBuiltinFunc{
							Name: "__str__",
							Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
								if len(args) < 1 {
									return runtime.NewString(""), nil
								}
								inst, ok := args[0].(*runtime.PyInstance)
								if !ok {
									return runtime.NewString(""), nil
								}
								return inst.Dict["_value_"], nil
							},
						}
					}

					// ── Flag / IntFlag specific dunders ─────────────────
					if isFlagType(cls) {
						// __or__: combine flags
						cls.Dict["__or__"] = &runtime.PyBuiltinFunc{
							Name: "__or__",
							Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
								if len(args) < 2 {
									return runtime.NotImplemented, nil
								}
								var aVal, bVal int64
								var ok bool
								if inst, ok2 := args[0].(*runtime.PyInstance); ok2 {
									aVal, ok = getIntValue(inst.Dict["_value_"])
									if !ok {
										return runtime.NotImplemented, nil
									}
								} else {
									return runtime.NotImplemented, nil
								}
								if inst, ok2 := args[1].(*runtime.PyInstance); ok2 {
									bVal, ok = getIntValue(inst.Dict["_value_"])
								} else {
									bVal, ok = getIntValue(args[1])
								}
								if !ok {
									return runtime.NotImplemented, nil
								}
								combined := aVal | bVal
								// Look up existing member or create composite
								return flagLookupOrCreate(cls, memberMap, value2member, combined, vm), nil
							},
						}

						// __and__
						cls.Dict["__and__"] = &runtime.PyBuiltinFunc{
							Name: "__and__",
							Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
								if len(args) < 2 {
									return runtime.NotImplemented, nil
								}
								var aVal, bVal int64
								var ok bool
								if inst, ok2 := args[0].(*runtime.PyInstance); ok2 {
									aVal, ok = getIntValue(inst.Dict["_value_"])
									if !ok {
										return runtime.NotImplemented, nil
									}
								} else {
									return runtime.NotImplemented, nil
								}
								if inst, ok2 := args[1].(*runtime.PyInstance); ok2 {
									bVal, ok = getIntValue(inst.Dict["_value_"])
								} else {
									bVal, ok = getIntValue(args[1])
								}
								if !ok {
									return runtime.NotImplemented, nil
								}
								combined := aVal & bVal
								return flagLookupOrCreate(cls, memberMap, value2member, combined, vm), nil
							},
						}

						// __xor__
						cls.Dict["__xor__"] = &runtime.PyBuiltinFunc{
							Name: "__xor__",
							Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
								if len(args) < 2 {
									return runtime.NotImplemented, nil
								}
								var aVal, bVal int64
								var ok bool
								if inst, ok2 := args[0].(*runtime.PyInstance); ok2 {
									aVal, ok = getIntValue(inst.Dict["_value_"])
									if !ok {
										return runtime.NotImplemented, nil
									}
								} else {
									return runtime.NotImplemented, nil
								}
								if inst, ok2 := args[1].(*runtime.PyInstance); ok2 {
									bVal, ok = getIntValue(inst.Dict["_value_"])
								} else {
									bVal, ok = getIntValue(args[1])
								}
								if !ok {
									return runtime.NotImplemented, nil
								}
								combined := aVal ^ bVal
								return flagLookupOrCreate(cls, memberMap, value2member, combined, vm), nil
							},
						}

						// __invert__
						cls.Dict["__invert__"] = &runtime.PyBuiltinFunc{
							Name: "__invert__",
							Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
								if len(args) < 1 {
									return runtime.NotImplemented, nil
								}
								inst, ok := args[0].(*runtime.PyInstance)
								if !ok {
									return runtime.NotImplemented, nil
								}
								val, ok := getIntValue(inst.Dict["_value_"])
								if !ok {
									return runtime.NotImplemented, nil
								}
								// Compute all-bits mask from all members
								var allBits int64
								for _, m := range memberMap {
									if mv, ok := getIntValue(m.Dict["_value_"]); ok {
										allBits |= mv
									}
								}
								inverted := allBits & ^val
								return flagLookupOrCreate(cls, memberMap, value2member, inverted, vm), nil
							},
						}

						// __bool__: flag value != 0
						cls.Dict["__bool__"] = &runtime.PyBuiltinFunc{
							Name: "__bool__",
							Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
								if len(args) < 1 {
									return runtime.False, nil
								}
								inst, ok := args[0].(*runtime.PyInstance)
								if !ok {
									return runtime.False, nil
								}
								val, ok := getIntValue(inst.Dict["_value_"])
								if !ok {
									return runtime.False, nil
								}
								if val != 0 {
									return runtime.True, nil
								}
								return runtime.False, nil
							},
						}
					}

					// Remove non-member entries from class dict that were processed
					// (auto sentinels were replaced by member instances above)

					return cls, nil
				},
			},
		}

		// ── EnumType.__call__ ───────────────────────────────────────────

		enumType.Dict["__call__"] = &runtime.PyBuiltinFunc{
			Name: "__call__",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				if len(args) < 2 {
					return nil, fmt.Errorf("TypeError: EnumType.__call__() requires at least 2 arguments")
				}
				cls := args[0].(*runtime.PyClass)
				callArgs := args[1:]

				// If creating a base enum class (Enum, IntEnum, etc), delegate to type
				if isEnumBase(cls) {
					return vm.DefaultClassCall(cls, callArgs, kwargs)
				}

				// Value-based member lookup: Color(1)
				if len(callArgs) == 1 {
					lookupValue := callArgs[0]

					// Search members by value
					namesList, ok := cls.Dict["_member_names_"]
					if !ok {
						return nil, fmt.Errorf("ValueError: %s has no members", cls.Name)
					}
					memberNames := namesList.(*runtime.PyList)

					for _, nameVal := range memberNames.Items {
						name := nameVal.(*runtime.PyString).Value
						member, ok := cls.Dict[name]
						if !ok {
							continue
						}
						inst, ok := member.(*runtime.PyInstance)
						if !ok {
							continue
						}
						memberValue := inst.Dict["_value_"]
						if vm.Equal(memberValue, lookupValue) {
							return inst, nil
						}
					}

					// Try _missing_ hook
					for _, mroCls := range cls.Mro {
						if missing, ok := mroCls.Dict["_missing_"]; ok {
							var result runtime.Value
							var err error
							switch fn := missing.(type) {
							case *runtime.PyFunction:
								result, err = vm.CallFunction(fn, []runtime.Value{cls, lookupValue}, nil)
							case *runtime.PyBuiltinFunc:
								result, err = fn.Fn([]runtime.Value{cls, lookupValue}, nil)
							case *runtime.PyClassMethod:
								result, err = vm.Call(fn.Func, []runtime.Value{cls, lookupValue}, nil)
							}
							if err != nil {
								return nil, err
							}
							if result != nil && result != runtime.None {
								return result, nil
							}
							break
						}
					}

					return nil, fmt.Errorf("ValueError: %s is not a valid %s", vm.Repr(lookupValue), cls.Name)
				}

				// Functional API: Color('Color', ['RED', 'GREEN', 'BLUE'])
				// For now, not implemented — just raise
				return nil, fmt.Errorf("TypeError: %s() takes 1 positional argument for value lookup", cls.Name)
			},
		}

		// ── Enum base class ─────────────────────────────────────────────

		enumClass = &runtime.PyClass{
			Name:      "Enum",
			Bases:     []*runtime.PyClass{objectClass},
			Dict:      make(map[string]runtime.Value),
			Metaclass: enumType,
		}
		enumClass.Mro, _ = vm.ComputeC3MRO(enumClass, enumClass.Bases)

		// _generate_next_value_ for Enum: sequential integers starting at 1
		enumClass.Dict["_generate_next_value_"] = &runtime.PyStaticMethod{
			Func: &runtime.PyBuiltinFunc{
				Name: "_generate_next_value_",
				Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
					// args: name, start, count, last_values
					if len(args) >= 3 {
						count, _ := getIntValue(args[2])
						return runtime.MakeInt(count + 1), nil
					}
					return runtime.MakeInt(1), nil
				},
			},
		}

		// _missing_ classmethod (returns None by default)
		enumClass.Dict["_missing_"] = &runtime.PyClassMethod{
			Func: &runtime.PyBuiltinFunc{
				Name: "_missing_",
				Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
					return runtime.None, nil
				},
			},
		}

		// name property
		enumClass.Dict["name"] = &runtime.PyProperty{
			Fget: &runtime.PyBuiltinFunc{
				Name: "name_getter",
				Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
					if len(args) < 1 {
						return runtime.None, nil
					}
					inst, ok := args[0].(*runtime.PyInstance)
					if !ok {
						return runtime.None, nil
					}
					if n, ok := inst.Dict["_name_"]; ok {
						return n, nil
					}
					return runtime.None, nil
				},
			},
		}

		// value property
		enumClass.Dict["value"] = &runtime.PyProperty{
			Fget: &runtime.PyBuiltinFunc{
				Name: "value_getter",
				Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
					if len(args) < 1 {
						return runtime.None, nil
					}
					inst, ok := args[0].(*runtime.PyInstance)
					if !ok {
						return runtime.None, nil
					}
					if v, ok := inst.Dict["_value_"]; ok {
						return v, nil
					}
					return runtime.None, nil
				},
			},
		}

		// ── IntEnum ─────────────────────────────────────────────────────

		intEnumClass = &runtime.PyClass{
			Name:      "IntEnum",
			Bases:     []*runtime.PyClass{enumClass},
			Dict:      make(map[string]runtime.Value),
			Metaclass: enumType,
		}
		intEnumClass.Mro, _ = vm.ComputeC3MRO(intEnumClass, intEnumClass.Bases)

		// _generate_next_value_ for IntEnum: same as Enum
		intEnumClass.Dict["_generate_next_value_"] = &runtime.PyStaticMethod{
			Func: &runtime.PyBuiltinFunc{
				Name: "_generate_next_value_",
				Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
					if len(args) >= 3 {
						count, _ := getIntValue(args[2])
						return runtime.MakeInt(count + 1), nil
					}
					return runtime.MakeInt(1), nil
				},
			},
		}

		// ── StrEnum ─────────────────────────────────────────────────────

		strEnumClass = &runtime.PyClass{
			Name:      "StrEnum",
			Bases:     []*runtime.PyClass{enumClass},
			Dict:      make(map[string]runtime.Value),
			Metaclass: enumType,
		}
		strEnumClass.Mro, _ = vm.ComputeC3MRO(strEnumClass, strEnumClass.Bases)

		// _generate_next_value_ for StrEnum: lowercased member name
		strEnumClass.Dict["_generate_next_value_"] = &runtime.PyStaticMethod{
			Func: &runtime.PyBuiltinFunc{
				Name: "_generate_next_value_",
				Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
					if len(args) >= 1 {
						if name, ok := args[0].(*runtime.PyString); ok {
							return runtime.NewString(strings.ToLower(name.Value)), nil
						}
					}
					return runtime.NewString(""), nil
				},
			},
		}

		// ── Flag ────────────────────────────────────────────────────────

		flagClass = &runtime.PyClass{
			Name:      "Flag",
			Bases:     []*runtime.PyClass{enumClass},
			Dict:      make(map[string]runtime.Value),
			Metaclass: enumType,
		}
		flagClass.Mro, _ = vm.ComputeC3MRO(flagClass, flagClass.Bases)

		// _generate_next_value_ for Flag: next power of 2
		flagClass.Dict["_generate_next_value_"] = &runtime.PyStaticMethod{
			Func: &runtime.PyBuiltinFunc{
				Name: "_generate_next_value_",
				Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
					// args: name, start, count, last_values
					if len(args) >= 4 {
						lastValues, ok := args[3].(*runtime.PyList)
						if ok && len(lastValues.Items) > 0 {
							// Find highest power of 2 used
							var maxVal int64
							for _, v := range lastValues.Items {
								if iv, ok := getIntValue(v); ok && iv > maxVal {
									maxVal = iv
								}
							}
							if maxVal == 0 {
								return runtime.MakeInt(1), nil
							}
							return runtime.MakeInt(maxVal << 1), nil
						}
					}
					return runtime.MakeInt(1), nil
				},
			},
		}

		// ── IntFlag ─────────────────────────────────────────────────────

		intFlagClass = &runtime.PyClass{
			Name:      "IntFlag",
			Bases:     []*runtime.PyClass{flagClass},
			Dict:      make(map[string]runtime.Value),
			Metaclass: enumType,
		}
		intFlagClass.Mro, _ = vm.ComputeC3MRO(intFlagClass, intFlagClass.Bases)

		intFlagClass.Dict["_generate_next_value_"] = flagClass.Dict["_generate_next_value_"]

		// ── @unique decorator ───────────────────────────────────────────

		uniqueFn := &runtime.PyBuiltinFunc{
			Name: "unique",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("TypeError: unique() takes exactly 1 argument")
				}
				cls, ok := args[0].(*runtime.PyClass)
				if !ok {
					return nil, fmt.Errorf("TypeError: unique() argument must be an enum class")
				}

				// Check for duplicate values
				namesList, ok := cls.Dict["_member_names_"]
				if !ok {
					return cls, nil
				}
				memberNames := namesList.(*runtime.PyList)

				// Build value → first-name map
				seen := make(map[string]string)
				var duplicates []string
				for _, nameVal := range memberNames.Items {
					name := nameVal.(*runtime.PyString).Value
					member, ok := cls.Dict[name]
					if !ok {
						continue
					}
					inst, ok := member.(*runtime.PyInstance)
					if !ok {
						continue
					}
					valRepr := vm.Repr(inst.Dict["_value_"])
					if firstName, exists := seen[valRepr]; exists {
						duplicates = append(duplicates, fmt.Sprintf("%s -> %s", name, firstName))
					} else {
						seen[valRepr] = name
					}
				}

				// Also check aliases (entries in _member_map_ not in _member_names_)
				mmDict, ok := cls.Dict["_member_map_"].(*runtime.PyDict)
				if ok {
					for _, k := range mmDict.Keys(vm) {
						ks, ok := k.(*runtime.PyString)
						if !ok {
							continue
						}
						v, ok := mmDict.DictGet(k, vm)
						if !ok {
							continue
						}
						// Check if this name is in _member_names_
						isInNames := false
						for _, n := range memberNames.Items {
							if ns, ok := n.(*runtime.PyString); ok && ns.Value == ks.Value {
								isInNames = true
								break
							}
						}
						if !isInNames {
							// This is an alias
							inst, ok := v.(*runtime.PyInstance)
							if ok {
								origName := ""
								if n, ok := inst.Dict["_name_"]; ok {
									if ns, ok := n.(*runtime.PyString); ok {
										origName = ns.Value
									}
								}
								duplicates = append(duplicates, fmt.Sprintf("%s -> %s", ks.Value, origName))
							}
						}
					}
				}

				if len(duplicates) > 0 {
					return nil, fmt.Errorf("ValueError: duplicate values found in <%s>: %s",
						cls.Name, strings.Join(duplicates, ", "))
				}
				return cls, nil
			},
		}

		// ── Build module ────────────────────────────────────────────────

		mod := runtime.NewModule("enum")
		mod.Dict["Enum"] = enumClass
		mod.Dict["IntEnum"] = intEnumClass
		mod.Dict["StrEnum"] = strEnumClass
		mod.Dict["Flag"] = flagClass
		mod.Dict["IntFlag"] = intFlagClass
		mod.Dict["EnumType"] = enumType
		mod.Dict["EnumMeta"] = enumType // alias
		mod.Dict["auto"] = autoClass
		mod.Dict["unique"] = uniqueFn
		return mod
	})
}

// flagLookupOrCreate looks up an existing member by value, or creates a composite flag member.
func flagLookupOrCreate(cls *runtime.PyClass, memberMap map[string]*runtime.PyInstance, value2member map[string]*runtime.PyInstance, val int64, vm *runtime.VM) *runtime.PyInstance {
	valKey := vm.Repr(runtime.MakeInt(val))
	if m, ok := value2member[valKey]; ok {
		return m
	}

	// Build composite name from individual flag bits
	var parts []string
	namesList, ok := cls.Dict["_member_names_"]
	if ok {
		for _, n := range namesList.(*runtime.PyList).Items {
			name := n.(*runtime.PyString).Value
			if m, ok := memberMap[name]; ok {
				if mv, ok := m.Dict["_value_"].(*runtime.PyInt); ok {
					if mv.Value != 0 && (val&mv.Value) == mv.Value {
						parts = append(parts, name)
					}
				}
			}
		}
	}

	compositeName := strings.Join(parts, "|")
	if compositeName == "" {
		compositeName = fmt.Sprintf("%d", val)
	}

	member := &runtime.PyInstance{
		Class: cls,
		Dict:  make(map[string]runtime.Value),
	}
	member.Dict["_name_"] = runtime.NewString(compositeName)
	member.Dict["_value_"] = runtime.MakeInt(val)
	member.Dict["name"] = runtime.NewString(compositeName)
	member.Dict["value"] = runtime.MakeInt(val)
	return member
}
