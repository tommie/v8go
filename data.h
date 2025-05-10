#ifndef V8GO_DATA_H
#define V8GO_DATA_H

#ifdef __cplusplus

#include "deps/include/v8-persistent-handle.h"

namespace v8 {
class Isolate;
class Data;
class FixedArray;
}  // namespace v8

class v8goData {
  v8::Global<v8::Data> ptr;

 public:
  v8goData(v8::Isolate* iso, v8::Local<v8::Data> val) : ptr(iso, val) {};
};

class v8goFixedArray {
  v8::Global<v8::FixedArray> ptr;

 public:
  v8goFixedArray(v8::Isolate* iso, v8::Local<v8::FixedArray> val)
      : ptr(iso, val) {};
  v8::Local<v8::FixedArray> ToLocal(v8::Isolate* iso) { return ptr.Get(iso); }
};

extern "C" {
#else

typedef struct v8goData v8goData;
typedef struct v8goFixedv8goFixedArray v8goFixedArray;

#endif

typedef struct m_ctx m_ctx;

int FixedArrayLength(v8goFixedArray* fixedArray, m_ctx* ctx);
v8goData* FixedArrayGet(v8goFixedArray* fixedArray, m_ctx* ctx);
void DataRelease(v8goData* data);

#ifdef __cplusplus
}
#endif
#endif
