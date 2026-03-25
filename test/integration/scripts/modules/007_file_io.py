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

# =========================================================================
# CPython-derived tests: StringIO (from Lib/test/test_memoryio.py)
# =========================================================================

def test_cpython_stringio_read_sizes():
    """CPython MemorySeekTestMixin.testRead + MemoryTestMixin.test_read"""
    buf = "1234567890"
    sio = io.StringIO(buf)
    # Read with various sizes
    expect(sio.read(1)).to_be("1")
    expect(sio.read(4)).to_be("2345")
    expect(sio.read(900)).to_be("67890")
    expect(sio.read()).to_be("")  # EOF

def test_cpython_stringio_read_no_args():
    """CPython MemorySeekTestMixin.testReadNoArgs"""
    buf = "1234567890"
    sio = io.StringIO(buf)
    expect(sio.read()).to_be(buf)
    expect(sio.read()).to_be("")  # EOF after reading all

def test_cpython_stringio_read_zero():
    """CPython MemoryTestMixin.test_read - read(0)"""
    buf = "1234567890"
    sio = io.StringIO(buf)
    expect(sio.read(0)).to_be("")
    expect(sio.read(1)).to_be("1")

def test_cpython_stringio_read_negative():
    """CPython MemoryTestMixin.test_read - read(-1) reads all"""
    buf = "1234567890"
    sio = io.StringIO(buf)
    expect(sio.read(-1)).to_be(buf)

def test_cpython_stringio_read_past_end():
    """CPython MemoryTestMixin.test_read - seek past end then read"""
    buf = "1234567890"
    sio = io.StringIO(buf)
    sio.seek(100)
    expect(sio.read(1)).to_be("")
    sio.seek(len(buf) + 1)
    expect(sio.read()).to_be("")

def test_cpython_stringio_read_tell():
    """CPython MemoryTestMixin.test_read - tell after full read"""
    sio = io.StringIO("1234567890")
    sio.read()
    expect(sio.tell()).to_be(10)

def test_cpython_stringio_seek_whence():
    """CPython MemoryTestMixin.test_seek - whence values"""
    buf = "1234567890"
    sio = io.StringIO(buf)
    sio.read(5)
    expect(sio.seek(0)).to_be(0)
    expect(sio.seek(0, 0)).to_be(0)
    expect(sio.read()).to_be(buf)
    expect(sio.seek(3)).to_be(3)
    expect(sio.read()).to_be(buf[3:])
    expect(sio.seek(len(buf))).to_be(len(buf))
    expect(sio.read()).to_be("")
    # seek to end with whence=2
    expect(sio.seek(0, 2)).to_be(len(buf))
    expect(sio.read()).to_be("")

def test_cpython_stringio_seek_negative():
    """CPython MemoryTestMixin.test_seek - negative seek raises ValueError"""
    sio = io.StringIO("1234567890")
    sio.read(5)
    raised = False
    try:
        sio.seek(-1)
    except ValueError:
        raised = True
    expect(raised).to_be(True)

def test_cpython_stringio_overseek():
    """CPython MemoryTestMixin.test_overseek"""
    buf = "1234567890"
    sio = io.StringIO(buf)
    expect(sio.seek(11)).to_be(11)
    expect(sio.read()).to_be("")
    expect(sio.tell()).to_be(11)
    expect(sio.getvalue()).to_be(buf)

def test_cpython_stringio_tell_seek():
    """CPython MemoryTestMixin.test_tell"""
    sio = io.StringIO("1234567890")
    expect(sio.tell()).to_be(0)
    sio.seek(5)
    expect(sio.tell()).to_be(5)
    sio.seek(10000)
    expect(sio.tell()).to_be(10000)

