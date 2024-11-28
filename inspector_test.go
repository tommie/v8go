package v8go_test

import (
	"testing"

	v8 "github.com/tommie/v8go"
)

func TestMonitorCreateDispose(t *testing.T) {
	t.Parallel()
	iso := v8.NewIsolate()
	defer iso.Dispose()
	client := v8.NewInspectorClient()
	defer client.Dispose()
	inspector := v8.NewInspector(iso, client)
	defer inspector.Dispose()
}
