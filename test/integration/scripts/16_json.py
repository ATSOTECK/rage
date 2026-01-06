# Test: JSON Module
# Tests json.dumps and json.loads

from test_framework import test, expect

import json

def test_dumps_basic_types():
    expect(json.dumps(None)).to_be("null")
    expect(json.dumps(True)).to_be("true")
    expect(json.dumps(False)).to_be("false")
    expect(json.dumps(42)).to_be("42")
    expect(json.dumps(-17)).to_be("-17")
    expect(json.dumps(3.14159)).to_be("3.14159")
    expect(json.dumps("hello")).to_be('"hello"')
    expect(json.dumps("")).to_be('""')

def test_dumps_escapes():
    expect(json.dumps("line1\nline2")).to_be('"line1\\nline2"')
    expect(json.dumps("col1\tcol2")).to_be('"col1\\tcol2"')
    expect(json.dumps('say "hello"')).to_be('"say \\"hello\\""')
    expect(json.dumps("path\\to\\file")).to_be('"path\\\\to\\\\file"')

def test_dumps_collections():
    expect(json.dumps([])).to_be("[]")
    # RAGE json module doesn't include spaces after commas
    expect(json.dumps([1, 2, 3])).to_be("[1,2,3]")
    expect(json.dumps([[1, 2], [3, 4]])).to_be("[[1,2],[3,4]]")
    expect(json.dumps({})).to_be("{}")
    # Dict roundtrip (order not guaranteed)
    expect(json.loads(json.dumps({"a": 1, "b": 2})) == {"a": 1, "b": 2}).to_be(True)
    expect(json.loads(json.dumps({"outer": {"inner": "value"}})) == {"outer": {"inner": "value"}}).to_be(True)
    expect(json.dumps((1, 2, 3))).to_be("[1,2,3]")

def test_dumps_indent():
    indented = json.dumps({"x": 1, "y": 2}, 2)
    expect("\n" in indented).to_be(True)
    expect("  " in indented).to_be(True)

def test_dumps_sort_keys():
    sorted_json = json.dumps({"z": 1, "a": 2, "m": 3}, None, None, True)
    expect(json.loads(sorted_json) == {"z": 1, "a": 2, "m": 3}).to_be(True)

def test_loads_basic_types():
    expect(json.loads("null")).to_be(None)
    expect(json.loads("true")).to_be(True)
    expect(json.loads("false")).to_be(False)
    expect(json.loads("42")).to_be(42)
    expect(json.loads("-17")).to_be(-17)
    expect(json.loads("3.14159")).to_be(3.14159)
    expect(json.loads('"hello"')).to_be("hello")
    expect(json.loads('""')).to_be("")

def test_loads_escapes():
    expect(json.loads('"line1\\nline2"')).to_be("line1\nline2")
    expect(json.loads('"col1\\tcol2"')).to_be("col1\tcol2")
    expect(json.loads('"say \\"hello\\""')).to_be('say "hello"')
    expect(json.loads('"path\\\\to\\\\file"')).to_be("path\\to\\file")
    expect(json.loads('"\\u0048\\u0069"')).to_be("Hi")

def test_loads_collections():
    expect(json.loads("[]")).to_be([])
    expect(json.loads("[1, 2, 3]")).to_be([1, 2, 3])
    expect(json.loads("[[1, 2], [3, 4]]")).to_be([[1, 2], [3, 4]])
    expect(json.loads('[1, "two", true, null]')).to_be([1, "two", True, None])
    expect(json.loads("{}")).to_be({})
    expect(json.loads('{"a": 1, "b": 2}')).to_be({"a": 1, "b": 2})
    expect(json.loads('{"outer": {"inner": "value"}}')).to_be({"outer": {"inner": "value"}})

def test_loads_whitespace():
    expect(json.loads('  {  "x"  :  1  }  ')).to_be({"x": 1})
    expect(json.loads('{\n  "x": 1\n}')).to_be({"x": 1})

def test_roundtrip():
    original_list = [1, 2.5, "hello", None, True, False]
    expect(json.loads(json.dumps(original_list)) == original_list).to_be(True)

    original_dict = {"name": "test", "value": 123, "active": True}
    expect(json.loads(json.dumps(original_dict)) == original_dict).to_be(True)

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
    expect(json.loads(json.dumps(complex_data)) == complex_data).to_be(True)

def test_type_preservation():
    expect(json.loads("42") == 42).to_be(True)
    expect(json.loads("3.14") == 3.14).to_be(True)
    expect(json.loads("[1,2,3]") == [1, 2, 3]).to_be(True)

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
