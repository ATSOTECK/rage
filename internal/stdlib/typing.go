package stdlib

import (
	"fmt"
	"strings"

	"github.com/ATSOTECK/rage/internal/runtime"
)

// InitTypingModule registers the typing module
func InitTypingModule() {
	// Register _GenericType type metatable (for List, Dict, etc.)
	genericTypeMT := &runtime.TypeMetatable{
		Name: "_GenericType",
		Methods: map[string]runtime.GoFunction{
			"__getitem__": genericTypeGetItem,
			"__repr__":    genericTypeRepr,
		},
	}
	runtime.RegisterTypeMetatable("_GenericType", genericTypeMT)

	// Register _GenericAlias type metatable
	genericAliasMT := &runtime.TypeMetatable{
		Name: "_GenericAlias",
		Methods: map[string]runtime.GoFunction{
			"__getitem__": genericAliasGetItem,
			"__repr__":    genericAliasRepr,
			"__args__":    genericAliasArgs,
			"__origin__":  genericAliasOrigin,
		},
	}
	runtime.RegisterTypeMetatable("_GenericAlias", genericAliasMT)

	// Register TypeVar type metatable
	typeVarMT := &runtime.TypeMetatable{
		Name: "TypeVar",
		Methods: map[string]runtime.GoFunction{
			"__repr__": typeVarRepr,
		},
	}
	runtime.RegisterTypeMetatable("TypeVar", typeVarMT)

	// Register _SpecialForm type metatable for Union, Optional, etc.
	specialFormMT := &runtime.TypeMetatable{
		Name: "_SpecialForm",
		Methods: map[string]runtime.GoFunction{
			"__getitem__": specialFormGetItem,
			"__repr__":    specialFormRepr,
		},
	}
	runtime.RegisterTypeMetatable("_SpecialForm", specialFormMT)

	// Register ParamSpec type metatable
	paramSpecMT := &runtime.TypeMetatable{
		Name: "ParamSpec",
		Methods: map[string]runtime.GoFunction{
			"__repr__": paramSpecRepr,
		},
	}
	runtime.RegisterTypeMetatable("ParamSpec", paramSpecMT)

	// Register TypeVarTuple type metatable
	typeVarTupleMT := &runtime.TypeMetatable{
		Name: "TypeVarTuple",
		Methods: map[string]runtime.GoFunction{
			"__repr__": typeVarTupleRepr,
		},
	}
	runtime.RegisterTypeMetatable("TypeVarTuple", typeVarTupleMT)

	runtime.NewModuleBuilder("typing").
		Doc("Support for type hints.").
		// Special forms (subscriptable)
		Const("Any", createSpecialForm("Any")).
		Const("Union", createSpecialForm("Union")).
		Const("Optional", createSpecialForm("Optional")).
		Const("Callable", createSpecialForm("Callable")).
		Const("Literal", createSpecialForm("Literal")).
		Const("Final", createSpecialForm("Final")).
		Const("ClassVar", createSpecialForm("ClassVar")).
		Const("Annotated", createSpecialForm("Annotated")).
		Const("TypeGuard", createSpecialForm("TypeGuard")).
		Const("Concatenate", createSpecialForm("Concatenate")).
		Const("NoReturn", createSpecialForm("NoReturn")).
		Const("Never", createSpecialForm("Never")).
		Const("Self", createSpecialForm("Self")).
		Const("Unpack", createSpecialForm("Unpack")).
		Const("Required", createSpecialForm("Required")).
		Const("NotRequired", createSpecialForm("NotRequired")).
		Const("ReadOnly", createSpecialForm("ReadOnly")).
		// Generic type aliases (subscriptable with __class_getitem__)
		Const("List", createGenericType("List", "list")).
		Const("Dict", createGenericType("Dict", "dict")).
		Const("Set", createGenericType("Set", "set")).
		Const("FrozenSet", createGenericType("FrozenSet", "frozenset")).
		Const("Tuple", createGenericType("Tuple", "tuple")).
		Const("Type", createGenericType("Type", "type")).
		Const("Sequence", createGenericType("Sequence", "Sequence")).
		Const("Mapping", createGenericType("Mapping", "Mapping")).
		Const("MutableMapping", createGenericType("MutableMapping", "MutableMapping")).
		Const("MutableSequence", createGenericType("MutableSequence", "MutableSequence")).
		Const("MutableSet", createGenericType("MutableSet", "MutableSet")).
		Const("Iterable", createGenericType("Iterable", "Iterable")).
		Const("Iterator", createGenericType("Iterator", "Iterator")).
		Const("Generator", createGenericType("Generator", "Generator")).
		Const("Coroutine", createGenericType("Coroutine", "Coroutine")).
		Const("AsyncGenerator", createGenericType("AsyncGenerator", "AsyncGenerator")).
		Const("AsyncIterable", createGenericType("AsyncIterable", "AsyncIterable")).
		Const("AsyncIterator", createGenericType("AsyncIterator", "AsyncIterator")).
		Const("Awaitable", createGenericType("Awaitable", "Awaitable")).
		Const("ContextManager", createGenericType("ContextManager", "ContextManager")).
		Const("AsyncContextManager", createGenericType("AsyncContextManager", "AsyncContextManager")).
		Const("Pattern", createGenericType("Pattern", "Pattern")).
		Const("Match", createGenericType("Match", "Match")).
		Const("IO", createGenericType("IO", "IO")).
		Const("TextIO", createGenericType("TextIO", "TextIO")).
		Const("BinaryIO", createGenericType("BinaryIO", "BinaryIO")).
		// Callable types
		Const("Hashable", createGenericType("Hashable", "Hashable")).
		Const("Sized", createGenericType("Sized", "Sized")).
		Const("Reversible", createGenericType("Reversible", "Reversible")).
		Const("SupportsInt", createGenericType("SupportsInt", "SupportsInt")).
		Const("SupportsFloat", createGenericType("SupportsFloat", "SupportsFloat")).
		Const("SupportsComplex", createGenericType("SupportsComplex", "SupportsComplex")).
		Const("SupportsBytes", createGenericType("SupportsBytes", "SupportsBytes")).
		Const("SupportsAbs", createGenericType("SupportsAbs", "SupportsAbs")).
		Const("SupportsRound", createGenericType("SupportsRound", "SupportsRound")).
		Const("SupportsIndex", createGenericType("SupportsIndex", "SupportsIndex")).
		// Counter, ChainMap, OrderedDict, defaultdict aliases
		Const("Counter", createGenericType("Counter", "Counter")).
		Const("ChainMap", createGenericType("ChainMap", "ChainMap")).
		Const("OrderedDict", createGenericType("OrderedDict", "OrderedDict")).
		Const("DefaultDict", createGenericType("DefaultDict", "defaultdict")).
		Const("Deque", createGenericType("Deque", "deque")).
		// Type creation functions
		Func("TypeVar", typingTypeVar).
		Func("Generic", typingGeneric).
		Func("Protocol", typingProtocol).
		Func("ParamSpec", typingParamSpec).
		Func("TypeVarTuple", typingTypeVarTuple).
		Func("NewType", typingNewType).
		Func("NamedTuple", typingNamedTuple).
		Func("TypedDict", typingTypedDict).
		// Utility functions
		Func("cast", typingCast).
		Func("get_type_hints", typingGetTypeHints).
		Func("get_origin", typingGetOrigin).
		Func("get_args", typingGetArgs).
		Func("is_typeddict", typingIsTypedDict).
		Func("overload", typingOverload).
		Func("final", typingFinal).
		Func("no_type_check", typingNoTypeCheck).
		Func("no_type_check_decorator", typingNoTypeCheckDecorator).
		Func("runtime_checkable", typingRuntimeCheckable).
		Func("dataclass_transform", typingDataclassTransform).
		Func("override", typingOverride).
		Func("reveal_type", typingRevealType).
		Func("assert_type", typingAssertType).
		Func("assert_never", typingAssertNever).
		Func("clear_overloads", typingClearOverloads).
		Func("get_overloads", typingGetOverloads).
		// Constants
		Const("TYPE_CHECKING", runtime.NewBool(false)).
		Const("EXCLUDED_ATTRIBUTES", runtime.NewTuple([]runtime.Value{})).
		Register()
}

