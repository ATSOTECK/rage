# Test: Multiple Inheritance with C3 Linearization and super()
# Tests MRO computation, cooperative multiple inheritance, and super() functionality

from test_framework import test, expect

# =============================================================================
# Basic Multiple Inheritance Classes
# =============================================================================

class A:
    def method(self):
        return ["A"]

class B(A):
    def method(self):
        return ["B"] + super().method()

class C(A):
    def method(self):
        return ["C"] + super().method()

class D(B, C):
    def method(self):
        return ["D"] + super().method()

# =============================================================================
# Cooperative __init__ Classes
# =============================================================================

class InitBase:
    def __init__(self):
        self.base_init = True

class InitMixin1(InitBase):
    def __init__(self):
        super().__init__()
        self.mixin1_init = True

class InitMixin2(InitBase):
    def __init__(self):
        super().__init__()
        self.mixin2_init = True

class InitFinal(InitMixin1, InitMixin2):
    def __init__(self):
        super().__init__()
        self.final_init = True

# =============================================================================
# Method Chaining Classes
# =============================================================================

class ProcessBase:
    def process(self, value):
        return value

class AddOne(ProcessBase):
    def process(self, value):
        return super().process(value) + 1

class Double(ProcessBase):
    def process(self, value):
        return super().process(value) * 2

class Combined(AddOne, Double):
    def process(self, value):
        return super().process(value)

# =============================================================================
# Complex Diamond Hierarchy
# =============================================================================

class DiamondA:
    def method(self):
        return "A"

class DiamondB(DiamondA):
    def method(self):
        return "B->" + super().method()

class DiamondC(DiamondA):
    def method(self):
        return "C->" + super().method()

class DiamondD(DiamondA):
    def method(self):
        return "D->" + super().method()

class DiamondE(DiamondB, DiamondC):
    def method(self):
        return "E->" + super().method()

class DiamondF(DiamondC, DiamondD):
    def method(self):
        return "F->" + super().method()

class DiamondG(DiamondE, DiamondF):
    def method(self):
        return "G->" + super().method()

# =============================================================================
# Deep Hierarchy Classes
# =============================================================================

class Level0:
    def get_level(self):
        return 0

class Level1(Level0):
    def get_level(self):
        return super().get_level() + 1

class Level2(Level1):
    def get_level(self):
        return super().get_level() + 1

class Level3(Level2):
    def get_level(self):
        return super().get_level() + 1

class Level4(Level3):
    def get_level(self):
        return super().get_level() + 1

# =============================================================================
# Two-argument super() Classes
# =============================================================================

class TwoArgA:
    def method(self):
        return "A"

class TwoArgB(TwoArgA):
    def method(self):
        return "B"

class TwoArgC(TwoArgB):
    def method(self):
        # Skip B, call A's method directly
        return super(TwoArgB, self).method()

# =============================================================================
# C3 Linearization Test Classes (from Python docs)
# =============================================================================

class O:
    pass

class C3A(O):
    pass

class C3B(O):
    pass

class C3C(O):
    pass

class C3D(O):
    pass

class C3E(O):
    pass

class K1(C3A, C3B, C3C):
    pass

class K2(C3D, C3B, C3E):
    pass

class K3(C3D, C3A):
    pass

class Z(K1, K2, K3):
    pass

# =============================================================================
# Mixin Pattern Classes
# =============================================================================

class LoggerMixin:
    def log(self, msg):
        return "LOG: " + msg

class ValidatorMixin:
    def validate(self, value):
        return value > 0

class ServiceBase:
    def __init__(self, name):
        self.name = name

class MyService(ServiceBase, LoggerMixin, ValidatorMixin):
    def process(self, value):
        if self.validate(value):
            return self.log(self.name + " processed " + str(value))
        return self.log(self.name + " rejected " + str(value))

# =============================================================================
# Tests
# =============================================================================

def test_super_zero_arg_diamond():
    """Test zero-argument super() in diamond inheritance"""
    d = D()
    result = d.method()
    # MRO: D -> B -> C -> A -> object
    # Each class adds itself and calls super()
    expect(result).to_be(["D", "B", "C", "A"])

def test_cooperative_init():
    """Test cooperative __init__ with super()"""
    f = InitFinal()
    # All __init__ methods should have been called
    expect(f.base_init).to_be(True)
    expect(f.mixin1_init).to_be(True)
    expect(f.mixin2_init).to_be(True)
    expect(f.final_init).to_be(True)

def test_method_chaining():
    """Test method result chaining through super()"""
    c = Combined()
    # MRO: Combined -> AddOne -> Double -> ProcessBase
    # process(5): AddOne calls super() then adds 1
    # Double calls super() then doubles
    # ProcessBase returns value
    # So: ProcessBase(5)=5, Double(5)=10, AddOne(10)=11
    expect(c.process(5)).to_be(11)

def test_super_two_arg():
    """Test two-argument super(Type, obj) form"""
    c = TwoArgC()
    # super(TwoArgB, self) starts searching after TwoArgB in MRO
    expect(c.method()).to_be("A")

