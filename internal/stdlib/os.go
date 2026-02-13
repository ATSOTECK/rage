package stdlib

import (
	"crypto/rand"
	"os"
	"os/user"
	"path/filepath"
	goruntime "runtime"
	"strings"

	gopherpy "github.com/ATSOTECK/rage/internal/runtime"
)

// InitOSModule registers the os module
func InitOSModule() {
	// Build environ dict
	environ := gopherpy.NewDict()
	for _, e := range os.Environ() {
		if idx := strings.Index(e, "="); idx != -1 {
			key := e[:idx]
			value := e[idx+1:]
			environ.Items[gopherpy.NewString(key)] = gopherpy.NewString(value)
		}
	}

	// Platform-specific values
	var name, sep, linesep, pathsep, devnull string
	var altsep gopherpy.Value

	if goruntime.GOOS == "windows" {
		name = "nt"
		sep = "\\"
		linesep = "\r\n"
		pathsep = ";"
		devnull = "nul"
		altsep = gopherpy.NewString("/")
	} else {
		name = "posix"
		sep = "/"
		linesep = "\n"
		pathsep = ":"
		devnull = "/dev/null"
		altsep = gopherpy.None
	}

	// Build os.path submodule
	pathModule := gopherpy.NewModuleBuilder("os.path").
		Doc("Common pathname manipulations.").
		Const("sep", gopherpy.NewString(sep)).
		Const("extsep", gopherpy.NewString(".")).
		Const("altsep", altsep).
		Const("pathsep", gopherpy.NewString(pathsep)).
		Const("curdir", gopherpy.NewString(".")).
		Const("pardir", gopherpy.NewString("..")).
		Const("devnull", gopherpy.NewString(devnull)).
		Func("join", osPathJoin).
		Func("split", osPathSplit).
		Func("splitext", osPathSplitext).
		Func("basename", osPathBasename).
		Func("dirname", osPathDirname).
		Func("exists", osPathExists).
		Func("isfile", osPathIsfile).
		Func("isdir", osPathIsdir).
		Func("isabs", osPathIsabs).
		Func("islink", osPathIslink).
		Func("abspath", osPathAbspath).
		Func("normpath", osPathNormpath).
		Func("realpath", osPathRealpath).
		Func("expanduser", osPathExpanduser).
		Func("expandvars", osPathExpandvars).
		Func("getsize", osPathGetsize).
		Func("getatime", osPathGetatime).
		Func("getmtime", osPathGetmtime).
		Func("getctime", osPathGetctime).
		Func("samefile", osPathSamefile).
		Func("commonpath", osPathCommonpath).
		Func("commonprefix", osPathCommonprefix).
		Func("relpath", osPathRelpath).
		Build()

	gopherpy.NewModuleBuilder("os").
		Doc("Miscellaneous operating system interfaces.").
		// Constants
		Const("name", gopherpy.NewString(name)).
		Const("sep", gopherpy.NewString(sep)).
		Const("linesep", gopherpy.NewString(linesep)).
		Const("pathsep", gopherpy.NewString(pathsep)).
		Const("curdir", gopherpy.NewString(".")).
		Const("pardir", gopherpy.NewString("..")).
		Const("extsep", gopherpy.NewString(".")).
		Const("altsep", altsep).
		Const("devnull", gopherpy.NewString(devnull)).
		Const("environ", environ).
		// Submodule
		SubModule("path", pathModule).
		// Environment functions
		Func("getenv", osGetenv).
		Func("putenv", osPutenv).
		Func("unsetenv", osUnsetenv).
		Func("getenvb", osGetenvb).
		// Directory functions
		Func("getcwd", osGetcwd).
		Func("getcwdb", osGetcwdb).
		Func("chdir", osChdir).
		Func("listdir", osListdir).
		Func("scandir", osScandir).
		Func("mkdir", osMkdir).
		Func("makedirs", osMakedirs).
		Func("rmdir", osRmdir).
		Func("removedirs", osRemovedirs).
		// File functions
		Func("remove", osRemove).
		Func("unlink", osRemove). // Alias for remove
		Func("rename", osRename).
		Func("renames", osRenames).
		Func("replace", osReplace).
		Func("stat", osStat).
		Func("lstat", osLstat).
		Func("link", osLink).
		Func("symlink", osSymlink).
		Func("readlink", osReadlink).
		Func("chmod", osChmod).
		Func("access", osAccess).
		Func("truncate", osTruncate).
		// Process functions
		Func("getpid", osGetpid).
		Func("getppid", osGetppid).
		Func("getuid", osGetuid).
		Func("getgid", osGetgid).
		Func("geteuid", osGeteuid).
		Func("getegid", osGetegid).
		Func("getlogin", osGetlogin).
		Func("uname", osUname).
		Func("cpu_count", osCpuCount).
		Func("urandom", osUrandom).
		// Path functions (also available at os level)
		Func("fspath", osFspath).
		Func("fsencode", osFsencode).
		Func("fsdecode", osFsdecode).
		// System functions
		Func("system", osSystem).
		Func("abort", osAbort).
		Func("_exit", osExit).
		Register()
}