// =====================================
// _SpecialForm - for Union, Optional, etc.
// =====================================

// PySpecialForm represents a special typing form like Union, Optional
type PySpecialForm struct {
	Name string
}

func (s *PySpecialForm) Type() string   { return "_SpecialForm" }
func (s *PySpecialForm) String() string { return fmt.Sprintf("typing.%s", s.Name) }

func createSpecialForm(name string) runtime.Value {
	form := &PySpecialForm{Name: name}
	ud := runtime.NewUserData(form)
	ud.Metatable = runtime.NewDict()
	ud.Metatable.Items[runtime.NewString("__type__")] = runtime.NewString("_SpecialForm")
	return ud
}

func specialFormGetItem(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected _SpecialForm object")
		return 0
	}
	form, ok := ud.Value.(*PySpecialForm)
	if !ok {
		vm.RaiseError("expected _SpecialForm object")
		return 0
	}

	// Get the type argument(s)
	arg := vm.Get(2)

	// Create a _GenericAlias for the subscripted form
	alias := &PyGenericAlias{
		Origin: form.Name,
		Args:   extractTypeArgs(arg),
	}

	udResult := runtime.NewUserData(alias)
	udResult.Metatable = runtime.NewDict()
	udResult.Metatable.Items[runtime.NewString("__type__")] = runtime.NewString("_GenericAlias")
	vm.Push(udResult)
	return 1
}

