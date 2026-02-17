package compiler

import (
	"testing"

	"github.com/ATSOTECK/rage/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParserLiterals(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected func(t *testing.T, expr model.Expr)
	}{
		{
			name:  "integer",
			input: "42",
			expected: func(t *testing.T, expr model.Expr) {
				lit, ok := expr.(*model.IntLit)
				require.True(t, ok, "expected IntLit")
				assert.Equal(t, "42", lit.Value)
			},
		},
		{
			name:  "float",
			input: "3.14",
			expected: func(t *testing.T, expr model.Expr) {
				lit, ok := expr.(*model.FloatLit)
				require.True(t, ok, "expected FloatLit")
				assert.Equal(t, "3.14", lit.Value)
			},
		},
		{
			name:  "string",
			input: `"hello"`,
			expected: func(t *testing.T, expr model.Expr) {
				lit, ok := expr.(*model.StringLit)
				require.True(t, ok, "expected StringLit")
				assert.Equal(t, "hello", lit.Value)
			},
		},
		{
			name:  "True",
			input: "True",
			expected: func(t *testing.T, expr model.Expr) {
				lit, ok := expr.(*model.BoolLit)
				require.True(t, ok, "expected BoolLit")
				assert.True(t, lit.Value)
			},
		},
		{
			name:  "False",
			input: "False",
			expected: func(t *testing.T, expr model.Expr) {
				lit, ok := expr.(*model.BoolLit)
				require.True(t, ok, "expected BoolLit")
				assert.False(t, lit.Value)
			},
		},
		{
			name:  "None",
			input: "None",
			expected: func(t *testing.T, expr model.Expr) {
				_, ok := expr.(*model.NoneLit)
				require.True(t, ok, "expected NoneLit")
			},
		},
		{
			name:  "identifier",
			input: "foo",
			expected: func(t *testing.T, expr model.Expr) {
				id, ok := expr.(*model.Identifier)
				require.True(t, ok, "expected Identifier")
				assert.Equal(t, "foo", id.Name)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parser := NewParser(test.input)
			module, errs := parser.Parse()

			require.Empty(t, errs, "unexpected parser errors")
			require.Len(t, module.Body, 1, "expected one statement")

			exprStmt, ok := module.Body[0].(*model.ExprStmt)
			require.True(t, ok, "expected expression statement")

			test.expected(t, exprStmt.Value)
		})
	}
}

func TestParserBinaryOperations(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		leftOp  string
		op      model.TokenKind
		rightOp string
	}{
		{"addition", "1 + 2", "1", model.TK_Plus, "2"},
		{"subtraction", "3 - 1", "3", model.TK_Minus, "1"},
		{"multiplication", "2 * 3", "2", model.TK_Star, "3"},
		{"division", "6 / 2", "6", model.TK_Slash, "2"},
		{"floor division", "7 // 2", "7", model.TK_DoubleSlash, "2"},
		{"modulo", "7 % 3", "7", model.TK_Percent, "3"},
		{"power", "2 ** 3", "2", model.TK_DoubleStar, "3"},
		{"bitwise and", "5 & 3", "5", model.TK_Ampersand, "3"},
		{"bitwise or", "5 | 3", "5", model.TK_Pipe, "3"},
		{"bitwise xor", "5 ^ 3", "5", model.TK_Caret, "3"},
		{"left shift", "1 << 2", "1", model.TK_LShift, "2"},
		{"right shift", "8 >> 2", "8", model.TK_RShift, "2"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parser := NewParser(test.input)
			module, errs := parser.Parse()

			require.Empty(t, errs, "unexpected parser errors")
			require.Len(t, module.Body, 1)

			exprStmt := module.Body[0].(*model.ExprStmt)
			binOp, ok := exprStmt.Value.(*model.BinaryOp)
			require.True(t, ok, "expected BinaryOp")

			assert.Equal(t, test.op, binOp.Op)

			left, ok := binOp.Left.(*model.IntLit)
			require.True(t, ok, "expected IntLit on left")
			assert.Equal(t, test.leftOp, left.Value)

			right, ok := binOp.Right.(*model.IntLit)
			require.True(t, ok, "expected IntLit on right")
			assert.Equal(t, test.rightOp, right.Value)
		})
	}
}

