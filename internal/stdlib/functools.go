package stdlib

import (
	"container/list"
	"fmt"
	"sync"

	"github.com/ATSOTECK/rage/internal/runtime"
)

// InitFunctoolsModule registers the functools module
func InitFunctoolsModule() {
	// Register partial type metatable
	partialMT := &runtime.TypeMetatable{
		Name: "functools.partial",
		Methods: map[string]runtime.GoFunction{
			"__call__": partialCall,
		},
	}
	runtime.RegisterTypeMetatable("functools.partial", partialMT)

	// Register _lru_cache_wrapper type metatable
	lruCacheMT := &runtime.TypeMetatable{
		Name: "functools._lru_cache_wrapper",
		Methods: map[string]runtime.GoFunction{
			"__call__":    lruCacheCall,
			"cache_info":  lruCacheCacheInfo,
			"cache_clear": lruCacheCacheClear,
		},
	}
	runtime.RegisterTypeMetatable("functools._lru_cache_wrapper", lruCacheMT)

	// Register cmp_to_key wrapper type metatable
	cmpKeyMT := &runtime.TypeMetatable{
		Name: "functools.KeyWrapper",
		Methods: map[string]runtime.GoFunction{
			"__lt__": cmpKeyLt,
			"__gt__": cmpKeyGt,
			"__eq__": cmpKeyEq,
			"__le__": cmpKeyLe,
			"__ge__": cmpKeyGe,
		},
	}
	runtime.RegisterTypeMetatable("functools.KeyWrapper", cmpKeyMT)

	runtime.NewModuleBuilder("functools").
		Doc("Higher-order functions and operations on callable objects.").
		Func("partial", functoolsPartial).
		Func("reduce", functoolsReduce).
		Func("wraps", functoolsWraps).
		Func("update_wrapper", functoolsUpdateWrapper).
		Func("cache", functoolsCache).
		Func("lru_cache", functoolsLruCache).
		Func("cmp_to_key", functoolsCmpToKey).
		Register()
}

// =====================================
// functools.partial
// =====================================

// PyPartial represents a partial function application
type PyPartial struct {
	Func   runtime.Value
	Args   []runtime.Value
	Kwargs map[string]runtime.Value
}

func (p *PyPartial) Type() string   { return "functools.partial" }
func (p *PyPartial) String() string { return fmt.Sprintf("functools.partial(%v)", p.Func) }

// functools.partial(func, *args, **kwargs)
func functoolsPartial(vm *runtime.VM) int {
	if !vm.RequireArgs("partial", 1) {
		return 0
	}

	fn := vm.Get(1)
	if !runtime.IsCallable(fn) {
		vm.RaiseError("the first argument must be callable")
		return 0
	}

	// Collect positional arguments (after the function)
	var args []runtime.Value
	for i := 2; i <= vm.GetTop(); i++ {
		args = append(args, vm.Get(i))
	}

	partial := &PyPartial{
		Func:   fn,
		Args:   args,
		Kwargs: make(map[string]runtime.Value),
	}

	ud := runtime.NewUserData(partial)
	ud.Metatable = runtime.NewDict()
	ud.Metatable.Items[runtime.NewString("__type__")] = runtime.NewString("functools.partial")
	// Store func attribute for introspection
	ud.Metatable.Items[runtime.NewString("func")] = fn
	ud.Metatable.Items[runtime.NewString("args")] = runtime.NewTuple(args)
	ud.Metatable.Items[runtime.NewString("keywords")] = runtime.NewDict()

	vm.Push(ud)
	return 1
}

// partialCall handles calling a partial object
func partialCall(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected partial object")
		return 0
	}
	partial, ok := ud.Value.(*PyPartial)
	if !ok {
		vm.RaiseError("expected partial object")
		return 0
	}

	// Combine stored args with new args
	allArgs := make([]runtime.Value, len(partial.Args))
	copy(allArgs, partial.Args)

	// Add new positional arguments
	for i := 2; i <= vm.GetTop(); i++ {
		allArgs = append(allArgs, vm.Get(i))
	}

	// Call the wrapped function
	result, err := vm.Call(partial.Func, allArgs, partial.Kwargs)
	if err != nil {
		vm.RaiseError("%v", err)
		return 0
	}

	if result != nil {
		vm.Push(result)
		return 1
	}
	return 0
}

