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

def test_truncate():
    trunc_file = tmp_dir + "/trunc.txt"
    f = open(trunc_file, "w+")
    f.write("Hello, World!")
    f.flush()
    new_size = f.truncate(5)
    f.close()
    expect(new_size).to_be(5)

    f = open(trunc_file, "r")
    content = f.read()
    f.close()
    expect(content).to_be("Hello")

def test_truncate_at_position():
    trunc_file = tmp_dir + "/trunc2.txt"
    f = open(trunc_file, "w+")
    f.write("Hello, World!")
    f.seek(5)
    f.truncate()
    f.close()

    f = open(trunc_file, "r")
    content = f.read()
    f.close()
    expect(content).to_be("Hello")

def test_isatty():
    isatty_file = tmp_dir + "/isatty.txt"
    f = open(isatty_file, "w")
    expect(f.isatty()).to_be(False)
    f.close()

def test_print_to_file():
    pf = tmp_dir + "/print_out.txt"
    f = open(pf, "w")
    print("hello", "world", file=f)
    print("line2", file=f, end="!\n")
    f.close()

    f = open(pf, "r")
    content = f.read()
    f.close()
    expect(content).to_be("hello world\nline2!\n")

def test_print_to_file_sep():
    pf = tmp_dir + "/print_sep.txt"
    f = open(pf, "w")
    print(1, 2, 3, sep=", ", file=f)
    f.close()

    f = open(pf, "r")
    content = f.read()
    f.close()
    expect(content).to_be("1, 2, 3\n")

def test_open_kwargs():
    kw_file = tmp_dir + "/kwargs.txt"
    f = open(kw_file, mode="w")
    f.write("kwargs work")
    f.close()

    f = open(kw_file, mode="r")
    content = f.read()
    f.close()
    expect(content).to_be("kwargs work")

def test_open_encoding_kwarg():
    enc_file = tmp_dir + "/enc.txt"
    f = open(enc_file, "w", encoding="utf-8")
    f.write("encoded")
    f.close()
    expect(f.encoding).to_be("utf-8")

def test_binary_read_write():
    bin_file = tmp_dir + "/binary.bin"
    f = open(bin_file, "wb")
    f.write(b"\x00\x01\x02\x03")
    f.close()

    f = open(bin_file, "rb")
    data = f.read()
    f.close()
    expect(len(data)).to_be(4)
    expect(data[0]).to_be(0)
    expect(data[3]).to_be(3)

def test_file_iteration():
    iter_file = tmp_dir + "/iter.txt"
    f = open(iter_file, "w")
    f.write("line1\nline2\nline3\n")
    f.close()

    lines = []
    for line in open(iter_file, "r"):
        lines.append(line)
    expect(len(lines)).to_be(3)
    expect(lines[0]).to_be("line1\n")
    expect(lines[2]).to_be("line3\n")

# StringIO tests

def test_stringio_basic():
    sio = io.StringIO()
    sio.write("hello ")
    sio.write("world")
    expect(sio.getvalue()).to_be("hello world")
    expect(sio.tell()).to_be(11)

def test_stringio_read():
    sio = io.StringIO("hello world")
    expect(sio.read(5)).to_be("hello")
    expect(sio.read()).to_be(" world")

def test_stringio_readline():
    sio = io.StringIO("line1\nline2\nline3")
    expect(sio.readline()).to_be("line1\n")
    expect(sio.readline()).to_be("line2\n")
    expect(sio.readline()).to_be("line3")
    expect(sio.readline()).to_be("")

def test_stringio_readlines():
    sio = io.StringIO("line1\nline2\nline3\n")
    lines = sio.readlines()
    expect(len(lines)).to_be(3)
    expect(lines[0]).to_be("line1\n")

def test_stringio_seek_tell():
    sio = io.StringIO("hello")
    sio.read(3)
    expect(sio.tell()).to_be(3)
    sio.seek(0)
    expect(sio.tell()).to_be(0)
    expect(sio.read()).to_be("hello")

def test_stringio_truncate():
    sio = io.StringIO("hello world")
    sio.truncate(5)
    expect(sio.getvalue()).to_be("hello")

