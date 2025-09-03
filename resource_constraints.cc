#include "deps/include/v8-isolate.h"
#include "resource_constraints.h"

using namespace v8;

extern "C" {

ResourceConstraintsPtr NewResourceConstraints() {
  ResourceConstraints* rc = new ResourceConstraints();
  return rc;
}

void ResourceConstraintsDelete(ResourceConstraintsPtr ptr) {
  if (ptr == nullptr) {
    return;
  }
  delete ptr;
}

void ResourceConstraintsConfigureDefaultsFromHeapSize(
    ResourceConstraintsPtr ptr,
    size_t initial_heap_size_in_bytes,
    size_t maximum_heap_size_in_bytes) {
  if (ptr == nullptr) {
    return;
  }
  ptr->ConfigureDefaultsFromHeapSize(initial_heap_size_in_bytes, maximum_heap_size_in_bytes);
}

void ResourceConstraintsConfigureDefaults(
    ResourceConstraintsPtr ptr,
    uint64_t physical_memory,
    uint64_t virtual_memory_limit) {
  if (ptr == nullptr) {
    return;
  }
  ptr->ConfigureDefaults(physical_memory, virtual_memory_limit);
}

void ResourceConstraintsSetStackLimit(ResourceConstraintsPtr ptr, uintptr_t value) {
  if (ptr == nullptr) {
    return;
  }
  ptr->set_stack_limit(reinterpret_cast<uint32_t*>(value));
}

uintptr_t ResourceConstraintsStackLimit(ResourceConstraintsPtr ptr) {
  if (ptr == nullptr) {
    return 0;
  }
  return reinterpret_cast<uintptr_t>(ptr->stack_limit());
}

void ResourceConstraintsSetCodeRangeSizeInBytes(ResourceConstraintsPtr ptr, size_t limit) {
  if (ptr == nullptr) {
    return;
  }
  ptr->set_code_range_size_in_bytes(limit);
}

size_t ResourceConstraintsCodeRangeSizeInBytes(ResourceConstraintsPtr ptr) {
  if (ptr == nullptr) {
    return 0;
  }
  return ptr->code_range_size_in_bytes();
}

void ResourceConstraintsSetMaxOldGenerationSizeInBytes(ResourceConstraintsPtr ptr, size_t limit) {
  if (ptr == nullptr) {
    return;
  }
  ptr->set_max_old_generation_size_in_bytes(limit);
}

size_t ResourceConstraintsMaxOldGenerationSizeInBytes(ResourceConstraintsPtr ptr) {
  if (ptr == nullptr) {
    return 0;
  }
  return ptr->max_old_generation_size_in_bytes();
}

void ResourceConstraintsSetMaxYoungGenerationSizeInBytes(ResourceConstraintsPtr ptr, size_t limit) {
  if (ptr == nullptr) {
    return;
  }
  ptr->set_max_young_generation_size_in_bytes(limit);
}

size_t ResourceConstraintsMaxYoungGenerationSizeInBytes(ResourceConstraintsPtr ptr) {
  if (ptr == nullptr) {
    return 0;
  }
  return ptr->max_young_generation_size_in_bytes();
}

void ResourceConstraintsSetInitialOldGenerationSizeInBytes(ResourceConstraintsPtr ptr, size_t initial_size) {
  if (ptr == nullptr) {
    return;
  }
  ptr->set_initial_old_generation_size_in_bytes(initial_size);
}

size_t ResourceConstraintsInitialOldGenerationSizeInBytes(ResourceConstraintsPtr ptr) {
  if (ptr == nullptr) {
    return 0;
  }
  return ptr->initial_old_generation_size_in_bytes();
}

void ResourceConstraintsSetInitialYoungGenerationSizeInBytes(ResourceConstraintsPtr ptr, size_t initial_size) {
  if (ptr == nullptr) {
    return;
  }
  ptr->set_initial_young_generation_size_in_bytes(initial_size);
}

size_t ResourceConstraintsInitialYoungGenerationSizeInBytes(ResourceConstraintsPtr ptr) {
  if (ptr == nullptr) {
    return 0;
  }
  return ptr->initial_young_generation_size_in_bytes();
}

}  // extern "C"
