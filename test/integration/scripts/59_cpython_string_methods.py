# Test: CPython String Methods - Deep Dive
# Adapted from CPython's test_str.py - covers additional string method edge cases
# beyond 09_strings.py and 30_cpython_str.py

from test_framework import test, expect

def test_split_no_args():
    # split() with no args splits on whitespace and strips leading/trailing
    expect("  hello  world  ".split()).to_be(["hello", "world"])
    expect("hello".split()).to_be(["hello"])
    expect("".split()).to_be([])
    expect("   ".split()).to_be([])
    # Tabs, newlines, multiple spaces
    expect("a\tb\nc  d".split()).to_be(["a", "b", "c", "d"])

def test_split_delimiter_edge_cases():
    expect("aXXbXXc".split("XX")).to_be(["a", "b", "c"])
    expect("XXaXXbXX".split("XX")).to_be(["", "a", "b", ""])
    # Split on multi-char delimiter
    expect("one<>two<>three".split("<>")).to_be(["one", "two", "three"])
    # Split with empty results
    expect("a,,b,,c".split(",")).to_be(["a", "", "b", "", "c"])

def test_split_maxsplit_edge():
    expect("a.b.c.d.e".split(".", 0)).to_be(["a.b.c.d.e"])
    expect("a.b.c.d.e".split(".", 1)).to_be(["a", "b.c.d.e"])
    expect("a.b.c.d.e".split(".", 10)).to_be(["a", "b", "c", "d", "e"])
    # maxsplit with whitespace split
    expect("a b c d".split(" ", 2)).to_be(["a", "b", "c d"])

def test_rsplit_basics():
    expect("a.b.c.d".rsplit(".")).to_be(["a", "b", "c", "d"])
    expect("a.b.c.d".rsplit(".", 1)).to_be(["a.b.c", "d"])
    expect("a.b.c.d".rsplit(".", 2)).to_be(["a.b", "c", "d"])
    expect("a b c d".rsplit(" ", 1)).to_be(["a b c", "d"])

def test_partition_edge_cases():
    # Partition with multi-char separator
    expect("hello::world::test".partition("::")).to_be(("hello", "::", "world::test"))
    expect("hello::world::test".rpartition("::")).to_be(("hello::world", "::", "test"))
    # String equals separator
    expect("abc".partition("abc")).to_be(("", "abc", ""))
    expect("abc".rpartition("abc")).to_be(("", "abc", ""))

def test_expandtabs_edge():
    expect("\t".expandtabs()).to_be("        ")
    expect("\t".expandtabs(4)).to_be("    ")
    expect("no tabs".expandtabs()).to_be("no tabs")
    expect("".expandtabs()).to_be("")
    expect("\t\t".expandtabs(4)).to_be("        ")

def test_center_edge_cases():
    # Already wider than requested
    expect("hello".center(3)).to_be("hello")
    expect("hello".center(5)).to_be("hello")
    # Odd padding
    expect("hi".center(7)).to_be("  hi   ")
    expect("hi".center(7, "*")).to_be("**hi***")
    expect("x".center(4, "-")).to_be("-x--")

def test_ljust_rjust_edge():
    expect("test".ljust(4)).to_be("test")
    expect("test".ljust(8, ".")).to_be("test....")
    expect("test".rjust(4)).to_be("test")
    expect("test".rjust(8, "0")).to_be("0000test")
    # Fill char is single character
    expect("a".ljust(5, "#")).to_be("a####")
    expect("a".rjust(5, "#")).to_be("####a")

def test_zfill_edge():
    expect("".zfill(5)).to_be("00000")
    expect("test".zfill(3)).to_be("test")
    expect("-".zfill(5)).to_be("-0000")
    expect("+".zfill(5)).to_be("+0000")
    # zfill with numbers
    expect("42".zfill(5)).to_be("00042")
    expect("-42".zfill(5)).to_be("-0042")
    expect("+42".zfill(5)).to_be("+0042")

def test_swapcase_edge():
    expect("".swapcase()).to_be("")
    expect("abc123".swapcase()).to_be("ABC123")
    expect("ABC123".swapcase()).to_be("abc123")
    expect("HeLLo WoRLd".swapcase()).to_be("hEllO wOrlD")

