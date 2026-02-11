from test_framework import test, expect

# === Basic: method accessing enclosing function variable ===
def test_method_access_outer_var():
    x = 10
    class MyClass:
        def get_x(self):
            return x
    obj = MyClass()
    expect(obj.get_x()).to_be(10)

# === Method referencing class name ===
def test_method_references_class_name():
    class Greeter:
        greeting = "hello"
        def greet(self):
            return Greeter.greeting
    g = Greeter()
    expect(g.greet()).to_be("hello")

# === Class body assignment creates attribute, not outer modification ===
def test_class_body_attribute():
    x = 100
    class MyClass:
        x = 42
    expect(x).to_be(100)
    expect(MyClass.x).to_be(42)

# === Deep nesting: function -> function -> class -> method ===
def test_deep_nesting():
    def outer():
        val = "deep"
        class Inner:
            def get_val(self):
                return val
        return Inner()
    obj = outer()
    expect(obj.get_val()).to_be("deep")

# === Multiple methods accessing different outer variables ===
def test_multiple_outer_vars():
    a = 1
    b = 2
    class Multi:
        def get_a(self):
            return a
        def get_b(self):
            return b
        def get_sum(self):
            return a + b
    m = Multi()
    expect(m.get_a()).to_be(1)
    expect(m.get_b()).to_be(2)
    expect(m.get_sum()).to_be(3)

# === Class name used to construct new instance from method ===
def test_class_constructor_in_method():
    class Vector:
        def __init__(self, x, y):
            self.x = x
            self.y = y
        def __add__(self, other):
            return Vector(self.x + other.x, self.y + other.y)
    v1 = Vector(1, 2)
    v2 = Vector(3, 4)
    v3 = v1 + v2
    expect(v3.x).to_be(4)
    expect(v3.y).to_be(6)

# === Class attribute shadowing outer variable ===
def test_class_attr_shadow():
    name = "outer"
    class MyClass:
        name = "class_attr"
        def get_outer_name(self):
            return name
        def get_class_name(self):
            return MyClass.name
    obj = MyClass()
    expect(obj.get_outer_name()).to_be("outer")
    expect(obj.get_class_name()).to_be("class_attr")
    expect(name).to_be("outer")

# === Class referencing self in __enter__/__exit__ ===
def test_class_cm_reference():
    class CountCM:
        count = 0
        def __enter__(self):
            CountCM.count = CountCM.count + 1
            return CountCM.count
        def __exit__(self, exc_type, exc_val, exc_tb):
            return False
    with CountCM() as c1:
        expect(c1).to_be(1)
    with CountCM() as c2:
        expect(c2).to_be(2)

# === CPython: nonlocal in class methods (testNonLocalMethod) ===
def test_nonlocal_in_methods():
    def f(x):
        class c:
            def inc(self):
                nonlocal x
                x = x + 1
                return x
            def dec(self):
                nonlocal x
                x = x - 1
                return x
        return c()
    obj = f(0)
    expect(obj.inc()).to_be(1)
    expect(obj.inc()).to_be(2)
    expect(obj.dec()).to_be(1)
    expect(obj.dec()).to_be(0)

# === CPython: nonlocal in class body (testNonLocalClass) ===
def test_nonlocal_in_class_body():
    def f(x):
        class c:
            nonlocal x
            x = x + 1
            def get(self):
                return x
        return c()
    obj = f(0)
    expect(obj.get()).to_be(1)

# === CPython: free var name collides with method name (testFreeVarInMethod) ===
def test_free_var_method_name_collision():
    def f():
        method_and_var = "var"
        class Test:
            def method_and_var(self):
                return "method"
            def test(self):
                return method_and_var  # enclosing var, not self.method_and_var
        return Test()
    t = f()
    expect(t.test()).to_be("var")
    expect(t.method_and_var()).to_be("method")

# === CPython: variable both bound and free (testBoundAndFree) ===
def test_bound_and_free():
    def f(x):
        class C:
            def m(self):
                return x
            a = x  # class attr from enclosing var
        return C
    C = f(3)
    inst = C()
    expect(inst.a).to_be(3)
    expect(inst.m()).to_be(3)

# === CPython: class body same-name + method reads enclosing (testLocalsClass) ===
def test_class_same_name_method_reads_enclosing():
    def f(x):
        class C:
            x = 12  # class attr, does NOT affect enclosing x
            def m(self):
                return x  # reads enclosing x, not class x
        return C
    C = f(1)
    expect(C.x).to_be(12)
    expect(C().m()).to_be(1)

# === CPython: __call__ accessing enclosing scope (testNestingThroughClass) ===
def test_callable_class_closure():
    def make_adder(x):
        class Adder:
            def __call__(self, y):
                return x + y
        return Adder()
    inc = make_adder(1)
    plus10 = make_adder(10)
    expect(inc(1)).to_be(2)
    expect(plus10(-2)).to_be(8)

# === CPython: comprehension in class body skips class scope ===
def test_comprehension_skips_class_scope():
    def f():
        y = 1
        class C:
            y = 2
            vals = [(x, y) for x in range(2)]
        return C
    C = f()
    expect(C.y).to_be(2)
    # Comprehension sees function's y=1, not class y=2
    expect(C.vals).to_be([(0, 1), (1, 1)])

# === CPython: nonlocal in generator (testNonLocalGenerator) ===
def test_nonlocal_in_generator():
    def f(x):
        def g(y):
            nonlocal x
            for i in range(y):
                x = x + 1
                yield x
        return g
    g = f(0)
    expect(list(g(5))).to_be([1, 2, 3, 4, 5])

test("method access outer var", test_method_access_outer_var)
test("method references class name", test_method_references_class_name)
test("class body attribute", test_class_body_attribute)
test("deep nesting", test_deep_nesting)
test("multiple outer vars", test_multiple_outer_vars)
test("class constructor in method", test_class_constructor_in_method)
test("class attr shadow", test_class_attr_shadow)
test("class CM reference", test_class_cm_reference)
test("nonlocal in methods", test_nonlocal_in_methods)
test("nonlocal in class body", test_nonlocal_in_class_body)
test("free var method name collision", test_free_var_method_name_collision)
test("bound and free", test_bound_and_free)
test("class same name method reads enclosing", test_class_same_name_method_reads_enclosing)
test("callable class closure", test_callable_class_closure)
test("comprehension skips class scope", test_comprehension_skips_class_scope)
test("nonlocal in generator", test_nonlocal_in_generator)
