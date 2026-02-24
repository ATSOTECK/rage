package runtime

import (
	"errors"
	"fmt"
	"testing"
)

// =====================================
// createException: from *PyClass
// =====================================

func TestCreateExceptionFromClass(t *testing.T) {
	vm := NewVM()

	cls := vm.builtinClass("ValueError")
	if cls == nil {
		t.Fatal("ValueError class not found")
	}

	exc := vm.createException(cls, nil)
	if exc == nil {
		t.Fatal("createException returned nil")
	}
	if exc.ExcType != cls {
		t.Errorf("ExcType = %v, want ValueError class", exc.ExcType)
	}
	if exc.Args == nil {
		t.Error("Args should not be nil")
	}
	if exc.Cause != nil {
		t.Error("Cause should be nil when no cause provided")
	}
}

func TestCreateExceptionFromNonExceptionClass(t *testing.T) {
	vm := NewVM()

	// Create a class that does NOT inherit from BaseException
	cls := &PyClass{
		Name:  "MyClass",
		Bases: nil,
		Dict:  make(map[string]Value),
	}
	cls.Mro = []*PyClass{cls}

	exc := vm.createException(cls, nil)
	if exc == nil {
		t.Fatal("createException returned nil")
	}
	// Should be wrapped as TypeError
	if exc.ExcType == nil || exc.ExcType.Name != "TypeError" {
		typeName := "<nil>"
		if exc.ExcType != nil {
			typeName = exc.ExcType.Name
		}
		t.Errorf("expected TypeError for non-exception class, got %s", typeName)
	}
}

// =====================================
// createException: from *PyString
// =====================================

func TestCreateExceptionFromString(t *testing.T) {
	vm := NewVM()

	msg := &PyString{Value: "something went wrong"}
	exc := vm.createException(msg, nil)

	if exc == nil {
		t.Fatal("createException returned nil")
	}
	// String exceptions get wrapped as Exception type
	if exc.ExcType == nil || exc.ExcType.Name != "Exception" {
		t.Error("string exception should have ExcType = Exception")
	}
	if exc.Message != "something went wrong" {
		t.Errorf("Message = %q, want %q", exc.Message, "something went wrong")
	}
	if exc.Args == nil || len(exc.Args.Items) != 1 {
		t.Fatal("expected 1 arg")
	}
	argStr, ok := exc.Args.Items[0].(*PyString)
	if !ok || argStr.Value != "something went wrong" {
		t.Errorf("args[0] = %v, want 'something went wrong'", exc.Args.Items[0])
	}
}

// =====================================
// createException: from *PyInstance
// =====================================

func TestCreateExceptionFromInstance(t *testing.T) {
	vm := NewVM()

	cls := vm.builtinClass("ValueError")
	inst := &PyInstance{
		Class: cls,
		Dict: map[string]Value{
			"args": &PyTuple{Items: []Value{&PyString{Value: "bad value"}}},
		},
	}

	exc := vm.createException(inst, nil)
	if exc == nil {
		t.Fatal("createException returned nil")
	}
	if exc.ExcType != cls {
		t.Error("ExcType should be ValueError")
	}
	if exc.Args == nil || len(exc.Args.Items) != 1 {
		t.Fatal("expected 1 arg from instance")
	}
}

func TestCreateExceptionFromNonExceptionInstance(t *testing.T) {
	vm := NewVM()

	cls := &PyClass{
		Name:  "NotAnException",
		Bases: nil,
		Dict:  make(map[string]Value),
	}
	cls.Mro = []*PyClass{cls}
	inst := &PyInstance{
		Class: cls,
		Dict:  map[string]Value{},
	}

	exc := vm.createException(inst, nil)
	if exc.ExcType == nil || exc.ExcType.Name != "TypeError" {
		t.Error("non-exception instance should produce TypeError")
	}
}

// =====================================
// createException: from *PyException (already an exception)
// =====================================

func TestCreateExceptionFromPyException(t *testing.T) {
	vm := NewVM()

	original := &PyException{
		ExcType: vm.builtinClass("ValueError"),
		Message: "original",
		Args:    &PyTuple{Items: []Value{&PyString{Value: "original"}}},
	}

	result := vm.createException(original, nil)
	// Should return the same object
	if result != original {
		t.Error("createException on *PyException should return same object")
	}
}

// =====================================
// createException: with cause (explicit chaining)
// =====================================

