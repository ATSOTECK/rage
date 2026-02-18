from test_framework import test, expect

# Test 1: ascii with plain ASCII string
def test_ascii_plain():
    expect(ascii("hello")).to_be("'hello'")
test("ascii plain string", test_ascii_plain)

# Test 2: ascii with non-ASCII characters
def test_ascii_non_ascii():
    result = ascii("café")
    expect(result).to_be("'caf\\xe9'")
test("ascii non-ASCII characters", test_ascii_non_ascii)

# Test 3: ascii with accented character (Latin-1 range)
def test_ascii_latin1():
    result = ascii("é")
    expect(result).to_be("'\\xe9'")
test("ascii Latin-1 character", test_ascii_latin1)

# Test 4: ascii with CJK characters
def test_ascii_cjk():
    result = ascii("日")
    expect(result).to_be("'\\u65e5'")
test("ascii CJK character", test_ascii_cjk)

# Test 5: ascii with integer
def test_ascii_int():
    expect(ascii(42)).to_be("42")
test("ascii integer", test_ascii_int)

# Test 6: ascii with None
def test_ascii_none():
    expect(ascii(None)).to_be("None")
test("ascii None", test_ascii_none)

# Test 7: ascii with bool
def test_ascii_bool():
    expect(ascii(True)).to_be("True")
    expect(ascii(False)).to_be("False")
test("ascii bool", test_ascii_bool)

# Test 8: ascii with list containing non-ASCII strings
def test_ascii_list():
    result = ascii(["café", "naïve"])
    expect(result).to_be("['caf\\xe9', 'na\\xefve']")
test("ascii list with non-ASCII", test_ascii_list)

# Test 9: ascii with tuple
def test_ascii_tuple():
    result = ascii(("café",))
    expect(result).to_be("('caf\\xe9',)")
test("ascii tuple with non-ASCII", test_ascii_tuple)

# Test 10: ascii with dict
def test_ascii_dict():
    result = ascii({"café": "naïve"})
    expect(result).to_be("{'caf\\xe9': 'na\\xefve'}")
test("ascii dict with non-ASCII", test_ascii_dict)

# Test 11: ascii with special characters
def test_ascii_special():
    result = ascii("line1\nline2")
    expect(result).to_be("'line1\\nline2'")
test("ascii special characters", test_ascii_special)

# Test 12: ascii with tab
def test_ascii_tab():
    result = ascii("a\tb")
    expect(result).to_be("'a\\tb'")
test("ascii tab", test_ascii_tab)

# Test 13: ascii with backslash
def test_ascii_backslash():
    result = ascii("a\\b")
    expect(result).to_be("'a\\\\b'")
test("ascii backslash", test_ascii_backslash)

# Test 14: ascii with single quote
def test_ascii_quote():
    result = ascii("it's")
    expect(result).to_be("'it\\'s'")
test("ascii single quote", test_ascii_quote)

# Test 15: ascii with empty string
def test_ascii_empty():
    expect(ascii("")).to_be("''")
test("ascii empty string", test_ascii_empty)

# Test 16: ascii with float
def test_ascii_float():
    expect(ascii(3.14)).to_be("3.14")
test("ascii float", test_ascii_float)

# Test 17: %a format specifier
def test_format_a():
    result = "%a" % "café"
    expect(result).to_be("'caf\\xe9'")
test("%a format specifier", test_format_a)
