package stdlib

import (
	"fmt"

	"github.com/ATSOTECK/rage/internal/runtime"
)

// InitCollectionsModule registers the collections module
func InitCollectionsModule() {
	// Register deque type metatable
	dequeMT := &runtime.TypeMetatable{
		Name: "deque",
		Methods: map[string]runtime.GoFunction{
			"append":      dequeAppend,
			"appendleft":  dequeAppendLeft,
			"pop":         dequePop,
			"popleft":     dequePopLeft,
			"extend":      dequeExtend,
			"extendleft":  dequeExtendLeft,
			"rotate":      dequeRotate,
			"clear":       dequeClear,
			"count":       dequeCount,
			"index":       dequeIndex,
			"insert":      dequeInsert,
			"remove":      dequeRemove,
			"reverse":     dequeReverse,
			"copy":        dequeCopy,
			"maxlen":      dequeMaxlen,
			"__len__":     dequeLen,
			"__iter__":    dequeIter,
			"__getitem__": dequeGetItem,
			"__repr__":    dequeRepr,
		},
	}
	runtime.RegisterTypeMetatable("deque", dequeMT)

	runtime.NewModuleBuilder("collections").
		Doc("Container datatypes.").
		// Types (constructors)
		Func("Counter", collectionsDictKind(runtime.DictCounter)).
		Func("deque", collectionsDeque).
		Func("defaultdict", collectionsDictKind(runtime.DictDefault)).
		Func("OrderedDict", collectionsDictKind(runtime.DictOrdered)).
		// namedtuple factory
		Func("namedtuple", collectionsNamedtuple).
		Register()
}

// =====================================
// Counter, defaultdict, OrderedDict
// =====================================

// collectionsDictKind returns the constructor for one of the dict subclasses
// (Counter, defaultdict, OrderedDict). They are real dicts tagged with a kind
// (see runtime/dict_kinds.go), so every dict operation works on them.
func collectionsDictKind(kind runtime.DictKind) runtime.GoFunction {
	return func(vm *runtime.VM) int {
		args := make([]runtime.Value, vm.GetTop())
		for i := range args {
			args[i] = vm.Get(i + 1)
		}
		d, err := vm.NewCollectionsDict(kind, args, vm.Kwargs())
		if err != nil {
			vm.RaiseError("%s", err.Error())
			return 0
		}
		vm.Push(d)
		return 1
	}
}

// =====================================
// deque
// =====================================

// PyDeque is a double-ended queue
type PyDeque struct {
	Items  []runtime.Value
	Maxlen int64 // -1 means unbounded
}

func (d *PyDeque) Type() string   { return "deque" }
func (d *PyDeque) String() string { return fmt.Sprintf("deque(%v)", d.Items) }

// collections.deque([iterable[, maxlen]])
func collectionsDeque(vm *runtime.VM) int {
	deque := &PyDeque{
		Items:  make([]runtime.Value, 0),
		Maxlen: -1,
	}

	// Check for maxlen parameter
	if vm.GetTop() >= 2 && !runtime.IsNone(vm.Get(2)) {
		deque.Maxlen = vm.CheckInt(2)
	}
	if m, ok := vm.Kwarg("maxlen"); ok && !runtime.IsNone(m) {
		n, ok := m.(*runtime.PyInt)
		if !ok {
			vm.RaiseError("TypeError: an integer is required")
			return 0
		}
		deque.Maxlen = n.Value
	}

	// Initialize from iterable
	if vm.GetTop() >= 1 && !runtime.IsNone(vm.Get(1)) {
		arg := vm.Get(1)
		switch v := arg.(type) {
		case *runtime.PyList:
			deque.Items = append(deque.Items, v.Items...)
		case *runtime.PyTuple:
			deque.Items = append(deque.Items, v.Items...)
		case *runtime.PyString:
			for _, ch := range v.Value {
				deque.Items = append(deque.Items, runtime.NewString(string(ch)))
			}
		}
	}

	// Enforce maxlen
	if deque.Maxlen >= 0 && int64(len(deque.Items)) > deque.Maxlen {
		drop := len(deque.Items) - int(deque.Maxlen)
		clearItems(deque.Items, 0, drop) // release evicted head refs for GC
		deque.Items = deque.Items[drop:]
	}

	ud := runtime.NewUserData(deque)
	ud.Metatable = runtime.NewDict()
	ud.Metatable.Items[runtime.NewString("__type__")] = runtime.NewString("deque")
	vm.Push(ud)
	return 1
}

