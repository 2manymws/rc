package rfc9111

import "net/http"

var defaultUnderstoodMethods = []string{
	// Although RFC 9111 does not explicitly define this, in general cacheable methods are GET and HEAD, so the default should be GET and HEAD.
	http.MethodGet,
	http.MethodHead,
	// http.MethodPost,
	// http.MethodPut,
	// http.MethodPatch,
	// http.MethodDelete,
	// http.MethodConnect,
	// http.MethodOptions,
	// http.MethodTrace,
}

var defaultUnderstoodStatusCodes = []int{
	http.StatusContinue,           // 100 / RFC 9110, 15.2.1
	http.StatusSwitchingProtocols, // 101 / RFC 9110, 15.2.2
	http.StatusProcessing,         // 102 / RFC 2518, 10.1
	http.StatusEarlyHints,         // 103 / RFC 8297

	http.StatusOK,                   // 200 / RFC 9110, 15.3.1
	http.StatusCreated,              // 201 / RFC 9110, 15.3.2
	http.StatusAccepted,             // 202 / RFC 9110, 15.3.3
	http.StatusNonAuthoritativeInfo, // 203 / RFC 9110, 15.3.4
	http.StatusNoContent,            // 204 / RFC 9110, 15.3.5
	http.StatusResetContent,         // 205 / RFC 9110, 15.3.6
	http.StatusPartialContent,       // 206 / RFC 9110, 15.3.7
	http.StatusMultiStatus,          // 207 / RFC 4918, 11.1
	http.StatusAlreadyReported,      // 208 / RFC 5842, 7.1
	http.StatusIMUsed,               // 226 / RFC 3229, 10.4.1

	http.StatusMultipleChoices,  // 300 / RFC 9110, 15.4.1
	http.StatusMovedPermanently, // 301 / RFC 9110, 15.4.2
	http.StatusFound,            // 302 / RFC 9110, 15.4.3
	http.StatusSeeOther,         // 303 / RFC 9110, 15.4.4
	http.StatusNotModified,      // 304 / RFC 9110, 15.4.5
	http.StatusUseProxy,         // 305 / RFC 9110, 15.4.6

	http.StatusTemporaryRedirect, // 307 / RFC 9110, 15.4.8
	http.StatusPermanentRedirect, // 308 / RFC 9110, 15.4.9

	http.StatusBadRequest,                   // 400 / RFC 9110, 15.5.1
	http.StatusUnauthorized,                 // 401 / RFC 9110, 15.5.2
	http.StatusPaymentRequired,              // 402 / RFC 9110, 15.5.3
	http.StatusForbidden,                    // 403 / RFC 9110, 15.5.4
	http.StatusNotFound,                     // 404 / RFC 9110, 15.5.5
	http.StatusMethodNotAllowed,             // 405 / RFC 9110, 15.5.6
	http.StatusNotAcceptable,                // 406 / RFC 9110, 15.5.7
	http.StatusProxyAuthRequired,            // 407 / RFC 9110, 15.5.8
	http.StatusRequestTimeout,               // 408 / RFC 9110, 15.5.9
	http.StatusConflict,                     // 409 / RFC 9110, 15.5.10
	http.StatusGone,                         // 410 / RFC 9110, 15.5.11
	http.StatusLengthRequired,               // 411 / RFC 9110, 15.5.12
	http.StatusPreconditionFailed,           // 412 / RFC 9110, 15.5.13
	http.StatusRequestEntityTooLarge,        // 413 / RFC 9110, 15.5.14
	http.StatusRequestURITooLong,            // 414 / RFC 9110, 15.5.15
	http.StatusUnsupportedMediaType,         // 415 / RFC 9110, 15.5.16
	http.StatusRequestedRangeNotSatisfiable, // 416 / RFC 9110, 15.5.17
	http.StatusExpectationFailed,            // 417 / RFC 9110, 15.5.18
	http.StatusTeapot,                       // 418 / RFC 9110, 15.5.19 (Unused)
	http.StatusMisdirectedRequest,           // 421 / RFC 9110, 15.5.20
	http.StatusUnprocessableEntity,          // 422 / RFC 9110, 15.5.21
	http.StatusLocked,                       // 423 / RFC 4918, 11.3
	http.StatusFailedDependency,             // 424 / RFC 4918, 11.4
	http.StatusTooEarly,                     // 425 / RFC 8470, 5.2.
	http.StatusUpgradeRequired,              // 426 / RFC 9110, 15.5.22
	http.StatusPreconditionRequired,         // 428 / RFC 6585, 3
	http.StatusTooManyRequests,              // 429 / RFC 6585, 4
	http.StatusRequestHeaderFieldsTooLarge,  // 431 / RFC 6585, 5
	http.StatusUnavailableForLegalReasons,   // 451 / RFC 7725, 3

	http.StatusInternalServerError,           // 500 / RFC 9110, 15.6.1
	http.StatusNotImplemented,                // 501 / RFC 9110, 15.6.2
	http.StatusBadGateway,                    // 502 / RFC 9110, 15.6.3
	http.StatusServiceUnavailable,            // 503 / RFC 9110, 15.6.4
	http.StatusGatewayTimeout,                // 504 / RFC 9110, 15.6.5
	http.StatusHTTPVersionNotSupported,       // 505 / RFC 9110, 15.6.6
	http.StatusVariantAlsoNegotiates,         // 506 / RFC 2295, 8.1
	http.StatusInsufficientStorage,           // 507 / RFC 4918, 11.5
	http.StatusLoopDetected,                  // 508 / RFC 5842, 7.2
	http.StatusNotExtended,                   // 510 / RFC 2774, 7
	http.StatusNetworkAuthenticationRequired, // 511 / RFC 6585, 6
}

// Responses with status codes that are defined as heuristically cacheable (e.g., 200, 203, 204, 206, 300, 301, 308, 404, 405, 410, 414, and 501 in this specification) can be reused by a cache with heuristic expiration unless otherwise indicated by the method definition or explicit cache controls [CACHING];
// https://httpwg.org/specs/rfc9110.html#rfc.section.15.1
var defaultHeuristicallyCacheableStatusCodes = []int{ //nostyle:varnames
	http.StatusOK,                   // 200 / RFC 9110, 15.3.1
	http.StatusNonAuthoritativeInfo, // 203 / RFC 9110, 15.3.4
	http.StatusNoContent,            // 204 / RFC 9110, 15.3.5
	http.StatusPartialContent,       // 206 / RFC 9110, 15.3.7

	http.StatusMultipleChoices,  // 300 / RFC 9110, 15.4.1
	http.StatusMovedPermanently, // 301 / RFC 9110, 15.4.2

	http.StatusPermanentRedirect, // 308 / RFC 9110, 15.4.9

	http.StatusNotFound,          // 404 / RFC 9110, 15.5.5
	http.StatusMethodNotAllowed,  // 405 / RFC 9110, 15.5.6
	http.StatusGone,              // 410 / RFC 9110, 15.5.11
	http.StatusRequestURITooLong, // 414 / RFC 9110, 15.5.15

	http.StatusNotImplemented, // 501 / RFC 9110, 15.6.2
}

var defaultHeuristicExpirationRatio = 0.1
