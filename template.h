#ifndef V8GO_TEMPLATE_H
#define V8GO_TEMPLATE_H
#ifdef __cplusplus

#include "deps/include/v8-persistent-handle.h"

namespace v8 {
class Isolate;
class Template;
}  // namespace v8

struct m_template {
  v8::Isolate* iso;
  v8::Persistent<v8::Template> ptr;
};

extern "C" {
#else

typedef struct m_template m_template;

#endif

typedef m_template* TemplatePtr;

#ifdef __cplusplus
}  // extern "C"
#endif
#endif
