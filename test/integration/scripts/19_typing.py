# Test: typing module
# Tests type hints, type variables, generic types, and utility functions

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
    expect(True, List is not None)
    expect(True, Dict is not None)
    expect(True, Set is not None)
    expect(True, Tuple is not None)
    expect(True, Optional is not None)
    expect(True, Union is not None)
    expect(True, Any is not None)
    expect(True, Callable is not None)

def test_subscripted_types():
    list_int = List[int]
    expect(True, list_int is not None)
    optional_str = Optional[str]
    expect(True, optional_str is not None)
    union_int_str = Union[int]
    expect(True, union_int_str is not None)
    tuple_int = Tuple[int]
    expect(True, tuple_int is not None)
    set_str = Set[str]
    expect(True, set_str is not None)

def test_typevar():
    expect(True, T is not None)
    expect(True, K is not None and V is not None)
    expect(True, Num is not None)

def test_get_origin_args():
    list_int = List[int]
    origin = get_origin(list_int)
    expect("List", origin)
    args = get_args(list_int)
    expect(1, len(args))
    expect("list", get_origin(List))
    expect(0, len(get_args(List)))

def test_cast():
    x = cast(int, "hello")
    expect("hello", x)
    y = cast(str, 42)
    expect(42, y)

def test_type_checking():
    expect(False, TYPE_CHECKING)

def test_newtype():
    user_id = UserId(42)
    expect(42, user_id)
    expect(True, isinstance(user_id, int))

def test_decorators():
    expect(5, process(5))
    expect("final", final_func())
    expect(11, unchecked_func(10))

def test_reveal_assert():
    val = reveal_type(42)
    expect(42, val)
    val2 = assert_type(42, int)
    expect(42, val2)

def test_get_overloads():
    overloads = get_overloads(process)
    expect(True, isinstance(overloads, list))
    clear_overloads()

def test_sequence_types():
    expect(True, Sequence is not None)
    expect(True, Mapping is not None)
    expect(True, Iterable is not None)
    expect(True, Iterator is not None)
    seq_int = Sequence[int]
    expect(True, seq_int is not None)
    iter_str = Iterable[str]
    expect(True, iter_str is not None)

def test_additional_types():
    expect(True, FrozenSet is not None)
    expect(True, Type is not None)
    frozenset_int = FrozenSet[int]
    expect(True, frozenset_int is not None)
    type_int = Type[int]
    expect(True, type_int is not None)

def test_protocol_generic():
    expect(True, Generic is not None)
    expect(True, Protocol is not None)

def test_special_forms():
    expect(True, Final is not None)
    expect(True, ClassVar is not None)
    expect(True, Literal is not None)
    final_int = Final[int]
    expect(True, final_int is not None)
    classvar_str = ClassVar[str]
    expect(True, classvar_str is not None)
    literal_one = Literal[1]
    expect(True, literal_one is not None)

def test_special_types():
    expect(True, NoReturn is not None)
    expect(True, Never is not None)
    expect(True, Self is not None)

def test_paramspec_typevartuple():
    expect(True, P is not None)
    expect(True, Ts is not None)

def test_namedtuple_typeddict():
    expect(True, Point is not None)
    expect(True, Movie is not None)
    expect(False, is_typeddict(Movie))
    hints = get_type_hints(sample_func)
    expect(True, isinstance(hints, dict))

def test_annotated():
    annotated_int = Annotated[int]
    expect(True, annotated_int is not None)

def test_required_notreq_readonly():
    expect(True, Required is not None)
    expect(True, NotRequired is not None)
    expect(True, ReadOnly is not None)
    required_int = Required[int]
    expect(True, required_int is not None)

def test_concatenate_typeguard_unpack():
    expect(True, Concatenate is not None)
    expect(True, TypeGuard is not None)
    expect(True, Unpack is not None)
    typeguard_bool = TypeGuard[bool]
    expect(True, typeguard_bool is not None)

def test_collection_aliases():
    expect(True, Counter is not None)
    expect(True, ChainMap is not None)
    expect(True, OrderedDict is not None)
    expect(True, DefaultDict is not None)
    expect(True, Deque is not None)
    counter_int = Counter[int]
    expect(True, counter_int is not None)
    deque_str = Deque[str]
    expect(True, deque_str is not None)

def test_context_managers():
    expect(True, ContextManager is not None)
    expect(True, AsyncContextManager is not None)

def test_async_types():
    expect(True, Coroutine is not None)
    expect(True, AsyncGenerator is not None)
    expect(True, AsyncIterable is not None)
    expect(True, AsyncIterator is not None)
    expect(True, Awaitable is not None)

def test_io_types():
    expect(True, IO is not None)
    expect(True, TextIO is not None)
    expect(True, BinaryIO is not None)

def test_pattern_match():
    expect(True, Pattern is not None)
    expect(True, Match is not None)
    pattern_str = Pattern[str]
    expect(True, pattern_str is not None)

def test_supports_protocols():
    expect(True, SupportsInt is not None)
    expect(True, SupportsFloat is not None)
    expect(True, SupportsAbs is not None)
    expect(True, SupportsRound is not None)
    expect(True, SupportsIndex is not None)

def test_abc_types():
    expect(True, Hashable is not None)
    expect(True, Sized is not None)
    expect(True, Reversible is not None)

def test_mutable_types():
    expect(True, MutableMapping is not None)
    expect(True, MutableSequence is not None)
    expect(True, MutableSet is not None)
    mutableseq_int = MutableSequence[int]
    expect(True, mutableseq_int is not None)

def test_generator():
    expect(True, Generator is not None)
    gen_int = Generator[int]
    expect(True, gen_int is not None)

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
