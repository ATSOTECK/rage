package stdlib

import (
	"bufio"
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

	// Register open() as a pending builtin
	runtime.RegisterPendingBuiltin("open", builtinOpen)

	// Also register an io module with constants
	runtime.NewModuleBuilder("io").
		Doc("Core tools for working with streams.").
		Const("SEEK_SET", runtime.NewInt(0)).
		Const("SEEK_CUR", runtime.NewInt(1)).
		Const("SEEK_END", runtime.NewInt(2)).
		Func("open", builtinOpen).
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

// builtinOpen implements the open() builtin function
// open(file, mode='r', encoding='utf-8') -> file object
func builtinOpen(vm *runtime.VM) int {
	nargs := vm.GetTop()
	if nargs < 1 {
		vm.RaiseError("open() missing required argument: 'file'")
		return 0
	}

	filename := vm.CheckString(1)

	mode := "r"
	if nargs >= 2 && !runtime.IsNone(vm.Get(2)) {
		mode = vm.CheckString(2)
	}

	encoding := "utf-8"
	if nargs >= 3 && !runtime.IsNone(vm.Get(3)) {
		encoding = vm.CheckString(3)
	}

	// Parse mode
	readable, writable, binary, _, flag, err := parseMode(mode)
	if err != nil {
		vm.RaiseError("ValueError: %v", err)
		return 0
	}

	// Open the file
	file, err := os.OpenFile(filename, flag, 0644)
	if err != nil {
		if os.IsNotExist(err) {
			vm.RaiseError("FileNotFoundError: [Errno 2] No such file or directory: '%s'", filename)
		} else if os.IsPermission(err) {
			vm.RaiseError("PermissionError: [Errno 13] Permission denied: '%s'", filename)
		} else if os.IsExist(err) {
			vm.RaiseError("FileExistsError: [Errno 17] File exists: '%s'", filename)
		} else {
			vm.RaiseError("OSError: %v", err)
		}
		return 0
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
	vm.Push(ud)
	return 1
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
		f.writer.Flush()
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
		f.writer.Flush()
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
	f.file.Sync()

	return 0
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
			f.writer.Flush()
		}
		f.file.Close()
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
