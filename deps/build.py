#!/usr/bin/env python3
import argparse
import glob
import os
import platform
import shutil
import subprocess
import sys

valid_archs = ['arm64', 'amd64']
# "amd64" is called "x86_64" on everything but Windows.
current_arch = platform.uname()[4].lower().replace("x86_64", "amd64")
default_arch = current_arch if current_arch in valid_archs else None

parser = argparse.ArgumentParser()
parser.add_argument('--verbose', '-v', default=False, action='store_true')
parser.add_argument('--debug', default=False, action='store_true')
parser.add_argument('--ccache', default=False, action='store_true')
parser.add_argument('--clang', action='store_true')
parser.add_argument('--no-clang', dest='clang', action='store_false')
parser.set_defaults(clang=None)
# GitHub file size limits: warning at 50 MB, hard limit at 100 MB.
# Symbol indices can add 15% in the final .ar, so we need margin.
parser.add_argument('--max-file-size', default=int(40e6))
parser.add_argument('--arch',
    dest='arch',
    action='store',
    choices=valid_archs,
    default=default_arch,
    required=default_arch is None)
parser.add_argument(
    '--os',
    dest='os',
    choices=['android', 'ios', 'linux', 'darwin', 'windows'],
    default=platform.system().lower())
args = parser.parse_args()

deps_path = os.path.dirname(os.path.realpath(__file__))
v8_path = os.path.join(deps_path, "v8")
tools_path = os.path.join(deps_path, "depot_tools")
is_windows = platform.system().lower() == "windows"
# Chromium only supports Clang for our targets, and its libc++ requires
# it.
is_clang = args.clang if args.clang is not None else True

def get_custom_deps():
    # These deps are unnecessary for building.
    deps = {
        "v8/testing/gmock"                      : None,
        "v8/test/wasm-js"                       : None,
        "v8/third_party/colorama/src"           : None,
        "v8/tools/gyp"                          : None,
        "v8/tools/luci-go"                      : None,
    }
    if args.os != "android":
        deps["v8/third_party/catapult"] = None
        deps["v8/third_party/android_tools"] = None
    return deps

gclient_sln = [
    { "name"        : "v8",
        "url"         : "https://chromium.googlesource.com/v8/v8.git",
        "deps_file"   : "DEPS",
        "managed"     : False,
        "custom_deps" : get_custom_deps(),
        "custom_vars": {
            "build_for_node" : True,
        },
    },
]

gn_args = """
is_debug=%s
is_clang=%s
target_os="%s"
target_cpu="%s"
v8_target_cpu="%s"
clang_use_chrome_plugins=false
use_custom_libcxx=true
use_sysroot=%s
use_glib=false
symbol_level=%s
strip_debug_info=%s
is_component_build=false
v8_monolithic=true
v8_use_external_startup_data=false
treat_warnings_as_errors=false
v8_embedder_string="-v8go"
v8_enable_gdbjit=false
v8_enable_i18n_support=true
icu_use_data_file=false
v8_enable_test_features=false
exclude_unwind_tables=true
v8_android_log_stdout=true
v8_enable_temporal_support=false
"""

# Chromium's pinned Debian sysroot, used on Linux. This makes the build
# independent of the host's glibc, and sets the minimum glibc version
# users need. It only affects Linux; Android uses the NDK sysroot, and
# macOS the Xcode SDK.
def use_sysroot():
    return args.os == "linux"

# Libraries in the sysroot that cgo users link with. librt is not
# included, since update_cgo.py doesn't add -lrt.
SYSROOT_TRIPLES = {"amd64": "x86_64-linux-gnu", "arm64": "aarch64-linux-gnu"}
SYSROOT_SHARED_LIBS = ["libc.so.6", "libm.so.6", "libdl.so.2", "libpthread.so.0", "libgcc_s.so.1"]
SYSROOT_LOADERS = {"amd64": "ld-linux-x86-64.so.2", "arm64": "ld-linux-aarch64.so.1"}
# The libc.so linker script adds this, which defines e.g. stat64 and
# pthread_atfork before glibc 2.33/2.34.
SYSROOT_STATIC_LIBS = ["libc_nonshared.a"]

# Defined by the linker or crt*.o, rather than any library.
LINKER_SYMBOLS = {"__dso_handle", "_GLOBAL_OFFSET_TABLE_", "__ehdr_start"}

def is_linker_symbol(name):
    # The linker defines __start_/__stop_ symbols for named sections.
    return name in LINKER_SYMBOLS or name.startswith(("__start_", "__stop_"))

