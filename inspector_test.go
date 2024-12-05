package v8go_test

import (
	"testing"

	v8 "github.com/tommie/v8go"
)

type consoleAPIMessageRecorder struct {
	messages []v8.ConsoleAPIMessage
}

func (r *consoleAPIMessageRecorder) ConsoleAPIMessage(msg v8.ConsoleAPIMessage) {
	r.messages = append(r.messages, msg)
}

func TestMonitorCreateDispose(t *testing.T) {
	recorder := consoleAPIMessageRecorder{}
	t.Parallel()
	iso := v8.NewIsolate()
	defer iso.Dispose()
	client := v8.NewInspectorClient(&recorder)
	defer client.Dispose()
	inspector := v8.NewInspector(iso, client)
	defer inspector.Dispose()
}