// =====================================
// functools.reduce
// =====================================

// functools.reduce(function, iterable[, initializer])
func functoolsReduce(vm *runtime.VM) int {
	if !vm.RequireArgs("reduce", 2) {
		return 0
	}

	fn := vm.Get(1)
	if !runtime.IsCallable(fn) {
		vm.RaiseError("the first argument must be callable")
		return 0
	}

	// Get iterable items
	items := getIterableItems(vm, 2)
	if items == nil {
		return 0
	}

	if len(items) == 0 {
		if vm.GetTop() >= 3 {
			// Return initializer if provided
			vm.Push(vm.Get(3))
			return 1
		}
		vm.RaiseError("reduce() of empty iterable with no initial value")
		return 0
	}

	// Start with initializer or first item
	var accumulator runtime.Value
	startIdx := 0
	if vm.GetTop() >= 3 {
		accumulator = vm.Get(3)
	} else {
		accumulator = items[0]
		startIdx = 1
	}

	// Apply function cumulatively
	for i := startIdx; i < len(items); i++ {
		result, err := vm.Call(fn, []runtime.Value{accumulator, items[i]}, nil)
		if err != nil {
			vm.RaiseError("%v", err)
			return 0
		}
		accumulator = result
	}

	vm.Push(accumulator)
	return 1
}

// =====================================
// functools.wraps and update_wrapper
// =====================================

// WRAPPER_ASSIGNMENTS are attributes to copy from wrapped function
var wrapperAssignments = []string{"__module__", "__name__", "__qualname__", "__annotations__", "__doc__"}

// functools.update_wrapper(wrapper, wrapped, assigned=..., updated=...)
func functoolsUpdateWrapper(vm *runtime.VM) int {
	if !vm.RequireArgs("update_wrapper", 2) {
		return 0
	}

	wrapper := vm.Get(1)
	wrapped := vm.Get(2)

	// Copy assigned attributes
	for _, attr := range wrapperAssignments {
		if val := getFunctionAttr(wrapped, attr); val != nil {
			setFunctionAttr(wrapper, attr, val)
		}
	}

	// Set __wrapped__ to point to the original function
	setFunctionAttr(wrapper, "__wrapped__", wrapped)

	vm.Push(wrapper)
	return 1
}

// functools.wraps(wrapped, assigned=..., updated=...)
// Returns a decorator that applies update_wrapper
func functoolsWraps(vm *runtime.VM) int {
	if !vm.RequireArgs("wraps", 1) {
		return 0
	}

	wrapped := vm.Get(1)

	// Create a decorator function that will call update_wrapper
	decorator := runtime.NewGoFunction("wraps_decorator", func(vm *runtime.VM) int {
		wrapper := vm.Get(1)
		// Copy attributes from wrapped to wrapper
		for _, attr := range wrapperAssignments {
			if val := getFunctionAttr(wrapped, attr); val != nil {
				setFunctionAttr(wrapper, attr, val)
			}
		}
		setFunctionAttr(wrapper, "__wrapped__", wrapped)
		vm.Push(wrapper)
		return 1
	})

	vm.Push(decorator)
	return 1
}

// Helper to get function attribute
func getFunctionAttr(fn runtime.Value, attr string) runtime.Value {
	switch f := fn.(type) {
	case *runtime.PyFunction:
		switch attr {
		case "__name__":
			return runtime.NewString(f.Name)
		case "__doc__":
			return runtime.None
		}
	case *runtime.PyBuiltinFunc:
		switch attr {
		case "__name__":
			return runtime.NewString(f.Name)
		}
	}
	return nil
}

