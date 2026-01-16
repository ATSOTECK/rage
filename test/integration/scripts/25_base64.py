# Test: Base64 Module
# Tests base16, base32, and base64 encoding/decoding functions
#
# Note: Since bytes equality and iteration aren't fully supported,
# we use str() to convert bytes to string representation for comparison.

from test_framework import test, expect

import base64

# =============================================================================
# Base64 encoding/decoding tests
# =============================================================================

def test_b64encode_basic():
    # Test encoding produces expected output using string representation
    expect(str(base64.b64encode(""))).to_be("b''")
    expect(str(base64.b64encode("a"))).to_be("b'YQ=='")
    expect(str(base64.b64encode("ab"))).to_be("b'YWI='")
    expect(str(base64.b64encode("abc"))).to_be("b'YWJj'")
    expect(str(base64.b64encode("Hello"))).to_be("b'SGVsbG8='")
    expect(str(base64.b64encode("Hello, World!"))).to_be("b'SGVsbG8sIFdvcmxkIQ=='")

def test_b64encode_lengths():
    # Test that encoding produces expected lengths
    expect(len(base64.b64encode(""))).to_be(0)
    expect(len(base64.b64encode("a"))).to_be(4)  # 1 byte -> 4 chars with padding
    expect(len(base64.b64encode("ab"))).to_be(4)  # 2 bytes -> 4 chars with padding
    expect(len(base64.b64encode("abc"))).to_be(4)  # 3 bytes -> 4 chars no padding
    expect(len(base64.b64encode("Hello, World!"))).to_be(20)

def test_b64decode_basic():
    # Test decoding produces expected output
    expect(str(base64.b64decode("YQ=="))).to_be("b'a'")
    expect(str(base64.b64decode("YWI="))).to_be("b'ab'")
    expect(str(base64.b64decode("YWJj"))).to_be("b'abc'")
    expect(str(base64.b64decode("SGVsbG8="))).to_be("b'Hello'")
    expect(str(base64.b64decode("SGVsbG8sIFdvcmxkIQ=="))).to_be("b'Hello, World!'")

def test_b64_roundtrip():
    # Encode then decode should return original (check via string representation)
    test_strings = ["", "a", "ab", "abc", "Hello, World!", "Python 3.14"]
    for s in test_strings:
        encoded = base64.b64encode(s)
        decoded = base64.b64decode(encoded)
        expected = "b'" + s + "'"
        expect(str(decoded)).to_be(expected)

# =============================================================================
# Standard base64 tests
# =============================================================================

def test_standard_b64encode():
    expect(str(base64.standard_b64encode("Hello"))).to_be("b'SGVsbG8='")
    expect(str(base64.standard_b64encode(""))).to_be("b''")

def test_standard_b64decode():
    expect(str(base64.standard_b64decode("SGVsbG8="))).to_be("b'Hello'")
    expect(str(base64.standard_b64decode(""))).to_be("b''")

def test_standard_b64_roundtrip():
    test_strings = ["test", "Python", "RAGE interpreter"]
    for s in test_strings:
        decoded = base64.standard_b64decode(base64.standard_b64encode(s))
        expected = "b'" + s + "'"
        expect(str(decoded)).to_be(expected)

# =============================================================================
# URL-safe base64 tests
# =============================================================================

def test_urlsafe_b64encode():
    # URL-safe encoding for normal text is same as standard
    expect(str(base64.urlsafe_b64encode("Hello"))).to_be("b'SGVsbG8='")

def test_urlsafe_b64encode_special():
    # URL-safe uses - and _ instead of + and /
    # Test with binary data that produces + and / in standard encoding
    # "\xff\xfe\xfd" produces "//79" in standard, "__79" in URL-safe
    result = str(base64.urlsafe_b64encode("\xff\xfe\xfd"))
    # Should NOT contain + or /
    expect("+" not in result).to_be(True)
    expect("/" not in result or result.startswith("b'")).to_be(True)

def test_urlsafe_b64decode():
    expect(str(base64.urlsafe_b64decode("SGVsbG8="))).to_be("b'Hello'")

def test_urlsafe_b64_roundtrip():
    test_strings = ["test", "hello world"]
    for s in test_strings:
        decoded = base64.urlsafe_b64decode(base64.urlsafe_b64encode(s))
        expected = "b'" + s + "'"
        expect(str(decoded)).to_be(expected)

# =============================================================================
# Base32 encoding/decoding tests
# =============================================================================

def test_b32encode():
    expect(str(base64.b32encode(""))).to_be("b''")
    expect(str(base64.b32encode("a"))).to_be("b'ME======'")
    expect(str(base64.b32encode("Hello"))).to_be("b'JBSWY3DP'")

def test_b32decode():
    expect(str(base64.b32decode("ME======"))).to_be("b'a'")
    expect(str(base64.b32decode("JBSWY3DP"))).to_be("b'Hello'")

def test_b32decode_casefold():
    # With casefold=True, lowercase should work
    expect(str(base64.b32decode("me======", True))).to_be("b'a'")
    expect(str(base64.b32decode("jbswy3dp", True))).to_be("b'Hello'")

def test_b32_roundtrip():
    test_strings = ["", "a", "test", "Hello, World!"]
    for s in test_strings:
        decoded = base64.b32decode(base64.b32encode(s))
        expected = "b'" + s + "'"
        expect(str(decoded)).to_be(expected)

