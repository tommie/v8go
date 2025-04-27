#include "script_compiler.h"
#include <iostream>
#include "context-macros.h"
#include "deps/include/v8-context.h"
#include "deps/include/v8-message.h"
#include "isolate-macros.h"

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
  context->ptr->GetIsolate()->ThrowError("Error importing module");
  return res;
}

extern RtnValue ScriptCompilerCompileModule(ContextPtr ctx) {
  puts("Compile module");
  fflush(stdout);
  std::cout << "Hello";
  LOCAL_CONTEXT(ctx);

  RtnValue rtn = {};

  Local<String> originURL;
  if (!String::NewFromUtf8(iso, "main.mjs").ToLocal(&originURL)) {
    rtn.error = ExceptionError(try_catch, iso, local_ctx);
    return rtn;
  }
  ScriptOrigin origin(originURL, 0, 0, false, -1, Local<Value>(), false, false,
                      true  // is_module
  );

  Local<String> source_text;

  if (!String::NewFromUtf8(iso, "import x from 'foo'; 1 + 1")
           .ToLocal(&source_text)) {
    rtn.error = ExceptionError(try_catch, iso, local_ctx);
    return rtn;
  }
  ScriptCompiler::Source source(source_text, origin);

  Local<Module> module;
  if (!ScriptCompiler::CompileModule(iso, &source).ToLocal(&module)) {
    // if you have a v8::TryCatch, you should check it here.
    rtn.error = ExceptionError(try_catch, iso, local_ctx);
    return rtn;
  }

  Maybe<bool> instantiateRes =
      module->InstantiateModule(local_ctx, ResolveModuleCallback);
  if (instantiateRes.IsNothing()) {
    rtn.error = ExceptionError(try_catch, iso, local_ctx);
    return rtn;
  }

  Local<Value> result;
  if (!module->Evaluate(local_ctx).ToLocal(&result)) {
    rtn.error = ExceptionError(try_catch, iso, local_ctx);
    return rtn;
  }

  m_value* new_val = new m_value;
  new_val->id = 0;
  new_val->iso = iso;
  new_val->ctx = ctx;
  new_val->ptr = Global<Value>(iso, result);

  rtn.value = tracked_value(ctx, new_val);
  return rtn;
}
