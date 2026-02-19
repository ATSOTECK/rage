# Test: Rich Comparisons
# Tests that __ne__, __le__, __ge__ are dispatched directly on user classes.

from test_framework import test, expect

class Cmp:
    def __init__(self, val):
        self.val = val

    def __eq__(self, other):
        return self.val == other.val

    def __ne__(self, other):
        # Custom __ne__: always returns "custom_ne" string (truthy)
        return "custom_ne"

    def __lt__(self, other):
        return self.val < other.val

    def __le__(self, other):
        return self.val <= other.val

    def __gt__(self, other):
        return self.val > other.val

    def __ge__(self, other):
        return self.val >= other.val


def test_ne_dispatch():
    """__ne__ is dispatched directly, not synthesized from __eq__"""
    a = Cmp(1)
    b = Cmp(2)
    # __ne__ returns "custom_ne" (truthy), so != should be True
    result = a != b
    expect(result).to_be("custom_ne")

    # Even for equal values, our __ne__ returns "custom_ne"
    c = Cmp(1)
    result2 = a != c
    expect(result2).to_be("custom_ne")

test("__ne__ dispatch", test_ne_dispatch)


def test_le_dispatch():
    """__le__ is dispatched directly"""
    a = Cmp(3)
    b = Cmp(5)
    expect(a <= b).to_be(True)
    expect(b <= a).to_be(False)
    expect(a <= Cmp(3)).to_be(True)

test("__le__ dispatch", test_le_dispatch)


def test_ge_dispatch():
    """__ge__ is dispatched directly"""
    a = Cmp(5)
    b = Cmp(3)
    expect(a >= b).to_be(True)
    expect(b >= a).to_be(False)
    expect(a >= Cmp(5)).to_be(True)

test("__ge__ dispatch", test_ge_dispatch)


class OnlyNe:
    """Class that only defines __ne__, not __eq__"""
    def __init__(self, val):
        self.val = val

    def __ne__(self, other):
        return self.val != other.val


def test_ne_without_eq():
    """__ne__ works independently of __eq__"""
    a = OnlyNe(1)
    b = OnlyNe(2)
    expect(a != b).to_be(True)

    c = OnlyNe(1)
    expect(a != c).to_be(False)

test("__ne__ without __eq__", test_ne_without_eq)


class OnlyLe:
    """Class with only __le__"""
    def __init__(self, val):
        self.val = val

    def __le__(self, other):
        return self.val <= other.val


def test_le_only():
    """__le__ works without __lt__"""
    a = OnlyLe(1)
    b = OnlyLe(2)
    expect(a <= b).to_be(True)
    expect(b <= a).to_be(False)

test("__le__ only", test_le_only)


class OnlyGe:
    """Class with only __ge__"""
    def __init__(self, val):
        self.val = val

    def __ge__(self, other):
        return self.val >= other.val


def test_ge_only():
    """__ge__ works without __gt__"""
    a = OnlyGe(5)
    b = OnlyGe(3)
    expect(a >= b).to_be(True)
    expect(b >= a).to_be(False)

test("__ge__ only", test_ge_only)


class ReflectedGe:
    """Class with __ge__ used as reflected __le__"""
    def __init__(self, val):
        self.val = val

    def __ge__(self, other):
        if not isinstance(other, ReflectedGe):
            return NotImplemented
        return self.val >= other.val


def test_reflected_le_via_ge():
    """a <= b tries b.__ge__(a) when a has no __le__"""
    a = ReflectedGe(2)
    b = ReflectedGe(5)
    # a <= b: a has no __le__, so tries b.__ge__(a) => 5 >= 2 => True
    expect(a <= b).to_be(True)
    expect(b <= a).to_be(False)

test("reflected __le__ via __ge__", test_reflected_le_via_ge)


def test_ne_in_if():
    """__ne__ dispatch works in if-statement (OpCompareNeJump)"""
    a = Cmp(1)
    b = Cmp(1)
    # __ne__ returns "custom_ne" (truthy), so even equal values are 'not equal'
    if a != b:
        result = "took_branch"
    else:
        result = "did_not"
    expect(result).to_be("took_branch")

test("__ne__ in if-statement", test_ne_in_if)


# ============================================================================
# Ported from CPython test_richcmp.py / test_compare.py
# ============================================================================

# --- Number class with all 6 comparison operators (from test_richcmp.py) ---

class Number:
    def __init__(self, x):
        self.x = x

    def __lt__(self, other):
        return self.x < other

    def __le__(self, other):
        return self.x <= other

    def __eq__(self, other):
        return self.x == other

    def __ne__(self, other):
        return self.x != other

    def __gt__(self, other):
        return self.x > other

    def __ge__(self, other):
        return self.x >= other

    def __repr__(self):
        return "Number(%r)" % (self.x, )