func specialFormRepr(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected _SpecialForm object")
		return 0
	}
	form, ok := ud.Value.(*PySpecialForm)
	if !ok {
		vm.RaiseError("expected _SpecialForm object")
		return 0
	}

	vm.Push(runtime.NewString(fmt.Sprintf("typing.%s", form.Name)))
	return 1
}

// =====================================
// _GenericAlias - for parameterized types
// =====================================

// PyGenericAlias represents a parameterized generic type like List[int]
type PyGenericAlias struct {
	Origin string
	Args   []runtime.Value
}

func (g *PyGenericAlias) Type() string { return "_GenericAlias" }
func (g *PyGenericAlias) String() string {
	if len(g.Args) == 0 {
		return fmt.Sprintf("typing.%s", g.Origin)
	}
	args := make([]string, len(g.Args))
	for i, arg := range g.Args {
		args[i] = valueToTypeString(arg)
	}
	return fmt.Sprintf("typing.%s[%s]", g.Origin, strings.Join(args, ", "))
}

// PyGenericType represents a subscriptable generic type like List, Dict
type PyGenericType struct {
	Name       string
	OriginType string // The underlying type (e.g., "list" for List)
}

func (g *PyGenericType) Type() string   { return "_GenericType" }
func (g *PyGenericType) String() string { return fmt.Sprintf("typing.%s", g.Name) }

// genericTypeGetItem handles subscripting generic types like List[int]
func genericTypeGetItem(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected _GenericType object")
		return 0
	}
	gt, ok := ud.Value.(*PyGenericType)
	if !ok {
		vm.RaiseError("expected _GenericType object")
		return 0
	}

	arg := vm.Get(2)
	alias := &PyGenericAlias{
		Origin: gt.Name,
		Args:   extractTypeArgs(arg),
	}

	udResult := runtime.NewUserData(alias)
	udResult.Metatable = runtime.NewDict()
	udResult.Metatable.Items[runtime.NewString("__type__")] = runtime.NewString("_GenericAlias")
	vm.Push(udResult)
	return 1
}

