package runtime

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Text conversion: repr(), str() and ascii().
//
// reprValue, strValue and asciiValue are the single implementation behind the
// repr/str/ascii builtins, print(), %-formatting, str.format() and f-strings.
// They propagate exceptions raised by user-defined __repr__/__str__ methods.
// The repr/str/ascii wrappers in conversions.go are for internal callers (error
// messages and the like) that cannot propagate errors.

// Text conversion recurses on the Go stack through nested containers and
// through user-defined __repr__/__str__ methods, which re-enter the VM and use
// much more stack per level. reprNestingBudget bounds the total so that a
// pathological input raises RecursionError (as CPython does) instead of
// overflowing the Go stack and killing the embedding process.
const (
	reprNestingBudget = 60000
	reprContainerCost = 1 // budget used by each nested container
	reprCallCost      = 8 // budget used by each nested __repr__/__str__ call
)

// reprEnter records that container v is being converted to text. It returns
// false when v is already being converted further up the call chain (a
// reference cycle), in which case the caller emits a "..." placeholder like
// CPython's Py_ReprEnter.
func (vm *VM) reprEnter(v Value) bool {
	if _, active := vm.reprActive[v]; active {
		return false
	}
	if vm.reprActive == nil {
		vm.reprActive = make(map[Value]struct{})
	}
	vm.reprActive[v] = struct{}{}
	return true
}

// reprLeave undoes a successful reprEnter.
func (vm *VM) reprLeave(v Value) {
	delete(vm.reprActive, v)
}

// enterReprNesting charges cost against the nesting budget, failing with
// RecursionError once it is exhausted. Callers must pair a nil return with
// leaveReprNesting(cost).
func (vm *VM) enterReprNesting(cost int) error {
	if vm.reprDepth+cost > reprNestingBudget {
		return fmt.Errorf("RecursionError: maximum recursion depth exceeded while getting the repr of an object")
	}
	vm.reprDepth += cost
	return nil
}

func (vm *VM) leaveReprNesting(cost int) {
	vm.reprDepth -= cost
}

