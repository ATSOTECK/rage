package stdlib

import (
	"fmt"

	"github.com/ATSOTECK/oink/internal/runtime"
)

// InitCollectionsModule registers the collections module
func InitCollectionsModule() {
	// Register Counter type metatable
	counterMT := &runtime.TypeMetatable{
		Name: "Counter",
		Methods: map[string]runtime.GoFunction{
			"most_common": counterMostCommon,
			"elements":    counterElements,
			"update":      counterUpdate,
			"subtract":    counterSubtract,
			"total":       counterTotal,
		},
	}
	runtime.RegisterTypeMetatable("Counter", counterMT)

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
		},
	}
	runtime.RegisterTypeMetatable("deque", dequeMT)

	// Register defaultdict type metatable
	defaultdictMT := &runtime.TypeMetatable{
		Name: "defaultdict",
		Methods: map[string]runtime.GoFunction{
			"default_factory": defaultdictGetFactory,
		},
	}
	runtime.RegisterTypeMetatable("defaultdict", defaultdictMT)

	runtime.NewModuleBuilder("collections").
		Doc("Container datatypes.").
		// Types (constructors)
		Func("Counter", collectionsCounter).
		Func("deque", collectionsDeque).
		Func("defaultdict", collectionsDefaultdict).
		// OrderedDict is just a regular dict in Python 3.7+
		Func("OrderedDict", collectionsOrderedDict).
		// namedtuple factory
		Func("namedtuple", collectionsNamedtuple).
		Register()
}

// =====================================
// Counter
// =====================================

// PyCounter wraps a dict that counts hashable objects
type PyCounter struct {
	Items map[runtime.Value]int64
}

func (c *PyCounter) Type() string   { return "Counter" }
func (c *PyCounter) String() string { return fmt.Sprintf("Counter(%v)", c.Items) }

// collections.Counter([iterable-or-mapping])
func collectionsCounter(vm *runtime.VM) int {
	counter := &PyCounter{Items: make(map[runtime.Value]int64)}

	if vm.GetTop() >= 1 && !runtime.IsNone(vm.Get(1)) {
		arg := vm.Get(1)
		switch v := arg.(type) {
		case *runtime.PyList:
			for _, item := range v.Items {
				counter.Items[item]++
			}
		case *runtime.PyTuple:
			for _, item := range v.Items {
				counter.Items[item]++
			}
		case *runtime.PyString:
			for _, ch := range v.Value {
				counter.Items[runtime.NewString(string(ch))]++
			}
		case *runtime.PyDict:
			for k, val := range v.Items {
				if intVal, ok := val.(*runtime.PyInt); ok {
					counter.Items[k] = intVal.Value
				}
			}
		}
	}

	ud := runtime.NewUserData(counter)
	ud.Metatable = runtime.NewDict()
	ud.Metatable.Items[runtime.NewString("__type__")] = runtime.NewString("Counter")
	vm.Push(ud)
	return 1
}

// Counter.most_common([n]) -> list of (elem, count) pairs
func counterMostCommon(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected Counter object")
		return 0
	}
	counter, ok := ud.Value.(*PyCounter)
	if !ok {
		vm.RaiseError("expected Counter object")
		return 0
	}

	// Get n parameter (default: all)
	n := int64(-1)
	if vm.GetTop() >= 2 && !runtime.IsNone(vm.Get(2)) {
		n = vm.CheckInt(2)
	}

	// Build list of (elem, count) pairs sorted by count
	type elemCount struct {
		elem  runtime.Value
		count int64
	}
	pairs := make([]elemCount, 0, len(counter.Items))
	for k, v := range counter.Items {
		pairs = append(pairs, elemCount{k, v})
	}

	// Sort by count (descending)
	for i := 0; i < len(pairs)-1; i++ {
		for j := i + 1; j < len(pairs); j++ {
			if pairs[j].count > pairs[i].count {
				pairs[i], pairs[j] = pairs[j], pairs[i]
			}
		}
	}

	// Limit to n
	if n >= 0 && int(n) < len(pairs) {
		pairs = pairs[:n]
	}

	// Convert to list of tuples
	result := make([]runtime.Value, len(pairs))
	for i, p := range pairs {
		result[i] = runtime.NewTuple([]runtime.Value{p.elem, runtime.NewInt(p.count)})
	}

	vm.Push(runtime.NewList(result))
	return 1
}

