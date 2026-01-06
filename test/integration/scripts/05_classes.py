# Test: Classes
# Tests class definitions, instances, methods, inheritance

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
    expect(3, p.x)
    expect(4, p.y)
    expect(7, p.sum())
    p2 = p.scale(2)
    expect(6, p2.x)
    expect(8, p2.y)

def test_multiple_instances():
    p3 = Point(10, 20)
    p4 = Point(100, 200)
    expect(10, p3.x)
    expect(100, p4.x)
    expect(30, p3.sum())
    expect(300, p4.sum())

def test_empty_class():
    e = Empty()
    expect(True, True)  # Just verify no error

def test_counter():
    c1 = Counter("first")
    c2 = Counter("second")
    expect("first", c1.name)
    expect("second", c2.name)

def test_inheritance():
    dog = Dog("Buddy")
    cat = Cat("Whiskers")
    expect("Buddy", dog.name)
    expect("Woof!", dog.speak())
    expect("Buddy fetches the ball", dog.fetch())
    expect("Whiskers", cat.name)
    expect("Meow!", cat.speak())

def test_calculator():
    calc = Calculator(10)
    expect(15, calc.add(5))
    expect(30, calc.multiply(3))

def test_method_chaining():
    builder = Builder()
    builder.add("a").add("b").add("c")
    expect(["a", "b", "c"], builder.get_parts())

def test_state_modification():
    acc = Account(100)
    expect(100, acc.get_balance())
    acc.deposit(50)
    expect(150, acc.get_balance())
    acc.withdraw(30)
    expect(120, acc.get_balance())

def test_multiple_inheritance():
    duck = Duck()
    expect("Flying", duck.fly())
    expect("Swimming", duck.swim())
    expect("Quack!", duck.quack())

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
