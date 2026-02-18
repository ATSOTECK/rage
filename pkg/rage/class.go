package rage

import (
	"errors"
	"fmt"

	"github.com/ATSOTECK/rage/internal/runtime"
)

// ErrStopIteration should be returned from Next callbacks to signal end of iteration.
var ErrStopIteration = errors.New("StopIteration: ")

// TypeError returns an error that becomes a Python TypeError.
func TypeError(msg string) error {
	return fmt.Errorf("TypeError: %s", msg)
}

// ValueError returns an error that becomes a Python ValueError.
func ValueError(msg string) error {
	return fmt.Errorf("ValueError: %s", msg)
}

// KeyError returns an error that becomes a Python KeyError.
func KeyError(msg string) error {
	return fmt.Errorf("KeyError: %s", msg)
}

// IndexError returns an error that becomes a Python IndexError.
func IndexError(msg string) error {
	return fmt.Errorf("IndexError: %s", msg)
}

// AttributeError returns an error that becomes a Python AttributeError.
func AttributeError(msg string) error {
	return fmt.Errorf("AttributeError: %s", msg)
}

// RuntimeError returns an error that becomes a Python RuntimeError.
func RuntimeError(msg string) error {
	return fmt.Errorf("RuntimeError: %s", msg)
}

// Object wraps a Python instance, providing Go methods to read and write attributes on self.
type Object struct {
	inst *runtime.PyInstance
}

// Get returns the value of an attribute on the instance.
func (o Object) Get(name string) Value {
	if o.inst.Dict != nil {
		if v, ok := o.inst.Dict[name]; ok {
			return fromRuntime(v)
		}
	}
	if o.inst.Slots != nil {
		if v, ok := o.inst.Slots[name]; ok {
			return fromRuntime(v)
		}
	}
	return None
}

// Set sets an attribute on the instance.
func (o Object) Set(name string, val Value) {
	if o.inst.Dict != nil {
		o.inst.Dict[name] = toRuntime(val)
	} else if o.inst.Slots != nil {
		o.inst.Slots[name] = toRuntime(val)
	}
}

// Has returns true if the instance has the named attribute.
func (o Object) Has(name string) bool {
	if o.inst.Dict != nil {
		_, ok := o.inst.Dict[name]
		return ok
	}
	if o.inst.Slots != nil {
		_, ok := o.inst.Slots[name]
		return ok
	}
	return false
}

// Delete removes an attribute from the instance.
func (o Object) Delete(name string) {
	if o.inst.Dict != nil {
		delete(o.inst.Dict, name)
	} else if o.inst.Slots != nil {
		delete(o.inst.Slots, name)
	}
}

// ClassName returns the name of the instance's class.
func (o Object) ClassName() string {
	return o.inst.Class.Name
}

// Class returns the ClassValue of this instance.
func (o Object) Class() ClassValue {
	return ClassValue{class: o.inst.Class}
}

// Type returns the Python type name of this object.
func (o Object) Type() string { return o.inst.Class.Name }

// String returns a string representation of this object.
func (o Object) String() string { return o.inst.String() }

// GoValue returns the underlying *runtime.PyInstance.
func (o Object) GoValue() any { return o.inst }

// toRuntime returns the underlying runtime value.
func (o Object) toRuntime() runtime.Value { return o.inst }

// ClassValue wraps a *runtime.PyClass, implementing rage.Value.
type ClassValue struct {
	class *runtime.PyClass
}

// Name returns the class name.
func (c ClassValue) Name() string { return c.class.Name }

// NewInstance creates a new instance of this class without calling __init__.
// Useful for Go code that wants to set up attributes manually.
func (c ClassValue) NewInstance() Object {
	inst := &runtime.PyInstance{
		Class: c.class,
		Dict:  make(map[string]runtime.Value),
	}
	return Object{inst: inst}
}

// Type returns "type".
func (c ClassValue) Type() string { return "type" }

// String returns the class string representation.
func (c ClassValue) String() string { return c.class.String() }

// GoValue returns the underlying *runtime.PyClass.
func (c ClassValue) GoValue() any { return c.class }

