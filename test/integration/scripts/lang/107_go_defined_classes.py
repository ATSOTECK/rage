# Test: Go-Defined Classes (ClassBuilder API)
# Tests that classes defined in Go via the ClassBuilder API work seamlessly from Python.
# The Go test harness registers these classes before running this script:
#   Person(name, age)      — __init__, method greet(), __str__
#   Animal(name)           — base class with speak()
#   Dog(name)              — inherits Animal, overrides speak()
#   Cat(name)              — inherits Animal, overrides speak()
#   Container(items)       — __len__, __getitem__, __contains__, __eq__, __bool__, __str__
#   Multiplier(factor)     — __call__
#   Rect(w, h)             — properties: area (read-only), width (read-write)
#   Counter(n)             — static_method from_string(s), class_method class_name()
#   Vec2(x, y)             — __add__, __sub__, __mul__, __neg__, __abs__, __str__, __repr__
#   GoBase(value)          — simple base class for Python to inherit from
#   Store()                — __setitem__, __getitem__
#   Formattable(v)         — __format__, __int__, __float__, __index__
#   Accumulator(start)     — __iadd__, __isub__
#   DynObj()               — __getattr__ (dynamic attributes)
#   WithAttrs              — class-level attributes via Attr()
#   CustomNew              — custom __new__
#   Roundable(v)           — __round__
#   RevList(items)         — __reversed__
#   GenericBox             — __class_getitem__
#   KwCall(prefix)         — __call__ with kwargs

from test_framework import test, expect

# ===========================================================================
# Basic class: __init__, methods, __str__
# ===========================================================================

def test_person_creation():
    p = Person("Alice", 30)
    expect(p.name).to_be("Alice")
    expect(p.age).to_be(30)

def test_person_method():
    p = Person("Bob", 25)
    expect(p.greet()).to_be("Hello, I'm Bob")

def test_person_str():
    p = Person("Charlie", 40)
    expect(str(p)).to_be("Person(Charlie, 40)")

def test_person_multiple_instances():
    p1 = Person("A", 1)
    p2 = Person("B", 2)
    p3 = Person("C", 3)
    expect(p1.name).to_be("A")
    expect(p2.name).to_be("B")
    expect(p3.name).to_be("C")
    expect(p1.greet()).to_be("Hello, I'm A")
    expect(p3.greet()).to_be("Hello, I'm C")

def test_person_isinstance():
    p = Person("D", 10)
    expect(isinstance(p, Person)).to_be(True)
    expect(isinstance(p, object)).to_be(True)

def test_person_type():
    p = Person("E", 20)
    expect(type(p).__name__).to_be("Person")

test("person_creation", test_person_creation)
test("person_method", test_person_method)
test("person_str", test_person_str)
test("person_multiple_instances", test_person_multiple_instances)
test("person_isinstance", test_person_isinstance)
test("person_type", test_person_type)

# ===========================================================================
# Inheritance: Go-defined base + Go-defined subclasses
# ===========================================================================

def test_inheritance_init():
    d = Dog("Rex")
    expect(d.name).to_be("Rex")
    c = Cat("Whiskers")
    expect(c.name).to_be("Whiskers")

def test_inheritance_override():
    d = Dog("Rex")
    expect(d.speak()).to_be("Rex says Woof!")
    c = Cat("Whiskers")
    expect(c.speak()).to_be("Whiskers says Meow!")

def test_inheritance_isinstance():
    d = Dog("Rex")
    expect(isinstance(d, Dog)).to_be(True)
    expect(isinstance(d, Animal)).to_be(True)
    expect(isinstance(d, Cat)).to_be(False)

def test_inheritance_polymorphism():
    animals = [Dog("Rex"), Cat("Whiskers"), Dog("Buddy"), Cat("Mittens")]
    sounds = [a.speak() for a in animals]
    expect(sounds[0]).to_be("Rex says Woof!")
    expect(sounds[1]).to_be("Whiskers says Meow!")
    expect(sounds[2]).to_be("Buddy says Woof!")
    expect(sounds[3]).to_be("Mittens says Meow!")

def test_inheritance_base_method():
    a = Animal("Generic")
    expect(a.speak()).to_be("...")

test("inheritance_init", test_inheritance_init)
test("inheritance_override", test_inheritance_override)
test("inheritance_isinstance", test_inheritance_isinstance)
test("inheritance_polymorphism", test_inheritance_polymorphism)
test("inheritance_base_method", test_inheritance_base_method)

