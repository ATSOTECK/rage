package compiler

import (
	"fmt"
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
		if val > 0o377 {
			l.addError(fmt.Sprintf("invalid octal escape \\%s: value exceeds 0o377", octal))
			val = val & 0xFF
		}
		return string(rune(val))
	case 'x':
		// Hex escape: \xNN — requires exactly 2 hex digits
		hex := l.consumeHexDigits(2)
		if len(hex) != 2 {
			l.addError("invalid \\x escape (must be followed by exactly 2 hex digits)")
			return "\\x" + hex
		}
		return string(rune(parseHex(hex)))
	case 'u':
		// Unicode escape: \uNNNN — requires exactly 4 hex digits
		hex := l.consumeHexDigits(4)
		if len(hex) != 4 {
			l.addError("invalid \\u escape (must be followed by exactly 4 hex digits)")
			return "\\u" + hex
		}
		return string(rune(parseHex(hex)))
	case 'U':
		// Unicode escape: \UNNNNNNNN — requires exactly 8 hex digits
		hex := l.consumeHexDigits(8)
		if len(hex) != 8 {
			l.addError("invalid \\U escape (must be followed by exactly 8 hex digits)")
			return "\\U" + hex
		}
		val := parseHex(hex)
		if val > 0x10FFFF {
			l.addError(fmt.Sprintf("invalid Unicode character U+%X", val))
			return "\uFFFD"
		}
		return string(rune(val))
	case 'N':
		// Named Unicode escape: \N{name} — not yet supported
		if !l.isAtEnd() && l.peek() == '{' {
			l.advance() // consume '{'
			name := ""
			for !l.isAtEnd() && l.peek() != '}' {
				name += string(l.advance())
			}
			if !l.isAtEnd() {
				l.advance() // consume '}'
			}
			if r, ok := lookupUnicodeName(name); ok {
				return string(r)
			}
			l.addError(fmt.Sprintf("unknown Unicode character name '%s'", name))
			return "\uFFFD"
		}
		l.addError("invalid \\N escape (must be \\N{name})")
		return "\\N"
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

// consumeHexDigits consumes up to n hex digits from input and returns them.
func (l *Lexer) consumeHexDigits(n int) string {
	hex := ""
	for i := 0; i < n && !l.isAtEnd(); i++ {
		next := l.peek()
		if isHexDigit(next) {
			hex += string(l.advance())
		} else {
			break
		}
	}
	return hex
}

// parseHex parses a hex string into an integer.
func parseHex(hex string) int {
	val := 0
	for _, c := range hex {
		val *= 16
		switch {
		case c >= '0' && c <= '9':
			val += int(c - '0')
		case c >= 'a' && c <= 'f':
			val += int(c-'a') + 10
		case c >= 'A' && c <= 'F':
			val += int(c-'A') + 10
		}
	}
	return val
}

// lookupUnicodeName maps a Unicode character name to its rune.
// Covers common names; returns false for unknown names.
func lookupUnicodeName(name string) (rune, bool) {
	name = strings.ToUpper(strings.TrimSpace(name))
	if r, ok := unicodeNames[name]; ok {
		return r, true
	}
	return 0, false
}

// unicodeNames maps Unicode character names to runes.
// This covers commonly used names. For full coverage, a comprehensive database would be needed.
var unicodeNames = map[string]rune{
	// ASCII control characters
	"NULL":                            0x0000,
	"BACKSPACE":                       0x0008,
	"CHARACTER TABULATION":            0x0009,
	"TAB":                             0x0009,
	"LINE FEED":                       0x000A,
	"CARRIAGE RETURN":                 0x000D,
	"ESCAPE":                          0x001B,
	"SPACE":                           0x0020,
	"DELETE":                          0x007F,
	"NO-BREAK SPACE":                  0x00A0,

	// Punctuation & symbols
	"EXCLAMATION MARK":                     0x0021,
	"QUOTATION MARK":                       0x0022,
	"NUMBER SIGN":                          0x0023,
	"PERCENT SIGN":                         0x0025,
	"AMPERSAND":                            0x0026,
	"APOSTROPHE":                           0x0027,
	"LEFT PARENTHESIS":                     0x0028,
	"RIGHT PARENTHESIS":                    0x0029,
	"ASTERISK":                             0x002A,
	"PLUS SIGN":                            0x002B,
	"COMMA":                                0x002C,
	"HYPHEN-MINUS":                         0x002D,
	"FULL STOP":                            0x002E,
	"SOLIDUS":                              0x002F,
	"COLON":                                0x003A,
	"SEMICOLON":                            0x003B,
	"LESS-THAN SIGN":                       0x003C,
	"EQUALS SIGN":                          0x003D,
	"GREATER-THAN SIGN":                    0x003E,
	"QUESTION MARK":                        0x003F,
	"COMMERCIAL AT":                        0x0040,
	"LEFT SQUARE BRACKET":                  0x005B,
	"REVERSE SOLIDUS":                      0x005C,
	"RIGHT SQUARE BRACKET":                 0x005D,
	"CIRCUMFLEX ACCENT":                    0x005E,
	"LOW LINE":                             0x005F,
	"GRAVE ACCENT":                         0x0060,
	"LEFT CURLY BRACKET":                   0x007B,
	"VERTICAL LINE":                        0x007C,
	"RIGHT CURLY BRACKET":                  0x007D,
	"TILDE":                                0x007E,

	// Latin letters (uppercase)
	"LATIN CAPITAL LETTER A": 'A', "LATIN CAPITAL LETTER B": 'B',
	"LATIN CAPITAL LETTER C": 'C', "LATIN CAPITAL LETTER D": 'D',
	"LATIN CAPITAL LETTER E": 'E', "LATIN CAPITAL LETTER F": 'F',
	"LATIN CAPITAL LETTER G": 'G', "LATIN CAPITAL LETTER H": 'H',
	"LATIN CAPITAL LETTER I": 'I', "LATIN CAPITAL LETTER J": 'J',
	"LATIN CAPITAL LETTER K": 'K', "LATIN CAPITAL LETTER L": 'L',
	"LATIN CAPITAL LETTER M": 'M', "LATIN CAPITAL LETTER N": 'N',
	"LATIN CAPITAL LETTER O": 'O', "LATIN CAPITAL LETTER P": 'P',
	"LATIN CAPITAL LETTER Q": 'Q', "LATIN CAPITAL LETTER R": 'R',
	"LATIN CAPITAL LETTER S": 'S', "LATIN CAPITAL LETTER T": 'T',
	"LATIN CAPITAL LETTER U": 'U', "LATIN CAPITAL LETTER V": 'V',
	"LATIN CAPITAL LETTER W": 'W', "LATIN CAPITAL LETTER X": 'X',
	"LATIN CAPITAL LETTER Y": 'Y', "LATIN CAPITAL LETTER Z": 'Z',

	// Latin letters (lowercase)
	"LATIN SMALL LETTER A": 'a', "LATIN SMALL LETTER B": 'b',
	"LATIN SMALL LETTER C": 'c', "LATIN SMALL LETTER D": 'd',
	"LATIN SMALL LETTER E": 'e', "LATIN SMALL LETTER F": 'f',
	"LATIN SMALL LETTER G": 'g', "LATIN SMALL LETTER H": 'h',
	"LATIN SMALL LETTER I": 'i', "LATIN SMALL LETTER J": 'j',
	"LATIN SMALL LETTER K": 'k', "LATIN SMALL LETTER L": 'l',
	"LATIN SMALL LETTER M": 'm', "LATIN SMALL LETTER N": 'n',
	"LATIN SMALL LETTER O": 'o', "LATIN SMALL LETTER P": 'p',
	"LATIN SMALL LETTER Q": 'q', "LATIN SMALL LETTER R": 'r',
	"LATIN SMALL LETTER S": 's', "LATIN SMALL LETTER T": 't',
	"LATIN SMALL LETTER U": 'u', "LATIN SMALL LETTER V": 'v',
	"LATIN SMALL LETTER W": 'w', "LATIN SMALL LETTER X": 'x',
	"LATIN SMALL LETTER Y": 'y', "LATIN SMALL LETTER Z": 'z',

	// Digits
	"DIGIT ZERO": '0', "DIGIT ONE": '1', "DIGIT TWO": '2',
	"DIGIT THREE": '3', "DIGIT FOUR": '4', "DIGIT FIVE": '5',
	"DIGIT SIX": '6', "DIGIT SEVEN": '7', "DIGIT EIGHT": '8',
	"DIGIT NINE": '9',

	// Common Latin extended
	"LATIN SMALL LETTER A WITH ACUTE":      0x00E1,
	"LATIN SMALL LETTER E WITH ACUTE":      0x00E9,
	"LATIN SMALL LETTER I WITH ACUTE":      0x00ED,
	"LATIN SMALL LETTER O WITH ACUTE":      0x00F3,
	"LATIN SMALL LETTER U WITH ACUTE":      0x00FA,
	"LATIN SMALL LETTER N WITH TILDE":      0x00F1,
	"LATIN SMALL LETTER U WITH DIAERESIS":  0x00FC,
	"LATIN SMALL LETTER A WITH GRAVE":      0x00E0,
	"LATIN SMALL LETTER E WITH GRAVE":      0x00E8,
	"LATIN SMALL LETTER A WITH CIRCUMFLEX": 0x00E2,
	"LATIN SMALL LETTER E WITH CIRCUMFLEX": 0x00EA,
	"LATIN SMALL LETTER C WITH CEDILLA":    0x00E7,
	"LATIN CAPITAL LETTER A WITH ACUTE":    0x00C1,
	"LATIN CAPITAL LETTER E WITH ACUTE":    0x00C9,
	"LATIN CAPITAL LETTER N WITH TILDE":    0x00D1,

	// Currency
	"CENT SIGN":     0x00A2,
	"POUND SIGN":    0x00A3,
	"CURRENCY SIGN": 0x00A4,
	"YEN SIGN":      0x00A5,
	"EURO SIGN":     0x20AC,
	"DOLLAR SIGN":   0x0024,

	// Math & technical
	"PLUS-MINUS SIGN":             0x00B1,
	"MULTIPLICATION SIGN":         0x00D7,
	"DIVISION SIGN":               0x00F7,
	"NOT SIGN":                    0x00AC,
	"MICRO SIGN":                  0x00B5,
	"DEGREE SIGN":                 0x00B0,
	"SUPERSCRIPT TWO":             0x00B2,
	"SUPERSCRIPT THREE":           0x00B3,
	"VULGAR FRACTION ONE HALF":    0x00BD,
	"VULGAR FRACTION ONE QUARTER": 0x00BC,
	"INFINITY":                    0x221E,
	"SQUARE ROOT":                 0x221A,
	"ALMOST EQUAL TO":             0x2248,
	"NOT EQUAL TO":                0x2260,
	"LESS-THAN OR EQUAL TO":       0x2264,
	"GREATER-THAN OR EQUAL TO":    0x2265,
	"GREEK SMALL LETTER PI":       0x03C0,
	"GREEK SMALL LETTER ALPHA":    0x03B1,
	"GREEK SMALL LETTER BETA":     0x03B2,
	"GREEK SMALL LETTER GAMMA":    0x03B3,
	"GREEK SMALL LETTER DELTA":    0x03B4,
	"GREEK SMALL LETTER EPSILON":  0x03B5,
	"GREEK SMALL LETTER LAMBDA":   0x03BB,
	"GREEK SMALL LETTER MU":       0x03BC,
	"GREEK SMALL LETTER SIGMA":    0x03C3,
	"GREEK SMALL LETTER OMEGA":    0x03C9,
	"GREEK CAPITAL LETTER SIGMA":  0x03A3,
	"GREEK CAPITAL LETTER OMEGA":  0x03A9,
	"GREEK CAPITAL LETTER DELTA":  0x0394,

	// Arrows & misc symbols
	"LEFTWARDS ARROW":             0x2190,
	"UPWARDS ARROW":               0x2191,
	"RIGHTWARDS ARROW":            0x2192,
	"DOWNWARDS ARROW":             0x2193,
	"LEFT RIGHT ARROW":            0x2194,
	"BULLET":                      0x2022,
	"HORIZONTAL ELLIPSIS":         0x2026,
	"EM DASH":                     0x2014,
	"EN DASH":                     0x2013,
	"LEFT SINGLE QUOTATION MARK":  0x2018,
	"RIGHT SINGLE QUOTATION MARK": 0x2019,
	"LEFT DOUBLE QUOTATION MARK":  0x201C,
	"RIGHT DOUBLE QUOTATION MARK": 0x201D,
	"MIDDLE DOT":                  0x00B7,
	"SECTION SIGN":                0x00A7,
	"PILCROW SIGN":                0x00B6,
	"COPYRIGHT SIGN":              0x00A9,
	"REGISTERED SIGN":             0x00AE,
	"TRADE MARK SIGN":             0x2122,
	"CHECK MARK":                  0x2713,
	"HEAVY CHECK MARK":            0x2714,
	"BALLOT X":                    0x2717,
	"SNOWFLAKE":                   0x2744,
	"BLACK STAR":                  0x2605,
	"WHITE STAR":                  0x2606,
	"BLACK HEART SUIT":            0x2665,

	// Box drawing
	"BOX DRAWINGS LIGHT HORIZONTAL": 0x2500,
	"BOX DRAWINGS LIGHT VERTICAL":   0x2502,

	// Special
	"REPLACEMENT CHARACTER":  0xFFFD,
	"ZERO WIDTH SPACE":       0x200B,
	"ZERO WIDTH NON-JOINER":  0x200C,
	"ZERO WIDTH JOINER":      0x200D,
	"LEFT-TO-RIGHT MARK":     0x200E,
	"RIGHT-TO-LEFT MARK":     0x200F,
	"BYTE ORDER MARK":        0xFEFF,
	"SUPERSCRIPT ONE":        0x00B9,
	"INVERTED EXCLAMATION MARK": 0x00A1,
	"INVERTED QUESTION MARK":    0x00BF,
}
