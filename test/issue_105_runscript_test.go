package main

import (
	"testing"

	"github.com/tommie/v8go"
)

func TestIssue105RunScript(t *testing.T) {
	iso := v8go.NewIsolate()
	ctx := v8go.NewContext(iso)

	if _, err := ctx.RunScript("const multiply = (a, b) => a * b", "math.js"); err != nil {
		t.Fatalf("RunScript(const multiply) failed: %v", err)
	}

	got, err := ctx.RunScript("multiply(3, 4)", "math.js")
	if err != nil {
		t.Fatalf("RunScript(multiply call) failed: %v", err)
	}

	if want := float64(3 * 4); got.Number() != want {
		t.Errorf("RunScript: got %v, want %v", got, want)
	}
}