// reprValue returns repr(v).
func (vm *VM) reprValue(v Value) (string, error) {
	switch val := v.(type) {
	case *PyNone:
		return "None", nil
	case *PyNotImplementedType:
		return "NotImplemented", nil
	case *PyBool:
		if val.Value {
			return "True", nil
		}
		return "False", nil
	case *PyInt:
		return intToDecimal(val)
	case *PyFloat:
		return floatRepr(val.Value), nil
	case *PyComplex:
		return formatComplex(val.Real, val.Imag), nil
	case *PyString:
		return strRepr(val.Value), nil
	case *PyBytes:
		return bytesRepr(val.Value), nil
	case *PyRange:
		if val.Step == 1 {
			return fmt.Sprintf("range(%d, %d)", val.Start, val.Stop), nil
		}
		return fmt.Sprintf("range(%d, %d, %d)", val.Start, val.Stop, val.Step), nil
	case *PyList:
		return vm.reprItems(val, "[", "]", "[...]", val.Items)
	case *PyTuple:
		if len(val.Items) == 1 {
			return vm.reprItems(val, "(", ",)", "(...)", val.Items)
		}
		return vm.reprItems(val, "(", ")", "(...)", val.Items)
	case *PyDict:
		return vm.reprDict(val)
	case *PySet:
		if val.SetLen() == 0 {
			return "set()", nil
		}
		items, err := vm.toList(val)
		if err != nil {
			return "", err
		}
		return vm.reprItems(val, "{", "}", "set(...)", items)
	case *PyFrozenSet:
		if val.FrozenSetLen() == 0 {
			return "frozenset()", nil
		}
		items, err := vm.toList(val)
		if err != nil {
			return "", err
		}
		return vm.reprItems(val, "frozenset({", "})", "frozenset(...)", items)
	case *PyInstance:
		return vm.reprInstance(val)
	case *PyException:
		return vm.reprException(val)
	case *PyClass:
		return fmt.Sprintf("<class '%s'>", val.Name), nil
	case *PyFunction:
		return fmt.Sprintf("<function %s>", val.Name), nil
	case *PyMethod:
		self, err := vm.reprValue(val.Instance)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("<bound method %s of %s>", methodQualName(val), self), nil
	case *PyBuiltinFunc:
		return fmt.Sprintf("<built-in function %s>", val.Name), nil
	case *PyGoFunc:
		return fmt.Sprintf("<built-in function %s>", val.Name), nil
	case *PyUserData:
		if s, found, err := vm.userDataText(val, "__repr__"); found {
			return s, err
		}
		return fmt.Sprintf("<userdata %T>", val.Value), nil
	case *PyModule:
		return fmt.Sprintf("<module '%s'>", val.Name), nil
	case *GenericAlias:
		return val.formatRepr(), nil
	case *UnionType:
		return val.formatRepr(), nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

// strValue returns str(v). For most builtin types str() and repr() agree.
func (vm *VM) strValue(v Value) (string, error) {
	switch val := v.(type) {
	case *PyString:
		return val.Value, nil
	case *PyInstance:
		return vm.strInstance(val)
	case *PyException:
		return vm.strException(val)
	case *PyUserData:
		if s, found, err := vm.userDataText(val, "__str__"); found {
			return s, err
		}
		return vm.reprValue(v)
	default:
		return vm.reprValue(v)
	}
}

// asciiValue returns ascii(v): repr(v) with non-ASCII characters escaped.
func (vm *VM) asciiValue(v Value) (string, error) {
	s, err := vm.reprValue(v)
	if err != nil {
		return "", err
	}
	return asciiEscape(s), nil
}

// reprItems renders items between open and close, emitting placeholder when
// container is already being rendered further up the call chain.
func (vm *VM) reprItems(container Value, open, close, placeholder string, items []Value) (string, error) {
	if !vm.reprEnter(container) {
		return placeholder, nil
	}
	defer vm.reprLeave(container)
	if err := vm.enterReprNesting(reprContainerCost); err != nil {
		return "", err
	}
	defer vm.leaveReprNesting(reprContainerCost)

	var b strings.Builder
	b.WriteString(open)
	for i, item := range items {
		if i > 0 {
			b.WriteString(", ")
		}
		s, err := vm.reprValue(item)
		if err != nil {
			return "", err
		}
		b.WriteString(s)
	}
	b.WriteString(close)
	return b.String(), nil
}

func (vm *VM) reprDict(d *PyDict) (string, error) {
	if !vm.reprEnter(d) {
		return "{...}", nil
	}
	defer vm.reprLeave(d)
	if err := vm.enterReprNesting(reprContainerCost); err != nil {
		return "", err
	}
	defer vm.leaveReprNesting(reprContainerCost)

	var b strings.Builder
	b.WriteByte('{')
	first := true
	for _, k := range d.Keys(vm) {
		v, ok := d.DictGet(k, vm)
		if !ok {
			continue
		}
		if !first {
			b.WriteString(", ")
		}
		first = false
		ks, err := vm.reprValue(k)
		if err != nil {
			return "", err
		}
		vs, err := vm.reprValue(v)
		if err != nil {
			return "", err
		}
		b.WriteString(ks)
		b.WriteString(": ")
		b.WriteString(vs)
	}
	b.WriteByte('}')
	return b.String(), nil
}

// callTextDunder calls a text-returning dunder (__repr__ or __str__) on inst.
func (vm *VM) callTextDunder(inst *PyInstance, name string) (string, bool, error) {
	if err := vm.enterReprNesting(reprCallCost); err != nil {
		return "", true, err
	}
	defer vm.leaveReprNesting(reprCallCost)
	result, found, err := vm.callDunder(inst, name)
	if !found || err != nil {
		return "", found, err
	}
	s, ok := result.(*PyString)
	if !ok {
		return "", true, fmt.Errorf("TypeError: %s returned non-string (type %s)", name, vm.typeName(result))
	}
	return s.Value, true, nil
}

func (vm *VM) reprInstance(inst *PyInstance) (string, error) {
	if s, found, err := vm.callTextDunder(inst, "__repr__"); found {
		return s, err
	}
	return vm.defaultRepr(inst)
}

func (vm *VM) strInstance(inst *PyInstance) (string, error) {
	if s, found, err := vm.callTextDunder(inst, "__str__"); found {
		return s, err
	}
	return vm.defaultStr(inst)
}

// defaultRepr implements object.__repr__, and BaseException.__repr__ for
// exception instances (whose MRO may place object before BaseException).
func (vm *VM) defaultRepr(v Value) (string, error) {
	inst, ok := v.(*PyInstance)
	if !ok {
		return vm.reprValue(v)
	}
	if vm.isExceptionClass(inst.Class) {
		return vm.exceptionRepr(inst.Class.Name, instanceExceptionArgs(inst))
	}
	return fmt.Sprintf("<%s object>", inst.Class.Name), nil
}

// defaultStr implements object.__str__ (which is repr(self)), and
// BaseException.__str__ for exception instances.
func (vm *VM) defaultStr(v Value) (string, error) {
	if inst, ok := v.(*PyInstance); ok && vm.isExceptionClass(inst.Class) {
		return vm.exceptionStr(vm.isKeyErrorClass(inst.Class), instanceExceptionArgs(inst))
	}
	return vm.reprValue(v)
}

// initObjectTextMethods installs object.__repr__, object.__str__ and
// object.__format__, reached through the MRO, super() and object.__repr__(x).
func (vm *VM) initObjectTextMethods(objectClass *PyClass) {
	textMethod := func(name string, fn func(Value) (string, error)) *PyBuiltinFunc {
		return &PyBuiltinFunc{
			Name: "object." + name,
			Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("TypeError: object.%s() takes exactly one argument (%d given)", name, len(args))
				}
				s, err := fn(args[0])
				if err != nil {
					return nil, err
				}
				return &PyString{Value: s}, nil
			},
		}
	}
	objectClass.Dict["__repr__"] = textMethod("__repr__", vm.defaultRepr)
	objectClass.Dict["__str__"] = textMethod("__str__", vm.defaultStr)
	objectClass.Dict["__format__"] = &PyBuiltinFunc{
		Name: "object.__format__",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("TypeError: object.__format__() takes exactly one argument (%d given)", len(args)-1)
			}
			spec, ok := args[1].(*PyString)
			if !ok {
				return nil, fmt.Errorf("TypeError: format_spec must be a str, not %s", vm.typeName(args[1]))
			}
			if spec.Value != "" {
				return nil, fmt.Errorf("TypeError: unsupported format string passed to %s.__format__", vm.typeName(args[0]))
			}
			s, err := vm.strValue(args[0])
			if err != nil {
				return nil, err
			}
			return &PyString{Value: s}, nil
		},
	}
}