func TestParserPrecedence(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"mul before add", "1 + 2 * 3", "(1 + (2 * 3))"},
		{"parens override", "(1 + 2) * 3", "((1 + 2) * 3)"},
		{"power right assoc", "2 ** 3 ** 2", "(2 ** (3 ** 2))"},
		{"comparison chain", "1 < 2 < 3", "(1 < 2 < 3)"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parser := NewParser(test.input)
			module, errs := parser.Parse()

			require.Empty(t, errs, "unexpected parser errors")
			require.Len(t, module.Body, 1)
		})
	}
}

func TestParserUnaryOperations(t *testing.T) {
	tests := []struct {
		name  string
		input string
		op    model.TokenKind
	}{
		{"negative", "-5", model.TK_Minus},
		{"positive", "+5", model.TK_Plus},
		{"bitwise not", "~5", model.TK_Tilde},
		{"logical not", "not True", model.TK_Not},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parser := NewParser(test.input)
			module, errs := parser.Parse()

			require.Empty(t, errs, "unexpected parser errors")
			require.Len(t, module.Body, 1)

			exprStmt := module.Body[0].(*model.ExprStmt)
			unaryOp, ok := exprStmt.Value.(*model.UnaryOp)
			require.True(t, ok, "expected UnaryOp")

			assert.Equal(t, test.op, unaryOp.Op)
		})
	}
}

func TestParserFunctionCall(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		funcName string
		argCount int
		kwCount  int
	}{
		{"no args", "foo()", "foo", 0, 0},
		{"one arg", "foo(1)", "foo", 1, 0},
		{"multiple args", "foo(1, 2, 3)", "foo", 3, 0},
		{"keyword arg", "foo(x=1)", "foo", 0, 1},
		{"mixed args", "foo(1, x=2)", "foo", 1, 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parser := NewParser(test.input)
			module, errs := parser.Parse()

			require.Empty(t, errs, "unexpected parser errors")
			require.Len(t, module.Body, 1)

			exprStmt := module.Body[0].(*model.ExprStmt)
			call, ok := exprStmt.Value.(*model.Call)
			require.True(t, ok, "expected Call")

			fn, ok := call.Func.(*model.Identifier)
			require.True(t, ok, "expected Identifier as function")
			assert.Equal(t, test.funcName, fn.Name)
			assert.Len(t, call.Args, test.argCount)
			assert.Len(t, call.Keywords, test.kwCount)
		})
	}
}

func TestParserAttributeAccess(t *testing.T) {
	parser := NewParser("obj.attr")
	module, errs := parser.Parse()

	require.Empty(t, errs)
	require.Len(t, module.Body, 1)

	exprStmt := module.Body[0].(*model.ExprStmt)
	attr, ok := exprStmt.Value.(*model.Attribute)
	require.True(t, ok, "expected Attribute")

	obj, ok := attr.Value.(*model.Identifier)
	require.True(t, ok)
	assert.Equal(t, "obj", obj.Name)
	assert.Equal(t, "attr", attr.Attr.Name)
}

func TestParserSubscript(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"index", "arr[0]"},
		{"slice", "arr[1:3]"},
		{"slice with step", "arr[::2]"},
		{"slice start only", "arr[1:]"},
		{"slice end only", "arr[:3]"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parser := NewParser(test.input)
			module, errs := parser.Parse()

			require.Empty(t, errs, "unexpected parser errors")
			require.Len(t, module.Body, 1)

			exprStmt := module.Body[0].(*model.ExprStmt)
			_, ok := exprStmt.Value.(*model.Subscript)
			require.True(t, ok, "expected Subscript")
		})
	}
}

