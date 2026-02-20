from test_framework import test, expect
import copy

# =====================
# copy.copy() tests
# =====================

# Test 1: Shallow copy of a list
def test_copy_list():
    original = [1, 2, [3, 4]]
    copied = copy.copy(original)
    expect(copied).to_be([1, 2, [3, 4]])
    # Modifying the nested list in copied should affect original (shallow)
    copied[2].append(5)
    expect(original[2]).to_be([3, 4, 5])

test("copy.copy list (shallow)", test_copy_list)

# Test 2: Shallow copy of a dict
def test_copy_dict():
    original = {"a": 1, "b": [2, 3]}
    copied = copy.copy(original)
    expect(copied["a"]).to_be(1)
    expect(copied["b"]).to_be([2, 3])
    # Shared reference
    copied["b"].append(4)
    expect(original["b"]).to_be([2, 3, 4])

test("copy.copy dict (shallow)", test_copy_dict)

# Test 3: Shallow copy of a set
def test_copy_set():
    original = {1, 2, 3}
    copied = copy.copy(original)
    copied.add(4)
    expect(4 in original).to_be(False)

test("copy.copy set (shallow)", test_copy_set)

# Test 4: Copy of immutable types returns same object
def test_copy_immutables():
    x = 42
    expect(copy.copy(x)).to_be(42)
    s = "hello"
    expect(copy.copy(s)).to_be("hello")
    t = (1, 2, 3)
    expect(copy.copy(t)).to_be((1, 2, 3))
    expect(copy.copy(None)).to_be(None)
    expect(copy.copy(True)).to_be(True)

test("copy.copy immutable types", test_copy_immutables)

# Test 5: Shallow copy of instance (default)
def test_copy_instance_default():
    class Point:
        def __init__(self, x, y):
            self.x = x
            self.y = y

    p1 = Point(1, 2)
    p2 = copy.copy(p1)
    expect(p2.x).to_be(1)
    expect(p2.y).to_be(2)
    # Modifying copy doesn't affect original
    p2.x = 10
    expect(p1.x).to_be(1)

test("copy.copy instance default", test_copy_instance_default)

# Test 6: __copy__ dunder method
def test_copy_dunder():
    class Special:
        def __init__(self, val):
            self.val = val
            self.copy_count = 0

        def __copy__(self):
            new = Special(self.val)
            new.copy_count = self.copy_count + 1
            return new

    s = Special(42)
    c = copy.copy(s)
    expect(c.val).to_be(42)
    expect(c.copy_count).to_be(1)
    expect(s.copy_count).to_be(0)

test("copy.copy uses __copy__", test_copy_dunder)

# Test 7: __copy__ inherited from parent
def test_copy_inherited():
    class Base:
        def __init__(self, val):
            self.val = val

        def __copy__(self):
            return type(self)(self.val * 2)

    class Child(Base):
        pass

    c = Child(5)
    copied = copy.copy(c)
    expect(copied.val).to_be(10)

test("copy.copy __copy__ inherited", test_copy_inherited)

# =====================
# copy.deepcopy() tests
# =====================

# Test 8: Deep copy of nested list
def test_deepcopy_list():
    original = [1, 2, [3, 4]]
    copied = copy.deepcopy(original)
    expect(copied).to_be([1, 2, [3, 4]])
    # Modifying nested list in copy should NOT affect original
    copied[2].append(5)
    expect(original[2]).to_be([3, 4])
    expect(copied[2]).to_be([3, 4, 5])

test("copy.deepcopy list (deep)", test_deepcopy_list)

# Test 9: Deep copy of nested dict
def test_deepcopy_dict():
    original = {"a": {"b": [1, 2]}}
    copied = copy.deepcopy(original)
    copied["a"]["b"].append(3)
    expect(original["a"]["b"]).to_be([1, 2])
    expect(copied["a"]["b"]).to_be([1, 2, 3])

test("copy.deepcopy dict (deep)", test_deepcopy_dict)

# Test 10: Deep copy of tuple with mutable contents
def test_deepcopy_tuple():
    inner = [1, 2]
    original = (inner, 3)
    copied = copy.deepcopy(original)
    copied[0].append(99)
    expect(inner).to_be([1, 2])
    expect(copied[0]).to_be([1, 2, 99])

test("copy.deepcopy tuple with mutable contents", test_deepcopy_tuple)

# Test 11: Deep copy of instance (default)
def test_deepcopy_instance_default():
    class Container:
        def __init__(self, data):
            self.data = data

    c1 = Container([1, 2, 3])
    c2 = copy.deepcopy(c1)
    c2.data.append(4)
    expect(c1.data).to_be([1, 2, 3])
    expect(c2.data).to_be([1, 2, 3, 4])

test("copy.deepcopy instance default", test_deepcopy_instance_default)

# Test 12: __deepcopy__ dunder method
def test_deepcopy_dunder():
    class Custom:
        def __init__(self, val):
            self.val = val
            self.deep = False

        def __deepcopy__(self, memo):
            new = Custom(self.val)
            new.deep = True
            return new

    c = Custom(99)
    d = copy.deepcopy(c)
    expect(d.val).to_be(99)
    expect(d.deep).to_be(True)
    expect(c.deep).to_be(False)

test("copy.deepcopy uses __deepcopy__", test_deepcopy_dunder)

# Test 13: __deepcopy__ inherited
def test_deepcopy_inherited():
    class Base:
        def __init__(self, val):
            self.val = val

        def __deepcopy__(self, memo):
            return type(self)(self.val + 100)

    class Child(Base):
        pass

    c = Child(5)
    copied = copy.deepcopy(c)
    expect(copied.val).to_be(105)

test("copy.deepcopy __deepcopy__ inherited", test_deepcopy_inherited)

# Test 14: Deep copy immutables
def test_deepcopy_immutables():
    expect(copy.deepcopy(42)).to_be(42)
    expect(copy.deepcopy("hi")).to_be("hi")
    expect(copy.deepcopy(None)).to_be(None)
    expect(copy.deepcopy(True)).to_be(True)
    expect(copy.deepcopy(3.14)).to_be(3.14)

test("copy.deepcopy immutable types", test_deepcopy_immutables)

# Test 15: Shallow copy instance with shared mutable attr
def test_shallow_shared_mutable():
    class Holder:
        def __init__(self, items):
            self.items = items

    shared = [1, 2, 3]
    h1 = Holder(shared)
    h2 = copy.copy(h1)
    # Shallow copy shares the list
    h2.items.append(4)
    expect(h1.items).to_be([1, 2, 3, 4])

test("copy.copy instance shares mutable attrs", test_shallow_shared_mutable)

# Test 16: Deep copy instance with nested instances
def test_deepcopy_nested_instances():
    class Inner:
        def __init__(self, val):
            self.val = val

    class Outer:
        def __init__(self, inner):
            self.inner = inner

    o1 = Outer(Inner(42))
    o2 = copy.deepcopy(o1)
    o2.inner.val = 99
    expect(o1.inner.val).to_be(42)
    expect(o2.inner.val).to_be(99)

test("copy.deepcopy nested instances", test_deepcopy_nested_instances)