def sysroot_path(arch):
    return os.path.join(v8_path, "build", "linux", "debian_bullseye_{}-sysroot".format(arch))

def install_sysroots():
    # Host tools, like mksnapshot, are built for the host CPU.
    for arch in sorted({args.arch, current_arch}):
        subprocess_check_call([sys.executable, "build/linux/sysroot_scripts/install-sysroot.py", "--arch=" + arch], cwd=v8_path)

def llvm_tool(name):
    """Returns the path to a tool in Chromium's pinned LLVM."""
    return os.path.join(v8_path, "third_party", "llvm-build", "Release+Asserts", "bin", name + (".exe" if is_windows else ""))

def nm(nm_args):
    nm_path = llvm_tool("llvm-nm")
    # stderr only contains "no symbols" for some members.
    return subprocess_check_output_text([nm_path] + nm_args, stderr=subprocess.DEVNULL).splitlines()

def check_sysroot_symbols(dest_path):
    """Fails if the archives need symbols the sysroot doesn't provide.

    Symbol versions are only bound when users link, so we check that
    every undefined symbol is available in the sysroot's libraries.
    Building against the host's headers instead can e.g. redirect strtol
    to __isoc23_strtol, which requires glibc 2.38.
    """
    archives = sorted(glob.glob(os.path.join(dest_path, "lib*.a")))
    defined = set()
    undefined = set()
    for line in nm(["--no-sort"] + archives):
        fields = line.split()
        if len(fields) < 2 or line.endswith(":"):
            continue
        kind, name = fields[-2], fields[-1]
        if kind == "U":
            undefined.add(name)
        elif kind not in ("w", "v"):
            defined.add(name)

    root = sysroot_path(args.arch)
    triple = SYSROOT_TRIPLES[args.arch]
    shared = [os.path.join(root, "lib", triple, lib) for lib in SYSROOT_SHARED_LIBS + [SYSROOT_LOADERS[args.arch]]]
    static = [os.path.join(root, "usr", "lib", triple, lib) for lib in SYSROOT_STATIC_LIBS]
    libgcc = glob.glob(os.path.join(root, "usr", "lib", "gcc", triple, "*", "libgcc.a"))
    if len(libgcc) != 1:
        raise RuntimeError("expected one libgcc.a in {}, found {}".format(root, libgcc))
    static += libgcc
    for path in shared + static:
        if not os.path.isfile(path):
            raise RuntimeError("sysroot library not found: {}".format(path))

    provided = set()
    for path in shared:
        provided.update(name.split("@")[0] for name in nm(["-D", "--defined-only", "-j", path]))
    provided.update(nm(["--defined-only", "-j"] + static))

    missing = sorted(name for name in undefined - defined - provided if not is_linker_symbol(name))
    if missing:
        raise RuntimeError("symbols not provided by {}:\n  {}".format(root, "\n  ".join(missing)))

def v8deps():
    spec = "solutions = %s\n" % gclient_sln
    spec += "target_os = [%r]" % (v8_os(),)
    env = os.environ.copy()
    env["PATH"] = tools_path + os.pathsep + env["PATH"]
    # --force: disable_crel() modifies v8/build, which would otherwise
    # stop a sync when build.py runs again on the same tree.
    subprocess_check_call(["gclient", "sync", "--force", "--delete_unversioned_trees", "--no-history", "--spec", spec],
                        cwd=deps_path,
                        env=env)

CREL_CFLAGS = 'cflags += [ "-Wa,--crel,--allow-experimental-crel" ]'
CREL_DISABLED = "# v8go: CREL disabled."

def disable_crel():
    """Stops Chromium's Clang from emitting CREL relocations.

    Only LLD reads them, while cgo users link with the system linker.
    There is no GN argument for this, so we edit the build config.
    """
    path = os.path.join(v8_path, "build", "config", "compiler", "BUILD.gn")
    with open(path, "rt") as f:
        source = f.read()

    if CREL_CFLAGS in source:
        with open(path, "wt") as f:
            f.write(source.replace(CREL_CFLAGS, CREL_DISABLED))
    elif CREL_DISABLED not in source:
        raise RuntimeError("CREL flags not found in {}; check if disable_crel() is still needed".format(path))