func TestParserList(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		elements int
	}{
		{"empty", "[]", 0},
		{"single", "[1]", 1},
		{"multiple", "[1, 2, 3]", 3},
		{"trailing comma", "[1, 2,]", 2},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parser := NewParser(test.input)
			module, errs := parser.Parse()

			require.Empty(t, errs, "unexpected parser errors")
			require.Len(t, module.Body, 1)

			exprStmt := module.Body[0].(*model.ExprStmt)
			list, ok := exprStmt.Value.(*model.List)
			require.True(t, ok, "expected List")
			assert.Len(t, list.Elts, test.elements)
		})
	}
}

func TestParserDict(t *testing.T) {
	tests := []struct {
		name  string
		input string
		pairs int
	}{
		{"empty", "{}", 0},
		{"single", `{"a": 1}`, 1},
		{"multiple", `{"a": 1, "b": 2}`, 2},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parser := NewParser(test.input)
			module, errs := parser.Parse()

			require.Empty(t, errs, "unexpected parser errors")
			require.Len(t, module.Body, 1)

			exprStmt := module.Body[0].(*model.ExprStmt)
			dict, ok := exprStmt.Value.(*model.Dict)
			require.True(t, ok, "expected Dict")
			assert.Len(t, dict.Keys, test.pairs)
			assert.Len(t, dict.Values, test.pairs)
		})
	}
}

func TestParserTuple(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		elements int
	}{
		{"empty", "()", 0},
		{"single trailing comma", "(1,)", 1},
		{"multiple", "(1, 2, 3)", 3},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parser := NewParser(test.input)
			module, errs := parser.Parse()

			require.Empty(t, errs, "unexpected parser errors")
			require.Len(t, module.Body, 1)

			exprStmt := module.Body[0].(*model.ExprStmt)
			tuple, ok := exprStmt.Value.(*model.Tuple)
			require.True(t, ok, "expected Tuple")
			assert.Len(t, tuple.Elts, test.elements)
		})
	}
}

func TestParserSet(t *testing.T) {
	parser := NewParser("{1, 2, 3}")
	module, errs := parser.Parse()

	require.Empty(t, errs)
	require.Len(t, module.Body, 1)

	exprStmt := module.Body[0].(*model.ExprStmt)
	set, ok := exprStmt.Value.(*model.Set)
	require.True(t, ok, "expected Set")
	assert.Len(t, set.Elts, 3)
}

func TestParserListComprehension(t *testing.T) {
	parser := NewParser("[x * 2 for x in range(10)]")
	module, errs := parser.Parse()

	require.Empty(t, errs)
	require.Len(t, module.Body, 1)

	exprStmt := module.Body[0].(*model.ExprStmt)
	listComp, ok := exprStmt.Value.(*model.ListComp)
	require.True(t, ok, "expected ListComp")
	assert.Len(t, listComp.Generators, 1)
}

func TestParserAssignment(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		targets int
	}{
		{"simple", "x = 1", 1},
		{"multiple targets", "x = y = 1", 2},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parser := NewParser(test.input)
			module, errs := parser.Parse()

			require.Empty(t, errs, "unexpected parser errors")
			require.Len(t, module.Body, 1)

			assign, ok := module.Body[0].(*model.Assign)
			require.True(t, ok, "expected Assign")
			assert.Len(t, assign.Targets, test.targets)
		})
	}
}

func TestParserAugmentedAssignment(t *testing.T) {
	tests := []struct {
		name  string
		input string
		op    model.TokenKind
	}{
		{"plus", "x += 1", model.TK_PlusAssign},
		{"minus", "x -= 1", model.TK_MinusAssign},
		{"mul", "x *= 2", model.TK_StarAssign},
		{"div", "x /= 2", model.TK_SlashAssign},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parser := NewParser(test.input)
			module, errs := parser.Parse()

			require.Empty(t, errs, "unexpected parser errors")
			require.Len(t, module.Body, 1)

			augAssign, ok := module.Body[0].(*model.AugAssign)
			require.True(t, ok, "expected AugAssign")
			assert.Equal(t, test.op, augAssign.Op)
		})
	}
}

