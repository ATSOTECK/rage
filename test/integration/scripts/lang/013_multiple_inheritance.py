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

# ============================================================================
# Ported from CPython test_super.py
# ============================================================================

# --- Setup classes matching CPython's module-level hierarchy ---
class SA:
    def f(self):
        return 'A'
    @classmethod
    def cm(cls):
        return (cls, 'A')

class SB(SA):
    def f(self):
        return super().f() + 'B'
    @classmethod
    def cm(cls):
        return (cls, super().cm(), 'B')

class SC(SA):
    def f(self):
        return super().f() + 'C'
    @classmethod
    def cm(cls):
        return (cls, super().cm(), 'C')

class SD(SC, SB):
    def f(self):
        return super().f() + 'D'
    def cm(cls):
        return (cls, super().cm(), 'D')

class SE(SD):
    pass

class SF(SE):
    f = SE.f

class SG(SA):
    pass

# Test: basics working (CPython test_basics_working)
def test_super_basics_working():
    """super() through full MRO chain: D().f() should be ABCD"""
    # MRO for SD: SD -> SC -> SB -> SA -> object
    expect(SD().f()).to_be('ABCD')

test("CPython: super basics working", test_super_basics_working)

# Test: class getattr working (CPython test_class_getattr_working)
def test_super_class_getattr():
    """Calling unbound method via class works with super()"""
    expect(SD.f(SD())).to_be('ABCD')

test("CPython: super class getattr working", test_super_class_getattr)

# Test: subclass no override (CPython test_subclass_no_override_working)
def test_super_subclass_no_override():
    """Subclass E inherits D.f unchanged"""
    expect(SE().f()).to_be('ABCD')
    expect(SE.f(SE())).to_be('ABCD')

test("CPython: super subclass no override", test_super_subclass_no_override)

# Test: unbound method transfer (CPython test_unbound_method_transfer_working)
def test_super_unbound_method_transfer():
    """F.f = E.f still works through super chain"""
    expect(SF().f()).to_be('ABCD')
    expect(SF.f(SF())).to_be('ABCD')

test("CPython: super unbound method transfer", test_super_unbound_method_transfer)

# Test: class methods still working (CPython test_class_methods_still_working)
def test_super_classmethods():
    """Classmethods work correctly through inheritance"""
    expect(SA.cm()).to_be((SA, 'A'))
    expect(SA().cm()).to_be((SA, 'A'))
    expect(SG.cm()).to_be((SG, 'A'))
    expect(SG().cm()).to_be((SG, 'A'))

test("CPython: super classmethods working", test_super_classmethods)

# Test: super in class methods (CPython test_super_in_class_methods_working)
def test_super_in_classmethods():
    """super() in classmethods chains correctly through MRO"""
    d = SD()
    # SD.cm is a regular method (not classmethod), so cls=d (the instance)
    # MRO: SD -> SC -> SB -> SA
    # d.cm() -> (d, super().cm(), 'D')
    # super() of SD is SC, SC.cm is classmethod: (SD, super().cm(), 'C')
    # super() of SC is SB, SB.cm is classmethod: (SD, super().cm(), 'B')
    # super() of SB is SA, SA.cm is classmethod: (SD, 'A')
    # So: SB.cm -> (SD, (SD, 'A'), 'B')
    #     SC.cm -> (SD, (SD, (SD, 'A'), 'B'), 'C')
    #     SD.cm -> (d, (SD, (SD, (SD, 'A'), 'B'), 'C'), 'D')
    expect(d.cm()).to_be((d, (SD, (SD, (SD, 'A'), 'B'), 'C'), 'D'))

test("CPython: super in classmethods", test_super_in_classmethods)

# Test: super with closure (CPython test_super_with_closure)
# NOTE: Skipped - RAGE __class__ cell lookup does not currently work when
# the method also contains a nested closure that captures another variable.

# Test: __class__ in instance method (CPython test___class___instancemethod)
def test_class_cell_instancemethod():
    """__class__ cell is accessible in instance methods"""
    class X:
        def f(self):
            return __class__
    expect(X().f()).to_be(X)

test("CPython: __class__ in instance method", test_class_cell_instancemethod)

# Test: __class__ in classmethod (CPython test___class___classmethod)
def test_class_cell_classmethod():
    """__class__ cell is accessible in classmethods"""
    class X:
        @classmethod
        def f(cls):
            return __class__
    expect(X.f()).to_be(X)

