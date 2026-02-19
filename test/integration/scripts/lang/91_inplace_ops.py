from test_framework import test, expect

# Test 1: __iadd__ is called when defined
def test_iadd_called(t):
    class Acc:
        def __init__(self, val):
            self.val = val
        def __iadd__(self, other):
            self.val += other
            return self

    a = Acc(10)
    original = a
    a += 5
    expect(a.val).to_be(15)
    expect(a is original).to_be(True)  # same object, mutated in-place

test("__iadd__ is called and mutates in place", test_iadd_called)

# Test 2: Falls back to __add__ when __iadd__ not defined
def test_fallback_to_add(t):
    class Num:
        def __init__(self, val):
            self.val = val
        def __add__(self, other):
            return Num(self.val + other)

    a = Num(10)
    original = a
    a += 5
    expect(a.val).to_be(15)
    expect(a is original).to_be(False)  # new object from __add__

test("falls back to __add__ when __iadd__ not defined", test_fallback_to_add)

# Test 3: __isub__
def test_isub(t):
    class Acc:
        def __init__(self, val):
            self.val = val
        def __isub__(self, other):
            self.val -= other
            return self

    a = Acc(10)
    a -= 3
    expect(a.val).to_be(7)

test("__isub__", test_isub)

# Test 4: __imul__
def test_imul(t):
    class Acc:
        def __init__(self, val):
            self.val = val
        def __imul__(self, other):
            self.val *= other
            return self

    a = Acc(5)
    a *= 3
    expect(a.val).to_be(15)

test("__imul__", test_imul)

# Test 5: __itruediv__
def test_itruediv(t):
    class Acc:
        def __init__(self, val):
            self.val = val
        def __itruediv__(self, other):
            self.val = self.val / other
            return self

    a = Acc(10)
    a /= 2
    expect(a.val).to_be(5.0)

test("__itruediv__", test_itruediv)

# Test 6: __ifloordiv__
def test_ifloordiv(t):
    class Acc:
        def __init__(self, val):
            self.val = val
        def __ifloordiv__(self, other):
            self.val = self.val // other
            return self

    a = Acc(10)
    a //= 3
    expect(a.val).to_be(3)

test("__ifloordiv__", test_ifloordiv)

# Test 7: __imod__
def test_imod(t):
    class Acc:
        def __init__(self, val):
            self.val = val
        def __imod__(self, other):
            self.val = self.val % other
            return self

    a = Acc(10)
    a %= 3
    expect(a.val).to_be(1)

test("__imod__", test_imod)

# Test 8: __ipow__
def test_ipow(t):
    class Acc:
        def __init__(self, val):
            self.val = val
        def __ipow__(self, other):
            self.val = self.val ** other
            return self

    a = Acc(2)
    a **= 10
    expect(a.val).to_be(1024)

test("__ipow__", test_ipow)

# Test 9: __imatmul__
def test_imatmul(t):
    class Acc:
        def __init__(self, val):
            self.val = val
        def __imatmul__(self, other):
            self.val = self.val + other  # just test it's called
            return self

    a = Acc(1)
    a @= 2
    expect(a.val).to_be(3)

test("__imatmul__", test_imatmul)

# Test 10: __iand__
def test_iand(t):
    class Acc:
        def __init__(self, val):
            self.val = val
        def __iand__(self, other):
            self.val = self.val & other
            return self

    a = Acc(0b1111)
    a &= 0b1010
    expect(a.val).to_be(0b1010)

test("__iand__", test_iand)

# Test 11: __ior__
def test_ior(t):
    class Acc:
        def __init__(self, val):
            self.val = val
        def __ior__(self, other):
            self.val = self.val | other
            return self

    a = Acc(0b1010)
    a |= 0b0101
    expect(a.val).to_be(0b1111)

test("__ior__", test_ior)

# Test 12: __ixor__
def test_ixor(t):
    class Acc:
        def __init__(self, val):
            self.val = val
        def __ixor__(self, other):
            self.val = self.val ^ other
            return self

    a = Acc(0b1111)
    a ^= 0b1010
    expect(a.val).to_be(0b0101)

test("__ixor__", test_ixor)

# Test 13: __ilshift__
def test_ilshift(t):
    class Acc:
        def __init__(self, val):
            self.val = val
        def __ilshift__(self, other):
            self.val = self.val << other
            return self

    a = Acc(1)
    a <<= 4
    expect(a.val).to_be(16)

test("__ilshift__", test_ilshift)

