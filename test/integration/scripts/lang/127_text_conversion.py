# Test: repr(), str(), ascii(), format() and str.format() text conversion
# Expected values match CPython 3.14

from test_framework import test, expect

# === str repr: quote selection and escaping ===
def test_str_repr_quotes():
    expect(repr("abc")).to_be("'abc'")
    expect(repr("it's")).to_be('"it\'s"')
    expect(repr('say "hi"')).to_be("'say \"hi\"'")
    expect(repr("both ' and \"")).to_be("'both \\' and \"'")
    expect(repr("")).to_be("''")

def test_str_repr_escapes():
    expect(repr("a\nb")).to_be("'a\\nb'")
    expect(repr("tab\there\r")).to_be("'tab\\there\\r'")
    expect(repr("back\\slash")).to_be("'back\\\\slash'")
    expect(repr("\x00\x1b\x7f")).to_be("'\\x00\\x1b\\x7f'")
    # Printable non-ASCII stays literal; non-printable is escaped
    expect(repr("caf\u00e9")).to_be("'caf\u00e9'")
    expect(repr("\x85\xa0\xad")).to_be("'\\x85\\xa0\\xad'")
    expect(repr("\u2028\u200b\u3000")).to_be("'\\u2028\\u200b\\u3000'")
    expect(repr("\U000e0001")).to_be("'\\U000e0001'")
    expect(repr("\U0001F600")).to_be("'\U0001F600'")

def test_str_repr_in_containers():
    expect(str(["it's", "a\nb"])).to_be("[\"it's\", 'a\\nb']")
    expect(repr(("x",))).to_be("('x',)")
    expect(str({"k'": "v\n"})).to_be("{\"k'\": 'v\\n'}")
    expect(str([None, True, 1.0, -0.0, (1+0j)])).to_be("[None, True, 1.0, -0.0, (1+0j)]")

def test_bytes_repr():
    expect(repr(b"abc")).to_be("b'abc'")
    expect(repr(b"\x00\xff'")).to_be("b\"\\x00\\xff'\"")
    expect(repr(b"'\"")).to_be("b'\\'\"'")
    expect(repr(b"\t\n\r\\")).to_be("b'\\t\\n\\r\\\\'")
    expect(str(b"x")).to_be("b'x'")

def test_ascii():
    expect(ascii("caf\u00e9")).to_be("'caf\\xe9'")
    expect(ascii("it's")).to_be('"it\'s"')
    expect(ascii("\u0100\U0001F600")).to_be("'\\u0100\\U0001f600'")
    expect(ascii(["\u00e9", 'a"'])).to_be("['\\xe9', 'a\"']")
    expect(ascii({"\u00e9": 1})).to_be("{'\\xe9': 1}")

# === Recursion guards ===
def test_self_referential_containers():
    a = [1, 2]
    a.append(a)
    expect(repr(a)).to_be("[1, 2, [...]]")
    expect(str(a)).to_be("[1, 2, [...]]")
    d = {}
    d["self"] = d
    expect(repr(d)).to_be("{'self': {...}}")
    t = ([],)
    t[0].append(t)
    expect(repr(t)).to_be("([(...)],)")
    # The same object twice (not a cycle) is shown in full
    inner = [1]
    expect(repr([inner, inner])).to_be("[[1], [1]]")

class RecursiveRepr:
    def __repr__(self):
        return repr(self)

def test_recursive_repr_raises():
    caught = False
    try:
        repr(RecursiveRepr())
    except RecursionError:
        caught = True
    expect(caught).to_be(True)

class BadRepr:
    def __repr__(self):
        raise ValueError("bad repr")

class BadStr:
    def __str__(self):
        return 42

def test_repr_errors_propagate():
    msg = ""
    try:
        repr([BadRepr()])
    except ValueError as e:
        msg = str(e)
    expect(msg).to_be("bad repr")
    msg = ""
    try:
        f"{BadRepr()!r}"
    except ValueError as e:
        msg = str(e)
    expect(msg).to_be("bad repr")
    msg = ""
    try:
        str(BadStr())
    except TypeError as e:
        msg = str(e)
    expect(msg).to_be("__str__ returned non-string (type int)")

