# Test: Built-in collection, string and numeric semantics matching CPython 3.14
# Regression tests for extended slices, set/dict ordering, dict views, special
# methods on built-ins, str methods, and int/float arithmetic edge cases.

from test_framework import test, expect


def raises(fn, exc_type):
    """Return the exception message if fn raises exc_type, else None."""
    try:
        fn()
    except exc_type as e:
        return str(e)
    return None


# ---------------------------------------------------------------------------
# Lists: extended slices
# ---------------------------------------------------------------------------

def test_extended_slice_assign():
    c = [1, 2, 3]
    c[::-1] = [7, 8, 9]
    expect(c).to_be([9, 8, 7])
    d = [0] * 5
    d[1:4:2] = ['x', 'y']
    expect(d).to_be([0, 'x', 0, 'y', 0])
    a = list(range(10))
    a[8:2:-2] = 'abc'
    expect(a).to_be([0, 1, 2, 3, 'c', 5, 'b', 7, 'a', 9])
    b = list(range(5))
    b[::-1] = b
    expect(b).to_be([4, 3, 2, 1, 0])


def test_extended_slice_size_mismatch():
    d = [0] * 5

    def assign():
        d[1:4:2] = ['x']
    msg = raises(assign, ValueError)
    expect(msg).to_be("attempt to assign sequence of size 1 to extended slice of size 2")
    expect(d).to_be([0, 0, 0, 0, 0])


def test_extended_slice_delete():
    a = list(range(10))
    del a[::2]
    expect(a).to_be([1, 3, 5, 7, 9])
    b = list(range(10))
    del b[::-3]
    expect(b).to_be([1, 2, 4, 5, 7, 8])
    c = list(range(5))
    del c[10:-10:-1]
    expect(c).to_be([])


def test_simple_slice_assign():
    a = list(range(5))
    a[3:1] = ['q']
    expect(a).to_be([0, 1, 2, 'q', 3, 4])
    a[1:3] = []
    expect(a).to_be([0, 'q', 3, 4])


def test_list_sort_key_called_once():
    calls = []

    def key(p):
        calls.append(p)
        return p[0]
    data = [(1, 'a'), (0, 'b'), (1, 'c'), (0, 'd'), (2, 'e')]
    data.sort(key=key, reverse=True)
    expect(data).to_be([(2, 'e'), (1, 'a'), (1, 'c'), (0, 'b'), (0, 'd')])
    expect(len(calls)).to_be(5)


def test_list_index_message():
    msg = raises(lambda: [1, 2, 3].index(5), ValueError)
    expect(msg).to_be("list.index(x): x not in list")
    expect([1, 2, 3, 2].index(2, 2)).to_be(3)


# ---------------------------------------------------------------------------
# Sets: deterministic, CPython-compatible iteration order
# ---------------------------------------------------------------------------

def test_set_iteration_order():
    expect(list({3, 1, 2})).to_be([1, 2, 3])
    expect(list({10, 3, 7, 100, 5})).to_be([3, 100, 5, 7, 10])
    expect(str({3, 1, 2})).to_be("{1, 2, 3}")
    # Colliding hashes are placed by CPython's probe sequence
    expect(list(set([9, 17, 1]))).to_be([9, 17, 1])
    expect(list(frozenset([200, 8, 0]))).to_be([200, 8, 0])


def test_set_operation_order():
    expect(list({1, 2, 3} | {10, 4})).to_be([1, 2, 3, 4, 10])
    expect(list({5, 4, 3, 2, 1} & {4, 2, 9})).to_be([2, 4])
    expect(list({1, 2, 3, 4} - {2})).to_be([1, 3, 4])
    expect(list({1, 2} ^ {2, 3})).to_be([1, 3])


def test_set_pop_order():
    s = {30, 10, 20}
    expect(s.pop()).to_be(10)
    expect(s.pop()).to_be(20)
    expect(s.pop()).to_be(30)


def test_set_inplace_ops_mutate():
    s = {1, 2}
    alias = s
    s |= {3}
    s -= {1}
    s &= {2, 3, 4}
    s ^= {2, 9}
    expect(alias is s).to_be(True)
    expect(sorted(alias)).to_be([3, 9])


def test_set_numeric_equality():
    expect(False in {0.0}).to_be(True)
    expect(0.0 == False).to_be(True)
    expect(len({1, 1.0, True})).to_be(1)
    expect(frozenset({1, 2}) in {frozenset({1, 2})}).to_be(True)


def test_hash_matches_cpython():
    expect(hash(5)).to_be(5)
    expect(hash(-1)).to_be(-2)
    expect(hash(1.5)).to_be(1152921504606846977)
    expect(hash((1, 2))).to_be(-3550055125485641917)
    expect(hash(frozenset({1, 2}))).to_be(-1826646154956904602)
    expect(hash(1.0) == hash(1)).to_be(True)


