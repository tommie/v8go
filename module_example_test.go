package v8go_test

import (
	"fmt"

	v8 "github.com/tommie/v8go"
)

func ExampleIsolate_CompileModule() {
	iso := v8.NewIsolate()
	defer iso.Dispose()
	ctx := v8.NewContext(iso)
	defer ctx.Close()

	// The resolver caches modules, so each specifier resolves to the same
	// Module.
	sources := map[string]string{"math.js": "export const multiply = (a, b) => a * b"}
	modules := map[string]*v8.Module{}
	resolver := v8.ModuleResolverFunc(func(ctx *v8.Context, spec string, attrs []v8.ImportAttribute, referrer *v8.Module) (*v8.Module, error) {
		if mod, ok := modules[spec]; ok {
			return mod, nil
		}
		mod, err := ctx.Isolate().CompileModule(sources[spec], spec)
		modules[spec] = mod
		return mod, err
	})

	mod, err := iso.CompileModule(`import { multiply } from "math.js"; export const result = multiply(3, 4);`, "main.js")
	if err != nil {
		panic(err)
	}
	if err := mod.Instantiate(ctx, resolver); err != nil {
		panic(err)
	}
	// The promise is settled, unless the module uses top-level await.
	if _, err := mod.Evaluate(ctx); err != nil {
		panic(err)
	}
	ns, err := mod.Namespace(ctx)
	if err != nil {
		panic(err)
	}
	result, err := ns.Get("result")
	if err != nil {
		panic(err)
	}
	fmt.Println(result)
	// Output: 12
}
