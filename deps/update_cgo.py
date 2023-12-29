#!/usr/bin/env python3
import argparse
import glob
import os.path

argp = argparse.ArgumentParser()
argp.add_argument("--cgo-file", default="cgo.go")
argp.add_argument("--manifest-paths", default="deps/*_*/libmanifest")
args = argp.parse_args()

def get_libs(manifest_path):
    with open(manifest_path, "rt") as f:
        return f.read().splitlines()

def get_all_libs(manifest_glob):
    for manifest_path in glob.glob(manifest_glob):
        os_arch = os.path.basename(os.path.dirname(manifest_path))
        os_arch = os_arch.split("_", 1)
        libs = get_libs(manifest_path)
        yield os_arch, libs

def format_ldflags_libs(os_arch_libs):
    for (os, arch), libs in sorted(os_arch_libs):
        # Since libraries are split without caring about dependencies,
        # we need to make it a group.
        #
        # Go explicitly allows start-group: https://cs.opensource.google/go/go/+/master:src/cmd/go/internal/work/security.go;l=205;drc=bc9da01e9d7de25f04173f7122e09fe0996aaa05
        #
        # However, XCode ld(1) does not support it, but says it "will continually search a static library": https://keith.github.io/xcode-man-pages/ld.1.html
        start_group = " -Wl,--start-group" if os != "darwin" else ""
        end_group = " -Wl,--end-group" if os != "darwin" else ""
        yield "// #cgo {},{} LDFLAGS:{} {}{}".format(
            os,
            arch,
            start_group,
            " ".join("-l{}".format(lib.replace(".a", "").replace("libv8", "v8")) for lib in libs),
            end_group,
        )

def replace_section(lines, begin, end, new_lines):
    inside = False
    for line in lines:
        if line.strip() == end and inside:
            inside = False
        if not inside:
            yield line
        if line.strip() == begin:
            inside = True
            yield from new_lines

def replace_section_in_file(path, begin, end, new_lines):
    with open(path, "rt") as f:
        cgo_content = f.read()

    cgo_lines = list(replace_section(cgo_content.splitlines(), begin, end, new_lines))

    with open(path, "wt") as f:
        for line in cgo_lines:
            print(line, file=f)

def main():
    new_lines = list(format_ldflags_libs(get_all_libs(args.manifest_paths)))

    replace_section_in_file(
        args.cgo_file,
        "// // Begin Generated Libs",
        "// // End Generated Libs",
        new_lines,
    )


if __name__ == "__main__":
    main()
