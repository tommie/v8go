#include "module.h"

#include <string>
#include <vector>

#include "_cgo_export.h"
#include "context-macros.h"
#include "context.h"
#include "deps/include/v8-script.h"
#include "isolate-macros.h"
#include "utils.h"

using namespace v8;

// The Go ModuleStatus constants mirror v8::Module::Status.
static_assert(Module::kUninstantiated == 0);
static_assert(Module::kInstantiating == 1);
static_assert(Module::kInstantiated == 2);
static_assert(Module::kEvaluating == 3);
static_assert(Module::kEvaluated == 4);
static_assert(Module::kErrored == 5);

// Finds the m_module for a module compiled by IsolateCompileModule, or
// returns nullptr. V8 hands the resolver a Local<Module>, and the Go
// side needs to know which Module it is.
static m_module* findModule(m_ctx* ctx, Local<Module> module) {
  auto range = ctx->modules.equal_range(module->GetIdentityHash());
  for (auto it = range.first; it != range.second; ++it) {
    if (it->second->ptr == module) {
      return it->second;
    }
  }
  return nullptr;
}

RtnModule IsolateCompileModule(IsolatePtr iso, const char* s, const char* o) {
  ISOLATE_SCOPE(iso);
  m_ctx* ctx = isolateInternalContext(iso);
  TryCatch try_catch(iso);
  Local<Context> local_ctx = ctx->ptr.Get(iso);
  Context::Scope context_scope(local_ctx);

  RtnModule rtn = {};

  Local<String> src, ogn;
  if (!String::NewFromUtf8(iso, s).ToLocal(&src) ||
      !String::NewFromUtf8(iso, o).ToLocal(&ogn)) {
    rtn.error = ExceptionError(try_catch, iso, local_ctx);
    return rtn;
  }

  ScriptOrigin origin(ogn,
                      0,               // resource_line_offset
                      0,               // resource_column_offset
                      false,           // resource_is_shared_cross_origin
                      -1,              // script_id
                      Local<Value>(),  // source_map_url
                      false,           // resource_is_opaque
                      false,           // is_wasm
                      true);           // is_module
  ScriptCompiler::Source source(src, origin);

  Local<Module> module;
  if (!ScriptCompiler::CompileModule(iso, &source).ToLocal(&module)) {
    rtn.error = ExceptionError(try_catch, iso, local_ctx);
    return rtn;
  }

  // Modules are owned by the internal context, like unbound scripts, so
  // they live until the Isolate is disposed. Keeping them indexed lets
  // ResolveModuleCallback find the referrer.
  m_module* mod = new m_module;
  mod->ptr.Reset(iso, module);
  ctx->modules.emplace(module->GetIdentityHash(), mod);
  rtn.ptr = mod;
  return rtn;
}

int ModuleGetStatus(IsolatePtr iso, ModulePtr module) {
  ISOLATE_SCOPE(iso);
  return module->ptr.Get(iso)->GetStatus();
}

int ModuleScriptId(IsolatePtr iso, ModulePtr module) {
  ISOLATE_SCOPE(iso);
  return module->ptr.Get(iso)->ScriptId();
}

static MaybeLocal<Module> ResolveModuleCallback(
    Local<Context> context,
    Local<String> specifier,
    Local<FixedArray> import_attributes,
    Local<Module> referrer) {
  Isolate* iso = Isolate::GetCurrent();
  int ctx_ref =
      context->GetEmbedderDataV2(ContextDataIndex::REF).As<Integer>()->Value();

  String::Utf8Value spec(iso, specifier);

  // The attributes are triples of key, value and source offset. The
  // strings are kept here until the Go resolver has returned.
  int n = import_attributes->Length() / 3;
  std::vector<std::string> strs(2 * n);
  std::vector<ModuleImportAttribute> attrs(n);
  for (int i = 0; i < n; ++i) {
    strs[2 * i] =
        *String::Utf8Value(iso, import_attributes->Get(3 * i).As<String>());
    strs[2 * i + 1] =
        *String::Utf8Value(iso, import_attributes->Get(3 * i + 1).As<String>());
    attrs[i].key = strs[2 * i].c_str();
    attrs[i].value = strs[2 * i + 1].c_str();
    attrs[i].source_offset =
        import_attributes->Get(3 * i + 2).As<Int32>()->Value();
  }

  m_module* ref = findModule(isolateInternalContext(iso), referrer);

  goResolveModule_return retval =
      goResolveModule(ctx_ref, *spec, attrs.data(), n, ref);
  if (retval.r1 != nullptr) {
    iso->ThrowException(retval.r1->ptr.Get(iso));
    return MaybeLocal<Module>();
  }
  return retval.r0->ptr.Get(iso);
}

RtnError ModuleInstantiate(ContextPtr ctx, ModulePtr module) {
  LOCAL_CONTEXT(ctx);

  RtnError rtn = {};
  if (module->ptr.Get(iso)
          ->InstantiateModule(local_ctx, ResolveModuleCallback)
          .IsNothing()) {
    rtn = ExceptionError(try_catch, iso, local_ctx);
  }
  return rtn;
}

RtnValue ModuleEvaluate(ContextPtr ctx, ModulePtr module) {
  LOCAL_CONTEXT(ctx);

  RtnValue rtn = {};
  Local<Value> result;
  if (!module->ptr.Get(iso)->Evaluate(local_ctx).ToLocal(&result)) {
    rtn.error = ExceptionError(try_catch, iso, local_ctx);
    return rtn;
  }
  rtn.value = track_value(ctx, result);
  return rtn;
}

RtnValue ModuleGetNamespace(ContextPtr ctx, ModulePtr module) {
  LOCAL_CONTEXT(ctx);

  RtnValue rtn = {};
  Local<Module> mod = module->ptr.Get(iso);
  // V8 aborts if the module isn't instantiated yet.
  if (mod->GetStatus() < Module::kInstantiated) {
    rtn.error.msg = CopyString("module is not instantiated");
    return rtn;
  }
  rtn.value = track_value(ctx, mod->GetModuleNamespace());
  return rtn;
}