def build_gn_args():
    is_debug = args.debug
    arch = v8_arch()
    # symbol_level = 1 includes line number information
    # symbol_level = 2 can be used for additional debug information, but it can increase the
    #   compiled library by an order of magnitude and further slow down compilation
    symbol_level = 1 if args.debug else 0
    strip_debug_info = not args.debug

    gnargs = gn_args % (
        str(bool(is_debug)).lower(),
        str(is_clang).lower(),
        v8_os(),
        arch,
        arch,
        str(use_sysroot()).lower(),
        symbol_level,
        str(strip_debug_info).lower(),
    )
    if args.ccache:
        gnargs += 'cc_wrapper="ccache"\n'
    if not is_clang and arch == "arm64":
        # https://chromium.googlesource.com/chromium/deps/icu/+/2958a507f15e475045906d73af39018d5038a93b
        # introduced -mmark-bti-property, which isn't supported by GCC.
        #
        # V8 itself fixed this in https://chromium-review.googlesource.com/c/v8/v8/+/3930160.
        gnargs += 'arm_control_flow_integrity="none"\n'

    return gnargs

def subprocess_check_call(cmdargs, *pargs, **kwargs):
    if args.verbose:
        print(sys.argv[0], ">", " ".join(cmdargs), file=sys.stderr)
    subprocess.check_call(cmd(cmdargs), *pargs, **kwargs)

def subprocess_check_output_text(cmdargs, *pargs, **kwargs):
    if args.verbose:
        print(sys.argv[0], ">", " ".join(cmdargs), file=sys.stderr)
    return subprocess.check_output(cmd(cmdargs), *pargs, **kwargs).decode('utf-8')

def cmd(args):
    return ["cmd", "/c"] + args if is_windows else args

def os_arch():
    return args.os + "_" + args.arch

def v8_os():
    return {"darwin": "mac", "windows": "win"}.get(args.os, args.os)

def v8_arch():
    if args.arch == "amd64":
        return "x64"
    return args.arch

def split_ar(src_fn, dest_fn, dest_obj_dn):
    """Extracts all files from src_fn to dest_obj_dn/ and makes a thin archive at dest_fn.

    GitHub's file size limit is 100 MiB, and the archive is hitting that.
    """
    dest_path = os.path.dirname(dest_fn)

    ar_path = os.path.abspath(os.path.join(v8_path, "third_party/llvm-build/Release+Asserts/bin/llvm-ar"))
    if args.os == "linux" and args.arch == "arm64" and not is_clang:
        ar_path = "aarch64-linux-gnu-ar"
    elif not os.access(ar_path, os.X_OK) or not is_clang:
        ar_path = "ar"

    if os.path.exists(dest_obj_dn):
        shutil.rmtree(dest_obj_dn)
    os.makedirs(dest_obj_dn)

    # Directories may have been flattened, causing duplicate file
    # names. ar(1) simply overwrites earlier files, causing
    # headache-inducing "undefined symbol" errors.
    ar_files = subprocess_check_output_text(
        [
            ar_path,
            "t",
            src_fn,
        ],
        cwd=v8_path)
    ar_files = ar_files.splitlines()

    # llvm-ar (--clang) for Darwin (but not Android) seems to mangle
    # the names to lowercase on extraction, while others do not.
    case_sensitive = args.os != "darwin"

    # Extracting files one-by-one is slow, so let's group them into
    # disjoint sets and use "ar N"... Complicated by the occasional
    # case mangling.
    ar_file_groups = allocate_disjoint_files(ar_files, case_sensitive)

    j = 0
    for i, ar_files in ar_file_groups:
        subprocess_check_call(
            [
                ar_path,
                "xN",
                "--output", dest_obj_dn,
                str(1 + i),
                src_fn,
            ] + ar_files,
            cwd=v8_path)
        for ar_file in ar_files:
            ar_file_canon = ar_file if case_sensitive else ar_file.lower()
            os.rename(os.path.join(dest_obj_dn, ar_file_canon), os.path.join(dest_obj_dn, "{}.{}".format(1 + j, ar_file)))
            j += 1

    file_groups = [] # [(file, size)]
    size = 0
    for fn in sorted(glob.glob(os.path.join(dest_obj_dn, "*"))):
        fsize = os.stat(fn).st_size
        if not file_groups or size + fsize >= args.max_file_size:
            file_groups.append([])
            size = 0
        file_groups[-1].append(os.path.relpath(fn, dest_path))
        size += fsize

    dest_stem, dest_ext = os.path.splitext(dest_fn)
    for fn in glob.glob(os.path.join(dest_path, "lib*.a")):
        os.unlink(fn)

    dest_fns = []
    for i, files in enumerate(file_groups):
        if len(file_groups) == 1:
            dest_fn = "{}{}".format(dest_stem, dest_ext)
        else:
            dest_fn = "{}-{}{}".format(dest_stem, i, dest_ext)

        dest_fns.append(os.path.relpath(dest_fn, dest_path))
        subprocess_check_call(
            [
                ar_path,
                "qsc",
                os.path.relpath(dest_fn, dest_path),
            ] + files,
            cwd=dest_path)

    with open(os.path.join(dest_path, "libmanifest"), "wt") as f:
        for dest_fn in dest_fns:
            print(dest_fn, file=f)