// genericTypeRepr returns the string representation of a generic type
func genericTypeRepr(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected _GenericType object")
		return 0
	}
	gt, ok := ud.Value.(*PyGenericType)
	if !ok {
		vm.RaiseError("expected _GenericType object")
		return 0
	}

	vm.Push(runtime.NewString(gt.String()))
	return 1
}

func createGenericType(name, origin string) runtime.Value {
	gt := &PyGenericType{Name: name, OriginType: origin}
	ud := runtime.NewUserData(gt)
	ud.Metatable = runtime.NewDict()
	ud.Metatable.Items[runtime.NewString("__type__")] = runtime.NewString("_GenericType")
	return ud
}

func genericAliasGetItem(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected _GenericAlias object")
		return 0
	}
	alias, ok := ud.Value.(*PyGenericAlias)
	if !ok {
		// Could be a PyGenericType
		if gt, ok := ud.Value.(*PyGenericType); ok {
			arg := vm.Get(2)
			newAlias := &PyGenericAlias{
				Origin: gt.Name,
				Args:   extractTypeArgs(arg),
			}
			udResult := runtime.NewUserData(newAlias)
			udResult.Metatable = runtime.NewDict()
			udResult.Metatable.Items[runtime.NewString("__type__")] = runtime.NewString("_GenericAlias")
			vm.Push(udResult)
			return 1
		}
		vm.RaiseError("expected _GenericAlias object")
		return 0
	}

	// Further subscripting
	arg := vm.Get(2)
	newAlias := &PyGenericAlias{
		Origin: alias.Origin,
		Args:   extractTypeArgs(arg),
	}

	udResult := runtime.NewUserData(newAlias)
	udResult.Metatable = runtime.NewDict()
	udResult.Metatable.Items[runtime.NewString("__type__")] = runtime.NewString("_GenericAlias")
	vm.Push(udResult)
	return 1
}

func genericAliasRepr(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected _GenericAlias object")
		return 0
	}
	alias, ok := ud.Value.(*PyGenericAlias)
	if !ok {
		if gt, ok := ud.Value.(*PyGenericType); ok {
			vm.Push(runtime.NewString(gt.String()))
			return 1
		}
		vm.RaiseError("expected _GenericAlias object")
		return 0
	}

	vm.Push(runtime.NewString(alias.String()))
	return 1
}

func genericAliasArgs(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected _GenericAlias object")
		return 0
	}
	alias, ok := ud.Value.(*PyGenericAlias)
	if !ok {
		vm.Push(runtime.NewTuple([]runtime.Value{}))
		return 1
	}

	vm.Push(runtime.NewTuple(alias.Args))
	return 1
}

func genericAliasOrigin(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected _GenericAlias object")
		return 0
	}
	alias, ok := ud.Value.(*PyGenericAlias)
	if !ok {
		if gt, ok := ud.Value.(*PyGenericType); ok {
			vm.Push(runtime.NewString(gt.OriginType))
			return 1
		}
		vm.RaiseError("expected _GenericAlias object")
		return 0
	}

	vm.Push(runtime.NewString(alias.Origin))
	return 1
}

// =====================================
// TypeVar
// =====================================

// PyTypeVar represents a type variable
type PyTypeVar struct {
	Name       string
	Bound      runtime.Value
	Covariant  bool
	Contravariant bool
	Constraints []runtime.Value
}

func (t *PyTypeVar) Type() string   { return "TypeVar" }
func (t *PyTypeVar) String() string { return fmt.Sprintf("~%s", t.Name) }