# ===========================================================================
# Dunder protocols: __len__, __getitem__, __contains__, __eq__, __bool__, __str__
# ===========================================================================

def test_container_len():
    c = Container([1, 2, 3])
    expect(len(c)).to_be(3)
    expect(len(Container([]))).to_be(0)
    expect(len(Container([10]))).to_be(1)

def test_container_getitem():
    c = Container([10, 20, 30])
    expect(c[0]).to_be(10)
    expect(c[1]).to_be(20)
    expect(c[2]).to_be(30)

def test_container_contains():
    c = Container([1, 2, 3, 4, 5])
    expect(1 in c).to_be(True)
    expect(3 in c).to_be(True)
    expect(5 in c).to_be(True)
    expect(6 in c).to_be(False)
    expect(0 in c).to_be(False)

def test_container_eq():
    expect(Container([1, 2]) == Container([1, 2])).to_be(True)
    expect(Container([1, 2]) == Container([1, 3])).to_be(False)
    expect(Container([]) == Container([])).to_be(True)
    expect(Container([1]) == Container([1, 2])).to_be(False)

def test_container_bool():
    expect(bool(Container([1]))).to_be(True)
    expect(bool(Container([1, 2, 3]))).to_be(True)
    expect(bool(Container([]))).to_be(False)

def test_container_str():
    expect(str(Container([1, 2, 3]))).to_be("Container(3 items)")
    expect(str(Container([]))).to_be("Container(0 items)")

def test_container_in_if():
    c = Container([10, 20, 30])
    if c:
        result = "truthy"
    else:
        result = "falsy"
    expect(result).to_be("truthy")

    empty = Container([])
    if empty:
        result2 = "truthy"
    else:
        result2 = "falsy"
    expect(result2).to_be("falsy")

test("container_len", test_container_len)
test("container_getitem", test_container_getitem)
test("container_contains", test_container_contains)
test("container_eq", test_container_eq)
test("container_bool", test_container_bool)
test("container_str", test_container_str)
test("container_in_if", test_container_in_if)

# ===========================================================================
# __call__: callable instances
# ===========================================================================

def test_callable_basic():
    double = Multiplier(2)
    expect(double(5)).to_be(10)
    expect(double(0)).to_be(0)
    expect(double(100)).to_be(200)

def test_callable_different_factors():
    triple = Multiplier(3)
    times_ten = Multiplier(10)
    expect(triple(7)).to_be(21)
    expect(times_ten(7)).to_be(70)

def test_callable_in_comprehension():
    double = Multiplier(2)
    result = [double(x) for x in range(1, 6)]
    expect(result).to_be([2, 4, 6, 8, 10])

def test_callable_as_key_function():
    items = [3, 1, 4, 1, 5]
    neg = Multiplier(-1)
    # Use callable to negate values for sorted
    result = sorted(items, key=neg)
    expect(result).to_be([5, 4, 3, 1, 1])

test("callable_basic", test_callable_basic)
test("callable_different_factors", test_callable_different_factors)
test("callable_in_comprehension", test_callable_in_comprehension)
test("callable_as_key_function", test_callable_as_key_function)

# ===========================================================================
# Properties: read-only and read-write
# ===========================================================================

def test_property_read_only():
    r = Rect(3, 4)
    expect(r.area).to_be(12)

def test_property_read_write():
    r = Rect(5, 6)
    expect(r.width).to_be(5)
    r.width = 10
    expect(r.width).to_be(10)

def test_property_derived_update():
    r = Rect(5, 6)
    expect(r.area).to_be(30)
    r.width = 10
    expect(r.area).to_be(60)

def test_property_multiple_rects():
    r1 = Rect(2, 3)
    r2 = Rect(4, 5)
    expect(r1.area).to_be(6)
    expect(r2.area).to_be(20)
    r1.width = 10
    expect(r1.area).to_be(30)
    expect(r2.area).to_be(20)  # r2 unchanged

test("property_read_only", test_property_read_only)
test("property_read_write", test_property_read_write)
test("property_derived_update", test_property_derived_update)
test("property_multiple_rects", test_property_multiple_rects)

# ===========================================================================
# Static methods and class methods
# ===========================================================================