def test_cpython_stringio_write_ops():
    """CPython MemoryTestMixin.write_ops - interleaved write/seek/truncate"""
    sio = io.StringIO()
    expect(sio.write("blah.")).to_be(5)
    expect(sio.seek(0)).to_be(0)
    expect(sio.write("Hello.")).to_be(6)
    expect(sio.tell()).to_be(6)
    expect(sio.seek(5)).to_be(5)
    expect(sio.tell()).to_be(5)
    expect(sio.write(" world\n\n\n")).to_be(9)
    expect(sio.seek(0)).to_be(0)
    expect(sio.write("h")).to_be(1)
    expect(sio.truncate(12)).to_be(12)
    expect(sio.tell()).to_be(1)

def test_cpython_stringio_write_getvalue():
    """CPython MemoryTestMixin.test_write - write then verify getvalue"""
    buf = "hello world\n"
    sio = io.StringIO(buf)
    # Replay write_ops
    sio.write("blah.")
    sio.seek(0)
    sio.write("Hello.")
    sio.seek(5)
    sio.write(" world\n\n\n")
    sio.seek(0)
    sio.write("h")
    sio.truncate(12)
    expect(sio.getvalue()).to_be(buf)

def test_cpython_stringio_writelines_repeat():
    """CPython MemoryTestMixin.test_writelines"""
    buf = "1234567890"
    sio = io.StringIO()
    sio.writelines([buf, buf, buf])
    expect(sio.getvalue()).to_be(buf * 3)
    sio.writelines([])
    expect(sio.getvalue()).to_be(buf * 3)

def test_cpython_stringio_truncate_detailed():
    """CPython MemoryTestMixin.test_truncate"""
    buf = "1234567890"
    sio = io.StringIO(buf)
    sio.seek(6)
    expect(sio.truncate(8)).to_be(8)
    expect(sio.getvalue()).to_be(buf[:8])
    expect(sio.truncate()).to_be(6)   # truncate at pos
    expect(sio.getvalue()).to_be(buf[:6])
    expect(sio.truncate(4)).to_be(4)
    expect(sio.getvalue()).to_be(buf[:4])
    expect(sio.tell()).to_be(6)       # tell unchanged after truncate
    sio.seek(0, 2)                    # seek to end
    sio.write(buf)
    expect(sio.getvalue()).to_be(buf[:4] + buf)

def test_cpython_stringio_readline_detailed():
    """CPython MemoryTestMixin.test_readline"""
    buf = "1234567890\n"
    sio = io.StringIO(buf * 2)
    expect(sio.readline()).to_be(buf)
    expect(sio.readline()).to_be(buf)
    expect(sio.readline()).to_be("")
    # readline with size limit
    sio.seek(0)
    expect(sio.readline(5)).to_be(buf[:5])
    expect(sio.readline(5)).to_be(buf[5:10])

def test_cpython_stringio_readline_no_trailing_newline():
    """CPython test_readline - buffer without trailing newline"""
    buf = "1234567890\n"
    sio = io.StringIO((buf * 3)[:-1])
    expect(sio.readline()).to_be(buf)
    expect(sio.readline()).to_be(buf)
    expect(sio.readline()).to_be(buf[:-1])
    expect(sio.readline()).to_be("")

def test_cpython_stringio_readlines_detailed():
    """CPython MemoryTestMixin.test_readlines"""
    buf = "1234567890\n"
    sio = io.StringIO(buf * 10)
    lines = sio.readlines()
    expect(len(lines)).to_be(10)
    expect(lines[0]).to_be(buf)
    expect(lines[9]).to_be(buf)
    # readlines from middle
    sio.seek(5)
    lines = sio.readlines()
    expect(lines[0]).to_be(buf[5:])
    expect(len(lines)).to_be(10)

def test_cpython_stringio_readlines_hint():
    """CPython MemoryTestMixin.test_readlines with hint"""
    buf = "1234567890\n"
    sio = io.StringIO(buf * 10)
    lines = sio.readlines(15)
    expect(len(lines)).to_be(2)