func TestParserIfStatement(t *testing.T) {
	input := `if x > 0:
    return True
elif x < 0:
    return False
else:
    return None`

	parser := NewParser(input)
	module, errs := parser.Parse()

	require.Empty(t, errs)
	require.Len(t, module.Body, 1)

	ifStmt, ok := module.Body[0].(*model.If)
	require.True(t, ok, "expected If")
	assert.NotNil(t, ifStmt.Test)
	assert.Len(t, ifStmt.Body, 1)
	assert.Len(t, ifStmt.OrElse, 1) // elif becomes nested If
}

func TestParserWhileStatement(t *testing.T) {
	input := `while x > 0:
    x -= 1`

	parser := NewParser(input)
	module, errs := parser.Parse()

	require.Empty(t, errs)
	require.Len(t, module.Body, 1)

	whileStmt, ok := module.Body[0].(*model.While)
	require.True(t, ok, "expected While")
	assert.NotNil(t, whileStmt.Test)
	assert.Len(t, whileStmt.Body, 1)
}

func TestParserForStatement(t *testing.T) {
	input := `for i in range(10):
    print(i)`

	parser := NewParser(input)
	module, errs := parser.Parse()

	require.Empty(t, errs)
	require.Len(t, module.Body, 1)

	forStmt, ok := module.Body[0].(*model.For)
	require.True(t, ok, "expected For")
	assert.NotNil(t, forStmt.Target)
	assert.NotNil(t, forStmt.Iter)
	assert.Len(t, forStmt.Body, 1)
}

func TestParserFunctionDef(t *testing.T) {
	input := `def add(a, b):
    return a + b`

	parser := NewParser(input)
	module, errs := parser.Parse()

	require.Empty(t, errs)
	require.Len(t, module.Body, 1)

	funcDef, ok := module.Body[0].(*model.FunctionDef)
	require.True(t, ok, "expected FunctionDef")
	assert.Equal(t, "add", funcDef.Name.Name)
	assert.Len(t, funcDef.Args.Args, 2)
	assert.Len(t, funcDef.Body, 1)
}

func TestParserFunctionDefWithDefaults(t *testing.T) {
	input := `def greet(name, greeting="Hello"):
    return greeting + " " + name`

	parser := NewParser(input)
	module, errs := parser.Parse()

	require.Empty(t, errs)
	require.Len(t, module.Body, 1)

	funcDef, ok := module.Body[0].(*model.FunctionDef)
	require.True(t, ok, "expected FunctionDef")
	assert.Len(t, funcDef.Args.Args, 2)
	assert.Len(t, funcDef.Args.Defaults, 1)
}

func TestParserFunctionDefWithAnnotations(t *testing.T) {
	input := `def add(a: int, b: int) -> int:
    return a + b`

	parser := NewParser(input)
	module, errs := parser.Parse()

	require.Empty(t, errs)
	require.Len(t, module.Body, 1)

	funcDef, ok := module.Body[0].(*model.FunctionDef)
	require.True(t, ok, "expected FunctionDef")
	assert.NotNil(t, funcDef.Returns)
	assert.NotNil(t, funcDef.Args.Args[0].Annotation)
	assert.NotNil(t, funcDef.Args.Args[1].Annotation)
}

func TestParserClassDef(t *testing.T) {
	input := `class MyClass:
    def __init__(self):
        pass`

	parser := NewParser(input)
	module, errs := parser.Parse()

	require.Empty(t, errs)
	require.Len(t, module.Body, 1)

	classDef, ok := module.Body[0].(*model.ClassDef)
	require.True(t, ok, "expected ClassDef")
	assert.Equal(t, "MyClass", classDef.Name.Name)
	assert.Len(t, classDef.Body, 1)
}

