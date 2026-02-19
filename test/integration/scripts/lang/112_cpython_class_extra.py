# Test: CPython Class Feature Extra Tests
# Adapted from CPython's Lib/test/test_class.py
# Covers: all dunder dispatch (__add__,__sub__,etc.), __repr__/__str__,
# __hash__, __bool__/__len__ truthiness, __contains__, __getitem__/__setitem__/__delitem__,
# __call__, comparison dunders, unary operators

from test_framework import test, expect

# ============================================================================
# Binary operator dunder dispatch
# ============================================================================

def test_add_dispatch():
    """__add__ is called for + operator."""
    class A:
        def __init__(self, val):
            self.val = val
        def __add__(self, other):
            return A(self.val + other.val)
    result = A(3) + A(4)
    expect(result.val).to_be(7)

test("__add__ dispatch", test_add_dispatch)

def test_sub_dispatch():
    """__sub__ is called for - operator."""
    class A:
        def __init__(self, val):
            self.val = val
        def __sub__(self, other):
            return A(self.val - other.val)
    result = A(10) - A(3)
    expect(result.val).to_be(7)

test("__sub__ dispatch", test_sub_dispatch)

def test_mul_dispatch():
    """__mul__ is called for * operator."""
    class A:
        def __init__(self, val):
            self.val = val
        def __mul__(self, other):
            return A(self.val * other.val)
    result = A(3) * A(4)
    expect(result.val).to_be(12)

test("__mul__ dispatch", test_mul_dispatch)

def test_truediv_dispatch():
    """__truediv__ is called for / operator."""
    class A:
        def __init__(self, val):
            self.val = val
        def __truediv__(self, other):
            return A(self.val / other.val)
    result = A(10) / A(2)
    expect(result.val).to_be(5.0)

test("__truediv__ dispatch", test_truediv_dispatch)

