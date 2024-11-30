#ifndef V8GO_VALUE_H
#define V8GO_VALUE_H
#ifdef __cplusplus

#include "deps/include/v8-persistent-handle.h"

namespace v8 {
class Isolate;
class Value;
} // namespace v8

typedef struct m_ctx m_ctx;

struct m_value {
  long id;
  v8::Isolate *iso;
  m_ctx *ctx;
  v8::Global<v8::Value> ptr;
};

extern "C" {
#else

typedef struct m_value m_value;

#endif

typedef m_value *ValuePtr;

#ifdef __cplusplus
}
#endif
#endif
