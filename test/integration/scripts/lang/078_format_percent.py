from test_framework import test, expect

# Test 1: Basic %s and %d
def test_basic():
    expect("hello %s" % "world").to_be("hello world")
    expect("%d" % 42).to_be("42")
    expect("%i" % 42).to_be("42")

test("basic %s and %d", test_basic)

# Test 2: Multiple positional args via tuple
def test_tuple_args():
    expect("%s is %d" % ("Alice", 30)).to_be("Alice is 30")
    expect("%d + %d = %d" % (1, 2, 3)).to_be("1 + 2 = 3")

test("tuple positional args", test_tuple_args)

# Test 3: Dict-based formatting
def test_dict_format():
    expect("%(name)s is %(age)d" % {"name": "Alice", "age": 30}).to_be("Alice is 30")
    expect("%(x)s and %(y)s" % {"x": "hello", "y": "world"}).to_be("hello and world")

test("dict-based formatting", test_dict_format)

# Test 4: Float formatting
def test_float():
    expect("%.2f" % 3.14159).to_be("3.14")
    expect("%f" % 1.5).to_be("1.500000")
    expect("%.0f" % 3.7).to_be("4")

test("float formatting", test_float)

# Test 5: Scientific notation
def test_scientific():
    result = "%e" % 123456.789
    expect(result.startswith("1.23456")).to_be(True)
    expect("e+" in result).to_be(True)
    result_upper = "%E" % 123456.789
    expect("E+" in result_upper).to_be(True)

test("scientific notation %e/%E", test_scientific)

# Test 6: General format %g/%G
def test_general():
    expect("%g" % 1.0).to_be("1")
    expect("%g" % 0.0001).to_be("0.0001")

test("general format %g/%G", test_general)

# Test 7: Hex and octal
def test_hex_oct():
    expect("%x" % 255).to_be("ff")
    expect("%X" % 255).to_be("FF")
    expect("%o" % 8).to_be("10")

test("hex %x/%X and octal %o", test_hex_oct)

# Test 8: Alt form with #
def test_alt_form():
    expect("%#x" % 255).to_be("0xff")
    expect("%#X" % 255).to_be("0XFF")
    expect("%#o" % 8).to_be("0o10")

test("alt form # flag", test_alt_form)

# Test 9: Character %c
def test_char():
    expect("%c" % 65).to_be("A")
    expect("%c" % 97).to_be("a")
    expect("%c" % "Z").to_be("Z")

test("character %c", test_char)

# Test 10: Sign flags + and space
def test_sign_flags():
    expect("%+d" % 42).to_be("+42")
    expect("%+d" % -42).to_be("-42")
    expect("% d" % 42).to_be(" 42")
    expect("% d" % -42).to_be("-42")

test("sign flags + and space", test_sign_flags)

# Test 11: Width and alignment
def test_width():
    expect("%10d" % 42).to_be("        42")
    expect("%-10d|" % 42).to_be("42        |")

test("width and alignment", test_width)

# Test 12: Zero padding
def test_zero_pad():
    expect("%05d" % 42).to_be("00042")
    expect("%05d" % -42).to_be("-0042")
    expect("%+06d" % 42).to_be("+00042")

test("zero padding", test_zero_pad)

# Test 13: Star width and precision
def test_star():
    expect("%*d" % (10, 42)).to_be("        42")
    expect("%.*f" % (2, 3.14159)).to_be("3.14")

test("star width and precision", test_star)

# Test 14: Literal %%
def test_literal_percent():
    expect("100%%" % ()).to_be("100%")
    expect("%d%%" % 50).to_be("50%")

test("literal %% escape", test_literal_percent)

# Test 15: %r repr formatting
def test_repr_format():
    expect("%r" % "hello").to_be("'hello'")
    expect("%r" % 42).to_be("42")

test("%r repr formatting", test_repr_format)

# Test 16: Precision on strings
def test_string_precision():
    expect("%.3s" % "hello").to_be("hel")
    expect("%.10s" % "hi").to_be("hi")

test("string precision truncation", test_string_precision)

# Test 17: Not enough args error
def test_not_enough_args():
    try:
        "%s %s" % ("only one",)
        expect("should have raised").to_be("error")
    except Exception as e:
        expect("not enough arguments" in str(e)).to_be(True)

test("not enough args error", test_not_enough_args)

# Test 18: Too many args error
def test_too_many_args():
    try:
        "%s" % ("a", "b")
        expect("should have raised").to_be("error")
    except Exception as e:
        expect("not all arguments converted" in str(e)).to_be(True)

test("too many args error", test_too_many_args)

# Test 19: Dict missing key error
def test_dict_missing_key():
    try:
        "%(missing)s" % {"name": "Alice"}
        expect("should have raised").to_be("error")
    except Exception as e:
        expect("missing" in str(e)).to_be(True)

test("dict missing key error", test_dict_missing_key)

# Test 20: Single value (not tuple)
def test_single_value():
    expect("%s" % 42).to_be("42")
    expect("%d" % True).to_be("1")

test("single value formatting", test_single_value)
