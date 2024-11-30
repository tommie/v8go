#ifndef V8GO_ISOLATE_H
#define V8GO_ISOLATE_H
#ifdef __cplusplus

// #include "libplatform/libplatform.h"
#include <memory>

#include "deps/include/v8-isolate.h"
#include "v8-script.h"

typedef v8::Isolate* IsolatePtr;
typedef v8::ScriptCompiler::CachedData* ScriptCompilerCachedDataPtr;

extern "C" {
#else
typedef struct v8Isolate v8Isolate;
typedef v8Isolate* IsolatePtr;

typedef struct v8ScriptCompilerCachedData v8ScriptCompilerCachedData;
typedef const v8ScriptCompilerCachedData* ScriptCompilerCachedDataPtr;

#endif

#include <stddef.h>
#include <stdint.h>

#include "unbound_script.h"
#include "value.h"

typedef struct {
  ScriptCompilerCachedDataPtr ptr;
  const uint8_t* data;
  int length;
  int rejected;
} ScriptCompilerCachedData;

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

#endif

#endif
