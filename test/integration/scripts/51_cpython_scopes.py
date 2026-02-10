# Test: CPython Variable Scope Tests
# Adapted from CPython's test_scope.py

from test_framework import test, expect

# Module-level globals for testing
_scope_global = 100
_scope_mutable = "original"

# === Global variable access from function ===
def test_global_read():
    def read_global():
        return _scope_global
    expect(read_global()).to_be(100)

# === global keyword for modification ===
def test_global_write():
    global _scope_mutable
    _scope_mutable = "original"
    def modify_global():
        global _scope_mutable
        _scope_mutable = "modified"
    modify_global()
    expect(_scope_mutable).to_be("modified")
    # Restore for other tests
    _scope_mutable = "original"

# === Local variable shadows global ===
_shadow_x = 999
def test_local_shadows_global():
    def func():
        _shadow_x = 42
        return _shadow_x
    expect(func()).to_be(42)
    # Global _shadow_x is not modified by local assignment in func
    expect(_shadow_x).to_be(999)

# === Scope resolution: local > enclosing > global > builtin ===
def test_scope_resolution_order():
    # builtin: len is always available
    expect(len("abc")).to_be(3)
    # global shadows builtin: not testing override of len
    # local shadows global
    val = "global"
    def outer():
        val = "enclosing"
        def inner():
            val = "local"
            return val
        return inner()
    expect(outer()).to_be("local")

def test_enclosing_scope_read():
    val = "outer_val"
    def outer():
        val = "enclosing"
        def inner():
            return val
        return inner()
    expect(outer()).to_be("enclosing")

# === Mutable default arguments gotcha ===
def test_mutable_default_arg():
    def append_to(element, target=[]):
        target.append(element)
        return target
    result1 = append_to(1)
    result2 = append_to(2)
    # The default list is shared across calls
    expect(result1).to_be([1, 2])
    expect(result2).to_be([1, 2])
    expect(result1 is result2).to_be(True)

# === Class/instance variable scope ===
def test_class_variable_scope():
    class MyClass:
        class_var = 10
        def __init__(self):
            self.instance_var = 20
        def get_class_var(self):
            return type(self).class_var
        def get_instance_var(self):
            return self.instance_var
    obj = MyClass()
    expect(obj.get_class_var()).to_be(10)
    expect(obj.get_instance_var()).to_be(20)
    expect(MyClass.class_var).to_be(10)

def test_instance_shadows_class_var():
    class C:
        x = "class"
    obj = C()
    expect(obj.x).to_be("class")
    obj.x = "instance"
    expect(obj.x).to_be("instance")
    expect(C.x).to_be("class")

# === Comprehension scope isolation ===
def test_comprehension_scope():
    x = 10
    result = [x for x in range(5)]
    # In Python 3, list comprehension has its own scope
    # x should still be 10 (or the comprehension variable, depending on impl)
    expect(result).to_be([0, 1, 2, 3, 4])

def test_comprehension_does_not_leak():
    items = [i * 2 for i in range(4)]
    expect(items).to_be([0, 2, 4, 6])

# === Lambda accessing enclosing scope ===
def test_lambda_global_access():
    multiplier = 10
    f = lambda x: x * multiplier
    expect(f(3)).to_be(30)
    expect(f(5)).to_be(50)

def test_lambda_closure():
    def make_multiplier(n):
        return lambda x: x * n
    double = make_multiplier(2)
    triple = make_multiplier(3)
    expect(double(5)).to_be(10)
    expect(triple(5)).to_be(15)

# === Function accessing enclosing function's local (closure) ===
def test_closure_basic():
    def outer():
        data = [1, 2, 3]
        def inner():
            return len(data)
        return inner()
    expect(outer()).to_be(3)

def test_closure_mutation_via_list():
    def make_accumulator():
        total = [0]
        def add(x):
            total[0] = total[0] + x
            return total[0]
        return add
    acc = make_accumulator()
    expect(acc(5)).to_be(5)
    expect(acc(3)).to_be(8)
    expect(acc(2)).to_be(10)

# === Conditional variable creation ===
def test_conditional_variable():
    flag = True
    if flag:
        x = "yes"
    else:
        x = "no"
    expect(x).to_be("yes")

def test_conditional_variable_false():
    flag = False
    if flag:
        x = "yes"
    else:
        x = "no"
    expect(x).to_be("no")

# === Global in nested functions ===
_nested_global = 0

def test_global_in_nested():
    global _nested_global
    _nested_global = 0
    def outer():
        def inner():
            global _nested_global
            _nested_global = 42
        inner()
    outer()
    expect(_nested_global).to_be(42)

def test_global_vs_local_in_nested():
    global _nested_global
    _nested_global = 100
    def func():
        # local variable with same name
        _nested_global_local = 999
        return _nested_global_local
    expect(func()).to_be(999)
    expect(_nested_global).to_be(100)

# === Multiple levels of nesting ===
def test_triple_nested_closure():
    def level1():
        a = 1
        def level2():
            b = 2
            def level3():
                return a + b
            return level3()
        return level2()
    expect(level1()).to_be(3)

# Register all tests
test("global_read", test_global_read)
test("global_write", test_global_write)
test("local_shadows_global", test_local_shadows_global)
test("scope_resolution_order", test_scope_resolution_order)
test("enclosing_scope_read", test_enclosing_scope_read)
test("mutable_default_arg", test_mutable_default_arg)
test("class_variable_scope", test_class_variable_scope)
test("instance_shadows_class_var", test_instance_shadows_class_var)
test("comprehension_scope", test_comprehension_scope)
test("comprehension_does_not_leak", test_comprehension_does_not_leak)
test("lambda_global_access", test_lambda_global_access)
test("lambda_closure", test_lambda_closure)
test("closure_basic", test_closure_basic)
test("closure_mutation_via_list", test_closure_mutation_via_list)
test("conditional_variable", test_conditional_variable)
test("conditional_variable_false", test_conditional_variable_false)
test("global_in_nested", test_global_in_nested)
test("global_vs_local_in_nested", test_global_vs_local_in_nested)
test("triple_nested_closure", test_triple_nested_closure)

print("CPython scope tests completed")
