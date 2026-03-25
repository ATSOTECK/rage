package stdlib

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ATSOTECK/rage/internal/runtime"
)

// getTypeName returns the Python type name for a value
func getTypeName(v runtime.Value) string {
	switch v.(type) {
	case *runtime.PyNone:
		return "NoneType"
	case *runtime.PyBool:
		return "bool"
	case *runtime.PyInt:
		return "int"
	case *runtime.PyFloat:
		return "float"
	case *runtime.PyString:
		return "str"
	case *runtime.PyBytes:
		return "bytes"
	case *runtime.PyList:
		return "list"
	case *runtime.PyTuple:
		return "tuple"
	case *runtime.PyDict:
		return "dict"
	default:
		return "object"
	}
}

// PyFile wraps an os.File with buffered I/O and Python-compatible state
type PyFile struct {
	file     *os.File      // Underlying OS file handle
	reader   *bufio.Reader // Buffered reader (for read modes)
	writer   *bufio.Writer // Buffered writer (for write modes)
	name     string        // File path/name
	mode     string        // Python mode string ('r', 'w', 'rb', etc.)
	encoding string        // Text encoding (default "utf-8")
	closed   bool          // Track if file has been closed
	readable bool          // Can read from file
	writable bool          // Can write to file
	binary   bool          // Binary mode flag
	seekable bool          // Can seek (regular files, not pipes/sockets)
}

func (f *PyFile) Type() string { return "TextIOWrapper" }
func (f *PyFile) String() string {
	if f.closed {
		return fmt.Sprintf("<_io.TextIOWrapper name='%s' mode='%s' encoding='%s'>", f.name, f.mode, f.encoding)
	}
	return fmt.Sprintf("<_io.TextIOWrapper name='%s' mode='%s' encoding='%s'>", f.name, f.mode, f.encoding)
}

// InitIOModule registers the io module and open() builtin
func InitIOModule() {
	// Register file type metatable
	fileMT := &runtime.TypeMetatable{
		Name: "file",
		Methods: map[string]runtime.GoFunction{
			// Read operations
			"read":      fileRead,
			"readline":  fileReadline,
			"readlines": fileReadlines,

			// Write operations
			"write":      fileWrite,
			"writelines": fileWritelines,

			// Position operations
			"seek": fileSeek,
			"tell": fileTell,

			// State operations
			"close":    fileClose,
			"flush":    fileFlush,
			"fileno":   fileFileno,
			"truncate": fileTruncate,
			"isatty":   fileIsatty,
			"readable": fileReadable,
			"writable": fileWritable,
			"seekable": fileSeekable,

			// Context manager protocol
			"__enter__": fileEnter,
			"__exit__":  fileExit,

			// Iterator protocol
			"__iter__": fileIter,
			"__next__": fileNext,
		},
		// Properties (accessed as f.name, f.closed, etc. without calling)
		Properties: map[string]runtime.GoFunction{
			"name":     fileName,
			"mode":     fileMode,
			"encoding": fileEncoding,
			"closed":   fileClosed,
		},
	}
	runtime.RegisterTypeMetatable("file", fileMT)

	// Register StringIO type metatable
	stringIOMT := &runtime.TypeMetatable{
		Name: "StringIO",
		Methods: map[string]runtime.GoFunction{
			"read":      stringIORead,
			"readline":  stringIOReadline,
			"readlines": stringIOReadlines,
			"write":     stringIOWrite,
			"writelines": stringIOWritelines,
			"seek":      stringIOSeek,
			"tell":      stringIOTell,
			"getvalue":  stringIOGetvalue,
			"truncate":  stringIOTruncate,
			"close":     stringIOClose,
			"flush":     stringIOFlush,
			"isatty":    stringIOIsatty,
			"readable":  stringIOReadable,
			"writable":  stringIOWritable,
			"seekable":  stringIOSeekable,
			"__enter__": stringIOEnter,
			"__exit__":  stringIOExit,
			"__iter__":  stringIOIter,
			"__next__":  stringIONext,
		},
		Properties: map[string]runtime.GoFunction{
			"closed": stringIOClosed,
		},
	}
	runtime.RegisterTypeMetatable("StringIO", stringIOMT)

	// Register BytesIO type metatable
	bytesIOMT := &runtime.TypeMetatable{
		Name: "BytesIO",
		Methods: map[string]runtime.GoFunction{
			"read":      bytesIORead,
			"readline":  bytesIOReadline,
			"readlines": bytesIOReadlines,
			"write":     bytesIOWrite,
			"writelines": bytesIOWritelines,
			"seek":      bytesIOSeek,
			"tell":      bytesIOTell,
			"getvalue":  bytesIOGetvalue,
			"truncate":  bytesIOTruncate,
			"close":     bytesIOClose,
			"flush":     bytesIOFlush,
			"isatty":    bytesIOIsatty,
			"readable":  bytesIOReadable,
			"writable":  bytesIOWritable,
			"seekable":  bytesIOSeekable,
			"__enter__": bytesIOEnter,
			"__exit__":  bytesIOExit,
			"__iter__":  bytesIOIter,
			"__next__":  bytesIONext,
		},
		Properties: map[string]runtime.GoFunction{
			"closed": bytesIOClosed,
		},
	}
	runtime.RegisterTypeMetatable("BytesIO", bytesIOMT)

	// Note: open() is NOT registered here — neither as a pending builtin nor in
	// the io module dict. It must be explicitly enabled via rage.WithFileIO() or
	// rage.WithBuiltin(rage.BuiltinOpen) for security.

	// Register the io module with constants and StringIO/BytesIO constructors
	runtime.NewModuleBuilder("io").
		Doc("Core tools for working with streams.").
		Const("SEEK_SET", runtime.NewInt(0)).
		Const("SEEK_CUR", runtime.NewInt(1)).
		Const("SEEK_END", runtime.NewInt(2)).
		Const("StringIO", stringIOConstructor).
		Const("BytesIO", bytesIOConstructor).
		Register()
}

