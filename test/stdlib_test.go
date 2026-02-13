package test

import (
	"testing"

	"github.com/ATSOTECK/rage/internal/runtime"
	"github.com/stretchr/testify/assert"
)

// =====================================
// Random Module Tests
// =====================================

func TestRandomRandom(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import random
random.seed(42)
r1 = random.random()
r2 = random.random()
`)
	r1 := vm.GetGlobal("r1").(*runtime.PyFloat).Value
	r2 := vm.GetGlobal("r2").(*runtime.PyFloat).Value

	// Values should be in [0, 1) and different
	assert.True(t, r1 >= 0 && r1 < 1)
	assert.True(t, r2 >= 0 && r2 < 1)
	assert.NotEqual(t, r1, r2)
}

func TestRandomRandint(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import random
random.seed(42)
results = []
for i in range(100):
    results.append(random.randint(1, 10))
min_val = min(results)
max_val = max(results)
`)
	minVal := vm.GetGlobal("min_val").(*runtime.PyInt).Value
	maxVal := vm.GetGlobal("max_val").(*runtime.PyInt).Value

	assert.True(t, minVal >= 1)
	assert.True(t, maxVal <= 10)
}

func TestRandomChoice(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import random
random.seed(42)
choices = ["a", "b", "c", "d"]
result = random.choice(choices)
`)
	result := vm.GetGlobal("result").(*runtime.PyString).Value
	assert.Contains(t, []string{"a", "b", "c", "d"}, result)
}

func TestRandomShuffle(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import random
random.seed(42)
items = [1, 2, 3, 4, 5]
random.shuffle(items)
total = sum(items)
`)
	// Sum should be unchanged
	total := vm.GetGlobal("total").(*runtime.PyInt).Value
	assert.Equal(t, int64(15), total)
}

func TestRandomSample(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import random
random.seed(42)
population = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
sample = random.sample(population, 3)
sample_len = len(sample)
`)
	sampleLen := vm.GetGlobal("sample_len").(*runtime.PyInt).Value
	assert.Equal(t, int64(3), sampleLen)
}

func TestRandomUniform(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import random
random.seed(42)
r = random.uniform(10.0, 20.0)
`)
	r := vm.GetGlobal("r").(*runtime.PyFloat).Value
	assert.True(t, r >= 10.0 && r <= 20.0)
}

func TestFromRandomImport(t *testing.T) {
	vm := runCodeWithStdlib(t, `
from random import randint, choice, seed
seed(42)
r = randint(1, 100)
c = choice([1, 2, 3])
`)
	r := vm.GetGlobal("r").(*runtime.PyInt).Value
	c := vm.GetGlobal("c").(*runtime.PyInt).Value

	assert.True(t, r >= 1 && r <= 100)
	assert.True(t, c >= 1 && c <= 3)
}

// =====================================
// String Module Tests
// =====================================

func TestStringConstants(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import string
lowercase = string.ascii_lowercase
uppercase = string.ascii_uppercase
digits = string.digits
`)
	lowercase := vm.GetGlobal("lowercase").(*runtime.PyString).Value
	uppercase := vm.GetGlobal("uppercase").(*runtime.PyString).Value
	digits := vm.GetGlobal("digits").(*runtime.PyString).Value

	assert.Equal(t, "abcdefghijklmnopqrstuvwxyz", lowercase)
	assert.Equal(t, "ABCDEFGHIJKLMNOPQRSTUVWXYZ", uppercase)
	assert.Equal(t, "0123456789", digits)
}

func TestStringAsciiLetters(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import string
letters = string.ascii_letters
has_a = "a" in letters
has_Z = "Z" in letters
`)
	letters := vm.GetGlobal("letters").(*runtime.PyString).Value
	assert.Equal(t, 52, len(letters))
}

func TestStringCapwords(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import string
result = string.capwords("hello world")
`)
	result := vm.GetGlobal("result").(*runtime.PyString).Value
	assert.Equal(t, "Hello World", result)
}

func TestFromStringImport(t *testing.T) {
	vm := runCodeWithStdlib(t, `
from string import digits, punctuation, whitespace
d = digits
p = punctuation
w = whitespace
`)
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
	vm := runCodeWithStdlib(t, `
import sys
version = sys.version
`)
	version := vm.GetGlobal("version").(*runtime.PyString).Value
	assert.Contains(t, version, "3.14")
	assert.Contains(t, version, "RAGE")
}

func TestSysVersionInfo(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import sys
major = sys.version_info[0]
minor = sys.version_info[1]
`)
	major := vm.GetGlobal("major").(*runtime.PyInt).Value
	minor := vm.GetGlobal("minor").(*runtime.PyInt).Value

	assert.Equal(t, int64(3), major)
	assert.Equal(t, int64(14), minor)
}

