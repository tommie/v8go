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

type IsolateWithInspector struct {
	iso             *v8.Isolate
	inspector       *v8.Inspector
	inspectorClient *v8.InspectorClient
}

func NewIsolateWithInspectorClient(handler v8.ConsoleAPIMessageHandler) *IsolateWithInspector {
	iso := v8.NewIsolate()
	client := v8.NewInspectorClient(handler)
	inspector := v8.NewInspector(iso, client)
	return &IsolateWithInspector{
		iso,
		inspector,
		client,
	}
}

func (iso *IsolateWithInspector) Dispose() {
	iso.inspector.Dispose()
	iso.inspectorClient.Dispose()
	iso.iso.Dispose()
}

type ContextWithInspector struct {
	*v8.Context
	iso *IsolateWithInspector
}

func (iso *IsolateWithInspector) NewContext() *ContextWithInspector {
	context := v8.NewContext(iso.iso)
	iso.inspector.ContextCreated(context)
	return &ContextWithInspector{context, iso}
}

func (ctx *ContextWithInspector) Dispose() {
	ctx.iso.inspector.ContextDestroyed(ctx.Context)
	ctx.Context.Close()
}

func TestMonitorCreateDispose(t *testing.T) {
	t.Parallel()
	recorder := consoleAPIMessageRecorder{}
	iso := NewIsolateWithInspectorClient(&recorder)
	defer iso.Dispose()
	context := iso.NewContext()
	defer context.Dispose()

	_, err := context.RunScript("console.log('Hello, world!'); console.error('Error, world!');", "")
	if err != nil {
		t.Error("Error occurred: " + err.Error())
		return
	}
	if len(recorder.messages) != 2 {
		t.Error("Expected exactly one message")
	} else {
		msg1 := recorder.messages[0]
		msg2 := recorder.messages[1]
		if msg1.ErrorLevel != v8.ErrorLevelLog {
			t.Errorf("Expected Log error level. Got %d", msg1.ErrorLevel)
		}
		if msg2.ErrorLevel != v8.ErrorLevelError {
			t.Errorf("Expected Error error level. Got: %d", msg2.ErrorLevel)
		}
		if msg1.Message != "Hello, world!" {
			t.Errorf("Expected Hello, World, got %s", msg1.Message)
		}
		if msg2.Message != "Error, world!" {
			t.Errorf("Expected Error, world!, got %s", msg2.Message)
		}
	}
}