# === Numbers ===
def test_float_repr():
    expect(repr(float("inf"))).to_be("inf")
    expect(repr(float("-inf"))).to_be("-inf")
    expect(repr(float("nan"))).to_be("nan")
    expect(str([float("inf")])).to_be("[inf]")
    expect(str(1e15)).to_be("1000000000000000.0")
    expect(repr(1e16)).to_be("1e+16")
    expect(repr(1.5e-7)).to_be("1.5e-07")
    expect(repr(-0.0)).to_be("-0.0")
    expect(repr(1e22)).to_be("1e+22")
    expect(repr(0.0001)).to_be("0.0001")
    expect(repr(0.00001)).to_be("1e-05")
    expect(repr(0.1 + 0.2)).to_be("0.30000000000000004")
    expect(repr(123456789012345678.0)).to_be("1.2345678901234568e+17")
    expect(repr(5e-324)).to_be("5e-324")

def test_complex_repr():
    expect(repr(1 + 2j)).to_be("(1+2j)")
    expect(repr(1e20j)).to_be("1e+20j")
    expect(repr(complex(float("nan"), float("inf")))).to_be("(nan+infj)")
    expect(repr(complex(1, -0.0))).to_be("(1-0j)")
    expect(repr(complex(0, -0.0))).to_be("-0j")
    expect(repr(1e15j)).to_be("1000000000000000j")

def test_big_int_text():
    b = int("123456789012345678901234567890")
    expect(str(b)).to_be("123456789012345678901234567890")
    expect(repr([b])).to_be("[123456789012345678901234567890]")
    expect(f"{b:,}").to_be("123,456,789,012,345,678,901,234,567,890")
    expect("%d" % b).to_be("123456789012345678901234567890")

def test_range_repr():
    expect(repr(range(5))).to_be("range(0, 5)")
    expect(repr(range(1, 10, 2))).to_be("range(1, 10, 2)")

# === Exceptions ===
class CustomError(Exception):
    pass

def test_exception_repr_str():
    expect(repr(ValueError("x"))).to_be("ValueError('x')")
    expect(repr(ValueError())).to_be("ValueError()")
    expect(repr(ValueError("it's", 2))).to_be("ValueError(\"it's\", 2)")
    expect(str(ValueError("it's", 2))).to_be("(\"it's\", 2)")
    expect(str(KeyError("k"))).to_be("'k'")
    expect(repr(KeyError("k"))).to_be("KeyError('k')")
    expect(str(KeyError())).to_be("")
    expect(repr(CustomError("a", 1))).to_be("CustomError('a', 1)")
    expect(str([ValueError("v")])).to_be("[ValueError('v')]")

def test_missing_key_message():
    msg = ""
    try:
        {}["nope"]
    except KeyError as e:
        msg = str(e)
    expect(msg).to_be("'nope'")
    msg = ""
    try:
        {}[(1, 2)]
    except KeyError as e:
        msg = str(e)
    expect(msg).to_be("(1, 2)")
    msg = ""
    try:
        {}.pop("k")
    except KeyError as e:
        msg = repr(e)
    expect(msg).to_be("KeyError('k')")

def test_raised_exception_text():
    msg = ""
    try:
        raise ValueError("x", 2)
    except ValueError as e:
        msg = str(e) + "|" + repr(e)
    expect(msg).to_be("('x', 2)|ValueError('x', 2)")
    msg = "unset"
    try:
        raise KeyError
    except KeyError as e:
        msg = str(e)
    expect(msg).to_be("")

# === format() and the format spec mini-language ===
def test_format_grouping():
    expect(format(1234567, ",")).to_be("1,234,567")
    expect(format(1234567, "_")).to_be("1_234_567")
    expect(format(0xFFFFFFFF, "_x")).to_be("ffff_ffff")
    expect(format(1234567.891, ",.2f")).to_be("1,234,567.89")
    expect(format(1234, "08,")).to_be("0,001,234")
    expect(format(1234567, "015,")).to_be("000,001,234,567")
    expect(format(-1234.5, ",")).to_be("-1,234.5")

