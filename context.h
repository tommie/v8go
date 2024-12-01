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

#endif
extern void ContextFree(ContextPtr ctx);

#ifdef __cplusplus
}  // extern "C"

#define LOCAL_CONTEXT(ctx)                      \
  Isolate* iso = ctx->iso;                      \
  Locker locker(iso);                           \
  Isolate::Scope isolate_scope(iso);            \
  HandleScope handle_scope(iso);                \
  TryCatch try_catch(iso);                      \
  Local<Context> local_ctx = ctx->ptr.Get(iso); \
  Context::Scope context_scope(local_ctx);

#endif

#endif
