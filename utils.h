#ifndef V8GO_UTILS_H
#define V8GO_UTILS_H

#include <string>

#include "deps/include/v8-primitive.h"

extern const char* CopyString(std::string str);
extern const char* CopyString(v8::String::Utf8Value& value);

#endif
