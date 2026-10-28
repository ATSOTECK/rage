package runtime

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// String methods that work on characters (code points) rather than bytes and
// follow CPython's argument handling: optional start/end arguments behave like
// slice bounds, and predicates use Unicode properties.

// isPySpace reports whether r is whitespace for str.split()/strip()/isspace().
// Python also treats the ASCII separators \x1c-\x1f as whitespace.
func isPySpace(r rune) bool {
	return unicode.IsSpace(r) || (r >= 0x1c && r <= 0x1f)
}

// isLineBreak reports whether r ends a line for str.splitlines().
func isLineBreak(r rune) bool {
	switch r {
	case '\n', '\r', '\v', '\f', 0x1c, 0x1d, 0x1e, 0x85, 0x2028, 0x2029:
		return true
	}
	return false
}

// isCased reports whether r has case (upper, lower or title).
func isCased(r rune) bool {
	return unicode.IsUpper(r) || unicode.IsLower(r) || unicode.IsTitle(r)
}

// digitRanges lists characters with Unicode Numeric_Type=Digit that are not
// decimal digits (superscripts, subscripts, circled digits, ...), which
// str.isdigit() accepts.
var digitRanges = &unicode.RangeTable{
	R16: []unicode.Range16{
		{Lo: 0x00b2, Hi: 0x00b3, Stride: 1},
		{Lo: 0x00b9, Hi: 0x00b9, Stride: 1},
		{Lo: 0x1369, Hi: 0x1371, Stride: 1},
		{Lo: 0x19da, Hi: 0x19da, Stride: 1},
		{Lo: 0x2070, Hi: 0x2070, Stride: 1},
		{Lo: 0x2074, Hi: 0x2079, Stride: 1},
		{Lo: 0x2080, Hi: 0x2089, Stride: 1},
		{Lo: 0x2460, Hi: 0x2468, Stride: 1},
		{Lo: 0x2474, Hi: 0x247c, Stride: 1},
		{Lo: 0x2488, Hi: 0x2490, Stride: 1},
		{Lo: 0x24ea, Hi: 0x24ea, Stride: 1},
		{Lo: 0x24f5, Hi: 0x24fd, Stride: 1},
		{Lo: 0x24ff, Hi: 0x24ff, Stride: 1},
		{Lo: 0x2776, Hi: 0x277e, Stride: 1},
		{Lo: 0x2780, Hi: 0x2788, Stride: 1},
		{Lo: 0x278a, Hi: 0x2792, Stride: 1},
	},
	R32: []unicode.Range32{
		{Lo: 0x10a40, Hi: 0x10a43, Stride: 1},
		{Lo: 0x1f100, Hi: 0x1f10a, Stride: 1},
	},
}

// hanNumerals lists common CJK ideographs with a numeric value, which
// str.isnumeric() accepts.
var hanNumerals = map[rune]bool{
	'〇': true, '一': true, '二': true, '三': true, '四': true, '五': true, '六': true,
	'七': true, '八': true, '九': true, '十': true, '百': true, '千': true, '万': true,
	'億': true, '兆': true, '零': true, '壹': true, '贰': true, '貳': true, '叁': true,
	'參': true, '肆': true, '伍': true, '陆': true, '陸': true, '柒': true, '捌': true,
	'玖': true, '拾': true, '佰': true, '仟': true, '廿': true, '卅': true,
}

func isPyDecimal(r rune) bool { return unicode.Is(unicode.Nd, r) }
func isPyDigit(r rune) bool   { return isPyDecimal(r) || unicode.Is(digitRanges, r) }
func isPyNumeric(r rune) bool {
	return isPyDigit(r) || unicode.Is(unicode.Nl, r) || unicode.Is(unicode.No, r) || hanNumerals[r]
}
func isPyAlnum(r rune) bool { return unicode.IsLetter(r) || isPyNumeric(r) }

