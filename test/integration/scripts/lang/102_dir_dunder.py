from test_framework import test, expect

# Test 1: dir() on an instance returns instance + class attributes
class Foo:
    class_var = 10
    def method(self):
        pass

def test_dir_instance():
    obj = Foo()
    obj.inst_var = 20
    d = dir(obj)
    expect("inst_var" in d).to_be(True)
    expect("class_var" in d).to_be(True)
    expect("method" in d).to_be(True)
test("dir() includes instance and class attrs", test_dir_instance)

# Test 2: dir() on a class returns class + inherited attributes
class Base:
    base_attr = 1

class Child(Base):
    child_attr = 2

def test_dir_class():
    d = dir(Child)
    expect("child_attr" in d).to_be(True)
    expect("base_attr" in d).to_be(True)
test("dir() on class includes inherited attrs", test_dir_class)

# Test 3: dir() result is sorted
class Multi:
    zebra = 1
    alpha = 2
    middle = 3

def test_dir_sorted():
    d = dir(Multi)
    names = [x for x in d if not x.startswith("_")]
    expect(names).to_be(["alpha", "middle", "zebra"])
test("dir() result is sorted", test_dir_sorted)

# Test 4: Custom __dir__ overrides default
class CustomDir:
    def __dir__(self):
        return ["a", "b", "c"]

def test_custom_dir():
    obj = CustomDir()
    expect(dir(obj)).to_be(["a", "b", "c"])
test("custom __dir__ overrides default", test_custom_dir)

# Test 5: __dir__ returning unsorted list
class UnsortedDir:
    def __dir__(self):
        return ["z", "a", "m"]

def test_unsorted_dir():
    obj = UnsortedDir()
    d = dir(obj)
    expect(d).to_be(["z", "a", "m"])
test("__dir__ result returned as-is", test_unsorted_dir)

# Test 6: __dir__ inherited from parent
class Parent:
    def __dir__(self):
        return ["inherited"]

class ChildDir(Parent):
    pass

def test_inherited_dir():
    obj = ChildDir()
    expect(dir(obj)).to_be(["inherited"])
test("__dir__ inherited from parent", test_inherited_dir)

# Test 7: dir() on a module
import sys

def test_dir_module():
    d = dir(sys)
    expect("path" in d).to_be(True)
    expect("version" in d).to_be(True)
test("dir() on module", test_dir_module)

# Test 8: dir() with no args returns scope names
def test_dir_no_args():
    local_var = 42
    d = dir()
    # Should contain builtin names at minimum
    expect("print" in d).to_be(True)
    expect("len" in d).to_be(True)
test("dir() with no args returns scope names", test_dir_no_args)

# Test 9: dir() includes __init__ and other dunders from object
class Simple:
    pass

def test_dir_includes_dunders():
    d = dir(Simple())
    expect("__setattr__" in d).to_be(True)
    expect("__delattr__" in d).to_be(True)
    expect("__getattribute__" in d).to_be(True)
test("dir() includes dunders from MRO", test_dir_includes_dunders)

# Test 10: dir() on instance with slots
class Slotted:
    __slots__ = ["x", "y"]
    def __init__(self):
        self.x = 1
        self.y = 2

def test_dir_slots():
    obj = Slotted()
    d = dir(obj)
    expect("x" in d).to_be(True)
    expect("y" in d).to_be(True)
test("dir() includes slot attributes", test_dir_slots)