def test_number_basic():
    """Check that comparisons involving Number objects give the same
    results as comparing the corresponding ints"""
    for a in range(3):
        for b in range(3):
            for typea in (int, Number):
                for typeb in (int, Number):
                    if typea == int and typeb == int:
                        continue
                    ta = typea(a)
                    tb = typeb(b)
                    # Test all six operators via lambdas
                    ops = [
                        lambda x, y: x < y,
                        lambda x, y: x <= y,
                        lambda x, y: x == y,
                        lambda x, y: x != y,
                        lambda x, y: x > y,
                        lambda x, y: x >= y,
                    ]
                    for op in ops:
                        realoutcome = op(a, b)
                        testoutcome = op(ta, tb)
                        expect(realoutcome).to_be(testoutcome)

test("number basic comparisons", test_number_basic)


def test_number_values():
    """Check all operators and all comparison results for Number class"""
    # a == b (0, 0)
    expect(Number(0) < 0).to_be(False)
    expect(Number(0) <= 0).to_be(True)
    expect(Number(0) == 0).to_be(True)
    expect(Number(0) != 0).to_be(False)
    expect(Number(0) > 0).to_be(False)
    expect(Number(0) >= 0).to_be(True)

    # a < b (0, 1)
    expect(Number(0) < 1).to_be(True)
    expect(Number(0) <= 1).to_be(True)
    expect(Number(0) == 1).to_be(False)
    expect(Number(0) != 1).to_be(True)
    expect(Number(0) > 1).to_be(False)
    expect(Number(0) >= 1).to_be(False)

    # a > b (1, 0)
    expect(Number(1) < 0).to_be(False)
    expect(Number(1) <= 0).to_be(False)
    expect(Number(1) == 0).to_be(False)
    expect(Number(1) != 0).to_be(True)
    expect(Number(1) > 0).to_be(True)
    expect(Number(1) >= 0).to_be(True)

test("number comparison values", test_number_values)


# --- Misbehaving comparisons returning non-bool (from test_richcmp.py) ---

def test_misbehaving_comparisons():
    """Comparisons that return non-bool values (e.g. 0 instead of False)"""
    class Misb:
        def __lt__(self, other):
            return 0
        def __gt__(self, other):
            return 0
        def __eq__(self, other):
            return 0
        def __le__(self, other):
            raise AssertionError("This shouldn't happen")
        def __ge__(self, other):
            raise AssertionError("This shouldn't happen")
        def __ne__(self, other):
            raise AssertionError("This shouldn't happen")

    a = Misb()
    b = Misb()
    expect(a < b).to_be(0)
    expect(a == b).to_be(0)
    expect(a > b).to_be(0)

test("misbehaving comparisons return non-bool", test_misbehaving_comparisons)


# --- Exceptions in __bool__ propagated by not operator (from test_richcmp.py) ---

def test_not_propagates_bool_exception():
    """Check that exceptions in __bool__ are properly propagated by not"""
    class Exc(Exception):
        pass
    class Bad:
        def __bool__(self):
            raise Exc

    try:
        not Bad()
        expect("no error").to_be("error")
    except Exc:
        expect(True).to_be(True)

test("not operator propagates __bool__ exception", test_not_propagates_bool_exception)


# --- TypeError messages for unsupported comparisons (from test_richcmp.py) ---

def test_exception_message_int_none_lt():
    """42 < None raises TypeError"""
    try:
        42 < None
        expect("no error").to_be("error")
    except TypeError:
        expect(True).to_be(True)

test("TypeError for int < None", test_exception_message_int_none_lt)


def test_exception_message_none_int_lt():
    """None < 42 raises TypeError"""
    try:
        None < 42
        expect("no error").to_be("error")
    except TypeError:
        expect(True).to_be(True)

test("TypeError for None < int", test_exception_message_none_int_lt)


def test_exception_message_int_none_gt():
    """42 > None raises TypeError"""
    try:
        42 > None
        expect("no error").to_be("error")
    except TypeError:
        expect(True).to_be(True)

test("TypeError for int > None", test_exception_message_int_none_gt)


def test_exception_message_str_none_lt():
    """'foo' < None raises TypeError"""
    try:
        "foo" < None
        expect("no error").to_be("error")
    except TypeError:
        expect(True).to_be(True)

test("TypeError for str < None", test_exception_message_str_none_lt)


def test_exception_message_str_int_ge():
    """'foo' >= 666 raises TypeError"""
    try:
        "foo" >= 666
        expect("no error").to_be("error")
    except TypeError:
        expect(True).to_be(True)

test("TypeError for str >= int", test_exception_message_str_int_ge)


def test_exception_message_int_none_le():
    """42 <= None raises TypeError"""
    try:
        42 <= None
        expect("no error").to_be("error")
    except TypeError:
        expect(True).to_be(True)

test("TypeError for int <= None", test_exception_message_int_none_le)


