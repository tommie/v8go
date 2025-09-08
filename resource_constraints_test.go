// Copyright 2024 Roger Chapman and the v8go contributors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package v8go_test

import (
	"testing"

	"github.com/tommie/v8go"
)

func TestNewIsolateWithConstraints(t *testing.T) {
	t.Parallel()

	iso := v8go.NewIsolate(v8go.WithResourceConstraints(&v8go.ResourceConstraints{
		8 * 1024 * 1024,
		16 * 1024 * 1024,
	}))
	defer iso.Dispose()

	ctx := v8go.NewContext(iso)
	defer ctx.Close()

	// First test - should work fine
	val, err := ctx.RunScript("1 + 2", "test.js")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !val.IsNumber() || val.Number() != 3 {
		t.Errorf("expected 3, got %v", val)
	}

	// Test
	val, err = ctx.RunScript(`
			const data = [];
			for (let i = 0; i < 1000 * 1000; i++) {
					data.push("large data chunk ".repeat(1000));
			}
			data.length;
		`, "memory-test.js")
	if err != nil {
		t.Logf("Memory test correctly returned error: %v", err)
	} else {
		t.Fatalf("Memory test completed unexpectedly: %v", val)
	}
}

func TestNewIsolateWithNilConstraints(t *testing.T) {
	t.Parallel()

	// Test that not providing constraints works (default behavior)
	iso := v8go.NewIsolate()
	defer iso.Dispose()

	// Test that the isolate was created successfully
	ctx := v8go.NewContext(iso)
	defer ctx.Close()

	// Run a simple script to verify the isolate works
	val, err := ctx.RunScript("'hello world'", "test.js")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !val.IsString() || val.String() != "hello world" {
		t.Errorf("expected 'hello world', got %v", val)
	}
}