// clearItems nils slots [from:to) of a deque's backing array so dropped elements
// become eligible for GC. A reslice alone retains them: they stay within cap.
func clearItems(items []runtime.Value, from, to int) {
	for i := from; i < to; i++ {
		items[i] = nil
	}
}

// deque.append(x)
func dequeAppend(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected deque object")
		return 0
	}
	deque, ok := ud.Value.(*PyDeque)
	if !ok {
		vm.RaiseError("expected deque object")
		return 0
	}

	x := vm.Get(2)
	deque.Items = append(deque.Items, x)

	// Enforce maxlen
	if deque.Maxlen >= 0 && int64(len(deque.Items)) > deque.Maxlen {
		deque.Items[0] = nil // release evicted head ref for GC (slice start advances past it)
		deque.Items = deque.Items[1:]
	}

	return 0
}

// deque.appendleft(x)
func dequeAppendLeft(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected deque object")
		return 0
	}
	deque, ok := ud.Value.(*PyDeque)
	if !ok {
		vm.RaiseError("expected deque object")
		return 0
	}

	x := vm.Get(2)
	deque.Items = append([]runtime.Value{x}, deque.Items...)

	// Enforce maxlen
	if deque.Maxlen >= 0 && int64(len(deque.Items)) > deque.Maxlen {
		clearItems(deque.Items, int(deque.Maxlen), len(deque.Items)) // release evicted tail refs for GC
		deque.Items = deque.Items[:deque.Maxlen]
	}

	return 0
}

// deque.pop() -> rightmost element
func dequePop(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected deque object")
		return 0
	}
	deque, ok := ud.Value.(*PyDeque)
	if !ok {
		vm.RaiseError("expected deque object")
		return 0
	}

	if len(deque.Items) == 0 {
		vm.RaiseError("pop from an empty deque")
		return 0
	}

	item := deque.Items[len(deque.Items)-1]
	deque.Items[len(deque.Items)-1] = nil // release popped ref for GC
	deque.Items = deque.Items[:len(deque.Items)-1]

	vm.Push(item)
	return 1
}

// deque.popleft() -> leftmost element
func dequePopLeft(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected deque object")
		return 0
	}
	deque, ok := ud.Value.(*PyDeque)
	if !ok {
		vm.RaiseError("expected deque object")
		return 0
	}

	if len(deque.Items) == 0 {
		vm.RaiseError("pop from an empty deque")
		return 0
	}

	item := deque.Items[0]
	deque.Items[0] = nil // release popped ref for GC (slice start advances past it)
	deque.Items = deque.Items[1:]

	vm.Push(item)
	return 1
}

// deque.extend(iterable)
func dequeExtend(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected deque object")
		return 0
	}
	deque, ok := ud.Value.(*PyDeque)
	if !ok {
		vm.RaiseError("expected deque object")
		return 0
	}

	arg := vm.Get(2)
	switch v := arg.(type) {
	case *runtime.PyList:
		deque.Items = append(deque.Items, v.Items...)
	case *runtime.PyTuple:
		deque.Items = append(deque.Items, v.Items...)
	case *runtime.PyString:
		for _, ch := range v.Value {
			deque.Items = append(deque.Items, runtime.NewString(string(ch)))
		}
	}

	// Enforce maxlen
	if deque.Maxlen >= 0 && int64(len(deque.Items)) > deque.Maxlen {
		drop := len(deque.Items) - int(deque.Maxlen)
		clearItems(deque.Items, 0, drop) // release evicted head refs for GC
		deque.Items = deque.Items[drop:]
	}

	return 0
}

// deque.extendleft(iterable)
func dequeExtendLeft(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected deque object")
		return 0
	}
	deque, ok := ud.Value.(*PyDeque)
	if !ok {
		vm.RaiseError("expected deque object")
		return 0
	}

	var toAdd []runtime.Value
	arg := vm.Get(2)
	switch v := arg.(type) {
	case *runtime.PyList:
		toAdd = v.Items
	case *runtime.PyTuple:
		toAdd = v.Items
	case *runtime.PyString:
		for _, ch := range v.Value {
			toAdd = append(toAdd, runtime.NewString(string(ch)))
		}
	}

	// Reverse and prepend
	for i := len(toAdd) - 1; i >= 0; i-- {
		deque.Items = append([]runtime.Value{toAdd[i]}, deque.Items...)
	}

	// Enforce maxlen
	if deque.Maxlen >= 0 && int64(len(deque.Items)) > deque.Maxlen {
		clearItems(deque.Items, int(deque.Maxlen), len(deque.Items)) // release evicted tail refs for GC
		deque.Items = deque.Items[:deque.Maxlen]
	}

	return 0
}

