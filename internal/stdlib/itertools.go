package stdlib

import (
	"github.com/ATSOTECK/rage/internal/runtime"
)

// InitItertoolsModule registers the itertools module
func InitItertoolsModule() {
	// Register iterator type metatables
	countIterMT := &runtime.TypeMetatable{
		Name: "itertools.count",
		Methods: map[string]runtime.GoFunction{
			"__iter__": countIterIter,
			"__next__": countIterNext,
		},
	}
	runtime.RegisterTypeMetatable("itertools.count", countIterMT)

	cycleIterMT := &runtime.TypeMetatable{
		Name: "itertools.cycle",
		Methods: map[string]runtime.GoFunction{
			"__iter__": cycleIterIter,
			"__next__": cycleIterNext,
		},
	}
	runtime.RegisterTypeMetatable("itertools.cycle", cycleIterMT)

	repeatIterMT := &runtime.TypeMetatable{
		Name: "itertools.repeat",
		Methods: map[string]runtime.GoFunction{
			"__iter__": repeatIterIter,
			"__next__": repeatIterNext,
		},
	}
	runtime.RegisterTypeMetatable("itertools.repeat", repeatIterMT)

	runtime.NewModuleBuilder("itertools").
		Doc("Functional tools for creating and using iterators.").
		// Infinite iterators
		Func("count", itertoolsCount).
		Func("cycle", itertoolsCycle).
		Func("repeat", itertoolsRepeat).
		// Iterators terminating on shortest input
		Func("accumulate", itertoolsAccumulate).
		Func("chain", itertoolsChain).
		Func("compress", itertoolsCompress).
		Func("dropwhile", itertoolsDropwhile).
		Func("filterfalse", itertoolsFilterfalse).
		Func("groupby", itertoolsGroupby).
		Func("islice", itertoolsIslice).
		Func("pairwise", itertoolsPairwise).
		Func("starmap", itertoolsStarmap).
		Func("takewhile", itertoolsTakewhile).
		Func("zip_longest", itertoolsZipLongest).
		// Combinatoric iterators
		Func("product", itertoolsProduct).
		Func("permutations", itertoolsPermutations).
		Func("combinations", itertoolsCombinations).
		Func("combinations_with_replacement", itertoolsCombinationsWithReplacement).
		Register()
}

// =====================================
// Infinite Iterators
// =====================================

// PyCountIter represents a count iterator
type PyCountIter struct {
	Start   int64
	Step    int64
	Current int64
}

func (c *PyCountIter) Type() string   { return "itertools.count" }
func (c *PyCountIter) String() string { return "count(...)" }

// itertools.count(start=0, step=1)
func itertoolsCount(vm *runtime.VM) int {
	start := vm.OptionalInt(1, 0)
	step := vm.OptionalInt(2, 1)

	iter := &PyCountIter{
		Start:   start,
		Step:    step,
		Current: start,
	}

	ud := runtime.NewUserData(iter)
	ud.Metatable = runtime.NewDict()
	ud.Metatable.Items[runtime.NewString("__type__")] = runtime.NewString("itertools.count")
	vm.Push(ud)
	return 1
}

func countIterIter(vm *runtime.VM) int {
	// Return self
	vm.Push(vm.Get(1))
	return 1
}

func countIterNext(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected count iterator")
		return 0
	}
	iter, ok := ud.Value.(*PyCountIter)
	if !ok {
		vm.RaiseError("expected count iterator")
		return 0
	}

	val := iter.Current
	iter.Current += iter.Step
	vm.Push(runtime.NewInt(val))
	return 1
}

// PyCycleIter represents a cycle iterator
type PyCycleIter struct {
	Items []runtime.Value
	Index int
}

func (c *PyCycleIter) Type() string   { return "itertools.cycle" }
func (c *PyCycleIter) String() string { return "cycle(...)" }

// itertools.cycle(iterable)
func itertoolsCycle(vm *runtime.VM) int {
	items := getIterableItems(vm, 1)
	if items == nil {
		return 0
	}

	if len(items) == 0 {
		vm.RaiseError("cycle() argument must be a non-empty iterable")
		return 0
	}

	iter := &PyCycleIter{
		Items: items,
		Index: 0,
	}

	ud := runtime.NewUserData(iter)
	ud.Metatable = runtime.NewDict()
	ud.Metatable.Items[runtime.NewString("__type__")] = runtime.NewString("itertools.cycle")
	vm.Push(ud)
	return 1
}

