from test_framework import test, expect

# Test 1: ExceptionGroup construction
def test_eg_construction():
    eg = ExceptionGroup("errors", [ValueError("v"), TypeError("t")])
    expect(eg.message).to_be("errors")
    expect(len(eg.exceptions)).to_be(2)
test("ExceptionGroup construction", test_eg_construction)

# Test 2: ExceptionGroup is instance of correct types
def test_eg_isinstance():
    eg = ExceptionGroup("eg", [ValueError("v")])
    expect(isinstance(eg, ExceptionGroup)).to_be(True)
    expect(isinstance(eg, BaseExceptionGroup)).to_be(True)
    expect(isinstance(eg, Exception)).to_be(True)
test("ExceptionGroup isinstance checks", test_eg_isinstance)

# Test 3: ExceptionGroup str representation
def test_eg_str():
    eg = ExceptionGroup("my errors", [ValueError("a"), ValueError("b")])
    s = str(eg)
    expect(s).to_be("my errors (2 sub-exceptions)")
test("ExceptionGroup str", test_eg_str)

# Test 4: ExceptionGroup str with 1 sub-exception
def test_eg_str_single():
    eg = ExceptionGroup("err", [ValueError("a")])
    s = str(eg)
    expect(s).to_be("err (1 sub-exception)")
test("ExceptionGroup str single", test_eg_str_single)

# Test 5: except* single type match
def test_except_star_single():
    result = None
    try:
        raise ExceptionGroup("eg", [ValueError("v1")])
    except* ValueError as eg:
        result = "caught"
    expect(result).to_be("caught")
test("except* single type match", test_except_star_single)

# Test 6: except* multiple clauses matching different types
def test_except_star_multi():
    caught_v = False
    caught_t = False
    try:
        raise ExceptionGroup("eg", [ValueError("v"), TypeError("t")])
    except* ValueError:
        caught_v = True
    except* TypeError:
        caught_t = True
    expect(caught_v).to_be(True)
    expect(caught_t).to_be(True)
test("except* multiple clauses", test_except_star_multi)

# Test 7: except* with named binding
def test_except_star_named():
    caught_eg = None
    try:
        raise ExceptionGroup("eg", [ValueError("val1"), ValueError("val2")])
    except* ValueError as eg:
        caught_eg = eg
    expect(caught_eg is not None).to_be(True)
test("except* with named binding", test_except_star_named)

# Test 8: except* all matched — no re-raise
def test_except_star_all_matched():
    outer_caught = False
    try:
        try:
            raise ExceptionGroup("eg", [ValueError("v")])
        except* ValueError:
            pass
    except Exception:
        outer_caught = True
    expect(outer_caught).to_be(False)
test("except* all matched no re-raise", test_except_star_all_matched)

# Test 9: except* partial match — unmatched re-raised
def test_except_star_partial():
    outer_caught = False
    try:
        try:
            raise ExceptionGroup("eg", [ValueError("v"), TypeError("t")])
        except* ValueError:
            pass
    except* TypeError:
        outer_caught = True
    expect(outer_caught).to_be(True)
test("except* partial match re-raises rest", test_except_star_partial)

# Test 10: except* none matched — all re-raised
def test_except_star_none_matched():
    outer_caught = False
    try:
        try:
            raise ExceptionGroup("eg", [ValueError("v"), TypeError("t")])
        except* KeyError:
            pass
    except Exception:
        outer_caught = True
    expect(outer_caught).to_be(True)
test("except* none matched re-raises all", test_except_star_none_matched)

# Test 11: subgroup returns matching subset
def test_subgroup():
    eg = ExceptionGroup("eg", [ValueError("v"), TypeError("t"), ValueError("v2")])
    sub = eg.subgroup(ValueError)
    expect(sub is not None).to_be(True)
    expect(len(sub.exceptions)).to_be(2)
test("subgroup returns matching subset", test_subgroup)

# Test 12: subgroup returns None when no match
def test_subgroup_none():
    eg = ExceptionGroup("eg", [ValueError("v")])
    sub = eg.subgroup(KeyError)
    expect(sub is None).to_be(True)
test("subgroup returns None for no match", test_subgroup_none)

