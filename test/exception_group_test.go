package test

import (
	"testing"

	"github.com/ATSOTECK/rage/internal/compiler"
	"github.com/ATSOTECK/rage/internal/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- ExceptionGroup construction ---

func TestExceptionGroupConstruction(t *testing.T) {
	vm := runCode(t, `
eg = ExceptionGroup("errors", [ValueError("v"), TypeError("t")])
msg = eg.message
count = len(eg.exceptions)
`)
	msg := vm.GetGlobal("msg")
	assert.Equal(t, "errors", msg.(*runtime.PyString).Value)
	count := vm.GetGlobal("count")
	assert.Equal(t, int64(2), count.(*runtime.PyInt).Value)
}

func TestExceptionGroupStr(t *testing.T) {
	vm := runCode(t, `
eg = ExceptionGroup("my errors", [ValueError("a"), ValueError("b")])
s = str(eg)
`)
	s := vm.GetGlobal("s")
	assert.Equal(t, "my errors (2 sub-exceptions)", s.(*runtime.PyString).Value)
}

func TestExceptionGroupStrSingle(t *testing.T) {
	vm := runCode(t, `
eg = ExceptionGroup("err", [ValueError("a")])
s = str(eg)
`)
	s := vm.GetGlobal("s")
	assert.Equal(t, "err (1 sub-exception)", s.(*runtime.PyString).Value)
}

func TestExceptionGroupIsInstance(t *testing.T) {
	vm := runCode(t, `
eg = ExceptionGroup("eg", [ValueError("v")])
is_eg = isinstance(eg, ExceptionGroup)
is_beg = isinstance(eg, BaseExceptionGroup)
is_exc = isinstance(eg, Exception)
`)
	assert.Equal(t, runtime.True, vm.GetGlobal("is_eg"))
	assert.Equal(t, runtime.True, vm.GetGlobal("is_beg"))
	assert.Equal(t, runtime.True, vm.GetGlobal("is_exc"))
}

func TestExceptionGroupRequiresString(t *testing.T) {
	runCodeExpectError(t, `
eg = ExceptionGroup(123, [ValueError("v")])
`, "message must be a string")
}

func TestExceptionGroupRequiresNonEmpty(t *testing.T) {
	runCodeExpectError(t, `
eg = ExceptionGroup("eg", [])
`, "must be non-empty")
}

// --- ExceptionGroup methods ---

func TestExceptionGroupSubgroup(t *testing.T) {
	vm := runCode(t, `
eg = ExceptionGroup("eg", [ValueError("v"), TypeError("t"), ValueError("v2")])
sub = eg.subgroup(ValueError)
has_sub = sub is not None
sub_count = len(sub.exceptions)
`)
	assert.Equal(t, runtime.True, vm.GetGlobal("has_sub"))
	assert.Equal(t, int64(2), vm.GetGlobal("sub_count").(*runtime.PyInt).Value)
}

func TestExceptionGroupSubgroupNone(t *testing.T) {
	vm := runCode(t, `
eg = ExceptionGroup("eg", [ValueError("v")])
sub = eg.subgroup(KeyError)
is_none = sub is None
`)
	assert.Equal(t, runtime.True, vm.GetGlobal("is_none"))
}

func TestExceptionGroupSplit(t *testing.T) {
	vm := runCode(t, `
eg = ExceptionGroup("eg", [ValueError("v"), TypeError("t")])
matched, rest = eg.split(ValueError)
has_matched = matched is not None
has_rest = rest is not None
matched_count = len(matched.exceptions)
rest_count = len(rest.exceptions)
`)
	assert.Equal(t, runtime.True, vm.GetGlobal("has_matched"))
	assert.Equal(t, runtime.True, vm.GetGlobal("has_rest"))
	assert.Equal(t, int64(1), vm.GetGlobal("matched_count").(*runtime.PyInt).Value)
	assert.Equal(t, int64(1), vm.GetGlobal("rest_count").(*runtime.PyInt).Value)
}

func TestExceptionGroupSplitAllMatch(t *testing.T) {
	vm := runCode(t, `
eg = ExceptionGroup("eg", [ValueError("v1"), ValueError("v2")])
matched, rest = eg.split(ValueError)
has_matched = matched is not None
rest_is_none = rest is None
`)
	assert.Equal(t, runtime.True, vm.GetGlobal("has_matched"))
	assert.Equal(t, runtime.True, vm.GetGlobal("rest_is_none"))
}

func TestExceptionGroupSplitNoneMatch(t *testing.T) {
	vm := runCode(t, `
eg = ExceptionGroup("eg", [ValueError("v")])
matched, rest = eg.split(KeyError)
matched_is_none = matched is None
has_rest = rest is not None
`)
	assert.Equal(t, runtime.True, vm.GetGlobal("matched_is_none"))
	assert.Equal(t, runtime.True, vm.GetGlobal("has_rest"))
}

func TestExceptionGroupDerive(t *testing.T) {
	vm := runCode(t, `
eg = ExceptionGroup("eg", [ValueError("v"), TypeError("t")])
derived = eg.derive([ValueError("new")])
msg = derived.message
count = len(derived.exceptions)
`)
	assert.Equal(t, "eg", vm.GetGlobal("msg").(*runtime.PyString).Value)
	assert.Equal(t, int64(1), vm.GetGlobal("count").(*runtime.PyInt).Value)
}

// --- except* basic matching ---

func TestExceptStarSingleMatch(t *testing.T) {
	vm := runCode(t, `
result = "not caught"
try:
    raise ExceptionGroup("eg", [ValueError("v1")])
except* ValueError as eg:
    result = "caught"
`)
	assert.Equal(t, "caught", vm.GetGlobal("result").(*runtime.PyString).Value)
}

func TestExceptStarMultipleClauses(t *testing.T) {
	vm := runCode(t, `
caught_v = False
caught_t = False
try:
    raise ExceptionGroup("eg", [ValueError("v"), TypeError("t")])
except* ValueError:
    caught_v = True
except* TypeError:
    caught_t = True
`)
	assert.Equal(t, runtime.True, vm.GetGlobal("caught_v"))
	assert.Equal(t, runtime.True, vm.GetGlobal("caught_t"))
}

func TestExceptStarNamedBinding(t *testing.T) {
	vm := runCode(t, `
caught = None
try:
    raise ExceptionGroup("eg", [ValueError("val1"), ValueError("val2")])
except* ValueError as eg:
    caught = eg
has_caught = caught is not None
`)
	assert.Equal(t, runtime.True, vm.GetGlobal("has_caught"))
}

func TestExceptStarAllMatchedNoReraise(t *testing.T) {
	vm := runCode(t, `
outer_caught = False
try:
    try:
        raise ExceptionGroup("eg", [ValueError("v")])
    except* ValueError:
        pass
except Exception:
    outer_caught = True
`)
	assert.Equal(t, runtime.False, vm.GetGlobal("outer_caught"))
}

func TestExceptStarPartialMatchReraiseRest(t *testing.T) {
	vm := runCode(t, `
outer_caught = False
try:
    try:
        raise ExceptionGroup("eg", [ValueError("v"), TypeError("t")])
    except* ValueError:
        pass
except* TypeError:
    outer_caught = True
`)
	assert.Equal(t, runtime.True, vm.GetGlobal("outer_caught"))
}

func TestExceptStarNoneMatchedReraiseAll(t *testing.T) {
	vm := runCode(t, `
outer_caught = False
try:
    try:
        raise ExceptionGroup("eg", [ValueError("v"), TypeError("t")])
    except* KeyError:
        pass
except Exception:
    outer_caught = True
`)
	assert.Equal(t, runtime.True, vm.GetGlobal("outer_caught"))
}

func TestExceptStarWithFinally(t *testing.T) {
	vm := runCode(t, `
finally_ran = False
try:
    raise ExceptionGroup("eg", [ValueError("v")])
except* ValueError:
    pass
finally:
    finally_ran = True
`)
	assert.Equal(t, runtime.True, vm.GetGlobal("finally_ran"))
}

func TestExceptStarNonGroupException(t *testing.T) {
	vm := runCode(t, `
caught = False
try:
    raise ValueError("plain")
except* ValueError:
    caught = True
`)
	assert.Equal(t, runtime.True, vm.GetGlobal("caught"))
}

// --- except* inside functions ---

func TestExceptStarInFunction(t *testing.T) {
	vm := runCode(t, `
def f():
    caught_v = False
    caught_t = False
    try:
        raise ExceptionGroup("eg", [ValueError("v"), TypeError("t")])
    except* ValueError:
        caught_v = True
    except* TypeError:
        caught_t = True
    return (caught_v, caught_t)
result = f()
`)
	result := vm.GetGlobal("result").(*runtime.PyTuple)
	assert.Equal(t, runtime.True, result.Items[0])
	assert.Equal(t, runtime.True, result.Items[1])
}

func TestExceptStarNestedInFunction(t *testing.T) {
	vm := runCode(t, `
def f():
    outer_caught = False
    try:
        try:
            raise ExceptionGroup("eg", [ValueError("v"), TypeError("t")])
        except* ValueError:
            pass
    except* TypeError:
        outer_caught = True
    return outer_caught
result = f()
`)
	assert.Equal(t, runtime.True, vm.GetGlobal("result"))
}

// --- except* with else ---

func TestExceptStarElseRuns(t *testing.T) {
	vm := runCode(t, `
else_ran = False
try:
    x = 1
except* ValueError:
    pass
else:
    else_ran = True
`)
	assert.Equal(t, runtime.True, vm.GetGlobal("else_ran"))
}

func TestExceptStarElseNotRunOnException(t *testing.T) {
	vm := runCode(t, `
else_ran = False
try:
    raise ExceptionGroup("eg", [ValueError("v")])
except* ValueError:
    pass
else:
    else_ran = True
`)
	assert.Equal(t, runtime.False, vm.GetGlobal("else_ran"))
}

// --- except* error cases ---

func TestExceptStarUncaughtPropagates(t *testing.T) {
	vm := runtime.NewVM()
	source := `
try:
    raise ExceptionGroup("eg", [ValueError("v")])
except* KeyError:
    pass
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)
	_, err := vm.Execute(code)
	require.Error(t, err)
}

func TestExceptStarMixedClausesError(t *testing.T) {
	source := `
try:
    pass
except ValueError:
    pass
except* TypeError:
    pass
`
	_, errs := compiler.CompileSource(source, "<test>")
	require.NotEmpty(t, errs, "expected error for mixing except and except*")
}

func TestExceptStarBareError(t *testing.T) {
	source := `
try:
    pass
except*:
    pass
`
	_, errs := compiler.CompileSource(source, "<test>")
	require.NotEmpty(t, errs, "expected error for bare except*")
}

// --- ExceptionGroup class registration ---

func TestExceptionGroupClassesExist(t *testing.T) {
	vm := runCode(t, `
beg = BaseExceptionGroup
eg = ExceptionGroup
beg_name = beg.__name__
eg_name = eg.__name__
`)
	assert.Equal(t, "BaseExceptionGroup", vm.GetGlobal("beg_name").(*runtime.PyString).Value)
	assert.Equal(t, "ExceptionGroup", vm.GetGlobal("eg_name").(*runtime.PyString).Value)
}

func TestExceptionGroupMessage(t *testing.T) {
	vm := runCode(t, `
eg = ExceptionGroup("test message", [ValueError("v")])
msg = eg.message
`)
	assert.Equal(t, "test message", vm.GetGlobal("msg").(*runtime.PyString).Value)
}

// --- Exception add_note / __notes__ ---

func TestExceptionAddNote(t *testing.T) {
	vm := runCode(t, `
e = ValueError("oops")
e.add_note("extra context")
notes = e.__notes__
count = len(notes)
first = notes[0]
`)
	assert.Equal(t, int64(1), vm.GetGlobal("count").(*runtime.PyInt).Value)
	assert.Equal(t, "extra context", vm.GetGlobal("first").(*runtime.PyString).Value)
}

func TestExceptionAddMultipleNotes(t *testing.T) {
	vm := runCode(t, `
e = TypeError("bad type")
e.add_note("note 1")
e.add_note("note 2")
e.add_note("note 3")
count = len(e.__notes__)
second = e.__notes__[1]
`)
	assert.Equal(t, int64(3), vm.GetGlobal("count").(*runtime.PyInt).Value)
	assert.Equal(t, "note 2", vm.GetGlobal("second").(*runtime.PyString).Value)
}

func TestExceptionNotesNotPresentBeforeAddNote(t *testing.T) {
	vm := runCode(t, `
e = ValueError("v")
has_notes = hasattr(e, "__notes__")
`)
	assert.Equal(t, runtime.False, vm.GetGlobal("has_notes"))
}

func TestExceptionNotesCaughtNotPresentBeforeAddNote(t *testing.T) {
	vm := runCode(t, `
try:
    raise ValueError("v")
except ValueError as e:
    has_notes = e.__notes__ is None
`)
	assert.Equal(t, runtime.True, vm.GetGlobal("has_notes"))
}

func TestExceptionAddNoteRequiresString(t *testing.T) {
	runCodeExpectError(t, `
e = ValueError("v")
e.add_note(42)
`, "note must be a str")
}

func TestExceptionAddNoteInExceptBlock(t *testing.T) {
	vm := runCode(t, `
try:
    raise ValueError("original")
except ValueError as e:
    e.add_note("caught and annotated")
    notes = e.__notes__
    count = len(notes)
    first = notes[0]
`)
	assert.Equal(t, int64(1), vm.GetGlobal("count").(*runtime.PyInt).Value)
	assert.Equal(t, "caught and annotated", vm.GetGlobal("first").(*runtime.PyString).Value)
}

func TestExceptionAddNotePreservedOnReraise(t *testing.T) {
	vm := runCode(t, `
note_found = False
try:
    try:
        raise ValueError("inner")
    except ValueError as e:
        e.add_note("added in inner handler")
        raise
except ValueError as e2:
    note_found = len(e2.__notes__) == 1 and e2.__notes__[0] == "added in inner handler"
`)
	assert.Equal(t, runtime.True, vm.GetGlobal("note_found"))
}

func TestExceptionAddNoteBeforeRaise(t *testing.T) {
	vm := runCode(t, `
e = ValueError("pre-annotated")
e.add_note("added before raise")
try:
    raise e
except ValueError as caught:
    count = len(caught.__notes__)
    first = caught.__notes__[0]
`)
	assert.Equal(t, int64(1), vm.GetGlobal("count").(*runtime.PyInt).Value)
	assert.Equal(t, "added before raise", vm.GetGlobal("first").(*runtime.PyString).Value)
}
