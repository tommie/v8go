// Copyright 2021 Roger Chapman and the v8go contributors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package v8go_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	v8 "github.com/tommie/v8go"
)

func TestFunctionTemplate(t *testing.T) {
	t.Parallel()

	iso := v8.NewIsolate()
	defer iso.Dispose()
	fn := v8.NewFunctionTemplate(iso, func(*v8.FunctionCallbackInfo) *v8.Value { return nil })
	if fn == nil {
		t.Error("expected FunctionTemplate, but got <nil>")
	}
}

func TestFunctionTemplate_panic_on_nil_isolate(t *testing.T) {
	t.Parallel()

	defer func() {
		if err := recover(); err == nil {
			t.Error("expected panic")
		}
	}()
	v8.NewFunctionTemplate(nil, func(*v8.FunctionCallbackInfo) *v8.Value {
		t.Error("unexpected call")
		return nil
	})
}

func TestFunctionTemplate_panic_on_nil_callback(t *testing.T) {
	t.Parallel()

	defer func() {
		if err := recover(); err == nil {
			t.Error("expected panic")
		}
	}()
	iso := v8.NewIsolate()
	defer iso.Dispose()
	v8.NewFunctionTemplate(iso, nil)
}
func TestFunctionTemplate_generates_values(t *testing.T) {
	t.Parallel()

	iso := v8.NewIsolate()
	defer iso.Dispose()
	global := v8.NewObjectTemplate(iso)
	printfn := v8.NewFunctionTemplate(iso, func(info *v8.FunctionCallbackInfo) *v8.Value {
		t.Logf("%+v\n", info.Args())
		return nil
	})
	global.Set("print", printfn, v8.ReadOnly)
	ctx := v8.NewContext(iso, global)
	defer ctx.Close()
	ctx.RunScript("print('foo', 'bar', 0, 1)", "")
	if ctx.RetainedValueCount() != 6 {
		t.Errorf("expected 6 retained values, got: %d", ctx.RetainedValueCount())
	}
}

func TestFunctionTemplate_releases_values(t *testing.T) {
	t.Parallel()

	iso := v8.NewIsolate()
	defer iso.Dispose()
	global := v8.NewObjectTemplate(iso)
	printfn := v8.NewFunctionTemplate(iso, func(info *v8.FunctionCallbackInfo) *v8.Value {
		defer info.Release()
		t.Logf("%+v\n", info.Args())
		return nil
	})
	global.Set("print", printfn, v8.ReadOnly)
	ctx := v8.NewContext(iso, global)
	defer ctx.Close()
	ctx.RunScript("print('foo', 'bar', 0, 1)", "")
	// there is a constant factor associated with the global.
	if ctx.RetainedValueCount() != 1 {
		t.Errorf("expected 1 retained values, got: %d", ctx.RetainedValueCount())
	}
}

func TestFunctionTemplateGetFunction(t *testing.T) {
	t.Parallel()

	t.Run("can_call", func(t *testing.T) {
		t.Parallel()

		iso := v8.NewIsolate()
		defer iso.Dispose()
		ctx := v8.NewContext(iso)
		defer ctx.Close()

		var args *v8.FunctionCallbackInfo
		tmpl := v8.NewFunctionTemplate(iso, func(info *v8.FunctionCallbackInfo) *v8.Value {
			args = info
			reply, _ := v8.NewValue(iso, "hello")
			return reply
		})
		fn := tmpl.GetFunction(ctx)
		ten, err := v8.NewValue(iso, int32(10))
		if err != nil {
			t.Fatal(err)
		}
		ret, err := fn.Call(v8.Undefined(iso), ten)
		if err != nil {
			t.Fatal(err)
		}
		if len(args.Args()) != 1 || args.Args()[0].String() != "10" {
			t.Fatalf("expected args [10], got: %+v", args.Args())
		}
		if !ret.IsString() || ret.String() != "hello" {
			t.Fatalf("expected return value of 'hello', was: %v", ret)
		}
	})

	t.Run("can_throw_string", func(t *testing.T) {
		t.Parallel()

		iso := v8.NewIsolate()
		defer iso.Dispose()

		tmpl := v8.NewFunctionTemplateWithError(
			iso,
			func(info *v8.FunctionCallbackInfo) (*v8.Value, error) {
				return nil, fmt.Errorf("fake error")
			},
		)
		global := v8.NewObjectTemplate(iso)
		global.Set("foo", tmpl)

		ctx := v8.NewContext(iso, global)
		defer ctx.Close()

		ret, err := ctx.RunScript(
			"(() => { try { foo(); return null; } catch (e) { return e; } })()",
			"",
		)
		if err != nil {
			t.Fatal(err)
		}
		if !ret.IsString() || ret.String() != "fake error" {
			t.Fatalf("expected return value of 'hello', was: %v", ret)
		}
	})

	t.Run("can_throw_exception", func(t *testing.T) {
		t.Parallel()

		iso := v8.NewIsolate()
		defer iso.Dispose()

		tmpl := v8.NewFunctionTemplateWithError(
			iso,
			func(info *v8.FunctionCallbackInfo) (*v8.Value, error) {
				return nil, v8.NewError(iso, "fake error")
			},
		)
		global := v8.NewObjectTemplate(iso)
		global.Set("foo", tmpl)

		ctx := v8.NewContext(iso, global)
		defer ctx.Close()

		ret, err := ctx.RunScript(
			"(() => { try { foo(); return null; } catch (e) { return e; } })()",
			"",
		)
		if err != nil {
			t.Fatal(err)
		}
		if !ret.IsNativeError() || !strings.Contains(ret.String(), "fake error") {
			t.Fatalf("expected return value of Error('hello'), was: %v", ret)
		}
	})
}

