# Test: __init_subclass__
# Tests PEP 487 __init_subclass__ hook for customizing subclass creation

from test_framework import test, expect

# Test 1: Basic hook is called when subclass is created
log = []

class Base:
    def __init_subclass__(cls, **kwargs):
        log.append(cls.__name__)

class Child(Base):
    pass

class GrandChild(Child):
    pass

def test_basic_hook():
    expect(log).to_be(["Child", "GrandChild"])

test("__init_subclass__ called on subclass creation", test_basic_hook)

# Test 2: cls argument is the new subclass
captured_cls = []

class Registry:
    def __init_subclass__(cls, **kwargs):
        captured_cls.append(cls)

class Plugin(Registry):
    pass

def test_cls_is_subclass():
    expect(captured_cls[0]).to_be(Plugin)
    expect(captured_cls[0].__name__).to_be("Plugin")

test("cls argument is the new subclass", test_cls_is_subclass)

# Test 3: kwargs forwarding from class definition
received_kwargs = {}

class Configurable:
    def __init_subclass__(cls, **kwargs):
        for k, v in kwargs.items():
            received_kwargs[k] = v

class Specific(Configurable, color="red", size=42):
    pass

def test_kwargs():
    expect(received_kwargs["color"]).to_be("red")
    expect(received_kwargs["size"]).to_be(42)

test("kwargs forwarded from class definition", test_kwargs)

# Test 4: Chaining with super().__init_subclass__()
chain_log = []

class A:
    def __init_subclass__(cls, **kwargs):
        super().__init_subclass__(**kwargs)
        chain_log.append("A")

class B(A):
    def __init_subclass__(cls, **kwargs):
        super().__init_subclass__(**kwargs)
        chain_log.append("B")

class C(B):
    pass

def test_chaining():
    # When B is created (subclass of A): A.__init_subclass__ called -> ["A"]
    # When C is created (subclass of B): B.__init_subclass__ calls super -> A.__init_subclass__ -> ["A", "A", "B"]
    expect(chain_log).to_be(["A", "A", "B"])

test("chaining with super().__init_subclass__()", test_chaining)

# Test 5: Default behavior (no custom __init_subclass__) doesn't error
class Plain:
    pass

class PlainChild(Plain):
    pass

def test_default():
    p = PlainChild()
    expect(type(p)).to_be(PlainChild)

test("default __init_subclass__ does nothing", test_default)

# Test 6: Plugin registration pattern
class PluginBase:
    plugins = []
    def __init_subclass__(cls, **kwargs):
        super().__init_subclass__(**kwargs)
        PluginBase.plugins.append(cls)

class PluginA(PluginBase):
    pass

class PluginB(PluginBase):
    pass

def test_plugin_pattern():
    expect(len(PluginBase.plugins)).to_be(2)
    expect(PluginBase.plugins[0].__name__).to_be("PluginA")
    expect(PluginBase.plugins[1].__name__).to_be("PluginB")

test("plugin registration pattern", test_plugin_pattern)

# Test 7: __init_subclass__ with class attributes set
attr_log = {}

class AttrBase:
    def __init_subclass__(cls, **kwargs):
        super().__init_subclass__(**kwargs)
        attr_log[cls.__name__] = True

class WithAttr(AttrBase):
    x = 10

def test_attrs_visible():
    expect(attr_log["WithAttr"]).to_be(True)
    expect(WithAttr.x).to_be(10)

test("__init_subclass__ works with class attributes", test_attrs_visible)

# Test 8: Multiple inheritance - hook from first parent in MRO
mi_log = []

class M1:
    def __init_subclass__(cls, **kwargs):
        super().__init_subclass__(**kwargs)
        mi_log.append("M1")

class M2:
    def __init_subclass__(cls, **kwargs):
        super().__init_subclass__(**kwargs)
        mi_log.append("M2")

# M1's __init_subclass__ is found first in MRO
class Combined(M1, M2):
    pass

def test_multiple_inheritance():
    # Combined's MRO: [Combined, M1, M2, object]
    # __init_subclass__ found on M1 (MRO[1]), which calls super -> M2's, which calls super -> object's
    expect("M1" in mi_log).to_be(True)

test("multiple inheritance __init_subclass__", test_multiple_inheritance)

# ============================================================================
# Ported from CPython test_subclassinit.py
# ============================================================================

# Test 9: init_subclass sets attribute on subclass, not parent (CPython test_init_subclass)
def test_init_subclass_basic_cpython():
    """From CPython: A.__init_subclass__ sets initialized=True on subclass only"""
    class ISA:
        initialized = False
        def __init_subclass__(cls):
            super().__init_subclass__()
            cls.initialized = True

    class ISB(ISA):
        pass

    expect(ISA.initialized).to_be(False)
    expect(ISB.initialized).to_be(True)

test("CPython: basic __init_subclass__ sets attr on subclass only", test_init_subclass_basic_cpython)

# Test 10: kwargs stored on subclass (CPython test_init_subclass_kwargs)
def test_init_subclass_kwargs_cpython():
    """From CPython: kwargs from class statement forwarded to __init_subclass__"""
    class KA:
        def __init_subclass__(cls, **kwargs):
            cls.kwargs = kwargs

    class KB(KA, x=3):
        pass

    expect(KB.kwargs).to_be({"x": 3})

test("CPython: kwargs forwarded to __init_subclass__", test_init_subclass_kwargs_cpython)

