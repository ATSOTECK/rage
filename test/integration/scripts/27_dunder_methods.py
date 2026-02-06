# Test dunder methods for custom classes

# Test 1: Arithmetic operators
class Number:
    def __init__(self, value):
        self.value = value

    def __add__(self, other):
        if isinstance(other, Number):
            return Number(self.value + other.value)
        return Number(self.value + other)

    def __radd__(self, other):
        return Number(other + self.value)

    def __sub__(self, other):
        if isinstance(other, Number):
            return Number(self.value - other.value)
        return Number(self.value - other)

    def __mul__(self, other):
        if isinstance(other, Number):
            return Number(self.value * other.value)
        return Number(self.value * other)

    def __truediv__(self, other):
        if isinstance(other, Number):
            return Number(self.value / other.value)
        return Number(self.value / other)

    def __floordiv__(self, other):
        if isinstance(other, Number):
            return Number(self.value // other.value)
        return Number(self.value // other)

    def __mod__(self, other):
        if isinstance(other, Number):
            return Number(self.value % other.value)
        return Number(self.value % other)

    def __pow__(self, other):
        if isinstance(other, Number):
            return Number(self.value ** other.value)
        return Number(self.value ** other)

    def __neg__(self):
        return Number(-self.value)

    def __pos__(self):
        return Number(+self.value)

    def __abs__(self):
        if self.value < 0:
            return Number(-self.value)
        return Number(self.value)

    def __repr__(self):
        return f"Number({self.value})"

n1 = Number(10)
n2 = Number(3)

print("Arithmetic tests:")
print((n1 + n2).value)  # 13
print((n1 - n2).value)  # 7
print((n1 * n2).value)  # 30
print((n1 / n2).value)  # 3.333...
print((n1 // n2).value) # 3
print((n1 % n2).value)  # 1
print((n1 ** n2).value) # 1000

# Test reverse operator
print((5 + n2).value)   # 8

# Test unary operators
print((-n1).value)      # -10
print((+n1).value)      # 10
print(abs(Number(-5)).value)  # 5

# Test 2: Comparison operators
class Comparable:
    def __init__(self, value):
        self.value = value

    def __eq__(self, other):
        if isinstance(other, Comparable):
            return self.value == other.value
        return self.value == other

    def __lt__(self, other):
        if isinstance(other, Comparable):
            return self.value < other.value
        return self.value < other

    def __le__(self, other):
        if isinstance(other, Comparable):
            return self.value <= other.value
        return self.value <= other

    def __gt__(self, other):
        if isinstance(other, Comparable):
            return self.value > other.value
        return self.value > other

    def __ge__(self, other):
        if isinstance(other, Comparable):
            return self.value >= other.value
        return self.value >= other

c1 = Comparable(5)
c2 = Comparable(5)
c3 = Comparable(10)

print("\nComparison tests:")
print(c1 == c2)  # True
print(c1 == c3)  # False
print(c1 < c3)   # True
print(c1 > c3)   # False
print(c1 <= c2)  # True
print(c1 >= c2)  # True

# Test 3: Container protocol
class Container:
    def __init__(self):
        self.items = {}

    def __len__(self):
        return len(self.items)

    def __getitem__(self, key):
        return self.items[key]

    def __setitem__(self, key, value):
        self.items[key] = value

    def __delitem__(self, key):
        del self.items[key]

    def __contains__(self, key):
        return key in self.items

c = Container()
c["a"] = 1
c["b"] = 2
c["c"] = 3

print("\nContainer tests:")
print(len(c))        # 3
print(c["a"])        # 1
print(c["b"])        # 2
print("a" in c)      # True
print("z" in c)      # False
del c["c"]
print(len(c))        # 2

# Test 4: Callable protocol
class Callable:
    def __init__(self, multiplier):
        self.multiplier = multiplier

    def __call__(self, x):
        return x * self.multiplier

double = Callable(2)
triple = Callable(3)

print("\nCallable tests:")
print(double(5))    # 10
print(triple(5))    # 15

# Test 5: Boolean protocol
class Truthy:
    def __init__(self, value):
        self.value = value

    def __bool__(self):
        return self.value > 0

class LenBased:
    def __init__(self, items):
        self.items = items

    def __len__(self):
        return len(self.items)

print("\nBoolean tests:")
print(bool(Truthy(1)))   # True
print(bool(Truthy(0)))   # False
print(bool(Truthy(-1)))  # False

print(bool(LenBased([1, 2, 3])))  # True
print(bool(LenBased([])))         # False

if Truthy(5):
    print("truthy")
else:
    print("falsy")

# Test 6: String representation
class Point:
    def __init__(self, x, y):
        self.x = x
        self.y = y

    def __str__(self):
        return f"Point({self.x}, {self.y})"

    def __repr__(self):
        return f"Point(x={self.x}, y={self.y})"

p = Point(3, 4)
print("\nString representation tests:")
print(str(p))   # Point(3, 4)
print(p)        # Point(3, 4) (uses __str__)

# Test 7: Hash support
class HashablePoint:
    def __init__(self, x, y):
        self.x = x
        self.y = y

    def __hash__(self):
        return self.x * 1000 + self.y

    def __eq__(self, other):
        if isinstance(other, HashablePoint):
            return self.x == other.x and self.y == other.y
        return False

h1 = HashablePoint(1, 2)
h2 = HashablePoint(1, 2)
h3 = HashablePoint(3, 4)

print("\nHash tests:")
print(hash(h1) == hash(h2))  # True (same hash)
d = {}
d[h1] = "first"
d[h2] = "second"  # Should overwrite since h1 == h2
print(len(d))  # 1
print(d[h1])   # second

# Test 8: Inheritance and MRO
class Base:
    def __str__(self):
        return "Base"

class Derived(Base):
    pass

class Override(Base):
    def __str__(self):
        return "Override"

print("\nInheritance tests:")
print(str(Base()))     # Base
print(str(Derived()))  # Base (inherited)
print(str(Override())) # Override

# Test 9: Bitwise operators
class BitNum:
    def __init__(self, value):
        self.value = value

    def __and__(self, other):
        if isinstance(other, BitNum):
            return BitNum(self.value & other.value)
        return BitNum(self.value & other)

    def __or__(self, other):
        if isinstance(other, BitNum):
            return BitNum(self.value | other.value)
        return BitNum(self.value | other)

    def __xor__(self, other):
        if isinstance(other, BitNum):
            return BitNum(self.value ^ other.value)
        return BitNum(self.value ^ other)

    def __invert__(self):
        return BitNum(~self.value)

b1 = BitNum(0b1100)
b2 = BitNum(0b1010)

print("\nBitwise tests:")
print((b1 & b2).value)   # 8 (0b1000)
print((b1 | b2).value)   # 14 (0b1110)
print((b1 ^ b2).value)   # 6 (0b0110)
print((~BitNum(0)).value)  # -1

print("\nAll tests passed!")
