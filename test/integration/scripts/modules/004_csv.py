# Test: CSV Module
# Tests csv.reader, csv.writer, csv.DictReader, csv.DictWriter, and utility functions

from test_framework import test, expect

import csv

def test_parse_row_basic():
    expect(csv.parse_row("a,b,c")).to_be(["a", "b", "c"])
    expect(csv.parse_row("1,2,3")).to_be(["1", "2", "3"])
    expect(csv.parse_row("name,30,city")).to_be(["name", "30", "city"])
    expect(csv.parse_row("a,,c")).to_be(["a", "", "c"])
    expect(csv.parse_row("only")).to_be(["only"])

def test_parse_row_quoted():
    expect(csv.parse_row('a,"b",c')).to_be(["a", "b", "c"])
    expect(csv.parse_row('a,"b,c",d')).to_be(["a", "b,c", "d"])
    expect(csv.parse_row('a,"",c')).to_be(["a", "", "c"])
    expect(csv.parse_row('a,"say ""hello""",c')).to_be(["a", 'say "hello"', "c"])

def test_parse_row_delimiters():
    expect(csv.parse_row("a;b;c", ";")).to_be(["a", "b", "c"])
    expect(csv.parse_row("a\tb\tc", "\t")).to_be(["a", "b", "c"])
    expect(csv.parse_row("a|b|c", "|")).to_be(["a", "b", "c"])

def test_format_row_basic():
    expect(csv.format_row(["a", "b", "c"])).to_be("a,b,c")
    expect(csv.format_row([1, 2, 3])).to_be("1,2,3")
    expect(csv.format_row(["name", 30, "city"])).to_be("name,30,city")
    expect(csv.format_row([])).to_be("")
    expect(csv.format_row(["only"])).to_be("only")

def test_format_row_quoting():
    expect(csv.format_row(["a", "b,c", "d"])).to_be('a,"b,c",d')
    expect(csv.format_row(["a", "b\nc", "d"])).to_be('a,"b\nc",d')
    expect(csv.format_row(["a", 'say "hi"', "c"])).to_be('a,"say ""hi""",c')

def test_format_row_modes():
    expect(csv.format_row(["a", "b"], ",", '"', csv.QUOTE_ALL)).to_be('"a","b"')
    expect(csv.format_row(["a", "b"], ",", '"', csv.QUOTE_NONE)).to_be("a,b")

def test_format_row_delimiters():
    expect(csv.format_row(["a", "b", "c"], ";")).to_be("a;b;c")
    expect(csv.format_row(["a", "b", "c"], "\t")).to_be("a\tb\tc")

def test_reader_basic():
    lines = ["a,b,c", "1,2,3", "x,y,z"]
    reader = csv.reader(lines)
    rows = []
    for row in reader:
        rows.append(row)
    expect(len(rows)).to_be(3)
    expect(rows[0]).to_be(["a", "b", "c"])
    expect(rows[1]).to_be(["1", "2", "3"])
    expect(rows[2]).to_be(["x", "y", "z"])

def test_reader_with_header():
    csv_data = ["name,age,city", "Alice,30,NYC", "Bob,25,LA"]
    reader = csv.reader(csv_data)
    all_rows = []
    for row in reader:
        all_rows.append(row)
    expect(all_rows[0]).to_be(["name", "age", "city"])
    expect(all_rows[1]).to_be(["Alice", "30", "NYC"])
    expect(all_rows[2]).to_be(["Bob", "25", "LA"])

def test_reader_quoted():
    quoted_lines = ['name,description', 'item1,"A simple item"', 'item2,"An item, with comma"']
    reader = csv.reader(quoted_lines)
    quoted_rows = []
    for row in reader:
        quoted_rows.append(row)
    expect(quoted_rows[0]).to_be(["name", "description"])
    expect(quoted_rows[1]).to_be(["item1", "A simple item"])
    expect(quoted_rows[2]).to_be(["item2", "An item, with comma"])

def test_reader_delimiter():
    semicolon_lines = ["a;b;c", "1;2;3"]
    reader = csv.reader(semicolon_lines, ";")
    semicolon_rows = []
    for row in reader:
        semicolon_rows.append(row)
    expect(semicolon_rows[0]).to_be(["a", "b", "c"])
    expect(semicolon_rows[1]).to_be(["1", "2", "3"])

