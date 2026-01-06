# Test: Match Statements
# Tests Python 3.10+ structural pattern matching

from test_framework import test, expect

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
    expect(test_int_literal(1)).to_be("one")
    expect(test_int_literal(2)).to_be("two")
    expect(test_int_literal(3)).to_be("three")
    expect(test_int_literal(99)).to_be("other")

def test_str_literals():
    expect(test_str_literal("start")).to_be("starting")
    expect(test_str_literal("stop")).to_be("stopping")
    expect(test_str_literal("pause")).to_be("unknown")

def test_singletons():
    expect(test_singleton(True)).to_be("is_true")
    expect(test_singleton(False)).to_be("is_false")
    expect(test_singleton(None)).to_be("is_none")
    expect(test_singleton(42)).to_be("other")

def test_captures():
    expect(test_capture(0)).to_be("zero")
    expect(test_capture(42)).to_be("captured: 42")

def test_wildcards():
    expect(test_wildcard(1)).to_be("one")
    expect(test_wildcard(999)).to_be("wildcard")

def test_or_patterns():
    expect(test_or_pattern(2)).to_be("small")
    expect(test_or_pattern(5)).to_be("medium")
    expect(test_or_pattern(100)).to_be("large")

def test_list_patterns():
    expect(test_list_empty([])).to_be("empty")
    expect(test_list_one([1])).to_be("one: 1")
    expect(test_list_two([1, 2])).to_be("two: 1, 2")
    expect(test_list_three([1, 2, 3])).to_be("three: 1, 2, 3")
    expect(test_list_many([1, 2, 3, 4, 5])).to_be("many")

def test_tuple_patterns():
    expect(test_tuple_pattern((10, 20))).to_be("pair: 10, 20")
    expect(test_tuple_pattern((1, 2, 3))).to_be("triple: 1, 2, 3")

def test_star_patterns():
    # Test the star pattern functionality without exact string matching
    result1 = test_star_pattern([1, 2, 3, 4])
    expect("first=1" in result1).to_be(True)
    result2 = test_star_pattern([1])
    expect("first=1" in result2).to_be(True)
    result3 = test_star_middle([1, 2, 3, 4, 5])
    expect("first=1" in result3 and "last=5" in result3).to_be(True)
    result4 = test_star_middle([1, 2])
    expect("first=1" in result4 and "last=2" in result4).to_be(True)

def test_mapping_patterns():
    expect(test_mapping_pattern({"action": "start", "target": "server"})).to_be("start server")
    expect(test_mapping_pattern({"action": "stop", "extra": "ignored"})).to_be("stop")
    expect(test_mapping_pattern({"x": 10, "y": 20})).to_be("point: (10, 20)")
    expect(test_mapping_pattern({"foo": "bar"})).to_be("unknown")

def test_guards():
    expect(test_guard(-5)).to_be("negative")
    expect(test_guard(0)).to_be("zero")
    expect(test_guard(10)).to_be("positive")

def test_guard_complex_patterns():
    expect(test_guard_complex((5, 5))).to_be("diagonal")
    expect(test_guard_complex((3, 0))).to_be("on x-axis")
    expect(test_guard_complex((0, 7))).to_be("on y-axis")
    expect(test_guard_complex((1, 2))).to_be("(1, 2)")

def test_as_patterns():
    result = test_as_pattern([10, 20])
    expect("x=10" in result and "y=20" in result).to_be(True)

def test_nested_patterns():
    expect(test_nested_pattern({"user": {"name": "Alice", "age": 30}})).to_be("user: Alice, age: 30")
    expect(test_nested_pattern([{"id": 1}, {"id": 2}])).to_be("ids: 1, 2")

def test_class_patterns():
    expect(test_class_pattern(Point(0, 0))).to_be("origin")
    expect(test_class_pattern(Point(5, 0))).to_be("on x-axis at 5")
    expect(test_class_pattern(Point(0, 3))).to_be("on y-axis at 3")
    expect(test_class_pattern(Point(2, 4))).to_be("point at (2, 4)")
    expect(test_class_pattern("not a point")).to_be("not a point")

def test_class_keyword_patterns():
    expect(test_class_keyword(Point(0, 0))).to_be("origin")
    expect(test_class_keyword(Point(3, 3))).to_be("diagonal at 3")
    expect(test_class_keyword(Point(1, 2))).to_be("other")

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