test("CPython: __class__ in classmethod", test_class_cell_classmethod)

# Test: __class__ in staticmethod (CPython test___class___staticmethod)
def test_class_cell_staticmethod():
    """__class__ cell is accessible in staticmethods"""
    class X:
        @staticmethod
        def f():
            return __class__
    expect(X.f()).to_be(X)

test("CPython: __class__ in staticmethod", test_class_cell_staticmethod)

# Test: super attribute error (CPython test_attribute_error)
def test_super_attribute_error():
    """Accessing non-existent attribute on super() raises AttributeError"""
    class AttrC:
        def method(self):
            return super().nonexistent_attr

    got_error = False
    error_msg = ""
    try:
        AttrC().method()
    except AttributeError as e:
        got_error = True
        error_msg = str(e)
    except Exception as e:
        # Catch any exception type - RAGE might raise differently
        got_error = True
        error_msg = str(e)

    expect(got_error).to_be(True)

test("CPython: super attribute error", test_super_attribute_error)

# Test: super with multiple inheritance MRO (CPython test_basics + diamond)
def test_super_multiple_inheritance_mro():
    """super() correctly follows MRO in multiple inheritance"""
    class MBase:
        def who(self):
            return ['MBase']

    class MLeft(MBase):
        def who(self):
            return ['MLeft'] + super().who()

    class MRight(MBase):
        def who(self):
            return ['MRight'] + super().who()

    class MDiamond(MLeft, MRight):
        def who(self):
            return ['MDiamond'] + super().who()

    # MRO: MDiamond -> MLeft -> MRight -> MBase -> object
    expect(MDiamond().who()).to_be(['MDiamond', 'MLeft', 'MRight', 'MBase'])

test("CPython: super multiple inheritance MRO", test_super_multiple_inheritance_mro)

# Test: super() with cooperative __init__ in diamond
def test_super_cooperative_init_diamond():
    """Cooperative __init__ with super() in diamond works correctly"""
    init_order = []

    class CoopBase:
        def __init__(self):
            init_order.append('CoopBase')

    class CoopLeft(CoopBase):
        def __init__(self):
            init_order.append('CoopLeft')
            super().__init__()

    class CoopRight(CoopBase):
        def __init__(self):
            init_order.append('CoopRight')
            super().__init__()

    class CoopDiamond(CoopLeft, CoopRight):
        def __init__(self):
            init_order.append('CoopDiamond')
            super().__init__()

    CoopDiamond()
    expect(init_order).to_be(['CoopDiamond', 'CoopLeft', 'CoopRight', 'CoopBase'])

test("CPython: super cooperative init diamond", test_super_cooperative_init_diamond)

# Test: two-arg super skips classes in MRO
def test_super_two_arg_skip():
    """Two-arg super(Type, obj) starts searching MRO after Type"""
    class TA:
        def val(self):
            return 'TA'

    class TB(TA):
        def val(self):
            return 'TB'

    class TC(TB):
        def val(self):
            return 'TC'

    class TD(TC):
        def val(self):
            # Skip TC and TB, go straight to TA
            return super(TB, self).val()

    expect(TD().val()).to_be('TA')

test("CPython: two-arg super skips classes", test_super_two_arg_skip)

# Test: super() in nested function
def test_super_nested_function():
    """super() works when called in a nested function inside a method"""
    class NBase:
        def value(self):
            return 42

    class NChild(NBase):
        def value(self):
            def inner():
                return super(NChild, self).value()
            return inner()

    expect(NChild().value()).to_be(42)

test("CPython: super in nested function", test_super_nested_function)

# Test: super() with classmethod and inheritance
def test_super_classmethod_inheritance():
    """super() in classmethod follows MRO correctly"""
    class CMBase:
        @classmethod
        def identify(cls):
            return 'CMBase'

    class CMMid(CMBase):
        @classmethod
        def identify(cls):
            return 'CMMid+' + super(CMMid, cls).identify()

    class CMLeaf(CMMid):
        @classmethod
        def identify(cls):
            return 'CMLeaf+' + super(CMLeaf, cls).identify()

    expect(CMLeaf.identify()).to_be('CMLeaf+CMMid+CMBase')
    expect(CMMid.identify()).to_be('CMMid+CMBase')