// upperFull and lowerFull apply the full (one-to-many) Unicode case mappings
// that Go's simple mappings miss.
var upperSpecial = map[rune]string{
	'ß': "SS", 'ŉ': "ʼN", 'ǰ': "J̌", 'ΐ': "Ϊ́", 'ΰ': "Ϋ́", 'և': "ԵՒ",
	'ﬀ': "FF", 'ﬁ': "FI", 'ﬂ': "FL", 'ﬃ': "FFI", 'ﬄ': "FFL", 'ﬅ': "ST", 'ﬆ': "ST",
}

func pyUpper(s string) string {
	var b strings.Builder
	for _, r := range s {
		if sp, ok := upperSpecial[r]; ok {
			b.WriteString(sp)
		} else {
			b.WriteRune(unicode.ToUpper(r))
		}
	}
	return b.String()
}

func pyLower(s string) string {
	runes := []rune(s)
	var b strings.Builder
	for i, r := range runes {
		switch {
		case r == 'İ':
			b.WriteString("i̇")
		case r == 'Σ' && isFinalSigma(runes, i):
			b.WriteRune('ς')
		default:
			b.WriteRune(unicode.ToLower(r))
		}
	}
	return b.String()
}

// isFinalSigma reports whether the capital sigma at runes[i] ends a word, in
// which case it lowercases to the final form ς.
func isFinalSigma(runes []rune, i int) bool {
	before := false
	for j := i - 1; j >= 0; j-- {
		if unicode.Is(unicode.Mn, runes[j]) || runes[j] == '\'' {
			continue
		}
		before = isCased(runes[j])
		break
	}
	if !before {
		return false
	}
	for j := i + 1; j < len(runes); j++ {
		if unicode.Is(unicode.Mn, runes[j]) || runes[j] == '\'' {
			continue
		}
		return !isCased(runes[j])
	}
	return true
}

// pyTitle implements str.title(): a character following a cased character is
// lowercased, any other is titlecased.
func pyTitle(s string) string {
	var b strings.Builder
	prevCased := false
	for _, r := range s {
		if prevCased {
			b.WriteString(pyLower(string(r)))
		} else if sp, ok := upperSpecial[r]; ok {
			// Titlecase of a special character: first letter upper, rest lower
			first, size := utf8.DecodeRuneInString(sp)
			b.WriteRune(first)
			b.WriteString(strings.ToLower(sp[size:]))
		} else {
			b.WriteRune(unicode.ToTitle(r))
		}
		prevCased = isCased(r)
	}
	return b.String()
}

// strSliceBounds resolves optional start/end arguments of find(), count(),
// startswith() and friends against a string of n characters, like CPython's
// ADJUST_INDICES: end is clamped to [0, n] and start to >= 0 (start may still
// exceed n, which makes most searches fail).
func (vm *VM) strSliceBounds(args []Value, n int) (int, int, error) {
	start, end := 0, n
	get := func(v Value, def int) (int, error) {
		if v == nil || v == None {
			return def, nil
		}
		return vm.sliceBound(v, def)
	}
	var err error
	if len(args) > 0 {
		if start, err = get(args[0], 0); err != nil {
			return 0, 0, err
		}
	}
	if len(args) > 1 {
		if end, err = get(args[1], n); err != nil {
			return 0, 0, err
		}
	}
	if end > n {
		end = n
	} else if end < 0 {
		end += n
		if end < 0 {
			end = 0
		}
	}
	if start < 0 {
		start += n
		if start < 0 {
			start = 0
		}
	}
	return start, end, nil
}

// runeSlice returns the characters [i, j) of s, where runes is []rune(s) or
// nil if s is ASCII.
func runeSlice(s string, runes []rune, i, j int) string {
	if runes == nil {
		return s[i:j]
	}
	return string(runes[i:j])
}

// strRunes returns the length of s in characters, plus its runes when s is not
// pure ASCII (nil otherwise, so byte offsets can be used directly).
func strRunes(s string) (int, []rune) {
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			r := []rune(s)
			return len(r), r
		}
	}
	return len(s), nil
}

// strArg requires v to be a str.
func (vm *VM) strArg(v Value, what string) (string, error) {
	if s, ok := v.(*PyString); ok {
		return s.Value, nil
	}
	return "", fmt.Errorf("TypeError: %s must be str, not %s", what, vm.typeName(v))
}

