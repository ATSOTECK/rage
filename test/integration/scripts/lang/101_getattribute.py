from test_framework import test, expect

# Test 1: __getattribute__ intercepts all attribute access
class Logged:
    def __init__(self):
        object.__setattr__(self, 'log', [])
        object.__setattr__(self, 'x', 42)

    def __getattribute__(self, name):
        if name == 'log':
            return object.__getattribute__(self, 'log')
        log = object.__getattribute__(self, 'log')
        log.append(name)
        return object.__getattribute__(self, name)

def test_intercepts_all():
    obj = Logged()
    val = obj.x
    expect(val).to_be(42)
    expect(obj.log).to_be(['x'])
test("__getattribute__ intercepts all access", test_intercepts_all)

# Test 2: __getattribute__ can return custom values
class AlwaysHello:
    def __getattribute__(self, name):
        return "hello"

def test_custom_return():
    obj = AlwaysHello()
    expect(obj.anything).to_be("hello")
    expect(obj.foo).to_be("hello")
test("__getattribute__ returns custom values", test_custom_return)

# Test 3: __getattribute__ raising AttributeError falls back to __getattr__
class Fallback:
    def __getattribute__(self, name):
        if name == "dynamic":
            raise AttributeError(name)
        return object.__getattribute__(self, name)

    def __getattr__(self, name):
        return "from_getattr_" + name

def test_fallback_to_getattr():
    obj = Fallback()
    expect(obj.dynamic).to_be("from_getattr_dynamic")
test("AttributeError falls back to __getattr__", test_fallback_to_getattr)

# Test 4: __getattribute__ without __getattr__ raises AttributeError
class StrictAccess:
    def __getattribute__(self, name):
        raise AttributeError(name)

def test_no_getattr_fallback():
    obj = StrictAccess()
    try:
        _ = obj.anything
        expect("should have raised").to_be("AttributeError")
    except AttributeError:
        expect(True).to_be(True)
test("no __getattr__ raises AttributeError", test_no_getattr_fallback)

# Test 5: __getattribute__ is not called for special/dunder methods by the VM
# (In CPython, dunders used by the VM bypass __getattribute__ on the instance)
class Counter:
    def __init__(self):
        object.__setattr__(self, 'count', 0)
        object.__setattr__(self, 'value', 10)

    def __getattribute__(self, name):
        if name == 'count':
            return object.__getattribute__(self, 'count')
        c = object.__getattribute__(self, 'count')
        object.__setattr__(self, 'count', c + 1)
        return object.__getattribute__(self, name)

def test_counting():
    obj = Counter()
    _ = obj.value
    _ = obj.value
    _ = obj.value
    expect(obj.count).to_be(3)
test("__getattribute__ tracks access count", test_counting)

# Test 6: __getattribute__ with inheritance
class Base:
    def __getattribute__(self, name):
        if name == "secret":
            return "base_secret"
        return object.__getattribute__(self, name)

class Child(Base):
    pass

def test_inherited():
    obj = Child()
    expect(obj.secret).to_be("base_secret")
test("__getattribute__ inherited by subclass", test_inherited)

# Test 7: Child can override __getattribute__
class Child2(Base):
    def __getattribute__(self, name):
        if name == "secret":
            return "child_secret"
        return object.__getattribute__(self, name)

def test_override():
    obj = Child2()
    expect(obj.secret).to_be("child_secret")
test("child overrides __getattribute__", test_override)

# Test 8: __getattribute__ with methods
class MethodIntercept:
    def __getattribute__(self, name):
        if name == "greet":
            return lambda: "intercepted"
        return object.__getattribute__(self, name)

    def greet(self):
        return "original"

def test_method_intercept():
    obj = MethodIntercept()
    expect(obj.greet()).to_be("intercepted")
test("__getattribute__ intercepts method access", test_method_intercept)

# Test 9: Non-AttributeError exceptions propagate
class ErrorAccess:
    def __getattribute__(self, name):
        raise ValueError("custom error")

def test_non_attr_error_propagates():
    obj = ErrorAccess()
    try:
        _ = obj.x
        expect("should have raised").to_be("ValueError")
    except ValueError as e:
        expect(str(e)).to_be("custom error")
test("non-AttributeError propagates", test_non_attr_error_propagates)

# Test 10: object.__getattribute__ works directly
class Normal:
    def __init__(self):
        self.x = 99

def test_object_getattribute():
    obj = Normal()
    val = object.__getattribute__(obj, "x")
    expect(val).to_be(99)
test("object.__getattribute__ works directly", test_object_getattribute)
