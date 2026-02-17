package test

import (
	"testing"

	"github.com/ATSOTECK/rage/internal/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =====================================
// len()
// =====================================

func TestBuiltinLen(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   int64
	}{
		{"str", `result = len("hello")`, 5},
		{"empty str", `result = len("")`, 0},
		{"list", `result = len([1, 2, 3])`, 3},
		{"empty list", `result = len([])`, 0},
		{"tuple", `result = len((1, 2, 3, 4))`, 4},
		{"dict", `result = len({"a": 1, "b": 2})`, 2},
		{"set", `result = len({1, 2, 3})`, 3},
		{"range", `result = len(range(10))`, 10},
		{"bytes", `result = len(b"abc")`, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := runCode(t, tt.source)
			result := vm.GetGlobal("result").(*runtime.PyInt)
			assert.Equal(t, tt.want, result.Value)
		})
	}
}

func TestBuiltinLenTypeError(t *testing.T) {
	runCodeExpectError(t, `len(42)`, "has no len()")
}

// =====================================
// abs()
// =====================================

func TestBuiltinAbs(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"positive int", `result = abs(5); expected = 5`},
		{"negative int", `result = abs(-5); expected = 5`},
		{"zero", `result = abs(0); expected = 0`},
		{"positive float", `result = abs(3.14); expected = 3.14`},
		{"negative float", `result = abs(-3.14); expected = 3.14`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := runCode(t, tt.source)
			result := vm.GetGlobal("result")
			expected := vm.GetGlobal("expected")
			switch r := result.(type) {
			case *runtime.PyInt:
				assert.Equal(t, expected.(*runtime.PyInt).Value, r.Value)
			case *runtime.PyFloat:
				assert.InDelta(t, expected.(*runtime.PyFloat).Value, r.Value, 1e-10)
			}
		})
	}
}

func TestBuiltinAbsComplex(t *testing.T) {
	vm := runCode(t, `result = abs(3+4j)`)
	result := vm.GetGlobal("result").(*runtime.PyFloat)
	assert.InDelta(t, 5.0, result.Value, 1e-10)
}

func TestBuiltinAbsTypeError(t *testing.T) {
	runCodeExpectError(t, `abs("hello")`, "bad operand type for abs()")
}

// =====================================
// min() / max()
// =====================================

func TestBuiltinMinMax(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   int64
	}{
		{"min list", `result = min([3, 1, 2])`, 1},
		{"min args", `result = min(5, 2, 8)`, 2},
		{"max list", `result = max([3, 1, 2])`, 3},
		{"max args", `result = max(5, 2, 8)`, 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := runCode(t, tt.source)
			result := vm.GetGlobal("result").(*runtime.PyInt)
			assert.Equal(t, tt.want, result.Value)
		})
	}
}

func TestBuiltinMinEmpty(t *testing.T) {
	runCodeExpectError(t, `min([])`, "empty")
}

func TestBuiltinMaxEmpty(t *testing.T) {
	runCodeExpectError(t, `max([])`, "empty")
}

func TestBuiltinMinMaxKey(t *testing.T) {
	vm := runCode(t, `
result = min(["hello", "hi", "hey"], key=len)
`)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "hi", result.Value)
}

// =====================================
// sum()
// =====================================

func TestBuiltinSum(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   int64
	}{
		{"basic", `result = sum([1, 2, 3])`, 6},
		{"empty", `result = sum([])`, 0},
		{"with start", `result = sum([1, 2, 3], 10)`, 16},
		{"range", `result = sum(range(5))`, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := runCode(t, tt.source)
			result := vm.GetGlobal("result").(*runtime.PyInt)
			assert.Equal(t, tt.want, result.Value)
		})
	}
}

func TestBuiltinSumFloat(t *testing.T) {
	vm := runCode(t, `result = sum([1.5, 2.5, 3.0])`)
	result := vm.GetGlobal("result").(*runtime.PyFloat)
	assert.InDelta(t, 7.0, result.Value, 1e-10)
}

// =====================================
// sorted()
// =====================================

func TestBuiltinSorted(t *testing.T) {
	vm := runCode(t, `
result = sorted([3, 1, 4, 1, 5, 9, 2, 6])
`)
	result := vm.GetGlobal("result").(*runtime.PyList)
	expected := []int64{1, 1, 2, 3, 4, 5, 6, 9}
	require.Equal(t, len(expected), len(result.Items))
	for i, want := range expected {
		assert.Equal(t, want, result.Items[i].(*runtime.PyInt).Value)
	}
}

