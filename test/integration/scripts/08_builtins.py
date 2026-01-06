# Test: Built-in Functions
# Tests commonly used built-in functions

results = {}

# Type constructors
results["int_from_float"] = int(3.7)
results["int_from_str"] = int("42")
results["int_from_bool"] = int(True)

results["float_from_int"] = float(42)
results["float_from_str"] = float("3.14")

results["str_from_int"] = str(42)
results["str_from_float"] = str(3.14)
results["str_from_bool"] = str(True)

results["bool_from_int_true"] = bool(1)
results["bool_from_int_false"] = bool(0)
results["bool_from_str_true"] = bool("hello")
results["bool_from_str_false"] = bool("")
results["bool_from_list_true"] = bool([1])
results["bool_from_list_false"] = bool([])

results["list_from_tuple"] = list((1, 2, 3))
results["list_from_str"] = list("abc")
results["list_from_range"] = list(range(5))

results["tuple_from_list"] = tuple([1, 2, 3])
results["tuple_from_str"] = tuple("abc")

# len function
results["len_str"] = len("hello")
results["len_list"] = len([1, 2, 3, 4, 5])
results["len_tuple"] = len((1, 2, 3))
results["len_dict"] = len({"a": 1, "b": 2})

# min and max
results["min_args"] = min(5, 2, 8, 1, 9)
results["max_args"] = max(5, 2, 8, 1, 9)
results["min_list"] = min([5, 2, 8, 1, 9])
results["max_list"] = max([5, 2, 8, 1, 9])

# sum
results["sum_list"] = sum([1, 2, 3, 4, 5])
results["sum_range"] = sum(range(10))
results["sum_empty"] = sum([])

# abs
results["abs_positive"] = abs(42)
results["abs_negative"] = abs(-42)
results["abs_float"] = abs(-3.14)
results["abs_zero"] = abs(0)

# ord and chr
results["ord_a"] = ord("a")
results["ord_A"] = ord("A")
results["chr_97"] = chr(97)
results["chr_65"] = chr(65)

# isinstance
results["isinstance_int"] = isinstance(42, int)
results["isinstance_str"] = isinstance("hello", str)
results["isinstance_list"] = isinstance([1, 2], list)
results["isinstance_not"] = isinstance(42, str)

# range
results["range_simple"] = list(range(5))
results["range_start_stop"] = list(range(2, 7))
results["range_with_step"] = list(range(0, 10, 2))
results["range_negative"] = list(range(10, 0, -1))
results["range_empty"] = list(range(5, 2))

# enumerate
results["enumerate_basic"] = list(enumerate(["a", "b", "c"]))
results["enumerate_start"] = list(enumerate(["x", "y"], 1))
results["enumerate_empty"] = list(enumerate([]))
results["enumerate_string"] = list(enumerate("hi"))

# zip
results["zip_basic"] = list(zip([1, 2, 3], ["a", "b", "c"]))
results["zip_unequal"] = list(zip([1, 2], ["a", "b", "c", "d"]))
results["zip_empty"] = list(zip())
results["zip_single"] = list(zip([1, 2, 3]))
results["zip_three"] = list(zip([1, 2], ["a", "b"], [True, False]))

# map
def double(x):
    return x * 2

def add(a, b):
    return a + b

results["map_basic"] = list(map(double, [1, 2, 3, 4]))
results["map_strings"] = list(map(str, [1, 2, 3]))
results["map_two_args"] = list(map(add, [1, 2, 3], [10, 20, 30]))
results["map_empty"] = list(map(double, []))

# filter
def is_even(x):
    return x % 2 == 0

def is_positive(x):
    return x > 0

results["filter_basic"] = list(filter(is_even, [1, 2, 3, 4, 5, 6]))
results["filter_positive"] = list(filter(is_positive, [-2, -1, 0, 1, 2]))
results["filter_none"] = list(filter(None, [0, 1, "", "hello", [], [1]]))
results["filter_empty"] = list(filter(is_even, []))
results["filter_all_fail"] = list(filter(is_even, [1, 3, 5]))

# reversed
results["reversed_list"] = list(reversed([1, 2, 3, 4, 5]))
results["reversed_string"] = list(reversed("hello"))
results["reversed_tuple"] = list(reversed((1, 2, 3)))
results["reversed_empty"] = list(reversed([]))
results["reversed_single"] = list(reversed([42]))

