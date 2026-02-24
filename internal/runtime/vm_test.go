package runtime

import (
	"testing"
)

// =====================================
// NewVM: Initialization
// =====================================

func TestNewVMCreatesValidVM(t *testing.T) {
	vm := NewVM()
	if vm == nil {
		t.Fatal("NewVM() returned nil")
	}
	if vm.Globals == nil {
		t.Error("Globals map should be initialized")
	}
	if vm.builtins == nil {
		t.Error("builtins map should be initialized")
	}
}

func TestNewVMHasBuiltinsPopulated(t *testing.T) {
	vm := NewVM()

	// Check that fundamental exception classes exist
	requiredClasses := []string{
		"BaseException", "Exception", "TypeError", "ValueError",
		"KeyError", "IndexError", "AttributeError", "RuntimeError",
		"StopIteration", "NameError", "ZeroDivisionError",
		"OverflowError", "ImportError", "OSError",
	}
	for _, name := range requiredClasses {
		v, ok := vm.builtins[name]
		if !ok {
			t.Errorf("builtin %q not found", name)
			continue
		}
		cls, ok := v.(*PyClass)
		if !ok {
			t.Errorf("builtin %q is %T, expected *PyClass", name, v)
			continue
		}
		if cls.Name != name {
			t.Errorf("builtin %q has Name=%q", name, cls.Name)
		}
	}
}

func TestNewVMCheckInterval(t *testing.T) {
	vm := NewVM()
	if vm.checkInterval != 1000 {
		t.Errorf("default checkInterval = %d, want 1000", vm.checkInterval)
	}
	if vm.checkCounter != 1000 {
		t.Errorf("default checkCounter = %d, want 1000", vm.checkCounter)
	}
}

func TestSetCheckInterval(t *testing.T) {
	vm := NewVM()
	vm.SetCheckInterval(500)
	if vm.checkInterval != 500 {
		t.Errorf("checkInterval = %d, want 500", vm.checkInterval)
	}
	if vm.checkCounter != 500 {
		t.Errorf("checkCounter = %d, want 500", vm.checkCounter)
	}

	// Minimum clamp to 1
	vm.SetCheckInterval(0)
	if vm.checkInterval != 1 {
		t.Errorf("checkInterval = %d, want 1 (clamped)", vm.checkInterval)
	}
	vm.SetCheckInterval(-10)
	if vm.checkInterval != 1 {
		t.Errorf("checkInterval = %d, want 1 (clamped from negative)", vm.checkInterval)
	}
}

// =====================================
// Stack Operations: push / pop / top / peek
// =====================================

// newVMWithFrame creates a VM with a frame suitable for stack operations.
func newVMWithFrame(stackSize int) *VM {
	vm := NewVM()
	frame := &Frame{
		Code: &CodeObject{
			Name:      "<test>",
			Filename:  "<test>",
			StackSize: stackSize,
		},
		Stack: make([]Value, stackSize),
		SP:    0,
	}
	vm.frames = []*Frame{frame}
	vm.frame = frame
	return vm
}

func TestPushAndPop(t *testing.T) {
	vm := newVMWithFrame(16)

	vm.push(MakeInt(42))
	vm.push(MakeInt(99))

	if vm.frame.SP != 2 {
		t.Fatalf("SP = %d, want 2", vm.frame.SP)
	}

	val := vm.pop()
	if v, ok := val.(*PyInt); !ok || v.Value != 99 {
		t.Errorf("first pop = %v, want 99", val)
	}

	val = vm.pop()
	if v, ok := val.(*PyInt); !ok || v.Value != 42 {
		t.Errorf("second pop = %v, want 42", val)
	}

	if vm.frame.SP != 0 {
		t.Errorf("SP after pops = %d, want 0", vm.frame.SP)
	}
}

func TestTop(t *testing.T) {
	vm := newVMWithFrame(16)

	vm.push(MakeInt(10))
	vm.push(MakeInt(20))

	val := vm.top()
	if v, ok := val.(*PyInt); !ok || v.Value != 20 {
		t.Errorf("top() = %v, want 20", val)
	}
	// top() should not remove the value
	if vm.frame.SP != 2 {
		t.Errorf("SP after top() = %d, want 2", vm.frame.SP)
	}
}

