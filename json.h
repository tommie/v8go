#ifndef V8GO_JSON_H
#define V8GO_JSON_H

#include "errors.h"

#ifdef __cplusplus
extern "C" {
#else
#endif

typedef struct m_ctx m_ctx;
typedef m_ctx* ContextPtr;

extern RtnValue JSONParse(ContextPtr ctx_ptr, const char* str);
const char* JSONStringify(ContextPtr ctx_ptr, ValuePtr val_ptr);

#ifdef __cplusplus
}
#endif
#endif