// parseMode parses Python file mode and returns flags
func parseMode(mode string) (readable, writable, binary, append bool, flag int, err error) {
	if mode == "" {
		mode = "r"
	}

	// Check for binary mode
	binary = strings.Contains(mode, "b")
	baseMode := strings.ReplaceAll(mode, "b", "")
	baseMode = strings.ReplaceAll(baseMode, "t", "") // text mode is default

	switch baseMode {
	case "r":
		readable = true
		flag = os.O_RDONLY
	case "w":
		writable = true
		flag = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	case "a":
		writable = true
		append = true
		flag = os.O_WRONLY | os.O_CREATE | os.O_APPEND
	case "r+":
		readable = true
		writable = true
		flag = os.O_RDWR
	case "w+":
		readable = true
		writable = true
		flag = os.O_RDWR | os.O_CREATE | os.O_TRUNC
	case "a+":
		readable = true
		writable = true
		append = true
		flag = os.O_RDWR | os.O_CREATE | os.O_APPEND
	case "x":
		writable = true
		flag = os.O_WRONLY | os.O_CREATE | os.O_EXCL
	case "x+":
		readable = true
		writable = true
		flag = os.O_RDWR | os.O_CREATE | os.O_EXCL
	default:
		err = fmt.Errorf("invalid mode: '%s'", mode)
	}
	return
}

// BuiltinOpen is the open() builtin function as a *PyBuiltinFunc so it supports kwargs.
// open(file, mode='r', encoding='utf-8') -> file object
var BuiltinOpen = &runtime.PyBuiltinFunc{
	Name: "open",
	Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("TypeError: open() missing required argument: 'file'")
		}

		fileArg, ok := args[0].(*runtime.PyString)
		if !ok {
			return nil, fmt.Errorf("TypeError: expected str for file, got %T", args[0])
		}
		filename := fileArg.Value

		mode := "r"
		if len(args) >= 2 && !runtime.IsNone(args[1]) {
			if s, ok := args[1].(*runtime.PyString); ok {
				mode = s.Value
			} else {
				return nil, fmt.Errorf("TypeError: expected str for mode, got %T", args[1])
			}
		}
		if v, ok := kwargs["mode"]; ok && !runtime.IsNone(v) {
			if s, ok := v.(*runtime.PyString); ok {
				mode = s.Value
			} else {
				return nil, fmt.Errorf("TypeError: expected str for mode, got %T", v)
			}
		}

		encoding := "utf-8"
		if len(args) >= 3 && !runtime.IsNone(args[2]) {
			if s, ok := args[2].(*runtime.PyString); ok {
				encoding = s.Value
			} else {
				return nil, fmt.Errorf("TypeError: expected str for encoding, got %T", args[2])
			}
		}
		if v, ok := kwargs["encoding"]; ok && !runtime.IsNone(v) {
			if s, ok := v.(*runtime.PyString); ok {
				encoding = s.Value
			} else {
				return nil, fmt.Errorf("TypeError: expected str for encoding, got %T", v)
			}
		}

		return openFile(filename, mode, encoding)
	},
}

// openFile is the shared implementation for open().
func openFile(filename, mode, encoding string) (runtime.Value, error) {
	// Parse mode
	readable, writable, binary, _, flag, err := parseMode(mode)
	if err != nil {
		return nil, fmt.Errorf("ValueError: %v", err)
	}

	// Open the file
	file, err := os.OpenFile(filename, flag, 0644)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("FileNotFoundError: [Errno 2] No such file or directory: '%s'", filename)
		} else if os.IsPermission(err) {
			return nil, fmt.Errorf("PermissionError: [Errno 13] Permission denied: '%s'", filename)
		} else if os.IsExist(err) {
			return nil, fmt.Errorf("FileExistsError: [Errno 17] File exists: '%s'", filename)
		}
		return nil, fmt.Errorf("OSError: %v", err)
	}

	pyFile := &PyFile{
		file:     file,
		name:     filename,
		mode:     mode,
		encoding: encoding,
		closed:   false,
		readable: readable,
		writable: writable,
		binary:   binary,
		seekable: true, // Assume regular file
	}

	if readable {
		pyFile.reader = bufio.NewReader(file)
	}
	if writable {
		pyFile.writer = bufio.NewWriter(file)
	}

	// Wrap in userdata with metatable
	ud := runtime.NewUserData(pyFile)
	ud.Metatable = runtime.NewDict()
	ud.Metatable.Items[runtime.NewString("__type__")] = runtime.NewString("file")
	return ud, nil
}

// Helper to extract PyFile from userdata
func getFile(vm *runtime.VM, argNum int) (*PyFile, bool) {
	ud := vm.ToUserData(argNum)
	if ud == nil {
		vm.RaiseError("expected file object")
		return nil, false
	}
	f, ok := ud.Value.(*PyFile)
	if !ok {
		vm.RaiseError("expected file object")
		return nil, false
	}
	return f, true
}

// file.read([size]) -> str or bytes
func fileRead(vm *runtime.VM) int {
	f, ok := getFile(vm, 1)
	if !ok {
		return 0
	}

	if f.closed {
		vm.RaiseError("ValueError: I/O operation on closed file")
		return 0
	}
	if !f.readable {
		vm.RaiseError("io.UnsupportedOperation: not readable")
		return 0
	}

	// Get size parameter (-1 means read all)
	size := int64(-1)
	if vm.GetTop() >= 2 && !runtime.IsNone(vm.Get(2)) {
		size = vm.CheckInt(2)
	}

	var data []byte
	var err error

	if size < 0 {
		// Read entire file
		data, err = io.ReadAll(f.reader)
	} else {
		data = make([]byte, size)
		n, readErr := io.ReadFull(f.reader, data)
		data = data[:n]
		if readErr != nil && readErr != io.EOF && readErr != io.ErrUnexpectedEOF {
			err = readErr
		}
	}

	if err != nil && err != io.EOF {
		vm.RaiseError("IOError: %v", err)
		return 0
	}

	if f.binary {
		vm.Push(runtime.NewBytes(data))
	} else {
		vm.Push(runtime.NewString(string(data)))
	}
	return 1
}

