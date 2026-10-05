package v8go

// #include <stdlib.h>
// #include "module.h"
import "C"

import (
	"errors"
	"fmt"
	"unsafe"
)

// TODO: Support dynamic import() through
// Isolate::SetHostImportModuleDynamicallyCallback, and import.meta
// through Isolate::SetHostInitializeImportMetaObjectCallback. Both are
// isolate-wide, so they don't affect the ModuleResolver passed to
// Module.Instantiate.

// Module is a compiled ECMAScript module (ESM). It is created with
// [Isolate.CompileModule], must be linked with [Module.Instantiate], and
// then run with [Module.Evaluate].
//
// A Module belongs to its Isolate, and is released when the Isolate is
// disposed.
type Module struct {
	ptr C.ModulePtr
	iso *Isolate
}

// ModuleStatus is the state of a Module.
type ModuleStatus int

// ModuleStatus values, in the order a Module normally passes through
// them. They mirror v8::Module::Status, which module.cc asserts.
const (
	ModuleUninstantiated ModuleStatus = 0
	ModuleInstantiating  ModuleStatus = 1
	ModuleInstantiated   ModuleStatus = 2
	ModuleEvaluating     ModuleStatus = 3
	ModuleEvaluated      ModuleStatus = 4
	ModuleErrored        ModuleStatus = 5
)

// String returns a name for the status.
func (s ModuleStatus) String() string {
	switch s {
	case ModuleUninstantiated:
		return "uninstantiated"
	case ModuleInstantiating:
		return "instantiating"
	case ModuleInstantiated:
		return "instantiated"
	case ModuleEvaluating:
		return "evaluating"
	case ModuleEvaluated:
		return "evaluated"
	case ModuleErrored:
		return "errored"
	default:
		return fmt.Sprintf("ModuleStatus(%d)", int(s))
	}
}

// ImportAttribute is an attribute in an import statement. E.g. this
// imports with the attribute type="json":
//
//	import data from "./data.json" with { type: "json" };
type ImportAttribute struct {
	Key   string
	Value string

	// SourceOffset is the offset of the key in the importing module's
	// source.
	SourceOffset int
}

// A ModuleResolver resolves the imports of a module.
type ModuleResolver interface {
	// ResolveModule returns the module the specifier refers to, as
	// imported by referrer.
	//
	// The resolver is responsible for caching: a module imported from
	// several places must resolve to the same Module each time, or it is
	// evaluated once per import. Doing so also makes cyclic imports work.
	//
	// If the error implements [ValueError], its value is thrown. Either
	// way, the error returned by Instantiate wraps it. Like
	// function callbacks, a panic aborts the process, since it would
	// unwind through V8.
	ResolveModule(ctx *Context, specifier string, attrs []ImportAttribute, referrer *Module) (*Module, error)
}

// ModuleResolverFunc is a function that implements [ModuleResolver].
type ModuleResolverFunc func(ctx *Context, specifier string, attrs []ImportAttribute, referrer *Module) (*Module, error)

// ResolveModule calls f.
func (f ModuleResolverFunc) ResolveModule(ctx *Context, specifier string, attrs []ImportAttribute, referrer *Module) (*Module, error) {
	return f(ctx, specifier, attrs, referrer)
}

// ErrNoModuleResolver is matched by errors.Is for errors from
// [Module.Instantiate], if a module has imports, but no resolver was
// given.
var ErrNoModuleResolver = errors.New("no module resolver")

// moduleInstantiation is the state of an ongoing Module.Instantiate.
type moduleInstantiation struct {
	resolver ModuleResolver

	// err is the last error from resolving a module. It is returned by
	// JSError.Unwrap, since the error only reaches V8 as an exception.
	err error
}

// CompileModule compiles source as an ECMAScript module. The origin
// (a.k.a. filename) is used in stack traces. If an error occurs, it is a
// *JSError.
func (i *Isolate) CompileModule(source, origin string) (*Module, error) {
	cSource := C.CString(source)
	cOrigin := C.CString(origin)
	defer C.free(unsafe.Pointer(cSource))
	defer C.free(unsafe.Pointer(cOrigin))

	rtn := C.IsolateCompileModule(i.ptr, cSource, cOrigin)
	if rtn.ptr == nil {
		return nil, newJSError(rtn.error)
	}
	m := &Module{ptr: rtn.ptr, iso: i}
	i.modules[m.ptr] = m
	return m, nil
}

