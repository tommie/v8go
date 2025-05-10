#include "data.h"
#include "context-macros.h"
#include "deps/include/v8-data.h"

using namespace v8;

int FixedArrayLength(v8goFixedArray* fixedArray, m_ctx* ctx) {
  LOCAL_CONTEXT(ctx)
  Local<FixedArray> arr = fixedArray->ToLocal(iso);
  return arr->Length();
}
