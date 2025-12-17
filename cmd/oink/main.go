package main

import (
	"fmt"

	"github.com/ATSOTECK/oink/internal/compiler"
	"github.com/ATSOTECK/oink/internal/runtime"
	"github.com/ATSOTECK/oink/internal/stdlib"
)

func main() {
	// Initialize standard library modules
	stdlib.InitAllModules()

	source := `
def factorial(n):
    if n <= 1:
        return 1
    return n * factorial(n - 1)

result = factorial(5)
print("5! =", result)

# Test list operations
nums = [1, 2, 3, 4, 5]
total = sum(nums)
print("Sum of", nums, "=", total)

# Test for loop
squares = []
for x in [1, 2, 3, 4]:
    squares.append(x * x)
print("Squares:", squares)

# Test while loop
count = 0
while count < 3:
    print("Count:", count)
    count = count + 1

# Test conditionals
x = 10
if x > 5:
    print("x is greater than 5")
else:
    print("x is not greater than 5")

# Test module imports
import math
print("Pi =", math.pi)
print("sqrt(16) =", math.sqrt(16))

from math import sin, cos, pi
print("sin(0) =", sin(0))
print("cos(0) =", cos(0))

# Test random module
import random
random.seed(42)
print("Random number:", random.randint(1, 100))
print("Random choice:", random.choice(["apple", "banana", "cherry"]))

# Test string module
import string
print("Digits:", string.digits)
print("Capwords:", string.capwords("hello world"))

# Test sys module
import sys
print("Python version:", sys.version)
print("Platform:", sys.platform)
`

	fmt.Println("=== Oink (Python 3.14) Demo ===")
	fmt.Println()
	fmt.Println("Source:")
	fmt.Println(source)
	fmt.Println("----------------------------")
	fmt.Println("Output:")
	fmt.Println()

	// Compile and run the source
	code, errs := compiler.CompileSource(source, "<demo>")
	if len(errs) > 0 {
		fmt.Println("Errors:")
		for _, err := range errs {
			fmt.Println(" ", err)
		}
		return
	}

	// Execute
	vm := runtime.NewVM()
	_, err := vm.Execute(code)
	if err != nil {
		fmt.Println("Runtime error:", err)
		return
	}
}