def test_cpython_stringio_iterator_detailed():
    """CPython MemoryTestMixin.test_iterator"""
    buf = "1234567890\n"
    sio = io.StringIO(buf * 10)
    i = 0
    for line in sio:
        expect(line).to_be(buf)
        i = i + 1
    expect(i).to_be(10)
    # Re-seek and iterate again
    sio.seek(0)
    i = 0
    for line in sio:
        expect(line).to_be(buf)
        i = i + 1
    expect(i).to_be(10)

def test_cpython_stringio_getvalue_after_read():
    """CPython MemoryTestMixin.test_getvalue"""
    buf = "1234567890"
    sio = io.StringIO(buf)
    expect(sio.getvalue()).to_be(buf)
    sio.read()
    expect(sio.getvalue()).to_be(buf)  # getvalue unaffected by read pos

def test_cpython_stringio_getvalue_large():
    """CPython MemoryTestMixin.test_getvalue - large buffer"""
    buf = "1234567890"
    sio = io.StringIO(buf * 1000)
    val = sio.getvalue()
    expect(val[-3:]).to_be("890")

def test_cpython_stringio_flush():
    """CPython MemoryTestMixin.test_flush"""
    sio = io.StringIO("1234567890")
    result = sio.flush()
    expect(result).to_be(None)

def test_cpython_stringio_flags():
    """CPython MemoryTestMixin.test_flags"""
    sio = io.StringIO()
    expect(sio.writable()).to_be(True)
    expect(sio.readable()).to_be(True)
    expect(sio.seekable()).to_be(True)
    expect(sio.isatty()).to_be(False)
    expect(sio.closed).to_be(False)
    sio.close()
    expect(sio.closed).to_be(True)

def test_cpython_stringio_closed_write():
    """CPython MemoryTestMixin.test_write - write to closed raises ValueError"""
    sio = io.StringIO()
    sio.close()
    raised = False
    try:
        sio.write("")
    except ValueError:
        raised = True
    expect(raised).to_be(True)

def test_cpython_stringio_closed_read():
    """CPython MemoryTestMixin.test_read - read from closed raises ValueError"""
    sio = io.StringIO("test")
    sio.close()
    raised = False
    try:
        sio.read()
    except ValueError:
        raised = True
    expect(raised).to_be(True)

def test_cpython_stringio_closed_getvalue():
    """CPython MemoryTestMixin.test_getvalue - getvalue on closed raises ValueError"""
    sio = io.StringIO("test")
    sio.close()
    raised = False
    try:
        sio.getvalue()
    except ValueError:
        raised = True
    expect(raised).to_be(True)

def test_cpython_stringio_closed_seek():
    """CPython MemoryTestMixin.test_seek - seek on closed raises ValueError"""
    sio = io.StringIO("test")
    sio.close()
    raised = False
    try:
        sio.seek(0)
    except ValueError:
        raised = True
    expect(raised).to_be(True)

def test_cpython_stringio_closed_tell():
    """CPython MemoryTestMixin.test_tell - tell on closed raises ValueError"""
    sio = io.StringIO("test")
    sio.close()
    raised = False
    try:
        sio.tell()
    except ValueError:
        raised = True
    expect(raised).to_be(True)

def test_cpython_stringio_closed_truncate():
    """CPython MemoryTestMixin.test_truncate - truncate on closed"""
    sio = io.StringIO("test")
    sio.close()
    raised = False
    try:
        sio.truncate(0)
    except ValueError:
        raised = True
    expect(raised).to_be(True)

def test_cpython_stringio_closed_readline():
    """CPython test_readline - readline on closed raises ValueError"""
    sio = io.StringIO("test\n")
    sio.close()
    raised = False
    try:
        sio.readline()
    except ValueError:
        raised = True
    expect(raised).to_be(True)