// deque.rotate(n=1)
func dequeRotate(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected deque object")
		return 0
	}
	deque, ok := ud.Value.(*PyDeque)
	if !ok {
		vm.RaiseError("expected deque object")
		return 0
	}

	n := vm.OptionalInt(2, 1)

	if len(deque.Items) == 0 {
		return 0
	}

	// Normalize n
	length := int64(len(deque.Items))
	n = n % length
	if n < 0 {
		n += length
	}

	// Rotate right by n
	if n > 0 {
		deque.Items = append(deque.Items[length-n:], deque.Items[:length-n]...)
	}

	return 0
}

// deque.clear()
func dequeClear(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected deque object")
		return 0
	}
	deque, ok := ud.Value.(*PyDeque)
	if !ok {
		vm.RaiseError("expected deque object")
		return 0
	}

	deque.Items = make([]runtime.Value, 0)
	return 0
}

// deque.count(x)
func dequeCount(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected deque object")
		return 0
	}
	deque, ok := ud.Value.(*PyDeque)
	if !ok {
		vm.RaiseError("expected deque object")
		return 0
	}

	x := vm.Get(2)
	count := int64(0)
	for _, item := range deque.Items {
		if vm.Equal(item, x) {
			count++
		}
	}

	vm.Push(runtime.NewInt(count))
	return 1
}

// deque.index(x[, start[, stop]])
func dequeIndex(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected deque object")
		return 0
	}
	deque, ok := ud.Value.(*PyDeque)
	if !ok {
		vm.RaiseError("expected deque object")
		return 0
	}

	x := vm.Get(2)
	start := vm.OptionalInt(3, 0)
	stop := vm.OptionalInt(4, int64(len(deque.Items)))

	// Normalize negative indices
	n := int64(len(deque.Items))
	if start < 0 {
		start += n
	}
	if start < 0 {
		start = 0
	}
	if stop < 0 {
		stop += n
	}
	if stop < 0 {
		stop = 0
	}

	for i := start; i < stop && i < n; i++ {
		if vm.Equal(deque.Items[i], x) {
			vm.Push(runtime.NewInt(i))
			return 1
		}
	}

	vm.RaiseError("deque.index(x): x not in deque")
	return 0
}

// deque.insert(i, x)
func dequeInsert(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected deque object")
		return 0
	}
	deque, ok := ud.Value.(*PyDeque)
	if !ok {
		vm.RaiseError("expected deque object")
		return 0
	}

	i := int(vm.CheckInt(2))
	x := vm.Get(3)

	if deque.Maxlen >= 0 && int64(len(deque.Items)) >= deque.Maxlen {
		vm.RaiseError("IndexError: deque already at its maximum size")
		return 0
	}

	// Handle negative index
	if i < 0 {
		i = len(deque.Items) + i
	}
	if i < 0 {
		i = 0
	}
	if i > len(deque.Items) {
		i = len(deque.Items)
	}

	// Insert
	deque.Items = append(deque.Items[:i], append([]runtime.Value{x}, deque.Items[i:]...)...)

	return 0
}

// deque.remove(value)
func dequeRemove(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected deque object")
		return 0
	}
	deque, ok := ud.Value.(*PyDeque)
	if !ok {
		vm.RaiseError("expected deque object")
		return 0
	}

	x := vm.Get(2)
	for i, item := range deque.Items {
		if vm.Equal(item, x) {
			copy(deque.Items[i:], deque.Items[i+1:])
			deque.Items[len(deque.Items)-1] = nil // release moved-out tail ref for GC
			deque.Items = deque.Items[:len(deque.Items)-1]
			return 0
		}
	}

	vm.RaiseError("deque.remove(x): x not in deque")
	return 0
}

// deque.reverse()
func dequeReverse(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected deque object")
		return 0
	}
	deque, ok := ud.Value.(*PyDeque)
	if !ok {
		vm.RaiseError("expected deque object")
		return 0
	}

	for i, j := 0, len(deque.Items)-1; i < j; i, j = i+1, j-1 {
		deque.Items[i], deque.Items[j] = deque.Items[j], deque.Items[i]
	}

	return 0
}

// deque.copy()
func dequeCopy(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected deque object")
		return 0
	}
	deque, ok := ud.Value.(*PyDeque)
	if !ok {
		vm.RaiseError("expected deque object")
		return 0
	}

	newDeque := &PyDeque{
		Items:  make([]runtime.Value, len(deque.Items)),
		Maxlen: deque.Maxlen,
	}
	copy(newDeque.Items, deque.Items)

	nud := runtime.NewUserData(newDeque)
	nud.Metatable = ud.Metatable // keep the deque type's methods
	vm.Push(nud)
	return 1
}

