package compiler

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Escape Sequence Tests
//
// Valid escapes are tested via CompileSource (verifying the string constant in
// the code object). Error cases are tested via the Lexer directly, since the
// parser discards lexer errors.
// =============================================================================

// helper: compile "x = <pyLiteral>" and return the string constant from the
// code object's Constants pool.
func compileStringLiteral(t *testing.T, pyLiteral string) string {
	t.Helper()
	source := "x = " + pyLiteral
	code, errs := CompileSource(source, "<test>")
	require.Empty(t, errs, "unexpected compile errors for %s: %v", pyLiteral, errs)
	require.NotNil(t, code)

	for _, c := range code.Constants {
		if s, ok := c.(string); ok {
			return s
		}
	}
	t.Fatalf("no string constant found in code object for source: %s", source)
	return ""
}

// helper: tokenize the given string literal and expect a lexer error containing
// the given substring.
func expectLexerError(t *testing.T, literal string, errSubstring string) {
	t.Helper()
	lexer := NewLexer(literal)
	_, errs := lexer.Tokenize()
	require.NotEmpty(t, errs, "expected lexer error for %s but got none", literal)

	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), errSubstring) {
			found = true
			break
		}
	}
	assert.True(t, found, "expected lexer error containing %q for %s, got: %v", errSubstring, literal, errs)
}

// =============================================================================
// Hex Escape: \xNN
// =============================================================================

