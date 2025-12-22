package test

import (
	"testing"

	"github.com/ATSOTECK/RAGE/internal/compiler"
	"github.com/ATSOTECK/RAGE/internal/runtime"
	"github.com/ATSOTECK/RAGE/internal/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =====================================
// Random Module Tests
// =====================================

func TestRandomRandom(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitAllModules()

	vm := runtime.NewVM()

	source := `
import random
random.seed(42)
r1 = random.random()
r2 = random.random()
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	r1 := vm.GetGlobal("r1").(*runtime.PyFloat).Value
	r2 := vm.GetGlobal("r2").(*runtime.PyFloat).Value

	// Values should be in [0, 1) and different
	assert.True(t, r1 >= 0 && r1 < 1)
	assert.True(t, r2 >= 0 && r2 < 1)
	assert.NotEqual(t, r1, r2)
}

func TestRandomRandint(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitAllModules()

	vm := runtime.NewVM()

	source := `
import random
random.seed(42)
results = []
for i in range(100):
    results.append(random.randint(1, 10))
min_val = min(results)
max_val = max(results)
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	minVal := vm.GetGlobal("min_val").(*runtime.PyInt).Value
	maxVal := vm.GetGlobal("max_val").(*runtime.PyInt).Value

	assert.True(t, minVal >= 1)
	assert.True(t, maxVal <= 10)
}

func TestRandomChoice(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitAllModules()

	vm := runtime.NewVM()

	source := `
import random
random.seed(42)
choices = ["a", "b", "c", "d"]
result = random.choice(choices)
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("result").(*runtime.PyString).Value
	assert.Contains(t, []string{"a", "b", "c", "d"}, result)
}

func TestRandomShuffle(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitAllModules()

	vm := runtime.NewVM()

	source := `
import random
random.seed(42)
items = [1, 2, 3, 4, 5]
random.shuffle(items)
total = sum(items)
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	// Sum should be unchanged
	total := vm.GetGlobal("total").(*runtime.PyInt).Value
	assert.Equal(t, int64(15), total)
}

func TestRandomSample(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitAllModules()

	vm := runtime.NewVM()

	source := `
import random
random.seed(42)
population = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
sample = random.sample(population, 3)
sample_len = len(sample)
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	sampleLen := vm.GetGlobal("sample_len").(*runtime.PyInt).Value
	assert.Equal(t, int64(3), sampleLen)
}

func TestRandomUniform(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitAllModules()

	vm := runtime.NewVM()

	source := `
import random
random.seed(42)
r = random.uniform(10.0, 20.0)
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	r := vm.GetGlobal("r").(*runtime.PyFloat).Value
	assert.True(t, r >= 10.0 && r <= 20.0)
}

func TestFromRandomImport(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitAllModules()

	vm := runtime.NewVM()

	source := `
from random import randint, choice, seed
seed(42)
r = randint(1, 100)
c = choice([1, 2, 3])
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	r := vm.GetGlobal("r").(*runtime.PyInt).Value
	c := vm.GetGlobal("c").(*runtime.PyInt).Value

	assert.True(t, r >= 1 && r <= 100)
	assert.True(t, c >= 1 && c <= 3)
}

// =====================================
// String Module Tests
// =====================================

func TestStringConstants(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitAllModules()

	vm := runtime.NewVM()

	source := `
import string
lowercase = string.ascii_lowercase
uppercase = string.ascii_uppercase
digits = string.digits
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	lowercase := vm.GetGlobal("lowercase").(*runtime.PyString).Value
	uppercase := vm.GetGlobal("uppercase").(*runtime.PyString).Value
	digits := vm.GetGlobal("digits").(*runtime.PyString).Value

	assert.Equal(t, "abcdefghijklmnopqrstuvwxyz", lowercase)
	assert.Equal(t, "ABCDEFGHIJKLMNOPQRSTUVWXYZ", uppercase)
	assert.Equal(t, "0123456789", digits)
}

func TestStringAsciiLetters(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitAllModules()

	vm := runtime.NewVM()

	source := `
import string
letters = string.ascii_letters
has_a = "a" in letters
has_Z = "Z" in letters
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	letters := vm.GetGlobal("letters").(*runtime.PyString).Value
	assert.Equal(t, 52, len(letters))
}

func TestStringCapwords(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitAllModules()

	vm := runtime.NewVM()

	source := `
import string
result = string.capwords("hello world")
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("result").(*runtime.PyString).Value
	assert.Equal(t, "Hello World", result)
}

func TestFromStringImport(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitAllModules()

	vm := runtime.NewVM()

	source := `
from string import digits, punctuation, whitespace
d = digits
p = punctuation
w = whitespace
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	d := vm.GetGlobal("d").(*runtime.PyString).Value
	p := vm.GetGlobal("p").(*runtime.PyString).Value
	w := vm.GetGlobal("w").(*runtime.PyString).Value

	assert.Equal(t, "0123456789", d)
	assert.Contains(t, p, "!")
	assert.Contains(t, w, " ")
	assert.Contains(t, w, "\n")
}

// =====================================
// Sys Module Tests
// =====================================

func TestSysVersion(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitAllModules()

	vm := runtime.NewVM()

	source := `
import sys
version = sys.version
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	version := vm.GetGlobal("version").(*runtime.PyString).Value
	assert.Contains(t, version, "3.14")
	assert.Contains(t, version, "RAGE")
}

func TestSysVersionInfo(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitAllModules()

	vm := runtime.NewVM()

	source := `
import sys
major = sys.version_info[0]
minor = sys.version_info[1]
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	major := vm.GetGlobal("major").(*runtime.PyInt).Value
	minor := vm.GetGlobal("minor").(*runtime.PyInt).Value

	assert.Equal(t, int64(3), major)
	assert.Equal(t, int64(14), minor)
}

func TestSysPlatform(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitAllModules()

	vm := runtime.NewVM()

	source := `
import sys
platform = sys.platform
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	platform := vm.GetGlobal("platform").(*runtime.PyString).Value
	// Should be one of the known platforms
	assert.Contains(t, []string{"darwin", "linux", "win32", "freebsd"}, platform)
}

func TestSysMaxsize(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitAllModules()

	vm := runtime.NewVM()

	source := `
import sys
maxsize = sys.maxsize
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	maxsize := vm.GetGlobal("maxsize").(*runtime.PyInt).Value
	assert.True(t, maxsize > 0)
}

func TestSysArgv(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitAllModules()

	vm := runtime.NewVM()

	source := `
import sys
argv = sys.argv
argv_len = len(argv)
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	// argv should be a list
	argv := vm.GetGlobal("argv")
	_, ok := argv.(*runtime.PyList)
	assert.True(t, ok, "sys.argv should be a list")
}

func TestSysRecursionLimit(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitAllModules()

	vm := runtime.NewVM()

	source := `
import sys
limit1 = sys.getrecursionlimit()
sys.setrecursionlimit(500)
limit2 = sys.getrecursionlimit()
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	limit1 := vm.GetGlobal("limit1").(*runtime.PyInt).Value
	limit2 := vm.GetGlobal("limit2").(*runtime.PyInt).Value

	assert.Equal(t, int64(1000), limit1)
	assert.Equal(t, int64(500), limit2)
}

func TestSysGetsizeof(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitAllModules()

	vm := runtime.NewVM()

	source := `
import sys
size_int = sys.getsizeof(42)
size_str = sys.getsizeof("hello")
size_list = sys.getsizeof([1, 2, 3])
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	sizeInt := vm.GetGlobal("size_int").(*runtime.PyInt).Value
	sizeStr := vm.GetGlobal("size_str").(*runtime.PyInt).Value
	sizeList := vm.GetGlobal("size_list").(*runtime.PyInt).Value

	assert.True(t, sizeInt > 0)
	assert.True(t, sizeStr > 0)
	assert.True(t, sizeList > 0)
}

func TestFromSysImport(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitAllModules()

	vm := runtime.NewVM()

	source := `
from sys import version, platform, maxsize
v = version
p = platform
m = maxsize
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	v := vm.GetGlobal("v").(*runtime.PyString).Value
	p := vm.GetGlobal("p").(*runtime.PyString).Value
	m := vm.GetGlobal("m").(*runtime.PyInt).Value

	assert.Contains(t, v, "3.14")
	assert.NotEmpty(t, p)
	assert.True(t, m > 0)
}

func TestSysByteorder(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitAllModules()

	vm := runtime.NewVM()

	source := `
import sys
byteorder = sys.byteorder
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	byteorder := vm.GetGlobal("byteorder").(*runtime.PyString).Value
	assert.Contains(t, []string{"little", "big"}, byteorder)
}

// =====================================
// Time Module Tests
// =====================================

func TestTimeTime(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitAllModules()

	vm := runtime.NewVM()

	source := `
import time
t1 = time.time()
t2 = time.time()
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	t1 := vm.GetGlobal("t1").(*runtime.PyFloat).Value
	t2 := vm.GetGlobal("t2").(*runtime.PyFloat).Value

	// Both should be positive and t2 >= t1
	assert.True(t, t1 > 0)
	assert.True(t, t2 >= t1)
}

func TestTimeTimeNs(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitAllModules()

	vm := runtime.NewVM()

	source := `
import time
ns = time.time_ns()
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	ns := vm.GetGlobal("ns").(*runtime.PyInt).Value
	assert.True(t, ns > 0)
}

func TestTimeLocaltime(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitAllModules()

	vm := runtime.NewVM()

	source := `
import time
lt = time.localtime()
year = lt[0]
month = lt[1]
day = lt[2]
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	year := vm.GetGlobal("year").(*runtime.PyInt).Value
	month := vm.GetGlobal("month").(*runtime.PyInt).Value
	day := vm.GetGlobal("day").(*runtime.PyInt).Value

	assert.True(t, year >= 2020)
	assert.True(t, month >= 1 && month <= 12)
	assert.True(t, day >= 1 && day <= 31)
}

func TestTimeStrftime(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitAllModules()

	vm := runtime.NewVM()

	source := `
import time
lt = time.localtime()
formatted = time.strftime("%Y-%m-%d", lt)
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	formatted := vm.GetGlobal("formatted").(*runtime.PyString).Value
	// Should match YYYY-MM-DD pattern
	assert.Regexp(t, `^\d{4}-\d{2}-\d{2}$`, formatted)
}

func TestTimePerfCounter(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitAllModules()

	vm := runtime.NewVM()

	source := `
import time
pc1 = time.perf_counter()
pc2 = time.perf_counter()
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	pc1 := vm.GetGlobal("pc1").(*runtime.PyFloat).Value
	pc2 := vm.GetGlobal("pc2").(*runtime.PyFloat).Value

	assert.True(t, pc1 >= 0)
	assert.True(t, pc2 >= pc1)
}

func TestFromTimeImport(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitAllModules()

	vm := runtime.NewVM()

	source := `
from time import time, localtime
t = time()
lt = localtime()
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	ti := vm.GetGlobal("t").(*runtime.PyFloat).Value
	assert.True(t, ti > 0)
}

// =====================================
// Re Module Tests
// =====================================

func TestReMatch(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitAllModules()

	vm := runtime.NewVM()

	source := `
import re
result = re.search(r"^hello", "hello world")
matched = result is not None
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	matched := vm.GetGlobal("matched").(*runtime.PyBool).Value
	assert.True(t, matched)
}

func TestReSearch(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitAllModules()

	vm := runtime.NewVM()

	source := `
import re
m = re.search(r"world", "hello world")
matched = m is not None
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	matched := vm.GetGlobal("matched").(*runtime.PyBool).Value
	assert.True(t, matched)
}

func TestReFindall(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitAllModules()

	vm := runtime.NewVM()

	source := `
import re
matches = re.findall(r"\d+", "abc123def456ghi789")
count = len(matches)
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	count := vm.GetGlobal("count").(*runtime.PyInt).Value
	assert.Equal(t, int64(3), count)
}

func TestReSub(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitAllModules()

	vm := runtime.NewVM()

	source := `
import re
result = re.sub(r"\d+", "X", "abc123def456")
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("result").(*runtime.PyString).Value
	assert.Equal(t, "abcXdefX", result)
}

func TestReSplit(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitAllModules()

	vm := runtime.NewVM()

	source := `
import re
parts = re.split(r"[,;]", "a,b;c,d")
count = len(parts)
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	count := vm.GetGlobal("count").(*runtime.PyInt).Value
	assert.Equal(t, int64(4), count)
}

func TestReCompile(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitAllModules()

	vm := runtime.NewVM()

	source := `
import re
pattern = re.compile(r"\d+")
matches = pattern.findall("abc123def456")
count = len(matches)
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	count := vm.GetGlobal("count").(*runtime.PyInt).Value
	assert.Equal(t, int64(2), count)
}

func TestReEscape(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitAllModules()

	vm := runtime.NewVM()

	source := `
import re
escaped = re.escape("hello.world")
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	escaped := vm.GetGlobal("escaped").(*runtime.PyString).Value
	assert.Contains(t, escaped, `\.`)
}

func TestFromReImport(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitAllModules()

	vm := runtime.NewVM()

	source := `
from re import search, findall, sub
result = search(r"^hello", "hello world")
matched = result is not None
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	matched := vm.GetGlobal("matched").(*runtime.PyBool).Value
	assert.True(t, matched)
}

// =====================================
// Collections Module Tests
// =====================================

func TestCollectionsCounter(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitAllModules()

	vm := runtime.NewVM()

	source := `
import collections
c = collections.Counter([1, 1, 1, 2, 2, 3])
total = c.total()
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	total := vm.GetGlobal("total").(*runtime.PyInt).Value
	assert.Equal(t, int64(6), total)
}

func TestCollectionsCounterMostCommon(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitAllModules()

	vm := runtime.NewVM()

	source := `
import collections
c = collections.Counter([1, 1, 1, 2, 2, 3])
mc = c.most_common(2)
count = len(mc)
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	count := vm.GetGlobal("count").(*runtime.PyInt).Value
	assert.Equal(t, int64(2), count)
}

func TestCollectionsDeque(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitAllModules()

	vm := runtime.NewVM()

	source := `
import collections
d = collections.deque([1, 2, 3])
d.append(4)
d.appendleft(0)
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	// Deque operations completed without error
}

func TestCollectionsDequePop(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitAllModules()

	vm := runtime.NewVM()

	source := `
import collections
d = collections.deque([1, 2, 3])
right = d.pop()
left = d.popleft()
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	right := vm.GetGlobal("right").(*runtime.PyInt).Value
	left := vm.GetGlobal("left").(*runtime.PyInt).Value

	assert.Equal(t, int64(3), right)
	assert.Equal(t, int64(1), left)
}

func TestCollectionsDequeRotate(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitAllModules()

	vm := runtime.NewVM()

	source := `
import collections
d = collections.deque([1, 2, 3, 4, 5])
d.rotate(2)
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	// Rotation completed without error
}

func TestCollectionsOrderedDict(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitAllModules()

	vm := runtime.NewVM()

	source := `
import collections
od = collections.OrderedDict()
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	od := vm.GetGlobal("od")
	_, ok := od.(*runtime.PyDict)
	assert.True(t, ok, "OrderedDict should be a dict")
}

func TestFromCollectionsImport(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitAllModules()

	vm := runtime.NewVM()

	source := `
from collections import Counter, deque
c = Counter([1, 1, 2])
d = deque([1, 2, 3])
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)
}