def test_cpython_stringio_closed_readlines():
    """CPython test_readlines - readlines on closed raises ValueError"""
    sio = io.StringIO("test\n")
    sio.close()
    raised = False
    try:
        sio.readlines()
    except ValueError:
        raised = True
    expect(raised).to_be(True)

def test_cpython_stringio_closed_writelines():
    """CPython test_writelines - writelines on closed raises ValueError"""
    sio = io.StringIO()
    sio.close()
    raised = False
    try:
        sio.writelines([])
    except ValueError:
        raised = True
    expect(raised).to_be(True)

def test_cpython_stringio_closed_next():
    """CPython test_iterator - __next__ on closed raises ValueError"""
    sio = io.StringIO("a\nb\n")
    sio.close()
    raised = False
    try:
        sio.__next__()
    except ValueError:
        raised = True
    expect(raised).to_be(True)

def test_cpython_stringio_init_none():
    """CPython MemoryTestMixin.test_init - StringIO(None) gives empty"""
    sio = io.StringIO()
    expect(sio.getvalue()).to_be("")

# =========================================================================
# CPython-derived tests: BytesIO (from Lib/test/test_memoryio.py)
# =========================================================================

def test_cpython_bytesio_read_sizes():
    """CPython MemorySeekTestMixin.testRead for BytesIO"""
    buf = b"1234567890"
    bio = io.BytesIO(buf)
    expect(bio.read(1)).to_be(b"1")
    expect(bio.read(4)).to_be(b"2345")
    expect(bio.read(900)).to_be(b"67890")
    expect(bio.read()).to_be(b"")

def test_cpython_bytesio_read_no_args():
    """CPython MemorySeekTestMixin.testReadNoArgs for BytesIO"""
    buf = b"1234567890"
    bio = io.BytesIO(buf)
    expect(bio.read()).to_be(buf)
    expect(bio.read()).to_be(b"")

def test_cpython_bytesio_read_zero():
    """CPython test_read - read(0) returns empty bytes"""
    bio = io.BytesIO(b"1234567890")
    expect(bio.read(0)).to_be(b"")

def test_cpython_bytesio_read_negative():
    """CPython test_read - read(-1) reads all"""
    buf = b"1234567890"
    bio = io.BytesIO(buf)
    expect(bio.read(-1)).to_be(buf)

def test_cpython_bytesio_seek_whence():
    """CPython MemorySeekTestMixin.testSeek for BytesIO"""
    buf = b"1234567890"
    bio = io.BytesIO(buf)
    bio.read(5)
    bio.seek(0)
    expect(bio.read()).to_be(buf)
    bio.seek(3)
    expect(bio.read()).to_be(buf[3:])

def test_cpython_bytesio_tell_detailed():
    """CPython MemorySeekTestMixin.testTell for BytesIO"""
    bio = io.BytesIO(b"1234567890")
    expect(bio.tell()).to_be(0)
    bio.seek(5)
    expect(bio.tell()).to_be(5)
    bio.seek(10000)
    expect(bio.tell()).to_be(10000)

def test_cpython_bytesio_relative_seek():
    """CPython PyBytesIOTest.test_relative_seek"""
    buf = b"1234567890"
    bio = io.BytesIO(buf)
    expect(bio.seek(-1, 1)).to_be(0)   # SEEK_CUR, clamped
    expect(bio.seek(3, 1)).to_be(3)
    expect(bio.seek(-4, 1)).to_be(0)   # clamped to 0
    expect(bio.seek(-1, 2)).to_be(9)   # SEEK_END - 1
    expect(bio.seek(1, 1)).to_be(10)
    expect(bio.seek(1, 2)).to_be(11)
    bio.seek(-3, 2)
    expect(bio.read()).to_be(buf[-3:])
    bio.seek(0)
    bio.seek(1, 1)
    expect(bio.read()).to_be(buf[1:])