def test_static_method_on_class():
    expect(Counter.from_string("hello")).to_be(5)
    expect(Counter.from_string("")).to_be(0)
    expect(Counter.from_string("ab")).to_be(2)

def test_static_method_on_instance():
    c = Counter(0)
    expect(c.from_string("test")).to_be(4)

def test_class_method():
    expect(Counter.class_name()).to_be("Counter")

def test_class_method_on_instance():
    c = Counter(0)
    expect(c.class_name()).to_be("Counter")

test("static_method_on_class", test_static_method_on_class)
test("static_method_on_instance", test_static_method_on_instance)
test("class_method", test_class_method)
test("class_method_on_instance", test_class_method_on_instance)

# ===========================================================================
# __add__ and __repr__
# ===========================================================================

def test_vec2_add():
    v1 = Vec2(1, 2)
    v2 = Vec2(3, 4)
    result = v1 + v2
    # Result is a list [x, y] since Vec2.__add__ returns a list
    expect(result).to_be([4, 6])

def test_vec2_str():
    expect(str(Vec2(5, 10))).to_be("Vec2(5, 10)")
    expect(str(Vec2(0, 0))).to_be("Vec2(0, 0)")
    expect(str(Vec2(-1, -2))).to_be("Vec2(-1, -2)")

def test_vec2_repr():
    expect(repr(Vec2(3, 4))).to_be("Vec2(3, 4)")

test("vec2_add", test_vec2_add)
test("vec2_str", test_vec2_str)
test("vec2_repr", test_vec2_repr)

# ===========================================================================
# __setitem__ / __getitem__
# ===========================================================================

def test_store_setitem_getitem():
    s = Store()
    s["x"] = 42
    s["y"] = 100
    expect(s["x"]).to_be(42)
    expect(s["y"]).to_be(100)

def test_store_overwrite():
    s = Store()
    s["key"] = 1
    expect(s["key"]).to_be(1)
    s["key"] = 999
    expect(s["key"]).to_be(999)

def test_store_string_values():
    s = Store()
    s["name"] = "Alice"
    expect(s["name"]).to_be("Alice")

test("store_setitem_getitem", test_store_setitem_getitem)
test("store_overwrite", test_store_overwrite)
test("store_string_values", test_store_string_values)

# ===========================================================================
# Python class inheriting from Go-defined class
# ===========================================================================

def test_python_inherits_go_class():
    class PyChild(GoBase):
        def doubled(self):
            return self.get_value() * 2

    c = PyChild(21)
    expect(c.get_value()).to_be(21)
    expect(c.doubled()).to_be(42)

def test_python_inherits_go_isinstance():
    class PyChild(GoBase):
        pass

    c = PyChild(5)
    expect(isinstance(c, PyChild)).to_be(True)
    expect(isinstance(c, GoBase)).to_be(True)

def test_python_overrides_go_method():
    class PyChild(GoBase):
        def get_value(self):
            return 999

    c = PyChild(1)
    expect(c.get_value()).to_be(999)

def test_python_extends_go_init():
    class Tagged(GoBase):
        def __init__(self, value, tag):
            GoBase.__init__(self, value)
            self.tag = tag

    t = Tagged(42, "important")
    expect(t.get_value()).to_be(42)
    expect(t.tag).to_be("important")

test("python_inherits_go_class", test_python_inherits_go_class)
test("python_inherits_go_isinstance", test_python_inherits_go_isinstance)
test("python_overrides_go_method", test_python_overrides_go_method)
test("python_extends_go_init", test_python_extends_go_init)

# ===========================================================================
# Go-defined class instances created from Go (NewInstance)
# ===========================================================================

def test_go_created_instance():
    # 'config' is a GoBase instance created directly from Go via NewInstance()
    expect(config.host).to_be("localhost")
    expect(config.port).to_be(8080)

def test_go_created_instance_method():
    expect(config.get("host")).to_be("localhost")
    expect(config.get("port")).to_be(8080)

def test_go_created_instance_isinstance():
    expect(isinstance(config, Config)).to_be(True)

test("go_created_instance", test_go_created_instance)
test("go_created_instance_method", test_go_created_instance_method)
test("go_created_instance_isinstance", test_go_created_instance_isinstance)

# ===========================================================================
# Mixed usage: Go classes in Python data structures
# ===========================================================================