// fillCharArg validates the fill character of center/ljust/rjust.
func (vm *VM) fillCharArg(args []Value) (string, error) {
	if len(args) < 2 {
		return " ", nil
	}
	s, ok := args[1].(*PyString)
	if !ok {
		return "", fmt.Errorf("TypeError: The fill character must be a unicode character, not %s", vm.typeName(args[1]))
	}
	if utf8.RuneCountInString(s.Value) != 1 {
		return "", fmt.Errorf("TypeError: The fill character must be exactly one character long")
	}
	return s.Value, nil
}

// splitWhitespace splits on runs of whitespace (str.split() with no sep).
func splitWhitespace(s string, maxsplit int) []string {
	runes := []rune(s)
	n := len(runes)
	var parts []string
	i := 0
	for maxsplit != 0 {
		for i < n && isPySpace(runes[i]) {
			i++
		}
		if i == n {
			break
		}
		j := i
		for i < n && !isPySpace(runes[i]) {
			i++
		}
		parts = append(parts, string(runes[j:i]))
		maxsplit--
	}
	if i < n {
		// Only reached when maxsplit ran out: the rest, minus leading space
		for i < n && isPySpace(runes[i]) {
			i++
		}
		if i < n {
			parts = append(parts, string(runes[i:]))
		}
	}
	return parts
}

// rsplitWhitespace splits on whitespace from the right.
func rsplitWhitespace(s string, maxsplit int) []string {
	runes := []rune(s)
	var parts []string
	i := len(runes) - 1
	for maxsplit != 0 {
		for i >= 0 && isPySpace(runes[i]) {
			i--
		}
		if i < 0 {
			break
		}
		j := i
		for i >= 0 && !isPySpace(runes[i]) {
			i--
		}
		parts = append(parts, string(runes[i+1:j+1]))
		maxsplit--
	}
	if i >= 0 {
		for i >= 0 && isPySpace(runes[i]) {
			i--
		}
		if i >= 0 {
			parts = append(parts, string(runes[:i+1]))
		}
	}
	for l, r := 0, len(parts)-1; l < r; l, r = l+1, r-1 {
		parts[l], parts[r] = parts[r], parts[l]
	}
	return parts
}

// rsplitSep splits on sep from the right, at most maxsplit times.
func rsplitSep(s, sep string, maxsplit int) []string {
	var parts []string
	for maxsplit != 0 {
		idx := strings.LastIndex(s, sep)
		if idx < 0 {
			break
		}
		parts = append(parts, s[idx+len(sep):])
		s = s[:idx]
		maxsplit--
	}
	parts = append(parts, s)
	for l, r := 0, len(parts)-1; l < r; l, r = l+1, r-1 {
		parts[l], parts[r] = parts[r], parts[l]
	}
	return parts
}

func stringList(parts []string) *PyList {
	items := make([]Value, len(parts))
	for i, p := range parts {
		items[i] = &PyString{Value: p}
	}
	return &PyList{Items: items}
}

func pyBool(b bool) Value {
	if b {
		return True
	}
	return False
}

