package main

import (
	"fmt"

	"github.com/ATSOTECK/oink/internal/compiler"
	"github.com/ATSOTECK/oink/internal/utils"
)

func main() {
	source := `def factorial(n):
    if n <= 1:
        return 1
    return n * factorial(n - 1)

result = factorial(5)
print(f"5! = {result}")
`

	fmt.Println("=== Python 3.14 Demo ===")
	fmt.Println()
	fmt.Println("Source:")
	fmt.Println(source)

	// Parse the source
	parser := compiler.NewParser(source)
	module, parseErrors := parser.Parse()

	if len(parseErrors) > 0 {
		fmt.Println("Parse Errors:")
		for _, err := range parseErrors {
			fmt.Println(" ", err)
		}
		return
	}

	fmt.Println("AST Structure:")
	fmt.Println("--------------")
	utils.PrintAST(module, 0)
}
