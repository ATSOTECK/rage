from test_framework import test, expect
from dataclasses import dataclass, field, fields, asdict, astuple, MISSING

# Test 1: Basic dataclass with __init__
def test_basic():
    @dataclass
    class Point:
        x: int
        y: int

    p = Point(3, 4)
    expect(p.x).to_be(3)
    expect(p.y).to_be(4)

test("basic dataclass", test_basic)

# Test 2: __repr__
def test_repr():
    @dataclass
    class Point:
        x: int
        y: int

    p = Point(1, 2)
    expect(repr(p)).to_be("Point(x=1, y=2)")

test("dataclass repr", test_repr)

# Test 3: __eq__
def test_eq():
    @dataclass
    class Point:
        x: int
        y: int

    p1 = Point(1, 2)
    p2 = Point(1, 2)
    p3 = Point(1, 3)
    expect(p1 == p2).to_be(True)
    expect(p1 == p3).to_be(False)
    expect(p1 != p3).to_be(True)

test("dataclass eq", test_eq)

# Test 4: Default values
def test_defaults():
    @dataclass
    class Config:
        name: str
        value: int = 42

    c = Config("test")
    expect(c.name).to_be("test")
    expect(c.value).to_be(42)

    c2 = Config("test2", 100)
    expect(c2.value).to_be(100)

test("dataclass defaults", test_defaults)

# Test 5: field(default=...)
def test_field_default():
    @dataclass
    class Item:
        name: str
        count: int = field(default=0)

    i = Item("thing")
    expect(i.count).to_be(0)

test("field default", test_field_default)

# Test 6: field(default_factory=...)
def test_default_factory():
    @dataclass
    class Container:
        items: list = field(default_factory=list)

    c1 = Container()
    c2 = Container()
    expect(c1.items).to_be([])
    expect(c2.items).to_be([])
    # Verify they are independent instances
    c1.items.append(1)
    expect(c1.items).to_be([1])
    expect(c2.items).to_be([])

test("field default_factory", test_default_factory)

# Test 7: order=True
def test_order():
    @dataclass(order=True)
    class Version:
        major: int
        minor: int

    v1 = Version(1, 0)
    v2 = Version(2, 0)
    v3 = Version(1, 5)
    expect(v1 < v2).to_be(True)
    expect(v2 > v1).to_be(True)
    expect(v1 < v3).to_be(True)
    expect(v1 <= v1).to_be(True)
    expect(v2 >= v1).to_be(True)

test("dataclass order", test_order)

# Test 8: frozen=True
def test_frozen():
    @dataclass(frozen=True)
    class Frozen:
        x: int
        y: int

    f = Frozen(1, 2)
    expect(f.x).to_be(1)
    expect(f.y).to_be(2)

    error = None
    try:
        f.x = 10
    except Exception as e:
        error = str(e)
    expect(error is not None).to_be(True)

test("frozen dataclass", test_frozen)

# Test 9: field(init=False)
def test_init_false():
    @dataclass
    class WithComputed:
        x: int
        y: int
        total: int = field(init=False, default=0)

    w = WithComputed(3, 4)
    expect(w.x).to_be(3)
    expect(w.y).to_be(4)
    expect(w.total).to_be(0)

test("field init=False", test_init_false)

# Test 10: field(repr=False)
def test_repr_false():
    @dataclass
    class Secret:
        name: str
        password: str = field(repr=False)

    s = Secret("admin", "secret123")
    r = repr(s)
    expect("admin" in r).to_be(True)
    expect("secret123" in r).to_be(False)

test("field repr=False", test_repr_false)

# Test 11: Inheritance
def test_inheritance():
    @dataclass
    class Base:
        x: int
        y: int

    @dataclass
    class Child(Base):
        z: int

    c = Child(1, 2, 3)
    expect(c.x).to_be(1)
    expect(c.y).to_be(2)
    expect(c.z).to_be(3)
    expect(repr(c)).to_be("Child(x=1, y=2, z=3)")

test("dataclass inheritance", test_inheritance)

