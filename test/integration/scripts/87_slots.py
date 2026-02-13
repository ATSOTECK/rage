from test_framework import test, expect

# Test 1: Basic slots with __init__
def test_basic_slots():
    class Point:
        __slots__ = ['x', 'y']

        def __init__(self, x, y):
            self.x = x
            self.y = y

    p = Point(3, 4)
    expect(p.x).to_be(3)
    expect(p.y).to_be(4)
    expect(p.x + p.y).to_be(7)

test("basic slots", test_basic_slots)

# Test 2: Rejecting assignment to non-slot attributes
def test_reject_non_slot():
    class Restricted:
        __slots__ = ['x']

    obj = Restricted()
    obj.x = 10
    expect(obj.x).to_be(10)

    error = None
    try:
        obj.y = 20
    except AttributeError as e:
        error = str(e)
    expect(error is not None).to_be(True)

test("reject non-slot attribute", test_reject_non_slot)

# Test 3: hasattr(obj, '__dict__') returns False for slotted instances
def test_no_dict():
    class Slotted:
        __slots__ = ['x']

    obj = Slotted()
    obj.x = 1
    expect(hasattr(obj, '__dict__')).to_be(False)

test("no __dict__ on slotted instances", test_no_dict)

# Test 4: Normal class still has __dict__
def test_normal_dict():
    class Normal:
        pass

    obj = Normal()
    obj.x = 1
    expect(hasattr(obj, '__dict__')).to_be(True)

test("normal class has __dict__", test_normal_dict)

# Test 5: Slots with inheritance
def test_slots_inheritance():
    class Base:
        __slots__ = ['x']

    class Child(Base):
        __slots__ = ['y']

    obj = Child()
    obj.x = 1
    obj.y = 2
    expect(obj.x).to_be(1)
    expect(obj.y).to_be(2)

    error = None
    try:
        obj.z = 3
    except AttributeError:
        error = True
    expect(error).to_be(True)

test("slots with inheritance", test_slots_inheritance)

# Test 6: Empty __slots__
def test_empty_slots():
    class Empty:
        __slots__ = []

    obj = Empty()
    error = None
    try:
        obj.x = 1
    except AttributeError:
        error = True
    expect(error).to_be(True)

test("empty slots", test_empty_slots)

# Test 7: Slots with methods
def test_slots_with_methods():
    class Vector:
        __slots__ = ['x', 'y']

        def __init__(self, x, y):
            self.x = x
            self.y = y

        def magnitude_squared(self):
            return self.x * self.x + self.y * self.y

    v = Vector(3, 4)
    expect(v.magnitude_squared()).to_be(25)

test("slots with methods", test_slots_with_methods)

# Test 8: Deleting slot attributes
def test_delete_slot():
    class Obj:
        __slots__ = ['x']

    o = Obj()
    o.x = 42
    expect(o.x).to_be(42)
    del o.x
    error = None
    try:
        _ = o.x
    except AttributeError:
        error = True
    expect(error).to_be(True)

test("delete slot attribute", test_delete_slot)

# Test 9: Slots with tuple
def test_slots_tuple():
    class TupleSlots:
        __slots__ = ('a', 'b')

    obj = TupleSlots()
    obj.a = 1
    obj.b = 2
    expect(obj.a + obj.b).to_be(3)

test("slots with tuple", test_slots_tuple)
