package runtime

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Format specification mini-language, shared by format(), str.format() and
// f-strings:
//
//	[[fill]align][sign]["z"]["#"]["0"][width][grouping]["." [precision][grouping]][type]

// formatSpec is a parsed standard format specifier.
type formatSpec struct {
	fill      rune
	align     byte // '<', '>', '^', '=' or 0 for the type's default
	sign      byte // '+', '-', ' ' or 0
	noNegZero bool // 'z': coerce negative zero to positive zero
	alternate bool // '#'
	width     int  // -1 when not given
	grouping  byte // ',' or '_' separator for the integer part, or 0
	precision int  // -1 when not given
	fracGroup byte // ',' or '_' separator for the fractional part, or 0
	typ       byte // presentation type, or 0
}

// parseFormatSpec parses spec for an object of type typeName whose default
// presentation type and alignment are defaultType and defaultAlign.
func parseFormatSpec(spec, typeName string, defaultType, defaultAlign byte) (*formatSpec, error) {
	fs := &formatSpec{fill: ' ', width: -1, precision: -1, typ: defaultType}
	runes := []rune(spec)
	n := len(runes)
	pos := 0

	isAlign := func(r rune) bool { return r == '<' || r == '>' || r == '^' || r == '=' }
	fillSpecified, alignSpecified := false, false
	if n >= 2 && isAlign(runes[1]) {
		fs.fill = runes[0]
		fs.align = byte(runes[1])
		fillSpecified, alignSpecified = true, true
		pos = 2
	} else if n >= 1 && isAlign(runes[0]) {
		fs.align = byte(runes[0])
		alignSpecified = true
		pos = 1
	}
	if pos < n && (runes[pos] == '+' || runes[pos] == '-' || runes[pos] == ' ') {
		fs.sign = byte(runes[pos])
		pos++
	}
	if pos < n && runes[pos] == 'z' {
		fs.noNegZero = true
		pos++
	}
	if pos < n && runes[pos] == '#' {
		fs.alternate = true
		pos++
	}
	if !fillSpecified && pos < n && runes[pos] == '0' {
		fs.fill = '0'
		if !alignSpecified && defaultAlign == '>' {
			fs.align = '='
		}
		pos++
	}
	if width, ok := parseSpecInt(runes, &pos); ok {
		fs.width = width
	}

	bothSeparators := fmt.Errorf("ValueError: Cannot specify both ',' and '_'.")
	if pos < n && runes[pos] == ',' {
		fs.grouping = ','
		pos++
	}
	if pos < n && runes[pos] == '_' {
		if fs.grouping != 0 {
			return nil, bothSeparators
		}
		fs.grouping = '_'
		pos++
	}
	if pos < n && runes[pos] == ',' && fs.grouping == '_' {
		return nil, bothSeparators
	}

	if pos < n && runes[pos] == '.' {
		pos++
		precision, ok := parseSpecInt(runes, &pos)
		if ok {
			fs.precision = precision
		}
		if pos < n && runes[pos] == ',' {
			fs.fracGroup = ','
			pos++
		}
		if pos < n && runes[pos] == '_' {
			if fs.fracGroup != 0 {
				return nil, bothSeparators
			}
			fs.fracGroup = '_'
			pos++
		}
		if pos < n && runes[pos] == ',' && fs.fracGroup == '_' {
			return nil, bothSeparators
		}
		if !ok && fs.fracGroup == 0 {
			return nil, fmt.Errorf("ValueError: Format specifier missing precision")
		}
	}

	if n-pos > 1 {
		return nil, fmt.Errorf("ValueError: Invalid format specifier '%s' for object of type '%s'", spec, typeName)
	}
	if n-pos == 1 {
		if runes[pos] >= utf8.RuneSelf {
			return nil, fmt.Errorf("ValueError: Unknown format code '\\x%x' for object of type '%s'", runes[pos], typeName)
		}
		fs.typ = byte(runes[pos])
	}

	if fs.grouping != 0 {
		switch fs.typ {
		case 'd', 'e', 'f', 'g', 'E', 'G', '%', 'F', 0:
		case 'b', 'o', 'x', 'X':
			if fs.grouping != '_' {
				return nil, invalidSeparator(fs.grouping, fs.typ)
			}
		default:
			return nil, invalidSeparator(fs.grouping, fs.typ)
		}
	}
	return fs, nil
}

