package test

import (
	"math"
	"testing"

	"github.com/ATSOTECK/rage/internal/runtime"
	"github.com/stretchr/testify/assert"
)

// =====================================
// Math Module
// =====================================

func TestMathTrigFunctions(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import math
s = math.sin(0)
c = math.cos(0)
t_ = math.tan(0)
`)
	assert.InDelta(t, 0.0, vm.GetGlobal("s").(*runtime.PyFloat).Value, 1e-10)
	assert.InDelta(t, 1.0, vm.GetGlobal("c").(*runtime.PyFloat).Value, 1e-10)
	assert.InDelta(t, 0.0, vm.GetGlobal("t_").(*runtime.PyFloat).Value, 1e-10)
}

func TestMathTrigInverse(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import math
a = math.asin(1.0)
b = math.acos(1.0)
c = math.atan(1.0)
d = math.atan2(1.0, 1.0)
`)
	assert.InDelta(t, math.Pi/2, vm.GetGlobal("a").(*runtime.PyFloat).Value, 1e-10)
	assert.InDelta(t, 0.0, vm.GetGlobal("b").(*runtime.PyFloat).Value, 1e-10)
	assert.InDelta(t, math.Pi/4, vm.GetGlobal("c").(*runtime.PyFloat).Value, 1e-10)
	assert.InDelta(t, math.Pi/4, vm.GetGlobal("d").(*runtime.PyFloat).Value, 1e-10)
}

func TestMathCeilFloorTrunc(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import math
c = math.ceil(3.2)
f = math.floor(3.8)
tr = math.trunc(3.9)
cn = math.ceil(-3.2)
fn = math.floor(-3.8)
`)
	assert.Equal(t, int64(4), vm.GetGlobal("c").(*runtime.PyInt).Value)
	assert.Equal(t, int64(3), vm.GetGlobal("f").(*runtime.PyInt).Value)
	assert.Equal(t, int64(3), vm.GetGlobal("tr").(*runtime.PyInt).Value)
	assert.Equal(t, int64(-3), vm.GetGlobal("cn").(*runtime.PyInt).Value)
	assert.Equal(t, int64(-4), vm.GetGlobal("fn").(*runtime.PyInt).Value)
}

func TestMathConstants(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import math
p = math.pi
e = math.e
tau = math.tau
inf = math.inf
`)
	assert.InDelta(t, math.Pi, vm.GetGlobal("p").(*runtime.PyFloat).Value, 1e-10)
	assert.InDelta(t, math.E, vm.GetGlobal("e").(*runtime.PyFloat).Value, 1e-10)
	assert.InDelta(t, 2*math.Pi, vm.GetGlobal("tau").(*runtime.PyFloat).Value, 1e-10)
	assert.True(t, math.IsInf(vm.GetGlobal("inf").(*runtime.PyFloat).Value, 1))
}

func TestMathLogExp(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import math
lg = math.log(math.e)
lg10 = math.log10(100)
lg2 = math.log2(8)
ex = math.exp(0)
ex1 = math.exp(1)
`)
	assert.InDelta(t, 1.0, vm.GetGlobal("lg").(*runtime.PyFloat).Value, 1e-10)
	assert.InDelta(t, 2.0, vm.GetGlobal("lg10").(*runtime.PyFloat).Value, 1e-10)
	assert.InDelta(t, 3.0, vm.GetGlobal("lg2").(*runtime.PyFloat).Value, 1e-10)
	assert.InDelta(t, 1.0, vm.GetGlobal("ex").(*runtime.PyFloat).Value, 1e-10)
	assert.InDelta(t, math.E, vm.GetGlobal("ex1").(*runtime.PyFloat).Value, 1e-10)
}

func TestMathSqrt(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import math
r = math.sqrt(16)
`)
	assert.InDelta(t, 4.0, vm.GetGlobal("r").(*runtime.PyFloat).Value, 1e-10)
}

func TestMathFabs(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import math
r = math.fabs(-3.14)
`)
	assert.InDelta(t, 3.14, vm.GetGlobal("r").(*runtime.PyFloat).Value, 1e-10)
}

func TestMathIsinfIsnan(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import math
ii = math.isinf(math.inf)
in_ = math.isnan(float("nan"))
fi = math.isinf(42.0)
fn = math.isnan(42.0)
`)
	assert.True(t, vm.GetGlobal("ii").(*runtime.PyBool).Value)
	assert.True(t, vm.GetGlobal("in_").(*runtime.PyBool).Value)
	assert.False(t, vm.GetGlobal("fi").(*runtime.PyBool).Value)
	assert.False(t, vm.GetGlobal("fn").(*runtime.PyBool).Value)
}