func TestEscapeHexValid(t *testing.T) {
	tests := []struct {
		name     string
		literal  string
		expected string
	}{
		{"\\x41 -> A", `"\x41"`, "A"},
		{"\\x61 -> a", `"\x61"`, "a"},
		{"\\x00 -> null byte", `"\x00"`, "\x00"},
		{"\\xFF -> U+00FF", `"\xFF"`, string(rune(0xFF))},
		{"\\x0a -> newline", `"\x0a"`, "\n"},
		{"\\x09 -> tab", `"\x09"`, "\t"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compileStringLiteral(t, tt.literal)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEscapeHexErrors(t *testing.T) {
	tests := []struct {
		name    string
		literal string
		errMsg  string
	}{
		{"only 1 hex digit", `"\x4"`, "invalid \\x escape"},
		{"\\x at end of string", `"\x"`, "invalid \\x escape"},
		{"invalid hex digits GG", `"\xGG"`, "invalid \\x escape"},
		{"invalid hex digits ZZ", `"\xZZ"`, "invalid \\x escape"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectLexerError(t, tt.literal, tt.errMsg)
		})
	}
}

// =============================================================================
// Unicode Escape: \uNNNN (4-digit)
// =============================================================================

func TestEscapeUnicode4Valid(t *testing.T) {
	tests := []struct {
		name     string
		literal  string
		expected string
	}{
		{"\\u0041 -> A", `"\u0041"`, "A"},
		{"\\u00e9 -> e-acute", `"\u00e9"`, "\u00e9"},
		{"\\u0020 -> space", `"\u0020"`, " "},
		{"\\u03C0 -> pi", `"\u03C0"`, "\u03C0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compileStringLiteral(t, tt.literal)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEscapeUnicode4Errors(t *testing.T) {
	tests := []struct {
		name    string
		literal string
		errMsg  string
	}{
		{"only 3 hex digits", `"\u004"`, "invalid \\u escape"},
		{"only 2 hex digits", `"\u00"`, "invalid \\u escape"},
		{"only 1 hex digit", `"\u0"`, "invalid \\u escape"},
		{"no hex digits", `"\u"`, "invalid \\u escape"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectLexerError(t, tt.literal, tt.errMsg)
		})
	}
}

// =============================================================================
// Unicode Escape: \UNNNNNNNN (8-digit)
// =============================================================================

func TestEscapeUnicode8Valid(t *testing.T) {
	tests := []struct {
		name     string
		literal  string
		expected string
	}{
		{"\\U00000041 -> A", `"\U00000041"`, "A"},
		{"\\U0001F600 -> emoji", `"\U0001F600"`, "\U0001F600"},
		{"\\U000003C0 -> pi", `"\U000003C0"`, "\u03C0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compileStringLiteral(t, tt.literal)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEscapeUnicode8Errors(t *testing.T) {
	tests := []struct {
		name    string
		literal string
		errMsg  string
	}{
		{"only 7 hex digits", `"\U0000004"`, "invalid \\U escape"},
		{"only 4 hex digits", `"\U0041"`, "invalid \\U escape"},
		{"no hex digits", `"\U"`, "invalid \\U escape"},
		{"out of range U+99999999", `"\U99999999"`, "invalid Unicode character"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectLexerError(t, tt.literal, tt.errMsg)
		})
	}
}

// =============================================================================
// Named Unicode Escape: \N{name}
// =============================================================================

func TestEscapeNamedUnicodeValid(t *testing.T) {
	tests := []struct {
		name     string
		literal  string
		expected string
	}{
		{"SPACE", `"\N{SPACE}"`, " "},
		{"GREEK SMALL LETTER PI", `"\N{GREEK SMALL LETTER PI}"`, "\u03C0"},
		{"EURO SIGN", `"\N{EURO SIGN}"`, "\u20AC"},
		{"COPYRIGHT SIGN", `"\N{COPYRIGHT SIGN}"`, "\u00A9"},
		{"LATIN SMALL LETTER A", `"\N{LATIN SMALL LETTER A}"`, "a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compileStringLiteral(t, tt.literal)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEscapeNamedUnicodeErrors(t *testing.T) {
	tests := []struct {
		name    string
		literal string
		errMsg  string
	}{
		{"unknown name", `"\N{UNKNOWN}"`, "unknown Unicode character name"},
		{"no braces", `"\Nx"`, "invalid \\N escape"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectLexerError(t, tt.literal, tt.errMsg)
		})
	}
}

// =============================================================================
// Octal Escape: \0, \12, \177, \777
// =============================================================================

func TestEscapeOctalValid(t *testing.T) {
	tests := []struct {
		name     string
		literal  string
		expected string
	}{
		{"\\0 -> null byte", `"\0"`, "\x00"},
		{"\\7 -> BEL", `"\7"`, "\x07"},
		{"\\12 -> newline (octal 12 = decimal 10)", `"\12"`, "\n"},
		{"\\101 -> A (octal 101 = 65)", `"\101"`, "A"},
		{"\\177 -> DEL (octal 177 = 127)", `"\177"`, "\x7F"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compileStringLiteral(t, tt.literal)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEscapeOctalOverflow(t *testing.T) {
	// \777 = octal 777 = 511, which exceeds 0o377 (255)
	expectLexerError(t, `"\777"`, "invalid octal escape")
}

// =============================================================================
// Standard Single-Character Escapes
// =============================================================================

func TestEscapeStandardSequences(t *testing.T) {
	tests := []struct {
		name     string
		literal  string
		expected string
	}{
		{"\\n -> newline", `"\n"`, "\n"},
		{"\\t -> tab", `"\t"`, "\t"},
		{"\\r -> carriage return", `"\r"`, "\r"},
		{"\\\\ -> backslash", `"\\"`, "\\"},
		{"\\a -> bell", `"\a"`, "\a"},
		{"\\b -> backspace", `"\b"`, "\b"},
		{"\\f -> form feed", `"\f"`, "\f"},
		{"\\v -> vertical tab", `"\v"`, "\v"},
		{"\\' -> single quote", `"\'"`, "'"},
		{"\\\" in single-quoted", `'\\"'`, `\"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compileStringLiteral(t, tt.literal)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// =============================================================================
// Unknown Escape: kept as-is (Python behavior)
// =============================================================================

func TestEscapeUnknownKeptAsIs(t *testing.T) {
	// In Python, unknown escapes like \z are kept as-is: literal backslash + z
	result := compileStringLiteral(t, `"\z"`)
	assert.Equal(t, "\\z", result)
}

func TestEscapeUnknownOtherChars(t *testing.T) {
	tests := []struct {
		name     string
		literal  string
		expected string
	}{
		{"\\q -> \\q", `"\q"`, "\\q"},
		{"\\w -> \\w", `"\w"`, "\\w"},
		{"\\p -> \\p", `"\p"`, "\\p"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compileStringLiteral(t, tt.literal)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// =============================================================================
// Combined / Multi-escape Strings
// =============================================================================

func TestEscapeCombinedSequences(t *testing.T) {
	// Multiple escapes in one string
	result := compileStringLiteral(t, `"hello\nworld\t!"`)
	assert.Equal(t, "hello\nworld\t!", result)
}

func TestEscapeHexInContext(t *testing.T) {
	// \x41 in the middle of a string
	result := compileStringLiteral(t, `"before\x41after"`)
	assert.Equal(t, "beforeAafter", result)
}

func TestEscapeUnicodeInContext(t *testing.T) {
	// \u0041 in the middle of a string
	result := compileStringLiteral(t, `"before\u0041after"`)
	assert.Equal(t, "beforeAafter", result)
}

// =============================================================================
// Raw Strings: Escapes not processed
// =============================================================================

func TestRawStringEscapesNotProcessed(t *testing.T) {
	tests := []struct {
		name     string
		literal  string
		expected string
	}{
		{"raw \\n kept", `r"\n"`, `\n`},
		{"raw \\t kept", `r"\t"`, `\t`},
		{"raw \\x41 kept", `r"\x41"`, `\x41`},
		{"raw \\u0041 kept", `r"\u0041"`, `\u0041`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compileStringLiteral(t, tt.literal)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// =============================================================================
// Direct Lexer Tests for Escape Sequences
//
// These test the lexer tokenization directly for more fine-grained validation
// of both valid and invalid escape sequences.
// =============================================================================

func TestLexerEscapeHexDirect(t *testing.T) {
	lexer := NewLexer(`"\x41"`)
	tokens, errs := lexer.Tokenize()
	require.Empty(t, errs)
	require.NotEmpty(t, tokens)
	assert.Equal(t, "A", tokens[0].Literal)
}

func TestLexerEscapeHexErrorDirect(t *testing.T) {
	lexer := NewLexer(`"\x4"`)
	_, errs := lexer.Tokenize()
	require.NotEmpty(t, errs, "expected error for \\x with only 1 hex digit")
	assert.Contains(t, errs[0].Error(), "invalid \\x escape")
}

func TestLexerEscapeUnicode4Direct(t *testing.T) {
	lexer := NewLexer(`"\u00e9"`)
	tokens, errs := lexer.Tokenize()
	require.Empty(t, errs)
	require.NotEmpty(t, tokens)
	assert.Equal(t, "\u00e9", tokens[0].Literal)
}

func TestLexerEscapeUnicode8Direct(t *testing.T) {
	lexer := NewLexer(`"\U0001F600"`)
	tokens, errs := lexer.Tokenize()
	require.Empty(t, errs)
	require.NotEmpty(t, tokens)
	assert.Equal(t, "\U0001F600", tokens[0].Literal)
}

func TestLexerEscapeOctalDirect(t *testing.T) {
	lexer := NewLexer(`"\101"`)
	tokens, errs := lexer.Tokenize()
	require.Empty(t, errs)
	require.NotEmpty(t, tokens)
	assert.Equal(t, "A", tokens[0].Literal)
}

func TestLexerEscapeOctalOverflowDirect(t *testing.T) {
	lexer := NewLexer(`"\777"`)
	_, errs := lexer.Tokenize()
	require.NotEmpty(t, errs)
	assert.Contains(t, errs[0].Error(), "invalid octal escape")
}

func TestLexerEscapeNamedDirect(t *testing.T) {
	lexer := NewLexer(`"\N{SPACE}"`)
	tokens, errs := lexer.Tokenize()
	require.Empty(t, errs)
	require.NotEmpty(t, tokens)
	assert.Equal(t, " ", tokens[0].Literal)
}

func TestLexerEscapeNamedPiDirect(t *testing.T) {
	lexer := NewLexer(`"\N{GREEK SMALL LETTER PI}"`)
	tokens, errs := lexer.Tokenize()
	require.Empty(t, errs)
	require.NotEmpty(t, tokens)
	assert.Equal(t, "\u03C0", tokens[0].Literal)
}

func TestLexerEscapeNamedUnknownDirect(t *testing.T) {
	lexer := NewLexer(`"\N{UNKNOWN}"`)
	_, errs := lexer.Tokenize()
	require.NotEmpty(t, errs)
	assert.Contains(t, errs[0].Error(), "unknown Unicode character name")
}

func TestLexerEscapeNamedNoBracesDirect(t *testing.T) {
	lexer := NewLexer(`"\Nx"`)
	_, errs := lexer.Tokenize()
	require.NotEmpty(t, errs)
	assert.Contains(t, errs[0].Error(), "invalid \\N escape")
}

// =============================================================================
// Edge case: \x at exact end of string (no digits follow)
// =============================================================================

func TestLexerEscapeHexAtEndDirect(t *testing.T) {
	lexer := NewLexer(`"\x"`)
	_, errs := lexer.Tokenize()
	require.NotEmpty(t, errs, "\\x at end of string should produce error")
	assert.Contains(t, errs[0].Error(), "invalid \\x escape")
}

func TestLexerEscapeHexOnlyOneDigitDirect(t *testing.T) {
	lexer := NewLexer(`"\xG"`)
	_, errs := lexer.Tokenize()
	require.NotEmpty(t, errs, "\\xG should produce error (G is not hex)")
	assert.Contains(t, errs[0].Error(), "invalid \\x escape")
}

// =============================================================================
// Edge case: \u and \U with insufficient digits at end of string
// =============================================================================

func TestLexerEscapeUnicode4AtEndDirect(t *testing.T) {
	lexer := NewLexer(`"\u"`)
	_, errs := lexer.Tokenize()
	require.NotEmpty(t, errs, "\\u at end of string should produce error")
	assert.Contains(t, errs[0].Error(), "invalid \\u escape")
}

func TestLexerEscapeUnicode8AtEndDirect(t *testing.T) {
	lexer := NewLexer(`"\U"`)
	_, errs := lexer.Tokenize()
	require.NotEmpty(t, errs, "\\U at end of string should produce error")
	assert.Contains(t, errs[0].Error(), "invalid \\U escape")
}

func TestLexerEscapeUnicode8OutOfRangeDirect(t *testing.T) {
	lexer := NewLexer(`"\U99999999"`)
	_, errs := lexer.Tokenize()
	require.NotEmpty(t, errs)
	assert.Contains(t, errs[0].Error(), "invalid Unicode character")
}
