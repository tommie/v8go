#ifndef V8GO_UNBOUND_SCRIPT_H
#define V8GO_UNBOUND_SCRIPT_H

#include <stdint.h>

#include "errors.h"

#ifdef __cplusplus

// If we could forward declare v8::ScriptCompiler::CachedData, we wouldn't need
// this include. But appears not to be possible.
// https://stackoverflow.com/a/1021809/158483
#include "deps/include/v8-script.h"

namespace v8 {
class UnboundScript;
class Isolate;
}  // namespace v8

struct m_unboundScript {
  v8::Persistent<v8::UnboundScript> ptr;
};

typedef v8::ScriptCompiler::CachedData* ScriptCompilerCachedDataPtr;
typedef v8::Isolate v8Isolate;

extern "C" {
#else

typedef struct v8ScriptCompilerCachedData v8ScriptCompilerCachedData;
typedef const v8ScriptCompilerCachedData* ScriptCompilerCachedDataPtr;
typedef struct v8Isolate v8Isolate;

#endif

typedef v8Isolate* IsolatePtr;

typedef struct m_ctx m_ctx;
typedef m_ctx* ContextPtr;

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

extern ScriptCompilerCachedData* UnboundScriptCreateCodeCache(
    IsolatePtr iso_ptr,
    UnboundScriptPtr us_ptr);
extern void ScriptCompilerCachedDataDelete(
    ScriptCompilerCachedData* cached_data);
extern RtnValue UnboundScriptRun(ContextPtr ctx_ptr, UnboundScriptPtr us_ptr);

#ifdef __cplusplus
}  // extern "C"
#endif
#endif
