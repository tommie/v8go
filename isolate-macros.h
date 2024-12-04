#ifndef V8GO_ISOLATE_MACROS_H
#define V8GO_ISOLATE_MACROS_H

#include "deps/include/v8-exception.h"
#include "deps/include/v8-isolate.h"
#include "deps/include/v8-locker.h"

#include "context.h"
#include "isolate.h"

static inline m_ctx* isolateInternalContext(v8::Isolate* iso) {
  return static_cast<m_ctx*>(iso->GetData(0));
}

#define ISOLATE_SCOPE(iso)           \
  Locker locker(iso);                \
  Isolate::Scope isolate_scope(iso); \
  HandleScope handle_scope(iso);

#define INTERNAL_CONTEXT(iso) m_ctx* ctx = static_cast<m_ctx*>(iso->GetData(0));

#endif
