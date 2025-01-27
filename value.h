#ifndef V8GO_VALUE_H
#define V8GO_VALUE_H

#include <stddef.h>
#include <stdint.h>

#include "errors.h"

#ifdef __cplusplus

#include "deps/include/v8-persistent-handle.h"

namespace v8 {
class Isolate;
class Value;
}  // namespace v8

typedef struct m_ctx m_ctx;

struct m_value {
  long id;
  v8::Isolate* iso;
  m_ctx* ctx;
  v8::Global<v8::Value> ptr;
};

typedef v8::Isolate v8Isolate;

extern "C" {
#else

typedef struct v8Isolate v8Isolate;
typedef struct m_value m_value;

#endif

typedef m_value* ValuePtr;
typedef v8Isolate* IsolatePtr;

typedef struct v8BackingStore v8BackingStore;
typedef v8BackingStore* BackingStorePtr;

typedef struct {
  const uint64_t* word_array;
  int word_count;
  int sign_bit;
} ValueBigInt;

typedef struct {
  const char* data;
  int length;
  RtnError error;
} RtnString;

void ValueRelease(ValuePtr ptr);
extern RtnString ValueToString(ValuePtr ptr);
const uint32_t* ValueToArrayIndex(ValuePtr ptr);
int ValueToBoolean(ValuePtr ptr);
int32_t ValueToInt32(ValuePtr ptr);
int64_t ValueToInteger(ValuePtr ptr);
double ValueToNumber(ValuePtr ptr);
RtnString ValueToDetailString(ValuePtr ptr);
uint32_t ValueToUint32(ValuePtr ptr);
extern ValueBigInt ValueToBigInt(ValuePtr ptr);
extern RtnValue ValueToObject(ValuePtr ptr);
int ValueSameValue(ValuePtr ptr, ValuePtr otherPtr);
int ValueIsUndefined(ValuePtr ptr);
int ValueIsNull(ValuePtr ptr);
int ValueIsNullOrUndefined(ValuePtr ptr);
int ValueIsTrue(ValuePtr ptr);
int ValueIsFalse(ValuePtr ptr);
int ValueIsName(ValuePtr ptr);
int ValueIsString(ValuePtr ptr);
int ValueIsSymbol(ValuePtr ptr);
int ValueIsFunction(ValuePtr ptr);
int ValueIsObject(ValuePtr ptr);
int ValueIsBigInt(ValuePtr ptr);
int ValueIsBoolean(ValuePtr ptr);
int ValueIsNumber(ValuePtr ptr);
int ValueIsExternal(ValuePtr ptr);
int ValueIsInt32(ValuePtr ptr);
int ValueIsUint32(ValuePtr ptr);
int ValueIsDate(ValuePtr ptr);
int ValueIsArgumentsObject(ValuePtr ptr);
int ValueIsBigIntObject(ValuePtr ptr);
int ValueIsNumberObject(ValuePtr ptr);
int ValueIsStringObject(ValuePtr ptr);
int ValueIsSymbolObject(ValuePtr ptr);
int ValueIsNativeError(ValuePtr ptr);
int ValueIsRegExp(ValuePtr ptr);
int ValueIsAsyncFunction(ValuePtr ptr);
int ValueIsGeneratorFunction(ValuePtr ptr);
int ValueIsGeneratorObject(ValuePtr ptr);
int ValueIsPromise(ValuePtr ptr);
int ValueIsMap(ValuePtr ptr);
int ValueIsSet(ValuePtr ptr);
int ValueIsMapIterator(ValuePtr ptr);
int ValueIsSetIterator(ValuePtr ptr);
int ValueIsWeakMap(ValuePtr ptr);
int ValueIsWeakSet(ValuePtr ptr);
int ValueIsArray(ValuePtr ptr);
int ValueIsArrayBuffer(ValuePtr ptr);
int ValueIsArrayBufferView(ValuePtr ptr);
int ValueIsTypedArray(ValuePtr ptr);
int ValueIsUint8Array(ValuePtr ptr);
int ValueIsUint8ClampedArray(ValuePtr ptr);
int ValueIsInt8Array(ValuePtr ptr);
int ValueIsUint16Array(ValuePtr ptr);
int ValueIsInt16Array(ValuePtr ptr);
int ValueIsUint32Array(ValuePtr ptr);
int ValueIsInt32Array(ValuePtr ptr);
int ValueIsFloat32Array(ValuePtr ptr);
int ValueIsFloat64Array(ValuePtr ptr);
int ValueIsBigInt64Array(ValuePtr ptr);
int ValueIsBigUint64Array(ValuePtr ptr);
int ValueIsDataView(ValuePtr ptr);
int ValueIsSharedArrayBuffer(ValuePtr ptr);
int ValueIsProxy(ValuePtr ptr);
int ValueIsWasmModuleObject(ValuePtr ptr);
int ValueIsModuleNamespaceObject(ValuePtr ptr);
int ValueStrictEquals(ValuePtr ptr, ValuePtr otherPtr);

extern ValuePtr NewValueNull(IsolatePtr iso_ptr);
extern ValuePtr NewValueUndefined(IsolatePtr iso_ptr);
extern ValuePtr NewValueInteger(IsolatePtr iso_ptr, int32_t v);
extern ValuePtr NewValueIntegerFromUnsigned(IsolatePtr iso_ptr, uint32_t v);
extern RtnValue NewValueString(IsolatePtr iso_ptr, const char* v, int v_length);
extern ValuePtr NewValueBoolean(IsolatePtr iso_ptr, int v);
extern ValuePtr NewValueNumber(IsolatePtr iso_ptr, double v);
extern ValuePtr NewValueBigInt(IsolatePtr iso_ptr, int64_t v);
extern ValuePtr NewValueBigIntFromUnsigned(IsolatePtr iso_ptr, uint64_t v);
extern RtnValue NewValueBigIntFromWords(IsolatePtr iso_ptr,
                                        int sign_bit,
                                        int word_count,
                                        const uint64_t* words);
extern ValuePtr NewValueError(IsolatePtr iso_ptr,
                              ErrorTypeIndex idx,
                              const char* message);

const char* ExceptionGetMessageString(ValuePtr ptr);

extern void ObjectSet(ValuePtr ptr, const char* key, ValuePtr val_ptr);
extern void ObjectSetAnyKey(ValuePtr ptr, ValuePtr key, ValuePtr val_ptr);
extern void ObjectSetIdx(ValuePtr ptr, uint32_t idx, ValuePtr val_ptr);
extern int ObjectSetInternalField(ValuePtr ptr, int idx, ValuePtr val_ptr);
extern int ObjectInternalFieldCount(ValuePtr ptr);
extern RtnValue ObjectGet(ValuePtr ptr, const char* key);
extern RtnValue ObjectGetAnyKey(ValuePtr ptr, ValuePtr key);
extern RtnValue ObjectGetIdx(ValuePtr ptr, uint32_t idx);
extern RtnValue ObjectGetInternalField(ValuePtr ptr, int idx);
int ObjectHas(ValuePtr ptr, const char* key);
int ObjectHasAnyKey(ValuePtr ptr, ValuePtr key);
int ObjectHasIdx(ValuePtr ptr, uint32_t idx);
int ObjectDelete(ValuePtr ptr, const char* key);
int ObjectDeleteAnyKey(ValuePtr ptr, ValuePtr key);
int ObjectDeleteIdx(ValuePtr ptr, uint32_t idx);
extern RtnValue ObjectGetPrototype(ValuePtr ptr);
extern void ObjectSetPrototype(ValuePtr ptr, ValuePtr proto_ptr);

extern void BackingStoreRelease(BackingStorePtr ptr);
extern void* BackingStoreData(BackingStorePtr ptr);
extern size_t BackingStoreByteLength(BackingStorePtr ptr);
extern BackingStorePtr SharedArrayBufferGetBackingStore(ValuePtr ptr);

#ifdef __cplusplus
}
#endif
#endif