// toRuntime returns the underlying runtime value.
func (c ClassValue) toRuntime() runtime.Value { return c.class }

// methodDef stores a Go function to be wrapped as an instance method at Build time.
// All methods are stored kwargs-aware internally.
type methodDef struct {
	fn func(s *State, self Object, args []Value, kwargs map[string]Value) (Value, error)
}

// classMethodDef stores a Go function to be wrapped as a class method at Build time.
type classMethodDef struct {
	fn func(s *State, cls ClassValue, args []Value, kwargs map[string]Value) (Value, error)
}

// staticMethodDef stores a Go function to be wrapped as a static method at Build time.
type staticMethodDef struct {
	fn func(s *State, args []Value, kwargs map[string]Value) (Value, error)
}

// propertyDef stores getter/setter functions to be wrapped at Build time.
type propertyDef struct {
	getter func(s *State, self Object) (Value, error)
	setter func(s *State, self Object, val Value) error // nil for read-only
}

// ClassBuilder provides a fluent API for building Python classes from Go.
type ClassBuilder struct {
	name         string
	bases        []*runtime.PyClass
	initFn       func(s *State, self Object, args []Value, kwargs map[string]Value) error
	methods      map[string]methodDef
	classMethods map[string]classMethodDef
	statics      map[string]staticMethodDef
	properties   map[string]propertyDef
}

// NewClass starts building a new Python class with the given name.
func NewClass(name string) *ClassBuilder {
	return &ClassBuilder{
		name:         name,
		methods:      make(map[string]methodDef),
		classMethods: make(map[string]classMethodDef),
		statics:      make(map[string]staticMethodDef),
		properties:   make(map[string]propertyDef),
	}
}

// Base sets a single base class. If not called, defaults to object.
func (b *ClassBuilder) Base(base ClassValue) *ClassBuilder {
	b.bases = []*runtime.PyClass{base.class}
	return b
}

// Bases sets multiple base classes for multiple inheritance.
func (b *ClassBuilder) Bases(bases ...ClassValue) *ClassBuilder {
	b.bases = make([]*runtime.PyClass, len(bases))
	for i, base := range bases {
		b.bases[i] = base.class
	}
	return b
}

// Init sets the __init__ method.
func (b *ClassBuilder) Init(fn func(s *State, self Object, args ...Value) error) *ClassBuilder {
	b.initFn = func(s *State, self Object, args []Value, kwargs map[string]Value) error {
		return fn(s, self, args...)
	}
	return b
}

// InitKw sets the __init__ method with keyword argument support.
func (b *ClassBuilder) InitKw(fn func(s *State, self Object, args []Value, kwargs map[string]Value) error) *ClassBuilder {
	b.initFn = fn
	return b
}

// Method adds a regular instance method.
func (b *ClassBuilder) Method(name string, fn func(s *State, self Object, args ...Value) (Value, error)) *ClassBuilder {
	b.methods[name] = methodDef{fn: func(s *State, self Object, args []Value, kwargs map[string]Value) (Value, error) {
		return fn(s, self, args...)
	}}
	return b
}

// MethodKw adds an instance method with keyword argument support.
func (b *ClassBuilder) MethodKw(name string, fn func(s *State, self Object, args []Value, kwargs map[string]Value) (Value, error)) *ClassBuilder {
	b.methods[name] = methodDef{fn: fn}
	return b
}

// Str sets the __str__ method.
func (b *ClassBuilder) Str(fn func(s *State, self Object) (string, error)) *ClassBuilder {
	b.methods["__str__"] = methodDef{fn: func(s *State, self Object, args []Value, kwargs map[string]Value) (Value, error) {
		str, err := fn(s, self)
		if err != nil {
			return nil, err
		}
		return String(str), nil
	}}
	return b
}

// Repr sets the __repr__ method.
func (b *ClassBuilder) Repr(fn func(s *State, self Object) (string, error)) *ClassBuilder {
	b.methods["__repr__"] = methodDef{fn: func(s *State, self Object, args []Value, kwargs map[string]Value) (Value, error) {
		str, err := fn(s, self)
		if err != nil {
			return nil, err
		}
		return String(str), nil
	}}
	return b
}

