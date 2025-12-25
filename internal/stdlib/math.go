package stdlib

import (
	"math"

	"github.com/ATSOTECK/rage/internal/runtime"
)

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
		Func("sqrt", mathSqrt).
		Func("pow", mathPow).
		Func("exp", mathExp).
		Func("log", mathLog).
		Func("log10", mathLog10).
		Func("log2", mathLog2).
		// Trigonometric functions
		Func("sin", mathSin).
		Func("cos", mathCos).
		Func("tan", mathTan).
		Func("asin", mathAsin).
		Func("acos", mathAcos).
		Func("atan", mathAtan).
		Func("atan2", mathAtan2).
		// Hyperbolic functions
		Func("sinh", mathSinh).
		Func("cosh", mathCosh).
		Func("tanh", mathTanh).
		// Rounding and absolute value
		Func("ceil", mathCeil).
		Func("floor", mathFloor).
		Func("trunc", mathTrunc).
		Func("fabs", mathFabs).
		Func("copysign", mathCopysign).
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

// Basic functions

func mathSqrt(vm *runtime.VM) int {
	x := vm.CheckFloat(1)
	vm.Push(runtime.NewFloat(math.Sqrt(x)))
	return 1
}

func mathPow(vm *runtime.VM) int {
	x := vm.CheckFloat(1)
	y := vm.CheckFloat(2)
	vm.Push(runtime.NewFloat(math.Pow(x, y)))
	return 1
}

func mathExp(vm *runtime.VM) int {
	x := vm.CheckFloat(1)
	vm.Push(runtime.NewFloat(math.Exp(x)))
	return 1
}

func mathLog(vm *runtime.VM) int {
	x := vm.CheckFloat(1)
	vm.Push(runtime.NewFloat(math.Log(x)))
	return 1
}

func mathLog10(vm *runtime.VM) int {
	x := vm.CheckFloat(1)
	vm.Push(runtime.NewFloat(math.Log10(x)))
	return 1
}

func mathLog2(vm *runtime.VM) int {
	x := vm.CheckFloat(1)
	vm.Push(runtime.NewFloat(math.Log2(x)))
	return 1
}

// Trigonometric functions

func mathSin(vm *runtime.VM) int {
	x := vm.CheckFloat(1)
	vm.Push(runtime.NewFloat(math.Sin(x)))
	return 1
}

func mathCos(vm *runtime.VM) int {
	x := vm.CheckFloat(1)
	vm.Push(runtime.NewFloat(math.Cos(x)))
	return 1
}

func mathTan(vm *runtime.VM) int {
	x := vm.CheckFloat(1)
	vm.Push(runtime.NewFloat(math.Tan(x)))
	return 1
}

func mathAsin(vm *runtime.VM) int {
	x := vm.CheckFloat(1)
	vm.Push(runtime.NewFloat(math.Asin(x)))
	return 1
}

func mathAcos(vm *runtime.VM) int {
	x := vm.CheckFloat(1)
	vm.Push(runtime.NewFloat(math.Acos(x)))
	return 1
}

func mathAtan(vm *runtime.VM) int {
	x := vm.CheckFloat(1)
	vm.Push(runtime.NewFloat(math.Atan(x)))
	return 1
}

func mathAtan2(vm *runtime.VM) int {
	y := vm.CheckFloat(1)
	x := vm.CheckFloat(2)
	vm.Push(runtime.NewFloat(math.Atan2(y, x)))
	return 1
}

// Hyperbolic functions

func mathSinh(vm *runtime.VM) int {
	x := vm.CheckFloat(1)
	vm.Push(runtime.NewFloat(math.Sinh(x)))
	return 1
}

func mathCosh(vm *runtime.VM) int {
	x := vm.CheckFloat(1)
	vm.Push(runtime.NewFloat(math.Cosh(x)))
	return 1
}

func mathTanh(vm *runtime.VM) int {
	x := vm.CheckFloat(1)
	vm.Push(runtime.NewFloat(math.Tanh(x)))
	return 1
}

// Rounding functions

func mathCeil(vm *runtime.VM) int {
	x := vm.CheckFloat(1)
	vm.Push(runtime.NewFloat(math.Ceil(x)))
	return 1
}

func mathFloor(vm *runtime.VM) int {
	x := vm.CheckFloat(1)
	vm.Push(runtime.NewFloat(math.Floor(x)))
	return 1
}

func mathTrunc(vm *runtime.VM) int {
	x := vm.CheckFloat(1)
	vm.Push(runtime.NewFloat(math.Trunc(x)))
	return 1
}

func mathFabs(vm *runtime.VM) int {
	x := vm.CheckFloat(1)
	vm.Push(runtime.NewFloat(math.Abs(x)))
	return 1
}

func mathCopysign(vm *runtime.VM) int {
	x := vm.CheckFloat(1)
	y := vm.CheckFloat(2)
	vm.Push(runtime.NewFloat(math.Copysign(x, y)))
	return 1
}

// Special functions

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