def test_go_instances_in_list():
    people = [Person("A", 1), Person("B", 2), Person("C", 3)]
    names = [p.name for p in people]
    expect(names).to_be(["A", "B", "C"])

def test_go_instances_in_dict():
    d = {"rex": Dog("Rex"), "whiskers": Cat("Whiskers")}
    expect(d["rex"].speak()).to_be("Rex says Woof!")
    expect(d["whiskers"].speak()).to_be("Whiskers says Meow!")

def test_go_instances_as_function_args():
    def describe(person):
        return person.name + " is " + str(person.age)
    expect(describe(Person("Alice", 30))).to_be("Alice is 30")

def test_go_instances_filtered():
    people = [Person("A", 10), Person("B", 25), Person("C", 5), Person("D", 30)]
    adults = [p.name for p in people if p.age >= 18]
    expect(adults).to_be(["B", "D"])

def test_go_instances_sorted():
    people = [Person("C", 30), Person("A", 10), Person("B", 20)]
    by_age = sorted(people, key=lambda p: p.age)
    names = [p.name for p in by_age]
    expect(names).to_be(["A", "B", "C"])

test("go_instances_in_list", test_go_instances_in_list)
test("go_instances_in_dict", test_go_instances_in_dict)
test("go_instances_as_function_args", test_go_instances_as_function_args)
test("go_instances_filtered", test_go_instances_filtered)
test("go_instances_sorted", test_go_instances_sorted)

# ===========================================================================
# Iterator protocol: __iter__ / __next__
# ===========================================================================

def test_iter_basic():
    r = GoRange(0, 5)
    result = []
    for x in r:
        result.append(x)
    expect(result).to_be([0, 1, 2, 3, 4])

def test_iter_empty():
    r = GoRange(5, 5)
    result = []
    for x in r:
        result.append(x)
    expect(result).to_be([])

def test_iter_in_list_comprehension():
    result = [x * 2 for x in GoRange(1, 4)]
    expect(result).to_be([2, 4, 6])

def test_iter_sum():
    total = 0
    for x in GoRange(1, 6):
        total += x
    expect(total).to_be(15)

def test_iter_multiple_passes():
    r = GoRange(0, 3)
    first = [x for x in r]
    second = [x for x in r]
    expect(first).to_be([0, 1, 2])
    expect(second).to_be([0, 1, 2])

test("iter_basic", test_iter_basic)
test("iter_empty", test_iter_empty)
test("iter_in_list_comprehension", test_iter_in_list_comprehension)
test("iter_sum", test_iter_sum)
test("iter_multiple_passes", test_iter_multiple_passes)

# ===========================================================================
# Comparison operators: __eq__, __lt__, __le__, __gt__, __ge__
# ===========================================================================

def test_temp_eq():
    expect(Temperature(100) == Temperature(100)).to_be(True)
    expect(Temperature(100) == Temperature(200)).to_be(False)

def test_temp_lt():
    expect(Temperature(50) < Temperature(100)).to_be(True)
    expect(Temperature(100) < Temperature(50)).to_be(False)
    expect(Temperature(50) < Temperature(50)).to_be(False)

def test_temp_le():
    expect(Temperature(50) <= Temperature(100)).to_be(True)
    expect(Temperature(50) <= Temperature(50)).to_be(True)
    expect(Temperature(100) <= Temperature(50)).to_be(False)

def test_temp_gt():
    expect(Temperature(100) > Temperature(50)).to_be(True)
    expect(Temperature(50) > Temperature(100)).to_be(False)
    expect(Temperature(50) > Temperature(50)).to_be(False)

def test_temp_ge():
    expect(Temperature(100) >= Temperature(50)).to_be(True)
    expect(Temperature(50) >= Temperature(50)).to_be(True)
    expect(Temperature(50) >= Temperature(100)).to_be(False)

def test_temp_sorting():
    temps = [Temperature(30), Temperature(10), Temperature(20)]
    sorted_temps = sorted(temps, key=lambda t: t.value)
    expect([t.value for t in sorted_temps]).to_be([10, 20, 30])

test("temp_eq", test_temp_eq)
test("temp_lt", test_temp_lt)
test("temp_le", test_temp_le)
test("temp_gt", test_temp_gt)
test("temp_ge", test_temp_ge)
test("temp_sorting", test_temp_sorting)

# ===========================================================================
# __hash__
# ===========================================================================