# Test 13: split returns (match, rest) tuple
def test_split():
    eg = ExceptionGroup("eg", [ValueError("v"), TypeError("t")])
    matched, rest = eg.split(ValueError)
    expect(matched is not None).to_be(True)
    expect(rest is not None).to_be(True)
    expect(len(matched.exceptions)).to_be(1)
    expect(len(rest.exceptions)).to_be(1)
test("split returns match and rest", test_split)

# Test 14: split with all matching
def test_split_all():
    eg = ExceptionGroup("eg", [ValueError("v1"), ValueError("v2")])
    matched, rest = eg.split(ValueError)
    expect(matched is not None).to_be(True)
    expect(rest is None).to_be(True)
test("split all matching", test_split_all)

# Test 15: split with none matching
def test_split_none():
    eg = ExceptionGroup("eg", [ValueError("v1")])
    matched, rest = eg.split(KeyError)
    expect(matched is None).to_be(True)
    expect(rest is not None).to_be(True)
test("split none matching", test_split_none)

# Test 16: except* with finally block
def test_except_star_finally():
    finally_ran = False
    try:
        raise ExceptionGroup("eg", [ValueError("v")])
    except* ValueError:
        pass
    finally:
        finally_ran = True
    expect(finally_ran).to_be(True)
test("except* with finally", test_except_star_finally)

# Test 17: non-ExceptionGroup caught by except*
def test_except_star_non_group():
    caught = False
    try:
        raise ValueError("plain")
    except* ValueError:
        caught = True
    expect(caught).to_be(True)
test("except* catches non-group exception", test_except_star_non_group)

# Test 18: ExceptionGroup message attribute
def test_eg_message():
    eg = ExceptionGroup("test message", [ValueError("v")])
    expect(eg.message).to_be("test message")
test("ExceptionGroup message attribute", test_eg_message)

# Test 19: add_note basic usage
def test_add_note():
    e = ValueError("oops")
    e.add_note("extra context")
    expect(len(e.__notes__)).to_be(1)
    expect(e.__notes__[0]).to_be("extra context")
test("add_note basic usage", test_add_note)

# Test 20: add_note multiple notes
def test_add_note_multiple():
    e = TypeError("bad type")
    e.add_note("note 1")
    e.add_note("note 2")
    e.add_note("note 3")
    expect(len(e.__notes__)).to_be(3)
    expect(e.__notes__[0]).to_be("note 1")
    expect(e.__notes__[1]).to_be("note 2")
    expect(e.__notes__[2]).to_be("note 3")
test("add_note multiple notes", test_add_note_multiple)

# Test 21: __notes__ not present before add_note
def test_notes_not_present():
    e = ValueError("v")
    expect(hasattr(e, "__notes__")).to_be(False)
test("__notes__ not present before add_note", test_notes_not_present)

# Test 22: add_note in except block
def test_add_note_in_except():
    try:
        raise ValueError("original")
    except ValueError as e:
        e.add_note("caught and annotated")
        expect(len(e.__notes__)).to_be(1)
        expect(e.__notes__[0]).to_be("caught and annotated")
test("add_note in except block", test_add_note_in_except)

# Test 23: add_note preserved on re-raise
def test_add_note_reraise():
    note_found = False
    try:
        try:
            raise ValueError("inner")
        except ValueError as e:
            e.add_note("added in inner handler")
            raise
    except ValueError as e2:
        note_found = len(e2.__notes__) == 1 and e2.__notes__[0] == "added in inner handler"
    expect(note_found).to_be(True)
test("add_note preserved on re-raise", test_add_note_reraise)

# Test 24: add_note before raise
def test_add_note_before_raise():
    e = ValueError("pre-annotated")
    e.add_note("added before raise")
    try:
        raise e
    except ValueError as caught:
        expect(len(caught.__notes__)).to_be(1)
        expect(caught.__notes__[0]).to_be("added before raise")
test("add_note before raise", test_add_note_before_raise)

# Test 25: add_note requires string argument
def test_add_note_type_check():
    e = ValueError("v")
    got_error = False
    try:
        e.add_note(42)
    except TypeError:
        got_error = True
    expect(got_error).to_be(True)
test("add_note requires string argument", test_add_note_type_check)
