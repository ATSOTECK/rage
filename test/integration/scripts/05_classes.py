# Test: Classes
# Tests class definitions, instances, methods, inheritance

results = {}

# =====================================
# Basic Class Definition
# =====================================

class Point:
    def __init__(self, x, y):
        self.x = x
        self.y = y

    def sum(self):
        return self.x + self.y

    def scale(self, factor):
        return Point(self.x * factor, self.y * factor)

# Create instance
p = Point(3, 4)
results["class_attr_x"] = p.x
results["class_attr_y"] = p.y
results["class_method_sum"] = p.sum()

# Method with return value that creates new instance
p2 = p.scale(2)
results["class_scale_x"] = p2.x
results["class_scale_y"] = p2.y

# =====================================
# Multiple Instances
# =====================================

p3 = Point(10, 20)
p4 = Point(100, 200)

# Each instance has separate state
results["multi_instance_p3_x"] = p3.x
results["multi_instance_p4_x"] = p4.x
results["multi_instance_p3_sum"] = p3.sum()
results["multi_instance_p4_sum"] = p4.sum()

# =====================================
# Class with No __init__
# =====================================

class Empty:
    pass

e = Empty()
results["empty_class_created"] = True

# =====================================
# Class Attributes vs Instance Attributes
# =====================================

class Counter:
    count = 0  # Class attribute

    def __init__(self, name):
        self.name = name  # Instance attribute

    def get_name(self):
        return self.name

c1 = Counter("first")
c2 = Counter("second")
results["counter_c1_name"] = c1.name
results["counter_c2_name"] = c2.name

# =====================================
# Simple Inheritance
# =====================================

class Animal:
    def __init__(self, name):
        self.name = name

    def speak(self):
        return "Some sound"

class Dog(Animal):
    def speak(self):
        return "Woof!"

    def fetch(self):
        return self.name + " fetches the ball"

class Cat(Animal):
    def speak(self):
        return "Meow!"

# Create instances
dog = Dog("Buddy")
cat = Cat("Whiskers")

results["dog_name"] = dog.name
results["dog_speak"] = dog.speak()
results["dog_fetch"] = dog.fetch()

results["cat_name"] = cat.name
results["cat_speak"] = cat.speak()

# =====================================
# Calling Methods That Return Values
# =====================================

class Calculator:
    def __init__(self, value):
        self.value = value

    def add(self, n):
        return self.value + n

    def multiply(self, n):
        return self.value * n

calc = Calculator(10)
results["calc_add_5"] = calc.add(5)
results["calc_multiply_3"] = calc.multiply(3)

# =====================================
# Method Chaining (via returning self)
# =====================================

class Builder:
    def __init__(self):
        self.parts = []

    def add(self, part):
        self.parts.append(part)
        return self

    def get_parts(self):
        return self.parts

builder = Builder()
builder.add("a").add("b").add("c")
results["builder_parts"] = builder.get_parts()

# =====================================
# Modifying Instance State
# =====================================

class Account:
    def __init__(self, balance):
        self.balance = balance

    def deposit(self, amount):
        self.balance = self.balance + amount

    def withdraw(self, amount):
        self.balance = self.balance - amount

    def get_balance(self):
        return self.balance

acc = Account(100)
results["account_initial"] = acc.get_balance()
acc.deposit(50)
results["account_after_deposit"] = acc.get_balance()
acc.withdraw(30)
results["account_after_withdraw"] = acc.get_balance()

# =====================================
# Multiple Inheritance - Basic
# =====================================

class Flyable:
    def fly(self):
        return "Flying"

class Swimmable:
    def swim(self):
        return "Swimming"

class Duck(Flyable, Swimmable):
    def quack(self):
        return "Quack!"

duck = Duck()
results["mi_duck_fly"] = duck.fly()
results["mi_duck_swim"] = duck.swim()
results["mi_duck_quack"] = duck.quack()

# =====================================
# Multiple Inheritance - Diamond Problem
# =====================================

class Base:
    def method(self):
        return "Base"

class Left(Base):
    def method(self):
        return "Left"

class Right(Base):
    def method(self):
        return "Right"

class Diamond(Left, Right):
    pass

diamond = Diamond()
# MRO should be [Diamond, Left, Right, Base], so method returns "Left"
results["mi_diamond_method"] = diamond.method()

# Diamond with no override in Left
class Left2(Base):
    pass  # No override

class Right2(Base):
    def method(self):
        return "Right2"

class Diamond2(Left2, Right2):
    pass

diamond2 = Diamond2()
# MRO should be [Diamond2, Left2, Right2, Base], Left2 has no method, so Right2's is used
results["mi_diamond2_method"] = diamond2.method()

# =====================================
# Multiple Inheritance - Method Resolution Order
# =====================================

class A:
    value = "A"

class B(A):
    value = "B"

class C(A):
    value = "C"

class D(B, C):
    pass

class E(C, B):
    pass

# D(B, C): MRO = [D, B, C, A], so value = "B"
# E(C, B): MRO = [E, C, B, A], so value = "C"
results["mi_mro_d_value"] = D.value
results["mi_mro_e_value"] = E.value

# =====================================
# Multiple Inheritance - Mixin Pattern
# =====================================

class JsonMixin:
    def to_json(self):
        return "json:" + self.data

class XmlMixin:
    def to_xml(self):
        return "xml:" + self.data

class DataModel:
    def __init__(self, data):
        self.data = data

class MyModel(DataModel, JsonMixin, XmlMixin):
    pass