func TestBuiltinSortedReverse(t *testing.T) {
	vm := runCode(t, `result = sorted([3, 1, 2], reverse=True)`)
	result := vm.GetGlobal("result").(*runtime.PyList)
	assert.Equal(t, int64(3), result.Items[0].(*runtime.PyInt).Value)
	assert.Equal(t, int64(2), result.Items[1].(*runtime.PyInt).Value)
	assert.Equal(t, int64(1), result.Items[2].(*runtime.PyInt).Value)
}

func TestBuiltinSortedKey(t *testing.T) {
	vm := runCode(t, `result = sorted(["banana", "apple", "cherry"], key=len)`)
	result := vm.GetGlobal("result").(*runtime.PyList)
	assert.Equal(t, "apple", result.Items[0].(*runtime.PyString).Value)
}

// =====================================
// int() conversions
// =====================================

func TestBuiltinIntConversions(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   int64
	}{
		{"no args", `result = int()`, 0},
		{"from string", `result = int("123")`, 123},
		{"from float", `result = int(3.9)`, 3},
		{"from bool", `result = int(True)`, 1},
		{"hex string", `result = int("ff", 16)`, 255},
		{"bin string", `result = int("1010", 2)`, 10},
		{"oct string", `result = int("77", 8)`, 63},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := runCode(t, tt.source)
			result := vm.GetGlobal("result").(*runtime.PyInt)
			assert.Equal(t, tt.want, result.Value)
		})
	}
}

func TestBuiltinIntErrors(t *testing.T) {
	tests := []struct {
		name   string
		source string
		errMsg string
	}{
		{"invalid string", `int("abc")`, "invalid literal"},
		{"empty string", `int("")`, "invalid literal"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runCodeExpectError(t, tt.source, tt.errMsg)
		})
	}
}

// =====================================
// float() conversions
// =====================================

func TestBuiltinFloatConversions(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   float64
	}{
		{"from string", `result = float("3.14")`, 3.14},
		{"from int", `result = float(42)`, 42.0},
		{"no args", `result = float()`, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := runCode(t, tt.source)
			result := vm.GetGlobal("result").(*runtime.PyFloat)
			assert.InDelta(t, tt.want, result.Value, 1e-10)
		})
	}
}