# Test 14: __irshift__
def test_irshift(t):
    class Acc:
        def __init__(self, val):
            self.val = val
        def __irshift__(self, other):
            self.val = self.val >> other
            return self

    a = Acc(16)
    a >>= 4
    expect(a.val).to_be(1)

test("__irshift__", test_irshift)

# Test 15: List-like mutation pattern with __iadd__
def test_list_like_iadd(t):
    class MyList:
        def __init__(self):
            self.items = []
        def __iadd__(self, other):
            for item in other:
                self.items.append(item)
            return self

    lst = MyList()
    lst += [1, 2, 3]
    lst += [4, 5]
    expect(lst.items).to_be([1, 2, 3, 4, 5])

test("list-like __iadd__ mutation pattern", test_list_like_iadd)

# Test 16: Inherited inplace dunder
def test_inherited_iadd(t):
    class Base:
        def __init__(self, val):
            self.val = val
        def __iadd__(self, other):
            self.val += other
            return self

    class Child(Base):
        pass

    c = Child(10)
    c += 5
    expect(c.val).to_be(15)

test("inherited __iadd__ from base class", test_inherited_iadd)

# Test 17: __iadd__ returning None makes variable None (no fallback)
def test_iadd_returns_none(t):
    class Num:
        def __init__(self, val):
            self.val = val
        def __iadd__(self, other):
            return None

    a = Num(10)
    a += 5
    expect(a).to_be(None)

test("__iadd__ returning None assigns None", test_iadd_returns_none)

# ============================================================================
# Ported from CPython test_augassign.py
# ============================================================================

# CPython testBasic: chain of augmented assignments on a plain variable
def test_cpython_basic_chain():
    x = 2
    x += 1
    x *= 2
    x **= 2
    x -= 8
    x //= 5
    x %= 3
    x &= 2
    x |= 5
    x ^= 1
    x /= 2
    expect(x).to_be(3.0)

test("cpython: basic augmented assignment chain", test_cpython_basic_chain)

# CPython testInList: chain of augmented assignments on a list element
def test_cpython_in_list():
    x = [2]
    x[0] += 1
    x[0] *= 2
    x[0] **= 2
    x[0] -= 8
    x[0] //= 5
    x[0] %= 3
    x[0] &= 2
    x[0] |= 5
    x[0] ^= 1
    x[0] /= 2
    expect(x[0]).to_be(3.0)

test("cpython: augmented assignment on list element", test_cpython_in_list)

# CPython testInDict: chain of augmented assignments on a dict value
def test_cpython_in_dict():
    x = {0: 2}
    x[0] += 1
    x[0] *= 2
    x[0] **= 2
    x[0] -= 8
    x[0] //= 5
    x[0] %= 3
    x[0] &= 2
    x[0] |= 5
    x[0] ^= 1
    x[0] /= 2
    expect(x[0]).to_be(3.0)

test("cpython: augmented assignment on dict value", test_cpython_in_dict)

# CPython testSequences: list += and *=
def test_cpython_sequences_concat_repeat():
    x = [1, 2]
    x += [3, 4]
    x *= 2
    expect(x).to_be([1, 2, 3, 4, 1, 2, 3, 4])

test("cpython: list += concatenation and *= repetition", test_cpython_sequences_concat_repeat)

# CPython testSequences: slice augmented assignment with identity preservation
def test_cpython_sequences_slice_augassign():
    x = [1, 2, 3]
    y = x
    x[1:2] *= 2
    expect(x).to_be([1, 2, 2, 3])
    expect(x is y).to_be(True)

test("cpython: slice *= augmented assignment preserves identity", test_cpython_sequences_slice_augassign)

# CPython testSequences: slice += augmented assignment
def test_cpython_sequences_slice_iadd():
    x = [1, 2, 2, 3]
    y = x
    y[1:2] += [1]
    expect(x).to_be([1, 2, 1, 2, 3])
    expect(x is y).to_be(True)

test("cpython: slice += augmented assignment preserves identity", test_cpython_sequences_slice_iadd)

# CPython testCustomMethods1: __add__ fallback creates new object, __iadd__ preserves identity
def test_cpython_custom_methods1_add_fallback():
    class aug_test:
        def __init__(self, value):
            self.val = value
        def __radd__(self, val):
            return self.val + val
        def __add__(self, val):
            return aug_test(self.val + val)

    x = aug_test(1)
    y = x
    x += 10

    # Fell back to __add__, so x is a new aug_test instance
    expect(x.val).to_be(11)
    expect(y is x).to_be(False)

test("cpython: += falls back to __add__ creating new object", test_cpython_custom_methods1_add_fallback)

