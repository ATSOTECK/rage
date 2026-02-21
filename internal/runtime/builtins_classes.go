package runtime

import (
	"fmt"
	"sort"
)

// initBuiltinsClasses registers class-related builtins: __build_class__, object, type,
// property, classmethod, staticmethod, super, __import__, and constants (None/True/False/NotImplemented).
func (vm *VM) initBuiltinsClasses() {
	vm.builtins["__import__"] = &PyBuiltinFunc{
		Name: "__import__",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			// __import__(name, globals=None, locals=None, fromlist=(), level=0)
			if len(args) < 1 {
				return nil, fmt.Errorf("TypeError: __import__() missing required argument: 'name'")
			}
			nameStr, ok := args[0].(*PyString)
			if !ok {
				return nil, fmt.Errorf("TypeError: __import__() argument 1 must be str, not %s", vm.typeName(args[0]))
			}
			moduleName := nameStr.Value

			// Extract globals (arg 1 or kwarg) for relative import resolution
			var globalsDict map[string]Value
			if len(args) > 1 {
				if d, ok := args[1].(*PyDict); ok {
					globalsDict = make(map[string]Value)
					for k, v := range d.Items {
						if ks, ok := k.(*PyString); ok {
							globalsDict[ks.Value] = v
						}
					}
				}
			}
			if globalsDict == nil && vm.frame != nil {
				globalsDict = vm.frame.Globals
			}

			// Extract fromlist (arg 3 or kwarg)
			var fromlistItems []string
			var fromlistVal Value
			if len(args) > 3 {
				fromlistVal = args[3]
			}
			if v, ok := kwargs["fromlist"]; ok {
				fromlistVal = v
			}
			if fromlistVal != nil && fromlistVal != None {
				switch fl := fromlistVal.(type) {
				case *PyTuple:
					for _, item := range fl.Items {
						if s, ok := item.(*PyString); ok {
							fromlistItems = append(fromlistItems, s.Value)
						}
					}
				case *PyList:
					for _, item := range fl.Items {
						if s, ok := item.(*PyString); ok {
							fromlistItems = append(fromlistItems, s.Value)
						}
					}
				}
			}

			// Extract level (arg 4 or kwarg)
			level := 0
			if len(args) > 4 {
				if li, ok := args[4].(*PyInt); ok {
					level = int(li.Value)
				}
			}
			if v, ok := kwargs["level"]; ok {
				if li, ok := v.(*PyInt); ok {
					level = int(li.Value)
				}
			}

			// Resolve relative imports
			resolvedName := moduleName
			if level > 0 {
				packageName := ""
				if globalsDict != nil {
					if pkgVal, ok := globalsDict["__package__"]; ok {
						if pkgStr, ok := pkgVal.(*PyString); ok {
							packageName = pkgStr.Value
						}
					}
					if packageName == "" {
						if nameVal, ok := globalsDict["__name__"]; ok {
							if nameStr, ok := nameVal.(*PyString); ok {
								packageName = nameStr.Value
							}
						}
					}
				}
				resolved, err := ResolveRelativeImport(moduleName, level, packageName)
				if err != nil {
					return nil, err
				}
				resolvedName = resolved
			}

			// Import each part of dotted name
			var rootMod, targetMod *PyModule
			parts := splitModuleName(resolvedName)
			for i := range parts {
				partialName := joinModuleName(parts[:i+1])
				mod, err := vm.ImportModule(partialName)
				if err != nil {
					return nil, err
				}
				if i == 0 {
					rootMod = mod
				}
				targetMod = mod
			}

			// fromlist non-empty → return target module; empty → return root module
			if len(fromlistItems) > 0 {
				return targetMod, nil
			}
			return rootMod, nil
		},
	}

	vm.builtins["None"] = None
	vm.builtins["True"] = True
	vm.builtins["False"] = False
	vm.builtins["NotImplemented"] = NotImplemented

	// __build_class__ is used to create classes
	vm.builtins["__build_class__"] = &PyBuiltinFunc{
		Name: "__build_class__",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("__build_class__: not enough arguments")
			}

			// First arg is the class body function
			bodyFunc, ok := args[0].(*PyFunction)
			if !ok {
				return nil, fmt.Errorf("__build_class__: first argument must be a function")
			}

			// Second arg is the class name
			nameVal, ok := args[1].(*PyString)
			if !ok {
				return nil, fmt.Errorf("__build_class__: second argument must be a string")
			}
			className := nameVal.Value

			// Remaining args are base classes — resolve __mro_entries__ for non-class bases
			originalBases := args[2:]
			var bases []*PyClass
			for _, baseArg := range originalBases {
				if base, ok := baseArg.(*PyClass); ok {
					bases = append(bases, base)
					continue
				}
				// Try __mro_entries__ for non-class bases (e.g. GenericAlias)
				origTuple := &PyTuple{Items: make([]Value, len(originalBases))}
				copy(origTuple.Items, originalBases)
				if mroEntries, err := vm.getAttr(baseArg, "__mro_entries__"); err == nil {
					if result, callErr := vm.call(mroEntries, []Value{origTuple}, nil); callErr == nil {
						if tup, ok := result.(*PyTuple); ok {
							for _, entry := range tup.Items {
								if cls, ok := entry.(*PyClass); ok {
									bases = append(bases, cls)
								}
							}
							continue
						}
						if lst, ok := result.(*PyList); ok {
							for _, entry := range lst.Items {
								if cls, ok := entry.(*PyClass); ok {
									bases = append(bases, cls)
								}
							}
							continue
						}
					}
				}
				// If __mro_entries__ not found or failed, skip non-class base
			}

			// If no bases specified, implicitly inherit from object (Python 3 behavior)
			objectClass := vm.builtins["object"].(*PyClass)
			if len(bases) == 0 {
				bases = []*PyClass{objectClass}
			}

			// Execute the class body to get the namespace and cells
			classDict, cells, orderedKeys, err := vm.callClassBody(bodyFunc)
			if err != nil {
				return nil, fmt.Errorf("__build_class__: error executing class body: %w", err)
			}

			// Check for metaclass kwarg
			typeClass := vm.builtins["type"].(*PyClass)
			var metaclass *PyClass
			if mc, ok := kwargs["metaclass"]; ok {
				if mcClass, ok := mc.(*PyClass); ok {
					metaclass = mcClass
				}
			}

			// If no explicit metaclass, infer from bases
			if metaclass == nil {
				for _, base := range bases {
					if base.Metaclass != nil && base.Metaclass != typeClass {
						if metaclass == nil {
							metaclass = base.Metaclass
						} else {
							// Check if new metaclass is subclass of current
							for _, m := range base.Metaclass.Mro {
								if m == metaclass {
									metaclass = base.Metaclass
									break
								}
							}
						}
					}
				}
			}

			var class *PyClass

			if metaclass != nil && metaclass != typeClass {
				// Metaclass-based class creation: call metaclass.__new__ and __init__

				// Convert to Python values for metaclass methods
				basesItems := make([]Value, len(bases))
				for i, b := range bases {
					basesItems[i] = b
				}
				basesTuple := &PyTuple{Items: basesItems}
				nsDict := &PyDict{Items: make(map[Value]Value), buckets: make(map[uint64][]dictEntry)}
				// Use ordered keys to preserve class body definition order
				for _, k := range orderedKeys {
					if v, ok := classDict[k]; ok {
						nsDict.DictSet(&PyString{Value: k}, v, vm)
					}
				}
				// Add any remaining entries not in orderedKeys
				for k, v := range classDict {
					found := false
					for _, ok := range orderedKeys {
						if ok == k {
							found = true
							break
						}
					}
					if !found {
						nsDict.DictSet(&PyString{Value: k}, v, vm)
					}
				}
				nameStr := &PyString{Value: className}

				// Call metaclass.__new__(mcs, name, bases, namespace) via MRO
				var newResult Value
				for _, cls := range metaclass.Mro {
					if newMethod, ok := cls.Dict["__new__"]; ok {
						newArgs := []Value{metaclass, nameStr, basesTuple, nsDict}
						switch nm := newMethod.(type) {
						case *PyFunction:
							newResult, err = vm.callFunction(nm, newArgs, kwargs)
						case *PyBuiltinFunc:
							newResult, err = nm.Fn(newArgs, kwargs)
						case *PyStaticMethod:
							newResult, err = vm.call(nm.Func, newArgs, kwargs)
						}
						if err != nil {
							return nil, err
						}
						break
					}
				}

				if newResult == nil {
					return nil, fmt.Errorf("TypeError: metaclass __new__ did not return a value")
				}

				if cls, ok := newResult.(*PyClass); ok {
					class = cls
					class.Metaclass = metaclass
					// Extract __slots__ if defined
					if class.Slots == nil {
						slots := extractSlots(class.Dict, bases)
						if slots != nil {
							class.Slots = slots
						}
					}

					// Call metaclass.__init__(cls, name, bases, namespace) via MRO
					for _, mroClass := range metaclass.Mro {
						if initMethod, ok := mroClass.Dict["__init__"]; ok {
							initArgs := []Value{class, nameStr, basesTuple, nsDict}
							switch im := initMethod.(type) {
							case *PyFunction:
								_, err = vm.callFunction(im, initArgs, kwargs)
							case *PyBuiltinFunc:
								_, err = im.Fn(initArgs, kwargs)
							}
							if err != nil {
								return nil, err
							}
							break
						}
					}
				} else {
					// If __new__ didn't return a *PyClass, just return it
					return newResult, nil
				}
			} else {
				// Standard class creation (no custom metaclass)
				slots := extractSlots(classDict, bases)
				class = &PyClass{
					Name:      className,
					Bases:     bases,
					Dict:      classDict,
					Metaclass: typeClass,
					Slots:     slots,
				}

				// Build MRO using C3 linearization for proper multiple inheritance
				mro, err := vm.ComputeC3MRO(class, bases)
				if err != nil {
					return nil, err
				}
				class.Mro = mro

				// Check if this class should use ABC abstract method checking
				for _, base := range bases {
					if base.IsABC {
						class.IsABC = true
						break
					}
				}

				// Collect abstract methods if this is an ABC class
				if class.IsABC {
					abstractMethods := make(map[string]bool)
					// Scan MRO (excluding current class) for abstract methods
					for _, cls := range mro[1:] {
						for name, val := range cls.Dict {
							if isAbstractValue(val) {
								abstractMethods[name] = true
							}
						}
					}
					// Scan current class: abstract methods add, concrete methods remove
					for name, val := range classDict {
						if isAbstractValue(val) {
							abstractMethods[name] = true
						} else {
							delete(abstractMethods, name)
						}
					}
					// Store as a PySet of strings for the instantiation guard
					if len(abstractMethods) > 0 {
						items := make([]Value, 0, len(abstractMethods))
						for name := range abstractMethods {
							items = append(items, &PyString{Value: name})
						}
						class.Dict["__abstractmethods__"] = &PyList{Items: items}
					}

					// Inject register() method for ABC classes (if not already defined)
					if _, hasRegister := class.Dict["register"]; !hasRegister {
						thisClass := class // capture for closure
						class.Dict["register"] = &PyBuiltinFunc{
							Name: "register",
							Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
								if len(args) != 1 {
									return nil, fmt.Errorf("TypeError: register() takes exactly 1 argument (%d given)", len(args))
								}
								subcls, ok := args[0].(*PyClass)
								if !ok {
									return nil, fmt.Errorf("TypeError: register() argument must be a class")
								}
								for _, existing := range thisClass.RegisteredSubclasses {
									if existing == subcls {
										return subcls, nil
									}
								}
								thisClass.RegisteredSubclasses = append(thisClass.RegisteredSubclasses, subcls)
								return subcls, nil
							},
						}
					}
				}
			}

			// Call __set_name__ on descriptors in the class dict
			if err := vm.callSetName(class); err != nil {
				return nil, err
			}

			// Call __init_subclass__ on parent classes
			if err := vm.callInitSubclass(class, kwargs); err != nil {
				return nil, err
			}

			// Populate the __class__ cell if present (for zero-argument super() support)
			// The __class__ cell is created by the compiler when methods use super()
			for i, cellName := range bodyFunc.Code.CellVars {
				if cellName == "__class__" && i < len(cells) && cells[i] != nil {
					cells[i].Value = class
					break
				}
			}

			return class, nil
		},
	}

	// property() creates a property descriptor
	vm.builtins["property"] = &PyBuiltinFunc{
		Name: "property",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			prop := &PyProperty{}

			// Handle positional args: property(fget=None, fset=None, fdel=None, doc=None)
			if len(args) > 0 && args[0] != None {
				prop.Fget = args[0]
			}
			if len(args) > 1 && args[1] != None {
				prop.Fset = args[1]
			}
			if len(args) > 2 && args[2] != None {
				prop.Fdel = args[2]
			}
			if len(args) > 3 {
				if s, ok := args[3].(*PyString); ok {
					prop.Doc = s.Value
				}
			}

			// Handle keyword args
			if fget, ok := kwargs["fget"]; ok && fget != None {
				prop.Fget = fget
			}
			if fset, ok := kwargs["fset"]; ok && fset != None {
				prop.Fset = fset
			}
			if fdel, ok := kwargs["fdel"]; ok && fdel != None {
				prop.Fdel = fdel
			}
			if doc, ok := kwargs["doc"]; ok {
				if s, ok := doc.(*PyString); ok {
					prop.Doc = s.Value
				}
			}

			return prop, nil
		},
	}

	// classmethod() wraps a function to bind the class as first argument
	vm.builtins["classmethod"] = &PyBuiltinFunc{
		Name: "classmethod",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("classmethod expected 1 argument, got %d", len(args))
			}
			return &PyClassMethod{Func: args[0]}, nil
		},
	}

	// staticmethod() wraps a function to prevent binding
	vm.builtins["staticmethod"] = &PyBuiltinFunc{
		Name: "staticmethod",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("staticmethod expected 1 argument, got %d", len(args))
			}
			return &PyStaticMethod{Func: args[0]}, nil
		},
	}

	// super() returns a proxy object for MRO-based method lookup
	vm.builtins["super"] = &PyBuiltinFunc{
		Name: "super",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			var thisClass *PyClass
			var instance Value

			if len(args) == 0 {
				// Zero-argument form: super()
				// Need to find __class__ from the enclosing scope and self from first arg
				// Look up the call stack to find the method context
				// We need to look at the caller's frame, not the current one (which is the builtin call)
				callerFrame := vm.frame
				if len(vm.frames) >= 1 {
					// The frames stack contains previous frames; vm.frame is current
					// For builtin calls, we want the calling frame
					callerFrame = vm.frame
				}
				if callerFrame != nil && callerFrame.Code != nil {
					// Try to get __class__ from closure cells
					for i, name := range callerFrame.Code.FreeVars {
						if name == "__class__" && i < len(callerFrame.Cells) {
							if cls, ok := callerFrame.Cells[i].Value.(*PyClass); ok {
								thisClass = cls
							}
						}
					}
					// Try to get self from first local variable (slot 0)
					if len(callerFrame.Code.VarNames) > 0 && len(callerFrame.Locals) > 0 {
						instance = callerFrame.Locals[0]
					}
				}
				if thisClass == nil {
					return nil, fmt.Errorf("super(): __class__ cell not found")
				}
				if instance == nil {
					return nil, fmt.Errorf("super(): self argument not found")
				}
			} else if len(args) == 2 {
				// Two-argument form: super(type, object-or-type)
				var ok bool
				thisClass, ok = args[0].(*PyClass)
				if !ok {
					return nil, fmt.Errorf("super() argument 1 must be type, not %s", vm.typeName(args[0]))
				}
				instance = args[1]
			} else if len(args) == 1 {
				// One-argument form: super(type) - unbound super
				var ok bool
				thisClass, ok = args[0].(*PyClass)
				if !ok {
					return nil, fmt.Errorf("super() argument 1 must be type, not %s", vm.typeName(args[0]))
				}
				instance = nil
			} else {
				return nil, fmt.Errorf("super() takes 0, 1, or 2 arguments (%d given)", len(args))
			}

			// Find the index of thisClass in the MRO of the instance's class
			var mro []*PyClass
			if inst, ok := instance.(*PyInstance); ok {
				mro = inst.Class.Mro
			} else if cls, ok := instance.(*PyClass); ok {
				// For class instances, check if thisClass is in the metaclass MRO
				// (used when super() is called inside a metaclass method)
				useMetaMro := false
				if cls.Metaclass != nil {
					for _, mc := range cls.Metaclass.Mro {
						if mc == thisClass {
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
			} else if instance != nil {
				return nil, fmt.Errorf("super(type, obj): obj must be an instance or subtype of type")
			}

			startIdx := 0
			if mro != nil {
				for i, cls := range mro {
					if cls == thisClass {
						startIdx = i + 1 // Start searching from the next class in MRO
						break
					}
				}
			}

			return &PySuper{
				ThisClass: thisClass,
				Instance:  instance,
				StartIdx:  startIdx,
			}, nil
		},
	}

	// object is the base class for all classes
	vm.builtins["object"] = &PyClass{
		Name:  "object",
		Bases: nil,
		Dict:  make(map[string]Value),
		Mro:   nil,
	}
	// Set object's MRO to just itself
	objectClass := vm.builtins["object"].(*PyClass)
	objectClass.Mro = []*PyClass{objectClass}

	// object.__getattribute__(self, name) - default attribute lookup (descriptor protocol)
	objectClass.Dict["__getattribute__"] = &PyBuiltinFunc{
		Name: "object.__getattribute__",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("object.__getattribute__() takes exactly 2 arguments (%d given)", len(args))
			}
			inst, ok := args[0].(*PyInstance)
			if !ok {
				return nil, fmt.Errorf("descriptor '__getattribute__' requires a 'object' instance")
			}
			name, ok := args[1].(*PyString)
			if !ok {
				return nil, fmt.Errorf("attribute name must be string, not '%s'", vm.typeName(args[1]))
			}
			return vm.defaultGetAttribute(inst, name.Value)
		},
	}

	// object.__setattr__(self, name, value) - direct instance dict assignment (bypasses user __setattr__)
	objectClass.Dict["__setattr__"] = &PyBuiltinFunc{
		Name: "object.__setattr__",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 3 {
				return nil, fmt.Errorf("object.__setattr__() takes exactly 3 arguments (%d given)", len(args))
			}
			inst, ok := args[0].(*PyInstance)
			if !ok {
				return nil, fmt.Errorf("descriptor '__setattr__' requires a 'object' instance")
			}
			name, ok := args[1].(*PyString)
			if !ok {
				return nil, fmt.Errorf("attribute name must be string, not '%s'", vm.typeName(args[1]))
			}
			// Respect property setters in MRO
			for _, cls := range inst.Class.Mro {
				if clsVal, ok := cls.Dict[name.Value]; ok {
					if prop, ok := clsVal.(*PyProperty); ok {
						if prop.Fset == nil {
							return nil, fmt.Errorf("property '%s' has no setter", name.Value)
						}
						_, err := vm.call(prop.Fset, []Value{inst, args[2]}, nil)
						if err != nil {
							return nil, err
						}
						return None, nil
					}
					break
				}
			}
			if inst.Slots != nil {
				if !isValidSlot(inst.Class, name.Value) {
					return nil, fmt.Errorf("AttributeError: '%s' object has no attribute '%s'", inst.Class.Name, name.Value)
				}
				inst.Slots[name.Value] = args[2]
			} else {
				inst.Dict[name.Value] = args[2]
			}
			return None, nil
		},
	}

	// object.__delattr__(self, name) - direct instance dict deletion (bypasses user __delattr__)
	objectClass.Dict["__delattr__"] = &PyBuiltinFunc{
		Name: "object.__delattr__",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("object.__delattr__() takes exactly 2 arguments (%d given)", len(args))
			}
			inst, ok := args[0].(*PyInstance)
			if !ok {
				return nil, fmt.Errorf("descriptor '__delattr__' requires a 'object' instance")
			}
			name, ok := args[1].(*PyString)
			if !ok {
				return nil, fmt.Errorf("attribute name must be string, not '%s'", vm.typeName(args[1]))
			}
			// Respect property deleters and custom descriptor __delete__ in MRO
			for _, cls := range inst.Class.Mro {
				if clsVal, ok := cls.Dict[name.Value]; ok {
					if prop, ok := clsVal.(*PyProperty); ok {
						if prop.Fdel == nil {
							return nil, fmt.Errorf("property '%s' has no deleter", name.Value)
						}
						_, err := vm.call(prop.Fdel, []Value{inst}, nil)
						if err != nil {
							return nil, err
						}
						return None, nil
					}
					// Check for custom descriptor with __delete__
					if descInst, ok := clsVal.(*PyInstance); ok {
						if _, found, err := vm.callDunder(descInst, "__delete__", inst); found {
							if err != nil {
								return nil, err
							}
							return None, nil
						}
					}
					break
				}
			}
			if inst.Slots != nil {
				if _, exists := inst.Slots[name.Value]; !exists {
					return nil, fmt.Errorf("AttributeError: '%s' object has no attribute '%s'", inst.Class.Name, name.Value)
				}
				delete(inst.Slots, name.Value)
			} else {
				if _, exists := inst.Dict[name.Value]; !exists {
					return nil, fmt.Errorf("AttributeError: '%s' object has no attribute '%s'", inst.Class.Name, name.Value)
				}
				delete(inst.Dict, name.Value)
			}
			return None, nil
		},
	}

	// object.__sizeof__(self)
	objectClass.Dict["__sizeof__"] = &PyBuiltinFunc{
		Name: "object.__sizeof__",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("descriptor '__sizeof__' requires an argument")
			}
			inst, ok := args[0].(*PyInstance)
			if !ok {
				return MakeInt(64), nil
			}
			// Base size for the instance struct
			var size int64 = 64
			if inst.Dict != nil {
				// 8 bytes per key-value pair estimate
				size += int64(len(inst.Dict) * 16)
			}
			if inst.Slots != nil {
				size += int64(len(inst.Slots) * 16)
			}
			return MakeInt(size), nil
		},
	}

	objectClass.Dict["__init_subclass__"] = &PyClassMethod{Func: &PyBuiltinFunc{
		Name: "__init_subclass__",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			return None, nil
		},
	}}

	// object.__new__(cls) - create a new instance of cls
	objectClass.Dict["__new__"] = &PyBuiltinFunc{
		Name: "object.__new__",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("object.__new__(): not enough arguments")
			}
			cls, ok := args[0].(*PyClass)
			if !ok {
				return nil, fmt.Errorf("object.__new__(X): X is not a type object (%s)", vm.typeName(args[0]))
			}
			if cls.Slots != nil {
				return &PyInstance{
					Class: cls,
					Slots: make(map[string]Value),
				}, nil
			}
			return &PyInstance{
				Class: cls,
				Dict:  make(map[string]Value),
			}, nil
		},
	}

	// type is a proper PyClass (the metaclass of all classes)
	typeClass := &PyClass{
		Name:  "type",
		Bases: []*PyClass{objectClass},
		Dict:  make(map[string]Value),
	}
	typeClass.Mro = []*PyClass{typeClass, objectClass}
	vm.builtins["type"] = typeClass

	// type.__new__(mcs, name_or_obj, bases, namespace) - static method
	typeClass.Dict["__new__"] = &PyStaticMethod{Func: &PyBuiltinFunc{
		Name: "type.__new__",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			// 2-arg form: type.__new__(type, x) -> type of x (called as type(x))
			if len(args) == 2 {
				switch v := args[1].(type) {
				case *PyInstance:
					return v.Class, nil
				case *PyClass:
					return typeClass, nil
				default:
					// Return a class with the type name
					typeName := vm.typeName(args[1])
					cls := &PyClass{Name: typeName}
					cls.Mro = []*PyClass{cls}
					return cls, nil
				}
			}
			// 4-arg form: type.__new__(mcs, name, bases, namespace)
			if len(args) == 4 {
				mcs, ok := args[0].(*PyClass)
				if !ok {
					return nil, fmt.Errorf("TypeError: type.__new__(X): X is not a type object (%s)", vm.typeName(args[0]))
				}
				nameStr, ok := args[1].(*PyString)
				if !ok {
					return nil, fmt.Errorf("TypeError: type.__new__() argument 1 must be str, not %s", vm.typeName(args[1]))
				}
				basesTuple, ok := args[2].(*PyTuple)
				if !ok {
					return nil, fmt.Errorf("TypeError: type.__new__() argument 2 must be tuple, not %s", vm.typeName(args[2]))
				}
				nsDict, ok := args[3].(*PyDict)
				if !ok {
					return nil, fmt.Errorf("TypeError: type.__new__() argument 3 must be dict, not %s", vm.typeName(args[3]))
				}

				// Convert bases tuple to []*PyClass
				var bases []*PyClass
				for _, b := range basesTuple.Items {
					if bc, ok := b.(*PyClass); ok {
						bases = append(bases, bc)
					}
				}
				if len(bases) == 0 {
					bases = []*PyClass{objectClass}
				}

				// Convert namespace dict to map[string]Value
				classDict := make(map[string]Value)
				for k, v := range nsDict.Items {
					if ks, ok := k.(*PyString); ok {
						classDict[ks.Value] = v
					}
				}

				slots := extractSlots(classDict, bases)
				cls := &PyClass{
					Name:      nameStr.Value,
					Bases:     bases,
					Dict:      classDict,
					Metaclass: mcs,
					Slots:     slots,
				}

				// Compute C3 MRO
				mro, err := vm.ComputeC3MRO(cls, bases)
				if err != nil {
					return nil, err
				}
				cls.Mro = mro

				// Handle ABC abstract method tracking
				if mcs.IsABC {
					cls.IsABC = true
				}
				if !cls.IsABC {
					for _, base := range bases {
						if base.IsABC {
							cls.IsABC = true
							break
						}
					}
				}
				if cls.IsABC {
					abstractMethods := make(map[string]bool)
					for _, mroClass := range mro[1:] {
						for name, val := range mroClass.Dict {
							if isAbstractValue(val) {
								abstractMethods[name] = true
							}
						}
					}
					for name, val := range classDict {
						if isAbstractValue(val) {
							abstractMethods[name] = true
						} else {
							delete(abstractMethods, name)
						}
					}
					if len(abstractMethods) > 0 {
						items := make([]Value, 0, len(abstractMethods))
						for name := range abstractMethods {
							items = append(items, &PyString{Value: name})
						}
						cls.Dict["__abstractmethods__"] = &PyList{Items: items}
					}
				}

				// Call __set_name__ on descriptors
				if err := vm.callSetName(cls); err != nil {
					return nil, err
				}

				return cls, nil
			}
			return nil, fmt.Errorf("type.__new__() takes 2 or 4 arguments (%d given)", len(args))
		},
	}}

	// type.__init__(cls, name, bases, namespace) - no-op
	typeClass.Dict["__init__"] = &PyBuiltinFunc{
		Name: "type.__init__",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			return None, nil
		},
	}

	// type.__call__(cls, *args, **kwargs) - default class instantiation
	typeClass.Dict["__call__"] = &PyBuiltinFunc{
		Name: "type.__call__",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("TypeError: type.__call__() requires at least 1 argument")
			}
			cls, ok := args[0].(*PyClass)
			if !ok {
				return nil, fmt.Errorf("TypeError: descriptor '__call__' requires a 'type' object")
			}
			return vm.defaultClassCall(cls, args[1:], kwargs)
		},
	}
}