func TestBuiltinFloatInfNan(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import math
inf = float("inf")
neg_inf = float("-inf")
nan = float("nan")
is_inf = math.isinf(inf)
is_neg_inf = math.isinf(neg_inf)
is_nan = math.isnan(nan)
`)
	assert.True(t, vm.GetGlobal("is_inf").(*runtime.PyBool).Value)
	assert.True(t, vm.GetGlobal("is_neg_inf").(*runtime.PyBool).Value)
	assert.True(t, vm.GetGlobal("is_nan").(*runtime.PyBool).Value)
}

func TestBuiltinFloatErrors(t *testing.T) {
	runCodeExpectError(t, `float("abc")`, "could not convert")
}

// =====================================
// bool() / str() conversions
// =====================================

func TestBuiltinBoolConversions(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   bool
	}{
		{"empty string", `result = bool("")`, false},
		{"non-empty string", `result = bool("a")`, true},
		{"zero", `result = bool(0)`, false},
		{"non-zero", `result = bool(1)`, true},
		{"empty list", `result = bool([])`, false},
		{"non-empty list", `result = bool([1])`, true},
		{"None", `result = bool(None)`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := runCode(t, tt.source)
			result := vm.GetGlobal("result").(*runtime.PyBool)
			assert.Equal(t, tt.want, result.Value)
		})
	}
}

func TestBuiltinStrConversions(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{"from int", `result = str(42)`, "42"},
		{"from float", `result = str(3.14)`, "3.14"},
		{"from bool", `result = str(True)`, "True"},
		{"from None", `result = str(None)`, "None"},
		{"from list", `result = str([1, 2, 3])`, "[1, 2, 3]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := runCode(t, tt.source)
			result := vm.GetGlobal("result").(*runtime.PyString)
			assert.Equal(t, tt.want, result.Value)
		})
	}
}

// =====================================
// isinstance() / issubclass()
// =====================================

func TestBuiltinIsinstance(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   bool
	}{
		{"int", `result = isinstance(42, int)`, true},
		{"str", `result = isinstance("hello", str)`, true},
		{"not match", `result = isinstance(42, str)`, false},
		{"bool is int", `result = isinstance(True, int)`, true},
		{"tuple of types", `result = isinstance(42, (str, int))`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := runCode(t, tt.source)
			result := vm.GetGlobal("result").(*runtime.PyBool)
			assert.Equal(t, tt.want, result.Value)
		})
	}
}

func TestBuiltinIssubclass(t *testing.T) {
	vm := runCode(t, `
class Animal:
    pass
class Dog(Animal):
    pass
result1 = issubclass(Dog, Animal)
result2 = issubclass(Animal, Dog)
result3 = issubclass(Dog, Dog)
`)
	assert.True(t, vm.GetGlobal("result1").(*runtime.PyBool).Value)
	assert.False(t, vm.GetGlobal("result2").(*runtime.PyBool).Value)
	assert.True(t, vm.GetGlobal("result3").(*runtime.PyBool).Value)
}

// =====================================
// hex() / oct() / bin()
// =====================================

func TestBuiltinHexOctBin(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{"hex 255", `result = hex(255)`, "0xff"},
		{"hex 0", `result = hex(0)`, "0x0"},
		{"hex neg", `result = hex(-42)`, "-0x2a"},
		{"oct 8", `result = oct(8)`, "0o10"},
		{"oct 0", `result = oct(0)`, "0o0"},
		{"oct neg", `result = oct(-8)`, "-0o10"},
		{"bin 10", `result = bin(10)`, "0b1010"},
		{"bin 0", `result = bin(0)`, "0b0"},
		{"bin neg", `result = bin(-10)`, "-0b1010"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := runCode(t, tt.source)
			result := vm.GetGlobal("result").(*runtime.PyString)
			assert.Equal(t, tt.want, result.Value)
		})
	}
}

// =====================================
// chr() / ord()
// =====================================

func TestBuiltinChrOrd(t *testing.T) {
	tests := []struct {
		name   string
		source string
		wantS  string
		wantI  int64
	}{
		{"chr 65", `result = chr(65)`, "A", 0},
		{"chr 48", `result = chr(48)`, "0", 0},
		{"ord A", `result = ord("A")`, "", 65},
		{"ord 0", `result = ord("0")`, "", 48},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := runCode(t, tt.source)
			if tt.wantS != "" {
				result := vm.GetGlobal("result").(*runtime.PyString)
				assert.Equal(t, tt.wantS, result.Value)
			} else {
				result := vm.GetGlobal("result").(*runtime.PyInt)
				assert.Equal(t, tt.wantI, result.Value)
			}
		})
	}
}

// =====================================
// range()
// =====================================

func TestBuiltinRange(t *testing.T) {
	vm := runCode(t, `
result = list(range(5))
`)
	result := vm.GetGlobal("result").(*runtime.PyList)
	assert.Equal(t, 5, len(result.Items))
	for i := 0; i < 5; i++ {
		assert.Equal(t, int64(i), result.Items[i].(*runtime.PyInt).Value)
	}
}

func TestBuiltinRangeNegativeStep(t *testing.T) {
	vm := runCode(t, `result = list(range(10, 0, -2))`)
	result := vm.GetGlobal("result").(*runtime.PyList)
	expected := []int64{10, 8, 6, 4, 2}
	require.Equal(t, len(expected), len(result.Items))
	for i, want := range expected {
		assert.Equal(t, want, result.Items[i].(*runtime.PyInt).Value)
	}
}

func TestBuiltinRangeZeroStep(t *testing.T) {
	runCodeExpectError(t, `range(0, 10, 0)`, "must not be zero")
}

// =====================================
// enumerate() / zip() / map() / filter()
// =====================================

func TestBuiltinEnumerate(t *testing.T) {
	vm := runCode(t, `
result = list(enumerate(["a", "b", "c"]))
first = result[0]
idx = first[0]
val = first[1]
`)
	idx := vm.GetGlobal("idx").(*runtime.PyInt)
	val := vm.GetGlobal("val").(*runtime.PyString)
	assert.Equal(t, int64(0), idx.Value)
	assert.Equal(t, "a", val.Value)
}

func TestBuiltinZip(t *testing.T) {
	vm := runCode(t, `
result = list(zip([1, 2, 3], ["a", "b", "c"]))
length = len(result)
first = result[0]
`)
	length := vm.GetGlobal("length").(*runtime.PyInt)
	assert.Equal(t, int64(3), length.Value)
	first := vm.GetGlobal("first").(*runtime.PyTuple)
	assert.Equal(t, int64(1), first.Items[0].(*runtime.PyInt).Value)
	assert.Equal(t, "a", first.Items[1].(*runtime.PyString).Value)
}

func TestBuiltinMap(t *testing.T) {
	vm := runCode(t, `
result = list(map(str, [1, 2, 3]))
`)
	result := vm.GetGlobal("result").(*runtime.PyList)
	assert.Equal(t, 3, len(result.Items))
	assert.Equal(t, "1", result.Items[0].(*runtime.PyString).Value)
}

func TestBuiltinFilter(t *testing.T) {
	vm := runCode(t, `
def is_even(x):
    return x % 2 == 0
result = list(filter(is_even, [1, 2, 3, 4, 5, 6]))
`)
	result := vm.GetGlobal("result").(*runtime.PyList)
	assert.Equal(t, 3, len(result.Items))
	assert.Equal(t, int64(2), result.Items[0].(*runtime.PyInt).Value)
	assert.Equal(t, int64(4), result.Items[1].(*runtime.PyInt).Value)
	assert.Equal(t, int64(6), result.Items[2].(*runtime.PyInt).Value)
}

// =====================================
// round()
// =====================================

func TestBuiltinRound(t *testing.T) {
	tests := []struct {
		name   string
		source string
		wantI  int64
		isInt  bool
	}{
		{"round int", `result = round(3)`, 3, true},
		{"round float", `result = round(3.7)`, 4, true},
		{"round floor", `result = round(3.2)`, 3, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := runCode(t, tt.source)
			result := vm.GetGlobal("result").(*runtime.PyInt)
			assert.Equal(t, tt.wantI, result.Value)
		})
	}
}

func TestBuiltinRoundNdigits(t *testing.T) {
	vm := runCode(t, `result = round(3.14159, 2)`)
	result := vm.GetGlobal("result").(*runtime.PyFloat)
	assert.InDelta(t, 3.14, result.Value, 0.005)
}

// =====================================
// divmod()
// =====================================

func TestBuiltinDivmod(t *testing.T) {
	vm := runCode(t, `
result = divmod(17, 5)
q = result[0]
r = result[1]
`)
	q := vm.GetGlobal("q").(*runtime.PyInt)
	r := vm.GetGlobal("r").(*runtime.PyInt)
	assert.Equal(t, int64(3), q.Value)
	assert.Equal(t, int64(2), r.Value)
}

func TestBuiltinDivmodFloat(t *testing.T) {
	vm := runCode(t, `
result = divmod(7.5, 2.0)
q = result[0]
r = result[1]
`)
	q := vm.GetGlobal("q").(*runtime.PyFloat)
	r := vm.GetGlobal("r").(*runtime.PyFloat)
	assert.InDelta(t, 3.0, q.Value, 1e-10)
	assert.InDelta(t, 1.5, r.Value, 1e-10)
}

func TestBuiltinDivmodZero(t *testing.T) {
	runCodeExpectError(t, `divmod(10, 0)`, "ZeroDivisionError")
}

// =====================================
// pow()
// =====================================

func TestBuiltinPow(t *testing.T) {
	vm := runCode(t, `result = pow(2, 10)`)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(1024), result.Value)
}

func TestBuiltinPowModular(t *testing.T) {
	vm := runCode(t, `result = pow(2, 10, 100)`)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(24), result.Value) // 1024 % 100 = 24
}

// =====================================
// all() / any()
// =====================================

func TestBuiltinAllAny(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   bool
	}{
		{"all true", `result = all([True, True, True])`, true},
		{"all with false", `result = all([True, False, True])`, false},
		{"all empty", `result = all([])`, true},
		{"any true", `result = any([False, True, False])`, true},
		{"any all false", `result = any([False, False, False])`, false},
		{"any empty", `result = any([])`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := runCode(t, tt.source)
			result := vm.GetGlobal("result").(*runtime.PyBool)
			assert.Equal(t, tt.want, result.Value)
		})
	}
}

// =====================================
// dir()
// =====================================

func TestBuiltinDir(t *testing.T) {
	vm := runCode(t, `
class Foo:
    x = 1
    def bar(self):
        pass
result = dir(Foo)
has_x = "x" in result
has_bar = "bar" in result
`)
	assert.True(t, vm.GetGlobal("has_x").(*runtime.PyBool).Value)
	assert.True(t, vm.GetGlobal("has_bar").(*runtime.PyBool).Value)
}

// =====================================
// callable()
// =====================================

func TestBuiltinCallable(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   bool
	}{
		{"function", `
def foo():
    pass
result = callable(foo)`, true},
		{"class", `
class Foo:
    pass
result = callable(Foo)`, true},
		{"int", `result = callable(42)`, false},
		{"string", `result = callable("hello")`, false},
		{"lambda", `result = callable(lambda x: x)`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := runCode(t, tt.source)
			result := vm.GetGlobal("result").(*runtime.PyBool)
			assert.Equal(t, tt.want, result.Value)
		})
	}
}

func TestBuiltinCallableWithCall(t *testing.T) {
	vm := runCode(t, `
class Foo:
    def __call__(self):
        return 42
f = Foo()
result = callable(f)
`)
	assert.True(t, vm.GetGlobal("result").(*runtime.PyBool).Value)
}

// =====================================
// reversed()
// =====================================

func TestBuiltinReversed(t *testing.T) {
	vm := runCode(t, `result = list(reversed([1, 2, 3]))`)
	result := vm.GetGlobal("result").(*runtime.PyList)
	assert.Equal(t, 3, len(result.Items))
	assert.Equal(t, int64(3), result.Items[0].(*runtime.PyInt).Value)
	assert.Equal(t, int64(2), result.Items[1].(*runtime.PyInt).Value)
	assert.Equal(t, int64(1), result.Items[2].(*runtime.PyInt).Value)
}

// =====================================
// type()
// =====================================

func TestBuiltinType(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{"int", `result = type(42).__name__`, "int"},
		{"str", `result = type("hello").__name__`, "str"},
		{"list", `result = type([]).__name__`, "list"},
		{"bool", `result = type(True).__name__`, "bool"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := runCode(t, tt.source)
			result := vm.GetGlobal("result").(*runtime.PyString)
			assert.Equal(t, tt.want, result.Value)
		})
	}
}

// =====================================
// repr()
// =====================================

func TestBuiltinRepr(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{"string", `result = repr("hello")`, "'hello'"},
		{"int", `result = repr(42)`, "42"},
		{"list", `result = repr([1, 2, 3])`, "[1, 2, 3]"},
		{"None", `result = repr(None)`, "None"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := runCode(t, tt.source)
			result := vm.GetGlobal("result").(*runtime.PyString)
			assert.Equal(t, tt.want, result.Value)
		})
	}
}

// =====================================
// hash()
// =====================================

func TestBuiltinHash(t *testing.T) {
	vm := runCode(t, `
h1 = hash(1)
h2 = hash(True)
same = h1 == h2
`)
	assert.True(t, vm.GetGlobal("same").(*runtime.PyBool).Value)
}

// =====================================
// getattr() / setattr() / hasattr()
// =====================================

func TestBuiltinGetSetHasattr(t *testing.T) {
	vm := runCode(t, `
class Obj:
    x = 10
o = Obj()
has_x = hasattr(o, "x")
val = getattr(o, "x")
has_y = hasattr(o, "y")
default_y = getattr(o, "y", 42)
setattr(o, "z", 99)
z_val = o.z
`)
	assert.True(t, vm.GetGlobal("has_x").(*runtime.PyBool).Value)
	assert.Equal(t, int64(10), vm.GetGlobal("val").(*runtime.PyInt).Value)
	assert.False(t, vm.GetGlobal("has_y").(*runtime.PyBool).Value)
	assert.Equal(t, int64(42), vm.GetGlobal("default_y").(*runtime.PyInt).Value)
	assert.Equal(t, int64(99), vm.GetGlobal("z_val").(*runtime.PyInt).Value)
}

// =====================================
// delattr()
// =====================================

func TestBuiltinDelattr(t *testing.T) {
	vm := runCode(t, `
class Obj:
    pass
o = Obj()
o.x = 42
has_before = hasattr(o, "x")
delattr(o, "x")
has_after = hasattr(o, "x")
`)
	assert.True(t, vm.GetGlobal("has_before").(*runtime.PyBool).Value)
	assert.False(t, vm.GetGlobal("has_after").(*runtime.PyBool).Value)
}