func TestMathGcd(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import math
r = math.gcd(12, 8)
`)
	assert.Equal(t, int64(4), vm.GetGlobal("r").(*runtime.PyInt).Value)
}

func TestMathFactorial(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import math
r = math.factorial(5)
`)
	assert.Equal(t, int64(120), vm.GetGlobal("r").(*runtime.PyInt).Value)
}

func TestMathPow(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import math
r = math.pow(2, 10)
`)
	assert.InDelta(t, 1024.0, vm.GetGlobal("r").(*runtime.PyFloat).Value, 1e-10)
}

// =====================================
// JSON Module
// =====================================

func TestJSONDumps(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import json
r1 = json.dumps({"a": 1, "b": 2})
r2 = json.dumps([1, 2, 3])
r3 = json.dumps("hello")
r4 = json.dumps(42)
r5 = json.dumps(True)
r6 = json.dumps(None)
`)
	// Dict ordering may vary, just check it's valid JSON-ish
	r1 := vm.GetGlobal("r1").(*runtime.PyString).Value
	assert.Contains(t, r1, `"a"`)
	assert.Contains(t, r1, `1`)

	r2str := vm.GetGlobal("r2").(*runtime.PyString).Value
	assert.True(t, r2str == "[1, 2, 3]" || r2str == "[1,2,3]")
	assert.Equal(t, `"hello"`, vm.GetGlobal("r3").(*runtime.PyString).Value)
	assert.Equal(t, "42", vm.GetGlobal("r4").(*runtime.PyString).Value)
	assert.Equal(t, "true", vm.GetGlobal("r5").(*runtime.PyString).Value)
	assert.Equal(t, "null", vm.GetGlobal("r6").(*runtime.PyString).Value)
}

func TestJSONLoads(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import json
r1 = json.loads('{"a": 1}')
val = r1["a"]
r2 = json.loads('[1, 2, 3]')
r3 = json.loads('"hello"')
r4 = json.loads('42')
r5 = json.loads('true')
r6 = json.loads('null')
`)
	assert.Equal(t, int64(1), vm.GetGlobal("val").(*runtime.PyInt).Value)
	r2 := vm.GetGlobal("r2").(*runtime.PyList)
	assert.Equal(t, 3, len(r2.Items))
	assert.Equal(t, "hello", vm.GetGlobal("r3").(*runtime.PyString).Value)
	assert.Equal(t, int64(42), vm.GetGlobal("r4").(*runtime.PyInt).Value)
	assert.True(t, vm.GetGlobal("r5").(*runtime.PyBool).Value)
	assert.Equal(t, runtime.None, vm.GetGlobal("r6"))
}

func TestJSONErrors(t *testing.T) {
	runCodeExpectError(t, `
import json
json.loads("not valid json {{{")
`, "JSON")
}

// =====================================
// Itertools Module
// =====================================

func TestItertoolsChain(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import itertools
result = list(itertools.chain([1, 2], [3, 4], [5]))
length = len(result)
`)
	assert.Equal(t, int64(5), vm.GetGlobal("length").(*runtime.PyInt).Value)
}

func TestItertoolsProduct(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import itertools
result = list(itertools.product([1, 2], ["a", "b"]))
length = len(result)
`)
	assert.Equal(t, int64(4), vm.GetGlobal("length").(*runtime.PyInt).Value)
}

func TestItertoolsPermutations(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import itertools
result = list(itertools.permutations([1, 2, 3], 2))
length = len(result)
`)
	assert.Equal(t, int64(6), vm.GetGlobal("length").(*runtime.PyInt).Value)
}

func TestItertoolsCombinations(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import itertools
result = list(itertools.combinations([1, 2, 3, 4], 2))
length = len(result)
`)
	assert.Equal(t, int64(6), vm.GetGlobal("length").(*runtime.PyInt).Value)
}

func TestItertoolsRepeat(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import itertools
result = list(itertools.repeat(42, 5))
length = len(result)
first = result[0]
`)
	assert.Equal(t, int64(5), vm.GetGlobal("length").(*runtime.PyInt).Value)
	assert.Equal(t, int64(42), vm.GetGlobal("first").(*runtime.PyInt).Value)
}