// Counter.elements() -> iterator over elements
func counterElements(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected Counter object")
		return 0
	}
	counter, ok := ud.Value.(*PyCounter)
	if !ok {
		vm.RaiseError("expected Counter object")
		return 0
	}

	// Return list of elements repeated by their count
	var result []runtime.Value
	for elem, count := range counter.Items {
		for i := int64(0); i < count; i++ {
			result = append(result, elem)
		}
	}

	vm.Push(runtime.NewList(result))
	return 1
}

// Counter.update([iterable-or-mapping])
func counterUpdate(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected Counter object")
		return 0
	}
	counter, ok := ud.Value.(*PyCounter)
	if !ok {
		vm.RaiseError("expected Counter object")
		return 0
	}

	if vm.GetTop() >= 2 && !runtime.IsNone(vm.Get(2)) {
		arg := vm.Get(2)
		switch v := arg.(type) {
		case *runtime.PyList:
			for _, item := range v.Items {
				counter.Items[item]++
			}
		case *runtime.PyTuple:
			for _, item := range v.Items {
				counter.Items[item]++
			}
		case *runtime.PyString:
			for _, ch := range v.Value {
				counter.Items[runtime.NewString(string(ch))]++
			}
		case *runtime.PyDict:
			for k, val := range v.Items {
				if intVal, ok := val.(*runtime.PyInt); ok {
					counter.Items[k] += intVal.Value
				}
			}
		case *runtime.PyUserData:
			if otherCounter, ok := v.Value.(*PyCounter); ok {
				for k, val := range otherCounter.Items {
					counter.Items[k] += val
				}
			}
		}
	}

	return 0
}

// Counter.subtract([iterable-or-mapping])
func counterSubtract(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected Counter object")
		return 0
	}
	counter, ok := ud.Value.(*PyCounter)
	if !ok {
		vm.RaiseError("expected Counter object")
		return 0
	}

	if vm.GetTop() >= 2 && !runtime.IsNone(vm.Get(2)) {
		arg := vm.Get(2)
		switch v := arg.(type) {
		case *runtime.PyList:
			for _, item := range v.Items {
				counter.Items[item]--
			}
		case *runtime.PyTuple:
			for _, item := range v.Items {
				counter.Items[item]--
			}
		case *runtime.PyString:
			for _, ch := range v.Value {
				counter.Items[runtime.NewString(string(ch))]--
			}
		case *runtime.PyDict:
			for k, val := range v.Items {
				if intVal, ok := val.(*runtime.PyInt); ok {
					counter.Items[k] -= intVal.Value
				}
			}
		case *runtime.PyUserData:
			if otherCounter, ok := v.Value.(*PyCounter); ok {
				for k, val := range otherCounter.Items {
					counter.Items[k] -= val
				}
			}
		}
	}

	return 0
}