# ---------------------------------------------------------------------------
# Dicts and dict views
# ---------------------------------------------------------------------------

def test_dict_copy_keeps_order():
    d = {'x': 1, 'y': 2}
    expect(list(dict(d, z=0))).to_be(['x', 'y', 'z'])
    big = {}
    for i in range(50):
        big[str(49 - i)] = i
    expect(list(dict(big))[:3]).to_be(['49', '48', '47'])


def test_dict_views_repr_and_ops():
    d = {'a': 1, 'b': 2}
    expect(str(d.keys())).to_be("dict_keys(['a', 'b'])")
    expect(str(d.values())).to_be("dict_values([1, 2])")
    expect(str(d.items())).to_be("dict_items([('a', 1), ('b', 2)])")
    expect(d.keys() & {'a'}).to_be({'a'})
    expect(sorted(d.keys() | {'c'})).to_be(['a', 'b', 'c'])
    expect(d.keys() - {'a'}).to_be({'b'})
    expect(sorted(d.keys() ^ {'a', 'z'})).to_be(['b', 'z'])
    expect(('a', 1) in d.items()).to_be(True)
    expect(2 in d.values()).to_be(True)
    expect(d.keys() == {'a', 'b'}).to_be(True)
    expect(len(d.items())).to_be(2)


def test_dict_views_are_live():
    d = {'a': 1}
    keys = d.keys()
    d['b'] = 2
    expect(list(keys)).to_be(['a', 'b'])
    expect(len(keys)).to_be(2)


def test_dict_inplace_or_mutates():
    d = {'a': 1}
    alias = d
    d |= {'b': 2}
    expect(alias is d).to_be(True)
    expect(alias).to_be({'a': 1, 'b': 2})


def test_dict_errors():
    expect(raises(lambda: dict([('a', 1, 2)]), ValueError)).to_be(
        "dictionary update sequence element #0 has length 3; 2 is required")
    expect(raises(lambda: {}.pop('x'), KeyError)).to_be("'x'")


# ---------------------------------------------------------------------------
# Special methods on built-in values
# ---------------------------------------------------------------------------

def test_builtin_dunder_methods():
    d = {}
    d.__setitem__('k', 1)
    expect(d).to_be({'k': 1})
    expect(d.__getitem__('k')).to_be(1)
    lst = [1]
    expect(lst.__iadd__([2])).to_be([1, 2])
    expect(lst).to_be([1, 2])
    lst.__delitem__(0)
    expect(lst).to_be([2])
    expect(list.__getitem__([5, 6], 1)).to_be(6)
    expect("ab".__contains__("a")).to_be(True)
    expect((1, 2).__len__()).to_be(2)
    expect({1, 2}.__contains__(2)).to_be(True)
    expect(hasattr({1}, '__getitem__')).to_be(False)
    expect(sorted(['b', 'a', 'c'], key={'a': 3, 'b': 1, 'c': 2}.__getitem__)).to_be(['b', 'c', 'a'])


# ---------------------------------------------------------------------------
# range, slice, zip
# ---------------------------------------------------------------------------

def test_range_slicing_and_repr():
    expect(range(0, 10, 2)[1:3]).to_be(range(2, 6, 2))
    expect(str(range(0, 10, 2)[1:3])).to_be("range(2, 6, 2)")
    expect(str(range(5))).to_be("range(0, 5)")
    expect(str(range(10)[::-1])).to_be("range(9, -1, -1)")
    expect(range(10)[-1]).to_be(9)


def test_slice_indices():
    expect(slice(None).indices(10)).to_be((0, 10, 1))
    expect(slice(-3, None, -1).indices(10)).to_be((7, -1, -1))
    expect(slice(1, 100, 2).indices(5)).to_be((1, 5, 2))


def test_zip_strict():
    msg = raises(lambda: list(zip([1, 2], [1], strict=True)), ValueError)
    expect(msg).to_be("zip() argument 2 is shorter than argument 1")
    msg = raises(lambda: list(zip([1], [1, 2], strict=True)), ValueError)
    expect(msg).to_be("zip() argument 2 is longer than argument 1")
    expect(list(zip([1, 2], 'ab', strict=True))).to_be([(1, 'a'), (2, 'b')])


# ---------------------------------------------------------------------------
# str methods
# ---------------------------------------------------------------------------

def test_str_split_maxsplit():
    expect("a b c".split(None, 1)).to_be(['a', 'b c'])
    expect("a b c".rsplit(None, 1)).to_be(['a b', 'c'])
    expect("  a  b  c  ".split(None, 1)).to_be(['a', 'b  c  '])
    expect("  a  b  c  ".rsplit(None, 1)).to_be(['  a  b', 'c'])
    expect("a b c".split(maxsplit=1)).to_be(['a', 'b c'])
    expect("a,b,c".rsplit(",", 1)).to_be(['a,b', 'c'])
    expect(raises(lambda: "x".split(""), ValueError)).to_be("empty separator")