// ==================== Environment Functions ====================

// os.getenv(key, default=None)
func osGetenv(vm *gopherpy.VM) int {
	key := vm.CheckString(1)
	value, exists := os.LookupEnv(key)
	if !exists {
		if vm.GetTop() >= 2 {
			vm.Push(vm.Get(2))
		} else {
			vm.Push(gopherpy.None)
		}
		return 1
	}
	vm.Push(gopherpy.NewString(value))
	return 1
}

// os.getenvb(key, default=None) - returns bytes
func osGetenvb(vm *gopherpy.VM) int {
	key := vm.CheckString(1)
	value, exists := os.LookupEnv(key)
	if !exists {
		if vm.GetTop() >= 2 {
			vm.Push(vm.Get(2))
		} else {
			vm.Push(gopherpy.None)
		}
		return 1
	}
	vm.Push(gopherpy.NewBytes([]byte(value)))
	return 1
}

// os.putenv(key, value)
func osPutenv(vm *gopherpy.VM) int {
	key := vm.CheckString(1)
	value := vm.CheckString(2)
	os.Setenv(key, value)
	return 0
}

// os.unsetenv(key)
func osUnsetenv(vm *gopherpy.VM) int {
	key := vm.CheckString(1)
	os.Unsetenv(key)
	return 0
}

// ==================== Directory Functions ====================

// os.getcwd()
func osGetcwd(vm *gopherpy.VM) int {
	cwd, err := os.Getwd()
	if err != nil {
		vm.RaiseError("OSError: %s", err.Error())
		return 0
	}
	vm.Push(gopherpy.NewString(cwd))
	return 1
}

// os.getcwdb() - returns bytes
func osGetcwdb(vm *gopherpy.VM) int {
	cwd, err := os.Getwd()
	if err != nil {
		vm.RaiseError("OSError: %s", err.Error())
		return 0
	}
	vm.Push(gopherpy.NewBytes([]byte(cwd)))
	return 1
}

// os.chdir(path)
func osChdir(vm *gopherpy.VM) int {
	path := vm.CheckString(1)
	if err := os.Chdir(path); err != nil {
		vm.RaiseError("OSError: %s", err.Error())
		return 0
	}
	return 0
}

// os.listdir(path='.')
func osListdir(vm *gopherpy.VM) int {
	path := vm.OptionalString(1, ".")

	entries, err := os.ReadDir(path)
	if err != nil {
		vm.RaiseError("OSError: %s", err.Error())
		return 0
	}

	items := make([]gopherpy.Value, len(entries))
	for i, entry := range entries {
		items[i] = gopherpy.NewString(entry.Name())
	}
	vm.Push(gopherpy.NewList(items))
	return 1
}

