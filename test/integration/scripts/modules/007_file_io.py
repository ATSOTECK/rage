# Test: File I/O Operations
# Tests open(), read(), write(), and file object methods

from test_framework import test, expect

import io

# Get temp directory from test runner
tmp_dir = __test_tmp_dir__

def test_write_and_read():
    test_file = tmp_dir + "/test.txt"
    f = open(test_file, "w")
    bytes_written = f.write("Hello, World!\n")
    f.write("Second line\n")
    f.close()
    expect(bytes_written).to_be(14)

    f = open(test_file, "r")
    content = f.read()
    f.close()
    expect(content).to_be("Hello, World!\nSecond line\n")

def test_readline():
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
    expect(line1).to_be("Line 1\n")
    expect(line2).to_be("Line 2\n")

def test_readlines():
    lines_file = tmp_dir + "/lines2.txt"
    f = open(lines_file, "w")
    f.write("Line 1\n")
    f.write("Line 2\n")
    f.write("Line 3\n")
    f.close()

    f = open(lines_file, "r")
    all_lines = f.readlines()
    f.close()
    expect(len(all_lines)).to_be(3)

def test_context_manager():
    with_file = tmp_dir + "/with_test.txt"
    with open(with_file, "w") as f:
        f.write("Context manager test\n")

    with open(with_file, "r") as f:
        with_content = f.read()

    expect(with_content).to_be("Context manager test\n")
    expect(f.closed).to_be(True)

def test_seek_tell():
    seek_file = tmp_dir + "/seek.txt"
    f = open(seek_file, "w")
    f.write("0123456789")
    f.close()

    f = open(seek_file, "r")
    f.seek(5)
    pos = f.tell()
    rest = f.read()
    f.close()
    expect(pos).to_be(5)
    expect(rest).to_be("56789")

def test_append_mode():
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
    expect(append_content).to_be("Original\nAppended\n")

def test_partial_read():
    partial_file = tmp_dir + "/partial.txt"
    f = open(partial_file, "w")
    f.write("Hello, World!")
    f.close()

    f = open(partial_file, "r")
    part1 = f.read(5)
    part2 = f.read(2)
    f.close()
    expect(part1).to_be("Hello")
    expect(part2).to_be(", ")

def test_writelines():
    writelines_file = tmp_dir + "/writelines.txt"
    lines = ["Line A\n", "Line B\n", "Line C\n"]
    f = open(writelines_file, "w")
    f.writelines(lines)
    f.close()

    f = open(writelines_file, "r")
    wl_content = f.read()
    f.close()
    expect(wl_content).to_be("Line A\nLine B\nLine C\n")

def test_file_properties():
    prop_file = tmp_dir + "/props.txt"
    f = open(prop_file, "w")
    expect("props.txt" in f.name).to_be(True)
    expect(f.mode).to_be("w")
    expect(f.closed).to_be(False)
    f.close()
    expect(f.closed).to_be(True)

def test_io_constants():
    expect(io.SEEK_SET).to_be(0)
    expect(io.SEEK_CUR).to_be(1)
    expect(io.SEEK_END).to_be(2)

def test_readable_writable():
    rws_file = tmp_dir + "/rws.txt"
    f = open(rws_file, "w")
    expect(f.readable()).to_be(False)
    expect(f.writable()).to_be(True)
    expect(f.seekable()).to_be(True)
    f.close()

    f = open(rws_file, "r")
    expect(f.readable()).to_be(True)
    expect(f.writable()).to_be(False)
    expect(f.seekable()).to_be(True)
    f.close()

test("write_and_read", test_write_and_read)
test("readline", test_readline)
test("readlines", test_readlines)
test("context_manager", test_context_manager)
test("seek_tell", test_seek_tell)
test("append_mode", test_append_mode)
test("partial_read", test_partial_read)
test("writelines", test_writelines)
test("file_properties", test_file_properties)
test("io_constants", test_io_constants)
test("readable_writable", test_readable_writable)

print("File I/O tests completed")
