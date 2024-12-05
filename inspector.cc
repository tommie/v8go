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

void InspectorClient::consoleAPIMessage(int contextGroupId,
                                        v8::Isolate::MessageErrorLevel level,
                                        const StringView& message,
                                        const StringView& url,
                                        unsigned lineNumber,
                                        unsigned columnNumber,
                                        V8StackTrace*) {
  StringViewData msg;
  msg.is8bit = message.is8Bit();
  msg.data = message.characters8();
  msg.length = message.length();
  goHandleConsoleAPIMessageCallback(_callbackRef, contextGroupId, level, msg);
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
