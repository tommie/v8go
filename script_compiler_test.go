package v8go_test

import (
	"errors"
	"reflect"
	"strings"
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

	mod, err := v8.CompileModule(ctx, `
		print("42")`, "")
	if err != nil {
		t.Errorf("Unexpected error: %#v", err)
	}
	err = mod.InstantiateModule(ctx, nil)
	if err != nil {
		t.Errorf("Unexpected error: %#v", err)
	}
	val, err := mod.Evaluate(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %#v", err)
	}
	t.Logf("Script id: %d", mod.ScriptID())

	p, _ := val.AsPromise()
	if s := p.State(); s != v8.Fulfilled {
		t.Errorf("Unexpected promise state: expected %q, got %q", v8.Fulfilled, s)
	}
	if !reflect.DeepEqual(lines, []string{"42"}) {
		t.Errorf("Unexpected output, got: %v", lines)
	}
}

func TestScriptCompilerMissingModule(t *testing.T) {
	t.Parallel()

	iso := v8.NewIsolate()
	defer iso.Dispose()
	global := v8.NewObjectTemplate(iso)
	ctx := v8.NewContext(iso, global)
	defer ctx.Close()

	mod, _ := v8.CompileModule(ctx, `import foo from "non-existing-module";`, "")
	modules := Resolver{}
	err := mod.InstantiateModule(ctx, modules)
	if err == nil {
		t.Fatal("Module instantiation should have failed because of missing module")
	}
	expected := "cannot resolve module 'non-existing-module': module not found"
	actual := err.Error()
	if !strings.Contains(actual, expected) {
		t.Errorf("Expected error to contain substring: '%s'. Actual message: %s", expected, actual)
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
		import foo from "a";
		print(1 + foo.a + foo.b)`, "")
	if err != nil {
		t.Errorf("Expected an error running script module: %v", err)
		return
	}
	modules := Resolver{
		"a": "import b from 'b'; export default { a: 2, b };",
		"b": "export default 3",
	}
	err = mod.InstantiateModule(ctx, LoggingResolver{modules, t})
	if err != nil {
		t.Errorf("Unexpected error: %#v", err)
		return
	}
	val, err := mod.Evaluate(ctx)
	if err != nil {
		t.Errorf("Expected an error running script module: %v", err)
	}
	p, _ := val.AsPromise()
	if s := p.State(); s != v8.Fulfilled {
		t.Errorf("Unexpected promise state: expected %q, got %q", v8.Fulfilled, s)
	}
	if !reflect.DeepEqual(lines, []string{"6"}) {
		t.Errorf("Unexpected output, got: %v", lines)
	}
}

type LoggingResolver struct {
	Resolver Resolver
	t        *testing.T
}

func (r LoggingResolver) ResolveModule(
	ctx *v8.Context,
	spec string,
	referrer *v8.Module,
) (*v8.Module, error) {
	r.t.Logf("Module ref: %#v", referrer)
	r.t.Logf("ResolveModule. ref: %d, spec: %s", referrer.ScriptID(), spec)
	res, err := r.Resolver.ResolveModule(ctx, spec, referrer)
	if err == nil {
		r.t.Logf("Module compiled. ref: %d, spec: %s", res.ScriptID(), spec)
	}
	return res, err
}

type Resolver map[string]string

func (r Resolver) ResolveModule(
	ctx *v8.Context,
	spec string,
	referrer *v8.Module,
) (*v8.Module, error) {
	script, found := r[spec]
	if !found {
		return nil, errors.New("module not found")
	}
	return v8.CompileModule(ctx, script, "")
}
