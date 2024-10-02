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
is_clang = args.clang if args.clang is not None else args.os != "linux"

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
use_custom_libcxx=false
use_sysroot=false
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
"""

def v8deps():
    spec = "solutions = %s\n" % gclient_sln
    spec += "target_os = [%r]" % (v8_os(),)
    env = os.environ.copy()
    env["PATH"] = tools_path + os.pathsep + env["PATH"]
    subprocess_check_call(["gclient", "sync", "--delete_unversioned_trees", "--no-history", "--spec", spec],
                        cwd=deps_path,
                        env=env)

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
    return args.os.replace('darwin', 'mac')

def v8_arch():
    if args.arch == "amd64":
        return "x64"
    return args.arch

def apply_mingw_patches():
    v8_build_path = os.path.join(v8_path, "build")
    apply_patch("0000-add-mingw-main-code-changes", v8_path)
    apply_patch("0001-add-mingw-toolchain", v8_build_path)
    update_last_change()
    zlib_path = os.path.join(v8_path, "third_party", "zlib")
    zlib_src_gn = os.path.join(deps_path, os_arch(), "zlib.gn")
    zlib_dst_gn = os.path.join(zlib_path, "BUILD.gn")
    shutil.copy(zlib_src_gn, zlib_dst_gn)

def apply_patch(patch_name, working_dir):
    patch_path = os.path.join(deps_path, os_arch(), patch_name + ".patch")
    subprocess_check_call(["git", "apply", "-v", patch_path], cwd=working_dir)

def update_last_change():
    out_path = os.path.join(v8_path, "build", "util", "LASTCHANGE")
    subprocess_check_call(["python", "build/util/lastchange.py", "-o", out_path], cwd=v8_path)

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
    if is_windows:
        apply_mingw_patches()

    gn_path = os.path.join(tools_path, "gn")
    assert(os.path.exists(gn_path))
    ninja_path = os.path.join(tools_path, "ninja" + (".exe" if is_windows else ""))
    assert(os.path.exists(ninja_path))

    build_path = os.path.join(deps_path, ".build", os_arch())

    gnargs = build_gn_args()

    subprocess_check_call([gn_path, "gen", build_path, "--args=" + gnargs.replace('\n', ' ')], cwd=v8_path)
    subprocess_check_call([ninja_path, "-v", "-C", build_path, "v8_monolith"], cwd=v8_path)

    dest_path = os.path.join(deps_path, os_arch())
    dest_obj_dn = os.path.join(dest_path, "obj")
    try:
        split_ar(
            os.path.join(build_path, "obj/libv8_monolith.a"),
            os.path.join(dest_path, "libv8.a"),
            dest_obj_dn)
    finally:
        if os.path.exists(dest_obj_dn):
            shutil.rmtree(dest_obj_dn)

if __name__ == "__main__":
    main()
