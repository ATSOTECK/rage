# Test: CPython String Edge Cases
# Adapted from CPython's test_str.py - covers methods and edge cases beyond 09_strings.py

from test_framework import test, expect

def test_str_find_rfind():
    s = "hello world hello"
    expect(s.find("hello")).to_be(0)
    expect(s.rfind("hello")).to_be(12)
    expect(s.find("hello", 1)).to_be(12)
    expect(s.find("xyz")).to_be(-1)
    expect(s.rfind("xyz")).to_be(-1)
    expect(s.index("world")).to_be(6)
    try:
        s.index("xyz")
        expect("no error").to_be("ValueError")
    except ValueError:
        expect(True).to_be(True)
    try:
        s.rindex("xyz")
        expect("no error").to_be("ValueError")
    except ValueError:
        expect(True).to_be(True)

def test_str_startswith_endswith():
    s = "hello world"
    expect(s.startswith("hello")).to_be(True)
    expect(s.startswith("world")).to_be(False)
    expect(s.endswith("world")).to_be(True)
    expect(s.endswith("hello")).to_be(False)
    # Tuple of prefixes/suffixes
    expect(s.startswith(("hello", "world"))).to_be(True)
    expect(s.endswith(("hello", "world"))).to_be(True)
    expect(s.startswith(("xyz", "abc"))).to_be(False)

def test_str_count():
    s = "abracadabra"
    expect(s.count("a")).to_be(5)
    expect(s.count("abra")).to_be(2)
    expect(s.count("z")).to_be(0)
    expect(s.count("a", 1)).to_be(4)
    expect(s.count("a", 1, 5)).to_be(2)

def test_str_center_ljust_rjust():
    expect("hi".center(10)).to_be("    hi    ")
    expect("hi".ljust(10)).to_be("hi        ")
    expect("hi".rjust(10)).to_be("        hi")
    expect("hi".center(10, "*")).to_be("****hi****")
    expect("hi".ljust(10, "-")).to_be("hi--------")
    expect("hi".rjust(10, ".")).to_be("........hi")
    # Width smaller than string returns original
    expect("hello".center(3)).to_be("hello")

def test_str_zfill():
    expect("42".zfill(5)).to_be("00042")
    expect("-42".zfill(5)).to_be("-0042")
    expect("+42".zfill(5)).to_be("+0042")
    expect("42".zfill(1)).to_be("42")

def test_str_expandtabs():
    expect("01\t012\t0123\t01234".expandtabs()).to_be("01      012     0123    01234")
    expect("01\t012\t0123\t01234".expandtabs(4)).to_be("01  012 0123    01234")

def test_str_partition_rpartition():
    s = "hello-world-python"
    expect(s.partition("-")).to_be(("hello", "-", "world-python"))
    expect(s.rpartition("-")).to_be(("hello-world", "-", "python"))
    # Separator not found
    expect(s.partition("@")).to_be(("hello-world-python", "", ""))
    expect(s.rpartition("@")).to_be(("", "", "hello-world-python"))

def test_str_title_swapcase_capitalize():
    expect("hello world".title()).to_be("Hello World")
    expect("Hello World".swapcase()).to_be("hELLO wORLD")
    expect("hello WORLD".capitalize()).to_be("Hello world")
    expect("".capitalize()).to_be("")

def test_str_isalpha_isdigit_etc():
    expect("hello".isalpha()).to_be(True)
    expect("hello1".isalpha()).to_be(False)
    expect("12345".isdigit()).to_be(True)
    expect("123a".isdigit()).to_be(False)
    expect("hello123".isalnum()).to_be(True)
    expect("hello 123".isalnum()).to_be(False)
    expect("   ".isspace()).to_be(True)
    expect(" a ".isspace()).to_be(False)
    expect("HELLO".isupper()).to_be(True)
    expect("Hello".isupper()).to_be(False)
    expect("hello".islower()).to_be(True)
    expect("Hello".islower()).to_be(False)
    # Empty string edge cases
    expect("".isalpha()).to_be(False)
    expect("".isdigit()).to_be(False)

def test_str_lstrip_rstrip():
    expect("  hello  ".lstrip()).to_be("hello  ")
    expect("  hello  ".rstrip()).to_be("  hello")
    expect("xxhelloxx".lstrip("x")).to_be("helloxx")
    expect("xxhelloxx".rstrip("x")).to_be("xxhello")
    expect("abcba".lstrip("abc")).to_be("")