// file.readline([size]) -> str or bytes
func fileReadline(vm *runtime.VM) int {
	f, ok := getFile(vm, 1)
	if !ok {
		return 0
	}

	if f.closed {
		vm.RaiseError("ValueError: I/O operation on closed file")
		return 0
	}
	if !f.readable {
		vm.RaiseError("io.UnsupportedOperation: not readable")
		return 0
	}

	// Optional size limit
	limit := int64(-1)
	if vm.GetTop() >= 2 && !runtime.IsNone(vm.Get(2)) {
		limit = vm.CheckInt(2)
	}

	line, err := f.reader.ReadString('\n')
	if err != nil && err != io.EOF {
		vm.RaiseError("IOError: %v", err)
		return 0
	}

	// Apply size limit
	if limit >= 0 && int64(len(line)) > limit {
		line = line[:limit]
	}

	if f.binary {
		vm.Push(runtime.NewBytes([]byte(line)))
	} else {
		vm.Push(runtime.NewString(line))
	}
	return 1
}

// file.readlines([hint]) -> list of str or bytes
func fileReadlines(vm *runtime.VM) int {
	f, ok := getFile(vm, 1)
	if !ok {
		return 0
	}

	if f.closed {
		vm.RaiseError("ValueError: I/O operation on closed file")
		return 0
	}
	if !f.readable {
		vm.RaiseError("io.UnsupportedOperation: not readable")
		return 0
	}

	// Optional hint (approximate bytes to read)
	hint := int64(-1)
	if vm.GetTop() >= 2 && !runtime.IsNone(vm.Get(2)) {
		hint = vm.CheckInt(2)
	}

	var lines []runtime.Value
	var bytesRead int64

	for {
		line, err := f.reader.ReadString('\n')
		if len(line) > 0 {
			if f.binary {
				lines = append(lines, runtime.NewBytes([]byte(line)))
			} else {
				lines = append(lines, runtime.NewString(line))
			}
			bytesRead += int64(len(line))
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			vm.RaiseError("IOError: %v", err)
			return 0
		}
		if hint >= 0 && bytesRead >= hint {
			break
		}
	}

	vm.Push(runtime.NewList(lines))
	return 1
}

// file.write(s) -> int (bytes written)
func fileWrite(vm *runtime.VM) int {
	f, ok := getFile(vm, 1)
	if !ok {
		return 0
	}

	if f.closed {
		vm.RaiseError("ValueError: I/O operation on closed file")
		return 0
	}
	if !f.writable {
		vm.RaiseError("io.UnsupportedOperation: not writable")
		return 0
	}

	arg := vm.Get(2)
	var data []byte

	if f.binary {
		// Binary mode: expect bytes or string
		switch v := arg.(type) {
		case *runtime.PyBytes:
			data = v.Value
		case *runtime.PyString:
			data = []byte(v.Value)
		default:
			vm.RaiseError("TypeError: a bytes-like object is required, not '%s'", getTypeName(arg))
			return 0
		}
	} else {
		// Text mode: expect string
		switch v := arg.(type) {
		case *runtime.PyString:
			data = []byte(v.Value)
		default:
			vm.RaiseError("TypeError: write() argument must be str, not %s", getTypeName(arg))
			return 0
		}
	}

	n, err := f.writer.Write(data)
	if err != nil {
		vm.RaiseError("IOError: %v", err)
		return 0
	}

	vm.Push(runtime.NewInt(int64(n)))
	return 1
}

// file.writelines(lines)
func fileWritelines(vm *runtime.VM) int {
	f, ok := getFile(vm, 1)
	if !ok {
		return 0
	}

	if f.closed {
		vm.RaiseError("ValueError: I/O operation on closed file")
		return 0
	}
	if !f.writable {
		vm.RaiseError("io.UnsupportedOperation: not writable")
		return 0
	}

	arg := vm.Get(2)
	list, ok := arg.(*runtime.PyList)
	if !ok {
		vm.RaiseError("TypeError: writelines() argument must be an iterable of strings")
		return 0
	}

	for _, item := range list.Items {
		var data []byte
		switch v := item.(type) {
		case *runtime.PyString:
			data = []byte(v.Value)
		case *runtime.PyBytes:
			data = v.Value
		default:
			vm.RaiseError("TypeError: write() argument must be str, not %s", getTypeName(item))
			return 0
		}

		_, err := f.writer.Write(data)
		if err != nil {
			vm.RaiseError("IOError: %v", err)
			return 0
		}
	}

	return 0 // writelines returns None
}

// file.seek(offset[, whence]) -> int (new position)
func fileSeek(vm *runtime.VM) int {
	f, ok := getFile(vm, 1)
	if !ok {
		return 0
	}

	if f.closed {
		vm.RaiseError("ValueError: I/O operation on closed file")
		return 0
	}
	if !f.seekable {
		vm.RaiseError("io.UnsupportedOperation: seek")
		return 0
	}

	offset := vm.CheckInt(2)

	// whence: 0 = SEEK_SET, 1 = SEEK_CUR, 2 = SEEK_END
	whence := 0
	if vm.GetTop() >= 3 && !runtime.IsNone(vm.Get(3)) {
		whence = int(vm.CheckInt(3))
	}

	// Flush writer before seeking
	if f.writer != nil {
		if err := f.writer.Flush(); err != nil {
			vm.RaiseError("OSError: %v", err)
			return 0
		}
	}

	pos, err := f.file.Seek(offset, whence)
	if err != nil {
		vm.RaiseError("OSError: %v", err)
		return 0
	}

	// Reset reader buffer on seek (important!)
	if f.reader != nil {
		f.reader.Reset(f.file)
	}

	vm.Push(runtime.NewInt(pos))
	return 1
}

// file.tell() -> int (current position)
func fileTell(vm *runtime.VM) int {
	f, ok := getFile(vm, 1)
	if !ok {
		return 0
	}

	if f.closed {
		vm.RaiseError("ValueError: I/O operation on closed file")
		return 0
	}

	// Get underlying position
	pos, err := f.file.Seek(0, io.SeekCurrent)
	if err != nil {
		vm.RaiseError("OSError: %v", err)
		return 0
	}

	// Adjust for buffered but unread data
	if f.reader != nil {
		pos -= int64(f.reader.Buffered())
	}
	// Adjust for buffered but unwritten data
	if f.writer != nil {
		pos += int64(f.writer.Buffered())
	}

	vm.Push(runtime.NewInt(pos))
	return 1
}

