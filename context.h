#ifndef V8GO_CONTEXT_H
#define V8GO_CONTEXT_H

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

extern "C" {
#else

typedef struct m_ctx m_ctx;
typedef m_ctx* ContextPtr;
typedef struct m_value m_value;
typedef m_value* ValuePtr;

#endif

extern void ContextFree(ContextPtr ctx);
extern ValuePtr ContextGlobal(ContextPtr ctx_ptr);

#ifdef __cplusplus
}  // extern "C"

#endif

#endif
