package runtime

import (
	"fmt"
	"strings"
	"unicode"
)

// getAttrString handles attribute access on *PyString values.
func (vm *VM) getAttrString(str *PyString, name string) (Value, error) {
	switch name {
	case "upper":
		return &PyBuiltinFunc{Name: "str.upper", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			return &PyString{Value: strings.ToUpper(str.Value)}, nil
		}}, nil
	case "lower":
		return &PyBuiltinFunc{Name: "str.lower", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			return &PyString{Value: strings.ToLower(str.Value)}, nil
		}}, nil
	case "split":
		return &PyBuiltinFunc{Name: "str.split", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			var strParts []string
			if len(args) == 0 {
				strParts = strings.Fields(str.Value)
			} else {
				sep := vm.str(args[0])
				maxSplit := -1
				if len(args) > 1 {
					maxSplit = int(vm.toInt(args[1]))
				}
				if maxSplit < 0 {
					strParts = strings.Split(str.Value, sep)
				} else {
					strParts = strings.SplitN(str.Value, sep, maxSplit+1)
				}
			}
			parts := make([]Value, len(strParts))
			for i, s := range strParts {
				parts[i] = &PyString{Value: s}
			}
			return &PyList{Items: parts}, nil
		}}, nil
	case "rsplit":
		return &PyBuiltinFunc{Name: "str.rsplit", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) == 0 {
				strParts := strings.Fields(str.Value)
				parts := make([]Value, len(strParts))
				for i, s := range strParts {
					parts[i] = &PyString{Value: s}
				}
				return &PyList{Items: parts}, nil
			}
			sep := vm.str(args[0])
			maxSplit := -1
			if len(args) > 1 {
				maxSplit = int(vm.toInt(args[1]))
			}
			if maxSplit < 0 {
				strParts := strings.Split(str.Value, sep)
				parts := make([]Value, len(strParts))
				for i, s := range strParts {
					parts[i] = &PyString{Value: s}
				}
				return &PyList{Items: parts}, nil
			}
			// rsplit from right
			s := str.Value
			var result []string
			for maxSplit > 0 {
				idx := strings.LastIndex(s, sep)
				if idx < 0 {
					break
				}
				result = append([]string{s[idx+len(sep):]}, result...)
				s = s[:idx]
				maxSplit--
			}
			result = append([]string{s}, result...)
			parts := make([]Value, len(result))
			for i, p := range result {
				parts[i] = &PyString{Value: p}
			}
			return &PyList{Items: parts}, nil
		}}, nil
	case "join":
		return &PyBuiltinFunc{Name: "str.join", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("join() takes exactly 1 argument")
			}
			items, err := vm.toList(args[0])
			if err != nil {
				return nil, err
			}
			var parts []string
			for _, item := range items {
				s, ok := item.(*PyString)
				if !ok {
					return nil, fmt.Errorf("TypeError: sequence item: expected str instance, %s found", vm.typeName(item))
				}
				parts = append(parts, s.Value)
			}
			return &PyString{Value: strings.Join(parts, str.Value)}, nil
		}}, nil
	case "strip":
		return &PyBuiltinFunc{Name: "str.strip", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) > 0 {
				chars := vm.str(args[0])
				return &PyString{Value: strings.Trim(str.Value, chars)}, nil
			}
			return &PyString{Value: strings.TrimSpace(str.Value)}, nil
		}}, nil
	case "lstrip":
		return &PyBuiltinFunc{Name: "str.lstrip", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) > 0 {
				chars := vm.str(args[0])
				return &PyString{Value: strings.TrimLeft(str.Value, chars)}, nil
			}
			return &PyString{Value: strings.TrimLeftFunc(str.Value, func(r rune) bool {
				return r == ' ' || r == '\t' || r == '\n' || r == '\r'
			})}, nil
		}}, nil
	case "rstrip":
		return &PyBuiltinFunc{Name: "str.rstrip", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) > 0 {
				chars := vm.str(args[0])
				return &PyString{Value: strings.TrimRight(str.Value, chars)}, nil
			}
			return &PyString{Value: strings.TrimRightFunc(str.Value, func(r rune) bool {
				return r == ' ' || r == '\t' || r == '\n' || r == '\r'
			})}, nil
		}}, nil
	case "replace":
		return &PyBuiltinFunc{Name: "str.replace", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("replace() takes at least 2 arguments")
			}
			old := vm.str(args[0])
			newStr := vm.str(args[1])
			count := -1
			if len(args) > 2 {
				count = int(vm.toInt(args[2]))
			}
			return &PyString{Value: strings.Replace(str.Value, old, newStr, count)}, nil
		}}, nil
	case "find":
		return &PyBuiltinFunc{Name: "str.find", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("find() takes at least 1 argument")
			}
			sub := vm.str(args[0])
			s := str.Value
			start := 0
			if len(args) > 1 {
				start = int(vm.toInt(args[1]))
				if start < 0 {
					start += len([]rune(s))
					if start < 0 {
						start = 0
					}
				}
			}
			if start > len(s) {
				return MakeInt(-1), nil
			}
			idx := strings.Index(s[start:], sub)
			if idx < 0 {
				return MakeInt(-1), nil
			}
			return MakeInt(int64(start + idx)), nil
		}}, nil
	case "rfind":
		return &PyBuiltinFunc{Name: "str.rfind", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("rfind() takes at least 1 argument")
			}
			sub := vm.str(args[0])
			idx := strings.LastIndex(str.Value, sub)
			return MakeInt(int64(idx)), nil
		}}, nil
	case "index":
		return &PyBuiltinFunc{Name: "str.index", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("index() takes at least 1 argument")
			}
			sub := vm.str(args[0])
			idx := strings.Index(str.Value, sub)
			if idx < 0 {
				return nil, fmt.Errorf("ValueError: substring not found")
			}
			return MakeInt(int64(idx)), nil
		}}, nil
	case "rindex":
		return &PyBuiltinFunc{Name: "str.rindex", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("rindex() takes at least 1 argument")
			}
			sub := vm.str(args[0])
			idx := strings.LastIndex(str.Value, sub)
			if idx < 0 {
				return nil, fmt.Errorf("ValueError: substring not found")
			}
			return MakeInt(int64(idx)), nil
		}}, nil
	case "startswith":
		return &PyBuiltinFunc{Name: "str.startswith", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("startswith() takes at least 1 argument")
			}
			// Handle tuple of prefixes
			if t, ok := args[0].(*PyTuple); ok {
				for _, item := range t.Items {
					prefix := vm.str(item)
					if strings.HasPrefix(str.Value, prefix) {
						return True, nil
					}
				}
				return False, nil
			}
			prefix := vm.str(args[0])
			if strings.HasPrefix(str.Value, prefix) {
				return True, nil
			}
			return False, nil
		}}, nil
	case "endswith":
		return &PyBuiltinFunc{Name: "str.endswith", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("endswith() takes at least 1 argument")
			}
			if t, ok := args[0].(*PyTuple); ok {
				for _, item := range t.Items {
					suffix := vm.str(item)
					if strings.HasSuffix(str.Value, suffix) {
						return True, nil
					}
				}
				return False, nil
			}
			suffix := vm.str(args[0])
			if strings.HasSuffix(str.Value, suffix) {
				return True, nil
			}
			return False, nil
		}}, nil
	case "count":
		return &PyBuiltinFunc{Name: "str.count", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("count() takes at least 1 argument")
			}
			sub := vm.str(args[0])
			s := str.Value
			start := 0
			end := len(s)
			if len(args) > 1 {
				start = int(vm.toInt(args[1]))
			}
			if len(args) > 2 {
				end = int(vm.toInt(args[2]))
			}
			if start > len(s) {
				return MakeInt(0), nil
			}
			if end > len(s) {
				end = len(s)
			}
			return MakeInt(int64(strings.Count(s[start:end], sub))), nil
		}}, nil
	case "center":
		return &PyBuiltinFunc{Name: "str.center", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("center() takes at least 1 argument")
			}
			width := int(vm.toInt(args[0]))
			fillChar := " "
			if len(args) > 1 {
				fillChar = vm.str(args[1])
			}
			s := str.Value
			if len(s) >= width {
				return &PyString{Value: s}, nil
			}
			total := width - len(s)
			left := total / 2
			right := total - left
			return &PyString{Value: strings.Repeat(fillChar, left) + s + strings.Repeat(fillChar, right)}, nil
		}}, nil
	case "ljust":
		return &PyBuiltinFunc{Name: "str.ljust", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("ljust() takes at least 1 argument")
			}
			width := int(vm.toInt(args[0]))
			fillChar := " "
			if len(args) > 1 {
				fillChar = vm.str(args[1])
			}
			s := str.Value
			if len(s) >= width {
				return &PyString{Value: s}, nil
			}
			return &PyString{Value: s + strings.Repeat(fillChar, width-len(s))}, nil
		}}, nil
	case "rjust":
		return &PyBuiltinFunc{Name: "str.rjust", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("rjust() takes at least 1 argument")
			}
			width := int(vm.toInt(args[0]))
			fillChar := " "
			if len(args) > 1 {
				fillChar = vm.str(args[1])
			}
			s := str.Value
			if len(s) >= width {
				return &PyString{Value: s}, nil
			}
			return &PyString{Value: strings.Repeat(fillChar, width-len(s)) + s}, nil
		}}, nil
	case "zfill":
		return &PyBuiltinFunc{Name: "str.zfill", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("zfill() takes exactly 1 argument")
			}
			width := int(vm.toInt(args[0]))
			s := str.Value
			if len(s) >= width {
				return &PyString{Value: s}, nil
			}
			if len(s) > 0 && (s[0] == '+' || s[0] == '-') {
				return &PyString{Value: string(s[0]) + strings.Repeat("0", width-len(s)) + s[1:]}, nil
			}
			return &PyString{Value: strings.Repeat("0", width-len(s)) + s}, nil
		}}, nil
	case "expandtabs":
		return &PyBuiltinFunc{Name: "str.expandtabs", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			tabSize := 8
			if len(args) > 0 {
				tabSize = int(vm.toInt(args[0]))
			}
			var result strings.Builder
			col := 0
			for _, ch := range str.Value {
				if ch == '\t' {
					spaces := tabSize - (col % tabSize)
					result.WriteString(strings.Repeat(" ", spaces))
					col += spaces
				} else if ch == '\n' || ch == '\r' {
					result.WriteRune(ch)
					col = 0
				} else {
					result.WriteRune(ch)
					col++
				}
			}
			return &PyString{Value: result.String()}, nil
		}}, nil
	case "partition":
		return &PyBuiltinFunc{Name: "str.partition", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("partition() takes exactly 1 argument")
			}
			sep := vm.str(args[0])
			idx := strings.Index(str.Value, sep)
			if idx < 0 {
				return &PyTuple{Items: []Value{
					&PyString{Value: str.Value},
					&PyString{Value: ""},
					&PyString{Value: ""},
				}}, nil
			}
			return &PyTuple{Items: []Value{
				&PyString{Value: str.Value[:idx]},
				&PyString{Value: sep},
				&PyString{Value: str.Value[idx+len(sep):]},
			}}, nil
		}}, nil
	case "rpartition":
		return &PyBuiltinFunc{Name: "str.rpartition", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("rpartition() takes exactly 1 argument")
			}
			sep := vm.str(args[0])
			idx := strings.LastIndex(str.Value, sep)
			if idx < 0 {
				return &PyTuple{Items: []Value{
					&PyString{Value: ""},
					&PyString{Value: ""},
					&PyString{Value: str.Value},
				}}, nil
			}
			return &PyTuple{Items: []Value{
				&PyString{Value: str.Value[:idx]},
				&PyString{Value: sep},
				&PyString{Value: str.Value[idx+len(sep):]},
			}}, nil
		}}, nil
	case "title":
		return &PyBuiltinFunc{Name: "str.title", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			return &PyString{Value: strings.Title(str.Value)}, nil
		}}, nil
	case "swapcase":
		return &PyBuiltinFunc{Name: "str.swapcase", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			var result strings.Builder
			for _, ch := range str.Value {
				if ch >= 'a' && ch <= 'z' {
					result.WriteRune(ch - 32)
				} else if ch >= 'A' && ch <= 'Z' {
					result.WriteRune(ch + 32)
				} else {
					result.WriteRune(ch)
				}
			}
			return &PyString{Value: result.String()}, nil
		}}, nil
	case "capitalize":
		return &PyBuiltinFunc{Name: "str.capitalize", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			s := str.Value
			if len(s) == 0 {
				return &PyString{Value: ""}, nil
			}
			runes := []rune(s)
			result := strings.ToUpper(string(runes[0:1])) + strings.ToLower(string(runes[1:]))
			return &PyString{Value: result}, nil
		}}, nil
	case "isalpha":
		return &PyBuiltinFunc{Name: "str.isalpha", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			s := str.Value
			if len(s) == 0 {
				return False, nil
			}
			for _, ch := range s {
				if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')) {
					return False, nil
				}
			}
			return True, nil
		}}, nil
	case "isdigit":
		return &PyBuiltinFunc{Name: "str.isdigit", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			s := str.Value
			if len(s) == 0 {
				return False, nil
			}
			for _, ch := range s {
				if ch < '0' || ch > '9' {
					return False, nil
				}
			}
			return True, nil
		}}, nil
	case "isalnum":
		return &PyBuiltinFunc{Name: "str.isalnum", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			s := str.Value
			if len(s) == 0 {
				return False, nil
			}
			for _, ch := range s {
				if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9')) {
					return False, nil
				}
			}
			return True, nil
		}}, nil
	case "isspace":
		return &PyBuiltinFunc{Name: "str.isspace", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			s := str.Value
			if len(s) == 0 {
				return False, nil
			}
			for _, ch := range s {
				if ch != ' ' && ch != '\t' && ch != '\n' && ch != '\r' && ch != '\f' && ch != '\v' {
					return False, nil
				}
			}
			return True, nil
		}}, nil
	case "isupper":
		return &PyBuiltinFunc{Name: "str.isupper", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			s := str.Value
			hasUpper := false
			for _, ch := range s {
				if ch >= 'a' && ch <= 'z' {
					return False, nil
				}
				if ch >= 'A' && ch <= 'Z' {
					hasUpper = true
				}
			}
			if hasUpper {
				return True, nil
			}
			return False, nil
		}}, nil
	case "islower":
		return &PyBuiltinFunc{Name: "str.islower", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			s := str.Value
			hasLower := false
			for _, ch := range s {
				if ch >= 'A' && ch <= 'Z' {
					return False, nil
				}
				if ch >= 'a' && ch <= 'z' {
					hasLower = true
				}
			}
			if hasLower {
				return True, nil
			}
			return False, nil
		}}, nil
	case "format":
		return &PyBuiltinFunc{Name: "str.format", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			return vm.strFormat(str.Value, args, kwargs)
		}}, nil
	case "splitlines":
		return &PyBuiltinFunc{Name: "str.splitlines", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			keepends := false
			if len(args) > 0 {
				keepends = vm.truthy(args[0])
			}
			s := str.Value
			if len(s) == 0 {
				return &PyList{Items: []Value{}}, nil
			}
			var lines []Value
			start := 0
			for i := 0; i < len(s); i++ {
				if s[i] == '\n' || s[i] == '\r' {
					end := i
					if s[i] == '\r' && i+1 < len(s) && s[i+1] == '\n' {
						i++
					}
					if keepends {
						lines = append(lines, &PyString{Value: s[start : i+1]})
					} else {
						lines = append(lines, &PyString{Value: s[start:end]})
					}
					start = i + 1
				}
			}
			if start < len(s) {
				lines = append(lines, &PyString{Value: s[start:]})
			}
			return &PyList{Items: lines}, nil
		}}, nil
	case "encode":
		return &PyBuiltinFunc{Name: "str.encode", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			return &PyBytes{Value: []byte(str.Value)}, nil
		}}, nil

	case "casefold":
		return &PyBuiltinFunc{Name: "str.casefold", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			// casefold is like lower() but more aggressive for caseless matching
			// Go's strings.ToLower handles most Unicode casefolding
			// For full CPython compatibility we'd need special cases (e.g. ss -> ss)
			// but strings.ToLower is a good approximation
			result := []rune{}
			for _, r := range str.Value {
				// Handle special casefold cases
				if r == '\u00df' || r == '\u1e9e' {
					result = append(result, 's', 's')
				} else {
					result = append(result, []rune(strings.ToLower(string(r)))...)
				}
			}
			return &PyString{Value: string(result)}, nil
		}}, nil

	case "isascii":
		return &PyBuiltinFunc{Name: "str.isascii", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			// Empty string returns True (CPython behavior)
			for _, r := range str.Value {
				if r > 127 {
					return False, nil
				}
			}
			return True, nil
		}}, nil

	case "isdecimal":
		return &PyBuiltinFunc{Name: "str.isdecimal", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(str.Value) == 0 {
				return False, nil
			}
			for _, r := range str.Value {
				// Decimal digits are Unicode category Nd with digit value
				if r < '0' || r > '9' {
					// Check for other Unicode decimal digits
					if !unicode.Is(unicode.Nd, r) {
						return False, nil
					}
				}
			}
			return True, nil
		}}, nil

	case "isnumeric":
		return &PyBuiltinFunc{Name: "str.isnumeric", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(str.Value) == 0 {
				return False, nil
			}
			for _, r := range str.Value {
				// Numeric includes digits, fractions, subscripts, superscripts, etc.
				if !unicode.Is(unicode.Nd, r) && !unicode.Is(unicode.Nl, r) && !unicode.Is(unicode.No, r) {
					return False, nil
				}
			}
			return True, nil
		}}, nil

	case "isidentifier":
		return &PyBuiltinFunc{Name: "str.isidentifier", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(str.Value) == 0 {
				return False, nil
			}
			for i, r := range str.Value {
				if i == 0 {
					if r != '_' && !unicode.IsLetter(r) {
						return False, nil
					}
				} else {
					if r != '_' && !unicode.IsLetter(r) && !unicode.IsDigit(r) {
						return False, nil
					}
				}
			}
			return True, nil
		}}, nil

	case "isprintable":
		return &PyBuiltinFunc{Name: "str.isprintable", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			// Empty string returns True (CPython behavior)
			for _, r := range str.Value {
				if !unicode.IsPrint(r) {
					return False, nil
				}
			}
			return True, nil
		}}, nil

	case "istitle":
		return &PyBuiltinFunc{Name: "str.istitle", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(str.Value) == 0 {
				return False, nil
			}
			// A titlecased string has uppercase after uncased chars
			// and lowercase after cased chars
			prevCased := false
			hasCased := false
			for _, r := range str.Value {
				if unicode.IsUpper(r) || unicode.IsTitle(r) {
					if prevCased {
						return False, nil
					}
					prevCased = true
					hasCased = true
				} else if unicode.IsLower(r) {
					if !prevCased {
						return False, nil
					}
					prevCased = true
					hasCased = true
				} else {
					prevCased = false
				}
			}
			if !hasCased {
				return False, nil
			}
			return True, nil
		}}, nil

	case "removeprefix":
		return &PyBuiltinFunc{Name: "str.removeprefix", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("TypeError: removeprefix() takes exactly one argument (%d given)", len(args))
			}
			prefix, ok := args[0].(*PyString)
			if !ok {
				return nil, fmt.Errorf("TypeError: removeprefix arg must be str, not '%s'", vm.typeName(args[0]))
			}
			if strings.HasPrefix(str.Value, prefix.Value) {
				return &PyString{Value: str.Value[len(prefix.Value):]}, nil
			}
			return str, nil
		}}, nil

	case "removesuffix":
		return &PyBuiltinFunc{Name: "str.removesuffix", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("TypeError: removesuffix() takes exactly one argument (%d given)", len(args))
			}
			suffix, ok := args[0].(*PyString)
			if !ok {
				return nil, fmt.Errorf("TypeError: removesuffix arg must be str, not '%s'", vm.typeName(args[0]))
			}
			if suffix.Value != "" && strings.HasSuffix(str.Value, suffix.Value) {
				return &PyString{Value: str.Value[:len(str.Value)-len(suffix.Value)]}, nil
			}
			return str, nil
		}}, nil

	case "format_map":
		return &PyBuiltinFunc{Name: "str.format_map", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("TypeError: format_map() takes exactly one argument (%d given)", len(args))
			}
			mapping, ok := args[0].(*PyDict)
			if !ok {
				return nil, fmt.Errorf("TypeError: format_map() argument must be a mapping, not '%s'", vm.typeName(args[0]))
			}
			// Build format args from the mapping
			template := str.Value
			var result strings.Builder
			i := 0
			for i < len(template) {
				if template[i] == '{' {
					if i+1 < len(template) && template[i+1] == '{' {
						result.WriteByte('{')
						i += 2
						continue
					}
					j := i + 1
					for j < len(template) && template[j] != '}' {
						j++
					}
					if j >= len(template) {
						return nil, fmt.Errorf("ValueError: Single '{' encountered in format string")
					}
					field := template[i+1 : j]
					var fieldName, formatSpec string
					if colonIdx := strings.Index(field, ":"); colonIdx >= 0 {
						fieldName = field[:colonIdx]
						formatSpec = field[colonIdx+1:]
					} else {
						fieldName = field
					}
					val, found := mapping.DictGet(&PyString{Value: fieldName}, vm)
					if !found {
						return nil, fmt.Errorf("KeyError: '%s'", fieldName)
					}
					if formatSpec != "" {
						formatted, err := vm.formatValue(val, formatSpec)
						if err != nil {
							return nil, err
						}
						result.WriteString(formatted)
					} else {
						result.WriteString(vm.str(val))
					}
					i = j + 1
				} else if template[i] == '}' {
					if i+1 < len(template) && template[i+1] == '}' {
						result.WriteByte('}')
						i += 2
						continue
					}
					return nil, fmt.Errorf("ValueError: Single '}' encountered in format string")
				} else {
					result.WriteByte(template[i])
					i++
				}
			}
			return &PyString{Value: result.String()}, nil
		}}, nil

	case "maketrans":
		return &PyBuiltinFunc{Name: "str.maketrans", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			return strMaketransImpl(vm, args)
		}}, nil

	case "translate":
		return &PyBuiltinFunc{Name: "str.translate", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("TypeError: translate() takes exactly one argument (%d given)", len(args))
			}
			table, ok := args[0].(*PyDict)
			if !ok {
				return nil, fmt.Errorf("TypeError: translate() argument must be a dict")
			}
			var result strings.Builder
			for _, r := range str.Value {
				key := MakeInt(int64(r))
				if val, found := table.DictGet(key, vm); found {
					if val == None {
						// Delete the character
						continue
					}
					switch v := val.(type) {
					case *PyString:
						result.WriteString(v.Value)
					case *PyInt:
						result.WriteRune(rune(v.Value))
					default:
						result.WriteRune(r)
					}
				} else {
					result.WriteRune(r)
				}
			}
			return &PyString{Value: result.String()}, nil
		}}, nil
	}
	return nil, fmt.Errorf("AttributeError: 'str' object has no attribute '%s'", name)
}

