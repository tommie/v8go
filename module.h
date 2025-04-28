#ifndef V8GO_MODULE_H
#define V8GO_MODULE_H

#ifdef __cplusplus

#include "deps/include/v8-persistent-handle.h"

namespace v8 {
class Module;
}

struct m_module {
  v8::Global<v8::Module> ptr;
};

extern "C" {
#else

typedef struct v8Module v8Module;
typedef struct m_module m_module;

#endif

#include "errors.h"

typedef struct m_ctx m_ctx;

extern RtnValue ModuleEvaluate(m_ctx* ctx, m_module* module);
extern RtnError ModuleInstantiateModule(m_ctx* ctx,
                                        m_module* module,
                                        void* resolveModuleHandle,
                                        void* resolveSourceModule);

#ifdef __cplusplus
}  // extern "C"
#endif

#endif
