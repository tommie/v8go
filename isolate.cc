#include "deps/include/v8-context.h"
#include "deps/include/v8-initialization.h"
#include "deps/include/v8-locker.h"
#include "deps/include/v8-platform.h"
#include "deps/include/v8-profiler.h"

#include "context.h"
#include "isolate.h"
#include "libplatform/libplatform.h"

using namespace v8;

auto default_platform = platform::NewDefaultPlatform();
ArrayBuffer::Allocator* default_allocator;

class FileOutputStream : public v8::OutputStream {
  public:
    explicit FileOutputStream(FILE* stream) : stream_(stream) {}

    virtual int GetChunkSize() {
      return 65536;  // big chunks == faster
    }

    virtual void EndOfStream() {}

    virtual WriteResult WriteAsciiChunk(char* data, int size) {
      const size_t len = static_cast<size_t>(size);
      size_t off = 0;

      while (off < len && !feof(stream_) && !ferror(stream_))
        off += fwrite(data + off, 1, len - off, stream_);

      return off == len ? kContinue : kAbort;
    }

  private:
    FILE* stream_;
};

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

IsolatePtr NewIsolate() {
  Isolate::CreateParams params;
  params.array_buffer_allocator = default_allocator;
  Isolate* iso = Isolate::New(params);
  Locker locker(iso);
  Isolate::Scope isolate_scope(iso);
  HandleScope handle_scope(iso);

  iso->SetCaptureStackTraceForUncaughtExceptions(true);

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

void IsolateFullGC(IsolatePtr iso) {
  if (iso == nullptr) {
    return;
  }
  iso->LowMemoryNotification();
}

int IsolateWriteSnapshot(IsolatePtr iso, const char* filename, int bForceGC) {
  if (iso == nullptr) {
    return 1;
  }

  FILE* fp = fopen(filename, "w");
  if (fp == NULL) {
    return errno;
  }

  if (bForceGC) {
    // force garbage collection
    iso->LowMemoryNotification();
  }

  Locker locker(iso);
  Isolate::Scope isolate_scope(iso);
  // Create a stack-allocated handle scope.
  HandleScope handle_scope(iso);

  const v8::HeapSnapshot* snap = iso->GetHeapProfiler()->TakeHeapSnapshot();
  FileOutputStream stream(fp);
  snap->Serialize(&stream, v8::HeapSnapshot::kJSON);

  int err = 0;
  if (fclose(fp)) {
    err = errno;
  }
  const_cast<v8::HeapSnapshot*>(snap)->Delete();
  return err;
}
}
