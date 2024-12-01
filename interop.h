#ifndef V8GO_INTEROP_H
#define V8GO_INTEROP_H

#include "errors.h"

typedef struct m_value m_value;
typedef m_value* ValuePtr;

typedef struct {
  ValuePtr value;
  RtnError error;
} RtnValue;

#endif
