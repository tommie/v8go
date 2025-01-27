#ifndef V8GO_OBJECT_H
#define V8GO_OBJECT_H

#include <stdint.h>

#include "errors.h"

#ifdef __cplusplus

extern "C" {
#endif

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

#ifdef __cplusplus
}  // extern "C"
#endif
#endif