// file.close()
func fileClose(vm *runtime.VM) int {
	f, ok := getFile(vm, 1)
	if !ok {
		return 0
	}

	if f.closed {
		return 0 // Closing already closed file is a no-op
	}

	// Flush any buffered writes
	if f.writer != nil {
		if err := f.writer.Flush(); err != nil {
			vm.RaiseError("OSError: %v", err)
			return 0
		}
	}

	err := f.file.Close()
	f.closed = true

	if err != nil {
		vm.RaiseError("OSError: %v", err)
		return 0
	}

	return 0
}

// file.flush()
func fileFlush(vm *runtime.VM) int {
	f, ok := getFile(vm, 1)
	if !ok {
		return 0
	}

	if f.closed {
		vm.RaiseError("ValueError: I/O operation on closed file")
		return 0
	}

	if f.writer != nil {
		err := f.writer.Flush()
		if err != nil {
			vm.RaiseError("IOError: %v", err)
			return 0
		}
	}

	// Also sync to disk
	_ = f.file.Sync()

	return 0
}

// file.truncate([size]) -> int (new file size)
func fileTruncate(vm *runtime.VM) int {
	f, ok := getFile(vm, 1)
	if !ok {
		return 0
	}

	if f.closed {
		vm.RaiseError("ValueError: I/O operation on closed file")
		return 0
	}
	if !f.writable {
		vm.RaiseError("io.UnsupportedOperation: File not open for writing")
		return 0
	}

	// Flush any buffered writes first
	if f.writer != nil {
		if err := f.writer.Flush(); err != nil {
			vm.RaiseError("OSError: %v", err)
			return 0
		}
	}

	// Compute the current logical position (before truncate)
	logicalPos, err := f.file.Seek(0, io.SeekCurrent)
	if err != nil {
		vm.RaiseError("OSError: %v", err)
		return 0
	}
	if f.reader != nil {
		logicalPos -= int64(f.reader.Buffered())
	}

	var size int64
	if vm.GetTop() >= 2 && !runtime.IsNone(vm.Get(2)) {
		size = vm.CheckInt(2)
	} else {
		// Default: truncate at current position
		size = logicalPos
	}

	if err := f.file.Truncate(size); err != nil {
		vm.RaiseError("OSError: %v", err)
		return 0
	}

	// Seek underlying file to logical position and reset reader
	// so tell() returns the correct position after truncate
	if f.reader != nil {
		f.file.Seek(logicalPos, io.SeekStart)
		f.reader.Reset(f.file)
	}

	vm.Push(runtime.NewInt(size))
	return 1
}

// file.isatty() -> bool
func fileIsatty(vm *runtime.VM) int {
	f, ok := getFile(vm, 1)
	if !ok {
		return 0
	}

	if f.closed {
		vm.RaiseError("ValueError: I/O operation on closed file")
		return 0
	}

	// Check if the file descriptor refers to a terminal
	stat, err := f.file.Stat()
	if err != nil {
		vm.Push(runtime.False)
		return 1
	}
	// Regular files and most non-device files are not ttys
	isTerminal := (stat.Mode() & os.ModeCharDevice) != 0
	vm.Push(runtime.NewBool(isTerminal))
	return 1
}

// file.fileno() -> int
func fileFileno(vm *runtime.VM) int {
	f, ok := getFile(vm, 1)
	if !ok {
		return 0
	}

	if f.closed {
		vm.RaiseError("ValueError: I/O operation on closed file")
		return 0
	}

	vm.Push(runtime.NewInt(int64(f.file.Fd())))
	return 1
}

// file.readable() -> bool
func fileReadable(vm *runtime.VM) int {
	f, ok := getFile(vm, 1)
	if !ok {
		return 0
	}

	vm.Push(runtime.NewBool(f.readable && !f.closed))
	return 1
}

// file.writable() -> bool
func fileWritable(vm *runtime.VM) int {
	f, ok := getFile(vm, 1)
	if !ok {
		return 0
	}

	vm.Push(runtime.NewBool(f.writable && !f.closed))
	return 1
}

// file.seekable() -> bool
func fileSeekable(vm *runtime.VM) int {
	f, ok := getFile(vm, 1)
	if !ok {
		return 0
	}

	vm.Push(runtime.NewBool(f.seekable && !f.closed))
	return 1
}

// file.__enter__() -> self
func fileEnter(vm *runtime.VM) int {
	// __enter__ just returns self (the file object)
	ud := vm.Get(1)
	vm.Push(ud)
	return 1
}

// file.__exit__(exc_type, exc_val, exc_tb) -> False
func fileExit(vm *runtime.VM) int {
	f, ok := getFile(vm, 1)
	if !ok {
		return 0
	}

	// Always close the file on exit
	if !f.closed {
		if f.writer != nil {
			_ = f.writer.Flush()
		}
		_ = f.file.Close()
		f.closed = true
	}

	// Return False to not suppress exceptions
	vm.Push(runtime.False)
	return 1
}

// file.__iter__() -> self
func fileIter(vm *runtime.VM) int {
	ud := vm.Get(1)
	vm.Push(ud)
	return 1
}

// file.__next__() -> next line
func fileNext(vm *runtime.VM) int {
	f, ok := getFile(vm, 1)
	if !ok {
		return 0
	}

	if f.closed {
		vm.RaiseError("ValueError: I/O operation on closed file")
		return 0
	}

	line, err := f.reader.ReadString('\n')
	if err == io.EOF && len(line) == 0 {
		vm.RaiseError("StopIteration")
		return 0
	}
	if err != nil && err != io.EOF {
		vm.RaiseError("IOError: %v", err)
		return 0
	}

	if f.binary {
		vm.Push(runtime.NewBytes([]byte(line)))
	} else {
		vm.Push(runtime.NewString(line))
	}
	return 1
}

// =====================================
// Property Getters (accessed without parentheses)
// =====================================

