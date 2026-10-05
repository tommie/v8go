#!/bin/sh
# Checks that a module using v8go builds and runs after go mod vendor.
#
# go mod vendor only copies directories of imported packages, so the
# header directories must be imported from cgo.go. See tommie/v8go#116.
#
# All v8go modules are replaced with this work tree, so nothing is
# downloaded, and the test covers uncommitted changes.

set -eu

src=$(cd "$(dirname "$0")/.." && pwd)
work=$(mktemp -d)
trap 'rm -r "$work"' EXIT

export GOWORK=off GOFLAGS=

cd "$work"
cat >go.mod <<EOF
module example.com/vendortest

go 1.19

require github.com/tommie/v8go v0.0.0

replace github.com/tommie/v8go => $src
EOF
for dir in "$src"/deps/*_*/; do
	[ -f "$dir/go.mod" ] || continue
	echo "replace github.com/tommie/v8go/deps/$(basename "$dir") => $dir" >>go.mod
done

cat >main.go <<'EOF'
package main

import (
	"fmt"

	v8 "github.com/tommie/v8go"
)

func main() {
	ctx := v8.NewContext()
	defer ctx.Isolate().Dispose()
	defer ctx.Close()

	val, err := ctx.RunScript("3 + 4", "value.js")
	if err != nil {
		panic(err)
	}
	fmt.Println(val)
}
EOF

go mod tidy
go mod vendor

# Fail early with a clear message, rather than a compiler error.
for f in deps/include/v8-template.h deps/include/cppgc/internal/api-constants.h deps/include_libcxx/vector deps/include_libcxx/__config_site deps/include_libcxxabi/cxxabi.h; do
	if [ ! -f "vendor/github.com/tommie/v8go/$f" ]; then
		echo "$f was not vendored" >&2
		exit 1
	fi
done

out=$(go run -mod=vendor .)
if [ "$out" != 7 ]; then
	echo "unexpected output: $out" >&2
	exit 1
fi
echo "vendored build OK"