// parseSpecInt parses a run of decimal digits, reporting whether any were found.
func parseSpecInt(runes []rune, pos *int) (int, bool) {
	start := *pos
	v := 0
	for *pos < len(runes) && runes[*pos] >= '0' && runes[*pos] <= '9' {
		v = v*10 + int(runes[*pos]-'0')
		if v > math.MaxInt32 {
			v = math.MaxInt32
		}
		*pos++
	}
	return v, *pos > start
}

func invalidSeparator(sep, typ byte) error {
	return fmt.Errorf("ValueError: Cannot specify '%c' with %s.", sep, quoteFormatCode(typ))
}

func unknownFormatCode(typ byte, typeName string) error {
	return fmt.Errorf("ValueError: Unknown format code %s for object of type '%s'", quoteFormatCode(typ), typeName)
}

func quoteFormatCode(typ byte) string {
	if typ > ' ' && typ < 0x7f {
		return "'" + string(typ) + "'"
	}
	return fmt.Sprintf("'\\x%x'", typ)
}

// pad pads lead+body to the spec's width. lead holds the sign and prefix of
// a number; '=' alignment inserts the padding between lead and body.
func (fs *formatSpec) pad(lead, body string, defaultAlign byte) string {
	n := utf8.RuneCountInString(lead) + utf8.RuneCountInString(body)
	if fs.width <= n {
		return lead + body
	}
	padding := fs.width - n
	fill := string(fs.fill)
	align := fs.align
	if align == 0 {
		align = defaultAlign
	}
	switch align {
	case '<':
		return lead + body + strings.Repeat(fill, padding)
	case '^':
		left := padding / 2
		return strings.Repeat(fill, left) + lead + body + strings.Repeat(fill, padding-left)
	case '=':
		return lead + strings.Repeat(fill, padding) + body
	default:
		return strings.Repeat(fill, padding) + lead + body
	}
}

// signText returns the sign character(s) for a number under sign option opt.
func signText(negative bool, opt byte) string {
	switch {
	case negative:
		return "-"
	case opt == '+':
		return "+"
	case opt == ' ':
		return " "
	}
	return ""
}

// layoutNumber assembles sign, prefix (e.g. "0x"), integer digits and rest
// (fraction, exponent, '%' suffix), applying digit grouping and padding.
// When zero-padding with grouping, the padding zeros are grouped too.
func (fs *formatSpec) layoutNumber(negative bool, prefix, intDigits, rest string, groupSize int) string {
	sign := signText(negative, fs.sign)
	if fs.grouping != 0 && intDigits != "" {
		minWidth := 0
		if fs.fill == '0' && fs.align == '=' && fs.width > 0 {
			minWidth = fs.width - len(sign) - len(prefix) - utf8.RuneCountInString(rest)
		}
		intDigits = groupDigits(intDigits, fs.grouping, groupSize, minWidth)
	}
	return fs.pad(sign+prefix, intDigits+rest, '>')
}

// groupDigits inserts sep between groups of size digits, counting from the
// right. If minWidth is positive the result is left-padded with zeros (which
// are grouped as well) to at least that width, like CPython's
// _PyUnicode_InsertThousandsGrouping.
func groupDigits(digits string, sep byte, size, minWidth int) string {
	var groups []string
	remaining := len(digits)
	for {
		l := min(size, max(remaining, minWidth, 1))
		zeros := max(0, l-remaining)
		chars := max(0, min(remaining, l))
		groups = append(groups, strings.Repeat("0", zeros)+digits[remaining-chars:remaining])
		remaining -= chars
		minWidth -= l
		if remaining <= 0 && minWidth <= 0 {
			break
		}
		minWidth-- // room for the separator
	}
	var b strings.Builder
	for i := len(groups) - 1; i >= 0; i-- {
		b.WriteString(groups[i])
		if i > 0 {
			b.WriteByte(sep)
		}
	}
	return b.String()
}

