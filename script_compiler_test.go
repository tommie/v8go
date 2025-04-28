package v8go_test

import (
	"fmt"
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
	err := mod.InstantiateModule(ctx, nil)
	if err != nil {
		t.Errorf("Unexpected error: %#v", err)
	}
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

	mod, err := v8.CompileModule(ctx, `
		import foo from "missing";
		print(1 + foo)`, "")
	if err != nil {
		t.Errorf("Expected an error running script module: %v", err)
		return
	}
	modules := Resolver{
		"missing": "export default 1",
	}
	err = mod.InstantiateModule(ctx, modules)
	if err != nil {
		t.Errorf("Unexpected error: %#v", err)
		return
	}
	_, err = mod.Evaluate(ctx)
	if err != nil {
		t.Errorf("Expected an error running script module: %v", err)
	}
}

type Resolver map[string]string

func (r Resolver) ResolveModule(
	ctx *v8.Context,
	spec string,
	referrer *v8.Module,
) (*v8.Module, error) {
	script, found := r[spec]
	if !found {
		return nil, fmt.Errorf("Cannot find module: %s", spec)
	}
	return v8.CompileModule(ctx, script, "")
}
