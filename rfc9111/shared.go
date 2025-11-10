package rfc9111

import (
	"net/http"
	"strings"
	"time"
)

// Shared is a shared cache that implements RFC 9111.
// The following features are not implemented
// - Private cache.
// - Request directives.
type Shared struct {
	understoodMethods                 []string
	understoodStatusCodes             []int
	heuristicallyCacheableStatusCodes []int
	heuristicExpirationRatio          float64
	storeRequestWithSetCookieHeader   bool
	extendedRules                     []ExtendedRule
}

// ExtendedRule is an extended rule.
// Like proxy_cache_valid of NGINX.
// Rules are applied only when there is no Cache-Control header and the expiration time cannot be calculated.
// THIS IS NOT RFC 9111.
type ExtendedRule interface { //nostyle:ifacenames
	// Cacheable returns true and and the expiration time if the response is cacheable.
	Cacheable(req *http.Request, res *http.Response) (ok bool, age time.Duration)
}

// SharedOption is an option for Shared.
type SharedOption func(*Shared) error

// UnderstoodMethods sets the understood methods.
func UnderstoodMethods(methods []string) SharedOption {
	return func(s *Shared) error {
		s.understoodMethods = methods
		return nil
	}
}

// UnderstoodStatusCodes sets the understood status codes.
func UnderstoodStatusCodes(statusCodes []int) SharedOption {
	return func(s *Shared) error {
		s.understoodStatusCodes = statusCodes
		return nil
	}
}

// HeuristicallyCacheableStatusCodes sets the heuristically cacheable status codes.
func HeuristicallyCacheableStatusCodes(statusCodes []int) SharedOption {
	return func(s *Shared) error {
		s.heuristicallyCacheableStatusCodes = statusCodes
		return nil
	}
}

// HeuristicExpirationRatio sets the heuristic expiration ratio.
func HeuristicExpirationRatio(ratio float64) SharedOption {
	return func(s *Shared) error {
		if ratio < 0 {
			return ErrNegativeRatio
		}
		s.heuristicExpirationRatio = ratio
		return nil
	}
}

// StoreRequestWithSetCookieHeader enables storing request with Set-Cookie header.
func StoreRequestWithSetCookieHeader() SharedOption {
	return func(s *Shared) error {
		s.storeRequestWithSetCookieHeader = true
		return nil
	}
}

// ExtendedRules sets the extended rules.
func ExtendedRules(rules []ExtendedRule) SharedOption {
	return func(s *Shared) error {
		s.extendedRules = rules
		return nil
	}
}

// NewShared returns a new Shared cache handler.
func NewShared(opts ...SharedOption) (*Shared, error) {
	s := &Shared{
		heuristicExpirationRatio: defaultHeuristicExpirationRatio,
	}

	um := make([]string, len(defaultUnderstoodMethods))
	_ = copy(um, defaultUnderstoodMethods)
	s.understoodMethods = um

	us := make([]int, len(defaultUnderstoodStatusCodes))
	_ = copy(us, defaultUnderstoodStatusCodes)
	s.understoodStatusCodes = us

	hs := make([]int, len(defaultHeuristicallyCacheableStatusCodes))
	_ = copy(hs, defaultHeuristicallyCacheableStatusCodes)
	s.heuristicallyCacheableStatusCodes = hs

	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}

	return s, nil
}

