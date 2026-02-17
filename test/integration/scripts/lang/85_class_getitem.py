from test_framework import test, expect

# Test 1: Basic __class_getitem__ on user-defined class
def test_basic_class_getitem():
    class MyGeneric:
        def __class_getitem__(cls, item):
            return "MyGeneric[parameterized]"

    result = MyGeneric[int]
    expect(result).to_be("MyGeneric[parameterized]")

test("basic __class_getitem__ on user class", test_basic_class_getitem)

# Test 2: __class_getitem__ receives the class as first arg
def test_class_getitem_receives_cls():
    class MyGeneric:
        def __class_getitem__(cls, item):
            return cls

    result = MyGeneric[int]
    expect(result).to_be(MyGeneric)

test("__class_getitem__ receives cls", test_class_getitem_receives_cls)

# Test 3: __class_getitem__ with multiple type args (tuple)
def test_multiple_type_args():
    class MyMapping:
        def __class_getitem__(cls, params):
            if isinstance(params, tuple):
                return len(params)
            return 1

    result = MyMapping[str, int]
    expect(result).to_be(2)

test("__class_getitem__ with multiple type args", test_multiple_type_args)

# Test 4: __class_getitem__ inherited from parent
def test_inherited_class_getitem():
    class Base:
        def __class_getitem__(cls, item):
            return 42

    class Child(Base):
        pass

    result = Child[int]
    expect(result).to_be(42)

test("__class_getitem__ inherited from parent", test_inherited_class_getitem)

# Test 5: __class_getitem__ returning arbitrary value
def test_return_arbitrary():
    class MyType:
        def __class_getitem__(cls, item):
            return [1, 2, 3]

    result = MyType[str]
    expect(result).to_be([1, 2, 3])

test("__class_getitem__ returning arbitrary value", test_return_arbitrary)

# Test 6: Built-in list[int] returns GenericAlias
def test_list_subscript():
    alias = list[int]
    result = str(alias)
    expect(result).to_be("list[int]")

test("list[int] returns GenericAlias", test_list_subscript)

# Test 7: Built-in dict[str, int] returns GenericAlias
def test_dict_subscript():
    alias = dict[str, int]
    result = str(alias)
    expect(result).to_be("dict[str, int]")

test("dict[str, int] returns GenericAlias", test_dict_subscript)

# Test 8: Built-in tuple[int, str] returns GenericAlias
def test_tuple_subscript():
    alias = tuple[int, str]
    result = str(alias)
    expect(result).to_be("tuple[int, str]")

test("tuple[int, str] returns GenericAlias", test_tuple_subscript)

# Test 9: Built-in set[int] returns GenericAlias
def test_set_subscript():
    alias = set[int]
    result = str(alias)
    expect(result).to_be("set[int]")

test("set[int] returns GenericAlias", test_set_subscript)

# Test 10: Built-in frozenset[int] returns GenericAlias
def test_frozenset_subscript():
    alias = frozenset[int]
    result = str(alias)
    expect(result).to_be("frozenset[int]")

test("frozenset[int] returns GenericAlias", test_frozenset_subscript)

# Test 11: GenericAlias __origin__ attribute
def test_generic_alias_origin():
    alias = list[int]
    expect(str(alias.__origin__)).to_be(str(list))

test("GenericAlias __origin__", test_generic_alias_origin)

# Test 12: GenericAlias __args__ attribute
def test_generic_alias_args():
    alias = dict[str, int]
    args = alias.__args__
    expect(len(args)).to_be(2)
    expect(str(args[0])).to_be(str(str))
    expect(str(args[1])).to_be(str(int))

test("GenericAlias __args__", test_generic_alias_args)

# Test 13: Non-subscriptable class raises TypeError
def test_non_subscriptable():
    class Plain:
        pass

    try:
        Plain[int]
        expect(True).to_be(False)  # Should not reach here
    except TypeError:
        expect(True).to_be(True)

test("non-subscriptable class raises TypeError", test_non_subscriptable)

# Test 14: __class_getitem__ as explicit classmethod
def test_explicit_classmethod():
    class MyType:
        @classmethod
        def __class_getitem__(cls, item):
            return "explicit"

    result = MyType[int]
    expect(result).to_be("explicit")

test("__class_getitem__ as explicit classmethod", test_explicit_classmethod)

# Test 15: __class_getitem__ child overrides parent
def test_override_class_getitem():
    class Base:
        def __class_getitem__(cls, item):
            return "base"

    class Child(Base):
        def __class_getitem__(cls, item):
            return "child"

    expect(Base[int]).to_be("base")
    expect(Child[int]).to_be("child")

test("__class_getitem__ child overrides parent", test_override_class_getitem)
