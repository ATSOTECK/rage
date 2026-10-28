package stdlib

import (
	"math"
	"math/big"

	"github.com/ATSOTECK/rage/internal/runtime"
)

// mathUnaryFloat wraps a unary float64 -> float64 function as a GoFunction.
func mathUnaryFloat(fn func(float64) float64) runtime.GoFunction {
	return func(vm *runtime.VM) int {
		x := vm.CheckFloat(1)
		vm.Push(runtime.NewFloat(fn(x)))
		return 1
	}
}

// mathBinaryFloat wraps a binary (float64, float64) -> float64 function as a GoFunction.
func mathBinaryFloat(fn func(float64, float64) float64) runtime.GoFunction {
	return func(vm *runtime.VM) int {
		x := vm.CheckFloat(1)
		y := vm.CheckFloat(2)
		vm.Push(runtime.NewFloat(fn(x, y)))
		return 1
	}
}

// InitMathModule registers the math module
func InitMathModule() {
	runtime.NewModuleBuilder("math").
		Doc("Mathematical functions and constants.").
		// Constants
		Const("pi", runtime.NewFloat(math.Pi)).
		Const("e", runtime.NewFloat(math.E)).
		Const("tau", runtime.NewFloat(math.Pi*2)).
		Const("inf", runtime.NewFloat(math.Inf(1))).
		Const("nan", runtime.NewFloat(math.NaN())).
		// Basic functions
		Func("sqrt", mathUnaryFloat(math.Sqrt)).
		Func("pow", mathBinaryFloat(math.Pow)).
		Func("exp", mathUnaryFloat(math.Exp)).
		Func("log", mathUnaryFloat(math.Log)).
		Func("log10", mathUnaryFloat(math.Log10)).
		Func("log2", mathUnaryFloat(math.Log2)).
		// Trigonometric functions
		Func("sin", mathUnaryFloat(math.Sin)).
		Func("cos", mathUnaryFloat(math.Cos)).
		Func("tan", mathUnaryFloat(math.Tan)).
		Func("asin", mathUnaryFloat(math.Asin)).
		Func("acos", mathUnaryFloat(math.Acos)).
		Func("atan", mathUnaryFloat(math.Atan)).
		Func("atan2", mathBinaryFloat(math.Atan2)).
		// Hyperbolic functions
		Func("sinh", mathUnaryFloat(math.Sinh)).
		Func("cosh", mathUnaryFloat(math.Cosh)).
		Func("tanh", mathUnaryFloat(math.Tanh)).
		// Rounding and absolute value
		Func("ceil", mathCeil).
		Func("floor", mathFloor).
		Func("trunc", mathTrunc).
		Func("fabs", mathUnaryFloat(math.Abs)).
		Func("copysign", mathBinaryFloat(math.Copysign)).
		// Special functions
		Func("factorial", mathFactorial).
		Func("gcd", mathGcd).
		Func("lcm", mathLcm).
		Func("isqrt", mathIsqrt).
		Func("comb", mathComb).
		Func("perm", mathPerm).
		Func("prod", mathProd).
		Func("fsum", mathFsum).
		Func("isnan", mathIsnan).
		Func("isinf", mathIsinf).
		Func("isfinite", mathIsfinite).
		Func("degrees", mathDegrees).
		Func("radians", mathRadians).
		Register()
}

// Special functions that cannot use the generic wrappers.

func mathIsnan(vm *runtime.VM) int {
	x := vm.CheckFloat(1)
	vm.Push(runtime.NewBool(math.IsNaN(x)))
	return 1
}

func mathIsinf(vm *runtime.VM) int {
	x := vm.CheckFloat(1)
	vm.Push(runtime.NewBool(math.IsInf(x, 0)))
	return 1
}

func mathIsfinite(vm *runtime.VM) int {
	x := vm.CheckFloat(1)
	vm.Push(runtime.NewBool(!math.IsNaN(x) && !math.IsInf(x, 0)))
	return 1
}

func mathDegrees(vm *runtime.VM) int {
	x := vm.CheckFloat(1)
	vm.Push(runtime.NewFloat(x * 180.0 / math.Pi))
	return 1
}

