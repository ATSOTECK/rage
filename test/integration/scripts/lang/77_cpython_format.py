# Test: CPython String Formatting Edge Cases
# Adapted from CPython's test_format.py and test_string.py

from test_framework import test, expect

# === f-string basics ===
def test_fstring_basic():
    name = "world"
    expect(f"hello {name}").to_be("hello world")
    x = 42
    expect(f"x={x}").to_be("x=42")
    expect(f"").to_be("")
    expect(f"no interpolation").to_be("no interpolation")

def test_fstring_expressions():
    expect(f"{1 + 2}").to_be("3")
    expect(f"{2 * 3 + 1}").to_be("7")
    expect(f"{'hello'.upper()}").to_be("HELLO")
    expect(f"{len('abc')}").to_be("3")

def test_fstring_conversions():
    expect(f"{'hello'!r}").to_be("'hello'")
    expect(f"{42!s}").to_be("42")

def test_fstring_nested_quotes():
    expect(f"{'hello'}").to_be("hello")
    d = {"key": "value"}
    expect(f"{d['key']}").to_be("value")

def test_fstring_multivalue():
    x = 1
    y = 2
    expect(f"{x} + {y} = {x + y}").to_be("1 + 2 = 3")

# === str.format() ===
def test_str_format_basic():
    expect("{} {}".format("hello", "world")).to_be("hello world")
    expect("{0} {1}".format("hello", "world")).to_be("hello world")
    expect("{1} {0}".format("hello", "world")).to_be("world hello")

def test_str_format_named():
    expect("{name} is {age}".format(name="Alice", age=30)).to_be("Alice is 30")

def test_str_format_types():
    expect("{:d}".format(42)).to_be("42")
    expect("{:f}".format(3.14)).to_be("3.140000")
    expect("{:.2f}".format(3.14159)).to_be("3.14")
    expect("{:s}".format("hello")).to_be("hello")

def test_str_format_padding():
    expect("{:10}".format("hello")).to_be("hello     ")
    expect("{:<10}".format("hello")).to_be("hello     ")
    expect("{:>10}".format("hello")).to_be("     hello")
    expect("{:^10}".format("hello")).to_be("  hello   ")
    expect("{:*^10}".format("hello")).to_be("**hello***")

def test_str_format_numbers():
    expect("{:05d}".format(42)).to_be("00042")
    expect("{:+d}".format(42)).to_be("+42")
    expect("{:+d}".format(-42)).to_be("-42")
    expect("{:x}".format(255)).to_be("ff")
    expect("{:o}".format(8)).to_be("10")
    expect("{:b}".format(10)).to_be("1010")

# === % formatting ===
def test_percent_format_basic():
    expect("hello %s" % "world").to_be("hello world")
    expect("%d items" % 42).to_be("42 items")
    expect("%s and %s" % ("a", "b")).to_be("a and b")

def test_percent_format_types():
    expect("%d" % 42).to_be("42")
    expect("%f" % 3.14).to_be("3.140000")
    expect("%.2f" % 3.14159).to_be("3.14")
    expect("%s" % "hello").to_be("hello")
    expect("%r" % "hello").to_be("'hello'")
    expect("%x" % 255).to_be("ff")
    expect("%o" % 8).to_be("10")

def test_percent_format_padding():
    expect("%10s" % "hello").to_be("     hello")
    expect("%-10s" % "hello").to_be("hello     ")
    expect("%05d" % 42).to_be("00042")

def test_percent_format_multiple():
    expect("%s=%d" % ("x", 42)).to_be("x=42")
    expect("%d+%d=%d" % (1, 2, 3)).to_be("1+2=3")

# === String multiplication and concatenation ===
def test_str_multiply():
    expect("ab" * 3).to_be("ababab")
    expect(3 * "xy").to_be("xyxyxy")
    expect("a" * 0).to_be("")
    expect("a" * -1).to_be("")

def test_str_concat():
    expect("hello" + " " + "world").to_be("hello world")
    expect("" + "a").to_be("a")
    expect("a" + "").to_be("a")

# === str() conversions ===
def test_str_conversion():
    expect(str(42)).to_be("42")
    expect(str(3.14)).to_be("3.14")
    expect(str(True)).to_be("True")
    expect(str(False)).to_be("False")
    expect(str(None)).to_be("None")
    expect(str([1, 2, 3])).to_be("[1, 2, 3]")
    expect(str((1, 2))).to_be("(1, 2)")
    expect(str({"a": 1})).to_be("{'a': 1}")

# === repr() ===
def test_repr_types():
    expect(repr(42)).to_be("42")
    expect(repr("hello")).to_be("'hello'")
    expect(repr(True)).to_be("True")
    expect(repr(None)).to_be("None")
    expect(repr([1, 2])).to_be("[1, 2]")
    expect(repr((1,))).to_be("(1,)")

# Register all tests
test("fstring_basic", test_fstring_basic)
test("fstring_expressions", test_fstring_expressions)
test("fstring_conversions", test_fstring_conversions)
test("fstring_nested_quotes", test_fstring_nested_quotes)
test("fstring_multivalue", test_fstring_multivalue)
test("str_format_basic", test_str_format_basic)
test("str_format_named", test_str_format_named)
test("str_format_types", test_str_format_types)
test("str_format_padding", test_str_format_padding)
test("str_format_numbers", test_str_format_numbers)
test("percent_format_basic", test_percent_format_basic)
test("percent_format_types", test_percent_format_types)
test("percent_format_padding", test_percent_format_padding)
test("percent_format_multiple", test_percent_format_multiple)
test("str_multiply", test_str_multiply)
test("str_concat", test_str_concat)
test("str_conversion", test_str_conversion)
test("repr_types", test_repr_types)

print("CPython format tests completed")
