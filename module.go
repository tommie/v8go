package v8go

// #include <stdlib.h>
// #include "module.h"
import "C"
import "unsafe"

// Module represents an ECMAScript Module (ESM). A module is obtained from
// [CompileModule]. Before a module can be used, it must be instantiated by
// calling [Module.InstantiateModule], after which it can be evaluated with
// [Module.Evaluate].
type Module struct {
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
) *C.m_module {
	defer C.free(unsafe.Pointer(buf))
	spec := C.GoString(buf)

	ctx := getContext(ctxref)
	if res, err := ctx.moduleResolver.ResolveModule(ctx, spec, nil); err == nil {
		return res.ptr
	}
	return nil
}

func (m Module) InstantiateModule(
	ctx *Context,
	resolver ResolveModuler,
) error {
	ctx.moduleResolver = resolver

	err := C.ModuleInstantiateModule(ctx.ptr, m.ptr, nil, nil)
	if err.msg == nil {
		return nil
	}
	return newJSError(err)
}