// Eq sets the __eq__ method.
func (b *ClassBuilder) Eq(fn func(s *State, self Object, other Value) (bool, error)) *ClassBuilder {
	b.methods["__eq__"] = methodDef{fn: func(s *State, self Object, args []Value, kwargs map[string]Value) (Value, error) {
		if len(args) < 1 {
			return False, nil
		}
		result, err := fn(s, self, args[0])
		if err != nil {
			return nil, err
		}
		return Bool(result), nil
	}}
	return b
}

// Ne sets the __ne__ method.
func (b *ClassBuilder) Ne(fn func(s *State, self Object, other Value) (bool, error)) *ClassBuilder {
	b.methods["__ne__"] = methodDef{fn: func(s *State, self Object, args []Value, kwargs map[string]Value) (Value, error) {
		if len(args) < 1 {
			return True, nil
		}
		result, err := fn(s, self, args[0])
		if err != nil {
			return nil, err
		}
		return Bool(result), nil
	}}
	return b
}

// Lt sets the __lt__ method.
func (b *ClassBuilder) Lt(fn func(s *State, self Object, other Value) (bool, error)) *ClassBuilder {
	b.methods["__lt__"] = methodDef{fn: func(s *State, self Object, args []Value, kwargs map[string]Value) (Value, error) {
		if len(args) < 1 {
			return nil, TypeError("'<' not supported")
		}
		result, err := fn(s, self, args[0])
		if err != nil {
			return nil, err
		}
		return Bool(result), nil
	}}
	return b
}

// Le sets the __le__ method.
func (b *ClassBuilder) Le(fn func(s *State, self Object, other Value) (bool, error)) *ClassBuilder {
	b.methods["__le__"] = methodDef{fn: func(s *State, self Object, args []Value, kwargs map[string]Value) (Value, error) {
		if len(args) < 1 {
			return nil, TypeError("'<=' not supported")
		}
		result, err := fn(s, self, args[0])
		if err != nil {
			return nil, err
		}
		return Bool(result), nil
	}}
	return b
}

// Gt sets the __gt__ method.
func (b *ClassBuilder) Gt(fn func(s *State, self Object, other Value) (bool, error)) *ClassBuilder {
	b.methods["__gt__"] = methodDef{fn: func(s *State, self Object, args []Value, kwargs map[string]Value) (Value, error) {
		if len(args) < 1 {
			return nil, TypeError("'>' not supported")
		}
		result, err := fn(s, self, args[0])
		if err != nil {
			return nil, err
		}
		return Bool(result), nil
	}}
	return b
}

// Ge sets the __ge__ method.
func (b *ClassBuilder) Ge(fn func(s *State, self Object, other Value) (bool, error)) *ClassBuilder {
	b.methods["__ge__"] = methodDef{fn: func(s *State, self Object, args []Value, kwargs map[string]Value) (Value, error) {
		if len(args) < 1 {
			return nil, TypeError("'>=' not supported")
		}
		result, err := fn(s, self, args[0])
		if err != nil {
			return nil, err
		}
		return Bool(result), nil
	}}
	return b
}

// Hash sets the __hash__ method.
func (b *ClassBuilder) Hash(fn func(s *State, self Object) (int64, error)) *ClassBuilder {
	b.methods["__hash__"] = methodDef{fn: func(s *State, self Object, args []Value, kwargs map[string]Value) (Value, error) {
		h, err := fn(s, self)
		if err != nil {
			return nil, err
		}
		return Int(h), nil
	}}
	return b
}

// Len sets the __len__ method.
func (b *ClassBuilder) Len(fn func(s *State, self Object) (int64, error)) *ClassBuilder {
	b.methods["__len__"] = methodDef{fn: func(s *State, self Object, args []Value, kwargs map[string]Value) (Value, error) {
		n, err := fn(s, self)
		if err != nil {
			return nil, err
		}
		return Int(n), nil
	}}
	return b
}

// GetItem sets the __getitem__ method.
func (b *ClassBuilder) GetItem(fn func(s *State, self Object, key Value) (Value, error)) *ClassBuilder {
	b.methods["__getitem__"] = methodDef{fn: func(s *State, self Object, args []Value, kwargs map[string]Value) (Value, error) {
		if len(args) < 1 {
			return nil, TypeError("__getitem__ requires a key argument")
		}
		return fn(s, self, args[0])
	}}
	return b
}