def test_str_padding():
    expect("ab".center(7, "*")).to_be("***ab**")
    expect("abc".center(6, "*")).to_be("*abc**")
    expect("é".zfill(3)).to_be("00é")
    expect("é".center(3, "-")).to_be("-é-")
    expect("-42".zfill(5)).to_be("-0042")
    msg = raises(lambda: "a".ljust(5, "ab"), TypeError)
    expect(msg).to_be("The fill character must be exactly one character long")


def test_str_start_end_args():
    expect("hello".startswith("l", 2)).to_be(True)
    expect("hello".startswith(("x", "ll"), 2)).to_be(True)
    expect("hello".endswith("ll", 0, 4)).to_be(True)
    expect("hello".startswith("", 6)).to_be(False)
    expect("abcabc".rindex("b", 0, 4)).to_be(1)
    expect("abcabc".find("c", -2)).to_be(5)
    expect("abcabc".rfind("a", 1)).to_be(3)
    expect("abcabc".count("b", 2)).to_be(1)
    expect("abc".find("", 4)).to_be(-1)
    expect("héllo".find("l")).to_be(2)


def test_str_case_and_predicates():
    expect("hello2world".title()).to_be("Hello2World")
    expect("they're bill's".title()).to_be("They'Re Bill'S")
    expect("²".isdigit()).to_be(True)
    expect("½".isnumeric()).to_be(True)
    expect("½".isdigit()).to_be(False)
    expect("héllo".isalpha()).to_be(True)
    expect("Straße".upper()).to_be("STRASSE")
    expect("ΣΑΣ".lower()).to_be("σας")
    expect("AbC".isupper()).to_be(False)
    expect("abc1".islower()).to_be(True)


def test_str_strip_none():
    expect("  x  ".strip(None)).to_be("x")
    expect("  x  ".lstrip(None)).to_be("x  ")
    expect("  x  ".rstrip(None)).to_be("  x")


def test_str_join_error():
    msg = raises(lambda: ",".join([1, 2]), TypeError)
    expect(msg).to_be("sequence item 0: expected str instance, int found")


def test_str_misc():
    expect("a\tb".expandtabs(0)).to_be("ab")
    expect("a\x0bb\x1cc".splitlines()).to_be(['a', 'b', 'c'])
    expect("hello".replace("l", "L", count=1)).to_be("heLlo")
    expect("café".encode("ascii", "ignore")).to_be(b"caf")
    expect(raises(lambda: b"\xff".decode(), UnicodeDecodeError) is not None).to_be(True)


def test_bytes_fromhex():
    expect(bytes.fromhex("4142")).to_be(b"AB")
    expect(bytes.fromhex("41 42 7a")).to_be(b"ABz")
    msg = raises(lambda: bytes.fromhex("zz"), ValueError)
    expect(msg).to_be("non-hexadecimal number found in fromhex() arg at position 0")


# ---------------------------------------------------------------------------
# Numbers
# ---------------------------------------------------------------------------

def test_modular_pow():
    expect(pow(2, -1, 7)).to_be(4)
    expect(pow(3, 4, -5)).to_be(-4)
    expect(pow(38, -1, 97)).to_be(23)
    msg = raises(lambda: pow(2, -1, 4), ValueError)
    expect(msg).to_be("base is not invertible for the given modulus")


