package runtime

// isInstanceOf checks if an instance is an instance of a class (including subclasses)
func (vm *VM) isInstanceOf(inst *PyInstance, cls *PyClass) bool {
	// Direct class match
	if inst.Class == cls {
		return true
	}
	// Check base classes recursively
	return vm.isSubclassOf(inst.Class, cls)
}

// isSubclassOf checks if cls is a subclass of target
func (vm *VM) isSubclassOf(cls, target *PyClass) bool {
	if cls == target {
		return true
	}
	for _, base := range cls.Bases {
		if vm.isSubclassOf(base, target) {
			return true
		}
	}
	return false
}

// hasMethod checks if an instance has a method via MRO lookup.
func (vm *VM) hasMethod(instance *PyInstance, name string) bool {
	for _, cls := range instance.Class.Mro {
		if _, ok := cls.Dict[name]; ok {
			return true
		}
	}
	return false
}

// callDunder looks up and calls a dunder method on an instance via MRO.
// Returns (result, found, error) - found is true if the method exists.
func (vm *VM) callDunder(instance *PyInstance, name string, args ...Value) (Value, bool, error) {
	// Search MRO for the method
	if len(instance.Class.Mro) > 0 {
		for _, cls := range instance.Class.Mro {
			if method, ok := cls.Dict[name]; ok {
				// Prepend instance as self
				allArgs := append([]Value{instance}, args...)
				switch fn := method.(type) {
				case *PyFunction:
					result, err := vm.callFunction(fn, allArgs, nil)
					return result, true, err
				case *PyBuiltinFunc:
					result, err := fn.Fn(allArgs, nil)
					return result, true, err
				}
			}
		}
	} else {
		// Fallback: check instance's class dict directly if MRO is empty
		if method, ok := instance.Class.Dict[name]; ok {
			allArgs := append([]Value{instance}, args...)
			switch fn := method.(type) {
			case *PyFunction:
				result, err := vm.callFunction(fn, allArgs, nil)
				return result, true, err
			case *PyBuiltinFunc:
				result, err := fn.Fn(allArgs, nil)
				return result, true, err
			}
		}
	}
	return nil, false, nil
}

// callDel calls __del__ on a PyInstance if the method exists in its MRO.
// Errors from __del__ are silently ignored (matching CPython behavior).
func (vm *VM) callDel(val Value) {
	inst, ok := val.(*PyInstance)
	if !ok {
		return
	}
	for _, cls := range inst.Class.Mro {
		if method, ok := cls.Dict["__del__"]; ok {
			args := []Value{inst}
			switch fn := method.(type) {
			case *PyFunction:
				vm.callFunction(fn, args, nil)
			case *PyBuiltinFunc:
				fn.Fn(args, nil)
			}
			return
		}
	}
}