// ComputeC3MRO computes the Method Resolution Order using C3 linearization algorithm.
// This properly handles multiple inheritance and detects inconsistent hierarchies.
func (vm *VM) ComputeC3MRO(class *PyClass, bases []*PyClass) ([]*PyClass, error) {
	// Base case: no bases
	if len(bases) == 0 {
		return []*PyClass{class}, nil
	}

	// Collect linearizations of all bases plus the list of bases itself
	// We need to copy slices to avoid modifying the originals
	var toMerge [][]*PyClass
	for _, base := range bases {
		// Copy the base's MRO
		baseMRO := make([]*PyClass, len(base.Mro))
		copy(baseMRO, base.Mro)
		toMerge = append(toMerge, baseMRO)
	}
	// Add the list of direct bases
	basesCopy := make([]*PyClass, len(bases))
	copy(basesCopy, bases)
	toMerge = append(toMerge, basesCopy)

	// Start with the class itself
	result := []*PyClass{class}

	// Merge until all lists are empty
	for {
		// Remove empty lists
		var nonEmpty [][]*PyClass
		for _, list := range toMerge {
			if len(list) > 0 {
				nonEmpty = append(nonEmpty, list)
			}
		}
		toMerge = nonEmpty

		if len(toMerge) == 0 {
			break
		}

		// Find a good head: a class that is not in the tail of any list
		var candidate *PyClass
		for _, list := range toMerge {
			head := list[0]
			inTail := false
			for _, other := range toMerge {
				// Check if head appears in the tail (positions 1+) of other
				for i := 1; i < len(other); i++ {
					if other[i] == head {
						inTail = true
						break
					}
				}
				if inTail {
					break
				}
			}
			if !inTail {
				candidate = head
				break
			}
		}

		if candidate == nil {
			// No valid candidate found - inconsistent hierarchy
			msg := fmt.Sprintf("Cannot create a consistent method resolution order (MRO) for bases %s",
				vm.formatBases(bases))
			return nil, &PyException{
				ExcType:  vm.builtins["TypeError"].(*PyClass),
				Args:     &PyTuple{Items: []Value{&PyString{Value: msg}}},
				Message:  "TypeError: " + msg,
				TypeName: "TypeError",
			}
		}

		// Add candidate to result
		result = append(result, candidate)

		// Remove candidate from the head of all lists where it appears
		for i := range toMerge {
			if len(toMerge[i]) > 0 && toMerge[i][0] == candidate {
				toMerge[i] = toMerge[i][1:]
			}
		}
	}

	return result, nil
}

