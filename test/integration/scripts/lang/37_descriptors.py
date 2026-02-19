from test_framework import test, expect

# Test 1: Data descriptor with __get__ and __set__
def test_data_descriptor():
    class Descriptor:
        def __get__(self, obj, objtype=None):
            if obj is None:
                return self
            return obj._value * 2

        def __set__(self, obj, value):
            obj._value = value

    class MyClass:
        attr = Descriptor()

    obj = MyClass()
    obj.attr = 5
    expect(obj.attr).to_be(10)
    obj.attr = 21
    expect(obj.attr).to_be(42)

test("data descriptor with __get__ and __set__", test_data_descriptor)

# Test 2: Non-data descriptor (only __get__) - instance dict shadows it
def test_non_data_descriptor():
    class NonDataDescriptor:
        def __get__(self, obj, objtype=None):
            return "from descriptor"

    class MyClass:
        attr = NonDataDescriptor()

    obj = MyClass()
    expect(obj.attr).to_be("from descriptor")

    # Instance dict should shadow non-data descriptor
    obj.__dict__["attr"] = "from instance"
    expect(obj.attr).to_be("from instance")

test("non-data descriptor shadowed by instance dict", test_non_data_descriptor)

# Test 3: Data descriptor overrides instance dict
def test_data_descriptor_precedence():
    class DataDescriptor:
        def __get__(self, obj, objtype=None):
            if obj is None:
                return self
            return "from descriptor"

        def __set__(self, obj, value):
            pass

    class MyClass:
        attr = DataDescriptor()

    obj = MyClass()
    obj.__dict__["attr"] = "from instance"
    # Data descriptor (has __set__) takes precedence over instance dict
    expect(obj.attr).to_be("from descriptor")

test("data descriptor overrides instance dict", test_data_descriptor_precedence)

# Test 4: Custom __delete__ descriptor
def test_descriptor_delete():
    deleted_attrs = []

    class Descriptor:
        def __get__(self, obj, objtype=None):
            return "value"

        def __set__(self, obj, value):
            pass

        def __delete__(self, obj):
            deleted_attrs.append("deleted")

    class MyClass:
        attr = Descriptor()

    obj = MyClass()
    del obj.attr
    expect(len(deleted_attrs)).to_be(1)
    expect(deleted_attrs[0]).to_be("deleted")

test("custom __delete__ descriptor", test_descriptor_delete)

# Test 5: Class-level descriptor access invokes __get__(None, cls)
def test_class_level_descriptor():
    class Descriptor:
        def __get__(self, obj, objtype=None):
            if obj is None:
                return "class-level access"
            return "instance access"

    class MyClass:
        attr = Descriptor()

    # Class-level access: __get__(None, MyClass)
    expect(MyClass.attr).to_be("class-level access")
    # Instance access: __get__(instance, MyClass)
    obj = MyClass()
    expect(obj.attr).to_be("instance access")

test("class-level descriptor access", test_class_level_descriptor)

# Test 6: Property descriptor still works (getter, setter, deleter)
def test_property_descriptor():
    class MyClass:
        def __init__(self):
            self._x = 0

        @property
        def x(self):
            return self._x

        @x.setter
        def x(self, value):
            self._x = value * 2

        @x.deleter
        def x(self):
            self._x = -1

    obj = MyClass()
    expect(obj.x).to_be(0)
    obj.x = 5
    expect(obj.x).to_be(10)
    del obj.x
    expect(obj.x).to_be(-1)

test("property descriptor (getter, setter, deleter)", test_property_descriptor)

# Test 7: Descriptor with inheritance
def test_descriptor_inheritance():
    class Descriptor:
        def __get__(self, obj, objtype=None):
            if obj is None:
                return self
            return "base descriptor"

        def __set__(self, obj, value):
            pass

    class Base:
        attr = Descriptor()

    class Child(Base):
        pass

    obj = Child()
    expect(obj.attr).to_be("base descriptor")

test("descriptor with inheritance", test_descriptor_inheritance)

# ============================================================================
# Additional descriptor tests adapted from CPython's Lib/test/test_descr.py
# ============================================================================

