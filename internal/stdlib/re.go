package stdlib

import (
	"regexp"
	"strings"

	"github.com/ATSOTECK/RAGE/internal/runtime"
)

// PyPattern represents a compiled regular expression
type PyPattern struct {
	Pattern string
	Regex   *regexp.Regexp
	Flags   int
}

func (p *PyPattern) Type() string   { return "re.Pattern" }
func (p *PyPattern) String() string { return "re.compile('" + p.Pattern + "')" }

// PyMatch represents a regex match result
type PyMatch struct {
	Pattern *PyPattern
	Input   string
	Start   int
	End     int
	Groups  []string
}

func (m *PyMatch) Type() string   { return "re.Match" }
func (m *PyMatch) String() string { return "<re.Match object>" }

// Regex flags
const (
	IGNORECASE = 2
	MULTILINE  = 8
	DOTALL     = 16
	VERBOSE    = 64
	ASCII      = 256
)

// InitReModule registers the re module
func InitReModule() {
	// Register Pattern type methods
	patternMT := &runtime.TypeMetatable{
		Name: "re.Pattern",
		Methods: map[string]runtime.GoFunction{
			"match":    patternMatch,
			"search":   patternSearch,
			"findall":  patternFindall,
			"finditer": patternFinditer,
			"sub":      patternSub,
			"subn":     patternSubn,
			"split":    patternSplit,
		},
	}
	runtime.RegisterTypeMetatable("re.Pattern", patternMT)

	// Register Match type methods
	matchMT := &runtime.TypeMetatable{
		Name: "re.Match",
		Methods: map[string]runtime.GoFunction{
			"group":  matchGroup,
			"groups": matchGroups,
			"start":  matchStart,
			"end":    matchEnd,
			"span":   matchSpan,
		},
	}
	runtime.RegisterTypeMetatable("re.Match", matchMT)

	runtime.NewModuleBuilder("re").
		Doc("Regular expression operations.").
		// Flags
		Const("IGNORECASE", runtime.NewInt(IGNORECASE)).
		Const("I", runtime.NewInt(IGNORECASE)).
		Const("MULTILINE", runtime.NewInt(MULTILINE)).
		Const("M", runtime.NewInt(MULTILINE)).
		Const("DOTALL", runtime.NewInt(DOTALL)).
		Const("S", runtime.NewInt(DOTALL)).
		Const("VERBOSE", runtime.NewInt(VERBOSE)).
		Const("X", runtime.NewInt(VERBOSE)).
		Const("ASCII", runtime.NewInt(ASCII)).
		Const("A", runtime.NewInt(ASCII)).
		// Functions
		Func("compile", reCompile).
		Func("match", reMatch).
		Func("search", reSearch).
		Func("findall", reFindall).
		Func("finditer", reFinditer).
		Func("sub", reSub).
		Func("subn", reSubn).
		Func("split", reSplit).
		Func("escape", reEscape).
		Register()
}

// RegisterTypeMetatable registers a type metatable globally
func init() {
	// This will be called when the package is imported
}

// compilePattern compiles a pattern with optional flags
func compilePattern(pattern string, flags int) (*PyPattern, error) {
	// Build regex pattern with flags
	regexPattern := pattern

	// Apply flags
	prefix := ""
	if flags&IGNORECASE != 0 {
		prefix += "(?i)"
	}
	if flags&MULTILINE != 0 {
		prefix += "(?m)"
	}
	if flags&DOTALL != 0 {
		prefix += "(?s)"
	}

	regexPattern = prefix + regexPattern

	re, err := regexp.Compile(regexPattern)
	if err != nil {
		return nil, err
	}

	return &PyPattern{
		Pattern: pattern,
		Regex:   re,
		Flags:   flags,
	}, nil
}

// re.compile(pattern[, flags]) -> Pattern
func reCompile(vm *runtime.VM) int {
	pattern := vm.CheckString(1)
	flags := 0
	if vm.GetTop() >= 2 {
		flags = int(vm.ToInt(2))
	}

	compiled, err := compilePattern(pattern, flags)
	if err != nil {
		vm.RaiseError("error in regular expression: %s", err.Error())
		return 0
	}

	ud := runtime.NewUserData(compiled)
	ud.Metatable = runtime.NewDict()
	ud.Metatable.Items[runtime.NewString("__type__")] = runtime.NewString("re.Pattern")
	vm.Push(ud)
	return 1
}

// createMatch creates a Match object from regex match
func createMatch(pattern *PyPattern, str string, loc []int, groups []string) runtime.Value {
	if loc == nil {
		return runtime.None
	}

	match := &PyMatch{
		Pattern: pattern,
		Input:   str,
		Start:   loc[0],
		End:     loc[1],
		Groups:  groups,
	}

	ud := runtime.NewUserData(match)
	ud.Metatable = runtime.NewDict()
	ud.Metatable.Items[runtime.NewString("__type__")] = runtime.NewString("re.Match")
	return ud
}

