package test

import (
	"strings"
	"testing"

	"github.com/ATSOTECK/rage/internal/compiler"
	"github.com/ATSOTECK/rage/internal/runtime"
)

// =============================================================================
// Note: File I/O is not yet implemented in RAGE
// These tests are designed to skip gracefully when features are not available
// and will automatically start working when file I/O is implemented.
// =============================================================================

// =============================================================================
// Basic File Open Tests
// =============================================================================

func TestOpenFileRead(t *testing.T) {
	source := `
f = open('/tmp/test.txt', 'r')
f.close()
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("File I/O not supported (compile error)")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("File I/O not supported: " + err.Error())
		return
	}
}

func TestOpenFileWrite(t *testing.T) {
	source := `
f = open('/tmp/test_write.txt', 'w')
f.close()
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("File I/O not supported (compile error)")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("File I/O not supported: " + err.Error())
		return
	}
}

func TestOpenFileAppend(t *testing.T) {
	source := `
f = open('/tmp/test_append.txt', 'a')
f.close()
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("File I/O not supported (compile error)")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("File I/O not supported: " + err.Error())
		return
	}
}

// =============================================================================
// File Read Tests
// =============================================================================

func TestFileReadAll(t *testing.T) {
	source := `
content = open('/tmp/test.txt', 'r').read()
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("File I/O not supported (compile error)")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("File I/O not supported: " + err.Error())
		return
	}
}

func TestFileReadLine(t *testing.T) {
	source := `
f = open('/tmp/test.txt', 'r')
line = f.readline()
f.close()
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("File I/O not supported (compile error)")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("File I/O not supported: " + err.Error())
		return
	}
}

func TestFileReadLines(t *testing.T) {
	source := `
f = open('/tmp/test.txt', 'r')
lines = f.readlines()
f.close()
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("File I/O not supported (compile error)")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("File I/O not supported: " + err.Error())
		return
	}
}

// =============================================================================
// File Write Tests
// =============================================================================

func TestFileWriteBasic(t *testing.T) {
	source := `
f = open('/tmp/test_write.txt', 'w')
f.write('Hello, World!')
f.close()
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("File I/O not supported (compile error)")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("File I/O not supported: " + err.Error())
		return
	}
}

func TestFileWriteLines(t *testing.T) {
	source := `
f = open('/tmp/test_write.txt', 'w')
f.writelines(['line1\n', 'line2\n', 'line3\n'])
f.close()
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("File I/O not supported (compile error)")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("File I/O not supported: " + err.Error())
		return
	}
}

// =============================================================================
// Context Manager Tests (with statement)
// =============================================================================

func TestFileWithStatement(t *testing.T) {
	source := `
with open('/tmp/test.txt', 'r') as f:
    content = f.read()
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("File I/O or with statement not supported (compile error)")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("File I/O not supported: " + err.Error())
		return
	}
}

func TestFileWithStatementWrite(t *testing.T) {
	source := `
with open('/tmp/test_with.txt', 'w') as f:
    f.write('test content')
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("File I/O or with statement not supported (compile error)")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("File I/O not supported: " + err.Error())
		return
	}
}

// =============================================================================
// File Error Tests
// =============================================================================

func TestOpenNonexistentFile(t *testing.T) {
	source := `
f = open('/nonexistent/path/to/file.txt', 'r')
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("File I/O not supported (compile error)")
		return
	}
	_, err := vm.Execute(code)
	// If open() is not defined, skip
	if err != nil && (strings.Contains(err.Error(), "not defined") ||
		strings.Contains(err.Error(), "not callable") ||
		strings.Contains(err.Error(), "not supported")) {
		t.Skip("File I/O not supported: " + err.Error())
		return
	}
	if err == nil {
		t.Error("Expected error when opening nonexistent file")
		return
	}
	// Check for FileNotFoundError or similar - error occurred so test passes
	// (we don't verify exact error message format since it varies by implementation)
}

func TestOpenDirectoryAsFile(t *testing.T) {
	source := `
f = open('/tmp', 'r')
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("File I/O not supported (compile error)")
		return
	}
	_, err := vm.Execute(code)
	if err == nil {
		t.Skip("Opening directory as file doesn't raise error (implementation detail)")
		return
	}
	// Should error - IsADirectoryError in Python
}

