# Test: CSV Module
# Tests csv.reader, csv.writer, csv.DictReader, csv.DictWriter, and utility functions

results = {}

import csv

# =====================================
# csv.parse_row - basic parsing
# =====================================

results["parse_row_simple"] = csv.parse_row("a,b,c")
results["parse_row_numbers"] = csv.parse_row("1,2,3")
results["parse_row_mixed"] = csv.parse_row("name,30,city")
results["parse_row_empty_fields"] = csv.parse_row("a,,c")
results["parse_row_single"] = csv.parse_row("only")

# =====================================
# csv.parse_row - quoted fields
# =====================================

results["parse_row_quoted"] = csv.parse_row('a,"b",c')
results["parse_row_quoted_comma"] = csv.parse_row('a,"b,c",d')
results["parse_row_quoted_empty"] = csv.parse_row('a,"",c')
results["parse_row_escaped_quote"] = csv.parse_row('a,"say ""hello""",c')

# =====================================
# csv.parse_row - custom delimiter
# =====================================

results["parse_row_semicolon"] = csv.parse_row("a;b;c", ";")
results["parse_row_tab"] = csv.parse_row("a\tb\tc", "\t")
results["parse_row_pipe"] = csv.parse_row("a|b|c", "|")

# =====================================
# csv.format_row - basic formatting
# =====================================

results["format_row_strings"] = csv.format_row(["a", "b", "c"])
results["format_row_numbers"] = csv.format_row([1, 2, 3])
results["format_row_mixed"] = csv.format_row(["name", 30, "city"])
results["format_row_empty_list"] = csv.format_row([])
results["format_row_single"] = csv.format_row(["only"])

# =====================================
# csv.format_row - quoting
# =====================================

results["format_row_needs_quote"] = csv.format_row(["a", "b,c", "d"])
results["format_row_has_newline"] = csv.format_row(["a", "b\nc", "d"])
results["format_row_has_quote"] = csv.format_row(["a", 'say "hi"', "c"])

# =====================================
# csv.format_row - quoting modes
# =====================================

results["format_row_quote_all"] = csv.format_row(["a", "b"], ",", '"', csv.QUOTE_ALL)
results["format_row_quote_none"] = csv.format_row(["a", "b"], ",", '"', csv.QUOTE_NONE)

# =====================================
# csv.format_row - custom delimiter
# =====================================

results["format_row_semicolon"] = csv.format_row(["a", "b", "c"], ";")
results["format_row_tab"] = csv.format_row(["a", "b", "c"], "\t")

# =====================================
# csv.reader - basic iteration
# =====================================

lines = ["a,b,c", "1,2,3", "x,y,z"]
reader = csv.reader(lines)
rows = []
for row in reader:
    rows.append(row)

results["reader_row_count"] = len(rows)
results["reader_first_row"] = rows[0]
results["reader_second_row"] = rows[1]
results["reader_third_row"] = rows[2]

# =====================================
# csv.reader - with header
# =====================================

csv_data = ["name,age,city", "Alice,30,NYC", "Bob,25,LA"]
reader = csv.reader(csv_data)
all_rows = []
for row in reader:
    all_rows.append(row)

results["reader_header"] = all_rows[0]
results["reader_data_row1"] = all_rows[1]
results["reader_data_row2"] = all_rows[2]

# =====================================
# csv.reader - quoted fields
# =====================================

quoted_lines = ['name,description', 'item1,"A simple item"', 'item2,"An item, with comma"']
reader = csv.reader(quoted_lines)
quoted_rows = []
for row in reader:
    quoted_rows.append(row)

results["reader_quoted_header"] = quoted_rows[0]
results["reader_quoted_simple"] = quoted_rows[1]
results["reader_quoted_with_comma"] = quoted_rows[2]

# =====================================
# csv.reader - custom delimiter
# =====================================

semicolon_lines = ["a;b;c", "1;2;3"]
reader = csv.reader(semicolon_lines, ";")
semicolon_rows = []
for row in reader:
    semicolon_rows.append(row)

results["reader_semicolon_row1"] = semicolon_rows[0]
results["reader_semicolon_row2"] = semicolon_rows[1]

# =====================================
# csv.DictReader - basic usage
# =====================================

csv_data = ["name,age,city", "Alice,30,NYC", "Bob,25,LA"]
dreader = csv.DictReader(csv_data)
dict_rows = []
for row in dreader:
    dict_rows.append(row)

results["dictreader_count"] = len(dict_rows)
# Check that rows are dicts with correct keys
results["dictreader_row1_name"] = dict_rows[0]["name"]
results["dictreader_row1_age"] = dict_rows[0]["age"]
results["dictreader_row1_city"] = dict_rows[0]["city"]
results["dictreader_row2_name"] = dict_rows[1]["name"]
results["dictreader_row2_age"] = dict_rows[1]["age"]
results["dictreader_row2_city"] = dict_rows[1]["city"]