# CPython testCustomMethods1: __iadd__ returning self preserves identity
def test_cpython_custom_methods1_iadd_identity():
    class aug_test:
        def __init__(self, value):
            self.val = value
        def __radd__(self, val):
            return self.val + val
        def __add__(self, val):
            return aug_test(self.val + val)

    class aug_test2(aug_test):
        def __iadd__(self, val):
            self.val = self.val + val
            return self

    x = aug_test2(2)
    y = x
    x += 10

    expect(y is x).to_be(True)
    expect(x.val).to_be(12)

test("cpython: __iadd__ returning self preserves identity", test_cpython_custom_methods1_iadd_identity)

# CPython testCustomMethods1: __iadd__ returning new instance
def test_cpython_custom_methods1_iadd_new_instance():
    class aug_test:
        def __init__(self, value):
            self.val = value
        def __radd__(self, val):
            return self.val + val
        def __add__(self, val):
            return aug_test(self.val + val)

    class aug_test3(aug_test):
        def __iadd__(self, val):
            return aug_test3(self.val + val)

    x = aug_test3(3)
    y = x
    x += 10

    expect(x.val).to_be(13)
    expect(y is x).to_be(False)

test("cpython: __iadd__ returning new instance replaces variable", test_cpython_custom_methods1_iadd_new_instance)

# CPython testCustomMethods1: __iadd__ = None blocks inheritance and fallback
def test_cpython_custom_methods1_iadd_none_typeerror():
    class aug_test:
        def __init__(self, value):
            self.val = value
        def __radd__(self, val):
            return self.val + val
        def __add__(self, val):
            return aug_test(self.val + val)

    class aug_test3(aug_test):
        def __iadd__(self, val):
            return aug_test3(self.val + val)

    class aug_test4(aug_test3):
        """Blocks inheritance, and fallback to __add__"""
        __iadd__ = None

    x = aug_test4(4)
    got_type_error = False
    try:
        x += 10
    except TypeError:
        got_type_error = True
    expect(got_type_error).to_be(True)

test("cpython: __iadd__ = None raises TypeError", test_cpython_custom_methods1_iadd_none_typeerror)