func TestParserClassDefWithBases(t *testing.T) {
	input := `class Child(Parent):
    pass`

	parser := NewParser(input)
	module, errs := parser.Parse()

	require.Empty(t, errs)
	require.Len(t, module.Body, 1)

	classDef, ok := module.Body[0].(*model.ClassDef)
	require.True(t, ok, "expected ClassDef")
	assert.Len(t, classDef.Bases, 1)
}

func TestParserImport(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"simple", "import os"},
		{"multiple", "import os, sys"},
		{"alias", "import numpy as np"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parser := NewParser(test.input)
			module, errs := parser.Parse()

			require.Empty(t, errs, "unexpected parser errors")
			require.Len(t, module.Body, 1)

			_, ok := module.Body[0].(*model.Import)
			require.True(t, ok, "expected Import")
		})
	}
}

func TestParserFromImport(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"simple", "from os import path"},
		{"multiple", "from os import path, getcwd"},
		{"star", "from os import *"},
		{"relative", "from . import module"},
		{"alias", "from os import path as p"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parser := NewParser(test.input)
			module, errs := parser.Parse()

			require.Empty(t, errs, "unexpected parser errors")
			require.Len(t, module.Body, 1)

			_, ok := module.Body[0].(*model.ImportFrom)
			require.True(t, ok, "expected ImportFrom")
		})
	}
}

func TestParserTryExcept(t *testing.T) {
	input := `try:
    risky()
except ValueError:
    handle()
except:
    default()
finally:
    cleanup()`

	parser := NewParser(input)
	module, errs := parser.Parse()

	require.Empty(t, errs)
	require.Len(t, module.Body, 1)

	tryStmt, ok := module.Body[0].(*model.Try)
	require.True(t, ok, "expected Try")
	assert.Len(t, tryStmt.Body, 1)
	assert.Len(t, tryStmt.Handlers, 2)
	assert.Len(t, tryStmt.FinalBody, 1)
}

func TestParserExceptStar(t *testing.T) {
	input := `try:
    risky()
except* ValueError as e:
    handle_value(e)
except* TypeError:
    handle_type()`

	parser := NewParser(input)
	module, errs := parser.Parse()

	require.Empty(t, errs)
	require.Len(t, module.Body, 1)

	tryStmt, ok := module.Body[0].(*model.Try)
	require.True(t, ok, "expected Try")
	assert.Len(t, tryStmt.Body, 1)
	require.Len(t, tryStmt.Handlers, 2)

	// First handler: except* ValueError as e
	h0 := tryStmt.Handlers[0]
	assert.True(t, h0.IsStar, "first handler should be except*")
	assert.NotNil(t, h0.Type)
	assert.NotNil(t, h0.Name)
	assert.Equal(t, "e", h0.Name.Name)

	// Second handler: except* TypeError
	h1 := tryStmt.Handlers[1]
	assert.True(t, h1.IsStar, "second handler should be except*")
	assert.NotNil(t, h1.Type)
	assert.Nil(t, h1.Name)
}

func TestParserExceptStarWithFinally(t *testing.T) {
	input := `try:
    risky()
except* ValueError:
    handle()
finally:
    cleanup()`

	parser := NewParser(input)
	module, errs := parser.Parse()

	require.Empty(t, errs)
	require.Len(t, module.Body, 1)

	tryStmt, ok := module.Body[0].(*model.Try)
	require.True(t, ok, "expected Try")
	require.Len(t, tryStmt.Handlers, 1)
	assert.True(t, tryStmt.Handlers[0].IsStar)
	assert.Len(t, tryStmt.FinalBody, 1)
}

func TestParserExceptStarMixedError(t *testing.T) {
	input := `try:
    risky()
except ValueError:
    handle()
except* TypeError:
    handle2()`

	parser := NewParser(input)
	_, errs := parser.Parse()

	require.NotEmpty(t, errs, "expected error for mixing except and except*")
}

func TestParserExceptStarBareError(t *testing.T) {
	input := `try:
    risky()
except*:
    handle()`

	parser := NewParser(input)
	_, errs := parser.Parse()

	require.NotEmpty(t, errs, "expected error for bare except*")
}