// groupFraction groups the fractional digits that follow the decimal point
// in rest (Python 3.14 "." precision grouping), counting from the left.
func groupFraction(rest string, sep byte) string {
	if sep == 0 || !strings.HasPrefix(rest, ".") {
		return rest
	}
	end := 1
	for end < len(rest) && rest[end] >= '0' && rest[end] <= '9' {
		end++
	}
	frac := rest[1:end]
	var b strings.Builder
	b.WriteByte('.')
	for i := 0; i < len(frac); i += 3 {
		if i > 0 {
			b.WriteByte(sep)
		}
		b.WriteString(frac[i:min(i+3, len(frac))])
	}
	b.WriteString(rest[end:])
	return b.String()
}

// splitNumber splits formatted number text into its sign, leading integer
// digits and the remainder.
func splitNumber(s string) (negative bool, intDigits, rest string) {
	if strings.HasPrefix(s, "-") {
		negative = true
		s = s[1:]
	}
	i := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	return negative, s[:i], s[i:]
}

// =====================================
// Per-type formatting
// =====================================

// formatValue formats val according to spec, like format(val, spec).
func (vm *VM) formatValue(val Value, spec string) (string, error) {
	switch v := val.(type) {
	case *PyString:
		return formatStringSpec(v.Value, spec)
	case *PyBool:
		if spec == "" {
			return vm.strValue(v)
		}
		n := int64(0)
		if v.Value {
			n = 1
		}
		return formatIntSpec(big.NewInt(n), spec, "bool")
	case *PyInt:
		if spec == "" {
			return intToDecimal(v)
		}
		return formatIntSpec(v.BigIntValue(), spec, "int")
	case *PyFloat:
		return formatFloatSpec(v.Value, spec)
	case *PyComplex:
		return formatComplexSpec(v.Real, v.Imag, spec)
	case *PyInstance:
		if result, found, err := vm.callDunder(v, "__format__", &PyString{Value: spec}); found {
			if err != nil {
				return "", err
			}
			if s, ok := result.(*PyString); ok {
				return s.Value, nil
			}
			return "", fmt.Errorf("TypeError: __format__ must return a str, not %s", vm.typeName(result))
		}
	case *PyUserData:
		if v.Metatable != nil {
			if method, err := vm.getAttr(v, "__format__"); err == nil {
				result, err := vm.call(method, []Value{&PyString{Value: spec}}, nil)
				if err != nil {
					return "", err
				}
				if s, ok := result.(*PyString); ok {
					return s.Value, nil
				}
				return "", fmt.Errorf("TypeError: __format__ must return a str, not %s", vm.typeName(result))
			}
		}
	}
	// object.__format__: only the empty spec is supported.
	if spec != "" {
		return "", fmt.Errorf("TypeError: unsupported format string passed to %s.__format__", vm.typeName(val))
	}
	return vm.strValue(val)
}

// formatStringSpec implements str.__format__.
func formatStringSpec(s, spec string) (string, error) {
	if spec == "" {
		return s, nil
	}
	fs, err := parseFormatSpec(spec, "str", 's', '<')
	if err != nil {
		return "", err
	}
	if fs.typ != 's' {
		return "", unknownFormatCode(fs.typ, "str")
	}
	switch {
	case fs.sign == ' ':
		return "", fmt.Errorf("ValueError: Space not allowed in string format specifier")
	case fs.sign != 0:
		return "", fmt.Errorf("ValueError: Sign not allowed in string format specifier")
	case fs.noNegZero:
		return "", fmt.Errorf("ValueError: Negative zero coercion (z) not allowed in string format specifier")
	case fs.alternate:
		return "", fmt.Errorf("ValueError: Alternate form (#) not allowed in string format specifier")
	case fs.align == '=':
		return "", fmt.Errorf("ValueError: '=' alignment not allowed in string format specifier")
	}
	if fs.precision >= 0 && utf8.RuneCountInString(s) > fs.precision {
		s = string([]rune(s)[:fs.precision])
	}
	return fs.pad("", s, '<'), nil
}

