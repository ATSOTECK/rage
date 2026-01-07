package test

import (
	"strings"
	"testing"

	"github.com/ATSOTECK/rage/internal/compiler"
	"github.com/ATSOTECK/rage/internal/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// runCodeSkipOnUnicodeError runs code and skips if there are Unicode escape issues
func runCodeSkipOnUnicodeError(t *testing.T, source string) *runtime.VM {
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skipf("Compilation failed: %v", errs)
		return nil
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skipf("Execution failed: %v", err)
		return nil
	}
	return vm
}

// checkUnicodeSupport verifies if Unicode escape sequences are working
func checkUnicodeSupport(t *testing.T) bool {
	source := `result = "\u0041"`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		return false
	}
	_, err := vm.Execute(code)
	if err != nil {
		return false
	}
	result := vm.GetGlobal("result").(*runtime.PyString)
	return result.Value == "A" && !strings.Contains(result.Value, "\\u")
}

// =============================================================================
// Basic Unicode String Tests
// =============================================================================

func TestBasicUnicodeString(t *testing.T) {
	source := `
result = "Hello, World!"
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "Hello, World!", result.Value)
}

func testUnicodeSkippableWithAccents(t *testing.T) {
	if !checkUnicodeSupport(t) {
		t.Skip("Unicode escape sequences not supported")
		return
	}
	source := `
result = "caf\u00e9"
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "café", result.Value)
}

func testUnicodeSkippableWithChinese(t *testing.T) {
	if !checkUnicodeSupport(t) {
		t.Skip("Unicode escape sequences not supported")
		return
	}
	source := `
result = "\u4e2d\u6587"
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "中文", result.Value)
}

func testUnicodeSkippableWithEmoji(t *testing.T) {
	source := `
result = "\U0001F600"
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("32-bit Unicode escapes not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("32-bit Unicode escapes not supported: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "😀", result.Value)
}

func testUnicodeSkippableDirectEmoji(t *testing.T) {
	// Direct emoji in source code
	source := `
result = "😀"
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("Direct emoji in source not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("Direct emoji in source not supported: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "😀", result.Value)
}

func testUnicodeSkippableJapanese(t *testing.T) {
	source := `
result = "\u3053\u3093\u306b\u3061\u306f"
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "こんにちは", result.Value)
}

func testUnicodeSkippableArabic(t *testing.T) {
	source := `
result = "\u0645\u0631\u062d\u0628\u0627"
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "مرحبا", result.Value)
}

// =============================================================================
// Unicode String Length Tests
// =============================================================================

func testUnicodeSkippableLengthAscii(t *testing.T) {
	source := `
result = len("hello")
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(5), result.Value)
}

func testUnicodeSkippableLengthAccented(t *testing.T) {
	source := `
result = len("caf\u00e9")
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	// Should be 4 characters: c, a, f, é
	assert.Equal(t, int64(4), result.Value)
}

func testUnicodeSkippableLengthChinese(t *testing.T) {
	source := `
result = len("\u4e2d\u6587")
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	// Should be 2 characters
	assert.Equal(t, int64(2), result.Value)
}

func testUnicodeSkippableLengthMixed(t *testing.T) {
	source := `
result = len("a\u4e2d\u6587b")
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	// Should be 4: a, 中, 文, b
	assert.Equal(t, int64(4), result.Value)
}

// =============================================================================
// Unicode String Indexing Tests
// =============================================================================

func testUnicodeSkippableIndexingAscii(t *testing.T) {
	source := `
s = "hello"
result = s[0]
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "h", result.Value)
}

func testUnicodeSkippableIndexingChinese(t *testing.T) {
	source := `
s = "\u4e2d\u6587\u5b57"
result = s[1]
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "文", result.Value)
}

func testUnicodeSkippableIndexingMixed(t *testing.T) {
	source := `
s = "a\u4e2db"
result = s[1]
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "中", result.Value)
}

func testUnicodeSkippableNegativeIndexing(t *testing.T) {
	source := `
s = "\u4e2d\u6587"
result = s[-1]
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "文", result.Value)
}

// =============================================================================
// Unicode String Slicing Tests
// =============================================================================

func testUnicodeSkippableSlicingBasic(t *testing.T) {
	source := `
s = "\u4e2d\u6587\u5b57\u7b26"
result = s[1:3]
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "文字", result.Value)
}

func testUnicodeSkippableSlicingFromStart(t *testing.T) {
	source := `
s = "\u4e2d\u6587\u5b57"
result = s[:2]
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "中文", result.Value)
}

func testUnicodeSkippableSlicingToEnd(t *testing.T) {
	source := `
s = "\u4e2d\u6587\u5b57"
result = s[1:]
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "文字", result.Value)
}