# =====================================
# csv.DictReader - with provided fieldnames
# =====================================

data_no_header = ["Alice,30,NYC", "Bob,25,LA"]
dreader = csv.DictReader(data_no_header, ["name", "age", "city"])
rows_with_fieldnames = []
for row in dreader:
    rows_with_fieldnames.append(row)

results["dictreader_fieldnames_count"] = len(rows_with_fieldnames)
results["dictreader_fieldnames_row1_name"] = rows_with_fieldnames[0]["name"]
results["dictreader_fieldnames_row2_city"] = rows_with_fieldnames[1]["city"]

# =====================================
# csv.writer - basic usage
# =====================================

w = csv.writer()
w.writerow(["name", "age", "city"])
w.writerow(["Alice", 30, "NYC"])
w.writerow(["Bob", 25, "LA"])
output = w.getvalue()

results["writer_has_header"] = "name,age,city" in output
results["writer_has_alice"] = "Alice,30,NYC" in output
results["writer_has_bob"] = "Bob,25,LA" in output
results["writer_line_count"] = len(output.split("\n")) - 1  # count newlines

# =====================================
# csv.writer - writerows
# =====================================

w2 = csv.writer()
w2.writerows([["a", "b"], ["c", "d"], ["e", "f"]])
output2 = w2.getvalue()

results["writer_writerows_line1"] = "a,b" in output2
results["writer_writerows_line2"] = "c,d" in output2
results["writer_writerows_line3"] = "e,f" in output2

# =====================================
# csv.writer - custom delimiter
# =====================================

w3 = csv.writer(";")
w3.writerow(["a", "b", "c"])
output3 = w3.getvalue()

results["writer_semicolon"] = "a;b;c" in output3

# =====================================
# csv.writer - quoting
# =====================================

w4 = csv.writer(",", '"', csv.QUOTE_ALL)
w4.writerow(["a", "b"])
output4 = w4.getvalue()

results["writer_quote_all"] = '"a","b"' in output4

# =====================================
# csv.DictWriter - basic usage
# =====================================

dw = csv.DictWriter(["name", "age", "city"])
dw.writeheader()
dw.writerow({"name": "Alice", "age": "30", "city": "NYC"})
dw.writerow({"name": "Bob", "age": "25", "city": "LA"})
dw_output = dw.getvalue()

results["dictwriter_has_header"] = "name,age,city" in dw_output
results["dictwriter_has_alice"] = "Alice,30,NYC" in dw_output
results["dictwriter_has_bob"] = "Bob,25,LA" in dw_output

# =====================================
# csv.DictWriter - writerows
# =====================================

dw2 = csv.DictWriter(["x", "y"])
dw2.writeheader()
dw2.writerows([{"x": "1", "y": "2"}, {"x": "3", "y": "4"}])
dw2_output = dw2.getvalue()

results["dictwriter_writerows_header"] = "x,y" in dw2_output
results["dictwriter_writerows_row1"] = "1,2" in dw2_output
results["dictwriter_writerows_row2"] = "3,4" in dw2_output

# =====================================
# csv.DictWriter - missing values (restval)
# =====================================

dw3 = csv.DictWriter(["a", "b", "c"], ",", '"', csv.QUOTE_MINIMAL, "\n", "N/A")
dw3.writerow({"a": "1", "c": "3"})  # missing "b"
dw3_output = dw3.getvalue()

results["dictwriter_restval"] = "1,N/A,3" in dw3_output

# =====================================
# Round-trip test
# =====================================

# Write some data
w_rt = csv.writer()
w_rt.writerow(["name", "score"])
w_rt.writerow(["Alice", "100"])
w_rt.writerow(["Bob", "95"])
csv_string = w_rt.getvalue()

# Read it back
lines_rt = csv_string.strip().split("\n")
reader_rt = csv.reader(lines_rt)
roundtrip_rows = []
for row in reader_rt:
    roundtrip_rows.append(row)

results["roundtrip_header"] = roundtrip_rows[0] == ["name", "score"]
results["roundtrip_row1"] = roundtrip_rows[1] == ["Alice", "100"]
results["roundtrip_row2"] = roundtrip_rows[2] == ["Bob", "95"]

# =====================================
# Constants
# =====================================

results["const_quote_minimal"] = csv.QUOTE_MINIMAL == 0
results["const_quote_all"] = csv.QUOTE_ALL == 1
results["const_quote_nonnumeric"] = csv.QUOTE_NONNUMERIC == 2
results["const_quote_none"] = csv.QUOTE_NONE == 3

print("CSV module tests completed")
