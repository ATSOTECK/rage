package compiler

import (
	"testing"

	"github.com/ATSOTECK/rage/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLexerBasicTokens(t *testing.T) {
	tests := []struct {
		input    string
		expected []model.TokenKind
	}{
		{"+", []model.TokenKind{model.TK_Plus, model.TK_EOF}},
		{"-", []model.TokenKind{model.TK_Minus, model.TK_EOF}},
		{"*", []model.TokenKind{model.TK_Star, model.TK_EOF}},
		{"/", []model.TokenKind{model.TK_Slash, model.TK_EOF}},
		{"**", []model.TokenKind{model.TK_DoubleStar, model.TK_EOF}},
		{"//", []model.TokenKind{model.TK_DoubleSlash, model.TK_EOF}},
		{"+=", []model.TokenKind{model.TK_PlusAssign, model.TK_EOF}},
		{"-=", []model.TokenKind{model.TK_MinusAssign, model.TK_EOF}},
		{"*=", []model.TokenKind{model.TK_StarAssign, model.TK_EOF}},
		{"/=", []model.TokenKind{model.TK_SlashAssign, model.TK_EOF}},
		{"**=", []model.TokenKind{model.TK_DoubleStarAssign, model.TK_EOF}},
		{"//=", []model.TokenKind{model.TK_DoubleSlashAssign, model.TK_EOF}},
		{"==", []model.TokenKind{model.TK_Equal, model.TK_EOF}},
		{"!=", []model.TokenKind{model.TK_NotEqual, model.TK_EOF}},
		{"<", []model.TokenKind{model.TK_Less, model.TK_EOF}},
		{">", []model.TokenKind{model.TK_Greater, model.TK_EOF}},
		{"<=", []model.TokenKind{model.TK_LessEqual, model.TK_EOF}},
		{">=", []model.TokenKind{model.TK_GreaterEqual, model.TK_EOF}},
		{"<<", []model.TokenKind{model.TK_LShift, model.TK_EOF}},
		{">>", []model.TokenKind{model.TK_RShift, model.TK_EOF}},
		{":=", []model.TokenKind{model.TK_Walrus, model.TK_EOF}},
		{"->", []model.TokenKind{model.TK_Arrow, model.TK_EOF}},
		{"...", []model.TokenKind{model.TK_Ellipsis, model.TK_EOF}},
		{"()", []model.TokenKind{model.TK_LParen, model.TK_RParen, model.TK_EOF}},
		{"[]", []model.TokenKind{model.TK_LBracket, model.TK_RBracket, model.TK_EOF}},
		{"{}", []model.TokenKind{model.TK_LBrace, model.TK_RBrace, model.TK_EOF}},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			lexer := NewLexer(test.input)
			tokens, errs := lexer.Tokenize()

			require.Empty(t, errs, "unexpected lexer errors")
			require.Len(t, tokens, len(test.expected), "token count mismatch")

			for i, tok := range tokens {
				assert.Equal(t, test.expected[i], tok.Kind, "token %d kind mismatch", i)
			}
		})
	}
}

func TestLexerKeywords(t *testing.T) {
	input := "if else elif while for def class return True False None and or not in is lambda try except finally raise with as async await break continue pass import from global nonlocal yield assert del match case type"
	expected := []model.TokenKind{
		model.TK_If, model.TK_Else, model.TK_Elif, model.TK_While, model.TK_For,
		model.TK_Def, model.TK_Class, model.TK_Return, model.TK_True, model.TK_False,
		model.TK_None, model.TK_And, model.TK_Or, model.TK_Not, model.TK_In,
		model.TK_Is, model.TK_Lambda, model.TK_Try, model.TK_Except, model.TK_Finally,
		model.TK_Raise, model.TK_With, model.TK_As, model.TK_Async, model.TK_Await,
		model.TK_Break, model.TK_Continue, model.TK_Pass, model.TK_Import, model.TK_From,
		model.TK_Global, model.TK_Nonlocal, model.TK_Yield, model.TK_Assert, model.TK_Del,
		model.TK_Match, model.TK_Case, model.TK_Type, model.TK_EOF,
	}

	lexer := NewLexer(input)
	tokens, errs := lexer.Tokenize()

	require.Empty(t, errs, "unexpected lexer errors")
	require.Len(t, tokens, len(expected), "token count mismatch")

	for i, tok := range tokens {
		assert.Equal(t, expected[i], tok.Kind, "token %d: expected %s, got %s", i, expected[i], tok.Kind)
	}
}

func TestLexerIdentifiers(t *testing.T) {
	tests := []struct {
		input   string
		literal string
	}{
		{"foo", "foo"},
		{"bar123", "bar123"},
		{"_private", "_private"},
		{"__dunder__", "__dunder__"},
		{"CamelCase", "CamelCase"},
		{"snake_case", "snake_case"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			lexer := NewLexer(test.input)
			tokens, errs := lexer.Tokenize()

			require.Empty(t, errs, "unexpected lexer errors")
			require.NotEmpty(t, tokens, "expected at least one token")

			assert.Equal(t, model.TK_Identifier, tokens[0].Kind)
			assert.Equal(t, test.literal, tokens[0].Literal)
		})
	}
}

