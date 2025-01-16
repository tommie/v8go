package v8go_test

import (
	"reflect"
	"testing"

	v8 "github.com/tommie/v8go"
)

type consoleAPIMessage struct {
	Message    string
	ErrorLevel v8.MessageErrorLevel
}
type consoleAPIMessageRecorder struct {
	messages []consoleAPIMessage
}

func (r *consoleAPIMessageRecorder) ConsoleAPIMessage(msg v8.ConsoleAPIMessage) {
	r.messages = append(r.messages, consoleAPIMessage{
		Message:    msg.Message,
		ErrorLevel: msg.ErrorLevel,
	})
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

func TestMonitorConsoleLogLevelt(t *testing.T) {
	t.Parallel()
	recorder := consoleAPIMessageRecorder{}
	iso := NewIsolateWithInspectorClient(&recorder)
	defer iso.Dispose()
	context := iso.NewContext()
	defer context.Dispose()

	_, err := context.RunScript(`
		console.log("Log msg");
		console.info("Info msg");
		console.debug("Debug msg");
		console.warn("Warn msg");
		console.error("Error msg");
	`, "")
	if err != nil {
		t.Fatal("Error occurred", err)
	}
	actual := recorder.messages
	expected := []consoleAPIMessage{
		{Message: "Log msg", ErrorLevel: v8.ErrorLevelLog},
		{Message: "Info msg", ErrorLevel: v8.ErrorLevelInfo},
		{Message: "Debug msg", ErrorLevel: v8.ErrorLevelDebug},
		{Message: "Warn msg", ErrorLevel: v8.ErrorLevelWarning},
		{Message: "Error msg", ErrorLevel: v8.ErrorLevelError},
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("Unexpected messages. \nExpected: %v\nGot: %v", expected, actual)
	}
}

// Verify utf-16 conversion. Internally, the strings are represented by a
// StringView, which is undocumented. Experiements shows that the values
// returned are an utf-16le encoded array, and a length.
//
// The length is assumed to be the size of the array, not the number of
// characters. This test verifies that, by writing a character that needs
// several utf-16 elements for endocing.
//
// https://v8.github.io/api/head/classv8__inspector_1_1StringView.html
func TestMonitorConsoleLogWideCharacters(t *testing.T) {
	t.Parallel()
	recorder := consoleAPIMessageRecorder{}
	iso := NewIsolateWithInspectorClient(&recorder)
	defer iso.Dispose()
	context := iso.NewContext()
	defer context.Dispose()

	_, err := context.RunScript(`
		console.log("This character takes up multiple utf-16 values: 𐀀");
	`, "")
	if err != nil {
		t.Fatal("Error occurred", err)
	}
	actual := recorder.messages
	expected := []consoleAPIMessage{
		{
			Message:    "This character takes up multiple utf-16 values: 𐀀",
			ErrorLevel: v8.ErrorLevelLog,
		},
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("Unexpected messages. \nExpected: %v\nGot: %v", expected, actual)
	}
}
