// Copyright 2022 Roger Chapman and the v8go contributors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package v8go

// #include <stdlib.h>
// #include "v8go.h"
import "C"
import "unsafe"

type Message struct {
	stackTrace       *StackTrace
	text             string
	source           string
	lineNumber       int
	posStart, posEnd int
	colStart, colEnd int
	wasmFuncIndex    int
}

func newMessageFromC(iso *Isolate, v C.RtnMessage) *Message {
	defer func() {
		C.free(unsafe.Pointer(v.text))
		C.free(unsafe.Pointer(v.source))
	}()

	return &Message{
		stackTrace:    newStackTraceFromC(iso, v.stackTrace),
		text:          C.GoString(v.text),
		source:        C.GoString(v.source),
		lineNumber:    int(v.lineNumber),
		posStart:      int(v.posStart),
		posEnd:        int(v.posEnd),
		colStart:      int(v.colStart),
		colEnd:        int(v.colEnd),
		wasmFuncIndex: int(v.wasmFuncIndex),
	}
}

func (m *Message) Text() string              { return m.text }
func (m *Message) Source() string            { return m.source }
func (m *Message) LineNumber() int           { return m.lineNumber }
func (m *Message) PositionRange() (int, int) { return m.posStart, m.posEnd }
func (m *Message) ColumnRange() (int, int)   { return m.colStart, m.colEnd }
func (m *Message) WASMFunctionIndex() int    { return m.wasmFuncIndex }
func (m *Message) StackTrace() *StackTrace   { return m.stackTrace }

type StackTrace struct {
	ptr C.StackTracePtr
	iso *Isolate
}

func newStackTraceFromC(iso *Isolate, ptr C.StackTracePtr) *StackTrace {
	return &StackTrace{ptr, iso}
}

func (st *StackTrace) finalizer() {
	// Using v8::PersistentBase::Reset() wouldn't be thread-safe to do
	// from this finalizer goroutine so just free the wrapper and let
	// the stack trace itself get cleaned up when the isolate is
	// disposed.
	C.StackTraceFreeWrapper(st.ptr)
	st.ptr = nil
}

func (st *StackTrace) NumFrames() int {
	return int(C.StackTraceNumFrames(st.ptr))
}

func (st *StackTrace) Frame(i int) *StackFrame {
	return newStackFrameFromC(C.StackTraceFrame(st.ptr, C.uint32_t(i)))
}

type StackFrame struct {
	ScriptName       string
	ScriptSource     string
	FunctionName     string
	LineNumber       int
	ColumnNumber     int
	IsEval           bool
	IsConstructor    bool
	IsWASM           bool
	IsUserJavaScript bool
}

func newStackFrameFromC(v C.RtnStackFrame) *StackFrame {
	defer func() {
		C.free(unsafe.Pointer(v.scriptName))
		C.free(unsafe.Pointer(v.scriptSource))
		C.free(unsafe.Pointer(v.functionName))
	}()

	return &StackFrame{
		ScriptName:       C.GoString(v.scriptName),
		ScriptSource:     C.GoString(v.scriptSource),
		FunctionName:     C.GoString(v.functionName),
		LineNumber:       int(v.lineNumber),
		ColumnNumber:     int(v.columnNumber),
		IsEval:           bool(v.isEval),
		IsConstructor:    bool(v.isConstructor),
		IsWASM:           bool(v.isWASM),
		IsUserJavaScript: bool(v.isUserJavaScript),
	}
}