func mathRadians(vm *runtime.VM) int {
	x := vm.CheckFloat(1)
	vm.Push(runtime.NewFloat(x * math.Pi / 180.0))
	return 1
}

// Integer functions, computed with arbitrary precision like CPython's.

// mathIntArg returns argument n as a big.Int, raising TypeError for non-ints.
func mathIntArg(vm *runtime.VM, n int) (*big.Int, bool) {
	i, err := vm.IndexInt(vm.Get(n))
	if err != nil {
		vm.RaiseError("%s", err.Error())
		return nil, false
	}
	return i.BigIntValue(), true
}

func mathFactorial(vm *runtime.VM) int {
	n, ok := mathIntArg(vm, 1)
	if !ok {
		return 0
	}
	if n.Sign() < 0 {
		vm.RaiseError("ValueError: factorial() not defined for negative values")
		return 0
	}
	if !n.IsInt64() || n.Int64() > 1<<20 {
		vm.RaiseError("OverflowError: factorial() argument should not exceed %d", 1<<20)
		return 0
	}
	vm.Push(runtime.MakeBigInt(new(big.Int).MulRange(1, n.Int64())))
	return 1
}

func mathGcd(vm *runtime.VM) int {
	result := new(big.Int)
	for i := 1; i <= vm.GetTop(); i++ {
		x, ok := mathIntArg(vm, i)
		if !ok {
			return 0
		}
		result.GCD(nil, nil, result.Abs(result), new(big.Int).Abs(x))
	}
	vm.Push(runtime.MakeBigInt(result))
	return 1
}

func mathLcm(vm *runtime.VM) int {
	result := big.NewInt(1)
	for i := 1; i <= vm.GetTop(); i++ {
		x, ok := mathIntArg(vm, i)
		if !ok {
			return 0
		}
		x = new(big.Int).Abs(x)
		if x.Sign() == 0 || result.Sign() == 0 {
			result.SetInt64(0)
			continue
		}
		g := new(big.Int).GCD(nil, nil, result, x)
		result.Mul(result, new(big.Int).Quo(x, g))
	}
	vm.Push(runtime.MakeBigInt(result))
	return 1
}

func mathIsqrt(vm *runtime.VM) int {
	n, ok := mathIntArg(vm, 1)
	if !ok {
		return 0
	}
	if n.Sign() < 0 {
		vm.RaiseError("ValueError: isqrt() argument must be nonnegative")
		return 0
	}
	vm.Push(runtime.MakeBigInt(new(big.Int).Sqrt(n)))
	return 1
}

// mathCombPermArgs reads (n, k) for comb/perm, both non-negative.
func mathCombPermArgs(vm *runtime.VM) (int64, int64, bool) {
	n, ok := mathIntArg(vm, 1)
	if !ok {
		return 0, 0, false
	}
	var k *big.Int
	if vm.GetTop() >= 2 && vm.Get(2) != runtime.None {
		if k, ok = mathIntArg(vm, 2); !ok {
			return 0, 0, false
		}
	} else {
		k = n
	}
	if n.Sign() < 0 {
		vm.RaiseError("ValueError: n must be a non-negative integer")
		return 0, 0, false
	}
	if k.Sign() < 0 {
		vm.RaiseError("ValueError: k must be a non-negative integer")
		return 0, 0, false
	}
	if !n.IsInt64() || !k.IsInt64() {
		vm.RaiseError("OverflowError: n or k too large")
		return 0, 0, false
	}
	return n.Int64(), k.Int64(), true
}

func mathComb(vm *runtime.VM) int {
	n, k, ok := mathCombPermArgs(vm)
	if !ok {
		return 0
	}
	if k > n {
		vm.Push(runtime.NewInt(0))
		return 1
	}
	vm.Push(runtime.MakeBigInt(new(big.Int).Binomial(n, k)))
	return 1
}

