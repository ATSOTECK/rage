package compiler

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/ATSOTECK/rage/internal/model"
)

// Lexer tokenizes Python source code.
type Lexer struct {
	source   string
	filename string

	pos       int // current position in source (byte offset)
	line      int // current line number (1-indexed)
	column    int // current column number (1-indexed)
	start     int // start position of current token
	startLine int
	startCol  int

	// Indentation tracking
	indentStack   []int         // stack of indentation levels
	pendingTokens []model.Token // tokens to emit before next scan
	atLineStart   bool          // are we at the start of a logical line?

	// Bracket nesting (suppress NEWLINE inside brackets)
	bracketDepth int

	// For f-string handling
	fStringStack []fStringState

	tokens []model.Token
	errors []LexError
}

type fStringState struct {
	quote      string // the quote style used (', ", ''', """)
	isRaw      bool
	braceDepth int
}

// LexError represents a lexical error.
type LexError struct {
	Pos     model.Position
	Message string
}

func (e LexError) Error() string {
	return fmt.Sprintf("%s: %s", e.Pos, e.Message)
}

// NewLexer creates a new lexer for the given source.
func NewLexer(source string) *Lexer {
	return NewLexerWithFilename(source, "")
}

// NewLexerWithFilename creates a new lexer with a filename for error messages.
func NewLexerWithFilename(source, filename string) *Lexer {
	return &Lexer{
		source:      source,
		filename:    filename,
		line:        1,
		column:      1,
		indentStack: []int{0},
		atLineStart: true,
	}
}

// Tokenize scans the entire source and returns all tokens.
func (l *Lexer) Tokenize() ([]model.Token, []LexError) {
	for {
		tok := l.NextToken()
		l.tokens = append(l.tokens, tok)
		if tok.Kind == model.TK_EOF {
			break
		}
	}
	return l.tokens, l.errors
}

// NextToken returns the next token from the source.
func (l *Lexer) NextToken() model.Token {
	// Return any pending tokens first (INDENT/DEDENT)
	if len(l.pendingTokens) > 0 {
		tok := l.pendingTokens[0]
		l.pendingTokens = l.pendingTokens[1:]
		return tok
	}

	// Handle indentation at the start of a line
	if l.atLineStart && l.bracketDepth == 0 {
		l.atLineStart = false
		if tok := l.handleIndentation(); tok != nil {
			return *tok
		}
	}

	l.skipWhitespace()

	if l.isAtEnd() {
		// Emit DEDENT tokens for remaining indentation levels
		if len(l.indentStack) > 1 {
			l.indentStack = l.indentStack[:len(l.indentStack)-1]
			for len(l.indentStack) > 1 {
				l.pendingTokens = append(l.pendingTokens, l.makeToken(model.TK_Dedent))
				l.indentStack = l.indentStack[:len(l.indentStack)-1]
			}
			return l.makeToken(model.TK_Dedent)
		}
		return l.makeToken(model.TK_EOF)
	}

	l.start = l.pos
	l.startLine = l.line
	l.startCol = l.column

	// Check for non-ASCII (Unicode) identifier start before advancing
	// This must be done before l.advance() to properly decode multi-byte UTF-8
	if l.peek() >= utf8.RuneSelf {
		r := l.peekRune()
		if isIdentifierStartRune(r) {
			l.advanceRune() // Advance past the first rune
			return l.scanIdentifier()
		}
		// Not a valid identifier start; advance and report error
		l.advanceRune()
		l.addError(fmt.Sprintf("unexpected character '%c'", r))
		return l.makeToken(model.TK_Illegal)
	}

	ch := l.advance()

	// Comments
	if ch == '#' {
		return l.scanComment()
	}

	// Newlines
	if ch == '\n' {
		l.atLineStart = true
		if l.bracketDepth > 0 {
			// Inside brackets, skip newlines
			return l.NextToken()
		}
		return l.makeToken(model.TK_Newline)
	}

	// Line continuation
	if ch == '\\' && l.peek() == '\n' {
		l.advance() // consume \n
		l.atLineStart = false
		return l.NextToken()
	}

	// String literals
	if ch == '\'' || ch == '"' {
		return l.scanString(rune(ch), false, false)
	}

	// String prefixes (r, b, f, u, fr, rf, br, rb)
	if isStringPrefix(ch) {
		if prefix := l.tryStringPrefix(ch); prefix != "" {
			return l.scanPrefixedString(prefix)
		}
	}

	// Numbers
	if isDigit(ch) || (ch == '.' && isDigit(l.peek())) {
		l.retreat()
		return l.scanNumber()
	}

	// Identifiers and keywords
	if isIdentifierStart(ch) {
		return l.scanIdentifier()
	}

	// Operators and delimiters
	return l.scanOperator(ch)
}

