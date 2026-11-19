package v8go_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	v8 "github.com/tommie/v8go"
)

// newModuleContext returns a context with a print function, whose
// arguments are appended to the returned slice.
func newModuleContext(t *testing.T) (*v8.Context, *[]string) {
	t.Helper()

	iso := v8.NewIsolate()
	t.Cleanup(iso.Dispose)

	var lines []string
	print := v8.NewFunctionTemplate(iso, func(info *v8.FunctionCallbackInfo) *v8.Value {
		lines = append(lines, info.Args()[0].String())
		return nil
	})
	global := v8.NewObjectTemplate(iso)
	fatalIf(t, global.Set("print", print))

	ctx := v8.NewContext(iso, global)
	t.Cleanup(ctx.Close)
	return ctx, &lines
}

var errModuleNotFound = errors.New("module not found")

// mapResolver compiles modules from a map of sources. Compiled modules
// are cached, so each specifier resolves to the same Module.
type mapResolver struct {
	sources  map[string]string
	modules  map[string]*v8.Module
	resolved []string
}

func newMapResolver(sources map[string]string) *mapResolver {
	return &mapResolver{sources: sources, modules: map[string]*v8.Module{}}
}

func (r *mapResolver) ResolveModule(ctx *v8.Context, spec string, attrs []v8.ImportAttribute, referrer *v8.Module) (*v8.Module, error) {
	r.resolved = append(r.resolved, spec)
	if mod, ok := r.modules[spec]; ok {
		return mod, nil
	}
	src, ok := r.sources[spec]
	if !ok {
		return nil, errModuleNotFound
	}
	mod, err := ctx.Isolate().CompileModule(src, spec)
	if err != nil {
		return nil, err
	}
	r.modules[spec] = mod
	return mod, nil
}

// runModule compiles, instantiates and evaluates a module, and returns
// the settled evaluation promise.
func runModule(t *testing.T, ctx *v8.Context, src string, resolver v8.ModuleResolver) (*v8.Module, *v8.Promise) {
	t.Helper()

	mod, err := ctx.Isolate().CompileModule(src, "main.js")
	fatalIf(t, err)
	fatalIf(t, mod.Instantiate(ctx, resolver))
	p, err := mod.Evaluate(ctx)
	fatalIf(t, err)
	ctx.PerformMicrotaskCheckpoint()
	return mod, p
}

func TestModuleWithoutImports(t *testing.T) {
	t.Parallel()

	ctx, lines := newModuleContext(t)
	mod, p := runModule(t, ctx, `print("42")`, nil)

	if got, want := p.State(), v8.Fulfilled; got != want {
		t.Errorf("State: got %v, want %v", got, want)
	}
	if got, want := mod.Status(), v8.ModuleEvaluated; got != want {
		t.Errorf("Status: got %v, want %v", got, want)
	}
	if got, want := *lines, []string{"42"}; !reflect.DeepEqual(got, want) {
		t.Errorf("lines: got %v, want %v", got, want)
	}
}

func TestModuleStatus(t *testing.T) {
	t.Parallel()

	ctx, _ := newModuleContext(t)
	mod, err := ctx.Isolate().CompileModule(`export default 1`, "main.js")
	fatalIf(t, err)
	if got, want := mod.Status(), v8.ModuleUninstantiated; got != want {
		t.Errorf("Status after compile: got %v, want %v", got, want)
	}
	fatalIf(t, mod.Instantiate(ctx, nil))
	if got, want := mod.Status(), v8.ModuleInstantiated; got != want {
		t.Errorf("Status after Instantiate: got %v, want %v", got, want)
	}
	if got, want := v8.ModuleEvaluated.String(), "evaluated"; got != want {
		t.Errorf("String: got %q, want %q", got, want)
	}
}

func TestModuleCompileSyntaxError(t *testing.T) {
	t.Parallel()

	ctx, _ := newModuleContext(t)
	_, err := ctx.Isolate().CompileModule(`export default {`, "main.js")
	var jsErr *v8.JSError
	if !errors.As(err, &jsErr) {
		t.Fatalf("CompileModule: got %v, want a *JSError", err)
	}
	if !strings.Contains(jsErr.Message, "SyntaxError") {
		t.Errorf("Message: got %q, want a SyntaxError", jsErr.Message)
	}
	if !strings.Contains(jsErr.Location, "main.js") {
		t.Errorf("Location: got %q, want main.js", jsErr.Location)
	}
}

