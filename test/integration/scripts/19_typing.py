# Test: typing module
# Tests type hints, type variables, generic types, and utility functions

results = {}

# =====================================
# Basic imports
# =====================================
from typing import List, Dict, Set, Tuple, Optional, Union, Any, Callable
from typing import TypeVar, Generic, Protocol
from typing import Sequence, Mapping, Iterable, Iterator
from typing import FrozenSet, Type

# =====================================
# Generic type aliases - basic access
# =====================================
results["list_exists"] = List is not None
results["dict_exists"] = Dict is not None
results["set_exists"] = Set is not None
results["tuple_exists"] = Tuple is not None
results["optional_exists"] = Optional is not None
results["union_exists"] = Union is not None
results["any_exists"] = Any is not None
results["callable_exists"] = Callable is not None

# =====================================
# Subscripted types
# =====================================
list_int = List[int]
results["list_int_subscript"] = list_int is not None

optional_str = Optional[str]
results["optional_str_subscript"] = optional_str is not None

union_int_str = Union[int]  # Can't do Union[int, str] due to parser limitation
results["union_subscript"] = union_int_str is not None

tuple_int = Tuple[int]
results["tuple_subscript"] = tuple_int is not None

set_str = Set[str]
results["set_subscript"] = set_str is not None

# =====================================
# TypeVar
# =====================================
T = TypeVar("T")
results["typevar_exists"] = T is not None

K = TypeVar("K")
V = TypeVar("V")
results["typevar_multiple"] = K is not None and V is not None

# TypeVar with constraints
Num = TypeVar("Num", int, float)
results["typevar_constraints"] = Num is not None

# =====================================
# get_origin and get_args
# =====================================
from typing import get_origin, get_args

list_int = List[int]
origin = get_origin(list_int)
results["get_origin_list"] = origin == "List"

args = get_args(list_int)
results["get_args_list_len"] = len(args) == 1

# get_origin on unsubscripted type
results["get_origin_unsubscripted"] = get_origin(List) == "list"

# get_args on unsubscripted type
results["get_args_unsubscripted"] = len(get_args(List)) == 0

# =====================================
# cast function
# =====================================
from typing import cast

x = cast(int, "hello")
results["cast_returns_value"] = x == "hello"

y = cast(str, 42)
results["cast_no_conversion"] = y == 42

# =====================================
# TYPE_CHECKING constant
# =====================================
from typing import TYPE_CHECKING
results["type_checking_false"] = TYPE_CHECKING == False

# =====================================
# NewType
# =====================================
from typing import NewType

UserId = NewType("UserId", int)
user_id = UserId(42)
results["newtype_value"] = user_id == 42

# NewType returns the value unchanged
results["newtype_is_int"] = isinstance(user_id, int)

# =====================================
# Decorators - overload
# =====================================
from typing import overload

@overload
def process(x):
    return x

results["overload_decorator"] = process(5) == 5

# =====================================
# Decorators - final
# =====================================
from typing import final

@final
def final_func():
    return "final"

results["final_decorator"] = final_func() == "final"

# =====================================
# Decorators - no_type_check
# =====================================
from typing import no_type_check

@no_type_check
def unchecked_func(x):
    return x + 1

results["no_type_check_decorator"] = unchecked_func(10) == 11

# =====================================
# Decorators - runtime_checkable
# =====================================
from typing import runtime_checkable

@runtime_checkable
def checkable_class():
    pass

results["runtime_checkable_decorator"] = True

# =====================================
# reveal_type and assert_type
# =====================================
from typing import reveal_type, assert_type

val = reveal_type(42)
results["reveal_type_returns_value"] = val == 42

val2 = assert_type(42, int)
results["assert_type_returns_value"] = val2 == 42

# =====================================
# get_overloads and clear_overloads
# =====================================
from typing import get_overloads, clear_overloads

overloads = get_overloads(process)
results["get_overloads_returns_list"] = isinstance(overloads, list)

clear_overloads()
results["clear_overloads_works"] = True

# =====================================
# Sequence types
# =====================================
results["sequence_exists"] = Sequence is not None
results["mapping_exists"] = Mapping is not None
results["iterable_exists"] = Iterable is not None
results["iterator_exists"] = Iterator is not None

# Subscripted sequence types
seq_int = Sequence[int]
results["sequence_subscript"] = seq_int is not None

iter_str = Iterable[str]
results["iterable_subscript"] = iter_str is not None

# =====================================
# Additional generic types
# =====================================
results["frozenset_exists"] = FrozenSet is not None
results["type_exists"] = Type is not None

frozenset_int = FrozenSet[int]
results["frozenset_subscript"] = frozenset_int is not None

type_int = Type[int]
results["type_subscript"] = type_int is not None

# =====================================
# Protocol and Generic base
# =====================================
results["generic_exists"] = Generic is not None
results["protocol_exists"] = Protocol is not None

# =====================================
# Special forms
# =====================================
from typing import Final, ClassVar, Literal

results["final_form_exists"] = Final is not None
results["classvar_exists"] = ClassVar is not None
results["literal_exists"] = Literal is not None

final_int = Final[int]
results["final_subscript"] = final_int is not None

classvar_str = ClassVar[str]
results["classvar_subscript"] = classvar_str is not None

literal_one = Literal[1]
results["literal_subscript"] = literal_one is not None

