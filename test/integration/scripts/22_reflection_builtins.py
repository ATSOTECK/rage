# Test: Reflection and Execution Builtins
# Tests repr, dir, globals, locals, vars, compile, exec, eval

from test_framework import test, expect

# Helper class for testing
class TestClass:
    class_attr = "class_value"

    def __init__(self):
        self.instance_attr = "instance_value"
        self.x = 10
        self.y = 20

    def method(self):
        return self.x + self.y

class CustomRepr:
    def __repr__(self):
        return "CustomRepr()"

# ===== repr() tests =====

def test_repr_basic_types():
    expect(repr(None)).to_be("None")
    expect(repr(True)).to_be("True")
    expect(repr(False)).to_be("False")
    expect(repr(42)).to_be("42")
    expect(repr(3.14)).to_be("3.14")

def test_repr_strings():
    # repr adds quotes around strings
    expect(repr("hello")).to_be("'hello'")
    expect(repr("")).to_be("''")

def test_repr_collections():
    expect(repr([1, 2, 3])).to_be("[1, 2, 3]")
    expect(repr([])).to_be("[]")
    expect(repr((1, 2))).to_be("(1, 2)")
    expect(repr((1,))).to_be("(1,)")
    expect(repr(())).to_be("()")

def test_repr_custom():
    obj = CustomRepr()
    expect(repr(obj)).to_be("CustomRepr()")

# ===== dir() tests =====

def test_dir_with_object():
    obj = TestClass()
    d = dir(obj)
    # Should include instance attributes
    expect("instance_attr" in d).to_be(True)
    expect("x" in d).to_be(True)
    expect("y" in d).to_be(True)
    # Should include class attributes
    expect("class_attr" in d).to_be(True)
    expect("method" in d).to_be(True)

def test_dir_with_class():
    d = dir(TestClass)
    expect("class_attr" in d).to_be(True)
    expect("method" in d).to_be(True)

def test_dir_no_args():
    # dir() without args returns names in current scope
    local_var = 42
    d = dir()
    # Should include local variables and builtins
    expect("local_var" in d).to_be(True)

# ===== globals() tests =====

def test_globals_basic():
    g = globals()
    # Should include module-level definitions
    expect("TestClass" in g).to_be(True)
    expect("CustomRepr" in g).to_be(True)

def test_globals_access():
    global_test_var = 100
    g = globals()
    g["global_test_var"] = 100
    expect(g["global_test_var"]).to_be(100)

# ===== locals() tests =====

def test_locals_basic():
    x = 10
    y = 20
    z = 30
    loc = locals()
    expect(loc["x"]).to_be(10)
    expect(loc["y"]).to_be(20)
    expect(loc["z"]).to_be(30)

def test_locals_in_function():
    def inner():
        a = 1
        b = 2
        return locals()

    result = inner()
    expect(result["a"]).to_be(1)
    expect(result["b"]).to_be(2)

# ===== vars() tests =====

def test_vars_with_object():
    obj = TestClass()
    v = vars(obj)
    expect(v["instance_attr"]).to_be("instance_value")
    expect(v["x"]).to_be(10)
    expect(v["y"]).to_be(20)

def test_vars_with_class():
    v = vars(TestClass)
    expect(v["class_attr"]).to_be("class_value")

def test_vars_no_args():
    # vars() without args is like locals()
    vars_test_var = 42
    v = vars()
    expect(v["vars_test_var"]).to_be(42)

# ===== eval() tests =====

def test_eval_simple_expression():
    result = eval("2 + 3")
    expect(result).to_be(5)

def test_eval_with_variables():
    x = 10
    y = 5
    result = eval("x * y")
    expect(result).to_be(50)

def test_eval_complex_expression():
    result = eval("2 ** 10")
    expect(result).to_be(1024)

def test_eval_with_globals():
    g = {"a": 100, "b": 50}
    result = eval("a + b", g)
    expect(result).to_be(150)

def test_eval_string_expression():
    result = eval("'hello' + ' ' + 'world'")
    expect(result).to_be("hello world")

def test_eval_list_expression():
    result = eval("[1, 2, 3] + [4, 5]")
    expect(result).to_be([1, 2, 3, 4, 5])

# ===== exec() tests =====

def test_exec_simple():
    exec("test_exec_var = 42")
    expect(test_exec_var).to_be(42)

def test_exec_multiple_statements():
    exec("a = 10\nb = 20\nc = a + b")
    expect(c).to_be(30)

def test_exec_with_globals():
    g = {}
    exec("x = 100", g)
    expect(g["x"]).to_be(100)

def test_exec_function_def():
    exec("def add_five(n): return n + 5")
    expect(add_five(10)).to_be(15)

def test_exec_class_def():
    exec("class Point:\n    def __init__(self, x):\n        self.x = x")
    p = Point(5)
    expect(p.x).to_be(5)

# ===== compile() tests =====

def test_compile_exec_mode():
    code = compile("result = 42", "<test>", "exec")
    exec(code)
    expect(result).to_be(42)

def test_compile_eval_mode():
    code = compile("10 + 20", "<test>", "eval")
    result = eval(code)
    expect(result).to_be(30)

def test_compile_reuse():
    # Compile once, execute multiple times
    code = compile("counter = counter + 1", "<test>", "exec")
    counter = 0
    exec(code)
    expect(counter).to_be(1)
    exec(code)
    expect(counter).to_be(2)
    exec(code)
    expect(counter).to_be(3)

# Run all tests
test("repr basic types", test_repr_basic_types)
test("repr strings", test_repr_strings)
test("repr collections", test_repr_collections)
test("repr custom", test_repr_custom)
test("dir with object", test_dir_with_object)
test("dir with class", test_dir_with_class)
test("dir no args", test_dir_no_args)
test("globals basic", test_globals_basic)
test("globals access", test_globals_access)
test("locals basic", test_locals_basic)
test("locals in function", test_locals_in_function)
test("vars with object", test_vars_with_object)
test("vars with class", test_vars_with_class)
test("vars no args", test_vars_no_args)
test("eval simple expression", test_eval_simple_expression)
test("eval with variables", test_eval_with_variables)
test("eval complex expression", test_eval_complex_expression)
test("eval with globals", test_eval_with_globals)
test("eval string expression", test_eval_string_expression)
test("eval list expression", test_eval_list_expression)
test("exec simple", test_exec_simple)
test("exec multiple statements", test_exec_multiple_statements)
test("exec with globals", test_exec_with_globals)
test("exec function def", test_exec_function_def)
test("exec class def", test_exec_class_def)
test("compile exec mode", test_compile_exec_mode)
test("compile eval mode", test_compile_eval_mode)
test("compile reuse", test_compile_reuse)
