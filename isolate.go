// Copyright 2019 Roger Chapman and the v8go contributors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package v8go

// #include <stdlib.h>
// #include "isolate.h"
import "C"

import (
	"io"
	"runtime/cgo"
	"sync"
	"unsafe"
)

// Isolate is a JavaScript VM instance with its own heap and
// garbage collector. Most applications will create one isolate
// with many V8 contexts for execution.
type Isolate struct {
	ptr C.IsolatePtr

	cbMutex sync.RWMutex
	cbSeq   int
	cbs     map[int]FunctionCallbackWithError

	null      *Value
	undefined *Value
}

// HeapStatistics represents V8 isolate heap statistics
type HeapStatistics struct {
	TotalHeapSize            uint64
	TotalHeapSizeExecutable  uint64
	TotalPhysicalSize        uint64
	TotalAvailableSize       uint64
	UsedHeapSize             uint64
	HeapSizeLimit            uint64
	MallocedMemory           uint64
	ExternalMemory           uint64
	PeakMallocedMemory       uint64
	NumberOfNativeContexts   uint64
	NumberOfDetachedContexts uint64
}

type resourceConstraints struct {
	InitialHeapSizeInBytes uint64
	MaxHeapSizeInBytes     uint64
}

// IsolateOption configures an Isolate on creation.
type IsolateOption func(*isolateConfig)

// isolateConfig holds the configuration for creating an isolate.
type isolateConfig struct {
	resourceConstraints *resourceConstraints
}

// WithResourceConstraints sets memory constraints for the isolate.
//
// When any isolate reaches its heap limit, v8go terminates the running
// script instead of letting V8 end the process. The returned error then
// matches ErrHeapLimitReached, and the isolate can be used again.
func WithResourceConstraints(initialHeapSizeInBytes, maxHeapSizeInBytes uint64) IsolateOption {
	return func(config *isolateConfig) {
		config.resourceConstraints = &resourceConstraints{
			InitialHeapSizeInBytes: initialHeapSizeInBytes,
			MaxHeapSizeInBytes:     maxHeapSizeInBytes,
		}
	}
}

// NewIsolate creates a new V8 isolate with the provided options.
// Only one thread may access a given isolate at a time, but different
// threads may access different isolates simultaneously.
// When an isolate is no longer used its resources should be freed
// by calling iso.Dispose().
// An *Isolate can be used as a v8go.ContextOption to create a new
// Context, rather than creating a new default Isolate.
func NewIsolate(opts ...IsolateOption) *Isolate {
	initializeIfNecessary()

	config := &isolateConfig{}
	for _, opt := range opts {
		opt(config)
	}

	var cConstraints C.IsolateConstraintsPtr
	if config.resourceConstraints != nil {
		cConstraints = &C.IsolateConstraints{
			initial_heap_size_in_bytes: C.size_t(config.resourceConstraints.InitialHeapSizeInBytes),
			maximum_heap_size_in_bytes: C.size_t(config.resourceConstraints.MaxHeapSizeInBytes),
		}
	}

	iso := &Isolate{
		ptr: C.NewIsolate(cConstraints),
		cbs: make(map[int]FunctionCallbackWithError),
	}
	iso.null = newValueNull(iso)
	iso.undefined = newValueUndefined(iso)
	return iso
}

// TerminateExecution terminates forcefully the current thread
// of JavaScript execution in the given isolate.
func (i *Isolate) TerminateExecution() {
	C.IsolateTerminateExecution(i.ptr)
}

// IsExecutionTerminating returns whether V8 is currently terminating
// Javascript execution. If true, there are still JavaScript frames
// on the stack and the termination exception is still active.
func (i *Isolate) IsExecutionTerminating() bool {
	return C.IsolateIsExecutionTerminating(i.ptr) == 1
}

type CompileOptions struct {
	CachedData *CompilerCachedData

	Mode CompileMode
}

// CompileUnboundScript will create an UnboundScript (i.e. context-indepdent)
// using the provided source JavaScript, origin (a.k.a. filename), and options.
// If options contain a non-null CachedData, compilation of the script will use
// that code cache.
// error will be of type `JSError` if not nil.
func (i *Isolate) CompileUnboundScript(
	source, origin string,
	opts CompileOptions,
) (*UnboundScript, error) {
	cSource := C.CString(source)
	cOrigin := C.CString(origin)
	defer C.free(unsafe.Pointer(cSource))
	defer C.free(unsafe.Pointer(cOrigin))

	var cOptions C.CompileOptions
	if opts.CachedData != nil {
		if opts.Mode != 0 {
			panic("On CompileOptions, Mode and CachedData can't both be set")
		}
		cOptions.compileOption = C.ScriptCompilerConsumeCodeCache
		cOptions.cachedData = C.ScriptCompilerCachedData{
			data:   (*C.uchar)(unsafe.Pointer(&opts.CachedData.Bytes[0])),
			length: C.int(len(opts.CachedData.Bytes)),
		}
	} else {
		cOptions.compileOption = C.int(opts.Mode)
	}

	rtn := C.IsolateCompileUnboundScript(i.ptr, cSource, cOrigin, cOptions)
	if rtn.ptr == nil {
		return nil, newJSError(rtn.error)
	}
	if opts.CachedData != nil {
		opts.CachedData.Rejected = int(rtn.cachedDataRejected) == 1
	}
	return &UnboundScript{
		ptr: rtn.ptr,
		iso: i,
	}, nil
}