func cycleIterIter(vm *runtime.VM) int {
	vm.Push(vm.Get(1))
	return 1
}

func cycleIterNext(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected cycle iterator")
		return 0
	}
	iter, ok := ud.Value.(*PyCycleIter)
	if !ok {
		vm.RaiseError("expected cycle iterator")
		return 0
	}

	val := iter.Items[iter.Index]
	iter.Index = (iter.Index + 1) % len(iter.Items)
	vm.Push(val)
	return 1
}

// PyRepeatIter represents a repeat iterator
type PyRepeatIter struct {
	Object runtime.Value
	Times  int64 // -1 means infinite
	Count  int64
}

func (r *PyRepeatIter) Type() string   { return "itertools.repeat" }
func (r *PyRepeatIter) String() string { return "repeat(...)" }

// itertools.repeat(object[, times])
func itertoolsRepeat(vm *runtime.VM) int {
	if !vm.RequireArgs("repeat", 1) {
		return 0
	}

	obj := vm.Get(1)
	times := vm.OptionalInt(2, -1) // -1 means infinite

	// If times is specified, return a list directly for easy iteration
	if times >= 0 {
		result := make([]runtime.Value, times)
		for i := int64(0); i < times; i++ {
			result[i] = obj
		}
		vm.Push(runtime.NewList(result))
		return 1
	}

	// For infinite repeat (no times specified), return an iterator
	iter := &PyRepeatIter{
		Object: obj,
		Times:  times,
		Count:  0,
	}

	ud := runtime.NewUserData(iter)
	ud.Metatable = runtime.NewDict()
	ud.Metatable.Items[runtime.NewString("__type__")] = runtime.NewString("itertools.repeat")
	vm.Push(ud)
	return 1
}

func repeatIterIter(vm *runtime.VM) int {
	vm.Push(vm.Get(1))
	return 1
}

func repeatIterNext(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected repeat iterator")
		return 0
	}
	iter, ok := ud.Value.(*PyRepeatIter)
	if !ok {
		vm.RaiseError("expected repeat iterator")
		return 0
	}

	if iter.Times >= 0 && iter.Count >= iter.Times {
		vm.RaiseError("StopIteration")
		return 0
	}

	iter.Count++
	vm.Push(iter.Object)
	return 1
}

// =====================================
// Iterators terminating on shortest input
// =====================================

// itertools.accumulate(iterable[, func, initial])
func itertoolsAccumulate(vm *runtime.VM) int {
	items := getIterableItems(vm, 1)
	if items == nil {
		return 0
	}

	// Get optional function (default is addition)
	var fn runtime.Value
	if vm.GetTop() >= 2 && !runtime.IsNone(vm.Get(2)) {
		fn = vm.Get(2)
	}

	// Get optional initial value
	var initial runtime.Value
	hasInitial := false
	if vm.GetTop() >= 3 && !runtime.IsNone(vm.Get(3)) {
		initial = vm.Get(3)
		hasInitial = true
	}

	if len(items) == 0 && !hasInitial {
		vm.Push(runtime.NewList([]runtime.Value{}))
		return 1
	}

	result := make([]runtime.Value, 0, len(items)+1)

	var acc runtime.Value
	startIdx := 0
	if hasInitial {
		acc = initial
		result = append(result, acc)
	} else {
		acc = items[0]
		result = append(result, acc)
		startIdx = 1
	}

	for i := startIdx; i < len(items); i++ {
		if fn != nil {
			// Call the function with acc and items[i]
			callResult, err := vm.Call(fn, []runtime.Value{acc, items[i]}, nil)
			if err != nil {
				vm.RaiseError("%v", err)
				return 0
			}
			acc = callResult
		} else {
			// Default: addition
			acc = addValues(acc, items[i])
		}
		result = append(result, acc)
	}

	vm.Push(runtime.NewList(result))
	return 1
}

// itertools.chain(*iterables)
func itertoolsChain(vm *runtime.VM) int {
	var result []runtime.Value

	for i := 1; i <= vm.GetTop(); i++ {
		items := getIterableItems(vm, i)
		if items == nil {
			return 0
		}
		result = append(result, items...)
	}

	vm.Push(runtime.NewList(result))
	return 1
}

