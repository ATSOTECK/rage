from test_framework import test, expect

# Test 1: Basic __new__ is called before __init__
order = []

class MyClass:
    def __new__(cls, val):
        order.append("new")
        instance = object.__new__(cls)
        return instance

    def __init__(self, val):
        order.append("init")
        self.val = val

obj = MyClass(10)
test("__new__ called before __init__", lambda: expect(order).to_equal(["new", "init"]))
test("__init__ sets attributes", lambda: expect(obj.val).to_equal(10))

# Test 2: Singleton pattern
class Singleton:
    _instance = None

    def __new__(cls):
        if cls._instance is None:
            cls._instance = object.__new__(cls)
        return cls._instance

s1 = Singleton()
s2 = Singleton()
test("singleton returns same instance", lambda: expect(s1 is s2).to_equal(True))

# Test 3: object.__new__(cls) explicit call
class Simple:
    pass

inst = object.__new__(Simple)
test("object.__new__ creates instance", lambda: expect(isinstance(inst, Simple)).to_equal(True))

# Test 4: __new__ with arguments
class Tagged:
    def __new__(cls, tag, value):
        instance = object.__new__(cls)
        instance.tag = tag
        return instance

    def __init__(self, tag, value):
        self.value = value

t = Tagged("hello", 42)
test("__new__ can set attributes", lambda: expect(t.tag).to_equal("hello"))
test("__init__ runs after __new__", lambda: expect(t.value).to_equal(42))

# Test 5: __init__ skipped when __new__ returns non-instance
init_called = False

class Factory:
    def __new__(cls):
        return "not an instance"

    def __init__(self):
        global init_called
        init_called = True

result = Factory()
test("__new__ can return non-instance", lambda: expect(result).to_equal("not an instance"))
test("__init__ skipped for non-instance", lambda: expect(init_called).to_equal(False))

# Test 6: Subclass inherits parent __new__
class Base:
    def __new__(cls):
        instance = object.__new__(cls)
        instance.created_by = "Base.__new__"
        return instance

class Child(Base):
    def __init__(self):
        self.name = "child"

c = Child()
test("subclass uses parent __new__", lambda: expect(c.created_by).to_equal("Base.__new__"))
test("subclass __init__ still runs", lambda: expect(c.name).to_equal("child"))

# Test 7: __new__ returning instance of different class skips __init__
class Other:
    pass

class Redirector:
    def __new__(cls):
        return object.__new__(Other)

    def __init__(self):
        # This should NOT be called since __new__ returns an Other instance
        pass

r = Redirector()
test("__new__ returns different class instance", lambda: expect(isinstance(r, Other)).to_equal(True))