def copy_libcxx(build_path, dest_path):
    """Copies Chromium's libc++ and libc++abi next to libv8.

    They are named *-cr.a to avoid confusion with the system library.
    Ninja produces thin archives, which only reference object files, so
    we create regular archives from the members.
    """
    ar_path = os.path.abspath(os.path.join(v8_path, "third_party/llvm-build/Release+Asserts/bin/llvm-ar"))

    if args.os == "windows":
        # libc++ is a source_set, so there is no archive to copy. libc++abi
        # is for the Itanium ABI, while Windows uses the MSVC ABI. Named
        # c++-cr.lib, since clang's MSVC driver maps -lc++-cr to that.
        objs = sorted(glob.glob(os.path.join(build_path, "obj", "buildtools", "third_party", "libc++", "libc++", "*.obj")))
        if not objs:
            raise RuntimeError("libc++ object files not found in {}".format(build_path))
        lib_tool(["/out:" + os.path.join(dest_path, "c++-cr.lib")] + objs, build_path)
        return

    for name in ("libc++", "libc++abi"):
        src = os.path.join(build_path, "obj", "buildtools", "third_party", name, name + ".a")
        dest = os.path.join(dest_path, name + "-cr.a")

        members = subprocess_check_output_text([ar_path, "t", src], cwd=build_path).splitlines()
        if not members:
            raise RuntimeError("archive is empty: {}".format(src))

        if os.path.exists(dest):
            os.unlink(dest)
        subprocess_check_call([ar_path, "qcs", dest] + members, cwd=build_path)

# Target triples of Chromium's Clang runtime libraries.
LINUX_TRIPLES = {"amd64": "x86_64-unknown-linux-gnu", "arm64": "aarch64-unknown-linux-gnu"}

def copy_builtins(dest_path):
    """Copies Clang's compiler-rt builtins next to libv8.

    V8 can call builtins, like __extendhfsf2 for Float16, that libgcc
    only provides from GCC 12. Chromium links these instead of libgcc,
    and so do we, for a known minimum.
    """
    clang_path = llvm_tool("clang")
    src = subprocess_check_output_text([clang_path, "--target=" + LINUX_TRIPLES[args.arch], "--rtlib=compiler-rt", "-print-libgcc-file-name"], stderr=subprocess.DEVNULL).strip()
    if not os.path.isfile(src):
        raise RuntimeError("compiler-rt builtins not found: {}".format(src))

    shutil.copyfile(src, os.path.join(dest_path, "libclang_rt.builtins-cr.a"))

# Object files per archive on Windows. The COFF symbol index, with long
# MSVC-mangled names, added 17-83% to the object file sizes, depending
# on the objects. Archives larger than WINDOWS_MAX_LIB_SIZE are split
# again, to stay below GitHub's 50 MB warning.
WINDOWS_MAX_FILE_SIZE = int(30e6)
WINDOWS_MAX_LIB_SIZE = int(45e6)

def lib_tool(lib_args, cwd):
    """Runs lld-link as lib.exe, using a response file for the arguments.

    Chromium's Windows LLVM has no llvm-ar, and the command line is
    limited to 32 KiB.
    """
    rsp = os.path.join(cwd, "v8go-lib.rsp")
    with open(rsp, "wt") as f:
        f.write("\n".join('"{}"'.format(a) for a in lib_args))
    try:
        return subprocess_check_output_text([llvm_tool("lld-link"), "/lib", "@" + rsp], cwd=cwd)
    finally:
        os.unlink(rsp)