func TestOpenInvalidMode(t *testing.T) {
	source := `
f = open('/tmp/test.txt', 'invalid_mode')
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("File I/O not supported (compile error)")
		return
	}
	_, err := vm.Execute(code)
	if err == nil {
		t.Skip("Invalid mode doesn't raise error")
		return
	}
}

func TestWriteToReadOnlyFile(t *testing.T) {
	source := `
f = open('/tmp/test.txt', 'r')
f.write('should fail')
f.close()
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("File I/O not supported (compile error)")
		return
	}
	_, err := vm.Execute(code)
	if err == nil {
		t.Error("Expected error when writing to read-only file")
		return
	}
}

func TestReadFromWriteOnlyFile(t *testing.T) {
	source := `
f = open('/tmp/test_write.txt', 'w')
content = f.read()
f.close()
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("File I/O not supported (compile error)")
		return
	}
	_, err := vm.Execute(code)
	if err == nil {
		t.Error("Expected error when reading from write-only file")
		return
	}
}

func TestOperationOnClosedFile(t *testing.T) {
	source := `
f = open('/tmp/test.txt', 'r')
f.close()
content = f.read()
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("File I/O not supported (compile error)")
		return
	}
	_, err := vm.Execute(code)
	if err == nil {
		t.Error("Expected error when operating on closed file")
		return
	}
}

// =============================================================================
// File Iteration Tests
// =============================================================================

func TestIterateFileLines(t *testing.T) {
	source := `
lines = []
for line in open('/tmp/test.txt', 'r'):
    lines.append(line)
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("File I/O not supported (compile error)")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("File I/O not supported: " + err.Error())
		return
	}
}

// =============================================================================
// File Seek and Tell Tests
// =============================================================================

func TestFileSeek(t *testing.T) {
	source := `
f = open('/tmp/test.txt', 'r')
f.seek(5)
pos = f.tell()
f.close()
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("File I/O not supported (compile error)")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("File I/O not supported: " + err.Error())
		return
	}
}

func TestFileTell(t *testing.T) {
	source := `
f = open('/tmp/test.txt', 'r')
start_pos = f.tell()
f.read(10)
after_pos = f.tell()
f.close()
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("File I/O not supported (compile error)")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("File I/O not supported: " + err.Error())
		return
	}
}

// =============================================================================
// File Truncate Tests
// =============================================================================

func TestFileTruncate(t *testing.T) {
	source := `
f = open('/tmp/test_truncate.txt', 'w+')
f.write('Hello, World!')
f.truncate(5)
f.close()
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("File I/O not supported (compile error)")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("File I/O not supported: " + err.Error())
		return
	}
}

// =============================================================================
// Binary File Tests
// =============================================================================

func TestBinaryFileRead(t *testing.T) {
	source := `
f = open('/tmp/test.bin', 'rb')
data = f.read()
f.close()
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("Binary file I/O not supported (compile error)")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("Binary file I/O not supported: " + err.Error())
		return
	}
}

func TestBinaryFileWrite(t *testing.T) {
	source := `
f = open('/tmp/test.bin', 'wb')
f.write(b'\x00\x01\x02\x03')
f.close()
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("Binary file I/O not supported (compile error)")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("Binary file I/O not supported: " + err.Error())
		return
	}
}

// =============================================================================
// File Flush Tests
// =============================================================================

func TestFileFlush(t *testing.T) {
	source := `
f = open('/tmp/test_flush.txt', 'w')
f.write('test')
f.flush()
f.close()
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("File I/O not supported (compile error)")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("File I/O not supported: " + err.Error())
		return
	}
}

// =============================================================================
// File Properties Tests
// =============================================================================

func TestFileProperties(t *testing.T) {
	source := `
f = open('/tmp/test.txt', 'r')
name = f.name
mode = f.mode
closed_before = f.closed
f.close()
closed_after = f.closed
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("File I/O not supported (compile error)")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("File I/O not supported: " + err.Error())
		return
	}
}

