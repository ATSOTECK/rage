# Test: String Operations
# Tests string methods and operations

results = {}

# Basic string operations
s = "Hello, World!"
results["str_len"] = len(s)
results["str_index"] = s[0]
results["str_negative_index"] = s[-1]

# String concatenation
results["str_concat"] = "Hello" + " " + "World"
results["str_repeat"] = "ab" * 4

# String methods - case
results["upper"] = "hello".upper()
results["lower"] = "HELLO".lower()

# String methods - whitespace
results["strip"] = "  hello  ".strip()

# String methods - split/join
results["split_default"] = "hello world python".split()
results["split_char"] = "a,b,c,d".split(",")
results["join_list"] = ",".join(["a", "b", "c"])

# String methods - replace
results["replace_all"] = "hello hello hello".replace("hello", "hi")

# String membership
results["in_str"] = "ell" in "hello"
results["not_in_str"] = "xyz" not in "hello"

# String comparison
results["str_eq"] = "hello" == "hello"
results["str_ne"] = "hello" != "world"
results["str_lt"] = "apple" < "banana"
results["str_gt"] = "banana" > "apple"

# Multi-character operations
text = "The quick brown fox"
words = text.split()
results["word_count"] = len(words)

# Empty string operations
results["empty_str_len"] = len("")
results["empty_str_bool"] = bool("")

# String multiplication edge cases
results["str_mult_zero"] = "hello" * 0
results["str_mult_one"] = "hello" * 1

# Case-insensitive comparison
s1 = "Hello"
s2 = "hello"
results["case_insensitive"] = s1.lower() == s2.lower()

print("Strings tests completed")