// Counter.total() -> sum of counts
func counterTotal(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected Counter object")
		return 0
	}
	counter, ok := ud.Value.(*PyCounter)
	if !ok {
		vm.RaiseError("expected Counter object")
		return 0
	}

	var total int64
	for _, count := range counter.Items {
		total += count
	}

	vm.Push(runtime.NewInt(total))
	return 1
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
		deque.Items = deque.Items[len(deque.Items)-int(deque.Maxlen):]
	}

	ud := runtime.NewUserData(deque)
	ud.Metatable = runtime.NewDict()
	ud.Metatable.Items[runtime.NewString("__type__")] = runtime.NewString("deque")
	vm.Push(ud)
	return 1
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
		deque.Items = deque.Items[len(deque.Items)-int(deque.Maxlen):]
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

	n := int64(1)
	if vm.GetTop() >= 2 && !runtime.IsNone(vm.Get(2)) {
		n = vm.CheckInt(2)
	}

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
		if valuesEqual(item, x) {
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
	start := int64(0)
	stop := int64(len(deque.Items))

	if vm.GetTop() >= 3 && !runtime.IsNone(vm.Get(3)) {
		start = vm.CheckInt(3)
	}
	if vm.GetTop() >= 4 && !runtime.IsNone(vm.Get(4)) {
		stop = vm.CheckInt(4)
	}

	for i := start; i < stop && i < int64(len(deque.Items)); i++ {
		if valuesEqual(deque.Items[i], x) {
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
		vm.RaiseError("deque already at its maximum size")
		return 0
	}

	// Handle negative index
	if i < 0 {
		i = len(deque.Items) + i + 1
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
		if valuesEqual(item, x) {
			deque.Items = append(deque.Items[:i], deque.Items[i+1:]...)
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

	vm.Push(runtime.NewUserData(newDeque))
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

// =====================================
// defaultdict
// =====================================

// PyDefaultDict is a dict with a default factory
type PyDefaultDict struct {
	Items          map[runtime.Value]runtime.Value
	DefaultFactory runtime.Value
}

func (d *PyDefaultDict) Type() string   { return "defaultdict" }
func (d *PyDefaultDict) String() string { return fmt.Sprintf("defaultdict(%v)", d.Items) }

// collections.defaultdict([default_factory[, ...]])
func collectionsDefaultdict(vm *runtime.VM) int {
	dd := &PyDefaultDict{
		Items:          make(map[runtime.Value]runtime.Value),
		DefaultFactory: runtime.None,
	}

	if vm.GetTop() >= 1 && !runtime.IsNone(vm.Get(1)) {
		dd.DefaultFactory = vm.Get(1)
	}

	// Initialize from additional args if provided
	if vm.GetTop() >= 2 {
		arg := vm.Get(2)
		if dict, ok := arg.(*runtime.PyDict); ok {
			for k, v := range dict.Items {
				dd.Items[k] = v
			}
		}
	}

	ud := runtime.NewUserData(dd)
	ud.Metatable = runtime.NewDict()
	ud.Metatable.Items[runtime.NewString("__type__")] = runtime.NewString("defaultdict")
	vm.Push(ud)
	return 1
}

// defaultdict.default_factory getter
func defaultdictGetFactory(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected defaultdict object")
		return 0
	}
	dd, ok := ud.Value.(*PyDefaultDict)
	if !ok {
		vm.RaiseError("expected defaultdict object")
		return 0
	}

	vm.Push(dd.DefaultFactory)
	return 1
}

// =====================================
// OrderedDict (just a regular dict in Python 3.7+)
// =====================================

// collections.OrderedDict([items])
func collectionsOrderedDict(vm *runtime.VM) int {
	dict := runtime.NewDict()

	if vm.GetTop() >= 1 && !runtime.IsNone(vm.Get(1)) {
		arg := vm.Get(1)
		if d, ok := arg.(*runtime.PyDict); ok {
			for k, v := range d.Items {
				dict.Items[k] = v
			}
		} else if l, ok := arg.(*runtime.PyList); ok {
			for _, item := range l.Items {
				if tuple, ok := item.(*runtime.PyTuple); ok && len(tuple.Items) >= 2 {
					dict.Items[tuple.Items[0]] = tuple.Items[1]
				}
			}
		}
	}

	vm.Push(dict)
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
			if s, ok := item.(*runtime.PyString); ok {
				fieldNames = append(fieldNames, s.Value)
			}
		}
	case *runtime.PyTuple:
		for _, item := range v.Items {
			if s, ok := item.(*runtime.PyString); ok {
				fieldNames = append(fieldNames, s.Value)
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

		// Create a dict-like tuple with _fields
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
// Helper functions
// =====================================

// valuesEqual compares two Values for equality
func valuesEqual(a, b runtime.Value) bool {
	switch va := a.(type) {
	case *runtime.PyInt:
		if vb, ok := b.(*runtime.PyInt); ok {
			return va.Value == vb.Value
		}
	case *runtime.PyFloat:
		if vb, ok := b.(*runtime.PyFloat); ok {
			return va.Value == vb.Value
		}
	case *runtime.PyString:
		if vb, ok := b.(*runtime.PyString); ok {
			return va.Value == vb.Value
		}
	case *runtime.PyBool:
		if vb, ok := b.(*runtime.PyBool); ok {
			return va.Value == vb.Value
		}
	case *runtime.PyNone:
		_, ok := b.(*runtime.PyNone)
		return ok
	}
	return a == b
}