// re.match(pattern, string[, flags]) -> Match or None
func reMatch(vm *runtime.VM) int {
	pattern := vm.CheckString(1)
	str := vm.CheckString(2)
	flags := 0
	if vm.GetTop() >= 3 {
		flags = int(vm.ToInt(3))
	}

	compiled, err := compilePattern("^"+pattern, flags)
	if err != nil {
		vm.RaiseError("error in regular expression: %s", err.Error())
		return 0
	}

	loc := compiled.Regex.FindStringIndex(str)
	if loc == nil {
		vm.Push(runtime.None)
		return 1
	}

	groups := compiled.Regex.FindStringSubmatch(str)
	vm.Push(createMatch(compiled, str, loc, groups))
	return 1
}

// re.search(pattern, string[, flags]) -> Match or None
func reSearch(vm *runtime.VM) int {
	pattern := vm.CheckString(1)
	str := vm.CheckString(2)
	flags := 0
	if vm.GetTop() >= 3 {
		flags = int(vm.ToInt(3))
	}

	compiled, err := compilePattern(pattern, flags)
	if err != nil {
		vm.RaiseError("error in regular expression: %s", err.Error())
		return 0
	}

	loc := compiled.Regex.FindStringIndex(str)
	if loc == nil {
		vm.Push(runtime.None)
		return 1
	}

	groups := compiled.Regex.FindStringSubmatch(str)
	vm.Push(createMatch(compiled, str, loc, groups))
	return 1
}

// re.findall(pattern, string[, flags]) -> list of strings
func reFindall(vm *runtime.VM) int {
	pattern := vm.CheckString(1)
	str := vm.CheckString(2)
	flags := 0
	if vm.GetTop() >= 3 {
		flags = int(vm.ToInt(3))
	}

	compiled, err := compilePattern(pattern, flags)
	if err != nil {
		vm.RaiseError("error in regular expression: %s", err.Error())
		return 0
	}

	matches := compiled.Regex.FindAllStringSubmatch(str, -1)
	result := make([]runtime.Value, 0, len(matches))

	for _, match := range matches {
		if len(match) == 1 {
			// No groups, return full match
			result = append(result, runtime.NewString(match[0]))
		} else if len(match) == 2 {
			// Single group, return just that group
			result = append(result, runtime.NewString(match[1]))
		} else {
			// Multiple groups, return tuple
			groups := make([]runtime.Value, len(match)-1)
			for i, g := range match[1:] {
				groups[i] = runtime.NewString(g)
			}
			result = append(result, runtime.NewTuple(groups))
		}
	}

	vm.Push(runtime.NewList(result))
	return 1
}

// re.finditer(pattern, string[, flags]) -> iterator of Match objects
func reFinditer(vm *runtime.VM) int {
	pattern := vm.CheckString(1)
	str := vm.CheckString(2)
	flags := 0
	if vm.GetTop() >= 3 {
		flags = int(vm.ToInt(3))
	}

	compiled, err := compilePattern(pattern, flags)
	if err != nil {
		vm.RaiseError("error in regular expression: %s", err.Error())
		return 0
	}

	allMatches := compiled.Regex.FindAllStringSubmatchIndex(str, -1)
	matches := make([]runtime.Value, 0, len(allMatches))

	for _, loc := range allMatches {
		groups := make([]string, 0)
		// Extract groups from indices
		for i := 0; i < len(loc); i += 2 {
			if loc[i] >= 0 && loc[i+1] >= 0 {
				groups = append(groups, str[loc[i]:loc[i+1]])
			} else {
				groups = append(groups, "")
			}
		}
		matches = append(matches, createMatch(compiled, str, loc[:2], groups))
	}

	vm.Push(runtime.NewList(matches))
	return 1
}

// re.sub(pattern, repl, string[, count, flags]) -> string
func reSub(vm *runtime.VM) int {
	pattern := vm.CheckString(1)
	repl := vm.CheckString(2)
	str := vm.CheckString(3)
	count := -1
	flags := 0

	if vm.GetTop() >= 4 {
		count = int(vm.ToInt(4))
	}
	if vm.GetTop() >= 5 {
		flags = int(vm.ToInt(5))
	}

	compiled, err := compilePattern(pattern, flags)
	if err != nil {
		vm.RaiseError("error in regular expression: %s", err.Error())
		return 0
	}

	var result string
	if count < 0 {
		result = compiled.Regex.ReplaceAllString(str, repl)
	} else {
		result = replaceN(compiled.Regex, str, repl, count)
	}

	vm.Push(runtime.NewString(result))
	return 1
}

