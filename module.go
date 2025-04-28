package v8go

// #include "module.h"
import "C"

type Module struct {
	ptr *C.m_module
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

func (m Module) InstantiateModule(ctx *Context) error {
	err := C.ModuleInstantiateModule(ctx.ptr, m.ptr)
	if err.msg == nil {
		return nil
	}
	return newJSError(err)
}
