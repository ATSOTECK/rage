from test_framework import test, expect

# Test 1: Basic __format__ with format() builtin
def test_basic_format():
    class Color:
        def __init__(self, r, g, b):
            self.r = r
            self.g = g
            self.b = b

        def __format__(self, spec):
            if spec == "hex":
                return "#{:02x}{:02x}{:02x}".format(self.r, self.g, self.b)
            return "Color({}, {}, {})".format(self.r, self.g, self.b)

    c = Color(255, 128, 0)
    expect(format(c, "hex")).to_be("#ff8000")
    expect(format(c, "")).to_be("Color(255, 128, 0)")

test("basic __format__ with format() builtin", test_basic_format)

# Test 2: format() with no spec passes empty string
def test_format_no_spec():
    class Obj:
        def __format__(self, spec):
            return "spec=" + repr(spec)

    expect(format(Obj())).to_be("spec=''")

test("format() with no spec passes empty string", test_format_no_spec)

# Test 3: __format__ with f-string
def test_format_fstring():
    class Num:
        def __init__(self, val):
            self.val = val

        def __format__(self, spec):
            if spec == "doubled":
                return str(self.val * 2)
            return str(self.val)

    n = Num(21)
    result = f"{n:doubled}"
    expect(result).to_be("42")

test("__format__ with f-string", test_format_fstring)

# Test 4: __format__ with str.format()
def test_format_str_format():
    class Tag:
        def __init__(self, name):
            self.name = name

        def __format__(self, spec):
            if spec == "upper":
                return self.name.upper()
            return self.name

    t = Tag("hello")
    expect("{:upper}".format(t)).to_be("HELLO")
    expect("{}".format(t)).to_be("hello")

test("__format__ with str.format()", test_format_str_format)

# Test 5: __format__ inherited from parent
def test_inherited_format():
    class Base:
        def __format__(self, spec):
            return "base:" + spec

    class Child(Base):
        pass

    expect(format(Child(), "x")).to_be("base:x")

test("__format__ inherited from parent", test_inherited_format)

# Test 6: Child overrides __format__
def test_override_format():
    class Base:
        def __format__(self, spec):
            return "base"

    class Child(Base):
        def __format__(self, spec):
            return "child"

    expect(format(Child())).to_be("child")

test("child overrides __format__", test_override_format)

# Test 7: format() on built-in types still works
def test_builtin_format():
    expect(format(42, "05d")).to_be("00042")
    expect(format(3.14, ".1f")).to_be("3.1")
    expect(format("hi", ">5")).to_be("   hi")

test("format() on built-in types", test_builtin_format)

# Test 8: f-string format spec on built-ins
def test_fstring_builtin_spec():
    x = 42
    expect(f"{x:05d}").to_be("00042")
    pi = 3.14159
    expect(f"{pi:.2f}").to_be("3.14")
    name = "hi"
    expect(f"{name:>5}").to_be("   hi")

test("f-string format spec on built-ins", test_fstring_builtin_spec)

# Test 9: __format__ with no __format__ falls back to str
def test_no_format_fallback():
    class Plain:
        def __str__(self):
            return "plain_str"

    # format() with empty spec should fall back to str()
    expect(format(Plain())).to_be("plain_str")

test("no __format__ falls back to __str__", test_no_format_fallback)

# Test 10: __format__ error propagates
def test_format_error_propagates():
    class Bad:
        def __format__(self, spec):
            raise ValueError("bad format: " + spec)

    try:
        format(Bad(), "xyz")
        expect(True).to_be(False)
    except ValueError as e:
        expect(str(e)).to_be("bad format: xyz")

test("__format__ error propagates", test_format_error_propagates)

# Test 11: f-string with conversion AND format spec
def test_fstring_conversion_and_spec():
    x = 42
    # !s converts to string first, then applies format spec
    expect(f"{x!s:>5}").to_be("   42")

test("f-string with conversion and format spec", test_fstring_conversion_and_spec)

# Test 12: f-string without format spec still works
def test_fstring_no_spec():
    x = 42
    expect(f"{x}").to_be("42")
    expect(f"hello {x} world").to_be("hello 42 world")

test("f-string without format spec still works", test_fstring_no_spec)
