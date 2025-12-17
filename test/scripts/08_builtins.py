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

print("Builtins tests completed")
