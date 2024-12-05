#include "deps/include/v8-inspector.h"

#include "inspector.h"

using namespace v8;
using namespace v8_inspector;

class InspectorClient : public V8InspectorClient {
  int _callbackRef;

 public:
  InspectorClient(int callbackRef) { _callbackRef = callbackRef; }
};

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

void DeleteInspectorClient(InspectorClientPtr client) {
  delete client;
}
}