// itertools.compress(data, selectors)
func itertoolsCompress(vm *runtime.VM) int {
	data := getIterableItems(vm, 1)
	if data == nil {
		return 0
	}

	selectors := getIterableItems(vm, 2)
	if selectors == nil {
		return 0
	}

	var result []runtime.Value
	minLen := len(data)
	if len(selectors) < minLen {
		minLen = len(selectors)
	}

	for i := 0; i < minLen; i++ {
		if runtime.IsTrue(selectors[i]) {
			result = append(result, data[i])
		}
	}

	vm.Push(runtime.NewList(result))
	return 1
}

// itertools.dropwhile(predicate, iterable)
func itertoolsDropwhile(vm *runtime.VM) int {
	if !vm.RequireArgs("dropwhile", 2) {
		return 0
	}

	predicate := vm.Get(1)
	items := getIterableItems(vm, 2)
	if items == nil {
		return 0
	}

	var result []runtime.Value
	dropping := true

	for _, item := range items {
		if dropping {
			callResult, err := vm.Call(predicate, []runtime.Value{item}, nil)
			if err != nil {
				vm.RaiseError("%v", err)
				return 0
			}
			if !runtime.IsTrue(callResult) {
				dropping = false
				result = append(result, item)
			}
		} else {
			result = append(result, item)
		}
	}

	vm.Push(runtime.NewList(result))
	return 1
}

// itertools.filterfalse(predicate, iterable)
func itertoolsFilterfalse(vm *runtime.VM) int {
	if !vm.RequireArgs("filterfalse", 2) {
		return 0
	}

	predicate := vm.Get(1)
	items := getIterableItems(vm, 2)
	if items == nil {
		return 0
	}

	var result []runtime.Value

	for _, item := range items {
		if runtime.IsNone(predicate) {
			// If predicate is None, filter items that are falsy
			if !runtime.IsTrue(item) {
				result = append(result, item)
			}
		} else {
			callResult, err := vm.Call(predicate, []runtime.Value{item}, nil)
			if err != nil {
				vm.RaiseError("%v", err)
				return 0
			}
			if !runtime.IsTrue(callResult) {
				result = append(result, item)
			}
		}
	}

	vm.Push(runtime.NewList(result))
	return 1
}

// itertools.groupby(iterable[, key])
func itertoolsGroupby(vm *runtime.VM) int {
	items := getIterableItems(vm, 1)
	if items == nil {
		return 0
	}

	var keyFunc runtime.Value
	if vm.GetTop() >= 2 && !runtime.IsNone(vm.Get(2)) {
		keyFunc = vm.Get(2)
	}

	if len(items) == 0 {
		vm.Push(runtime.NewList([]runtime.Value{}))
		return 1
	}

	var result []runtime.Value
	var currentKey runtime.Value
	var currentGroup []runtime.Value

	for i, item := range items {
		var key runtime.Value
		if keyFunc != nil {
			callResult, err := vm.Call(keyFunc, []runtime.Value{item}, nil)
			if err != nil {
				vm.RaiseError("%v", err)
				return 0
			}
			key = callResult
		} else {
			key = item
		}

		if i == 0 {
			currentKey = key
			currentGroup = []runtime.Value{item}
		} else if vm.Equal(key, currentKey) {
			currentGroup = append(currentGroup, item)
		} else {
			// Save current group and start new one
			result = append(result, runtime.NewTuple([]runtime.Value{
				currentKey,
				runtime.NewList(currentGroup),
			}))
			currentKey = key
			currentGroup = []runtime.Value{item}
		}
	}

	// Don't forget the last group
	if len(currentGroup) > 0 {
		result = append(result, runtime.NewTuple([]runtime.Value{
			currentKey,
			runtime.NewList(currentGroup),
		}))
	}

	vm.Push(runtime.NewList(result))
	return 1
}