// typing.TypeVar(name, *constraints, bound=None, covariant=False, contravariant=False)
func typingTypeVar(vm *runtime.VM) int {
	if vm.GetTop() < 1 {
		vm.RaiseError("TypeVar() requires at least 1 argument")
		return 0
	}

	name := vm.CheckString(1)

	typeVar := &PyTypeVar{
		Name:        name,
		Bound:       runtime.None,
		Covariant:   false,
		Contravariant: false,
		Constraints: []runtime.Value{},
	}

	// Collect constraints (positional arguments after name)
	for i := 2; i <= vm.GetTop(); i++ {
		arg := vm.Get(i)
		// Skip if it's a keyword argument (handled separately in real implementation)
		typeVar.Constraints = append(typeVar.Constraints, arg)
	}

	ud := runtime.NewUserData(typeVar)
	ud.Metatable = runtime.NewDict()
	ud.Metatable.Items[runtime.NewString("__type__")] = runtime.NewString("TypeVar")
	ud.Metatable.Items[runtime.NewString("__name__")] = runtime.NewString(name)
	vm.Push(ud)
	return 1
}

func typeVarRepr(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected TypeVar object")
		return 0
	}
	tv, ok := ud.Value.(*PyTypeVar)
	if !ok {
		vm.RaiseError("expected TypeVar object")
		return 0
	}

	vm.Push(runtime.NewString(fmt.Sprintf("~%s", tv.Name)))
	return 1
}

// =====================================
// ParamSpec
// =====================================

// PyParamSpec represents a parameter specification variable
type PyParamSpec struct {
	Name string
}

func (p *PyParamSpec) Type() string   { return "ParamSpec" }
func (p *PyParamSpec) String() string { return fmt.Sprintf("~%s", p.Name) }

func typingParamSpec(vm *runtime.VM) int {
	if vm.GetTop() < 1 {
		vm.RaiseError("ParamSpec() requires at least 1 argument")
		return 0
	}

	name := vm.CheckString(1)

	ps := &PyParamSpec{Name: name}

	ud := runtime.NewUserData(ps)
	ud.Metatable = runtime.NewDict()
	ud.Metatable.Items[runtime.NewString("__type__")] = runtime.NewString("ParamSpec")
	ud.Metatable.Items[runtime.NewString("__name__")] = runtime.NewString(name)
	vm.Push(ud)
	return 1
}

func paramSpecRepr(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected ParamSpec object")
		return 0
	}
	ps, ok := ud.Value.(*PyParamSpec)
	if !ok {
		vm.RaiseError("expected ParamSpec object")
		return 0
	}

	vm.Push(runtime.NewString(fmt.Sprintf("~%s", ps.Name)))
	return 1
}

// =====================================
// TypeVarTuple
// =====================================

// PyTypeVarTuple represents a type variable tuple
type PyTypeVarTuple struct {
	Name string
}

func (t *PyTypeVarTuple) Type() string   { return "TypeVarTuple" }
func (t *PyTypeVarTuple) String() string { return fmt.Sprintf("*%s", t.Name) }

func typingTypeVarTuple(vm *runtime.VM) int {
	if vm.GetTop() < 1 {
		vm.RaiseError("TypeVarTuple() requires at least 1 argument")
		return 0
	}

	name := vm.CheckString(1)

	tvt := &PyTypeVarTuple{Name: name}

	ud := runtime.NewUserData(tvt)
	ud.Metatable = runtime.NewDict()
	ud.Metatable.Items[runtime.NewString("__type__")] = runtime.NewString("TypeVarTuple")
	ud.Metatable.Items[runtime.NewString("__name__")] = runtime.NewString(name)
	vm.Push(ud)
	return 1
}

func typeVarTupleRepr(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected TypeVarTuple object")
		return 0
	}
	tvt, ok := ud.Value.(*PyTypeVarTuple)
	if !ok {
		vm.RaiseError("expected TypeVarTuple object")
		return 0
	}

	vm.Push(runtime.NewString(fmt.Sprintf("*%s", tvt.Name)))
	return 1
}

