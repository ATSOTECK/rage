# Test: Match Statements
# Tests Python 3.10+ structural pattern matching

results = {}

# === Literal patterns ===

# Integer literal
def test_int_literal(x):
    match x:
        case 1:
            return "one"
        case 2:
            return "two"
        case 3:
            return "three"
        case _:
            return "other"

results["int_1"] = test_int_literal(1)
results["int_2"] = test_int_literal(2)
results["int_3"] = test_int_literal(3)
results["int_other"] = test_int_literal(99)

# String literal
def test_str_literal(cmd):
    match cmd:
        case "start":
            return "starting"
        case "stop":
            return "stopping"
        case _:
            return "unknown"

results["str_start"] = test_str_literal("start")
results["str_stop"] = test_str_literal("stop")
results["str_unknown"] = test_str_literal("pause")

# === Singleton patterns (True, False, None) ===

def test_singleton(x):
    match x:
        case True:
            return "is_true"
        case False:
            return "is_false"
        case None:
            return "is_none"
        case _:
            return "other"

results["singleton_true"] = test_singleton(True)
results["singleton_false"] = test_singleton(False)
results["singleton_none"] = test_singleton(None)
results["singleton_other"] = test_singleton(42)

# === Capture patterns ===

def test_capture(x):
    match x:
        case 0:
            return "zero"
        case n:
            return "captured: " + str(n)

results["capture_zero"] = test_capture(0)
results["capture_num"] = test_capture(42)

# === Wildcard pattern ===

def test_wildcard(x):
    match x:
        case 1:
            return "one"
        case _:
            return "wildcard"

results["wildcard_match"] = test_wildcard(1)
results["wildcard_default"] = test_wildcard(999)

# === Or patterns ===

def test_or_pattern(x):
    match x:
        case 1 | 2 | 3:
            return "small"
        case 4 | 5 | 6:
            return "medium"
        case _:
            return "large"

results["or_small"] = test_or_pattern(2)
results["or_medium"] = test_or_pattern(5)
results["or_large"] = test_or_pattern(100)

# === Sequence patterns ===

# List pattern with exact length
def test_list_empty(lst):
    match lst:
        case []:
            return "empty"
        case _:
            return "not empty"

def test_list_one(lst):
    match lst:
        case [x]:
            return "one: " + str(x)
        case _:
            return "not one"

def test_list_two(lst):
    match lst:
        case [x, y]:
            return "two: " + str(x) + ", " + str(y)
        case _:
            return "not two"

def test_list_three(lst):
    match lst:
        case [x, y, z]:
            return "three: " + str(x) + ", " + str(y) + ", " + str(z)
        case _:
            return "not three"

def test_list_many(lst):
    match lst:
        case [_, _, _, _, *_]:
            return "many"
        case _:
            return "not many"

results["list_empty"] = test_list_empty([])
results["list_one"] = test_list_one([1])
results["list_two"] = test_list_two([1, 2])
results["list_three"] = test_list_three([1, 2, 3])
results["list_many"] = test_list_many([1, 2, 3, 4, 5])

# Tuple pattern
def test_tuple_pattern(tup):
    match tup:
        case (x, y):
            return "pair: " + str(x) + ", " + str(y)
        case (x, y, z):
            return "triple: " + str(x) + ", " + str(y) + ", " + str(z)
        case _:
            return "other"

results["tuple_pair"] = test_tuple_pattern((10, 20))
results["tuple_triple"] = test_tuple_pattern((1, 2, 3))

# Sequence with star pattern
def test_star_pattern(lst):
    match lst:
        case [first, *rest]:
            return "first=" + str(first) + ", rest=" + str(rest)
        case _:
            return "empty"

results["star_simple"] = test_star_pattern([1, 2, 3, 4])
results["star_one"] = test_star_pattern([1])

