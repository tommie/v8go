#include "module.h"
#include "context-macros.h"

extern RtnValue ModuleEvaluate(ContextPtr ctx, m_module* module) {
  LOCAL_CONTEXT(ctx);
  v8::Local<v8::Module> local_mod = module->ptr.Get(iso);

  RtnValue rtn = {};

  v8::Local<v8::Value> result;
  if (!local_mod->Evaluate(local_ctx).ToLocal(&result)) {
    rtn.error = ExceptionError(try_catch, iso, local_ctx);
    return rtn;
  }

  m_value* new_val = new m_value;
  new_val->id = 0;
  new_val->iso = iso;
  new_val->ctx = ctx;
  new_val->ptr = v8::Global<v8::Value>(iso, result);

  rtn.value = tracked_value(ctx, new_val);
  return rtn;
}
