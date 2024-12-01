#ifndef V8GO_CONTEXT_H
#define V8GO_CONTEXT_H

#include "interop.h"
#include "isolate.h"

#ifdef __cplusplus

#include <unordered_map>
#include <vector>
#include "deps/include/v8-persistent-handle.h"
#include "value.h"

namespace v8 {
class Isolate;
class Context;
}  // namespace v8

typedef struct m_unboundScript m_unboundScript;

struct m_ctx {
  v8::Isolate* iso;
  std::unordered_map<long, m_value*> vals;
  std::vector<m_unboundScript*> unboundScripts;
  v8::Persistent<v8::Context> ptr;
  long nextValId;
};
typedef m_ctx* ContextPtr;

extern m_value* tracked_value(m_ctx* ctx, m_value* val);
typedef v8::Isolate* IsolatePtr;

extern "C" {
#else

typedef struct m_ctx m_ctx;
typedef m_ctx* ContextPtr;

typedef struct m_value m_value;
typedef m_value* ValuePtr;

#endif

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