# Test 11: __init_subclass__ raising an error (CPython test_init_subclass_error)
def test_init_subclass_error():
    """From CPython: error in __init_subclass__ propagates"""
    class EA:
        def __init_subclass__(cls):
            raise RuntimeError("init_subclass error")

    got_error = False
    try:
        class EB(EA):
            pass
    except RuntimeError:
        got_error = True

    expect(got_error).to_be(True)

test("CPython: __init_subclass__ error propagates", test_init_subclass_error)

# Test 12: Wrong parameters to __init_subclass__ (CPython test_init_subclass_wrong)
# NOTE: Skipped - RAGE does not currently raise TypeError for missing required
# __init_subclass__ parameters.

# Test 13: Skipped intermediate classes (CPython test_init_subclass_skipped)
def test_init_subclass_skipped():
    """From CPython: intermediate class without __init_subclass__ still inherits it"""
    class BaseWithInit:
        def __init_subclass__(cls, **kwargs):
            super().__init_subclass__(**kwargs)
            cls.initialized = cls

    class BaseWithoutInit(BaseWithInit):
        pass

    class SkipA(BaseWithoutInit):
        pass

    expect(SkipA.initialized).to_be(SkipA)
    expect(BaseWithoutInit.initialized).to_be(BaseWithoutInit)

test("CPython: __init_subclass__ inherited through intermediate class", test_init_subclass_skipped)

# Test 14: Diamond inheritance with __init_subclass__ (CPython test_init_subclass_diamond)
# Adapted: test the diamond pattern where __init_subclass__ chains through super()
def test_init_subclass_diamond():
    """Diamond __init_subclass__ with super chaining"""
    call_order = []

    class DBase:
        def __init_subclass__(cls, **kwargs):
            super().__init_subclass__(**kwargs)
            call_order.append("DBase:" + cls.__name__)

    class DLeft(DBase):
        def __init_subclass__(cls, **kwargs):
            super().__init_subclass__(**kwargs)
            call_order.append("DLeft:" + cls.__name__)

    # Reset to only capture the diamond subclass creation
    call_order.clear()

    class DDiamond(DLeft):
        pass

    # DLeft.__init_subclass__ calls super -> DBase.__init_subclass__
    expect(call_order).to_be(["DBase:DDiamond", "DLeft:DDiamond"])

test("CPython: diamond __init_subclass__ with super chain", test_init_subclass_diamond)

# Test 15: Multiple kwargs forwarding
def test_init_subclass_multiple_kwargs():
    """Test multiple keyword arguments forwarded through __init_subclass__"""
    class MKBase:
        def __init_subclass__(cls, name=None, version=None, **kwargs):
            super().__init_subclass__(**kwargs)
            cls.name = name
            cls.version = version

    class MKChild(MKBase, name="plugin", version=2):
        pass

    expect(MKChild.name).to_be("plugin")
    expect(MKChild.version).to_be(2)

test("CPython: multiple kwargs in __init_subclass__", test_init_subclass_multiple_kwargs)

# Test 16: __init_subclass__ with super chaining in diamond
def test_init_subclass_super_chain():
    """Test that super().__init_subclass__() chains correctly in diamond"""
    call_order = []

    class ChainA:
        def __init_subclass__(cls, **kwargs):
            super().__init_subclass__(**kwargs)
            call_order.append("ChainA")

    class ChainB(ChainA):
        def __init_subclass__(cls, **kwargs):
            super().__init_subclass__(**kwargs)
            call_order.append("ChainB")

    class ChainC(ChainA):
        def __init_subclass__(cls, **kwargs):
            super().__init_subclass__(**kwargs)
            call_order.append("ChainC")

    # Reset to capture only D's creation
    call_order.clear()

    class ChainD(ChainB, ChainC):
        pass

    # MRO for D: D -> ChainB -> ChainC -> ChainA -> object
    # ChainB.__init_subclass__ calls super -> ChainC, which calls super -> ChainA
    expect(call_order).to_be(["ChainA", "ChainC", "ChainB"])

test("CPython: super chain in diamond __init_subclass__", test_init_subclass_super_chain)

# Test 17: __init_subclass__ not called on the defining class itself
def test_init_subclass_not_called_on_self():
    """__init_subclass__ should not be called when the class itself is defined"""
    call_count = [0]

    class Counter:
        def __init_subclass__(cls, **kwargs):
            super().__init_subclass__(**kwargs)
            call_count[0] += 1

    # Counter was just defined, hook should not have been called
    expect(call_count[0]).to_be(0)

    class Sub1(Counter):
        pass
    expect(call_count[0]).to_be(1)

    class Sub2(Counter):
        pass
    expect(call_count[0]).to_be(2)

test("CPython: __init_subclass__ not called on defining class", test_init_subclass_not_called_on_self)

# Test 18: __init_subclass__ with default kwargs
def test_init_subclass_default_kwargs():
    """Test __init_subclass__ with default keyword argument values"""
    class DefaultBase:
        def __init_subclass__(cls, debug=False, **kwargs):
            super().__init_subclass__(**kwargs)
            cls.debug = debug

    class NoDebug(DefaultBase):
        pass

    class WithDebug(DefaultBase, debug=True):
        pass

    expect(NoDebug.debug).to_be(False)
    expect(WithDebug.debug).to_be(True)

test("CPython: __init_subclass__ with default kwargs", test_init_subclass_default_kwargs)

# Test 19: __init_subclass__ error with extra kwargs not consumed
# NOTE: Skipped - RAGE does not currently raise TypeError for unconsumed
# kwargs in __init_subclass__.
