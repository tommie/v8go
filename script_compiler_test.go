package v8go_test

import (
	"testing"

	v8 "github.com/tommie/v8go"
)

func TestScriptCompilerModuleWithoutImports(t *testing.T) {
	t.Parallel()

	iso := v8.NewIsolate()
	defer iso.Dispose()
	ctx := v8.NewContext(iso)
	defer ctx.Close()

	val, err := v8.CompileModule(ctx, `
		export default 1 + 1;`, "")
	if err != nil {
		t.Errorf("Unexpected error: %#v", err)
	}

	p, _ := val.AsPromise()
	if s := p.State(); s != v8.Fulfilled {
		t.Errorf("Unexpected promise state: expected %q, got %q", v8.Fulfilled, s)
	}
	val = p.Result()
	// How to read values from the module?
	// We get a fulfilled promise, but it's undefined?
	if rtn := val.String(); rtn != "1" {
		t.Errorf("script returned an unexpected value: expected %q, got %q", "2", rtn)
	}
}

func TestScriptCompilerImportingNonExistingModule(t *testing.T) {
	t.Parallel()

	iso := v8.NewIsolate()
	defer iso.Dispose()
	ctx := v8.NewContext(iso)
	defer ctx.Close()

	_, err := v8.CompileModule(ctx, `
		import foo from "missing";
		1 + 1;`, "")

	if err == nil {
		t.Error("Expected an error running script module")
	}
}
