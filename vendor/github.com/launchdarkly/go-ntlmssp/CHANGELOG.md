# Change log

All notable changes to the project will be documented in this file. This project adheres to [Semantic Versioning](http://semver.org).

## [1.0.2](https://github.com/launchdarkly/go-ntlmssp/compare/v1.0.1...v1.0.2) (2025-03-27)


### Bug Fixes

* Bump golang.org/x/crypto for CVE-2025-22869 ([#4](https://github.com/launchdarkly/go-ntlmssp/issues/4)) ([6d0a61e](https://github.com/launchdarkly/go-ntlmssp/commit/6d0a61e0eddc91513ad6c927159eda85397e9bca))

## [1.0.1] - 2020-06-11
### Added:
- Added `go.mod` file so this package can now be consumed as a module. Since the major version is still `1` and it does not have any module-only dependencies, it can still be used by non-module projects as well.
- Added CircleCI build.

### Fixed:
- Fixed a broken unit test.

## [1.0.0] - 2019-05-31
Initial release of this fork. The only change from the upstream code is https://github.com/launchdarkly/go-ntlmssp/pull/1.
