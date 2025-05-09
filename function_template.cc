#include "_cgo_export.h"

#include "deps/include/v8-context.h"
#include "deps/include/v8-function.h"
#include "isolate-macros.h"
#include "template-macros.h"
#include "template.h"

using namespace v8;

void FunctionTemplateCallback(const FunctionCallbackInfo<Value>& info) {
  Isolate* iso = info.GetIsolate();
  ISOLATE_SCOPE(iso);

  // This callback function can be called from any Context, which we only know
  // at runtime. We extract the Context reference from the embedder data so that
  // we can use the context registry to match the Context on the Go side
  Local<Context> local_ctx = iso->GetCurrentContext();
  int ctx_ref = local_ctx->GetEmbedderData(1).As<Integer>()->Value();
  m_ctx* ctx = goContext(ctx_ref);

  int callback_ref = info.Data().As<Integer>()->Value();

  m_value* _this = new m_value;
  _this->id = 0;
  _this->iso = iso;
  _this->ctx = ctx;
  _this->ptr.Reset(iso, Global<Value>(iso, info.This()));

  std::vector<ValuePtr> thisAndArgs(1 + info.Length());
  thisAndArgs[0] = tracked_value(ctx, _this);
  for (size_t i = 1; i < thisAndArgs.size(); ++i) {
    m_value* val = new m_value;
    val->id = 0;
    val->iso = iso;
    val->ctx = ctx;
    val->ptr.Reset(iso, Global<Value>(iso, info[i - 1]));
    thisAndArgs[i] = tracked_value(ctx, val);
  }

  goFunctionCallback_return retval =
      goFunctionCallback(ctx_ref, callback_ref, thisAndArgs.data(), thisAndArgs.size() - 1);
  if (retval.r1 != nullptr) {
    iso->ThrowException(retval.r1->ptr.Get(iso));
  } else if (retval.r0 != nullptr) {
    info.GetReturnValue().Set(retval.r0->ptr.Get(iso));
  } else {
    info.GetReturnValue().SetUndefined();
  }
}

TemplatePtr NewFunctionTemplate(IsolatePtr iso, int callback_ref) {
  Locker locker(iso);
  Isolate::Scope isolate_scope(iso);
  HandleScope handle_scope(iso);

  // (rogchap) We only need to store one value, callback_ref, into the
  // C++ callback function data, but if we needed to store more items we could
  // use an V8::Array; this would require the internal context from
  // iso->GetData(0)
  Local<Integer> cbData = Integer::New(iso, callback_ref);

  m_template* ot = new m_template;
  ot->iso = iso;
  ot->ptr.Reset(iso,
                FunctionTemplate::New(iso, FunctionTemplateCallback, cbData));
  return ot;
}

RtnValue FunctionTemplateGetFunction(m_template* ptr, m_ctx* ctx) {
  LOCAL_TEMPLATE(ptr);
  TryCatch try_catch(iso);
  Local<Context> local_ctx = ctx->ptr.Get(iso);
  Context::Scope context_scope(local_ctx);

  Local<FunctionTemplate> fn_tmpl = tmpl.As<FunctionTemplate>();
  RtnValue rtn = {};
  Local<Function> fn;
  if (!fn_tmpl->GetFunction(local_ctx).ToLocal(&fn)) {
    rtn.error = ExceptionError(try_catch, iso, local_ctx);
    return rtn;
  }

  m_value* val = new m_value;
  val->id = 0;
  val->iso = iso;
  val->ctx = ctx;
  val->ptr = Global<Value>(iso, fn);
  rtn.value = tracked_value(ctx, val);
  return rtn;
}

m_template* FunctionTemplateInstanceTemplate(m_template* ptr) {
  LOCAL_TEMPLATE(ptr);
  Local<FunctionTemplate> fn_tmpl = tmpl.As<FunctionTemplate>();
  m_template* ot = new m_template;
  ot->iso = iso;
  ot->ptr.Reset(iso, fn_tmpl->InstanceTemplate());

  return ot;
}

m_template* FunctionTemplatePrototypeTemplate(m_template* ptr) {
  LOCAL_TEMPLATE(ptr);
  Local<FunctionTemplate> fn_tmpl = tmpl.As<FunctionTemplate>();
  m_template* ot = new m_template;
  ot->iso = iso;
  ot->ptr.Reset(iso, fn_tmpl->PrototypeTemplate());

  return ot;
}

void FunctionTemplateInherit(TemplatePtr ptr, TemplatePtr base) {
  LOCAL_TEMPLATE(ptr);
  Local<FunctionTemplate> fn_tmpl = tmpl.As<FunctionTemplate>();
  Local<FunctionTemplate> base_tmp = base->ptr.Get(iso).As<FunctionTemplate>();
  fn_tmpl->Inherit(base_tmp);
}
