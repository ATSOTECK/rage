package stdlib

import (
	"math"

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
		Func("isnan", mathIsnan).
		Func("isinf", mathIsinf).
		Func("isfinite", mathIsfinite).
		Func("degrees", mathDegrees).
		Func("radians", mathRadians).
		Register()
}

// Special functions that cannot use the generic wrappers.

func mathFactorial(vm *runtime.VM) int {
	n := vm.CheckInt(1)
	if n < 0 {
		vm.RaiseError("factorial() not defined for negative values")
		return 0
	}
	result := int64(1)
	for i := int64(2); i <= n; i++ {
		result *= i
	}
	vm.Push(runtime.NewInt(result))
	return 1
}

func mathGcd(vm *runtime.VM) int {
	a := vm.CheckInt(1)
	b := vm.CheckInt(2)
	if a < 0 {
		a = -a
	}
	if b < 0 {
		b = -b
	}
	for b != 0 {
		a, b = b, a%b
	}
	vm.Push(runtime.NewInt(a))
	return 1
}

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

func mathCeil(vm *runtime.VM) int {
	v := vm.Get(1)
	if inst, ok := v.(*runtime.PyInstance); ok {
		if result, found := vm.CallDunder(inst, "__ceil__"); found {
			vm.Push(result)
			return 1
		}
		vm.RaiseError("TypeError: type %s doesn't define __ceil__ method", inst.Class.Name)
		return 0
	}
	x := vm.CheckFloat(1)
	vm.Push(runtime.NewInt(int64(math.Ceil(x))))
	return 1
}

func mathFloor(vm *runtime.VM) int {
	v := vm.Get(1)
	if inst, ok := v.(*runtime.PyInstance); ok {
		if result, found := vm.CallDunder(inst, "__floor__"); found {
			vm.Push(result)
			return 1
		}
		vm.RaiseError("TypeError: type %s doesn't define __floor__ method", inst.Class.Name)
		return 0
	}
	x := vm.CheckFloat(1)
	vm.Push(runtime.NewInt(int64(math.Floor(x))))
	return 1
}

func mathTrunc(vm *runtime.VM) int {
	v := vm.Get(1)
	if inst, ok := v.(*runtime.PyInstance); ok {
		if result, found := vm.CallDunder(inst, "__trunc__"); found {
			vm.Push(result)
			return 1
		}
		vm.RaiseError("TypeError: type %s doesn't define __trunc__ method", inst.Class.Name)
		return 0
	}
	x := vm.CheckFloat(1)
	vm.Push(runtime.NewInt(int64(math.Trunc(x))))
	return 1
}
