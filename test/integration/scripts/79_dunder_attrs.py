from test_framework import test, expect

# Test __getattr__ as fallback for missing attributes
class Dynamic:
    def __init__(self):
        self.existing = "I exist"

    def __getattr__(self, name):
        return f"dynamic_{name}"

obj = Dynamic()
test("__getattr__ not called for existing attrs", lambda: expect(obj.existing).to_equal("I exist"))
test("__getattr__ called for missing attrs", lambda: expect(obj.missing).to_equal("dynamic_missing"))
test("__getattr__ with different name", lambda: expect(obj.foo).to_equal("dynamic_foo"))

# Test __setattr__ intercepting all assignments
class Logged:
    def __init__(self):
        object.__setattr__(self, 'log', [])

    def __setattr__(self, name, value):
        self.log.append(name)
        object.__setattr__(self, name, value)

obj2 = Logged()
obj2.x = 1
obj2.y = 2
test("__setattr__ intercepts init and post-init", lambda: expect(len(obj2.log)).to_equal(2))
test("__setattr__ logs correct names", lambda: expect(obj2.log).to_equal(["x", "y"]))
test("__setattr__ still sets value via object.__setattr__", lambda: expect(obj2.x).to_equal(1))

# Test object.__setattr__ as bypass mechanism
class Strict:
    def __setattr__(self, name, value):
        if name.startswith("_"):
            raise AttributeError(f"cannot set private attribute {name}")
        object.__setattr__(self, name, value)

obj3 = Strict()
obj3.public = 42
test("__setattr__ allows public attrs", lambda: expect(obj3.public).to_equal(42))

try:
    obj3._private = 1
    test("__setattr__ blocks private attrs", lambda: expect(True).to_equal(False))
except AttributeError:
    test("__setattr__ blocks private attrs", lambda: expect(True).to_equal(True))

# Test __delattr__ intercepting del obj.attr
class TrackDel:
    def __init__(self):
        object.__setattr__(self, 'deleted', [])
        object.__setattr__(self, 'x', 10)
        object.__setattr__(self, 'y', 20)

    def __delattr__(self, name):
        self.deleted.append(name)
        object.__delattr__(self, name)

obj4 = TrackDel()
del obj4.x
test("__delattr__ intercepts del", lambda: expect(obj4.deleted).to_equal(["x"]))
test("__delattr__ actually deletes via object.__delattr__", lambda: expect(obj4.y).to_equal(20))

try:
    _ = obj4.x
    test("deleted attr raises AttributeError", lambda: expect(True).to_equal(False))
except AttributeError:
    test("deleted attr raises AttributeError", lambda: expect(True).to_equal(True))

# Test delattr() builtin triggers __delattr__
class TrackDel2:
    def __init__(self):
        object.__setattr__(self, 'deleted', [])
        object.__setattr__(self, 'a', 1)

    def __delattr__(self, name):
        self.deleted.append(name)
        object.__delattr__(self, name)

obj5 = TrackDel2()
delattr(obj5, 'a')
test("delattr() builtin triggers __delattr__", lambda: expect(obj5.deleted).to_equal(["a"]))

# Test hasattr with __getattr__
class HasAttrTest:
    def __init__(self):
        self.real = True

    def __getattr__(self, name):
        if name == "virtual":
            return True
        raise AttributeError(name)

obj6 = HasAttrTest()
test("hasattr True for real attr", lambda: expect(hasattr(obj6, "real")).to_equal(True))
test("hasattr True for __getattr__ attr", lambda: expect(hasattr(obj6, "virtual")).to_equal(True))
test("hasattr False when __getattr__ raises", lambda: expect(hasattr(obj6, "nope")).to_equal(False))

# Test getattr with default + __getattr__
test("getattr finds real attr", lambda: expect(getattr(obj6, "real")).to_equal(True))
test("getattr finds __getattr__ attr", lambda: expect(getattr(obj6, "virtual")).to_equal(True))
test("getattr uses default when __getattr__ raises", lambda: expect(getattr(obj6, "nope", "default")).to_equal("default"))

# Test __setattr__ in __init__ intercepts self.x = val
class InitIntercept:
    def __init__(self):
        self.count = 0
        self.x = 10

    def __setattr__(self, name, value):
        if name == "count":
            object.__setattr__(self, name, value)
        else:
            object.__setattr__(self, "count", self.count + 1)
            object.__setattr__(self, name, value)

obj7 = InitIntercept()
obj7.y = 20
test("__setattr__ intercepts during __init__", lambda: expect(obj7.count).to_equal(2))
test("__setattr__ intercepts post-init", lambda: expect(obj7.x).to_equal(10))

# Test inheritance of __getattr__
class Base:
    def __getattr__(self, name):
        return "from_base"

class Child(Base):
    pass

obj8 = Child()
test("__getattr__ inherited from base", lambda: expect(obj8.anything).to_equal("from_base"))