// sortedNameList converts a set of names to a sorted PyList of PyStrings.
func (vm *VM) sortedNameList(names map[string]bool) *PyList {
	sorted := make([]string, 0, len(names))
	for k := range names {
		sorted = append(sorted, k)
	}
	sort.Strings(sorted)
	items := make([]Value, len(sorted))
	for i, s := range sorted {
		items[i] = &PyString{Value: s}
	}
	return &PyList{Items: items}
}

// isAbstractValue checks if a value is marked as abstract
func isAbstractValue(v Value) bool {
	switch val := v.(type) {
	case *PyFunction:
		return val.IsAbstract
	case *PyProperty:
		if fn, ok := val.Fget.(*PyFunction); ok {
			return fn.IsAbstract
		}
	case *PyClassMethod:
		if fn, ok := val.Func.(*PyFunction); ok {
			return fn.IsAbstract
		}
	case *PyStaticMethod:
		if fn, ok := val.Func.(*PyFunction); ok {
			return fn.IsAbstract
		}
	}
	return false
}

// formatBases formats a list of base classes for error messages
func (vm *VM) formatBases(bases []*PyClass) string {
	if len(bases) == 0 {
		return ""
	}
	names := make([]string, len(bases))
	for i, b := range bases {
		names[i] = b.Name
	}
	result := names[0]
	for i := 1; i < len(names); i++ {
		result += ", " + names[i]
	}
	return result
}