// SetItem sets the __setitem__ method.
func (b *ClassBuilder) SetItem(fn func(s *State, self Object, key, val Value) error) *ClassBuilder {
	b.methods["__setitem__"] = methodDef{fn: func(s *State, self Object, args []Value, kwargs map[string]Value) (Value, error) {
		if len(args) < 2 {
			return nil, TypeError("__setitem__ requires key and value arguments")
		}
		err := fn(s, self, args[0], args[1])
		if err != nil {
			return nil, err
		}
		return None, nil
	}}
	return b
}

// DelItem sets the __delitem__ method.
func (b *ClassBuilder) DelItem(fn func(s *State, self Object, key Value) error) *ClassBuilder {
	b.methods["__delitem__"] = methodDef{fn: func(s *State, self Object, args []Value, kwargs map[string]Value) (Value, error) {
		if len(args) < 1 {
			return nil, TypeError("__delitem__ requires a key argument")
		}
		err := fn(s, self, args[0])
		if err != nil {
			return nil, err
		}
		return None, nil
	}}
	return b
}

// Contains sets the __contains__ method.
func (b *ClassBuilder) Contains(fn func(s *State, self Object, item Value) (bool, error)) *ClassBuilder {
	b.methods["__contains__"] = methodDef{fn: func(s *State, self Object, args []Value, kwargs map[string]Value) (Value, error) {
		if len(args) < 1 {
			return False, nil
		}
		result, err := fn(s, self, args[0])
		if err != nil {
			return nil, err
		}
		return Bool(result), nil
	}}
	return b
}

// Bool sets the __bool__ method.
func (b *ClassBuilder) Bool(fn func(s *State, self Object) (bool, error)) *ClassBuilder {
	b.methods["__bool__"] = methodDef{fn: func(s *State, self Object, args []Value, kwargs map[string]Value) (Value, error) {
		result, err := fn(s, self)
		if err != nil {
			return nil, err
		}
		return Bool(result), nil
	}}
	return b
}

// Call sets the __call__ method, making instances callable.
func (b *ClassBuilder) Call(fn func(s *State, self Object, args ...Value) (Value, error)) *ClassBuilder {
	b.methods["__call__"] = methodDef{fn: func(s *State, self Object, args []Value, kwargs map[string]Value) (Value, error) {
		return fn(s, self, args...)
	}}
	return b
}

// Iter sets the __iter__ method. Return self for objects that are their own iterator.
func (b *ClassBuilder) Iter(fn func(s *State, self Object) (Value, error)) *ClassBuilder {
	b.methods["__iter__"] = methodDef{fn: func(s *State, self Object, args []Value, kwargs map[string]Value) (Value, error) {
		return fn(s, self)
	}}
	return b
}

// Next sets the __next__ method. Return ErrStopIteration to signal end of iteration.
func (b *ClassBuilder) Next(fn func(s *State, self Object) (Value, error)) *ClassBuilder {
	b.methods["__next__"] = methodDef{fn: func(s *State, self Object, args []Value, kwargs map[string]Value) (Value, error) {
		return fn(s, self)
	}}
	return b
}

// Enter sets the __enter__ method for context managers.
func (b *ClassBuilder) Enter(fn func(s *State, self Object) (Value, error)) *ClassBuilder {
	b.methods["__enter__"] = methodDef{fn: func(s *State, self Object, args []Value, kwargs map[string]Value) (Value, error) {
		return fn(s, self)
	}}
	return b
}

// Exit sets the __exit__ method for context managers.
// Return true to suppress the exception, false to propagate it.
// excType, excVal, and excTb are None when no exception occurred.
func (b *ClassBuilder) Exit(fn func(s *State, self Object, excType, excVal, excTb Value) (bool, error)) *ClassBuilder {
	b.methods["__exit__"] = methodDef{fn: func(s *State, self Object, args []Value, kwargs map[string]Value) (Value, error) {
		var excType, excVal, excTb Value = None, None, None
		if len(args) > 0 {
			excType = args[0]
		}
		if len(args) > 1 {
			excVal = args[1]
		}
		if len(args) > 2 {
			excTb = args[2]
		}
		suppress, err := fn(s, self, excType, excVal, excTb)
		if err != nil {
			return nil, err
		}
		return Bool(suppress), nil
	}}
	return b
}