// strMethod returns the bound str method name, or nil if it is not handled
// here.
func (vm *VM) strMethod(str *PyString, name string) Value {
	s := str.Value
	method := func(fn func(args []Value, kwargs map[string]Value) (Value, error)) Value {
		return &PyBuiltinFunc{Name: "str." + name, Fn: fn}
	}
	noKwargs := func(kwargs map[string]Value) error {
		if len(kwargs) > 0 {
			return fmt.Errorf("TypeError: str.%s() takes no keyword arguments", name)
		}
		return nil
	}
	argRange := func(args []Value, min, max int) error {
		if len(args) < min || len(args) > max {
			if min == max {
				return fmt.Errorf("TypeError: str.%s() takes exactly %d argument(s) (%d given)", name, min, len(args))
			}
			return fmt.Errorf("TypeError: %s expected at most %d arguments, got %d", name, max, len(args))
		}
		return nil
	}
	// predicate builds an is*() method: true if s is non-empty and every
	// character satisfies ok
	predicate := func(ok func(rune) bool) Value {
		return method(func(args []Value, kwargs map[string]Value) (Value, error) {
			if s == "" {
				return False, nil
			}
			for _, r := range s {
				if !ok(r) {
					return False, nil
				}
			}
			return True, nil
		})
	}
	// splitArgs reads (sep=None, maxsplit=-1)
	splitArgs := func(args []Value, kwargs map[string]Value) (sep string, hasSep bool, maxsplit int, err error) {
		maxsplit = -1
		if len(args) > 2 {
			return "", false, 0, fmt.Errorf("TypeError: %s() takes at most 2 arguments (%d given)", name, len(args))
		}
		sepVal, maxVal := Value(None), Value(nil)
		if len(args) > 0 {
			sepVal = args[0]
		}
		if len(args) > 1 {
			maxVal = args[1]
		}
		for k, v := range kwargs {
			switch k {
			case "sep":
				sepVal = v
			case "maxsplit":
				maxVal = v
			default:
				return "", false, 0, fmt.Errorf("TypeError: %s() got an unexpected keyword argument '%s'", name, k)
			}
		}
		if sepVal != None {
			sv, ok := sepVal.(*PyString)
			if !ok {
				return "", false, 0, fmt.Errorf("TypeError: must be str or None, not %s", vm.typeName(sepVal))
			}
			if sv.Value == "" {
				return "", false, 0, fmt.Errorf("ValueError: empty separator")
			}
			sep, hasSep = sv.Value, true
		}
		if maxVal != nil {
			m, err := vm.getIntIndex(maxVal)
			if err != nil {
				return "", false, 0, err
			}
			maxsplit = int(m)
		}
		return sep, hasSep, maxsplit, nil
	}
	// stripChars reads the optional chars argument of strip/lstrip/rstrip
	stripChars := func(args []Value) (func(rune) bool, error) {
		if len(args) > 1 {
			return nil, fmt.Errorf("TypeError: %s expected at most 1 argument, got %d", name, len(args))
		}
		if len(args) == 0 || args[0] == None {
			return isPySpace, nil
		}
		chars, ok := args[0].(*PyString)
		if !ok {
			return nil, fmt.Errorf("TypeError: %s arg must be None or str", name)
		}
		return func(r rune) bool { return strings.ContainsRune(chars.Value, r) }, nil
	}
	// search implements find/rfind/index/rindex
	search := func(last, raise bool) Value {
		return method(func(args []Value, kwargs map[string]Value) (Value, error) {
			if err := noKwargs(kwargs); err != nil {
				return nil, err
			}
			if err := argRange(args, 1, 3); err != nil {
				return nil, err
			}
			sub, err := vm.strArg(args[0], "argument")
			if err != nil {
				return nil, err
			}
			n, runes := strRunes(s)
			start, end, err := vm.strSliceBounds(args[1:], n)
			if err != nil {
				return nil, err
			}
			idx := -1
			if end-start >= utf8.RuneCountInString(sub) {
				hay := runeSlice(s, runes, start, end)
				var b int
				if last {
					b = strings.LastIndex(hay, sub)
				} else {
					b = strings.Index(hay, sub)
				}
				if b >= 0 {
					idx = start + utf8.RuneCountInString(hay[:b])
				}
			}
			if idx < 0 && raise {
				return nil, fmt.Errorf("ValueError: substring not found")
			}
			return MakeInt(int64(idx)), nil
		})
	}
	// affix implements startswith/endswith
	affix := func(suffix bool) Value {
		return method(func(args []Value, kwargs map[string]Value) (Value, error) {
			if err := noKwargs(kwargs); err != nil {
				return nil, err
			}
			if err := argRange(args, 1, 3); err != nil {
				return nil, err
			}
			n, runes := strRunes(s)
			start, end, err := vm.strSliceBounds(args[1:], n)
			if err != nil {
				return nil, err
			}
			match := func(v Value) (bool, error) {
				a, ok := v.(*PyString)
				if !ok {
					kind := "startswith"
					if suffix {
						kind = "endswith"
					}
					return false, fmt.Errorf("TypeError: %s first arg must be str or a tuple of str, not %s", kind, vm.typeName(v))
				}
				alen := utf8.RuneCountInString(a.Value)
				if start > n || end-start < alen {
					return false, nil
				}
				if suffix {
					return runeSlice(s, runes, end-alen, end) == a.Value, nil
				}
				return runeSlice(s, runes, start, start+alen) == a.Value, nil
			}
			if t, ok := args[0].(*PyTuple); ok {
				for _, item := range t.Items {
					m, err := match(item)
					if err != nil {
						return nil, err
					}
					if m {
						return True, nil
					}
				}
				return False, nil
			}
			m, err := match(args[0])
			if err != nil {
				return nil, err
			}
			return pyBool(m), nil
		})
	}
	// pad implements center/ljust/rjust
	pad := func(kind string) Value {
		return method(func(args []Value, kwargs map[string]Value) (Value, error) {
			if err := noKwargs(kwargs); err != nil {
				return nil, err
			}
			if err := argRange(args, 1, 2); err != nil {
				return nil, err
			}
			w, err := vm.getIntIndex(args[0])
			if err != nil {
				return nil, err
			}
			fill, err := vm.fillCharArg(args)
			if err != nil {
				return nil, err
			}
			width := int(w)
			n := utf8.RuneCountInString(s)
			if n >= width {
				return str, nil
			}
			total := width - n
			left := 0
			switch kind {
			case "center":
				// CPython's split, which favours the left for odd widths
				left = total/2 + (total & width & 1)
			case "rjust":
				left = total
			}
			return &PyString{Value: strings.Repeat(fill, left) + s + strings.Repeat(fill, total-left)}, nil
		})
	}

	switch name {
	case "upper":
		return method(func(args []Value, kwargs map[string]Value) (Value, error) {
			return &PyString{Value: pyUpper(s)}, nil
		})
	case "lower":
		return method(func(args []Value, kwargs map[string]Value) (Value, error) {
			return &PyString{Value: pyLower(s)}, nil
		})
	case "casefold":
		return method(func(args []Value, kwargs map[string]Value) (Value, error) {
			folded := strings.NewReplacer("ß", "ss", "ẞ", "ss", "ς", "σ").Replace(pyLower(s))
			return &PyString{Value: folded}, nil
		})
	case "title":
		return method(func(args []Value, kwargs map[string]Value) (Value, error) {
			return &PyString{Value: pyTitle(s)}, nil
		})
	case "capitalize":
		return method(func(args []Value, kwargs map[string]Value) (Value, error) {
			if s == "" {
				return str, nil
			}
			first, size := utf8.DecodeRuneInString(s)
			head := pyTitle(string(first))
			return &PyString{Value: head + pyLower(s[size:])}, nil
		})
	case "swapcase":
		return method(func(args []Value, kwargs map[string]Value) (Value, error) {
			var b strings.Builder
			for _, r := range s {
				switch {
				case unicode.IsUpper(r):
					b.WriteString(pyLower(string(r)))
				case unicode.IsLower(r):
					b.WriteString(pyUpper(string(r)))
				default:
					b.WriteRune(r)
				}
			}
			return &PyString{Value: b.String()}, nil
		})
	case "split", "rsplit":
		return method(func(args []Value, kwargs map[string]Value) (Value, error) {
			sep, hasSep, maxsplit, err := splitArgs(args, kwargs)
			if err != nil {
				return nil, err
			}
			switch {
			case !hasSep && name == "split":
				return stringList(splitWhitespace(s, maxsplit)), nil
			case !hasSep:
				return stringList(rsplitWhitespace(s, maxsplit)), nil
			case name == "split":
				if maxsplit < 0 {
					return stringList(strings.Split(s, sep)), nil
				}
				return stringList(strings.SplitN(s, sep, maxsplit+1)), nil
			}
			return stringList(rsplitSep(s, sep, maxsplit)), nil
		})
	case "splitlines":
		return method(func(args []Value, kwargs map[string]Value) (Value, error) {
			keepends := false
			if len(args) > 0 {
				keepends = vm.truthy(args[0])
			}
			if v, ok := kwargs["keepends"]; ok {
				keepends = vm.truthy(v)
			}
			runes := []rune(s)
			var lines []string
			start := 0
			for i := 0; i < len(runes); i++ {
				if !isLineBreak(runes[i]) {
					continue
				}
				end := i
				if runes[i] == '\r' && i+1 < len(runes) && runes[i+1] == '\n' {
					i++
				}
				if keepends {
					end = i + 1
				}
				lines = append(lines, string(runes[start:end]))
				start = i + 1
			}
			if start < len(runes) {
				lines = append(lines, string(runes[start:]))
			}
			return stringList(lines), nil
		})
	case "join":
		return method(func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("TypeError: str.join() takes exactly one argument (%d given)", len(args))
			}
			items, err := vm.toList(args[0])
			if err != nil {
				return nil, fmt.Errorf("TypeError: can only join an iterable")
			}
			parts := make([]string, len(items))
			for i, item := range items {
				p, ok := item.(*PyString)
				if !ok {
					return nil, fmt.Errorf("TypeError: sequence item %d: expected str instance, %s found", i, vm.typeName(item))
				}
				parts[i] = p.Value
			}
			return &PyString{Value: strings.Join(parts, s)}, nil
		})
	case "strip", "lstrip", "rstrip":
		return method(func(args []Value, kwargs map[string]Value) (Value, error) {
			strip, err := stripChars(args)
			if err != nil {
				return nil, err
			}
			r := s
			if name != "rstrip" {
				r = strings.TrimLeftFunc(r, strip)
			}
			if name != "lstrip" {
				r = strings.TrimRightFunc(r, strip)
			}
			return &PyString{Value: r}, nil
		})
	case "replace":
		return method(func(args []Value, kwargs map[string]Value) (Value, error) {
			if c, ok := kwargs["count"]; ok && len(args) == 2 {
				args = append(args, c)
			}
			if err := argRange(args, 2, 3); err != nil {
				return nil, err
			}
			old, err := vm.strArg(args[0], "replace() argument 1")
			if err != nil {
				return nil, err
			}
			repl, err := vm.strArg(args[1], "replace() argument 2")
			if err != nil {
				return nil, err
			}
			count := -1
			if len(args) > 2 {
				c, err := vm.getIntIndex(args[2])
				if err != nil {
					return nil, err
				}
				count = int(c)
			}
			return &PyString{Value: strings.Replace(s, old, repl, count)}, nil
		})
	case "find":
		return search(false, false)
	case "rfind":
		return search(true, false)
	case "index":
		return search(false, true)
	case "rindex":
		return search(true, true)
	case "count":
		return method(func(args []Value, kwargs map[string]Value) (Value, error) {
			if err := noKwargs(kwargs); err != nil {
				return nil, err
			}
			if err := argRange(args, 1, 3); err != nil {
				return nil, err
			}
			sub, err := vm.strArg(args[0], "argument")
			if err != nil {
				return nil, err
			}
			n, runes := strRunes(s)
			start, end, err := vm.strSliceBounds(args[1:], n)
			if err != nil {
				return nil, err
			}
			subLen := utf8.RuneCountInString(sub)
			if end-start < subLen {
				return MakeInt(0), nil
			}
			if subLen == 0 {
				return MakeInt(int64(end - start + 1)), nil
			}
			return MakeInt(int64(strings.Count(runeSlice(s, runes, start, end), sub))), nil
		})
	case "startswith":
		return affix(false)
	case "endswith":
		return affix(true)
	case "center", "ljust", "rjust":
		return pad(name)
	case "zfill":
		return method(func(args []Value, kwargs map[string]Value) (Value, error) {
			if err := argRange(args, 1, 1); err != nil {
				return nil, err
			}
			w, err := vm.getIntIndex(args[0])
			if err != nil {
				return nil, err
			}
			n := utf8.RuneCountInString(s)
			if n >= int(w) {
				return str, nil
			}
			zeros := strings.Repeat("0", int(w)-n)
			if s != "" && (s[0] == '+' || s[0] == '-') {
				return &PyString{Value: s[:1] + zeros + s[1:]}, nil
			}
			return &PyString{Value: zeros + s}, nil
		})
	case "expandtabs":
		return method(func(args []Value, kwargs map[string]Value) (Value, error) {
			tabsize := int64(8)
			var tv Value
			if len(args) > 0 {
				tv = args[0]
			} else if v, ok := kwargs["tabsize"]; ok {
				tv = v
			}
			if tv != nil {
				t, err := vm.getIntIndex(tv)
				if err != nil {
					return nil, err
				}
				tabsize = t
			}
			var b strings.Builder
			col := int64(0)
			for _, r := range s {
				switch {
				case r == '\t':
					if tabsize > 0 {
						spaces := tabsize - col%tabsize
						b.WriteString(strings.Repeat(" ", int(spaces)))
						col += spaces
					}
				case r == '\n' || r == '\r':
					b.WriteRune(r)
					col = 0
				default:
					b.WriteRune(r)
					col++
				}
			}
			return &PyString{Value: b.String()}, nil
		})
	case "partition", "rpartition":
		return method(func(args []Value, kwargs map[string]Value) (Value, error) {
			if err := argRange(args, 1, 1); err != nil {
				return nil, err
			}
			sep, err := vm.strArg(args[0], "argument")
			if err != nil {
				return nil, err
			}
			if sep == "" {
				return nil, fmt.Errorf("ValueError: empty separator")
			}
			var idx int
			if name == "partition" {
				idx = strings.Index(s, sep)
			} else {
				idx = strings.LastIndex(s, sep)
			}
			empty := &PyString{Value: ""}
			if idx < 0 {
				if name == "partition" {
					return &PyTuple{Items: []Value{str, empty, empty}}, nil
				}
				return &PyTuple{Items: []Value{empty, empty, str}}, nil
			}
			return &PyTuple{Items: []Value{
				&PyString{Value: s[:idx]},
				&PyString{Value: sep},
				&PyString{Value: s[idx+len(sep):]},
			}}, nil
		})
	case "isalpha":
		return predicate(unicode.IsLetter)
	case "isdecimal":
		return predicate(isPyDecimal)
	case "isdigit":
		return predicate(isPyDigit)
	case "isnumeric":
		return predicate(isPyNumeric)
	case "isalnum":
		return predicate(isPyAlnum)
	case "isspace":
		return predicate(isPySpace)
	case "isupper", "islower":
		return method(func(args []Value, kwargs map[string]Value) (Value, error) {
			cased := false
			for _, r := range s {
				if name == "isupper" && (unicode.IsLower(r) || unicode.IsTitle(r)) ||
					name == "islower" && (unicode.IsUpper(r) || unicode.IsTitle(r)) {
					return False, nil
				}
				if isCased(r) {
					cased = true
				}
			}
			return pyBool(cased), nil
		})
	case "isidentifier":
		return method(func(args []Value, kwargs map[string]Value) (Value, error) {
			if s == "" {
				return False, nil
			}
			for i, r := range s {
				start := r == '_' || unicode.IsLetter(r) || unicode.Is(unicode.Nl, r)
				if i == 0 && !start {
					return False, nil
				}
				if !start && !unicode.In(r, unicode.Mn, unicode.Mc, unicode.Nd, unicode.Pc) {
					return False, nil
				}
			}
			return True, nil
		})
	case "encode":
		return method(func(args []Value, kwargs map[string]Value) (Value, error) {
			encoding, errors := "utf-8", "strict"
			if len(args) > 0 {
				encoding = vm.str(args[0])
			}
			if len(args) > 1 {
				errors = vm.str(args[1])
			}
			if v, ok := kwargs["encoding"]; ok {
				encoding = vm.str(v)
			}
			if v, ok := kwargs["errors"]; ok {
				errors = vm.str(v)
			}
			b, err := encodeString(s, encoding, errors)
			if err != nil {
				return nil, err
			}
			return &PyBytes{Value: b}, nil
		})
	}
	return nil
}