func TestItertoolsCount(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import itertools
c = itertools.count(10, 2)
r1 = next(c)
r2 = next(c)
r3 = next(c)
`)
	assert.Equal(t, int64(10), vm.GetGlobal("r1").(*runtime.PyInt).Value)
	assert.Equal(t, int64(12), vm.GetGlobal("r2").(*runtime.PyInt).Value)
	assert.Equal(t, int64(14), vm.GetGlobal("r3").(*runtime.PyInt).Value)
}

func TestItertoolsCycle(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import itertools
c = itertools.cycle([1, 2, 3])
results = []
for _ in range(7):
    results.append(next(c))
length = len(results)
last = results[6]
`)
	assert.Equal(t, int64(7), vm.GetGlobal("length").(*runtime.PyInt).Value)
	assert.Equal(t, int64(1), vm.GetGlobal("last").(*runtime.PyInt).Value)
}

// =====================================
// Functools Module
// =====================================

func TestFunctoolsReduce(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import functools
def add(a, b):
    return a + b
result = functools.reduce(add, [1, 2, 3, 4, 5])
`)
	assert.Equal(t, int64(15), vm.GetGlobal("result").(*runtime.PyInt).Value)
}

func TestFunctoolsReduceInitial(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import functools
def add(a, b):
    return a + b
result = functools.reduce(add, [1, 2, 3], 10)
`)
	assert.Equal(t, int64(16), vm.GetGlobal("result").(*runtime.PyInt).Value)
}

func TestFunctoolsPartial(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import functools
def add(a, b):
    return a + b
add5 = functools.partial(add, 5)
result = add5(3)
`)
	assert.Equal(t, int64(8), vm.GetGlobal("result").(*runtime.PyInt).Value)
}

// =====================================
// Datetime Module
// =====================================

func TestDatetimeImport(t *testing.T) {
	// Verify datetime module can be imported
	vm := runCodeWithStdlib(t, `
import datetime
has_date = hasattr(datetime, "date")
has_timedelta = hasattr(datetime, "timedelta")
`)
	assert.True(t, vm.GetGlobal("has_date").(*runtime.PyBool).Value)
	assert.True(t, vm.GetGlobal("has_timedelta").(*runtime.PyBool).Value)
}

func TestDatetimeTimedelta(t *testing.T) {
	// timedelta(days, seconds, ...) - use positional args
	vm := runCodeWithStdlib(t, `
import datetime
td = datetime.timedelta(2)
d = td.days()
total = td.total_seconds()
`)
	assert.Equal(t, int64(2), vm.GetGlobal("d").(*runtime.PyInt).Value)
	assert.InDelta(t, 172800.0, vm.GetGlobal("total").(*runtime.PyFloat).Value, 1.0)
}

// =====================================
// CSV Module
// =====================================

func TestCSVParseRow(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import csv
row = csv.parse_row("a,b,c")
row_len = len(row)
first = row[0]
last = row[2]
`)
	assert.Equal(t, int64(3), vm.GetGlobal("row_len").(*runtime.PyInt).Value)
	assert.Equal(t, "a", vm.GetGlobal("first").(*runtime.PyString).Value)
	assert.Equal(t, "c", vm.GetGlobal("last").(*runtime.PyString).Value)
}

func TestCSVFormatRow(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import csv
result = csv.format_row(["hello", "world", "test"])
has_content = len(result) > 0
has_comma = "," in result
`)
	assert.True(t, vm.GetGlobal("has_content").(*runtime.PyBool).Value)
	assert.True(t, vm.GetGlobal("has_comma").(*runtime.PyBool).Value)
}

// =====================================
// Base64 Module
// =====================================

func TestBase64EncodeDecode(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import base64
data = b"Hello"
encoded = base64.b64encode(data)
decoded = base64.b64decode(encoded)
enc_len = len(encoded)
dec_len = len(decoded)
`)
	assert.Equal(t, int64(8), vm.GetGlobal("enc_len").(*runtime.PyInt).Value)  // "SGVsbG8=" is 8 chars
	assert.Equal(t, int64(5), vm.GetGlobal("dec_len").(*runtime.PyInt).Value)  // "Hello" is 5 bytes
}