// Dunder adds an arbitrary dunder method.
func (b *ClassBuilder) Dunder(name string, fn func(s *State, self Object, args ...Value) (Value, error)) *ClassBuilder {
	b.methods[name] = methodDef{fn: func(s *State, self Object, args []Value, kwargs map[string]Value) (Value, error) {
		return fn(s, self, args...)
	}}
	return b
}

// Property adds a read-only property.
func (b *ClassBuilder) Property(name string, getter func(s *State, self Object) (Value, error)) *ClassBuilder {
	b.properties[name] = propertyDef{getter: getter}
	return b
}

// PropertyWithSetter adds a read-write property.
func (b *ClassBuilder) PropertyWithSetter(name string, getter func(s *State, self Object) (Value, error), setter func(s *State, self Object, val Value) error) *ClassBuilder {
	b.properties[name] = propertyDef{getter: getter, setter: setter}
	return b
}

// ClassMethod adds a class method. The first argument to fn is the class, not an instance.
func (b *ClassBuilder) ClassMethod(name string, fn func(s *State, cls ClassValue, args ...Value) (Value, error)) *ClassBuilder {
	b.classMethods[name] = classMethodDef{fn: func(s *State, cls ClassValue, args []Value, kwargs map[string]Value) (Value, error) {
		return fn(s, cls, args...)
	}}
	return b
}

// ClassMethodKw adds a class method with keyword argument support.
func (b *ClassBuilder) ClassMethodKw(name string, fn func(s *State, cls ClassValue, args []Value, kwargs map[string]Value) (Value, error)) *ClassBuilder {
	b.classMethods[name] = classMethodDef{fn: fn}
	return b
}

// StaticMethod adds a static method. No self or cls is passed.
func (b *ClassBuilder) StaticMethod(name string, fn func(s *State, args ...Value) (Value, error)) *ClassBuilder {
	b.statics[name] = staticMethodDef{fn: func(s *State, args []Value, kwargs map[string]Value) (Value, error) {
		return fn(s, args...)
	}}
	return b
}

// StaticMethodKw adds a static method with keyword argument support.
func (b *ClassBuilder) StaticMethodKw(name string, fn func(s *State, args []Value, kwargs map[string]Value) (Value, error)) *ClassBuilder {
	b.statics[name] = staticMethodDef{fn: fn}
	return b
}

