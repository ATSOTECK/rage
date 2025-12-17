package main

import (
	"fmt"

	"github.com/ATSOTECK/oink/internal/compiler"
)

func main() {
	source := `def factorial(n):
    """Calculate the factorial of n."""
    if n <= 1:
        return 1
    return n * factorial(n - 1)

result = factorial(5)
print(f"5! = {result}")
`

	fmt.Println("=== Oink Python 3.14 Lexer Demo ===")
	fmt.Println()
	fmt.Println("Source:")
	fmt.Println(source)
	fmt.Println("Tokens:")
	fmt.Println("--------")

	lexer := compiler.NewLexer(source)
	tokens, errors := lexer.Tokenize()

	for _, tok := range tokens {
		fmt.Println(tok)
	}

	if len(errors) > 0 {
		fmt.Println()
		fmt.Println("Errors:")
		for _, err := range errors {
			fmt.Println(err)
		}
	}
}