def test_format_int_types():
    expect(format(255, "#x")).to_be("0xff")
    expect(format(255, "#X")).to_be("0XFF")
    expect(format(5, "#b")).to_be("0b101")
    expect(format(8, "#o")).to_be("0o10")
    expect(format(-255, "#x")).to_be("-0xff")
    expect(format(255, "#010x")).to_be("0x000000ff")
    expect(format(65, "c")).to_be("A")
    expect(format(1234, "n")).to_be("1234")
    expect(format(3, "f")).to_be("3.000000")
    expect(format(3, "%")).to_be("300.000000%")
    expect(format(True, "")).to_be("True")
    expect(format(True, ">5")).to_be("    1")

def test_format_align_sign():
    expect(format(-42, "=10")).to_be("-       42")
    expect(format(-42, "0=10")).to_be("-000000042")
    expect(format(42, "+")).to_be("+42")
    expect(format(42, " ")).to_be(" 42")
    expect(format(42, "*^10")).to_be("****42****")
    expect(format("ab", "-^6")).to_be("--ab--")
    expect(format("ab", "05")).to_be("ab000")
    expect(format(123, "<010")).to_be("1230000000")
    expect(format(3.14159, "+08.2f")).to_be("+0003.14")

def test_format_float_types():
    expect(format(0.5, "%")).to_be("50.000000%")
    expect(format(0.125, ".1%")).to_be("12.5%")
    expect(format(1 / 3, ".3")).to_be("0.333")
    expect(format(100.0, ".3")).to_be("1e+02")
    expect(format(2.0, ".3")).to_be("2.0")
    expect(format(123456.0, ".3")).to_be("1.23e+05")
    expect(format(1e-5, "g")).to_be("1e-05")
    expect(format(123456789.0, "g")).to_be("1.23457e+08")
    expect(format(1.0, "#g")).to_be("1.00000")
    expect(format(1.5, "#.0f")).to_be("2.")
    expect(format(-0.0, "z.1f")).to_be("0.0")
    expect(format(float("inf"), "f")).to_be("inf")
    expect(format(float("inf"), "F")).to_be("INF")
    expect(format(float("nan"), "+")).to_be("+nan")
    expect(format(float("inf"), "010")).to_be("0000000inf")
    expect(format(2.675, ".2f")).to_be("2.67")
    expect(format(1234.5678912, ",.6_f")).to_be("1,234.567_891")

def test_format_complex():
    expect(format(1 + 2j, "")).to_be("(1+2j)")
    expect(format(1 + 2j, ".2f")).to_be("1.00+2.00j")
    expect(format(1.5j, ">10")).to_be("      1.5j")
    expect(format(-1 - 2j, "+.1f")).to_be("-1.0-2.0j")

def test_format_errors():
    msg = ""
    try:
        format(1, "s")
    except ValueError as e:
        msg = str(e)
    expect(msg).to_be("Unknown format code 's' for object of type 'int'")
    msg = ""
    try:
        format("a", "+")
    except ValueError as e:
        msg = str(e)
    expect(msg).to_be("Sign not allowed in string format specifier")
    msg = ""
    try:
        format(1, ".2d")
    except ValueError as e:
        msg = str(e)
    expect(msg).to_be("Precision not allowed in integer format specifier")
    msg = ""
    try:
        format(1, ",x")
    except ValueError as e:
        msg = str(e)
    expect(msg).to_be("Cannot specify ',' with 'x'.")
    msg = ""
    try:
        format([1], ">5")
    except TypeError as e:
        msg = str(e)
    expect(msg).to_be("unsupported format string passed to list.__format__")

class Formattable:
    def __format__(self, spec):
        return "F[" + spec + "]"

    def __str__(self):
        return "STR"

def test_format_dunder():
    expect(format(Formattable(), "abc")).to_be("F[abc]")
    expect(f"{Formattable()}").to_be("F[]")
    expect(f"{Formattable():xyz}").to_be("F[xyz]")
    expect("{}".format(Formattable())).to_be("F[]")
    expect(str(Formattable())).to_be("STR")