// file.name -> str (property)
func fileName(vm *runtime.VM) int {
	f, ok := getFile(vm, 1)
	if !ok {
		return 0
	}
	vm.Push(runtime.NewString(f.name))
	return 1
}

// file.mode -> str (property)
func fileMode(vm *runtime.VM) int {
	f, ok := getFile(vm, 1)
	if !ok {
		return 0
	}
	vm.Push(runtime.NewString(f.mode))
	return 1
}

// file.encoding -> str or None (property)
func fileEncoding(vm *runtime.VM) int {
	f, ok := getFile(vm, 1)
	if !ok {
		return 0
	}
	if f.binary {
		vm.Push(runtime.None)
	} else {
		vm.Push(runtime.NewString(f.encoding))
	}
	return 1
}

// file.closed -> bool (property)
func fileClosed(vm *runtime.VM) int {
	f, ok := getFile(vm, 1)
	if !ok {
		return 0
	}
	vm.Push(runtime.NewBool(f.closed))
	return 1
}

// =====================================
// StringIO — in-memory text stream
// =====================================

// PyStringIO is an in-memory text stream.
type PyStringIO struct {
	buf    []byte
	pos    int
	closed bool
}

func (s *PyStringIO) Type() string   { return "StringIO" }
func (s *PyStringIO) String() string { return "<_io.StringIO>" }

func getStringIO(vm *runtime.VM, argNum int) (*PyStringIO, bool) {
	ud := vm.ToUserData(argNum)
	if ud == nil {
		vm.RaiseError("expected StringIO object")
		return nil, false
	}
	s, ok := ud.Value.(*PyStringIO)
	if !ok {
		vm.RaiseError("expected StringIO object")
		return nil, false
	}
	return s, true
}

func newStringIOUserData(s *PyStringIO) *runtime.PyUserData {
	ud := runtime.NewUserData(s)
	ud.Metatable = runtime.NewDict()
	ud.Metatable.Items[runtime.NewString("__type__")] = runtime.NewString("StringIO")
	return ud
}

var stringIOConstructor = &runtime.PyBuiltinFunc{
	Name: "StringIO",
	Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
		s := &PyStringIO{}
		if len(args) >= 1 && !runtime.IsNone(args[0]) {
			str, ok := args[0].(*runtime.PyString)
			if !ok {
				return nil, fmt.Errorf("TypeError: initial_value must be str, not %T", args[0])
			}
			s.buf = []byte(str.Value)
		}
		return newStringIOUserData(s), nil
	},
}

func stringIORead(vm *runtime.VM) int {
	s, ok := getStringIO(vm, 1)
	if !ok {
		return 0
	}
	if s.closed {
		vm.RaiseError("ValueError: I/O operation on closed file")
		return 0
	}
	size := int64(-1)
	if vm.GetTop() >= 2 && !runtime.IsNone(vm.Get(2)) {
		size = vm.CheckInt(2)
	}
	// Clamp read position to buffer length (may be past end after seek)
	readPos := s.pos
	if readPos > len(s.buf) {
		readPos = len(s.buf)
	}
	if size < 0 {
		// Read all remaining
		data := s.buf[readPos:]
		if readPos < len(s.buf) {
			s.pos = len(s.buf)
		}
		// else: keep pos as-is (past end) — CPython preserves overseek position
		vm.Push(runtime.NewString(string(data)))
	} else {
		end := readPos + int(size)
		if end > len(s.buf) {
			end = len(s.buf)
		}
		if end > readPos {
			s.pos = end
		}
		// else: pos unchanged if nothing read
		data := s.buf[readPos:end]
		vm.Push(runtime.NewString(string(data)))
	}
	return 1
}

func stringIOReadline(vm *runtime.VM) int {
	s, ok := getStringIO(vm, 1)
	if !ok {
		return 0
	}
	if s.closed {
		vm.RaiseError("ValueError: I/O operation on closed file")
		return 0
	}
	limit := int64(-1)
	if vm.GetTop() >= 2 && !runtime.IsNone(vm.Get(2)) {
		limit = vm.CheckInt(2)
	}
	if s.pos >= len(s.buf) {
		vm.Push(runtime.NewString(""))
		return 1
	}
	// Find next newline
	rest := s.buf[s.pos:]
	idx := bytes.IndexByte(rest, '\n')
	var line []byte
	if idx >= 0 {
		line = rest[:idx+1]
	} else {
		line = rest
	}
	if limit >= 0 && int64(len(line)) > limit {
		line = line[:limit]
	}
	s.pos += len(line)
	vm.Push(runtime.NewString(string(line)))
	return 1
}

func stringIOReadlines(vm *runtime.VM) int {
	s, ok := getStringIO(vm, 1)
	if !ok {
		return 0
	}
	if s.closed {
		vm.RaiseError("ValueError: I/O operation on closed file")
		return 0
	}
	hint := int64(-1)
	if vm.GetTop() >= 2 && !runtime.IsNone(vm.Get(2)) {
		hint = vm.CheckInt(2)
	}
	var lines []runtime.Value
	var bytesRead int64
	for s.pos < len(s.buf) {
		rest := s.buf[s.pos:]
		idx := bytes.IndexByte(rest, '\n')
		var line []byte
		if idx >= 0 {
			line = rest[:idx+1]
		} else {
			line = rest
		}
		s.pos += len(line)
		lines = append(lines, runtime.NewString(string(line)))
		bytesRead += int64(len(line))
		if hint >= 0 && bytesRead >= hint {
			break
		}
	}
	vm.Push(runtime.NewList(lines))
	return 1
}

func stringIOWrite(vm *runtime.VM) int {
	s, ok := getStringIO(vm, 1)
	if !ok {
		return 0
	}
	if s.closed {
		vm.RaiseError("ValueError: I/O operation on closed file")
		return 0
	}
	arg := vm.Get(2)
	str, ok2 := arg.(*runtime.PyString)
	if !ok2 {
		vm.RaiseError("TypeError: string argument expected, got '%s'", getTypeName(arg))
		return 0
	}
	data := []byte(str.Value)
	n := len(data)

	// If writing at current position, may need to expand buffer
	if s.pos == len(s.buf) {
		s.buf = append(s.buf, data...)
	} else {
		// Overwrite from current position, extending if necessary
		end := s.pos + n
		if end > len(s.buf) {
			s.buf = append(s.buf[:s.pos], data...)
		} else {
			copy(s.buf[s.pos:end], data)
		}
	}
	s.pos += n
	vm.Push(runtime.NewInt(int64(n)))
	return 1
}

