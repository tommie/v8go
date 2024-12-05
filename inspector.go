package v8go

// #include "inspector.h"
import "C"
import (
	"sync"
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
	ptr     C.InspectorClientPtr
	handler ConsoleAPIMessageHandler
}

// registry is a simple map of int->something. Allows passing an int to C-code
// that can be used in a callback; then Go code can retrieve the right object
// afterwards.
// This can be generalised, e.g. Isolate has similar functionality handling
// callback functions
type registry[T any] struct {
	entries map[C.int]T
	id      C.int
	lock    sync.RWMutex
}

func newRegistry[T any]() *registry[T] {
	return &registry[T]{
		entries: make(map[C.int]T),
	}
}

var clientRegistry = newRegistry[*InspectorClient]()

func (r *registry[T]) register(instance T) C.int {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.id++
	r.entries[r.id] = instance
	return r.id
}

func (r *registry[T]) unregister(id C.int) {
	r.lock.Lock()
	defer r.lock.Unlock()
	delete(r.entries, id)
}

func (r *registry[T]) get(id C.int) T {
	r.lock.RLock()
	defer r.lock.RUnlock()
	// What if not found? Return two parameters? Panic? (my preference)
	return r.entries[id]
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
	result := &InspectorClient{
		handler: handler,
	}
	ref := clientRegistry.register(result)
	result.ptr = C.NewInspectorClient(ref)
	return result
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
	callbackRef C.int,
	contextGroupId C.int,
	errorLevel C.int,
	message C.StringViewData,
	url C.StringViewData,
	lineNumber C.uint,
	columnNumber C.uint,
) {
	// Convert data to Go data
	client := clientRegistry.get(callbackRef)
	// TODO, Stack trace
	client.handler.ConsoleAPIMessage(ConsoleAPIMessage{
		ErrorLevel:   MessageErrorLevel(errorLevel),
		Message:      stringViewToString(message),
		Url:          stringViewToString(url),
		LineNumber:   uint(lineNumber),
		ColumnNumber: uint(columnNumber),
	})
	// client.handleConsoleAPIMessageCallback(data)
}