// Storable returns true if the response is storable in the cache.
func (s *Shared) Storable(req *http.Request, res *http.Response, now time.Time) (bool, time.Time) {
	// 3. Storing Responses in Caches (https://www.rfc-editor.org/rfc/rfc9111#section-3)
	// - the request method is understood by the cache;
	if !contains(req.Method, s.understoodMethods) {
		return s.storableWithExtendedRules(req, res, now)
	}

	// - the response status code is final (see https://www.rfc-editor.org/rfc/rfc9110#section-15);
	if !isFinalStatusCode(res.StatusCode) {
		return s.storableWithExtendedRules(req, res, now)
	}

	rescc := ParseResponseCacheControlHeader(res.Header.Values("Cache-Control"))

	// - if the response status code is 206 or 304, or the must-understand cache directive (see https://www.rfc-editor.org/rfc/rfc9111#section-5.2.2.3) is present: the cache understands the response status code;
	if (contains(res.StatusCode, []int{http.StatusPartialContent, http.StatusNotModified}) &&
		!contains(res.StatusCode, s.understoodStatusCodes)) ||
		(rescc.MustUnderstand && !contains(res.StatusCode, s.understoodStatusCodes)) {
		return s.storableWithExtendedRules(req, res, now)
	}

	// - the no-store cache directive is not present in the response (see https://www.rfc-editor.org/rfc/rfc9111#section-5.2.2.5);
	if rescc.NoStore {
		return false, time.Time{}
	}

	// - if the cache is shared: the private response directive is either not present or allows a shared cache to store a modified response; see https://www.rfc-editor.org/rfc/rfc9111#section-5.2.2.7);
	if rescc.Private {
		return false, time.Time{}
	}

	// - if the cache is shared: the Authorization header field is not present in the request (see https://www.rfc-editor.org/rfc/rfc9111#section-11.6.2 of [HTTP]) or a response directive is present that explicitly allows shared caching (see https://www.rfc-editor.org/rfc/rfc9111#section-3.5);
	// In this specification, the following response directives have such an effect: must-revalidate (https://www.rfc-editor.org/rfc/rfc9111#section-5.2.2.2), public (https://www.rfc-editor.org/rfc/rfc9111#section-5.2.2.9), and s-maxage (https://www.rfc-editor.org/rfc/rfc9111#section-5.2.2.10).
	if req.Header.Get("Authorization") != "" && !rescc.MustRevalidate && !rescc.Public && rescc.SMaxAge == nil {
		return false, time.Time{}
	}

	// In RFC 9111, Servers that wish to control caching of responses with Set-Cookie headers are encouraged to emit appropriate Cache-Control response header fields (see https://www.rfc-editor.org/rfc/rfc9111#section-7.3).
	// But to beat on the safe side, this package does not store responses with Set-Cookie headers by default, similar to NGINX.
	// THIS IS NOT RFC 9111.
	if res.Header.Get("Set-Cookie") != "" && !s.storeRequestWithSetCookieHeader {
		return false, time.Time{}
	}

	// - the response contains at least one of the following:

	//   * a public response directive (see https://www.rfc-editor.org/rfc/rfc9111#section-5.2.2.9);
	if rescc.Public {
		exp := CalclateExpires(rescc, res.Header, s.heuristicExpirationRatio, now)
		return true, exp
	}
	//   * a private response directive, if the cache is not shared (see https://www.rfc-editor.org/rfc/rfc9111#section-5.2.2.7);
	// THE CACHE IS SHARED

	//   * an Expires header field (see https://www.rfc-editor.org/rfc/rfc9111#section-5.3);
	if res.Header.Get("Expires") != "" {
		exp := CalclateExpires(rescc, res.Header, s.heuristicExpirationRatio, now)
		return true, exp
	}
	//   * a max-age response directive (see https://www.rfc-editor.org/rfc/rfc9111#section-5.2.2.1);
	if rescc.MaxAge != nil {
		exp := CalclateExpires(rescc, res.Header, s.heuristicExpirationRatio, now)
		return true, exp
	}
	//   * if the cache is shared: an s-maxage response directive (see https://www.rfc-editor.org/rfc/rfc9111#section-5.2.2.10);
	if rescc.SMaxAge != nil {
		exp := CalclateExpires(rescc, res.Header, s.heuristicExpirationRatio, now)
		return true, exp
	}
	//   * a cache extension that allows it to be cached (see https://www.rfc-editor.org/rfc/rfc9111#section-5.2.3); or
	// NOT IMPLEMENTED

	//   * a status code that is defined as heuristically cacheable (see https://www.rfc-editor.org/rfc/rfc9111#section-4.2.2).
	if contains(res.StatusCode, s.heuristicallyCacheableStatusCodes) {
		exp := CalclateExpires(rescc, res.Header, s.heuristicExpirationRatio, now)
		// Only store if we can calculate an expiration time
		if exp.Sub(time.Time{}) != 0 {
			return true, exp
		}
	}

	return s.storableWithExtendedRules(req, res, now)
}