# sorted
results["sorted_basic"] = sorted([3, 1, 4, 1, 5, 9, 2, 6])
results["sorted_string"] = sorted("hello")
results["sorted_reverse"] = sorted([3, 1, 2], reverse=True)
results["sorted_empty"] = sorted([])
results["sorted_single"] = sorted([42])
results["sorted_already"] = sorted([1, 2, 3, 4, 5])

def get_len(x):
    return len(x)
results["sorted_key"] = sorted(["apple", "pie", "a"], key=get_len)

# all
results["all_true"] = all([True, True, True])
results["all_false"] = all([True, False, True])
results["all_empty"] = all([])
results["all_numbers_true"] = all([1, 2, 3])
results["all_numbers_false"] = all([1, 0, 3])
results["all_strings"] = all(["a", "b", "c"])
results["all_strings_empty"] = all(["a", "", "c"])

# any
results["any_true"] = any([False, False, True])
results["any_false"] = any([False, False, False])
results["any_empty"] = any([])
results["any_numbers_true"] = any([0, 0, 1])
results["any_numbers_false"] = any([0, 0, 0])
results["any_mixed"] = any([0, "", [], "hello"])

# Attribute access builtins
class AttrTest:
    def __init__(self):
        self.x = 10
        self.y = 20

obj = AttrTest()

# hasattr
results["hasattr_exists"] = hasattr(obj, "x")
results["hasattr_missing"] = hasattr(obj, "z")
results["hasattr_method"] = hasattr(obj, "__init__")

# getattr
results["getattr_exists"] = getattr(obj, "x")
results["getattr_default"] = getattr(obj, "z", 99)
results["getattr_default_none"] = getattr(obj, "missing", None)

# setattr
setattr(obj, "z", 30)
results["setattr_new"] = obj.z
setattr(obj, "x", 100)
results["setattr_existing"] = obj.x

# delattr
setattr(obj, "temp", 42)
results["delattr_before"] = hasattr(obj, "temp")
delattr(obj, "temp")
results["delattr_after"] = hasattr(obj, "temp")

# pow
results["pow_int"] = pow(2, 3)
results["pow_large"] = pow(2, 10)
results["pow_mod"] = pow(2, 3, 5)
results["pow_mod2"] = pow(7, 2, 13)
results["pow_zero"] = pow(5, 0)
results["pow_one"] = pow(5, 1)

# divmod
results["divmod_basic"] = divmod(17, 5)
results["divmod_exact"] = divmod(10, 2)
results["divmod_neg1"] = divmod(-17, 5)
results["divmod_neg2"] = divmod(17, -5)

# hex
results["hex_255"] = hex(255)
results["hex_16"] = hex(16)
results["hex_0"] = hex(0)
results["hex_neg"] = hex(-255)

# oct
results["oct_8"] = oct(8)
results["oct_64"] = oct(64)
results["oct_0"] = oct(0)
results["oct_neg"] = oct(-8)

# bin
results["bin_5"] = bin(5)
results["bin_255"] = bin(255)
results["bin_0"] = bin(0)
results["bin_neg"] = bin(-5)

# round
results["round_up"] = round(3.7)
results["round_down"] = round(3.2)
results["round_half_even1"] = round(2.5)  # Banker's rounding
results["round_half_even2"] = round(3.5)  # Banker's rounding
results["round_digits"] = round(3.14159, 2)
results["round_digits4"] = round(3.14159, 4)
results["round_neg_digits"] = round(1234, -2)
results["round_negative"] = round(-2.5)
results["round_zero"] = round(0.0)

# callable
def test_func():
    pass

class TestClass:
    pass

class CallableClass:
    def __call__(self):
        pass

results["callable_func"] = callable(test_func)
results["callable_class"] = callable(TestClass)
results["callable_callable_inst"] = callable(CallableClass())
results["callable_inst"] = callable(TestClass())
results["callable_int"] = callable(42)
results["callable_str"] = callable("hello")
results["callable_list"] = callable([1, 2])
results["callable_builtin"] = callable(print)
results["callable_none"] = callable(None)

print("Builtins tests completed")