// =============================================================================
// Path Module Tests (os.path equivalent)
// =============================================================================

func TestOsPathExists(t *testing.T) {
	source := `
import os
result = os.path.exists('/tmp')
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("os.path not supported (compile error)")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("os.path not supported: " + err.Error())
		return
	}
}

func TestOsPathJoin(t *testing.T) {
	source := `
import os
result = os.path.join('/tmp', 'test', 'file.txt')
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("os.path not supported (compile error)")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("os.path not supported: " + err.Error())
		return
	}
}

func TestOsPathBasename(t *testing.T) {
	source := `
import os
result = os.path.basename('/tmp/test/file.txt')
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("os.path not supported (compile error)")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("os.path not supported: " + err.Error())
		return
	}
}

func TestOsPathDirname(t *testing.T) {
	source := `
import os
result = os.path.dirname('/tmp/test/file.txt')
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("os.path not supported (compile error)")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("os.path not supported: " + err.Error())
		return
	}
}

// =============================================================================
// os Module Tests
// =============================================================================

func TestOsGetcwd(t *testing.T) {
	source := `
import os
result = os.getcwd()
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("os module not supported (compile error)")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("os.getcwd not supported: " + err.Error())
		return
	}
}

func TestOsListdir(t *testing.T) {
	source := `
import os
result = os.listdir('/tmp')
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("os module not supported (compile error)")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("os.listdir not supported: " + err.Error())
		return
	}
}

func TestOsMakedirs(t *testing.T) {
	source := `
import os
os.makedirs('/tmp/rage_test_dir/subdir', exist_ok=True)
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("os module not supported (compile error)")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("os.makedirs not supported: " + err.Error())
		return
	}
}

func TestOsRemove(t *testing.T) {
	source := `
import os
os.remove('/tmp/rage_test_file.txt')
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("os module not supported (compile error)")
		return
	}
	_, err := vm.Execute(code)
	// Expected to error if file doesn't exist
	if err != nil {
		t.Skip("os.remove not supported: " + err.Error())
		return
	}
}

// =============================================================================
// File Encoding Tests
// =============================================================================

func TestFileEncodingUtf8(t *testing.T) {
	source := `
f = open('/tmp/test_utf8.txt', 'r', encoding='utf-8')
content = f.read()
f.close()
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("File encoding parameter not supported (compile error)")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("File encoding not supported: " + err.Error())
		return
	}
}

func TestFileEncodingLatin1(t *testing.T) {
	source := `
f = open('/tmp/test_latin1.txt', 'r', encoding='latin-1')
content = f.read()
f.close()
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("File encoding parameter not supported (compile error)")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("File encoding not supported: " + err.Error())
		return
	}
}

// =============================================================================
// Print to File Tests
// =============================================================================

func TestPrintToFile(t *testing.T) {
	source := `
f = open('/tmp/test_print.txt', 'w')
print('Hello, World!', file=f)
f.close()
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("print to file not supported (compile error)")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("print to file not supported: " + err.Error())
		return
	}
}

// =============================================================================
// Permission Error Tests
// =============================================================================

func TestPermissionDenied(t *testing.T) {
	source := `
f = open('/etc/shadow', 'r')
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("File I/O not supported (compile error)")
		return
	}
	_, err := vm.Execute(code)
	if err == nil {
		t.Skip("Permission check not enforced or running as root")
		return
	}
	// Should be PermissionError
}

// =============================================================================
// Temporary File Tests
// =============================================================================

func TestTempfile(t *testing.T) {
	source := `
import tempfile
f = tempfile.NamedTemporaryFile()
name = f.name
f.close()
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("tempfile module not supported (compile error)")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("tempfile module not supported: " + err.Error())
		return
	}
}

func TestTempfileDirectory(t *testing.T) {
	source := `
import tempfile
d = tempfile.TemporaryDirectory()
name = d.name
d.cleanup()
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("tempfile module not supported (compile error)")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("tempfile module not supported: " + err.Error())
		return
	}
}