func TestParserExceptNotStar(t *testing.T) {
	// Regular except should NOT set IsStar
	input := `try:
    risky()
except ValueError:
    handle()`

	parser := NewParser(input)
	module, errs := parser.Parse()

	require.Empty(t, errs)
	tryStmt := module.Body[0].(*model.Try)
	assert.False(t, tryStmt.Handlers[0].IsStar, "regular except should not be star")
}

func TestParserWithStatement(t *testing.T) {
	input := `with open("file") as f:
    data = f.read()`

	parser := NewParser(input)
	module, errs := parser.Parse()

	require.Empty(t, errs)
	require.Len(t, module.Body, 1)

	withStmt, ok := module.Body[0].(*model.With)
	require.True(t, ok, "expected With")
	assert.Len(t, withStmt.Items, 1)
	assert.NotNil(t, withStmt.Items[0].OptionalVar)
}

func TestParserDecorator(t *testing.T) {
	input := `@decorator
def func():
    pass`

	parser := NewParser(input)
	module, errs := parser.Parse()

	require.Empty(t, errs)
	require.Len(t, module.Body, 1)

	funcDef, ok := module.Body[0].(*model.FunctionDef)
	require.True(t, ok, "expected FunctionDef")
	assert.Len(t, funcDef.Decorators, 1)
}

func TestParserLambda(t *testing.T) {
	parser := NewParser("lambda x, y: x + y")
	module, errs := parser.Parse()

	require.Empty(t, errs)
	require.Len(t, module.Body, 1)

	exprStmt := module.Body[0].(*model.ExprStmt)
	lambda, ok := exprStmt.Value.(*model.Lambda)
	require.True(t, ok, "expected Lambda")
	assert.Len(t, lambda.Args.Args, 2)
}

func TestParserTernary(t *testing.T) {
	parser := NewParser("x if condition else y")
	module, errs := parser.Parse()

	require.Empty(t, errs)
	require.Len(t, module.Body, 1)

	exprStmt := module.Body[0].(*model.ExprStmt)
	ifExpr, ok := exprStmt.Value.(*model.IfExpr)
	require.True(t, ok, "expected IfExpr")
	assert.NotNil(t, ifExpr.Test)
	assert.NotNil(t, ifExpr.Body)
	assert.NotNil(t, ifExpr.OrElse)
}

func TestParserWalrus(t *testing.T) {
	parser := NewParser("(n := 10)")
	module, errs := parser.Parse()

	require.Empty(t, errs)
	require.Len(t, module.Body, 1)

	exprStmt := module.Body[0].(*model.ExprStmt)
	namedExpr, ok := exprStmt.Value.(*model.NamedExpr)
	require.True(t, ok, "expected NamedExpr")
	assert.Equal(t, "n", namedExpr.Target.Name)
}

func TestParserComplex(t *testing.T) {
	input := `def factorial(n):
    if n <= 1:
        return 1
    return n * factorial(n - 1)

result = factorial(5)
print(result)`

	parser := NewParser(input)
	module, errs := parser.Parse()

	require.Empty(t, errs)
	assert.Len(t, module.Body, 3)
}

func TestParserAsync(t *testing.T) {
	input := `async def fetch():
    await something()`

	parser := NewParser(input)
	module, errs := parser.Parse()

	require.Empty(t, errs)
	require.Len(t, module.Body, 1)

	funcDef, ok := module.Body[0].(*model.FunctionDef)
	require.True(t, ok, "expected FunctionDef")
	assert.True(t, funcDef.IsAsync)
}

func TestParserStarredExpr(t *testing.T) {
	parser := NewParser("[*items]")
	module, errs := parser.Parse()

	require.Empty(t, errs)
	require.Len(t, module.Body, 1)

	exprStmt := module.Body[0].(*model.ExprStmt)
	list, ok := exprStmt.Value.(*model.List)
	require.True(t, ok, "expected List")

	_, ok = list.Elts[0].(*model.Starred)
	require.True(t, ok, "expected Starred")
}