# Test 8: Computed attribute descriptor (counter pattern)
def test_computed_attribute_descriptor():
    """Descriptor that increments on each access (from CPython test_descr.py)."""
    class computed_attribute:
        def __init__(self, get_fn, set_fn, del_fn):
            self._get = get_fn
            self._set = set_fn
            self._del = del_fn
        def __get__(self, obj, objtype=None):
            return self._get(obj)
        def __set__(self, obj, value):
            return self._set(obj, value)
        def __delete__(self, obj):
            return self._del(obj)

    class C:
        def __init__(self):
            self._x = 0
        def _get_x(self):
            x = self._x
            self._x = x + 1
            return x
        def _set_x(self, x):
            self._x = x
        def _del_x(self):
            del self._x
        x = computed_attribute(_get_x, _set_x, _del_x)

    a = C()
    expect(a.x).to_be(0)
    expect(a.x).to_be(1)
    expect(a.x).to_be(2)
    a.x = 10
    expect(a.x).to_be(10)
    expect(a.x).to_be(11)

test("computed attribute descriptor (counter)", test_computed_attribute_descriptor)

# Test 9: classmethod as descriptor
def test_classmethod_descriptor():
    """classmethod works as a descriptor through inheritance (from CPython test_descr.py)."""
    class C:
        def foo(*a):
            return a
        goo = classmethod(foo)

    c = C()
    expect(C.goo(1)[1]).to_be(1)
    expect(c.goo(1)[1]).to_be(1)
    # C.goo(1) => (C, 1)
    result = C.goo(1)
    expect(result[0] == C).to_be(True)
    # c.goo(1) => (C, 1) — class, not instance
    result = c.goo(1)
    expect(result[0] == C).to_be(True)

    class D(C):
        pass
    d = D()
    # Inherited classmethod gets D as class
    result = D.goo(1)
    expect(result[0] == D).to_be(True)
    result = d.goo(1)
    expect(result[0] == D).to_be(True)

test("classmethod as descriptor", test_classmethod_descriptor)

# Test 10: staticmethod as descriptor
def test_staticmethod_descriptor():
    """staticmethod works as a descriptor through inheritance (from CPython test_descr.py)."""
    class C:
        def foo(*a):
            return a
        goo = staticmethod(foo)

    c = C()
    # staticmethod doesn't bind any implicit argument
    expect(C.goo(1)).to_be((1,))
    expect(c.goo(1)).to_be((1,))

    class D(C):
        pass
    d = D()
    expect(D.goo(1)).to_be((1,))
    expect(d.goo(1)).to_be((1,))

test("staticmethod as descriptor", test_staticmethod_descriptor)

# Test 11: Dynamic attribute assignment on classes
def test_dynamic_class_attributes():
    """Dynamically assigned class attributes are visible to instances (from CPython test_descr.py)."""
    class C:
        pass
    class D(C):
        pass

    C.foo = 1
    expect(C.foo).to_be(1)
    expect(D.foo).to_be(1)  # Inherited

    a = C()
    C.bar = 42
    expect(a.bar).to_be(42)  # Instance sees new class attr

    # Dynamic method assignment
    C.method = lambda self: 99
    expect(a.method()).to_be(99)

test("dynamic class attributes", test_dynamic_class_attributes)

# Test 12: Data descriptor priority over instance __dict__
def test_data_descriptor_priority_detailed():
    """Data descriptor always wins over instance dict, non-data does not (from CPython test_descr.py)."""
    class DataDesc:
        def __init__(self, name):
            self.name = name
        def __get__(self, obj, objtype=None):
            if obj is None:
                return self
            return "data_" + self.name
        def __set__(self, obj, value):
            pass  # Ignore sets

    class NonDataDesc:
        def __init__(self, name):
            self.name = name
        def __get__(self, obj, objtype=None):
            if obj is None:
                return self
            return "nondata_" + self.name

    class C:
        x = DataDesc("x")
        y = NonDataDesc("y")

    obj = C()
    # Data descriptor wins
    expect(obj.x).to_be("data_x")
    obj.__dict__["x"] = "instance_x"
    expect(obj.x).to_be("data_x")  # Still descriptor

    # Non-data descriptor loses to instance dict
    expect(obj.y).to_be("nondata_y")
    obj.__dict__["y"] = "instance_y"
    expect(obj.y).to_be("instance_y")  # Instance wins

