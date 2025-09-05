#include "deps/include/v8-context.h"
#include "deps/include/v8-initialization.h"
#include "deps/include/v8-locker.h"
#include "deps/include/v8-platform.h"

#include "context.h"
#include "isolate.h"
#include "libplatform/libplatform.h"

using namespace v8;

auto default_platform = platform::NewDefaultPlatform();
ArrayBuffer::Allocator* default_allocator;

extern "C" {

/********** Isolate **********/

#define ISOLATE_SCOPE(iso)           \
  Locker locker(iso);                \
  Isolate::Scope isolate_scope(iso); \
  HandleScope handle_scope(iso);

void Init() {
#ifdef _WIN32
  V8::InitializeExternalStartupData(".");
#endif
  V8::InitializePlatform(default_platform.get());
  V8::Initialize();

  default_allocator = ArrayBuffer::Allocator::NewDefaultAllocator();
  return;
}

size_t NearMemoryLimitCallback(void* data, size_t current_heap_limit, size_t initial_heap_limit)
{
  auto iso = static_cast<Isolate*>(data);
  iso->TerminateExecution();
  return current_heap_limit * 2;
}

IsolatePtr NewIsolateWithOptions(IsolateConstraints constraints, int has_constraints) {
  Isolate::CreateParams params;
  params.array_buffer_allocator = default_allocator;

  if (has_constraints) {
    ResourceConstraints rc;
    if (constraints.stack_limit != 0) {
      rc.set_stack_limit(reinterpret_cast<uint32_t*>(constraints.stack_limit));
    }
    if (constraints.code_range_size_in_bytes != 0) {
      rc.set_code_range_size_in_bytes(constraints.code_range_size_in_bytes);
    }
    if (constraints.max_old_generation_size_in_bytes != 0) {
      rc.set_max_old_generation_size_in_bytes(constraints.max_old_generation_size_in_bytes);
    }
    if (constraints.max_young_generation_size_in_bytes != 0) {
      rc.set_max_young_generation_size_in_bytes(constraints.max_young_generation_size_in_bytes);
    }
    if (constraints.initial_old_generation_size_in_bytes != 0) {
      rc.set_initial_old_generation_size_in_bytes(constraints.initial_old_generation_size_in_bytes);
    }
    if (constraints.initial_young_generation_size_in_bytes != 0) {
      rc.set_initial_young_generation_size_in_bytes(constraints.initial_young_generation_size_in_bytes);
    }
    params.constraints = rc;
  }

  Isolate* iso = Isolate::New(params);
  Locker locker(iso);
  Isolate::Scope isolate_scope(iso);
  HandleScope handle_scope(iso);

  iso->SetCaptureStackTraceForUncaughtExceptions(true);

  // Try to catch the OOM condition and stop execution before killing the process
  iso->AddNearHeapLimitCallback(NearMemoryLimitCallback, iso);

  // Create a Context for internal use
  m_ctx* ctx = new m_ctx;
  ctx->ptr.Reset(iso, Context::New(iso));
  ctx->iso = iso;
  iso->SetData(0, ctx);

  return iso;
}

void IsolatePerformMicrotaskCheckpoint(IsolatePtr iso) {
  ISOLATE_SCOPE(iso)
  iso->PerformMicrotaskCheckpoint();
}

void IsolateDispose(IsolatePtr iso) {
  if (iso == nullptr) {
    return;
  }
  auto ctx = static_cast<m_ctx*>(iso->GetData(0));
  ContextFree(ctx);

  iso->Dispose();
}

void IsolateTerminateExecution(IsolatePtr iso) {
  iso->TerminateExecution();
}

int IsolateIsExecutionTerminating(IsolatePtr iso) {
  return iso->IsExecutionTerminating();
}

IsolateHStatistics IsolationGetHeapStatistics(IsolatePtr iso) {
  if (iso == nullptr) {
    return IsolateHStatistics{0};
  }
  v8::HeapStatistics hs;
  iso->GetHeapStatistics(&hs);

  return IsolateHStatistics{hs.total_heap_size(),
                            hs.total_heap_size_executable(),
                            hs.total_physical_size(),
                            hs.total_available_size(),
                            hs.used_heap_size(),
                            hs.heap_size_limit(),
                            hs.malloced_memory(),
                            hs.external_memory(),
                            hs.peak_malloced_memory(),
                            hs.number_of_native_contexts(),
                            hs.number_of_detached_contexts()};
}
}
