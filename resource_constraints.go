// Copyright 2024 Roger Chapman and the v8go contributors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package v8go

// #include <stdlib.h>
// #include "resource_constraints.h"
import "C"

import (
	"runtime"
)

// ResourceConstraints represents V8 resource constraints that specify
// the limits of the runtime's memory use. You must set the heap size
// before initializing the VM - the size cannot be adjusted after the
// VM is initialized.
type ResourceConstraints struct {
	ptr C.ResourceConstraintsPtr
}

// NewResourceConstraints creates a new ResourceConstraints object.
func NewResourceConstraints() *ResourceConstraints {
	rc := &ResourceConstraints{
		ptr: C.NewResourceConstraints(),
	}
	runtime.SetFinalizer(rc, (*ResourceConstraints).finalizer)
	return rc
}

// ConfigureDefaultsFromHeapSize configures the constraints with reasonable
// default values based on the provided heap size limit. The heap size
// includes both the young and the old generation.
//
// initialHeapSizeInBytes: The initial heap size or zero. By default V8
// starts with a small heap and dynamically grows it to match the set of
// live objects. This may lead to ineffective garbage collections at startup
// if the live set is large. Setting the initial heap size avoids such
// garbage collections. Note that this does not affect young generation
// garbage collections.
//
// maximumHeapSizeInBytes: The hard limit for the heap size. When the heap
// size approaches this limit, V8 will perform series of garbage collections
// and invoke the NearHeapLimitCallback. If the garbage collections do not
// help and the callback does not increase the limit, then V8 will crash
// with V8::FatalProcessOutOfMemory.
func (rc *ResourceConstraints) ConfigureDefaultsFromHeapSize(initialHeapSizeInBytes, maximumHeapSizeInBytes uint64) {
	C.ResourceConstraintsConfigureDefaultsFromHeapSize(rc.ptr, C.size_t(initialHeapSizeInBytes), C.size_t(maximumHeapSizeInBytes))
}

// ConfigureDefaults configures the constraints with reasonable default
// values based on the capabilities of the current device the VM is running on.
//
// physicalMemory: The total amount of physical memory on the current device, in bytes.
// virtualMemoryLimit: The amount of virtual memory on the current device, in bytes,
// or zero, if there is no limit.
func (rc *ResourceConstraints) ConfigureDefaults(physicalMemory, virtualMemoryLimit uint64) {
	C.ResourceConstraintsConfigureDefaults(rc.ptr, C.uint64_t(physicalMemory), C.uint64_t(virtualMemoryLimit))
}

// SetStackLimit sets the address beyond which the VM's stack may not grow.
func (rc *ResourceConstraints) SetStackLimit(stackLimit uintptr) {
	C.ResourceConstraintsSetStackLimit(rc.ptr, C.uintptr_t(stackLimit))
}

// StackLimit gets the address beyond which the VM's stack may not grow.
func (rc *ResourceConstraints) StackLimit() uintptr {
	return uintptr(C.ResourceConstraintsStackLimit(rc.ptr))
}

// SetCodeRangeSizeInBytes sets the amount of virtual memory reserved for
// generated code. This is relevant for 64-bit architectures that rely on
// code range for calls in code.
func (rc *ResourceConstraints) SetCodeRangeSizeInBytes(limit uint64) {
	C.ResourceConstraintsSetCodeRangeSizeInBytes(rc.ptr, C.size_t(limit))
}

// CodeRangeSizeInBytes gets the amount of virtual memory reserved for
// generated code.
func (rc *ResourceConstraints) CodeRangeSizeInBytes() uint64 {
	return uint64(C.ResourceConstraintsCodeRangeSizeInBytes(rc.ptr))
}

// SetMaxOldGenerationSizeInBytes sets the maximum size of the old generation.
// When the old generation approaches this limit, V8 will perform series of
// garbage collections and invoke the NearHeapLimitCallback. If the garbage
// collections do not help and the callback does not increase the limit,
// then V8 will crash with V8::FatalProcessOutOfMemory.
func (rc *ResourceConstraints) SetMaxOldGenerationSizeInBytes(limit uint64) {
	C.ResourceConstraintsSetMaxOldGenerationSizeInBytes(rc.ptr, C.size_t(limit))
}

// MaxOldGenerationSizeInBytes gets the maximum size of the old generation.
func (rc *ResourceConstraints) MaxOldGenerationSizeInBytes() uint64 {
	return uint64(C.ResourceConstraintsMaxOldGenerationSizeInBytes(rc.ptr))
}

// SetMaxYoungGenerationSizeInBytes sets the maximum size of the young generation,
// which consists of two semi-spaces and a large object space. This affects
// frequency of Scavenge garbage collections and should be typically much
// smaller that the old generation.
func (rc *ResourceConstraints) SetMaxYoungGenerationSizeInBytes(limit uint64) {
	C.ResourceConstraintsSetMaxYoungGenerationSizeInBytes(rc.ptr, C.size_t(limit))
}

// MaxYoungGenerationSizeInBytes gets the maximum size of the young generation.
func (rc *ResourceConstraints) MaxYoungGenerationSizeInBytes() uint64 {
	return uint64(C.ResourceConstraintsMaxYoungGenerationSizeInBytes(rc.ptr))
}

// SetInitialOldGenerationSizeInBytes sets the initial size of the old generation.
func (rc *ResourceConstraints) SetInitialOldGenerationSizeInBytes(initialSize uint64) {
	C.ResourceConstraintsSetInitialOldGenerationSizeInBytes(rc.ptr, C.size_t(initialSize))
}

// InitialOldGenerationSizeInBytes gets the initial size of the old generation.
func (rc *ResourceConstraints) InitialOldGenerationSizeInBytes() uint64 {
	return uint64(C.ResourceConstraintsInitialOldGenerationSizeInBytes(rc.ptr))
}

// SetInitialYoungGenerationSizeInBytes sets the initial size of the young generation.
func (rc *ResourceConstraints) SetInitialYoungGenerationSizeInBytes(initialSize uint64) {
	C.ResourceConstraintsSetInitialYoungGenerationSizeInBytes(rc.ptr, C.size_t(initialSize))
}

// InitialYoungGenerationSizeInBytes gets the initial size of the young generation.
func (rc *ResourceConstraints) InitialYoungGenerationSizeInBytes() uint64 {
	return uint64(C.ResourceConstraintsInitialYoungGenerationSizeInBytes(rc.ptr))
}

// finalizer is called when the ResourceConstraints is being garbage collected.
func (rc *ResourceConstraints) finalizer() {
	if rc.ptr != nil {
		C.ResourceConstraintsDelete(rc.ptr)
		rc.ptr = nil
	}
}
