#ifndef V8GO_CONTEXT_H
#define V8GO_CONTEXT_H

#include "errors.h"
#include "isolate.h"

#ifdef __cplusplus

#include "deps/include/v8-persistent-handle.h"

#include <unordered_map>
#include <vector>
#include "value.h"

namespace v8 {
class Value;
class Isolate;
class Context;
}  // namespace v8

typedef v8::Isolate v8Isolate;
typedef struct m_unboundScript m_unboundScript;

// ContextDataIndex defines the indexes for "embedder data".
enum ContextDataIndex {
  // We start at 1, as slot 0 has special meaning for the Chrome debugger

  // Is an integer "handle" created in Go code, so given a specific V8 context,
  // Go code can find its corresponding *Context value.
  REF = 1,

  // WRAPPER_PTR is the pointer being transferred between C++ and Go, a pointer
  // to a C++ object responsible for maintenance tasks.
  WRAPPER_PTR = 2,
};

// A wrapper on top of V8's Context object providing additional maintenance
// tasks for V8 go, e.g., releaseing unreleased Value objects associated with
// the context.
//
// Exactly one wrapper instance must exist for a V8 context. A pointer to the
// corresponding wrapper is stored as embedded data in the V8 context.
struct m_ctx {
  v8::Isolate* iso;
  std::unordered_map<long, m_value*> vals;
  std::vector<m_unboundScript*> unboundScripts;
  v8::Persistent<v8::Context> ptr;
  long nextValId;

  // Retrieves the wrapper context for a specific V8 context.
  static m_ctx* FromV8Context(v8::Local<v8::Context> ctx);
};
typedef m_ctx* ContextPtr;

extern m_value* tracked_value(m_ctx* ctx, m_value* val);
extern m_value* track_value(m_ctx* ctx, v8::Local<v8::Value> val);

extern "C" {
#else
typedef struct v8Isolate v8Isolate;
#endif
typedef v8Isolate* IsolatePtr;

typedef struct m_value m_value;
typedef m_value* ValuePtr;

typedef struct m_template m_template;
typedef m_template* TemplatePtr;

extern ContextPtr NewContext(IsolatePtr iso_ptr,
                             TemplatePtr global_template_ptr,
                             int ref);
extern int ContextRetainedValueCount(ContextPtr ctx);
extern ValuePtr ContextGlobal(ContextPtr ctx_ptr);
extern void ContextFree(ContextPtr ctx);
extern RtnValue RunScript(ContextPtr ctx_ptr,
                          const char* source,
                          const char* origin);

#ifdef __cplusplus
}  // extern "C"

#endif

#endif