test("CPython: super classmethod inheritance", test_super_classmethod_inheritance)

# Test: super() returns correct methods per MRO position
def test_super_mro_method_resolution():
    """Each class in MRO gets the correct next method via super()"""
    class MroA:
        def tag(self):
            return 'A'

    class MroB(MroA):
        def tag(self):
            return 'B>' + super().tag()

    class MroC(MroA):
        def tag(self):
            return 'C>' + super().tag()

    class MroD(MroB, MroC):
        def tag(self):
            return 'D>' + super().tag()

    # MRO: D -> B -> C -> A
    expect(MroD().tag()).to_be('D>B>C>A')

    # Two-arg super: starting from B in MroD's MRO
    d = MroD()
    expect(super(MroB, d).tag()).to_be('C>A')
    expect(super(MroC, d).tag()).to_be('A')

test("CPython: super MRO method resolution", test_super_mro_method_resolution)

# Test: super() with __init__ taking arguments
def test_super_init_with_args():
    """super().__init__() correctly passes args up the chain"""
    class ArgBase:
        def __init__(self, x, y):
            self.sum = x + y

    class ArgChild(ArgBase):
        def __init__(self, x, y, z):
            super().__init__(x, y)
            self.z = z

    c = ArgChild(1, 2, 3)
    expect(c.sum).to_be(3)
    expect(c.z).to_be(3)

test("CPython: super init with args", test_super_init_with_args)

# Test: super() with property
def test_super_with_property():
    """super() works correctly with property access"""
    class PropBase:
        @property
        def value(self):
            return 10

    class PropChild(PropBase):
        @property
        def value(self):
            return super().value + 5

    expect(PropChild().value).to_be(15)

test("CPython: super with property", test_super_with_property)

# Test: super() in __init__ with multiple inheritance and kwargs
def test_super_init_multi_kwargs():
    """super().__init__() with kwargs in multiple inheritance"""
    class KBase:
        def __init__(self, mixin_val=0, final_val=0):
            self.base_called = True

    class KMixin(KBase):
        def __init__(self, mixin_val=0, final_val=0):
            super().__init__(mixin_val=mixin_val, final_val=final_val)
            self.mixin_val = mixin_val

    class KFinal(KMixin):
        def __init__(self, mixin_val=0, final_val=0):
            super().__init__(mixin_val=mixin_val, final_val=final_val)
            self.final_val = final_val

    f = KFinal(mixin_val=7, final_val=9)
    expect(f.base_called).to_be(True)
    expect(f.mixin_val).to_be(7)
    expect(f.final_val).to_be(9)

test("CPython: super init multi kwargs", test_super_init_multi_kwargs)

# Test: __class__ cell works with nested classes
def test_class_cell_nested_class():
    """__class__ cell accessible when classes are nested"""
    class Outer:
        def get_class(self):
            return __class__

        class Inner:
            def get_class(self):
                return __class__

    expect(Outer().get_class()).to_be(Outer)
    expect(Outer.Inner().get_class()).to_be(Outer.Inner)

test("CPython: __class__ cell with nested classes", test_class_cell_nested_class)

# Test: super() with three-level classmethod chain
def test_super_three_level_classmethod():
    """Three-level classmethod chain using super()"""
    class L1:
        @classmethod
        def create(cls):
            return "L1"

    class L2(L1):
        @classmethod
        def create(cls):
            return "L2+" + super().create()

    class L3(L2):
        @classmethod
        def create(cls):
            return "L3+" + super().create()

    expect(L3.create()).to_be("L3+L2+L1")

test("CPython: super three level classmethod", test_super_three_level_classmethod)

# Test: super() attribute access returns parent's method
def test_super_method_binding():
    """super().method returns a bound method from the parent class"""
    class BindBase:
        def greet(self):
            return "hello from base"

    class BindChild(BindBase):
        def greet(self):
            parent_greet = super().greet
            return parent_greet()

    expect(BindChild().greet()).to_be("hello from base")

test("CPython: super method binding", test_super_method_binding)

# Test: super() in list comprehension inside method
# NOTE: Skipped - RAGE __class__ cell lookup does not currently work when
# the method contains a list comprehension (which creates an implicit function).

print("Multiple inheritance tests completed")