def test_exception_message_int_none_ge():
    """42 >= None raises TypeError"""
    try:
        42 >= None
        expect("no error").to_be("error")
    except TypeError:
        expect(True).to_be(True)

test("TypeError for int >= None", test_exception_message_int_none_ge)


def test_exception_message_int_list_lt():
    """42 < [] raises TypeError"""
    try:
        42 < []
        expect("no error").to_be("error")
    except TypeError:
        expect(True).to_be(True)

test("TypeError for int < list", test_exception_message_int_list_lt)


def test_exception_message_tuple_list_gt():
    """() > [] raises TypeError"""
    try:
        () > []
        expect("no error").to_be("error")
    except TypeError:
        expect(True).to_be(True)

test("TypeError for tuple > list", test_exception_message_tuple_list_gt)


def test_exception_message_none_none_ge():
    """None >= None raises TypeError"""
    try:
        None >= None
        expect("no error").to_be("error")
    except TypeError:
        expect(True).to_be(True)

test("TypeError for None >= None", test_exception_message_none_none_ge)


def test_exception_message_custom_class():
    """Custom class < int raises TypeError"""
    class Spam:
        pass
    try:
        Spam() < 42
        expect("no error").to_be("error")
    except TypeError:
        expect(True).to_be(True)

test("TypeError for custom class < int", test_exception_message_custom_class)


def test_exception_message_int_custom_class():
    """int < Custom class raises TypeError"""
    class Spam:
        pass
    try:
        42 < Spam()
        expect("no error").to_be("error")
    except TypeError:
        expect(True).to_be(True)

test("TypeError for int < custom class", test_exception_message_int_custom_class)


def test_exception_message_custom_le_custom():
    """Custom class <= Custom class raises TypeError"""
    class Spam:
        pass
    try:
        Spam() <= Spam()
        expect("no error").to_be("error")
    except TypeError:
        expect(True).to_be(True)

test("TypeError for custom class <= custom class", test_exception_message_custom_le_custom)


# --- List comparison behavior (from test_richcmp.py ListTest) ---

def test_list_coverage():
    """Exercise all comparisons for lists"""
    x = [42]
    expect(x < x).to_be(False)
    expect(x <= x).to_be(True)
    expect(x == x).to_be(True)
    expect(x != x).to_be(False)
    expect(x > x).to_be(False)
    expect(x >= x).to_be(True)

    y = [42, 42]
    expect(x < y).to_be(True)
    expect(x <= y).to_be(True)
    expect(x == y).to_be(False)
    expect(x != y).to_be(True)
    expect(x > y).to_be(False)
    expect(x >= y).to_be(False)

test("list comparison coverage", test_list_coverage)


def test_list_badentry():
    """Exceptions in item __eq__ are propagated in list comparisons"""
    class Exc(Exception):
        pass
    class Bad:
        def __eq__(self, other):
            raise Exc

    x = [Bad()]
    y = [Bad()]

    try:
        x == y
        expect("no error").to_be("error")
    except Exc:
        expect(True).to_be(True)

test("list comparison propagates item __eq__ exception", test_list_badentry)


def test_list_goodentry():
    """Custom __lt__ works in list element comparison"""
    class Good:
        def __lt__(self, other):
            return True

    x = [Good()]
    y = [Good()]
    expect(x < y).to_be(True)

test("list comparison with custom __lt__ on elements", test_list_goodentry)


# --- Dict equality (no ordering) (from test_richcmp.py DictTest) ---

def test_dict_equality():
    """Verify that __eq__ and __ne__ work for dicts"""
    d1 = {1: "a", 2: "b", 3: "c"}
    d2 = {3: "c", 1: "a", 2: "b"}
    d3 = {1: "a", 2: "b", 3: "d"}

    expect(d1 == d1).to_be(True)
    expect(d1 == d2).to_be(True)
    expect(d1 != d3).to_be(True)
    expect(d1 == d3).to_be(False)

test("dict equality", test_dict_equality)


def test_dict_ordering_raises():
    """Dict ordering comparisons (<, <=, >, >=) raise TypeError"""
    d1 = {1: "a"}
    d2 = {2: "b"}

    try:
        d1 < d2
        expect("no error").to_be("error")
    except TypeError:
        expect(True).to_be(True)

    try:
        d1 <= d2
        expect("no error").to_be("error")
    except TypeError:
        expect(True).to_be(True)

    try:
        d1 > d2
        expect("no error").to_be("error")
    except TypeError:
        expect(True).to_be(True)

    try:
        d1 >= d2
        expect("no error").to_be("error")
    except TypeError:
        expect(True).to_be(True)

test("dict ordering raises TypeError", test_dict_ordering_raises)


# --- NotImplemented returns falling back to reflected (from test_compare.py) ---

