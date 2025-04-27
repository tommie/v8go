// Copyright 2020 Roger Chapman and the v8go contributors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package v8go_test

import (
	"fmt"
	"math/big"
	"runtime"
	"testing"

	v8 "github.com/tommie/v8go"
)

func TestObjectTemplate(t *testing.T) {
	t.Parallel()
	iso := v8.NewIsolate()
	defer iso.Dispose()
	obj := v8.NewObjectTemplate(iso)

	setError := func(t *testing.T, err error) {
		if err != nil {
			t.Errorf("failed to set property: %v", err)
		}
	}

	val, _ := v8.NewValue(iso, "bar")
	objVal := v8.NewObjectTemplate(iso)
	bigbigint, _ := new(
		big.Int,
	).SetString("36893488147419099136", 10)
	// larger than a single word size (64bit)
	bigbignegint, _ := new(big.Int).SetString("-36893488147419099136", 10)

	tests := [...]struct {
		name  string
		value interface{}
	}{
		{"str", "foo"},
		{"i32", int32(1)},
		{"u32", uint32(1)},
		{"i64", int64(1)},
		{"u64", uint64(1)},
		{"float64", float64(1)},
		{"bigint", big.NewInt(1)},
		{"biguint", new(big.Int).SetUint64(1 << 63)},
		{"bigbigint", bigbigint},
		{"bigbignegint", bigbignegint},
		{"bool", true},
		{"val", val},
		{"obj", objVal},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			setError(t, obj.Set(tt.name, tt.value, 0))
		})
	}
}

func TestObjectTemplateSetSymbol(t *testing.T) {
	t.Parallel()
	iso := v8.NewIsolate()
	defer iso.Dispose()
	obj := v8.NewObjectTemplate(iso)

	val, _ := v8.NewValue(iso, "bar")
	objVal := v8.NewObjectTemplate(iso)

	if err := obj.SetSymbol(v8.SymbolIterator(iso), val); err != nil {
		t.Errorf("failed to set property: %v", err)
	}
	if err := obj.SetSymbol(v8.SymbolIterator(iso), objVal); err != nil {
		t.Errorf("failed to set template property: %v", err)
	}
}

func TestObjectTemplate_panic_on_nil_isolate(t *testing.T) {
	t.Parallel()

	defer func() {
		if err := recover(); err == nil {
			t.Error("expected panic")
		}
	}()
	v8.NewObjectTemplate(nil)
}

