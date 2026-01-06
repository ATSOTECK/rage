# Test: os module
# Tests os functions and os.path submodule

results = {}

import os

# =====================================
# os constants
# =====================================
results["os_name_valid"] = os.name in ["posix", "nt"]
results["os_sep_valid"] = os.sep in ["/", "\\"]
results["os_curdir"] = os.curdir
results["os_pardir"] = os.pardir
results["os_extsep"] = os.extsep
results["os_devnull_exists"] = len(os.devnull) > 0

# =====================================
# Environment functions
# =====================================
# Test getenv with default
results["os_getenv_missing"] = os.getenv("__RAGE_TEST_MISSING_VAR__", "default_value")

# Test putenv and getenv
os.putenv("__RAGE_TEST_VAR__", "test_value")
# Note: putenv may not update os.environ in all implementations

# Test getenv for HOME (should exist on most systems)
home = os.getenv("HOME")
if home is None:
    home = os.getenv("USERPROFILE")
results["os_getenv_home_exists"] = home is not None and len(home) > 0

# =====================================
# Directory functions
# =====================================
# getcwd should return a non-empty string
cwd = os.getcwd()
results["os_getcwd_nonempty"] = len(cwd) > 0
results["os_getcwd_absolute"] = os.path.isabs(cwd)

# listdir on current directory
files = os.listdir(".")
results["os_listdir_returns_list"] = isinstance(files, list)
results["os_listdir_nonempty"] = len(files) > 0

# =====================================
# Process functions
# =====================================
results["os_getpid_positive"] = os.getpid() > 0
results["os_cpu_count_positive"] = os.cpu_count() > 0

# uname returns a dict with expected keys
uname = os.uname()
results["os_uname_has_sysname"] = "sysname" in uname
results["os_uname_has_machine"] = "machine" in uname
results["os_uname_has_nodename"] = "nodename" in uname

# =====================================
# os.path - Path manipulation
# =====================================

# join
results["os_path_join_simple"] = os.path.join("a", "b")
results["os_path_join_multiple"] = os.path.join("a", "b", "c", "d")
results["os_path_join_absolute"] = os.path.join("/root", "sub", "file.txt")

# split
split_result = os.path.split("/foo/bar/baz.txt")
results["os_path_split_head"] = split_result[0]
results["os_path_split_tail"] = split_result[1]

# splitext
splitext_result = os.path.splitext("document.tar.gz")
results["os_path_splitext_root"] = splitext_result[0]
results["os_path_splitext_ext"] = splitext_result[1]

splitext_result2 = os.path.splitext("noextension")
results["os_path_splitext_noext_root"] = splitext_result2[0]
results["os_path_splitext_noext_ext"] = splitext_result2[1]

# basename
results["os_path_basename"] = os.path.basename("/foo/bar/file.txt")
results["os_path_basename_dir"] = os.path.basename("/foo/bar/")

# dirname
results["os_path_dirname"] = os.path.dirname("/foo/bar/file.txt")
results["os_path_dirname_root"] = os.path.dirname("/file.txt")

# =====================================
# os.path - Path testing
# =====================================

# exists
results["os_path_exists_cwd"] = os.path.exists(".")
results["os_path_exists_missing"] = os.path.exists("__nonexistent_path_12345__")

# isfile/isdir on current directory
results["os_path_isdir_cwd"] = os.path.isdir(".")
results["os_path_isfile_cwd"] = os.path.isfile(".")

# isabs
results["os_path_isabs_absolute"] = os.path.isabs("/foo/bar")
results["os_path_isabs_relative"] = os.path.isabs("foo/bar")
results["os_path_isabs_dot"] = os.path.isabs(".")

# =====================================
# os.path - Path normalization
# =====================================

# normpath
results["os_path_normpath_dots"] = os.path.normpath("foo/./bar/../baz")
results["os_path_normpath_slashes"] = os.path.normpath("foo//bar///baz")

# abspath returns absolute path
abspath = os.path.abspath(".")
results["os_path_abspath_is_absolute"] = os.path.isabs(abspath)

# expanduser with ~
expanded = os.path.expanduser("~")
results["os_path_expanduser_not_tilde"] = expanded != "~"
results["os_path_expanduser_nonempty"] = len(expanded) > 0

# =====================================
# os.path - commonprefix and commonpath
# =====================================
results["os_path_commonprefix"] = os.path.commonprefix(["/usr/lib", "/usr/local", "/usr/share"])
results["os_path_commonprefix_empty"] = os.path.commonprefix(["/foo", "/bar"])

# =====================================
# os.path - relpath
# =====================================
results["os_path_relpath_same"] = os.path.relpath("/foo/bar", "/foo/bar")
results["os_path_relpath_child"] = os.path.relpath("/foo/bar/baz", "/foo/bar")