def test_not_implemented_fallback():
    """When __eq__ returns NotImplemented, reflected is tried"""
    class A:
        def __eq__(self, other):
            return NotImplemented

    class B:
        def __eq__(self, other):
            return True

    a = A()
    b = B()
    # a == b: A.__eq__ returns NotImplemented, so tries B.__eq__(a) => True
    expect(a == b).to_be(True)
    expect(b == a).to_be(True)

test("NotImplemented fallback to reflected __eq__", test_not_implemented_fallback)


def test_not_implemented_lt_fallback():
    """When __lt__ returns NotImplemented, reflected __gt__ is tried"""
    class A:
        def __lt__(self, other):
            return NotImplemented

    class B:
        def __gt__(self, other):
            return True

    a = A()
    b = B()
    # a < b: A.__lt__ returns NotImplemented, so tries B.__gt__(a) => True
    expect(a < b).to_be(True)

test("NotImplemented __lt__ fallback to reflected __gt__", test_not_implemented_lt_fallback)


def test_not_implemented_le_fallback():
    """When __le__ returns NotImplemented, reflected __ge__ is tried"""
    class A:
        def __le__(self, other):
            return NotImplemented

    class B:
        def __ge__(self, other):
            return True

    a = A()
    b = B()
    expect(a <= b).to_be(True)

test("NotImplemented __le__ fallback to reflected __ge__", test_not_implemented_le_fallback)


# --- __ne__ defaulting to not __eq__ (from test_compare.py) ---

def test_ne_defaults_to_not_eq():
    """__ne__ defaults to not __eq__ when no custom __ne__ is defined"""
    class CmpEq:
        def __init__(self, arg):
            self.arg = arg
        def __eq__(self, other):
            return self.arg == other

    a = CmpEq(1)
    b = CmpEq(1)
    c = CmpEq(2)
    expect(a == b).to_be(True)
    expect(a != b).to_be(False)
    expect(a != c).to_be(True)

test("__ne__ defaults to not __eq__", test_ne_defaults_to_not_eq)


# --- __ne__ high priority: reflected __ne__ tried when __eq__ returns NotImplemented ---
# (from test_compare.py test_ne_high_priority)

def test_ne_high_priority():
    """object.__ne__() should allow reflected __ne__() to be tried"""
    calls = []
    class Left:
        def __eq__(self, other):
            calls.append("Left.__eq__")
            return NotImplemented
    class Right:
        def __eq__(self, other):
            calls.append("Right.__eq__")
            return NotImplemented
        def __ne__(self, other):
            calls.append("Right.__ne__")
            return NotImplemented

    Left() != Right()
    expect(calls).to_be(["Left.__eq__", "Right.__ne__"])

test("__ne__ high priority - reflected __ne__ tried", test_ne_high_priority)


# --- __ne__ low priority: object.__ne__() should not invoke reflected __eq__() ---
# (from test_compare.py test_ne_low_priority)

def test_ne_low_priority():
    """object.__ne__() should not invoke reflected __eq__()"""
    calls = []
    class Base:
        def __eq__(self, other):
            calls.append("Base.__eq__")
            return NotImplemented
    class Derived(Base):
        def __eq__(self, other):
            calls.append("Derived.__eq__")
            return NotImplemented
        def __ne__(self, other):
            calls.append("Derived.__ne__")
            return NotImplemented

    Base() != Derived()
    expect(calls).to_be(["Derived.__ne__", "Base.__eq__"])

test("__ne__ low priority - no reflected __eq__ invocation", test_ne_low_priority)


# --- No default delegation between operations except __ne__ ---
# (from test_compare.py test_other_delegation)

def test_no_delegation_eq():
    """__eq__ should not delegate to other comparison methods"""
    class C:
        def __ne__(self, other):
            raise AssertionError("unexpected")
        def __lt__(self, other):
            raise AssertionError("unexpected")
        def __le__(self, other):
            raise AssertionError("unexpected")
        def __gt__(self, other):
            raise AssertionError("unexpected")
        def __ge__(self, other):
            raise AssertionError("unexpected")

    # __eq__ not defined, should use default identity comparison
    expect(C() == object()).to_be(False)

test("__eq__ does not delegate to other comparison ops", test_no_delegation_eq)


def test_no_delegation_lt():
    """__lt__ should not delegate to other comparison methods"""
    class C:
        def __ne__(self, other):
            raise AssertionError("unexpected")
        def __eq__(self, other):
            raise AssertionError("unexpected")
        def __le__(self, other):
            raise AssertionError("unexpected")
        def __gt__(self, other):
            raise AssertionError("unexpected")
        def __ge__(self, other):
            raise AssertionError("unexpected")

    try:
        C() < object()
        expect("no error").to_be("error")
    except TypeError:
        expect(True).to_be(True)

test("__lt__ does not delegate to other comparison ops", test_no_delegation_lt)


