#ifndef V8GO_INSPECTOR_H
#define V8GO_INSPECTOR_H

#ifdef __cplusplus

namespace v8 {
class Isolate;
};

namespace v8_inspector {
class V8Inspector;
class V8InspectorClient;
};  // namespace v8_inspector

typedef v8::Isolate v8Isolate;
typedef v8_inspector::V8Inspector* InspectorPtr;
typedef v8_inspector::V8InspectorClient* InspectorClientPtr;

extern "C" {
#else
typedef struct v8Inspector v8Inspector;
typedef v8Inspector* InspectorPtr;

typedef struct v8InspectorClient v8InspectorClient;
typedef v8InspectorClient* InspectorClientPtr;

typedef struct v8Isolate v8Isolate;

typedef _Bool bool;

#endif

#include <stddef.h>
#include <stdint.h>

typedef struct m_ctx m_ctx;
typedef m_ctx* ContextPtr;

extern InspectorPtr CreateInspector(v8Isolate* iso, InspectorClientPtr client);
extern void DeleteInspector(InspectorPtr inspector);
extern void InspectorContextCreated(InspectorPtr inspector, ContextPtr context);
extern void InspectorContextDestroyed(InspectorPtr inspector,
                                      ContextPtr context);

extern InspectorClientPtr NewInspectorClient(int callbackRef);
extern void DeleteInspectorClient(InspectorClientPtr client);

typedef struct StringViewData {
  bool is8bit;
  void const* data;
  int length;
} StringViewData;

#ifdef __cplusplus
}  // extern "C"
#endif

#endif