func (l *Lexer) handleIndentation() *model.Token {
	// Count leading whitespace
	indent := 0
	for !l.isAtEnd() {
		ch := l.peek()
		if ch == ' ' {
			indent++
			l.advance()
		} else if ch == '\t' {
			// Tabs align to 8-space boundaries
			indent = (indent/8 + 1) * 8
			l.advance()
		} else {
			break
		}
	}

	// Skip blank lines and comment-only lines
	if l.peek() == '\n' || l.peek() == '#' || l.isAtEnd() {
		return nil
	}

	currentIndent := l.indentStack[len(l.indentStack)-1]

	if indent > currentIndent {
		l.indentStack = append(l.indentStack, indent)
		tok := l.makeToken(model.TK_Indent)
		return &tok
	} else if indent < currentIndent {
		// Generate DEDENT tokens
		for len(l.indentStack) > 1 && l.indentStack[len(l.indentStack)-1] > indent {
			l.indentStack = l.indentStack[:len(l.indentStack)-1]
			l.pendingTokens = append(l.pendingTokens, l.makeToken(model.TK_Dedent))
		}

		if l.indentStack[len(l.indentStack)-1] != indent {
			l.addError("unindent does not match any outer indentation level")
		}

		if len(l.pendingTokens) > 0 {
			tok := l.pendingTokens[0]
			l.pendingTokens = l.pendingTokens[1:]
			return &tok
		}
	}

	return nil
}

func (l *Lexer) scanComment() model.Token {
	for !l.isAtEnd() && l.peek() != '\n' {
		l.advance()
	}
	return l.makeTokenWithLiteral(model.TK_Comment, l.source[l.start:l.pos])
}

func (l *Lexer) scanString(quote rune, isRaw, isBytes bool) model.Token {
	// Check for triple quotes
	triple := false
	if l.peek() == byte(quote) && l.peekN(1) == byte(quote) {
		triple = true
		l.advance()
		l.advance()
	}

	var builder strings.Builder

	for !l.isAtEnd() {
		ch := l.advance()

		if !triple && ch == '\n' {
			l.addError("EOL while scanning string literal")
			break
		}

		if ch == byte(quote) {
			if triple {
				if l.peek() == byte(quote) && l.peekN(1) == byte(quote) {
					l.advance()
					l.advance()
					break
				}
				builder.WriteByte(ch)
			} else {
				break
			}
			continue
		}

		if ch == '\\' && !isRaw {
			if l.isAtEnd() {
				l.addError("EOL while scanning string literal")
				break
			}
			escaped := l.advance()
			builder.WriteByte(l.processEscape(escaped))
		} else {
			builder.WriteByte(ch)
		}
	}

	kind := model.TK_StringLit
	if isBytes {
		kind = model.TK_BytesLit
	}

	return l.makeTokenWithLiteral(kind, builder.String())
}

func (l *Lexer) scanPrefixedString(prefix string) model.Token {
	isRaw := strings.ContainsAny(prefix, "rR")
	isBytes := strings.ContainsAny(prefix, "bB")
	isFString := strings.ContainsAny(prefix, "fF")

	quote := l.advance()
	if quote != '\'' && quote != '"' {
		l.addError("expected string quote after prefix")
		return l.makeToken(model.TK_Illegal)
	}

	if isFString {
		return l.scanFString(rune(quote), isRaw)
	}

	return l.scanString(rune(quote), isRaw, isBytes)
}