def test_title_edge():
    expect("".title()).to_be("")
    expect("hello world".title()).to_be("Hello World")
    # title with non-alpha separators
    expect("hello-world".title()).to_be("Hello-World")
    # Simple single word
    expect("hello".title()).to_be("Hello")
    expect("a".title()).to_be("A")

def test_isdigit_deep():
    expect("0".isdigit()).to_be(True)
    expect("0123456789".isdigit()).to_be(True)
    expect("".isdigit()).to_be(False)
    expect(" ".isdigit()).to_be(False)
    expect("12.34".isdigit()).to_be(False)
    expect("-1".isdigit()).to_be(False)

def test_isalpha_deep():
    expect("abc".isalpha()).to_be(True)
    expect("ABC".isalpha()).to_be(True)
    expect("abcDEF".isalpha()).to_be(True)
    expect("".isalpha()).to_be(False)
    expect("abc123".isalpha()).to_be(False)
    expect("abc def".isalpha()).to_be(False)

def test_isalnum_deep():
    expect("abc123".isalnum()).to_be(True)
    expect("abc".isalnum()).to_be(True)
    expect("123".isalnum()).to_be(True)
    expect("".isalnum()).to_be(False)
    expect("abc 123".isalnum()).to_be(False)
    expect("abc!".isalnum()).to_be(False)

def test_isspace_deep():
    expect(" ".isspace()).to_be(True)
    expect("\t".isspace()).to_be(True)
    expect("\n".isspace()).to_be(True)
    expect("\r".isspace()).to_be(True)
    expect(" \t\n\r".isspace()).to_be(True)
    expect("".isspace()).to_be(False)
    expect(" a ".isspace()).to_be(False)

def test_isupper_islower_deep():
    expect("HELLO".isupper()).to_be(True)
    expect("HELLO123".isupper()).to_be(True)
    expect("Hello".isupper()).to_be(False)
    expect("hello".islower()).to_be(True)
    expect("hello123".islower()).to_be(True)
    expect("Hello".islower()).to_be(False)
    expect("123".isupper()).to_be(False)
    expect("123".islower()).to_be(False)

def test_count_with_range():
    s = "ababababab"
    expect(s.count("ab")).to_be(5)
    expect(s.count("ab", 2)).to_be(4)
    expect(s.count("ab", 2, 6)).to_be(2)
    expect(s.count("ab", 0, 0)).to_be(0)
    expect("".count("")).to_be(1)
    expect("abc".count("")).to_be(4)  # Empty string found between each char + start + end

def test_str_encode():
    # encode returns bytes
    b = "hello".encode()
    expect(type(b).__name__).to_be("bytes")
    expect(b).to_be(b"hello")

def test_str_concatenation_patterns():
    # Building strings in different ways
    parts = ["hello", " ", "world"]
    result = ""
    for p in parts:
        result = result + p
    expect(result).to_be("hello world")
    # Join is preferred
    expect("".join(parts)).to_be("hello world")

def test_str_find_with_range():
    s = "abcabcabc"
    expect(s.find("abc")).to_be(0)
    expect(s.find("abc", 1)).to_be(3)
    expect(s.find("abc", 4)).to_be(6)
    expect(s.find("abc", 7)).to_be(-1)
    expect(s.rfind("abc")).to_be(6)
    # rfind without range restriction
    expect("hello world hello".rfind("hello")).to_be(12)
    expect("hello world hello".rfind("world")).to_be(6)

def test_str_strip_chars():
    expect("www.example.com".strip("cmowz.")).to_be("example")
    expect("xxxhelloxxx".strip("x")).to_be("hello")
    expect("   spacey   ".strip()).to_be("spacey")

def test_str_multiply_and_add():
    # Combining multiply and add
    expect("ha" * 3).to_be("hahaha")
    expect(3 * "ha").to_be("hahaha")
    expect("a" * 5 + "b" * 3).to_be("aaaaabbb")
    expect(("ab" + "cd") * 2).to_be("abcdabcd")

def test_str_comparison_mixed():
    # String comparison is lexicographic
    expect("abc" == "abc").to_be(True)
    expect("abc" < "abd").to_be(True)
    expect("abc" < "abcd").to_be(True)
    expect("" < "a").to_be(True)
    expect("z" > "a").to_be(True)
    expect("Z" < "a").to_be(True)  # uppercase before lowercase in ASCII
    expect("9" < "A").to_be(True)  # digits before letters in ASCII

