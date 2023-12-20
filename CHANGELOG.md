# Changelog

## [v0.7.0](https://github.com/2manymws/rc/compare/v0.6.1...v0.7.0) - 2023-12-20
### Breaking Changes 🛠
- Fix Cacher.Store signature by @k1LoW in https://github.com/2manymws/rc/pull/36
- Support extended rules like proxy_cache_valid of NGINX by @k1LoW in https://github.com/2manymws/rc/pull/38
- Does not store responses with Set-Cookie headers by default, similar to NGINX by @k1LoW in https://github.com/2manymws/rc/pull/39

## [v0.6.1](https://github.com/2manymws/rc/compare/v0.6.0...v0.6.1) - 2023-12-18
### Other Changes
- Add license for RFC 9111 by @k1LoW in https://github.com/2manymws/rc/pull/34

## [v0.6.0](https://github.com/2manymws/rc/compare/v0.5.2...v0.6.0) - 2023-12-18
### New Features 🎉
- Support logger (log/slog.Logger) by @k1LoW in https://github.com/2manymws/rc/pull/33

## [v0.5.2](https://github.com/2manymws/rc/compare/v0.5.1...v0.5.2) - 2023-12-15

## [v0.5.1](https://github.com/2manymws/rc/compare/v0.5.0...v0.5.1) - 2023-12-15
### Fix bug 🐛
- For SNI compatibility, also compare req.Host by @k1LoW in https://github.com/2manymws/rc/pull/29

## [v0.5.0](https://github.com/k1LoW/rc/compare/v0.4.1...v0.5.0) - 2023-12-15
### Breaking Changes 🛠
- Support Age header by @k1LoW in https://github.com/k1LoW/rc/pull/28
### New Features 🎉
- Add options by @k1LoW in https://github.com/k1LoW/rc/pull/24
### Fix bug 🐛
- Fix no-cache decision logic by @k1LoW in https://github.com/k1LoW/rc/pull/22
### Other Changes
- Fix URL of comment by @k1LoW in https://github.com/k1LoW/rc/pull/25
- Although RFC 9111 does not explicitly define this, in general cacheable methods are GET and HEAD, so the default should be GET and HEAD. by @k1LoW in https://github.com/k1LoW/rc/pull/26
- Although not explicitly specified in RFC 9111, allow GET, HEAD, OPTIONS, and TRACE methods, which are specified as safe methods in RFC 9110. by @k1LoW in https://github.com/k1LoW/rc/pull/27

## [v0.4.1](https://github.com/k1LoW/rc/compare/v0.4.0...v0.4.1) - 2023-12-14
### Other Changes
- Rename ErrDoNotUseCache to ErrShouldNotUseCache by @k1LoW in https://github.com/k1LoW/rc/pull/19
- Rename do to originRequester by @k1LoW in https://github.com/k1LoW/rc/pull/20

## [v0.4.0](https://github.com/k1LoW/rc/compare/v0.3.1...v0.4.0) - 2023-12-14
### Breaking Changes 🛠
- Change the interface to be ready for RFC 9111 by @k1LoW in https://github.com/k1LoW/rc/pull/18
### Other Changes
- Fix benchmark condition by @k1LoW in https://github.com/k1LoW/rc/pull/13
- Use rcutil v0.5.0 by @k1LoW in https://github.com/k1LoW/rc/pull/14
- Add gostyle-action by @k1LoW in https://github.com/k1LoW/rc/pull/15
- Run 2 benchmarks on same runner by @k1LoW in https://github.com/k1LoW/rc/pull/16

## [v0.3.1](https://github.com/k1LoW/rc/compare/v0.3.0...v0.3.1) - 2023-09-05
### Breaking Changes 🛠
- Change error by @k1LoW in https://github.com/k1LoW/rc/pull/10

## [v0.3.0](https://github.com/k1LoW/rc/compare/v0.2.0...v0.3.0) - 2023-09-05
### Breaking Changes 🛠
- Separate `rcutil` package by @k1LoW in https://github.com/k1LoW/rc/pull/8

## [v0.2.0](https://github.com/k1LoW/rc/compare/v0.1.1...v0.2.0) - 2023-09-05
### Breaking Changes 🛠
- Create `rcutil` package by @k1LoW in https://github.com/k1LoW/rc/pull/7

## [v0.1.1](https://github.com/k1LoW/rc/compare/v0.1.0...v0.1.1) - 2023-09-04
### Other Changes
- Add Usage by @k1LoW in https://github.com/k1LoW/rc/pull/5

## [v0.1.0](https://github.com/k1LoW/rc/commits/v0.1.0) - 2023-09-04
### Other Changes
- Change to use local file caching in `testutil.*Cacher` for benchmarking by @k1LoW in https://github.com/k1LoW/rc/pull/2
- Add benchmark by @k1LoW in https://github.com/k1LoW/rc/pull/3
