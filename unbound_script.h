#ifndef V8GO_UNBOUND_SCRIPT_H
#define V8GO_UNBOUND_SCRIPT_H
#include "errors.h"

#ifdef __cplusplus

#include "deps/include/v8-script.h"

namespace v8 {
class UnboundScript;
}

struct m_unboundScript {
  v8::Persistent<v8::UnboundScript> ptr;
};

extern "C" {
#else
#endif

typedef struct m_unboundScript m_unboundScript;
typedef m_unboundScript* UnboundScriptPtr;

typedef struct {
  UnboundScriptPtr ptr;
  int cachedDataRejected;
  RtnError error;
} RtnUnboundScript;

#ifdef __cplusplus
}  // extern "C"
#endif

#endif