// formatIntSpec implements int.__format__ for a non-empty spec.
func formatIntSpec(n *big.Int, spec, typeName string) (string, error) {
	fs, err := parseFormatSpec(spec, typeName, 'd', '>')
	if err != nil {
		return "", err
	}
	switch fs.typ {
	case 'b', 'c', 'd', 'o', 'x', 'X', 'n':
	case 'e', 'E', 'f', 'F', 'g', 'G', '%':
		f, acc := new(big.Float).SetInt(n).Float64()
		if math.IsInf(f, 0) && acc != big.Exact {
			return "", fmt.Errorf("OverflowError: int too large to convert to float")
		}
		return formatFloatWithSpec(f, fs, typeName)
	default:
		return "", unknownFormatCode(fs.typ, typeName)
	}
	if fs.precision >= 0 {
		return "", fmt.Errorf("ValueError: Precision not allowed in integer format specifier")
	}
	if fs.noNegZero {
		return "", fmt.Errorf("ValueError: Negative zero coercion (z) not allowed in integer format specifier")
	}
	if fs.typ == 'c' {
		if fs.sign != 0 {
			return "", fmt.Errorf("ValueError: Sign not allowed with integer format specifier 'c'")
		}
		if fs.alternate {
			return "", fmt.Errorf("ValueError: Alternate form (#) not allowed with integer format specifier 'c'")
		}
		if !n.IsInt64() || n.Int64() < 0 || n.Int64() > 0x10ffff {
			return "", fmt.Errorf("OverflowError: %%c arg not in range(0x110000)")
		}
		return fs.layoutNumber(false, "", "", string(rune(n.Int64())), 3), nil
	}

	base, prefix := 10, ""
	switch fs.typ {
	case 'b':
		base, prefix = 2, "0b"
	case 'o':
		base, prefix = 8, "0o"
	case 'x':
		base, prefix = 16, "0x"
	case 'X':
		base, prefix = 16, "0X"
	}
	if !fs.alternate {
		prefix = ""
	}
	digits := new(big.Int).Abs(n).Text(base)
	if fs.typ == 'X' {
		digits = strings.ToUpper(digits)
	}
	if base == 10 && len(digits) > maxIntStringDigits {
		return "", fmt.Errorf("ValueError: Exceeds the limit (%d digits) for integer string conversion; use sys.set_int_max_str_digits() to increase the limit", maxIntStringDigits)
	}
	groupSize := 3
	if base != 10 {
		groupSize = 4
	}
	return fs.layoutNumber(n.Sign() < 0, prefix, digits, "", groupSize), nil
}

// formatFloatSpec implements float.__format__.
func formatFloatSpec(x float64, spec string) (string, error) {
	fs, err := parseFormatSpec(spec, "float", 0, '>')
	if err != nil {
		return "", err
	}
	return formatFloatWithSpec(x, fs, "float")
}