def test_float_modulo_and_floordiv():
    expect(str(-0.0 % 1.0)).to_be("0.0")
    expect(1 // 0.1).to_be(9.0)
    expect(divmod(-7.5, 2)).to_be((-4.0, 0.5))
    expect(divmod(7, -2)).to_be((-4, -1))


def test_int_parsing_underscores():
    expect(raises(lambda: int("1__0"), ValueError)).to_be("invalid literal for int() with base 10: '1__0'")
    expect(raises(lambda: int("_1"), ValueError) is not None).to_be(True)
    expect(raises(lambda: int("1_"), ValueError) is not None).to_be(True)
    expect(raises(lambda: int("010", 0), ValueError) is not None).to_be(True)
    expect(int(" -1_000 ")).to_be(-1000)
    expect(int("0x_1f", 16)).to_be(31)
    expect(int("0_0", 0)).to_be(0)
    expect(raises(lambda: float("1__0"), ValueError) is not None).to_be(True)
    expect(float("1_000.5")).to_be(1000.5)


def test_big_ints():
    expect(str(2 ** 100)).to_be("1267650600228229401496703205376")
    expect(str(2 ** 64 // 3)).to_be("6148914691236517205")
    expect(str(1 << 70)).to_be("1180591620717411303424")
    expect(str(divmod(2 ** 70, 7))).to_be("(168655945816773043346, 2)")
    expect(str(sum([2 ** 64, 2 ** 64]))).to_be("36893488147419103232")
    expect(hex(2 ** 70)).to_be("0x400000000000000000")
    expect(str(int("123456789012345678901234567890"))).to_be("123456789012345678901234567890")
    expect(str(pow(2, 100, 2 ** 70 | 1))).to_be("1180591620716337561601")
    expect(str(round(2 ** 70, -5))).to_be("1180591620717411300000")


def test_math_integer_functions():
    import math
    expect(str(math.factorial(25))).to_be("15511210043330985984000000")
    expect(str(math.comb(100, 50))).to_be("100891344545564193334812497256")
    expect(math.perm(5, 2)).to_be(20)
    expect(math.isqrt(99)).to_be(9)
    expect(math.gcd(12, 18, 27)).to_be(3)
    expect(math.lcm(4, 6)).to_be(12)
    expect(math.prod([1, 2, 3, 4])).to_be(24)
    expect(math.fsum([0.1] * 10)).to_be(1.0)
    expect(str(math.floor(1e20))).to_be("100000000000000000000")
    expect((2 ** 70).to_bytes(9, "big")).to_be(b"@\x00\x00\x00\x00\x00\x00\x00\x00")
    expect((-129).to_bytes(2, "big", signed=True)).to_be(b"\xff\x7f")


def test_round_and_sum():
    expect(round(2.675, 2)).to_be(2.67)
    expect(round(1250, -2)).to_be(1200)
    expect(round(-25, -1)).to_be(-20)
    expect(round(2.5)).to_be(2)
    expect(sum([0.1] * 10)).to_be(1.0)
    expect(sum([1, 2], start=10)).to_be(13)


def test_min_max_kwargs():
    expect(max([], default=None)).to_be(None)
    expect(max(1, 2, key=lambda x: -x)).to_be(1)
    msg = raises(lambda: min(1, 2, default=0), TypeError)
    expect(msg).to_be("Cannot specify a default for min() with multiple positional arguments")
    expect(raises(lambda: max([]), ValueError)).to_be("max() iterable argument is empty")


def test_operator_error_messages():
    expect(raises(lambda: [1] + (2,), TypeError)).to_be('can only concatenate list (not "tuple") to list')
    expect(raises(lambda: None - 1, TypeError)).to_be("unsupported operand type(s) for -: 'NoneType' and 'int'")


test("extended_slice_assign", test_extended_slice_assign)
test("extended_slice_size_mismatch", test_extended_slice_size_mismatch)
test("extended_slice_delete", test_extended_slice_delete)
test("simple_slice_assign", test_simple_slice_assign)
test("list_sort_key_called_once", test_list_sort_key_called_once)
test("list_index_message", test_list_index_message)
test("set_iteration_order", test_set_iteration_order)
test("set_operation_order", test_set_operation_order)
test("set_pop_order", test_set_pop_order)
test("set_inplace_ops_mutate", test_set_inplace_ops_mutate)
test("set_numeric_equality", test_set_numeric_equality)
test("hash_matches_cpython", test_hash_matches_cpython)
test("dict_copy_keeps_order", test_dict_copy_keeps_order)
test("dict_views_repr_and_ops", test_dict_views_repr_and_ops)
test("dict_views_are_live", test_dict_views_are_live)
test("dict_inplace_or_mutates", test_dict_inplace_or_mutates)
test("dict_errors", test_dict_errors)
test("builtin_dunder_methods", test_builtin_dunder_methods)
test("range_slicing_and_repr", test_range_slicing_and_repr)
test("slice_indices", test_slice_indices)
test("zip_strict", test_zip_strict)
test("str_split_maxsplit", test_str_split_maxsplit)
test("str_padding", test_str_padding)
test("str_start_end_args", test_str_start_end_args)
test("str_case_and_predicates", test_str_case_and_predicates)
test("str_strip_none", test_str_strip_none)
test("str_join_error", test_str_join_error)
test("str_misc", test_str_misc)
test("bytes_fromhex", test_bytes_fromhex)
test("modular_pow", test_modular_pow)
test("float_modulo_and_floordiv", test_float_modulo_and_floordiv)
test("int_parsing_underscores", test_int_parsing_underscores)
test("big_ints", test_big_ints)
test("math_integer_functions", test_math_integer_functions)
test("round_and_sum", test_round_and_sum)
test("min_max_kwargs", test_min_max_kwargs)
test("operator_error_messages", test_operator_error_messages)

print("Built-in type semantics tests completed")