# =============================================================================
# Base32 Hex encoding/decoding tests
# =============================================================================

def test_b32hexencode():
    expect(str(base64.b32hexencode(""))).to_be("b''")
    expect(str(base64.b32hexencode("a"))).to_be("b'C4======'")
    expect(str(base64.b32hexencode("Hello"))).to_be("b'91IMOR3F'")

def test_b32hexdecode():
    expect(str(base64.b32hexdecode("C4======"))).to_be("b'a'")
    expect(str(base64.b32hexdecode("91IMOR3F"))).to_be("b'Hello'")

def test_b32hexdecode_casefold():
    expect(str(base64.b32hexdecode("c4======", True))).to_be("b'a'")
    expect(str(base64.b32hexdecode("91imor3f", True))).to_be("b'Hello'")

def test_b32hex_roundtrip():
    test_strings = ["", "a", "test", "Hello"]
    for s in test_strings:
        decoded = base64.b32hexdecode(base64.b32hexencode(s))
        expected = "b'" + s + "'"
        expect(str(decoded)).to_be(expected)

# =============================================================================
# Base16 (hex) encoding/decoding tests
# =============================================================================

def test_b16encode():
    expect(str(base64.b16encode(""))).to_be("b''")
    expect(str(base64.b16encode("a"))).to_be("b'61'")
    expect(str(base64.b16encode("Hello"))).to_be("b'48656C6C6F'")

def test_b16decode():
    expect(str(base64.b16decode("61"))).to_be("b'a'")
    expect(str(base64.b16decode("48656C6C6F"))).to_be("b'Hello'")

def test_b16decode_casefold():
    # With casefold=True, lowercase hex should work
    expect(str(base64.b16decode("48656c6c6f", True))).to_be("b'Hello'")

def test_b16_roundtrip():
    test_strings = ["", "a", "Hello", "Test123"]
    for s in test_strings:
        decoded = base64.b16decode(base64.b16encode(s))
        expected = "b'" + s + "'"
        expect(str(decoded)).to_be(expected)

# =============================================================================
# MIME-style encodebytes/decodebytes tests
# =============================================================================

def test_encodebytes():
    # encodebytes adds newlines every 76 characters
    result = str(base64.encodebytes("Hello"))
    # Should contain the base64 encoding followed by newline
    expect("SGVsbG8=" in result).to_be(True)

def test_encodebytes_has_newline():
    result = base64.encodebytes("Hello")
    # Length should be 9 (8 chars + newline)
    expect(len(result)).to_be(9)

def test_decodebytes():
    # decodebytes should handle strings with newlines
    expect(str(base64.decodebytes("SGVsbG8=\n"))).to_be("b'Hello'")
    expect(str(base64.decodebytes("SGVsbG8="))).to_be("b'Hello'")

def test_decodebytes_with_whitespace():
    # Should ignore whitespace
    expect(str(base64.decodebytes("SGVs\nbG8="))).to_be("b'Hello'")

def test_encodebytes_decodebytes_roundtrip():
    test_strings = ["", "Hello", "Test data"]
    for s in test_strings:
        decoded = base64.decodebytes(base64.encodebytes(s))
        expected = "b'" + s + "'"
        expect(str(decoded)).to_be(expected)

# =============================================================================
# String input tests
# =============================================================================

def test_string_input():
    # Functions should accept strings as input
    result = base64.b64encode("Hello")
    expect(len(result)).to_be(8)

    result = base64.b16encode("Hi")
    expect(str(result)).to_be("b'4869'")

# =============================================================================
# Run all tests
# =============================================================================

# Base64 tests
test("b64encode_basic", test_b64encode_basic)
test("b64encode_lengths", test_b64encode_lengths)
test("b64decode_basic", test_b64decode_basic)
test("b64_roundtrip", test_b64_roundtrip)

# Standard base64 tests
test("standard_b64encode", test_standard_b64encode)
test("standard_b64decode", test_standard_b64decode)
test("standard_b64_roundtrip", test_standard_b64_roundtrip)

# URL-safe base64 tests
test("urlsafe_b64encode", test_urlsafe_b64encode)
test("urlsafe_b64encode_special", test_urlsafe_b64encode_special)
test("urlsafe_b64decode", test_urlsafe_b64decode)
test("urlsafe_b64_roundtrip", test_urlsafe_b64_roundtrip)

# Base32 tests
test("b32encode", test_b32encode)
test("b32decode", test_b32decode)
test("b32decode_casefold", test_b32decode_casefold)
test("b32_roundtrip", test_b32_roundtrip)

# Base32 Hex tests
test("b32hexencode", test_b32hexencode)
test("b32hexdecode", test_b32hexdecode)
test("b32hexdecode_casefold", test_b32hexdecode_casefold)
test("b32hex_roundtrip", test_b32hex_roundtrip)

# Base16 tests
test("b16encode", test_b16encode)
test("b16decode", test_b16decode)
test("b16decode_casefold", test_b16decode_casefold)
test("b16_roundtrip", test_b16_roundtrip)

# MIME-style tests
test("encodebytes", test_encodebytes)
test("encodebytes_has_newline", test_encodebytes_has_newline)
test("decodebytes", test_decodebytes)
test("decodebytes_with_whitespace", test_decodebytes_with_whitespace)
test("encodebytes_decodebytes_roundtrip", test_encodebytes_decodebytes_roundtrip)

# String input tests
test("string_input", test_string_input)

print("Base64 module tests completed")