// encodeString implements str.encode for UTF-8, ASCII and Latin-1.
func encodeString(s, encoding, errors string) ([]byte, error) {
	norm := strings.ReplaceAll(strings.ToLower(encoding), "_", "-")
	var limit rune
	switch norm {
	case "utf-8", "utf8", "u8":
		return []byte(s), nil
	case "ascii", "us-ascii":
		limit, norm = 0x80, "ascii"
	case "latin-1", "latin1", "iso-8859-1", "iso8859-1", "l1":
		limit, norm = 0x100, "latin-1"
	default:
		return nil, fmt.Errorf("LookupError: unknown encoding: %s", encoding)
	}
	out := make([]byte, 0, len(s))
	pos := 0
	for _, r := range s {
		if r < limit {
			out = append(out, byte(r))
		} else {
			switch errors {
			case "ignore":
			case "replace":
				out = append(out, '?')
			default:
				esc := fmt.Sprintf("\\x%02x", r)
				if r > 0xffff {
					esc = fmt.Sprintf("\\U%08x", r)
				} else if r > 0xff {
					esc = fmt.Sprintf("\\u%04x", r)
				}
				return nil, fmt.Errorf("UnicodeEncodeError: '%s' codec can't encode character '%s' in position %d: ordinal not in range(%d)", norm, esc, pos, limit)
			}
		}
		pos++
	}
	return out, nil
}