// builtinTextMethod returns the bound __repr__, __str__ or __format__ method
// of a builtin value (e.g. "abc".__repr__), or nil if v is not a builtin value.
func (vm *VM) builtinTextMethod(v Value, name string) Value {
	switch v.(type) {
	case *PyInstance, *PyClass, *PyModule, *PyUserData, *PySuper:
		// Resolved through their own attribute lookup
		return nil
	}
	return &PyBuiltinFunc{
		Name: vm.typeName(v) + "." + name,
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			var s string
			var err error
			switch name {
			case "__repr__":
				if len(args) != 0 {
					return nil, fmt.Errorf("TypeError: %s.__repr__() takes no arguments (%d given)", vm.typeName(v), len(args))
				}
				s, err = vm.reprValue(v)
			case "__str__":
				if len(args) != 0 {
					return nil, fmt.Errorf("TypeError: %s.__str__() takes no arguments (%d given)", vm.typeName(v), len(args))
				}
				s, err = vm.strValue(v)
			default:
				if len(args) != 1 {
					return nil, fmt.Errorf("TypeError: %s.__format__() takes exactly one argument (%d given)", vm.typeName(v), len(args))
				}
				spec, ok := args[0].(*PyString)
				if !ok {
					return nil, fmt.Errorf("TypeError: argument to __format__ must be str, not %s", vm.typeName(args[0]))
				}
				s, err = vm.formatValue(v, spec.Value)
			}
			if err != nil {
				return nil, err
			}
			return &PyString{Value: s}, nil
		},
	}
}

// methodQualName returns "Class.name" for a bound method, naming the class in
// the MRO that defines the function (like CPython's __qualname__).
func methodQualName(m *PyMethod) string {
	var mro []*PyClass
	switch owner := m.Instance.(type) {
	case *PyInstance:
		mro = owner.Class.Mro
	case *PyClass:
		mro = owner.Mro
	}
	for _, cls := range mro {
		if fn, ok := cls.Dict[m.Func.Name]; ok {
			switch f := fn.(type) {
			case *PyFunction:
				if f == m.Func {
					return cls.Name + "." + m.Func.Name
				}
			case *PyClassMethod:
				if f.Func == Value(m.Func) {
					return cls.Name + "." + m.Func.Name
				}
			}
		}
	}
	return m.Func.Name
}

