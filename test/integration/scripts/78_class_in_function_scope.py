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

test("method access outer var", test_method_access_outer_var)
test("method references class name", test_method_references_class_name)
test("class body attribute", test_class_body_attribute)
test("deep nesting", test_deep_nesting)
test("multiple outer vars", test_multiple_outer_vars)
test("class constructor in method", test_class_constructor_in_method)
test("class attr shadow", test_class_attr_shadow)
test("class CM reference", test_class_cm_reference)
