package compiler

import (
	"testing"
)

// Seed corpus entries for fuzzing.
var seeds = []string{
	// Valid Python
	`x = 1 + 2`,
	`print("hello world")`,
	`def foo(a, b): return a + b`,
	`class Foo:\n    def __init__(self): self.x = 1`,
	`[x * x for x in range(10)]`,
	`{k: v for k, v in items}`,
	`{x for x in range(10)}`,
	`(x for x in range(10))`,
	`if x > 0:\n    pass\nelif x < 0:\n    pass\nelse:\n    pass`,
	`for i in range(10):\n    break`,
	`while True:\n    continue`,
	`try:\n    x = 1\nexcept:\n    pass\nfinally:\n    pass`,
	`with open("f") as f:\n    pass`,
	`lambda x, y: x + y`,
	`a, *b, c = [1, 2, 3, 4, 5]`,
	`x = 1 if True else 0`,
	`import os`,
	`from os import path`,
	`raise ValueError("bad")`,
	`assert x == 1, "expected 1"`,
	`del x`,
	`global x`,
	`nonlocal x`,
	`yield x`,
	`yield from iterable`,
	`async def foo(): await bar()`,

	// Edge cases
	`""""""`,
	`'''triple'''`,
	`f"hello {name!r:.2f}"`,
	`b"\x00\xff"`,
	`0xFF + 0o77 + 0b1010`,
	`1_000_000`,
	`1.5e-10`,
	`1+2j`,

	// Unicode
	`变量 = 42`,
	`π = 3.14`,

	// Nested structures
	`[[[[[]]]]]`,
	`((((()))))`,
	`{{{1: {2: {3: 4}}}}}`,

	// Long expressions
	`x = 1 + 2 * 3 - 4 / 5 ** 6 % 7 & 8 | 9 ^ 10 << 11 >> 12`,

	// Decorators
	`@decorator\ndef foo(): pass`,
	`@decorator(arg)\nclass Foo: pass`,

	// Walrus operator
	`if (n := len(a)) > 10: print(n)`,

	// Match statement
	`match x:\n    case 1: pass\n    case _: pass`,

	// Empty/whitespace
	``,
	`   `,
	"\t\t\t",
	"\n\n\n",

	// Potentially confusing
	`'''`,
	`"""`,
	`(`,
	`)`,
	`[`,
	`{`,
	`def`,
	`if`,
	`:`,
	`\`,
}

func FuzzLexer(f *testing.F) {
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("lexer panicked on input %q: %v", input, r)
			}
		}()

		lexer := NewLexer(input)
		_, _ = lexer.Tokenize()
	})
}

func FuzzParser(f *testing.F) {
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("parser panicked on input %q: %v", input, r)
			}
		}()

		parser := NewParser(input)
		_, _ = parser.Parse()
	})
}

func FuzzCompileSource(f *testing.F) {
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("CompileSource panicked on input %q: %v", input, r)
			}
		}()

		_, _ = CompileSource(input, "fuzz")
	})
}