// GetHeapStatistics returns heap statistics for an isolate.
func (i *Isolate) GetHeapStatistics() HeapStatistics {
	hs := C.IsolationGetHeapStatistics(i.ptr)

	return HeapStatistics{
		TotalHeapSize:            uint64(hs.total_heap_size),
		TotalHeapSizeExecutable:  uint64(hs.total_heap_size_executable),
		TotalPhysicalSize:        uint64(hs.total_physical_size),
		TotalAvailableSize:       uint64(hs.total_available_size),
		UsedHeapSize:             uint64(hs.used_heap_size),
		HeapSizeLimit:            uint64(hs.heap_size_limit),
		MallocedMemory:           uint64(hs.malloced_memory),
		ExternalMemory:           uint64(hs.external_memory),
		PeakMallocedMemory:       uint64(hs.peak_malloced_memory),
		NumberOfNativeContexts:   uint64(hs.number_of_native_contexts),
		NumberOfDetachedContexts: uint64(hs.number_of_detached_contexts),
	}
}

// LowMemoryNotification tells V8 that the system is running low on
// memory. V8 then performs a full garbage collection, and frees other
// memory it can, e.g. caches.
func (i *Isolate) LowMemoryNotification() {
	C.IsolateLowMemoryNotification(i.ptr)
}

// WriteHeapSnapshot takes a snapshot of the JavaScript heap and writes it
// to w, in the JSON format that Chrome DevTools can load (".heapsnapshot"
// files). Comparing snapshots taken at different times is useful for
// finding memory leaks. V8 performs a full garbage collection before
// taking the snapshot.
//
// The isolate is locked while the snapshot is written. The returned
// error is the first error returned by w.
func (i *Isolate) WriteHeapSnapshot(w io.Writer) error {
	hw := heapSnapshotWriter{w: w}
	h := cgo.NewHandle(&hw)
	defer h.Delete()

	C.IsolateWriteHeapSnapshot(i.ptr, C.uintptr_t(h))
	return hw.err
}

// heapSnapshotWriter is the state of a WriteHeapSnapshot call.
type heapSnapshotWriter struct {
	w   io.Writer
	err error
}

// goWriteHeapSnapshotChunk is called by C code for each chunk of a heap
// snapshot. writerRef is a [cgo.Handle] of a *heapSnapshotWriter. It
// returns zero to abort the snapshot.
//
//export goWriteHeapSnapshotChunk
func goWriteHeapSnapshotChunk(writerRef C.uintptr_t, data *C.char, size C.int) C.int {
	hw := cgo.Handle(writerRef).Value().(*heapSnapshotWriter)

	// io.Writer implementations must not retain the slice, so it can
	// refer to the C memory.
	n, err := hw.w.Write(unsafe.Slice((*byte)(unsafe.Pointer(data)), int(size)))
	if err == nil && n < int(size) {
		err = io.ErrShortWrite
	}
	if err != nil {
		hw.err = err
		return 0
	}
	return 1
}

// Dispose will dispose the Isolate VM; subsequent calls will panic.
func (i *Isolate) Dispose() {
	if i.ptr == nil {
		return
	}
	C.IsolateDispose(i.ptr)
	i.ptr = nil
}

// ThrowException schedules an exception to be thrown when returning to
// JavaScript. When an exception has been scheduled it is illegal to invoke
// any JavaScript operation; the caller must return immediately and only after
// the exception has been handled does it become legal to invoke JavaScript operations.
func (i *Isolate) ThrowException(value *Value) *Value {
	if i.ptr == nil {
		panic("Isolate has been disposed")
	}
	return &Value{
		ptr: C.IsolateThrowException(i.ptr, value.ptr),
	}
}

// Deprecated: use `iso.Dispose()`.
func (i *Isolate) Close() {
	i.Dispose()
}

func (i *Isolate) apply(opts *contextOptions) {
	opts.iso = i
}

func (i *Isolate) registerCallback(cb FunctionCallbackWithError) int {
	i.cbMutex.Lock()
	i.cbSeq++
	ref := i.cbSeq
	i.cbs[ref] = cb
	i.cbMutex.Unlock()
	return ref
}

func (i *Isolate) getCallback(ref int) FunctionCallbackWithError {
	i.cbMutex.RLock()
	defer i.cbMutex.RUnlock()
	return i.cbs[ref]
}
