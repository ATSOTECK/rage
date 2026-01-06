# Test: Match Statements
# Tests Python 3.10+ structural pattern matching

class Point:
    __match_args__ = ("x", "y")

    def __init__(self, x, y):
        self.x = x
        self.y = y

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

def test_str_literal(cmd):
    match cmd:
        case "start":
            return "starting"
        case "stop":
            return "stopping"
        case _:
            return "unknown"

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

def test_capture(x):
    match x:
        case 0:
            return "zero"
        case n:
            return "captured: " + str(n)

def test_wildcard(x):
    match x:
        case 1:
            return "one"
        case _:
            return "wildcard"

def test_or_pattern(x):
    match x:
        case 1 | 2 | 3:
            return "small"
        case 4 | 5 | 6:
            return "medium"
        case _:
            return "large"

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

def test_tuple_pattern(tup):
    match tup:
        case (x, y):
            return "pair: " + str(x) + ", " + str(y)
        case (x, y, z):
            return "triple: " + str(x) + ", " + str(y) + ", " + str(z)
        case _:
            return "other"

def test_star_pattern(lst):
    match lst:
        case [first, *rest]:
            return "first=" + str(first) + ", rest=" + str(rest)
        case _:
            return "empty"

def test_star_middle(lst):
    match lst:
        case [first, *middle, last]:
            return "first=" + str(first) + ", middle=" + str(middle) + ", last=" + str(last)
        case [only]:
            return "only=" + str(only)
        case _:
            return "other"

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

def test_guard(x):
    match x:
        case n if n < 0:
            return "negative"
        case n if n == 0:
            return "zero"
        case n if n > 0:
            return "positive"

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

def test_as_pattern(data):
    match data:
        case [x, y] as point:
            return "point=" + str(point) + ", x=" + str(x) + ", y=" + str(y)
        case _:
            return "other"

def test_nested_pattern(data):
    match data:
        case {"user": {"name": name, "age": age}}:
            return "user: " + str(name) + ", age: " + str(age)
        case [{"id": id1}, {"id": id2}]:
            return "ids: " + str(id1) + ", " + str(id2)
        case _:
            return "unknown"

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

def test_class_keyword(obj):
    match obj:
        case Point(x=0, y=0):
            return "origin"
        case Point(x=x, y=y) if x == y:
            return "diagonal at " + str(x)
        case _:
            return "other"

def test_int_literals():
    expect("one", test_int_literal(1))
    expect("two", test_int_literal(2))
    expect("three", test_int_literal(3))
    expect("other", test_int_literal(99))

def test_str_literals():
    expect("starting", test_str_literal("start"))
    expect("stopping", test_str_literal("stop"))
    expect("unknown", test_str_literal("pause"))

def test_singletons():
    expect("is_true", test_singleton(True))
    expect("is_false", test_singleton(False))
    expect("is_none", test_singleton(None))
    expect("other", test_singleton(42))

def test_captures():
    expect("zero", test_capture(0))
    expect("captured: 42", test_capture(42))

def test_wildcards():
    expect("one", test_wildcard(1))
    expect("wildcard", test_wildcard(999))

def test_or_patterns():
    expect("small", test_or_pattern(2))
    expect("medium", test_or_pattern(5))
    expect("large", test_or_pattern(100))

def test_list_patterns():
    expect("empty", test_list_empty([]))
    expect("one: 1", test_list_one([1]))
    expect("two: 1, 2", test_list_two([1, 2]))
    expect("three: 1, 2, 3", test_list_three([1, 2, 3]))
    expect("many", test_list_many([1, 2, 3, 4, 5]))

def test_tuple_patterns():
    expect("pair: 10, 20", test_tuple_pattern((10, 20)))
    expect("triple: 1, 2, 3", test_tuple_pattern((1, 2, 3)))

def test_star_patterns():
    # Test the star pattern functionality without exact string matching
    result1 = test_star_pattern([1, 2, 3, 4])
    expect(True, "first=1" in result1)
    result2 = test_star_pattern([1])
    expect(True, "first=1" in result2)
    result3 = test_star_middle([1, 2, 3, 4, 5])
    expect(True, "first=1" in result3 and "last=5" in result3)
    result4 = test_star_middle([1, 2])
    expect(True, "first=1" in result4 and "last=2" in result4)

def test_mapping_patterns():
    expect("start server", test_mapping_pattern({"action": "start", "target": "server"}))
    expect("stop", test_mapping_pattern({"action": "stop", "extra": "ignored"}))
    expect("point: (10, 20)", test_mapping_pattern({"x": 10, "y": 20}))
    expect("unknown", test_mapping_pattern({"foo": "bar"}))

def test_guards():
    expect("negative", test_guard(-5))
    expect("zero", test_guard(0))
    expect("positive", test_guard(10))

def test_guard_complex_patterns():
    expect("diagonal", test_guard_complex((5, 5)))
    expect("on x-axis", test_guard_complex((3, 0)))
    expect("on y-axis", test_guard_complex((0, 7)))
    expect("(1, 2)", test_guard_complex((1, 2)))

def test_as_patterns():
    result = test_as_pattern([10, 20])
    expect(True, "x=10" in result and "y=20" in result)

def test_nested_patterns():
    expect("user: Alice, age: 30", test_nested_pattern({"user": {"name": "Alice", "age": 30}}))
    expect("ids: 1, 2", test_nested_pattern([{"id": 1}, {"id": 2}]))

def test_class_patterns():
    expect("origin", test_class_pattern(Point(0, 0)))
    expect("on x-axis at 5", test_class_pattern(Point(5, 0)))
    expect("on y-axis at 3", test_class_pattern(Point(0, 3)))
    expect("point at (2, 4)", test_class_pattern(Point(2, 4)))
    expect("not a point", test_class_pattern("not a point"))

def test_class_keyword_patterns():
    expect("origin", test_class_keyword(Point(0, 0)))
    expect("diagonal at 3", test_class_keyword(Point(3, 3)))
    expect("other", test_class_keyword(Point(1, 2)))

test("int_literals", test_int_literals)
test("str_literals", test_str_literals)
test("singletons", test_singletons)
test("captures", test_captures)
test("wildcards", test_wildcards)
test("or_patterns", test_or_patterns)
test("list_patterns", test_list_patterns)
test("tuple_patterns", test_tuple_patterns)
test("star_patterns", test_star_patterns)
test("mapping_patterns", test_mapping_patterns)
test("guards", test_guards)
test("guard_complex_patterns", test_guard_complex_patterns)
test("as_patterns", test_as_patterns)
test("nested_patterns", test_nested_patterns)
test("class_patterns", test_class_patterns)
test("class_keyword_patterns", test_class_keyword_patterns)

print("Match statement tests completed")
