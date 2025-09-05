// Copyright 2024 Roger Chapman and the v8go contributors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package v8go

// ResourceConstraints represents V8 resource constraints that specify
// the limits of the runtime's memory use. You must set the heap size
// before initializing the VM - the size cannot be adjusted after the
// VM is initialized.
type ResourceConstraints struct {
	// StackLimit sets the address beyond which the VM's stack may not grow.
	StackLimit uintptr

	// CodeRangeSizeInBytes sets the amount of virtual memory reserved for
	// generated code. This is relevant for 64-bit architectures that rely on
	// code range for calls in code.
	CodeRangeSizeInBytes uint64

	// MaxOldGenerationSizeInBytes sets the maximum size of the old generation.
	// When the old generation approaches this limit, V8 will perform series of
	// garbage collections and invoke the NearHeapLimitCallback. If the garbage
	// collections do not help and the callback does not increase the limit,
	// then V8 will crash with V8::FatalProcessOutOfMemory.
	MaxOldGenerationSizeInBytes uint64

	// MaxYoungGenerationSizeInBytes sets the maximum size of the young generation,
	// which consists of two semi-spaces and a large object space. This affects
	// frequency of Scavenge garbage collections and should be typically much
	// smaller that the old generation.
	MaxYoungGenerationSizeInBytes uint64

	// InitialOldGenerationSizeInBytes sets the initial size of the old generation.
	InitialOldGenerationSizeInBytes uint64

	// InitialYoungGenerationSizeInBytes sets the initial size of the young generation.
	InitialYoungGenerationSizeInBytes uint64
}

// ConfigureDefaultsFromHeapSize returns ResourceConstraints configured with reasonable
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
func ConfigureDefaultsFromHeapSize(initialHeapSizeInBytes, maximumHeapSizeInBytes uint64) *ResourceConstraints {
	// We'll let V8 configure the defaults and return reasonable values
	// For now, we'll set some sensible defaults based on the heap size
	return &ResourceConstraints{
		MaxOldGenerationSizeInBytes:       maximumHeapSizeInBytes * 8 / 10, // 80% for old generation
		MaxYoungGenerationSizeInBytes:     maximumHeapSizeInBytes * 2 / 10, // 20% for young generation
		InitialOldGenerationSizeInBytes:   initialHeapSizeInBytes * 8 / 10,
		InitialYoungGenerationSizeInBytes: initialHeapSizeInBytes * 2 / 10,
	}
}

// ConfigureDefaults returns ResourceConstraints configured with reasonable default
// values based on the capabilities of the current device the VM is running on.
//
// physicalMemory: The total amount of physical memory on the current device, in bytes.
// virtualMemoryLimit: The amount of virtual memory on the current device, in bytes,
// or zero, if there is no limit.
func ConfigureDefaults(physicalMemory, virtualMemoryLimit uint64) *ResourceConstraints {
	// Use a reasonable fraction of physical memory (e.g., 1/8th for max heap)
	maxHeap := physicalMemory / 8
	initialHeap := maxHeap / 4

	return &ResourceConstraints{
		MaxOldGenerationSizeInBytes:       maxHeap * 8 / 10,
		MaxYoungGenerationSizeInBytes:     maxHeap * 2 / 10,
		InitialOldGenerationSizeInBytes:   initialHeap * 8 / 10,
		InitialYoungGenerationSizeInBytes: initialHeap * 2 / 10,
	}
}
