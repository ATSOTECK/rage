from test_framework import test, expect

# Basic creation and repr
def test_basic_union():
    t = int | str
    expect(repr(t)).to_be("int | str")
test("basic union creation", test_basic_union)

def test_union_with_classes():
    class Foo:
        pass
    class Bar:
        pass
    t = Foo | Bar
    expect(repr(t)).to_be("Foo | Bar")
test("union with user classes", test_union_with_classes)

# __args__ attribute
def test_args():
    t = int | str
    args = t.__args__
    expect(isinstance(args, tuple)).to_be(True)
    expect(len(args)).to_be(2)
test("__args__ attribute", test_args)

# Flattening nested unions
def test_flattening():
    t = int | str | float
    expect(repr(t)).to_be("int | str | float")
    expect(len(t.__args__)).to_be(3)
test("flattening nested unions", test_flattening)

# Deduplication
def test_dedup():
    t = int | str | int
    expect(repr(t)).to_be("int | str")
    expect(len(t.__args__)).to_be(2)
test("deduplication", test_dedup)

# None -> NoneType
def test_none():
    t = int | None
    expect(repr(t)).to_be("int | NoneType")
test("None converts to NoneType", test_none)

# isinstance with union
def test_isinstance():
    expect(isinstance(42, int | str)).to_be(True)
    expect(isinstance("hi", int | str)).to_be(True)
    expect(isinstance(3.14, int | str)).to_be(False)
    expect(isinstance(None, int | None)).to_be(True)
    expect(isinstance(42, int | None)).to_be(True)
test("isinstance with union", test_isinstance)

# isinstance with user classes
def test_isinstance_classes():
    class Animal:
        pass
    class Dog(Animal):
        pass
    class Cat(Animal):
        pass
    class Fish:
        pass
    d = Dog()
    expect(isinstance(d, Dog | Cat)).to_be(True)
    expect(isinstance(d, Cat | Fish)).to_be(False)
    expect(isinstance(d, Animal | Fish)).to_be(True)
test("isinstance with user classes", test_isinstance_classes)

# issubclass with union
def test_issubclass():
    class Animal:
        pass
    class Dog(Animal):
        pass
    class Cat(Animal):
        pass
    expect(issubclass(Dog, Dog | Cat)).to_be(True)
    expect(issubclass(Dog, Animal | Cat)).to_be(True)
test("issubclass with union", test_issubclass)

# Exception matching
def test_exception_matching():
    caught = False
    try:
        raise ValueError("test")
    except ValueError | TypeError:
        caught = True
    expect(caught).to_be(True)
test("exception matching with union", test_exception_matching)

def test_exception_no_match():
    caught_right = False
    try:
        try:
            raise KeyError("test")
        except ValueError | TypeError:
            caught_right = False
    except KeyError:
        caught_right = True
    expect(caught_right).to_be(True)
test("exception union no false match", test_exception_no_match)

# Chaining multiple unions
def test_chaining():
    t = int | str | float | bool
    expect(len(t.__args__)).to_be(4)
test("chaining multiple unions", test_chaining)

# Equality
def test_equality():
    t1 = int | str
    t2 = int | str
    expect(t1 == t2).to_be(True)
test("union equality", test_equality)

def test_inequality():
    t1 = int | str
    t2 = int | float
    expect(t1 == t2).to_be(False)
test("union inequality", test_inequality)

# str() works
def test_str():
    t = int | str
    expect(str(t)).to_be("int | str")
test("str() on union", test_str)

# Bitwise OR still works for ints
def test_int_or():
    expect(5 | 3).to_be(7)
test("int bitwise OR unaffected", test_int_or)