def test_cpython_bytesio_write_ops():
    """CPython MemoryTestMixin.write_ops for BytesIO"""
    bio = io.BytesIO()
    expect(bio.write(b"blah.")).to_be(5)
    expect(bio.seek(0)).to_be(0)
    expect(bio.write(b"Hello.")).to_be(6)
    expect(bio.tell()).to_be(6)
    expect(bio.seek(5)).to_be(5)
    expect(bio.tell()).to_be(5)
    expect(bio.write(b" world\n\n\n")).to_be(9)
    expect(bio.seek(0)).to_be(0)
    expect(bio.write(b"h")).to_be(1)
    expect(bio.truncate(12)).to_be(12)
    expect(bio.tell()).to_be(1)

def test_cpython_bytesio_writelines_repeat():
    """CPython MemoryTestMixin.test_writelines for BytesIO"""
    buf = b"1234567890"
    bio = io.BytesIO()
    bio.writelines([buf, buf, buf])
    expect(bio.getvalue()).to_be(buf + buf + buf)

def test_cpython_bytesio_truncate_detailed():
    """CPython MemoryTestMixin.test_truncate for BytesIO"""
    buf = b"1234567890"
    bio = io.BytesIO(buf)
    bio.seek(6)
    expect(bio.truncate(8)).to_be(8)
    expect(bio.getvalue()).to_be(buf[:8])
    expect(bio.truncate()).to_be(6)
    expect(bio.getvalue()).to_be(buf[:6])
    expect(bio.truncate(4)).to_be(4)
    expect(bio.getvalue()).to_be(buf[:4])
    expect(bio.tell()).to_be(6)
    bio.seek(0, 2)
    bio.write(buf)
    expect(bio.getvalue()).to_be(buf[:4] + buf)

def test_cpython_bytesio_readline_detailed():
    """CPython MemoryTestMixin.test_readline for BytesIO"""
    buf = b"1234567890\n"
    bio = io.BytesIO(buf * 2)
    expect(bio.readline()).to_be(buf)
    expect(bio.readline()).to_be(buf)
    expect(bio.readline()).to_be(b"")
    bio.seek(0)
    expect(bio.readline(5)).to_be(buf[:5])
    expect(bio.readline(5)).to_be(buf[5:10])

def test_cpython_bytesio_readlines_detailed():
    """CPython MemoryTestMixin.test_readlines for BytesIO"""
    buf = b"1234567890\n"
    bio = io.BytesIO(buf * 10)
    lines = bio.readlines()
    expect(len(lines)).to_be(10)
    bio.seek(5)
    lines = bio.readlines()
    expect(lines[0]).to_be(buf[5:])
    expect(len(lines)).to_be(10)

def test_cpython_bytesio_iterator():
    """CPython MemoryTestMixin.test_iterator for BytesIO"""
    buf = b"1234567890\n"
    bio = io.BytesIO(buf * 10)
    i = 0
    for line in bio:
        expect(line).to_be(buf)
        i = i + 1
    expect(i).to_be(10)
    # Re-seek and iterate again
    bio.seek(0)
    i = 0
    for line in bio:
        i = i + 1
    expect(i).to_be(10)

def test_cpython_bytesio_getvalue_after_read():
    """CPython MemoryTestMixin.test_getvalue for BytesIO"""
    buf = b"1234567890"
    bio = io.BytesIO(buf)
    expect(bio.getvalue()).to_be(buf)
    bio.read()
    expect(bio.getvalue()).to_be(buf)

def test_cpython_bytesio_overseek():
    """CPython MemoryTestMixin.test_overseek for BytesIO"""
    buf = b"1234567890"
    bio = io.BytesIO(buf)
    expect(bio.seek(11)).to_be(11)
    expect(bio.read()).to_be(b"")
    expect(bio.tell()).to_be(11)
    expect(bio.getvalue()).to_be(buf)

