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

void DataRelease(v8goData* data) {
  delete data;
}