func (s *Shared) Handle(req *http.Request, cachedReq *http.Request, cachedRes *http.Response, do func(*http.Request) (*http.Response, error), now time.Time) (useCached bool, r *http.Response, _ error) {
	defer func() {
		// 5.1 Age (https://www.rfc-editor.org/rfc/rfc9111#section-5.1)
		if r != nil {
			setAgeHeader(useCached, r.Header, now)
		}
	}()

	if cachedReq == nil || cachedRes == nil {
		res, err := do(req)
		return false, res, err
	}

	// 4. Constructing Responses from Caches
	// When presented with a request, a cache MUST NOT reuse a stored response unless:

	// - the presented target URI (https://www.rfc-editor.org/rfc/rfc9110#section-7.1 of [HTTP]) and that of the stored response match, and
	if cachedReq.Host == "" || req.Host != cachedReq.Host || req.URL.Path != cachedReq.URL.Path || req.URL.RawQuery != cachedReq.URL.RawQuery {
		// For SNI compatibility, also compare req.Host
		res, err := do(req)
		return false, res, err
	}

	// - the request method associated with the stored response allows it to be used for the presented request, and
	if req.Method != cachedReq.Method { // FIXME: more strictly
		res, err := do(req)
		return false, res, err
	}

	// - request header fields nominated by the stored response (if any) match those presented (see https://www.rfc-editor.org/rfc/rfc9111#section-4.1)
	if v := cachedRes.Header.Values("Vary"); len(v) != 0 {
		vary := strings.Join(v, ",")
		if strings.Contains(vary, "*") {
			res, err := do(req)
			return false, res, err
		}
		for _, h := range strings.Split(vary, ",") {
			h = strings.TrimSpace(h)
			if req.Header.Get(h) != cachedReq.Header.Get(h) { // FIXME: more strictly
				res, err := do(req)
				return false, res, err
			}
		}
	}

	rescc := ParseResponseCacheControlHeader(cachedRes.Header.Values("Cache-Control"))

	// - the stored response does not contain the no-cache directive (https://www.rfc-editor.org/rfc/rfc9111#section-5.2.2.4), unless it is successfully validated (https://www.rfc-editor.org/rfc/rfc9111#section-4.3)
	if rescc.NoCache {
		// The no-cache response directive, in its unqualified form (without an argument), indicates that the response MUST NOT be used to satisfy any other request without forwarding it for validation and receiving a successful response; see https://www.rfc-editor.org/rfc/rfc9111#section-4.3.
		if req.Method == http.MethodGet || req.Method == http.MethodHead {
			if cachedRes.Header.Get("ETag") != "" {
				req.Header.Set("If-None-Match", cachedRes.Header.Get("ETag"))
			}
			if cachedRes.Header.Get("Last-Modified") != "" {
				req.Header.Set("If-Modified-Since", cachedRes.Header.Get("Last-Modified"))
			}
			res, err := do(req)
			if err != nil {
				return false, res, err
			}
			if res.StatusCode != http.StatusNotModified {
				return false, res, nil
			}
			// The qualified form of the no-cache response directive, with an argument that lists one or more field names, indicates that a cache MAY use the response to satisfy a subsequent request, subject to any other restrictions on caching, if the listed header fields are excluded from the subsequent response or the subsequent response has been successfully revalidated with the origin server (updating or removing those fields).
		} else {
			res, err := do(req)
			return false, res, err
		}
	}

	expires := CalclateExpires(rescc, cachedRes.Header, s.heuristicExpirationRatio, now)

	// - the stored response is one of the following:
	//   * fresh (see https://www.rfc-editor.org/rfc/rfc9111#section-4.2), or
	if expires.Sub(now) > 0 {
		return true, cachedRes, nil
	}

	//   * allowed to be served stale (see https://www.rfc-editor.org/rfc/rfc9111#section-4.2.4), or
	if !rescc.NoCache && !rescc.MustRevalidate && rescc.SMaxAge == nil && !rescc.ProxyRevalidate {
		//     > A cache MUST NOT generate a stale response if it is prohibited by an explicit in-protocol directive (e.g., by a no-cache response directive, a must-revalidate response directive, or an applicable s-maxage or proxy-revalidate response directive; see https://www.rfc-editor.org/rfc/rfc9111#section-5.2.2).
		reqcc := ParseRequestCacheControlHeader(req.Header.Values("Cache-Control"))
		//     > A cache MUST NOT generate a stale response unless it is disconnected or doing so is explicitly permitted by the client or origin server (e.g., by the max-stale request directive in https://www.rfc-editor.org/rfc/rfc9111#section-5.2.1, extension directives such as those defined in [RFC5861], or configuration in accordance with an out-of-band contract).

		// stale-while-revalidate: https://www.rfc-editor.org/rfc/rfc5861
		// Permits serving stale response while revalidating in background
		if rescc.StaleWhileRevalidate != nil {
			age := now.Sub(expires)
			swr := time.Duration(*rescc.StaleWhileRevalidate) * time.Second
			if age >= 0 && age < swr {
				// Within stale-while-revalidate window, use cached response
				// and trigger background revalidation
				go func() {
					// Background revalidation: do() will fetch from origin and update cache
					_, _ = do(req) //nostyle:handlerrors
				}()
				return true, cachedRes, nil
			}
		}

		if reqcc.MaxStale != nil {
			if expires.Add(time.Duration(*reqcc.MaxStale)*time.Second).Sub(now) > 0 {
				return true, cachedRes, nil
			}
		}
	}

	//   * successfully validated (see https://www.rfc-editor.org/rfc/rfc9111#section-4.3).
	if req.Method == http.MethodGet || req.Method == http.MethodHead {
		if cachedRes.Header.Get("ETag") != "" {
			req.Header.Set("If-None-Match", cachedRes.Header.Get("ETag"))
		}
		if cachedRes.Header.Get("Last-Modified") != "" {
			req.Header.Set("If-Modified-Since", cachedRes.Header.Get("Last-Modified"))
		}
		res, err := do(req)
		if err != nil {
			// stale-if-error: https://www.rfc-editor.org/rfc/rfc5861
			// Permits serving stale response when error occurs
			if rescc.StaleIfError != nil {
				age := now.Sub(expires)
				sie := time.Duration(*rescc.StaleIfError) * time.Second
				if age >= 0 && age < sie {
					// Within stale-if-error window, use cached response on error
					return true, cachedRes, nil
				}
			}
			return false, res, err
		}
		// stale-if-error also applies to 5xx errors (500, 502, 503, 504)
		if rescc.StaleIfError != nil && (res.StatusCode == http.StatusInternalServerError ||
			res.StatusCode == http.StatusBadGateway ||
			res.StatusCode == http.StatusServiceUnavailable ||
			res.StatusCode == http.StatusGatewayTimeout) {
			age := now.Sub(expires)
			sie := time.Duration(*rescc.StaleIfError) * time.Second
			if age >= 0 && age < sie {
				// Within stale-if-error window, use cached response on 5xx error
				return true, cachedRes, nil
			}
		}
		if res.StatusCode == http.StatusNotModified {
			return true, cachedRes, nil
		}
		return false, res, nil
	}

	res, err := do(req)
	if err != nil {
		// stale-if-error: https://www.rfc-editor.org/rfc/rfc5861
		// Permits serving stale response when error occurs
		if rescc.StaleIfError != nil {
			age := now.Sub(expires)
			sie := time.Duration(*rescc.StaleIfError) * time.Second
			if age >= 0 && age < sie {
				// Within stale-if-error window, use cached response on error
				return true, cachedRes, nil
			}
		}
		return false, res, err
	}
	// stale-if-error also applies to 5xx errors (500, 502, 503, 504)
	if rescc.StaleIfError != nil && (res.StatusCode == http.StatusInternalServerError ||
		res.StatusCode == http.StatusBadGateway ||
		res.StatusCode == http.StatusServiceUnavailable ||
		res.StatusCode == http.StatusGatewayTimeout) {
		age := now.Sub(expires)
		sie := time.Duration(*rescc.StaleIfError) * time.Second
		if age >= 0 && age < sie {
			// Within stale-if-error window, use cached response on 5xx error
			return true, cachedRes, nil
		}
	}
	return false, res, err
}