def test_cpython_bytesio_closed_ops():
    """CPython - operations on closed BytesIO raise ValueError"""
    bio = io.BytesIO(b"test")
    bio.close()
    ops_raised = []
    try:
        bio.read()
    except ValueError:
        ops_raised.append("read")
    try:
        bio.write(b"x")
    except ValueError:
        ops_raised.append("write")
    try:
        bio.getvalue()
    except ValueError:
        ops_raised.append("getvalue")
    try:
        bio.seek(0)
    except ValueError:
        ops_raised.append("seek")
    try:
        bio.tell()
    except ValueError:
        ops_raised.append("tell")
    expect(len(ops_raised)).to_be(5)

def test_cpython_bytesio_flush():
    """CPython MemoryTestMixin.test_flush for BytesIO"""
    bio = io.BytesIO(b"1234567890")
    result = bio.flush()
    expect(result).to_be(None)

# =========================================================================
# CPython-derived tests: File objects (from Lib/test/test_file.py)
# =========================================================================

def test_cpython_file_attributes():
    """CPython AutoFileTests.testAttributes"""
    f = open(tmp_dir + "/cpython_attrs.txt", "w")
    name = f.name  # shouldn't blow up
    mode = f.mode
    closed = f.closed
    expect(closed).to_be(False)
    f.close()

def test_cpython_file_errors():
    """CPython AutoFileTests.testErrors"""
    f = open(tmp_dir + "/cpython_errors.txt", "w")
    expect(f.isatty()).to_be(False)
    expect(f.closed).to_be(False)
    f.close()
    expect(f.closed).to_be(True)

def test_cpython_file_exit_closes():
    """CPython AutoFileTests.testMethods - __exit__ closes file"""
    f = open(tmp_dir + "/cpython_exit.txt", "w")
    f.__exit__(None, None, None)
    expect(f.closed).to_be(True)

def test_cpython_file_closed_ops():
    """CPython AutoFileTests.testMethods - methods raise on closed file"""
    f = open(tmp_dir + "/cpython_closed.txt", "w")
    f.close()
    ops_raised = []
    try:
        f.read()
    except (ValueError, Exception):
        ops_raised.append("read")
    try:
        f.write("")
    except (ValueError, Exception):
        ops_raised.append("write")
    try:
        f.readline()
    except (ValueError, Exception):
        ops_raised.append("readline")
    try:
        f.readlines()
    except (ValueError, Exception):
        ops_raised.append("readlines")
    try:
        f.seek(0)
    except (ValueError, Exception):
        ops_raised.append("seek")
    try:
        f.tell()
    except (ValueError, Exception):
        ops_raised.append("tell")
    expect(len(ops_raised)).to_be(6)

def test_cpython_file_read_when_writing():
    """CPython AutoFileTests.testReadWhenWriting"""
    f = open(tmp_dir + "/cpython_rw.txt", "w")
    raised = False
    try:
        f.read()
    except Exception:
        raised = True
    f.close()
    expect(raised).to_be(True)

def test_cpython_file_bad_mode():
    """CPython OtherFileTests.testBadModeArgument"""
    raised = False
    try:
        f = open(tmp_dir + "/cpython_badmode.txt", "qwerty")
    except ValueError:
        raised = True
    expect(raised).to_be(True)

def test_cpython_file_truncate_after_read():
    """CPython OtherFileTests.testTruncateOnWindows"""
    path = tmp_dir + "/cpython_trunc.txt"
    f = open(path, "w")
    f.write("12345678901")
    f.close()

    f = open(path, "r+")
    data = f.read(5)
    expect(data).to_be("12345")
    expect(f.tell()).to_be(5)
    f.truncate()
    expect(f.tell()).to_be(5)
    f.close()

    # Verify file size by reading
    f = open(path, "r")
    content = f.read()
    f.close()
    expect(len(content)).to_be(5)