// itertools.islice(iterable, stop) or islice(iterable, start, stop[, step])
func itertoolsIslice(vm *runtime.VM) int {
	if !vm.RequireArgs("islice", 2) {
		return 0
	}

	arg1 := vm.Get(1)
	var items []runtime.Value

	// Check if first argument is one of our special iterator types
	if ud, ok := arg1.(*runtime.PyUserData); ok {
		switch iter := ud.Value.(type) {
		case *PyCountIter:
			// Generate items from count iterator
			stop := int64(0)
			start := int64(0)
			step := int64(1)

			if vm.GetTop() == 2 {
				if !runtime.IsNone(vm.Get(2)) {
					stop = vm.ToInt(2)
				}
			} else {
				if !runtime.IsNone(vm.Get(2)) {
					start = vm.ToInt(2)
				}
				if !runtime.IsNone(vm.Get(3)) {
					stop = vm.ToInt(3)
				}
				if vm.GetTop() >= 4 && !runtime.IsNone(vm.Get(4)) {
					step = vm.ToInt(4)
				}
			}

			if step <= 0 {
				vm.RaiseError("step for islice() must be positive")
				return 0
			}

			var result []runtime.Value
			current := iter.Start
			idx := int64(0)
			for idx < stop {
				if idx >= start && (idx-start)%step == 0 {
					result = append(result, runtime.NewInt(current))
				}
				current += iter.Step
				idx++
			}
			vm.Push(runtime.NewList(result))
			return 1

		case *PyCycleIter:
			// Generate items from cycle iterator
			stop := int64(0)
			start := int64(0)
			step := int64(1)

			if vm.GetTop() == 2 {
				if !runtime.IsNone(vm.Get(2)) {
					stop = vm.ToInt(2)
				}
			} else {
				if !runtime.IsNone(vm.Get(2)) {
					start = vm.ToInt(2)
				}
				if !runtime.IsNone(vm.Get(3)) {
					stop = vm.ToInt(3)
				}
				if vm.GetTop() >= 4 && !runtime.IsNone(vm.Get(4)) {
					step = vm.ToInt(4)
				}
			}

			if step <= 0 {
				vm.RaiseError("step for islice() must be positive")
				return 0
			}

			var result []runtime.Value
			cycleIdx := iter.Index
			for idx := int64(0); idx < stop; idx++ {
				if idx >= start && (idx-start)%step == 0 {
					result = append(result, iter.Items[cycleIdx])
				}
				cycleIdx = (cycleIdx + 1) % len(iter.Items)
			}
			vm.Push(runtime.NewList(result))
			return 1

		case *PyRepeatIter:
			// Generate items from repeat iterator
			stop := int64(0)
			start := int64(0)
			step := int64(1)

			if vm.GetTop() == 2 {
				if !runtime.IsNone(vm.Get(2)) {
					stop = vm.ToInt(2)
				}
			} else {
				if !runtime.IsNone(vm.Get(2)) {
					start = vm.ToInt(2)
				}
				if !runtime.IsNone(vm.Get(3)) {
					stop = vm.ToInt(3)
				}
				if vm.GetTop() >= 4 && !runtime.IsNone(vm.Get(4)) {
					step = vm.ToInt(4)
				}
			}

			if step <= 0 {
				vm.RaiseError("step for islice() must be positive")
				return 0
			}

			// Respect repeat's times limit if set
			effectiveStop := stop
			if iter.Times >= 0 && iter.Times < stop {
				effectiveStop = iter.Times
			}

			var result []runtime.Value
			for idx := int64(0); idx < effectiveStop; idx++ {
				if idx >= start && (idx-start)%step == 0 {
					result = append(result, iter.Object)
				}
			}
			vm.Push(runtime.NewList(result))
			return 1
		}
	}

	// Regular iterable
	items = getIterableItems(vm, 1)
	if items == nil {
		return 0
	}

	var start, stop, step int64

	if vm.GetTop() == 2 {
		// islice(iterable, stop)
		start = 0
		if runtime.IsNone(vm.Get(2)) {
			stop = int64(len(items))
		} else {
			stop = vm.ToInt(2)
		}
		step = 1
	} else {
		// islice(iterable, start, stop[, step])
		if runtime.IsNone(vm.Get(2)) {
			start = 0
		} else {
			start = vm.ToInt(2)
		}
		if runtime.IsNone(vm.Get(3)) {
			stop = int64(len(items))
		} else {
			stop = vm.ToInt(3)
		}
		step = 1
		if vm.GetTop() >= 4 && !runtime.IsNone(vm.Get(4)) {
			step = vm.ToInt(4)
		}
	}

	if step <= 0 {
		vm.RaiseError("step for islice() must be positive")
		return 0
	}

	if stop > int64(len(items)) {
		stop = int64(len(items))
	}

	var result []runtime.Value
	for i := start; i < stop; i += step {
		result = append(result, items[i])
	}

	vm.Push(runtime.NewList(result))
	return 1
}