def test_star_middle(lst):
    match lst:
        case [first, *middle, last]:
            return "first=" + str(first) + ", middle=" + str(middle) + ", last=" + str(last)
        case [only]:
            return "only=" + str(only)
        case _:
            return "other"

results["star_middle"] = test_star_middle([1, 2, 3, 4, 5])
results["star_two_elem"] = test_star_middle([1, 2])

# === Mapping patterns ===

def test_mapping_pattern(d):
    match d:
        case {"action": "start", "target": t}:
            return "start " + str(t)
        case {"action": "stop"}:
            return "stop"
        case {"x": x, "y": y}:
            return "point: (" + str(x) + ", " + str(y) + ")"
        case _:
            return "unknown"

results["map_start"] = test_mapping_pattern({"action": "start", "target": "server"})
results["map_stop"] = test_mapping_pattern({"action": "stop", "extra": "ignored"})
results["map_point"] = test_mapping_pattern({"x": 10, "y": 20})
results["map_unknown"] = test_mapping_pattern({"foo": "bar"})

# === Guard conditions ===

def test_guard(x):
    match x:
        case n if n < 0:
            return "negative"
        case n if n == 0:
            return "zero"
        case n if n > 0:
            return "positive"

results["guard_neg"] = test_guard(-5)
results["guard_zero"] = test_guard(0)
results["guard_pos"] = test_guard(10)

def test_guard_complex(point):
    match point:
        case (x, y) if x == y:
            return "diagonal"
        case (x, 0):
            return "on x-axis"
        case (0, y):
            return "on y-axis"
        case (x, y):
            return "(" + str(x) + ", " + str(y) + ")"

results["guard_diag"] = test_guard_complex((5, 5))
results["guard_xaxis"] = test_guard_complex((3, 0))
results["guard_yaxis"] = test_guard_complex((0, 7))
results["guard_general"] = test_guard_complex((1, 2))

# === As patterns ===

def test_as_pattern(data):
    match data:
        case [x, y] as point:
            return "point=" + str(point) + ", x=" + str(x) + ", y=" + str(y)
        case _:
            return "other"

results["as_pattern"] = test_as_pattern([10, 20])

# === Nested patterns ===

def test_nested_pattern(data):
    match data:
        case {"user": {"name": name, "age": age}}:
            return "user: " + str(name) + ", age: " + str(age)
        case [{"id": id1}, {"id": id2}]:
            return "ids: " + str(id1) + ", " + str(id2)
        case _:
            return "unknown"

results["nested_dict"] = test_nested_pattern({"user": {"name": "Alice", "age": 30}})
results["nested_list"] = test_nested_pattern([{"id": 1}, {"id": 2}])

# === Class patterns ===

class Point:
    __match_args__ = ("x", "y")

    def __init__(self, x, y):
        self.x = x
        self.y = y

def test_class_pattern(obj):
    match obj:
        case Point(0, 0):
            return "origin"
        case Point(x, 0):
            return "on x-axis at " + str(x)
        case Point(0, y):
            return "on y-axis at " + str(y)
        case Point(x, y):
            return "point at (" + str(x) + ", " + str(y) + ")"
        case _:
            return "not a point"

results["class_origin"] = test_class_pattern(Point(0, 0))
results["class_xaxis"] = test_class_pattern(Point(5, 0))
results["class_yaxis"] = test_class_pattern(Point(0, 3))
results["class_general"] = test_class_pattern(Point(2, 4))
results["class_not_point"] = test_class_pattern("not a point")

# Class pattern with keyword args
def test_class_keyword(obj):
    match obj:
        case Point(x=0, y=0):
            return "origin"
        case Point(x=x, y=y) if x == y:
            return "diagonal at " + str(x)
        case _:
            return "other"

results["class_kw_origin"] = test_class_keyword(Point(0, 0))
results["class_kw_diag"] = test_class_keyword(Point(3, 3))
results["class_kw_other"] = test_class_keyword(Point(1, 2))

# Print summary
print("Match statement tests completed")
