from test_framework import expect, test


def test_walrus_in_listcomp():
    def func():
        _ = [y := x for x in range(5)]
        return y

    expect(func()).to_be(4)


test(
    "walrus operator in list comprehension stores to enclosing scope",
    test_walrus_in_listcomp,
)


def test_walrus_in_listcomp_use_after():
    def func():
        vals = [y := x * 2 for x in range(3)]
        return (vals, y)

    result = func()
    expect(result[0]).to_be([0, 2, 4])
    expect(result[1]).to_be(4)


test(
    "walrus in listcomp - both result and variable accessible",
    test_walrus_in_listcomp_use_after,
)


def test_walrus_in_listcomp_with_existing_var():
    def func():
        y = 100
        _ = [y := x for x in range(3)]
        return y

    expect(func()).to_be(2)


test(
    "walrus in listcomp overwrites existing variable",
    test_walrus_in_listcomp_with_existing_var,
)


def test_walrus_module_level():
    result = [z := x for x in range(4)]
    expect(z).to_be(3)
    expect(result).to_be([0, 1, 2, 3])


test("walrus operator at module level", test_walrus_module_level)

print("Walrus operator comprehension tests completed")
