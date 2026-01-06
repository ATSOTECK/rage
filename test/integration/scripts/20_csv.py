# Test: CSV Module
# Tests csv.reader, csv.writer, csv.DictReader, csv.DictWriter, and utility functions

from test_framework import test, expect

import csv

def test_parse_row_basic():
    expect(["a", "b", "c"], csv.parse_row("a,b,c"))
    expect(["1", "2", "3"], csv.parse_row("1,2,3"))
    expect(["name", "30", "city"], csv.parse_row("name,30,city"))
    expect(["a", "", "c"], csv.parse_row("a,,c"))
    expect(["only"], csv.parse_row("only"))

def test_parse_row_quoted():
    expect(["a", "b", "c"], csv.parse_row('a,"b",c'))
    expect(["a", "b,c", "d"], csv.parse_row('a,"b,c",d'))
    expect(["a", "", "c"], csv.parse_row('a,"",c'))
    expect(["a", 'say "hello"', "c"], csv.parse_row('a,"say ""hello""",c'))

def test_parse_row_delimiters():
    expect(["a", "b", "c"], csv.parse_row("a;b;c", ";"))
    expect(["a", "b", "c"], csv.parse_row("a\tb\tc", "\t"))
    expect(["a", "b", "c"], csv.parse_row("a|b|c", "|"))

def test_format_row_basic():
    expect("a,b,c", csv.format_row(["a", "b", "c"]))
    expect("1,2,3", csv.format_row([1, 2, 3]))
    expect("name,30,city", csv.format_row(["name", 30, "city"]))
    expect("", csv.format_row([]))
    expect("only", csv.format_row(["only"]))

def test_format_row_quoting():
    expect('a,"b,c",d', csv.format_row(["a", "b,c", "d"]))
    expect('a,"b\nc",d', csv.format_row(["a", "b\nc", "d"]))
    expect('a,"say ""hi""",c', csv.format_row(["a", 'say "hi"', "c"]))

def test_format_row_modes():
    expect('"a","b"', csv.format_row(["a", "b"], ",", '"', csv.QUOTE_ALL))
    expect("a,b", csv.format_row(["a", "b"], ",", '"', csv.QUOTE_NONE))

def test_format_row_delimiters():
    expect("a;b;c", csv.format_row(["a", "b", "c"], ";"))
    expect("a\tb\tc", csv.format_row(["a", "b", "c"], "\t"))

def test_reader_basic():
    lines = ["a,b,c", "1,2,3", "x,y,z"]
    reader = csv.reader(lines)
    rows = []
    for row in reader:
        rows.append(row)
    expect(3, len(rows))
    expect(["a", "b", "c"], rows[0])
    expect(["1", "2", "3"], rows[1])
    expect(["x", "y", "z"], rows[2])

def test_reader_with_header():
    csv_data = ["name,age,city", "Alice,30,NYC", "Bob,25,LA"]
    reader = csv.reader(csv_data)
    all_rows = []
    for row in reader:
        all_rows.append(row)
    expect(["name", "age", "city"], all_rows[0])
    expect(["Alice", "30", "NYC"], all_rows[1])
    expect(["Bob", "25", "LA"], all_rows[2])

def test_reader_quoted():
    quoted_lines = ['name,description', 'item1,"A simple item"', 'item2,"An item, with comma"']
    reader = csv.reader(quoted_lines)
    quoted_rows = []
    for row in reader:
        quoted_rows.append(row)
    expect(["name", "description"], quoted_rows[0])
    expect(["item1", "A simple item"], quoted_rows[1])
    expect(["item2", "An item, with comma"], quoted_rows[2])

def test_reader_delimiter():
    semicolon_lines = ["a;b;c", "1;2;3"]
    reader = csv.reader(semicolon_lines, ";")
    semicolon_rows = []
    for row in reader:
        semicolon_rows.append(row)
    expect(["a", "b", "c"], semicolon_rows[0])
    expect(["1", "2", "3"], semicolon_rows[1])

def test_dictreader_basic():
    csv_data = ["name,age,city", "Alice,30,NYC", "Bob,25,LA"]
    dreader = csv.DictReader(csv_data)
    dict_rows = []
    for row in dreader:
        dict_rows.append(row)
    expect(2, len(dict_rows))
    expect("Alice", dict_rows[0]["name"])
    expect("30", dict_rows[0]["age"])
    expect("NYC", dict_rows[0]["city"])
    expect("Bob", dict_rows[1]["name"])
    expect("25", dict_rows[1]["age"])
    expect("LA", dict_rows[1]["city"])

