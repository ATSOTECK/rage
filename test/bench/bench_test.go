package bench

import (
	"testing"

	"github.com/ATSOTECK/rage/internal/compiler"
	"github.com/ATSOTECK/rage/internal/runtime"
	"github.com/ATSOTECK/rage/pkg/rage"
)

// Python source snippets for benchmarks.
var (
	srcArithmetic = `
x = 0
i = 0
while i < 1000:
    x = x + i * 2 - i // 3
    i = i + 1
`

	srcStringConcat = `
s = ""
i = 0
while i < 100:
    s = s + "hello"
    i = i + 1
`

	srcListComprehension = `
result = [x * x for x in range(500)]
`

	srcFunctionCalls = `
def add(a, b):
    return a + b

x = 0
i = 0
while i < 1000:
    x = add(x, i)
    i = i + 1
`

	srcClassInstantiation = `
class Point:
    def __init__(self, x, y):
        self.x = x
        self.y = y
    def magnitude(self):
        return (self.x ** 2 + self.y ** 2) ** 0.5

i = 0
while i < 500:
    p = Point(3, 4)
    m = p.magnitude()
    i = i + 1
`

	srcRecursion = `
def fib(n):
    if n <= 1:
        return n
    return fib(n - 1) + fib(n - 2)

result = fib(20)
`

	srcCompileOnly = `
def example(x, y, z):
    if x > 0:
        for i in range(y):
            z = z + i
    elif x < 0:
        while z > 0:
            z = z - 1
    else:
        try:
            z = x / y
        except ZeroDivisionError:
            z = 0
    return z

class Foo:
    def __init__(self, val):
        self.val = val
    def get(self):
        return self.val
`
)

func BenchmarkCompile(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		code, errs := compiler.CompileSource(srcCompileOnly, "bench")
		if code == nil || len(errs) > 0 {
			b.Fatal("compilation failed")
		}
	}
}

func BenchmarkExecute(b *testing.B) {
	code, errs := compiler.CompileSource(srcArithmetic, "bench")
	if code == nil || len(errs) > 0 {
		b.Fatal("compilation failed")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vm := runtime.NewVM()
		_, err := vm.Execute(code)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCompileAndExecute(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := rage.Run(srcArithmetic)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkArithmetic(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := rage.Run(srcArithmetic)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStringConcat(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := rage.Run(srcStringConcat)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkListComprehension(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := rage.Run(srcListComprehension)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFunctionCalls(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := rage.Run(srcFunctionCalls)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkClassInstantiation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := rage.Run(srcClassInstantiation)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRecursion(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := rage.Run(srcRecursion)
		if err != nil {
			b.Fatal(err)
		}
	}
}