func TestGlobalObjectTemplate(t *testing.T) {
	t.Parallel()
	iso := v8.NewIsolate()
	defer iso.Dispose()
	tests := [...]struct {
		global   func() *v8.ObjectTemplate
		source   string
		validate func(t *testing.T, val *v8.Value)
	}{
		{
			func() *v8.ObjectTemplate {
				gbl := v8.NewObjectTemplate(iso)
				gbl.Set("foo", "bar")
				return gbl
			},
			"foo",
			func(t *testing.T, val *v8.Value) {
				if !val.IsString() {
					t.Errorf("expect value %q to be of type String", val)
					return
				}
				if val.String() != "bar" {
					t.Errorf("unexpected value: %v", val)
				}
			},
		},
		{
			func() *v8.ObjectTemplate {
				foo := v8.NewObjectTemplate(iso)
				foo.Set("bar", "baz")
				gbl := v8.NewObjectTemplate(iso)
				gbl.Set("foo", foo)
				return gbl
			},
			"foo.bar",
			func(t *testing.T, val *v8.Value) {
				if val.String() != "baz" {
					t.Errorf("unexpected value: %v", val)
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.source, func(t *testing.T) {
			ctx := v8.NewContext(iso, tt.global())
			val, err := ctx.RunScript(tt.source, "test.js")
			if err != nil {
				t.Fatalf("unexpected error runing script: %v", err)
			}
			tt.validate(t, val)
			ctx.Close()
		})
	}
}

func TestObjectTemplateNewInstance(t *testing.T) {
	t.Parallel()
	iso := v8.NewIsolate()
	defer iso.Dispose()
	tmpl := v8.NewObjectTemplate(iso)
	if _, err := tmpl.NewInstance(nil); err == nil {
		t.Error("expected error but got <nil>")
	}

	tmpl.Set("foo", "bar")
	ctx := v8.NewContext(iso)
	defer ctx.Close()
	obj, _ := tmpl.NewInstance(ctx)
	if foo, _ := obj.Get("foo"); foo.String() != "bar" {
		t.Errorf("unexpected value for object property: %v", foo)
	}
}

func TestObjectTemplateSetAccessorProperty_OnlyGetter(t *testing.T) {
	// Create an accessor property that has only a getter.
	// Setting the value from JS should not have side effects.
	t.Parallel()
	iso := v8.NewIsolate()
	defer iso.Dispose()

	get := v8.NewFunctionTemplateWithError(iso,
		func(*v8.FunctionCallbackInfo) (*v8.Value, error) { return v8.NewValue(iso, "Value") },
	)
	tmpl := v8.NewObjectTemplate(iso)
	tmpl.SetAccessorProperty("prop", get, nil, v8.None)

	global := v8.NewObjectTemplate(iso)
	global.Set("obj", tmpl)
	ctx := v8.NewContext(iso, global)
	defer ctx.Close()

	values, err := ctx.RunScript(`
		const val1 = obj.prop;
		obj.prop = "foo";
		const val2 = obj.prop;
		[val1, val2].join(", ")
	`, "")
	if err != nil {
		t.Fatal("Script error", err)
	}
	if values.String() != "Value, Value" {
		t.Errorf("Unexpected values. Expected: 'Value, Value', got %s", values)
	}
}

func TestObjectTemplateSetAccessorProperty_GetterAnSetter(t *testing.T) {
	t.Parallel()
	iso := v8.NewIsolate()
	defer iso.Dispose()
	var value *v8.Value

	var get = v8.NewFunctionTemplate(iso, func(*v8.FunctionCallbackInfo) *v8.Value {
		return value
	})
	var set = v8.NewFunctionTemplate(iso, func(i *v8.FunctionCallbackInfo) *v8.Value {
		value = i.Args()[0] // A property setter will always have _one_ argument.
		return nil
	})
	tmpl := v8.NewObjectTemplate(iso)
	tmpl.SetAccessorProperty("prop", get, set, v8.None)

	global := v8.NewObjectTemplate(iso)
	global.Set("obj", tmpl)
	ctx := v8.NewContext(iso, global)
	defer ctx.Close()

	values, err := ctx.RunScript(`
		obj.prop = "foo";
		const val1 = obj.prop;
		obj.prop = "bar";
		const val2 = obj.prop;
		[val1, val2].join(", ")
	`, "")
	if err != nil {
		t.Fatal("Script error", err)
	}
	if values.String() != "foo, bar" {
		t.Errorf("Unexpected values. Expected: 'foo, bar', got %s", values)
	}
}

func TestObjectTemplate_garbageCollection(t *testing.T) {
	t.Parallel()

	iso := v8.NewIsolate()

	tmpl := v8.NewObjectTemplate(iso)
	tmpl.Set("foo", "bar")
	ctx := v8.NewContext(iso, tmpl)

	ctx.Close()
	iso.Dispose()

	runtime.GC()
}

func ExampleObjectTemplate_SetAccessorProperty() {
	iso := v8.NewIsolate()
	defer iso.Dispose()
	tmpl := v8.NewObjectTemplate(iso)
	tmpl.SetAccessorProperty(
		"prop",
		// Getter
		v8.NewFunctionTemplateWithError(
			iso,
			func(*v8.FunctionCallbackInfo) (*v8.Value, error) {
				return v8.NewValue(iso, "Value")
			},
		),
		nil, // Setter
		v8.None,
	)

	global := v8.NewObjectTemplate(iso)
	global.Set("obj", tmpl)
	ctx := v8.NewContext(iso, global)
	defer ctx.Close()

	value, _ := ctx.RunScript("obj.prop", "")
	fmt.Printf("Property value: %s\n", value.String())
	// Output:
	// Property value: Value
}

func TestObjectTemplateSetCallAsFunctionHandler(t *testing.T) {
	t.Parallel()
	iso := v8.NewIsolate()
	defer iso.Dispose()
	tmpl := v8.NewObjectTemplate(iso)
	if _, err := tmpl.NewInstance(nil); err == nil {
		t.Error("expected error but got <nil>")
	}

	ctx := v8.NewContext(iso)
	defer ctx.Close()

	tmpl.SetCallAsFunctionHandler(func(info *v8.FunctionCallbackInfo) (*v8.Value, error) {
		return v8.NewValue(iso, "42")
	})
	instance, err := tmpl.NewInstance(ctx)
	if err != nil {
		t.Fatalf("Error creating instance: %v", err)
	}
	ctx.Global().Set("obj", instance)

	res, err := ctx.RunScript(`obj()`, "")
	resStr := res.String()
	if resStr != "42" {
		t.Errorf(`unexpected result. Expected "42", got: %s`, resStr)
	}
}

func TestObjectTemplateMarkAsUndetectable(t *testing.T) {
	t.Parallel()

	iso := v8.NewIsolate()
	defer iso.Dispose()
	obj := v8.NewObjectTemplate(iso)
	obj.MarkAsUndetectable()
	ctx := v8.NewContext(iso)
	defer ctx.Close()
	ft := v8.NewFunctionTemplate(
		iso,
		func(info *v8.FunctionCallbackInfo) *v8.Value { return nil },
	)
	v, err := v8.NewValue(iso, "42")
	if err != nil {
		t.Fatalf("Error creating value: %v", err)
	}
	ft.InstanceTemplate().Set("val", v)
	instance, err := ft.InstanceTemplate().NewInstance(ctx)
	if err != nil {
		t.Fatalf("Error getting new instance: %v", err)
	}
	ctx.Global().Set("obj", instance)
	res, err := ctx.RunScript("'undefined'", "")
	if err != nil {
		t.Fatalf("Error run 'typeof obj': %v", err)
	}
	str := res.String()
	if str != "undefined" {
		t.Errorf(`Expected 'typeof obj' to return "undefined", got: %s`, str)
	}
	res, err = ctx.RunScript("obj.val", "")
	if err != nil {
		t.Fatalf("Error evaluating 'obj.val': %v", err)
	}
	str = res.String()
	if str != "42" {
		t.Errorf(`Expected 'typeof obj' to return "undefined", got: %s`, str)
	}

	/*
	  v8::HandleScope scope(env->GetIsolate());

	  Local<v8::FunctionTemplate> desc =
	      v8::FunctionTemplate::New(env->GetIsolate());
	  desc->InstanceTemplate()->MarkAsUndetectable();  // undetectable
	  desc->InstanceTemplate()->SetCallAsFunctionHandler(ReturnThis);  // callable

	  Local<v8::Object> obj = desc->GetFunction(env.local())
	                              .ToLocalChecked()
	                              ->NewInstance(env.local())
	                              .ToLocalChecked();

	  CHECK(obj->IsUndetectable());

	  CHECK(
	      env->Global()->Set(env.local(), v8_str("undetectable"), obj).FromJust());

	  ExpectString("undetectable.toString()", "[object Object]");
	  ExpectString("typeof undetectable", "undefined");
	  ExpectString("typeof(undetectable)", "undefined");
	  ExpectBoolean("typeof undetectable == 'undefined'", true);
	  ExpectBoolean("typeof undetectable == 'object'", false);
	  ExpectBoolean("if (undetectable) { true; } else { false; }", false);
	  ExpectBoolean("!undetectable", true);
	*/
}