// itertools.pairwise(iterable)
func itertoolsPairwise(vm *runtime.VM) int {
	items := getIterableItems(vm, 1)
	if items == nil {
		return 0
	}

	if len(items) < 2 {
		vm.Push(runtime.NewList([]runtime.Value{}))
		return 1
	}

	result := make([]runtime.Value, len(items)-1)
	for i := 0; i < len(items)-1; i++ {
		result[i] = runtime.NewTuple([]runtime.Value{items[i], items[i+1]})
	}

	vm.Push(runtime.NewList(result))
	return 1
}

// itertools.starmap(function, iterable)
func itertoolsStarmap(vm *runtime.VM) int {
	if !vm.RequireArgs("starmap", 2) {
		return 0
	}

	fn := vm.Get(1)
	items := getIterableItems(vm, 2)
	if items == nil {
		return 0
	}

	result := make([]runtime.Value, len(items))
	for i, item := range items {
		// Each item should be an iterable that we unpack as arguments
		args := getIterableItemsFromValue(vm, item)
		if args == nil {
			vm.RaiseError("starmap() argument 2 must be an iterable of iterables")
			return 0
		}

		callResult, err := vm.Call(fn, args, nil)
		if err != nil {
			vm.RaiseError("%v", err)
			return 0
		}
		result[i] = callResult
	}

	vm.Push(runtime.NewList(result))
	return 1
}

// itertools.takewhile(predicate, iterable)
func itertoolsTakewhile(vm *runtime.VM) int {
	if !vm.RequireArgs("takewhile", 2) {
		return 0
	}

	predicate := vm.Get(1)
	items := getIterableItems(vm, 2)
	if items == nil {
		return 0
	}

	var result []runtime.Value

	for _, item := range items {
		callResult, err := vm.Call(predicate, []runtime.Value{item}, nil)
		if err != nil {
			vm.RaiseError("%v", err)
			return 0
		}
		if !runtime.IsTrue(callResult) {
			break
		}
		result = append(result, item)
	}

	vm.Push(runtime.NewList(result))
	return 1
}

// itertools.zip_longest(*iterables, fillvalue=None)
func itertoolsZipLongest(vm *runtime.VM) int {
	if vm.GetTop() < 1 {
		vm.Push(runtime.NewList([]runtime.Value{}))
		return 1
	}

	// Collect all iterables
	var iterables [][]runtime.Value
	var fillvalue runtime.Value = runtime.None

	for i := 1; i <= vm.GetTop(); i++ {
		arg := vm.Get(i)
		// Check if this is a keyword argument dict for fillvalue
		if dict, ok := arg.(*runtime.PyDict); ok {
			for k, v := range dict.Items {
				if keyStr, ok := k.(*runtime.PyString); ok && keyStr.Value == "fillvalue" {
					fillvalue = v
				}
			}
			continue
		}
		items := getIterableItemsFromValue(vm, arg)
		if items == nil {
			vm.RaiseError("zip_longest argument must be an iterable")
			return 0
		}
		iterables = append(iterables, items)
	}

	if len(iterables) == 0 {
		vm.Push(runtime.NewList([]runtime.Value{}))
		return 1
	}

	// Find the maximum length
	maxLen := 0
	for _, it := range iterables {
		if len(it) > maxLen {
			maxLen = len(it)
		}
	}

	result := make([]runtime.Value, maxLen)
	for i := 0; i < maxLen; i++ {
		tuple := make([]runtime.Value, len(iterables))
		for j, it := range iterables {
			if i < len(it) {
				tuple[j] = it[i]
			} else {
				tuple[j] = fillvalue
			}
		}
		result[i] = runtime.NewTuple(tuple)
	}

	vm.Push(runtime.NewList(result))
	return 1
}

// =====================================
// Combinatoric Iterators
// =====================================

