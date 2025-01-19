#include "object.h"
#include "deps/include/v8-object.h"
#include "isolate-macros.h"
#include "utils.h"
#include "value-macros.h"
#include "value.h"

using namespace v8;

/********** Object **********/

#define LOCAL_OBJECT(ptr) \
  LOCAL_VALUE(ptr)        \
  Local<Object> obj = value.As<Object>()

void ObjectSet(ValuePtr ptr, const char* key, ValuePtr prop_val) {
  LOCAL_OBJECT(ptr);
  Local<String> key_val =
      String::NewFromUtf8(iso, key, NewStringType::kNormal).ToLocalChecked();
  obj->Set(local_ctx, key_val, prop_val->ptr.Get(iso)).Check();
}

void ObjectSetAnyKey(ValuePtr ptr, ValuePtr key, ValuePtr prop_val) {
  LOCAL_OBJECT(ptr);
  Local<Value> local_key = key->ptr.Get(iso);
  obj->Set(local_ctx, local_key, prop_val->ptr.Get(iso)).Check();
}

void ObjectSetIdx(ValuePtr ptr, uint32_t idx, ValuePtr prop_val) {
  LOCAL_OBJECT(ptr);
  obj->Set(local_ctx, idx, prop_val->ptr.Get(iso)).Check();
}

int ObjectSetInternalField(ValuePtr ptr, int idx, ValuePtr val_ptr) {
  LOCAL_OBJECT(ptr);
  m_value* prop_val = static_cast<m_value*>(val_ptr);

  if (idx >= obj->InternalFieldCount()) {
    return 0;
  }

  obj->SetInternalField(idx, prop_val->ptr.Get(iso));

  return 1;
}

int ObjectInternalFieldCount(ValuePtr ptr) {
  LOCAL_OBJECT(ptr);
  return obj->InternalFieldCount();
}

RtnValue ObjectGet(ValuePtr ptr, const char* key) {
  LOCAL_OBJECT(ptr);
  RtnValue rtn = {};

  Local<String> key_val;
  if (!String::NewFromUtf8(iso, key, NewStringType::kNormal)
           .ToLocal(&key_val)) {
    rtn.error = ExceptionError(try_catch, iso, local_ctx);
    return rtn;
  }
  Local<Value> result;
  if (!obj->Get(local_ctx, key_val).ToLocal(&result)) {
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

RtnValue ObjectGetAnyKey(ValuePtr ptr, ValuePtr key) {
  LOCAL_OBJECT(ptr);
  RtnValue rtn = {};

  Local<Value> local_key = key->ptr.Get(iso);
  Local<Value> result;
  if (!obj->Get(local_ctx, local_key).ToLocal(&result)) {
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

RtnValue ObjectGetInternalField(ValuePtr ptr, int idx) {
  LOCAL_OBJECT(ptr);
  RtnValue rtn = {};

  if (idx >= obj->InternalFieldCount()) {
    rtn.error.msg = CopyString("internal field index out of range");
    return rtn;
  }

  Local<Data> result = obj->GetInternalField(idx);
  if (!result->IsValue()) {
    rtn.error.msg = CopyString("the internal field did not contain a Value");
    return rtn;
  }

  m_value* new_val = new m_value;
  new_val->id = 0;
  new_val->iso = iso;
  new_val->ctx = ctx;
  new_val->ptr = Global<Value>(iso, result.As<Value>());

  rtn.value = tracked_value(ctx, new_val);
  return rtn;
}

RtnValue ObjectGetIdx(ValuePtr ptr, uint32_t idx) {
  LOCAL_OBJECT(ptr);
  RtnValue rtn = {};

  Local<Value> result;
  if (!obj->Get(local_ctx, idx).ToLocal(&result)) {
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

int ObjectHas(ValuePtr ptr, const char* key) {
  LOCAL_OBJECT(ptr);
  Local<String> key_val =
      String::NewFromUtf8(iso, key, NewStringType::kNormal).ToLocalChecked();
  return obj->Has(local_ctx, key_val).ToChecked();
}

int ObjectHasAnyKey(ValuePtr ptr, ValuePtr key) {
  LOCAL_OBJECT(ptr);
  Local<Value> local_key = key->ptr.Get(iso);
  return obj->Has(local_ctx, local_key).ToChecked();
}

int ObjectHasIdx(ValuePtr ptr, uint32_t idx) {
  LOCAL_OBJECT(ptr);
  return obj->Has(local_ctx, idx).ToChecked();
}

int ObjectDelete(ValuePtr ptr, const char* key) {
  LOCAL_OBJECT(ptr);
  Local<String> key_val =
      String::NewFromUtf8(iso, key, NewStringType::kNormal).ToLocalChecked();
  return obj->Delete(local_ctx, key_val).ToChecked();
}

int ObjectDeleteAnyKey(ValuePtr ptr, ValuePtr key) {
  LOCAL_OBJECT(ptr);
  Local<Value> local_key = key->ptr.Get(iso);
  return obj->Delete(local_ctx, local_key).ToChecked();
}

int ObjectDeleteIdx(ValuePtr ptr, uint32_t idx) {
  LOCAL_OBJECT(ptr);
  return obj->Delete(local_ctx, idx).ToChecked();
}
