#include "module.h"
#include "_cgo_export.h"
#include "context-macros.h"
#include "context.h"

extern int ModuleScriptId(v8Isolate* iso, m_module* module) {
  v8::Local<v8::Module> local_mod = module->ptr.Get(iso);
  return local_mod->ScriptId();
}

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

v8::MaybeLocal<v8::Module> ResolveModuleCallback(
    v8::Local<v8::Context> context,
    v8::Local<v8::String> specifier,
    v8::Local<v8::FixedArray> import_attributes,
    v8::Local<v8::Module> referrer) {
  v8Isolate* iso = context->GetIsolate();
  int ctx_ref = context->GetEmbedderData(1).As<v8::Integer>()->Value();
  std::size_t cap = specifier->Utf8LengthV2(iso);
  char* buf = static_cast<char*>(malloc(cap));
  specifier->WriteUtf8V2(iso, buf, cap);
  m_module* mod = resolveModuleCallback(ctx_ref, buf);
  if (mod) {
    v8::MaybeLocal<v8::Module> ret = mod->ptr.Get(iso);
    return ret;
  }
  v8::MaybeLocal<v8::Module> res;
  context->GetIsolate()->ThrowError("Error importing module");
  return res;
}

extern RtnError ModuleInstantiateModule(m_ctx* ctx,
                                        m_module* module,
                                        void* resolveModuleHandle,
                                        void* resolveSourceModule) {
  LOCAL_CONTEXT(ctx);

  v8::Local<v8::Module> local_mod = module->ptr.Get(iso);

  v8::Maybe<bool> instantiateRes =
      local_mod->InstantiateModule(local_ctx, ResolveModuleCallback);

  RtnError rtn;
  rtn.location = 0;
  rtn.msg = 0;
  rtn.stack = 0;
  if (instantiateRes.IsNothing()) {
    rtn = ExceptionError(try_catch, iso, local_ctx);
  }
  return rtn;
}