// itertools.product(*iterables, repeat=1)
func itertoolsProduct(vm *runtime.VM) int {
	if vm.GetTop() < 1 {
		vm.Push(runtime.NewList([]runtime.Value{runtime.NewTuple([]runtime.Value{})}))
		return 1
	}

	// Collect all iterables and check for repeat keyword
	var pools [][]runtime.Value
	repeat := int64(1)

	for i := 1; i <= vm.GetTop(); i++ {
		arg := vm.Get(i)
		// Check if this is a keyword argument dict for repeat
		if dict, ok := arg.(*runtime.PyDict); ok {
			for k, v := range dict.Items {
				if keyStr, ok := k.(*runtime.PyString); ok && keyStr.Value == "repeat" {
					if intVal, ok := v.(*runtime.PyInt); ok {
						repeat = intVal.Value
					}
				}
			}
			continue
		}
		items := getIterableItemsFromValue(vm, arg)
		if items == nil {
			vm.RaiseError("product argument must be an iterable")
			return 0
		}
		pools = append(pools, items)
	}

	// Apply repeat
	if repeat > 1 {
		originalPools := pools
		pools = make([][]runtime.Value, 0, len(originalPools)*int(repeat))
		for r := int64(0); r < repeat; r++ {
			pools = append(pools, originalPools...)
		}
	}

	if len(pools) == 0 {
		vm.Push(runtime.NewList([]runtime.Value{runtime.NewTuple([]runtime.Value{})}))
		return 1
	}

	// Calculate total product size
	totalSize := 1
	for _, pool := range pools {
		if len(pool) == 0 {
			vm.Push(runtime.NewList([]runtime.Value{}))
			return 1
		}
		totalSize *= len(pool)
	}

	result := make([]runtime.Value, 0, totalSize)
	indices := make([]int, len(pools))

	for {
		// Create tuple from current indices
		tuple := make([]runtime.Value, len(pools))
		for i, idx := range indices {
			tuple[i] = pools[i][idx]
		}
		result = append(result, runtime.NewTuple(tuple))

		// Increment indices (like a multi-base counter)
		i := len(indices) - 1
		for i >= 0 {
			indices[i]++
			if indices[i] < len(pools[i]) {
				break
			}
			indices[i] = 0
			i--
		}
		if i < 0 {
			break
		}
	}

	vm.Push(runtime.NewList(result))
	return 1
}

// itertools.permutations(iterable[, r])
func itertoolsPermutations(vm *runtime.VM) int {
	items := getIterableItems(vm, 1)
	if items == nil {
		return 0
	}

	n := len(items)
	r := n
	if vm.GetTop() >= 2 && !runtime.IsNone(vm.Get(2)) {
		r = int(vm.ToInt(2))
	}

	if r > n || r < 0 {
		vm.Push(runtime.NewList([]runtime.Value{}))
		return 1
	}

	if r == 0 {
		vm.Push(runtime.NewList([]runtime.Value{runtime.NewTuple([]runtime.Value{})}))
		return 1
	}

	// Generate permutations
	var result []runtime.Value
	indices := make([]int, n)
	for i := range indices {
		indices[i] = i
	}

	cycles := make([]int, r)
	for i := range cycles {
		cycles[i] = n - i
	}

	// First permutation
	perm := make([]runtime.Value, r)
	for i := 0; i < r; i++ {
		perm[i] = items[indices[i]]
	}
	result = append(result, runtime.NewTuple(perm))

	for {
		found := false
		for i := r - 1; i >= 0; i-- {
			cycles[i]--
			if cycles[i] == 0 {
				// Rotate indices[i:] left by 1
				temp := indices[i]
				copy(indices[i:], indices[i+1:])
				indices[n-1] = temp
				cycles[i] = n - i
			} else {
				j := cycles[i]
				indices[i], indices[n-j] = indices[n-j], indices[i]

				perm := make([]runtime.Value, r)
				for k := 0; k < r; k++ {
					perm[k] = items[indices[k]]
				}
				result = append(result, runtime.NewTuple(perm))
				found = true
				break
			}
		}
		if !found {
			break
		}
	}

	vm.Push(runtime.NewList(result))
	return 1
}

