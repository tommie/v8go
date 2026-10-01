// Copyright 2019 Roger Chapman and the v8go contributors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package v8go

// #include <stdlib.h>
// #include "errors.h"
import "C"
import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// ErrHeapLimitReached is matched by errors.Is for errors from executing
// JavaScript that was terminated because the isolate's heap limit was
// reached. The process isn't killed; the isolate can be used again.
var ErrHeapLimitReached = errors.New("heap limit reached")

// JSError is an error that is returned if there is are any
// JavaScript exceptions handled in the context. When used with the fmt
// verb `%+v`, will output the JavaScript stack trace, if available.
type JSError struct {
	Message    string
	Location   string
	StackTrace string

	// cause is returned by Unwrap, e.g. ErrHeapLimitReached.
	cause error

	// message is the serialized exception message, parsed by
	// ExceptionMessage. A string keeps JSError comparable.
	message string
}

// Message describes a JavaScript exception, as reported by V8.
type Message struct {
	// Text is the message, e.g. "Uncaught ReferenceError: c is not defined".
	Text string

	// ScriptResourceName is the name of the script, e.g. the origin
	// passed to RunScript.
	ScriptResourceName string

	// SourceLine is the line of source code where the exception happened.
	SourceLine string

	// LineNumber is 1-based, or 0 if unknown.
	LineNumber int

	// StartPosition and EndPosition are 0-based offsets into the script.
	StartPosition, EndPosition int

	// StartColumn and EndColumn are 0-based.
	StartColumn, EndColumn int

	// WASMFunctionIndex is the WebAssembly function, or -1.
	WASMFunctionIndex int

	// StackTrace is where the exception was thrown, innermost first.
	StackTrace []StackFrame
}

// StackFrame is a frame in the stack trace of a Message.
type StackFrame struct {
	ScriptName   string
	FunctionName string

	// LineNumber and ColumnNumber are 1-based.
	LineNumber, ColumnNumber int

	IsEval           bool
	IsConstructor    bool
	IsWASM           bool
	IsUserJavaScript bool
}

func newJSError(rtnErr C.RtnError) error {
	err := &JSError{
		Message:    C.GoString(rtnErr.msg),
		Location:   C.GoString(rtnErr.location),
		StackTrace: C.GoString(rtnErr.stack),
	}
	if rtnErr.message != nil {
		err.message = C.GoStringN(rtnErr.message, rtnErr.message_length)
	}
	if rtnErr.heap_limit_reached != 0 {
		err.cause = ErrHeapLimitReached
	}
	C.ErrorRelease(rtnErr)
	return err
}

// ExceptionMessage returns details about where the exception happened.
// It returns nil unless the isolate was created WithExceptionMessages,
// or if V8 reported no message, e.g. when execution was terminated.
func (e *JSError) ExceptionMessage() *Message {
	if e.message == "" {
		return nil
	}
	return parseMessage(e.message)
}

// parseMessage parses the format written by SerializeMessage in
// errors.cc. It panics if it's malformed, since that is a bug.
func parseMessage(s string) *Message {
	r := messageReader{s: s}
	m := &Message{
		Text:               r.string(),
		ScriptResourceName: r.string(),
		SourceLine:         r.string(),
		LineNumber:         r.int(),
		StartPosition:      r.int(),
		EndPosition:        r.int(),
		StartColumn:        r.int(),
		EndColumn:          r.int(),
		WASMFunctionIndex:  r.int(),
	}
	if n := r.int(); n > 0 {
		m.StackTrace = make([]StackFrame, n)
		for i := range m.StackTrace {
			f := &m.StackTrace[i]
			f.ScriptName = r.string()
			f.FunctionName = r.string()
			f.LineNumber = r.int()
			f.ColumnNumber = r.int()
			flags := r.int()
			f.IsEval = flags&1 != 0
			f.IsConstructor = flags&2 != 0
			f.IsWASM = flags&4 != 0
			f.IsUserJavaScript = flags&8 != 0
		}
	}
	if r.s != "" {
		panic(fmt.Sprintf("v8go: %d trailing bytes in exception message", len(r.s)))
	}
	return m
}

type messageReader struct {
	s string
}

func (r *messageReader) int() int {
	if len(r.s) < 4 {
		panic("v8go: truncated exception message")
	}
	v := int32(binary.LittleEndian.Uint32([]byte(r.s[:4])))
	r.s = r.s[4:]
	return int(v)
}

func (r *messageReader) string() string {
	n := r.int()
	if n < 0 || len(r.s) < n {
		panic("v8go: truncated exception message")
	}
	v := r.s[:n]
	r.s = r.s[n:]
	return v
}

// Unwrap returns the underlying cause, if known.
func (e *JSError) Unwrap() error {
	return e.cause
}

func (e *JSError) Error() string {
	return e.Message
}

// Format implements the fmt.Formatter interface to provide a custom formatter
// primarily to output the javascript stack trace with %+v
func (e *JSError) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') && e.StackTrace != "" {
			// The StackTrace starts with the Message, so only the former needs to be printed
			io.WriteString(s, e.StackTrace)

			// If it was a compile time error, then there wouldn't be a runtime stack trace,
			// but StackTrace will still include the Message, making them equal. In this case,
			// we want to include the Location where the compilation failed.
			if e.StackTrace == e.Message && e.Location != "" {
				fmt.Fprintf(s, " (at %s)", e.Location)
			}
			return
		}
		fallthrough
	case 's':
		io.WriteString(s, e.Message)
	case 'q':
		fmt.Fprintf(s, "%q", e.Message)
	}
}
