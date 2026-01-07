# Test framework module for integration tests

__test_passed__ = 0
__test_failed__ = 0
__test_failures__ = ""
__test_use_color__ = False

# ANSI color codes
_COLOR_RESET = "\033[0m"
_COLOR_RED = "\033[31m"
_COLOR_GREEN = "\033[32m"
_COLOR_BOLD = "\033[1m"

def _color(code, text):
    if __test_use_color__:
        return f"{code}{text}{_COLOR_RESET}"
    return text

class Expectation:
    def __init__(self, actual):
        self.actual = actual

    def to_be(self, expected):
        if self.actual != expected:
            raise Exception("Expected " + str(expected) + " but got " + str(self.actual))

def expect(actual):
    return Expectation(actual)

def test(name, fn):
    global __test_passed__, __test_failed__, __test_failures__
    try:
        fn()
        __test_passed__ = __test_passed__ + 1
        status = _color(_COLOR_GREEN + _COLOR_BOLD, "✓ PASS")
        print(f"{status} {name}")
    except Exception as e:
        __test_failed__ = __test_failed__ + 1
        __test_failures__ = __test_failures__ + name + ": " + str(e) + "\n"
        status = _color(_COLOR_RED + _COLOR_BOLD, "✗ FAIL")
        error_msg = _color(_COLOR_RED, str(e))
        print(f"{status} {name}: {error_msg}")
