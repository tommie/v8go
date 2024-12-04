#ifndef V8GO_CONTEXT_MACROS_H
#define V8GO_CONTEXT_MACROS_H

#include "deps/include/v8-context.h"
#include "deps/include/v8-isolate.h"
#include "deps/include/v8-local-handle.h"
#include "deps/include/v8-locker.h"

#include "context.h"

#define LOCAL_CONTEXT(ctx)                              \
  v8::Isolate* iso = ctx->iso;                          \
  v8::Locker locker(iso);                               \
  v8::Isolate::Scope isolate_scope(iso);                \
  v8::HandleScope handle_scope(iso);                    \
  v8::TryCatch try_catch(iso);                          \
  v8::Local<v8::Context> local_ctx = ctx->ptr.Get(iso); \
  v8::Context::Scope context_scope(local_ctx);

#endif