def split_windows_lib(build_path, dest_path):
    """Archives the monolith's object files into v8-N.lib files.

    The monolith's members are paths to object files in the build
    directory, so we archive those, rather than extracting the archive.
    clang's MSVC driver maps -lv8-0 to v8-0.lib.
    """
    monolith = lib_tool(["/list", os.path.join("obj", "v8_monolith.lib")], build_path).split()
    missing = [m for m in monolith if not os.path.isfile(os.path.join(build_path, m))]
    if not monolith or missing:
        raise RuntimeError("v8_monolith.lib members not found in {}: {}".format(build_path, missing[:5]))
    inputs = monolith

    groups = []
    size = 0
    for obj in inputs:
        fsize = os.stat(os.path.join(build_path, obj)).st_size
        if not groups or size + fsize >= WINDOWS_MAX_FILE_SIZE:
            groups.append([])
            size = 0
        groups[-1].append(obj)
        size += fsize

    for fn in glob.glob(os.path.join(dest_path, "v8-*.lib")):
        os.unlink(fn)
    dest_fns = []
    while groups:
        objs = groups.pop(0)
        dest_fn = "v8-{}.lib".format(len(dest_fns))
        lib_tool(["/out:" + os.path.join(dest_path, dest_fn)] + objs, build_path)
        if os.stat(os.path.join(dest_path, dest_fn)).st_size > WINDOWS_MAX_LIB_SIZE and len(objs) > 1:
            os.unlink(os.path.join(dest_path, dest_fn))
            groups[:0] = [objs[:len(objs) // 2], objs[len(objs) // 2:]]
            continue
        dest_fns.append(dest_fn)

    # The pieces must have the same members as the monolith.
    pieces = []
    for dest_fn in dest_fns:
        pieces += lib_tool(["/list", os.path.join(dest_path, dest_fn)], build_path).split()
    if sorted(monolith) != sorted(pieces):
        raise RuntimeError("split .lib members differ from v8_monolith.lib: {} vs {}".format(len(pieces), len(monolith)))

    with open(os.path.join(dest_path, "libmanifest"), "wt") as f:
        for dest_fn in dest_fns:
            print(dest_fn, file=f)

def allocate_disjoint_files(ar_files, case_sensitive=True):
    ar_file_counts = {} # file -> count
    for ar_file in ar_files:
        ar_file_counts[ar_file] = ar_file_counts.get(ar_file, 0) + 1
    ar_file_counts = list(ar_file_counts.items())
    ar_file_counts.sort(key=lambda item: -item[1])

    ar_file_groups = [] # [(index, files)]
    while ar_file_counts:
        canon_file_set = {} # canon file -> (file, count)
        file_set = set()
        max_count = 0
        for ar_file, count in ar_file_counts:
            ar_file_canon = ar_file if case_sensitive else ar_file.lower()
            if ar_file_canon in canon_file_set: continue
            canon_file_set[ar_file_canon] = (ar_file, count)
            file_set.add(ar_file)
            max_count = max(max_count, count)

        ar_file_counts = [(ar_file, count) for ar_file, count in ar_file_counts if ar_file not in file_set]
        groups = [(i, []) for i in range(max_count)]
        for ar_file, count in canon_file_set.values():
            for i in range(count):
                groups[i][1].append(ar_file)
        ar_file_groups.extend(groups)

    return ar_file_groups

def main():
    v8deps()
    disable_crel()
    if use_sysroot():
        install_sysroots()

    gn_path = os.path.join(tools_path, "gn" + (".bat" if is_windows else ""))
    assert(os.path.exists(gn_path))
    ninja_path = os.path.join(tools_path, "ninja" + (".bat" if is_windows else ""))
    assert(os.path.exists(ninja_path))

    build_path = os.path.join(deps_path, ".build", os_arch())

    gnargs = build_gn_args()

    subprocess_check_call([gn_path, "gen", build_path, "--args=" + gnargs.replace('\n', ' ')], cwd=v8_path)
    ninja_targets = ["v8_monolith", "libc++"]
    if args.os != "windows":
        ninja_targets.append("libc++abi")
    subprocess_check_call([ninja_path, "-v", "-C", build_path] + ninja_targets, cwd=v8_path)

    dest_path = os.path.join(deps_path, os_arch())
    if args.os == "windows":
        os.makedirs(dest_path, exist_ok=True)
        split_windows_lib(build_path, dest_path)
        copy_libcxx(build_path, dest_path)
        return

    dest_obj_dn = os.path.join(dest_path, "obj")
    try:
        split_ar(
            os.path.join(build_path, "obj/libv8_monolith.a"),
            os.path.join(dest_path, "libv8.a"),
            dest_obj_dn)
        copy_libcxx(build_path, dest_path)
        if args.os == "linux":
            copy_builtins(dest_path)
        if use_sysroot():
            check_sysroot_symbols(dest_path)
    finally:
        if os.path.exists(dest_obj_dn):
            shutil.rmtree(dest_obj_dn)

if __name__ == "__main__":
    main()
