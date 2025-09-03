#ifndef V8GO_RESOURCE_CONSTRAINTS_H
#define V8GO_RESOURCE_CONSTRAINTS_H

#ifdef __cplusplus

namespace v8 {
class ResourceConstraints;
}
typedef v8::ResourceConstraints v8ResourceConstraints;

extern "C" {
#else
typedef struct v8ResourceConstraints v8ResourceConstraints;
#endif

typedef v8ResourceConstraints* ResourceConstraintsPtr;

extern ResourceConstraintsPtr NewResourceConstraints();
extern void ResourceConstraintsDelete(ResourceConstraintsPtr ptr);

extern void ResourceConstraintsConfigureDefaultsFromHeapSize(
    ResourceConstraintsPtr ptr,
    size_t initial_heap_size_in_bytes,
    size_t maximum_heap_size_in_bytes);

extern void ResourceConstraintsConfigureDefaults(
    ResourceConstraintsPtr ptr,
    uint64_t physical_memory,
    uint64_t virtual_memory_limit);

extern void ResourceConstraintsSetStackLimit(ResourceConstraintsPtr ptr, uintptr_t value);
extern uintptr_t ResourceConstraintsStackLimit(ResourceConstraintsPtr ptr);

extern void ResourceConstraintsSetCodeRangeSizeInBytes(ResourceConstraintsPtr ptr, size_t limit);
extern size_t ResourceConstraintsCodeRangeSizeInBytes(ResourceConstraintsPtr ptr);

extern void ResourceConstraintsSetMaxOldGenerationSizeInBytes(ResourceConstraintsPtr ptr, size_t limit);
extern size_t ResourceConstraintsMaxOldGenerationSizeInBytes(ResourceConstraintsPtr ptr);

extern void ResourceConstraintsSetMaxYoungGenerationSizeInBytes(ResourceConstraintsPtr ptr, size_t limit);
extern size_t ResourceConstraintsMaxYoungGenerationSizeInBytes(ResourceConstraintsPtr ptr);

extern void ResourceConstraintsSetInitialOldGenerationSizeInBytes(ResourceConstraintsPtr ptr, size_t initial_size);
extern size_t ResourceConstraintsInitialOldGenerationSizeInBytes(ResourceConstraintsPtr ptr);

extern void ResourceConstraintsSetInitialYoungGenerationSizeInBytes(ResourceConstraintsPtr ptr, size_t initial_size);
extern size_t ResourceConstraintsInitialYoungGenerationSizeInBytes(ResourceConstraintsPtr ptr);

#ifdef __cplusplus
}  // extern "C"
#endif
#endif
