package v8go

// #include "inspector.h"
import "C"
import (
	"runtime/cgo"
	"unicode/utf16"
)

type MessageErrorLevel uint8

const (
	ErrorLevelLog MessageErrorLevel = 1 << iota
	ErrorLevelDebug
	ErrorLevelInfo
	ErrorLevelError
	ErrorLevelWarning
	ErrorLevelAll = ErrorLevelLog | ErrorLevelDebug | ErrorLevelInfo | ErrorLevelError | ErrorLevelWarning
)

type Inspector struct {
	ptr C.InspectorPtr
}

type InspectorClient struct {
	ptr          C.InspectorClientPtr
	clientHandle cgo.Handle
}

type ConsoleAPIMessage struct {
	contextGroupId int
	ErrorLevel     MessageErrorLevel
	Message        string
	Url            string
	LineNumber     uint
	ColumnNumber   uint
	// stackTrace StackTrace
}

type ConsoleAPIMessageHandler interface {
	ConsoleAPIMessage(message ConsoleAPIMessage)
}

type ConsoleAPIMessageHandlerFunc func(msg ConsoleAPIMessage)

func (f ConsoleAPIMessageHandlerFunc) ConsoleAPIMessage(msg ConsoleAPIMessage) {
	f(msg)
}

func NewInspector(iso *Isolate, client *InspectorClient) *Inspector {
	ptr := C.CreateInspector(iso.ptr, client.ptr)
	return &Inspector{
		ptr: ptr,
	}
}

func (i *Inspector) Dispose() {
	C.DeleteInspector(i.ptr)
}

func (i *Inspector) ContextCreated(ctx *Context) {
	C.InspectorContextCreated(i.ptr, ctx.ptr)
}

func (i *Inspector) ContextDestroyed(ctx *Context) {
	C.InspectorContextDestroyed(i.ptr, ctx.ptr)
}

func NewInspectorClient(handler ConsoleAPIMessageHandler) *InspectorClient {
	clientHandle := cgo.NewHandle(handler)
	ptr := C.NewInspectorClient(C.uintptr_t(clientHandle))
	return &InspectorClient{
		clientHandle: clientHandle,
		ptr:          ptr,
	}
}

func (c *InspectorClient) Dispose() {
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

//
//export goHandleConsoleAPIMessageCallback
func goHandleConsoleAPIMessageCallback(
	callbackRef C.uintptr_t,
	contextGroupId C.int,
	errorLevel C.int,
	message C.StringViewData,
	url C.StringViewData,
	lineNumber C.uint,
	columnNumber C.uint,
) {
	// Convert data to Go data
	handle := cgo.Handle(callbackRef)
	if client, ok := handle.Value().(ConsoleAPIMessageHandler); ok {
		// TODO, Stack trace
		client.ConsoleAPIMessage(ConsoleAPIMessage{
			ErrorLevel:   MessageErrorLevel(errorLevel),
			Message:      stringViewToString(message),
			Url:          stringViewToString(url),
			LineNumber:   uint(lineNumber),
			ColumnNumber: uint(columnNumber),
		})
	}
}