def test_cpython_file_exclusive_mode():
    """Test 'x' mode - exclusive creation"""
    path = tmp_dir + "/cpython_exclusive.txt"
    f = open(path, "x")
    f.write("exclusive")
    f.close()

    # Opening again with 'x' should fail
    raised = False
    try:
        f = open(path, "x")
    except FileExistsError:
        raised = True
    expect(raised).to_be(True)

def test_cpython_file_nonexistent():
    """CPython - FileNotFoundError for missing files"""
    raised = False
    try:
        f = open(tmp_dir + "/nonexistent_file_12345.txt", "r")
    except FileNotFoundError:
        raised = True
    expect(raised).to_be(True)

def test_cpython_file_double_close():
    """Closing already-closed file is a no-op"""
    f = open(tmp_dir + "/cpython_double_close.txt", "w")
    f.close()
    f.close()  # Should not raise
    expect(f.closed).to_be(True)

def test_cpython_file_write_read_roundtrip():
    """Write and read back with r+ mode"""
    path = tmp_dir + "/cpython_roundtrip.txt"
    f = open(path, "w")
    f.write("hello world")
    f.close()

    f = open(path, "r+")
    content = f.read()
    expect(content).to_be("hello world")
    f.seek(0)
    f.write("HELLO")
    f.seek(0)
    content = f.read()
    expect(content).to_be("HELLO world")
    f.close()

def test_cpython_file_binary_encoding():
    """Binary mode should have encoding=None"""
    path = tmp_dir + "/cpython_bin_enc.bin"
    f = open(path, "wb")
    expect(f.encoding).to_be(None)
    f.close()

def test_cpython_file_context_exception():
    """Context manager doesn't suppress exceptions"""
    path = tmp_dir + "/cpython_ctx_exc.txt"
    raised = False
    try:
        with open(path, "w") as f:
            f.write("test")
            x = 1 / 0
    except ZeroDivisionError:
        raised = True
    expect(raised).to_be(True)
    expect(f.closed).to_be(True)

def test_cpython_file_readline_after_iteration():
    """CPython testIteration - readline after iteration at EOF"""
    path = tmp_dir + "/cpython_iter_rl.txt"
    f = open(path, "w")
    f.write("line1\nline2\nline3\n")
    f.close()

    f = open(path, "r")
    for line in f:
        pass  # exhaust iterator
    # Reading after iteration hit EOF shouldn't hurt
    rest = f.readline()
    expect(rest).to_be("")
    rest2 = f.read()
    expect(rest2).to_be("")
    f.close()

# Run all tests

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

# CPython-derived StringIO tests
test("CPython: StringIO read sizes", test_cpython_stringio_read_sizes)
test("CPython: StringIO read no args", test_cpython_stringio_read_no_args)
test("CPython: StringIO read(0)", test_cpython_stringio_read_zero)
test("CPython: StringIO read(-1)", test_cpython_stringio_read_negative)
test("CPython: StringIO read past end", test_cpython_stringio_read_past_end)
test("CPython: StringIO read tell", test_cpython_stringio_read_tell)
test("CPython: StringIO seek whence", test_cpython_stringio_seek_whence)
test("CPython: StringIO seek negative", test_cpython_stringio_seek_negative)
test("CPython: StringIO overseek", test_cpython_stringio_overseek)
test("CPython: StringIO tell+seek", test_cpython_stringio_tell_seek)
test("CPython: StringIO write ops", test_cpython_stringio_write_ops)
test("CPython: StringIO write getvalue", test_cpython_stringio_write_getvalue)
test("CPython: StringIO writelines repeat", test_cpython_stringio_writelines_repeat)
test("CPython: StringIO truncate detailed", test_cpython_stringio_truncate_detailed)
test("CPython: StringIO readline detailed", test_cpython_stringio_readline_detailed)
test("CPython: StringIO readline no trailing newline", test_cpython_stringio_readline_no_trailing_newline)
test("CPython: StringIO readlines detailed", test_cpython_stringio_readlines_detailed)
test("CPython: StringIO readlines hint", test_cpython_stringio_readlines_hint)
test("CPython: StringIO iterator detailed", test_cpython_stringio_iterator_detailed)
test("CPython: StringIO getvalue after read", test_cpython_stringio_getvalue_after_read)
test("CPython: StringIO getvalue large", test_cpython_stringio_getvalue_large)
test("CPython: StringIO flush", test_cpython_stringio_flush)
test("CPython: StringIO flags", test_cpython_stringio_flags)
test("CPython: StringIO closed write", test_cpython_stringio_closed_write)
test("CPython: StringIO closed read", test_cpython_stringio_closed_read)
test("CPython: StringIO closed getvalue", test_cpython_stringio_closed_getvalue)
test("CPython: StringIO closed seek", test_cpython_stringio_closed_seek)
test("CPython: StringIO closed tell", test_cpython_stringio_closed_tell)
test("CPython: StringIO closed truncate", test_cpython_stringio_closed_truncate)
test("CPython: StringIO closed readline", test_cpython_stringio_closed_readline)
test("CPython: StringIO closed readlines", test_cpython_stringio_closed_readlines)
test("CPython: StringIO closed writelines", test_cpython_stringio_closed_writelines)
test("CPython: StringIO closed next", test_cpython_stringio_closed_next)
test("CPython: StringIO init none", test_cpython_stringio_init_none)