func TestLexerNumbers(t *testing.T) {
	tests := []struct {
		input   string
		kind    model.TokenKind
		literal string
	}{
		{"0", model.TK_IntLit, "0"},
		{"123", model.TK_IntLit, "123"},
		{"1_000_000", model.TK_IntLit, "1_000_000"},
		{"0x1F", model.TK_IntLit, "0x1F"},
		{"0xFF", model.TK_IntLit, "0xFF"},
		{"0o777", model.TK_IntLit, "0o777"},
		{"0b1010", model.TK_IntLit, "0b1010"},
		{"3.14", model.TK_FloatLit, "3.14"},
		{"0.5", model.TK_FloatLit, "0.5"},
		{"1e10", model.TK_FloatLit, "1e10"},
		{"1E-5", model.TK_FloatLit, "1E-5"},
		{"3.14e+2", model.TK_FloatLit, "3.14e+2"},
		{"1j", model.TK_ImaginaryLit, "1j"},
		{"3.14j", model.TK_ImaginaryLit, "3.14j"},
		{"1e10J", model.TK_ImaginaryLit, "1e10J"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			lexer := NewLexer(test.input)
			tokens, errs := lexer.Tokenize()

			require.Empty(t, errs, "unexpected lexer errors")
			require.NotEmpty(t, tokens, "expected at least one token")

			assert.Equal(t, test.kind, tokens[0].Kind)
			assert.Equal(t, test.literal, tokens[0].Literal)
		})
	}
}

func TestLexerStrings(t *testing.T) {
	tests := []struct {
		input   string
		literal string
	}{
		{`"hello"`, "hello"},
		{`'hello'`, "hello"},
		{`"hello\nworld"`, "hello\nworld"},
		{`"hello\tworld"`, "hello\tworld"},
		{`"escaped\"quote"`, `escaped"quote`},
		{`'escaped\'quote'`, `escaped'quote`},
		{`""`, ""},
		{`''`, ""},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			lexer := NewLexer(test.input)
			tokens, errs := lexer.Tokenize()

			require.Empty(t, errs, "unexpected lexer errors")
			require.NotEmpty(t, tokens, "expected at least one token")

			assert.Equal(t, model.TK_StringLit, tokens[0].Kind)
			assert.Equal(t, test.literal, tokens[0].Literal)
		})
	}
}

func TestLexerTripleQuotedStrings(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		literal string
	}{
		{"double triple", `"""hello world"""`, "hello world"},
		{"single triple", `'''hello world'''`, "hello world"},
		{"multiline", "\"\"\"line1\nline2\"\"\"", "line1\nline2"},
		{"with quotes", `"""he said "hi" there"""`, `he said "hi" there`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			lexer := NewLexer(test.input)
			tokens, errs := lexer.Tokenize()

			require.Empty(t, errs, "unexpected lexer errors")
			require.NotEmpty(t, tokens, "expected at least one token")

			assert.Equal(t, model.TK_StringLit, tokens[0].Kind)
			assert.Equal(t, test.literal, tokens[0].Literal)
		})
	}
}

func TestLexerRawStrings(t *testing.T) {
	tests := []struct {
		input   string
		literal string
	}{
		{`r"hello\nworld"`, `hello\nworld`},
		{`R"hello\tworld"`, `hello\tworld`},
		{`r'path\to\file'`, `path\to\file`},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			lexer := NewLexer(test.input)
			tokens, errs := lexer.Tokenize()

			require.Empty(t, errs, "unexpected lexer errors")
			require.NotEmpty(t, tokens, "expected at least one token")

			assert.Equal(t, model.TK_StringLit, tokens[0].Kind)
			assert.Equal(t, test.literal, tokens[0].Literal)
		})
	}
}

func TestLexerBytesStrings(t *testing.T) {
	tests := []struct {
		input   string
		literal string
	}{
		{`b"hello"`, "hello"},
		{`B"hello"`, "hello"},
		{`b'hello'`, "hello"},
		{`br"hello\n"`, `hello\n`},
		{`rb"hello\n"`, `hello\n`},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			lexer := NewLexer(test.input)
			tokens, errs := lexer.Tokenize()

			require.Empty(t, errs, "unexpected lexer errors")
			require.NotEmpty(t, tokens, "expected at least one token")

			assert.Equal(t, model.TK_BytesLit, tokens[0].Kind)
			assert.Equal(t, test.literal, tokens[0].Literal)
		})
	}
}

