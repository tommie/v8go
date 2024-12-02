#ifndef V8GO_INTEROP_H
#define V8GO_INTEROP_H

#include "errors.h"
#include "forward-declarations.h"

// A go-friendly return type combining value and error.
// Maybe this really belongs in value.h, but seems to be used from multiple
// places.
typedef struct {
  ValuePtr value;
  RtnError error;
} RtnValue;

#endif
