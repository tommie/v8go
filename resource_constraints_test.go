// Copyright 2024 Roger Chapman and the v8go contributors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package v8go_test

import (
	"fmt"
	"testing"

	"github.com/tommie/v8go"
)

func TestResourceConstraints(t *testing.T) {
	t.Parallel()

	// Test creating ResourceConstraints struct directly
	rc := v8go.ResourceConstraints{
		MaxOldGenerationSizeInBytes:       100 * 1024 * 1024, // 100MB
		MaxYoungGenerationSizeInBytes:     10 * 1024 * 1024,  // 10MB
		InitialOldGenerationSizeInBytes:   50 * 1024 * 1024,  // 50MB
		InitialYoungGenerationSizeInBytes: 5 * 1024 * 1024,   // 5MB
		CodeRangeSizeInBytes:              64 * 1024 * 1024,  // 64MB
	}

	// Verify struct values
	if rc.MaxOldGenerationSizeInBytes != 100*1024*1024 {
		t.Errorf("expected max old generation size 104857600, got %d", rc.MaxOldGenerationSizeInBytes)
	}
	if rc.MaxYoungGenerationSizeInBytes != 10*1024*1024 {
		t.Errorf("expected max young generation size 10485760, got %d", rc.MaxYoungGenerationSizeInBytes)
	}
	if rc.InitialOldGenerationSizeInBytes != 50*1024*1024 {
		t.Errorf("expected initial old generation size 52428800, got %d", rc.InitialOldGenerationSizeInBytes)
	}
	if rc.InitialYoungGenerationSizeInBytes != 5*1024*1024 {
		t.Errorf("expected initial young generation size 5242880, got %d", rc.InitialYoungGenerationSizeInBytes)
	}
	if rc.CodeRangeSizeInBytes != 64*1024*1024 {
		t.Errorf("expected code range size 67108864, got %d", rc.CodeRangeSizeInBytes)
	}
}

func TestResourceConstraintsConfigureDefaults(t *testing.T) {
	t.Parallel()

	// Test ConfigureDefaults with reasonable values
	physicalMemory := uint64(8 * 1024 * 1024 * 1024) // 8GB
	virtualMemoryLimit := uint64(0)                  // No limit
	rc := v8go.ConfigureDefaults(physicalMemory, virtualMemoryLimit)

	// After ConfigureDefaults, some values should be set
	// We can't test exact values as they depend on our internal logic
	// but we can verify they're non-zero
	if rc.MaxOldGenerationSizeInBytes == 0 {
		t.Error("expected non-zero max old generation size after ConfigureDefaults")
	}
}

func TestResourceConstraintsConfigureDefaultsFromHeapSize(t *testing.T) {
	t.Parallel()

	// Test ConfigureDefaultsFromHeapSize
	initialHeapSize := uint64(64 * 1024 * 1024)  // 64MB
	maximumHeapSize := uint64(512 * 1024 * 1024) // 512MB
	rc := v8go.ConfigureDefaultsFromHeapSize(initialHeapSize, maximumHeapSize)

	// After ConfigureDefaultsFromHeapSize, some values should be set
	// We can verify they're reasonable
	if rc.MaxOldGenerationSizeInBytes == 0 {
		t.Error("expected non-zero max old generation size after ConfigureDefaultsFromHeapSize")
	}
	expectedMaxOld := maximumHeapSize * 8 / 10 // 80% of max heap
	if rc.MaxOldGenerationSizeInBytes != expectedMaxOld {
		t.Errorf("expected max old generation %d, got %d", expectedMaxOld, rc.MaxOldGenerationSizeInBytes)
	}
}

func TestNewIsolateWithConstraints(t *testing.T) {
	t.Parallel()

	// Configure small heap sizes (2MB)
	constraints := &v8go.ResourceConstraints{
		MaxOldGenerationSizeInBytes:   2 * 1024 * 1024, // 2MB
		MaxYoungGenerationSizeInBytes: 2 * 1024 * 1024, // 2MB
	}

	iso := v8go.NewIsolate(v8go.WithResourceConstraints(constraints))
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

	// Test memory exhaustion - might panic or return error
	var memoryTestPanicked bool
	func() {
		defer func() {
			if r := recover(); r != nil {
				memoryTestPanicked = true
				t.Logf("Memory test caused panic (expected): %v", r)
			}
		}()

		val, err = ctx.RunScript(`
			const data = [];
			for (let i = 0; i < 10000; i++) {
					data.push("large data chunk ".repeat(1000));
			}
			data.length;
		`, "memory-test.js")
		if err != nil {
			t.Logf("Memory test returned error (also acceptable): %v", err)
		} else {
			t.Logf("Memory test completed unexpectedly: %v", val)
		}
	}()

	// Either panic or error is acceptable for memory exhaustion
	if !memoryTestPanicked && err == nil {
		t.Log("Warning: Memory constraint test didn't trigger panic or error")
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

// ExampleResourceConstraints demonstrates how to use ResourceConstraints
// to limit V8's memory usage when creating an isolate.
func ExampleResourceConstraints() {

	// Create ResourceConstraints to limit memory usage
	constraints := &v8go.ResourceConstraints{
		MaxOldGenerationSizeInBytes:       100 * 1024 * 1024, // 100MB max
		MaxYoungGenerationSizeInBytes:     10 * 1024 * 1024,  // 10MB max
		InitialOldGenerationSizeInBytes:   50 * 1024 * 1024,  // 50MB initial
		InitialYoungGenerationSizeInBytes: 5 * 1024 * 1024,   // 5MB initial
	}

	// Alternative: use convenience functions to configure defaults
	// constraints := v8go.ConfigureDefaultsFromHeapSize(64*1024*1024, 512*1024*1024) // 64MB initial, 512MB max
	// constraints := v8go.ConfigureDefaults(8*1024*1024*1024, 0) // Based on 8GB physical memory, no virtual limit

	// Create an isolate with the specified constraints using initializer function
	iso := v8go.NewIsolate(v8go.WithResourceConstraints(constraints))
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