def test_dictreader_basic():
    csv_data = ["name,age,city", "Alice,30,NYC", "Bob,25,LA"]
    dreader = csv.DictReader(csv_data)
    dict_rows = []
    for row in dreader:
        dict_rows.append(row)
    expect(len(dict_rows)).to_be(2)
    expect(dict_rows[0]["name"]).to_be("Alice")
    expect(dict_rows[0]["age"]).to_be("30")
    expect(dict_rows[0]["city"]).to_be("NYC")
    expect(dict_rows[1]["name"]).to_be("Bob")
    expect(dict_rows[1]["age"]).to_be("25")
    expect(dict_rows[1]["city"]).to_be("LA")

def test_dictreader_fieldnames():
    data_no_header = ["Alice,30,NYC", "Bob,25,LA"]
    dreader = csv.DictReader(data_no_header, ["name", "age", "city"])
    rows_with_fieldnames = []
    for row in dreader:
        rows_with_fieldnames.append(row)
    expect(len(rows_with_fieldnames)).to_be(2)
    expect(rows_with_fieldnames[0]["name"]).to_be("Alice")
    expect(rows_with_fieldnames[1]["city"]).to_be("LA")

def test_writer_basic():
    w = csv.writer()
    w.writerow(["name", "age", "city"])
    w.writerow(["Alice", 30, "NYC"])
    w.writerow(["Bob", 25, "LA"])
    output = w.getvalue()
    expect("name,age,city" in output).to_be(True)
    expect("Alice,30,NYC" in output).to_be(True)
    expect("Bob,25,LA" in output).to_be(True)
    expect(len(output.split("\n")) - 1).to_be(3)

def test_writer_writerows():
    w2 = csv.writer()
    w2.writerows([["a", "b"], ["c", "d"], ["e", "f"]])
    output2 = w2.getvalue()
    expect("a,b" in output2).to_be(True)
    expect("c,d" in output2).to_be(True)
    expect("e,f" in output2).to_be(True)

def test_writer_delimiter():
    w3 = csv.writer(";")
    w3.writerow(["a", "b", "c"])
    output3 = w3.getvalue()
    expect("a;b;c" in output3).to_be(True)

def test_writer_quote_all():
    w4 = csv.writer(",", '"', csv.QUOTE_ALL)
    w4.writerow(["a", "b"])
    output4 = w4.getvalue()
    expect('"a","b"' in output4).to_be(True)

def test_dictwriter_basic():
    dw = csv.DictWriter(["name", "age", "city"])
    dw.writeheader()
    dw.writerow({"name": "Alice", "age": "30", "city": "NYC"})
    dw.writerow({"name": "Bob", "age": "25", "city": "LA"})
    dw_output = dw.getvalue()
    expect("name,age,city" in dw_output).to_be(True)
    expect("Alice,30,NYC" in dw_output).to_be(True)
    expect("Bob,25,LA" in dw_output).to_be(True)

def test_dictwriter_writerows():
    dw2 = csv.DictWriter(["x", "y"])
    dw2.writeheader()
    dw2.writerows([{"x": "1", "y": "2"}, {"x": "3", "y": "4"}])
    dw2_output = dw2.getvalue()
    expect("x,y" in dw2_output).to_be(True)
    expect("1,2" in dw2_output).to_be(True)
    expect("3,4" in dw2_output).to_be(True)

def test_dictwriter_restval():
    dw3 = csv.DictWriter(["a", "b", "c"], ",", '"', csv.QUOTE_MINIMAL, "\n", "N/A")
    dw3.writerow({"a": "1", "c": "3"})
    dw3_output = dw3.getvalue()
    expect("1,N/A,3" in dw3_output).to_be(True)

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

    expect(roundtrip_rows[0]).to_be(["name", "score"])
    expect(roundtrip_rows[1]).to_be(["Alice", "100"])
    expect(roundtrip_rows[2]).to_be(["Bob", "95"])

def test_constants():
    expect(csv.QUOTE_MINIMAL).to_be(0)
    expect(csv.QUOTE_ALL).to_be(1)
    expect(csv.QUOTE_NONNUMERIC).to_be(2)
    expect(csv.QUOTE_NONE).to_be(3)

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