// =====================================
// Generic base class
// =====================================

// typing.Generic - base class for generic types
func typingGeneric(vm *runtime.VM) int {
	// Return a marker object that can be used as a base class
	ud := runtime.NewUserData(&struct{ Name string }{Name: "Generic"})
	ud.Metatable = runtime.NewDict()
	ud.Metatable.Items[runtime.NewString("__type__")] = runtime.NewString("_GenericBase")
	vm.Push(ud)
	return 1
}

// =====================================
// Protocol
// =====================================

// typing.Protocol - base class for protocol classes
func typingProtocol(vm *runtime.VM) int {
	ud := runtime.NewUserData(&struct{ Name string }{Name: "Protocol"})
	ud.Metatable = runtime.NewDict()
	ud.Metatable.Items[runtime.NewString("__type__")] = runtime.NewString("_ProtocolBase")
	vm.Push(ud)
	return 1
}

// =====================================
// NewType
// =====================================

// typing.NewType(name, tp) - creates a distinct type
func typingNewType(vm *runtime.VM) int {
	if vm.GetTop() < 2 {
		vm.RaiseError("NewType() requires 2 arguments")
		return 0
	}

	name := vm.CheckString(1)
	tp := vm.Get(2)

	// Create a callable that just returns its argument (NewType is a no-op at runtime)
	newType := runtime.NewGoFunction(name, func(vm *runtime.VM) int {
		if vm.GetTop() < 1 {
			vm.RaiseError("%s() requires 1 argument", name)
			return 0
		}
		vm.Push(vm.Get(1))
		return 1
	})

	// Store the supertype info in the function metadata (if needed)
	_ = tp

	vm.Push(newType)
	return 1
}

// =====================================
// NamedTuple
// =====================================

// typing.NamedTuple(typename, fields) - creates a named tuple type
func typingNamedTuple(vm *runtime.VM) int {
	if vm.GetTop() < 1 {
		vm.RaiseError("NamedTuple() requires at least 1 argument")
		return 0
	}

	typename := vm.CheckString(1)

	var fieldNames []string

	// Get field names from second argument if provided
	if vm.GetTop() >= 2 && !runtime.IsNone(vm.Get(2)) {
		arg := vm.Get(2)
		switch v := arg.(type) {
		case *runtime.PyList:
			for _, item := range v.Items {
				// Each item should be a tuple (name, type)
				if tuple, ok := item.(*runtime.PyTuple); ok && len(tuple.Items) >= 1 {
					if s, ok := tuple.Items[0].(*runtime.PyString); ok {
						fieldNames = append(fieldNames, s.Value)
					}
				}
			}
		case *runtime.PyDict:
			for k := range v.Items {
				if s, ok := k.(*runtime.PyString); ok {
					fieldNames = append(fieldNames, s.Value)
				}
			}
		}
	}

	// Create a factory function that creates named tuples
	factory := runtime.NewGoFunction(typename, func(vm *runtime.VM) int {
		items := make([]runtime.Value, len(fieldNames))
		for i := 0; i < len(fieldNames); i++ {
			if i+1 <= vm.GetTop() {
				items[i] = vm.Get(i + 1)
			} else {
				items[i] = runtime.None
			}
		}

		// Create a dict-like structure with _fields
		result := runtime.NewDict()
		result.Items[runtime.NewString("_fields")] = runtime.NewTuple(func() []runtime.Value {
			fields := make([]runtime.Value, len(fieldNames))
			for i, name := range fieldNames {
				fields[i] = runtime.NewString(name)
			}
			return fields
		}())

		for i, name := range fieldNames {
			result.Items[runtime.NewString(name)] = items[i]
		}

		vm.Push(result)
		return 1
	})

	vm.Push(factory)
	return 1
}

// =====================================
// TypedDict
// =====================================