def test_temp_hash():
    t1 = Temperature(100)
    t2 = Temperature(100)
    expect(hash(t1)).to_be(hash(t2))

def test_temp_in_dict_as_key():
    d = {}
    d[Temperature(100)] = "boiling"
    d[Temperature(0)] = "freezing"
    expect(d[Temperature(100)]).to_be("boiling")
    expect(d[Temperature(0)]).to_be("freezing")

def test_temp_in_set():
    s = {Temperature(10), Temperature(20), Temperature(10)}
    expect(len(s)).to_be(2)

test("temp_hash", test_temp_hash)
test("temp_in_dict_as_key", test_temp_in_dict_as_key)
test("temp_in_set", test_temp_in_set)

# ===========================================================================
# __delitem__
# ===========================================================================

def test_delitem_basic():
    l = Ledger()
    l["x"] = 10
    l["y"] = 20
    expect(l["x"]).to_be(10)
    del l["x"]
    # After deletion, accessing deleted key returns None (our Get fallback)
    expect(l["x"]).to_be(None)
    expect(l["y"]).to_be(20)

def test_delitem_missing_key():
    l = Ledger()
    caught = False
    try:
        del l["nonexistent"]
    except KeyError:
        caught = True
    expect(caught).to_be(True)

test("delitem_basic", test_delitem_basic)
test("delitem_missing_key", test_delitem_missing_key)

# ===========================================================================
# Context manager: __enter__ / __exit__
# ===========================================================================

def test_context_manager_basic():
    cm = GoContextManager("test")
    expect(cm.status()).to_be("entered=False exited=False error=False")
    with cm as ctx:
        expect(cm.status()).to_be("entered=True exited=False error=False")
    expect(cm.status()).to_be("entered=True exited=True error=False")

def test_context_manager_with_exception():
    cm = GoContextManager("err_test")
    caught = False
    try:
        with cm:
            raise ValueError("test error")
    except ValueError:
        caught = True
    expect(caught).to_be(True)
    expect(cm.status()).to_be("entered=True exited=True error=True")

def test_context_manager_returns_self():
    cm = GoContextManager("self_test")
    with cm as ctx:
        expect(ctx.name).to_be("self_test")

test("context_manager_basic", test_context_manager_basic)
test("context_manager_with_exception", test_context_manager_with_exception)
test("context_manager_returns_self", test_context_manager_returns_self)

# ===========================================================================
# Error raising from Go methods
# ===========================================================================

def test_go_raises_value_error():
    er = ErrorRaiser()
    caught = False
    msg = ""
    try:
        er.raise_value_error()
    except ValueError as e:
        caught = True
        msg = str(e)
    expect(caught).to_be(True)
    expect("bad value from Go" in msg).to_be(True)

def test_go_raises_type_error():
    er = ErrorRaiser()
    caught = False
    try:
        er.raise_type_error()
    except TypeError:
        caught = True
    expect(caught).to_be(True)

def test_go_raises_key_error():
    er = ErrorRaiser()
    caught = False
    try:
        er.raise_key_error()
    except KeyError:
        caught = True
    expect(caught).to_be(True)

test("go_raises_value_error", test_go_raises_value_error)
test("go_raises_type_error", test_go_raises_type_error)
test("go_raises_key_error", test_go_raises_key_error)

# ===========================================================================
# Vec2 extended: __sub__, __mul__, __neg__, __abs__
# ===========================================================================

def test_vec2_sub():
    v1 = Vec2(5, 10)
    v2 = Vec2(2, 3)
    result = v1 - v2
    expect(result).to_be([3, 7])

def test_vec2_mul():
    v = Vec2(3, 4)
    result = v * 2
    expect(result).to_be([6, 8])

def test_vec2_neg():
    v = Vec2(3, -4)
    result = -v
    expect(result).to_be([-3, 4])

def test_vec2_abs():
    v = Vec2(-3, -4)
    result = abs(v)
    expect(result).to_be([3, 4])

test("vec2_sub", test_vec2_sub)
test("vec2_mul", test_vec2_mul)
test("vec2_neg", test_vec2_neg)
test("vec2_abs", test_vec2_abs)

# ===========================================================================
# __format__, __int__, __float__, __index__
# ===========================================================================

def test_format_default():
    f = Formattable(42)
    expect(format(f, "")).to_be("42")

