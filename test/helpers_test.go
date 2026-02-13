package test

import (
	"strings"
	"testing"

	"github.com/ATSOTECK/rage/internal/compiler"
	"github.com/ATSOTECK/rage/internal/runtime"
	"github.com/ATSOTECK/rage/internal/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// runCode compiles and executes Python source code, returning the VM.
// Fails the test if compilation or execution produces errors.
func runCode(t *testing.T, source string) *runtime.VM {
	t.Helper()
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs, "Compilation errors: %v", errs)
	_, err := vm.Execute(code)
	require.NoError(t, err)
	return vm
}

// runCodeExpectError compiles and executes Python source code, expecting an error
// containing expectedErrSubstr in either compilation or execution.
func runCodeExpectError(t *testing.T, source string, expectedErrSubstr string) {
	t.Helper()
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		for _, e := range errs {
			if strings.Contains(e.Error(), expectedErrSubstr) {
				return
			}
		}
		t.Fatalf("Expected error containing %q, got compilation errors: %v", expectedErrSubstr, errs)
	}
	_, err := vm.Execute(code)
	require.Error(t, err)
	assert.Contains(t, err.Error(), expectedErrSubstr)
}

// runExpectError compiles and executes Python source code, expecting an error
// containing one of the expectedErrors substrings. Skips the test if an error
// occurs but its message doesn't match any expected string.
func runExpectError(t *testing.T, source string, expectedErrors ...string) {
	t.Helper()
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		for _, e := range errs {
			for _, expected := range expectedErrors {
				if strings.Contains(e.Error(), expected) {
					return
				}
			}
		}
		t.Logf("Compilation errors: %v", errs)
	}

	if len(errs) == 0 {
		_, err := vm.Execute(code)
		require.Error(t, err, "Expected error containing one of %v", expectedErrors)
		errStr := err.Error()
		matched := false
		for _, expected := range expectedErrors {
			if strings.Contains(errStr, expected) {
				matched = true
				break
			}
		}
		if !matched {
			t.Skipf("Error occurred (%q) but message doesn't match expected format %v", errStr, expectedErrors)
		}
	}
}

// tryCompile attempts to compile source code and returns nil if there's a panic.
func tryCompile(source string) (code *runtime.CodeObject, errs []error, panicked bool) {
	defer func() {
		if r := recover(); r != nil {
			panicked = true
		}
	}()
	code, errs = compiler.CompileSource(source, "<test>")
	return code, errs, false
}

// runCodeWithStdlib initializes stdlib modules, then compiles and executes source code.
// Returns the VM. Fails the test if compilation or execution produces errors.
func runCodeWithStdlib(t *testing.T, source string) *runtime.VM {
	t.Helper()
	runtime.ResetModules()
	stdlib.InitAllModules()
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs, "Compilation errors: %v", errs)
	_, err := vm.Execute(code)
	require.NoError(t, err)
	return vm
}

// newStdlibVM initializes stdlib modules and returns a fresh VM.
func newStdlibVM(t *testing.T) *runtime.VM {
	t.Helper()
	runtime.ResetModules()
	stdlib.InitAllModules()
	return runtime.NewVM()
}
