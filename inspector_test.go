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
