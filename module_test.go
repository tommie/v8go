package v8go_test

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

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

	mod, err := v8.CompileModule(iso, `
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

	mod, _ := v8.CompileModule(iso, `import foo from "non-existing-module";`, "")
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

	mod, err := v8.CompileModule(iso, `
		import foo from "a";
		print(1 + foo.a + foo.b)`, "")
	if err != nil {
		t.Errorf("Expected an error running script module: %v", err)
		return
	}
	t.Logf("Compiled. Status: %d", mod.GetStatus())
	modules := Resolver{
		"a": "import b from 'b'; export default { a: 2, b };",
		"b": "export default 3",
	}
	err = mod.InstantiateModule(ctx, LoggingResolver{modules, t})
	t.Logf("Instantiated. Status: %d", mod.GetStatus())
	if err != nil {
		t.Errorf("Unexpected error: %#v", err)
		return
	}
	val, err := mod.Evaluate(ctx)
	if err != nil {
		t.Errorf("Expected an error running script module: %v", err)
	}
	t.Logf("Evaluated. Status: %d", mod.GetStatus())
	p, _ := val.AsPromise()
	if s := p.State(); s != v8.Fulfilled {
		t.Errorf("Unexpected promise state: expected %q, got %q", v8.Fulfilled, s)
	}
	if !reflect.DeepEqual(lines, []string{"6"}) {
		t.Errorf("Unexpected output, got: %v", lines)
	}
}

func TestSameModuleImportedMultipleTimes(t *testing.T) {
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

	modules := Resolver{
		"./c.js": `
			let val = 0;
			export const inc = () => {
				val++;
				return val;
			}
		`,
		"./a.js": `
		import { inc } from './c.js' with { key: "data" }; 
			export default inc();`,
		"./b.js": "import { inc } from './c.js'; export default inc();",
	}
	mod, err := v8.CompileModule(iso, `
		import a from "./a.js";
		import b from "./b.js";
		export const result = a + b;
		print(a + b)`, "https://example.com/root.js")
	t.Log("Module status after compile", mod.GetStatus())
	if err != nil {
		t.Errorf("Expected an error running script module: %v", err)
		return
	}
	err = mod.InstantiateModule(ctx, LoggingResolver{InitResolver(modules), t})
	t.Logf("Instantiated. Status: %d", mod.GetStatus())
	if err != nil {
		t.Errorf("Unexpected error: %#v", err)
		return
	}
	val, err := mod.Evaluate(ctx)
	if err != nil {
		t.Errorf("Expected an error running script module: %v", err)
	}
	p, err := val.AsPromise()
	if err != nil {
		t.Errorf("Error getting module as promise: %v", err)
	}
	// moduleEval, err := WaitForPromise(ctx, p)
	if err != nil {
		t.Errorf("Error resolving module promise: %v", err)
	}
	if s := p.State(); s != v8.Fulfilled {
		t.Errorf("Unexpected promise state: expected %q, got %q", v8.Fulfilled, s)
	}
	if !reflect.DeepEqual(lines, []string{"3"}) {
		t.Errorf("Unexpected output, got: %v", lines)
	}
	// ns := mod.GetModuleNamespace()
	// t.Log("Module eval", moduleEval.IsModuleNamespaceObject())
	// t.Log("Module status", mod.GetStatus())
	// t.Log("Module namespace", ns.IsModuleNamespaceObject())
	// t.Log("Module is object", ns.IsObject())
	// t.Log("Module desc", ns.DetailString())
	// obj, err := ns.AsObject()
	// fatalIf(t, err)
	// res, err := obj.Get("result")
	// fatalIf(t, err)
	// t.Log("Val", res.String())
	t.Error("ping")
}

type LoggingResolver struct {
	Resolver v8.ResolveModuler
	t        *testing.T
}

func (r LoggingResolver) ResolveModule(
	ctx *v8.Context,
	spec string,
	attr v8.ImportAttributes,
	referrer *v8.Module,
) (*v8.Module, error) {
	r.t.Logf(
		"ResolveModule. spec: %s -  ref: %d, Attr.len, %d",
		spec,
		referrer.ScriptID(),
		attr.Length(ctx),
	)
	res, err := r.Resolver.ResolveModule(ctx, spec, attr, referrer)
	if err == nil {
		r.t.Logf("Module compiled. ref: %d, status: %d", res.ScriptID(), res.GetStatus())
	}
	return res, err
}

type Resolver map[string]string

func (r Resolver) ResolveModule(
	ctx *v8.Context,
	spec string,
	attr v8.ImportAttributes,
	referrer *v8.Module,
) (*v8.Module, error) {
	script, found := r[spec]
	if !found {
		return nil, errors.New("module not found")
	}
	return v8.CompileModule(ctx.Isolate(), script, spec)
}

type CachingModuleResolver struct {
	Resolver Resolver
	Modules  map[string]*v8.Module
}

func (r *CachingModuleResolver) ResolveModule(
	ctx *v8.Context,
	spec string,
	attr v8.ImportAttributes,
	referrer *v8.Module,
) (*v8.Module, error) {
	fmt.Println("Look for module", spec)
	if module, found := r.Modules[spec]; found {
		fmt.Println("Found!")
		return module, nil
	}
	script, found := r.Resolver[spec]
	if !found {
		return nil, errors.New("module source not found")
	}
	module, err := v8.CompileModule(ctx.Isolate(), script, spec)
	if err == nil {
		r.Modules[spec] = module
	}
	return module, err
}

func InitResolver(sources map[string]string) v8.ResolveModuler {
	return &CachingModuleResolver{Resolver: sources, Modules: make(map[string]*v8.Module)}
}

func WaitForPromise(ctx *v8.Context, p *v8.Promise) (*v8.Value, error) {
	timeout := time.After(time.Second)
	resolve := make(chan *v8.Value)
	reject := make(chan error)

	p.Then(func(info *v8.FunctionCallbackInfo) *v8.Value {
		args := info.Args()
		if len(args) > 0 {
			resolve <- args[0]
			return args[0]
		} else {
			resolve <- nil
			return nil
		}
	}, func(info *v8.FunctionCallbackInfo) *v8.Value {
		args := info.Args()
		if len(args) > 0 && args[0] != nil {
			val := args[0]
			if val.IsNativeError() {
				exc, _ := val.AsException()
				reject <- exc
			} else {
				reject <- fmt.Errorf("Non-Error rejection: %s", val.String())
			}
		} else {
			reject <- errors.New("Suspicious reject call. No argument provided")
		}
		return nil
	})
	go ctx.PerformMicrotaskCheckpoint()

	select {
	case <-timeout:
		return nil, errors.New("Timeout waiting for promise")
	case val := <-resolve:
		return val, nil
	case err := <-reject:
		return nil, err
	}
}