# CPython testCustomMethods2: all inplace operators dispatch to correct dunders
def test_cpython_custom_methods2_iadd_dispatch():
    output = []

    class testall:
        def __add__(self, val):
            output.append("__add__ called")
        def __radd__(self, val):
            output.append("__radd__ called")
        def __iadd__(self, val):
            output.append("__iadd__ called")
            return self

        def __sub__(self, val):
            output.append("__sub__ called")
        def __rsub__(self, val):
            output.append("__rsub__ called")
        def __isub__(self, val):
            output.append("__isub__ called")
            return self

        def __mul__(self, val):
            output.append("__mul__ called")
        def __rmul__(self, val):
            output.append("__rmul__ called")
        def __imul__(self, val):
            output.append("__imul__ called")
            return self

        def __floordiv__(self, val):
            output.append("__floordiv__ called")
            return self
        def __ifloordiv__(self, val):
            output.append("__ifloordiv__ called")
            return self
        def __rfloordiv__(self, val):
            output.append("__rfloordiv__ called")
            return self

        def __truediv__(self, val):
            output.append("__truediv__ called")
            return self
        def __rtruediv__(self, val):
            output.append("__rtruediv__ called")
            return self
        def __itruediv__(self, val):
            output.append("__itruediv__ called")
            return self

        def __mod__(self, val):
            output.append("__mod__ called")
        def __rmod__(self, val):
            output.append("__rmod__ called")
        def __imod__(self, val):
            output.append("__imod__ called")
            return self

        def __pow__(self, val):
            output.append("__pow__ called")
        def __rpow__(self, val):
            output.append("__rpow__ called")
        def __ipow__(self, val):
            output.append("__ipow__ called")
            return self

        def __or__(self, val):
            output.append("__or__ called")
        def __ror__(self, val):
            output.append("__ror__ called")
        def __ior__(self, val):
            output.append("__ior__ called")
            return self

        def __and__(self, val):
            output.append("__and__ called")
        def __rand__(self, val):
            output.append("__rand__ called")
        def __iand__(self, val):
            output.append("__iand__ called")
            return self

        def __xor__(self, val):
            output.append("__xor__ called")
        def __rxor__(self, val):
            output.append("__rxor__ called")
        def __ixor__(self, val):
            output.append("__ixor__ called")
            return self

        def __rshift__(self, val):
            output.append("__rshift__ called")
        def __rrshift__(self, val):
            output.append("__rrshift__ called")
        def __irshift__(self, val):
            output.append("__irshift__ called")
            return self

        def __lshift__(self, val):
            output.append("__lshift__ called")
        def __rlshift__(self, val):
            output.append("__rlshift__ called")
        def __ilshift__(self, val):
            output.append("__ilshift__ called")
            return self

    x = testall()

    # __add__, __radd__, __iadd__
    x + 1
    1 + x
    x += 1

    # __sub__, __rsub__, __isub__
    x - 1
    1 - x
    x -= 1

    # __mul__, __rmul__, __imul__
    x * 1
    1 * x
    x *= 1

    # __truediv__, __rtruediv__, __itruediv__
    x / 1
    1 / x
    x /= 1

    # __floordiv__, __rfloordiv__, __ifloordiv__
    x // 1
    1 // x
    x //= 1

    # __mod__, __rmod__, __imod__
    x % 1
    1 % x
    x %= 1

    # __pow__, __rpow__, __ipow__
    x ** 1
    1 ** x
    x **= 1

    # __or__, __ror__, __ior__
    x | 1
    1 | x
    x |= 1

    # __and__, __rand__, __iand__
    x & 1
    1 & x
    x &= 1

    # __xor__, __rxor__, __ixor__
    x ^ 1
    1 ^ x
    x ^= 1

    # __rshift__, __rrshift__, __irshift__
    x >> 1
    1 >> x
    x >>= 1

    # __lshift__, __rlshift__, __ilshift__
    x << 1
    1 << x
    x <<= 1

    expected = [
        "__add__ called",
        "__radd__ called",
        "__iadd__ called",
        "__sub__ called",
        "__rsub__ called",
        "__isub__ called",
        "__mul__ called",
        "__rmul__ called",
        "__imul__ called",
        "__truediv__ called",
        "__rtruediv__ called",
        "__itruediv__ called",
        "__floordiv__ called",
        "__rfloordiv__ called",
        "__ifloordiv__ called",
        "__mod__ called",
        "__rmod__ called",
        "__imod__ called",
        "__pow__ called",
        "__rpow__ called",
        "__ipow__ called",
        "__or__ called",
        "__ror__ called",
        "__ior__ called",
        "__and__ called",
        "__rand__ called",
        "__iand__ called",
        "__xor__ called",
        "__rxor__ called",
        "__ixor__ called",
        "__rshift__ called",
        "__rrshift__ called",
        "__irshift__ called",
        "__lshift__ called",
        "__rlshift__ called",
        "__ilshift__ called",
    ]
    expect(output).to_be(expected)

test("cpython: all augmented operators dispatch to correct dunders", test_cpython_custom_methods2_iadd_dispatch)

# CPython testCustomMethods2 (matmul variant): @= dispatches to __imatmul__
def test_cpython_custom_methods2_matmul_dispatch():
    output = []

    class testmatmul:
        def __matmul__(self, val):
            output.append("__matmul__ called")
        def __rmatmul__(self, val):
            output.append("__rmatmul__ called")
        def __imatmul__(self, val):
            output.append("__imatmul__ called")
            return self

    x = testmatmul()
    x @ 1
    1 @ x
    x @= 1

    expected = [
        "__matmul__ called",
        "__rmatmul__ called",
        "__imatmul__ called",
    ]
    expect(output).to_be(expected)

test("cpython: @= dispatches to __imatmul__", test_cpython_custom_methods2_matmul_dispatch)

# Additional: augmented assignment on dict with string keys
def test_cpython_dict_string_keys():
    d = {"a": 10, "b": 20}
    d["a"] += 5
    d["b"] *= 3
    expect(d["a"]).to_be(15)
    expect(d["b"]).to_be(60)

test("cpython: augmented assignment on dict with string keys", test_cpython_dict_string_keys)

# Additional: list identity preserved with +=
def test_cpython_list_identity_iadd():
    x = [1, 2, 3]
    y = x
    x += [4, 5]
    # In CPython, list += extends in place, so x is y
    expect(x is y).to_be(True)
    expect(x).to_be([1, 2, 3, 4, 5])

test("cpython: list += preserves identity (extends in place)", test_cpython_list_identity_iadd)

# Additional: list identity preserved with *=
def test_cpython_list_identity_imul():
    x = [1, 2]
    y = x
    x *= 3
    # In CPython, list *= repeats in place, so x is y
    expect(x is y).to_be(True)
    expect(x).to_be([1, 2, 1, 2, 1, 2])