def test_floordiv_dispatch():
    """__floordiv__ is called for // operator."""
    class A:
        def __init__(self, val):
            self.val = val
        def __floordiv__(self, other):
            return A(self.val // other.val)
    result = A(7) // A(2)
    expect(result.val).to_be(3)

test("__floordiv__ dispatch", test_floordiv_dispatch)

def test_mod_dispatch():
    """__mod__ is called for % operator."""
    class A:
        def __init__(self, val):
            self.val = val
        def __mod__(self, other):
            return A(self.val % other.val)
    result = A(7) % A(3)
    expect(result.val).to_be(1)

test("__mod__ dispatch", test_mod_dispatch)

def test_pow_dispatch():
    """__pow__ is called for ** operator."""
    class A:
        def __init__(self, val):
            self.val = val
        def __pow__(self, other):
            return A(self.val ** other.val)
    result = A(2) ** A(8)
    expect(result.val).to_be(256)

test("__pow__ dispatch", test_pow_dispatch)

def test_lshift_dispatch():
    """__lshift__ is called for << operator."""
    class A:
        def __init__(self, val):
            self.val = val
        def __lshift__(self, other):
            return A(self.val << other.val)
    result = A(1) << A(3)
    expect(result.val).to_be(8)

test("__lshift__ dispatch", test_lshift_dispatch)

def test_rshift_dispatch():
    """__rshift__ is called for >> operator."""
    class A:
        def __init__(self, val):
            self.val = val
        def __rshift__(self, other):
            return A(self.val >> other.val)
    result = A(16) >> A(2)
    expect(result.val).to_be(4)

test("__rshift__ dispatch", test_rshift_dispatch)

def test_and_dispatch():
    """__and__ is called for & operator."""
    class A:
        def __init__(self, val):
            self.val = val
        def __and__(self, other):
            return A(self.val & other.val)
    result = A(0b1100) & A(0b1010)
    expect(result.val).to_be(0b1000)

test("__and__ dispatch", test_and_dispatch)

def test_or_dispatch():
    """__or__ is called for | operator."""
    class A:
        def __init__(self, val):
            self.val = val
        def __or__(self, other):
            return A(self.val | other.val)
    result = A(0b1100) | A(0b1010)
    expect(result.val).to_be(0b1110)

test("__or__ dispatch", test_or_dispatch)

def test_xor_dispatch():
    """__xor__ is called for ^ operator."""
    class A:
        def __init__(self, val):
            self.val = val
        def __xor__(self, other):
            return A(self.val ^ other.val)
    result = A(0b1100) ^ A(0b1010)
    expect(result.val).to_be(0b0110)

test("__xor__ dispatch", test_xor_dispatch)

# ============================================================================
# Reflected (right-hand) binary operators
# ============================================================================

def test_radd_dispatch():
    """__radd__ is called when left operand doesn't support the op."""
    class A:
        def __init__(self, val):
            self.val = val
        def __radd__(self, other):
            return A(other + self.val)
    result = 10 + A(5)
    expect(result.val).to_be(15)

test("__radd__ dispatch", test_radd_dispatch)

def test_rsub_dispatch():
    """__rsub__ is called for reflected subtraction."""
    class A:
        def __init__(self, val):
            self.val = val
        def __rsub__(self, other):
            return A(other - self.val)
    result = 10 - A(3)
    expect(result.val).to_be(7)

test("__rsub__ dispatch", test_rsub_dispatch)

def test_rmul_dispatch():
    """__rmul__ is called for reflected multiplication."""
    class A:
        def __init__(self, val):
            self.val = val
        def __rmul__(self, other):
            return A(other * self.val)
    result = 3 * A(4)
    expect(result.val).to_be(12)

test("__rmul__ dispatch", test_rmul_dispatch)

def test_rtruediv_dispatch():
    """__rtruediv__ is called for reflected division."""
    class A:
        def __init__(self, val):
            self.val = val
        def __rtruediv__(self, other):
            return A(other / self.val)
    result = 10 / A(2)
    expect(result.val).to_be(5.0)

test("__rtruediv__ dispatch", test_rtruediv_dispatch)

def test_rfloordiv_dispatch():
    """__rfloordiv__ is called for reflected floor division."""
    class A:
        def __init__(self, val):
            self.val = val
        def __rfloordiv__(self, other):
            return A(other // self.val)
    result = 7 // A(2)
    expect(result.val).to_be(3)

test("__rfloordiv__ dispatch", test_rfloordiv_dispatch)

def test_rmod_dispatch():
    """__rmod__ is called for reflected modulo."""
    class A:
        def __init__(self, val):
            self.val = val
        def __rmod__(self, other):
            return A(other % self.val)
    result = 7 % A(3)
    expect(result.val).to_be(1)

test("__rmod__ dispatch", test_rmod_dispatch)

def test_rpow_dispatch():
    """__rpow__ is called for reflected power."""
    class A:
        def __init__(self, val):
            self.val = val
        def __rpow__(self, other):
            return A(other ** self.val)
    result = 2 ** A(10)
    expect(result.val).to_be(1024)

test("__rpow__ dispatch", test_rpow_dispatch)

# ============================================================================
# __repr__ and __str__
# ============================================================================

def test_repr_dispatch():
    """repr() calls __repr__."""
    class A:
        def __repr__(self):
            return "A(repr)"
    expect(repr(A())).to_be("A(repr)")

test("__repr__ dispatch", test_repr_dispatch)

def test_str_dispatch():
    """str() calls __str__."""
    class A:
        def __str__(self):
            return "A(str)"
    expect(str(A())).to_be("A(str)")

test("__str__ dispatch", test_str_dispatch)

def test_str_falls_back_to_repr():
    """str() falls back to __repr__ when __str__ is not defined."""
    class A:
        def __repr__(self):
            return "A(repr)"
    expect(str(A())).to_be("A(repr)")

test("str() falls back to __repr__", test_str_falls_back_to_repr)

def test_repr_and_str_different():
    """__repr__ and __str__ can return different values."""
    class A:
        def __repr__(self):
            return "repr_value"
        def __str__(self):
            return "str_value"
    a = A()
    expect(repr(a)).to_be("repr_value")
    expect(str(a)).to_be("str_value")

test("__repr__ and __str__ different", test_repr_and_str_different)

def test_repr_in_containers():
    """repr() is used when objects are in containers."""
    class A:
        def __init__(self, n):
            self.n = n
        def __repr__(self):
            return "A(" + str(self.n) + ")"
    result = repr([A(1), A(2)])
    expect(result).to_be("[A(1), A(2)]")

test("__repr__ used in container repr", test_repr_in_containers)

# ============================================================================
# __hash__ custom implementations
# ============================================================================

def test_hash_dispatch():
    """hash() calls __hash__."""
    class A:
        def __hash__(self):
            return 42
    expect(hash(A())).to_be(42)

test("__hash__ dispatch", test_hash_dispatch)

def test_hash_consistent():
    """Same hash for equal objects (custom)."""
    class Point:
        def __init__(self, x, y):
            self.x = x
            self.y = y
        def __hash__(self):
            return self.x * 31 + self.y
        def __eq__(self, other):
            if not isinstance(other, Point):
                return False
            return self.x == other.x and self.y == other.y
    p1 = Point(1, 2)
    p2 = Point(1, 2)
    expect(hash(p1)).to_be(hash(p2))
    expect(p1 == p2).to_be(True)

test("__hash__ consistent with __eq__", test_hash_consistent)

def test_hash_as_dict_key():
    """Objects with __hash__ and __eq__ can be dict keys."""
    class Key:
        def __init__(self, val):
            self.val = val
        def __hash__(self):
            return self.val
        def __eq__(self, other):
            if isinstance(other, Key):
                return self.val == other.val
            return False
    d = {}
    d[Key(1)] = "one"
    d[Key(2)] = "two"
    expect(d[Key(1)]).to_be("one")
    expect(d[Key(2)]).to_be("two")

test("__hash__ objects as dict keys", test_hash_as_dict_key)

def test_hash_returns_int():
    """__hash__ must return an integer."""
    class A:
        def __hash__(self):
            return 0
    # Should not raise
    expect(hash(A())).to_be(0)

test("__hash__ returns int", test_hash_returns_int)

# ============================================================================
# __bool__ / __len__ for truthiness
# ============================================================================

def test_bool_dispatch():
    """bool() calls __bool__."""
    class AlwaysTrue:
        def __bool__(self):
            return True
    class AlwaysFalse:
        def __bool__(self):
            return False
    expect(bool(AlwaysTrue())).to_be(True)
    expect(bool(AlwaysFalse())).to_be(False)

test("__bool__ dispatch", test_bool_dispatch)

def test_len_truthiness():
    """__len__ determines truthiness when __bool__ is absent."""
    class Empty:
        def __len__(self):
            return 0
    class NonEmpty:
        def __len__(self):
            return 5
    expect(bool(Empty())).to_be(False)
    expect(bool(NonEmpty())).to_be(True)

test("__len__ determines truthiness", test_len_truthiness)

def test_bool_overrides_len():
    """__bool__ takes precedence over __len__."""
    class Weird:
        def __len__(self):
            return 0
        def __bool__(self):
            return True
    expect(bool(Weird())).to_be(True)

test("__bool__ overrides __len__", test_bool_overrides_len)

def test_bool_in_if():
    """__bool__ is used in if statements."""
    class Falsy:
        def __bool__(self):
            return False
    class Truthy:
        def __bool__(self):
            return True
    if Truthy():
        r1 = "truthy"
    else:
        r1 = "falsy"
    if Falsy():
        r2 = "truthy"
    else:
        r2 = "falsy"
    expect(r1).to_be("truthy")
    expect(r2).to_be("falsy")

test("__bool__ in if statements", test_bool_in_if)

def test_len_in_if():
    """__len__ is used in if statements when no __bool__."""
    class Empty:
        def __len__(self):
            return 0
    class NonEmpty:
        def __len__(self):
            return 3
    if NonEmpty():
        r1 = "truthy"
    else:
        r1 = "falsy"
    if Empty():
        r2 = "truthy"
    else:
        r2 = "falsy"
    expect(r1).to_be("truthy")
    expect(r2).to_be("falsy")

test("__len__ in if statements", test_len_in_if)

def test_bool_with_and_or():
    """__bool__ affects and/or short-circuiting."""
    class Truthy:
        def __init__(self, val):
            self.val = val
        def __bool__(self):
            return True
    class Falsy:
        def __init__(self, val):
            self.val = val
        def __bool__(self):
            return False
    # 'and' returns first falsy or last truthy
    result = Truthy("a") and Truthy("b")
    expect(result.val).to_be("b")
    result = Falsy("a") and Truthy("b")
    expect(result.val).to_be("a")
    # 'or' returns first truthy or last falsy
    result = Truthy("a") or Truthy("b")
    expect(result.val).to_be("a")
    result = Falsy("a") or Falsy("b")
    expect(result.val).to_be("b")

test("__bool__ with and/or operators", test_bool_with_and_or)

# ============================================================================
# __contains__ on custom classes
# ============================================================================

def test_contains_dispatch():
    """__contains__ is called for 'in' operator."""
    class Box:
        def __init__(self, items):
            self.items = items
        def __contains__(self, item):
            return item in self.items
    b = Box([1, 2, 3])
    expect(1 in b).to_be(True)
    expect(4 in b).to_be(False)
    expect(4 not in b).to_be(True)

test("__contains__ dispatch", test_contains_dispatch)

def test_contains_with_custom_eq():
    """__contains__ interacts with custom __eq__."""
    class Wrapper:
        def __init__(self, val):
            self.val = val
        def __eq__(self, other):
            if isinstance(other, Wrapper):
                return self.val == other.val
            return self.val == other

    class Container:
        def __init__(self, items):
            self.items = items
        def __contains__(self, item):
            for i in self.items:
                if i == item:
                    return True
            return False

    c = Container([Wrapper(1), Wrapper(2), Wrapper(3)])
    expect(Wrapper(2) in c).to_be(True)
    expect(Wrapper(5) in c).to_be(False)

test("__contains__ with custom __eq__", test_contains_with_custom_eq)

def test_in_falls_back_to_iter():
    """When __contains__ is absent, 'in' falls back to __iter__."""
    class Iterable:
        def __init__(self, data):
            self.data = data
        def __iter__(self):
            return iter(self.data)
    it = Iterable([10, 20, 30])
    expect(20 in it).to_be(True)
    expect(40 in it).to_be(False)

test("in falls back to __iter__", test_in_falls_back_to_iter)

# ============================================================================
# __getitem__, __setitem__, __delitem__
# ============================================================================

def test_getitem_dispatch():
    """__getitem__ is called for obj[key]."""
    class A:
        def __getitem__(self, key):
            return key * 2
    a = A()
    expect(a[3]).to_be(6)
    expect(a["hi"]).to_be("hihi")

test("__getitem__ dispatch", test_getitem_dispatch)

def test_setitem_dispatch():
    """__setitem__ is called for obj[key] = val."""
    class A:
        def __init__(self):
            self.store = {}
        def __setitem__(self, key, val):
            self.store[key] = val
        def __getitem__(self, key):
            return self.store[key]
    a = A()
    a[1] = "one"
    a[2] = "two"
    expect(a[1]).to_be("one")
    expect(a[2]).to_be("two")

test("__setitem__ dispatch", test_setitem_dispatch)

def test_delitem_dispatch():
    """__delitem__ is called for del obj[key]."""
    deleted_keys = []
    class A:
        def __delitem__(self, key):
            deleted_keys.append(key)
    a = A()
    del a[1]
    del a["hello"]
    expect(deleted_keys).to_be([1, "hello"])

test("__delitem__ dispatch", test_delitem_dispatch)

def test_getitem_setitem_delitem_combined():
    """Combined __getitem__, __setitem__, __delitem__."""
    class DictLike:
        def __init__(self):
            self.data = {}
        def __getitem__(self, key):
            return self.data[key]
        def __setitem__(self, key, val):
            self.data[key] = val
        def __delitem__(self, key):
            del self.data[key]
        def __contains__(self, key):
            return key in self.data
    d = DictLike()
    d["x"] = 10
    expect(d["x"]).to_be(10)
    expect("x" in d).to_be(True)
    del d["x"]
    expect("x" in d).to_be(False)

test("combined getitem/setitem/delitem", test_getitem_setitem_delitem_combined)

def test_getitem_with_slice():
    """__getitem__ receives slice objects for slice notation."""
    class SliceAware:
        def __init__(self, data):
            self.data = data
        def __getitem__(self, key):
            return self.data[key]
    s = SliceAware([0, 1, 2, 3, 4])
    expect(s[1]).to_be(1)
    expect(s[1:4]).to_be([1, 2, 3])

test("__getitem__ with slice", test_getitem_with_slice)

# ============================================================================
# __call__ on instances
# ============================================================================

def test_call_dispatch():
    """__call__ makes instances callable."""
    class Adder:
        def __init__(self, base):
            self.base = base
        def __call__(self, x):
            return self.base + x
    add5 = Adder(5)
    expect(add5(10)).to_be(15)
    expect(add5(0)).to_be(5)

test("__call__ dispatch", test_call_dispatch)

def test_call_with_kwargs():
    """__call__ can receive keyword arguments."""
    class Formatter:
        def __call__(self, name, greeting="Hello"):
            return greeting + " " + name
    f = Formatter()
    expect(f("World")).to_be("Hello World")
    expect(f("World", greeting="Hi")).to_be("Hi World")

test("__call__ with kwargs", test_call_with_kwargs)

def test_call_with_varargs():
    """__call__ can receive *args."""
    class Summer:
        def __call__(self, *args):
            return sum(args)
    s = Summer()
    expect(s(1, 2, 3)).to_be(6)
    expect(s()).to_be(0)

test("__call__ with *args", test_call_with_varargs)

def test_callable_check():
    """callable() returns True for objects with __call__."""
    class WithCall:
        def __call__(self):
            pass
    class WithoutCall:
        pass
    expect(callable(WithCall())).to_be(True)
    expect(callable(WithoutCall())).to_be(False)
    expect(callable(lambda: None)).to_be(True)
    expect(callable(42)).to_be(False)

test("callable() check", test_callable_check)

def test_call_stateful():
    """__call__ can maintain state across calls."""
    class Counter:
        def __init__(self):
            self.count = 0
        def __call__(self):
            self.count = self.count + 1
            return self.count
    c = Counter()
    expect(c()).to_be(1)
    expect(c()).to_be(2)
    expect(c()).to_be(3)

test("__call__ stateful", test_call_stateful)

# ============================================================================
# Comparison dunder methods
# ============================================================================

def test_lt_dispatch():
    """__lt__ is called for < operator."""
    class A:
        def __init__(self, val):
            self.val = val
        def __lt__(self, other):
            return self.val < other.val
    expect(A(1) < A(2)).to_be(True)
    expect(A(2) < A(1)).to_be(False)
    expect(A(1) < A(1)).to_be(False)

test("__lt__ dispatch", test_lt_dispatch)

def test_le_dispatch():
    """__le__ is called for <= operator."""
    class A:
        def __init__(self, val):
            self.val = val
        def __le__(self, other):
            return self.val <= other.val
    expect(A(1) <= A(2)).to_be(True)
    expect(A(2) <= A(2)).to_be(True)
    expect(A(3) <= A(2)).to_be(False)

test("__le__ dispatch", test_le_dispatch)

def test_gt_dispatch():
    """__gt__ is called for > operator."""
    class A:
        def __init__(self, val):
            self.val = val
        def __gt__(self, other):
            return self.val > other.val
    expect(A(2) > A(1)).to_be(True)
    expect(A(1) > A(1)).to_be(False)

test("__gt__ dispatch", test_gt_dispatch)

def test_ge_dispatch():
    """__ge__ is called for >= operator."""
    class A:
        def __init__(self, val):
            self.val = val
        def __ge__(self, other):
            return self.val >= other.val
    expect(A(2) >= A(1)).to_be(True)
    expect(A(2) >= A(2)).to_be(True)
    expect(A(1) >= A(2)).to_be(False)

test("__ge__ dispatch", test_ge_dispatch)

def test_eq_dispatch():
    """__eq__ is called for == operator."""
    class A:
        def __init__(self, val):
            self.val = val
        def __eq__(self, other):
            if isinstance(other, A):
                return self.val == other.val
            return NotImplemented
    expect(A(1) == A(1)).to_be(True)
    expect(A(1) == A(2)).to_be(False)

test("__eq__ dispatch", test_eq_dispatch)

def test_ne_dispatch():
    """__ne__ is called for != operator."""
    class A:
        def __init__(self, val):
            self.val = val
        def __ne__(self, other):
            if isinstance(other, A):
                return self.val != other.val
            return NotImplemented
    expect(A(1) != A(2)).to_be(True)
    expect(A(1) != A(1)).to_be(False)

test("__ne__ dispatch", test_ne_dispatch)

def test_comparison_chaining():
    """Chained comparisons with custom classes."""
    class A:
        def __init__(self, val):
            self.val = val
        def __lt__(self, other):
            if isinstance(other, A):
                return self.val < other.val
            return self.val < other
        def __gt__(self, other):
            if isinstance(other, A):
                return self.val > other.val
            return self.val > other
    a = A(5)
    # 3 < a < 8  =>  3 < a and a < 8
    expect(A(3) < a).to_be(True)
    expect(a < A(8)).to_be(True)

test("comparison chaining", test_comparison_chaining)

def test_comparison_returns_notimplemented():
    """When __eq__ returns NotImplemented, fallback works."""
    class A:
        def __init__(self, val):
            self.val = val
        def __eq__(self, other):
            if isinstance(other, A):
                return self.val == other.val
            return NotImplemented
    a = A(1)
    # Comparing with non-A should use identity
    expect(a == 1).to_be(False)
    expect(a != 1).to_be(True)

test("comparison NotImplemented fallback", test_comparison_returns_notimplemented)

def test_sorted_with_lt():
    """sorted() uses __lt__ for ordering."""
    class Sortable:
        def __init__(self, val):
            self.val = val
        def __lt__(self, other):
            return self.val < other.val
        def __repr__(self):
            return "S(" + str(self.val) + ")"
    items = [Sortable(3), Sortable(1), Sortable(2)]
    result = sorted(items)
    expect(result[0].val).to_be(1)
    expect(result[1].val).to_be(2)
    expect(result[2].val).to_be(3)

test("sorted() uses __lt__", test_sorted_with_lt)

# ============================================================================
# Unary operators
# ============================================================================

def test_neg_dispatch():
    """__neg__ is called for unary - operator."""
    class A:
        def __init__(self, val):
            self.val = val
        def __neg__(self):
            return A(-self.val)
    result = -A(5)
    expect(result.val).to_be(-5)
    result = -A(-3)
    expect(result.val).to_be(3)

test("__neg__ dispatch", test_neg_dispatch)

def test_pos_dispatch():
    """__pos__ is called for unary + operator."""
    class A:
        def __init__(self, val):
            self.val = val
        def __pos__(self):
            return A(abs(self.val))
    result = +A(-5)
    expect(result.val).to_be(5)
    result = +A(3)
    expect(result.val).to_be(3)

test("__pos__ dispatch", test_pos_dispatch)

def test_abs_dispatch():
    """abs() calls __abs__."""
    class A:
        def __init__(self, val):
            self.val = val
        def __abs__(self):
            if self.val < 0:
                return A(-self.val)
            return A(self.val)
    expect(abs(A(-5)).val).to_be(5)
    expect(abs(A(3)).val).to_be(3)

test("__abs__ dispatch", test_abs_dispatch)

def test_invert_dispatch():
    """__invert__ is called for ~ operator."""
    class A:
        def __init__(self, val):
            self.val = val
        def __invert__(self):
            return A(~self.val)
    result = ~A(0)
    expect(result.val).to_be(-1)
    result = ~A(5)
    expect(result.val).to_be(-6)

test("__invert__ dispatch", test_invert_dispatch)

# ============================================================================
# Combined dunder method class (adapted from CPython's AllTests)
# ============================================================================

def test_all_dunders_on_one_class():
    """Multiple dunders on one class work together."""
    class Vector:
        def __init__(self, x, y):
            self.x = x
            self.y = y

        def __add__(self, other):
            return Vector(self.x + other.x, self.y + other.y)

        def __sub__(self, other):
            return Vector(self.x - other.x, self.y - other.y)

        def __mul__(self, scalar):
            return Vector(self.x * scalar, self.y * scalar)

        def __neg__(self):
            return Vector(-self.x, -self.y)

        def __eq__(self, other):
            if isinstance(other, Vector):
                return self.x == other.x and self.y == other.y
            return False

        def __repr__(self):
            return "Vector(" + str(self.x) + ", " + str(self.y) + ")"

        def __str__(self):
            return "<" + str(self.x) + ", " + str(self.y) + ">"

        def __bool__(self):
            return self.x != 0 or self.y != 0

        def __len__(self):
            # Number of components
            return 2

        def __getitem__(self, idx):
            if idx == 0:
                return self.x
            if idx == 1:
                return self.y
            raise IndexError("index out of range")

        def __hash__(self):
            return self.x * 31 + self.y

    v1 = Vector(1, 2)
    v2 = Vector(3, 4)

    # Arithmetic
    v3 = v1 + v2
    expect(v3.x).to_be(4)
    expect(v3.y).to_be(6)

    v4 = v2 - v1
    expect(v4.x).to_be(2)
    expect(v4.y).to_be(2)

    v5 = v1 * 3
    expect(v5.x).to_be(3)
    expect(v5.y).to_be(6)

    # Negation
    v6 = -v1
    expect(v6.x).to_be(-1)
    expect(v6.y).to_be(-2)

    # Equality
    expect(v1 == Vector(1, 2)).to_be(True)
    expect(v1 == v2).to_be(False)

    # String representations
    expect(repr(v1)).to_be("Vector(1, 2)")
    expect(str(v1)).to_be("<1, 2>")

    # Truthiness
    expect(bool(v1)).to_be(True)
    expect(bool(Vector(0, 0))).to_be(False)

    # Indexing
    expect(v1[0]).to_be(1)
    expect(v1[1]).to_be(2)

    # Length
    expect(len(v1)).to_be(2)

    # Hash for dict keys
    d = {}
    d[v1] = "v1"
    expect(d[Vector(1, 2)]).to_be("v1")

test("all dunders on one class", test_all_dunders_on_one_class)

# ============================================================================
# Dunder method inheritance
# ============================================================================

def test_dunder_inheritance():
    """Dunder methods are inherited from parent class."""
    class Base:
        def __init__(self, val):
            self.val = val
        def __add__(self, other):
            return type(self)(self.val + other.val)
        def __repr__(self):
            return type(self).__name__ + "(" + str(self.val) + ")"

    class Child(Base):
        pass

    c1 = Child(1)
    c2 = Child(2)
    c3 = c1 + c2
    expect(c3.val).to_be(3)
    expect(repr(c3)).to_be("Child(3)")

test("dunder inheritance", test_dunder_inheritance)

def test_dunder_override():
    """Child can override parent dunder methods."""
    class Base:
        def __init__(self, val):
            self.val = val
        def __str__(self):
            return "Base:" + str(self.val)

    class Child(Base):
        def __str__(self):
            return "Child:" + str(self.val)

    expect(str(Base(1))).to_be("Base:1")
    expect(str(Child(2))).to_be("Child:2")

test("dunder override", test_dunder_override)

# ============================================================================
# Edge cases
# ============================================================================

def test_eq_identity_fallback():
    """Without __eq__, == falls back to identity."""
    class A:
        pass
    a = A()
    b = A()
    expect(a == a).to_be(True)
    expect(a == b).to_be(False)

test("__eq__ identity fallback", test_eq_identity_fallback)

def test_object_default_truthiness():
    """Objects without __bool__ or __len__ are always truthy."""
    class A:
        pass
    expect(bool(A())).to_be(True)

test("default object truthiness", test_object_default_truthiness)

def test_getitem_key_error():
    """__getitem__ raising KeyError propagates."""
    class A:
        def __getitem__(self, key):
            raise KeyError(key)
    a = A()
    try:
        a["missing"]
        expect("no error").to_be("KeyError")
    except KeyError:
        expect(True).to_be(True)

test("__getitem__ KeyError propagation", test_getitem_key_error)

def test_setitem_type_check():
    """__setitem__ can do type checking."""
    class TypedDict:
        def __init__(self):
            self.data = {}
        def __setitem__(self, key, val):
            if not isinstance(key, str):
                raise TypeError("key must be str")
            self.data[key] = val
        def __getitem__(self, key):
            return self.data[key]
    td = TypedDict()
    td["valid"] = 42
    expect(td["valid"]).to_be(42)
    try:
        td[123] = "bad"
        expect("no error").to_be("TypeError")
    except TypeError:
        expect(True).to_be(True)

test("__setitem__ type check", test_setitem_type_check)

print("CPython class extra tests completed")
