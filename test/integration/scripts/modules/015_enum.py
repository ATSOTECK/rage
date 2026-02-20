# Test: Enum Module
# Tests Enum, IntEnum, StrEnum, Flag, IntFlag, auto(), and @unique

from test_framework import test, expect
from enum import Enum, IntEnum, StrEnum, Flag, IntFlag, auto, unique

# === Basic Enum ===

class Color(Enum):
    RED = 1
    GREEN = 2
    BLUE = 3

def test_basic_enum_access():
    expect(Color.RED.name).to_be("RED")
    expect(Color.RED.value).to_be(1)
    expect(Color.GREEN.name).to_be("GREEN")
    expect(Color.GREEN.value).to_be(2)
    expect(Color.BLUE.name).to_be("BLUE")
    expect(Color.BLUE.value).to_be(3)

test("basic enum member access", test_basic_enum_access)

# === Name-based lookup ===

def test_name_lookup():
    expect(Color['RED']).to_be(Color.RED)
    expect(Color['GREEN']).to_be(Color.GREEN)
    expect(Color['BLUE']).to_be(Color.BLUE)

test("enum name lookup with []", test_name_lookup)

# === Value-based lookup ===

def test_value_lookup():
    expect(Color(1)).to_be(Color.RED)
    expect(Color(2)).to_be(Color.GREEN)
    expect(Color(3)).to_be(Color.BLUE)

test("enum value lookup with ()", test_value_lookup)

# === Singleton identity ===

def test_singleton():
    expect(Color.RED is Color.RED).to_be(True)
    expect(Color(1) is Color.RED).to_be(True)
    expect(Color['RED'] is Color.RED).to_be(True)

test("enum singleton identity", test_singleton)

# === Iteration ===

def test_iteration():
    members = list(Color)
    expect(len(members)).to_be(3)
    expect(members[0]).to_be(Color.RED)
    expect(members[1]).to_be(Color.GREEN)
    expect(members[2]).to_be(Color.BLUE)

test("enum iteration", test_iteration)

# === Membership ===

def test_membership():
    expect(Color.RED in Color).to_be(True)
    expect(Color.GREEN in Color).to_be(True)

test("enum membership test", test_membership)

# === Repr and Str ===

def test_repr_str():
    expect(repr(Color.RED)).to_be("<Color.RED: 1>")
    expect(str(Color.RED)).to_be("Color.RED")

test("enum repr and str", test_repr_str)

# === Equality ===

def test_equality():
    expect(Color.RED == Color.RED).to_be(True)
    expect(Color.RED == Color.GREEN).to_be(False)
    expect(Color.RED != Color.GREEN).to_be(True)

test("enum equality", test_equality)

# === Invalid value lookup ===

def test_invalid_value():
    try:
        Color(99)
        expect(True).to_be(False)  # Should not reach here
    except ValueError:
        expect(True).to_be(True)

test("enum invalid value raises ValueError", test_invalid_value)

# === Invalid name lookup ===

def test_invalid_name():
    try:
        Color['PURPLE']
        expect(True).to_be(False)
    except KeyError:
        expect(True).to_be(True)

test("enum invalid name raises KeyError", test_invalid_name)

# === auto() ===

class Priority(Enum):
    LOW = auto()
    MEDIUM = auto()
    HIGH = auto()

def test_auto():
    expect(Priority.LOW.value).to_be(1)
    expect(Priority.MEDIUM.value).to_be(2)
    expect(Priority.HIGH.value).to_be(3)

test("auto() generates sequential values", test_auto)

# === Aliases ===

class Status(Enum):
    ACTIVE = 1
    RUNNING = 1   # alias for ACTIVE
    STOPPED = 2

def test_aliases():
    expect(Status.ACTIVE is Status.RUNNING).to_be(True)
    expect(Status.ACTIVE.name).to_be("ACTIVE")
    expect(Status.RUNNING.name).to_be("ACTIVE")  # alias uses original name
    expect(Status(1) is Status.ACTIVE).to_be(True)
    # Aliases are not in iteration
    members = list(Status)
    expect(len(members)).to_be(2)  # ACTIVE and STOPPED only

test("enum aliases", test_aliases)

# === @unique decorator ===

def test_unique_passes():
    @unique
    class Direction(Enum):
        NORTH = 1
        SOUTH = 2
        EAST = 3
        WEST = 4
    expect(Direction.NORTH.value).to_be(1)

test("@unique passes for unique values", test_unique_passes)

def test_unique_fails():
    try:
        @unique
        class BadEnum(Enum):
            A = 1
            B = 1
        expect(True).to_be(False)
    except ValueError:
        expect(True).to_be(True)

test("@unique fails for duplicate values", test_unique_fails)

# === IntEnum ===

