#ifndef V8GO_INTEROP_H
#define V8GO_INTEROP_H

#include "errors.h"

typedef struct m_value m_value;
typedef m_value* ValuePtr;

// A go-friendly return type combining value and error.
// Maybe this really belongs in value.h, but seems to be used from multiple
// places.
typedef struct {
  ValuePtr value;
  RtnError error;
} RtnValue;

#endif