func stringIOWritelines(vm *runtime.VM) int {
	s, ok := getStringIO(vm, 1)
	if !ok {
		return 0
	}
	if s.closed {
		vm.RaiseError("ValueError: I/O operation on closed file")
		return 0
	}
	arg := vm.Get(2)
	list, ok2 := arg.(*runtime.PyList)
	if !ok2 {
		vm.RaiseError("TypeError: writelines() argument must be an iterable of strings")
		return 0
	}
	for _, item := range list.Items {
		str, ok3 := item.(*runtime.PyString)
		if !ok3 {
			vm.RaiseError("TypeError: writelines() argument must be an iterable of strings")
			return 0
		}
		data := []byte(str.Value)
		if s.pos == len(s.buf) {
			s.buf = append(s.buf, data...)
		} else {
			end := s.pos + len(data)
			if end > len(s.buf) {
				s.buf = append(s.buf[:s.pos], data...)
			} else {
				copy(s.buf[s.pos:end], data)
			}
		}
		s.pos += len(data)
	}
	return 0
}

func stringIOSeek(vm *runtime.VM) int {
	s, ok := getStringIO(vm, 1)
	if !ok {
		return 0
	}
	if s.closed {
		vm.RaiseError("ValueError: I/O operation on closed file")
		return 0
	}
	offset := int(vm.CheckInt(2))
	whence := 0
	if vm.GetTop() >= 3 && !runtime.IsNone(vm.Get(3)) {
		whence = int(vm.CheckInt(3))
	}
	var newPos int
	switch whence {
	case 0: // SEEK_SET
		newPos = offset
	case 1: // SEEK_CUR
		newPos = s.pos + offset
	case 2: // SEEK_END
		newPos = len(s.buf) + offset
	default:
		vm.RaiseError("ValueError: invalid whence (%d, should be 0, 1 or 2)", whence)
		return 0
	}
	if newPos < 0 {
		vm.RaiseError("ValueError: Negative seek position %d", newPos)
		return 0
	}
	s.pos = newPos
	vm.Push(runtime.NewInt(int64(newPos)))
	return 1
}

func stringIOTell(vm *runtime.VM) int {
	s, ok := getStringIO(vm, 1)
	if !ok {
		return 0
	}
	if s.closed {
		vm.RaiseError("ValueError: I/O operation on closed file")
		return 0
	}
	vm.Push(runtime.NewInt(int64(s.pos)))
	return 1
}

func stringIOGetvalue(vm *runtime.VM) int {
	s, ok := getStringIO(vm, 1)
	if !ok {
		return 0
	}
	if s.closed {
		vm.RaiseError("ValueError: I/O operation on closed file")
		return 0
	}
	vm.Push(runtime.NewString(string(s.buf)))
	return 1
}

func stringIOTruncate(vm *runtime.VM) int {
	s, ok := getStringIO(vm, 1)
	if !ok {
		return 0
	}
	if s.closed {
		vm.RaiseError("ValueError: I/O operation on closed file")
		return 0
	}
	size := int64(s.pos)
	if vm.GetTop() >= 2 && !runtime.IsNone(vm.Get(2)) {
		size = vm.CheckInt(2)
	}
	if size < 0 {
		vm.RaiseError("ValueError: Negative size value %d", size)
		return 0
	}
	if int(size) < len(s.buf) {
		s.buf = s.buf[:size]
	} else {
		// Pad with zero bytes if size exceeds current length
		for len(s.buf) < int(size) {
			s.buf = append(s.buf, 0)
		}
	}
	vm.Push(runtime.NewInt(size))
	return 1
}

func stringIOClose(vm *runtime.VM) int {
	s, ok := getStringIO(vm, 1)
	if !ok {
		return 0
	}
	s.closed = true
	return 0
}

func stringIOFlush(vm *runtime.VM) int {
	s, ok := getStringIO(vm, 1)
	if !ok {
		return 0
	}
	if s.closed {
		vm.RaiseError("ValueError: I/O operation on closed file")
		return 0
	}
	return 0 // No-op for in-memory streams
}

func stringIOIsatty(vm *runtime.VM) int {
	s, ok := getStringIO(vm, 1)
	if !ok {
		return 0
	}
	if s.closed {
		vm.RaiseError("ValueError: I/O operation on closed file")
		return 0
	}
	vm.Push(runtime.False)
	return 1
}

func stringIOReadable(vm *runtime.VM) int {
	vm.Push(runtime.True)
	return 1
}

func stringIOWritable(vm *runtime.VM) int {
	vm.Push(runtime.True)
	return 1
}

func stringIOSeekable(vm *runtime.VM) int {
	vm.Push(runtime.True)
	return 1
}

func stringIOClosed(vm *runtime.VM) int {
	s, ok := getStringIO(vm, 1)
	if !ok {
		return 0
	}
	vm.Push(runtime.NewBool(s.closed))
	return 1
}

func stringIOEnter(vm *runtime.VM) int {
	ud := vm.Get(1)
	vm.Push(ud)
	return 1
}

func stringIOExit(vm *runtime.VM) int {
	s, ok := getStringIO(vm, 1)
	if !ok {
		return 0
	}
	s.closed = true
	vm.Push(runtime.False)
	return 1
}

func stringIOIter(vm *runtime.VM) int {
	ud := vm.Get(1)
	vm.Push(ud)
	return 1
}

func stringIONext(vm *runtime.VM) int {
	s, ok := getStringIO(vm, 1)
	if !ok {
		return 0
	}
	if s.closed {
		vm.RaiseError("ValueError: I/O operation on closed file")
		return 0
	}
	if s.pos >= len(s.buf) {
		vm.RaiseError("StopIteration")
		return 0
	}
	rest := s.buf[s.pos:]
	idx := bytes.IndexByte(rest, '\n')
	var line []byte
	if idx >= 0 {
		line = rest[:idx+1]
	} else {
		line = rest
	}
	s.pos += len(line)
	vm.Push(runtime.NewString(string(line)))
	return 1
}