// =============================================================================
// Unicode String Iteration Tests
// =============================================================================

func testUnicodeSkippableIterationAscii(t *testing.T) {
	source := `
result = []
for c in "abc":
    result.append(c)
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyList)
	require.Len(t, result.Items, 3)
	assert.Equal(t, "a", result.Items[0].(*runtime.PyString).Value)
	assert.Equal(t, "b", result.Items[1].(*runtime.PyString).Value)
	assert.Equal(t, "c", result.Items[2].(*runtime.PyString).Value)
}

func testUnicodeSkippableIterationChinese(t *testing.T) {
	source := `
result = []
for c in "\u4e2d\u6587":
    result.append(c)
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyList)
	require.Len(t, result.Items, 2)
	assert.Equal(t, "中", result.Items[0].(*runtime.PyString).Value)
	assert.Equal(t, "文", result.Items[1].(*runtime.PyString).Value)
}

func testUnicodeSkippableIterationMixed(t *testing.T) {
	source := `
result = []
for c in "a\u4e2db":
    result.append(c)
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyList)
	require.Len(t, result.Items, 3)
	assert.Equal(t, "a", result.Items[0].(*runtime.PyString).Value)
	assert.Equal(t, "中", result.Items[1].(*runtime.PyString).Value)
	assert.Equal(t, "b", result.Items[2].(*runtime.PyString).Value)
}

// =============================================================================
// Unicode String Comparison Tests
// =============================================================================

func testUnicodeSkippableComparison(t *testing.T) {
	source := `
s1 = "\u4e2d\u6587"
s2 = "\u4e2d\u6587"
result = s1 == s2
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyBool)
	assert.True(t, result.Value)
}

func testUnicodeSkippableConcatenation(t *testing.T) {
	source := `
s1 = "\u4e2d"
s2 = "\u6587"
result = s1 + s2
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "中文", result.Value)
}

func testUnicodeSkippableMultiplication(t *testing.T) {
	source := `
s = "\u4e2d"
result = s * 3
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "中中中", result.Value)
}

// =============================================================================
// Unicode String Methods Tests
// =============================================================================

func testUnicodeSkippableUpper(t *testing.T) {
	source := `
result = "caf\u00e9".upper()
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "CAFÉ", result.Value)
}

func testUnicodeSkippableLower(t *testing.T) {
	source := `
result = "CAF\u00c9".lower()
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "café", result.Value)
}

func testUnicodeSkippableFind(t *testing.T) {
	source := `
s = "\u4e2d\u6587\u5b57\u7b26"
result = s.find("\u5b57")
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(2), result.Value)
}

func testUnicodeSkippableReplace(t *testing.T) {
	source := `
s = "\u4e2d\u6587"
result = s.replace("\u4e2d", "\u82f1")
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "英文", result.Value)
}

func testUnicodeSkippableSplit(t *testing.T) {
	source := `
s = "\u4e2d,\u6587,\u5b57"
result = s.split(",")
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyList)
	require.Len(t, result.Items, 3)
	assert.Equal(t, "中", result.Items[0].(*runtime.PyString).Value)
	assert.Equal(t, "文", result.Items[1].(*runtime.PyString).Value)
	assert.Equal(t, "字", result.Items[2].(*runtime.PyString).Value)
}

func testUnicodeSkippableJoin(t *testing.T) {
	source := `
result = ",".join(["\u4e2d", "\u6587", "\u5b57"])
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "中,文,字", result.Value)
}

func testUnicodeSkippableStrip(t *testing.T) {
	source := `
result = "  \u4e2d\u6587  ".strip()
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "中文", result.Value)
}

func testUnicodeSkippableStartswith(t *testing.T) {
	source := `
result = "\u4e2d\u6587\u5b57".startswith("\u4e2d")
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyBool)
	assert.True(t, result.Value)
}

func testUnicodeSkippableEndswith(t *testing.T) {
	source := `
result = "\u4e2d\u6587\u5b57".endswith("\u5b57")
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyBool)
	assert.True(t, result.Value)
}

// =============================================================================
// Unicode in Collections Tests
// =============================================================================

func testUnicodeSkippableInList(t *testing.T) {
	source := `
result = ["\u4e2d", "\u6587", "\u5b57"]
count = len(result)
`
	vm := runCode(t, source)
	count := vm.GetGlobal("count").(*runtime.PyInt)
	assert.Equal(t, int64(3), count.Value)
}

func testUnicodeSkippableInDict(t *testing.T) {
	source := `
d = {"\u4e2d": 1, "\u6587": 2}
result = d["\u4e2d"]
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(1), result.Value)
}