// Build creates the Python class and registers it in the given State.
// Returns a ClassValue that can be passed to State.SetGlobal.
func (b *ClassBuilder) Build(s *State) ClassValue {
	vm := s.vm
	objectClass := vm.GetBuiltin("object").(*runtime.PyClass)

	bases := b.bases
	if len(bases) == 0 {
		bases = []*runtime.PyClass{objectClass}
	}

	cls := &runtime.PyClass{
		Name:  b.name,
		Bases: bases,
		Dict:  make(map[string]runtime.Value),
	}

	mro, err := vm.ComputeC3MRO(cls, cls.Bases)
	if err != nil {
		// Fallback: simple linear MRO
		cls.Mro = []*runtime.PyClass{cls}
		for _, base := range bases {
			cls.Mro = append(cls.Mro, base.Mro...)
		}
	} else {
		cls.Mro = mro
	}

	// Set metaclass to type
	if typeClass, ok := vm.GetBuiltin("type").(*runtime.PyClass); ok {
		cls.Metaclass = typeClass
	}

	// Add __init__ if provided
	if b.initFn != nil {
		initFn := b.initFn
		cls.Dict["__init__"] = makeInstanceMethodKw(b.name, "__init__", s, func(s *State, self Object, args []Value, kwargs map[string]Value) (Value, error) {
			err := initFn(s, self, args, kwargs)
			if err != nil {
				return nil, err
			}
			return None, nil
		})
	}

	// Add instance methods
	for name, def := range b.methods {
		cls.Dict[name] = makeInstanceMethodKw(b.name, name, s, def.fn)
	}

	// Add class methods
	for name, def := range b.classMethods {
		fn := def.fn
		className := b.name
		cls.Dict[name] = &runtime.PyClassMethod{
			Func: &runtime.PyBuiltinFunc{
				Name: className + "." + name,
				Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
					if len(args) < 1 {
						return nil, fmt.Errorf("TypeError: %s.%s() requires a class argument", className, name)
					}
					clsArg, ok := args[0].(*runtime.PyClass)
					if !ok {
						return nil, fmt.Errorf("TypeError: %s.%s() first argument must be a class", className, name)
					}
					cv := ClassValue{class: clsArg}
					rageArgs := make([]Value, len(args)-1)
					for i := 1; i < len(args); i++ {
						rageArgs[i-1] = fromRuntime(args[i])
					}
					rageKwargs := convertKwargs(kwargs)
					result, err := fn(s, cv, rageArgs, rageKwargs)
					if err != nil {
						return nil, err
					}
					if result == nil {
						return runtime.None, nil
					}
					return toRuntime(result), nil
				},
			},
		}
	}

	// Add static methods
	for name, def := range b.statics {
		fn := def.fn
		className := b.name
		cls.Dict[name] = &runtime.PyStaticMethod{
			Func: &runtime.PyBuiltinFunc{
				Name: className + "." + name,
				Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
					rageArgs := make([]Value, len(args))
					for i, a := range args {
						rageArgs[i] = fromRuntime(a)
					}
					rageKwargs := convertKwargs(kwargs)
					result, err := fn(s, rageArgs, rageKwargs)
					if err != nil {
						return nil, err
					}
					if result == nil {
						return runtime.None, nil
					}
					return toRuntime(result), nil
				},
			},
		}
	}

	// Add properties
	for name, def := range b.properties {
		prop := &runtime.PyProperty{}
		if def.getter != nil {
			getter := def.getter
			prop.Fget = makeInstanceMethodKw(b.name, name+".fget", s, func(s *State, self Object, args []Value, kwargs map[string]Value) (Value, error) {
				return getter(s, self)
			})
		}
		if def.setter != nil {
			setter := def.setter
			prop.Fset = makeInstanceMethodKw(b.name, name+".fset", s, func(s *State, self Object, args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) < 1 {
					return nil, TypeError("property setter requires a value")
				}
				err := setter(s, self, args[0])
				if err != nil {
					return nil, err
				}
				return None, nil
			})
		}
		cls.Dict[name] = prop
	}

	return ClassValue{class: cls}
}

// makeInstanceMethodKw creates a *PyBuiltinFunc that extracts self from args[0],
// wraps it in Object, converts kwargs, and calls the Go function.
func makeInstanceMethodKw(className, methodName string, s *State, fn func(s *State, self Object, args []Value, kwargs map[string]Value) (Value, error)) *runtime.PyBuiltinFunc {
	return &runtime.PyBuiltinFunc{
		Name: className + "." + methodName,
		Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("TypeError: %s.%s() requires self", className, methodName)
			}
			inst, ok := args[0].(*runtime.PyInstance)
			if !ok {
				return nil, fmt.Errorf("TypeError: %s.%s() self must be an instance, got %T", className, methodName, args[0])
			}
			self := Object{inst: inst}
			rageArgs := make([]Value, len(args)-1)
			for i := 1; i < len(args); i++ {
				rageArgs[i-1] = fromRuntime(args[i])
			}
			rageKwargs := convertKwargs(kwargs)
			result, err := fn(s, self, rageArgs, rageKwargs)
			if err != nil {
				return nil, err
			}
			if result == nil {
				return runtime.None, nil
			}
			return toRuntime(result), nil
		},
	}
}

// convertKwargs converts runtime kwargs to rage kwargs.
func convertKwargs(kwargs map[string]runtime.Value) map[string]Value {
	if len(kwargs) == 0 {
		return nil
	}
	result := make(map[string]Value, len(kwargs))
	for k, v := range kwargs {
		result[k] = fromRuntime(v)
	}
	return result
}