func TestSysPlatform(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import sys
platform = sys.platform
`)
	platform := vm.GetGlobal("platform").(*runtime.PyString).Value
	assert.Contains(t, []string{"darwin", "linux", "win32", "freebsd"}, platform)
}

func TestSysMaxsize(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import sys
maxsize = sys.maxsize
`)
	maxsize := vm.GetGlobal("maxsize").(*runtime.PyInt).Value
	assert.True(t, maxsize > 0)
}

func TestSysArgv(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import sys
argv = sys.argv
argv_len = len(argv)
`)
	argv := vm.GetGlobal("argv")
	_, ok := argv.(*runtime.PyList)
	assert.True(t, ok, "sys.argv should be a list")
}

func TestSysRecursionLimit(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import sys
limit1 = sys.getrecursionlimit()
sys.setrecursionlimit(500)
limit2 = sys.getrecursionlimit()
`)
	limit1 := vm.GetGlobal("limit1").(*runtime.PyInt).Value
	limit2 := vm.GetGlobal("limit2").(*runtime.PyInt).Value

	assert.Equal(t, int64(1000), limit1)
	assert.Equal(t, int64(500), limit2)
}

func TestSysGetsizeof(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import sys
size_int = sys.getsizeof(42)
size_str = sys.getsizeof("hello")
size_list = sys.getsizeof([1, 2, 3])
`)
	sizeInt := vm.GetGlobal("size_int").(*runtime.PyInt).Value
	sizeStr := vm.GetGlobal("size_str").(*runtime.PyInt).Value
	sizeList := vm.GetGlobal("size_list").(*runtime.PyInt).Value

	assert.True(t, sizeInt > 0)
	assert.True(t, sizeStr > 0)
	assert.True(t, sizeList > 0)
}

func TestFromSysImport(t *testing.T) {
	vm := runCodeWithStdlib(t, `
from sys import version, platform, maxsize
v = version
p = platform
m = maxsize
`)
	v := vm.GetGlobal("v").(*runtime.PyString).Value
	p := vm.GetGlobal("p").(*runtime.PyString).Value
	m := vm.GetGlobal("m").(*runtime.PyInt).Value

	assert.Contains(t, v, "3.14")
	assert.NotEmpty(t, p)
	assert.True(t, m > 0)
}

func TestSysByteorder(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import sys
byteorder = sys.byteorder
`)
	byteorder := vm.GetGlobal("byteorder").(*runtime.PyString).Value
	assert.Contains(t, []string{"little", "big"}, byteorder)
}

// =====================================
// Time Module Tests
// =====================================

func TestTimeTime(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import time
t1 = time.time()
t2 = time.time()
`)
	t1 := vm.GetGlobal("t1").(*runtime.PyFloat).Value
	t2 := vm.GetGlobal("t2").(*runtime.PyFloat).Value

	assert.True(t, t1 > 0)
	assert.True(t, t2 >= t1)
}

func TestTimeTimeNs(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import time
ns = time.time_ns()
`)
	ns := vm.GetGlobal("ns").(*runtime.PyInt).Value
	assert.True(t, ns > 0)
}

func TestTimeLocaltime(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import time
lt = time.localtime()
year = lt[0]
month = lt[1]
day = lt[2]
`)
	year := vm.GetGlobal("year").(*runtime.PyInt).Value
	month := vm.GetGlobal("month").(*runtime.PyInt).Value
	day := vm.GetGlobal("day").(*runtime.PyInt).Value

	assert.True(t, year >= 2020)
	assert.True(t, month >= 1 && month <= 12)
	assert.True(t, day >= 1 && day <= 31)
}

func TestTimeStrftime(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import time
lt = time.localtime()
formatted = time.strftime("%Y-%m-%d", lt)
`)
	formatted := vm.GetGlobal("formatted").(*runtime.PyString).Value
	assert.Regexp(t, `^\d{4}-\d{2}-\d{2}$`, formatted)
}

