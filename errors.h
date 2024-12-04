#ifndef V8GO_ERRORS_H
#define V8GO_ERRORS_H

typedef struct {
  const char* msg;
  const char* location;
  const char* stack;
} RtnError;

#ifdef __cplusplus

#include "deps/include/v8-local-handle.h"

namespace v8 {
class Isolate;
class Context;
class TryCatch;
}  // namespace v8

extern "C" {

extern RtnError ExceptionError(v8::TryCatch& try_catch,
                               v8::Isolate* iso,
                               v8::Local<v8::Context> ctx);
#endif

typedef struct m_value m_value;
typedef m_value* ValuePtr;

typedef enum {
  ERROR_RANGE = 1,
  ERROR_REFERENCE,
  ERROR_SYNTAX,
  ERROR_TYPE,
  ERROR_WASM_COMPILE,
  ERROR_WASM_LINK,
  ERROR_WASM_RUNTIME,
  ERROR_GENERIC,
} ErrorTypeIndex;

typedef struct {
  ValuePtr value;
  RtnError error;
} RtnValue;

#ifdef __cplusplus
}
#endif
#endif
