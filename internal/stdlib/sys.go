package stdlib

import (
	"math"
	"os"
	goruntime "runtime"
	"unsafe"

	gopherpy "github.com/ATSOTECK/oink/internal/runtime"
)

// Version information
const (
	versionMajor  = 3
	versionMinor  = 14
	versionMicro  = 0
	versionLevel  = "alpha" // alpha, beta, candidate, final
	versionSerial = 1
)

// sysArgv holds the command line arguments (can be set before running)
var sysArgv []string

// SetArgv sets the command line arguments for sys.argv
func SetArgv(args []string) {
	sysArgv = args
}

// InitSysModule registers the sys module
func InitSysModule() {
	// Build argv list
	var argvItems []gopherpy.Value
	if len(sysArgv) > 0 {
		for _, arg := range sysArgv {
			argvItems = append(argvItems, gopherpy.NewString(arg))
		}
	} else {
		// Use os.Args by default
		for _, arg := range os.Args {
			argvItems = append(argvItems, gopherpy.NewString(arg))
		}
	}

	// Build version_info tuple
	versionInfo := gopherpy.NewTuple([]gopherpy.Value{
		gopherpy.NewInt(versionMajor),
		gopherpy.NewInt(versionMinor),
		gopherpy.NewInt(versionMicro),
		gopherpy.NewString(versionLevel),
		gopherpy.NewInt(versionSerial),
	})

	// Build platform string
	platform := getPlatform()

	// Build implementation info
	implementation := gopherpy.NewModuleBuilder("sys.implementation").
		Const("name", gopherpy.NewString("oink")).
		Const("version", versionInfo).
		Const("cache_tag", gopherpy.NewString("oink-314")).
		Build()

	gopherpy.NewModuleBuilder("sys").
		Doc("System-specific parameters and functions.").
		// Version info
		Const("version", gopherpy.NewString("3.14.0 (oink)")).
		Const("version_info", versionInfo).
		Const("hexversion", gopherpy.NewInt(0x030e00a1)). // 3.14.0a1
		Const("implementation", implementation).
		// Platform info
		Const("platform", gopherpy.NewString(platform)).
		Const("byteorder", gopherpy.NewString(getByteOrder())).
		// Size limits
		Const("maxsize", gopherpy.NewInt(math.MaxInt64)).
		Const("maxunicode", gopherpy.NewInt(0x10FFFF)).
		Const("float_info", buildFloatInfo()).
		Const("int_info", buildIntInfo()).
		// Paths
		Const("argv", gopherpy.NewList(argvItems)).
		Const("path", gopherpy.NewList([]gopherpy.Value{})).
		Const("modules", gopherpy.NewDict()).
		Const("executable", gopherpy.NewString(getExecutable())).
		Const("prefix", gopherpy.NewString("")).
		Const("exec_prefix", gopherpy.NewString("")).
		// Standard streams (placeholders - actual IO not implemented)
		Const("stdin", gopherpy.None).
		Const("stdout", gopherpy.None).
		Const("stderr", gopherpy.None).
		// Recursion
		Const("_recursion_limit", gopherpy.NewInt(1000)).
		// Functions
		Func("exit", sysExit).
		Func("getrecursionlimit", sysGetRecursionLimit).
		Func("setrecursionlimit", sysSetRecursionLimit).
		Func("getsizeof", sysGetSizeof).
		Func("getrefcount", sysGetRefcount).
		Func("intern", sysIntern).
		Register()
}

// getPlatform returns the platform identifier
func getPlatform() string {
	switch goruntime.GOOS {
	case "darwin":
		return "darwin"
	case "linux":
		return "linux"
	case "windows":
		return "win32"
	case "freebsd":
		return "freebsd"
	default:
		return goruntime.GOOS
	}
}

// getByteOrder returns the native byte order
func getByteOrder() string {
	// Check endianness
	var x uint32 = 0x01020304
	if *(*byte)(unsafe.Pointer(&x)) == 0x04 {
		return "little"
	}
	return "big"
}