// Helper to set function attribute
func setFunctionAttr(fn runtime.Value, attr string, val runtime.Value) {
	switch f := fn.(type) {
	case *runtime.PyFunction:
		switch attr {
		case "__name__":
			if s, ok := val.(*runtime.PyString); ok {
				f.Name = s.Value
			}
		// __doc__ is not currently writable as CodeObject doesn't store it
		}
	}
}

// =====================================
// functools.cache (unbounded memoization)
// =====================================

// functools.cache
// Same as lru_cache(maxsize=None)
func functoolsCache(vm *runtime.VM) int {
	if !vm.RequireArgs("cache", 1) {
		return 0
	}

	fn := vm.Get(1)
	if !runtime.IsCallable(fn) {
		vm.RaiseError("cache() argument must be callable")
		return 0
	}

	wrapper := createLruCacheWrapper(fn, -1) // -1 means unlimited
	vm.Push(wrapper)
	return 1
}

// =====================================
// functools.lru_cache
// =====================================

// PyLruCache represents an LRU cache wrapper
type PyLruCache struct {
	Func    runtime.Value
	MaxSize int // -1 means unlimited
	Cache   map[string]runtime.Value
	Order   *list.List // For LRU ordering
	Keys    map[string]*list.Element
	Hits    int64
	Misses  int64
	mu      sync.Mutex
}

func (l *PyLruCache) Type() string   { return "functools._lru_cache_wrapper" }
func (l *PyLruCache) String() string { return "<functools._lru_cache_wrapper object>" }

// CacheInfo represents cache statistics
type CacheInfo struct {
	Hits    int64
	Misses  int64
	MaxSize int
	CurrSize int
}

// functools.lru_cache(maxsize=128, typed=False)
func functoolsLruCache(vm *runtime.VM) int {
	maxSize := 128 // Default

	// Check if called with function directly: @lru_cache
	if vm.GetTop() >= 1 {
		arg := vm.Get(1)
		if runtime.IsCallable(arg) {
			// Called as @lru_cache without parens
			wrapper := createLruCacheWrapper(arg, maxSize)
			vm.Push(wrapper)
			return 1
		}
		// Called with maxsize argument
		if !runtime.IsNone(arg) {
			if intVal, ok := arg.(*runtime.PyInt); ok {
				maxSize = int(intVal.Value)
			}
		} else {
			maxSize = -1 // None means unlimited
		}
	}

	// Return a decorator
	decorator := runtime.NewGoFunction("lru_cache_decorator", func(vm *runtime.VM) int {
		fn := vm.Get(1)
		if !runtime.IsCallable(fn) {
			vm.RaiseError("lru_cache() argument must be callable")
			return 0
		}
		wrapper := createLruCacheWrapper(fn, maxSize)
		vm.Push(wrapper)
		return 1
	})

	vm.Push(decorator)
	return 1
}

// createLruCacheWrapper creates an LRU cache wrapper for a function
func createLruCacheWrapper(fn runtime.Value, maxSize int) *runtime.PyUserData {
	cache := &PyLruCache{
		Func:    fn,
		MaxSize: maxSize,
		Cache:   make(map[string]runtime.Value),
		Order:   list.New(),
		Keys:    make(map[string]*list.Element),
	}

	ud := runtime.NewUserData(cache)
	ud.Metatable = runtime.NewDict()
	ud.Metatable.Items[runtime.NewString("__type__")] = runtime.NewString("functools._lru_cache_wrapper")

	return ud
}

