#include "context.h"
#include "unbound_script.h"
#include "value.h"

void ContextFree(ContextPtr ctx) {
  if (ctx == nullptr) {
    return;
  }
  ctx->ptr.Reset();

  for (auto it = ctx->vals.begin(); it != ctx->vals.end(); ++it) {
    auto value = it->second;
    value->ptr.Reset();
    delete value;
  }
  ctx->vals.clear();

  for (m_unboundScript* us : ctx->unboundScripts) {
    us->ptr.Reset();
    delete us;
  }

  delete ctx;
}

m_value* tracked_value(m_ctx* ctx, m_value* val) {
  // (rogchap) we track values against a context so that when the context is
  // closed (either manually or GC'd by Go) we can also release all the
  // values associated with the context;
  if (val->id == 0) {
    val->id = ++ctx->nextValId;
    ctx->vals[val->id] = val;
  }

  return val;
}
