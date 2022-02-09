// Copyright 2019 Roger Chapman and the v8go contributors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package v8go

// #include <stdlib.h>
// #include "v8go.h"
import "C"
import (
	"fmt"
	"io"
)

var DisableLegacyJSErrorStrings = false

// JSError is an error that is returned if there is are any
// JavaScript exceptions handled in the context. When used with the fmt
// verb `%+v`, will output the JavaScript stack trace, if available.
type JSError struct {
	*Value

	m *Message

	// Deprecated fields. Use ExceptionMessage() instead.

	Message    string
	Location   string
	StackTrace string
}

func newJSError(ctx *Context, rtnErr C.RtnError) error {
	m := newMessageFromC(rtnErr.excMessage)
	startCol, _ := m.ColumnRange()

	var v *Value
	if rtnErr.exception != nil {
		v = &Value{rtnErr.exception, ctx}
	}
	e := &JSError{
		Value: v,
		m:     m,
	}

	if !DisableLegacyJSErrorStrings {
		if rtnErr.exception != nil {
			v = &Value{rtnErr.exception, ctx}
			if o, err := v.AsObject(); err == nil {
				if so, err := o.Get("stack"); err == nil {
					e.StackTrace = so.String()
				}
			}
			e.Message = v.String()
		} else {
			e.Message = m.Text()
		}
		e.Location = fmt.Sprint(m.ScriptResourceName(), ":", m.LineNumber(), ":", startCol+1)
	}
	return e
}

func (e *JSError) Error() string {
	return e.Message
}

func (e *JSError) ExceptionMessage() *Message {
	return e.m
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