// lruCacheCall handles calling the cached function
func lruCacheCall(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected lru_cache wrapper")
		return 0
	}
	cache, ok := ud.Value.(*PyLruCache)
	if !ok {
		vm.RaiseError("expected lru_cache wrapper")
		return 0
	}

	// Build cache key from arguments
	var args []runtime.Value
	for i := 2; i <= vm.GetTop(); i++ {
		args = append(args, vm.Get(i))
	}
	key := makeKey(args)

	cache.mu.Lock()
	defer cache.mu.Unlock()

	// Check cache
	if result, found := cache.Cache[key]; found {
		cache.Hits++
		// Move to front (most recently used)
		if elem, ok := cache.Keys[key]; ok {
			cache.Order.MoveToFront(elem)
		}
		vm.Push(result)
		return 1
	}

	// Cache miss
	cache.Misses++
	cache.mu.Unlock() // Unlock during function call

	result, err := vm.Call(cache.Func, args, nil)

	cache.mu.Lock() // Re-lock to update cache
	if err != nil {
		vm.RaiseError("%v", err)
		return 0
	}

	// Store in cache
	cache.Cache[key] = result
	elem := cache.Order.PushFront(key)
	cache.Keys[key] = elem

	// Evict if over maxsize
	if cache.MaxSize > 0 && cache.Order.Len() > cache.MaxSize {
		oldest := cache.Order.Back()
		if oldest != nil {
			oldKey := oldest.Value.(string)
			delete(cache.Cache, oldKey)
			delete(cache.Keys, oldKey)
			cache.Order.Remove(oldest)
		}
	}

	vm.Push(result)
	return 1
}

// lruCacheCacheInfo returns cache statistics
func lruCacheCacheInfo(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected lru_cache wrapper")
		return 0
	}
	cache, ok := ud.Value.(*PyLruCache)
	if !ok {
		vm.RaiseError("expected lru_cache wrapper")
		return 0
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()

	// Return a named tuple-like dict
	info := runtime.NewDict()
	info.Items[runtime.NewString("hits")] = runtime.NewInt(cache.Hits)
	info.Items[runtime.NewString("misses")] = runtime.NewInt(cache.Misses)
	if cache.MaxSize < 0 {
		info.Items[runtime.NewString("maxsize")] = runtime.None
	} else {
		info.Items[runtime.NewString("maxsize")] = runtime.NewInt(int64(cache.MaxSize))
	}
	info.Items[runtime.NewString("currsize")] = runtime.NewInt(int64(len(cache.Cache)))

	vm.Push(info)
	return 1
}

// lruCacheCacheClear clears the cache
func lruCacheCacheClear(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected lru_cache wrapper")
		return 0
	}
	cache, ok := ud.Value.(*PyLruCache)
	if !ok {
		vm.RaiseError("expected lru_cache wrapper")
		return 0
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()

	cache.Cache = make(map[string]runtime.Value)
	cache.Order = list.New()
	cache.Keys = make(map[string]*list.Element)
	cache.Hits = 0
	cache.Misses = 0

	return 0
}

// makeKey creates a cache key from arguments
func makeKey(args []runtime.Value) string {
	result := ""
	for i, arg := range args {
		if i > 0 {
			result += ","
		}
		result += valueToKey(arg)
	}
	return result
}

// valueToKey converts a value to a string key for caching.
// Uses length-prefixed strings to avoid collisions between values
// whose string representations contain the separator character.
func valueToKey(v runtime.Value) string {
	switch val := v.(type) {
	case *runtime.PyInt:
		return fmt.Sprintf("i:%d", val.Value)
	case *runtime.PyFloat:
		return fmt.Sprintf("f:%f", val.Value)
	case *runtime.PyString:
		return fmt.Sprintf("s:%d:%s", len(val.Value), val.Value)
	case *runtime.PyBool:
		return fmt.Sprintf("b:%t", val.Value)
	case *runtime.PyTuple:
		result := "t:("
		for i, item := range val.Items {
			if i > 0 {
				result += ","
			}
			result += valueToKey(item)
		}
		return result + ")"
	case *runtime.PyNone:
		return "n:None"
	default:
		// For unhashable types, use identity (pointer)
		return fmt.Sprintf("o:%p", v)
	}
}

// =====================================
// functools.cmp_to_key
// =====================================

// PyCmpKey represents a key wrapper for comparison functions
type PyCmpKey struct {
	CmpFunc runtime.Value
	Object  runtime.Value
}

func (c *PyCmpKey) Type() string   { return "functools.KeyWrapper" }
func (c *PyCmpKey) String() string { return fmt.Sprintf("<functools.KeyWrapper object for %v>", c.Object) }