// getExecutable returns the path to the executable
func getExecutable() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	return exe
}

// buildFloatInfo creates the sys.float_info struct
func buildFloatInfo() *gopherpy.PyDict {
	info := gopherpy.NewDict()
	info.Items[gopherpy.NewString("max")] = gopherpy.NewFloat(math.MaxFloat64)
	info.Items[gopherpy.NewString("min")] = gopherpy.NewFloat(math.SmallestNonzeroFloat64)
	info.Items[gopherpy.NewString("epsilon")] = gopherpy.NewFloat(math.Nextafter(1.0, 2.0) - 1.0)
	info.Items[gopherpy.NewString("dig")] = gopherpy.NewInt(15)
	info.Items[gopherpy.NewString("mant_dig")] = gopherpy.NewInt(53)
	info.Items[gopherpy.NewString("max_exp")] = gopherpy.NewInt(1024)
	info.Items[gopherpy.NewString("min_exp")] = gopherpy.NewInt(-1021)
	info.Items[gopherpy.NewString("radix")] = gopherpy.NewInt(2)
	return info
}

// buildIntInfo creates the sys.int_info struct
func buildIntInfo() *gopherpy.PyDict {
	info := gopherpy.NewDict()
	info.Items[gopherpy.NewString("bits_per_digit")] = gopherpy.NewInt(64)
	info.Items[gopherpy.NewString("sizeof_digit")] = gopherpy.NewInt(8)
	return info
}

// Recursion limit (can be modified)
var recursionLimit int64 = 1000

// sys.exit([code])
func sysExit(vm *gopherpy.VM) int {
	code := 0
	if vm.GetTop() >= 1 {
		arg := vm.Get(1)
		switch v := arg.(type) {
		case *gopherpy.PyInt:
			code = int(v.Value)
		case *gopherpy.PyString:
			// Print message and exit with code 1
			println(v.Value)
			code = 1
		case *gopherpy.PyNone:
			code = 0
		default:
			code = 1
		}
	}
	os.Exit(code)
	return 0 // Never reached
}

// sys.getrecursionlimit()
func sysGetRecursionLimit(vm *gopherpy.VM) int {
	vm.Push(gopherpy.NewInt(recursionLimit))
	return 1
}

// sys.setrecursionlimit(limit)
func sysSetRecursionLimit(vm *gopherpy.VM) int {
	limit := vm.CheckInt(1)
	if limit <= 0 {
		vm.RaiseError("recursion limit must be positive")
		return 0
	}
	recursionLimit = limit
	return 0
}

// sys.getsizeof(obj) - returns approximate size of object
func sysGetSizeof(vm *gopherpy.VM) int {
	obj := vm.Get(1)

	// Return approximate sizes based on type
	var size int64
	switch v := obj.(type) {
	case *gopherpy.PyNone:
		size = 16
	case *gopherpy.PyBool:
		size = 24
	case *gopherpy.PyInt:
		size = 28
	case *gopherpy.PyFloat:
		size = 24
	case *gopherpy.PyString:
		size = 49 + int64(len(v.Value))
	case *gopherpy.PyList:
		size = 56 + int64(len(v.Items)*8)
	case *gopherpy.PyTuple:
		size = 40 + int64(len(v.Items)*8)
	case *gopherpy.PyDict:
		size = 64 + int64(len(v.Items)*24)
	case *gopherpy.PySet:
		size = 64 + int64(len(v.Items)*16)
	default:
		size = 64
	}

	vm.Push(gopherpy.NewInt(size))
	return 1
}

// sys.getrefcount(obj) - always returns 2 (fake reference count)
func sysGetRefcount(vm *gopherpy.VM) int {
	// Go uses garbage collection, so we just return a placeholder
	vm.Push(gopherpy.NewInt(2))
	return 1
}

// sys.intern(string) - return interned string (no-op in our implementation)
func sysIntern(vm *gopherpy.VM) int {
	s := vm.CheckString(1)
	vm.Push(gopherpy.NewString(s))
	return 1
}