func (l *Lexer) scanFString(quote rune, isRaw bool) model.Token {
	// Check for triple quotes
	triple := false
	if l.peek() == byte(quote) && l.peekN(1) == byte(quote) {
		triple = true
		l.advance()
		l.advance()
	}

	quoteStr := string(quote)
	if triple {
		quoteStr = string([]rune{quote, quote, quote})
	}

	// Push f-string state
	l.fStringStack = append(l.fStringStack, fStringState{
		quote:      quoteStr,
		isRaw:      isRaw,
		braceDepth: 0,
	})

	// For now, we'll treat f-strings as regular strings
	// A full implementation would need to tokenize the expressions inside {}
	var builder strings.Builder

	for !l.isAtEnd() {
		ch := l.advance()

		if !triple && ch == '\n' {
			l.addError("EOL while scanning f-string literal")
			break
		}

		if ch == byte(quote) {
			if triple {
				if l.peek() == byte(quote) && l.peekN(1) == byte(quote) {
					l.advance()
					l.advance()
					break
				}
				builder.WriteByte(ch)
			} else {
				break
			}
			continue
		}

		if ch == '{' {
			if l.peek() == '{' {
				l.advance()
				builder.WriteString("{{")
				continue
			}
			// For a complete implementation, we'd start tokenizing the expression here
			builder.WriteByte(ch)
			braceDepth := 1
			for !l.isAtEnd() && braceDepth > 0 {
				c := l.advance()
				builder.WriteByte(c)
				if c == '{' {
					braceDepth++
				} else if c == '}' {
					braceDepth--
				}
			}
			continue
		}

		if ch == '}' {
			if l.peek() == '}' {
				l.advance()
				builder.WriteString("}}")
				continue
			}
			l.addError("single '}' is not allowed in f-string")
		}

		if ch == '\\' && !isRaw {
			if l.isAtEnd() {
				l.addError("EOL while scanning f-string literal")
				break
			}
			escaped := l.advance()
			builder.WriteByte(l.processEscape(escaped))
		} else {
			builder.WriteByte(ch)
		}
	}

	// Pop f-string state
	if len(l.fStringStack) > 0 {
		l.fStringStack = l.fStringStack[:len(l.fStringStack)-1]
	}

	return l.makeTokenWithLiteral(model.TK_StringLit, builder.String())
}

func (l *Lexer) processEscape(ch byte) byte {
	switch ch {
	case 'n':
		return '\n'
	case 't':
		return '\t'
	case 'r':
		return '\r'
	case '\\':
		return '\\'
	case '\'':
		return '\''
	case '"':
		return '"'
	case '0':
		return '\x00'
	case 'a':
		return '\a'
	case 'b':
		return '\b'
	case 'f':
		return '\f'
	case 'v':
		return '\v'
	default:
		// For \x, \u, \U, \N we'd need more complex handling
		return ch
	}
}

func (l *Lexer) scanNumber() model.Token {
	l.start = l.pos
	l.startLine = l.line
	l.startCol = l.column

	ch := l.advance()

	// Check for hex, octal, binary
	if ch == '0' {
		next := l.peek()
		switch next {
		case 'x', 'X':
			return l.scanHexNumber()
		case 'o', 'O':
			return l.scanOctalNumber()
		case 'b', 'B':
			return l.scanBinaryNumber()
		}
	}

	// Decimal integer or float
	isFloat := false
	hasExponent := false

	// Consume digits before decimal point
	if ch != '.' {
		l.scanDigits()
	}

	// Check for decimal point
	if l.peek() == '.' && l.peekN(1) != '.' { // Avoid confusing with ...
		if isDigit(l.peekN(1)) || !isIdentifierStart(l.peekN(1)) {
			isFloat = true
			l.advance() // consume '.'
			l.scanDigits()
		}
	}

	// Check for exponent
	if l.peek() == 'e' || l.peek() == 'E' {
		isFloat = true
		hasExponent = true
		l.advance()
		if l.peek() == '+' || l.peek() == '-' {
			l.advance()
		}
		if !isDigit(l.peek()) {
			l.addError("invalid decimal literal")
		}
		l.scanDigits()
	}
	_ = hasExponent

	// Check for imaginary
	if l.peek() == 'j' || l.peek() == 'J' {
		l.advance()
		return l.makeTokenWithLiteral(model.TK_ImaginaryLit, l.source[l.start:l.pos])
	}

	if isFloat {
		return l.makeTokenWithLiteral(model.TK_FloatLit, l.source[l.start:l.pos])
	}
	return l.makeTokenWithLiteral(model.TK_IntLit, l.source[l.start:l.pos])
}