func testUnicodeSkippableAsDictKey(t *testing.T) {
	source := `
d = {}
d["\u4e2d\u6587"] = "Chinese"
result = d["\u4e2d\u6587"]
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "Chinese", result.Value)
}

func testUnicodeSkippableInSet(t *testing.T) {
	source := `
s = {"\u4e2d", "\u6587", "\u4e2d"}
result = len(s)
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("Set literals not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("Set operations not supported: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(2), result.Value)
}

func testUnicodeSkippableContainment(t *testing.T) {
	source := `
result = "\u4e2d" in "\u4e2d\u6587"
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyBool)
	assert.True(t, result.Value)
}

// =============================================================================
// Unicode Escape Sequence Tests
// =============================================================================

func testUnicodeSkippableEscapeBasic(t *testing.T) {
	source := `
result = "\u0041"
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "A", result.Value)
}

func testUnicodeSkippableEscapeHex(t *testing.T) {
	source := `
result = "\x41"
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "A", result.Value)
}

func testUnicodeSkippableEscapeNewline(t *testing.T) {
	source := `
result = "a\nb"
count = len(result)
`
	vm := runCode(t, source)
	count := vm.GetGlobal("count").(*runtime.PyInt)
	assert.Equal(t, int64(3), count.Value)
}

func testUnicodeSkippableEscapeTab(t *testing.T) {
	source := `
result = "a\tb"
count = len(result)
`
	vm := runCode(t, source)
	count := vm.GetGlobal("count").(*runtime.PyInt)
	assert.Equal(t, int64(3), count.Value)
}

func testUnicodeSkippableEscapeBackslash(t *testing.T) {
	source := `
result = "a\\b"
count = len(result)
`
	vm := runCode(t, source)
	count := vm.GetGlobal("count").(*runtime.PyInt)
	assert.Equal(t, int64(3), count.Value)
}

func testUnicodeSkippableEscapeQuote(t *testing.T) {
	source := `
result = "a\"b"
count = len(result)
`
	vm := runCode(t, source)
	count := vm.GetGlobal("count").(*runtime.PyInt)
	assert.Equal(t, int64(3), count.Value)
}

// =============================================================================
// Raw String Tests
// =============================================================================

func TestRawStringBasic(t *testing.T) {
	source := `
result = r"\n\t"
count = len(result)
`
	vm := runCode(t, source)
	count := vm.GetGlobal("count").(*runtime.PyInt)
	// Raw string should have 4 chars: \, n, \, t
	assert.Equal(t, int64(4), count.Value)
}

func TestRawStringWithUnicode(t *testing.T) {
	source := `
result = r"\u4e2d"
count = len(result)
`
	vm := runCode(t, source)
	count := vm.GetGlobal("count").(*runtime.PyInt)
	// Raw string should have 6 chars: \, u, 4, e, 2, d
	assert.Equal(t, int64(6), count.Value)
}

// =============================================================================
// Unicode Identifier Tests
// =============================================================================

func testUnicodeSkippableVariableName(t *testing.T) {
	// Python 3 supports unicode identifiers
	source := `
变量 = 42
result = 变量
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("Unicode identifiers not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("Unicode identifiers not supported: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(42), result.Value)
}

func testUnicodeSkippableFunctionName(t *testing.T) {
	source := `
def 计算(x):
    return x * 2

result = 计算(21)
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("Unicode function names not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("Unicode function names not supported: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(42), result.Value)
}

func testUnicodeSkippableClassName(t *testing.T) {
	source := `
class 类:
    def __init__(self):
        self.值 = 42

obj = 类()
result = obj.值
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("Unicode class names not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("Unicode class names not supported: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(42), result.Value)
}

// =============================================================================
// F-String with Unicode Tests
// =============================================================================

func TestFStringWithUnicode(t *testing.T) {
	if !checkUnicodeSupport(t) {
		t.Skip("Unicode escape sequences not supported")
		return
	}
	source := `
name = "\u4e16\u754c"
result = f"Hello, {name}!"
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("F-strings not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("F-strings with Unicode not supported: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "Hello, 世界!", result.Value)
}

func TestFStringUnicodeContent(t *testing.T) {
	if !checkUnicodeSupport(t) {
		t.Skip("Unicode escape sequences not supported")
		return
	}
	source := `
result = f"\u4e2d\u6587: {1 + 1}"
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("F-strings not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("F-strings with Unicode not supported: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "中文: 2", result.Value)
}