// =====================================
// BytesIO — in-memory binary stream
// =====================================

// PyBytesIO is an in-memory binary stream.
type PyBytesIO struct {
	buf    []byte
	pos    int
	closed bool
}

func (b *PyBytesIO) Type() string   { return "BytesIO" }
func (b *PyBytesIO) String() string { return "<_io.BytesIO>" }

func getBytesIO(vm *runtime.VM, argNum int) (*PyBytesIO, bool) {
	ud := vm.ToUserData(argNum)
	if ud == nil {
		vm.RaiseError("expected BytesIO object")
		return nil, false
	}
	b, ok := ud.Value.(*PyBytesIO)
	if !ok {
		vm.RaiseError("expected BytesIO object")
		return nil, false
	}
	return b, true
}

func newBytesIOUserData(b *PyBytesIO) *runtime.PyUserData {
	ud := runtime.NewUserData(b)
	ud.Metatable = runtime.NewDict()
	ud.Metatable.Items[runtime.NewString("__type__")] = runtime.NewString("BytesIO")
	return ud
}

var bytesIOConstructor = &runtime.PyBuiltinFunc{
	Name: "BytesIO",
	Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
		b := &PyBytesIO{}
		if len(args) >= 1 && !runtime.IsNone(args[0]) {
			switch v := args[0].(type) {
			case *runtime.PyBytes:
				b.buf = make([]byte, len(v.Value))
				copy(b.buf, v.Value)
			case *runtime.PyString:
				b.buf = []byte(v.Value)
			default:
				return nil, fmt.Errorf("TypeError: a bytes-like object is required, not '%s'", getTypeName(args[0]))
			}
		}
		return newBytesIOUserData(b), nil
	},
}

func bytesIORead(vm *runtime.VM) int {
	b, ok := getBytesIO(vm, 1)
	if !ok {
		return 0
	}
	if b.closed {
		vm.RaiseError("ValueError: I/O operation on closed file")
		return 0
	}
	size := int64(-1)
	if vm.GetTop() >= 2 && !runtime.IsNone(vm.Get(2)) {
		size = vm.CheckInt(2)
	}
	// Clamp read position to buffer length (may be past end after seek)
	readPos := b.pos
	if readPos > len(b.buf) {
		readPos = len(b.buf)
	}
	if size < 0 {
		data := make([]byte, len(b.buf)-readPos)
		copy(data, b.buf[readPos:])
		if readPos < len(b.buf) {
			b.pos = len(b.buf)
		}
		vm.Push(runtime.NewBytes(data))
	} else {
		end := readPos + int(size)
		if end > len(b.buf) {
			end = len(b.buf)
		}
		data := make([]byte, end-readPos)
		copy(data, b.buf[readPos:end])
		if end > readPos {
			b.pos = end
		}
		vm.Push(runtime.NewBytes(data))
	}
	return 1
}

func bytesIOReadline(vm *runtime.VM) int {
	b, ok := getBytesIO(vm, 1)
	if !ok {
		return 0
	}
	if b.closed {
		vm.RaiseError("ValueError: I/O operation on closed file")
		return 0
	}
	limit := int64(-1)
	if vm.GetTop() >= 2 && !runtime.IsNone(vm.Get(2)) {
		limit = vm.CheckInt(2)
	}
	if b.pos >= len(b.buf) {
		vm.Push(runtime.NewBytes(nil))
		return 1
	}
	rest := b.buf[b.pos:]
	idx := bytes.IndexByte(rest, '\n')
	var line []byte
	if idx >= 0 {
		line = rest[:idx+1]
	} else {
		line = rest
	}
	if limit >= 0 && int64(len(line)) > limit {
		line = line[:limit]
	}
	result := make([]byte, len(line))
	copy(result, line)
	b.pos += len(line)
	vm.Push(runtime.NewBytes(result))
	return 1
}

func bytesIOReadlines(vm *runtime.VM) int {
	b, ok := getBytesIO(vm, 1)
	if !ok {
		return 0
	}
	if b.closed {
		vm.RaiseError("ValueError: I/O operation on closed file")
		return 0
	}
	hint := int64(-1)
	if vm.GetTop() >= 2 && !runtime.IsNone(vm.Get(2)) {
		hint = vm.CheckInt(2)
	}
	var lines []runtime.Value
	var bytesRead int64
	for b.pos < len(b.buf) {
		rest := b.buf[b.pos:]
		idx := bytes.IndexByte(rest, '\n')
		var line []byte
		if idx >= 0 {
			line = rest[:idx+1]
		} else {
			line = rest
		}
		result := make([]byte, len(line))
		copy(result, line)
		b.pos += len(line)
		lines = append(lines, runtime.NewBytes(result))
		bytesRead += int64(len(line))
		if hint >= 0 && bytesRead >= hint {
			break
		}
	}
	vm.Push(runtime.NewList(lines))
	return 1
}

func bytesIOWrite(vm *runtime.VM) int {
	b, ok := getBytesIO(vm, 1)
	if !ok {
		return 0
	}
	if b.closed {
		vm.RaiseError("ValueError: I/O operation on closed file")
		return 0
	}
	arg := vm.Get(2)
	var data []byte
	switch v := arg.(type) {
	case *runtime.PyBytes:
		data = v.Value
	case *runtime.PyString:
		data = []byte(v.Value)
	default:
		vm.RaiseError("TypeError: a bytes-like object is required, not '%s'", getTypeName(arg))
		return 0
	}
	n := len(data)
	if b.pos == len(b.buf) {
		b.buf = append(b.buf, data...)
	} else {
		end := b.pos + n
		if end > len(b.buf) {
			b.buf = append(b.buf[:b.pos], data...)
		} else {
			copy(b.buf[b.pos:end], data)
		}
	}
	b.pos += n
	vm.Push(runtime.NewInt(int64(n)))
	return 1
}