func TestCreateExceptionWithCause(t *testing.T) {
	vm := NewVM()

	causeClass := vm.builtinClass("TypeError")
	cause := &PyException{
		ExcType: causeClass,
		Message: "the cause",
		Args:    &PyTuple{Items: []Value{&PyString{Value: "the cause"}}},
	}

	mainClass := vm.builtinClass("ValueError")
	exc := vm.createException(mainClass, cause)

	if exc.Cause == nil {
		t.Fatal("Cause should not be nil")
	}
	if exc.Cause != cause {
		t.Error("Cause should be the cause exception")
	}
	if !exc.SuppressContext {
		t.Error("SuppressContext should be true when Cause is set")
	}
}

func TestCreateExceptionWithNoneCause(t *testing.T) {
	vm := NewVM()

	mainClass := vm.builtinClass("ValueError")
	exc := vm.createException(mainClass, None)

	if exc.Cause != nil {
		t.Error("Cause should be nil when cause is None")
	}
	if !exc.SuppressContext {
		t.Error("SuppressContext should be true even with None cause")
	}
}

func TestCreateExceptionCauseChaining(t *testing.T) {
	vm := NewVM()

	// Create cause from a class
	causeClass := vm.builtinClass("TypeError")
	mainClass := vm.builtinClass("ValueError")

	exc := vm.createException(mainClass, causeClass)
	if exc.Cause == nil {
		t.Fatal("Cause should not be nil")
	}
	if exc.Cause.ExcType != causeClass {
		t.Error("Cause ExcType should be TypeError")
	}
}

// Attaching cause to an existing *PyException
func TestCreateExceptionAttachCauseToPyException(t *testing.T) {
	vm := NewVM()

	original := &PyException{
		ExcType: vm.builtinClass("RuntimeError"),
		Message: "runtime issue",
		Args:    &PyTuple{Items: []Value{&PyString{Value: "runtime issue"}}},
	}

	cause := &PyException{
		ExcType: vm.builtinClass("OSError"),
		Message: "os issue",
		Args:    &PyTuple{Items: []Value{&PyString{Value: "os issue"}}},
	}

	result := vm.createException(original, cause)
	if result != original {
		t.Error("should return same exception object")
	}
	if result.Cause != cause {
		t.Error("Cause should be attached")
	}
	if !result.SuppressContext {
		t.Error("SuppressContext should be true")
	}
}

// =====================================
// exceptionMatches: class matching
// =====================================

func TestExceptionMatchesExactClass(t *testing.T) {
	vm := NewVM()

	cls := vm.builtinClass("ValueError")
	exc := &PyException{ExcType: cls}

	if !vm.exceptionMatches(exc, cls) {
		t.Error("exception should match its own class")
	}
}

func TestExceptionMatchesParentClass(t *testing.T) {
	vm := NewVM()

	// KeyError -> LookupError -> Exception -> BaseException
	keyErrCls := vm.builtinClass("KeyError")
	lookupErrCls := vm.builtinClass("LookupError")
	exceptionCls := vm.builtinClass("Exception")
	baseExcCls := vm.builtinClass("BaseException")

	exc := &PyException{ExcType: keyErrCls}

	if !vm.exceptionMatches(exc, lookupErrCls) {
		t.Error("KeyError should match LookupError (parent)")
	}
	if !vm.exceptionMatches(exc, exceptionCls) {
		t.Error("KeyError should match Exception (grandparent)")
	}
	if !vm.exceptionMatches(exc, baseExcCls) {
		t.Error("KeyError should match BaseException (root)")
	}
}

func TestExceptionDoesNotMatchUnrelatedClass(t *testing.T) {
	vm := NewVM()

	valErrCls := vm.builtinClass("ValueError")
	typeErrCls := vm.builtinClass("TypeError")

	exc := &PyException{ExcType: valErrCls}

	if vm.exceptionMatches(exc, typeErrCls) {
		t.Error("ValueError should NOT match TypeError")
	}
}

// =====================================
// exceptionMatches: tuple of types
// =====================================

func TestExceptionMatchesTupleOfTypes(t *testing.T) {
	vm := NewVM()

	valErrCls := vm.builtinClass("ValueError")
	typeErrCls := vm.builtinClass("TypeError")
	keyErrCls := vm.builtinClass("KeyError")

	exc := &PyException{ExcType: valErrCls}

	tuple := &PyTuple{Items: []Value{typeErrCls, keyErrCls, valErrCls}}
	if !vm.exceptionMatches(exc, tuple) {
		t.Error("ValueError should match (TypeError, KeyError, ValueError)")
	}

	noMatchTuple := &PyTuple{Items: []Value{typeErrCls, keyErrCls}}
	if vm.exceptionMatches(exc, noMatchTuple) {
		t.Error("ValueError should NOT match (TypeError, KeyError)")
	}
}

