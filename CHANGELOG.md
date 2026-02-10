# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- Remove obsolete `20250331-stringhasher-arm64.patch`; the fix is now upstream in V8 and patch application fails as already-applied.

## [v0.34.0] - 2025-10-07

### Added

- Add support for ObjectTemplate.MarkAsUndetectable.
- Add `Value.StrictEquals` providing strict equality checks in Go code.
- Add resource constraint options to `NewIsolate` in [#111](https://github.com/tommie/v8go/pull/111).
- Add support for `Value.TypeOf()` corresponding to JavaScript `typeof` operator in [#104](https://github.com/tommie/v8go/pull/104).

### Changed

## [v0.33.0] - 2025-05-15

### Added
- Add support for `FunctionTemplate.Inherit` to set up prototype inheritance.

### Changed
- Auto-bumped V8 to 13.6.233.10.

## [v0.32.0] - 2025-04-30

### Changed
- Auto-bumped V8 to 13.6.233.8.

## [v0.31.0] - 2025-04-02

### Changed
- Auto-bumped V8 to 13.5.212.10.

## [v0.30.0] - 2025-02-26

### Changed
- Auto-bumped V8 to 13.3.415.23.

## [v0.29.0] - 2025-02-19

### Added

- Add support to setup an `Inspector`, and `InspectorClient` to receive output from `console` messages in JS code.
- Add `FunctionTemplate.InstanceTemplate` to add own properties to instances when the function is used as a constructor.
- Add `FunctionTemplate.PrototypeTemplates` to add own properties to the prototype of an instance when the function is used as a constructor.

### Changed
- Auto-bumped V8 to 13.3.415.22.

## [v0.28.0] - 2025-01-08

### Changed
- Auto-bumped V8 to 13.1.201.22.

## [v0.27.0] - 2024-12-19

### Changed
- Auto-bumped V8 to 13.1.201.19.

## [v0.26.0] - 2024-12-11

### Changed
- Auto-bumped V8 to 13.1.201.16.

## [v0.25.0] - 2024-12-04

### Changed
- Auto-bumped V8 to 13.1.201.15.

## [v0.24.0] - 2024-11-20

### Changed
- Auto-bumped V8 to 13.1.201.9.

## [v0.23.0] - 2024-11-13

### Changed
- Auto-bumped V8 to 13.1.201.8.

## [v0.22.0] - 2024-10-16

### Changed
- Auto-bumped V8 to 13.0.245.16.

## [v0.21.0] - 2024-10-09

### Changed
- Auto-bumped V8 to 12.9.202.27.

## [v0.20.0] - 2024-10-02

### Changed
- Auto-bumped V8 to 12.9.202.24.

## [v0.19.0] - 2024-09-25

### Changed
- Auto-bumped V8 to 12.9.202.22.

## [v0.18.0] - 2024-09-18

### Changed
- Auto-bumped V8 to 12.9.202.18.

## [v0.17.0] - 2024-08-07

### Changed

## [v0.16.0] - 2024-08-07

### Changed
- Auto-bumped V8 to 12.7.224.18.

## [v0.15.0] - 2024-07-25

### Changed
- Auto-bumped V8 to 12.7.224.16.

## [v0.14.0] - 2024-07-17

### Changed
- Auto-bumped V8 to 12.6.228.28.

## [v0.13.1] - 2024-06-28

### Added
- Made scheduled V8 rebuilding and releases fully automatic.

### Changed

### Fixed
- The build broke because CopyablePersistentTraits was removed in V8 4683daaf774982b1aa0e46ae38f2bdf9c995c90d.
- Removed bad release 0.13.0.

## [v0.13.0] - 2024-06-28

### Changed
- Auto-bumped V8 to 12.6.228.21.

## [v0.12.0] - 2024-05-01

### Changed
- Auto-bumped V8 to c546005d65b58039ccaf3f81be3772eee45454f9.
- The build workflows commit directly, instead of creating PRs.

### Fixed
- Release tags need to include the patch number.
  Thank you for the report, @GraphR00t.
  (https://github.com/tommie/v8go/issues/62)

## [v0.11] - 2024-03-02

### Changed
- Bumped V8 to d5c51572dec5b2a385b57549fe195c319f1284e2.

## [v0.10.1] - 2023-12-30

### Changed
- Required Go version changed to 1.19 for CGo across modules.

### Fixed
- Split the built libraries into separate submodules to work around Go's 500 MB module size limit.
- The library modules still had the old x86_64 names.

## [v0.10.0] - 2023-12-29

### Changed
- Required Go version changed to 1.17 (needed for SharedArrayBuffer support)
- Forked from rogchap.com/v8go
- x86_64 is now referred to as amd64 in V8 builds, conforming to `GOARCH`

### Added
- Support for getting the underlying data (as a `[]byte`) from a SharedArrayBuffer
- Support for `Symbol`
- Support for `FunctionCallback` to return an error
- Support for exceptions as Go errors
- Support for Android on amd64 and arm64

### Fixed
- Upgrade to V8 12.0.267.10


## [v0.9.0] - 2023-03-30

### Fixed
- Upgrade to V8 11.1.277.13

## [v0.8.0] - 2023-01-19

### Added
- Added support for Value.release() and FunctionCallbackInfo.release(). This is useful when using v8go in a long-running context.

### Fixed
- Use string length to ensure null character-containing strings in Go/JS are not terminated early.
- Object.Set with an empty key string is now supported
- Upgrade to V8 10.9.194.9
- Upgrade V8 build OS to Ubuntu 22.04

## [v0.7.0] - 2021-12-09

### Added
- Support for calling constructors functions with NewInstance on Function
- Access "this" from function callback
- value.SameValue(otherValue) function to compare values for sameness
- Undefined, Null functions to get these constant values for the isolate
- Support for calling a method on an object.
- Support for calling `IsExecutionTerminating` on isolate to check if execution is still terminating.
- Support for setting and getting internal fields for template object instances
- Support for CPU profiling
- Add V8 build for Apple Silicon
- Add support for throwing an exception directly via the isolate's ThrowException function.
- Support for compiling a context-dependent UnboundScript which can be run in any context of the isolate it was compiled in.
- Support for creating a code cache from an UnboundScript which can be used to create an UnboundScript in other isolates
to run a pre-compiled script in new contexts.
- Included compile error location in `%+v` formatting of JSError
- Enable i18n support

### Changed
- Removed error return value from NewIsolate which never fails
- Removed error return value from NewContext which never fails
- Removed error return value from Context.Isolate() which never fails
- Removed error return value from NewObjectTemplate and NewFunctionTemplate. Panic if given a nil argument.
- Function Call accepts receiver as first argument. This **subtle breaking change** will compile old code but interpret the first argument as the receiver. Use `Undefined` to prepend an argument to fix old Call use.
- Removed Windows support until its build issues are addressed.
- Upgrade to V8 9.6.180.12

### Fixed
- Add some missing error propagation
- Fix crash from template finalizer releasing V8 data, let it be disposed with the isolate
- Fix crash by keeping alive the template while its C++ pointer is still being used
- Fix crash from accessing function template callbacks outside of `RunScript`, such as in `JSONStringify`

## [v0.6.0] - 2021-05-11

### Added
- Promise resolver and promise result
- Convert a Value to a Function and invoke it. Thanks to [@robfig](https://github.com/robfig)
- Windows static binary. Thanks to [@cleiner](https://github.com/cleiner)
- Setting/unsetting of V8 feature flags
- Register promise callbacks in Go. Thanks to [@robfig](https://github.com/robfig)
- Get Function from a template for a given context. Thanks to [@robfig](https://github.com/robfig)

### Changed
- Upgrade to V8 9.0.257.18

### Fixed
- Go GC attempting to free C memory (via finalizer) of values after an Isolate is disposed causes a panic

## [v0.5.1] - 2021-02-19

### Fixed
- Memory being held by Values after the associated Context is closed

## [v0.5.0] - 2021-02-08

### Added
- Support for the BigInt value to the big.Int Go type
- Create Object Templates with primitive values, including other Object Templates
- Configure Object Template as the global object of any new Context
- Function Templates with callbacks to Go
- Value to Object type, including Get/Set/Has/Delete methods
- Get Global Object from the Context
- Convert an Object Template to an instance of an Object

### Changed
- NewContext() API has been improved to handle optional global object, as well as optional Isolate
- Package error messages are now prefixed with `v8go` rather than the struct name
- Deprecated `iso.Close()` in favor of `iso.Dispose()` to keep consistancy with the C++ API
- Upgraded V8 to 8.8.278.14
- Licence BSD 3-Clause (same as V8 and Go)

## [v0.4.0] - 2021-01-14

### Added
- Value methods for checking value kind (is string, number, array etc)
- C formatting via `clang-format` to aid future development
- Support of vendoring with `go mod vendor`
- Value methods to convert to primitive data types

### Changed
- Use g++ (default for cgo) for linux builds of the static v8 lib

## [v0.3.0] - 2020-12-18

### Added
- Support for Windows via [MSYS2](https://www.msys2.org/). Thanks to [@neptoess](https://github.com/neptoess)

### Changed
- Upgraded V8 to 8.7.220.31

## [v0.2.0] - 2020-01-25

### Added
- Manually dispose of the isolate when required
- Monitor isolate heap statistics. Thanks to [@mehrdadrad](https://github.com/mehrdadrad)

### Changed
- Upgrade V8 to 8.0.426.15

## [v0.1.0] - 2019-09-22

### Changed
- Upgrade V8 to 7.7.299.9

## [v0.0.1] - 2019-09-2020

### Added
- Create V8 Isolate
- Create Contexts
- Run JavaScript scripts
- Get Values back from JavaScript in Go
- Get detailed JavaScript errors in Go, including stack traces
- Terminate long running scripts from any Goroutine
