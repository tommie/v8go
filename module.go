package v8go

// #include <stdlib.h>
// #include "module.h"
// #include "data.h"
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

// ImportAttributes represents the attributes for each module import.
//
// NOTE: ImportAttributes cannot be used AFTER ResolveModule returns
type ImportAttributes struct {
	fixedArray *C.v8goFixedArray
}

func (a ImportAttributes) All(ctx *Context) []ImportAttribute {
	l := int(C.FixedArrayLength(a.fixedArray, ctx.ptr)) / 3
	if l == 0 {
		return nil
	}
	res := make([]ImportAttribute, l)
	for i := 0; i < l; i++ {
		res[i] = a.get(ctx, i)
	}
	return res
}

func (a ImportAttributes) get(ctx *Context, i int) ImportAttribute {
	d1 := C.FixedArrayGet(a.fixedArray, ctx.ptr, C.int(i*3))
	d2 := C.FixedArrayGet(a.fixedArray, ctx.ptr, C.int(i*3+1))
	d3 := C.FixedArrayGet(a.fixedArray, ctx.ptr, C.int(i*3+2))
	defer C.DataRelease(d1)
	defer C.DataRelease(d2)
	defer C.DataRelease(d3)
	v1 := Value{ptr: C.DataAsValue(d1, ctx.ptr), ctx: ctx}
	v2 := Value{ptr: C.DataAsValue(d2, ctx.ptr), ctx: ctx}
	v3 := Value{ptr: C.DataAsValue(d3, ctx.ptr), ctx: ctx}
	return ImportAttribute{
		Key:      v1.String(),
		Value:    v2.String(),
		Location: int(v3.Int32()),
	}
}

// ImportAttribute represents a single import attribute in a module import
// statment. E.g., the following script has a single import attribute.
//
//	import foo from "foo.js" with { data: "value" }
type ImportAttribute struct {
	Key   string
	Value string
	// Location is the zero-based index in the string where the key is found
	Location int
}

type FixedArray struct{}

type ResolveModuler interface {
	ResolveModule(ctx *Context, spec string, attr ImportAttributes, referrer *Module) (*Module, error)
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
	importAttributes *C.v8goFixedArray,
	referrer *C.m_module,
) (*C.m_module, C.ValuePtr) {
	defer C.free(unsafe.Pointer(buf))
	spec := C.GoString(buf)

	ctx := getContext(ctxref)
	ref := &Module{ptr: referrer}
	res, err := ctx.moduleResolver.ResolveModule(ctx, spec, ImportAttributes{importAttributes}, ref)
	if err == nil {
		return res.ptr, nil
	} else {
		err = fmt.Errorf("cannot resolve module '%s': %w", spec, err)
		return nil, NewError(ctx.iso, err.Error()).Value.ptr
	}
}

func (m Module) InstantiateModule(ctx *Context, resolver ResolveModuler) error {
	ctx.moduleResolver = resolver

	err := C.ModuleInstantiateModule(ctx.ptr, m.ptr)
	if err.msg == nil {
		return nil
	}
	return newJSError(err)
}

func (m Module) ScriptID() int {
	return int(C.ModuleScriptId(m.ptr))
}

func (m Module) GetStatus() int {
	return int(C.ModuleGetStatus(m.ptr))
}

func (m Module) IsSourceTextModule() bool {
	return bool(C.ModuleIsSourceTextModule(m.ptr))
}

// GetModuleNamespace returns the module namespace. This is an exotic object
// containing the exports of the module.
//
// See also: https://tc39.es/ecma262/#sec-module-namespace-exotic-objects
func (m Module) GetModuleNamespace() *Value {
	var res = C.ModuleGetModuleNamespace(m.iso, m.ptr)
	return &Value{res, nil}
}