// os.scandir(path='.') - returns list of DirEntry-like dicts
func osScandir(vm *gopherpy.VM) int {
	path := vm.OptionalString(1, ".")

	entries, err := os.ReadDir(path)
	if err != nil {
		vm.RaiseError("OSError: %s", err.Error())
		return 0
	}

	items := make([]gopherpy.Value, len(entries))
	for i, entry := range entries {
		info, _ := entry.Info()
		entryDict := gopherpy.NewDict()
		entryDict.Items[gopherpy.NewString("name")] = gopherpy.NewString(entry.Name())
		entryDict.Items[gopherpy.NewString("path")] = gopherpy.NewString(filepath.Join(path, entry.Name()))
		entryDict.Items[gopherpy.NewString("is_dir")] = gopherpy.NewBool(entry.IsDir())
		entryDict.Items[gopherpy.NewString("is_file")] = gopherpy.NewBool(entry.Type().IsRegular())
		entryDict.Items[gopherpy.NewString("is_symlink")] = gopherpy.NewBool(entry.Type()&os.ModeSymlink != 0)
		if info != nil {
			entryDict.Items[gopherpy.NewString("size")] = gopherpy.NewInt(info.Size())
		}
		items[i] = entryDict
	}
	vm.Push(gopherpy.NewList(items))
	return 1
}

// os.mkdir(path, mode=0o777)
func osMkdir(vm *gopherpy.VM) int {
	path := vm.CheckString(1)
	mode := os.FileMode(0777)
	if vm.GetTop() >= 2 {
		mode = os.FileMode(vm.CheckInt(2))
	}
	if err := os.Mkdir(path, mode); err != nil {
		vm.RaiseError("OSError: %s", err.Error())
		return 0
	}
	return 0
}

// os.makedirs(path, mode=0o777, exist_ok=False)
func osMakedirs(vm *gopherpy.VM) int {
	path := vm.CheckString(1)
	mode := os.FileMode(0777)
	existOk := false

	if vm.GetTop() >= 2 {
		mode = os.FileMode(vm.CheckInt(2))
	}
	if vm.GetTop() >= 3 {
		existOk = vm.ToBool(3)
	}

	err := os.MkdirAll(path, mode)
	if err != nil {
		if !existOk || !os.IsExist(err) {
			vm.RaiseError("OSError: %s", err.Error())
			return 0
		}
	}
	return 0
}

// os.rmdir(path)
func osRmdir(vm *gopherpy.VM) int {
	path := vm.CheckString(1)
	if err := os.Remove(path); err != nil {
		vm.RaiseError("OSError: %s", err.Error())
		return 0
	}
	return 0
}

// os.removedirs(path)
func osRemovedirs(vm *gopherpy.VM) int {
	path := vm.CheckString(1)

	// Remove leaf directory and all empty parent directories
	for {
		if err := os.Remove(path); err != nil {
			// Stop when we can't remove anymore (not empty or doesn't exist)
			break
		}
		parent := filepath.Dir(path)
		if parent == path || parent == "." || parent == "/" {
			break
		}
		path = parent
	}
	return 0
}

// ==================== File Functions ====================

// os.remove(path) / os.unlink(path)
func osRemove(vm *gopherpy.VM) int {
	path := vm.CheckString(1)
	if err := os.Remove(path); err != nil {
		vm.RaiseError("OSError: %s", err.Error())
		return 0
	}
	return 0
}

// os.rename(src, dst)
func osRename(vm *gopherpy.VM) int {
	src := vm.CheckString(1)
	dst := vm.CheckString(2)
	if err := os.Rename(src, dst); err != nil {
		vm.RaiseError("OSError: %s", err.Error())
		return 0
	}
	return 0
}

// os.renames(old, new) - recursive rename
func osRenames(vm *gopherpy.VM) int {
	oldPath := vm.CheckString(1)
	newPath := vm.CheckString(2)

	// Create parent directories for new path
	newDir := filepath.Dir(newPath)
	if err := os.MkdirAll(newDir, 0777); err != nil {
		vm.RaiseError("OSError: %s", err.Error())
		return 0
	}

	// Rename
	if err := os.Rename(oldPath, newPath); err != nil {
		vm.RaiseError("OSError: %s", err.Error())
		return 0
	}

	// Remove empty old directories
	oldDir := filepath.Dir(oldPath)
	for {
		if err := os.Remove(oldDir); err != nil {
			break
		}
		parent := filepath.Dir(oldDir)
		if parent == oldDir || parent == "." || parent == "/" {
			break
		}
		oldDir = parent
	}

	return 0
}

