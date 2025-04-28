#include "script_compiler.h"
#include "context-macros.h"
#include "deps/include/v8-context.h"
#include "deps/include/v8-message.h"
#include "isolate-macros.h"
#include "module.h"

using namespace v8;

const int ScriptCompilerNoCompileOptions =
    v8::ScriptCompiler::kNoCompileOptions;
const int ScriptCompilerConsumeCodeCache =
    v8::ScriptCompiler::kConsumeCodeCache;
const int ScriptCompilerEagerCompile = v8::ScriptCompiler::kEagerCompile;

m_unboundScript* tracked_unbound_script(m_ctx* ctx, m_unboundScript* us) {
  ctx->unboundScripts.push_back(us);

  return us;
}

RtnUnboundScript IsolateCompileUnboundScript(IsolatePtr iso,
                                             const char* s,
                                             const char* o,
                                             CompileOptions opts) {
  ISOLATE_SCOPE(iso);
  m_ctx* ctx = isolateInternalContext(iso);
  TryCatch try_catch(iso);
  Local<Context> local_ctx = ctx->ptr.Get(iso);
  Context::Scope context_scope(local_ctx);

  RtnUnboundScript rtn = {};

  Local<String> src =
      String::NewFromUtf8(iso, s, NewStringType::kNormal).ToLocalChecked();
  Local<String> ogn =
      String::NewFromUtf8(iso, o, NewStringType::kNormal).ToLocalChecked();

  ScriptCompiler::CompileOptions option =
      static_cast<ScriptCompiler::CompileOptions>(opts.compileOption);

  ScriptCompiler::CachedData* cached_data = nullptr;

  if (opts.cachedData.data) {
    cached_data = new ScriptCompiler::CachedData(opts.cachedData.data,
                                                 opts.cachedData.length);
  }

  ScriptOrigin script_origin(ogn);

  ScriptCompiler::Source source(src, script_origin, cached_data);

  Local<UnboundScript> unbound_script;
  if (!ScriptCompiler::CompileUnboundScript(iso, &source, option)
           .ToLocal(&unbound_script)) {
    rtn.error = ExceptionError(try_catch, iso, local_ctx);
    return rtn;
  };

  if (cached_data) {
    rtn.cachedDataRejected = cached_data->rejected;
  }

  m_unboundScript* us = new m_unboundScript;
  us->ptr.Reset(iso, unbound_script);
  rtn.ptr = tracked_unbound_script(ctx, us);
  return rtn;
}

MaybeLocal<Module> ResolveModuleCallback(Local<Context> context,
                                         Local<String> specifier,
                                         Local<FixedArray> import_attributes,
                                         Local<Module> referrer) {
  Local<Module> res;
  context->GetIsolate()->ThrowError("Error importing module");
  return res;
}

extern m_module* ScriptCompilerCompileModule(ContextPtr ctx,
                                             const char* s,
                                             const char* o) {
  LOCAL_CONTEXT(ctx);

  Local<String> src =
      String::NewFromUtf8(iso, s, NewStringType::kNormal).ToLocalChecked();
  Local<String> ogn =
      String::NewFromUtf8(iso, o, NewStringType::kNormal).ToLocalChecked();

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

  Local<Module> module;
  if (!ScriptCompiler::CompileModule(iso, &source).ToLocal(&module)) {
    return 0;
    // rtn.error = ExceptionError(try_catch, iso, local_ctx);
    // return rtn;
  }

  Maybe<bool> instantiateRes =
      module->InstantiateModule(local_ctx, ResolveModuleCallback);
  if (instantiateRes.IsNothing()) {
    return 0;
    // rtn.error = ExceptionError(try_catch, iso, local_ctx);
    // return rtn;
  }

  m_module* retVal = new m_module;
  retVal->ptr.Reset(iso, module);
  return retVal;
}
