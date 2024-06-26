#!/usr/bin/env python3
"""
Modifies a Keep-A-Changelog file.
"""

import argparse
import contextlib
import datetime
import os
import re
import sys

VERSION_HEADING_RE = re.compile(r'^## \[v(\d+)\.(\d+)\.(\d+)\]')

def get_latest_version(f):
    for line in f:
        m = VERSION_HEADING_RE.match(line)
        if m: return m.groups()

def add_versions(v, d):
    zero = False
    for vv, dd in zip(v, d):
        if zero:
            yield 0
        else:
            iv = int(vv)
            id = int(dd)
            yield iv + id
            if id: zero = True

@contextlib.contextmanager
def stdout_or_inplace(path, inplace):
    if not inplace:
        yield sys.stdout
        return

    with open(path + '.new', 'wt') as f:
        try:
            yield f
            os.rename(path + '.new', path)
        except:
            os.unlink(path + '.new')
            raise

def main():
    argp = argparse.ArgumentParser(description=__doc__)
    argp.add_argument('--changed', default=[], action='append', help='append a line to the unreleased Changed section')
    argp.add_argument('--unchanged', default=[], metavar='REGEX', action='append', help='removes matching lines from the unreleased Changed section')
    argp.add_argument('--inplace', '-i', action='store_true', help='modify the original file instead of writing to stdout')
    argp.add_argument('--release', metavar='[+]X.Y.Z', help='make a release, with an absolute or relative version number')
    argp.add_argument('file', type=argparse.FileType('r'), help='file to read')
    args = argp.parse_args()

    unchanged_re = re.compile('^(?:' + '|'.join('(?:' + s + ')' for s in args.unchanged) + ')$')

    make_release = args.release
    if make_release and make_release.startswith('+'):
        make_release = '.'.join(str(i) for i in add_versions(get_latest_version(args.file), make_release[1:].split('.')))
        args.file.seek(0)

    with stdout_or_inplace(args.file.name, args.inplace) as outf:
        in_unreleased = False
        old_version = None
        for line in args.file:
            line = line.rstrip(os.linesep)

            if line == '## [Unreleased]':
                in_unreleased = True
                break
            elif line.startswith('## [v'):
                old_version = line
                break

            print(line, file=outf)

        if make_release:
            print('## [Unreleased]', file=outf)
            print(file=outf)
            print('### Changed', file=outf)
            print(file=outf)
            print('## [v{}] - {}'.format(make_release, datetime.datetime.utcnow().strftime('%Y-%m-%d')), file=outf)
        elif not in_unreleased:
            print('## [Unreleased]', file=outf)
        else:
            print(line, file=outf)

        in_changed = False
        if not old_version:
            for line in args.file:
                line = line.rstrip(os.linesep)

                if line == '### Changed':
                    in_changed = True
                    break
                elif line.startswith('## [v'):
                    old_version = line
                    break

                print(line, file=outf)

        if not in_changed:
            if not in_unreleased:
                print(file=outf)
            print('### Changed', file=outf)
        else:
            print(line, file=outf)

        if not old_version:
            for line in args.file:
                line = line.rstrip(os.linesep)

                if not line:
                    break

                if unchanged_re.match(line):
                    continue

                print(line, file=outf)

        for chline in args.changed:
            print(chline, file=outf)
        print(file=outf)

        if old_version:
            print(old_version, file=outf)

        for line in args.file:
            print(line.rstrip(os.linesep), file=outf)

if __name__ == '__main__':
    main()