// os.replace(src, dst)
func osReplace(vm *gopherpy.VM) int {
	src := vm.CheckString(1)
	dst := vm.CheckString(2)
	// Remove destination first to allow cross-filesystem moves
	os.Remove(dst)
	if err := os.Rename(src, dst); err != nil {
		vm.RaiseError("OSError: %s", err.Error())
		return 0
	}
	return 0
}

// os.stat(path) - returns dict with file info
func osStat(vm *gopherpy.VM) int {
	path := vm.CheckString(1)
	info, err := os.Stat(path)
	if err != nil {
		vm.RaiseError("OSError: %s", err.Error())
		return 0
	}
	vm.Push(buildStatResult(info))
	return 1
}

// os.lstat(path) - like stat but doesn't follow symlinks
func osLstat(vm *gopherpy.VM) int {
	path := vm.CheckString(1)
	info, err := os.Lstat(path)
	if err != nil {
		vm.RaiseError("OSError: %s", err.Error())
		return 0
	}
	vm.Push(buildStatResult(info))
	return 1
}

// buildStatResult creates a dict from os.FileInfo
func buildStatResult(info os.FileInfo) *gopherpy.PyDict {
	result := gopherpy.NewDict()
	result.Items[gopherpy.NewString("st_mode")] = gopherpy.NewInt(int64(info.Mode()))
	result.Items[gopherpy.NewString("st_size")] = gopherpy.NewInt(info.Size())
	result.Items[gopherpy.NewString("st_mtime")] = gopherpy.NewFloat(float64(info.ModTime().Unix()))
	result.Items[gopherpy.NewString("st_atime")] = gopherpy.NewFloat(float64(info.ModTime().Unix()))
	result.Items[gopherpy.NewString("st_ctime")] = gopherpy.NewFloat(float64(info.ModTime().Unix()))
	result.Items[gopherpy.NewString("st_nlink")] = gopherpy.NewInt(1)
	result.Items[gopherpy.NewString("st_uid")] = gopherpy.NewInt(0)
	result.Items[gopherpy.NewString("st_gid")] = gopherpy.NewInt(0)

	// Determine file type for S_IF* constants
	mode := info.Mode()
	var stMode int64 = int64(mode.Perm())
	if mode.IsDir() {
		stMode |= 0o040000 // S_IFDIR
	} else if mode.IsRegular() {
		stMode |= 0o100000 // S_IFREG
	} else if mode&os.ModeSymlink != 0 {
		stMode |= 0o120000 // S_IFLNK
	}
	result.Items[gopherpy.NewString("st_mode")] = gopherpy.NewInt(stMode)

	return result
}

// os.link(src, dst) - create hard link
func osLink(vm *gopherpy.VM) int {
	src := vm.CheckString(1)
	dst := vm.CheckString(2)
	if err := os.Link(src, dst); err != nil {
		vm.RaiseError("OSError: %s", err.Error())
		return 0
	}
	return 0
}

// os.symlink(src, dst) - create symbolic link
func osSymlink(vm *gopherpy.VM) int {
	src := vm.CheckString(1)
	dst := vm.CheckString(2)
	if err := os.Symlink(src, dst); err != nil {
		vm.RaiseError("OSError: %s", err.Error())
		return 0
	}
	return 0
}

// os.readlink(path) - read symbolic link target
func osReadlink(vm *gopherpy.VM) int {
	path := vm.CheckString(1)
	target, err := os.Readlink(path)
	if err != nil {
		vm.RaiseError("OSError: %s", err.Error())
		return 0
	}
	vm.Push(gopherpy.NewString(target))
	return 1
}

