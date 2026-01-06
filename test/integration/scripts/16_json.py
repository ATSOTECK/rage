# Test: JSON Module
# Tests json.dumps and json.loads

import json

def test_dumps_basic_types():
    expect("null", json.dumps(None))
    expect("true", json.dumps(True))
    expect("false", json.dumps(False))
    expect("42", json.dumps(42))
    expect("-17", json.dumps(-17))
    expect("3.14159", json.dumps(3.14159))
    expect('"hello"', json.dumps("hello"))
    expect('""', json.dumps(""))

def test_dumps_escapes():
    expect('"line1\\nline2"', json.dumps("line1\nline2"))
    expect('"col1\\tcol2"', json.dumps("col1\tcol2"))
    expect('"say \\"hello\\""', json.dumps('say "hello"'))
    expect('"path\\\\to\\\\file"', json.dumps("path\\to\\file"))

def test_dumps_collections():
    expect("[]", json.dumps([]))
    # RAGE json module doesn't include spaces after commas
    expect("[1,2,3]", json.dumps([1, 2, 3]))
    expect("[[1,2],[3,4]]", json.dumps([[1, 2], [3, 4]]))
    expect("{}", json.dumps({}))
    # Dict roundtrip (order not guaranteed)
    expect(True, json.loads(json.dumps({"a": 1, "b": 2})) == {"a": 1, "b": 2})
    expect(True, json.loads(json.dumps({"outer": {"inner": "value"}})) == {"outer": {"inner": "value"}})
    expect("[1,2,3]", json.dumps((1, 2, 3)))

def test_dumps_indent():
    indented = json.dumps({"x": 1, "y": 2}, 2)
    expect(True, "\n" in indented)
    expect(True, "  " in indented)

def test_dumps_sort_keys():
    sorted_json = json.dumps({"z": 1, "a": 2, "m": 3}, None, None, True)
    expect(True, json.loads(sorted_json) == {"z": 1, "a": 2, "m": 3})

def test_loads_basic_types():
    expect(None, json.loads("null"))
    expect(True, json.loads("true"))
    expect(False, json.loads("false"))
    expect(42, json.loads("42"))
    expect(-17, json.loads("-17"))
    expect(3.14159, json.loads("3.14159"))
    expect("hello", json.loads('"hello"'))
    expect("", json.loads('""'))

def test_loads_escapes():
    expect("line1\nline2", json.loads('"line1\\nline2"'))
    expect("col1\tcol2", json.loads('"col1\\tcol2"'))
    expect('say "hello"', json.loads('"say \\"hello\\""'))
    expect("path\\to\\file", json.loads('"path\\\\to\\\\file"'))
    expect("Hi", json.loads('"\\u0048\\u0069"'))

def test_loads_collections():
    expect([], json.loads("[]"))
    expect([1, 2, 3], json.loads("[1, 2, 3]"))
    expect([[1, 2], [3, 4]], json.loads("[[1, 2], [3, 4]]"))
    expect([1, "two", True, None], json.loads('[1, "two", true, null]'))
    expect({}, json.loads("{}"))
    expect({"a": 1, "b": 2}, json.loads('{"a": 1, "b": 2}'))
    expect({"outer": {"inner": "value"}}, json.loads('{"outer": {"inner": "value"}}'))

def test_loads_whitespace():
    expect({"x": 1}, json.loads('  {  "x"  :  1  }  '))
    expect({"x": 1}, json.loads('{\n  "x": 1\n}'))

def test_roundtrip():
    original_list = [1, 2.5, "hello", None, True, False]
    expect(True, json.loads(json.dumps(original_list)) == original_list)

    original_dict = {"name": "test", "value": 123, "active": True}
    expect(True, json.loads(json.dumps(original_dict)) == original_dict)

    complex_data = {
        "users": [
            {"name": "Alice", "age": 30},
            {"name": "Bob", "age": 25}
        ],
        "metadata": {
            "version": 1,
            "active": True
        }
    }
    expect(True, json.loads(json.dumps(complex_data)) == complex_data)

def test_type_preservation():
    expect(True, json.loads("42") == 42)
    expect(True, json.loads("3.14") == 3.14)
    expect(True, json.loads("[1,2,3]") == [1, 2, 3])

test("dumps_basic_types", test_dumps_basic_types)
test("dumps_escapes", test_dumps_escapes)
test("dumps_collections", test_dumps_collections)
test("dumps_indent", test_dumps_indent)
test("dumps_sort_keys", test_dumps_sort_keys)
test("loads_basic_types", test_loads_basic_types)
test("loads_escapes", test_loads_escapes)
test("loads_collections", test_loads_collections)
test("loads_whitespace", test_loads_whitespace)
test("roundtrip", test_roundtrip)
test("type_preservation", test_type_preservation)

print("JSON module tests completed")