// re.subn(pattern, repl, string[, count, flags]) -> (string, count)
func reSubn(vm *runtime.VM) int {
	pattern := vm.CheckString(1)
	repl := vm.CheckString(2)
	str := vm.CheckString(3)
	maxCount := -1
	flags := 0

	if vm.GetTop() >= 4 {
		maxCount = int(vm.ToInt(4))
	}
	if vm.GetTop() >= 5 {
		flags = int(vm.ToInt(5))
	}

	compiled, err := compilePattern(pattern, flags)
	if err != nil {
		vm.RaiseError("error in regular expression: %s", err.Error())
		return 0
	}

	var result string
	var count int

	if maxCount < 0 {
		matches := compiled.Regex.FindAllString(str, -1)
		count = len(matches)
		result = compiled.Regex.ReplaceAllString(str, repl)
	} else {
		result, count = replaceNCount(compiled.Regex, str, repl, maxCount)
	}

	vm.Push(runtime.NewTuple([]runtime.Value{
		runtime.NewString(result),
		runtime.NewInt(int64(count)),
	}))
	return 1
}

func replaceN(re *regexp.Regexp, str, repl string, n int) string {
	if n == 0 {
		return str
	}

	result := str
	for i := 0; (n < 0 || i < n) && re.MatchString(result); i++ {
		result = re.ReplaceAllStringFunc(result, func(match string) string {
			return repl
		})
		// Only replace first occurrence
		loc := re.FindStringIndex(result)
		if loc == nil {
			break
		}
		result = result[:loc[0]] + repl + result[loc[1]:]
		if n > 0 {
			n--
		}
		if n == 0 {
			break
		}
	}
	return result
}

func replaceNCount(re *regexp.Regexp, str, repl string, n int) (string, int) {
	if n == 0 {
		return str, 0
	}

	count := 0
	result := str

	for (n < 0 || count < n) && re.MatchString(result) {
		loc := re.FindStringIndex(result)
		if loc == nil {
			break
		}
		result = result[:loc[0]] + repl + result[loc[1]:]
		count++
	}

	return result, count
}

// re.split(pattern, string[, maxsplit, flags]) -> list
func reSplit(vm *runtime.VM) int {
	pattern := vm.CheckString(1)
	str := vm.CheckString(2)
	maxsplit := -1
	flags := 0

	if vm.GetTop() >= 3 {
		maxsplit = int(vm.ToInt(3))
	}
	if vm.GetTop() >= 4 {
		flags = int(vm.ToInt(4))
	}

	compiled, err := compilePattern(pattern, flags)
	if err != nil {
		vm.RaiseError("error in regular expression: %s", err.Error())
		return 0
	}

	var parts []string
	if maxsplit < 0 {
		parts = compiled.Regex.Split(str, -1)
	} else {
		parts = compiled.Regex.Split(str, maxsplit+1)
	}

	result := make([]runtime.Value, len(parts))
	for i, p := range parts {
		result[i] = runtime.NewString(p)
	}

	vm.Push(runtime.NewList(result))
	return 1
}

// re.escape(string) -> string with special chars escaped
func reEscape(vm *runtime.VM) int {
	str := vm.CheckString(1)
	escaped := regexp.QuoteMeta(str)
	vm.Push(runtime.NewString(escaped))
	return 1
}

// Pattern methods

func patternMatch(vm *runtime.VM) int {
	ud := vm.CheckUserData(1, "re.Pattern")
	pattern := ud.Value.(*PyPattern)
	str := vm.CheckString(2)

	// Anchor at start for match
	loc := pattern.Regex.FindStringIndex(str)
	if loc == nil || loc[0] != 0 {
		vm.Push(runtime.None)
		return 1
	}

	groups := pattern.Regex.FindStringSubmatch(str)
	vm.Push(createMatch(pattern, str, loc, groups))
	return 1
}

func patternSearch(vm *runtime.VM) int {
	ud := vm.CheckUserData(1, "re.Pattern")
	pattern := ud.Value.(*PyPattern)
	str := vm.CheckString(2)

	loc := pattern.Regex.FindStringIndex(str)
	if loc == nil {
		vm.Push(runtime.None)
		return 1
	}

	groups := pattern.Regex.FindStringSubmatch(str)
	vm.Push(createMatch(pattern, str, loc, groups))
	return 1
}