// os.chmod(path, mode)
func osChmod(vm *gopherpy.VM) int {
	path := vm.CheckString(1)
	mode := os.FileMode(vm.CheckInt(2))
	if err := os.Chmod(path, mode); err != nil {
		vm.RaiseError("OSError: %s", err.Error())
		return 0
	}
	return 0
}

// os.access(path, mode)
func osAccess(vm *gopherpy.VM) int {
	path := vm.CheckString(1)
	mode := int(vm.CheckInt(2))

	// Access mode constants: F_OK=0, R_OK=4, W_OK=2, X_OK=1
	info, err := os.Stat(path)
	if err != nil {
		vm.Push(gopherpy.NewBool(false))
		return 1
	}

	// F_OK - existence check
	if mode == 0 {
		vm.Push(gopherpy.NewBool(true))
		return 1
	}

	// Simplified permission check based on file mode
	perm := info.Mode().Perm()
	result := true

	if mode&4 != 0 { // R_OK
		result = result && (perm&0444 != 0)
	}
	if mode&2 != 0 { // W_OK
		result = result && (perm&0222 != 0)
	}
	if mode&1 != 0 { // X_OK
		result = result && (perm&0111 != 0)
	}

	vm.Push(gopherpy.NewBool(result))
	return 1
}

// os.truncate(path, length)
func osTruncate(vm *gopherpy.VM) int {
	path := vm.CheckString(1)
	length := vm.CheckInt(2)
	if err := os.Truncate(path, length); err != nil {
		vm.RaiseError("OSError: %s", err.Error())
		return 0
	}
	return 0
}

// ==================== Process Functions ====================

// os.getpid()
func osGetpid(vm *gopherpy.VM) int {
	vm.Push(gopherpy.NewInt(int64(os.Getpid())))
	return 1
}

// os.getppid()
func osGetppid(vm *gopherpy.VM) int {
	vm.Push(gopherpy.NewInt(int64(os.Getppid())))
	return 1
}

// os.getuid()
func osGetuid(vm *gopherpy.VM) int {
	vm.Push(gopherpy.NewInt(int64(os.Getuid())))
	return 1
}

// os.getgid()
func osGetgid(vm *gopherpy.VM) int {
	vm.Push(gopherpy.NewInt(int64(os.Getgid())))
	return 1
}

// os.geteuid()
func osGeteuid(vm *gopherpy.VM) int {
	vm.Push(gopherpy.NewInt(int64(os.Geteuid())))
	return 1
}

// os.getegid()
func osGetegid(vm *gopherpy.VM) int {
	vm.Push(gopherpy.NewInt(int64(os.Getegid())))
	return 1
}

// os.getlogin()
func osGetlogin(vm *gopherpy.VM) int {
	u, err := user.Current()
	if err != nil {
		vm.RaiseError("OSError: %s", err.Error())
		return 0
	}
	vm.Push(gopherpy.NewString(u.Username))
	return 1
}

// os.uname() - returns dict with system info
func osUname(vm *gopherpy.VM) int {
	result := gopherpy.NewDict()
	result.Items[gopherpy.NewString("sysname")] = gopherpy.NewString(goruntime.GOOS)
	result.Items[gopherpy.NewString("machine")] = gopherpy.NewString(goruntime.GOARCH)
	result.Items[gopherpy.NewString("release")] = gopherpy.NewString("")
	result.Items[gopherpy.NewString("version")] = gopherpy.NewString("")

	hostname, _ := os.Hostname()
	result.Items[gopherpy.NewString("nodename")] = gopherpy.NewString(hostname)

	vm.Push(result)
	return 1
}

// os.cpu_count()
func osCpuCount(vm *gopherpy.VM) int {
	vm.Push(gopherpy.NewInt(int64(goruntime.NumCPU())))
	return 1
}

// os.urandom(n) - return n random bytes
func osUrandom(vm *gopherpy.VM) int {
	n := vm.CheckInt(1)
	if n < 0 {
		vm.RaiseError("ValueError: negative argument not allowed")
		return 0
	}

	buf := make([]byte, n)
	// Use crypto/rand for secure random bytes
	_, err := cryptoRand(buf)
	if err != nil {
		vm.RaiseError("OSError: %s", err.Error())
		return 0
	}
	vm.Push(gopherpy.NewBytes(buf))
	return 1
}

