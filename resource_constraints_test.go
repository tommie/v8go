// Copyright 2024 Roger Chapman and the v8go contributors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package v8go_test

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/tommie/v8go"
)

func TestResourceConstraints(t *testing.T) {
	t.Parallel()

	rc := v8go.NewResourceConstraints()
	defer runtime.SetFinalizer(rc, nil) // Allow GC to clean up in tests

	// Test setting and getting heap sizes
	rc.SetMaxOldGenerationSizeInBytes(100 * 1024 * 1024) // 100MB
	if size := rc.MaxOldGenerationSizeInBytes(); size != 100*1024*1024 {
		t.Errorf("expected max old generation size 104857600, got %d", size)
	}

	rc.SetMaxYoungGenerationSizeInBytes(10 * 1024 * 1024) // 10MB
	if size := rc.MaxYoungGenerationSizeInBytes(); size != 10*1024*1024 {
		t.Errorf("expected max young generation size 10485760, got %d", size)
	}

	rc.SetInitialOldGenerationSizeInBytes(50 * 1024 * 1024) // 50MB
	if size := rc.InitialOldGenerationSizeInBytes(); size != 50*1024*1024 {
		t.Errorf("expected initial old generation size 52428800, got %d", size)
	}

	rc.SetInitialYoungGenerationSizeInBytes(5 * 1024 * 1024) // 5MB
	if size := rc.InitialYoungGenerationSizeInBytes(); size != 5*1024*1024 {
		t.Errorf("expected initial young generation size 5242880, got %d", size)
	}

	// Test code range size
	rc.SetCodeRangeSizeInBytes(64 * 1024 * 1024) // 64MB
	if size := rc.CodeRangeSizeInBytes(); size != 64*1024*1024 {
		t.Errorf("expected code range size 67108864, got %d", size)
	}
}

func TestResourceConstraintsConfigureDefaults(t *testing.T) {
	t.Parallel()

	rc := v8go.NewResourceConstraints()
	defer runtime.SetFinalizer(rc, nil) // Allow GC to clean up in tests

	// Test ConfigureDefaults with reasonable values
	physicalMemory := uint64(8 * 1024 * 1024 * 1024) // 8GB
	virtualMemoryLimit := uint64(0)                  // No limit
	rc.ConfigureDefaults(physicalMemory, virtualMemoryLimit)

	// After ConfigureDefaults, some values should be set
	// We can't test exact values as they depend on V8's internal logic
	// but we can verify they're non-zero
	if size := rc.MaxOldGenerationSizeInBytes(); size == 0 {
		t.Error("expected non-zero max old generation size after ConfigureDefaults")
	}
}

func TestResourceConstraintsConfigureDefaultsFromHeapSize(t *testing.T) {
	t.Parallel()

	rc := v8go.NewResourceConstraints()
	defer runtime.SetFinalizer(rc, nil) // Allow GC to clean up in tests

	// Test ConfigureDefaultsFromHeapSize
	initialHeapSize := uint64(64 * 1024 * 1024)  // 64MB
	maximumHeapSize := uint64(512 * 1024 * 1024) // 512MB
	rc.ConfigureDefaultsFromHeapSize(initialHeapSize, maximumHeapSize)

	// After ConfigureDefaultsFromHeapSize, some values should be set
	// We can't test exact values as they depend on V8's internal logic
	// but we can verify they're non-zero
	if size := rc.MaxOldGenerationSizeInBytes(); size == 0 {
		t.Error("expected non-zero max old generation size after ConfigureDefaultsFromHeapSize")
	}
}

func TestNewIsolateWithConstraints(t *testing.T) {
	t.Parallel()

	rc := v8go.NewResourceConstraints()
	defer runtime.SetFinalizer(rc, nil) // Allow GC to clean up in tests

	// Configure reasonable heap sizes
	rc.SetMaxOldGenerationSizeInBytes(100 * 1024 * 1024)  // 100MB
	rc.SetMaxYoungGenerationSizeInBytes(10 * 1024 * 1024) // 10MB

	// Create isolate with constraints
	iso := v8go.NewIsolateWithConstraints(rc)
	defer iso.Dispose()

	// Test that the isolate was created successfully
	ctx := v8go.NewContext(iso)
	defer ctx.Close()

	// Run a simple script to verify the isolate works
	val, err := ctx.RunScript("1 + 2", "test.js")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !val.IsNumber() || val.Number() != 3 {
		t.Errorf("expected 3, got %v", val)
	}
}

func TestNewIsolateWithNilConstraints(t *testing.T) {
	t.Parallel()

	// Test that passing nil constraints works (should behave like NewIsolate)
	iso := v8go.NewIsolateWithConstraints(nil)
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

// ExampleResourceConstraints demonstrates how to use ResourceConstraints
// to limit V8's memory usage when creating an isolate.
func ExampleResourceConstraints() {

	// Create ResourceConstraints to limit memory usage
	constraints := v8go.NewResourceConstraints()

	// Set maximum heap sizes to limit memory usage
	constraints.SetMaxOldGenerationSizeInBytes(100 * 1024 * 1024)  // 100MB max
	constraints.SetMaxYoungGenerationSizeInBytes(10 * 1024 * 1024) // 10MB max

	// Configure initial sizes to avoid garbage collection churn at startup
	constraints.SetInitialOldGenerationSizeInBytes(50 * 1024 * 1024)  // 50MB initial
	constraints.SetInitialYoungGenerationSizeInBytes(5 * 1024 * 1024) // 5MB initial

	// Alternative: use convenience methods to configure defaults
	// constraints.ConfigureDefaultsFromHeapSize(64*1024*1024, 512*1024*1024) // 64MB initial, 512MB max
	// constraints.ConfigureDefaults(8*1024*1024*1024, 0) // Based on 8GB physical memory, no virtual limit

	// Create an isolate with the specified constraints
	iso := v8go.NewIsolateWithConstraints(constraints)
	defer iso.Dispose()

	// Create a context and run JavaScript
	ctx := v8go.NewContext(iso)
	defer ctx.Close()

	script := `
		const data = [];
		for (let i = 0; i < 100; i++) {
			data.push({ id: i, message: "Hello from constrained V8!" });
		}
		data.length;
	`

	result, err := ctx.RunScript(script, "example.js")
	if err != nil {
		panic(err)
	}

	// Get heap statistics to see actual memory usage
	stats := iso.GetHeapStatistics()
	_ = stats // Use stats to check memory usage if needed

	fmt.Print(result.String())
	// Output: 100
}
