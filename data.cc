#include "data.h"
#include "context-macros.h"
#include "deps/include/v8-data.h"

using namespace v8;

int FixedArrayLength(v8goFixedArray* fixedArray, m_ctx* ctx) {
  LOCAL_CONTEXT(ctx)
  Local<FixedArray> arr = fixedArray->ToLocal(iso);
  return arr->Length();
}

v8goData* FixedArrayGet(v8goFixedArray* fixedArray, m_ctx* ctx, int i) {
  LOCAL_CONTEXT(ctx)
  Local<FixedArray> arr = fixedArray->ToLocal(iso);
  return new v8goData(iso, arr->Get(local_ctx, i));
}

m_value* DataAsValue(v8goData* data, m_ctx* ctx) {
  LOCAL_CONTEXT(ctx);
  auto value = Local<Value>::Cast(data->ToLocal(iso));
  m_value* retVal = new m_value;
  retVal->id = 0;
  retVal->iso = iso;
  retVal->ctx = ctx;
  retVal->ptr = Global<Value>(iso, value);
  return tracked_value(ctx, retVal);
}

void DataRelease(v8goData* data) {
  delete data;
}
