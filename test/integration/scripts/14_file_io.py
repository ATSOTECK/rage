# Test: File I/O Operations
# Tests open(), read(), write(), and file object methods

import io

results = {}

# Get temp directory from test runner
tmp_dir = __test_tmp_dir__

# =====================================
# Basic File Writing and Reading
# =====================================

# Write to file
test_file = tmp_dir + "/test.txt"
f = open(test_file, "w")
bytes_written = f.write("Hello, World!\n")
f.write("Second line\n")
f.close()

results["write_bytes"] = bytes_written

# Read entire file
f = open(test_file, "r")
content = f.read()
f.close()
results["read_all"] = content

# =====================================
# readline and readlines
# =====================================

lines_file = tmp_dir + "/lines.txt"
f = open(lines_file, "w")
f.write("Line 1\n")
f.write("Line 2\n")
f.write("Line 3\n")
f.close()

f = open(lines_file, "r")
line1 = f.readline()
line2 = f.readline()
f.close()
results["readline1"] = line1
results["readline2"] = line2

f = open(lines_file, "r")
all_lines = f.readlines()
f.close()
results["readlines_count"] = len(all_lines)

# =====================================
# Context Manager (with statement)
# =====================================

with_file = tmp_dir + "/with_test.txt"
with open(with_file, "w") as f:
    f.write("Context manager test\n")

with open(with_file, "r") as f:
    with_content = f.read()

# File should be closed after with block
results["with_content"] = with_content
results["with_closed"] = f.closed

# =====================================
# Seek and Tell
# =====================================

seek_file = tmp_dir + "/seek.txt"
f = open(seek_file, "w")
f.write("0123456789")
f.close()

f = open(seek_file, "r")
f.seek(5)
pos = f.tell()
rest = f.read()
f.close()
results["seek_pos"] = pos
results["seek_rest"] = rest

# =====================================
# Append Mode
# =====================================

append_file = tmp_dir + "/append.txt"
f = open(append_file, "w")
f.write("Original\n")
f.close()

f = open(append_file, "a")
f.write("Appended\n")
f.close()

f = open(append_file, "r")
append_content = f.read()
f.close()
results["append_content"] = append_content

# =====================================
# Partial Read
# =====================================

partial_file = tmp_dir + "/partial.txt"
f = open(partial_file, "w")
f.write("Hello, World!")
f.close()

f = open(partial_file, "r")
part1 = f.read(5)
part2 = f.read(2)
f.close()
results["partial1"] = part1
results["partial2"] = part2

# =====================================
# writelines
# =====================================

writelines_file = tmp_dir + "/writelines.txt"
lines = ["Line A\n", "Line B\n", "Line C\n"]
f = open(writelines_file, "w")
f.writelines(lines)
f.close()

f = open(writelines_file, "r")
wl_content = f.read()
f.close()
results["writelines_content"] = wl_content

# =====================================
# File Properties
# =====================================

prop_file = tmp_dir + "/props.txt"
f = open(prop_file, "w")
# Only check filename, not full path (temp dirs vary)
results["prop_name_contains"] = "props.txt" in f.name
results["prop_mode"] = f.mode
results["prop_closed_before"] = f.closed
f.close()
results["prop_closed_after"] = f.closed

# =====================================
# io module constants
# =====================================

results["seek_set"] = io.SEEK_SET
results["seek_cur"] = io.SEEK_CUR
results["seek_end"] = io.SEEK_END

# =====================================
# readable/writable/seekable
# =====================================

rws_file = tmp_dir + "/rws.txt"
f = open(rws_file, "w")
results["write_readable"] = f.readable()
results["write_writable"] = f.writable()
results["write_seekable"] = f.seekable()
f.close()

f = open(rws_file, "r")
results["read_readable"] = f.readable()
results["read_writable"] = f.writable()
results["read_seekable"] = f.seekable()
f.close()

print("File I/O tests completed")
