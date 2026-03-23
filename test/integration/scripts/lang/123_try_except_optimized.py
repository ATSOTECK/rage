from test_framework import expect, test

# Test that exceptions from optimized arithmetic fallback paths are catchable


class BadAdd:
    """Object that raises TypeError when used in addition."""

    pass


def test_try_except_increment_fallback():
    caught = False
    try:
        x = BadAdd()
        x += 1
    except:
        caught = True
    expect(caught).to_be(True)


test("try/except catches increment fallback error", test_try_except_increment_fallback)


def test_try_except_decrement_fallback():
    caught = False
    try:
        x = BadAdd()
        x -= 1
    except:
        caught = True
    expect(caught).to_be(True)


test("try/except catches decrement fallback error", test_try_except_decrement_fallback)


def test_try_except_negate_fallback():
    caught = False
    try:
        x = BadAdd()
        x = -x
    except:
        caught = True
    expect(caught).to_be(True)


test("try/except catches negate fallback error", test_try_except_negate_fallback)


def test_try_except_add_non_int():
    caught = False
    try:
        result = BadAdd() + BadAdd()
    except:
        caught = True
    expect(caught).to_be(True)


test("try/except catches add non-int error", test_try_except_add_non_int)


def test_try_except_subtract_non_int():
    caught = False
    try:
        result = BadAdd() - BadAdd()
    except:
        caught = True
    expect(caught).to_be(True)


test("try/except catches subtract non-int error", test_try_except_subtract_non_int)


def test_try_except_multiply_non_int():
    caught = False
    try:
        result = BadAdd() * BadAdd()
    except:
        caught = True
    expect(caught).to_be(True)


test("try/except catches multiply non-int error", test_try_except_multiply_non_int)


# Test that exceptions from OpGetIter are catchable
def test_try_except_get_iter():
    caught = False
    try:
        for x in 42:
            pass
    except:
        caught = True
    expect(caught).to_be(True)


test("try/except catches iter on non-iterable", test_try_except_get_iter)


# Test that exceptions from OpLen* are catchable
def test_try_except_len_type_error():
    caught = False
    try:
        x = len(42)
    except:
        caught = True
    expect(caught).to_be(True)


test("try/except catches len() TypeError", test_try_except_len_type_error)


# Test that normal optimized paths still work correctly
def test_normal_increment():
    x = 5
    x += 1
    expect(x).to_be(6)


test("normal increment still works", test_normal_increment)


def test_normal_arithmetic():
    expect(3 + 4).to_be(7)
    expect(10 - 3).to_be(7)
    expect(6 * 7).to_be(42)


test("normal optimized arithmetic still works", test_normal_arithmetic)


def test_normal_len():
    expect(len([1, 2, 3])).to_be(3)
    expect(len("hello")).to_be(5)
    expect(len((1, 2))).to_be(2)
    expect(len({"a": 1})).to_be(1)


test("normal len still works", test_normal_len)


def test_normal_iter():
    result = []
    for x in [1, 2, 3]:
        result.append(x)
    expect(result).to_be([1, 2, 3])


test("normal iteration still works", test_normal_iter)

print("Optimized opcode try/except tests completed")