// deque.maxlen property
func dequeMaxlen(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected deque object")
		return 0
	}
	deque, ok := ud.Value.(*PyDeque)
	if !ok {
		vm.RaiseError("expected deque object")
		return 0
	}

	if deque.Maxlen < 0 {
		vm.Push(runtime.None)
	} else {
		vm.Push(runtime.NewInt(deque.Maxlen))
	}
	return 1
}

// dequeArg returns the deque passed as argument 1, raising if it isn't one.
func dequeArg(vm *runtime.VM) *PyDeque {
	if ud := vm.ToUserData(1); ud != nil {
		if d, ok := ud.Value.(*PyDeque); ok {
			return d
		}
	}
	vm.RaiseError("TypeError: expected deque object")
	return nil
}

// deque.__len__
func dequeLen(vm *runtime.VM) int {
	d := dequeArg(vm)
	if d == nil {
		return 0
	}
	vm.Push(runtime.NewInt(int64(len(d.Items))))
	return 1
}

// deque.__iter__ iterates over a snapshot of the items.
func dequeIter(vm *runtime.VM) int {
	d := dequeArg(vm)
	if d == nil {
		return 0
	}
	items := make([]runtime.Value, len(d.Items))
	copy(items, d.Items)
	vm.Push(&runtime.PyIterator{Items: items})
	return 1
}

// deque.__getitem__ (integer indexes only, as in CPython)
func dequeGetItem(vm *runtime.VM) int {
	d := dequeArg(vm)
	if d == nil {
		return 0
	}
	i, ok := vm.Get(2).(*runtime.PyInt)
	if !ok {
		vm.RaiseError("TypeError: sequence index must be integer, not '%s'", vm.TypeNameOf(vm.Get(2)))
		return 0
	}
	idx := i.Value
	if idx < 0 {
		idx += int64(len(d.Items))
	}
	if idx < 0 || idx >= int64(len(d.Items)) {
		vm.RaiseError("IndexError: deque index out of range")
		return 0
	}
	vm.Push(d.Items[idx])
	return 1
}

// deque.__repr__: deque([1, 2]) or deque([1, 2], maxlen=3)
func dequeRepr(vm *runtime.VM) int {
	d := dequeArg(vm)
	if d == nil {
		return 0
	}
	s := "deque(" + vm.Repr(runtime.NewList(d.Items))
	if d.Maxlen >= 0 {
		s += fmt.Sprintf(", maxlen=%d", d.Maxlen)
	}
	vm.Push(runtime.NewString(s + ")"))
	return 1
}

// =====================================
// namedtuple
// =====================================

// collections.namedtuple(typename, field_names)
func collectionsNamedtuple(vm *runtime.VM) int {
	typename := vm.CheckString(1)
	arg := vm.Get(2)

	var fieldNames []string
	switch v := arg.(type) {
	case *runtime.PyString:
		// Split by comma or space
		current := ""
		for _, ch := range v.Value {
			if ch == ',' || ch == ' ' {
				if current != "" {
					fieldNames = append(fieldNames, current)
					current = ""
				}
			} else {
				current += string(ch)
			}
		}
		if current != "" {
			fieldNames = append(fieldNames, current)
		}
	case *runtime.PyList:
		for _, item := range v.Items {
			s, ok := item.(*runtime.PyString)
			if !ok {
				vm.RaiseError("TypeError: Field names must be strings, not '%s'", vm.TypeNameOf(item))
				return 0
			}
			fieldNames = append(fieldNames, s.Value)
		}
	case *runtime.PyTuple:
		for _, item := range v.Items {
			s, ok := item.(*runtime.PyString)
			if !ok {
				vm.RaiseError("TypeError: Field names must be strings, not '%s'", vm.TypeNameOf(item))
				return 0
			}
			fieldNames = append(fieldNames, s.Value)
		}
	}

	var defaults []runtime.Value
	if d, ok := vm.Kwarg("defaults"); ok && !runtime.IsNone(d) {
		items, err := vm.ToList(d)
		if err != nil {
			vm.RaiseError("%s", err.Error())
			return 0
		}
		defaults = items
	}

	cls, err := vm.NewNamedTupleClass(typename, fieldNames, defaults)
	if err != nil {
		vm.RaiseError("%s", err.Error())
		return 0
	}
	vm.Push(cls)
	return 1
}
