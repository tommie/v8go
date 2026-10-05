#include "deps/include/v8-context.h"
#include "deps/include/v8-initialization.h"
#include "deps/include/v8-locker.h"
#include "deps/include/v8-platform.h"
#include "deps/include/v8-profiler.h"
#include "deps/include/v8-promise.h"

#include "_cgo_export.h"
#include "context.h"
#include "isolate-macros.h"
#include "isolate.h"
#include "libplatform/libplatform.h"

using namespace v8;

auto default_platform = platform::NewDefaultPlatform();
ArrayBuffer::Allocator* default_allocator;

// Forwards heap snapshot chunks to a Go io.Writer.
class GoOutputStream : public OutputStream {
 public:
  // writerRef is a cgo.Handle used by goWriteHeapSnapshotChunk.
  explicit GoOutputStream(uintptr_t writerRef) : writerRef_(writerRef) {}

  int GetChunkSize() override { return 64 * 1024; }

  void EndOfStream() override {}

  WriteResult WriteAsciiChunk(char* data, int size) override {
    return goWriteHeapSnapshotChunk(writerRef_, data, size) ? kContinue
                                                            : kAbort;
  }

 private:
  uintptr_t writerRef_;
};

extern "C" {

/********** Isolate **********/

// Per-isolate state, in data slot 1. Slot 0 holds the internal context.
struct IsolateState {
  // Set when NearMemoryLimitCallback terminates execution, and cleared
  // when the termination is reported.
  bool heap_limit_reached = false;

  // Whether errors include a serialized exception message.
  bool exception_messages = false;
};

#define ISOLATE_STATE_SLOT 1

void Init() {
#ifdef _WIN32
  V8::InitializeExternalStartupData(".");
#endif
  V8::InitializePlatform(default_platform.get());
  V8::Initialize();

  default_allocator = ArrayBuffer::Allocator::NewDefaultAllocator();
  return;
}

size_t NearMemoryLimitCallback(void* data,
                               size_t current_heap_limit,
                               size_t initial_heap_limit) {
  auto iso = static_cast<Isolate*>(data);
  auto state = static_cast<IsolateState*>(iso->GetData(ISOLATE_STATE_SLOT));
  state->heap_limit_reached = true;
  iso->TerminateExecution();

  // if we return the initial heap limit, the VM will crash, so here we give it
  // room to exit gracefully
  return current_heap_limit * 2;
}

IsolatePtr NewIsolate(IsolateConstraintsPtr constraints) {
  Isolate::CreateParams params;
  params.array_buffer_allocator = default_allocator;

  if (constraints != nullptr) {
    ResourceConstraints rc;
    rc.ConfigureDefaultsFromHeapSize(constraints->initial_heap_size_in_bytes,
                                     constraints->maximum_heap_size_in_bytes);
    params.constraints = rc;
  }

  Isolate* iso = Isolate::New(params);
  Locker locker(iso);
  Isolate::Scope isolate_scope(iso);
  HandleScope handle_scope(iso);

  iso->SetCaptureStackTraceForUncaughtExceptions(true);

  // Try to catch the OOM condition and stop execution before killing the
  // process
  iso->SetData(ISOLATE_STATE_SLOT, new IsolateState);
  iso->AddNearHeapLimitCallback(NearMemoryLimitCallback, iso);
  // The callback raises the heap limit, so the isolate can be reused
  // after the termination. Without this, the raised limit is kept, and
  // raised again on every termination.
  iso->AutomaticallyRestoreInitialHeapLimit(0.5);

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
  auto state = static_cast<IsolateState*>(iso->GetData(ISOLATE_STATE_SLOT));

  iso->Dispose();
  delete state;
}

void IsolateSetExceptionMessages(IsolatePtr iso, int enabled) {
  auto state = static_cast<IsolateState*>(iso->GetData(ISOLATE_STATE_SLOT));
  state->exception_messages = enabled;
}

int IsolateExceptionMessages(IsolatePtr iso) {
  auto state = static_cast<IsolateState*>(iso->GetData(ISOLATE_STATE_SLOT));
  return state->exception_messages;
}

int IsolateTakeHeapLimitReached(IsolatePtr iso) {
  auto state = static_cast<IsolateState*>(iso->GetData(ISOLATE_STATE_SLOT));
  if (!state->heap_limit_reached) {
    return 0;
  }
  state->heap_limit_reached = false;
  return 1;
}

void IsolateTerminateExecution(IsolatePtr iso) {
  iso->TerminateExecution();
}

int IsolateIsExecutionTerminating(IsolatePtr iso) {
  return iso->IsExecutionTerminating();
}

void IsolateLowMemoryNotification(IsolatePtr iso) {
  ISOLATE_SCOPE(iso);
  iso->LowMemoryNotification();
}

void IsolateWriteHeapSnapshot(IsolatePtr iso, uintptr_t writerRef) {
  ISOLATE_SCOPE(iso);

  // This runs a full garbage collection first.
  const HeapSnapshot* snapshot = iso->GetHeapProfiler()->TakeHeapSnapshot();
  GoOutputStream stream(writerRef);
  snapshot->Serialize(&stream, HeapSnapshot::kJSON);
  const_cast<HeapSnapshot*>(snapshot)->Delete();
}

// The Go PromiseRejectEvent constants mirror v8::PromiseRejectEvent.
static_assert(kPromiseRejectWithNoHandler == 0);
static_assert(kPromiseHandlerAddedAfterReject == 1);

static void PromiseRejectedCallback(PromiseRejectMessage message) {
  Isolate* iso = Isolate::GetCurrent();
  Local<Promise> promise = message.GetPromise();

  Local<Context> local_ctx;
  if (!promise->GetCreationContext(iso).ToLocal(&local_ctx)) {
    // Promises are always created in a context, but if there is none,
    // there is no Go Context to report it in either.
    return;
  }

  // The context is looked up through the Go registry, rather than
  // through a pointer stored in the V8 context, since the V8 context
  // can outlive a closed Context.
  int ctx_ref = local_ctx->GetEmbedderDataV2(ContextDataIndex::REF)
                    .As<Integer>()
                    ->Value();
  m_ctx* ctx = goContext(ctx_ref);
  if (ctx == nullptr) {
    return;
  }

  // The value is empty for kPromiseHandlerAddedAfterReject.
  Local<Value> value = message.GetValue();
  goPromiseRejectedCallback(
      ctx_ref, message.GetEvent(), track_value(ctx, promise),
      value.IsEmpty() ? nullptr : track_value(ctx, value));
}

void IsolateSetPromiseRejectedCallback(IsolatePtr iso, bool enable) {
  ISOLATE_SCOPE(iso);
  iso->SetPromiseRejectCallback(enable ? PromiseRejectedCallback : nullptr);
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
