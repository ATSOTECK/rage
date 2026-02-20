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

# ============================================================================
# Additional metaclass tests adapted from CPython's Lib/test/test_metaclass.py
# ============================================================================

# Test 7: Metaclass inherited from parent
class InheritedMeta(type):
    def __new__(mcs, name, bases, namespace):
        namespace['meta_tag'] = "from_inherited_meta"
        return super().__new__(mcs, name, bases, namespace)

class BaseMeta(metaclass=InheritedMeta):
    pass

class ChildMeta(BaseMeta):
    pass

def test_metaclass_inheritance():
    """Child class inherits metaclass from parent."""
    expect(BaseMeta.meta_tag).to_be("from_inherited_meta")
    expect(ChildMeta.meta_tag).to_be("from_inherited_meta")

test("metaclass inheritance from parent", test_metaclass_inheritance)

# Test 8: Metaclass with __new__ that modifies class name
class RenamingMeta(type):
    def __new__(mcs, name, bases, namespace):
        namespace['original_name'] = name
        return super().__new__(mcs, "Renamed_" + name, bases, namespace)

class Original(metaclass=RenamingMeta):
    pass

def test_metaclass_rename():
    """Metaclass __new__ can change the class name."""
    expect(Original.__name__).to_be("Renamed_Original")
    expect(Original.original_name).to_be("Original")

test("metaclass __new__ modifies name", test_metaclass_rename)

# Test 9: Metaclass with keyword arguments
class KwargMeta(type):
    def __new__(mcs, name, bases, namespace, **kwargs):
        for k, v in kwargs.items():
            if k != "metaclass":
                namespace[k] = v
        return super().__new__(mcs, name, bases, namespace)

    def __init__(cls, name, bases, namespace, **kwargs):
        super().__init__(name, bases, namespace)

class WithKwargs(metaclass=KwargMeta, x=10, y=20):
    pass

def test_metaclass_kwargs():
    """Metaclass receives keyword arguments from class statement."""
    expect(WithKwargs.x).to_be(10)
    expect(WithKwargs.y).to_be(20)

test("metaclass with keyword arguments", test_metaclass_kwargs)

# Test 10: Metaclass __init__ receives correct arguments
init_log = []

class LoggingInitMeta(type):
    def __init__(cls, name, bases, namespace):
        init_log.append(name)
        super().__init__(name, bases, namespace)

class LoggedA(metaclass=LoggingInitMeta):
    pass

class LoggedB(metaclass=LoggingInitMeta):
    pass

def test_metaclass_init_called():
    """Metaclass __init__ is called for each class creation."""
    expect("LoggedA" in init_log).to_be(True)
    expect("LoggedB" in init_log).to_be(True)

test("metaclass __init__ called for each class", test_metaclass_init_called)

# Test 11: Metaclass that validates class body
class ValidatingMeta(type):
    def __new__(mcs, name, bases, namespace):
        if name != "ValidatedBase" and "required_method" not in namespace:
            raise TypeError(name + " must define required_method")
        return super().__new__(mcs, name, bases, namespace)

class ValidatedBase(metaclass=ValidatingMeta):
    pass

def test_metaclass_validates():
    """Metaclass can validate class body and raise errors."""
    # This should succeed
    class Good(ValidatedBase):
        def required_method(self):
            pass
    expect(hasattr(Good, "required_method")).to_be(True)

    # This should fail
    got_error = False
    try:
        class Bad(ValidatedBase):
            pass
    except TypeError:
        got_error = True
    expect(got_error).to_be(True)

test("metaclass validates class body", test_metaclass_validates)

# Test 12: Metaclass with __call__ that modifies instance
class SingletonMeta(type):
    def __call__(cls, *args, **kwargs):
        if not hasattr(cls, '_instance'):
            cls._instance = super().__call__(*args, **kwargs)
        return cls._instance

class Singleton(metaclass=SingletonMeta):
    def __init__(self):
        self.created = True

def test_singleton_metaclass():
    """Metaclass __call__ can implement singleton pattern."""
    a = Singleton()
    b = Singleton()
    expect(a is b).to_be(True)
    expect(a.created).to_be(True)

test("singleton metaclass pattern", test_singleton_metaclass)

# Test 13: Metaclass __new__ returns instance of different class
class ProxyMeta(type):
    def __new__(mcs, name, bases, namespace):
        if name == "Proxy":
            return super().__new__(mcs, name, bases, namespace)
        # For subclasses, create with extra attribute
        namespace['_proxied'] = True
        return super().__new__(mcs, name, bases, namespace)

class Proxy(metaclass=ProxyMeta):
    _proxied = False

class ProxyChild(Proxy):
    pass

def test_metaclass_new_modification():
    """Metaclass __new__ distinguishes base from child."""
    expect(Proxy._proxied).to_be(False)
    expect(ProxyChild._proxied).to_be(True)

test("metaclass __new__ distinguishes base from child", test_metaclass_new_modification)

# Test 14: type() 3-arg form creates class
def test_type_3arg():
    """type(name, bases, dict) creates a new class."""
    MyDynamic = type("MyDynamic", (object,), {"x": 42, "greet": lambda self: "hello"})
    obj = MyDynamic()
    expect(obj.x).to_be(42)
    expect(obj.greet()).to_be("hello")
    expect(MyDynamic.__name__).to_be("MyDynamic")

test("type() 3-arg class creation", test_type_3arg)

# Test 15: type() 3-arg with inheritance
def test_type_3arg_inheritance():
    """type(name, bases, dict) respects inheritance."""
    class Base:
        def method(self):
            return "base"
    Child = type("Child", (Base,), {"extra": 99})
    obj = Child()
    expect(obj.method()).to_be("base")
    expect(obj.extra).to_be(99)

test("type() 3-arg with inheritance", test_type_3arg_inheritance)
