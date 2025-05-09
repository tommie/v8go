#ifndef V8GO_SCRIPT_COMPILER_H
#define V8GO_SCRIPT_COMPILER_H

#include "unbound_script.h"

#ifdef __cplusplus

#include "deps/include/v8-script.h"

namespace v8 {
class Isolate;
}
typedef v8::Isolate v8Isolate;
class m_module;

extern "C" {

#else

typedef struct v8Isolate v8Isolate;
typedef struct m_module m_module;

#endif

typedef v8Isolate* IsolatePtr;

typedef struct {
  ScriptCompilerCachedData cachedData;
  int compileOption;
} CompileOptions;

// ScriptCompiler::CompileOptions values
extern const int ScriptCompilerNoCompileOptions;
extern const int ScriptCompilerConsumeCodeCache;
extern const int ScriptCompilerEagerCompile;

extern RtnUnboundScript IsolateCompileUnboundScript(IsolatePtr iso_ptr,
                                                    const char* source,
                                                    const char* origin,
                                                    CompileOptions options);

extern m_module* ScriptCompilerCompileModule(v8Isolate* iso,
                                             const char* source,
                                             const char* origin);

#ifdef __cplusplus
}
#endif
#endif