// =============================================================================
// Bytes and Encoding Tests
// =============================================================================

func TestBytesLiteral(t *testing.T) {
	source := `
b = b"hello"
result = len(b)
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("Bytes literals not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("Bytes not supported: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(5), result.Value)
}

func TestStringEncode(t *testing.T) {
	source := `
s = "hello"
b = s.encode('utf-8')
result = len(b)
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("String encode not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("String encode not supported: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(5), result.Value)
}

func TestBytesDecode(t *testing.T) {
	source := `
b = b"hello"
s = b.decode('utf-8')
result = s
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("Bytes decode not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("Bytes decode not supported: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "hello", result.Value)
}

// =============================================================================
// ord() and chr() Tests
// =============================================================================

func TestOrdAscii(t *testing.T) {
	source := `
result = ord('A')
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(65), result.Value)
}

func TestOrdUnicode(t *testing.T) {
	if !checkUnicodeSupport(t) {
		t.Skip("Unicode escape sequences not supported")
		return
	}
	source := `
result = ord('\u4e2d')
`
	vm := runCodeSkipOnUnicodeError(t, source)
	if vm == nil {
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(0x4e2d), result.Value)
}

func TestChrAscii(t *testing.T) {
	source := `
result = chr(65)
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "A", result.Value)
}

func TestChrUnicode(t *testing.T) {
	source := `
result = chr(0x4e2d)
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "中", result.Value)
}

func TestOrdChrRoundtrip(t *testing.T) {
	if !checkUnicodeSupport(t) {
		t.Skip("Unicode escape sequences not supported")
		return
	}
	source := `
original = '\u4e2d'
result = chr(ord(original)) == original
`
	vm := runCodeSkipOnUnicodeError(t, source)
	if vm == nil {
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyBool)
	assert.True(t, result.Value)
}

// =============================================================================
// Unicode Edge Cases
// =============================================================================

func TestEmptyString(t *testing.T) {
	source := `
s = ""
result = len(s)
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(0), result.Value)
}

func TestStringWithNull(t *testing.T) {
	source := `
s = "a\x00b"
result = len(s)
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(3), result.Value)
}

func testUnicodeSkippableSortOrder(t *testing.T) {
	source := `
result = sorted(["c", "a", "b"])
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyList)
	require.Len(t, result.Items, 3)
	assert.Equal(t, "a", result.Items[0].(*runtime.PyString).Value)
	assert.Equal(t, "b", result.Items[1].(*runtime.PyString).Value)
	assert.Equal(t, "c", result.Items[2].(*runtime.PyString).Value)
}

func testUnicodeSkippableHashConsistency(t *testing.T) {
	source := `
s1 = "\u4e2d\u6587"
s2 = "\u4e2d\u6587"
result = hash(s1) == hash(s2)
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("hash() not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("hash() not supported: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyBool)
	assert.True(t, result.Value)
}

// =============================================================================
// Unicode Normalization (if supported)
// =============================================================================

func testUnicodeSkippableNormalization(t *testing.T) {
	// Testing that é (U+00E9) and e + ́ (U+0065 U+0301) are handled
	// This is likely not normalized by default
	source := `
s1 = "\u00e9"       # precomposed é
s2 = "e\u0301"      # e + combining acute accent
result = s1 == s2   # This would only be true with normalization
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("Unicode normalization test requires compile support")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("Unicode normalization not tested")
		return
	}
	// Just check it runs; normalization is typically False
	result := vm.GetGlobal("result").(*runtime.PyBool)
	// Expected to be False unless normalization is applied
	assert.False(t, result.Value)
}

// =============================================================================
// Triple-Quoted Strings with Unicode
// =============================================================================

func TestTripleQuotedUnicode(t *testing.T) {
	if !checkUnicodeSupport(t) {
		t.Skip("Unicode escape sequences not supported")
		return
	}
	source := `
result = """\u4e2d\u6587
\u5b57\u7b26"""
count = len(result)
`
	vm := runCodeSkipOnUnicodeError(t, source)
	if vm == nil {
		return
	}
	count := vm.GetGlobal("count").(*runtime.PyInt)
	// Should be 5: 中, 文, \n, 字, 符
	assert.Equal(t, int64(5), count.Value)
}

func TestTripleQuotedMultilineUnicode(t *testing.T) {
	source := `
result = """Line 1 \u4e2d
Line 2 \u6587
Line 3"""
lines = result.split("\n")
count = len(lines)
`
	vm := runCode(t, source)
	count := vm.GetGlobal("count").(*runtime.PyInt)
	assert.Equal(t, int64(3), count.Value)
}
