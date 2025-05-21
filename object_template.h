#ifndef V8GO_OBJECT_TEMPLATE_H
#define V8GO_OBJECT_TEMPLATE_H

#include "errors.h"
#include "template.h"

#ifdef __cplusplus

namespace v8 {
class Isolate;
}

typedef v8::Isolate v8Isolate;

extern "C" {
#else

typedef struct v8Isolate v8Isolate;

#endif

typedef struct m_ctx m_ctx;
typedef struct m_template m_template;

extern TemplatePtr NewObjectTemplate(v8Isolate* iso_ptr);
extern RtnValue ObjectTemplateNewInstance(m_template* ptr, m_ctx* ctx_ptr);
extern void ObjectTemplateSetInternalFieldCount(m_template* ptr,
                                                int field_count);
extern int ObjectTemplateInternalFieldCount(m_template* ptr);
extern void ObjectTemplateSetAccessorProperty(m_template* ptr,
                                              const char* key,
                                              m_template* get,
                                              m_template* set,
                                              int attributes);
extern void ObjectTemplateMarkAsUndetectable(m_template* ptr);
extern void ObjectTemplateSetCallAsFunctionHandler(m_template* ptr,
                                                   int callback_ref);

#ifdef __cplusplus
}
#endif
#endif