// storableWithExtendedRules returns true if the response is storable with extended rules.
func (s *Shared) storableWithExtendedRules(req *http.Request, res *http.Response, now time.Time) (bool, time.Time) {
	if res.Header.Get("Cache-Control") != "" {
		return false, time.Time{}
	}

	for _, rule := range s.extendedRules {
		ok, age := rule.Cacheable(req, res)
		if ok {
			// Add Expires header field
			od := originDate(res.Header, now)
			expires := od.Add(age) //nostyle:varnames
			res.Header.Set("Expires", expires.UTC().Format(http.TimeFormat))
			return true, expires
		}
	}
	return false, time.Time{}
}

func CalclateExpires(d *ResponseDirectives, resHeader http.Header, heuristicExpirationRatio float64, now time.Time) time.Time {
	// 	4.2.1. Calculating Freshness Lifetime
	// A cache can calculate the freshness lifetime (denoted as freshness_lifetime) of a response by evaluating the following rules and using the first match:

	// - If the cache is shared and the s-maxage response directive (https://www.rfc-editor.org/rfc/rfc9111#section-5.2.2.10) is present, use its value
	if d.SMaxAge != nil {
		od := originDate(resHeader, now)
		return od.Add(time.Duration(*d.SMaxAge) * time.Second)
	}
	// - If the max-age response directive (https://www.rfc-editor.org/rfc/rfc9111#section-5.2.2.1) is present, use its value
	if d.MaxAge != nil {
		od := originDate(resHeader, now)
		return od.Add(time.Duration(*d.MaxAge) * time.Second)
	}
	if resHeader.Get("Expires") != "" {
		// - If the Expires response header field (https://www.rfc-editor.org/rfc/rfc9111#section-5.3) is present, use its value minus the value of the Date response header field (using the time the message was received if it is not present, as per Section 6.6.1 of [HTTP])
		et, err := http.ParseTime(resHeader.Get("Expires"))
		if err == nil {
			od := originDate(resHeader, now)
			return now.Add(et.Sub(od))
		}
	}
	// Otherwise, no explicit expiration time is present in the response. A heuristic freshness lifetime might be applicable; see https://www.rfc-editor.org/rfc/rfc9111#section-4.2.2.
	if resHeader.Get("Last-Modified") != "" {
		lt, err := http.ParseTime(resHeader.Get("Last-Modified"))
		if err == nil {
			// If the response has a Last-Modified header field (https://www.rfc-editor.org/rfc/rfc9110#section-8.8.2 of [HTTP]), caches are encouraged to use a heuristic expiration value that is no more than some fraction of the interval since that time. A typical setting of this fraction might be 10%.
			od := originDate(resHeader, now)
			return od.Add(time.Duration(float64(od.Sub(lt)) * heuristicExpirationRatio))
		}
	}

	// Can't calculate expires
	return time.Time{}
}

func isFinalStatusCode(status int) bool {
	if status >= 100 && status < 200 {
		return false
	}
	return true
}

func originDate(resHeader http.Header, now time.Time) time.Time {
	if resHeader.Get("Date") != "" {
		t, err := http.ParseTime(resHeader.Get("Date"))
		if err == nil {
			return t
		}
	}
	// A recipient with a clock that receives a response with an invalid Date header field value MAY replace that value with the time that response was received. (https://www.rfc-editor.org/rfc/rfc9110#section-6.6.1 of [HTTP])
	//
	// ...using the time the message was received if it is not present, as per https://www.rfc-editor.org/rfc/rfc9110#section-6.6.1 of [HTTP]
	// (https://www.rfc-editor.org/rfc/rfc9111#section-4.2.1)
	return now
}

func contains[T comparable](v T, vv []T) bool {
	for _, vvv := range vv {
		if vvv == v {
			return true
		}
	}
	return false
}
