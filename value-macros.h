#ifndef V8GO_VALUE_MACROS_H
#define V8GO_VALUE_MACROS_H

#define LOCAL_VALUE(val)                   \
  Isolate* iso = val->iso;                 \
  Locker locker(iso);                      \
  Isolate::Scope isolate_scope(iso);       \
  HandleScope handle_scope(iso);           \
  TryCatch try_catch(iso);                 \
  m_ctx* ctx = val->ctx;                   \
  Local<Context> local_ctx;                \
  if (ctx != nullptr) {                    \
    local_ctx = ctx->ptr.Get(iso);         \
  } else {                                 \
    ctx = isolateInternalContext(iso);     \
    local_ctx = ctx->ptr.Get(iso);         \
  }                                        \
  Context::Scope context_scope(local_ctx); \
  Local<Value> value = val->ptr.Get(iso);

#endif