func patternFindall(vm *runtime.VM) int {
	ud := vm.CheckUserData(1, "re.Pattern")
	pattern := ud.Value.(*PyPattern)
	str := vm.CheckString(2)

	matches := pattern.Regex.FindAllStringSubmatch(str, -1)
	result := make([]runtime.Value, 0, len(matches))

	for _, match := range matches {
		if len(match) == 1 {
			result = append(result, runtime.NewString(match[0]))
		} else if len(match) == 2 {
			result = append(result, runtime.NewString(match[1]))
		} else {
			groups := make([]runtime.Value, len(match)-1)
			for i, g := range match[1:] {
				groups[i] = runtime.NewString(g)
			}
			result = append(result, runtime.NewTuple(groups))
		}
	}

	vm.Push(runtime.NewList(result))
	return 1
}

func patternFinditer(vm *runtime.VM) int {
	ud := vm.CheckUserData(1, "re.Pattern")
	pattern := ud.Value.(*PyPattern)
	str := vm.CheckString(2)

	allMatches := pattern.Regex.FindAllStringSubmatchIndex(str, -1)
	matches := make([]runtime.Value, 0, len(allMatches))

	for _, loc := range allMatches {
		groups := make([]string, 0)
		for i := 0; i < len(loc); i += 2 {
			if loc[i] >= 0 && loc[i+1] >= 0 {
				groups = append(groups, str[loc[i]:loc[i+1]])
			} else {
				groups = append(groups, "")
			}
		}
		matches = append(matches, createMatch(pattern, str, loc[:2], groups))
	}

	vm.Push(runtime.NewList(matches))
	return 1
}

func patternSub(vm *runtime.VM) int {
	ud := vm.CheckUserData(1, "re.Pattern")
	pattern := ud.Value.(*PyPattern)
	repl := vm.CheckString(2)
	str := vm.CheckString(3)

	result := pattern.Regex.ReplaceAllString(str, repl)
	vm.Push(runtime.NewString(result))
	return 1
}

func patternSubn(vm *runtime.VM) int {
	ud := vm.CheckUserData(1, "re.Pattern")
	pattern := ud.Value.(*PyPattern)
	repl := vm.CheckString(2)
	str := vm.CheckString(3)

	matches := pattern.Regex.FindAllString(str, -1)
	count := len(matches)
	result := pattern.Regex.ReplaceAllString(str, repl)

	vm.Push(runtime.NewTuple([]runtime.Value{
		runtime.NewString(result),
		runtime.NewInt(int64(count)),
	}))
	return 1
}

func patternSplit(vm *runtime.VM) int {
	ud := vm.CheckUserData(1, "re.Pattern")
	pattern := ud.Value.(*PyPattern)
	str := vm.CheckString(2)

	parts := pattern.Regex.Split(str, -1)
	result := make([]runtime.Value, len(parts))
	for i, p := range parts {
		result[i] = runtime.NewString(p)
	}

	vm.Push(runtime.NewList(result))
	return 1
}

// Match methods

func matchGroup(vm *runtime.VM) int {
	ud := vm.CheckUserData(1, "re.Match")
	match := ud.Value.(*PyMatch)

	groupNum := 0
	if vm.GetTop() >= 2 {
		groupNum = int(vm.ToInt(2))
	}

	if groupNum < 0 || groupNum >= len(match.Groups) {
		vm.RaiseError("no such group")
		return 0
	}

	vm.Push(runtime.NewString(match.Groups[groupNum]))
	return 1
}

func matchGroups(vm *runtime.VM) int {
	ud := vm.CheckUserData(1, "re.Match")
	match := ud.Value.(*PyMatch)

	// Return all groups except group 0 (the full match)
	if len(match.Groups) <= 1 {
		vm.Push(runtime.NewTuple([]runtime.Value{}))
		return 1
	}

	groups := make([]runtime.Value, len(match.Groups)-1)
	for i, g := range match.Groups[1:] {
		groups[i] = runtime.NewString(g)
	}

	vm.Push(runtime.NewTuple(groups))
	return 1
}

func matchStart(vm *runtime.VM) int {
	ud := vm.CheckUserData(1, "re.Match")
	match := ud.Value.(*PyMatch)
	vm.Push(runtime.NewInt(int64(match.Start)))
	return 1
}

func matchEnd(vm *runtime.VM) int {
	ud := vm.CheckUserData(1, "re.Match")
	match := ud.Value.(*PyMatch)
	vm.Push(runtime.NewInt(int64(match.End)))
	return 1
}

func matchSpan(vm *runtime.VM) int {
	ud := vm.CheckUserData(1, "re.Match")
	match := ud.Value.(*PyMatch)
	vm.Push(runtime.NewTuple([]runtime.Value{
		runtime.NewInt(int64(match.Start)),
		runtime.NewInt(int64(match.End)),
	}))
	return 1
}

// Helper to register type metatables
func init() {
	// Ensure strings package is imported for escape functionality
	_ = strings.Contains
}
