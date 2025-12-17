# Test: Data Types
# Tests basic Python data types

results = {}

# None type
results["none_value"] = None
results["none_is_none"] = None is None

# Boolean type
results["bool_true"] = True
results["bool_false"] = False
results["bool_from_int"] = bool(1)
results["bool_from_zero"] = bool(0)
results["bool_from_string"] = bool("hello")
results["bool_from_empty"] = bool("")

# Integer type
results["int_positive"] = 42
results["int_negative"] = -17
results["int_zero"] = 0
results["int_large"] = 1000000
results["int_from_float"] = int(3.7)
results["int_from_string"] = int("123")

# Float type
results["float_positive"] = 3.14
results["float_negative"] = -2.5
results["float_zero"] = 0.0
results["float_from_int"] = float(42)
results["float_from_string"] = float("3.14")

# String type
results["str_single"] = 'hello'
results["str_double"] = "world"
results["str_empty"] = ""
results["str_with_spaces"] = "hello world"
results["str_concat"] = "hello" + " " + "world"
results["str_repeat"] = "ab" * 3
results["str_len"] = len("hello")
results["str_index"] = "hello"[1]
results["str_negative_index"] = "hello"[-1]

# List type
results["list_empty"] = []
results["list_ints"] = [1, 2, 3]
results["list_mixed"] = [1, "two", 3.0, True]
results["list_nested"] = [[1, 2], [3, 4]]
results["list_len"] = len([1, 2, 3, 4, 5])
results["list_index"] = [10, 20, 30][1]
results["list_concat"] = [1, 2] + [3, 4]
results["list_repeat"] = [1, 2] * 2
results["list_negative_index"] = [1, 2, 3][-1]

# Tuple type
results["tuple_empty"] = ()
results["tuple_single"] = (1,)
results["tuple_multi"] = (1, 2, 3)
results["tuple_mixed"] = (1, "two", 3.0)
results["tuple_len"] = len((1, 2, 3))
results["tuple_index"] = (10, 20, 30)[1]

# Dictionary type
results["dict_empty"] = {}
results["dict_simple"] = {"a": 1, "b": 2}
results["dict_mixed_values"] = {"int": 1, "str": "hello", "bool": True}
results["dict_len"] = len({"a": 1, "b": 2, "c": 3})
results["dict_access"] = {"x": 10, "y": 20}["x"]
results["dict_get"] = {"a": 1}.get("a")
results["dict_get_default"] = {"a": 1}.get("b", 99)

# Range type
results["range_simple"] = list(range(5))
results["range_start_stop"] = list(range(2, 7))
results["range_with_step"] = list(range(0, 10, 2))

# isinstance checks
results["isinstance_int"] = isinstance(42, int)
results["isinstance_str"] = isinstance("hello", str)
results["isinstance_list"] = isinstance([1, 2], list)

print("Data types tests completed")
