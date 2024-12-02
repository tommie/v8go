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

// It's not necessary to have forward declarations in a shared file; they can
// easily be redeclared where needed. Some of these are not _so_ trivial thouch,
// so types that are declared in may files may be extracted here.

#ifdef __cplusplus

// This block is visible when C++ code is compiled. Create the forward
// declarations in the namespace and class name they belong.

namespace v8 {
class Isolate;
};

typedef v8::Isolate v8Isolate;

extern "C" {

#else

// This block is visible when compiling Go code.

// Go code can't use C++ types and namespace, so their full type isn't declared,
// only as anonymous structs. Go only needs to know that it uses pointers to
// these types; but not the full definitions of those types.

typedef struct v8Isolate v8Isolate;

#endif

typedef struct m_ctx m_ctx;
typedef struct m_value m_value;

// Pointer types. Go code needs explicit pointer types to use pointers from
// C-code
typedef v8Isolate* IsolatePtr;
typedef m_ctx* ContextPtr;
typedef m_value* ValuePtr;

#ifdef __cplusplus
}  // extern "C"
#endif
#endif
