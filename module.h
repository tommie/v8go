#ifndef V8GO_MODULE_H
#define V8GO_MODULE_H

#ifdef __cplusplus

#include "deps/include/v8-persistent-handle.h"
#include "isolate-macros.h"

namespace v8 {
class Module;
class Isolate;
}  // namespace v8

class m_module {
 public:
  v8::Isolate* iso;
  v8::Global<v8::Module> ptr;
  m_module(v8::Isolate* iso, v8::Local<v8::Module> mod) {
    ISOLATE_SCOPE(iso);
    this->iso = iso;
    this->ptr.Reset(iso, mod);
  }
};

typedef v8::Isolate v8Isolate;

extern "C" {
#else

typedef struct v8Module v8Module;
typedef struct m_module m_module;
typedef struct v8Isolate v8Isolate;

#endif

#include <stdbool.h>
#include "errors.h"

typedef struct m_value m_value;
typedef struct m_ctx m_ctx;

extern int ModuleScriptId(m_module* module);
extern bool ModuleIsSourceTextModule(m_module* module);
extern RtnValue ModuleEvaluate(m_ctx* ctx, m_module* module);
extern RtnError ModuleInstantiateModule(m_ctx* ctx,
                                        m_module* module,
                                        void* resolveModuleHandle,
                                        void* resolveSourceModule);
extern void ModuleDelete(m_module* module);

#ifdef __cplusplus
}  // extern "C"
#endif

#endif
