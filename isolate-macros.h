#ifndef V8GO_ISOLATE_MACROS_H
#define V8GO_ISOLATE_MACROS_H

#define ISOLATE_SCOPE(iso)           \
  Locker locker(iso);                \
  Isolate::Scope isolate_scope(iso); \
  HandleScope handle_scope(iso);

#define INTERNAL_CONTEXT(iso) \
  m_cm_ctx* ctx = static_cast<m_ctx*>(iso->GetData(0));

#endif
