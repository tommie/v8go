// Copyright 2021 the v8go contributors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package v8go

// TODO: Can v8go.h be removed?

// #include <stdlib.h>
// #include "script_compiler.h"
import "C"

import "unsafe"

type CompileMode C.int

var (
	CompileModeDefault = CompileMode(C.ScriptCompilerNoCompileOptions)
	CompileModeEager   = CompileMode(C.ScriptCompilerEagerCompile)
)

type CompilerCachedData struct {
	Bytes    []byte
	Rejected bool
}

func CompileModule(ctx *Context, source, origin string) (*Value, error) {
	cSource := C.CString(source)
	cOrigin := C.CString(origin)
	defer C.free(unsafe.Pointer(cSource))
	defer C.free(unsafe.Pointer(cOrigin))

	return valueResult(ctx, C.ScriptCompilerCompileModule(ctx.ptr, cSource, cOrigin))
}