// userDataText calls a __repr__/__str__ method registered in the type
// metatable of a userdata value, if there is one.
func (vm *VM) userDataText(ud *PyUserData, name string) (string, bool, error) {
	if ud.Metatable == nil {
		return "", false, nil
	}
	method, err := vm.getAttr(ud, name)
	if err != nil {
		return "", false, nil
	}
	result, err := vm.call(method, nil, nil)
	if err != nil {
		return "", true, err
	}
	s, ok := result.(*PyString)
	if !ok {
		return "", true, fmt.Errorf("TypeError: %s returned non-string (type %s)", name, vm.typeName(result))
	}
	return s.Value, true, nil
}

// =====================================
// Exceptions
// =====================================

// instanceExceptionArgs returns the args of an exception instance.
func instanceExceptionArgs(inst *PyInstance) []Value {
	if t, ok := inst.Dict["args"].(*PyTuple); ok {
		return t.Items
	}
	return nil
}

// exceptionArgs returns the args of a raised exception. Exceptions created
// from Go code may carry only a message.
func exceptionArgs(exc *PyException) []Value {
	if exc.Args != nil {
		return exc.Args.Items
	}
	if exc.Message != "" {
		return []Value{&PyString{Value: exc.Message}}
	}
	return nil
}

// isKeyErrorClass reports whether cls is KeyError or a subclass of it.
// KeyError is special-cased by str() (it shows the repr of the missing key).
func (vm *VM) isKeyErrorClass(cls *PyClass) bool {
	keyError := vm.builtinClass("KeyError")
	if keyError == nil {
		return cls.Name == "KeyError"
	}
	for _, c := range cls.Mro {
		if c == keyError {
			return true
		}
	}
	return false
}

// classDefinesDunder reports whether a class in cls's MRO, other than the
// object base class, defines name itself.
func classDefinesDunder(cls *PyClass, name string) bool {
	for _, c := range cls.Mro {
		if c.Name == "object" && len(c.Bases) == 0 {
			continue
		}
		if _, ok := c.Dict[name]; ok {
			return true
		}
	}
	return false
}

// exceptionStr implements BaseException.__str__ (and KeyError.__str__).
func (vm *VM) exceptionStr(isKeyError bool, args []Value) (string, error) {
	switch len(args) {
	case 0:
		return "", nil
	case 1:
		if isKeyError {
			return vm.reprValue(args[0])
		}
		return vm.strValue(args[0])
	default:
		return vm.reprValue(&PyTuple{Items: args})
	}
}

// exceptionRepr implements BaseException.__repr__: Name(arg) or Name(args...).
func (vm *VM) exceptionRepr(name string, args []Value) (string, error) {
	if len(args) == 1 {
		s, err := vm.reprValue(args[0])
		if err != nil {
			return "", err
		}
		return name + "(" + s + ")", nil
	}
	s, err := vm.reprValue(&PyTuple{Items: args})
	if err != nil {
		return "", err
	}
	return name + s, nil
}

func (vm *VM) strException(exc *PyException) (string, error) {
	if exc.Instance != nil {
		return vm.strInstance(exc.Instance)
	}
	if exc.ExcType != nil {
		if classDefinesDunder(exc.ExcType, "__str__") {
			// The instance is gone; Message holds its __str__ result.
			return exc.Message, nil
		}
		return vm.exceptionStr(vm.isKeyErrorClass(exc.ExcType), exceptionArgs(exc))
	}
	return vm.exceptionStr(exc.TypeName == "KeyError", exceptionArgs(exc))
}

func (vm *VM) reprException(exc *PyException) (string, error) {
	if exc.Instance != nil {
		return vm.reprInstance(exc.Instance)
	}
	return vm.exceptionRepr(exc.Type(), exceptionArgs(exc))
}

// =====================================
// Scalar text helpers (no VM needed)
// =====================================

