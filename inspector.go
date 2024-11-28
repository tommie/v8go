package v8go

// #include "inspector.h"
import "C"

type Inspector struct {
	ptr C.InspectorPtr
}

type InspectorClient struct {
	ptr C.InspectorClientPtr
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

func NewInspectorClient() *InspectorClient {
	ptr := C.NewInspectorClient()
	return &InspectorClient{ptr: ptr}
}

func (c *InspectorClient) Dispose() {
	C.DeleteInspectorClient(c.ptr)
}