test("data descriptor priority detailed", test_data_descriptor_priority_detailed)

# Test 13: Descriptor __delete__ removes from instance
def test_descriptor_delete_detailed():
    """__delete__ descriptor protocol in detail (from CPython test_descr.py)."""
    log = []

    class LoggingDescriptor:
        def __get__(self, obj, objtype=None):
            log.append("get")
            if obj is None:
                return self
            return getattr(obj, "_val", "unset")
        def __set__(self, obj, value):
            log.append("set")
            obj._val = value
        def __delete__(self, obj):
            log.append("delete")
            if hasattr(obj, "_val"):
                del obj._val

    class C:
        attr = LoggingDescriptor()

    obj = C()
    # Get default
    val = obj.attr
    expect(val).to_be("unset")
    expect("get" in log).to_be(True)

    # Set
    obj.attr = 42
    expect("set" in log).to_be(True)
    val = obj.attr
    expect(val).to_be(42)

    # Delete
    del obj.attr
    expect("delete" in log).to_be(True)
    val = obj.attr
    expect(val).to_be("unset")

test("descriptor __delete__ detailed", test_descriptor_delete_detailed)

# Test 14: Descriptor accessed from class returns descriptor
def test_descriptor_class_access():
    """Non-data descriptor __get__(None, cls) returns descriptor itself (from CPython test_descr.py)."""
    class Desc:
        def __get__(self, obj, objtype=None):
            if obj is None:
                return "class_access"
            return "instance_access"

    class C:
        attr = Desc()

    expect(C.attr).to_be("class_access")
    expect(C().attr).to_be("instance_access")

test("descriptor class vs instance access", test_descriptor_class_access)

# Test 15: Property with getter, setter, deleter (decorator style)
def test_property_decorator_style():
    """Property with all three decorators (from CPython test_descr.py)."""
    class C:
        def __init__(self):
            self._foo = 0

        @property
        def foo(self):
            return self._foo

        @foo.setter
        def foo(self, value):
            self._foo = abs(value)

        @foo.deleter
        def foo(self):
            self._foo = -1

    c = C()
    expect(c.foo).to_be(0)
    c.foo = -42
    expect(c._foo).to_be(42)
    expect(c.foo).to_be(42)
    del c.foo
    expect(c._foo).to_be(-1)

test("property decorator style", test_property_decorator_style)

# Test 16: Descriptor with multiple classes
def test_descriptor_shared_across_classes():
    """Same descriptor instance shared by multiple classes."""
    class Tracker:
        def __init__(self):
            self.access_count = 0
        def __get__(self, obj, objtype=None):
            self.access_count = self.access_count + 1
            if obj is None:
                return self
            return "accessed"
        def __set__(self, obj, value):
            pass

    shared = Tracker()

    class A:
        attr = shared
    class B:
        attr = shared

    a = A()
    b = B()
    a.attr
    b.attr
    a.attr
    # shared.access_count should be 3
    expect(shared.access_count).to_be(3)

test("descriptor shared across classes", test_descriptor_shared_across_classes)

# Test 17: Descriptor with MRO (from CPython test_descr.py)
def test_descriptor_mro():
    """Descriptor resolution follows MRO."""
    class DescA:
        def __get__(self, obj, objtype=None):
            return "A"
        def __set__(self, obj, value):
            pass

    class DescB:
        def __get__(self, obj, objtype=None):
            return "B"
        def __set__(self, obj, value):
            pass

    class Base:
        attr = DescA()

    class Child(Base):
        attr = DescB()

    obj = Child()
    # Child's descriptor should win
    expect(obj.attr).to_be("B")

    # Base's descriptor still active for Base instances
    base_obj = Base()
    expect(base_obj.attr).to_be("A")

test("descriptor MRO resolution", test_descriptor_mro)