func (l *Lexer) scanDigits() {
	for isDigit(l.peek()) || l.peek() == '_' {
		l.advance()
	}
}

func (l *Lexer) scanHexNumber() model.Token {
	l.advance() // consume 'x' or 'X'
	if !isHexDigit(l.peek()) {
		l.addError("invalid hexadecimal literal")
	}
	for isHexDigit(l.peek()) || l.peek() == '_' {
		l.advance()
	}
	return l.makeTokenWithLiteral(model.TK_IntLit, l.source[l.start:l.pos])
}

func (l *Lexer) scanOctalNumber() model.Token {
	l.advance() // consume 'o' or 'O'
	if !isOctalDigit(l.peek()) {
		l.addError("invalid octal literal")
	}
	for isOctalDigit(l.peek()) || l.peek() == '_' {
		l.advance()
	}
	return l.makeTokenWithLiteral(model.TK_IntLit, l.source[l.start:l.pos])
}

func (l *Lexer) scanBinaryNumber() model.Token {
	l.advance() // consume 'b' or 'B'
	if !isBinaryDigit(l.peek()) {
		l.addError("invalid binary literal")
	}
	for isBinaryDigit(l.peek()) || l.peek() == '_' {
		l.advance()
	}
	return l.makeTokenWithLiteral(model.TK_IntLit, l.source[l.start:l.pos])
}

func (l *Lexer) scanIdentifier() model.Token {
	for !l.isAtEnd() {
		ch := l.peek()
		if ch < utf8.RuneSelf {
			// ASCII: use fast byte-based check
			if !isIdentifierChar(ch) {
				break
			}
			l.advance()
		} else {
			// Non-ASCII: decode full rune and check
			r := l.peekRune()
			if r == utf8.RuneError || !isIdentifierCharRune(r) {
				break
			}
			l.advanceRune()
		}
	}
	literal := l.source[l.start:l.pos]
	kind := model.LookupIdent(literal)
	return l.makeTokenWithLiteral(kind, literal)
}