func TestFunctionCallbackInfoThis(t *testing.T) {
	t.Parallel()

	iso := v8.NewIsolate()
	defer iso.Dispose()

	foo := v8.NewObjectTemplate(iso)
	foo.Set("name", "foobar")

	var this *v8.Object
	barfn := v8.NewFunctionTemplate(iso, func(info *v8.FunctionCallbackInfo) *v8.Value {
		this = info.This()
		return nil
	})
	foo.Set("bar", barfn)

	global := v8.NewObjectTemplate(iso)
	global.Set("foo", foo)

	ctx := v8.NewContext(iso, global)
	defer ctx.Close()
	ctx.RunScript("foo.bar()", "")

	v, _ := this.Get("name")
	if v.String() != "foobar" {
		t.Errorf("expected this.name to be foobar, but got %q", v)
	}
}

func TestFunctionTemplate_instance_template(t *testing.T) {
	t.Parallel()

	iso := v8.NewIsolate()
	defer iso.Dispose()

	constructor := v8.NewFunctionTemplate(iso,
		// A constructore doesn't need to return a value
		func(info *v8.FunctionCallbackInfo) *v8.Value { return nil })
	constructor.InstanceTemplate().
		Set("getBar", v8.NewFunctionTemplateWithError(iso, func(info *v8.FunctionCallbackInfo) (*v8.Value, error) {
			return v8.NewValue(iso, "Bar")
		}))
	global := v8.NewObjectTemplate(iso)
	global.Set("Foo", constructor)
	ctx := v8.NewContext(iso, global)
	defer ctx.Close()

	val, err := ctx.RunScript("const foo = new Foo(); foo.getBar()", "")
	if err != nil || val == nil {
		t.Fatal("Script error", err)
	}
	if val.String() != "Bar" {
		t.Errorf("Unexpected value. Expected 'Bar'. Got: '%s'", val.String())
	}

	// The function is an "own property" of the instance, and the only one
	val, err = ctx.RunScript(`[
		Object.getOwnPropertyNames(foo).includes("getBar"),
		Object.getOwnPropertyNames(Foo.prototype).includes("getBar"),
	].join(", ")`, "")

	if err != nil || val == nil {
		t.Fatal("Script error", err)
	}
	if val.String() != "true, false" {
		t.Errorf(`Unexpected value.
	[own property of instance, own property of prototype]
	Expected 'true, false'. Got: '%s'`,
			val.String(),
		)
	}
}

