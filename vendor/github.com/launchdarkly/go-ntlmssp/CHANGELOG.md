# Change log

All notable changes to the project will be documented in this file. This project adheres to [Semantic Versioning](http://semver.org).

## [1.0.5](https://github.com/launchdarkly/go-ntlmssp/compare/v1.0.4...v1.0.5) (2026-08-24)


### Bug Fixes

* Bump minimum Go version to 1.26 ([#16](https://github.com/launchdarkly/go-ntlmssp/issues/16)) ([116bd40](https://github.com/launchdarkly/go-ntlmssp/commit/116bd403a0b037f53619f1074c9b4f6eb3b58761))

## [1.0.4](https://github.com/launchdarkly/go-ntlmssp/compare/v1.0.3...v1.0.4) (2026-08-11)


### Bug Fixes

* Remove golang.org/x/crypto dependency by vendoring md4 ([#12](https://github.com/launchdarkly/go-ntlmssp/issues/12)) ([7fc6c8a](https://github.com/launchdarkly/go-ntlmssp/commit/7fc6c8ad9dc2d536ee76128c922b1c9565cae79e))

## [1.0.3](https://github.com/launchdarkly/go-ntlmssp/compare/v1.0.2...v1.0.3) (2025-11-21)


### Bug Fixes

* Bump golang.org/x/crypto from v0.36.0 -&gt; v0.45.0 ([#9](https://github.com/launchdarkly/go-ntlmssp/issues/9)) ([46647fb](https://github.com/launchdarkly/go-ntlmssp/commit/46647fbbfcaff202ed6ba2cb089e1f146d495352))
* Bump minimum go to 1.24 ([#8](https://github.com/launchdarkly/go-ntlmssp/issues/8)) ([0af9fa0](https://github.com/launchdarkly/go-ntlmssp/commit/0af9fa0e02afd7e452834816f517256037ff1a3c))

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
