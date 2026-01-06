# Test: Classes
# Tests class definitions, instances, methods, inheritance

from test_framework import test, expect

# Define all classes at module level

class Point:
    def __init__(self, x, y):
        self.x = x
        self.y = y
    def sum(self):
        return self.x + self.y
    def scale(self, factor):
        return Point(self.x * factor, self.y * factor)

class Empty:
    pass

class Counter:
    count = 0
    def __init__(self, name):
        self.name = name
    def get_name(self):
        return self.name

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

class Calculator:
    def __init__(self, value):
        self.value = value
    def add(self, n):
        return self.value + n
    def multiply(self, n):
        return self.value * n

class Builder:
    def __init__(self):
        self.parts = []
    def add(self, part):
        self.parts.append(part)
        return self
    def get_parts(self):
        return self.parts

class Account:
    def __init__(self, balance):
        self.balance = balance
    def deposit(self, amount):
        self.balance = self.balance + amount
    def withdraw(self, amount):
        self.balance = self.balance - amount
    def get_balance(self):
        return self.balance

class Flyable:
    def fly(self):
        return "Flying"

class Swimmable:
    def swim(self):
        return "Swimming"

class Duck(Flyable, Swimmable):
    def quack(self):
        return "Quack!"

def test_basic_class():
    p = Point(3, 4)
    expect(p.x).to_be(3)
    expect(p.y).to_be(4)
    expect(p.sum()).to_be(7)
    p2 = p.scale(2)
    expect(p2.x).to_be(6)
    expect(p2.y).to_be(8)

def test_multiple_instances():
    p3 = Point(10, 20)
    p4 = Point(100, 200)
    expect(p3.x).to_be(10)
    expect(p4.x).to_be(100)
    expect(p3.sum()).to_be(30)
    expect(p4.sum()).to_be(300)

def test_empty_class():
    e = Empty()
    expect(True).to_be(True)  # Just verify no error

def test_counter():
    c1 = Counter("first")
    c2 = Counter("second")
    expect(c1.name).to_be("first")
    expect(c2.name).to_be("second")

def test_inheritance():
    dog = Dog("Buddy")
    cat = Cat("Whiskers")
    expect(dog.name).to_be("Buddy")
    expect(dog.speak()).to_be("Woof!")
    expect(dog.fetch()).to_be("Buddy fetches the ball")
    expect(cat.name).to_be("Whiskers")
    expect(cat.speak()).to_be("Meow!")

def test_calculator():
    calc = Calculator(10)
    expect(calc.add(5)).to_be(15)
    expect(calc.multiply(3)).to_be(30)

def test_method_chaining():
    builder = Builder()
    builder.add("a").add("b").add("c")
    expect(builder.get_parts()).to_be(["a", "b", "c"])

def test_state_modification():
    acc = Account(100)
    expect(acc.get_balance()).to_be(100)
    acc.deposit(50)
    expect(acc.get_balance()).to_be(150)
    acc.withdraw(30)
    expect(acc.get_balance()).to_be(120)

def test_multiple_inheritance():
    duck = Duck()
    expect(duck.fly()).to_be("Flying")
    expect(duck.swim()).to_be("Swimming")
    expect(duck.quack()).to_be("Quack!")

test("basic_class", test_basic_class)
test("multiple_instances", test_multiple_instances)
test("empty_class", test_empty_class)
test("counter", test_counter)
test("inheritance", test_inheritance)
test("calculator", test_calculator)
test("method_chaining", test_method_chaining)
test("state_modification", test_state_modification)
test("multiple_inheritance", test_multiple_inheritance)

print("Classes tests completed")
