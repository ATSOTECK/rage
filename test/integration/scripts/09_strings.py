# Test: String Operations
# Tests string methods and operations

from test_framework import test, expect

def test_str_basic():
    s = "Hello, World!"
    expect(len(s)).to_be(13)
    expect(s[0]).to_be("H")
    expect(s[-1]).to_be("!")

def test_str_concat():
    expect("Hello" + " " + "World").to_be("Hello World")
    expect("ab" * 4).to_be("abababab")

def test_str_case():
    expect("hello".upper()).to_be("HELLO")
    expect("HELLO".lower()).to_be("hello")

def test_str_strip():
    expect("  hello  ".strip()).to_be("hello")

def test_str_split_join():
    expect("hello world python".split()).to_be(["hello", "world", "python"])
    expect("a,b,c,d".split(",")).to_be(["a", "b", "c", "d"])
    expect(",".join(["a", "b", "c"])).to_be("a,b,c")

def test_str_replace():
    expect("hello hello hello".replace("hello", "hi")).to_be("hi hi hi")

def test_str_membership():
    expect("ell" in "hello").to_be(True)
    expect("xyz" not in "hello").to_be(True)

def test_str_comparison():
    expect("hello" == "hello").to_be(True)
    expect("hello" != "world").to_be(True)
    expect("apple" < "banana").to_be(True)
    expect("banana" > "apple").to_be(True)

def test_str_word_count():
    text = "The quick brown fox"
    words = text.split()
    expect(len(words)).to_be(4)

def test_str_empty():
    expect(len("")).to_be(0)
    expect(bool("")).to_be(False)

def test_str_mult_edge():
    expect("hello" * 0).to_be("")
    expect("hello" * 1).to_be("hello")

def test_str_case_insensitive():
    s1 = "Hello"
    s2 = "hello"
    expect(s1.lower() == s2.lower()).to_be(True)

test("str_basic", test_str_basic)
test("str_concat", test_str_concat)
test("str_case", test_str_case)
test("str_strip", test_str_strip)
test("str_split_join", test_str_split_join)
test("str_replace", test_str_replace)
test("str_membership", test_str_membership)
test("str_comparison", test_str_comparison)
test("str_word_count", test_str_word_count)
test("str_empty", test_str_empty)
test("str_mult_edge", test_str_mult_edge)
test("str_case_insensitive", test_str_case_insensitive)

print("Strings tests completed")
