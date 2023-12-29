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
    for (os, arch), libs in os_arch_libs:
        yield "// #cgo {},{} LDFLAGS: {}".format(
            os,
            arch,
            " ".join("-l{}".format(lib.replace(".a", "").replace("libv8", "v8")) for lib in libs),
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
