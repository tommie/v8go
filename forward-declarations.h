#ifndef V8GO_FORWARD_DECLARATIONS_H
#define V8GO_FORWARD_DECLARATIONS_H

// Create "forward declarations" of structs and classes.
//
// When a type is defined, the compiler doesn't need to know the full
// definition of the contained types. But the compiler needs to know the _size_
// of those types. By creating forward declarations of used classes and structs,
// pointers to these can be used in header files; without including the full
// type definition, when all we need is a pointer; as the pointer has a well
// defined type.
//
// This can speed up compilation significantly, when used cleverly.

#ifdef __cplusplus

// Create types available to C++ implementation code.

namespace v8 {
class Isolate;
};

typedef v8::Isolate v8Isolate;

extern "C" {

#else

// Go code can't use C++ types, so declare these as structs. For these types,
// we create specific pointer types for them, such that Go code can use the
// pointers.

typedef struct v8Isolate v8Isolate;

#endif

typedef struct m_ctx m_ctx;

// Pointer types. Go code needs explicit pointer types to use pointers from
// C-code
typedef v8Isolate* IsolatePtr;
typedef m_ctx* ContextPtr;

#ifdef __cplusplus
}  // extern "C"
#endif
#endif
