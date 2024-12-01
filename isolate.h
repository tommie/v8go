#ifndef V8GO_ISOLATE_H
#define V8GO_ISOLATE_H

#include "unbound_script.h"

#ifdef __cplusplus

// #include "libplatform/libplatform.h"
#include <memory>

namespace v8 {
class Isolate;
}

typedef struct m_ctx m_ctx;
typedef v8::Isolate* IsolatePtr;

extern "C" {
#else

typedef struct v8Isolate v8Isolate;
typedef v8Isolate* IsolatePtr;

#endif

#include <stddef.h>
#include <stdint.h>

#include "value.h"

// ScriptCompiler::CompileOptions values
extern const int ScriptCompilerNoCompileOptions;
extern const int ScriptCompilerConsumeCodeCache;
extern const int ScriptCompilerEagerCompile;

typedef struct {
  ScriptCompilerCachedData cachedData;
  int compileOption;
} CompileOptions;

typedef struct {
  size_t total_heap_size;
  size_t total_heap_size_executable;
  size_t total_physical_size;
  size_t total_available_size;
  size_t used_heap_size;
  size_t heap_size_limit;
  size_t malloced_memory;
  size_t external_memory;
  size_t peak_malloced_memory;
  size_t number_of_native_contexts;
  size_t number_of_detached_contexts;
} IsolateHStatistics;

extern IsolatePtr NewIsolate();
extern void IsolatePerformMicrotaskCheckpoint(IsolatePtr ptr);
extern void IsolateDispose(IsolatePtr ptr);
extern void IsolateTerminateExecution(IsolatePtr ptr);
extern int IsolateIsExecutionTerminating(IsolatePtr ptr);
extern IsolateHStatistics IsolationGetHeapStatistics(IsolatePtr ptr);

extern ValuePtr IsolateThrowException(IsolatePtr iso, ValuePtr value);

extern RtnUnboundScript IsolateCompileUnboundScript(IsolatePtr iso_ptr,
                                                    const char* source,
                                                    const char* origin,
                                                    CompileOptions options);

#ifdef __cplusplus
}  // extern "C"

#define ISOLATE_SCOPE(iso)           \
  Locker locker(iso);                \
  Isolate::Scope isolate_scope(iso); \
  HandleScope handle_scope(iso);

#define INTERNAL_CONTEXT(iso) \
  m_cm_ctx* ctx = static_cast<m_ctx*>(iso->GetData(0));

#endif

#endif
