package v8go_test

import (
	"fmt"
	"runtime/cgo"
	"strings"

	v8 "github.com/tommie/v8go"
)

// This example exposes a Go [strings.Builder] to JavaScript as the class
// StringBuilder. Each JavaScript instance stores a [cgo.Handle] to its Go
// object in an internal field.
func ExampleNewValueExternalHandle() {
	iso := v8.NewIsolate()
	defer iso.Dispose()

	// v8go has no callback when a JavaScript object is garbage collected, so
	// the handles are deleted once the context is gone.
	var handles []cgo.Handle
	defer func() {
		for _, h := range handles {
			h.Delete()
		}
	}()

	constructor := v8.NewFunctionTemplate(iso, func(info *v8.FunctionCallbackInfo) *v8.Value {
		h := cgo.NewHandle(&strings.Builder{})
		handles = append(handles, h)
		info.This().SetInternalField(0, v8.NewValueExternalHandle(iso, h))
		return nil
	})
	constructor.InstanceTemplate().SetInternalFieldCount(1)

	proto := constructor.PrototypeTemplate()
	proto.Set("write", v8.NewFunctionTemplateWithError(iso, func(info *v8.FunctionCallbackInfo) (*v8.Value, error) {
		sb, err := builderFromThis(info)
		if err != nil {
			return nil, err
		}
		for _, arg := range info.Args() {
			sb.WriteString(arg.String())
		}
		return nil, nil
	}))
	proto.Set("toString", v8.NewFunctionTemplateWithError(iso, func(info *v8.FunctionCallbackInfo) (*v8.Value, error) {
		sb, err := builderFromThis(info)
		if err != nil {
			return nil, err
		}
		return v8.NewValue(iso, sb.String())
	}))

	global := v8.NewObjectTemplate(iso)
	global.Set("StringBuilder", constructor)
	ctx := v8.NewContext(iso, global)
	defer ctx.Close()

	v, err := ctx.RunScript(`
		const sb = new StringBuilder();
		sb.write("Hello, ", "World");
		sb.write("!");
		sb.toString();`, "builder.js")
	if err != nil {
		panic(err)
	}
	fmt.Println(v.String())

	// An object can inherit the methods without being created by the
	// constructor. It then has no internal field.
	_, err = ctx.RunScript(`
		const fake = { __proto__: StringBuilder.prototype };
		fake.write("oops");`, "fake.js")
	fmt.Println(err)

	// Output:
	// Hello, World!
	// TypeError: Illegal invocation
}

// builderFromThis returns the Go object wrapped by the receiver of a method
// call. It returns a TypeError, as built-in methods do, if the receiver was not
// created by the StringBuilder constructor.
func builderFromThis(info *v8.FunctionCallbackInfo) (*strings.Builder, error) {
	this := info.This()
	if this.InternalFieldCount() < 1 {
		return nil, v8.NewTypeError(info.Context().Isolate(), "Illegal invocation")
	}
	h := this.GetInternalField(0).ExternalHandle()
	if h == 0 {
		return nil, v8.NewTypeError(info.Context().Isolate(), "Illegal invocation")
	}
	return h.Value().(*strings.Builder), nil
}