test("cpython: list *= preserves identity (repeats in place)", test_cpython_list_identity_imul)

# Additional: __iadd__ returning NotImplemented falls back to __add__
def test_cpython_iadd_notimplemented_fallback():
    class MyNum:
        def __init__(self, val):
            self.val = val
        def __iadd__(self, other):
            return NotImplemented
        def __add__(self, other):
            return MyNum(self.val + other)

    x = MyNum(5)
    y = x
    x += 10
    expect(x.val).to_be(15)
    expect(x is y).to_be(False)  # new object from __add__ fallback

test("cpython: __iadd__ returning NotImplemented falls back to __add__", test_cpython_iadd_notimplemented_fallback)

# Additional: augmented assignment with nested subscript
def test_cpython_nested_subscript():
    x = [[1, 2], [3, 4]]
    x[0][1] += 10
    x[1][0] *= 2
    expect(x[0][1]).to_be(12)
    expect(x[1][0]).to_be(6)

test("cpython: augmented assignment on nested subscript", test_cpython_nested_subscript)

# Additional: augmented assignment on attribute
def test_cpython_augassign_attribute():
    class Obj:
        def __init__(self):
            self.x = 10
            self.y = 3

    o = Obj()
    o.x += 5
    o.y *= 4
    expect(o.x).to_be(15)
    expect(o.y).to_be(12)

test("cpython: augmented assignment on attributes", test_cpython_augassign_attribute)

# Additional: all augmented ops on integers (comprehensive check)
def test_cpython_all_int_augops():
    # +=
    x = 10
    x += 5
    expect(x).to_be(15)

    # -=
    x = 10
    x -= 3
    expect(x).to_be(7)

    # *=
    x = 4
    x *= 5
    expect(x).to_be(20)

    # /= (always returns float)
    x = 10
    x /= 4
    expect(x).to_be(2.5)

    # //=
    x = 10
    x //= 3
    expect(x).to_be(3)

    # %=
    x = 10
    x %= 3
    expect(x).to_be(1)

    # **=
    x = 2
    x **= 8
    expect(x).to_be(256)

    # &=
    x = 0xFF
    x &= 0x0F
    expect(x).to_be(0x0F)

    # |=
    x = 0xF0
    x |= 0x0F
    expect(x).to_be(0xFF)

    # ^=
    x = 0xFF
    x ^= 0x0F
    expect(x).to_be(0xF0)

    # <<=
    x = 1
    x <<= 8
    expect(x).to_be(256)

    # >>=
    x = 256
    x >>= 4
    expect(x).to_be(16)

test("cpython: all augmented ops on plain integers", test_cpython_all_int_augops)

# Additional: all augmented ops on floats
def test_cpython_all_float_augops():
    x = 2.5
    x += 1.5
    expect(x).to_be(4.0)

    x = 10.0
    x -= 3.5
    expect(x).to_be(6.5)

    x = 3.0
    x *= 2.5
    expect(x).to_be(7.5)

    x = 7.5
    x /= 2.5
    expect(x).to_be(3.0)

    x = 7.5
    x //= 2.0
    expect(x).to_be(3.0)

    x = 7.5
    x %= 2.0
    expect(x).to_be(1.5)

    x = 2.0
    x **= 3.0
    expect(x).to_be(8.0)

test("cpython: all augmented ops on floats", test_cpython_all_float_augops)

# Additional: string += concatenation
def test_cpython_string_iadd():
    s = "hello"
    s += " "
    s += "world"
    expect(s).to_be("hello world")

test("cpython: string += concatenation", test_cpython_string_iadd)

# Additional: string *= repetition
def test_cpython_string_imul():
    s = "ab"
    s *= 3
    expect(s).to_be("ababab")

test("cpython: string *= repetition", test_cpython_string_imul)

# Additional: tuple += concatenation (creates new tuple, no identity preservation)
def test_cpython_tuple_iadd():
    t = (1, 2)
    orig = t
    t += (3, 4)
    expect(t).to_be((1, 2, 3, 4))
    expect(t is orig).to_be(False)  # tuples are immutable, new object

test("cpython: tuple += concatenation creates new tuple", test_cpython_tuple_iadd)

# Additional: tuple *= repetition (creates new tuple)
def test_cpython_tuple_imul():
    t = (1, 2)
    orig = t
    t *= 3
    expect(t).to_be((1, 2, 1, 2, 1, 2))
    expect(t is orig).to_be(False)

test("cpython: tuple *= repetition creates new tuple", test_cpython_tuple_imul)