class HttpStatus(IntEnum):
    OK = 200
    NOT_FOUND = 404
    SERVER_ERROR = 500

def test_intenum_comparison():
    expect(HttpStatus.OK == 200).to_be(True)
    expect(HttpStatus.OK < HttpStatus.NOT_FOUND).to_be(True)
    expect(HttpStatus.NOT_FOUND > 300).to_be(True)
    expect(HttpStatus.OK != 404).to_be(True)

test("IntEnum comparison with ints", test_intenum_comparison)

def test_intenum_arithmetic():
    expect(HttpStatus.OK + 4).to_be(204)
    expect(HttpStatus.NOT_FOUND - 4).to_be(400)
    expect(HttpStatus.OK * 2).to_be(400)

test("IntEnum arithmetic", test_intenum_arithmetic)

def test_intenum_int():
    expect(int(HttpStatus.OK)).to_be(200)
    expect(int(HttpStatus.NOT_FOUND)).to_be(404)

test("IntEnum int() conversion", test_intenum_int)

# === StrEnum ===

class Animal(StrEnum):
    CAT = "cat"
    DOG = "dog"
    BIRD = "bird"

def test_strenum():
    expect(Animal.CAT == "cat").to_be(True)
    expect(Animal.DOG == "dog").to_be(True)
    expect(str(Animal.CAT)).to_be("cat")

test("StrEnum comparison and str", test_strenum)

# === StrEnum with auto() ===

class Direction(StrEnum):
    NORTH = auto()
    SOUTH = auto()
    EAST = auto()

def test_strenum_auto():
    expect(Direction.NORTH.value).to_be("north")
    expect(Direction.SOUTH.value).to_be("south")
    expect(Direction.EAST.value).to_be("east")

test("StrEnum auto() gives lowercase name", test_strenum_auto)

# === Flag ===

class Perm(Flag):
    READ = auto()
    WRITE = auto()
    EXECUTE = auto()

def test_flag_auto():
    expect(Perm.READ.value).to_be(1)
    expect(Perm.WRITE.value).to_be(2)
    expect(Perm.EXECUTE.value).to_be(4)

test("Flag auto() generates powers of 2", test_flag_auto)

def test_flag_or():
    rw = Perm.READ | Perm.WRITE
    expect(rw.value).to_be(3)

test("Flag bitwise OR", test_flag_or)

def test_flag_and():
    rw = Perm.READ | Perm.WRITE
    result = rw & Perm.READ
    expect(result.value).to_be(1)

test("Flag bitwise AND", test_flag_and)

def test_flag_xor():
    rw = Perm.READ | Perm.WRITE
    result = rw ^ Perm.READ
    expect(result.value).to_be(2)

test("Flag bitwise XOR", test_flag_xor)

def test_flag_invert():
    result = ~Perm.READ
    expect(result.value).to_be(6)  # WRITE | EXECUTE = 2 | 4 = 6

test("Flag bitwise invert", test_flag_invert)

def test_flag_bool():
    expect(bool(Perm.READ)).to_be(True)

test("Flag bool", test_flag_bool)

# === IntFlag ===

class IntPerm(IntFlag):
    R = 4
    W = 2
    X = 1

def test_intflag():
    rw = IntPerm.R | IntPerm.W
    expect(rw.value).to_be(6)
    expect(IntPerm.R == 4).to_be(True)
    expect(IntPerm.R > IntPerm.X).to_be(True)

test("IntFlag basics", test_intflag)

# === Custom methods on enum ===

class Planet(Enum):
    MERCURY = 1
    VENUS = 2
    EARTH = 3

    def is_inner(self):
        return self.value <= 2

def test_custom_method():
    expect(Planet.MERCURY.is_inner()).to_be(True)
    expect(Planet.EARTH.is_inner()).to_be(False)

test("custom methods on enum", test_custom_method)

# === __members__ attribute ===

def test_members():
    members = Color.__members__
    expect(members['RED']).to_be(Color.RED)
    expect(members['GREEN']).to_be(Color.GREEN)
    expect(members['BLUE']).to_be(Color.BLUE)

test("__members__ attribute", test_members)

# === Hash works correctly ===

def test_hash():
    d = {Color.RED: "red", Color.GREEN: "green"}
    expect(d[Color.RED]).to_be("red")
    expect(d[Color.GREEN]).to_be("green")

test("enum members as dict keys", test_hash)

# === IntEnum auto() ===

class Level(IntEnum):
    LOW = auto()
    MED = auto()
    HIGH = auto()

def test_intenum_auto():
    expect(Level.LOW.value).to_be(1)
    expect(Level.MED.value).to_be(2)
    expect(Level.HIGH.value).to_be(3)
    expect(Level.LOW == 1).to_be(True)

test("IntEnum auto() values", test_intenum_auto)