# === str.format() ===
class Point:
    def __init__(self, x):
        self.x = x

def test_str_format_fields():
    expect("{0.real} {0.imag}".format(3)).to_be("3 0")
    expect("{0[1]} {1[k]}".format([1, 2], {"k": "v"})).to_be("2 v")
    expect("{0.x}".format(Point(9))).to_be("9")
    expect("{a[0]}{a[1]}".format(a="xy")).to_be("xy")
    expect("{00}".format("z")).to_be("z")
    expect("{{}} {{{}}}".format(5)).to_be("{} {5}")

def test_str_format_nested_spec():
    expect("{:>{w}}|".format("a", w=4)).to_be("   a|")
    expect("{:.{p}f}".format(3.14159, p=2)).to_be("3.14")
    expect("{:{fill}{align}{width}}".format("x", fill="*", align="^", width=7)).to_be("***x***")
    expect("{0:{1}}|{0:>{1}}|".format("z", 4)).to_be("z   |   z|")
    expect("{:.{}f}".format(2.71828, 3)).to_be("2.718")

def test_str_format_conversions():
    expect("{!r} {!s} {!a}".format("x", "y", "\u00e9")).to_be("'x' y '\\xe9'")
    expect("{0!r:>6}".format("x")).to_be("   'x'")

def format_error(template):
    try:
        return template.format(1, a=2)
    except ValueError as e:
        return "ValueError: " + str(e)
    except IndexError as e:
        return "IndexError: " + str(e)
    except KeyError as e:
        return "KeyError: " + str(e)

def test_str_format_errors():
    expect(format_error("{")).to_be("ValueError: Single '{' encountered in format string")
    expect(format_error("}")).to_be("ValueError: Single '}' encountered in format string")
    expect(format_error("{0")).to_be("ValueError: expected '}' before end of string")
    expect(format_error("{a!x}")).to_be("ValueError: Unknown conversion specifier x")
    expect(format_error("{0.}")).to_be("ValueError: Empty attribute in format string")
    expect(format_error("{a:{a:{a}}}")).to_be("ValueError: Max string recursion exceeded")
    expect(format_error("{}{0}")).to_be("ValueError: cannot switch from automatic field numbering to manual field specification")
    expect(format_error("{5}")).to_be("IndexError: Replacement index 5 out of range for positional args tuple")
    expect(format_error("{z}")).to_be("KeyError: 'z'")

def test_format_map():
    expect("{x} {y}".format_map({"x": 1, "y": "two"})).to_be("1 two")
    expect("{a[k]}".format_map({"a": {"k": 3}})).to_be("3")
    msg = ""
    try:
        "{0}".format_map({})
    except ValueError as e:
        msg = str(e)
    expect(msg).to_be("Format string contains positional fields")

# === f-strings ===
def test_fstring_nested_spec():
    expect(f"{10:{'>'}{5}}").to_be("   10")
    w = 6
    p = 2
    expect(f"{3.14159:{w}.{p}f}|").to_be("  3.14|")
    expect(f"{'x':{'*'}^{w}}|").to_be("**x***|")

def make_padder():
    width = 8
    def pad():
        return f"{'a':>{width}}|"
    return pad

def test_fstring_nested_spec_closure():
    expect(make_padder()()).to_be("       a|")

def test_fstring_conversions():
    e_acute = "\u00e9"
    expect(f"{e_acute!a}").to_be("'\\xe9'")
    expect(f"{'x'!r:>5}").to_be("  'x'")
    expect(f"{1/3:.3}").to_be("0.333")
    expect(f"{[1.0, 'a']}").to_be("[1.0, 'a']")

