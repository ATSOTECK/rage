# Test: String Operations
# Tests string methods and operations

def test_str_basic():
    s = "Hello, World!"
    expect(13, len(s))
    expect("H", s[0])
    expect("!", s[-1])

def test_str_concat():
    expect("Hello World", "Hello" + " " + "World")
    expect("abababab", "ab" * 4)

def test_str_case():
    expect("HELLO", "hello".upper())
    expect("hello", "HELLO".lower())

def test_str_strip():
    expect("hello", "  hello  ".strip())

def test_str_split_join():
    expect(["hello", "world", "python"], "hello world python".split())
    expect(["a", "b", "c", "d"], "a,b,c,d".split(","))
    expect("a,b,c", ",".join(["a", "b", "c"]))

def test_str_replace():
    expect("hi hi hi", "hello hello hello".replace("hello", "hi"))

def test_str_membership():
    expect(True, "ell" in "hello")
    expect(True, "xyz" not in "hello")

def test_str_comparison():
    expect(True, "hello" == "hello")
    expect(True, "hello" != "world")
    expect(True, "apple" < "banana")
    expect(True, "banana" > "apple")

def test_str_word_count():
    text = "The quick brown fox"
    words = text.split()
    expect(4, len(words))

def test_str_empty():
    expect(0, len(""))
    expect(False, bool(""))

def test_str_mult_edge():
    expect("", "hello" * 0)
    expect("hello", "hello" * 1)

def test_str_case_insensitive():
    s1 = "Hello"
    s2 = "hello"
    expect(True, s1.lower() == s2.lower())

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
