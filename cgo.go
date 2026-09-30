// Copyright 2019 Roger Chapman and the v8go contributors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package v8go

//go:generate clang-format -i --verbose -style=Chromium v8go.h v8go.cc

// V8 is built with Chromium's libc++, which we must also use. The
// CGO_CXXFLAGS environment variable must contain -nostdinc++, since it
// is not allowed in #cgo directives.

// #cgo CXXFLAGS: -fno-rtti -fPIC -std=c++20 -stdlib=libc++ -I${SRCDIR}/deps/include -Wall
// #cgo CXXFLAGS: -I${SRCDIR}/deps/include_libcxx -I${SRCDIR}/deps/include_libcxxabi
// #cgo CXXFLAGS: -D_LIBCPP_HARDENING_MODE=_LIBCPP_HARDENING_MODE_EXTENSIVE
// #cgo CXXFLAGS: -DV8_COMPRESS_POINTERS -DV8_31BIT_SMIS_ON_64BIT_ARCH -DV8_ENABLE_SANDBOX
// #cgo CXXFLAGS: -DV8_DEPRECATION_WARNINGS -DV8_IMMINENT_DEPRECATION_WARNINGS
import "C"
