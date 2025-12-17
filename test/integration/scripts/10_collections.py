# Test: Collection Operations
# Tests list, dict, and tuple operations

results = {}

# =====================================
# List Operations
# =====================================

# List append
lst = [1, 2, 3]
lst.append(4)
results["list_append"] = lst

# List extend
lst = [1, 2]
lst.extend([3, 4, 5])
results["list_extend"] = lst

# List pop
lst = [1, 2, 3, 4, 5]
popped = lst.pop()
results["list_pop_last"] = popped
results["list_after_pop"] = lst

# List with negative indexing
lst = [1, 2, 3, 4, 5]
results["list_neg_index"] = lst[-2]

# List membership
results["list_in"] = 3 in [1, 2, 3, 4]
results["list_not_in"] = 5 not in [1, 2, 3, 4]

# =====================================
# Dictionary Operations
# =====================================

# Dict access
d = {"a": 1, "b": 2, "c": 3}
results["dict_get_key"] = d["a"]

# Dict get with default
results["dict_get_exists"] = d.get("a")
results["dict_get_default"] = d.get("z", 99)
results["dict_get_none"] = d.get("z")

# Dict membership
d = {"a": 1, "b": 2}
results["dict_in_key"] = "a" in d
results["dict_not_in_key"] = "z" not in d

# Dict len
results["dict_len"] = len({"a": 1, "b": 2, "c": 3})

# =====================================
# Tuple Operations
# =====================================

# Tuple slicing
t = (1, 2, 3, 4, 5)
results["tuple_neg_index"] = t[-1]

# Tuple membership
results["tuple_in"] = 2 in (1, 2, 3)

# Tuple unpacking
t = (1, 2, 3)
a, b, c = t
results["tuple_unpack"] = [a, b, c]

# Tuple with single element
t = (42,)
results["tuple_single"] = len(t)

print("Collections tests completed")
