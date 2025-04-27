#include "object_template.h"
#include "context.h"
#include "deps/include/v8-context.h"
#include "deps/include/v8-isolate.h"
#include "deps/include/v8-locker.h"
#include "deps/include/v8-template.h"
#include "function_template.h"
#include "template-macros.h"

using namespace v8;

TemplatePtr NewObjectTemplate(v8Isolate* iso) {
  Locker locker(iso);
  Isolate::Scope isolate_scope(iso);
  HandleScope handle_scope(iso);

  m_template* ot = new m_template;
  ot->iso = iso;
  ot->ptr.Reset(iso, ObjectTemplate::New(iso));
  return ot;
}

RtnValue ObjectTemplateNewInstance(TemplatePtr ptr, m_ctx* ctx) {
  LOCAL_TEMPLATE(ptr);
  TryCatch try_catch(iso);
  Local<Context> local_ctx = ctx->ptr.Get(iso);
  Context::Scope context_scope(local_ctx);

  RtnValue rtn = {};

  Local<ObjectTemplate> obj_tmpl = tmpl.As<ObjectTemplate>();
  Local<Object> obj;
  if (!obj_tmpl->NewInstance(local_ctx).ToLocal(&obj)) {
    rtn.error = ExceptionError(try_catch, iso, local_ctx);
    return rtn;
  }

  m_value* val = new m_value;
  val->id = 0;
  val->iso = iso;
  val->ctx = ctx;
  val->ptr = Global<Value>(iso, obj);
  rtn.value = tracked_value(ctx, val);
  return rtn;
}

void ObjectTemplateSetInternalFieldCount(TemplatePtr ptr, int field_count) {
  LOCAL_TEMPLATE(ptr);

  Local<ObjectTemplate> obj_tmpl = tmpl.As<ObjectTemplate>();
  obj_tmpl->SetInternalFieldCount(field_count);
}

void ObjectTemplateMarkAsUndetectable(TemplatePtr ptr) {
  LOCAL_TEMPLATE(ptr);

  Local<ObjectTemplate> obj_tmpl = tmpl.As<ObjectTemplate>();
  obj_tmpl->MarkAsUndetectable();
}

int ObjectTemplateInternalFieldCount(TemplatePtr ptr) {
  LOCAL_TEMPLATE(ptr);

  Local<ObjectTemplate> obj_tmpl = tmpl.As<ObjectTemplate>();
  return obj_tmpl->InternalFieldCount();
}

void ObjectTemplateSetAccessorProperty(TemplatePtr ptr,
                                       const char* key,
                                       TemplatePtr get,
                                       TemplatePtr set,
                                       int attributes) {
  LOCAL_TEMPLATE(ptr);

  Local<String> key_val =
      String::NewFromUtf8(iso, key, NewStringType::kNormal).ToLocalChecked();
  Local<ObjectTemplate> obj_tmpl = tmpl.As<ObjectTemplate>();
  Local<FunctionTemplate> get_tmpl =
      get ? get->ptr.Get(iso).As<FunctionTemplate>()
          : Local<FunctionTemplate>();
  Local<FunctionTemplate> set_tmpl =
      set ? set->ptr.Get(iso).As<FunctionTemplate>()
          : Local<FunctionTemplate>();

  return obj_tmpl->SetAccessorProperty(key_val, get_tmpl, set_tmpl,
                                       (PropertyAttribute)attributes);
}

void ObjectTemplateSetCallAsFunctionHandler(TemplatePtr ptr, int callback_ref) {
  LOCAL_TEMPLATE(ptr);

  Local<Integer> cbData = Integer::New(iso, callback_ref);

  Local<ObjectTemplate> obj_tmpl = tmpl.As<ObjectTemplate>();
  obj_tmpl->SetCallAsFunctionHandler(FunctionTemplateCallback, cbData);
}
