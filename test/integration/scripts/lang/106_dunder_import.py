from test_framework import test, expect

# Test 1: __import__ basic module import
def test_import_basic():
    mod = __import__("math")
    expect(mod is not None).to_be(True)
    expect(hasattr(mod, "sqrt")).to_be(True)
test("__import__ basic module", test_import_basic)

# Test 2: __import__ returns working module
def test_import_usable():
    mod = __import__("math")
    result = mod.sqrt(16)
    expect(result).to_be(4.0)
test("__import__ returns usable module", test_import_usable)

# Test 3: __import__ with fromlist returns target module
def test_import_fromlist():
    mod = __import__("math", fromlist=["sqrt"])
    expect(hasattr(mod, "sqrt")).to_be(True)
test("__import__ with fromlist", test_import_fromlist)

# Test 4: __import__ different modules
def test_import_string():
    mod = __import__("string")
    expect(hasattr(mod, "ascii_letters")).to_be(True)
test("__import__ string module", test_import_string)

# Test 5: __import__ json module
def test_import_json():
    mod = __import__("json")
    result = mod.dumps([1, 2, 3])
    expect(result).to_be("[1,2,3]")
test("__import__ json module", test_import_json)

# Test 6: __import__ returns same cached module
def test_import_cached():
    mod1 = __import__("math")
    mod2 = __import__("math")
    expect(mod1 is mod2).to_be(True)
test("__import__ returns cached module", test_import_cached)

# Test 7: __import__ with invalid name
def test_import_invalid():
    got_error = False
    try:
        __import__("nonexistent_module_xyz")
    except Exception:
        got_error = True
    expect(got_error).to_be(True)
test("__import__ invalid module raises error", test_import_invalid)

# Test 8: __import__ requires string argument
def test_import_type_check():
    got_error = False
    try:
        __import__(42)
    except TypeError:
        got_error = True
    expect(got_error).to_be(True)
test("__import__ requires string argument", test_import_type_check)
