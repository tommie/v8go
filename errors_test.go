// Copyright 2019 Roger Chapman and the v8go contributors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package v8go_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	v8 "github.com/tommie/v8go"
)

func TestJSErrorFormat(t *testing.T) {
	t.Parallel()
	tests := [...]struct {
		name            string
		err             error
		defaultVerb     string
		defaultVerbFlag string
		stringVerb      string
		quoteVerb       string
	}{
		{"WithStack", &v8.JSError{Message: "msg", StackTrace: "stack"}, "msg", "stack", "msg", `"msg"`},
		{"WithoutStack", &v8.JSError{Message: "msg"}, "msg", "msg", "msg", `"msg"`},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if s := fmt.Sprintf("%v", tt.err); s != tt.defaultVerb {
				t.Errorf("incorrect format for %%v: %s", s)
			}
			if s := fmt.Sprintf("%+v", tt.err); s != tt.defaultVerbFlag {
				t.Errorf("incorrect format for %%+v: %s", s)
			}
			if s := fmt.Sprintf("%s", tt.err); s != tt.stringVerb {
				t.Errorf("incorrect format for %%s: %s", s)
			}
			if s := fmt.Sprintf("%q", tt.err); s != tt.quoteVerb {
				t.Errorf("incorrect format for %%q: %s", s)
			}
		})
	}
}

func TestJSErrorOutput(t *testing.T) {
	t.Parallel()
	ctx := v8.NewContext(nil)
	defer ctx.Isolate().Dispose()
	defer ctx.Close()

	math := `
	function add(a, b) {
		return a + b;
	}

	function addMore(a, b) {
		return add(a, c);
	}`

	main := `
	let a = add(3, 5);
	let b = addMore(a, 6);
	b;
	`

	ctx.RunScript(math, "math.js")
	_, err := ctx.RunScript(main, "main.js")
	if err == nil {
		t.Error("expected error but got <nil>")
		return
	}
	e, ok := err.(*v8.JSError)
	if !ok {
		t.Errorf("expected error of type JSError, got %T", err)
	}
	if e.Message != "ReferenceError: c is not defined" {
		t.Errorf("unexpected error message: %q", e.Message)
	}
	if e.Location != "math.js:7:17" {
		t.Errorf("unexpected error location: %q", e.Location)
	}
	expectedStack := `ReferenceError: c is not defined
    at addMore (math.js:7:17)
    at main.js:3:10`

	if e.StackTrace != expectedStack {
		t.Errorf("unexpected error stack trace: %q", e.StackTrace)
	}
}

func TestJSErrorExceptionMessage(t *testing.T) {
	t.Parallel()

	math := `
	function add(a, b) {
		return a + b;
	}

	function addMore(a, b) {
		return add(a, c);
	}`

	run := func(t *testing.T, iso *v8.Isolate) *v8.JSError {
		t.Helper()
		ctx := v8.NewContext(iso)
		defer ctx.Close()
		if _, err := ctx.RunScript(math, "math.js"); err != nil {
			t.Fatalf("RunScript failed: %v", err)
		}
		_, err := ctx.RunScript("\n\tlet b = addMore(1, 6);\n", "main.js")
		var e *v8.JSError
		if !errors.As(err, &e) {
			t.Fatalf("RunScript error: got %v, want a JSError", err)
		}
		return e
	}

	t.Run("enabled", func(t *testing.T) {
		iso := v8.NewIsolate(v8.WithExceptionMessages())
		defer iso.Dispose()

		e := run(t, iso)
		m := e.ExceptionMessage()
		if m == nil {
			t.Fatal("ExceptionMessage: got nil")
		}
		want := v8.Message{
			Text:               "Uncaught ReferenceError: c is not defined",
			ScriptResourceName: "math.js",
			SourceLine:         "\t\treturn add(a, c);",
			LineNumber:         7,
			StartPosition:      85,
			EndPosition:        86,
			StartColumn:        16,
			EndColumn:          17,
			WASMFunctionIndex:  -1,
			StackTrace: []v8.StackFrame{
				{ScriptName: "math.js", FunctionName: "addMore", LineNumber: 7, ColumnNumber: 17, IsUserJavaScript: true},
				{ScriptName: "main.js", LineNumber: 2, ColumnNumber: 10, IsUserJavaScript: true},
			},
		}
		if !reflect.DeepEqual(*m, want) {
			t.Errorf("ExceptionMessage: got %+v, want %+v", *m, want)
		}

		// The message is data, so equal errors still compare equal.
		if e2 := run(t, iso); *e2 != *e {
			t.Errorf("JSError values differ: %+v and %+v", *e, *e2)
		}
	})

	t.Run("disabled", func(t *testing.T) {
		iso := v8.NewIsolate()
		defer iso.Dispose()

		if m := run(t, iso).ExceptionMessage(); m != nil {
			t.Errorf("ExceptionMessage: got %+v, want nil", m)
		}
	})

	t.Run("terminated", func(t *testing.T) {
		iso := v8.NewIsolate(v8.WithExceptionMessages())
		defer iso.Dispose()
		// TerminateExecution only has an effect while JavaScript runs.
		global := v8.NewObjectTemplate(iso)
		global.Set("stop", v8.NewFunctionTemplate(iso, func(info *v8.FunctionCallbackInfo) *v8.Value {
			iso.TerminateExecution()
			return nil
		}))
		ctx := v8.NewContext(iso, global)
		defer ctx.Close()

		_, err := ctx.RunScript("stop(); for (;;) {}", "loop.js")
		var e *v8.JSError
		if !errors.As(err, &e) {
			t.Fatalf("RunScript error: got %v, want a JSError", err)
		}
		if m := e.ExceptionMessage(); m != nil {
			t.Errorf("ExceptionMessage: got %+v, want nil", m)
		}
	})
}

func TestJSErrorFormat_forSyntaxError(t *testing.T) {
	t.Parallel()
	iso := v8.NewIsolate()
	defer iso.Dispose()
	ctx := v8.NewContext(iso)
	defer ctx.Close()

	script := `
		let x = 1;
		let y = x + ;
		let z = x + z;
	`
	_, err := ctx.RunScript(script, "xyz.js")
	jsErr := err.(*v8.JSError)
	if jsErr.StackTrace != jsErr.Message {
		t.Errorf("unexpected StackTrace %q not equal to Message %q", jsErr.StackTrace, jsErr.Message)
	}
	if jsErr.Location == "" {
		t.Errorf("missing Location")
	}

	msg := fmt.Sprintf("%+v", err)
	if msg != "SyntaxError: Unexpected token ';' (at xyz.js:3:15)" {
		t.Errorf("unexpected verbose error message: %q", msg)
	}
}