// =====================================
// exceptionMatches: MRO matching through tuple
// =====================================

func TestExceptionMatchesMROThroughTuple(t *testing.T) {
	vm := NewVM()

	keyErrCls := vm.builtinClass("KeyError")
	exceptionCls := vm.builtinClass("Exception")

	exc := &PyException{ExcType: keyErrCls}

	tuple := &PyTuple{Items: []Value{exceptionCls}}
	if !vm.exceptionMatches(exc, tuple) {
		t.Error("KeyError should match (Exception,) via MRO")
	}
}

// =====================================
// exceptionMatches: non-class type returns false
// =====================================

func TestExceptionMatchesNonClassReturnsFalse(t *testing.T) {
	vm := NewVM()

	exc := &PyException{ExcType: vm.builtinClass("ValueError")}

	if vm.exceptionMatches(exc, MakeInt(42)) {
		t.Error("should not match an integer")
	}
	if vm.exceptionMatches(exc, &PyString{Value: "ValueError"}) {
		t.Error("should not match a string")
	}
	if vm.exceptionMatches(exc, None) {
		t.Error("should not match None")
	}
}

// =====================================
// exceptionMatches: TypeName fallback (no ExcType)
// =====================================

func TestExceptionMatchesTypeNameFallback(t *testing.T) {
	vm := NewVM()

	exc := &PyException{TypeName: "ValueError"}

	valErrCls := vm.builtinClass("ValueError")
	if !vm.exceptionMatches(exc, valErrCls) {
		t.Error("exception with TypeName='ValueError' should match ValueError class")
	}

	typeErrCls := vm.builtinClass("TypeError")
	if vm.exceptionMatches(exc, typeErrCls) {
		t.Error("exception with TypeName='ValueError' should NOT match TypeError class")
	}
}

// =====================================
// wrapGoError: type detection from error strings
// =====================================

func TestWrapGoErrorValueError(t *testing.T) {
	vm := NewVM()

	err := fmt.Errorf("ValueError: invalid literal for int()")
	exc := vm.wrapGoError(err)

	if exc.ExcType == nil || exc.ExcType.Name != "ValueError" {
		t.Errorf("expected ValueError, got %v", exc.ExcType)
	}
	if exc.Message != "invalid literal for int()" {
		t.Errorf("Message = %q, want %q", exc.Message, "invalid literal for int()")
	}
}

func TestWrapGoErrorTypeError(t *testing.T) {
	vm := NewVM()

	err := fmt.Errorf("TypeError: unsupported operand type(s)")
	exc := vm.wrapGoError(err)

	if exc.ExcType == nil || exc.ExcType.Name != "TypeError" {
		t.Errorf("expected TypeError, got %v", exc.ExcType)
	}
}

func TestWrapGoErrorKeyError(t *testing.T) {
	vm := NewVM()

	err := fmt.Errorf("KeyError: 'missing'")
	exc := vm.wrapGoError(err)

	if exc.ExcType == nil || exc.ExcType.Name != "KeyError" {
		t.Errorf("expected KeyError, got %v", exc.ExcType)
	}
}

func TestWrapGoErrorIndexError(t *testing.T) {
	vm := NewVM()

	err := fmt.Errorf("IndexError: list index out of range")
	exc := vm.wrapGoError(err)

	if exc.ExcType == nil || exc.ExcType.Name != "IndexError" {
		t.Errorf("expected IndexError, got %v", exc.ExcType)
	}
}

func TestWrapGoErrorZeroDivisionError(t *testing.T) {
	vm := NewVM()

	err := fmt.Errorf("ZeroDivisionError: division by zero")
	exc := vm.wrapGoError(err)

	if exc.ExcType == nil || exc.ExcType.Name != "ZeroDivisionError" {
		t.Errorf("expected ZeroDivisionError, got %v", exc.ExcType)
	}
}

func TestWrapGoErrorFallbackToRuntimeError(t *testing.T) {
	vm := NewVM()

	err := errors.New("something unknown happened")
	exc := vm.wrapGoError(err)

	if exc.ExcType == nil || exc.ExcType.Name != "RuntimeError" {
		t.Errorf("expected RuntimeError fallback, got %v", exc.ExcType)
	}
	if exc.Message != "something unknown happened" {
		t.Errorf("Message = %q, want %q", exc.Message, "something unknown happened")
	}
}

