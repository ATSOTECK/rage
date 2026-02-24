package stdlib

import (
	"math/rand"
	"sync"
	"time"

	"github.com/ATSOTECK/rage/internal/runtime"
)

// randMu protects globalRand from concurrent access.
// *rand.Rand is not thread-safe, and multiple VM instances may call random functions concurrently.
var randMu sync.Mutex

// globalRand is the default random source
var globalRand = rand.New(rand.NewSource(time.Now().UnixNano()))

// InitRandomModule registers the random module
func InitRandomModule() {
	runtime.NewModuleBuilder("random").
		Doc("Random number generation.").
		// Core functions
		Func("random", randomRandom).
		Func("seed", randomSeed).
		// Integer functions
		Func("randint", randomRandint).
		Func("randrange", randomRandrange).
		Func("getrandbits", randomGetrandbits).
		// Sequence functions
		Func("choice", randomChoice).
		Func("choices", randomChoices).
		Func("shuffle", randomShuffle).
		Func("sample", randomSample).
		// Float functions
		Func("uniform", randomUniform).
		Func("triangular", randomTriangular).
		Func("gauss", randomGauss).
		Func("normalvariate", randomGauss). // alias
		Register()
}

// random() -> float in [0.0, 1.0)
func randomRandom(vm *runtime.VM) int {
	randMu.Lock()
	f := globalRand.Float64()
	randMu.Unlock()
	vm.Push(runtime.NewFloat(f))
	return 1
}

// seed(n) -> None
func randomSeed(vm *runtime.VM) int {
	n := vm.ToInt(1)
	randMu.Lock()
	if n == 0 {
		globalRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	} else {
		globalRand = rand.New(rand.NewSource(n))
	}
	randMu.Unlock()
	return 0
}

// randint(a, b) -> integer in [a, b] (inclusive)
func randomRandint(vm *runtime.VM) int {
	a := vm.CheckInt(1)
	b := vm.CheckInt(2)
	if a > b {
		vm.RaiseError("empty range for randint()")
		return 0
	}
	randMu.Lock()
	result := a + globalRand.Int63n(b-a+1)
	randMu.Unlock()
	vm.Push(runtime.NewInt(result))
	return 1
}

// randrange(stop) or randrange(start, stop[, step]) -> random int from range
func randomRandrange(vm *runtime.VM) int {
	if !vm.RequireArgs("randrange", 1) {
		return 0
	}

	top := vm.GetTop()

	var start int64
	var stop int64
	var step int64 = 1

	switch top {
	case 1:
		stop = vm.CheckInt(1)
	case 2:
		start = vm.CheckInt(1)
		stop = vm.CheckInt(2)
	case 3:
		start = vm.CheckInt(1)
		stop = vm.CheckInt(2)
		step = vm.CheckInt(3)
	default:
		vm.RaiseError("randrange expected 1 to 3 arguments, got %d", top)
		return 0
	}

	if step == 0 {
		vm.RaiseError("zero step for randrange()")
		return 0
	}

	var n int64
	if step > 0 {
		n = (stop - start + step - 1) / step
	} else {
		n = (start - stop - step - 1) / (-step)
	}

	if n <= 0 {
		vm.RaiseError("empty range for randrange()")
		return 0
	}

	randMu.Lock()
	result := start + step*globalRand.Int63n(n)
	randMu.Unlock()
	vm.Push(runtime.NewInt(result))
	return 1
}

// getrandbits(k) -> integer with k random bits
func randomGetrandbits(vm *runtime.VM) int {
	k := vm.CheckInt(1)
	if k < 0 {
		vm.RaiseError("number of bits must be non-negative")
		return 0
	}
	if k == 0 {
		vm.Push(runtime.NewInt(0))
		return 1
	}
	if k > 63 {
		k = 63 // Limit to int64 range
	}
	randMu.Lock()
	result := globalRand.Int63n(1 << k)
	randMu.Unlock()
	vm.Push(runtime.NewInt(result))
	return 1
}

// choice(seq) -> random element from sequence
func randomChoice(vm *runtime.VM) int {
	seq := vm.Get(1)

	var items []runtime.Value
	switch s := seq.(type) {
	case *runtime.PyList:
		items = s.Items
	case *runtime.PyTuple:
		items = s.Items
	case *runtime.PyString:
		// Convert string to list of characters
		for _, ch := range s.Value {
			items = append(items, runtime.NewString(string(ch)))
		}
	default:
		vm.RaiseError("choice() argument must be a sequence")
		return 0
	}

	if len(items) == 0 {
		vm.RaiseError("cannot choose from an empty sequence")
		return 0
	}

	randMu.Lock()
	idx := globalRand.Intn(len(items))
	randMu.Unlock()
	vm.Push(items[idx])
	return 1
}

