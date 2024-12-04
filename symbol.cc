#include "symbol.h"
#include "context.h"
#include "isolate-macros.h"
#include "utils.h"
#include "value-macros.h"

using namespace v8;

ValuePtr BuiltinSymbol(IsolatePtr iso, SymbolIndex idx) {
  ISOLATE_SCOPE(iso);
  INTERNAL_CONTEXT(iso);
  Local<Symbol> sym;
  switch (idx) {
    case SYMBOL_ASYNC_ITERATOR:
      sym = Symbol::GetAsyncIterator(iso);
      break;
    case SYMBOL_HAS_INSTANCE:
      sym = Symbol::GetHasInstance(iso);
      break;
    case SYMBOL_IS_CONCAT_SPREADABLE:
      sym = Symbol::GetIsConcatSpreadable(iso);
      break;
    case SYMBOL_ITERATOR:
      sym = Symbol::GetIterator(iso);
      break;
    case SYMBOL_MATCH:
      sym = Symbol::GetMatch(iso);
      break;
    case SYMBOL_REPLACE:
      sym = Symbol::GetReplace(iso);
      break;
    case SYMBOL_SEARCH:
      sym = Symbol::GetSearch(iso);
      break;
    case SYMBOL_SPLIT:
      sym = Symbol::GetSplit(iso);
      break;
    case SYMBOL_TO_PRIMITIVE:
      sym = Symbol::GetToPrimitive(iso);
      break;
    case SYMBOL_TO_STRING_TAG:
      sym = Symbol::GetToStringTag(iso);
      break;
    case SYMBOL_UNSCOPABLES:
      sym = Symbol::GetUnscopables(iso);
      break;
    default:
      return nullptr;
  }
  m_value* val = new m_value;
  val->id = 0;
  val->iso = iso;
  val->ctx = ctx;
  val->ptr = Global<Value>(iso, sym);
  return tracked_value(ctx, val);
}

const char* SymbolDescription(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  Local<Symbol> sym = value.As<Symbol>();
  Local<Value> descr = sym->Description(iso);
  String::Utf8Value utf8(iso, descr);
  return CopyString(utf8);
}