// typing.TypedDict(typename, fields) - creates a typed dictionary class
func typingTypedDict(vm *runtime.VM) int {
	if vm.GetTop() < 1 {
		vm.RaiseError("TypedDict() requires at least 1 argument")
		return 0
	}

	typename := vm.CheckString(1)

	// Create a factory that returns a dict
	factory := runtime.NewGoFunction(typename, func(vm *runtime.VM) int {
		result := runtime.NewDict()

		// Copy fields from argument if provided
		if vm.GetTop() >= 1 {
			if dict, ok := vm.Get(1).(*runtime.PyDict); ok {
				for k, v := range dict.Items {
					result.Items[k] = v
				}
			}
		}

		vm.Push(result)
		return 1
	})

	vm.Push(factory)
	return 1
}

// =====================================
// Utility functions
// =====================================

// typing.cast(typ, val) - cast a value to a type (no-op at runtime)
func typingCast(vm *runtime.VM) int {
	if vm.GetTop() < 2 {
		vm.RaiseError("cast() requires 2 arguments")
		return 0
	}

	// Return the value unchanged - cast is a no-op at runtime
	vm.Push(vm.Get(2))
	return 1
}

// typing.get_type_hints(obj) - get type hints for an object
func typingGetTypeHints(vm *runtime.VM) int {
	// Return an empty dict - actual type hints require introspection
	vm.Push(runtime.NewDict())
	return 1
}

// typing.get_origin(tp) - get the origin of a generic type
func typingGetOrigin(vm *runtime.VM) int {
	if vm.GetTop() < 1 {
		vm.Push(runtime.None)
		return 1
	}

	arg := vm.Get(1)
	if ud, ok := arg.(*runtime.PyUserData); ok {
		if alias, ok := ud.Value.(*PyGenericAlias); ok {
			vm.Push(runtime.NewString(alias.Origin))
			return 1
		}
		if gt, ok := ud.Value.(*PyGenericType); ok {
			vm.Push(runtime.NewString(gt.OriginType))
			return 1
		}
	}

	vm.Push(runtime.None)
	return 1
}

// typing.get_args(tp) - get the type arguments of a generic type
func typingGetArgs(vm *runtime.VM) int {
	if vm.GetTop() < 1 {
		vm.Push(runtime.NewTuple([]runtime.Value{}))
		return 1
	}

	arg := vm.Get(1)
	if ud, ok := arg.(*runtime.PyUserData); ok {
		if alias, ok := ud.Value.(*PyGenericAlias); ok {
			vm.Push(runtime.NewTuple(alias.Args))
			return 1
		}
	}

	vm.Push(runtime.NewTuple([]runtime.Value{}))
	return 1
}

// typing.is_typeddict(tp) - check if a type is a TypedDict
func typingIsTypedDict(vm *runtime.VM) int {
	// For now, return False - would need more sophisticated type introspection
	vm.Push(runtime.NewBool(false))
	return 1
}

// typing.overload - decorator for overloaded functions (no-op at runtime)
func typingOverload(vm *runtime.VM) int {
	if vm.GetTop() < 1 {
		vm.RaiseError("overload() requires 1 argument")
		return 0
	}
	// Return the function unchanged
	vm.Push(vm.Get(1))
	return 1
}

// typing.final - decorator marking a method/class as final
func typingFinal(vm *runtime.VM) int {
	if vm.GetTop() < 1 {
		vm.RaiseError("final() requires 1 argument")
		return 0
	}
	// Return the argument unchanged
	vm.Push(vm.Get(1))
	return 1
}

// typing.no_type_check - decorator to skip type checking
func typingNoTypeCheck(vm *runtime.VM) int {
	if vm.GetTop() < 1 {
		vm.RaiseError("no_type_check() requires 1 argument")
		return 0
	}
	// Return the argument unchanged
	vm.Push(vm.Get(1))
	return 1
}

