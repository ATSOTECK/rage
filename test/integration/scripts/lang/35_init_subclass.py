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