// callInitSubclass calls __init_subclass__ on the parent class after a new class is created.
// It walks the MRO starting from index 1 (skipping the new class itself) to find the hook.
// callSetName iterates through the class dict and calls __set_name__(owner, name)
// on any descriptor that defines it. Called during class creation.
func (vm *VM) callSetName(class *PyClass) error {
	for name, val := range class.Dict {
		if inst, ok := val.(*PyInstance); ok {
			_, found, err := vm.callDunder(inst, "__set_name__", class, &PyString{Value: name})
			if err != nil {
				return fmt.Errorf("RuntimeError: __set_name__ of '%s' descriptor '%s' raised: %w",
					inst.Class.Name, name, err)
			}
			_ = found
		}
	}
	return nil
}

func (vm *VM) callInitSubclass(class *PyClass, kwargs map[string]Value) error {
	// Filter out "metaclass" from kwargs
	var filteredKwargs map[string]Value
	if len(kwargs) > 0 {
		filteredKwargs = make(map[string]Value, len(kwargs))
		for k, v := range kwargs {
			if k != "metaclass" {
				filteredKwargs[k] = v
			}
		}
	}

	// Walk MRO starting from index 1 (skip the new class itself)
	for i := 1; i < len(class.Mro); i++ {
		if method, ok := class.Mro[i].Dict["__init_subclass__"]; ok {
			args := []Value{class}
			var err error
			switch m := method.(type) {
			case *PyClassMethod:
				_, err = vm.call(m.Func, args, filteredKwargs)
			case *PyFunction:
				_, err = vm.callFunction(m, args, filteredKwargs)
			case *PyBuiltinFunc:
				_, err = m.Fn(args, filteredKwargs)
			}
			if err != nil {
				return err
			}
			break
		}
	}
	return nil
}