def test_format_hex():
    f = Formattable(255)
    expect(format(f, "hex")).to_be("0xff")

def test_format_bin():
    f = Formattable(10)
    expect(format(f, "bin")).to_be("0b1010")

def test_int_conv():
    f = Formattable(42)
    expect(int(f)).to_be(42)

def test_float_conv():
    f = Formattable(3)
    expect(float(f)).to_be(3.5)

def test_index_conv():
    f = Formattable(1)
    items = [10, 20, 30]
    expect(items[f]).to_be(20)

test("format_default", test_format_default)
test("format_hex", test_format_hex)
test("format_bin", test_format_bin)
test("int_conv", test_int_conv)
test("float_conv", test_float_conv)
test("index_conv", test_index_conv)

# ===========================================================================
# In-place operators: __iadd__, __isub__
# ===========================================================================

def test_iadd():
    a = Accumulator(10)
    a += 5
    expect(a.v).to_be(15)

def test_isub():
    a = Accumulator(10)
    a -= 3
    expect(a.v).to_be(7)

def test_iadd_chain():
    a = Accumulator(0)
    a += 10
    a += 20
    a += 30
    expect(a.v).to_be(60)

test("iadd", test_iadd)
test("isub", test_isub)
test("iadd_chain", test_iadd_chain)

# ===========================================================================
# __getattr__ (dynamic attributes)
# ===========================================================================

def test_getattr_dynamic():
    d = DynObj()
    expect(d.foo).to_be("dynamic:foo")
    expect(d.bar).to_be("dynamic:bar")
    expect(d.anything).to_be("dynamic:anything")

test("getattr_dynamic", test_getattr_dynamic)

# ===========================================================================
# Class-level attributes via Attr()
# ===========================================================================

def test_class_attrs():
    expect(WithAttrs.VERSION).to_be(1)
    expect(WithAttrs.NAME).to_be("WithAttrs")

def test_class_attrs_on_instance():
    w = WithAttrs()
    expect(w.VERSION).to_be(1)
    expect(w.NAME).to_be("WithAttrs")

test("class_attrs", test_class_attrs)
test("class_attrs_on_instance", test_class_attrs_on_instance)

# ===========================================================================
# Custom __new__
# ===========================================================================

def test_custom_new():
    obj = CustomNew(42)
    expect(obj.from_new).to_be(True)
    expect(obj.v).to_be(42)

def test_custom_new_no_args():
    obj = CustomNew()
    expect(obj.from_new).to_be(True)

test("custom_new", test_custom_new)
test("custom_new_no_args", test_custom_new_no_args)

# ===========================================================================
# __round__
# ===========================================================================

def test_round_basic():
    r = Roundable(3.7)
    expect(round(r)).to_be(4)

def test_round_exact():
    r = Roundable(5.0)
    expect(round(r)).to_be(5)

test("round_basic", test_round_basic)
test("round_exact", test_round_exact)

# ===========================================================================
# __reversed__
# ===========================================================================

def test_reversed_basic():
    r = RevList([1, 2, 3])
    expect(list(reversed(r))).to_be([3, 2, 1])

def test_reversed_empty():
    r = RevList([])
    expect(list(reversed(r))).to_be([])

def test_reversed_single():
    r = RevList([42])
    expect(list(reversed(r))).to_be([42])

test("reversed_basic", test_reversed_basic)
test("reversed_empty", test_reversed_empty)
test("reversed_single", test_reversed_single)

# ===========================================================================
# __class_getitem__ (generics syntax)
# ===========================================================================

def test_class_getitem():
    result = GenericBox["int"]
    expect(result).to_be("GenericBox[int]")

def test_class_getitem_str():
    result = GenericBox["str"]
    expect(result).to_be("GenericBox[str]")

test("class_getitem", test_class_getitem)
test("class_getitem_str", test_class_getitem_str)

# ===========================================================================
# __call__ with kwargs (CallKw)
# ===========================================================================

def test_callkw_no_kwargs():
    c = KwCall("hello")
    expect(c("world")).to_be("hello world")

def test_callkw_with_sep():
    c = KwCall("a")
    expect(c("b", "c", sep="-")).to_be("a-b-c")

test("callkw_no_kwargs", test_callkw_no_kwargs)
test("callkw_with_sep", test_callkw_with_sep)

print("Go-defined classes tests completed")
