from test_framework import test, expect

# Test 1: Basic __del__ called on del statement
def test_basic_del():
    results = []

    class Resource:
        def __del__(self):
            results.append("deleted")

    r = Resource()
    del r
    expect(results).to_be(["deleted"])

test("basic __del__ called on del", test_basic_del)

# Test 2: __del__ receives self
def test_del_receives_self():
    results = []

    class Named:
        def __init__(self, name):
            self.name = name

        def __del__(self):
            results.append(self.name)

    obj = Named("foo")
    del obj
    expect(results).to_be(["foo"])

test("__del__ receives self", test_del_receives_self)

# Test 3: __del__ inherited from parent
def test_inherited_del():
    results = []

    class Base:
        def __del__(self):
            results.append("base_del")

    class Child(Base):
        pass

    c = Child()
    del c
    expect(results).to_be(["base_del"])

test("__del__ inherited from parent", test_inherited_del)

# Test 4: Child overrides __del__
def test_override_del():
    results = []

    class Base:
        def __del__(self):
            results.append("base")

    class Child(Base):
        def __del__(self):
            results.append("child")

    c = Child()
    del c
    expect(results).to_be(["child"])

test("child overrides __del__", test_override_del)

# Test 5: __del__ with global variable
def test_global_del():
    results = []

    class Tracker:
        def __del__(self):
            results.append("tracked")

    def make_and_delete():
        t = Tracker()
        del t

    make_and_delete()
    expect(results).to_be(["tracked"])

test("__del__ in local scope", test_global_del)

# Test 6: __del__ not called on non-instance types
def test_del_non_instance():
    x = [1, 2, 3]
    del x  # Should not crash
    expect(True).to_be(True)

test("del on non-instance does not crash", test_del_non_instance)

# Test 7: __del__ errors are silently ignored
def test_del_errors_ignored():
    class BadDel:
        def __del__(self):
            raise ValueError("oops")

    b = BadDel()
    del b  # Should not raise
    expect(True).to_be(True)

test("__del__ errors silently ignored", test_del_errors_ignored)

# Test 8: Multiple objects deleted
def test_multiple_del():
    results = []

    class Obj:
        def __init__(self, n):
            self.n = n
        def __del__(self):
            results.append(self.n)

    a = Obj(1)
    b = Obj(2)
    c = Obj(3)
    del a
    del b
    del c
    expect(results).to_be([1, 2, 3])

test("multiple objects deleted in order", test_multiple_del)

# Test 9: del at module level (global scope)
deleted_global = []

class GlobalResource:
    def __del__(self):
        deleted_global.append("gone")

gr = GlobalResource()
del gr

def test_global_scope_del():
    expect(deleted_global).to_be(["gone"])

test("__del__ at module/global scope", test_global_scope_del)

# Test 10: __del__ with class that has __init__
def test_del_with_init():
    log = []

    class Lifecycle:
        def __init__(self):
            log.append("init")
        def __del__(self):
            log.append("del")

    obj = Lifecycle()
    del obj
    expect(log).to_be(["init", "del"])

test("__del__ with __init__ lifecycle", test_del_with_init)

# Test 11: del obj.attr calls __del__ on the attribute value
def test_del_attr_calls_del():
    log = []

    class Inner:
        def __init__(self, name):
            self.name = name
        def __del__(self):
            log.append(self.name)

    class Outer:
        def __init__(self):
            self.child = Inner("child")

    o = Outer()
    del o.child
    expect(log).to_be(["child"])

test("del obj.attr calls __del__ on attr value", test_del_attr_calls_del)

# Test 12: del obj.attr does not call __del__ for non-instance attrs
def test_del_attr_non_instance():
    class Container:
        def __init__(self):
            self.data = [1, 2, 3]

    c = Container()
    del c.data
    expect(True).to_be(True)

test("del obj.attr on non-instance attr no crash", test_del_attr_non_instance)

# Test 13: del obj.attr with multiple attributes
def test_del_attr_multiple():
    log = []

    class Resource:
        def __init__(self, name):
            self.name = name
        def __del__(self):
            log.append(self.name)

    class Owner:
        def __init__(self):
            self.a = Resource("first")
            self.b = Resource("second")

    o = Owner()
    del o.a
    del o.b
    expect(log).to_be(["first", "second"])

test("del obj.attr multiple attrs", test_del_attr_multiple)