// cryptoRand reads random bytes using crypto/rand
func cryptoRand(b []byte) (int, error) {
	return rand.Read(b)
}

// ==================== Path Functions ====================

// os.fspath(path)
func osFspath(vm *gopherpy.VM) int {
	arg := vm.Get(1)
	switch v := arg.(type) {
	case *gopherpy.PyString:
		vm.Push(arg)
	case *gopherpy.PyBytes:
		vm.Push(arg)
	default:
		vm.RaiseError("TypeError: expected str or bytes, got %T", v)
		return 0
	}
	return 1
}

// os.fsencode(filename)
func osFsencode(vm *gopherpy.VM) int {
	s := vm.CheckString(1)
	vm.Push(gopherpy.NewBytes([]byte(s)))
	return 1
}

// os.fsdecode(filename)
func osFsdecode(vm *gopherpy.VM) int {
	arg := vm.Get(1)
	switch v := arg.(type) {
	case *gopherpy.PyString:
		vm.Push(arg)
	case *gopherpy.PyBytes:
		vm.Push(gopherpy.NewString(string(v.Value)))
	default:
		vm.RaiseError("TypeError: expected bytes or str, got %T", v)
		return 0
	}
	return 1
}

// ==================== System Functions ====================

// os.system(command)
func osSystem(vm *gopherpy.VM) int {
	// Not fully implemented - would need exec
	vm.RaiseError("NotImplementedError: os.system() is not supported")
	return 0
}

// os.abort()
func osAbort(vm *gopherpy.VM) int {
	os.Exit(1)
	return 0
}

// os._exit(code)
func osExit(vm *gopherpy.VM) int {
	code := 0
	if vm.GetTop() >= 1 {
		code = int(vm.CheckInt(1))
	}
	os.Exit(code)
	return 0
}

// ==================== os.path Functions ====================

// os.path.join(*paths)
func osPathJoin(vm *gopherpy.VM) int {
	nargs := vm.GetTop()
	if nargs == 0 {
		vm.Push(gopherpy.NewString(""))
		return 1
	}

	paths := make([]string, nargs)
	for i := 1; i <= nargs; i++ {
		paths[i-1] = vm.CheckString(i)
	}

	vm.Push(gopherpy.NewString(filepath.Join(paths...)))
	return 1
}

// os.path.split(path)
func osPathSplit(vm *gopherpy.VM) int {
	path := vm.CheckString(1)
	dir, file := filepath.Split(path)
	// Remove trailing separator from dir (Python behavior)
	if len(dir) > 1 && (dir[len(dir)-1] == '/' || dir[len(dir)-1] == '\\') {
		dir = dir[:len(dir)-1]
	}
	vm.Push(gopherpy.NewTuple([]gopherpy.Value{
		gopherpy.NewString(dir),
		gopherpy.NewString(file),
	}))
	return 1
}

// os.path.splitext(path)
func osPathSplitext(vm *gopherpy.VM) int {
	path := vm.CheckString(1)
	ext := filepath.Ext(path)
	root := path[:len(path)-len(ext)]
	vm.Push(gopherpy.NewTuple([]gopherpy.Value{
		gopherpy.NewString(root),
		gopherpy.NewString(ext),
	}))
	return 1
}

// os.path.basename(path)
func osPathBasename(vm *gopherpy.VM) int {
	path := vm.CheckString(1)
	vm.Push(gopherpy.NewString(filepath.Base(path)))
	return 1
}

// os.path.dirname(path)
func osPathDirname(vm *gopherpy.VM) int {
	path := vm.CheckString(1)
	vm.Push(gopherpy.NewString(filepath.Dir(path)))
	return 1
}

// os.path.exists(path)
func osPathExists(vm *gopherpy.VM) int {
	path := vm.CheckString(1)
	_, err := os.Stat(path)
	vm.Push(gopherpy.NewBool(err == nil))
	return 1
}