func TestPeek(t *testing.T) {
	vm := newVMWithFrame(16)

	vm.push(MakeInt(10)) // index 0 from bottom
	vm.push(MakeInt(20)) // index 1 from bottom
	vm.push(MakeInt(30)) // index 2 from bottom (top)

	// peek(0) should be top
	val := vm.peek(0)
	if v, ok := val.(*PyInt); !ok || v.Value != 30 {
		t.Errorf("peek(0) = %v, want 30", val)
	}

	// peek(1) should be second from top
	val = vm.peek(1)
	if v, ok := val.(*PyInt); !ok || v.Value != 20 {
		t.Errorf("peek(1) = %v, want 20", val)
	}

	// peek(2) should be bottom
	val = vm.peek(2)
	if v, ok := val.(*PyInt); !ok || v.Value != 10 {
		t.Errorf("peek(2) = %v, want 10", val)
	}

	// SP unchanged
	if vm.frame.SP != 3 {
		t.Errorf("SP after peeks = %d, want 3", vm.frame.SP)
	}
}

func TestPopEmptyStackPanics(t *testing.T) {
	vm := newVMWithFrame(16)

	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic on empty pop, got none")
		}
		msg, ok := r.(string)
		if !ok {
			t.Errorf("panic value is %T, expected string", r)
			return
		}
		if msg != "stack underflow: cannot pop from empty stack" {
			t.Errorf("unexpected panic message: %q", msg)
		}
	}()

	vm.pop()
}

func TestTopEmptyStackPanics(t *testing.T) {
	vm := newVMWithFrame(16)

	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic on empty top, got none")
		}
	}()

	vm.top()
}

func TestPeekInvalidIndexPanics(t *testing.T) {
	vm := newVMWithFrame(16)
	vm.push(MakeInt(1))

	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic on invalid peek, got none")
		}
	}()

	vm.peek(5) // way beyond stack size
}

// =====================================
// Stack growth (push beyond initial capacity)
// =====================================

func TestPushGrowsStack(t *testing.T) {
	vm := newVMWithFrame(4) // Start with small stack

	// Push more than 4 items to force growth
	for i := 0; i < 10; i++ {
		vm.push(MakeInt(int64(i)))
	}

	if vm.frame.SP != 10 {
		t.Errorf("SP = %d, want 10", vm.frame.SP)
	}

	// Verify values are intact after growth
	for i := 9; i >= 0; i-- {
		val := vm.pop()
		v, ok := val.(*PyInt)
		if !ok {
			t.Fatalf("expected *PyInt at position %d, got %T", i, val)
		}
		if v.Value != int64(i) {
			t.Errorf("pop[%d] = %d, want %d", i, v.Value, i)
		}
	}
}

// =====================================
// ensureStack
// =====================================

func TestEnsureStack(t *testing.T) {
	vm := newVMWithFrame(4)

	// Push 2 items
	vm.push(MakeInt(1))
	vm.push(MakeInt(2))

	// Ensure we have room for 10 more
	vm.ensureStack(10)

	if len(vm.frame.Stack) < 12 {
		t.Errorf("stack length = %d, expected at least 12", len(vm.frame.Stack))
	}

	// Existing values should still be there
	if vm.frame.SP != 2 {
		t.Errorf("SP = %d, want 2", vm.frame.SP)
	}
	val := vm.peek(0)
	if v, ok := val.(*PyInt); !ok || v.Value != 2 {
		t.Errorf("peek(0) after ensureStack = %v, want 2", val)
	}
}

func TestEnsureStackNoOpWhenSufficient(t *testing.T) {
	vm := newVMWithFrame(100)
	origLen := len(vm.frame.Stack)

	vm.push(MakeInt(1))
	vm.ensureStack(5) // 1 + 5 = 6, well within 100

	if len(vm.frame.Stack) != origLen {
		t.Errorf("stack should not have been reallocated, was %d now %d", origLen, len(vm.frame.Stack))
	}
}

// =====================================
// pendingMemError
// =====================================

func TestPendingMemErrorOnStackGrowth(t *testing.T) {
	vm := newVMWithFrame(2)
	vm.maxMemoryBytes = 1    // Extremely low limit
	vm.allocatedBytes = 0

	// Push enough to trigger stack growth
	vm.push(MakeInt(1))
	vm.push(MakeInt(2))
	vm.push(MakeInt(3)) // This should trigger growth + pendingMemError

	if !vm.pendingMemError {
		t.Error("expected pendingMemError to be true after stack growth with low memory limit")
	}
}