def test_no_delegation_le():
    """__le__ should not delegate to other comparison methods"""
    class C:
        def __ne__(self, other):
            raise AssertionError("unexpected")
        def __eq__(self, other):
            raise AssertionError("unexpected")
        def __lt__(self, other):
            raise AssertionError("unexpected")
        def __gt__(self, other):
            raise AssertionError("unexpected")
        def __ge__(self, other):
            raise AssertionError("unexpected")

    try:
        C() <= object()
        expect("no error").to_be("error")
    except TypeError:
        expect(True).to_be(True)

test("__le__ does not delegate to other comparison ops", test_no_delegation_le)


def test_no_delegation_gt():
    """__gt__ should not delegate to other comparison methods"""
    class C:
        def __ne__(self, other):
            raise AssertionError("unexpected")
        def __eq__(self, other):
            raise AssertionError("unexpected")
        def __lt__(self, other):
            raise AssertionError("unexpected")
        def __le__(self, other):
            raise AssertionError("unexpected")
        def __ge__(self, other):
            raise AssertionError("unexpected")

    try:
        C() > object()
        expect("no error").to_be("error")
    except TypeError:
        expect(True).to_be(True)

test("__gt__ does not delegate to other comparison ops", test_no_delegation_gt)


def test_no_delegation_ge():
    """__ge__ should not delegate to other comparison methods"""
    class C:
        def __ne__(self, other):
            raise AssertionError("unexpected")
        def __eq__(self, other):
            raise AssertionError("unexpected")
        def __lt__(self, other):
            raise AssertionError("unexpected")
        def __le__(self, other):
            raise AssertionError("unexpected")
        def __gt__(self, other):
            raise AssertionError("unexpected")

    try:
        C() >= object()
        expect("no error").to_be("error")
    except TypeError:
        expect(True).to_be(True)

test("__ge__ does not delegate to other comparison ops", test_no_delegation_ge)


# --- Identity comparison as default (from test_compare.py test_id_comparisons) ---

def test_id_comparisons():
    """Default comparison uses identity (is) for objects without __eq__"""
    class Empty:
        pass

    objs = []
    for i in range(5):
        objs.append(Empty())

    for a in objs:
        for b in objs:
            expect(a == b).to_be(a is b)

test("default identity comparison", test_id_comparisons)


# --- Candidate set comparisons (from test_compare.py test_comparisons) ---

def test_candidate_set_comparisons():
    """Values that are 'equal' compare equal; different values don't"""
    class CmpVal:
        def __init__(self, arg):
            self.arg = arg
        def __eq__(self, other):
            return self.arg == other
        def __hash__(self):
            return hash(self.arg)

    class Empty:
        pass

    set1 = [2, 2.0, CmpVal(2.0)]
    set2 = [[1], (3,), None, Empty()]
    candidates = set1 + set2

    for a in candidates:
        for b in candidates:
            if ((a in set1) and (b in set1)) or a is b:
                expect(a == b).to_be(True)
            else:
                expect(a != b).to_be(True)

test("candidate set comparisons", test_candidate_set_comparisons)


# --- object comparison: equality only, no ordering (from test_compare.py test_objects) ---

def test_objects_equality_only():
    """object instances support == and != but not ordering"""
    a = object()
    b = object()

    # Same object
    expect(a == a).to_be(True)
    expect(a != a).to_be(False)

    # Different objects
    expect(a == b).to_be(False)
    expect(a != b).to_be(True)

    # Ordering should raise TypeError
    try:
        a < b
        expect("no error").to_be("error")
    except TypeError:
        expect(True).to_be(True)

    try:
        a <= b
        expect("no error").to_be("error")
    except TypeError:
        expect(True).to_be(True)

    try:
        a > b
        expect("no error").to_be("error")
    except TypeError:
        expect(True).to_be(True)

    try:
        a >= b
        expect("no error").to_be("error")
    except TypeError:
        expect(True).to_be(True)

test("object equality only, no ordering", test_objects_equality_only)


# --- Tuple comparison behavior (from test_compare.py test_sequences) ---

def test_tuple_comparisons():
    """Tuples support full ordering"""
    t1 = (1, 2)
    t2 = (2, 3)

    expect(t1 == t1).to_be(True)
    expect(t1 != t1).to_be(False)
    expect(t1 < t1).to_be(False)
    expect(t1 <= t1).to_be(True)
    expect(t1 > t1).to_be(False)
    expect(t1 >= t1).to_be(True)

    expect(t1 < t2).to_be(True)
    expect(t1 <= t2).to_be(True)
    expect(t1 == t2).to_be(False)
    expect(t1 != t2).to_be(True)
    expect(t1 > t2).to_be(False)
    expect(t1 >= t2).to_be(False)

test("tuple comparisons", test_tuple_comparisons)


def test_list_tuple_not_equal():
    """Lists and tuples with same elements are not equal"""
    t = (1, 2)
    l = [1, 2]
    expect(t == l).to_be(False)
    expect(t != l).to_be(True)
    expect(l == t).to_be(False)
    expect(l != t).to_be(True)

