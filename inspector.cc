#include "deps/include/v8-inspector.h"

#include "_cgo_export.h"
#include "context-macros.h"
#include "inspector.h"

using namespace v8;
using namespace v8_inspector;

class InspectorClient : public V8InspectorClient {
  int _callbackRef;

 public:
  InspectorClient(int callbackRef) { _callbackRef = callbackRef; }
  void consoleAPIMessage(int contextGroupId,
                         v8::Isolate::MessageErrorLevel level,
                         const StringView& message,
                         const StringView& url,
                         unsigned lineNumber,
                         unsigned columnNumber,
                         V8StackTrace*) override;
};

StringViewData ConvertStringView(const StringView& view) {
  StringViewData msg;
  msg.is8bit = view.is8Bit();
  // The ? isn't necessary, the two functions return the sama pointer. But that
  // has been considered an implementation detail that may change.
  msg.data =
      view.is8Bit() ? (void*)view.characters8() : (void*)view.characters16();
  msg.length = view.length();
  return msg;
}

void InspectorClient::consoleAPIMessage(int contextGroupId,
                                        v8::Isolate::MessageErrorLevel level,
                                        const StringView& message,
                                        const StringView& url,
                                        unsigned lineNumber,
                                        unsigned columnNumber,
                                        V8StackTrace*) {
  goHandleConsoleAPIMessageCallback(
      _callbackRef, contextGroupId, level, ConvertStringView(message),
      ConvertStringView(url), lineNumber, columnNumber);
}

extern "C" {

InspectorPtr CreateInspector(v8Isolate* iso, InspectorClientPtr client) {
  InspectorPtr inspector = V8Inspector::create(iso, client).release();
  return inspector;
}

void DeleteInspector(InspectorPtr inspector) {
  delete inspector;
}

/********** InspectorClient **********/

InspectorClientPtr NewInspectorClient(int callbackRef) {
  return new InspectorClient(callbackRef);
}

void InspectorContextCreated(InspectorPtr inspector, ContextPtr context) {
  LOCAL_CONTEXT(context);
  int groupId = 1;
  StringView name = StringView((const uint8_t*)"Test", 4);
  V8ContextInfo info = V8ContextInfo(local_ctx, groupId, name);
  inspector->contextCreated(info);
}

void InspectorContextDestroyed(InspectorPtr inspector, ContextPtr context) {
  LOCAL_CONTEXT(context);
  inspector->contextDestroyed(local_ctx);
}

void DeleteInspectorClient(InspectorClientPtr client) {
  delete client;
}
}