# === printf-style % formatting ===
def test_percent_format():
    expect("%r" % "it's").to_be('"it\'s"')
    expect("%a" % "\u00e9").to_be("'\\xe9'")
    expect("%s" % {"k": 1}).to_be("{'k': 1}")
    expect("%d %i" % (3.9, True)).to_be("3 1")
    expect("%#x %#o %#X" % (0, 0, 255)).to_be("0x0 0o0 0XFF")
    expect("%.3d|%05.1f|%-5d|" % (7, -2.5, 42)).to_be("007|-02.5|42   |")
    expect("%f %E %g" % (float("inf"), float("nan"), 1e20)).to_be("inf NAN 1e+20")
    expect("%#g %e" % (1.0, 12345.678)).to_be("1.00000 1.234568e+04")
    expect("%c%c" % (72, "i")).to_be("Hi")
    expect("%.1s|%5.1s|" % ("\u00e9!", "xyz")).to_be("\u00e9|    x|")

def test_percent_format_errors():
    msg = ""
    try:
        "%x" % 1.5
    except TypeError as e:
        msg = str(e)
    expect(msg).to_be("%x format: an integer is required, not float")
    msg = ""
    try:
        "%d" % "s"
    except TypeError as e:
        msg = str(e)
    expect(msg).to_be("%d format: a real number is required, not str")
    msg = ""
    try:
        "%y" % 1
    except ValueError as e:
        msg = str(e)
    expect(msg).to_be("unsupported format character 'y' (0x79) at index 1")

# === Text dunders on builtins and object ===
class Plain:
    pass

class SuperRepr:
    def __repr__(self):
        return "S+" + super().__repr__()

def test_text_dunders():
    expect("it's".__repr__()).to_be('"it\'s"')
    expect((1.5).__repr__()).to_be("1.5")
    expect([1, "a"].__repr__()).to_be("[1, 'a']")
    expect((5).__format__("03d")).to_be("005")
    p = Plain()
    expect(object.__repr__(p) == repr(p)).to_be(True)
    expect(p.__str__() == p.__repr__()).to_be(True)
    expect("Plain object" in object.__repr__(p)).to_be(True)
    expect(repr(SuperRepr()).startswith("S+<")).to_be(True)

# === Enum text ===
from enum import Enum, IntEnum

class Color(Enum):
    RED = 1

class Num(IntEnum):
    ONE = 1

def test_enum_format():
    expect(f"{Color.RED:>12}|").to_be("   Color.RED|")
    expect(format(Color.RED)).to_be("Color.RED")
    expect(str(Num.ONE)).to_be("1")
    expect(f"{Num.ONE}").to_be("1")
    expect(f"{Num.ONE:03d}").to_be("001")
    expect(repr(Num.ONE)).to_be("<Num.ONE: 1>")

test("str_repr_quotes", test_str_repr_quotes)
test("str_repr_escapes", test_str_repr_escapes)
test("str_repr_in_containers", test_str_repr_in_containers)
test("bytes_repr", test_bytes_repr)
test("ascii", test_ascii)
test("self_referential_containers", test_self_referential_containers)
test("recursive_repr_raises", test_recursive_repr_raises)
test("repr_errors_propagate", test_repr_errors_propagate)
test("float_repr", test_float_repr)
test("complex_repr", test_complex_repr)
test("big_int_text", test_big_int_text)
test("range_repr", test_range_repr)
test("exception_repr_str", test_exception_repr_str)
test("missing_key_message", test_missing_key_message)
test("raised_exception_text", test_raised_exception_text)
test("format_grouping", test_format_grouping)
test("format_int_types", test_format_int_types)
test("format_align_sign", test_format_align_sign)
test("format_float_types", test_format_float_types)
test("format_complex", test_format_complex)
test("format_errors", test_format_errors)
test("format_dunder", test_format_dunder)
test("str_format_fields", test_str_format_fields)
test("str_format_nested_spec", test_str_format_nested_spec)
test("str_format_conversions", test_str_format_conversions)
test("str_format_errors", test_str_format_errors)
test("format_map", test_format_map)
test("fstring_nested_spec", test_fstring_nested_spec)
test("fstring_nested_spec_closure", test_fstring_nested_spec_closure)
test("fstring_conversions", test_fstring_conversions)
test("percent_format", test_percent_format)
test("percent_format_errors", test_percent_format_errors)
test("text_dunders", test_text_dunders)
test("enum_format", test_enum_format)

print("Text conversion tests completed")
