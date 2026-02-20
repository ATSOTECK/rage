from test_framework import test, expect

# === casefold ===
test("casefold basic", lambda: expect("Hello".casefold()).to_be("hello"))
test("casefold already lower", lambda: expect("hello".casefold()).to_be("hello"))
test("casefold empty", lambda: expect("".casefold()).to_be(""))
test("casefold mixed", lambda: expect("ABC123def".casefold()).to_be("abc123def"))
test("casefold german ss", lambda: expect("Straße".casefold()).to_be("strasse"))

# === isascii ===
test("isascii all ascii", lambda: expect("hello".isascii()).to_be(True))
test("isascii empty", lambda: expect("".isascii()).to_be(True))
test("isascii with digits", lambda: expect("abc123".isascii()).to_be(True))
test("isascii non-ascii", lambda: expect("héllo".isascii()).to_be(False))
test("isascii space", lambda: expect(" ".isascii()).to_be(True))
test("isascii control char", lambda: expect("\n\t".isascii()).to_be(True))

# === isdecimal ===
test("isdecimal digits", lambda: expect("12345".isdecimal()).to_be(True))
test("isdecimal empty", lambda: expect("".isdecimal()).to_be(False))
test("isdecimal with letter", lambda: expect("123a".isdecimal()).to_be(False))
test("isdecimal single", lambda: expect("0".isdecimal()).to_be(True))
test("isdecimal space", lambda: expect("1 2".isdecimal()).to_be(False))

# === isnumeric ===
test("isnumeric digits", lambda: expect("12345".isnumeric()).to_be(True))
test("isnumeric empty", lambda: expect("".isnumeric()).to_be(False))
test("isnumeric with letter", lambda: expect("123a".isnumeric()).to_be(False))

# === isidentifier ===
test("isidentifier valid", lambda: expect("hello".isidentifier()).to_be(True))
test("isidentifier underscore", lambda: expect("_foo".isidentifier()).to_be(True))
test("isidentifier with digits", lambda: expect("foo123".isidentifier()).to_be(True))
test("isidentifier starts with digit", lambda: expect("123foo".isidentifier()).to_be(False))
test("isidentifier empty", lambda: expect("".isidentifier()).to_be(False))
test("isidentifier keyword", lambda: expect("class".isidentifier()).to_be(True))  # True in CPython
test("isidentifier with space", lambda: expect("foo bar".isidentifier()).to_be(False))
test("isidentifier with hyphen", lambda: expect("foo-bar".isidentifier()).to_be(False))

# === isprintable ===
test("isprintable normal", lambda: expect("hello".isprintable()).to_be(True))
test("isprintable empty", lambda: expect("".isprintable()).to_be(True))
test("isprintable with space", lambda: expect("hello world".isprintable()).to_be(True))
test("isprintable with newline", lambda: expect("hello\n".isprintable()).to_be(False))
test("isprintable with tab", lambda: expect("hello\t".isprintable()).to_be(False))
test("isprintable with null", lambda: expect("hello\x00".isprintable()).to_be(False))

# === istitle ===
test("istitle title case", lambda: expect("Hello World".istitle()).to_be(True))
test("istitle all upper", lambda: expect("HELLO".istitle()).to_be(False))
test("istitle all lower", lambda: expect("hello".istitle()).to_be(False))
test("istitle empty", lambda: expect("".istitle()).to_be(False))
test("istitle single word", lambda: expect("Hello".istitle()).to_be(True))
test("istitle with number", lambda: expect("Hello 123 World".istitle()).to_be(True))
test("istitle mixed wrong", lambda: expect("Hello world".istitle()).to_be(False))
test("istitle apostrophe", lambda: expect("They'Re".istitle()).to_be(True))

# === removeprefix ===
test("removeprefix match", lambda: expect("HelloWorld".removeprefix("Hello")).to_be("World"))
test("removeprefix no match", lambda: expect("HelloWorld".removeprefix("World")).to_be("HelloWorld"))
test("removeprefix empty prefix", lambda: expect("Hello".removeprefix("")).to_be("Hello"))
test("removeprefix empty string", lambda: expect("".removeprefix("Hello")).to_be(""))
test("removeprefix full match", lambda: expect("Hello".removeprefix("Hello")).to_be(""))

# === removesuffix ===
test("removesuffix match", lambda: expect("HelloWorld".removesuffix("World")).to_be("Hello"))
test("removesuffix no match", lambda: expect("HelloWorld".removesuffix("Hello")).to_be("HelloWorld"))
test("removesuffix empty suffix", lambda: expect("Hello".removesuffix("")).to_be("Hello"))
test("removesuffix empty string", lambda: expect("".removesuffix("Hello")).to_be(""))
test("removesuffix full match", lambda: expect("Hello".removesuffix("Hello")).to_be(""))

# === format_map ===
test("format_map basic", lambda: expect("{name} is {age}".format_map({"name": "Alice", "age": 30})).to_be("Alice is 30"))
test("format_map empty", lambda: expect("hello".format_map({})).to_be("hello"))
test("format_map single", lambda: expect("{x}".format_map({"x": 42})).to_be("42"))
test("format_map escaped braces", lambda: expect("{{literal}}".format_map({})).to_be("{literal}"))

# format_map with format spec
test("format_map with spec", lambda: expect("{x:.2f}".format_map({"x": 3.14159})).to_be("3.14"))

# format_map key error
def test_format_map_keyerror():
    try:
        "{missing}".format_map({})
        return False
    except KeyError:
        return True

test("format_map missing key raises KeyError", lambda: expect(test_format_map_keyerror()).to_be(True))

# === maketrans and translate ===
# 2-arg form
test("maketrans 2-arg", lambda: expect("hello".translate("hello".maketrans("helo", "HELO"))).to_be("HELLO"))

# 3-arg form (delete chars)
test("maketrans 3-arg delete", lambda: expect("hello world".translate("hello world".maketrans("", "", " "))).to_be("helloworld"))

# 2-arg translate mapping
test("translate basic", lambda: expect("abc".translate("abc".maketrans("abc", "xyz"))).to_be("xyz"))

# 3-arg with replacement and deletion
test("maketrans replace and delete", lambda: expect("hello world!".translate("".maketrans("o", "0", "!"))).to_be("hell0 w0rld"))

# translate with no matches
test("translate no match", lambda: expect("hello".translate("hello".maketrans("xyz", "XYZ"))).to_be("hello"))

# 1-arg dict form
test("maketrans 1-arg dict", lambda: expect("abc".translate(str.maketrans({ord("a"): "X", ord("b"): None}))).to_be("Xc"))

# str.maketrans class method call
test("str.maketrans class call", lambda: expect("abc".translate(str.maketrans("a", "A"))).to_be("Abc"))

# === Error cases ===
def test_removeprefix_type_error():
    try:
        "hello".removeprefix(123)
        return False
    except TypeError:
        return True

test("removeprefix type error", lambda: expect(test_removeprefix_type_error()).to_be(True))

def test_removesuffix_type_error():
    try:
        "hello".removesuffix(123)
        return False
    except TypeError:
        return True

test("removesuffix type error", lambda: expect(test_removesuffix_type_error()).to_be(True))

def test_maketrans_unequal():
    try:
        str.maketrans("ab", "xyz")
        return False
    except ValueError:
        return True

test("maketrans unequal lengths", lambda: expect(test_maketrans_unequal()).to_be(True))