func TestFunctionTemplate_prototype_template(t *testing.T) {
	t.Parallel()

	iso := v8.NewIsolate()
	defer iso.Dispose()

	constructor := v8.NewFunctionTemplate(iso,
		// This works as a constructor, so we don't need to return any values.
		func(info *v8.FunctionCallbackInfo) *v8.Value { return nil })
	constructor.PrototypeTemplate().
		Set("getBar", v8.NewFunctionTemplateWithError(iso,
			func(info *v8.FunctionCallbackInfo) (*v8.Value, error) {
				return v8.NewValue(iso, "Bar")
			}))
	global := v8.NewObjectTemplate(iso)
	global.Set("Foo", constructor)

	ctx := v8.NewContext(iso, global)
	defer ctx.Close()

	val, err := ctx.RunScript("const foo = new Foo(); foo.getBar()", "")
	if err != nil || val == nil {
		t.Fatal("Script error", err)
	}
	if val.String() != "Bar" {
		t.Fatalf("Unexpected value. Expected 'Bar'. Got: '%s'", val.String())
	}

	val, err = ctx.RunScript(`[
		Object.getOwnPropertyNames(foo).includes("getBar"),
		Object.getOwnPropertyNames(Foo.prototype).includes("getBar"),
	].join(", ")`, "")
	if err != nil || val == nil {
		t.Fatal("Script error", err)
	}
	if val.String() != "false, true" {
		t.Errorf(`Unexpected value.
	[own property of instance, own property of prototype]
	Expected 'false, true'. Got: '%s'`,
			val.String(),
		)
	}
}

func TestFunctionTemplate_inherit(t *testing.T) {
	t.Parallel()

	iso := v8.NewIsolate()
	defer iso.Dispose()

	// A constructor doesn't need to return a value
	superClassConstructor := v8.NewFunctionTemplate(iso,
		func(info *v8.FunctionCallbackInfo) *v8.Value { return nil })
	superClassConstructor.PrototypeTemplate().
		Set("superString", v8.NewFunctionTemplateWithError(iso,
			func(info *v8.FunctionCallbackInfo) (*v8.Value, error) {
				return v8.NewValue(iso, "super")
			}))
	subClassConstructor := v8.NewFunctionTemplate(iso,
		func(info *v8.FunctionCallbackInfo) *v8.Value { return nil })
	subClassConstructor.PrototypeTemplate().
		Set("subString", v8.NewFunctionTemplateWithError(iso,
			func(info *v8.FunctionCallbackInfo) (*v8.Value, error) {
				return v8.NewValue(iso, "sub")
			}))
	global := v8.NewObjectTemplate(iso)
	global.Set("Super", superClassConstructor)
	global.Set("Sub", subClassConstructor)
	subClassConstructor.Inherit(superClassConstructor)

	ctx := v8.NewContext(iso, global)
	defer ctx.Close()

	val1, err1 := ctx.RunScript("const sup = new Super(); sup.superString()", "")
	val2, err2 := ctx.RunScript("const sub = new Sub(); sub.superString()", "")
	val3, err3 := ctx.RunScript("sup.subString && sup.subString()", "")
	val4, err4 := ctx.RunScript("sub.subString && sub.subString()", "")
	subInheritsFromSuper, err5 := ctx.RunScript(
		"Object.getPrototypeOf(Sub.prototype) === Super.prototype", "")
	superInheritsFromObject, err6 := ctx.RunScript(
		"Object.getPrototypeOf(Super.prototype) === Object.prototype", "")
	if err := errors.Join(err1, err2, err3, err4, err5, err6); err != nil {
		t.Fatal("Script error", err)
	}
	if !val1.IsString() || val1.String() != "super" {
		t.Errorf("Expcted super.superString() to return 'super', got %s", val1.String())
	}
	if !val2.IsString() || val2.String() != "super" {
		t.Errorf("Expcted sub.superString() to return 'super', got %s", val1.String())
	}
	if !val3.IsUndefined() {
		t.Errorf("Expcted super.subString() to return undefined, got %s", val3.String())
	}
	if !val4.IsString() || val4.String() != "sub" {
		t.Errorf("Expcted sub.subString() to return 'sub', got %s", val4.String())
	}
	if !subInheritsFromSuper.Boolean() {
		t.Errorf("Expected Sub to inherit directly from Super")
	}
	if !superInheritsFromObject.Boolean() {
		t.Errorf("Expected Super to inherit directly from Super")
	}
}

func ExampleFunctionTemplate() {
	iso := v8.NewIsolate()
	defer iso.Dispose()
	global := v8.NewObjectTemplate(iso)
	printfn := v8.NewFunctionTemplate(iso, func(info *v8.FunctionCallbackInfo) *v8.Value {
		fmt.Printf("%+v\n", info.Args())
		return nil
	})
	global.Set("print", printfn, v8.ReadOnly)
	ctx := v8.NewContext(iso, global)
	defer ctx.Close()
	ctx.RunScript("print('foo', 'bar', 0, 1)", "")
	// Output:
	// [foo bar 0 1]
}