func mathPerm(vm *runtime.VM) int {
	n, k, ok := mathCombPermArgs(vm)
	if !ok {
		return 0
	}
	if k > n {
		vm.Push(runtime.NewInt(0))
		return 1
	}
	if k == 0 {
		vm.Push(runtime.NewInt(1))
		return 1
	}
	vm.Push(runtime.MakeBigInt(new(big.Int).MulRange(n-k+1, n)))
	return 1
}

func mathProd(vm *runtime.VM) int {
	items, err := vm.ToList(vm.Get(1))
	if err != nil {
		vm.RaiseError("%s", err.Error())
		return 0
	}
	var result runtime.Value = runtime.NewInt(1)
	if vm.GetTop() >= 2 {
		result = vm.Get(2)
	}
	for _, item := range items {
		result, err = vm.BinaryOp(runtime.OpBinaryMultiply, result, item)
		if err != nil {
			vm.RaiseError("%s", err.Error())
			return 0
		}
	}
	vm.Push(result)
	return 1
}

// mathFsum is math.fsum: an exactly rounded float sum (Shewchuk's algorithm,
// as in CPython).
func mathFsum(vm *runtime.VM) int {
	items, err := vm.ToList(vm.Get(1))
	if err != nil {
		vm.RaiseError("%s", err.Error())
		return 0
	}
	var partials []float64
	for _, item := range items {
		var x float64
		switch v := item.(type) {
		case *runtime.PyFloat:
			x = v.Value
		case *runtime.PyInt:
			x = runtime.IntToFloat64(v)
		case *runtime.PyBool:
			if v.Value {
				x = 1
			}
		default:
			vm.RaiseError("TypeError: must be real number, not %s", vm.TypeNameOf(item))
			return 0
		}
		i := 0
		for _, y := range partials {
			if math.Abs(x) < math.Abs(y) {
				x, y = y, x
			}
			hi := x + y
			lo := y - (hi - x)
			if lo != 0 {
				partials[i] = lo
				i++
			}
			x = hi
		}
		partials = append(partials[:i], x)
	}
	// Sum the partials from the top, handling the half-way rounding case
	total := 0.0
	n := len(partials)
	if n > 0 {
		n--
		total = partials[n]
		for n > 0 {
			x := total
			n--
			y := partials[n]
			total = x + y
			yr := total - x
			lo := y - yr
			if lo != 0 {
				if n > 0 && ((lo < 0 && partials[n-1] < 0) || (lo > 0 && partials[n-1] > 0)) {
					y = lo * 2
					x = total + y
					yr = x - total
					if y == yr {
						total = x
					}
				}
				break
			}
		}
	}
	vm.Push(runtime.NewFloat(total))
	return 1
}

// mathRoundTo implements floor/ceil/trunc: ints are returned unchanged and
// floats are converted exactly, giving big ints for large values.
func mathRoundTo(vm *runtime.VM, dunder string, fn func(float64) float64) int {
	v := vm.Get(1)
	switch n := v.(type) {
	case *runtime.PyInstance:
		if result, found := vm.CallDunder(n, dunder); found {
			vm.Push(result)
			return 1
		}
		vm.RaiseError("TypeError: type %s doesn't define %s method", n.Class.Name, dunder)
		return 0
	case *runtime.PyInt:
		vm.Push(n)
		return 1
	case *runtime.PyBool:
		if n.Value {
			vm.Push(runtime.NewInt(1))
		} else {
			vm.Push(runtime.NewInt(0))
		}
		return 1
	}
	x := fn(vm.CheckFloat(1))
	if math.IsInf(x, 0) {
		vm.RaiseError("OverflowError: cannot convert float infinity to integer")
		return 0
	}
	if math.IsNaN(x) {
		vm.RaiseError("ValueError: cannot convert float NaN to integer")
		return 0
	}
	i, _ := new(big.Float).SetFloat64(x).Int(nil)
	vm.Push(runtime.MakeBigInt(i))
	return 1
}

func mathCeil(vm *runtime.VM) int  { return mathRoundTo(vm, "__ceil__", math.Ceil) }
func mathFloor(vm *runtime.VM) int { return mathRoundTo(vm, "__floor__", math.Floor) }
func mathTrunc(vm *runtime.VM) int { return mathRoundTo(vm, "__trunc__", math.Trunc) }
