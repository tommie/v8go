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
typedef v8_inspector::V8Inspector v8Inspector;
typedef v8_inspector::V8InspectorClient v8InspectorClient;

extern "C" {
#else
typedef struct v8Inspector v8Inspector;
typedef struct v8InspectorClient v8InspectorClient;

typedef struct v8Isolate v8Isolate;

typedef _Bool bool;

#endif

#include <stddef.h>
#include <stdint.h>

typedef struct m_ctx m_ctx;

extern v8Inspector* CreateInspector(v8Isolate* iso, v8InspectorClient* client);
extern void DeleteInspector(v8Inspector* inspector);
extern void InspectorContextCreated(v8Inspector* inspector, m_ctx* context);
extern void InspectorContextDestroyed(v8Inspector* inspector, m_ctx* context);

extern v8InspectorClient* NewInspectorClient(uintptr_t callbackRef);
extern void DeleteInspectorClient(v8InspectorClient* client);

typedef struct StringViewData {
  bool is8bit;
  void const* data;
  int length;
} StringViewData;

#ifdef __cplusplus
}  // extern "C"
#endif

#endif
