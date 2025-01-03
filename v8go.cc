// Copyright 2019 Roger Chapman and the v8go contributors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

#include "v8go.h"

#include <stdio.h>

#include <cstdlib>
#include <cstring>
#include "utils.h"

#include "context-macros.h"
#include "isolate-macros.h"
#include "template-macros.h"
#include "template.h"
#include "value-macros.h"

using namespace v8;

void FunctionTemplateCallback(const FunctionCallbackInfo<Value>& info);

const int ScriptCompilerNoCompileOptions = ScriptCompiler::kNoCompileOptions;
const int ScriptCompilerConsumeCodeCache = ScriptCompiler::kConsumeCodeCache;
const int ScriptCompilerEagerCompile = ScriptCompiler::kEagerCompile;

m_unboundScript* tracked_unbound_script(m_ctx* ctx, m_unboundScript* us) {
  ctx->unboundScripts.push_back(us);

  return us;
}

extern "C" {

/********** Isolate **********/

#define ISOLATE_SCOPE_INTERNAL_CONTEXT(iso) \
  ISOLATE_SCOPE(iso);                       \
  m_ctx* ctx = isolateInternalContext(iso);

RtnUnboundScript IsolateCompileUnboundScript(IsolatePtr iso,
                                             const char* s,
                                             const char* o,
                                             CompileOptions opts) {
  ISOLATE_SCOPE_INTERNAL_CONTEXT(iso);
  TryCatch try_catch(iso);
  Local<Context> local_ctx = ctx->ptr.Get(iso);
  Context::Scope context_scope(local_ctx);

  RtnUnboundScript rtn = {};

  Local<String> src =
      String::NewFromUtf8(iso, s, NewStringType::kNormal).ToLocalChecked();
  Local<String> ogn =
      String::NewFromUtf8(iso, o, NewStringType::kNormal).ToLocalChecked();

  ScriptCompiler::CompileOptions option =
      static_cast<ScriptCompiler::CompileOptions>(opts.compileOption);

  ScriptCompiler::CachedData* cached_data = nullptr;

  if (opts.cachedData.data) {
    cached_data = new ScriptCompiler::CachedData(opts.cachedData.data,
                                                 opts.cachedData.length);
  }

  ScriptOrigin script_origin(ogn);

  ScriptCompiler::Source source(src, script_origin, cached_data);

  Local<UnboundScript> unbound_script;
  if (!ScriptCompiler::CompileUnboundScript(iso, &source, option)
           .ToLocal(&unbound_script)) {
    rtn.error = ExceptionError(try_catch, iso, local_ctx);
    return rtn;
  };

  if (cached_data) {
    rtn.cachedDataRejected = cached_data->rejected;
  }

  m_unboundScript* us = new m_unboundScript;
  us->ptr.Reset(iso, unbound_script);
  rtn.ptr = tracked_unbound_script(ctx, us);
  return rtn;
}

/********** Exceptions & Errors **********/

ValuePtr IsolateThrowException(IsolatePtr iso, ValuePtr value) {
  ISOLATE_SCOPE(iso);
  m_ctx* ctx = value->ctx;

  Local<Value> throw_ret_val = iso->ThrowException(value->ptr.Get(iso));

  m_value* new_val = new m_value;
  new_val->id = 0;
  new_val->iso = iso;
  new_val->ctx = ctx;
  new_val->ptr = Global<Value>(iso, throw_ret_val);

  return tracked_value(ctx, new_val);
}

/********** CpuProfiler **********/

CPUProfiler* NewCPUProfiler(IsolatePtr iso_ptr) {
  Isolate* iso = static_cast<Isolate*>(iso_ptr);
  Locker locker(iso);
  Isolate::Scope isolate_scope(iso);
  HandleScope handle_scope(iso);

  CPUProfiler* c = new CPUProfiler;
  c->iso = iso;
  c->ptr = CpuProfiler::New(iso);
  return c;
}

void CPUProfilerDispose(CPUProfiler* profiler) {
  if (profiler->ptr == nullptr) {
    return;
  }
  profiler->ptr->Dispose();

  delete profiler;
}

void CPUProfilerStartProfiling(CPUProfiler* profiler, const char* title) {
  if (profiler->iso == nullptr) {
    return;
  }

  Locker locker(profiler->iso);
  Isolate::Scope isolate_scope(profiler->iso);
  HandleScope handle_scope(profiler->iso);

  Local<String> title_str =
      String::NewFromUtf8(profiler->iso, title, NewStringType::kNormal)
          .ToLocalChecked();
  profiler->ptr->StartProfiling(title_str);
}

CPUProfileNode* NewCPUProfileNode(const CpuProfileNode* ptr_) {
  int count = ptr_->GetChildrenCount();
  CPUProfileNode** children = new CPUProfileNode*[count];
  for (int i = 0; i < count; ++i) {
    children[i] = NewCPUProfileNode(ptr_->GetChild(i));
  }

  CPUProfileNode* root = new CPUProfileNode{
      ptr_,
      ptr_->GetNodeId(),
      ptr_->GetScriptId(),
      ptr_->GetScriptResourceNameStr(),
      ptr_->GetFunctionNameStr(),
      ptr_->GetLineNumber(),
      ptr_->GetColumnNumber(),
      ptr_->GetHitCount(),
      ptr_->GetBailoutReason(),
      count,
      children,
  };
  return root;
}

CPUProfile* CPUProfilerStopProfiling(CPUProfiler* profiler, const char* title) {
  if (profiler->iso == nullptr) {
    return nullptr;
  }

  Locker locker(profiler->iso);
  Isolate::Scope isolate_scope(profiler->iso);
  HandleScope handle_scope(profiler->iso);

  Local<String> title_str =
      String::NewFromUtf8(profiler->iso, title, NewStringType::kNormal)
          .ToLocalChecked();

  CPUProfile* profile = new CPUProfile;
  profile->ptr = profiler->ptr->StopProfiling(title_str);

  Local<String> str = profile->ptr->GetTitle();
  String::Utf8Value t(profiler->iso, str);
  profile->title = CopyString(t);

  CPUProfileNode* root = NewCPUProfileNode(profile->ptr->GetTopDownRoot());
  profile->root = root;

  profile->startTime = profile->ptr->GetStartTime();
  profile->endTime = profile->ptr->GetEndTime();

  return profile;
}

void CPUProfileNodeDelete(CPUProfileNode* node) {
  for (int i = 0; i < node->childrenCount; ++i) {
    CPUProfileNodeDelete(node->children[i]);
  }

  delete[] node->children;
  delete node;
}

void CPUProfileDelete(CPUProfile* profile) {
  if (profile->ptr == nullptr) {
    return;
  }
  profile->ptr->Delete();
  free((void*)profile->title);

  CPUProfileNodeDelete(profile->root);

  delete profile;
}

/********** Template **********/

void TemplateFreeWrapper(TemplatePtr tmpl) {
  tmpl->ptr.Clear();  // Just does `val_ = 0;` without calling V8::DisposeGlobal
  delete tmpl;
}

void TemplateSetValue(TemplatePtr ptr,
                      const char* name,
                      ValuePtr val,
                      int attributes) {
  LOCAL_TEMPLATE(ptr);

  Local<String> prop_name =
      String::NewFromUtf8(iso, name, NewStringType::kNormal).ToLocalChecked();
  tmpl->Set(prop_name, val->ptr.Get(iso), (PropertyAttribute)attributes);
}

int TemplateSetAnyValue(TemplatePtr ptr,
                        ValuePtr key,
                        ValuePtr val,
                        int attributes) {
  LOCAL_TEMPLATE(ptr);

  Local<Value> local_key = key->ptr.Get(iso);
  if (!local_key->IsName()) {
    return false;
  }
  tmpl->Set(local_key.As<Name>(), val->ptr.Get(iso),
            (PropertyAttribute)attributes);
  return true;
}

void TemplateSetTemplate(TemplatePtr ptr,
                         const char* name,
                         TemplatePtr obj,
                         int attributes) {
  LOCAL_TEMPLATE(ptr);

  Local<String> prop_name =
      String::NewFromUtf8(iso, name, NewStringType::kNormal).ToLocalChecked();
  tmpl->Set(prop_name, obj->ptr.Get(iso), (PropertyAttribute)attributes);
}

int TemplateSetAnyTemplate(TemplatePtr ptr,
                           ValuePtr key,
                           TemplatePtr obj,
                           int attributes) {
  LOCAL_TEMPLATE(ptr);

  Local<Value> local_key = key->ptr.Get(iso);
  if (!local_key->IsName()) {
    return false;
  }
  tmpl->Set(Local<Name>::Cast(local_key), obj->ptr.Get(iso),
            (PropertyAttribute)attributes);
  return true;
}

/********** Context **********/

RtnValue JSONParse(ContextPtr ctx, const char* str) {
  LOCAL_CONTEXT(ctx);
  RtnValue rtn = {};

  Local<String> v8Str;
  if (!String::NewFromUtf8(iso, str, NewStringType::kNormal).ToLocal(&v8Str)) {
    rtn.error = ExceptionError(try_catch, iso, local_ctx);
  }

  Local<Value> result;
  if (!JSON::Parse(local_ctx, v8Str).ToLocal(&result)) {
    rtn.error = ExceptionError(try_catch, iso, local_ctx);
    return rtn;
  }
  m_value* val = new m_value;
  val->id = 0;
  val->iso = iso;
  val->ctx = ctx;
  val->ptr = Global<Value>(iso, result);

  rtn.value = tracked_value(ctx, val);
  return rtn;
}

const char* JSONStringify(ContextPtr ctx, ValuePtr val) {
  Isolate* iso;
  Local<Context> local_ctx;

  if (ctx != nullptr) {
    iso = ctx->iso;
  } else {
    iso = val->iso;
  }

  Locker locker(iso);
  Isolate::Scope isolate_scope(iso);
  HandleScope handle_scope(iso);

  if (ctx != nullptr) {
    local_ctx = ctx->ptr.Get(iso);
  } else {
    if (val->ctx != nullptr) {
      local_ctx = val->ctx->ptr.Get(iso);
    } else {
      m_ctx* ctx = isolateInternalContext(iso);
      local_ctx = ctx->ptr.Get(iso);
    }
  }

  Context::Scope context_scope(local_ctx);

  Local<String> str;
  if (!JSON::Stringify(local_ctx, val->ptr.Get(iso)).ToLocal(&str)) {
    return nullptr;
  }
  String::Utf8Value json(iso, str);
  return CopyString(json);
}

void ValueRelease(ValuePtr ptr) {
  if (ptr == nullptr) {
    return;
  }

  ptr->ctx->vals.erase(ptr->id);
  ptr->ptr.Reset();
  delete ptr;
}

ValuePtr NewValueInteger(IsolatePtr iso, int32_t v) {
  ISOLATE_SCOPE_INTERNAL_CONTEXT(iso);
  m_value* val = new m_value;
  val->id = 0;
  val->iso = iso;
  val->ctx = ctx;
  val->ptr = Global<Value>(iso, Integer::New(iso, v));
  return tracked_value(ctx, val);
}

ValuePtr NewValueIntegerFromUnsigned(IsolatePtr iso, uint32_t v) {
  ISOLATE_SCOPE_INTERNAL_CONTEXT(iso);
  m_value* val = new m_value;
  val->id = 0;
  val->iso = iso;
  val->ctx = ctx;
  val->ptr = Global<Value>(iso, Integer::NewFromUnsigned(iso, v));
  return tracked_value(ctx, val);
}

RtnValue NewValueString(IsolatePtr iso, const char* v, int v_length) {
  ISOLATE_SCOPE_INTERNAL_CONTEXT(iso);
  TryCatch try_catch(iso);
  RtnValue rtn = {};
  Local<String> str;
  if (!String::NewFromUtf8(iso, v, NewStringType::kNormal, v_length)
           .ToLocal(&str)) {
    rtn.error = ExceptionError(try_catch, iso, ctx->ptr.Get(iso));
    return rtn;
  }
  m_value* val = new m_value;
  val->id = 0;
  val->iso = iso;
  val->ctx = ctx;
  val->ptr = Global<Value>(iso, str);
  rtn.value = tracked_value(ctx, val);
  return rtn;
}

ValuePtr NewValueNull(IsolatePtr iso) {
  ISOLATE_SCOPE_INTERNAL_CONTEXT(iso);
  m_value* val = new m_value;
  val->id = 0;
  val->iso = iso;
  val->ctx = ctx;
  val->ptr = Global<Value>(iso, Null(iso));
  return tracked_value(ctx, val);
}

ValuePtr NewValueUndefined(IsolatePtr iso) {
  ISOLATE_SCOPE_INTERNAL_CONTEXT(iso);
  m_value* val = new m_value;
  val->id = 0;
  val->iso = iso;
  val->ctx = ctx;
  val->ptr = Global<Value>(iso, Undefined(iso));
  return tracked_value(ctx, val);
}

ValuePtr NewValueBoolean(IsolatePtr iso, int v) {
  ISOLATE_SCOPE_INTERNAL_CONTEXT(iso);
  m_value* val = new m_value;
  val->id = 0;
  val->iso = iso;
  val->ctx = ctx;
  val->ptr = Global<Value>(iso, Boolean::New(iso, v));
  return tracked_value(ctx, val);
}

ValuePtr NewValueNumber(IsolatePtr iso, double v) {
  ISOLATE_SCOPE_INTERNAL_CONTEXT(iso);
  m_value* val = new m_value;
  val->id = 0;
  val->iso = iso;
  val->ctx = ctx;
  val->ptr = Global<Value>(iso, Number::New(iso, v));
  return tracked_value(ctx, val);
}

ValuePtr NewValueBigInt(IsolatePtr iso, int64_t v) {
  ISOLATE_SCOPE_INTERNAL_CONTEXT(iso);
  m_value* val = new m_value;
  val->id = 0;
  val->iso = iso;
  val->ctx = ctx;
  val->ptr = Global<Value>(iso, BigInt::New(iso, v));
  return tracked_value(ctx, val);
}

ValuePtr NewValueBigIntFromUnsigned(IsolatePtr iso, uint64_t v) {
  ISOLATE_SCOPE_INTERNAL_CONTEXT(iso);
  m_value* val = new m_value;
  val->id = 0;
  val->iso = iso;
  val->ctx = ctx;
  val->ptr = Global<Value>(iso, BigInt::NewFromUnsigned(iso, v));
  return tracked_value(ctx, val);
}

RtnValue NewValueBigIntFromWords(IsolatePtr iso,
                                 int sign_bit,
                                 int word_count,
                                 const uint64_t* words) {
  ISOLATE_SCOPE_INTERNAL_CONTEXT(iso);
  TryCatch try_catch(iso);
  Local<Context> local_ctx = ctx->ptr.Get(iso);

  RtnValue rtn = {};
  Local<BigInt> bigint;
  if (!BigInt::NewFromWords(local_ctx, sign_bit, word_count, words)
           .ToLocal(&bigint)) {
    rtn.error = ExceptionError(try_catch, iso, local_ctx);
    return rtn;
  }
  m_value* val = new m_value;
  val->id = 0;
  val->iso = iso;
  val->ctx = ctx;
  val->ptr = Global<Value>(iso, bigint);
  rtn.value = tracked_value(ctx, val);
  return rtn;
}

ValuePtr NewValueError(IsolatePtr iso,
                       ErrorTypeIndex idx,
                       const char* message) {
  ISOLATE_SCOPE_INTERNAL_CONTEXT(iso);
  Local<Context> local_ctx = ctx->ptr.Get(iso);
  Context::Scope context_scope(local_ctx);

  Local<String> local_msg = String::NewFromUtf8(iso, message).ToLocalChecked();
  Local<Value> v;
  switch (idx) {
    case ERROR_RANGE:
      v = Exception::RangeError(local_msg);
      break;
    case ERROR_REFERENCE:
      v = Exception::ReferenceError(local_msg);
      break;
    case ERROR_SYNTAX:
      v = Exception::SyntaxError(local_msg);
      break;
    case ERROR_TYPE:
      v = Exception::TypeError(local_msg);
      break;
    case ERROR_WASM_COMPILE:
      v = Exception::WasmCompileError(local_msg);
      break;
    case ERROR_WASM_LINK:
      v = Exception::WasmLinkError(local_msg);
      break;
    case ERROR_WASM_RUNTIME:
      v = Exception::WasmRuntimeError(local_msg);
      break;
    case ERROR_GENERIC:
      v = Exception::Error(local_msg);
      break;
    default:
      return nullptr;
  }
  m_value* val = new m_value;
  val->id = 0;
  val->iso = iso;
  val->ctx = ctx;
  val->ptr = Global<Value>(iso, v);
  return tracked_value(ctx, val);
}

const uint32_t* ValueToArrayIndex(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  Local<Uint32> array_index;
  if (!value->ToArrayIndex(local_ctx).ToLocal(&array_index)) {
    return nullptr;
  }

  uint32_t* idx = (uint32_t*)malloc(sizeof(uint32_t));
  *idx = array_index->Value();
  return idx;
}

int ValueToBoolean(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->BooleanValue(iso);
}

int32_t ValueToInt32(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->Int32Value(local_ctx).ToChecked();
}

int64_t ValueToInteger(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IntegerValue(local_ctx).ToChecked();
}

double ValueToNumber(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->NumberValue(local_ctx).ToChecked();
}

RtnString ValueToDetailString(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  RtnString rtn = {0};
  Local<String> str;
  if (!value->ToDetailString(local_ctx).ToLocal(&str)) {
    rtn.error = ExceptionError(try_catch, iso, local_ctx);
    return rtn;
  }
  String::Utf8Value ds(iso, str);
  rtn.data = CopyString(ds);
  rtn.length = ds.length();
  return rtn;
}

RtnString ValueToString(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  RtnString rtn = {0};
  // String::Utf8Value will result in an empty string if conversion to a string
  // fails
  // TODO: Consider propagating the JS error. A fallback value could be returned
  // in Value.String()
  String::Utf8Value src(iso, value);
  char* data = static_cast<char*>(malloc(src.length()));
  memcpy(data, *src, src.length());
  rtn.data = data;
  rtn.length = src.length();
  return rtn;
}

uint32_t ValueToUint32(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->Uint32Value(local_ctx).ToChecked();
}

ValueBigInt ValueToBigInt(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  Local<BigInt> bint;
  if (!value->ToBigInt(local_ctx).ToLocal(&bint)) {
    return {nullptr, 0};
  }

  int word_count = bint->WordCount();
  int sign_bit = 0;
  uint64_t* words = (uint64_t*)malloc(sizeof(uint64_t) * word_count);
  bint->ToWordsArray(&sign_bit, &word_count, words);
  ValueBigInt rtn = {words, word_count, sign_bit};
  return rtn;
}

RtnValue ValueToObject(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  RtnValue rtn = {};
  Local<Object> obj;
  if (!value->ToObject(local_ctx).ToLocal(&obj)) {
    rtn.error = ExceptionError(try_catch, iso, local_ctx);
    return rtn;
  }
  m_value* new_val = new m_value;
  new_val->id = 0;
  new_val->iso = iso;
  new_val->ctx = ctx;
  new_val->ptr = Global<Value>(iso, obj);
  rtn.value = tracked_value(ctx, new_val);
  return rtn;
}

int ValueSameValue(ValuePtr val1, ValuePtr val2) {
  Isolate* iso = val1->iso;
  ISOLATE_SCOPE(iso);
  Local<Value> value1 = val1->ptr.Get(iso);
  Local<Value> value2 = val2->ptr.Get(iso);

  return value1->SameValue(value2);
}

int ValueIsUndefined(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsUndefined();
}

int ValueIsNull(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsNull();
}

int ValueIsNullOrUndefined(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsNullOrUndefined();
}

int ValueIsTrue(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsTrue();
}

int ValueIsFalse(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsFalse();
}

int ValueIsName(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsName();
}

int ValueIsString(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsString();
}

int ValueIsSymbol(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsSymbol();
}

int ValueIsFunction(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsFunction();
}

int ValueIsObject(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsObject();
}

int ValueIsBigInt(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsBigInt();
}

int ValueIsBoolean(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsBoolean();
}

int ValueIsNumber(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsNumber();
}

int ValueIsExternal(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsExternal();
}

int ValueIsInt32(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsInt32();
}

int ValueIsUint32(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsUint32();
}

int ValueIsDate(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsDate();
}

int ValueIsArgumentsObject(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsArgumentsObject();
}

int ValueIsBigIntObject(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsBigIntObject();
}

int ValueIsNumberObject(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsNumberObject();
}

int ValueIsStringObject(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsStringObject();
}

int ValueIsSymbolObject(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsSymbolObject();
}

int ValueIsNativeError(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsNativeError();
}

int ValueIsRegExp(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsRegExp();
}

int ValueIsAsyncFunction(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsAsyncFunction();
}

int ValueIsGeneratorFunction(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsGeneratorFunction();
}

int ValueIsGeneratorObject(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsGeneratorObject();
}

int ValueIsPromise(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsPromise();
}

int ValueIsMap(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsMap();
}

int ValueIsSet(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsSet();
}

int ValueIsMapIterator(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsMapIterator();
}

int ValueIsSetIterator(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsSetIterator();
}

int ValueIsWeakMap(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsWeakMap();
}

int ValueIsWeakSet(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsWeakSet();
}

int ValueIsArray(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsArray();
}

int ValueIsArrayBuffer(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsArrayBuffer();
}

int ValueIsArrayBufferView(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsArrayBufferView();
}

int ValueIsTypedArray(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsTypedArray();
}

int ValueIsUint8Array(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsUint8Array();
}

int ValueIsUint8ClampedArray(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsUint8ClampedArray();
}

int ValueIsInt8Array(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsInt8Array();
}

int ValueIsUint16Array(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsUint16Array();
}

int ValueIsInt16Array(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsInt16Array();
}

int ValueIsUint32Array(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsUint32Array();
}

int ValueIsInt32Array(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsInt32Array();
}

int ValueIsFloat32Array(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsFloat32Array();
}

int ValueIsFloat64Array(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsFloat64Array();
}

int ValueIsBigInt64Array(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsBigInt64Array();
}

int ValueIsBigUint64Array(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsBigUint64Array();
}

int ValueIsDataView(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsDataView();
}

int ValueIsSharedArrayBuffer(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsSharedArrayBuffer();
}

int ValueIsProxy(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsProxy();
}

int ValueIsWasmModuleObject(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsWasmModuleObject();
}

int ValueIsModuleNamespaceObject(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsModuleNamespaceObject();
}

/********** Exception **********/

const char* ExceptionGetMessageString(ValuePtr ptr) {
  LOCAL_VALUE(ptr);

  Local<Message> local_msg = Exception::CreateMessage(iso, value);
  Local<String> local_str = local_msg->Get();
  String::Utf8Value utf8(iso, local_str);
  return CopyString(utf8);
}

/********** Promise **********/

RtnValue NewPromiseResolver(ContextPtr ctx) {
  LOCAL_CONTEXT(ctx);
  RtnValue rtn = {};
  Local<Promise::Resolver> resolver;
  if (!Promise::Resolver::New(local_ctx).ToLocal(&resolver)) {
    rtn.error = ExceptionError(try_catch, iso, local_ctx);
    return rtn;
  }
  m_value* val = new m_value;
  val->id = 0;
  val->iso = iso;
  val->ctx = ctx;
  val->ptr = Global<Value>(iso, resolver);
  rtn.value = tracked_value(ctx, val);
  return rtn;
}

ValuePtr PromiseResolverGetPromise(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  Local<Promise::Resolver> resolver = value.As<Promise::Resolver>();
  Local<Promise> promise = resolver->GetPromise();
  m_value* promise_val = new m_value;
  promise_val->id = 0;
  promise_val->iso = iso;
  promise_val->ctx = ctx;
  promise_val->ptr = Global<Value>(iso, promise);
  return tracked_value(ctx, promise_val);
}

int PromiseResolverResolve(ValuePtr ptr, ValuePtr resolve_val) {
  LOCAL_VALUE(ptr);
  Local<Promise::Resolver> resolver = value.As<Promise::Resolver>();
  return resolver->Resolve(local_ctx, resolve_val->ptr.Get(iso)).ToChecked();
}

int PromiseResolverReject(ValuePtr ptr, ValuePtr reject_val) {
  LOCAL_VALUE(ptr);
  Local<Promise::Resolver> resolver = value.As<Promise::Resolver>();
  return resolver->Reject(local_ctx, reject_val->ptr.Get(iso)).ToChecked();
}

int PromiseState(ValuePtr ptr) {
  LOCAL_VALUE(ptr)
  Local<Promise> promise = value.As<Promise>();
  return promise->State();
}

RtnValue PromiseThen(ValuePtr ptr, int callback_ref) {
  LOCAL_VALUE(ptr)
  RtnValue rtn = {};
  Local<Promise> promise = value.As<Promise>();
  Local<Integer> cbData = Integer::New(iso, callback_ref);
  Local<Function> func;
  if (!Function::New(local_ctx, FunctionTemplateCallback, cbData)
           .ToLocal(&func)) {
    rtn.error = ExceptionError(try_catch, iso, local_ctx);
    return rtn;
  }
  Local<Promise> result;
  if (!promise->Then(local_ctx, func).ToLocal(&result)) {
    rtn.error = ExceptionError(try_catch, iso, local_ctx);
    return rtn;
  }
  m_value* result_val = new m_value;
  result_val->id = 0;
  result_val->iso = iso;
  result_val->ctx = ctx;
  result_val->ptr = Global<Value>(iso, result);
  rtn.value = tracked_value(ctx, result_val);
  return rtn;
}

RtnValue PromiseThen2(ValuePtr ptr, int on_fulfilled_ref, int on_rejected_ref) {
  LOCAL_VALUE(ptr)
  RtnValue rtn = {};
  Local<Promise> promise = value.As<Promise>();
  Local<Integer> onFulfilledData = Integer::New(iso, on_fulfilled_ref);
  Local<Function> onFulfilledFunc;
  if (!Function::New(local_ctx, FunctionTemplateCallback, onFulfilledData)
           .ToLocal(&onFulfilledFunc)) {
    rtn.error = ExceptionError(try_catch, iso, local_ctx);
    return rtn;
  }
  Local<Integer> onRejectedData = Integer::New(iso, on_rejected_ref);
  Local<Function> onRejectedFunc;
  if (!Function::New(local_ctx, FunctionTemplateCallback, onRejectedData)
           .ToLocal(&onRejectedFunc)) {
    rtn.error = ExceptionError(try_catch, iso, local_ctx);
    return rtn;
  }
  Local<Promise> result;
  if (!promise->Then(local_ctx, onFulfilledFunc, onRejectedFunc)
           .ToLocal(&result)) {
    rtn.error = ExceptionError(try_catch, iso, local_ctx);
    return rtn;
  }
  m_value* result_val = new m_value;
  result_val->id = 0;
  result_val->iso = iso;
  result_val->ctx = ctx;
  result_val->ptr = Global<Value>(iso, result);
  rtn.value = tracked_value(ctx, result_val);
  return rtn;
}

RtnValue PromiseCatch(ValuePtr ptr, int callback_ref) {
  LOCAL_VALUE(ptr)
  RtnValue rtn = {};
  Local<Promise> promise = value.As<Promise>();
  Local<Integer> cbData = Integer::New(iso, callback_ref);
  Local<Function> func;
  if (!Function::New(local_ctx, FunctionTemplateCallback, cbData)
           .ToLocal(&func)) {
    rtn.error = ExceptionError(try_catch, iso, local_ctx);
    return rtn;
  }
  Local<Promise> result;
  if (!promise->Catch(local_ctx, func).ToLocal(&result)) {
    rtn.error = ExceptionError(try_catch, iso, local_ctx);
    return rtn;
  }
  m_value* result_val = new m_value;
  result_val->id = 0;
  result_val->iso = iso;
  result_val->ctx = ctx;
  result_val->ptr = Global<Value>(iso, result);
  rtn.value = tracked_value(ctx, result_val);
  return rtn;
}

ValuePtr PromiseResult(ValuePtr ptr) {
  LOCAL_VALUE(ptr)
  Local<Promise> promise = value.As<Promise>();
  Local<Value> result = promise->Result();
  m_value* result_val = new m_value;
  result_val->id = 0;
  result_val->iso = iso;
  result_val->ctx = ctx;
  result_val->ptr = Global<Value>(iso, result);
  return tracked_value(ctx, result_val);
}

/********** Function **********/

static void buildCallArguments(Isolate* iso,
                               Local<Value>* argv,
                               int argc,
                               ValuePtr args[]) {
  for (int i = 0; i < argc; i++) {
    argv[i] = args[i]->ptr.Get(iso);
  }
}

RtnValue FunctionCall(ValuePtr ptr, ValuePtr recv, int argc, ValuePtr args[]) {
  LOCAL_VALUE(ptr)

  RtnValue rtn = {};
  Local<Function> fn = Local<Function>::Cast(value);
  Local<Value> argv[argc];
  buildCallArguments(iso, argv, argc, args);

  Local<Value> local_recv = recv->ptr.Get(iso);

  Local<Value> result;
  if (!fn->Call(local_ctx, local_recv, argc, argv).ToLocal(&result)) {
    rtn.error = ExceptionError(try_catch, iso, local_ctx);
    return rtn;
  }
  m_value* rtnval = new m_value;
  rtnval->id = 0;
  rtnval->iso = iso;
  rtnval->ctx = ctx;
  rtnval->ptr = Global<Value>(iso, result);
  rtn.value = tracked_value(ctx, rtnval);
  return rtn;
}

RtnValue FunctionNewInstance(ValuePtr ptr, int argc, ValuePtr args[]) {
  LOCAL_VALUE(ptr)
  RtnValue rtn = {};
  Local<Function> fn = Local<Function>::Cast(value);
  Local<Value> argv[argc];
  buildCallArguments(iso, argv, argc, args);
  Local<Object> result;
  if (!fn->NewInstance(local_ctx, argc, argv).ToLocal(&result)) {
    rtn.error = ExceptionError(try_catch, iso, local_ctx);
    return rtn;
  }
  m_value* rtnval = new m_value;
  rtnval->id = 0;
  rtnval->iso = iso;
  rtnval->ctx = ctx;
  rtnval->ptr = Global<Value>(iso, result);
  rtn.value = tracked_value(ctx, rtnval);
  return rtn;
}

ValuePtr FunctionSourceMapUrl(ValuePtr ptr) {
  LOCAL_VALUE(ptr)
  Local<Function> fn = Local<Function>::Cast(value);
  Local<Value> result = fn->GetScriptOrigin().SourceMapUrl();
  m_value* rtnval = new m_value;
  rtnval->id = 0;
  rtnval->iso = iso;
  rtnval->ctx = ctx;
  rtnval->ptr = Global<Value>(iso, result);
  return tracked_value(ctx, rtnval);
}

/********** v8::V8 **********/

const char* Version() {
  return V8::GetVersion();
}

void SetFlags(const char* flags) {
  V8::SetFlagsFromString(flags);
}

/********** SharedArrayBuffer & BackingStore ***********/

struct v8BackingStore {
  v8BackingStore(std::shared_ptr<v8::BackingStore>&& ptr)
      : backing_store{ptr} {}
  std::shared_ptr<v8::BackingStore> backing_store;
};

BackingStorePtr SharedArrayBufferGetBackingStore(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  auto buffer = Local<SharedArrayBuffer>::Cast(value);
  auto backing_store = buffer->GetBackingStore();
  auto proxy = new v8BackingStore(std::move(backing_store));
  return proxy;
}

void BackingStoreRelease(BackingStorePtr ptr) {
  if (ptr == nullptr) {
    return;
  }
  ptr->backing_store.reset();
  delete ptr;
}

void* BackingStoreData(BackingStorePtr ptr) {
  if (ptr == nullptr) {
    return nullptr;
  }

  return ptr->backing_store->Data();
}

size_t BackingStoreByteLength(BackingStorePtr ptr) {
  if (ptr == nullptr) {
    return 0;
  }
  return ptr->backing_store->ByteLength();
}
}
