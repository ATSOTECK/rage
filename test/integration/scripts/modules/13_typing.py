# Test: typing module
# Tests type hints, type variables, generic types, and utility functions

from test_framework import test, expect

from typing import List, Dict, Set, Tuple, Optional, Union, Any, Callable
from typing import TypeVar, Generic, Protocol
from typing import Sequence, Mapping, Iterable, Iterator
from typing import FrozenSet, Type
from typing import get_origin, get_args
from typing import cast
from typing import TYPE_CHECKING
from typing import NewType
from typing import overload, final, no_type_check, runtime_checkable
from typing import reveal_type, assert_type
from typing import get_overloads, clear_overloads
from typing import Final, ClassVar, Literal
from typing import NoReturn, Never, Self
from typing import ParamSpec, TypeVarTuple
from typing import NamedTuple, TypedDict, is_typeddict, get_type_hints
from typing import Annotated
from typing import Required, NotRequired, ReadOnly
from typing import Concatenate, TypeGuard, Unpack
from typing import Counter, ChainMap, OrderedDict, DefaultDict, Deque
from typing import ContextManager, AsyncContextManager
from typing import Coroutine, AsyncGenerator, AsyncIterable, AsyncIterator, Awaitable
from typing import IO, TextIO, BinaryIO
from typing import Pattern, Match
from typing import SupportsInt, SupportsFloat, SupportsAbs, SupportsRound, SupportsIndex
from typing import Hashable, Sized, Reversible
from typing import MutableMapping, MutableSequence, MutableSet
from typing import Generator

T = TypeVar("T")
K = TypeVar("K")
V = TypeVar("V")
Num = TypeVar("Num", int, float)
P = ParamSpec("P")
Ts = TypeVarTuple("Ts")
UserId = NewType("UserId", int)
Point = NamedTuple("Point", [("x", int), ("y", int)])
Movie = TypedDict("Movie", {})

@overload
def process(x):
    return x

@final
def final_func():
    return "final"

@no_type_check
def unchecked_func(x):
    return x + 1

@runtime_checkable
def checkable_class():
    pass

def sample_func(x, y):
    return x + y

def test_generic_types_exist():
    expect(List is not None).to_be(True)
    expect(Dict is not None).to_be(True)
    expect(Set is not None).to_be(True)
    expect(Tuple is not None).to_be(True)
    expect(Optional is not None).to_be(True)
    expect(Union is not None).to_be(True)
    expect(Any is not None).to_be(True)
    expect(Callable is not None).to_be(True)

def test_subscripted_types():
    list_int = List[int]
    expect(list_int is not None).to_be(True)
    optional_str = Optional[str]
    expect(optional_str is not None).to_be(True)
    union_int_str = Union[int]
    expect(union_int_str is not None).to_be(True)
    tuple_int = Tuple[int]
    expect(tuple_int is not None).to_be(True)
    set_str = Set[str]
    expect(set_str is not None).to_be(True)

def test_typevar():
    expect(T is not None).to_be(True)
    expect(K is not None and V is not None).to_be(True)
    expect(Num is not None).to_be(True)

def test_get_origin_args():
    list_int = List[int]
    origin = get_origin(list_int)
    expect(origin).to_be("List")
    args = get_args(list_int)
    expect(len(args)).to_be(1)
    expect(get_origin(List)).to_be("list")
    expect(len(get_args(List))).to_be(0)

def test_cast():
    x = cast(int, "hello")
    expect(x).to_be("hello")
    y = cast(str, 42)
    expect(y).to_be(42)

def test_type_checking():
    expect(TYPE_CHECKING).to_be(False)

def test_newtype():
    user_id = UserId(42)
    expect(user_id).to_be(42)
    expect(isinstance(user_id, int)).to_be(True)

def test_decorators():
    expect(process(5)).to_be(5)
    expect(final_func()).to_be("final")
    expect(unchecked_func(10)).to_be(11)

def test_reveal_assert():
    val = reveal_type(42)
    expect(val).to_be(42)
    val2 = assert_type(42, int)
    expect(val2).to_be(42)

def test_get_overloads():
    overloads = get_overloads(process)
    expect(isinstance(overloads, list)).to_be(True)
    clear_overloads()

def test_sequence_types():
    expect(Sequence is not None).to_be(True)
    expect(Mapping is not None).to_be(True)
    expect(Iterable is not None).to_be(True)
    expect(Iterator is not None).to_be(True)
    seq_int = Sequence[int]
    expect(seq_int is not None).to_be(True)
    iter_str = Iterable[str]
    expect(iter_str is not None).to_be(True)

def test_additional_types():
    expect(FrozenSet is not None).to_be(True)
    expect(Type is not None).to_be(True)
    frozenset_int = FrozenSet[int]
    expect(frozenset_int is not None).to_be(True)
    type_int = Type[int]
    expect(type_int is not None).to_be(True)

def test_protocol_generic():
    expect(Generic is not None).to_be(True)
    expect(Protocol is not None).to_be(True)