// functools.cmp_to_key(mycmp)
func functoolsCmpToKey(vm *runtime.VM) int {
	if !vm.RequireArgs("cmp_to_key", 1) {
		return 0
	}

	cmpFunc := vm.Get(1)
	if !runtime.IsCallable(cmpFunc) {
		vm.RaiseError("cmp_to_key() argument must be callable")
		return 0
	}

	// Return a factory that creates KeyWrapper objects
	factory := runtime.NewGoFunction("cmp_to_key_factory", func(vm *runtime.VM) int {
		obj := vm.Get(1)

		wrapper := &PyCmpKey{
			CmpFunc: cmpFunc,
			Object:  obj,
		}

		ud := runtime.NewUserData(wrapper)
		ud.Metatable = runtime.NewDict()
		ud.Metatable.Items[runtime.NewString("__type__")] = runtime.NewString("functools.KeyWrapper")

		vm.Push(ud)
		return 1
	})

	vm.Push(factory)
	return 1
}

// Helper for comparison operators
func getCmpKeyPair(vm *runtime.VM) (*PyCmpKey, *PyCmpKey) {
	ud1 := vm.ToUserData(1)
	ud2 := vm.ToUserData(2)
	if ud1 == nil || ud2 == nil {
		return nil, nil
	}
	k1, ok1 := ud1.Value.(*PyCmpKey)
	k2, ok2 := ud2.Value.(*PyCmpKey)
	if !ok1 || !ok2 {
		return nil, nil
	}
	return k1, k2
}

// callCmpFunc calls the comparison function and returns the result
func callCmpFunc(vm *runtime.VM, k1, k2 *PyCmpKey) (int64, bool) {
	result, err := vm.Call(k1.CmpFunc, []runtime.Value{k1.Object, k2.Object}, nil)
	if err != nil {
		vm.RaiseError("%v", err)
		return 0, false
	}
	if intVal, ok := result.(*runtime.PyInt); ok {
		return intVal.Value, true
	}
	vm.RaiseError("comparison function must return an integer")
	return 0, false
}

func cmpKeyLt(vm *runtime.VM) int {
	k1, k2 := getCmpKeyPair(vm)
	if k1 == nil || k2 == nil {
		vm.RaiseError("comparison requires two KeyWrapper objects")
		return 0
	}
	cmp, ok := callCmpFunc(vm, k1, k2)
	if !ok {
		return 0
	}
	vm.Push(runtime.NewBool(cmp < 0))
	return 1
}

func cmpKeyGt(vm *runtime.VM) int {
	k1, k2 := getCmpKeyPair(vm)
	if k1 == nil || k2 == nil {
		vm.RaiseError("comparison requires two KeyWrapper objects")
		return 0
	}
	cmp, ok := callCmpFunc(vm, k1, k2)
	if !ok {
		return 0
	}
	vm.Push(runtime.NewBool(cmp > 0))
	return 1
}

func cmpKeyEq(vm *runtime.VM) int {
	k1, k2 := getCmpKeyPair(vm)
	if k1 == nil || k2 == nil {
		vm.RaiseError("comparison requires two KeyWrapper objects")
		return 0
	}
	cmp, ok := callCmpFunc(vm, k1, k2)
	if !ok {
		return 0
	}
	vm.Push(runtime.NewBool(cmp == 0))
	return 1
}

func cmpKeyLe(vm *runtime.VM) int {
	k1, k2 := getCmpKeyPair(vm)
	if k1 == nil || k2 == nil {
		vm.RaiseError("comparison requires two KeyWrapper objects")
		return 0
	}
	cmp, ok := callCmpFunc(vm, k1, k2)
	if !ok {
		return 0
	}
	vm.Push(runtime.NewBool(cmp <= 0))
	return 1
}

func cmpKeyGe(vm *runtime.VM) int {
	k1, k2 := getCmpKeyPair(vm)
	if k1 == nil || k2 == nil {
		vm.RaiseError("comparison requires two KeyWrapper objects")
		return 0
	}
	cmp, ok := callCmpFunc(vm, k1, k2)
	if !ok {
		return 0
	}
	vm.Push(runtime.NewBool(cmp >= 0))
	return 1
}