// itertools.combinations(iterable, r)
func itertoolsCombinations(vm *runtime.VM) int {
	if !vm.RequireArgs("combinations", 2) {
		return 0
	}

	items := getIterableItems(vm, 1)
	if items == nil {
		return 0
	}

	n := len(items)
	r := int(vm.ToInt(2))

	if r > n || r < 0 {
		vm.Push(runtime.NewList([]runtime.Value{}))
		return 1
	}

	if r == 0 {
		vm.Push(runtime.NewList([]runtime.Value{runtime.NewTuple([]runtime.Value{})}))
		return 1
	}

	var result []runtime.Value
	indices := make([]int, r)
	for i := range indices {
		indices[i] = i
	}

	// First combination
	comb := make([]runtime.Value, r)
	for i := 0; i < r; i++ {
		comb[i] = items[indices[i]]
	}
	result = append(result, runtime.NewTuple(comb))

	for {
		// Find rightmost index that can be incremented
		i := r - 1
		for i >= 0 && indices[i] == i+n-r {
			i--
		}
		if i < 0 {
			break
		}

		indices[i]++
		for j := i + 1; j < r; j++ {
			indices[j] = indices[j-1] + 1
		}

		comb := make([]runtime.Value, r)
		for k := 0; k < r; k++ {
			comb[k] = items[indices[k]]
		}
		result = append(result, runtime.NewTuple(comb))
	}

	vm.Push(runtime.NewList(result))
	return 1
}

// itertools.combinations_with_replacement(iterable, r)
func itertoolsCombinationsWithReplacement(vm *runtime.VM) int {
	if !vm.RequireArgs("combinations_with_replacement", 2) {
		return 0
	}

	items := getIterableItems(vm, 1)
	if items == nil {
		return 0
	}

	n := len(items)
	r := int(vm.ToInt(2))

	if n == 0 && r > 0 {
		vm.Push(runtime.NewList([]runtime.Value{}))
		return 1
	}

	if r == 0 {
		vm.Push(runtime.NewList([]runtime.Value{runtime.NewTuple([]runtime.Value{})}))
		return 1
	}

	if r < 0 {
		vm.Push(runtime.NewList([]runtime.Value{}))
		return 1
	}

	var result []runtime.Value
	indices := make([]int, r)
	// All indices start at 0

	// First combination
	comb := make([]runtime.Value, r)
	for i := 0; i < r; i++ {
		comb[i] = items[indices[i]]
	}
	result = append(result, runtime.NewTuple(comb))

	for {
		// Find rightmost index that can be incremented
		i := r - 1
		for i >= 0 && indices[i] == n-1 {
			i--
		}
		if i < 0 {
			break
		}

		// Increment and reset all indices to the right
		newVal := indices[i] + 1
		for j := i; j < r; j++ {
			indices[j] = newVal
		}

		comb := make([]runtime.Value, r)
		for k := 0; k < r; k++ {
			comb[k] = items[indices[k]]
		}
		result = append(result, runtime.NewTuple(comb))
	}

	vm.Push(runtime.NewList(result))
	return 1
}

// =====================================
// Helper functions
// =====================================

// getIterableItems extracts items from an iterable at the given stack position
// using the runtime's ToList method.
func getIterableItems(vm *runtime.VM, pos int) []runtime.Value {
	if pos > vm.GetTop() {
		vm.RaiseError("missing required argument")
		return nil
	}
	items, err := vm.ToList(vm.Get(pos))
	if err != nil {
		vm.RaiseError("%v", err)
		return nil
	}
	return items
}

// getIterableItemsFromValue extracts items from an iterable value
// using the runtime's ToList method.
func getIterableItemsFromValue(vm *runtime.VM, v runtime.Value) []runtime.Value {
	items, err := vm.ToList(v)
	if err != nil {
		return nil
	}
	return items
}

// addValues adds two numeric values
func addValues(a, b runtime.Value) runtime.Value {
	switch va := a.(type) {
	case *runtime.PyInt:
		switch vb := b.(type) {
		case *runtime.PyInt:
			return runtime.NewInt(va.Value + vb.Value)
		case *runtime.PyFloat:
			return runtime.NewFloat(float64(va.Value) + vb.Value)
		}
	case *runtime.PyFloat:
		switch vb := b.(type) {
		case *runtime.PyInt:
			return runtime.NewFloat(va.Value + float64(vb.Value))
		case *runtime.PyFloat:
			return runtime.NewFloat(va.Value + vb.Value)
		}
	case *runtime.PyString:
		if vb, ok := b.(*runtime.PyString); ok {
			return runtime.NewString(va.Value + vb.Value)
		}
	case *runtime.PyList:
		if vb, ok := b.(*runtime.PyList); ok {
			result := make([]runtime.Value, len(va.Items)+len(vb.Items))
			copy(result, va.Items)
			copy(result[len(va.Items):], vb.Items)
			return runtime.NewList(result)
		}
	}
	// Default: return b (for unsupported types)
	return b
}
