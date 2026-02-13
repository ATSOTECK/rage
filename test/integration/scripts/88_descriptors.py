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
