#include <stdlib.h>
#include <sstream>

#include "deps/include/v8-debug.h"
#include "deps/include/v8-exception.h"
#include "deps/include/v8-isolate.h"
#include "deps/include/v8-message.h"
#include "deps/include/v8-primitive.h"

#include "errors.h"
#include "isolate.h"
#include "utils.h"

using namespace v8;

// The serialization format is parsed by parseMessage in errors.go.
// Integers are 32 bits, little-endian. Strings are a length, followed by
// UTF-8 bytes.
static void AppendInt(std::string& buf, int32_t v) {
  // A single append, which a byte-by-byte push_back loop isn't optimized
  // into. The swap is a no-op on all our targets.
  uint32_t le = static_cast<uint32_t>(v);
#if __BYTE_ORDER__ == __ORDER_BIG_ENDIAN__
  le = __builtin_bswap32(le);
#endif
  buf.append(reinterpret_cast<const char*>(&le), sizeof(le));
}

static void AppendString(std::string& buf, Isolate* iso, Local<Value> v) {
  if (v.IsEmpty() || v->IsUndefined()) {
    AppendInt(buf, 0);
    return;
  }
  String::Utf8Value s(iso, v);
  AppendInt(buf, s.length());
  buf.append(*s, s.length());
}

static std::string SerializeMessage(Isolate* iso,
                                    Local<Context> ctx,
                                    Local<Message> msg) {
  std::string buf;
  // Avoids growing the buffer repeatedly. The header is 36 bytes of
  // lengths and integers, plus the text, script name and source line,
  // e.g. 60, 40 and 150 bytes. A larger message only costs reallocations.
  buf.reserve(256);
  AppendString(buf, iso, msg->Get());
  AppendString(buf, iso, msg->GetScriptResourceName());
  Local<String> source_line;
  AppendString(buf, iso,
               msg->GetSourceLine(ctx).ToLocal(&source_line)
                   ? source_line.As<Value>()
                   : Local<Value>());
  AppendInt(buf, msg->GetLineNumber(ctx).FromMaybe(0));
  AppendInt(buf, msg->GetStartPosition());
  AppendInt(buf, msg->GetEndPosition());
  AppendInt(buf, msg->GetStartColumn());
  AppendInt(buf, msg->GetEndColumn());
  AppendInt(buf, msg->GetWasmFunctionIndex());

  // Captured because of SetCaptureStackTraceForUncaughtExceptions.
  Local<StackTrace> trace = msg->GetStackTrace();
  int num_frames = trace.IsEmpty() ? 0 : trace->GetFrameCount();
  // The count, and per frame 20 bytes of lengths and integers, plus the
  // script and function names, e.g. 30 and 14 bytes.
  buf.reserve(buf.size() + 4 + num_frames * 64);
  AppendInt(buf, num_frames);
  for (int i = 0; i < num_frames; i++) {
    Local<StackFrame> frame = trace->GetFrame(iso, i);
    AppendString(buf, iso, frame->GetScriptName());
    AppendString(buf, iso, frame->GetFunctionName());
    AppendInt(buf, frame->GetLineNumber());
    AppendInt(buf, frame->GetColumn());
    AppendInt(buf, (frame->IsEval() ? 1 : 0) |
                       (frame->IsConstructor() ? 2 : 0) |
                       (frame->IsWasm() ? 4 : 0) |
                       (frame->IsUserJavaScript() ? 8 : 0));
  }
  return buf;
}

RtnError ExceptionError(TryCatch& try_catch, Isolate* iso, Local<Context> ctx) {
  HandleScope handle_scope(iso);

  RtnError rtn = {};

  if (try_catch.HasTerminated()) {
    if (IsolateTakeHeapLimitReached(iso)) {
      // The script has unwound, so its garbage can be collected. This
      // makes AutomaticallyRestoreInitialHeapLimit restore the limit,
      // which it only does when the heap is small enough. Otherwise, the
      // next script reaching the limit would raise it further.
      iso->LowMemoryNotification();
      rtn.msg = CopyString("ExecutionTerminated: heap limit reached");
      rtn.heap_limit_reached = 1;
    } else {
      rtn.msg = CopyString(
          "ExecutionTerminated: script execution has been terminated");
    }
    return rtn;
  }

  String::Utf8Value exception(iso, try_catch.Exception());
  rtn.msg = CopyString(exception);

  Local<Message> msg = try_catch.Message();
  if (!msg.IsEmpty()) {
    String::Utf8Value origin(iso, msg->GetScriptOrigin().ResourceName());
    std::ostringstream sb;
    sb << *origin;
    Maybe<int> line = try_catch.Message()->GetLineNumber(ctx);
    if (line.IsJust()) {
      sb << ":" << line.ToChecked();
    }
    Maybe<int> start = try_catch.Message()->GetStartColumn(ctx);
    if (start.IsJust()) {
      sb << ":"
         << start.ToChecked() + 1;  // + 1 to match output from stack trace
    }
    rtn.location = CopyString(sb.str());

    if (IsolateExceptionMessages(iso)) {
      std::string buf = SerializeMessage(iso, ctx, msg);
      rtn.message = CopyString(buf);
      rtn.message_length = buf.size();
    }
  }

  Local<Value> mstack;
  if (try_catch.StackTrace(ctx).ToLocal(&mstack)) {
    String::Utf8Value stack(iso, mstack);
    rtn.stack = CopyString(stack);
  }

  return rtn;
}

void ErrorRelease(RtnError err) {
  free(err.msg);
  free(err.location);
  free(err.stack);
  free(err.message);
}