model = MyModel("test")
results["mi_mixin_json"] = model.to_json()
results["mi_mixin_xml"] = model.to_xml()
results["mi_mixin_data"] = model.data

# =====================================
# Multiple Inheritance - Three Bases
# =====================================

class Printable:
    def show(self):
        return "Printable"

class Saveable:
    def save(self):
        return "Saved"

class Loadable:
    def load(self):
        return "Loaded"

class Document(Printable, Saveable, Loadable):
    def __init__(self, content):
        self.content = content

doc = Document("Hello")
results["mi_three_show"] = doc.show()
results["mi_three_save"] = doc.save()
results["mi_three_load"] = doc.load()
results["mi_three_content"] = doc.content

# =====================================
# Multiple Inheritance - Instance Attributes
# =====================================

class NameMixin:
    def set_name(self, name):
        self.name = name
        return self

class AgeMixin:
    def set_age(self, age):
        self.age = age
        return self

class Person(NameMixin, AgeMixin):
    def __init__(self):
        self.name = ""
        self.age = 0

person = Person()
person.set_name("Alice").set_age(30)
results["mi_person_name"] = person.name
results["mi_person_age"] = person.age

# =====================================
# C3 MRO Linearization Tests
# =====================================

# Test __mro__ attribute
class MroBase:
    pass

class MroChild(MroBase):
    pass

# __mro__ should include class, base, and object
mro_child = MroChild.__mro__
results["mro_length"] = len(mro_child)
results["mro_first_is_child"] = mro_child[0] is MroChild
results["mro_second_is_base"] = mro_child[1] is MroBase
results["mro_last_is_object"] = mro_child[-1] is object

# Test __bases__ attribute
results["bases_child_length"] = len(MroChild.__bases__)
results["bases_child_first"] = MroChild.__bases__[0] is MroBase
results["bases_empty_class"] = len(Empty.__bases__)  # Should be 1 (object)
results["bases_empty_is_object"] = Empty.__bases__[0] is object

# Test diamond inheritance MRO
class DiamondTop:
    def value(self):
        return "top"

class DiamondLeft(DiamondTop):
    def value(self):
        return "left"

class DiamondRight(DiamondTop):
    def value(self):
        return "right"

class DiamondBottom(DiamondLeft, DiamondRight):
    pass

# MRO should be: DiamondBottom -> DiamondLeft -> DiamondRight -> DiamondTop -> object
diamond_mro = DiamondBottom.__mro__
results["diamond_mro_length"] = len(diamond_mro)
results["diamond_mro_0"] = diamond_mro[0] is DiamondBottom
results["diamond_mro_1"] = diamond_mro[1] is DiamondLeft
results["diamond_mro_2"] = diamond_mro[2] is DiamondRight
results["diamond_mro_3"] = diamond_mro[3] is DiamondTop
results["diamond_mro_4"] = diamond_mro[4] is object

# Method resolution follows MRO
db = DiamondBottom()
results["diamond_method_resolution"] = db.value()

# =====================================
# super() with C3 MRO
# =====================================

class SuperBase:
    def greet(self):
        return "Base"

class SuperLeft(SuperBase):
    def greet(self):
        return "Left->" + super(SuperLeft, self).greet()

class SuperRight(SuperBase):
    def greet(self):
        return "Right->" + super(SuperRight, self).greet()

class SuperChild(SuperLeft, SuperRight):
    def greet(self):
        return "Child->" + super(SuperChild, self).greet()

sc = SuperChild()
# MRO: SuperChild -> SuperLeft -> SuperRight -> SuperBase -> object
# So super chain should be: Child->Left->Right->Base
results["super_chain"] = sc.greet()

# =====================================
# Complex MRO with super()
# =====================================

class O:
    def m(self):
        return "O"

class A_mro(O):
    def m(self):
        return "A->" + super(A_mro, self).m()

class B_mro(O):
    def m(self):
        return "B->" + super(B_mro, self).m()

class C_mro(O):
    def m(self):
        return "C->" + super(C_mro, self).m()

class D_mro(A_mro, B_mro):
    def m(self):
        return "D->" + super(D_mro, self).m()

class E_mro(B_mro, C_mro):
    def m(self):
        return "E->" + super(E_mro, self).m()

class F_mro(D_mro, E_mro):
    def m(self):
        return "F->" + super(F_mro, self).m()

f = F_mro()
# Complex MRO should correctly linearize
results["complex_mro_length"] = len(F_mro.__mro__)
results["complex_super_chain"] = f.m()

# =====================================
# Inconsistent MRO Detection
# =====================================

class X:
    pass

class Y:
    pass

class A_bad(X, Y):
    pass

class B_bad(Y, X):
    pass

# Trying to create C(A_bad, B_bad) should fail with TypeError
inconsistent_mro_error = False
try:
    class C_bad(A_bad, B_bad):
        pass
except TypeError:
    inconsistent_mro_error = True

results["inconsistent_mro_detected"] = inconsistent_mro_error

# =====================================
# Class membership in MRO (using 'in')
# =====================================

results["object_in_mro"] = object in DiamondBottom.__mro__
results["top_in_mro"] = DiamondTop in DiamondBottom.__mro__
results["left_in_mro"] = DiamondLeft in DiamondBottom.__mro__
results["right_in_mro"] = DiamondRight in DiamondBottom.__mro__
results["unrelated_not_in_mro"] = MroBase not in DiamondBottom.__mro__

print("Classes tests completed")