func TestNoPendingMemErrorWithoutLimit(t *testing.T) {
	vm := newVMWithFrame(2)
	// maxMemoryBytes = 0 means unlimited

	vm.push(MakeInt(1))
	vm.push(MakeInt(2))
	vm.push(MakeInt(3)) // triggers growth but no limit

	if vm.pendingMemError {
		t.Error("pendingMemError should be false with unlimited memory")
	}
}

// =====================================
// builtinClass helper
// =====================================

func TestBuiltinClass(t *testing.T) {
	vm := NewVM()

	tests := []struct {
		name string
		want bool
	}{
		{"TypeError", true},
		{"ValueError", true},
		{"KeyError", true},
		{"IndexError", true},
		{"RuntimeError", true},
		{"BaseException", true},
		{"Exception", true},
		{"NonExistentError", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cls := vm.builtinClass(tt.name)
			if tt.want && cls == nil {
				t.Errorf("builtinClass(%q) = nil, want *PyClass", tt.name)
			}
			if !tt.want && cls != nil {
				t.Errorf("builtinClass(%q) = %v, want nil", tt.name, cls)
			}
			if tt.want && cls != nil && cls.Name != tt.name {
				t.Errorf("builtinClass(%q).Name = %q", tt.name, cls.Name)
			}
		})
	}
}

func TestBuiltinClassNonClassValue(t *testing.T) {
	vm := NewVM()
	// Put a non-class value in builtins
	vm.builtins["not_a_class"] = MakeInt(42)

	cls := vm.builtinClass("not_a_class")
	if cls != nil {
		t.Error("builtinClass should return nil for non-*PyClass value")
	}
}

// =====================================
// Resource Limits
// =====================================

func TestSetMaxRecursionDepth(t *testing.T) {
	vm := NewVM()

	vm.SetMaxRecursionDepth(100)
	if vm.MaxRecursionDepth() != 100 {
		t.Errorf("MaxRecursionDepth = %d, want 100", vm.MaxRecursionDepth())
	}

	vm.SetMaxRecursionDepth(0)
	if vm.MaxRecursionDepth() != 0 {
		t.Errorf("MaxRecursionDepth = %d, want 0 (unlimited)", vm.MaxRecursionDepth())
	}

	// Negative values clamped to 0
	vm.SetMaxRecursionDepth(-5)
	if vm.MaxRecursionDepth() != 0 {
		t.Errorf("MaxRecursionDepth = %d, want 0 (clamped)", vm.MaxRecursionDepth())
	}
}

func TestSetMaxMemoryBytes(t *testing.T) {
	vm := NewVM()

	vm.SetMaxMemoryBytes(1024 * 1024)
	if vm.MaxMemoryBytes() != 1024*1024 {
		t.Errorf("MaxMemoryBytes = %d, want %d", vm.MaxMemoryBytes(), 1024*1024)
	}

	vm.SetMaxMemoryBytes(0)
	if vm.MaxMemoryBytes() != 0 {
		t.Errorf("MaxMemoryBytes = %d, want 0 (unlimited)", vm.MaxMemoryBytes())
	}

	vm.SetMaxMemoryBytes(-1)
	if vm.MaxMemoryBytes() != 0 {
		t.Errorf("MaxMemoryBytes = %d, want 0 (clamped)", vm.MaxMemoryBytes())
	}
}

func TestSetMaxCollectionSize(t *testing.T) {
	vm := NewVM()

	vm.SetMaxCollectionSize(10000)
	if vm.MaxCollectionSize() != 10000 {
		t.Errorf("MaxCollectionSize = %d, want 10000", vm.MaxCollectionSize())
	}

	vm.SetMaxCollectionSize(0)
	if vm.MaxCollectionSize() != 0 {
		t.Errorf("MaxCollectionSize = %d, want 0", vm.MaxCollectionSize())
	}
}

// =====================================
// trackAlloc / TrackAlloc
// =====================================

func TestTrackAllocWithinLimit(t *testing.T) {
	vm := NewVM()
	vm.SetMaxMemoryBytes(1000)

	err := vm.TrackAlloc(500)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if vm.AllocatedBytes() != 500 {
		t.Errorf("AllocatedBytes = %d, want 500", vm.AllocatedBytes())
	}
}