func TestLexerFStrings(t *testing.T) {
	tests := []struct {
		input   string
		literal string
	}{
		{`f"hello"`, "hello"},
		{`F"hello"`, "hello"},
		{`f"hello {name}"`, "hello {name}"},
		{`f"2 + 2 = {2 + 2}"`, "2 + 2 = {2 + 2}"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			lexer := NewLexer(test.input)
			tokens, errs := lexer.Tokenize()

			require.Empty(t, errs, "unexpected lexer errors")
			require.NotEmpty(t, tokens, "expected at least one token")

			assert.Equal(t, model.TK_StringLit, tokens[0].Kind)
			assert.Equal(t, test.literal, tokens[0].Literal)
		})
	}
}

func TestLexerComments(t *testing.T) {
	input := "x = 1 # this is a comment"
	expected := []model.TokenKind{
		model.TK_Identifier, model.TK_Assign, model.TK_IntLit,
		model.TK_Comment, model.TK_EOF,
	}

	lexer := NewLexer(input)
	tokens, errs := lexer.Tokenize()

	require.Empty(t, errs, "unexpected lexer errors")
	require.Len(t, tokens, len(expected), "token count mismatch")

	for i, tok := range tokens {
		assert.Equal(t, expected[i], tok.Kind, "token %d kind mismatch", i)
	}
}

func TestLexerIndentation(t *testing.T) {
	input := `if True:
    x = 1
    if False:
        y = 2
    z = 3
w = 4`

	lexer := NewLexer(input)
	tokens, errs := lexer.Tokenize()

	require.Empty(t, errs, "unexpected lexer errors")

	indentCount := 0
	dedentCount := 0
	for _, tok := range tokens {
		if tok.Kind == model.TK_Indent {
			indentCount++
		}
		if tok.Kind == model.TK_Dedent {
			dedentCount++
		}
	}

	assert.Equal(t, 2, indentCount, "INDENT token count mismatch")
	assert.Equal(t, 2, dedentCount, "DEDENT token count mismatch")
}

func TestLexerNewlineInBrackets(t *testing.T) {
	input := `foo(
    1,
    2
)`

	lexer := NewLexer(input)
	tokens, errs := lexer.Tokenize()

	require.Empty(t, errs, "unexpected lexer errors")

	for i, tok := range tokens {
		if tok.Kind == model.TK_Newline && i < len(tokens)-2 {
			t.Errorf("unexpected NEWLINE inside brackets at position %d", i)
		}
	}
}

func TestLexerLineContinuation(t *testing.T) {
	input := `x = 1 + \
2 + 3`
	expected := []model.TokenKind{
		model.TK_Identifier, model.TK_Assign, model.TK_IntLit,
		model.TK_Plus, model.TK_IntLit, model.TK_Plus, model.TK_IntLit,
		model.TK_EOF,
	}

	lexer := NewLexer(input)
	tokens, errs := lexer.Tokenize()

	require.Empty(t, errs, "unexpected lexer errors")
	require.Len(t, tokens, len(expected), "token count mismatch")

	for i, tok := range tokens {
		assert.Equal(t, expected[i], tok.Kind, "token %d kind mismatch", i)
	}
}

func TestLexerComplexExpression(t *testing.T) {
	input := `def factorial(n):
    if n <= 1:
        return 1
    return n * factorial(n - 1)`

	lexer := NewLexer(input)
	tokens, errs := lexer.Tokenize()

	require.Empty(t, errs, "unexpected lexer errors")

	hasDefKeyword := false
	hasReturnKeyword := false
	for _, tok := range tokens {
		if tok.Kind == model.TK_Def {
			hasDefKeyword = true
		}
		if tok.Kind == model.TK_Return {
			hasReturnKeyword = true
		}
	}

	assert.True(t, hasDefKeyword, "expected to find 'def' keyword")
	assert.True(t, hasReturnKeyword, "expected to find 'return' keyword")
}

func TestLexerPosition(t *testing.T) {
	input := "x = 1\ny = 2"
	lexer := NewLexer(input)
	tokens, _ := lexer.Tokenize()

	require.NotEmpty(t, tokens, "expected at least one token")

	assert.Equal(t, 1, tokens[0].Pos.Line, "first token line")
	assert.Equal(t, 1, tokens[0].Pos.Column, "first token column")

	for _, tok := range tokens {
		if tok.Kind == model.TK_Identifier && tok.Literal == "y" {
			assert.Equal(t, 2, tok.Pos.Line, "'y' should be on line 2")
			break
		}
	}
}

func TestLexerWithFilename(t *testing.T) {
	input := "x = 1"
	lexer := NewLexerWithFilename(input, "test.py")
	tokens, _ := lexer.Tokenize()

	require.NotEmpty(t, tokens, "expected at least one token")
	assert.Equal(t, "test.py", tokens[0].Pos.Filename)
}

func TestLexerErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"unexpected char", "x = $y"},
		{"invalid indent", "    x = 1\n  y = 2"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			lexer := NewLexer(test.input)
			_, errs := lexer.Tokenize()

			assert.NotEmpty(t, errs, "expected lexer errors")
		})
	}
}