func (l *Lexer) scanOperator(ch byte) model.Token {
	switch ch {
	case '+':
		if l.match('=') {
			return l.makeToken(model.TK_PlusAssign)
		}
		return l.makeToken(model.TK_Plus)

	case '-':
		if l.match('=') {
			return l.makeToken(model.TK_MinusAssign)
		}
		if l.match('>') {
			return l.makeToken(model.TK_Arrow)
		}
		return l.makeToken(model.TK_Minus)

	case '*':
		if l.match('*') {
			if l.match('=') {
				return l.makeToken(model.TK_DoubleStarAssign)
			}
			return l.makeToken(model.TK_DoubleStar)
		}
		if l.match('=') {
			return l.makeToken(model.TK_StarAssign)
		}
		return l.makeToken(model.TK_Star)

	case '/':
		if l.match('/') {
			if l.match('=') {
				return l.makeToken(model.TK_DoubleSlashAssign)
			}
			return l.makeToken(model.TK_DoubleSlash)
		}
		if l.match('=') {
			return l.makeToken(model.TK_SlashAssign)
		}
		return l.makeToken(model.TK_Slash)

	case '%':
		if l.match('=') {
			return l.makeToken(model.TK_PercentAssign)
		}
		return l.makeToken(model.TK_Percent)

	case '@':
		if l.match('=') {
			return l.makeToken(model.TK_AtAssign)
		}
		return l.makeToken(model.TK_At)

	case '&':
		if l.match('=') {
			return l.makeToken(model.TK_AmpersandAssign)
		}
		return l.makeToken(model.TK_Ampersand)

	case '|':
		if l.match('=') {
			return l.makeToken(model.TK_PipeAssign)
		}
		return l.makeToken(model.TK_Pipe)

	case '^':
		if l.match('=') {
			return l.makeToken(model.TK_CaretAssign)
		}
		return l.makeToken(model.TK_Caret)

	case '~':
		return l.makeToken(model.TK_Tilde)

	case '<':
		if l.match('<') {
			if l.match('=') {
				return l.makeToken(model.TK_LShiftAssign)
			}
			return l.makeToken(model.TK_LShift)
		}
		if l.match('=') {
			return l.makeToken(model.TK_LessEqual)
		}
		return l.makeToken(model.TK_Less)

	case '>':
		if l.match('>') {
			if l.match('=') {
				return l.makeToken(model.TK_RShiftAssign)
			}
			return l.makeToken(model.TK_RShift)
		}
		if l.match('=') {
			return l.makeToken(model.TK_GreaterEqual)
		}
		return l.makeToken(model.TK_Greater)

	case '=':
		if l.match('=') {
			return l.makeToken(model.TK_Equal)
		}
		return l.makeToken(model.TK_Assign)

	case '!':
		if l.match('=') {
			return l.makeToken(model.TK_NotEqual)
		}
		l.addError("unexpected character '!'")
		return l.makeToken(model.TK_Illegal)

	case ':':
		if l.match('=') {
			return l.makeToken(model.TK_Walrus)
		}
		return l.makeToken(model.TK_Colon)

	case '.':
		if l.match('.') {
			if l.match('.') {
				return l.makeToken(model.TK_Ellipsis)
			}
			// Two dots is invalid in Python
			l.retreat()
		}
		return l.makeToken(model.TK_Dot)

	case ',':
		return l.makeToken(model.TK_Comma)

	case ';':
		return l.makeToken(model.TK_Semicolon)

	case '(':
		l.bracketDepth++
		return l.makeToken(model.TK_LParen)

	case ')':
		l.bracketDepth--
		return l.makeToken(model.TK_RParen)

	case '[':
		l.bracketDepth++
		return l.makeToken(model.TK_LBracket)

	case ']':
		l.bracketDepth--
		return l.makeToken(model.TK_RBracket)

	case '{':
		l.bracketDepth++
		return l.makeToken(model.TK_LBrace)

	case '}':
		l.bracketDepth--
		return l.makeToken(model.TK_RBrace)

	default:
		l.addError(fmt.Sprintf("unexpected character '%c'", ch))
		return l.makeToken(model.TK_Illegal)
	}
}

// Helper methods

func (l *Lexer) isAtEnd() bool {
	return l.pos >= len(l.source)
}

func (l *Lexer) peek() byte {
	if l.isAtEnd() {
		return 0
	}
	return l.source[l.pos]
}

func (l *Lexer) peekN(n int) byte {
	if l.pos+n >= len(l.source) {
		return 0
	}
	return l.source[l.pos+n]
}

func (l *Lexer) advance() byte {
	if l.isAtEnd() {
		return 0
	}
	ch := l.source[l.pos]
	l.pos++
	if ch == '\n' {
		l.line++
		l.column = 1
	} else {
		l.column++
	}
	return ch
}

func (l *Lexer) retreat() {
	if l.pos > 0 {
		l.pos--
		if l.source[l.pos] == '\n' {
			l.line--
			// Recalculate column (find last newline before current position)
			l.column = 1
			for i := l.pos - 1; i >= 0; i-- {
				if l.source[i] == '\n' {
					break
				}
				l.column++
			}
		} else {
			l.column--
		}
	}
}

// peekRune returns the next rune without advancing. Returns 0 if at end.
func (l *Lexer) peekRune() rune {
	if l.isAtEnd() {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.source[l.pos:])
	return r
}

// advanceRune advances past one rune and returns it.
func (l *Lexer) advanceRune() rune {
	if l.isAtEnd() {
		return 0
	}
	r, width := utf8.DecodeRuneInString(l.source[l.pos:])
	l.pos += width
	if r == '\n' {
		l.line++
		l.column = 1
	} else {
		l.column++
	}
	return r
}

func (l *Lexer) match(expected byte) bool {
	if l.isAtEnd() || l.source[l.pos] != expected {
		return false
	}
	l.advance()
	return true
}

