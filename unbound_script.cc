/********** UnboundScript & ScriptCompilerCachedData **********/

#include "unbound_script.h"
#include "context-macros.h"
#include "isolate-macros.h"

namespace v8 {
class Isolate;
}
using namespace v8;

ScriptCompilerCachedData* UnboundScriptCreateCodeCache(
    Isolate* iso,
    UnboundScriptPtr us_ptr) {
  ISOLATE_SCOPE(iso);

  Local<UnboundScript> unbound_script = us_ptr->ptr.Get(iso);

  ScriptCompiler::CachedData* cached_data =
      ScriptCompiler::CreateCodeCache(unbound_script);

  ScriptCompilerCachedData* cd = new ScriptCompilerCachedData;
  cd->ptr = cached_data;
  cd->data = cached_data->data;
  cd->length = cached_data->length;
  cd->rejected = cached_data->rejected;
  return cd;
}

void ScriptCompilerCachedDataDelete(ScriptCompilerCachedData* cached_data) {
  delete cached_data->ptr;
  delete cached_data;
}

// This can only run in contexts that belong to the same isolate
// the script was compiled in
RtnValue UnboundScriptRun(ContextPtr ctx, UnboundScriptPtr us_ptr) {
  LOCAL_CONTEXT(ctx)

  RtnValue rtn = {};

  Local<UnboundScript> unbound_script = us_ptr->ptr.Get(iso);

  Local<Script> script = unbound_script->BindToCurrentContext();
  Local<Value> result;
  if (!script->Run(local_ctx).ToLocal(&result)) {
    rtn.error = ExceptionError(try_catch, iso, local_ctx);
    return rtn;
  }
  m_value* val = new m_value;
  val->id = 0;
  val->iso = iso;
  val->ctx = ctx;
  val->ptr = Global<Value>(iso, result);

  rtn.value = tracked_value(ctx, val);
  return rtn;
}
