package rage

import (
	"fmt"

	"github.com/ATSOTECK/rage/internal/runtime"
)

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
type methodDef struct {
	fn func(s *State, self Object, args ...Value) Value
}

// classMethodDef stores a Go function to be wrapped as a class method at Build time.
type classMethodDef struct {
	fn func(s *State, cls ClassValue, args ...Value) Value
}

// staticMethodDef stores a Go function to be wrapped as a static method at Build time.
type staticMethodDef struct {
	fn func(s *State, args ...Value) Value
}

// propertyDef stores getter/setter functions to be wrapped at Build time.
type propertyDef struct {
	getter func(s *State, self Object) Value
	setter func(s *State, self Object, val Value) // nil for read-only
}

// ClassBuilder provides a fluent API for building Python classes from Go.
type ClassBuilder struct {
	name         string
	base         *runtime.PyClass
	initFn       func(s *State, self Object, args ...Value)
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

// Base sets the base class. If not called, defaults to object.
func (b *ClassBuilder) Base(base ClassValue) *ClassBuilder {
	b.base = base.class
	return b
}

// Init sets the __init__ method.
func (b *ClassBuilder) Init(fn func(s *State, self Object, args ...Value)) *ClassBuilder {
	b.initFn = fn
	return b
}

// Method adds a regular instance method.
func (b *ClassBuilder) Method(name string, fn func(s *State, self Object, args ...Value) Value) *ClassBuilder {
	b.methods[name] = methodDef{fn: fn}
	return b
}

// Str sets the __str__ method.
func (b *ClassBuilder) Str(fn func(s *State, self Object) string) *ClassBuilder {
	b.methods["__str__"] = methodDef{fn: func(s *State, self Object, args ...Value) Value {
		return String(fn(s, self))
	}}
	return b
}

// Repr sets the __repr__ method.
func (b *ClassBuilder) Repr(fn func(s *State, self Object) string) *ClassBuilder {
	b.methods["__repr__"] = methodDef{fn: func(s *State, self Object, args ...Value) Value {
		return String(fn(s, self))
	}}
	return b
}

// Eq sets the __eq__ method.
func (b *ClassBuilder) Eq(fn func(s *State, self Object, other Value) bool) *ClassBuilder {
	b.methods["__eq__"] = methodDef{fn: func(s *State, self Object, args ...Value) Value {
		if len(args) < 1 {
			return False
		}
		return Bool(fn(s, self, args[0]))
	}}
	return b
}

// Len sets the __len__ method.
func (b *ClassBuilder) Len(fn func(s *State, self Object) int64) *ClassBuilder {
	b.methods["__len__"] = methodDef{fn: func(s *State, self Object, args ...Value) Value {
		return Int(fn(s, self))
	}}
	return b
}

// GetItem sets the __getitem__ method.
func (b *ClassBuilder) GetItem(fn func(s *State, self Object, key Value) Value) *ClassBuilder {
	b.methods["__getitem__"] = methodDef{fn: func(s *State, self Object, args ...Value) Value {
		if len(args) < 1 {
			return None
		}
		return fn(s, self, args[0])
	}}
	return b
}

// SetItem sets the __setitem__ method.
func (b *ClassBuilder) SetItem(fn func(s *State, self Object, key, val Value)) *ClassBuilder {
	b.methods["__setitem__"] = methodDef{fn: func(s *State, self Object, args ...Value) Value {
		if len(args) >= 2 {
			fn(s, self, args[0], args[1])
		}
		return None
	}}
	return b
}

// Contains sets the __contains__ method.
func (b *ClassBuilder) Contains(fn func(s *State, self Object, item Value) bool) *ClassBuilder {
	b.methods["__contains__"] = methodDef{fn: func(s *State, self Object, args ...Value) Value {
		if len(args) < 1 {
			return False
		}
		return Bool(fn(s, self, args[0]))
	}}
	return b
}

// Bool sets the __bool__ method.
func (b *ClassBuilder) Bool(fn func(s *State, self Object) bool) *ClassBuilder {
	b.methods["__bool__"] = methodDef{fn: func(s *State, self Object, args ...Value) Value {
		return Bool(fn(s, self))
	}}
	return b
}

// Call sets the __call__ method, making instances callable.
func (b *ClassBuilder) Call(fn func(s *State, self Object, args ...Value) Value) *ClassBuilder {
	b.methods["__call__"] = methodDef{fn: fn}
	return b
}

// Dunder adds an arbitrary dunder method.
func (b *ClassBuilder) Dunder(name string, fn func(s *State, self Object, args ...Value) Value) *ClassBuilder {
	b.methods[name] = methodDef{fn: fn}
	return b
}

// Property adds a read-only property.
func (b *ClassBuilder) Property(name string, getter func(s *State, self Object) Value) *ClassBuilder {
	b.properties[name] = propertyDef{getter: getter}
	return b
}

// PropertyWithSetter adds a read-write property.
func (b *ClassBuilder) PropertyWithSetter(name string, getter func(s *State, self Object) Value, setter func(s *State, self Object, val Value)) *ClassBuilder {
	b.properties[name] = propertyDef{getter: getter, setter: setter}
	return b
}

// ClassMethod adds a class method. The first argument to fn is the class, not an instance.
func (b *ClassBuilder) ClassMethod(name string, fn func(s *State, cls ClassValue, args ...Value) Value) *ClassBuilder {
	b.classMethods[name] = classMethodDef{fn: fn}
	return b
}

// StaticMethod adds a static method. No self or cls is passed.
func (b *ClassBuilder) StaticMethod(name string, fn func(s *State, args ...Value) Value) *ClassBuilder {
	b.statics[name] = staticMethodDef{fn: fn}
	return b
}

// Build creates the Python class and registers it in the given State.
// Returns a ClassValue that can be passed to State.SetGlobal.
func (b *ClassBuilder) Build(s *State) ClassValue {
	vm := s.vm
	objectClass := vm.GetBuiltin("object").(*runtime.PyClass)

	base := b.base
	if base == nil {
		base = objectClass
	}

	cls := &runtime.PyClass{
		Name:  b.name,
		Bases: []*runtime.PyClass{base},
		Dict:  make(map[string]runtime.Value),
	}

	mro, err := vm.ComputeC3MRO(cls, cls.Bases)
	if err != nil {
		// Fallback: simple linear MRO
		cls.Mro = append([]*runtime.PyClass{cls}, base.Mro...)
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
		cls.Dict["__init__"] = makeInstanceMethod(b.name, "__init__", s, func(s *State, self Object, args ...Value) Value {
			initFn(s, self, args...)
			return None
		})
	}

	// Add instance methods
	for name, def := range b.methods {
		cls.Dict[name] = makeInstanceMethod(b.name, name, s, def.fn)
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
						return nil, fmt.Errorf("%s.%s() requires a class argument", className, name)
					}
					clsArg, ok := args[0].(*runtime.PyClass)
					if !ok {
						return nil, fmt.Errorf("%s.%s() first argument must be a class", className, name)
					}
					cv := ClassValue{class: clsArg}
					rageArgs := make([]Value, len(args)-1)
					for i := 1; i < len(args); i++ {
						rageArgs[i-1] = fromRuntime(args[i])
					}
					result := fn(s, cv, rageArgs...)
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
					result := fn(s, rageArgs...)
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
			prop.Fget = makeInstanceMethod(b.name, name+".fget", s, func(s *State, self Object, args ...Value) Value {
				return getter(s, self)
			})
		}
		if def.setter != nil {
			setter := def.setter
			prop.Fset = makeInstanceMethod(b.name, name+".fset", s, func(s *State, self Object, args ...Value) Value {
				if len(args) >= 1 {
					setter(s, self, args[0])
				}
				return None
			})
		}
		cls.Dict[name] = prop
	}

	return ClassValue{class: cls}
}

// makeInstanceMethod creates a *PyBuiltinFunc that extracts self from args[0],
// wraps it in Object, and calls the Go function.
func makeInstanceMethod(className, methodName string, s *State, fn func(s *State, self Object, args ...Value) Value) *runtime.PyBuiltinFunc {
	return &runtime.PyBuiltinFunc{
		Name: className + "." + methodName,
		Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("%s.%s() requires self", className, methodName)
			}
			inst, ok := args[0].(*runtime.PyInstance)
			if !ok {
				return nil, fmt.Errorf("%s.%s() self must be an instance, got %T", className, methodName, args[0])
			}
			self := Object{inst: inst}
			rageArgs := make([]Value, len(args)-1)
			for i := 1; i < len(args); i++ {
				rageArgs[i-1] = fromRuntime(args[i])
			}
			result := fn(s, self, rageArgs...)
			if result == nil {
				return runtime.None, nil
			}
			return toRuntime(result), nil
		},
	}
}