// decodeBytes implements bytes.decode for UTF-8, ASCII and Latin-1.
func decodeBytes(data []byte, encoding, errors string) (string, error) {
	norm := strings.ReplaceAll(strings.ToLower(encoding), "_", "-")
	switch norm {
	case "utf-8", "utf8", "u8":
		norm = "utf-8"
	case "ascii", "us-ascii":
		norm = "ascii"
	case "latin-1", "latin1", "iso-8859-1", "iso8859-1", "l1":
		var b strings.Builder
		for _, c := range data {
			b.WriteRune(rune(c))
		}
		return b.String(), nil
	default:
		return "", fmt.Errorf("LookupError: unknown encoding: %s", encoding)
	}
	var b strings.Builder
	for i := 0; i < len(data); {
		r, size := rune(data[i]), 1
		valid := data[i] < 0x80
		if !valid && norm == "utf-8" {
			r, size = utf8.DecodeRune(data[i:])
			valid = r != utf8.RuneError || size > 1
		}
		if valid {
			b.WriteRune(r)
			i += size
			continue
		}
		switch errors {
		case "ignore":
		case "replace":
			b.WriteRune(utf8.RuneError)
		default:
			reason := "ordinal not in range(128)"
			if norm == "utf-8" {
				reason = "invalid start byte"
				if data[i] >= 0xc2 && data[i] <= 0xf4 {
					if i+1 >= len(data) {
						reason = "unexpected end of data"
					} else {
						reason = "invalid continuation byte"
					}
				}
			}
			return "", fmt.Errorf("UnicodeDecodeError: '%s' codec can't decode byte 0x%02x in position %d: %s", norm, data[i], i, reason)
		}
		i++
	}
	return b.String(), nil
}

// bytesFromHex implements bytes.fromhex: pairs of hex digits, optionally
// separated by ASCII whitespace.
func bytesFromHex(s string) (Value, error) {
	hexVal := func(c byte) int {
		switch {
		case c >= '0' && c <= '9':
			return int(c - '0')
		case c >= 'a' && c <= 'f':
			return int(c-'a') + 10
		case c >= 'A' && c <= 'F':
			return int(c-'A') + 10
		}
		return -1
	}
	isSpace := func(c byte) bool {
		return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\v' || c == '\f'
	}
	var out []byte
	for i := 0; i < len(s); {
		if isSpace(s[i]) {
			i++
			continue
		}
		hi := hexVal(s[i])
		if hi < 0 {
			return nil, fmt.Errorf("ValueError: non-hexadecimal number found in fromhex() arg at position %d", i)
		}
		if i+1 >= len(s) || hexVal(s[i+1]) < 0 {
			return nil, fmt.Errorf("ValueError: non-hexadecimal number found in fromhex() arg at position %d", i+1)
		}
		out = append(out, byte(hi<<4|hexVal(s[i+1])))
		i += 2
	}
	if out == nil {
		out = []byte{}
	}
	return &PyBytes{Value: out}, nil
}
