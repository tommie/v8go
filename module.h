#ifndef V8GO_MODULE_H
#define V8GO_MODULE_H

#include "errors.h"

#ifdef __cplusplus

#include "deps/include/v8-persistent-handle.h"

namespace v8 {
class Isolate;
class Module;
}  // namespace v8

struct m_module {
  v8::Global<v8::Module> ptr;
};

typedef v8::Isolate v8Isolate;

extern "C" {
#else

typedef struct v8Isolate v8Isolate;

#endif

typedef v8Isolate* IsolatePtr;

typedef struct m_ctx m_ctx;
typedef m_ctx* ContextPtr;

typedef struct m_module m_module;
typedef m_module* ModulePtr;

typedef struct {
  ModulePtr ptr;
  RtnError error;
} RtnModule;

// An import attribute, as given to the module resolver. The strings are
// only valid during the call.
typedef struct {
  const char* key;
  const char* value;
  int source_offset;
} ModuleImportAttribute;

extern RtnModule IsolateCompileModule(IsolatePtr iso,
                                      const char* source,
                                      const char* origin);
extern int ModuleGetStatus(IsolatePtr iso, ModulePtr module);
extern int ModuleScriptId(IsolatePtr iso, ModulePtr module);
extern RtnError ModuleInstantiate(ContextPtr ctx, ModulePtr module);
extern RtnValue ModuleEvaluate(ContextPtr ctx, ModulePtr module);
extern RtnValue ModuleGetNamespace(ContextPtr ctx, ModulePtr module);

#ifdef __cplusplus
}  // extern "C"
#endif

#endif