func formatFloatWithSpec(x float64, fs *formatSpec, typeName string) (string, error) {
	typ := fs.typ
	switch typ {
	case 0, 'e', 'E', 'f', 'F', 'g', 'G', 'n', '%':
	default:
		return "", unknownFormatCode(typ, typeName)
	}
	flags := 0
	if fs.alternate {
		flags |= fmtFloatAlt
	}
	if fs.noNegZero {
		flags |= fmtFloatNoNegZero
	}
	precision, defaultPrecision := fs.precision, 6
	if typ == 0 {
		// Like repr() without a precision, otherwise like 'g' but always
		// with at least one digit after the decimal point.
		flags |= fmtFloatAddDot0
		typ = 'r'
		defaultPrecision = 0
	}
	if typ == 'n' {
		typ = 'g'
	}
	suffix := ""
	if typ == '%' {
		typ = 'f'
		x *= 100
		suffix = "%"
	}
	if precision < 0 {
		precision = defaultPrecision
	} else if typ == 'r' {
		typ = 'g'
	}
	negative, intDigits, rest := splitNumber(formatFloat(x, typ, precision, flags))
	rest = groupFraction(rest, fs.fracGroup)
	return fs.layoutNumber(negative, "", intDigits, rest+suffix, 3), nil
}

// formatComplexSpec implements complex.__format__.
func formatComplexSpec(re, im float64, spec string) (string, error) {
	if spec == "" {
		return formatComplex(re, im), nil
	}
	fs, err := parseFormatSpec(spec, "complex", 0, '>')
	if err != nil {
		return "", err
	}
	if fs.fill == '0' {
		return "", fmt.Errorf("ValueError: Zero padding is not allowed in complex format specifier")
	}
	if fs.align == '=' {
		return "", fmt.Errorf("ValueError: '=' alignment flag is not allowed in complex format specifier")
	}
	typ := fs.typ
	switch typ {
	case 0, 'e', 'E', 'f', 'F', 'g', 'G', 'n':
	default:
		return "", unknownFormatCode(typ, "complex")
	}
	flags := 0
	if fs.alternate {
		flags |= fmtFloatAlt
	}
	if fs.noNegZero {
		flags |= fmtFloatNoNegZero
	}
	precision, defaultPrecision := fs.precision, 6
	skipRe, addParens := false, false
	if typ == 0 {
		// Like str(): omit a +0 real part, otherwise add parentheses.
		typ = 'r'
		defaultPrecision = 0
		if re == 0 && !math.Signbit(re) {
			skipRe = true
		} else {
			addParens = true
		}
	}
	if typ == 'n' {
		typ = 'g'
	}
	if precision < 0 {
		precision = defaultPrecision
	} else if typ == 'r' {
		typ = 'g'
	}

	part := func(x float64, signOpt byte) string {
		negative, intDigits, rest := splitNumber(formatFloat(x, typ, precision, flags))
		if fs.grouping != 0 && intDigits != "" {
			intDigits = groupDigits(intDigits, fs.grouping, 3, 0)
		}
		return signText(negative, signOpt) + intDigits + groupFraction(rest, fs.fracGroup)
	}
	var body string
	if skipRe {
		body = part(im, fs.sign) + "j"
	} else {
		body = part(re, fs.sign) + part(im, '+') + "j"
	}
	if addParens {
		body = "(" + body + ")"
	}
	return fs.pad("", body, '>'), nil
}

// =====================================
// Float to text
// =====================================

// Flags for formatFloat, mirroring CPython's Py_DTSF_* flags.
const (
	fmtFloatSign      = 1 << iota // always emit a sign character
	fmtFloatAddDot0               // append ".0" when the result would look like an integer
	fmtFloatAlt                   // alternate form: keep the decimal point and trailing zeros
	fmtFloatNoNegZero             // drop the sign of a result that rounds to zero
)

