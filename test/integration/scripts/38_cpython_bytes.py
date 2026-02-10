# Test: CPython Bytes Edge Cases
# Adapted from CPython's test_bytes.py

from test_framework import test, expect

def test_bytes_constructor():
    expect(bytes()).to_be(b"")
    expect(bytes(3)).to_be(b"\x00\x00\x00")
    expect(bytes([65, 66, 67])).to_be(b"ABC")
    expect(bytes(b"hello")).to_be(b"hello")

def test_bytes_from_string():
    expect(bytes("hello", "utf-8")).to_be(b"hello")
    expect(bytes("abc", "ascii")).to_be(b"abc")

def test_bytes_len():
    expect(len(b"")).to_be(0)
    expect(len(b"hello")).to_be(5)
    expect(len(bytes(10))).to_be(10)

def test_bytes_indexing():
    b = b"hello"
    expect(b[0]).to_be(104)  # ord('h')
    expect(b[1]).to_be(101)  # ord('e')
    expect(b[-1]).to_be(111) # ord('o')

def test_bytes_slicing():
    b = b"hello world"
    expect(b[0:5]).to_be(b"hello")
    expect(b[6:]).to_be(b"world")
    expect(b[:5]).to_be(b"hello")
    expect(b[::2]).to_be(b"hlowrd")

def test_bytes_contains():
    b = b"hello"
    expect(104 in b).to_be(True)   # ord('h')
    expect(120 in b).to_be(False)  # ord('x')

def test_bytes_concat():
    expect(b"hello" + b" world").to_be(b"hello world")
    expect(b"ab" + b"cd").to_be(b"abcd")

def test_bytes_repeat():
    expect(b"ab" * 3).to_be(b"ababab")
    expect(3 * b"x").to_be(b"xxx")
    expect(b"ab" * 0).to_be(b"")

def test_bytes_comparison():
    expect(b"abc" == b"abc").to_be(True)
    expect(b"abc" != b"def").to_be(True)
    expect(b"abc" < b"abd").to_be(True)
    expect(b"abc" > b"abb").to_be(True)
    expect(b"abc" <= b"abc").to_be(True)

def test_bytes_methods_find():
    b = b"hello world"
    expect(b.find(b"world")).to_be(6)
    expect(b.find(b"xyz")).to_be(-1)
    expect(b.find(b"l")).to_be(2)
    expect(b.find(b"l", 3)).to_be(3)

def test_bytes_methods_count():
    b = b"hello world"
    expect(b.count(b"l")).to_be(3)
    expect(b.count(b"o")).to_be(2)
    expect(b.count(b"xyz")).to_be(0)

def test_bytes_methods_replace():
    expect(b"hello".replace(b"l", b"r")).to_be(b"herro")
    expect(b"aaa".replace(b"a", b"b", 2)).to_be(b"bba")

def test_bytes_methods_split():
    expect(b"a b c".split()).to_be([b"a", b"b", b"c"])
    expect(b"a,b,c".split(b",")).to_be([b"a", b"b", b"c"])
    expect(b"a,,b".split(b",")).to_be([b"a", b"", b"b"])

def test_bytes_methods_join():
    expect(b",".join([b"a", b"b", b"c"])).to_be(b"a,b,c")
    expect(b" ".join([b"hello", b"world"])).to_be(b"hello world")
    expect(b"".join([b"a", b"b"])).to_be(b"ab")

def test_bytes_methods_strip():
    expect(b"  hello  ".strip()).to_be(b"hello")
    expect(b"  hello  ".lstrip()).to_be(b"hello  ")
    expect(b"  hello  ".rstrip()).to_be(b"  hello")

def test_bytes_methods_upper_lower():
    expect(b"hello".upper()).to_be(b"HELLO")
    expect(b"HELLO".lower()).to_be(b"hello")
    expect(b"Hello World".upper()).to_be(b"HELLO WORLD")

def test_bytes_methods_startswith_endswith():
    b = b"hello world"
    expect(b.startswith(b"hello")).to_be(True)
    expect(b.startswith(b"world")).to_be(False)
    expect(b.endswith(b"world")).to_be(True)
    expect(b.endswith(b"hello")).to_be(False)

def test_bytes_hex():
    expect(b"ABC".hex()).to_be("414243")
    expect(b"\x00\xff".hex()).to_be("00ff")
    expect(b"".hex()).to_be("")

def test_bytes_decode():
    expect(b"hello".decode()).to_be("hello")
    expect(b"hello".decode("utf-8")).to_be("hello")
    expect(b"hello".decode("ascii")).to_be("hello")

def test_bytes_bool():
    expect(bool(b"")).to_be(False)
    expect(bool(b"x")).to_be(True)

def test_bytes_iteration():
    result = []
    for byte in b"ABC":
        result.append(byte)
    expect(result).to_be([65, 66, 67])

def test_bytes_immutability():
    try:
        b = b"hello"
        b[0] = 72
        expect("no error").to_be("TypeError")
    except TypeError:
        expect(True).to_be(True)

def test_bytes_repr():
    expect(repr(b"hello")).to_be("b'hello'")

# Register all tests
test("bytes_constructor", test_bytes_constructor)
test("bytes_from_string", test_bytes_from_string)
test("bytes_len", test_bytes_len)
test("bytes_indexing", test_bytes_indexing)
test("bytes_slicing", test_bytes_slicing)
test("bytes_contains", test_bytes_contains)
test("bytes_concat", test_bytes_concat)
test("bytes_repeat", test_bytes_repeat)
test("bytes_comparison", test_bytes_comparison)
test("bytes_methods_find", test_bytes_methods_find)
test("bytes_methods_count", test_bytes_methods_count)
test("bytes_methods_replace", test_bytes_methods_replace)
test("bytes_methods_split", test_bytes_methods_split)
test("bytes_methods_join", test_bytes_methods_join)
test("bytes_methods_strip", test_bytes_methods_strip)
test("bytes_methods_upper_lower", test_bytes_methods_upper_lower)
test("bytes_methods_startswith_endswith", test_bytes_methods_startswith_endswith)
test("bytes_hex", test_bytes_hex)
test("bytes_decode", test_bytes_decode)
test("bytes_bool", test_bytes_bool)
test("bytes_iteration", test_bytes_iteration)
test("bytes_immutability", test_bytes_immutability)
test("bytes_repr", test_bytes_repr)

print("CPython bytes tests completed")