func TestTimePerfCounter(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import time
pc1 = time.perf_counter()
pc2 = time.perf_counter()
`)
	pc1 := vm.GetGlobal("pc1").(*runtime.PyFloat).Value
	pc2 := vm.GetGlobal("pc2").(*runtime.PyFloat).Value

	assert.True(t, pc1 >= 0)
	assert.True(t, pc2 >= pc1)
}

func TestFromTimeImport(t *testing.T) {
	vm := runCodeWithStdlib(t, `
from time import time, localtime
t = time()
lt = localtime()
`)
	ti := vm.GetGlobal("t").(*runtime.PyFloat).Value
	assert.True(t, ti > 0)
}

// =====================================
// Re Module Tests
// =====================================

func TestReMatch(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import re
result = re.search(r"^hello", "hello world")
matched = result is not None
`)
	matched := vm.GetGlobal("matched").(*runtime.PyBool).Value
	assert.True(t, matched)
}

func TestReSearch(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import re
m = re.search(r"world", "hello world")
matched = m is not None
`)
	matched := vm.GetGlobal("matched").(*runtime.PyBool).Value
	assert.True(t, matched)
}

func TestReFindall(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import re
matches = re.findall(r"\d+", "abc123def456ghi789")
count = len(matches)
`)
	count := vm.GetGlobal("count").(*runtime.PyInt).Value
	assert.Equal(t, int64(3), count)
}

func TestReSub(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import re
result = re.sub(r"\d+", "X", "abc123def456")
`)
	result := vm.GetGlobal("result").(*runtime.PyString).Value
	assert.Equal(t, "abcXdefX", result)
}

func TestReSplit(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import re
parts = re.split(r"[,;]", "a,b;c,d")
count = len(parts)
`)
	count := vm.GetGlobal("count").(*runtime.PyInt).Value
	assert.Equal(t, int64(4), count)
}

func TestReCompile(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import re
pattern = re.compile(r"\d+")
matches = pattern.findall("abc123def456")
count = len(matches)
`)
	count := vm.GetGlobal("count").(*runtime.PyInt).Value
	assert.Equal(t, int64(2), count)
}

func TestReEscape(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import re
escaped = re.escape("hello.world")
`)
	escaped := vm.GetGlobal("escaped").(*runtime.PyString).Value
	assert.Contains(t, escaped, `\.`)
}

func TestFromReImport(t *testing.T) {
	vm := runCodeWithStdlib(t, `
from re import search, findall, sub
result = search(r"^hello", "hello world")
matched = result is not None
`)
	matched := vm.GetGlobal("matched").(*runtime.PyBool).Value
	assert.True(t, matched)
}

// =====================================
// Collections Module Tests
// =====================================

func TestCollectionsCounter(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import collections
c = collections.Counter([1, 1, 1, 2, 2, 3])
total = c.total()
`)
	total := vm.GetGlobal("total").(*runtime.PyInt).Value
	assert.Equal(t, int64(6), total)
}

func TestCollectionsCounterMostCommon(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import collections
c = collections.Counter([1, 1, 1, 2, 2, 3])
mc = c.most_common(2)
count = len(mc)
`)
	count := vm.GetGlobal("count").(*runtime.PyInt).Value
	assert.Equal(t, int64(2), count)
}

func TestCollectionsDeque(t *testing.T) {
	runCodeWithStdlib(t, `
import collections
d = collections.deque([1, 2, 3])
d.append(4)
d.appendleft(0)
`)
}

func TestCollectionsDequePop(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import collections
d = collections.deque([1, 2, 3])
right = d.pop()
left = d.popleft()
`)
	right := vm.GetGlobal("right").(*runtime.PyInt).Value
	left := vm.GetGlobal("left").(*runtime.PyInt).Value

	assert.Equal(t, int64(3), right)
	assert.Equal(t, int64(1), left)
}

func TestCollectionsDequeRotate(t *testing.T) {
	runCodeWithStdlib(t, `
import collections
d = collections.deque([1, 2, 3, 4, 5])
d.rotate(2)
`)
}

func TestCollectionsOrderedDict(t *testing.T) {
	vm := runCodeWithStdlib(t, `
import collections
od = collections.OrderedDict()
`)
	od := vm.GetGlobal("od")
	_, ok := od.(*runtime.PyDict)
	assert.True(t, ok, "OrderedDict should be a dict")
}

func TestFromCollectionsImport(t *testing.T) {
	runCodeWithStdlib(t, `
from collections import Counter, deque
c = Counter([1, 1, 2])
d = deque([1, 2, 3])
`)
}
