#ifndef V8GO_UNBOUND_SCRIPT_H
#define V8GO_UNBOUND_SCRIPT_H
#include "errors.h"

#ifdef __cplusplus

#include <memory>
#include "deps/include/v8-script.h"

namespace v8 {
class UnboundScript;
}

struct m_unboundScript {
  v8::Persistent<v8::UnboundScript> ptr;
};

typedef v8::ScriptCompiler::CachedData* ScriptCompilerCachedDataPtr;

extern "C" {
#else

typedef struct v8ScriptCompilerCachedData v8ScriptCompilerCachedData;
typedef const v8ScriptCompilerCachedData* ScriptCompilerCachedDataPtr;

#endif

typedef struct m_unboundScript m_unboundScript;
typedef m_unboundScript* UnboundScriptPtr;

typedef struct {
  UnboundScriptPtr ptr;
  int cachedDataRejected;
  RtnError error;
} RtnUnboundScript;

typedef struct {
  ScriptCompilerCachedDataPtr ptr;
  const uint8_t* data;
  int length;
  int rejected;
} ScriptCompilerCachedData;

#ifdef __cplusplus
}  // extern "C"
#endif
#endif
