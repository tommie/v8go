#include "module.h"
#include <stdio.h>
#include "_cgo_export.h"
#include "context-macros.h"
#include "context.h"
#include "data.h"
#include "isolate-macros.h"
#include "value.h"

using namespace v8;

extern int ModuleGetStatus(m_module* module) {
  Isolate* iso = module->iso;
  ISOLATE_SCOPE(iso);

  Local<Module> local_mod = module->ptr.Get(iso);
  return local_mod->GetStatus();
}

extern int ModuleScriptId(m_module* module) {
  ISOLATE_SCOPE(module->iso);
  Local<Module> local_mod = module->ptr.Get(module->iso);
  return local_mod->ScriptId();
}

extern bool ModuleIsSourceTextModule(m_module* module) {
  ISOLATE_SCOPE(module->iso);
  Local<Module> local_mod = module->ptr.Get(module->iso);
  return local_mod->IsSourceTextModule();
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
  v8Isolate* iso = v8::Isolate::GetCurrent();
  int ctx_ref = context->GetEmbedderDataV2(ContextDataIndex::REF)
                    .As<v8::Integer>()
                    ->Value();
  std::size_t cap = specifier->Utf8Length(iso) + 1;
  char* buf = static_cast<char*>(malloc(cap));
  specifier->WriteUtf8(iso, buf, cap, v8::String::WriteFlags::kNullTerminate);
  m_module ref(iso, referrer);
  v8goFixedArray attributes(iso, import_attributes);
  resolveModuleCallback_return retval =
      resolveModuleCallback(ctx_ref, buf, &attributes, &ref);
  if (retval.r1 != nullptr) {
    iso->ThrowException(retval.r1->ptr.Get(iso));
    return MaybeLocal<Module>();
  }
  return retval.r0->ptr.Get(iso);
}

extern RtnError ModuleInstantiateModule(m_ctx* ctx, m_module* module) {
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

extern void ModuleDelete(m_module* module) {
  delete module;
}

extern ValuePtr ModuleGetModuleNamespace(v8Isolate* iso, m_module* module) {
  ISOLATE_SCOPE(iso);
  INTERNAL_CONTEXT(iso);
  v8::Context::Scope context_scope(ctx->ptr.Get(iso));
  Local<Module> mod = module->ptr.Get(iso);
  Local<Value> result = mod->GetModuleNamespace();

  m_value* new_val = new m_value;
  new_val->id = 0;
  new_val->iso = iso;
  new_val->ctx = ctx;
  new_val->ptr = Global<Value>(iso, result);

  return tracked_value(ctx, new_val);
}

extern m_module* ScriptCompilerCompileModule(Isolate* iso,
                                             const char* s,
                                             const char* o) {
  v8::Locker locker(iso);
  v8::Isolate::Scope iso_scope(iso);
  v8::TryCatch try_catch(iso);
  v8::HandleScope handle_scope(iso);
  INTERNAL_CONTEXT(iso);
  v8::Context::Scope context_scope(ctx->ptr.Get(iso));

  Local<String> src = String::NewFromUtf8(iso, s).ToLocalChecked();
  Local<String> ogn = String::NewFromUtf8(iso, o).ToLocalChecked();

  ScriptOrigin origin(ogn,    // resource_name
                      0, 0,   // resource_line_offset, resource_column_offset
                      false,  //  resource_is_shared_cross_origin
                      -1,     // script_id
                      Local<Value>(),  // source_map_url
                      false,           // resource_is_opaque
                      false,           // is_wasm
                      true             // is_module
                            // Local<Data> host_defined_options = Local<Data>())
  );

  ScriptCompiler::Source source(src, origin);

  Local<Module> module =
      ScriptCompiler::CompileModule(iso, &source).ToLocalChecked();
  if (try_catch.HasCaught()) {
    return 0;
  }

  m_module* retVal = new m_module(iso, module);
  return retVal;
}