// formatFloat converts x to text like CPython's PyOS_double_to_string. typ is
// one of 'e', 'E', 'f', 'F', 'g', 'G' or 'r' (shortest round-trip repr; prec is
// ignored).
func formatFloat(x float64, typ byte, prec int, flags int) string {
	upper := typ == 'E' || typ == 'F' || typ == 'G'
	if upper {
		typ += 'a' - 'A'
	}

	if math.IsNaN(x) || math.IsInf(x, 0) {
		word := "inf"
		if math.IsNaN(x) {
			word = "nan"
		}
		if upper {
			word = strings.ToUpper(word)
		}
		if math.IsInf(x, -1) {
			return "-" + word
		}
		if flags&fmtFloatSign != 0 {
			return "+" + word
		}
		return word
	}

	if typ == 'g' && prec == 0 {
		prec = 1
	}
	digits, decpt := floatDigits(math.Abs(x), typ, prec)
	alt := flags&fmtFloatAlt != 0
	addDot0 := flags&fmtFloatAddDot0 != 0

	// Choose between fixed and exponent notation and work out how many
	// digits (vEnd, relative to the start of digits) to show.
	useExp := false
	vEnd := len(digits)
	switch typ {
	case 'e':
		useExp = true
		vEnd = prec + 1
	case 'f':
		vEnd = decpt + prec
	case 'g':
		limit := prec
		if addDot0 {
			limit--
		}
		if decpt <= -4 || decpt > limit {
			useExp = true
		}
		if alt {
			vEnd = prec
		}
	case 'r':
		if decpt <= -4 || decpt > 16 {
			useExp = true
		}
	}
	exp := 0
	if useExp {
		exp = decpt - 1
		decpt = 1
	}
	vStart := 0
	if decpt <= 0 {
		vStart = decpt - 1
	}
	if !useExp && addDot0 {
		vEnd = max(vEnd, decpt+1)
	} else {
		vEnd = max(vEnd, decpt)
	}

	var b strings.Builder
	negative := math.Signbit(x)
	if negative && flags&fmtFloatNoNegZero != 0 && digits == "0" {
		negative = false
	}
	if negative {
		b.WriteByte('-')
	} else if flags&fmtFloatSign != 0 {
		b.WriteByte('+')
	}
	if decpt <= 0 {
		b.WriteString(strings.Repeat("0", decpt-vStart))
		b.WriteByte('.')
		b.WriteString(strings.Repeat("0", -decpt))
	}
	if decpt > 0 && decpt <= len(digits) {
		b.WriteString(digits[:decpt])
		b.WriteByte('.')
		b.WriteString(digits[decpt:])
	} else {
		b.WriteString(digits)
	}
	if len(digits) < decpt {
		b.WriteString(strings.Repeat("0", decpt-len(digits)))
		b.WriteByte('.')
		b.WriteString(strings.Repeat("0", vEnd-decpt))
	} else {
		b.WriteString(strings.Repeat("0", vEnd-len(digits)))
	}

	s := b.String()
	if !alt && strings.HasSuffix(s, ".") {
		s = s[:len(s)-1]
	}
	if useExp {
		e := byte('e')
		if upper {
			e = 'E'
		}
		s += string(e) + fmt.Sprintf("%+03d", exp)
	}
	return s
}

// floatDigits returns the significant decimal digits of abs (without trailing
// zeros, "0" for zero) and the position of the decimal point relative to the
// first digit, correctly rounded for typ and prec.
func floatDigits(abs float64, typ byte, prec int) (string, int) {
	if typ == 'f' {
		s := strconv.FormatFloat(abs, 'f', prec, 64)
		intPart, frac, _ := strings.Cut(s, ".")
		all := intPart + frac
		trimmed := strings.TrimLeft(all, "0")
		if trimmed == "" {
			return "0", 1
		}
		decpt := len(intPart) - (len(all) - len(trimmed))
		return strings.TrimRight(trimmed, "0"), decpt
	}
	var s string
	switch typ {
	case 'e':
		s = strconv.FormatFloat(abs, 'e', prec, 64)
	case 'g':
		s = strconv.FormatFloat(abs, 'e', prec-1, 64)
	default:
		s = strconv.FormatFloat(abs, 'e', -1, 64)
	}
	mant, expText, _ := strings.Cut(s, "e")
	exp, _ := strconv.Atoi(expText)
	digits := strings.TrimRight(strings.Replace(mant, ".", "", 1), "0")
	if digits == "" {
		return "0", 1
	}
	return digits, exp + 1
}