def test_str_format_method():
    expect("{} {}".format("hello", "world")).to_be("hello world")
    expect("{0} {1}".format("hello", "world")).to_be("hello world")
    expect("{1} {0}".format("hello", "world")).to_be("world hello")
    expect("{name}".format(name="world")).to_be("world")

def test_str_format_spec():
    expect("{:>10}".format("hi")).to_be("        hi")
    expect("{:<10}".format("hi")).to_be("hi        ")
    expect("{:^10}".format("hi")).to_be("    hi    ")
    expect("{:.2f}".format(3.14159)).to_be("3.14")
    expect("{:05d}".format(42)).to_be("00042")

def test_str_multiply_edge():
    expect("a" * -1).to_be("")
    expect("" * 1000).to_be("")
    expect("a" * 0).to_be("")

def test_str_comparison_order():
    expect("a" < "b").to_be(True)
    expect("b" > "a").to_be(True)
    expect("abc" < "abd").to_be(True)
    expect("abc" < "abcd").to_be(True)
    expect("" < "a").to_be(True)

def test_str_contains_empty():
    expect("" in "hello").to_be(True)
    expect("" in "").to_be(True)

def test_str_split_maxsplit():
    expect("a b c d".split(" ", 1)).to_be(["a", "b c d"])
    expect("a b c d".split(" ", 2)).to_be(["a", "b", "c d"])
    expect("a b c d".rsplit(" ", 1)).to_be(["a b c", "d"])

def test_str_splitlines():
    expect("hello\nworld\n".splitlines()).to_be(["hello", "world"])
    expect("hello\nworld\n".splitlines(True)).to_be(["hello\n", "world\n"])
    expect("".splitlines()).to_be([])

def test_str_join_types():
    expect(" ".join(["a", "b", "c"])).to_be("a b c")
    expect("".join(["a", "b", "c"])).to_be("abc")
    try:
        " ".join([1, 2, 3])
        expect("no error").to_be("TypeError")
    except TypeError:
        expect(True).to_be(True)

def test_str_replace_count():
    expect("aaa".replace("a", "b", 1)).to_be("baa")
    expect("aaa".replace("a", "b", 2)).to_be("bba")
    expect("aaa".replace("a", "b")).to_be("bbb")

def test_str_slice_step():
    s = "abcdefgh"
    expect(s[::2]).to_be("aceg")
    expect(s[::-1]).to_be("hgfedcba")
    expect(s[1:5:2]).to_be("bd")

def test_str_immutability():
    try:
        s = "hello"
        s[0] = "x"
        expect("no error").to_be("TypeError")
    except TypeError:
        expect(True).to_be(True)

def test_str_ord_chr():
    expect(ord("A")).to_be(65)
    expect(chr(65)).to_be("A")
    expect(ord(chr(1000))).to_be(1000)
    expect(chr(ord("z"))).to_be("z")

def test_str_repr_vs_str():
    expect(str(42)).to_be("42")
    expect(str(3.14)).to_be("3.14")
    expect(repr("hello")).to_be("'hello'")

# Register all tests
test("str_find_rfind", test_str_find_rfind)
test("str_startswith_endswith", test_str_startswith_endswith)
test("str_count", test_str_count)
test("str_center_ljust_rjust", test_str_center_ljust_rjust)
test("str_zfill", test_str_zfill)
test("str_expandtabs", test_str_expandtabs)
test("str_partition_rpartition", test_str_partition_rpartition)
test("str_title_swapcase_capitalize", test_str_title_swapcase_capitalize)
test("str_isalpha_isdigit_etc", test_str_isalpha_isdigit_etc)
test("str_lstrip_rstrip", test_str_lstrip_rstrip)
test("str_format_method", test_str_format_method)
test("str_format_spec", test_str_format_spec)
test("str_multiply_edge", test_str_multiply_edge)
test("str_comparison_order", test_str_comparison_order)
test("str_contains_empty", test_str_contains_empty)
test("str_split_maxsplit", test_str_split_maxsplit)
test("str_splitlines", test_str_splitlines)
test("str_join_types", test_str_join_types)
test("str_replace_count", test_str_replace_count)
test("str_slice_step", test_str_slice_step)
test("str_immutability", test_str_immutability)
test("str_ord_chr", test_str_ord_chr)
test("str_repr_vs_str", test_str_repr_vs_str)

print("CPython str tests completed")
