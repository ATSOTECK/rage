# Test framework module for integration tests

__test_passed__ = 0
__test_failed__ = 0
__test_failures__ = ""

def expect(expected, actual):
    if expected != actual:
        raise Exception("Expected " + str(expected) + " but got " + str(actual))

def test(name, fn):
    global __test_passed__, __test_failed__, __test_failures__
    try:
        fn()
        __test_passed__ = __test_passed__ + 1
        print("[PASS] " + name)
    except Exception as e:
        __test_failed__ = __test_failed__ + 1
        __test_failures__ = __test_failures__ + name + ": " + str(e) + "\n"
        print("[FAIL] " + name + ": " + str(e))