func TestTrackAllocExceedsLimit(t *testing.T) {
	vm := NewVM()
	vm.SetMaxMemoryBytes(100)

	err := vm.TrackAlloc(200)
	if err == nil {
		t.Error("expected MemoryError, got nil")
	}
}

func TestTrackAllocUnlimited(t *testing.T) {
	vm := NewVM()
	// maxMemoryBytes = 0 (unlimited)

	err := vm.TrackAlloc(999999999)
	if err != nil {
		t.Errorf("unexpected error with unlimited memory: %v", err)
	}
}

// =====================================
// checkCollectionSize
// =====================================

func TestCheckCollectionSizeWithinLimit(t *testing.T) {
	vm := NewVM()
	vm.SetMaxCollectionSize(100)

	err := vm.checkCollectionSize(50, "list")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCheckCollectionSizeExceedsLimit(t *testing.T) {
	vm := NewVM()
	vm.SetMaxCollectionSize(100)

	err := vm.checkCollectionSize(100, "list")
	if err == nil {
		t.Error("expected MemoryError, got nil")
	}
}

func TestCheckCollectionSizeUnlimited(t *testing.T) {
	vm := NewVM()
	// maxCollectionSize = 0 (unlimited)

	err := vm.checkCollectionSize(999999999, "list")
	if err != nil {
		t.Errorf("unexpected error with unlimited collection size: %v", err)
	}
}

// =====================================
// Multiple pushes then pops maintain LIFO order
// =====================================

func TestStackLIFOOrder(t *testing.T) {
	vm := newVMWithFrame(64)

	values := []int64{10, 20, 30, 40, 50}
	for _, v := range values {
		vm.push(MakeInt(v))
	}

	// Pop in reverse order
	for i := len(values) - 1; i >= 0; i-- {
		val := vm.pop()
		v, ok := val.(*PyInt)
		if !ok {
			t.Fatalf("expected *PyInt, got %T", val)
		}
		if v.Value != values[i] {
			t.Errorf("pop order wrong: got %d, want %d", v.Value, values[i])
		}
	}
}

// =====================================
// Push various types
// =====================================

func TestPushVariousTypes(t *testing.T) {
	vm := newVMWithFrame(16)

	vm.push(None)
	vm.push(True)
	vm.push(&PyString{Value: "hello"})
	vm.push(&PyFloat{Value: 3.14})
	vm.push(MakeInt(42))

	// Pop and verify types
	if _, ok := vm.pop().(*PyInt); !ok {
		t.Error("expected *PyInt")
	}
	if _, ok := vm.pop().(*PyFloat); !ok {
		t.Error("expected *PyFloat")
	}
	if _, ok := vm.pop().(*PyString); !ok {
		t.Error("expected *PyString")
	}
	if _, ok := vm.pop().(*PyBool); !ok {
		t.Error("expected *PyBool")
	}
	if _, ok := vm.pop().(*PyNone); !ok {
		t.Error("expected *PyNone")
	}
}

// =====================================
// Exception class MRO in builtins
// =====================================

func TestExceptionClassMRO(t *testing.T) {
	vm := NewVM()

	// ValueError should have Exception and BaseException in MRO
	ve := vm.builtinClass("ValueError")
	if ve == nil {
		t.Fatal("ValueError not found in builtins")
	}

	foundException := false
	foundBaseException := false
	for _, cls := range ve.Mro {
		if cls.Name == "Exception" {
			foundException = true
		}
		if cls.Name == "BaseException" {
			foundBaseException = true
		}
	}
	if !foundException {
		t.Error("ValueError MRO should include Exception")
	}
	if !foundBaseException {
		t.Error("ValueError MRO should include BaseException")
	}
}

func TestExceptionSubclassMRO(t *testing.T) {
	vm := NewVM()

	// KeyError -> LookupError -> Exception -> BaseException
	ke := vm.builtinClass("KeyError")
	if ke == nil {
		t.Fatal("KeyError not found")
	}

	expected := []string{"KeyError", "LookupError", "Exception", "BaseException"}
	if len(ke.Mro) != len(expected) {
		t.Fatalf("KeyError MRO length = %d, want %d", len(ke.Mro), len(expected))
	}
	for i, name := range expected {
		if ke.Mro[i].Name != name {
			t.Errorf("KeyError.Mro[%d].Name = %q, want %q", i, ke.Mro[i].Name, name)
		}
	}
}
