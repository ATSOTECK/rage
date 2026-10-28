# Test: collections module types
# Counter, defaultdict and OrderedDict are dicts; namedtuple builds a real
# class; deque supports len/iteration/indexing.

from test_framework import test, expect
from collections import Counter, defaultdict, OrderedDict, deque, namedtuple

def test_counter_basics():
    c = Counter("hello")
    expect((c["l"], c["z"], len(c), "z" in c)).to_be((2, 0, 4, False))
    expect(c.most_common(1)).to_be([("l", 2)])
    expect(sorted(c.elements())).to_be(["e", "h", "l", "l", "o"])
    expect(c.total()).to_be(5)
    expect(isinstance(c, dict)).to_be(True)

def test_counter_update_and_arithmetic():
    c = Counter(a=3, b=1)
    c.update("aab")
    expect(dict(c)).to_be({"a": 5, "b": 2})
    c.subtract({"a": 4})
    expect(dict(c)).to_be({"a": 1, "b": 2})
    expect(dict(Counter(a=3, b=1) + Counter(a=1, c=2))).to_be({"a": 4, "b": 1, "c": 2})
    expect(dict(Counter(a=3, b=1) - Counter(a=1, b=5))).to_be({"a": 2})
    expect(dict(Counter(a=3) | Counter(a=5, b=1))).to_be({"a": 5, "b": 1})
    expect(dict(Counter(a=3, b=2) & Counter(a=1))).to_be({"a": 1})

def test_counter_repr_is_most_common_first():
    expect(repr(Counter("hello"))).to_be("Counter({'l': 2, 'h': 1, 'e': 1, 'o': 1})")
    expect(repr(Counter())).to_be("Counter()")

def test_defaultdict():
    d = defaultdict(int)
    for w in "abca":
        d[w] += 1
    expect(dict(d)).to_be({"a": 2, "b": 1, "c": 1})
    g = defaultdict(list)
    g["k"].append(1)
    expect((g["k"], g.get("missing"), "missing" in g, len(g))).to_be(([1], None, False, 1))
    expect(isinstance(d, dict)).to_be(True)

def test_defaultdict_without_factory_raises():
    d = defaultdict()
    caught = False
    try:
        d["x"]
    except KeyError:
        caught = True
    expect(caught).to_be(True)

def test_ordered_dict():
    o = OrderedDict([("x", 1), ("y", 2), ("z", 3)])
    o.move_to_end("x")
    expect(list(o)).to_be(["y", "z", "x"])
    expect(o.popitem(last=False)).to_be(("y", 2))
    expect(repr(o)).to_be("OrderedDict({'z': 3, 'x': 1})")

def test_deque():
    d = deque([1, 2, 3], maxlen=3)
    d.append(4)
    expect((list(d), len(d), d[0], d[-1], 3 in d, 9 in d)).to_be(([2, 3, 4], 3, 2, 4, True, False))
    expect((bool(deque()), bool(d))).to_be((False, True))
    expect(repr(d)).to_be("deque([2, 3, 4], maxlen=3)")
    expect(list(d.copy())).to_be([2, 3, 4])

def test_namedtuple():
    P = namedtuple("P", "x y")
    p = P(1, y=2)
    a, b = p
    expect((p.x, p[1], p[-1], len(p), tuple(p), a, b)).to_be((1, 2, 2, 2, (1, 2), 1, 2))
    expect((p == (1, 2), p == P(1, 2), p < P(1, 3))).to_be((True, True, True))
    expect(repr(p)).to_be("P(x=1, y=2)")
    expect((p._asdict(), p._replace(x=9), P._make([3, 4]), P._fields)).to_be(({"x": 1, "y": 2}, P(9, 2), P(3, 4), ("x", "y")))
    expect(len({P(1, 2), P(1, 2)})).to_be(1)

def test_namedtuple_defaults_and_match():
    Q = namedtuple("Q", ["a", "b", "c"], defaults=[0, 0])
    expect((Q(1), Q(1, 2, 3))).to_be((Q(1, 0, 0), Q(1, 2, 3)))
    match Q(5):
        case Q(a, b, c):
            got = (a, b, c)
    expect(got).to_be((5, 0, 0))

test("counter_basics", test_counter_basics)
test("counter_update_and_arithmetic", test_counter_update_and_arithmetic)
test("counter_repr_is_most_common_first", test_counter_repr_is_most_common_first)
test("defaultdict", test_defaultdict)
test("defaultdict_without_factory_raises", test_defaultdict_without_factory_raises)
test("ordered_dict", test_ordered_dict)
test("deque", test_deque)
test("namedtuple", test_namedtuple)
test("namedtuple_defaults_and_match", test_namedtuple_defaults_and_match)

print("collections tests completed")
