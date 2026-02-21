package compiler

import (
	"strings"

	"github.com/ATSOTECK/rage/internal/model"
)

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
			builder.WriteString(l.processEscapeSequence(escaped))
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

	// Ensure fStringStack is popped on all exit paths
	defer func() {
		if len(l.fStringStack) > 0 {
			l.fStringStack = l.fStringStack[:len(l.fStringStack)-1]
		}
	}()

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
			builder.WriteString(l.processEscapeSequence(escaped))
		} else {
			builder.WriteByte(ch)
		}
	}

	// f-string state is popped by defer

	return l.makeTokenWithLiteral(model.TK_FStringLit, builder.String())
}

// processEscapeSequence handles escape sequences and returns the resulting string.
// It may consume additional characters from the lexer for multi-char escapes.
func (l *Lexer) processEscapeSequence(ch byte) string {
	switch ch {
	case 'n':
		return "\n"
	case 't':
		return "\t"
	case 'r':
		return "\r"
	case '\\':
		return "\\"
	case '\'':
		return "'"
	case '"':
		return "\""
	case 'a':
		return "\a"
	case 'b':
		return "\b"
	case 'f':
		return "\f"
	case 'v':
		return "\v"
	case '0', '1', '2', '3', '4', '5', '6', '7':
		// Octal escape: \0, \00, \000, \1, \12, \123, etc.
		octal := string(ch)
		// Consume up to 2 more octal digits
		for i := 0; i < 2 && !l.isAtEnd(); i++ {
			next := l.peek()
			if next >= '0' && next <= '7' {
				octal += string(l.advance())
			} else {
				break
			}
		}
		// Parse octal value
		var val int
		for _, c := range octal {
			val = val*8 + int(c-'0')
		}
		if val > 255 {
			val = 255 // Clamp to byte range
		}
		return string(byte(val))
	case 'x':
		// Hex escape: \xNN
		if l.isAtEnd() {
			return "\\x"
		}
		hex := ""
		for i := 0; i < 2 && !l.isAtEnd(); i++ {
			next := l.peek()
			if isHexDigit(next) {
				hex += string(l.advance())
			} else {
				break
			}
		}
		if len(hex) == 0 {
			return "\\x"
		}
		val := 0
		for _, c := range hex {
			val *= 16
			if c >= '0' && c <= '9' {
				val += int(c - '0')
			} else if c >= 'a' && c <= 'f' {
				val += int(c-'a') + 10
			} else if c >= 'A' && c <= 'F' {
				val += int(c-'A') + 10
			}
		}
		return string(byte(val))
	default:
		// Unknown escape, return as-is (Python behavior)
		return "\\" + string(ch)
	}
}

// processEscape is kept for backward compatibility but delegates to processEscapeSequence
func (l *Lexer) processEscape(ch byte) byte {
	result := l.processEscapeSequence(ch)
	if len(result) == 1 {
		return result[0]
	}
	// For multi-byte results, return just the first byte (this shouldn't happen in normal use)
	return result[0]
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
