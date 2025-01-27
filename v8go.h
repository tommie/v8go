// Copyright 2019 Roger Chapman and the v8go contributors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

#ifndef V8GO_H
#define V8GO_H

#include "context.h"
#include "errors.h"
#include "isolate.h"
#include "value.h"

#ifdef __cplusplus

#include "deps/include/libplatform/libplatform.h"
#include "deps/include/v8-profiler.h"
#include "deps/include/v8.h"

typedef v8::CpuProfiler* CpuProfilerPtr;
typedef v8::CpuProfile* CpuProfilePtr;
typedef const v8::CpuProfileNode* CpuProfileNodePtr;

extern "C" {
#else

// Opaque to cgo, but useful to treat it as a pointer to a distinct type

typedef struct v8CpuProfiler v8CpuProfiler;
typedef v8CpuProfiler* CpuProfilerPtr;

typedef struct v8CpuProfile v8CpuProfile;
typedef v8CpuProfile* CpuProfilePtr;

typedef struct v8CpuProfileNode v8CpuProfileNode;
typedef const v8CpuProfileNode* CpuProfileNodePtr;

#endif

#include "unbound_script.h"
#include "value.h"

// Opaque to both C and C++
typedef struct v8BackingStore v8BackingStore;
typedef v8BackingStore* BackingStorePtr;

#include <stddef.h>
#include <stdint.h>

typedef struct m_template m_template;

typedef m_template* TemplatePtr;

typedef struct {
  CpuProfilerPtr ptr;
  IsolatePtr iso;
} CPUProfiler;

typedef struct CPUProfileNode {
  CpuProfileNodePtr ptr;
  unsigned nodeId;
  int scriptId;
  const char* scriptResourceName;
  const char* functionName;
  int lineNumber;
  int columnNumber;
  unsigned hitCount;
  const char* bailoutReason;
  int childrenCount;
  struct CPUProfileNode** children;
} CPUProfileNode;

typedef struct {
  CpuProfilePtr ptr;
  const char* title;
  CPUProfileNode* root;
  int64_t startTime;
  int64_t endTime;
} CPUProfile;

extern void Init();

extern CPUProfiler* NewCPUProfiler(IsolatePtr iso_ptr);
extern void CPUProfilerDispose(CPUProfiler* ptr);
extern void CPUProfilerStartProfiling(CPUProfiler* ptr, const char* title);
extern CPUProfile* CPUProfilerStopProfiling(CPUProfiler* ptr,
                                            const char* title);
extern void CPUProfileDelete(CPUProfile* ptr);

extern void TemplateFreeWrapper(TemplatePtr ptr);
extern void TemplateSetValue(TemplatePtr ptr,
                             const char* name,
                             ValuePtr val_ptr,
                             int attributes);
extern int TemplateSetAnyValue(TemplatePtr ptr,
                               ValuePtr key,
                               ValuePtr val_ptr,
                               int attributes);
extern void TemplateSetTemplate(TemplatePtr ptr,
                                const char* name,
                                TemplatePtr obj_ptr,
                                int attributes);
extern int TemplateSetAnyTemplate(TemplatePtr ptr,
                                  ValuePtr key,
                                  TemplatePtr obj_ptr,
                                  int attributes);

const char* ExceptionGetMessageString(ValuePtr ptr);

extern RtnValue NewPromiseResolver(ContextPtr ctx_ptr);
extern ValuePtr PromiseResolverGetPromise(ValuePtr ptr);
int PromiseResolverResolve(ValuePtr ptr, ValuePtr val_ptr);
int PromiseResolverReject(ValuePtr ptr, ValuePtr val_ptr);
int PromiseState(ValuePtr ptr);
RtnValue PromiseThen(ValuePtr ptr, int callback_ref);
RtnValue PromiseThen2(ValuePtr ptr, int on_fulfilled_ref, int on_rejected_ref);
RtnValue PromiseCatch(ValuePtr ptr, int callback_ref);
extern ValuePtr PromiseResult(ValuePtr ptr);

extern RtnValue FunctionCall(ValuePtr ptr,
                             ValuePtr recv,
                             int argc,
                             ValuePtr argv[]);
RtnValue FunctionNewInstance(ValuePtr ptr, int argc, ValuePtr args[]);
ValuePtr FunctionSourceMapUrl(ValuePtr ptr);

const char* Version();
extern void SetFlags(const char* flags);

#ifdef __cplusplus
}  // extern "C"
#endif
#endif  // V8GO_H