// Status returns the module's status.
func (m *Module) Status() ModuleStatus {
	return ModuleStatus(C.ModuleGetStatus(m.iso.ptr, m.ptr))
}

// ScriptID returns the V8 script ID of the module.
func (m *Module) ScriptID() int {
	return int(C.ModuleScriptId(m.iso.ptr, m.ptr))
}

// Instantiate links the module, and its imports, in ctx. Imports are
// resolved with resolver, which may be nil if there are no imports. A
// module can only be instantiated once, and it then belongs to ctx.
//
// If an error occurs, it is a *JSError. If resolving a module failed,
// the resolver's error can be matched with errors.Is and errors.As.
func (m *Module) Instantiate(ctx *Context, resolver ModuleResolver) error {
	m.checkIsolate(ctx)

	// V8 calls the resolver synchronously, without a data pointer, so it
	// is passed through the Context. The previous state is restored, in
	// case a resolver instantiates another module.
	inst := &moduleInstantiation{resolver: resolver}
	prev := ctx.instantiation
	ctx.instantiation = inst
	defer func() { ctx.instantiation = prev }()

	rtn := C.ModuleInstantiate(ctx.ptr, m.ptr)
	if rtn.msg == nil {
		return nil
	}
	err := newJSError(rtn).(*JSError)
	if err.cause == nil {
		err.cause = inst.err
	}
	return err
}

// Evaluate runs the module, and its imports. The module must have been
// instantiated in ctx.
//
// It returns a promise. A module without top-level await settles it
// before returning. Exceptions thrown by the module reject the promise,
// rather than returning an error.
func (m *Module) Evaluate(ctx *Context) (*Promise, error) {
	m.checkIsolate(ctx)
	v, err := valueResult(ctx, C.ModuleEvaluate(ctx.ptr, m.ptr))
	if err != nil {
		return nil, err
	}
	return v.AsPromise()
}

// Namespace returns the module namespace object, which holds the
// exports of the module. The module must have been instantiated in ctx.
// Before evaluation, the exports exist, but are not initialized.
//
// See https://tc39.es/ecma262/#sec-module-namespace-exotic-objects.
func (m *Module) Namespace(ctx *Context) (*Object, error) {
	m.checkIsolate(ctx)
	return objectResult(ctx, C.ModuleGetNamespace(ctx.ptr, m.ptr))
}

func (m *Module) checkIsolate(ctx *Context) {
	if ctx.iso != m.iso {
		panic("attempted to use a module in a context that belongs to a different isolate")
	}
}

//export goResolveModule
func goResolveModule(
	ctxref int,
	specifier *C.char,
	cattrs *C.ModuleImportAttribute,
	nattrs C.int,
	referrer C.ModulePtr,
) (C.ModulePtr, C.ValuePtr) {
	ctx := getContext(ctxref)
	spec := C.GoString(specifier)

	var attrs []ImportAttribute
	if nattrs > 0 {
		attrs = make([]ImportAttribute, nattrs)
		for i, a := range unsafe.Slice(cattrs, nattrs) {
			attrs[i] = ImportAttribute{
				Key:          C.GoString(a.key),
				Value:        C.GoString(a.value),
				SourceOffset: int(a.source_offset),
			}
		}
	}

	mod, err := resolveModule(ctx, spec, attrs, ctx.iso.modules[referrer])
	if err != nil {
		ctx.instantiation.err = fmt.Errorf("cannot resolve module %q: %w", spec, err)
		if verr, ok := err.(ValueError); ok {
			return nil, verr.value().ptr
		}
		return nil, NewError(ctx.iso, ctx.instantiation.err.Error()).ptr
	}
	return mod.ptr, nil
}

func resolveModule(ctx *Context, spec string, attrs []ImportAttribute, referrer *Module) (*Module, error) {
	if ctx.instantiation.resolver == nil {
		return nil, ErrNoModuleResolver
	}
	mod, err := ctx.instantiation.resolver.ResolveModule(ctx, spec, attrs, referrer)
	if err != nil {
		return nil, err
	}
	if mod == nil {
		return nil, errors.New("resolver returned no module")
	}
	if mod.iso != ctx.iso {
		return nil, errors.New("resolver returned a module from a different isolate")
	}
	return mod, nil
}
