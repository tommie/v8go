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
	context := v8.NewContext(iso)
	defer context.Close()
	inspector.ContextCreated(context)
	defer inspector.ContextDestroyed(context)
	_, err := context.RunScript("console.log('Hello, world!')", "")
	if err != nil {
		t.Error("Error occurred: " + err.Error())
		return
	}
	if len(recorder.messages) != 1 {
		t.Error("Expected exactly one message")
	} else if recorder.messages[0].Message != "Hello, world!" {
		t.Error("Expected Hello, World")
	}
}