// os.path.isfile(path)
func osPathIsfile(vm *gopherpy.VM) int {
	path := vm.CheckString(1)
	info, err := os.Stat(path)
	vm.Push(gopherpy.NewBool(err == nil && info.Mode().IsRegular()))
	return 1
}

// os.path.isdir(path)
func osPathIsdir(vm *gopherpy.VM) int {
	path := vm.CheckString(1)
	info, err := os.Stat(path)
	vm.Push(gopherpy.NewBool(err == nil && info.IsDir()))
	return 1
}

// os.path.isabs(path)
func osPathIsabs(vm *gopherpy.VM) int {
	path := vm.CheckString(1)
	vm.Push(gopherpy.NewBool(filepath.IsAbs(path)))
	return 1
}

// os.path.islink(path)
func osPathIslink(vm *gopherpy.VM) int {
	path := vm.CheckString(1)
	info, err := os.Lstat(path)
	vm.Push(gopherpy.NewBool(err == nil && info.Mode()&os.ModeSymlink != 0))
	return 1
}

// os.path.abspath(path)
func osPathAbspath(vm *gopherpy.VM) int {
	path := vm.CheckString(1)
	abs, err := filepath.Abs(path)
	if err != nil {
		vm.RaiseError("OSError: %s", err.Error())
		return 0
	}
	vm.Push(gopherpy.NewString(abs))
	return 1
}

// os.path.normpath(path)
func osPathNormpath(vm *gopherpy.VM) int {
	path := vm.CheckString(1)
	vm.Push(gopherpy.NewString(filepath.Clean(path)))
	return 1
}

// os.path.realpath(path)
func osPathRealpath(vm *gopherpy.VM) int {
	path := vm.CheckString(1)
	real, err := filepath.EvalSymlinks(path)
	if err != nil {
		// Fall back to absolute path if symlink resolution fails
		abs, err2 := filepath.Abs(path)
		if err2 != nil {
			vm.RaiseError("OSError: %s", err.Error())
			return 0
		}
		vm.Push(gopherpy.NewString(abs))
		return 1
	}
	vm.Push(gopherpy.NewString(real))
	return 1
}

// os.path.expanduser(path)
func osPathExpanduser(vm *gopherpy.VM) int {
	path := vm.CheckString(1)

	if !strings.HasPrefix(path, "~") {
		vm.Push(gopherpy.NewString(path))
		return 1
	}

	var home string
	u, err := user.Current()
	if err == nil {
		home = u.HomeDir
	} else {
		home = os.Getenv("HOME")
		if home == "" && goruntime.GOOS == "windows" {
			home = os.Getenv("USERPROFILE")
		}
	}

	if home == "" {
		vm.Push(gopherpy.NewString(path))
		return 1
	}

	if path == "~" {
		vm.Push(gopherpy.NewString(home))
		return 1
	}

	if strings.HasPrefix(path, "~/") || strings.HasPrefix(path, "~\\") {
		vm.Push(gopherpy.NewString(filepath.Join(home, path[2:])))
		return 1
	}

	vm.Push(gopherpy.NewString(path))
	return 1
}

// os.path.expandvars(path)
func osPathExpandvars(vm *gopherpy.VM) int {
	path := vm.CheckString(1)
	vm.Push(gopherpy.NewString(os.ExpandEnv(path)))
	return 1
}

// os.path.getsize(path)
func osPathGetsize(vm *gopherpy.VM) int {
	path := vm.CheckString(1)
	info, err := os.Stat(path)
	if err != nil {
		vm.RaiseError("OSError: %s", err.Error())
		return 0
	}
	vm.Push(gopherpy.NewInt(info.Size()))
	return 1
}

// os.path.getatime(path)
func osPathGetatime(vm *gopherpy.VM) int {
	path := vm.CheckString(1)
	info, err := os.Stat(path)
	if err != nil {
		vm.RaiseError("OSError: %s", err.Error())
		return 0
	}
	vm.Push(gopherpy.NewFloat(float64(info.ModTime().Unix())))
	return 1
}