def test_str_bool():
    expect(bool("")).to_be(False)
    expect(bool("x")).to_be(True)
    expect(bool(" ")).to_be(True)
    expect(bool("0")).to_be(True)

def test_str_iteration():
    chars = []
    for c in "hello":
        chars.append(c)
    expect(chars).to_be(["h", "e", "l", "l", "o"])
    # list() on string
    expect(list("abc")).to_be(["a", "b", "c"])

def test_str_min_max():
    expect(min("hello")).to_be("e")
    expect(max("hello")).to_be("o")
    expect(min("zAb")).to_be("A")
    expect(max("zAb")).to_be("z")

def test_str_format_number_padding():
    # Number formatting with various specs
    expect("{:d}".format(42)).to_be("42")
    expect("{:05d}".format(42)).to_be("00042")
    expect("{:>10d}".format(42)).to_be("        42")
    expect("{:<10d}".format(42)).to_be("42        ")
    expect("{:+d}".format(42)).to_be("+42")
    expect("{:+d}".format(-42)).to_be("-42")

def test_str_format_float_specs():
    expect("{:.2f}".format(3.14159)).to_be("3.14")
    expect("{:.0f}".format(3.7)).to_be("4")
    expect("{:10.2f}".format(3.14)).to_be("      3.14")
    expect("{:e}".format(1000.0)).to_be("1.000000e+03")

def test_str_multiple_replace():
    s = "the cat sat on the mat"
    s = s.replace("cat", "dog")
    s = s.replace("mat", "rug")
    expect(s).to_be("the dog sat on the rug")

def test_str_chained_methods():
    expect("  Hello World  ".strip().lower()).to_be("hello world")
    expect("hello".upper().replace("L", "R")).to_be("HERRO")
    expect("  abc  ".strip().upper()).to_be("ABC")

def test_str_splitlines_variations():
    # Different line endings
    expect("a\nb\nc".splitlines()).to_be(["a", "b", "c"])
    expect("a\r\nb\r\nc".splitlines()).to_be(["a", "b", "c"])
    expect("a\rb\rc".splitlines()).to_be(["a", "b", "c"])
    expect("single line".splitlines()).to_be(["single line"])
    expect("".splitlines()).to_be([])
    # keepends
    expect("a\nb\n".splitlines(True)).to_be(["a\n", "b\n"])

# Register all tests
test("split_no_args", test_split_no_args)
test("split_delimiter_edge_cases", test_split_delimiter_edge_cases)
test("split_maxsplit_edge", test_split_maxsplit_edge)
test("rsplit_basics", test_rsplit_basics)
test("partition_edge_cases", test_partition_edge_cases)
test("expandtabs_edge", test_expandtabs_edge)
test("center_edge_cases", test_center_edge_cases)
test("ljust_rjust_edge", test_ljust_rjust_edge)
test("zfill_edge", test_zfill_edge)
test("swapcase_edge", test_swapcase_edge)
test("title_edge", test_title_edge)
test("isdigit_deep", test_isdigit_deep)
test("isalpha_deep", test_isalpha_deep)
test("isalnum_deep", test_isalnum_deep)
test("isspace_deep", test_isspace_deep)
test("isupper_islower_deep", test_isupper_islower_deep)
test("count_with_range", test_count_with_range)
test("str_encode", test_str_encode)
test("str_concatenation_patterns", test_str_concatenation_patterns)
test("str_find_with_range", test_str_find_with_range)
test("str_strip_chars", test_str_strip_chars)
test("str_multiply_and_add", test_str_multiply_and_add)
test("str_comparison_mixed", test_str_comparison_mixed)
test("str_bool", test_str_bool)
test("str_iteration", test_str_iteration)
test("str_min_max", test_str_min_max)
test("str_format_number_padding", test_str_format_number_padding)
test("str_format_float_specs", test_str_format_float_specs)
test("str_multiple_replace", test_str_multiple_replace)
test("str_chained_methods", test_str_chained_methods)
test("str_splitlines_variations", test_str_splitlines_variations)

print("CPython string methods tests completed")