func TestModuleNamespace(t *testing.T) {
	t.Parallel()

	ctx, _ := newModuleContext(t)
	mod, err := ctx.Isolate().CompileModule(`
		export const strVal = "str";
		export const numVal = 42;
		export default "Default export";
	`, "main.js")
	fatalIf(t, err)

	if _, err := mod.Namespace(ctx); err == nil {
		t.Error("Namespace before Instantiate: got nil error, want an error")
	}

	fatalIf(t, mod.Instantiate(ctx, nil))
	_, err = mod.Evaluate(ctx)
	fatalIf(t, err)

	ns, err := mod.Namespace(ctx)
	fatalIf(t, err)
	if !ns.IsModuleNamespaceObject() {
		t.Error("IsModuleNamespaceObject: got false, want true")
	}

	for key, want := range map[string]string{
		"strVal":  "str",
		"numVal":  "42",
		"default": "Default export",
	} {
		got, err := ns.Get(key)
		fatalIf(t, err)
		if got.String() != want {
			t.Errorf("Get(%q): got %q, want %q", key, got, want)
		}
	}
}

func TestModuleImports(t *testing.T) {
	t.Parallel()

	ctx, lines := newModuleContext(t)
	r := newMapResolver(map[string]string{
		"a": `import b from "b"; export default { a: 2, b };`,
		"b": `export default 3;`,
	})
	_, p := runModule(t, ctx, `
		import foo from "a";
		print(1 + foo.a + foo.b);
	`, r)

	if got, want := p.State(), v8.Fulfilled; got != want {
		t.Errorf("State: got %v, want %v", got, want)
	}
	if got, want := *lines, []string{"6"}; !reflect.DeepEqual(got, want) {
		t.Errorf("lines: got %v, want %v", got, want)
	}
}

func TestModuleImportedTwiceEvaluatesOnce(t *testing.T) {
	t.Parallel()

	ctx, lines := newModuleContext(t)
	r := newMapResolver(map[string]string{
		"./c.js": `
			let val = 0;
			export const inc = () => ++val;
		`,
		"./a.js": `import { inc } from "./c.js"; export default inc();`,
		"./b.js": `import { inc } from "./c.js"; export default inc();`,
	})
	runModule(t, ctx, `
		import a from "./a.js";
		import b from "./b.js";
		print(a + b);
	`, r)

	if got, want := *lines, []string{"3"}; !reflect.DeepEqual(got, want) {
		t.Errorf("lines: got %v, want %v", got, want)
	}
}

func TestModuleCyclicImports(t *testing.T) {
	t.Parallel()

	ctx, lines := newModuleContext(t)
	r := newMapResolver(map[string]string{
		"a": `import { b } from "b"; export const a = () => "a" + b();`,
		"b": `import { a } from "a"; export const b = () => "b"; export const ab = () => a();`,
	})
	_, p := runModule(t, ctx, `
		import { ab } from "b";
		print(ab());
	`, r)

	if got, want := p.State(), v8.Fulfilled; got != want {
		t.Errorf("State: got %v, want %v", got, want)
	}
	if got, want := *lines, []string{"ab"}; !reflect.DeepEqual(got, want) {
		t.Errorf("lines: got %v, want %v", got, want)
	}
}

func TestModuleResolverReferrer(t *testing.T) {
	t.Parallel()

	ctx, _ := newModuleContext(t)
	iso := ctx.Isolate()
	mod, err := iso.CompileModule(`import "a";`, "main.js")
	fatalIf(t, err)
	a, err := iso.CompileModule(`import "b";`, "a")
	fatalIf(t, err)
	b, err := iso.CompileModule(``, "b")
	fatalIf(t, err)

	referrers := map[string]*v8.Module{}
	err = mod.Instantiate(ctx, v8.ModuleResolverFunc(func(ctx *v8.Context, spec string, attrs []v8.ImportAttribute, referrer *v8.Module) (*v8.Module, error) {
		referrers[spec] = referrer
		return map[string]*v8.Module{"a": a, "b": b}[spec], nil
	}))
	fatalIf(t, err)

	if got, want := referrers, map[string]*v8.Module{"a": mod, "b": a}; !reflect.DeepEqual(got, want) {
		t.Errorf("referrers: got %v, want %v", got, want)
	}
}

func TestModuleImportAttributes(t *testing.T) {
	t.Parallel()

	ctx, _ := newModuleContext(t)
	src := `import "a" with { type: "json", foo: "bar" };`
	mod, err := ctx.Isolate().CompileModule(src, "main.js")
	fatalIf(t, err)
	a, err := ctx.Isolate().CompileModule(``, "a")
	fatalIf(t, err)

	var got []v8.ImportAttribute
	err = mod.Instantiate(ctx, v8.ModuleResolverFunc(func(ctx *v8.Context, spec string, attrs []v8.ImportAttribute, referrer *v8.Module) (*v8.Module, error) {
		got = attrs
		return a, nil
	}))
	fatalIf(t, err)

	want := []v8.ImportAttribute{
		{Key: "foo", Value: "bar", SourceOffset: strings.Index(src, "foo")},
		{Key: "type", Value: "json", SourceOffset: strings.Index(src, "type")},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("attrs: got %+v, want %+v", got, want)
	}
}

