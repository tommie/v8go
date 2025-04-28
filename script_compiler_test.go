package v8go_test

import (
	"reflect"
	"testing"

	v8 "github.com/tommie/v8go"
)

func TestScriptCompilerModuleWithoutImports(t *testing.T) {
	t.Parallel()

	iso := v8.NewIsolate()
	defer iso.Dispose()

	var lines []string
	ft := v8.NewFunctionTemplateWithError(
		iso,
		func(info *v8.FunctionCallbackInfo) (*v8.Value, error) {
			lines = append(lines, info.Args()[0].String())
			return nil, nil
		},
	)
	global := v8.NewObjectTemplate(iso)
	global.Set("print", ft)

	ctx := v8.NewContext(iso, global)
	defer ctx.Close()

	mod, _ := v8.CompileModule(ctx, `
		print("42")`, "")
	val, err := mod.Evaluate(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %#v", err)
	}

	p, _ := val.AsPromise()
	if s := p.State(); s != v8.Fulfilled {
		t.Errorf("Unexpected promise state: expected %q, got %q", v8.Fulfilled, s)
	}
	if !reflect.DeepEqual(lines, []string{"42"}) {
		t.Errorf("Unexpected output, got: %v", lines)
	}
}

func TestScriptCompilerImportingNonExistingModule(t *testing.T) {
	t.Parallel()

	iso := v8.NewIsolate()
	defer iso.Dispose()
	ctx := v8.NewContext(iso)
	defer ctx.Close()

	// mod, err := v8.CompileModule(ctx, `
	// 	import foo from "missing";
	// 	1 + 1;`, "")
	// if err != nil {
	// 	t.Errorf("Expected an error running script module: %v", err)
	// }
	// _, err = mod.Evaluate(ctx)
	// if err != nil {
	// 	t.Errorf("Expected an error running script module: %v", err)
	// }
}
