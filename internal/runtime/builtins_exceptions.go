package runtime

import (
	"fmt"
	"strings"
)

// initExceptionClasses sets up the exception class hierarchy
func (vm *VM) initExceptionClasses() {
	// BaseException is the root of all exceptions
	baseException := &PyClass{
		Name:  "BaseException",
		Bases: nil,
		Dict:  make(map[string]Value),
	}
	baseException.Mro = []*PyClass{baseException}
	// add_note(note) — appends a note string to __notes__
	baseException.Dict["add_note"] = &PyBuiltinFunc{
		Name: "BaseException.add_note",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("add_note() missing required argument: 'note'")
			}
			self, ok := args[0].(*PyInstance)
			if !ok {
				return nil, fmt.Errorf("add_note() requires an exception instance")
			}
			note, ok := args[1].(*PyString)
			if !ok {
				return nil, fmt.Errorf("TypeError: note must be a str, not '%s'", vm.typeName(args[1]))
			}
			notes, exists := self.Dict["__notes__"]
			if !exists {
				notesList := &PyList{Items: []Value{note}}
				self.Dict["__notes__"] = notesList
			} else if notesList, ok := notes.(*PyList); ok {
				notesList.Items = append(notesList.Items, note)
			}
			return None, nil
		},
	}
	vm.builtins["BaseException"] = baseException

	// Exception inherits from BaseException (most exceptions derive from this)
	exception := &PyClass{
		Name:  "Exception",
		Bases: []*PyClass{baseException},
		Dict:  make(map[string]Value),
	}
	exception.Mro = []*PyClass{exception, baseException}
	vm.builtins["Exception"] = exception

	// Helper to create exception class inheriting from Exception
	makeExc := func(name string, parent *PyClass) *PyClass {
		cls := &PyClass{
			Name:  name,
			Bases: []*PyClass{parent},
			Dict:  make(map[string]Value),
		}
		// Build MRO by prepending self to parent's MRO
		cls.Mro = append([]*PyClass{cls}, parent.Mro...)
		vm.builtins[name] = cls
		return cls
	}

	// Standard exceptions inheriting from Exception
	makeExc("ValueError", exception)
	makeExc("TypeError", exception)
	attrError := makeExc("AttributeError", exception)
	makeExc("FrozenInstanceError", attrError)
	nameError := makeExc("NameError", exception)
	makeExc("UnboundLocalError", nameError)
	makeExc("RuntimeError", exception)
	makeExc("AssertionError", exception)
	makeExc("StopIteration", exception)
	makeExc("NotImplementedError", exception)
	makeExc("RecursionError", exception)
	makeExc("MemoryError", exception)
	makeExc("SyntaxError", exception)
	makeExc("EOFError", exception)
	makeExc("BufferError", exception)

	// LookupError and its subclasses
	lookupError := makeExc("LookupError", exception)
	makeExc("KeyError", lookupError)
	makeExc("IndexError", lookupError)

	// ArithmeticError and its subclasses
	arithmeticError := makeExc("ArithmeticError", exception)
	makeExc("ZeroDivisionError", arithmeticError)
	makeExc("OverflowError", arithmeticError)
	makeExc("FloatingPointError", arithmeticError)

	// OSError and its subclasses
	osError := makeExc("OSError", exception)
	makeExc("FileNotFoundError", osError)
	makeExc("PermissionError", osError)
	makeExc("FileExistsError", osError)
	makeExc("IOError", osError) // IOError is an alias for OSError in Python 3
	makeExc("TimeoutError", osError)
	connError := makeExc("ConnectionError", osError)
	makeExc("ConnectionRefusedError", connError)
	makeExc("ConnectionResetError", connError)
	makeExc("ConnectionAbortedError", connError)
	makeExc("BrokenPipeError", connError)
	makeExc("IsADirectoryError", osError)
	makeExc("NotADirectoryError", osError)
	makeExc("InterruptedError", osError)
	makeExc("BlockingIOError", osError)
	makeExc("ChildProcessError", osError)
	makeExc("ProcessLookupError", osError)

	// ImportError and its subclass
	importError := makeExc("ImportError", exception)
	makeExc("ModuleNotFoundError", importError)

	// UnicodeError and its subclasses (under ValueError)
	valError := vm.builtins["ValueError"].(*PyClass)
	unicodeError := makeExc("UnicodeError", valError)
	makeExc("UnicodeDecodeError", unicodeError)
	makeExc("UnicodeEncodeError", unicodeError)
	makeExc("UnicodeTranslateError", unicodeError)

	// Warning hierarchy
	warning := makeExc("Warning", exception)
	makeExc("DeprecationWarning", warning)
	makeExc("PendingDeprecationWarning", warning)
	makeExc("RuntimeWarning", warning)
	makeExc("SyntaxWarning", warning)
	makeExc("UserWarning", warning)
	makeExc("FutureWarning", warning)
	makeExc("ImportWarning", warning)
	makeExc("UnicodeWarning", warning)
	makeExc("BytesWarning", warning)
	makeExc("ResourceWarning", warning)
	makeExc("EncodingWarning", warning)

	// BaseException subclasses (not Exception)
	makeExc("GeneratorExit", baseException)
	makeExc("SystemExit", baseException)
	makeExc("KeyboardInterrupt", baseException)
	makeExc("StopAsyncIteration", baseException)

	// BaseExceptionGroup inherits from BaseException
	baseExcGroup := &PyClass{
		Name:  "BaseExceptionGroup",
		Bases: []*PyClass{baseException},
		Dict:  make(map[string]Value),
	}
	baseExcGroup.Mro = []*PyClass{baseExcGroup, baseException}
	vm.builtins["BaseExceptionGroup"] = baseExcGroup

	// ExceptionGroup inherits from Exception and BaseExceptionGroup
	excGroup := &PyClass{
		Name:  "ExceptionGroup",
		Bases: []*PyClass{exception, baseExcGroup},
		Dict:  make(map[string]Value),
	}
	excGroup.Mro = []*PyClass{excGroup, exception, baseExcGroup, baseException}
	vm.builtins["ExceptionGroup"] = excGroup

	// Shared __init__ for both exception group classes
	egInit := &PyBuiltinFunc{Name: "ExceptionGroup.__init__", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("TypeError: ExceptionGroup.__init__() requires at least 2 arguments (message, exceptions)")
		}
		self := args[0]
		inst, ok := self.(*PyInstance)
		if !ok {
			return nil, fmt.Errorf("TypeError: ExceptionGroup.__init__() expected instance")
		}
		msgVal := args[1]
		msgStr, ok := msgVal.(*PyString)
		if !ok {
			return nil, fmt.Errorf("TypeError: ExceptionGroup message must be a string")
		}
		excsVal := args[2]
		var excItems []Value
		switch ev := excsVal.(type) {
		case *PyList:
			excItems = ev.Items
		case *PyTuple:
			excItems = ev.Items
		default:
			return nil, fmt.Errorf("TypeError: ExceptionGroup exceptions must be a list or tuple")
		}
		if len(excItems) == 0 {
			return nil, fmt.Errorf("ValueError: ExceptionGroup exceptions must be non-empty")
		}
		// Convert to []*PyException for VM use and validate
		pyExcs := make([]*PyException, len(excItems))
		tupleItems := make([]Value, len(excItems))
		for i, item := range excItems {
			switch e := item.(type) {
			case *PyException:
				pyExcs[i] = e
				tupleItems[i] = e
			case *PyInstance:
				if vm.isExceptionClass(e.Class) {
					pyExcs[i] = vm.createException(e, nil)
					tupleItems[i] = e
				} else {
					return nil, fmt.Errorf("TypeError: exceptions must be instances of BaseException")
				}
			default:
				return nil, fmt.Errorf("TypeError: exceptions must be instances of BaseException")
			}
		}
		inst.Dict["message"] = msgStr
		inst.Dict["exceptions"] = &PyTuple{Items: tupleItems}
		inst.Dict["args"] = &PyTuple{Items: []Value{msgStr}}
		inst.Dict["__eg_exceptions__"] = &pyExceptionList{items: pyExcs}
		return None, nil
	}}

	// __str__ for exception groups
	egStr := &PyBuiltinFunc{Name: "ExceptionGroup.__str__", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
		if len(args) < 1 {
			return &PyString{Value: "ExceptionGroup()"}, nil
		}
		inst, ok := args[0].(*PyInstance)
		if !ok {
			return &PyString{Value: "ExceptionGroup()"}, nil
		}
		msg := "ExceptionGroup"
		if m, ok := inst.Dict["message"].(*PyString); ok {
			msg = m.Value
		}
		excsTuple, _ := inst.Dict["exceptions"].(*PyTuple)
		count := 0
		if excsTuple != nil {
			count = len(excsTuple.Items)
		}
		sub := "sub-exception"
		if count != 1 {
			sub = "sub-exceptions"
		}
		return &PyString{Value: fmt.Sprintf("%s (%d %s)", msg, count, sub)}, nil
	}}

	// __repr__ for exception groups
	egRepr := &PyBuiltinFunc{Name: "ExceptionGroup.__repr__", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
		if len(args) < 1 {
			return &PyString{Value: "ExceptionGroup()"}, nil
		}
		inst, ok := args[0].(*PyInstance)
		if !ok {
			return &PyString{Value: "ExceptionGroup()"}, nil
		}
		msg := ""
		if m, ok := inst.Dict["message"].(*PyString); ok {
			msg = m.Value
		}
		excsTuple, _ := inst.Dict["exceptions"].(*PyTuple)
		className := "ExceptionGroup"
		if inst.Class != nil {
			className = inst.Class.Name
		}
		excReprs := "[]"
		if excsTuple != nil {
			parts := make([]string, len(excsTuple.Items))
			for i, e := range excsTuple.Items {
				parts[i] = vm.repr(e)
			}
			excReprs = "[" + strings.Join(parts, ", ") + "]"
		}
		return &PyString{Value: fmt.Sprintf("%s('%s', %s)", className, msg, excReprs)}, nil
	}}

	// subgroup(condition) — filter exceptions matching type
	egSubgroup := &PyBuiltinFunc{Name: "ExceptionGroup.subgroup", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("TypeError: subgroup() requires 1 argument")
		}
		inst, ok := args[0].(*PyInstance)
		if !ok {
			return None, nil
		}
		condition := args[1]
		egExcs := vm.getEGExceptions(inst)
		if egExcs == nil {
			return None, nil
		}
		var matched []*PyException
		for _, exc := range egExcs {
			if vm.exceptionMatches(exc, condition) {
				matched = append(matched, exc)
			}
		}
		if len(matched) == 0 {
			return None, nil
		}
		msg := ""
		if m, ok := inst.Dict["message"].(*PyString); ok {
			msg = m.Value
		}
		return vm.buildExceptionGroup(msg, matched, vm.isBaseExceptionGroup(inst.Class))
	}}

	// split(condition) — return (matched, rest) tuple
	egSplit := &PyBuiltinFunc{Name: "ExceptionGroup.split", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("TypeError: split() requires 1 argument")
		}
		inst, ok := args[0].(*PyInstance)
		if !ok {
			return &PyTuple{Items: []Value{None, None}}, nil
		}
		condition := args[1]
		egExcs := vm.getEGExceptions(inst)
		if egExcs == nil {
			return &PyTuple{Items: []Value{None, None}}, nil
		}
		var matched, rest []*PyException
		for _, exc := range egExcs {
			if vm.exceptionMatches(exc, condition) {
				matched = append(matched, exc)
			} else {
				rest = append(rest, exc)
			}
		}
		msg := ""
		if m, ok := inst.Dict["message"].(*PyString); ok {
			msg = m.Value
		}
		isBase := vm.isBaseExceptionGroup(inst.Class)
		var matchGroup, restGroup Value
		if len(matched) > 0 {
			matchGroup, _ = vm.buildExceptionGroup(msg, matched, isBase)
		} else {
			matchGroup = None
		}
		if len(rest) > 0 {
			restGroup, _ = vm.buildExceptionGroup(msg, rest, isBase)
		} else {
			restGroup = None
		}
		return &PyTuple{Items: []Value{matchGroup, restGroup}}, nil
	}}

	// derive(excs) — create new group of same class
	egDerive := &PyBuiltinFunc{Name: "ExceptionGroup.derive", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("TypeError: derive() requires 1 argument")
		}
		inst, ok := args[0].(*PyInstance)
		if !ok {
			return None, nil
		}
		excsVal := args[1]
		var excItems []Value
		switch ev := excsVal.(type) {
		case *PyList:
			excItems = ev.Items
		case *PyTuple:
			excItems = ev.Items
		default:
			return nil, fmt.Errorf("TypeError: derive() argument must be a list or tuple")
		}
		pyExcs := make([]*PyException, len(excItems))
		for i, item := range excItems {
			switch e := item.(type) {
			case *PyException:
				pyExcs[i] = e
			default:
				pyExcs[i] = vm.createException(e, nil)
			}
		}
		msg := ""
		if m, ok := inst.Dict["message"].(*PyString); ok {
			msg = m.Value
		}
		return vm.buildExceptionGroup(msg, pyExcs, vm.isBaseExceptionGroup(inst.Class))
	}}

	// Register methods on both classes
	for _, cls := range []*PyClass{baseExcGroup, excGroup} {
		cls.Dict["__init__"] = egInit
		cls.Dict["__str__"] = egStr
		cls.Dict["__repr__"] = egRepr
		cls.Dict["subgroup"] = egSubgroup
		cls.Dict["split"] = egSplit
		cls.Dict["derive"] = egDerive
	}
}