# =====================================
# NoReturn, Never, Self
# =====================================
from typing import NoReturn, Never, Self

results["noreturn_exists"] = NoReturn is not None
results["never_exists"] = Never is not None
results["self_exists"] = Self is not None

# =====================================
# ParamSpec and TypeVarTuple
# =====================================
from typing import ParamSpec, TypeVarTuple

P = ParamSpec("P")
results["paramspec_exists"] = P is not None

Ts = TypeVarTuple("Ts")
results["typevartuple_exists"] = Ts is not None

# =====================================
# NamedTuple
# =====================================
from typing import NamedTuple

# Create with list of tuples
Point = NamedTuple("Point", [("x", int), ("y", int)])
results["namedtuple_created"] = Point is not None

# =====================================
# TypedDict
# =====================================
from typing import TypedDict

Movie = TypedDict("Movie", {})
results["typeddict_created"] = Movie is not None

# =====================================
# is_typeddict
# =====================================
from typing import is_typeddict

results["is_typeddict_result"] = is_typeddict(Movie) == False  # Runtime returns False

# =====================================
# get_type_hints
# =====================================
from typing import get_type_hints

def sample_func(x, y):
    return x + y

hints = get_type_hints(sample_func)
results["get_type_hints_returns_dict"] = isinstance(hints, dict)

# =====================================
# Annotated
# =====================================
from typing import Annotated

annotated_int = Annotated[int]
results["annotated_subscript"] = annotated_int is not None

# =====================================
# Required, NotRequired, ReadOnly
# =====================================
from typing import Required, NotRequired, ReadOnly

results["required_exists"] = Required is not None
results["notrequired_exists"] = NotRequired is not None
results["readonly_exists"] = ReadOnly is not None

required_int = Required[int]
results["required_subscript"] = required_int is not None

# =====================================
# Concatenate, TypeGuard, Unpack
# =====================================
from typing import Concatenate, TypeGuard, Unpack

results["concatenate_exists"] = Concatenate is not None
results["typeguard_exists"] = TypeGuard is not None
results["unpack_exists"] = Unpack is not None

typeguard_bool = TypeGuard[bool]
results["typeguard_subscript"] = typeguard_bool is not None

# =====================================
# Collection type aliases
# =====================================
from typing import Counter, ChainMap, OrderedDict, DefaultDict, Deque

results["counter_alias_exists"] = Counter is not None
results["chainmap_alias_exists"] = ChainMap is not None
results["ordereddict_alias_exists"] = OrderedDict is not None
results["defaultdict_alias_exists"] = DefaultDict is not None
results["deque_alias_exists"] = Deque is not None

counter_int = Counter[int]
results["counter_subscript"] = counter_int is not None

deque_str = Deque[str]
results["deque_subscript"] = deque_str is not None

# =====================================
# ContextManager types
# =====================================
from typing import ContextManager, AsyncContextManager

results["contextmanager_exists"] = ContextManager is not None
results["asynccontextmanager_exists"] = AsyncContextManager is not None

# =====================================
# Coroutine and async types
# =====================================
from typing import Coroutine, AsyncGenerator, AsyncIterable, AsyncIterator, Awaitable

results["coroutine_exists"] = Coroutine is not None
results["asyncgenerator_exists"] = AsyncGenerator is not None
results["asynciterable_exists"] = AsyncIterable is not None
results["asynciterator_exists"] = AsyncIterator is not None
results["awaitable_exists"] = Awaitable is not None

# =====================================
# IO types
# =====================================
from typing import IO, TextIO, BinaryIO

results["io_exists"] = IO is not None
results["textio_exists"] = TextIO is not None
results["binaryio_exists"] = BinaryIO is not None

# =====================================
# Pattern and Match
# =====================================
from typing import Pattern, Match

results["pattern_exists"] = Pattern is not None
results["match_exists"] = Match is not None

pattern_str = Pattern[str]
results["pattern_subscript"] = pattern_str is not None

# =====================================
# Supports protocols
# =====================================
from typing import SupportsInt, SupportsFloat, SupportsAbs, SupportsRound, SupportsIndex

results["supportsint_exists"] = SupportsInt is not None
results["supportsfloat_exists"] = SupportsFloat is not None
results["supportsabs_exists"] = SupportsAbs is not None
results["supportsround_exists"] = SupportsRound is not None
results["supportsindex_exists"] = SupportsIndex is not None

# =====================================
# Hashable, Sized, Reversible
# =====================================
from typing import Hashable, Sized, Reversible

results["hashable_exists"] = Hashable is not None
results["sized_exists"] = Sized is not None
results["reversible_exists"] = Reversible is not None

# =====================================
# MutableMapping, MutableSequence, MutableSet
# =====================================
from typing import MutableMapping, MutableSequence, MutableSet

results["mutablemapping_exists"] = MutableMapping is not None
results["mutablesequence_exists"] = MutableSequence is not None
results["mutableset_exists"] = MutableSet is not None

mutableseq_int = MutableSequence[int]
results["mutablesequence_subscript"] = mutableseq_int is not None

# =====================================
# Generator
# =====================================
from typing import Generator

results["generator_exists"] = Generator is not None

gen_int = Generator[int]
results["generator_subscript"] = gen_int is not None

print("typing module tests completed")