func (l *Lexer) skipWhitespace() {
	for !l.isAtEnd() {
		ch := l.peek()
		switch ch {
		case ' ', '\t', '\r':
			l.advance()
		default:
			return
		}
	}
}

func (l *Lexer) makeToken(kind model.TokenKind) model.Token {
	return model.Token{
		Kind: kind,
		Pos: model.Position{
			Filename: l.filename,
			Line:     l.startLine,
			Column:   l.startCol,
			Offset:   l.start,
		},
		EndPos: model.Position{
			Filename: l.filename,
			Line:     l.line,
			Column:   l.column,
			Offset:   l.pos,
		},
	}
}

func (l *Lexer) makeTokenWithLiteral(kind model.TokenKind, literal string) model.Token {
	tok := l.makeToken(kind)
	tok.Literal = literal
	return tok
}

func (l *Lexer) addError(msg string) {
	l.errors = append(l.errors, LexError{
		Pos: model.Position{
			Filename: l.filename,
			Line:     l.line,
			Column:   l.column,
			Offset:   l.pos,
		},
		Message: msg,
	})
}

func (l *Lexer) tryStringPrefix(ch byte) string {
	// Check for string prefixes: r, R, b, B, f, F, u, U, fr, Fr, fR, FR, rf, rF, Rf, RF, br, Br, bR, BR, rb, rB, Rb, RB
	lower := toLower(ch)
	prefix := string(ch)

	if lower == 'u' {
		if l.peek() == '\'' || l.peek() == '"' {
			return prefix
		}
		return ""
	}

	if lower == 'r' || lower == 'b' || lower == 'f' {
		next := l.peek()
		if next == '\'' || next == '"' {
			return prefix
		}

		nextLower := toLower(next)
		if lower == 'r' && (nextLower == 'b' || nextLower == 'f') {
			if l.peekN(1) == '\'' || l.peekN(1) == '"' {
				l.advance()
				return prefix + string(next)
			}
		}
		if lower == 'b' && nextLower == 'r' {
			if l.peekN(1) == '\'' || l.peekN(1) == '"' {
				l.advance()
				return prefix + string(next)
			}
		}
		if lower == 'f' && nextLower == 'r' {
			if l.peekN(1) == '\'' || l.peekN(1) == '"' {
				l.advance()
				return prefix + string(next)
			}
		}
	}

	return ""
}

// Character classification helpers

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isHexDigit(ch byte) bool {
	return isDigit(ch) || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

func isOctalDigit(ch byte) bool {
	return ch >= '0' && ch <= '7'
}

func isBinaryDigit(ch byte) bool {
	return ch == '0' || ch == '1'
}

func isIdentifierStart(ch byte) bool {
	// Only handles ASCII; non-ASCII requires rune-based checking
	if ch < utf8.RuneSelf {
		return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
	}
	// Non-ASCII byte: return true only for UTF-8 leading bytes (0xC0-0xFF)
	// Continuation bytes (0x80-0xBF) cannot start an identifier
	return ch >= 0xC0
}

func isIdentifierChar(ch byte) bool {
	// Only handles ASCII; non-ASCII requires rune-based checking
	if ch < utf8.RuneSelf {
		return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') || ch == '_'
	}
	// Non-ASCII: allow any byte >= 0x80 (both leading and continuation bytes)
	// as they may be part of a valid multi-byte Unicode character
	return true
}

// isIdentifierStartRune checks if a rune can start a Python identifier
func isIdentifierStartRune(r rune) bool {
	return unicode.IsLetter(r) || r == '_'
}

// isIdentifierCharRune checks if a rune can be part of a Python identifier
func isIdentifierCharRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}

func isStringPrefix(ch byte) bool {
	lower := toLower(ch)
	return lower == 'r' || lower == 'b' || lower == 'f' || lower == 'u'
}

func toLower(ch byte) byte {
	if ch >= 'A' && ch <= 'Z' {
		return ch + ('a' - 'A')
	}
	return ch
}

// GetTokens returns all tokens (for backwards compatibility).
func (l *Lexer) GetTokens() []model.Token {
	tokens, _ := l.Tokenize()
	return tokens
}

// Errors returns any lexical errors encountered.
func (l *Lexer) Errors() []LexError {
	return l.errors
}
