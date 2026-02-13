# Test: Metaclasses
# Tests metaclass __new__, __init__, __call__, and type() behavior

from test_framework import test, expect

# Test 1: type(x) 1-arg form still works
class Foo:
    pass

def test_type_1arg():
    f = Foo()
    expect(type(f) == Foo).to_be(True)
    expect(type(42).__name__).to_be("int")
    expect(type("hello").__name__).to_be("str")

test("type(x) returns class of instance", test_type_1arg)

# Test 2: Basic metaclass with __new__ that modifies namespace
class Meta(type):
    def __new__(mcs, name, bases, namespace):
        namespace['added_by_meta'] = True
        return super().__new__(mcs, name, bases, namespace)

class MyClass(metaclass=Meta):
    pass

def test_meta_new():
    expect(MyClass.added_by_meta).to_be(True)

test("metaclass __new__ modifies namespace", test_meta_new)

# Test 3: Metaclass __init__ that sets class attributes
class InitMeta(type):
    def __init__(cls, name, bases, namespace):
        super().__init__(name, bases, namespace)
        cls.initialized = True

class InitClass(metaclass=InitMeta):
    pass

def test_meta_init():
    expect(InitClass.initialized).to_be(True)

test("metaclass __init__ sets class attributes", test_meta_init)

# Test 4: Metaclass __call__ that counts instantiations
instance_count = 0

class CountingMeta(type):
    def __call__(cls, *args, **kwargs):
        global instance_count
        instance_count = instance_count + 1
        return super().__call__(*args, **kwargs)

class Counted(metaclass=CountingMeta):
    pass

def test_meta_call():
    a = Counted()
    b = Counted()
    c = Counted()
    expect(instance_count).to_be(3)

test("metaclass __call__ counts instantiations", test_meta_call)

# Test 5: Metaclass with both __new__ and __init__
class FullMeta(type):
    def __new__(mcs, name, bases, namespace):
        namespace['created'] = True
        return super().__new__(mcs, name, bases, namespace)
    def __init__(cls, name, bases, namespace):
        super().__init__(name, bases, namespace)
        cls.initialized_too = True

class FullClass(metaclass=FullMeta):
    x = 42

def test_full_meta():
    expect(FullClass.created).to_be(True)
    expect(FullClass.initialized_too).to_be(True)
    expect(FullClass.x).to_be(42)

test("metaclass with __new__ and __init__", test_full_meta)

# Test 6: Instances of metaclass-created classes work normally
class NormalMeta(type):
    def __new__(mcs, name, bases, namespace):
        return super().__new__(mcs, name, bases, namespace)

class NormalClass(metaclass=NormalMeta):
    def __init__(self, val):
        self.val = val

def test_normal_instances():
    obj = NormalClass(99)
    expect(obj.val).to_be(99)

test("instances of metaclass-created classes work normally", test_normal_instances)
