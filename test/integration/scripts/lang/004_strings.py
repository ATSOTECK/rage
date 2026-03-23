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

def test_utf8_basic():
    # Basic Unicode strings
    expect("こんにちは").to_be("こんにちは")
    expect("Привет").to_be("Привет")
    expect("مرحبا").to_be("مرحبا")
    expect("🎉🎊🎁").to_be("🎉🎊🎁")

def test_utf8_len():
    # len() should return number of characters, not bytes
    expect(len("hello")).to_be(5)
    expect(len("café")).to_be(4)
    expect(len("日本語")).to_be(3)
    expect(len("Ω")).to_be(1)

def test_utf8_indexing():
    s = "αβγδε"
    expect(s[0]).to_be("α")
    expect(s[1]).to_be("β")
    expect(s[-1]).to_be("ε")
    expect(s[2]).to_be("γ")

def test_utf8_slicing():
    s = "日本語テスト"
    expect(s[0:3]).to_be("日本語")
    expect(s[3:]).to_be("テスト")
    expect(s[:3]).to_be("日本語")
    expect(s[::2]).to_be("日語ス")

def test_utf8_concat():
    expect("Hello " + "世界").to_be("Hello 世界")
    expect("🌟" * 3).to_be("🌟🌟🌟")

def test_utf8_membership():
    expect("日" in "日本語").to_be(True)
    expect("中" not in "日本語").to_be(True)
    expect("α" in "αβγ").to_be(True)

def test_utf8_comparison():
    expect("α" < "β").to_be(True)
    expect("日本" == "日本").to_be(True)
    expect("café" != "cafe").to_be(True)

def test_utf8_methods():
    # upper/lower for ASCII in mixed strings
    expect("Café".lower()).to_be("café")
    expect("Café".upper()).to_be("CAFÉ")
    # split and join with Unicode
    expect("日,本,語".split(",")).to_be(["日", "本", "語"])
    expect("-".join(["α", "β", "γ"])).to_be("α-β-γ")

def test_utf8_identifiers():
    # Unicode variable names (Python 3 feature)
    α = 1
    β = 2
    γ = α + β
    expect(γ).to_be(3)

    日本語 = "Japanese"
    expect(日本語).to_be("Japanese")

    переменная = 42
    expect(переменная).to_be(42)

def test_utf8_function_names():
    def 挨拶():
        return "こんにちは"
    expect(挨拶()).to_be("こんにちは")

    def приветствие(имя):
        return "Привет, " + имя
    expect(приветствие("мир")).to_be("Привет, мир")

def test_utf8_class_names():
    class 人:
        def __init__(self, 名前):
            self.名前 = 名前
        def 挨拶(self):
            return "こんにちは、" + self.名前 + "です"

    太郎 = 人("太郎")
    expect(太郎.名前).to_be("太郎")
    expect(太郎.挨拶()).to_be("こんにちは、太郎です")

def test_utf8_mixed():
    # Mixed ASCII and Unicode
    message = "Price: €100 (¥15000)"
    expect("€" in message).to_be(True)
    expect("¥" in message).to_be(True)
    expect(len(message)).to_be(20)

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
test("utf8_basic", test_utf8_basic)
test("utf8_len", test_utf8_len)
test("utf8_indexing", test_utf8_indexing)
test("utf8_slicing", test_utf8_slicing)
test("utf8_concat", test_utf8_concat)
test("utf8_membership", test_utf8_membership)
test("utf8_comparison", test_utf8_comparison)
test("utf8_methods", test_utf8_methods)
test("utf8_identifiers", test_utf8_identifiers)
test("utf8_function_names", test_utf8_function_names)
test("utf8_class_names", test_utf8_class_names)
test("utf8_mixed", test_utf8_mixed)

# --- len() on Unicode strings (character count, not byte count) ---

def test_len_chinese():
    expect(len("世界")).to_be(2)

def test_len_mixed_ascii_unicode():
    expect(len("hello 世界")).to_be(8)

def test_len_cafe():
    expect(len("café")).to_be(4)

def test_len_japanese():
    expect(len("日本語")).to_be(3)

def test_len_cyrillic():
    expect(len("Привет")).to_be(6)

def test_len_single_multibyte():
    expect(len("日")).to_be(1)

def test_len_unicode_consistency():
    s = "こんにちは"
    count = 0
    for c in s:
        count = count + 1
    expect(len(s)).to_be(count)
    expect(count).to_be(5)

def test_len_unicode_variable():
    s = "αβγδε"
    expect(len(s)).to_be(5)

test("len_chinese", test_len_chinese)
test("len_mixed_ascii_unicode", test_len_mixed_ascii_unicode)
test("len_cafe", test_len_cafe)
test("len_japanese", test_len_japanese)
test("len_cyrillic", test_len_cyrillic)
test("len_single_multibyte", test_len_single_multibyte)
test("len_unicode_consistency", test_len_unicode_consistency)
test("len_unicode_variable", test_len_unicode_variable)

print("Strings tests completed")