func TestBase64UrlSafe(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import base64
data = b"Hello"
encoded = base64.urlsafe_b64encode(data)
decoded = base64.urlsafe_b64decode(encoded)
enc_len = len(encoded)
dec_len = len(decoded)
`)
	assert.Equal(t, int64(8), vm.GetGlobal("enc_len").(*runtime.PyInt).Value)
	assert.Equal(t, int64(5), vm.GetGlobal("dec_len").(*runtime.PyInt).Value)
}

// =====================================
// Copy Module
// =====================================

func TestCopyCopy(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import copy
original = [1, 2, [3, 4]]
shallow = copy.copy(original)
same_len = len(shallow) == len(original)
not_same = shallow is not original
`)
	assert.True(t, vm.GetGlobal("same_len").(*runtime.PyBool).Value)
	assert.True(t, vm.GetGlobal("not_same").(*runtime.PyBool).Value)
}

func TestCopyDeepcopy(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import copy
original = [1, 2, [3, 4]]
deep = copy.deepcopy(original)
same_len = len(deep) == len(original)
not_same = deep is not original
inner_not_same = deep[2] is not original[2]
`)
	assert.True(t, vm.GetGlobal("same_len").(*runtime.PyBool).Value)
	assert.True(t, vm.GetGlobal("not_same").(*runtime.PyBool).Value)
	assert.True(t, vm.GetGlobal("inner_not_same").(*runtime.PyBool).Value)
}

// =====================================
// Operator Module
// =====================================

func TestOperatorLengthHint(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import operator
result = operator.length_hint([1, 2, 3])
`)
	assert.Equal(t, int64(3), vm.GetGlobal("result").(*runtime.PyInt).Value)
}

func TestOperatorIndex(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import operator
result = operator.index(42)
`)
	assert.Equal(t, int64(42), vm.GetGlobal("result").(*runtime.PyInt).Value)
}

// =====================================
// ABC Module
// =====================================

func TestABCAbstractmethod(t *testing.T) {
	// Can't instantiate class with abstract methods
	runCodeExpectError(t, `
from abc import ABC, abstractmethod
class Animal(ABC):
    @abstractmethod
    def speak(self):
        pass
a = Animal()
`, "abstract")
}

func TestABCConcreteSubclass(t *testing.T) {
	vm := runCodeWithStdlib(t, `
from abc import ABC, abstractmethod
class Animal(ABC):
    @abstractmethod
    def speak(self):
        pass
class Dog(Animal):
    def speak(self):
        return "woof"
d = Dog()
result = d.speak()
`)
	assert.Equal(t, "woof", vm.GetGlobal("result").(*runtime.PyString).Value)
}

func TestABCRegister(t *testing.T) {
	vm := runCodeWithStdlib(t, `
from abc import ABC
class MyABC(ABC):
    pass
class Regular:
    pass
MyABC.register(Regular)
r = Regular()
result = isinstance(r, MyABC)
`)
	assert.True(t, vm.GetGlobal("result").(*runtime.PyBool).Value)
}

// =====================================
// Math module: additional functions
// =====================================

func TestMathCopysign(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import math
r1 = math.copysign(5, -1)
r2 = math.copysign(-5, 1)
`)
	assert.InDelta(t, -5.0, vm.GetGlobal("r1").(*runtime.PyFloat).Value, 1e-10)
	assert.InDelta(t, 5.0, vm.GetGlobal("r2").(*runtime.PyFloat).Value, 1e-10)
}

func TestMathHyperbolic(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import math
s = math.sinh(0)
c = math.cosh(0)
t_ = math.tanh(0)
`)
	assert.InDelta(t, 0.0, vm.GetGlobal("s").(*runtime.PyFloat).Value, 1e-10)
	assert.InDelta(t, 1.0, vm.GetGlobal("c").(*runtime.PyFloat).Value, 1e-10)
	assert.InDelta(t, 0.0, vm.GetGlobal("t_").(*runtime.PyFloat).Value, 1e-10)
}

func TestMathDegRad(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import math
d = math.degrees(math.pi)
r = math.radians(180)
`)
	assert.InDelta(t, 180.0, vm.GetGlobal("d").(*runtime.PyFloat).Value, 1e-10)
	assert.InDelta(t, math.Pi, vm.GetGlobal("r").(*runtime.PyFloat).Value, 1e-10)
}

func TestMathIsfinite(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import math
r1 = math.isfinite(42.0)
r2 = math.isfinite(math.inf)
r3 = math.isfinite(float("nan"))
`)
	assert.True(t, vm.GetGlobal("r1").(*runtime.PyBool).Value)
	assert.False(t, vm.GetGlobal("r2").(*runtime.PyBool).Value)
	assert.False(t, vm.GetGlobal("r3").(*runtime.PyBool).Value)
}
