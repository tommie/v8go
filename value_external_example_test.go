package v8go_test

import (
	"fmt"
	"strings"

	v8 "github.com/tommie/v8go"
)

// This example exposes a Go [strings.Builder] to JavaScript as the class
// StringBuilder. Each JavaScript instance stores its Go object, wrapped in an
// External, in an internal field.
func ExampleNewValue_external() {
	iso := v8.NewIsolate()
	defer iso.Dispose()

	// Each callback releases its arguments, so the objects can be garbage
	// collected, and the strings.Builders with them.
	constructor := v8.NewFunctionTemplateWithError(iso, func(info *v8.FunctionCallbackInfo) (*v8.Value, error) {
		defer info.Release()

		// Without new, this is the global object, which has no internal
		// fields.
		this := info.This()
		if this.InternalFieldCount() < 1 {
			return nil, v8.NewTypeError(iso, "Constructor StringBuilder requires 'new'")
		}

		ext, err := v8.NewValue(iso, &strings.Builder{})
		if err != nil {
			return nil, err
		}
		defer ext.Release()
		return nil, this.SetInternalField(0, ext)
	})
	constructor.InstanceTemplate().SetInternalFieldCount(1)

	proto := constructor.PrototypeTemplate()
	proto.Set("write", v8.NewFunctionTemplateWithError(iso, func(info *v8.FunctionCallbackInfo) (*v8.Value, error) {
		defer info.Release()

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
		defer info.Release()

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

	_, err = ctx.RunScript(`StringBuilder()`, "call.js")
	fmt.Println(err)

	// Output:
	// Hello, World!
	// TypeError: Illegal invocation
	// TypeError: Constructor StringBuilder requires 'new'
}

// builderFromThis returns the Go object wrapped by the receiver of a method
// call. It returns a TypeError, as built-in methods do, if the receiver was not
// created by the StringBuilder constructor. That includes objects from other
// classes wrapping Go values, so the type is checked.
func builderFromThis(info *v8.FunctionCallbackInfo) (*strings.Builder, error) {
	this := info.This()
	if this.InternalFieldCount() < 1 {
		return nil, v8.NewTypeError(info.Context().Isolate(), "Illegal invocation")
	}

	field := this.GetInternalField(0)
	defer field.Release()
	v, _ := field.External()
	sb, ok := v.(*strings.Builder)
	if !ok {
		return nil, v8.NewTypeError(info.Context().Isolate(), "Illegal invocation")
	}
	return sb, nil
}
