from test_framework import test, expect

# Test 1: Basic __set_name__ is called during class creation
def test_basic_set_name():
    class Descriptor:
        def __set_name__(self, owner, name):
            self.owner = owner
            self.name = name

    class MyClass:
        attr = Descriptor()

    expect(MyClass.attr.name).to_be("attr")
    expect(MyClass.attr.owner).to_be(MyClass)

test("basic __set_name__ called during class creation", test_basic_set_name)

# Test 2: __set_name__ with multiple descriptors
def test_multiple_descriptors():
    names = []

    class Descriptor:
        def __set_name__(self, owner, name):
            self.name = name
            names.append(name)

    class MyClass:
        x = Descriptor()
        y = Descriptor()
        z = Descriptor()

    expect(MyClass.x.name).to_be("x")
    expect(MyClass.y.name).to_be("y")
    expect(MyClass.z.name).to_be("z")
    expect(len(names)).to_be(3)

test("__set_name__ with multiple descriptors", test_multiple_descriptors)

# Test 3: __set_name__ combined with __get__/__set__
def test_set_name_with_data_descriptor():
    class Field:
        def __set_name__(self, owner, name):
            self.attr_name = "_" + name

        def __get__(self, obj, objtype=None):
            if obj is None:
                return self
            return getattr(obj, self.attr_name, None)

        def __set__(self, obj, value):
            setattr(obj, self.attr_name, value)

    class Person:
        name = Field()
        age = Field()

    p = Person()
    p.name = "Alice"
    p.age = 30
    expect(p.name).to_be("Alice")
    expect(p.age).to_be(30)
    expect(p._name).to_be("Alice")
    expect(p._age).to_be(30)

test("__set_name__ combined with data descriptor", test_set_name_with_data_descriptor)

# Test 4: __set_name__ receives the correct owner class
def test_set_name_owner():
    owners = []

    class Descriptor:
        def __set_name__(self, owner, name):
            owners.append(owner)

    class A:
        x = Descriptor()

    class B:
        y = Descriptor()

    expect(owners[0]).to_be(A)
    expect(owners[1]).to_be(B)

test("__set_name__ receives correct owner class", test_set_name_owner)

# Test 5: __set_name__ not called on non-descriptor objects
def test_set_name_only_on_instances():
    # Regular values in class dict should not trigger __set_name__
    class MyClass:
        x = 42
        y = "hello"
        z = [1, 2, 3]

    expect(MyClass.x).to_be(42)
    expect(MyClass.y).to_be("hello")

test("__set_name__ not called on regular values", test_set_name_only_on_instances)

# Test 6: __set_name__ with inheritance - descriptors in parent are NOT re-called
def test_set_name_inheritance():
    calls = []

    class Descriptor:
        def __set_name__(self, owner, name):
            calls.append((owner, name))
            self.name = name

    class Base:
        attr = Descriptor()

    class Child(Base):
        pass

    # __set_name__ should only be called once for Base, not again for Child
    expect(len(calls)).to_be(1)

test("__set_name__ not re-called for inherited descriptors", test_set_name_inheritance)

# Test 7: __set_name__ error propagates
def test_set_name_error():
    class BadDescriptor:
        def __set_name__(self, owner, name):
            raise ValueError("bad descriptor")

    try:
        class MyClass:
            x = BadDescriptor()
        expect("should not reach here").to_be("unreachable")
    except Exception as e:
        expect("bad descriptor" in str(e)).to_be(True)

test("__set_name__ error propagates", test_set_name_error)

# Test 8: __set_name__ called before __init_subclass__
def test_set_name_before_init_subclass():
    order = []

    class Descriptor:
        def __set_name__(self, owner, name):
            order.append("set_name")

    class Base:
        def __init_subclass__(cls, **kwargs):
            order.append("init_subclass")

    class Child(Base):
        attr = Descriptor()

    expect(order[0]).to_be("set_name")
    expect(order[1]).to_be("init_subclass")

test("__set_name__ called before __init_subclass__", test_set_name_before_init_subclass)