test("list and tuple not equal", test_list_tuple_not_equal)


# --- Comparison classes with various method subsets (from test_compare.py) ---

def test_comp_eq_only():
    """Class with only __eq__ supports equality but not ordering"""
    class CompEq:
        def __init__(self, x):
            self.x = x
        def __eq__(self, other):
            return self.x == other.x

    a1 = CompEq(1)
    a2 = CompEq(1)
    b = CompEq(2)

    expect(a1 == a2).to_be(True)
    expect(a1 != a2).to_be(False)
    expect(a1 == b).to_be(False)
    expect(a1 != b).to_be(True)

    try:
        a1 < b
        expect("no error").to_be("error")
    except TypeError:
        expect(True).to_be(True)

test("class with only __eq__", test_comp_eq_only)


def test_comp_ne_only():
    """Class with only __ne__"""
    class CompNe:
        def __init__(self, x):
            self.x = x
        def __ne__(self, other):
            return self.x != other.x

    a1 = CompNe(1)
    a2 = CompNe(1)
    b = CompNe(2)

    expect(a1 != a2).to_be(False)
    expect(a1 != b).to_be(True)

test("class with only __ne__", test_comp_ne_only)


def test_comp_eq_ne():
    """Class with __eq__ and __ne__"""
    class CompEqNe:
        def __init__(self, x):
            self.x = x
        def __eq__(self, other):
            return self.x == other.x
        def __ne__(self, other):
            return self.x != other.x

    a1 = CompEqNe(1)
    a2 = CompEqNe(1)
    b = CompEqNe(2)

    expect(a1 == a2).to_be(True)
    expect(a1 != a2).to_be(False)
    expect(a1 == b).to_be(False)
    expect(a1 != b).to_be(True)

test("class with __eq__ and __ne__", test_comp_eq_ne)


def test_comp_lt_only():
    """Class with only __lt__: < works, reflected > works too"""
    class CompLt:
        def __init__(self, x):
            self.x = x
        def __lt__(self, other):
            return self.x < other.x

    a = CompLt(1)
    b = CompLt(2)

    expect(a < b).to_be(True)
    expect(b < a).to_be(False)
    # b > a should try a.__lt__(b) as reflected
    expect(b > a).to_be(True)
    expect(a > b).to_be(False)

test("class with only __lt__", test_comp_lt_only)


def test_comp_gt_only():
    """Class with only __gt__: > works, reflected < works too"""
    class CompGt:
        def __init__(self, x):
            self.x = x
        def __gt__(self, other):
            return self.x > other.x

    a = CompGt(1)
    b = CompGt(2)

    expect(b > a).to_be(True)
    expect(a > b).to_be(False)
    # a < b should try b.__gt__(a) as reflected
    expect(a < b).to_be(True)
    expect(b < a).to_be(False)

test("class with only __gt__", test_comp_gt_only)


def test_comp_lt_gt():
    """Class with __lt__ and __gt__"""
    class CompLtGt:
        def __init__(self, x):
            self.x = x
        def __lt__(self, other):
            return self.x < other.x
        def __gt__(self, other):
            return self.x > other.x

    a = CompLtGt(1)
    b = CompLtGt(2)
    c = CompLtGt(1)

    expect(a < b).to_be(True)
    expect(b > a).to_be(True)
    expect(a > b).to_be(False)
    expect(b < a).to_be(False)
    expect(a < c).to_be(False)
    expect(a > c).to_be(False)

test("class with __lt__ and __gt__", test_comp_lt_gt)


def test_comp_le_only():
    """Class with only __le__: <= works, reflected >= works too"""
    class CompLe:
        def __init__(self, x):
            self.x = x
        def __le__(self, other):
            return self.x <= other.x

    a = CompLe(1)
    b = CompLe(2)
    c = CompLe(1)

    expect(a <= b).to_be(True)
    expect(b <= a).to_be(False)
    expect(a <= c).to_be(True)
    # reflected: b >= a should try a.__le__(b)
    expect(b >= a).to_be(True)

test("class with only __le__", test_comp_le_only)


def test_comp_ge_only():
    """Class with only __ge__: >= works, reflected <= works too"""
    class CompGe:
        def __init__(self, x):
            self.x = x
        def __ge__(self, other):
            return self.x >= other.x

    a = CompGe(1)
    b = CompGe(2)
    c = CompGe(1)

    expect(b >= a).to_be(True)
    expect(a >= b).to_be(False)
    expect(a >= c).to_be(True)
    # reflected: a <= b should try b.__ge__(a)
    expect(a <= b).to_be(True)

test("class with only __ge__", test_comp_ge_only)