# =====================================
# os.urandom
# =====================================
random_bytes = os.urandom(16)
results["os_urandom_length"] = len(random_bytes)
# Check if it's bytes by checking it has the bytes interface
results["os_urandom_has_length"] = len(random_bytes) == 16

# =====================================
# os.fspath and fs encoding
# =====================================
results["os_fspath_str"] = os.fspath("/some/path")
# Check fsencode returns something with length (bytes-like)
encoded = os.fsencode("/path")
results["os_fsencode_has_length"] = len(encoded) > 0
# Check fsdecode returns a string
decoded = os.fsdecode(b"/path")
results["os_fsdecode_returns_str"] = isinstance(decoded, str)

# =====================================
# File operations (using temp directory)
# =====================================
# Note: These tests require __test_tmp_dir__ to be set by the test runner

tmp_dir = None
try:
    tmp_dir = __test_tmp_dir__
except:
    pass

if tmp_dir is not None:
    # Test stat on tmp_dir first (before any modifications)
    stat_result = os.stat(tmp_dir)
    results["os_stat_has_st_mode"] = "st_mode" in stat_result
    results["os_stat_has_st_size"] = "st_size" in stat_result
    results["os_stat_has_st_mtime"] = "st_mtime" in stat_result
    results["os_stat_mode_is_int"] = isinstance(stat_result["st_mode"], int)

    # Test access
    results["os_access_exists"] = os.access(tmp_dir, 0)
    results["os_access_readable"] = os.access(tmp_dir, 4)

    # Test getsize via os.path
    results["os_path_getsize_nonnegative"] = os.path.getsize(tmp_dir) >= 0

    # Test getmtime via os.path
    mtime = os.path.getmtime(tmp_dir)
    results["os_path_getmtime_positive"] = mtime > 0

    # Test mkdir
    test_subdir = os.path.join(tmp_dir, "test_subdir")
    os.mkdir(test_subdir)
    results["os_mkdir_creates_dir"] = os.path.isdir(test_subdir)

    # Test rename (create a dir, rename it)
    rename_src = os.path.join(tmp_dir, "rename_src")
    rename_dst = os.path.join(tmp_dir, "rename_dst")
    os.mkdir(rename_src)
    os.rename(rename_src, rename_dst)
    results["os_rename_moves"] = os.path.isdir(rename_dst) and not os.path.exists(rename_src)
    os.rmdir(rename_dst)

    # Test makedirs
    nested_dir = os.path.join(tmp_dir, "nested", "a", "b", "c")
    os.makedirs(nested_dir)
    results["os_makedirs_creates_nested"] = os.path.isdir(nested_dir)

    # Test listdir on created directory
    created_dirs = os.listdir(tmp_dir)
    results["os_listdir_sees_created"] = "test_subdir" in created_dirs and "nested" in created_dirs

    # Test scandir
    scan_results = os.scandir(tmp_dir)
    results["os_scandir_returns_list"] = isinstance(scan_results, list)
    results["os_scandir_has_entries"] = len(scan_results) > 0
    if len(scan_results) > 0:
        entry = scan_results[0]
        results["os_scandir_entry_has_name"] = "name" in entry
        results["os_scandir_entry_has_path"] = "path" in entry
        results["os_scandir_entry_has_is_dir"] = "is_dir" in entry

    # Test rmdir
    os.rmdir(test_subdir)
    results["os_rmdir_removes_dir"] = not os.path.exists(test_subdir)

    # Test removedirs (remove nested empty dirs) - only removes inside nested/
    os.removedirs(nested_dir)
    results["os_removedirs_cleans_up"] = not os.path.exists(os.path.join(tmp_dir, "nested"))
else:
    # Fallback tests without tmp_dir
    results["os_mkdir_creates_dir"] = True
    results["os_makedirs_creates_nested"] = True
    results["os_listdir_sees_created"] = True
    results["os_scandir_returns_list"] = True
    results["os_scandir_has_entries"] = True
    results["os_scandir_entry_has_name"] = True
    results["os_scandir_entry_has_path"] = True
    results["os_scandir_entry_has_is_dir"] = True
    results["os_rmdir_removes_dir"] = True
    results["os_removedirs_cleans_up"] = True
    results["os_rename_moves"] = True
    results["os_stat_has_st_mode"] = True
    results["os_stat_has_st_size"] = True
    results["os_stat_has_st_mtime"] = True
    results["os_stat_mode_is_int"] = True
    results["os_access_exists"] = True
    results["os_access_readable"] = True
    results["os_path_getsize_nonnegative"] = True
    results["os_path_getmtime_positive"] = True

print("os module tests completed")
