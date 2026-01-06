# Test: JSON Module
# Tests json.dumps and json.loads

results = {}

import json

# =====================================
# json.dumps - basic types
# =====================================

results["dumps_none"] = json.dumps(None)
results["dumps_true"] = json.dumps(True)
results["dumps_false"] = json.dumps(False)
results["dumps_int"] = json.dumps(42)
results["dumps_negative_int"] = json.dumps(-17)
results["dumps_float"] = json.dumps(3.14159)
results["dumps_string"] = json.dumps("hello")
results["dumps_empty_string"] = json.dumps("")

# =====================================
# json.dumps - escape sequences
# =====================================

results["dumps_newline"] = json.dumps("line1\nline2")
results["dumps_tab"] = json.dumps("col1\tcol2")
results["dumps_quote"] = json.dumps('say "hello"')
results["dumps_backslash"] = json.dumps("path\\to\\file")

# =====================================
# json.dumps - collections
# =====================================

results["dumps_empty_list"] = json.dumps([])
results["dumps_list"] = json.dumps([1, 2, 3])
results["dumps_nested_list"] = json.dumps([[1, 2], [3, 4]])
results["dumps_mixed_list"] = json.dumps([1, "two", True, None])

results["dumps_empty_dict"] = json.dumps({})
results["dumps_dict"] = json.dumps({"a": 1, "b": 2})
results["dumps_nested_dict"] = json.dumps({"outer": {"inner": "value"}})

results["dumps_tuple"] = json.dumps((1, 2, 3))

# =====================================
# json.dumps - with indent
# =====================================

indented = json.dumps({"x": 1, "y": 2}, 2)
results["dumps_indent_has_newlines"] = "\n" in indented
results["dumps_indent_has_spaces"] = "  " in indented

# =====================================
# json.dumps - with sort_keys
# =====================================

# sort_keys test - just verify it produces valid JSON that can be parsed back
sorted_json = json.dumps({"z": 1, "a": 2, "m": 3}, None, None, True)
results["dumps_sort_keys_valid"] = json.loads(sorted_json) == {"z": 1, "a": 2, "m": 3}

# =====================================
# json.loads - basic types
# =====================================

results["loads_null"] = json.loads("null")
results["loads_true"] = json.loads("true")
results["loads_false"] = json.loads("false")
results["loads_int"] = json.loads("42")
results["loads_negative_int"] = json.loads("-17")
results["loads_float"] = json.loads("3.14159")
results["loads_string"] = json.loads('"hello"')
results["loads_empty_string"] = json.loads('""')

# =====================================
# json.loads - escape sequences
# =====================================

results["loads_newline"] = json.loads('"line1\\nline2"')
results["loads_tab"] = json.loads('"col1\\tcol2"')
results["loads_quote"] = json.loads('"say \\"hello\\""')
results["loads_backslash"] = json.loads('"path\\\\to\\\\file"')
results["loads_unicode"] = json.loads('"\\u0048\\u0069"')

# =====================================
# json.loads - collections
# =====================================

results["loads_empty_list"] = json.loads("[]")
results["loads_list"] = json.loads("[1, 2, 3]")
results["loads_nested_list"] = json.loads("[[1, 2], [3, 4]]")
results["loads_mixed_list"] = json.loads('[1, "two", true, null]')

results["loads_empty_dict"] = json.loads("{}")
results["loads_dict"] = json.loads('{"a": 1, "b": 2}')
results["loads_nested_dict"] = json.loads('{"outer": {"inner": "value"}}')

# =====================================
# json.loads - whitespace handling
# =====================================

results["loads_with_spaces"] = json.loads('  {  "x"  :  1  }  ')
results["loads_with_newlines"] = json.loads('{\n  "x": 1\n}')

# =====================================
# Round-trip tests
# =====================================

original_list = [1, 2.5, "hello", None, True, False]
roundtrip_list = json.loads(json.dumps(original_list))
results["roundtrip_list"] = roundtrip_list == original_list

original_dict = {"name": "test", "value": 123, "active": True}
roundtrip_dict = json.loads(json.dumps(original_dict))
results["roundtrip_dict"] = roundtrip_dict == original_dict

# Complex nested structure
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
roundtrip_complex = json.loads(json.dumps(complex_data))
results["roundtrip_complex"] = roundtrip_complex == complex_data

# =====================================
# Type preservation
# =====================================

results["type_int_preserved"] = json.loads("42") == 42
results["type_float_preserved"] = json.loads("3.14") == 3.14
results["type_list_is_list"] = json.loads("[1,2,3]") == [1, 2, 3]

print("JSON module tests completed")