def test_dictreader_fieldnames():
    data_no_header = ["Alice,30,NYC", "Bob,25,LA"]
    dreader = csv.DictReader(data_no_header, ["name", "age", "city"])
    rows_with_fieldnames = []
    for row in dreader:
        rows_with_fieldnames.append(row)
    expect(2, len(rows_with_fieldnames))
    expect("Alice", rows_with_fieldnames[0]["name"])
    expect("LA", rows_with_fieldnames[1]["city"])

def test_writer_basic():
    w = csv.writer()
    w.writerow(["name", "age", "city"])
    w.writerow(["Alice", 30, "NYC"])
    w.writerow(["Bob", 25, "LA"])
    output = w.getvalue()
    expect(True, "name,age,city" in output)
    expect(True, "Alice,30,NYC" in output)
    expect(True, "Bob,25,LA" in output)
    expect(3, len(output.split("\n")) - 1)

def test_writer_writerows():
    w2 = csv.writer()
    w2.writerows([["a", "b"], ["c", "d"], ["e", "f"]])
    output2 = w2.getvalue()
    expect(True, "a,b" in output2)
    expect(True, "c,d" in output2)
    expect(True, "e,f" in output2)

def test_writer_delimiter():
    w3 = csv.writer(";")
    w3.writerow(["a", "b", "c"])
    output3 = w3.getvalue()
    expect(True, "a;b;c" in output3)

def test_writer_quote_all():
    w4 = csv.writer(",", '"', csv.QUOTE_ALL)
    w4.writerow(["a", "b"])
    output4 = w4.getvalue()
    expect(True, '"a","b"' in output4)

def test_dictwriter_basic():
    dw = csv.DictWriter(["name", "age", "city"])
    dw.writeheader()
    dw.writerow({"name": "Alice", "age": "30", "city": "NYC"})
    dw.writerow({"name": "Bob", "age": "25", "city": "LA"})
    dw_output = dw.getvalue()
    expect(True, "name,age,city" in dw_output)
    expect(True, "Alice,30,NYC" in dw_output)
    expect(True, "Bob,25,LA" in dw_output)

def test_dictwriter_writerows():
    dw2 = csv.DictWriter(["x", "y"])
    dw2.writeheader()
    dw2.writerows([{"x": "1", "y": "2"}, {"x": "3", "y": "4"}])
    dw2_output = dw2.getvalue()
    expect(True, "x,y" in dw2_output)
    expect(True, "1,2" in dw2_output)
    expect(True, "3,4" in dw2_output)

def test_dictwriter_restval():
    dw3 = csv.DictWriter(["a", "b", "c"], ",", '"', csv.QUOTE_MINIMAL, "\n", "N/A")
    dw3.writerow({"a": "1", "c": "3"})
    dw3_output = dw3.getvalue()
    expect(True, "1,N/A,3" in dw3_output)

def test_roundtrip():
    w_rt = csv.writer()
    w_rt.writerow(["name", "score"])
    w_rt.writerow(["Alice", "100"])
    w_rt.writerow(["Bob", "95"])
    csv_string = w_rt.getvalue()

    lines_rt = csv_string.strip().split("\n")
    reader_rt = csv.reader(lines_rt)
    roundtrip_rows = []
    for row in reader_rt:
        roundtrip_rows.append(row)

    expect(["name", "score"], roundtrip_rows[0])
    expect(["Alice", "100"], roundtrip_rows[1])
    expect(["Bob", "95"], roundtrip_rows[2])

def test_constants():
    expect(0, csv.QUOTE_MINIMAL)
    expect(1, csv.QUOTE_ALL)
    expect(2, csv.QUOTE_NONNUMERIC)
    expect(3, csv.QUOTE_NONE)

test("parse_row_basic", test_parse_row_basic)
test("parse_row_quoted", test_parse_row_quoted)
test("parse_row_delimiters", test_parse_row_delimiters)
test("format_row_basic", test_format_row_basic)
test("format_row_quoting", test_format_row_quoting)
test("format_row_modes", test_format_row_modes)
test("format_row_delimiters", test_format_row_delimiters)
test("reader_basic", test_reader_basic)
test("reader_with_header", test_reader_with_header)
test("reader_quoted", test_reader_quoted)
test("reader_delimiter", test_reader_delimiter)
test("dictreader_basic", test_dictreader_basic)
test("dictreader_fieldnames", test_dictreader_fieldnames)
test("writer_basic", test_writer_basic)
test("writer_writerows", test_writer_writerows)
test("writer_delimiter", test_writer_delimiter)
test("writer_quote_all", test_writer_quote_all)
test("dictwriter_basic", test_dictwriter_basic)
test("dictwriter_writerows", test_dictwriter_writerows)
test("dictwriter_restval", test_dictwriter_restval)
test("roundtrip", test_roundtrip)
test("constants", test_constants)

print("CSV module tests completed")