# Test 12: fields() function
def test_fields_fn():
    @dataclass
    class Point:
        x: int
        y: int

    fs = fields(Point)
    expect(len(fs)).to_be(2)

    # Also works with instance
    p = Point(1, 2)
    fs2 = fields(p)
    expect(len(fs2)).to_be(2)

test("fields() function", test_fields_fn)

# Test 13: asdict()
def test_asdict():
    @dataclass
    class Point:
        x: int
        y: int

    p = Point(3, 4)
    d = asdict(p)
    expect(d["x"]).to_be(3)
    expect(d["y"]).to_be(4)

test("asdict()", test_asdict)

# Test 14: astuple()
def test_astuple():
    @dataclass
    class Point:
        x: int
        y: int

    p = Point(3, 4)
    t = astuple(p)
    expect(t).to_be((3, 4))

test("astuple()", test_astuple)

# Test 15: __post_init__
def test_post_init():
    @dataclass
    class WithPostInit:
        x: int
        y: int

        def __post_init__(self):
            self.total = self.x + self.y

    w = WithPostInit(3, 4)
    expect(w.total).to_be(7)

test("__post_init__", test_post_init)

# Test 16: kwargs in constructor
def test_kwargs():
    @dataclass
    class Point:
        x: int
        y: int

    p = Point(x=10, y=20)
    expect(p.x).to_be(10)
    expect(p.y).to_be(20)

test("kwargs in constructor", test_kwargs)

# Test 17: @dataclass() with empty parens
def test_empty_parens():
    @dataclass()
    class Empty:
        x: int

    e = Empty(42)
    expect(e.x).to_be(42)

test("@dataclass() with parens", test_empty_parens)

# Test 18: init=False option
def test_no_init():
    @dataclass(init=False)
    class Manual:
        x: int
        y: int

        def __init__(self, val):
            self.x = val
            self.y = val * 2

    m = Manual(5)
    expect(m.x).to_be(5)
    expect(m.y).to_be(10)

test("init=False option", test_no_init)

# Test 19: repr=False option
def test_no_repr():
    @dataclass(repr=False)
    class NoRepr:
        x: int

    n = NoRepr(1)
    r = repr(n)
    # Should use default object repr, not dataclass repr
    expect("NoRepr(x=1)" in r).to_be(False)

test("repr=False option", test_no_repr)

# Test 20: eq=False option
def test_no_eq():
    @dataclass(eq=False)
    class NoEq:
        x: int

    a = NoEq(1)
    b = NoEq(1)
    # Without eq, identity comparison
    expect(a == b).to_be(False)
    expect(a == a).to_be(True)

test("eq=False option", test_no_eq)

# Test 21: MISSING sentinel
def test_missing():
    expect(MISSING is not None).to_be(True)
    expect(str(MISSING)).to_be("MISSING")

test("MISSING sentinel", test_missing)

# Test 22: Multiple default factories are independent
def test_independent_factories():
    @dataclass
    class MultiFactory:
        a: list = field(default_factory=list)
        b: dict = field(default_factory=dict)

    m1 = MultiFactory()
    m2 = MultiFactory()
    m1.a.append(1)
    expect(m2.a).to_be([])
    expect(m1.a).to_be([1])

test("independent factories", test_independent_factories)

# Test 23: field(compare=False)
def test_compare_false():
    @dataclass
    class WithIgnored:
        x: int
        label: str = field(compare=False)

    a = WithIgnored(1, "alpha")
    b = WithIgnored(1, "beta")
    expect(a == b).to_be(True)

test("field compare=False", test_compare_false)

# Test 24: frozen dataclass hash
def test_frozen_hash():
    @dataclass(frozen=True)
    class FrozenPoint:
        x: int
        y: int

    p1 = FrozenPoint(1, 2)
    p2 = FrozenPoint(1, 2)
    expect(hash(p1) == hash(p2)).to_be(True)

test("frozen dataclass hash", test_frozen_hash)

# Test 25: Mixed positional and keyword args
def test_mixed_args():
    @dataclass
    class Mixed:
        a: int
        b: int
        c: int = 30

    m = Mixed(1, b=2)
    expect(m.a).to_be(1)
    expect(m.b).to_be(2)
    expect(m.c).to_be(30)

test("mixed positional and keyword args", test_mixed_args)