func TestWrapGoErrorPassthroughPyException(t *testing.T) {
	vm := NewVM()

	original := &PyException{
		ExcType: vm.builtinClass("ValueError"),
		Message: "already an exception",
	}

	result := vm.wrapGoError(original)
	if result != original {
		t.Error("wrapGoError should return the same *PyException")
	}
}

func TestWrapGoErrorOSError(t *testing.T) {
	vm := NewVM()

	err := fmt.Errorf("OSError: [Errno 2] No such file or directory")
	exc := vm.wrapGoError(err)

	if exc.ExcType == nil || exc.ExcType.Name != "OSError" {
		t.Errorf("expected OSError, got %v", exc.ExcType)
	}
}

func TestWrapGoErrorFileNotFoundError(t *testing.T) {
	vm := NewVM()

	err := fmt.Errorf("FileNotFoundError: No such file or directory: 'foo.txt'")
	exc := vm.wrapGoError(err)

	if exc.ExcType == nil || exc.ExcType.Name != "FileNotFoundError" {
		t.Errorf("expected FileNotFoundError, got %v", exc.ExcType)
	}
}

// =====================================
// isExceptionClass
// =====================================

func TestIsExceptionClass(t *testing.T) {
	vm := NewVM()

	tests := []struct {
		name string
		want bool
	}{
		{"ValueError", true},
		{"TypeError", true},
		{"BaseException", true},
		{"Exception", true},
		{"KeyError", true},
		{"StopIteration", true},
		{"GeneratorExit", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cls := vm.builtinClass(tt.name)
			if cls == nil {
				t.Fatalf("class %q not found", tt.name)
			}
			got := vm.isExceptionClass(cls)
			if got != tt.want {
				t.Errorf("isExceptionClass(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestIsExceptionClassFalseForNonException(t *testing.T) {
	vm := NewVM()

	cls := &PyClass{
		Name:  "RegularClass",
		Bases: nil,
		Dict:  make(map[string]Value),
	}
	cls.Mro = []*PyClass{cls}

	if vm.isExceptionClass(cls) {
		t.Error("non-exception class should not be identified as exception class")
	}
}

// =====================================
// buildTraceback
// =====================================

func TestBuildTracebackEmpty(t *testing.T) {
	vm := NewVM()
	// No frames
	vm.frames = nil

	tb := vm.buildTraceback()
	if len(tb) != 0 {
		t.Errorf("expected empty traceback, got %d entries", len(tb))
	}
}

func TestBuildTracebackSingleFrame(t *testing.T) {
	vm := NewVM()

	code := &CodeObject{
		Name:     "<module>",
		Filename: "test.py",
		LineNoTab: []LineEntry{
			{StartOffset: 0, EndOffset: 10, Line: 5},
		},
	}
	frame := &Frame{
		Code: code,
		IP:   3,
	}
	vm.frames = []*Frame{frame}

	tb := vm.buildTraceback()
	if len(tb) != 1 {
		t.Fatalf("expected 1 traceback entry, got %d", len(tb))
	}
	if tb[0].Filename != "test.py" {
		t.Errorf("Filename = %q, want %q", tb[0].Filename, "test.py")
	}
	if tb[0].Function != "<module>" {
		t.Errorf("Function = %q, want %q", tb[0].Function, "<module>")
	}
	if tb[0].Line != 5 {
		t.Errorf("Line = %d, want 5", tb[0].Line)
	}
}

func TestBuildTracebackMultipleFrames(t *testing.T) {
	vm := NewVM()

	code1 := &CodeObject{
		Name:     "<module>",
		Filename: "main.py",
		LineNoTab: []LineEntry{
			{StartOffset: 0, EndOffset: 20, Line: 1},
		},
	}
	code2 := &CodeObject{
		Name:     "foo",
		Filename: "main.py",
		LineNoTab: []LineEntry{
			{StartOffset: 0, EndOffset: 10, Line: 10},
		},
	}
	frame1 := &Frame{Code: code1, IP: 5}
	frame2 := &Frame{Code: code2, IP: 3}
	vm.frames = []*Frame{frame1, frame2}

	tb := vm.buildTraceback()
	if len(tb) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(tb))
	}
	// Traceback is built in reverse order (most recent frame first)
	if tb[0].Function != "foo" {
		t.Errorf("tb[0].Function = %q, want 'foo'", tb[0].Function)
	}
	if tb[1].Function != "<module>" {
		t.Errorf("tb[1].Function = %q, want '<module>'", tb[1].Function)
	}
}

// =====================================
// PyException formatting
// =====================================

func TestPyExceptionTypeMethod(t *testing.T) {
	vm := NewVM()

	exc := &PyException{
		ExcType: vm.builtinClass("ValueError"),
		Message: "bad value",
	}
	if exc.Type() != "ValueError" {
		t.Errorf("Type() = %q, want %q", exc.Type(), "ValueError")
	}

	// TypeName fallback
	exc2 := &PyException{
		TypeName: "CustomError",
		Message:  "custom",
	}
	if exc2.Type() != "CustomError" {
		t.Errorf("Type() = %q, want %q", exc2.Type(), "CustomError")
	}
}

func TestPyExceptionError(t *testing.T) {
	vm := NewVM()

	exc := &PyException{
		ExcType: vm.builtinClass("ValueError"),
		Args:    &PyTuple{Items: []Value{&PyString{Value: "bad input"}}},
		Message: "bad input",
	}
	errStr := exc.Error()
	if errStr != "ValueError: bad input" {
		t.Errorf("Error() = %q, want %q", errStr, "ValueError: bad input")
	}
}

// =====================================
// Exception chaining fields
// =====================================

func TestExceptionCauseAndContext(t *testing.T) {
	vm := NewVM()

	cause := &PyException{
		ExcType: vm.builtinClass("OSError"),
		Message: "disk full",
		Args:    &PyTuple{Items: []Value{&PyString{Value: "disk full"}}},
	}

	context := &PyException{
		ExcType: vm.builtinClass("TypeError"),
		Message: "wrong type",
		Args:    &PyTuple{Items: []Value{&PyString{Value: "wrong type"}}},
	}

	exc := &PyException{
		ExcType:        vm.builtinClass("ValueError"),
		Message:        "bad value",
		Args:           &PyTuple{Items: []Value{&PyString{Value: "bad value"}}},
		Cause:          cause,
		Context:        context,
		SuppressContext: true,
	}

	if exc.Cause != cause {
		t.Error("Cause not set correctly")
	}
	if exc.Context != context {
		t.Error("Context not set correctly")
	}
	if !exc.SuppressContext {
		t.Error("SuppressContext should be true")
	}
}

// =====================================
// wrapGoError: various exception types
// =====================================

func TestWrapGoErrorAllCommonTypes(t *testing.T) {
	vm := NewVM()

	tests := []struct {
		errMsg   string
		wantType string
	}{
		{"ValueError: bad", "ValueError"},
		{"TypeError: wrong", "TypeError"},
		{"KeyError: 'x'", "KeyError"},
		{"IndexError: out of range", "IndexError"},
		{"AttributeError: no attr", "AttributeError"},
		{"RuntimeError: runtime", "RuntimeError"},
		{"StopIteration: done", "StopIteration"},
		{"ZeroDivisionError: div by zero", "ZeroDivisionError"},
		{"OverflowError: too big", "OverflowError"},
		{"RecursionError: too deep", "RecursionError"},
		{"NameError: undefined", "NameError"},
		{"ImportError: no module", "ImportError"},
		{"MemoryError: out of memory", "MemoryError"},
		{"NotImplementedError: not impl", "NotImplementedError"},
	}

	for _, tt := range tests {
		t.Run(tt.wantType, func(t *testing.T) {
			err := fmt.Errorf("%s", tt.errMsg)
			exc := vm.wrapGoError(err)
			if exc.ExcType == nil {
				t.Fatalf("ExcType is nil for %q", tt.errMsg)
			}
			if exc.ExcType.Name != tt.wantType {
				t.Errorf("got type %q, want %q", exc.ExcType.Name, tt.wantType)
			}
		})
	}
}

// =====================================
// wrapGoError: message extraction
// =====================================

func TestWrapGoErrorMessageExtraction(t *testing.T) {
	vm := NewVM()

	err := fmt.Errorf("ValueError: invalid literal for int() with base 10: 'abc'")
	exc := vm.wrapGoError(err)

	if exc.Message != "invalid literal for int() with base 10: 'abc'" {
		t.Errorf("Message = %q, want extracted message without prefix", exc.Message)
	}

	// Args should contain the extracted message
	if exc.Args == nil || len(exc.Args.Items) != 1 {
		t.Fatal("expected 1 arg")
	}
	argStr, ok := exc.Args.Items[0].(*PyString)
	if !ok {
		t.Fatalf("expected *PyString arg, got %T", exc.Args.Items[0])
	}
	if argStr.Value != exc.Message {
		t.Errorf("args[0] = %q, want %q", argStr.Value, exc.Message)
	}
}
