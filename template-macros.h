#ifndef V8GO_TEMPLATE_MACROS_H
#define V8GO_TEMPLATE_MACROS_H

#include "deps/include/v8-isolate.h"
#include "deps/include/v8-local-handle.h"
#include "deps/include/v8-locker.h"
#include "deps/include/v8-template.h"

#define LOCAL_TEMPLATE(tmpl_ptr)     \
  Isolate* iso = tmpl_ptr->iso;      \
  Locker locker(iso);                \
  Isolate::Scope isolate_scope(iso); \
  HandleScope handle_scope(iso);     \
  Local<Template> tmpl = tmpl_ptr->ptr.Get(iso);

#endif
