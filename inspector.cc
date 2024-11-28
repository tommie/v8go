#include "deps/include/v8-inspector.h"

#include "inspector.h"

using namespace v8;
using namespace v8_inspector;

class InspectorClient : public V8InspectorClient {};

extern "C" {

InspectorPtr CreateInspector(v8Isolate* iso, InspectorClientPtr client) {
  InspectorPtr inspector = V8Inspector::create(iso, client).release();
  return inspector;
}

void DeleteInspector(InspectorPtr inspector) {
  delete inspector;
}

/********** InspectorClient **********/

InspectorClientPtr NewInspectorClient() {
  return new InspectorClient();
}

void DeleteInspectorClient(InspectorClientPtr client) {
  delete client;
}
}