def test_stringio_context_manager():
    with io.StringIO("test") as sio:
        content = sio.read()
    expect(content).to_be("test")
    expect(sio.closed).to_be(True)

def test_stringio_writelines():
    sio = io.StringIO()
    sio.writelines(["hello\n", "world\n"])
    expect(sio.getvalue()).to_be("hello\nworld\n")

def test_stringio_iteration():
    sio = io.StringIO("a\nb\nc\n")
    lines = []
    for line in sio:
        lines.append(line)
    expect(len(lines)).to_be(3)
    expect(lines[0]).to_be("a\n")

def test_stringio_capabilities():
    sio = io.StringIO()
    expect(sio.readable()).to_be(True)
    expect(sio.writable()).to_be(True)
    expect(sio.seekable()).to_be(True)
    expect(sio.isatty()).to_be(False)

# BytesIO tests

def test_bytesio_basic():
    bio = io.BytesIO()
    bio.write(b"hello ")
    bio.write(b"world")
    expect(bio.getvalue()).to_be(b"hello world")
    expect(bio.tell()).to_be(11)

def test_bytesio_read():
    bio = io.BytesIO(b"hello world")
    expect(bio.read(5)).to_be(b"hello")
    expect(bio.read()).to_be(b" world")

def test_bytesio_readline():
    bio = io.BytesIO(b"line1\nline2\nline3")
    expect(bio.readline()).to_be(b"line1\n")
    expect(bio.readline()).to_be(b"line2\n")
    expect(bio.readline()).to_be(b"line3")

def test_bytesio_seek_tell():
    bio = io.BytesIO(b"hello")
    bio.read(3)
    expect(bio.tell()).to_be(3)
    bio.seek(0)
    expect(bio.read()).to_be(b"hello")

def test_bytesio_truncate():
    bio = io.BytesIO(b"hello world")
    bio.truncate(5)
    expect(bio.getvalue()).to_be(b"hello")

def test_bytesio_context_manager():
    with io.BytesIO(b"test") as bio:
        content = bio.read()
    expect(content).to_be(b"test")
    expect(bio.closed).to_be(True)

def test_bytesio_capabilities():
    bio = io.BytesIO()
    expect(bio.readable()).to_be(True)
    expect(bio.writable()).to_be(True)
    expect(bio.seekable()).to_be(True)
    expect(bio.isatty()).to_be(False)

# Print to StringIO

def test_print_to_stringio():
    sio = io.StringIO()
    print("hello", "world", file=sio)
    expect(sio.getvalue()).to_be("hello world\n")

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
test("truncate", test_truncate)
test("truncate_at_position", test_truncate_at_position)
test("isatty", test_isatty)
test("print_to_file", test_print_to_file)
test("print_to_file_sep", test_print_to_file_sep)
test("open_kwargs", test_open_kwargs)
test("open_encoding_kwarg", test_open_encoding_kwarg)
test("binary_read_write", test_binary_read_write)
test("file_iteration", test_file_iteration)
test("stringio_basic", test_stringio_basic)
test("stringio_read", test_stringio_read)
test("stringio_readline", test_stringio_readline)
test("stringio_readlines", test_stringio_readlines)
test("stringio_seek_tell", test_stringio_seek_tell)
test("stringio_truncate", test_stringio_truncate)
test("stringio_context_manager", test_stringio_context_manager)
test("stringio_writelines", test_stringio_writelines)
test("stringio_iteration", test_stringio_iteration)
test("stringio_capabilities", test_stringio_capabilities)
test("bytesio_basic", test_bytesio_basic)
test("bytesio_read", test_bytesio_read)
test("bytesio_readline", test_bytesio_readline)
test("bytesio_seek_tell", test_bytesio_seek_tell)
test("bytesio_truncate", test_bytesio_truncate)
test("bytesio_context_manager", test_bytesio_context_manager)
test("bytesio_capabilities", test_bytesio_capabilities)
test("print_to_stringio", test_print_to_stringio)

print("File I/O tests completed")
