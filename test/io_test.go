package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ATSOTECK/rage/internal/compiler"
	"github.com/ATSOTECK/rage/internal/runtime"
	"github.com/ATSOTECK/rage/internal/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =====================================
// File I/O Tests
// =====================================

func setupIOTest(t *testing.T) *runtime.VM {
	t.Helper()
	vm := newStdlibVM(t)
	stdlib.InitIOModule() // Explicitly enable IO module
	return vm
}

func TestFileOpenAndRead(t *testing.T) {
	vm := setupIOTest(t)

	// Create a temp file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("Hello, World!"), 0644)
	require.NoError(t, err)

	source := `
f = open("` + testFile + `", "r")
content = f.read()
f.close()
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err = vm.Execute(code)
	require.NoError(t, err)

	content := vm.GetGlobal("content").(*runtime.PyString).Value
	assert.Equal(t, "Hello, World!", content)
}

func TestFileWrite(t *testing.T) {
	vm := setupIOTest(t)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "output.txt")

	source := `
f = open("` + testFile + `", "w")
n = f.write("Test output\n")
f.close()
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	// Verify file content
	content, err := os.ReadFile(testFile)
	require.NoError(t, err)
	assert.Equal(t, "Test output\n", string(content))

	// Check bytes written
	n := vm.GetGlobal("n").(*runtime.PyInt).Value
	assert.Equal(t, int64(12), n)
}

func TestFileContextManager(t *testing.T) {
	vm := setupIOTest(t)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "context.txt")
	err := os.WriteFile(testFile, []byte("Line 1\nLine 2\nLine 3\n"), 0644)
	require.NoError(t, err)

	source := `
with open("` + testFile + `", "r") as f:
    content = f.read()
closed = f.closed
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err = vm.Execute(code)
	require.NoError(t, err)

	content := vm.GetGlobal("content").(*runtime.PyString).Value
	assert.Equal(t, "Line 1\nLine 2\nLine 3\n", content)

	// File should be closed after with block
	closed := vm.GetGlobal("closed").(*runtime.PyBool).Value
	assert.True(t, closed)
}

func TestFileReadline(t *testing.T) {
	vm := setupIOTest(t)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "lines.txt")
	err := os.WriteFile(testFile, []byte("Line 1\nLine 2\nLine 3\n"), 0644)
	require.NoError(t, err)

	source := `
f = open("` + testFile + `", "r")
line1 = f.readline()
line2 = f.readline()
f.close()
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err = vm.Execute(code)
	require.NoError(t, err)

	line1 := vm.GetGlobal("line1").(*runtime.PyString).Value
	line2 := vm.GetGlobal("line2").(*runtime.PyString).Value
	assert.Equal(t, "Line 1\n", line1)
	assert.Equal(t, "Line 2\n", line2)
}

func TestFileReadlines(t *testing.T) {
	vm := setupIOTest(t)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "lines.txt")
	err := os.WriteFile(testFile, []byte("Line 1\nLine 2\nLine 3\n"), 0644)
	require.NoError(t, err)

	source := `
f = open("` + testFile + `", "r")
lines = f.readlines()
num_lines = len(lines)
f.close()
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err = vm.Execute(code)
	require.NoError(t, err)

	numLines := vm.GetGlobal("num_lines").(*runtime.PyInt).Value
	assert.Equal(t, int64(3), numLines)
}

func TestFileSeekTell(t *testing.T) {
	vm := setupIOTest(t)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "seek.txt")
	err := os.WriteFile(testFile, []byte("0123456789"), 0644)
	require.NoError(t, err)

	source := `
f = open("` + testFile + `", "r")
f.seek(5)
pos = f.tell()
rest = f.read()
f.close()
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err = vm.Execute(code)
	require.NoError(t, err)

	pos := vm.GetGlobal("pos").(*runtime.PyInt).Value
	assert.Equal(t, int64(5), pos)

	rest := vm.GetGlobal("rest").(*runtime.PyString).Value
	assert.Equal(t, "56789", rest)
}

func TestFileAppendMode(t *testing.T) {
	vm := setupIOTest(t)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "append.txt")
	err := os.WriteFile(testFile, []byte("Original\n"), 0644)
	require.NoError(t, err)

	source := `
f = open("` + testFile + `", "a")
f.write("Appended\n")
f.close()
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err = vm.Execute(code)
	require.NoError(t, err)

	content, err := os.ReadFile(testFile)
	require.NoError(t, err)
	assert.Equal(t, "Original\nAppended\n", string(content))
}

func TestFileBinaryMode(t *testing.T) {
	vm := setupIOTest(t)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "binary.bin")

	// Write binary data directly from Go to test reading
	err := os.WriteFile(testFile, []byte{0x00, 0x01, 0x02, 0x03}, 0644)
	require.NoError(t, err)

	source := `
f = open("` + testFile + `", "rb")
data = f.read()
data_len = len(data)
f.close()
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err = vm.Execute(code)
	require.NoError(t, err)

	dataLen := vm.GetGlobal("data_len").(*runtime.PyInt).Value
	assert.Equal(t, int64(4), dataLen)
}

func TestFileNotFound(t *testing.T) {
	vm := setupIOTest(t)

	source := `f = open("/nonexistent/path/file.txt", "r")`

	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "FileNotFoundError")
}

func TestFileClosedError(t *testing.T) {
	vm := setupIOTest(t)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "closed.txt")
	err := os.WriteFile(testFile, []byte("test"), 0644)
	require.NoError(t, err)

	source := `
f = open("` + testFile + `", "r")
f.close()
f.read()
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err = vm.Execute(code)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "I/O operation on closed file")
}

func TestFileWritelines(t *testing.T) {
	vm := setupIOTest(t)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "writelines.txt")

	source := `
lines = ["Line 1\n", "Line 2\n", "Line 3\n"]
f = open("` + testFile + `", "w")
f.writelines(lines)
f.close()
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	content, err := os.ReadFile(testFile)
	require.NoError(t, err)
	assert.Equal(t, "Line 1\nLine 2\nLine 3\n", string(content))
}

func TestFileReadPartial(t *testing.T) {
	vm := setupIOTest(t)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "partial.txt")
	err := os.WriteFile(testFile, []byte("Hello, World!"), 0644)
	require.NoError(t, err)

	source := `
f = open("` + testFile + `", "r")
part1 = f.read(5)
part2 = f.read(2)
f.close()
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err = vm.Execute(code)
	require.NoError(t, err)

	part1 := vm.GetGlobal("part1").(*runtime.PyString).Value
	part2 := vm.GetGlobal("part2").(*runtime.PyString).Value
	assert.Equal(t, "Hello", part1)
	assert.Equal(t, ", ", part2)
}