def test_comp_le_ge():
    """Class with __le__ and __ge__"""
    class CompLeGe:
        def __init__(self, x):
            self.x = x
        def __le__(self, other):
            return self.x <= other.x
        def __ge__(self, other):
            return self.x >= other.x

    a = CompLeGe(1)
    b = CompLeGe(2)
    c = CompLeGe(1)

    expect(a <= b).to_be(True)
    expect(b >= a).to_be(True)
    expect(a <= c).to_be(True)
    expect(a >= c).to_be(True)
    expect(a >= b).to_be(False)
    expect(b <= a).to_be(False)

test("class with __le__ and __ge__", test_comp_le_ge)


# --- Different-class comparison interactions (from test_compare.py) ---

def test_cross_class_eq():
    """Equality across different user classes with __eq__"""
    class A:
        def __init__(self, x):
            self.x = x
        def __eq__(self, other):
            return self.x == other.x

    class B:
        def __init__(self, x):
            self.x = x
        def __eq__(self, other):
            return self.x == other.x

    a = A(1)
    b1 = B(1)
    b2 = B(2)

    expect(a == b1).to_be(True)
    expect(a != b1).to_be(False)
    expect(a == b2).to_be(False)
    expect(a != b2).to_be(True)

test("cross-class equality comparison", test_cross_class_eq)


def test_cross_class_lt_gt():
    """Ordering across different user classes with __lt__ and __gt__"""
    class A:
        def __init__(self, x):
            self.x = x
        def __lt__(self, other):
            return self.x < other.x

    class B:
        def __init__(self, x):
            self.x = x
        def __gt__(self, other):
            return self.x > other.x

    a = A(1)
    b = B(2)

    expect(a < b).to_be(True)
    expect(b > a).to_be(True)

test("cross-class lt/gt comparison", test_cross_class_lt_gt)


# --- Number comparisons (from test_compare.py test_numbers) ---

def test_number_type_comparisons():
    """Compare int and float types"""
    i1 = 1001
    i2 = 1002

    expect(i1 == i1).to_be(True)
    expect(i1 < i2).to_be(True)
    expect(i1 <= i2).to_be(True)
    expect(i1 != i2).to_be(True)
    expect(i1 > i2).to_be(False)
    expect(i1 >= i2).to_be(False)

    f1 = 1001.0
    f2 = 1001.1

    expect(f1 == f1).to_be(True)
    expect(f1 < f2).to_be(True)
    expect(f1 <= f2).to_be(True)
    expect(f1 != f2).to_be(True)
    expect(f1 > f2).to_be(False)
    expect(f1 >= f2).to_be(False)

    # int and float with same value
    expect(i1 == f1).to_be(True)
    expect(i1 != f1).to_be(False)
    expect(i1 <= f1).to_be(True)
    expect(i1 >= f1).to_be(True)

test("int and float comparisons", test_number_type_comparisons)


# --- Complex number equality only (from test_compare.py) ---

def test_complex_equality_only():
    """Complex numbers support equality but not ordering"""
    c1 = 1001 + 0j
    c2 = 1001 + 1j

    expect(c1 == c1).to_be(True)
    expect(c1 != c1).to_be(False)
    expect(c1 == c2).to_be(False)
    expect(c1 != c2).to_be(True)

    # Ordering should raise TypeError
    try:
        c1 < c2
        expect("no error").to_be("error")
    except TypeError:
        expect(True).to_be(True)

    try:
        c1 <= c2
        expect("no error").to_be("error")
    except TypeError:
        expect(True).to_be(True)

    try:
        c1 > c2
        expect("no error").to_be("error")
    except TypeError:
        expect(True).to_be(True)

    try:
        c1 >= c2
        expect("no error").to_be("error")
    except TypeError:
        expect(True).to_be(True)

test("complex number equality only, no ordering", test_complex_equality_only)


# --- int/float and complex equality (from test_compare.py) ---

def test_int_complex_equality():
    """int and complex with same value are equal"""
    i = 1001
    c = 1001 + 0j

    expect(i == c).to_be(True)
    expect(c == i).to_be(True)
    expect(i != c).to_be(False)

test("int and complex equality", test_int_complex_equality)


# --- Mapping comparisons (from test_compare.py test_mappings) ---

def test_mapping_comparisons():
    """Dict equality with different insertion orders"""
    d1 = {1: "a", 2: "b"}
    d2 = {2: "b", 3: "c"}
    d3 = {3: "c", 2: "b"}

    expect(d1 == d1).to_be(True)
    expect(d1 != d1).to_be(False)
    expect(d1 == d2).to_be(False)
    expect(d1 != d2).to_be(True)
    expect(d2 == d3).to_be(True)
    expect(d2 != d3).to_be(False)

test("mapping comparisons", test_mapping_comparisons)


# --- Sequence comparisons (from test_compare.py test_sequences) ---