// typing.no_type_check_decorator - decorator factory
func typingNoTypeCheckDecorator(vm *runtime.VM) int {
	if vm.GetTop() < 1 {
		vm.RaiseError("no_type_check_decorator() requires 1 argument")
		return 0
	}
	// Return the argument unchanged
	vm.Push(vm.Get(1))
	return 1
}

// typing.runtime_checkable - decorator for runtime-checkable protocols
func typingRuntimeCheckable(vm *runtime.VM) int {
	if vm.GetTop() < 1 {
		vm.RaiseError("runtime_checkable() requires 1 argument")
		return 0
	}
	// Return the argument unchanged
	vm.Push(vm.Get(1))
	return 1
}

// typing.dataclass_transform - decorator for dataclass-like transforms
func typingDataclassTransform(vm *runtime.VM) int {
	// Return an identity decorator
	decorator := runtime.NewGoFunction("dataclass_transform", func(vm *runtime.VM) int {
		if vm.GetTop() < 1 {
			vm.RaiseError("dataclass_transform() decorator requires 1 argument")
			return 0
		}
		vm.Push(vm.Get(1))
		return 1
	})
	vm.Push(decorator)
	return 1
}

// typing.override - decorator marking a method as an override
func typingOverride(vm *runtime.VM) int {
	if vm.GetTop() < 1 {
		vm.RaiseError("override() requires 1 argument")
		return 0
	}
	// Return the argument unchanged
	vm.Push(vm.Get(1))
	return 1
}

// typing.reveal_type(obj) - reveal the inferred type of an expression
func typingRevealType(vm *runtime.VM) int {
	if vm.GetTop() < 1 {
		vm.RaiseError("reveal_type() requires 1 argument")
		return 0
	}
	// At runtime, just return the value
	vm.Push(vm.Get(1))
	return 1
}

// typing.assert_type(val, typ) - assert that val has the specified type
func typingAssertType(vm *runtime.VM) int {
	if vm.GetTop() < 2 {
		vm.RaiseError("assert_type() requires 2 arguments")
		return 0
	}
	// At runtime, just return the value
	vm.Push(vm.Get(1))
	return 1
}

// typing.assert_never(arg) - assert that code is unreachable
func typingAssertNever(vm *runtime.VM) int {
	vm.RaiseError("Expected code to be unreachable")
	return 0
}

// typing.clear_overloads() - clear all overload definitions
func typingClearOverloads(vm *runtime.VM) int {
	// No-op at runtime
	return 0
}

// typing.get_overloads(func) - get all overload definitions for a function
func typingGetOverloads(vm *runtime.VM) int {
	// Return empty list - overloads aren't tracked at runtime
	vm.Push(runtime.NewList([]runtime.Value{}))
	return 1
}

// =====================================
// Helper functions
// =====================================

// extractTypeArgs converts a subscript argument to a slice of type arguments
func extractTypeArgs(arg runtime.Value) []runtime.Value {
	switch v := arg.(type) {
	case *runtime.PyTuple:
		return v.Items
	default:
		return []runtime.Value{arg}
	}
}

// valueToTypeString converts a Value to a string representation for types
func valueToTypeString(v runtime.Value) string {
	switch val := v.(type) {
	case *runtime.PyString:
		return val.Value
	case *runtime.PyClass:
		return val.Name
	case *runtime.PyUserData:
		if alias, ok := val.Value.(*PyGenericAlias); ok {
			return alias.String()
		}
		if gt, ok := val.Value.(*PyGenericType); ok {
			return gt.String()
		}
		if sf, ok := val.Value.(*PySpecialForm); ok {
			return sf.String()
		}
		if tv, ok := val.Value.(*PyTypeVar); ok {
			return tv.String()
		}
		return fmt.Sprintf("%v", val.Value)
	case *runtime.PyNone:
		return "None"
	case *runtime.PyBuiltinFunc:
		return val.Name
	default:
		return fmt.Sprintf("%v", v)
	}
}