func bytesIOWritelines(vm *runtime.VM) int {
	b, ok := getBytesIO(vm, 1)
	if !ok {
		return 0
	}
	if b.closed {
		vm.RaiseError("ValueError: I/O operation on closed file")
		return 0
	}
	arg := vm.Get(2)
	list, ok2 := arg.(*runtime.PyList)
	if !ok2 {
		vm.RaiseError("TypeError: writelines() argument must be an iterable of bytes")
		return 0
	}
	for _, item := range list.Items {
		var data []byte
		switch v := item.(type) {
		case *runtime.PyBytes:
			data = v.Value
		case *runtime.PyString:
			data = []byte(v.Value)
		default:
			vm.RaiseError("TypeError: a bytes-like object is required, not '%s'", getTypeName(item))
			return 0
		}
		if b.pos == len(b.buf) {
			b.buf = append(b.buf, data...)
		} else {
			end := b.pos + len(data)
			if end > len(b.buf) {
				b.buf = append(b.buf[:b.pos], data...)
			} else {
				copy(b.buf[b.pos:end], data)
			}
		}
		b.pos += len(data)
	}
	return 0
}

func bytesIOSeek(vm *runtime.VM) int {
	b, ok := getBytesIO(vm, 1)
	if !ok {
		return 0
	}
	if b.closed {
		vm.RaiseError("ValueError: I/O operation on closed file")
		return 0
	}
	offset := int(vm.CheckInt(2))
	whence := 0
	if vm.GetTop() >= 3 && !runtime.IsNone(vm.Get(3)) {
		whence = int(vm.CheckInt(3))
	}
	var newPos int
	switch whence {
	case 0:
		newPos = offset
		if newPos < 0 {
			vm.RaiseError("ValueError: Negative seek position %d", newPos)
			return 0
		}
	case 1:
		newPos = b.pos + offset
		if newPos < 0 {
			newPos = 0
		}
	case 2:
		newPos = len(b.buf) + offset
		if newPos < 0 {
			newPos = 0
		}
	default:
		vm.RaiseError("ValueError: invalid whence (%d, should be 0, 1 or 2)", whence)
		return 0
	}
	b.pos = newPos
	vm.Push(runtime.NewInt(int64(newPos)))
	return 1
}

func bytesIOTell(vm *runtime.VM) int {
	b, ok := getBytesIO(vm, 1)
	if !ok {
		return 0
	}
	if b.closed {
		vm.RaiseError("ValueError: I/O operation on closed file")
		return 0
	}
	vm.Push(runtime.NewInt(int64(b.pos)))
	return 1
}

func bytesIOGetvalue(vm *runtime.VM) int {
	b, ok := getBytesIO(vm, 1)
	if !ok {
		return 0
	}
	if b.closed {
		vm.RaiseError("ValueError: I/O operation on closed file")
		return 0
	}
	result := make([]byte, len(b.buf))
	copy(result, b.buf)
	vm.Push(runtime.NewBytes(result))
	return 1
}

func bytesIOTruncate(vm *runtime.VM) int {
	b, ok := getBytesIO(vm, 1)
	if !ok {
		return 0
	}
	if b.closed {
		vm.RaiseError("ValueError: I/O operation on closed file")
		return 0
	}
	size := int64(b.pos)
	if vm.GetTop() >= 2 && !runtime.IsNone(vm.Get(2)) {
		size = vm.CheckInt(2)
	}
	if size < 0 {
		vm.RaiseError("ValueError: Negative size value %d", size)
		return 0
	}
	if int(size) < len(b.buf) {
		b.buf = b.buf[:size]
	} else {
		for len(b.buf) < int(size) {
			b.buf = append(b.buf, 0)
		}
	}
	vm.Push(runtime.NewInt(size))
	return 1
}

func bytesIOClose(vm *runtime.VM) int {
	b, ok := getBytesIO(vm, 1)
	if !ok {
		return 0
	}
	b.closed = true
	return 0
}

func bytesIOFlush(vm *runtime.VM) int {
	b, ok := getBytesIO(vm, 1)
	if !ok {
		return 0
	}
	if b.closed {
		vm.RaiseError("ValueError: I/O operation on closed file")
		return 0
	}
	return 0
}

func bytesIOIsatty(vm *runtime.VM) int {
	b, ok := getBytesIO(vm, 1)
	if !ok {
		return 0
	}
	if b.closed {
		vm.RaiseError("ValueError: I/O operation on closed file")
		return 0
	}
	vm.Push(runtime.False)
	return 1
}

func bytesIOReadable(vm *runtime.VM) int {
	vm.Push(runtime.True)
	return 1
}

func bytesIOWritable(vm *runtime.VM) int {
	vm.Push(runtime.True)
	return 1
}

func bytesIOSeekable(vm *runtime.VM) int {
	vm.Push(runtime.True)
	return 1
}

func bytesIOClosed(vm *runtime.VM) int {
	b, ok := getBytesIO(vm, 1)
	if !ok {
		return 0
	}
	vm.Push(runtime.NewBool(b.closed))
	return 1
}

func bytesIOEnter(vm *runtime.VM) int {
	ud := vm.Get(1)
	vm.Push(ud)
	return 1
}

func bytesIOExit(vm *runtime.VM) int {
	b, ok := getBytesIO(vm, 1)
	if !ok {
		return 0
	}
	b.closed = true
	vm.Push(runtime.False)
	return 1
}

func bytesIOIter(vm *runtime.VM) int {
	ud := vm.Get(1)
	vm.Push(ud)
	return 1
}

func bytesIONext(vm *runtime.VM) int {
	b, ok := getBytesIO(vm, 1)
	if !ok {
		return 0
	}
	if b.closed {
		vm.RaiseError("ValueError: I/O operation on closed file")
		return 0
	}
	if b.pos >= len(b.buf) {
		vm.RaiseError("StopIteration")
		return 0
	}
	rest := b.buf[b.pos:]
	idx := bytes.IndexByte(rest, '\n')
	var line []byte
	if idx >= 0 {
		line = rest[:idx+1]
	} else {
		line = rest
	}
	result := make([]byte, len(line))
	copy(result, line)
	b.pos += len(line)
	vm.Push(runtime.NewBytes(result))
	return 1
}