// os.path.getmtime(path)
func osPathGetmtime(vm *gopherpy.VM) int {
	path := vm.CheckString(1)
	info, err := os.Stat(path)
	if err != nil {
		vm.RaiseError("OSError: %s", err.Error())
		return 0
	}
	vm.Push(gopherpy.NewFloat(float64(info.ModTime().Unix())))
	return 1
}

// os.path.getctime(path)
func osPathGetctime(vm *gopherpy.VM) int {
	path := vm.CheckString(1)
	info, err := os.Stat(path)
	if err != nil {
		vm.RaiseError("OSError: %s", err.Error())
		return 0
	}
	// Go doesn't expose ctime directly; use mtime as fallback
	vm.Push(gopherpy.NewFloat(float64(info.ModTime().Unix())))
	return 1
}

// os.path.samefile(path1, path2)
func osPathSamefile(vm *gopherpy.VM) int {
	path1 := vm.CheckString(1)
	path2 := vm.CheckString(2)

	abs1, err1 := filepath.Abs(path1)
	abs2, err2 := filepath.Abs(path2)

	if err1 != nil || err2 != nil {
		vm.RaiseError("OSError: cannot resolve paths")
		return 0
	}

	real1, _ := filepath.EvalSymlinks(abs1)
	real2, _ := filepath.EvalSymlinks(abs2)

	if real1 == "" {
		real1 = abs1
	}
	if real2 == "" {
		real2 = abs2
	}

	vm.Push(gopherpy.NewBool(real1 == real2))
	return 1
}

// os.path.commonpath(paths)
func osPathCommonpath(vm *gopherpy.VM) int {
	list := vm.CheckList(1)
	if len(list.Items) == 0 {
		vm.RaiseError("ValueError: commonpath() arg is an empty sequence")
		return 0
	}

	paths := make([]string, len(list.Items))
	for i, item := range list.Items {
		s, ok := item.(*gopherpy.PyString)
		if !ok {
			vm.RaiseError("TypeError: expected str, got %T", item)
			return 0
		}
		paths[i] = filepath.Clean(s.Value)
	}

	// Find common prefix component by component
	first := strings.Split(paths[0], string(filepath.Separator))
	commonLen := len(first)

	for _, p := range paths[1:] {
		parts := strings.Split(p, string(filepath.Separator))
		for i := 0; i < commonLen && i < len(parts); i++ {
			if first[i] != parts[i] {
				commonLen = i
				break
			}
		}
		if len(parts) < commonLen {
			commonLen = len(parts)
		}
	}

	common := strings.Join(first[:commonLen], string(filepath.Separator))
	if common == "" && filepath.IsAbs(paths[0]) {
		common = string(filepath.Separator)
	}

	vm.Push(gopherpy.NewString(common))
	return 1
}

// os.path.commonprefix(list)
func osPathCommonprefix(vm *gopherpy.VM) int {
	list := vm.CheckList(1)
	if len(list.Items) == 0 {
		vm.Push(gopherpy.NewString(""))
		return 1
	}

	paths := make([]string, len(list.Items))
	for i, item := range list.Items {
		s, ok := item.(*gopherpy.PyString)
		if !ok {
			vm.RaiseError("TypeError: expected str, got %T", item)
			return 0
		}
		paths[i] = s.Value
	}

	prefix := paths[0]
	for _, p := range paths[1:] {
		for !strings.HasPrefix(p, prefix) {
			prefix = prefix[:len(prefix)-1]
			if prefix == "" {
				vm.Push(gopherpy.NewString(""))
				return 1
			}
		}
	}

	vm.Push(gopherpy.NewString(prefix))
	return 1
}

// os.path.relpath(path, start='.')
func osPathRelpath(vm *gopherpy.VM) int {
	path := vm.CheckString(1)
	start := vm.OptionalString(2, ".")

	rel, err := filepath.Rel(start, path)
	if err != nil {
		vm.RaiseError("ValueError: %s", err.Error())
		return 0
	}
	vm.Push(gopherpy.NewString(rel))
	return 1
}
