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

import os

def test_dump_and_load_basic():
    # Test basic dump and load with a dictionary
    data = {"name": "test", "value": 42, "active": True}
    filename = "/tmp/test_json_dump_basic.json"

    # Write using json.dump
    f = open(filename, "w")
    json.dump(data, f)
    f.close()

    # Read using json.load
    f = open(filename, "r")
    loaded = json.load(f)
    f.close()

    expect(loaded == data).to_be(True)

    # Cleanup
    os.remove(filename)

def test_dump_and_load_list():
    # Test with a list
    data = [1, 2, 3, "four", None, True]
    filename = "/tmp/test_json_dump_list.json"

    f = open(filename, "w")
    json.dump(data, f)
    f.close()

    f = open(filename, "r")
    loaded = json.load(f)
    f.close()

    expect(loaded == data).to_be(True)

    os.remove(filename)

def test_dump_with_indent():
    # Test dump with indentation
    data = {"x": 1, "y": 2}
    filename = "/tmp/test_json_dump_indent.json"

    f = open(filename, "w")
    json.dump(data, f, 2)
    f.close()

    f = open(filename, "r")
    content = f.read()
    f.close()

    expect("\n" in content).to_be(True)
    expect("  " in content).to_be(True)

    # Verify it can be loaded back
    f = open(filename, "r")
    loaded = json.load(f)
    f.close()

    expect(loaded == data).to_be(True)

    os.remove(filename)

def test_dump_complex_nested():
    # Test with complex nested structure
    data = {
        "users": [
            {"name": "Alice", "age": 30},
            {"name": "Bob", "age": 25}
        ],
        "metadata": {
            "version": 1,
            "active": True,
            "tags": ["test", "json"]
        }
    }
    filename = "/tmp/test_json_dump_complex.json"

    f = open(filename, "w")
    json.dump(data, f)
    f.close()

    f = open(filename, "r")
    loaded = json.load(f)
    f.close()

    expect(loaded == data).to_be(True)

    os.remove(filename)

def test_dump_with_context_manager():
    # Test using with statement (context manager)
    data = {"key": "value"}
    filename = "/tmp/test_json_dump_ctx.json"

    with open(filename, "w") as f:
        json.dump(data, f)

    with open(filename, "r") as f:
        loaded = json.load(f)

    expect(loaded == data).to_be(True)

    os.remove(filename)

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
test("dump_and_load_basic", test_dump_and_load_basic)
test("dump_and_load_list", test_dump_and_load_list)
test("dump_with_indent", test_dump_with_indent)
test("dump_complex_nested", test_dump_complex_nested)
test("dump_with_context_manager", test_dump_with_context_manager)

print("JSON module tests completed")
