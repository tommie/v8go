package v8go

// #include "inspector.h"
import "C"
import (
	"runtime/cgo"
	"strconv"
	"unicode/utf16"
)

// Represents the level of console output from JavaScript. E.g., `console.log`,
// `console.error`, etc.
//
// The values reflect the values of v8::Isolate::MessageErrorLevel
//
// See also: https://v8.github.io/api/head/classv8_1_1Isolate.html
type MessageErrorLevel uint8

func (lvl MessageErrorLevel) String() string {
	switch lvl {
	case ErrorLevelLog:
		return "log"
	case ErrorLevelDebug:
		return "debug"
	case ErrorLevelError:
		return "error"
	case ErrorLevelInfo:
		return "info"
	case ErrorLevelWarning:
		return "warning"
	default:
		return strconv.Itoa(int(lvl))
	}
}

const (
	ErrorLevelLog MessageErrorLevel = 1 << iota
	ErrorLevelDebug
	ErrorLevelInfo
	ErrorLevelError
	ErrorLevelWarning
	ErrorLevelAll = ErrorLevelLog | ErrorLevelDebug | ErrorLevelInfo | ErrorLevelError | ErrorLevelWarning
)

// An Inspector in v8 provides access to internals of the engine, such as
// console output
//
// To receive console output, you need to first create an [InspectorClient]
// which will handle the interaction for a specific [Context].
//
// After a Context is created, you need to register it with the Inspector using
// [Inspector.ContextCreated], and cleanup using [Inspector.ContextDestroyed].
//
// See also: https://v8.github.io/api/head/classv8__inspector_1_1V8Inspector.html
type Inspector struct {
	ptr *C.v8Inspector
}

// An InspectorClient is the bridge from the [Inspector] to your code.
type InspectorClient struct {
	ptr          *C.v8InspectorClient
	clientHandle cgo.Handle
}

// ConsoleAPIMessage contains the information from v8 from console function
// calls.
//
// The fields correspond to the arguments for the C++ function
// v8_inspector::InspectorClient::consoleAPIMessage
//
// Note: Stack traces are not supported.
//
// See also: https://v8.github.io/api/head/classv8__inspector_1_1V8InspectorClient.html
type ConsoleAPIMessage struct {
	contextGroupId int
	ErrorLevel     MessageErrorLevel
	Message        string
	Url            string
	LineNumber     uint
	ColumnNumber   uint
}

// A ConsoleAPIMessageHandler will receive JavaScript `console` API calls.
type ConsoleAPIMessageHandler interface {
	ConsoleAPIMessage(message ConsoleAPIMessage)
}

// NewInspector creates an [Inspector] for a specific [Isolate] iso
// communicating with the [InspectorClient] client.
//
// Before disposing the iso, be sure to dispose the inspector using
// [Inspector.Dispose]
func NewInspector(iso *Isolate, client *InspectorClient) *Inspector {
	ptr := C.CreateInspector(iso.ptr, client.ptr)
	return &Inspector{
		ptr: ptr,
	}
}

// Dispose the [Inspector]. Call this before disposing the [Isolate] and the
// [InspectorClient] that this is connected to.
func (i *Inspector) Dispose() {
	C.DeleteInspector(i.ptr)
}

// ContextCreated tells the inspector that a new [Context] has been created.
// This must be called before the [InspectorClient] can be used.
func (i *Inspector) ContextCreated(ctx *Context) {
	C.InspectorContextCreated(i.ptr, ctx.ptr)
}

// ContextDestroyed must be called before a [Context] is closed.
func (i *Inspector) ContextDestroyed(ctx *Context) {
	C.InspectorContextDestroyed(i.ptr, ctx.ptr)
}

// Create a new [InspectorClient] passing a handler that will receive the
// callbacks from v8.
func NewInspectorClient(handler ConsoleAPIMessageHandler) *InspectorClient {
	clientHandle := cgo.NewHandle(handler)
	ptr := C.NewInspectorClient(C.uintptr_t(clientHandle))
	return &InspectorClient{
		clientHandle: clientHandle,
		ptr:          ptr,
	}
}

// Dispose frees up resources taken up by the [InspectorClient]. Be sure to call
// this after calling [Inspector.Dispose]
func (c *InspectorClient) Dispose() {
	c.clientHandle.Delete()
	C.DeleteInspectorClient(c.ptr)
}

func stringViewToString(d C.StringViewData) string {
	if d.is8bit {
		data := C.GoBytes(d.data, d.length)
		return string(data)
	} else {
		data := C.GoBytes(d.data, d.length*2)
		shorts := make([]uint16, len(data)/2)
		for i := 0; i < len(data); i += 2 {
			shorts[i/2] = (uint16(data[i+1]) << 8) | uint16(data[i])
		}
		return string(utf16.Decode(shorts))
	}
}

// goHandleConsoleAPIMessageCallback is called by C code when a console message
// is written. The correct [InspectorClient] is retrieved by the cgoHandle,
// which is a [cgo.Handle].
//
//export goHandleConsoleAPIMessageCallback
func goHandleConsoleAPIMessageCallback(
	cgoHandle C.uintptr_t,
	contextGroupId C.int,
	errorLevel C.int,
	message C.StringViewData,
	url C.StringViewData,
	lineNumber C.uint,
	columnNumber C.uint,
) {
	handle := cgo.Handle(cgoHandle)
	if client, ok := handle.Value().(ConsoleAPIMessageHandler); ok {
		client.ConsoleAPIMessage(ConsoleAPIMessage{
			contextGroupId: int(contextGroupId),
			ErrorLevel:     MessageErrorLevel(errorLevel),
			Message:        stringViewToString(message),
			Url:            stringViewToString(url),
			LineNumber:     uint(lineNumber),
			ColumnNumber:   uint(columnNumber),
		})
	}
}
