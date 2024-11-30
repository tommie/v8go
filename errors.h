#ifndef V8GO_ERRORS_H
#define V8GO_ERRORS_H

typedef struct {
  const char* msg;
  const char* location;
  const char* stack;
} RtnError;

#endif
