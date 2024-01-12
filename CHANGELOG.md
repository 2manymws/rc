# Changelog

## [v0.9.1](https://github.com/2manymws/rc/compare/v0.9.0...v0.9.1) - 2024-01-12
### Breaking Changes 🛠
- No request body is used for cache handling by default. by @k1LoW in https://github.com/2manymws/rc/pull/56

## [v0.9.0](https://github.com/2manymws/rc/compare/v0.8.2...v0.9.0) - 2024-01-12
### Breaking Changes 🛠
- If cache is not used, no Age header is given by @k1LoW in https://github.com/2manymws/rc/pull/54
### Fix bug 🐛
- Keep and use requests unaffected by the next middleware. by @k1LoW in https://github.com/2manymws/rc/pull/53

## [v0.8.2](https://github.com/2manymws/rc/compare/v0.8.1...v0.8.2) - 2024-01-11
### Fix bug 🐛
- Fix badkey log by @k1LoW in https://github.com/2manymws/rc/pull/52
### Other Changes
- Log more info by @k1LoW in https://github.com/2manymws/rc/pull/50

## [v0.8.1](https://github.com/2manymws/rc/compare/v0.8.0...v0.8.1) - 2024-01-11
### Fix bug 🐛
- Fix Set-Cookie handling by @k1LoW in https://github.com/2manymws/rc/pull/48

## [v0.8.0](https://github.com/2manymws/rc/compare/v0.7.3...v0.8.0) - 2024-01-04
### Breaking Changes 🛠
- Use req.Host only ( does not use req.URL.Host ) by @k1LoW in https://github.com/2manymws/rc/pull/46

## [v0.7.3](https://github.com/2manymws/rc/compare/v0.7.2...v0.7.3) - 2024-01-04
### Fix bug 🐛
- Use the Date header field value first. by @k1LoW in https://github.com/2manymws/rc/pull/44

## [v0.7.2](https://github.com/2manymws/rc/compare/v0.7.1...v0.7.2) - 2023-12-22
### Other Changes
- Revert "Add CacherHandler" by @k1LoW in https://github.com/2manymws/rc/pull/42

## [v0.7.1](https://github.com/2manymws/rc/compare/v0.7.0...v0.7.1) - 2023-12-21
### Other Changes
- Add CacherHandler by @k1LoW in https://github.com/2manymws/rc/pull/40

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
