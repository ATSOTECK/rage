from test_framework import test, expect

# We can't easily capture print output in the test framework,
# so we test that the kwargs are accepted without errors and
# verify type checking.

# Test 1: sep keyword
def test_sep():
    # Should not raise - just verify it's accepted
    print(1, 2, 3, sep=", ")
    print(1, 2, 3, sep="")
    print(1, 2, 3, sep=None)  # None means default (space)
    expect(True).to_be(True)

test("sep keyword accepted", test_sep)

# Test 2: end keyword
def test_end():
    print("hello", end="")
    print(" world")  # should appear on same line
    print("a", end="!\n")
    print("b", end=None)  # None means default (newline)
    expect(True).to_be(True)

test("end keyword accepted", test_end)

# Test 3: sep and end together
def test_sep_and_end():
    print("a", "b", "c", sep="-", end=".\n")
    expect(True).to_be(True)

test("sep and end together", test_sep_and_end)

# Test 4: flush keyword accepted
def test_flush():
    print("hello", flush=True)
    print("world", flush=False)
    expect(True).to_be(True)

test("flush keyword accepted", test_flush)

# Test 5: file keyword accepted
def test_file():
    print("hello", file=None)
    expect(True).to_be(True)

test("file keyword accepted", test_file)

# Test 6: sep must be string or None
def test_sep_type_error():
    try:
        print("a", "b", sep=42)
        expect("should have raised").to_be("error")
    except Exception as e:
        expect("sep must be None or a string" in str(e)).to_be(True)

test("sep type error", test_sep_type_error)

# Test 7: end must be string or None
def test_end_type_error():
    try:
        print("a", end=42)
        expect("should have raised").to_be("error")
    except Exception as e:
        expect("end must be None or a string" in str(e)).to_be(True)

test("end type error", test_end_type_error)

# Test 8: empty print with end
def test_empty_print():
    print(end="---\n")
    expect(True).to_be(True)

test("empty print with end", test_empty_print)