def test_mro_attribute():
    """Test accessing __mro__ attribute"""
    mro = D.__mro__
    # MRO: D -> B -> C -> A -> object (5 classes)
    expect(len(mro)).to_be(5)
    expect(mro[0].__name__).to_be("D")
    expect(mro[1].__name__).to_be("B")
    expect(mro[2].__name__).to_be("C")
    expect(mro[3].__name__).to_be("A")
    expect(mro[4].__name__).to_be("object")

def test_deep_hierarchy():
    """Test super() through deep hierarchy"""
    obj = Level4()
    expect(obj.get_level()).to_be(4)

def test_complex_diamond():
    """Test complex 7-class diamond with full chain"""
    g = DiamondG()
    # MRO: G -> E -> B -> F -> C -> D -> A -> object
    expect(g.method()).to_be("G->E->B->F->C->D->A")

def test_c3_linearization_order():
    """Test C3 linearization produces correct MRO"""
    mro_names = [cls.__name__ for cls in Z.__mro__]
    # C3 linearization for Z(K1, K2, K3) should produce:
    # Z, K1, K2, K3, D, A, B, C, E, O, object
    expected = ["Z", "K1", "K2", "K3", "C3D", "C3A", "C3B", "C3C", "C3E", "O", "object"]
    expect(mro_names).to_be(expected)

def test_mixin_pattern():
    """Test practical mixin pattern usage"""
    svc = MyService("TestService")
    result = svc.process(42)
    expect(result).to_be("LOG: TestService processed 42")

    result2 = svc.process(-1)
    expect(result2).to_be("LOG: TestService rejected -1")

def test_super_with_args():
    """Test super() calling methods with arguments"""
    class Base:
        def calc(self, x, y):
            return x + y

    class Child(Base):
        def calc(self, x, y):
            return super().calc(x, y) * 2

    c = Child()
    expect(c.calc(3, 4)).to_be(14)  # (3 + 4) * 2 = 14

def test_super_multiple_methods():
    """Test super() with multiple methods in same class"""
    class Base:
        def method_a(self):
            return "Base.a"
        def method_b(self):
            return "Base.b"

    class Child(Base):
        def method_a(self):
            return "Child.a:" + super().method_a()
        def method_b(self):
            return "Child.b:" + super().method_b()

    c = Child()
    expect(c.method_a()).to_be("Child.a:Base.a")
    expect(c.method_b()).to_be("Child.b:Base.b")

def test_super_partial_override():
    """Test super() when only some classes override a method"""
    class PA:
        def method(self):
            return ["A"]

    class PB(PA):
        # B does NOT override method
        pass

    class PC(PA):
        def method(self):
            return ["C"] + super().method()

    class PD(PB, PC):
        def method(self):
            return ["D"] + super().method()

    d = PD()
    # MRO: PD -> PB -> PC -> PA
    # PD.method() calls super() -> PB has no method -> PC.method()
    expect(d.method()).to_be(["D", "C", "A"])

def test_inconsistent_mro_error():
    """Test that inconsistent MRO raises TypeError"""
    class IA:
        pass
    class IB:
        pass
    class IX(IA, IB):
        pass
    class IY(IB, IA):
        pass

    error_raised = False
    try:
        class IZ(IX, IY):
            pass
    except TypeError as e:
        error_raised = True
        expect("Cannot create a consistent method resolution order" in str(e)).to_be(True)

    expect(error_raised).to_be(True)

def test_super_in_init_with_kwargs():
    """Test super().__init__ with keyword arguments"""
    class KWBase:
        def __init__(self, name="default"):
            self.name = name

    class KWChild(KWBase):
        def __init__(self, name, age):
            super().__init__(name=name)
            self.age = age

    c = KWChild("Alice", 30)
    expect(c.name).to_be("Alice")
    expect(c.age).to_be(30)

def test_bases_attribute():
    """Test accessing __bases__ attribute"""
    bases = D.__bases__
    expect(len(bases)).to_be(2)
    expect(bases[0].__name__).to_be("B")
    expect(bases[1].__name__).to_be("C")

def test_name_attribute():
    """Test accessing __name__ attribute on classes"""
    expect(D.__name__).to_be("D")
    expect(InitFinal.__name__).to_be("InitFinal")
    expect(DiamondG.__name__).to_be("DiamondG")

# Run all tests
test("super_zero_arg_diamond", test_super_zero_arg_diamond)
test("cooperative_init", test_cooperative_init)
test("method_chaining", test_method_chaining)
test("super_two_arg", test_super_two_arg)
test("mro_attribute", test_mro_attribute)
test("deep_hierarchy", test_deep_hierarchy)
test("complex_diamond", test_complex_diamond)
test("c3_linearization_order", test_c3_linearization_order)
test("mixin_pattern", test_mixin_pattern)
test("super_with_args", test_super_with_args)
test("super_multiple_methods", test_super_multiple_methods)
test("super_partial_override", test_super_partial_override)
test("inconsistent_mro_error", test_inconsistent_mro_error)
test("super_in_init_with_kwargs", test_super_in_init_with_kwargs)
test("bases_attribute", test_bases_attribute)
test("name_attribute", test_name_attribute)

print("Multiple inheritance tests completed")
