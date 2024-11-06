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

// SetAccessorPropertyCallback is a simplified version of SetAccessorProperty
// that automatically create the [FunctionTemplate] for the callbacks when the
// caller doesn't need the template(s).
func (o *ObjectTemplate) SetAccessorPropertyCallback(
	key string,
	get FunctionCallbackWithError,
	set FunctionCallbackWithError,
	attributes PropertyAttribute,
) {
	var (
		getter *FunctionTemplate
		setter *FunctionTemplate
	)
	if get != nil {
		getter = NewFunctionTemplateWithError(o.iso, get)
	}
	if set != nil {
		setter = NewFunctionTemplateWithError(o.iso, set)
	}
	o.SetAccessorProperty(key, getter, setter, attributes)
}

// InternalFieldCount returns the number of internal fields that instances of this
// template will have.
func (o *ObjectTemplate) InternalFieldCount() uint32 {
	return uint32(C.ObjectTemplateInternalFieldCount(o.ptr))
}

func (o *ObjectTemplate) apply(opts *contextOptions) {
	opts.gTmpl = o
}