// choices(seq, k=1) -> list of k random elements (with replacement)
func randomChoices(vm *runtime.VM) int {
	seq := vm.Get(1)
	k := int(vm.ToInt(2))
	if k <= 0 {
		k = 1
	}

	var items []runtime.Value
	switch s := seq.(type) {
	case *runtime.PyList:
		items = s.Items
	case *runtime.PyTuple:
		items = s.Items
	case *runtime.PyString:
		for _, ch := range s.Value {
			items = append(items, runtime.NewString(string(ch)))
		}
	default:
		vm.RaiseError("choices() argument must be a sequence")
		return 0
	}

	if len(items) == 0 {
		vm.RaiseError("cannot choose from an empty sequence")
		return 0
	}

	result := make([]runtime.Value, k)
	randMu.Lock()
	for i := 0; i < k; i++ {
		result[i] = items[globalRand.Intn(len(items))]
	}
	randMu.Unlock()
	vm.Push(runtime.NewList(result))
	return 1
}

// shuffle(seq) -> None (shuffles list in place)
func randomShuffle(vm *runtime.VM) int {
	seq := vm.Get(1)

	list, ok := seq.(*runtime.PyList)
	if !ok {
		vm.RaiseError("shuffle() argument must be a list")
		return 0
	}

	// Fisher-Yates shuffle
	n := len(list.Items)
	randMu.Lock()
	for i := n - 1; i > 0; i-- {
		j := globalRand.Intn(i + 1)
		list.Items[i], list.Items[j] = list.Items[j], list.Items[i]
	}
	randMu.Unlock()

	return 0
}

// sample(seq, k) -> list of k unique random elements
func randomSample(vm *runtime.VM) int {
	seq := vm.Get(1)
	k := int(vm.CheckInt(2))

	var items []runtime.Value
	switch s := seq.(type) {
	case *runtime.PyList:
		items = s.Items
	case *runtime.PyTuple:
		items = s.Items
	case *runtime.PyString:
		for _, ch := range s.Value {
			items = append(items, runtime.NewString(string(ch)))
		}
	default:
		vm.RaiseError("sample() argument must be a sequence")
		return 0
	}

	n := len(items)
	if k < 0 || k > n {
		vm.RaiseError("sample larger than population or negative")
		return 0
	}

	// Create a copy and shuffle first k elements
	pool := make([]runtime.Value, n)
	copy(pool, items)

	result := make([]runtime.Value, k)
	randMu.Lock()
	for i := 0; i < k; i++ {
		j := globalRand.Intn(n - i)
		result[i] = pool[j]
		pool[j] = pool[n-i-1]
	}
	randMu.Unlock()

	vm.Push(runtime.NewList(result))
	return 1
}

// uniform(a, b) -> random float in [a, b]
func randomUniform(vm *runtime.VM) int {
	a := vm.CheckFloat(1)
	b := vm.CheckFloat(2)
	randMu.Lock()
	result := a + (b-a)*globalRand.Float64()
	randMu.Unlock()
	vm.Push(runtime.NewFloat(result))
	return 1
}

// triangular(low, high, mode) -> random float with triangular distribution
func randomTriangular(vm *runtime.VM) int {
	top := vm.GetTop()

	var low, high, mode float64 = 0.0, 1.0, 0.5

	switch top {
	case 0:
		// defaults
	case 1:
		high = vm.CheckFloat(1)
	case 2:
		low = vm.CheckFloat(1)
		high = vm.CheckFloat(2)
		mode = (low + high) / 2
	case 3:
		low = vm.CheckFloat(1)
		high = vm.CheckFloat(2)
		mode = vm.CheckFloat(3)
	}

	randMu.Lock()
	u := globalRand.Float64()
	randMu.Unlock()
	c := (mode - low) / (high - low)

	var result float64
	if u < c {
		result = low + (high-low)*sqrtFloat(u*c)
	} else {
		result = high - (high-low)*sqrtFloat((1-u)*(1-c))
	}

	vm.Push(runtime.NewFloat(result))
	return 1
}

// gauss(mu, sigma) -> random float with Gaussian distribution
func randomGauss(vm *runtime.VM) int {
	mu := vm.CheckFloat(1)
	sigma := vm.CheckFloat(2)
	randMu.Lock()
	result := globalRand.NormFloat64()*sigma + mu
	randMu.Unlock()
	vm.Push(runtime.NewFloat(result))
	return 1
}

// sqrtFloat is a simple square root for floats
func sqrtFloat(x float64) float64 {
	if x < 0 {
		return 0
	}
	// Newton's method
	z := x
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}