def test_sequence_comparisons():
    """List and tuple ordering comparisons"""
    l1 = [1, 2]
    l2 = [2, 3]
    expect(l1 < l2).to_be(True)
    expect(l1 <= l2).to_be(True)
    expect(l1 == l2).to_be(False)
    expect(l1 != l2).to_be(True)
    expect(l1 > l2).to_be(False)
    expect(l1 >= l2).to_be(False)

    # Same list
    expect(l1 == l1).to_be(True)
    expect(l1 <= l1).to_be(True)
    expect(l1 >= l1).to_be(True)
    expect(l1 < l1).to_be(False)
    expect(l1 > l1).to_be(False)

test("sequence comparisons", test_sequence_comparisons)


# --- Subclass priority in comparisons (from test_compare.py) ---

def test_subclass_priority():
    """Subclass comparison methods are tried first"""
    class Base:
        def __init__(self, x):
            self.x = x
        def __eq__(self, other):
            return self.x == other.x
        def __lt__(self, other):
            return self.x < other.x
        def __gt__(self, other):
            return self.x > other.x

    class Derived(Base):
        def __eq__(self, other):
            return self.x == other.x
        def __lt__(self, other):
            return self.x < other.x
        def __gt__(self, other):
            return self.x > other.x

    b = Base(1)
    d = Derived(2)

    # When comparing Base with Derived, Derived's methods get priority
    expect(b < d).to_be(True)
    expect(d > b).to_be(True)
    expect(b == d).to_be(False)
    expect(d == b).to_be(False)

test("subclass comparison priority", test_subclass_priority)


# --- String comparisons (from test_compare.py test_str_subclass) ---

def test_string_comparisons():
    """String ordering comparisons"""
    s1 = "a"
    s2 = "b"

    expect(s1 == s1).to_be(True)
    expect(s1 < s2).to_be(True)
    expect(s1 <= s2).to_be(True)
    expect(s1 != s2).to_be(True)
    expect(s1 > s2).to_be(False)
    expect(s1 >= s2).to_be(False)

    expect(s2 > s1).to_be(True)
    expect(s2 >= s1).to_be(True)

test("string comparisons", test_string_comparisons)


# --- Mixed type list comparisons ---

def test_list_different_lengths():
    """List comparison with different lengths"""
    expect([1, 2] < [1, 2, 3]).to_be(True)
    expect([1, 2, 3] > [1, 2]).to_be(True)
    expect([1, 2] == [1, 2, 3]).to_be(False)
    expect([1, 2] != [1, 2, 3]).to_be(True)
    expect([1, 2] <= [1, 2, 3]).to_be(True)
    expect([1, 2, 3] >= [1, 2]).to_be(True)

test("list comparison with different lengths", test_list_different_lengths)


def test_tuple_different_lengths():
    """Tuple comparison with different lengths"""
    expect((1, 2) < (1, 2, 3)).to_be(True)
    expect((1, 2, 3) > (1, 2)).to_be(True)
    expect((1, 2) == (1, 2, 3)).to_be(False)
    expect((1, 2) != (1, 2, 3)).to_be(True)
    expect((1, 2) <= (1, 2, 3)).to_be(True)
    expect((1, 2, 3) >= (1, 2)).to_be(True)

test("tuple comparison with different lengths", test_tuple_different_lengths)


# --- Empty container comparisons ---

def test_empty_container_comparisons():
    """Empty containers compare as expected"""
    expect([] == []).to_be(True)
    expect([] != []).to_be(False)
    expect([] < [1]).to_be(True)
    expect([] <= []).to_be(True)
    expect([] >= []).to_be(True)
    expect([] > []).to_be(False)

    expect(() == ()).to_be(True)
    expect(() != ()).to_be(False)
    expect(() < (1,)).to_be(True)
    expect(() <= ()).to_be(True)
    expect(() >= ()).to_be(True)
    expect(() > ()).to_be(False)

    expect({} == {}).to_be(True)
    expect({} != {}).to_be(False)

test("empty container comparisons", test_empty_container_comparisons)


# --- NotImplemented bilateral returns TypeError ---

def test_both_return_not_implemented():
    """When both sides return NotImplemented, appropriate behavior occurs"""
    class A:
        def __eq__(self, other):
            return NotImplemented
    class B:
        def __eq__(self, other):
            return NotImplemented

    a = A()
    b = B()
    # When both __eq__ return NotImplemented, falls back to identity
    expect(a == b).to_be(False)
    expect(a != b).to_be(True)

test("both return NotImplemented for __eq__", test_both_return_not_implemented)


def test_both_return_not_implemented_lt():
    """When both __lt__ and reflected __gt__ return NotImplemented, TypeError"""
    class A:
        def __lt__(self, other):
            return NotImplemented
    class B:
        def __gt__(self, other):
            return NotImplemented

    try:
        A() < B()
        expect("no error").to_be("error")
    except TypeError:
        expect(True).to_be(True)

test("both return NotImplemented for __lt__/__gt__ raises TypeError", test_both_return_not_implemented_lt)