def test_special_forms():
    expect(Final is not None).to_be(True)
    expect(ClassVar is not None).to_be(True)
    expect(Literal is not None).to_be(True)
    final_int = Final[int]
    expect(final_int is not None).to_be(True)
    classvar_str = ClassVar[str]
    expect(classvar_str is not None).to_be(True)
    literal_one = Literal[1]
    expect(literal_one is not None).to_be(True)

def test_special_types():
    expect(NoReturn is not None).to_be(True)
    expect(Never is not None).to_be(True)
    expect(Self is not None).to_be(True)

def test_paramspec_typevartuple():
    expect(P is not None).to_be(True)
    expect(Ts is not None).to_be(True)

def test_namedtuple_typeddict():
    expect(Point is not None).to_be(True)
    expect(Movie is not None).to_be(True)
    expect(is_typeddict(Movie)).to_be(False)
    hints = get_type_hints(sample_func)
    expect(isinstance(hints, dict)).to_be(True)

def test_annotated():
    annotated_int = Annotated[int]
    expect(annotated_int is not None).to_be(True)

def test_required_notreq_readonly():
    expect(Required is not None).to_be(True)
    expect(NotRequired is not None).to_be(True)
    expect(ReadOnly is not None).to_be(True)
    required_int = Required[int]
    expect(required_int is not None).to_be(True)

def test_concatenate_typeguard_unpack():
    expect(Concatenate is not None).to_be(True)
    expect(TypeGuard is not None).to_be(True)
    expect(Unpack is not None).to_be(True)
    typeguard_bool = TypeGuard[bool]
    expect(typeguard_bool is not None).to_be(True)

def test_collection_aliases():
    expect(Counter is not None).to_be(True)
    expect(ChainMap is not None).to_be(True)
    expect(OrderedDict is not None).to_be(True)
    expect(DefaultDict is not None).to_be(True)
    expect(Deque is not None).to_be(True)
    counter_int = Counter[int]
    expect(counter_int is not None).to_be(True)
    deque_str = Deque[str]
    expect(deque_str is not None).to_be(True)

def test_context_managers():
    expect(ContextManager is not None).to_be(True)
    expect(AsyncContextManager is not None).to_be(True)

def test_async_types():
    expect(Coroutine is not None).to_be(True)
    expect(AsyncGenerator is not None).to_be(True)
    expect(AsyncIterable is not None).to_be(True)
    expect(AsyncIterator is not None).to_be(True)
    expect(Awaitable is not None).to_be(True)

def test_io_types():
    expect(IO is not None).to_be(True)
    expect(TextIO is not None).to_be(True)
    expect(BinaryIO is not None).to_be(True)

def test_pattern_match():
    expect(Pattern is not None).to_be(True)
    expect(Match is not None).to_be(True)
    pattern_str = Pattern[str]
    expect(pattern_str is not None).to_be(True)

def test_supports_protocols():
    expect(SupportsInt is not None).to_be(True)
    expect(SupportsFloat is not None).to_be(True)
    expect(SupportsAbs is not None).to_be(True)
    expect(SupportsRound is not None).to_be(True)
    expect(SupportsIndex is not None).to_be(True)

def test_abc_types():
    expect(Hashable is not None).to_be(True)
    expect(Sized is not None).to_be(True)
    expect(Reversible is not None).to_be(True)

def test_mutable_types():
    expect(MutableMapping is not None).to_be(True)
    expect(MutableSequence is not None).to_be(True)
    expect(MutableSet is not None).to_be(True)
    mutableseq_int = MutableSequence[int]
    expect(mutableseq_int is not None).to_be(True)

def test_generator():
    expect(Generator is not None).to_be(True)
    gen_int = Generator[int]
    expect(gen_int is not None).to_be(True)

test("generic_types_exist", test_generic_types_exist)
test("subscripted_types", test_subscripted_types)
test("typevar", test_typevar)
test("get_origin_args", test_get_origin_args)
test("cast", test_cast)
test("type_checking", test_type_checking)
test("newtype", test_newtype)
test("decorators", test_decorators)
test("reveal_assert", test_reveal_assert)
test("get_overloads", test_get_overloads)
test("sequence_types", test_sequence_types)
test("additional_types", test_additional_types)
test("protocol_generic", test_protocol_generic)
test("special_forms", test_special_forms)
test("special_types", test_special_types)
test("paramspec_typevartuple", test_paramspec_typevartuple)
test("namedtuple_typeddict", test_namedtuple_typeddict)
test("annotated", test_annotated)
test("required_notreq_readonly", test_required_notreq_readonly)
test("concatenate_typeguard_unpack", test_concatenate_typeguard_unpack)
test("collection_aliases", test_collection_aliases)
test("context_managers", test_context_managers)
test("async_types", test_async_types)
test("io_types", test_io_types)
test("pattern_match", test_pattern_match)
test("supports_protocols", test_supports_protocols)
test("abc_types", test_abc_types)
test("mutable_types", test_mutable_types)
test("generator", test_generator)

print("typing module tests completed")