// strMaketransImpl implements str.maketrans(x[, y[, z]]).
func strMaketransImpl(vm *VM, args []Value) (Value, error) {
	result := &PyDict{Items: make(map[Value]Value), buckets: make(map[uint64][]dictEntry)}
	if len(args) == 1 {
		d, ok := args[0].(*PyDict)
		if !ok {
			return nil, fmt.Errorf("TypeError: if you give only one argument to maketrans it must be a dict")
		}
		for _, key := range d.Keys(vm) {
			val, _ := d.DictGet(key, vm)
			var intKey Value
			switch k := key.(type) {
			case *PyString:
				runes := []rune(k.Value)
				if len(runes) != 1 {
					return nil, fmt.Errorf("ValueError: string keys in translate table must be of length 1, found %d", len(runes))
				}
				intKey = MakeInt(int64(runes[0]))
			case *PyInt:
				intKey = k
			default:
				return nil, fmt.Errorf("TypeError: keys in translate table must be strings or integers")
			}
			switch v := val.(type) {
			case *PyString:
				runes := []rune(v.Value)
				if len(runes) == 1 {
					result.DictSet(intKey, MakeInt(int64(runes[0])), vm)
				} else {
					result.DictSet(intKey, val, vm)
				}
			case *PyInt:
				result.DictSet(intKey, val, vm)
			case *PyNone:
				result.DictSet(intKey, None, vm)
			default:
				result.DictSet(intKey, val, vm)
			}
		}
		return result, nil
	}
	if len(args) < 2 {
		return nil, fmt.Errorf("TypeError: maketrans requires 1 or 2-3 string arguments")
	}
	x, ok1 := args[0].(*PyString)
	y, ok2 := args[1].(*PyString)
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("TypeError: maketrans arguments must be strings")
	}
	xRunes := []rune(x.Value)
	yRunes := []rune(y.Value)
	if len(xRunes) != len(yRunes) {
		return nil, fmt.Errorf("ValueError: the first two maketrans arguments must have equal length")
	}
	for i, r := range xRunes {
		result.DictSet(MakeInt(int64(r)), MakeInt(int64(yRunes[i])), vm)
	}
	if len(args) >= 3 {
		z, ok := args[2].(*PyString)
		if !ok {
			return nil, fmt.Errorf("TypeError: maketrans third argument must be a string")
		}
		for _, r := range z.Value {
			result.DictSet(MakeInt(int64(r)), None, vm)
		}
	}
	return result, nil
}