func TestModuleResolverErrors(t *testing.T) {
	t.Parallel()

	tsts := []struct {
		name     string
		resolver v8.ModuleResolver
		want     string
		wantIs   error
	}{
		{
			name:     "notFound",
			resolver: newMapResolver(nil),
			want:     `Error: cannot resolve module "a": module not found`,
			wantIs:   errModuleNotFound,
		},
		{
			name: "valueError",
			resolver: v8.ModuleResolverFunc(func(ctx *v8.Context, spec string, attrs []v8.ImportAttribute, referrer *v8.Module) (*v8.Module, error) {
				return nil, v8.NewTypeError(ctx.Isolate(), "custom")
			}),
			want: "TypeError: custom",
		},
		{
			name: "nilModule",
			resolver: v8.ModuleResolverFunc(func(ctx *v8.Context, spec string, attrs []v8.ImportAttribute, referrer *v8.Module) (*v8.Module, error) {
				return nil, nil
			}),
			want: `Error: cannot resolve module "a": resolver returned no module`,
		},
		{
			name:     "noResolver",
			resolver: nil,
			want:     `Error: cannot resolve module "a": no module resolver`,
			wantIs:   v8.ErrNoModuleResolver,
		},
		{
			name: "otherIsolate",
			resolver: v8.ModuleResolverFunc(func(ctx *v8.Context, spec string, attrs []v8.ImportAttribute, referrer *v8.Module) (*v8.Module, error) {
				iso := v8.NewIsolate()
				defer iso.Dispose()
				return iso.CompileModule(``, spec)
			}),
			want: `Error: cannot resolve module "a": resolver returned a module from a different isolate`,
		},
	}
	for _, tst := range tsts {
		tst := tst
		t.Run(tst.name, func(t *testing.T) {
			t.Parallel()

			ctx, _ := newModuleContext(t)
			mod, err := ctx.Isolate().CompileModule(`import "a";`, "main.js")
			fatalIf(t, err)

			err = mod.Instantiate(ctx, tst.resolver)
			if err == nil {
				t.Fatal("Instantiate: got nil error, want an error")
			}
			if got := err.Error(); got != tst.want {
				t.Errorf("Instantiate: got %q, want %q", got, tst.want)
			}
			if tst.wantIs != nil && !errors.Is(err, tst.wantIs) {
				t.Errorf("Instantiate: got %v, want errors.Is %v", err, tst.wantIs)
			}
			var exc *v8.Exception
			if got, want := errors.As(err, &exc), tst.name == "valueError"; got != want {
				t.Errorf("errors.As(*Exception): got %v, want %v", got, want)
			}
		})
	}
}

func TestModuleEvaluateThrows(t *testing.T) {
	t.Parallel()

	ctx, _ := newModuleContext(t)
	mod, p := runModule(t, ctx, `throw new Error("boom");`, nil)

	if got, want := p.State(), v8.Rejected; got != want {
		t.Fatalf("State: got %v, want %v", got, want)
	}
	if got, want := p.Result().String(), "Error: boom"; got != want {
		t.Errorf("Result: got %q, want %q", got, want)
	}
	if got, want := mod.Status(), v8.ModuleErrored; got != want {
		t.Errorf("Status: got %v, want %v", got, want)
	}
}

func TestModuleTopLevelAwait(t *testing.T) {
	t.Parallel()

	ctx, lines := newModuleContext(t)
	// Microtasks run as Evaluate returns, so the module awaits a
	// promise only Go resolves.
	gate, err := v8.NewPromiseResolver(ctx)
	fatalIf(t, err)
	fatalIf(t, ctx.Global().Set("gate", gate.GetPromise()))

	mod, err := ctx.Isolate().CompileModule(`
		print(await gate);
	`, "main.js")
	fatalIf(t, err)
	fatalIf(t, mod.Instantiate(ctx, nil))
	p, err := mod.Evaluate(ctx)
	fatalIf(t, err)

	if got, want := p.State(), v8.Pending; got != want {
		t.Errorf("State before resolving: got %v, want %v", got, want)
	}
	done, err := v8.NewValue(ctx.Isolate(), "done")
	fatalIf(t, err)
	gate.Resolve(done)
	ctx.PerformMicrotaskCheckpoint()
	if got, want := p.State(), v8.Fulfilled; got != want {
		t.Errorf("State after resolving: got %v, want %v", got, want)
	}
	if got, want := *lines, []string{"done"}; !reflect.DeepEqual(got, want) {
		t.Errorf("lines: got %v, want %v", got, want)
	}
}

func TestModuleDifferentIsolatePanics(t *testing.T) {
	t.Parallel()

	ctx, _ := newModuleContext(t)
	iso := v8.NewIsolate()
	defer iso.Dispose()
	mod, err := iso.CompileModule(``, "main.js")
	fatalIf(t, err)

	if recoverPanic(func() { mod.Instantiate(ctx, nil) }) == nil {
		t.Error("Instantiate: got no panic, want a panic")
	}
}