// intToDecimal converts an int to decimal text, enforcing CPython's limit on
// the number of digits (sys.int_info.default_max_str_digits).
func intToDecimal(i *PyInt) (string, error) {
	if i.BigValue == nil {
		return strconv.FormatInt(i.Value, 10), nil
	}
	s := i.BigValue.String()
	digits := len(s)
	if s[0] == '-' {
		digits--
	}
	if digits > maxIntStringDigits {
		return "", fmt.Errorf("ValueError: Exceeds the limit (%d digits) for integer string conversion; use sys.set_int_max_str_digits() to increase the limit", maxIntStringDigits)
	}
	return s, nil
}

// strRepr returns the repr of a str: single quotes unless the string contains
// ' and no ", with backslashes, the quote character and non-printable
// characters escaped. Printable non-ASCII characters are kept as-is.
func strRepr(s string) string {
	quote := byte('\'')
	if strings.IndexByte(s, '\'') >= 0 && strings.IndexByte(s, '"') < 0 {
		quote = '"'
	}
	var b strings.Builder
	b.Grow(len(s) + 2)
	b.WriteByte(quote)
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && size == 1 {
			// Not valid UTF-8; show the raw byte.
			fmt.Fprintf(&b, "\\x%02x", s[i])
			i++
			continue
		}
		i += size
		switch {
		case r == rune(quote) || r == '\\':
			b.WriteByte('\\')
			b.WriteRune(r)
		case r == '\t':
			b.WriteString(`\t`)
		case r == '\n':
			b.WriteString(`\n`)
		case r == '\r':
			b.WriteString(`\r`)
		case r < 0x20 || r == 0x7f:
			fmt.Fprintf(&b, "\\x%02x", r)
		case r < 0x7f || unicode.IsPrint(r):
			b.WriteRune(r)
		case r <= 0xff:
			fmt.Fprintf(&b, "\\x%02x", r)
		case r <= 0xffff:
			fmt.Fprintf(&b, "\\u%04x", r)
		default:
			fmt.Fprintf(&b, "\\U%08x", r)
		}
	}
	b.WriteByte(quote)
	return b.String()
}

// bytesRepr returns the repr of a bytes object, with the same quote selection
// rule as str.
func bytesRepr(data []byte) string {
	quote := byte('\'')
	if bytes.IndexByte(data, '\'') >= 0 && bytes.IndexByte(data, '"') < 0 {
		quote = '"'
	}
	var b strings.Builder
	b.Grow(len(data) + 3)
	b.WriteByte('b')
	b.WriteByte(quote)
	for _, c := range data {
		switch {
		case c == quote || c == '\\':
			b.WriteByte('\\')
			b.WriteByte(c)
		case c == '\t':
			b.WriteString(`\t`)
		case c == '\n':
			b.WriteString(`\n`)
		case c == '\r':
			b.WriteString(`\r`)
		case c < 0x20 || c >= 0x7f:
			fmt.Fprintf(&b, "\\x%02x", c)
		default:
			b.WriteByte(c)
		}
	}
	b.WriteByte(quote)
	return b.String()
}

// asciiEscape replaces non-ASCII characters with \x, \u or \U escapes.
func asciiEscape(s string) string {
	ascii := true
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			ascii = false
			break
		}
	}
	if ascii {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		switch {
		case r == utf8.RuneError && size == 1:
			fmt.Fprintf(&b, "\\x%02x", s[i])
		case r < utf8.RuneSelf:
			b.WriteByte(byte(r))
		case r <= 0xff:
			fmt.Fprintf(&b, "\\x%02x", r)
		case r <= 0xffff:
			fmt.Fprintf(&b, "\\u%04x", r)
		default:
			fmt.Fprintf(&b, "\\U%08x", r)
		}
		i += size
	}
	return b.String()
}

// asciiRepr returns ascii() of a str.
func asciiRepr(s string) string {
	return asciiEscape(strRepr(s))
}

// floatRepr returns repr(f): the shortest string that round-trips, using
// scientific notation only for exponents < -4 or >= 16.
func floatRepr(f float64) string {
	return formatFloat(f, 'r', 0, fmtFloatAddDot0)
}

// formatComplex returns repr() of a complex number, matching CPython:
// "3j" when the real part is +0, otherwise "(1+2j)".
func formatComplex(real, imag float64) string {
	if real == 0 && !math.Signbit(real) {
		return formatFloat(imag, 'r', 0, 0) + "j"
	}
	return "(" + formatFloat(real, 'r', 0, 0) + formatFloat(imag, 'r', 0, fmtFloatSign) + "j)"
}
