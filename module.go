package v8go

// #include <stdlib.h>
// #include "module.h"
import "C"
import (
	"fmt"
	"unsafe"
)

// Module represents an ECMAScript Module (ESM). A module is obtained from
// [CompileModule]. Before a module can be used, it must be instantiated by
// calling [Module.InstantiateModule], after which it can be evaluated with
// [Module.Evaluate].
type Module struct {
	iso *C.v8Isolate
	ptr *C.m_module
}

type FixedArray struct{}

type ResolveModuler interface {
	ResolveModule(ctx *Context, spec string, referrer *Module) (*Module, error)
}

// Evaluate evaluates the module.
//
// The returned valus is a promise. If the module is evaluated synchronously,
// the promise will have settled upon return.
//
// If evaluating the script fails, the promise will be rejected; but if an
// imported module cannot be evaluated, Evaluate will return an error.
func (m Module) Evaluate(ctx *Context) (*Value, error) {
	retVal := C.ModuleEvaluate(ctx.ptr, m.ptr)
	return valueResult(ctx, retVal)
}

//export resolveModuleCallback
func resolveModuleCallback(
	ctxref int,
	buf *C.char,
	referrer *C.m_module,
) (*C.m_module, C.ValuePtr) {
	defer C.free(unsafe.Pointer(buf))
	spec := C.GoString(buf)

	ctx := getContext(ctxref)
	ref := &Module{ptr: referrer}
	res, err := ctx.moduleResolver.ResolveModule(ctx, spec, ref)
	if err == nil {
		return res.ptr, nil
	} else {
		err = fmt.Errorf("cannot resolve module '%s': %w", spec, err)
		return nil, NewError(ctx.iso, err.Error()).Value.ptr
	}
}

func (m Module) InstantiateModule(ctx *Context, resolver ResolveModuler) error {
	ctx.moduleResolver = resolver

	err := C.ModuleInstantiateModule(ctx.ptr, m.ptr, nil, nil)
	if err.msg == nil {
		return nil
	}
	return newJSError(err)
}

func (m Module) ScriptID() int {
	return int(C.ModuleScriptId(m.ptr))
}

func (m Module) IsSourceTextModule() bool {
	return bool(C.ModuleIsSourceTextModule(m.ptr))
}
