package main

import (
	"fmt"

	"github.com/ATSOTECK/oink/internal/compiler"
	"github.com/ATSOTECK/oink/internal/runtime"
)

func main() {
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
`

	fmt.Println("=== Python 3.14 Demo ===")
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
