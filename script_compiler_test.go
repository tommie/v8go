package v8go_test

import (
	"testing"

	v8 "github.com/tommie/v8go"
)

func TestScriptCompilerModuleWithoutImports(t *testing.T) {
	// t.Parallel()

	iso := v8.NewIsolate()
	defer iso.Dispose()
	ctx := v8.NewContext(iso)
	defer ctx.Close()

	res, err := v8.CompileModule(ctx)

	if err != nil {
		t.Log("ERROR")
		t.Log(err.Error())
		t.Errorf("Unexpected error: %#v", err)
	}
	t.Errorf("Res: %v", res)
}