# CPython-derived BytesIO tests
test("CPython: BytesIO read sizes", test_cpython_bytesio_read_sizes)
test("CPython: BytesIO read no args", test_cpython_bytesio_read_no_args)
test("CPython: BytesIO read(0)", test_cpython_bytesio_read_zero)
test("CPython: BytesIO read(-1)", test_cpython_bytesio_read_negative)
test("CPython: BytesIO seek whence", test_cpython_bytesio_seek_whence)
test("CPython: BytesIO tell detailed", test_cpython_bytesio_tell_detailed)
test("CPython: BytesIO relative seek", test_cpython_bytesio_relative_seek)
test("CPython: BytesIO write ops", test_cpython_bytesio_write_ops)
test("CPython: BytesIO writelines repeat", test_cpython_bytesio_writelines_repeat)
test("CPython: BytesIO truncate detailed", test_cpython_bytesio_truncate_detailed)
test("CPython: BytesIO readline detailed", test_cpython_bytesio_readline_detailed)
test("CPython: BytesIO readlines detailed", test_cpython_bytesio_readlines_detailed)
test("CPython: BytesIO iterator", test_cpython_bytesio_iterator)
test("CPython: BytesIO getvalue after read", test_cpython_bytesio_getvalue_after_read)
test("CPython: BytesIO overseek", test_cpython_bytesio_overseek)
test("CPython: BytesIO closed ops", test_cpython_bytesio_closed_ops)
test("CPython: BytesIO flush", test_cpython_bytesio_flush)

# CPython-derived file tests
test("CPython: file attributes", test_cpython_file_attributes)
test("CPython: file errors", test_cpython_file_errors)
test("CPython: file __exit__ closes", test_cpython_file_exit_closes)
test("CPython: file closed ops", test_cpython_file_closed_ops)
test("CPython: file read when writing", test_cpython_file_read_when_writing)
test("CPython: file bad mode", test_cpython_file_bad_mode)
test("CPython: file truncate after read", test_cpython_file_truncate_after_read)
test("CPython: file exclusive mode", test_cpython_file_exclusive_mode)
test("CPython: file nonexistent", test_cpython_file_nonexistent)
test("CPython: file double close", test_cpython_file_double_close)
test("CPython: file write read roundtrip", test_cpython_file_write_read_roundtrip)
test("CPython: file binary encoding", test_cpython_file_binary_encoding)
test("CPython: file context exception", test_cpython_file_context_exception)
test("CPython: file readline after iteration", test_cpython_file_readline_after_iteration)

print("File I/O tests completed")
