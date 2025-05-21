// Copyright 2020 Roger Chapman and the v8go contributors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package v8go

// #include <stdlib.h>
// #include "object_template.h"
import "C"
import (
	"errors"
	"runtime"
	"unsafe"
)

// PropertyAttribute are the attribute flags for a property on an Object.
// Typical usage when setting an Object or TemplateObject property, and
// can also be validated when accessing a property.
type PropertyAttribute uint8

const (
	// None.
	None PropertyAttribute = 0
	// ReadOnly, ie. not writable.
	ReadOnly PropertyAttribute = 1 << iota
	// DontEnum, ie. not enumerable.
	DontEnum
	// DontDelete, ie. not configurable.
	DontDelete
)

// ObjectTemplate is used to create objects at runtime.
// Properties added to an ObjectTemplate are added to each object created from the ObjectTemplate.
type ObjectTemplate struct {
	*template
}

// NewObjectTemplate creates a new ObjectTemplate.
// The *ObjectTemplate can be used as a v8go.ContextOption to create a global object in a Context.
func NewObjectTemplate(iso *Isolate) *ObjectTemplate {
	if iso == nil {
		panic("nil Isolate argument not supported")
	}

	tmpl := &template{
		ptr: C.NewObjectTemplate(iso.ptr),
		iso: iso,
	}
	runtime.SetFinalizer(tmpl, (*template).finalizer)
	return &ObjectTemplate{tmpl}
}

// NewInstance creates a new Object based on the template.
func (o *ObjectTemplate) NewInstance(ctx *Context) (*Object, error) {
	if ctx == nil {
		return nil, errors.New("v8go: Context cannot be <nil>")
	}

	rtn := C.ObjectTemplateNewInstance(o.ptr, ctx.ptr)
	runtime.KeepAlive(o)
	return objectResult(ctx, rtn)
}

// SetInternalFieldCount sets the number of internal fields that instances of this
// template will have.
func (o *ObjectTemplate) SetInternalFieldCount(fieldCount uint32) {
	C.ObjectTemplateSetInternalFieldCount(o.ptr, C.int(fieldCount))
}

// SetAccessorProperty creates a named accessor property, i.e., a property that
// is implemented as a function call. Arguments get and set represents the
// getter and setter, and can both be nil.
//
// Note: The [ReadOnly] should not be used with a readonly property. If set is
// nil, the property will be readonly, and passing [None] is a sensible default.
//
// This corresponds to ObjectTemplate::SetAccessorProperty in the C++ API.
func (o *ObjectTemplate) SetAccessorProperty(
	key string,
	get *FunctionTemplate,
	set *FunctionTemplate,
	attributes PropertyAttribute,
) {
	ckey := C.CString(key)
	defer C.free(unsafe.Pointer(ckey))
	var (
		getter C.TemplatePtr
		setter C.TemplatePtr
	)
	if get != nil {
		getter = get.ptr
	}
	if set != nil {
		setter = set.ptr
	}
	C.ObjectTemplateSetAccessorProperty(o.ptr, ckey, getter, setter, C.int(attributes))
}

// InternalFieldCount returns the number of internal fields that instances of this
// template will have.
func (o *ObjectTemplate) InternalFieldCount() uint32 {
	return uint32(C.ObjectTemplateInternalFieldCount(o.ptr))
}

func (o *ObjectTemplate) apply(opts *contextOptions) {
	opts.gTmpl = o
}

// MarkAsUndetectable marks object instances of the template as undetectable. Undetectable
// objects behave like undefined, but you can access properties defined on undetectable
// objects.
//
// Note: Undetectable objects MUST have a CallAsFunctionHandler, see
// [ObjectTemplate.SetCallAsFunctionHandler]
func (o *ObjectTemplate) MarkAsUndetectable() {
	C.ObjectTemplateMarkAsUndetectable(o.ptr)
}

// SetCallAsFunctionHandler sets the callback to be used when calling instances created
// from this template. If no callback is set, instances behave like normal JavaScript
// objects that cannot be called as a function.
func (o *ObjectTemplate) SetCallAsFunctionHandler(callback FunctionCallbackWithError) {
	if callback == nil {
		panic("nil callback argument not supported")
	}
	cbref := o.iso.registerCallback(callback)
	C.ObjectTemplateSetCallAsFunctionHandler(
		o.ptr,
		C.int(cbref),
	)
}
