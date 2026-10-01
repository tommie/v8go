#include <stdlib.h>
#include <sstream>

#include "deps/include/v8-exception.h"
#include "deps/include/v8-isolate.h"
#include "deps/include/v8-message.h"
#include "deps/include/v8-primitive.h"

#include "errors.h"
#include "isolate.h"
#include "utils.h"

using namespace v8;

RtnError ExceptionError(TryCatch& try_catch, Isolate* iso, Local<Context> ctx) {
  HandleScope handle_scope(iso);

  RtnError rtn = {nullptr, nullptr, nullptr, 0};

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
}
