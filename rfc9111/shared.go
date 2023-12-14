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
	// 3. Storing Responses in Caches (https://httpwg.org/specs/rfc9111.html#rfc.section.3)
	// - the request method is understood by the cache;
	if !contains(req.Method, s.understoodMethods) {
		return false, time.Time{}
	}

	// - the response status code is final (see https://httpwg.org/specs/rfc9110.html#rfc.section.15);
	if contains(res.StatusCode, []int{
		http.StatusContinue,
		http.StatusSwitchingProtocols,
		http.StatusProcessing,
		http.StatusEarlyHints,
	}) {
		return false, time.Time{}
	}

	rescc := ParseResponseCacheControlHeader(res.Header.Values("Cache-Control"))

	// - if the response status code is 206 or 304, or the must-understand cache directive (see https://httpwg.org/specs/rfc9111.html#rfc.section.5.2.2.3) is present: the cache understands the response status code;
	if contains(res.StatusCode, []int{
		http.StatusPartialContent,
		http.StatusNotModified,
	}) || (rescc.MustUnderstand && !contains(res.StatusCode, s.understoodStatusCodes)) {
		return false, time.Time{}
	}

	// - the no-store cache directive is not present in the response (see https://httpwg.org/specs/rfc9111.html#rfc.section.5.2.2.5);
	if rescc.NoStore {
		return false, time.Time{}
	}

	// - if the cache is shared: the private response directive is either not present or allows a shared cache to store a modified response; see https://httpwg.org/specs/rfc9111.html#rfc.section.5.2.2.7);
	if rescc.Private {
		return false, time.Time{}
	}

	// - if the cache is shared: the Authorization header field is not present in the request (see https://httpwg.org/specs/rfc9111.html#rfc.section.11.6.2 of [HTTP]) or a response directive is present that explicitly allows shared caching (see https://httpwg.org/specs/rfc9111.html#rfc.section.3.5);
	// In this specification, the following response directives have such an effect: must-revalidate (https://httpwg.org/specs/rfc9111.html#rfc.section.5.2.2.2), public (https://httpwg.org/specs/rfc9111.html#rfc.section.5.2.2.9), and s-maxage (https://httpwg.org/specs/rfc9111.html#rfc.section.5.2.2.10).
	if req.Header.Get("Authorization") != "" && !rescc.MustRevalidate && !rescc.Public && rescc.SMaxAge == nil {
		return false, time.Time{}
	}

	expires := CalclateExpires(rescc, res.Header, s.heuristicExpirationRatio, now)
	if expires.Sub(now) <= 0 {
		return false, time.Time{}
	}

	// - the response contains at least one of the following:

	//   * a public response directive (see https://httpwg.org/specs/rfc9111.html#rfc.section.5.2.2.9);
	if rescc.Public {
		return true, expires
	}
	//   * a private response directive, if the cache is not shared (see https://httpwg.org/specs/rfc9111.html#rfc.section.5.2.2.7);
	// THE CACHE IS SHARED

	//   * an Expires header field (see https://httpwg.org/specs/rfc9111.html#rfc.section.5.3);
	if res.Header.Get("Expires") != "" {
		return true, expires
	}
	//   * a max-age response directive (see https://httpwg.org/specs/rfc9111.html#rfc.section.5.2.2.1);
	if rescc.MaxAge != nil {
		return true, expires
	}
	//   * if the cache is shared: an s-maxage response directive (see https://httpwg.org/specs/rfc9111.html#rfc.section.5.2.2.10);
	if rescc.SMaxAge != nil {
		return true, expires
	}
	//   * a cache extension that allows it to be cached (see https://httpwg.org/specs/rfc9111.html#rfc.section.5.2.3); or
	// NOT IMPLEMENTED

	//   * a status code that is defined as heuristically cacheable (see https://httpwg.org/specs/rfc9111.html#rfc.section.4.2.2).
	if contains(res.StatusCode, s.heuristicallyCacheableStatusCodes) {
		return true, expires
	}

	return false, time.Time{}
}

func (s *Shared) Handle(req *http.Request, cachedReq *http.Request, cachedRes *http.Response, do func(*http.Request) (*http.Response, error), now time.Time) (bool, *http.Response, error) {
	if cachedReq == nil || cachedRes == nil {
		res, err := do(req)
		return false, res, err
	}

	// 4. Constructing Responses from Caches
	// When presented with a request, a cache MUST NOT reuse a stored response unless:

	// - the presented target URI (https://httpwg.org/specs/rfc9110.html#rfc.section.7.1 of [HTTP]) and that of the stored response match, and
	if req.URL.String() != cachedReq.URL.String() {
		res, err := do(req)
		return false, res, err
	}

	// - the request method associated with the stored response allows it to be used for the presented request, and
	if req.Method != cachedReq.Method { // FIXME: more strictly
		res, err := do(req)
		return false, res, err
	}

	// - request header fields nominated by the stored response (if any) match those presented (see https://httpwg.org/specs/rfc9111.html#rfc.section.4.1)
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

	// - the stored response does not contain the no-cache directive (https://httpwg.org/specs/rfc9111.html#rfc.section.5.2.2.4), unless it is successfully validated (https://httpwg.org/specs/rfc9111.html#rfc.section.4.3)
	if rescc.NoCache {
		// The no-cache response directive, in its unqualified form (without an argument), indicates that the response MUST NOT be used to satisfy any other request without forwarding it for validation and receiving a successful response; see https://httpwg.org/specs/rfc9111.html#rfc.section.4.3.
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
	//   * fresh (see https://httpwg.org/specs/rfc9111.html#rfc.section.4.2), or
	if expires.Sub(now) > 0 {
		return true, cachedRes, nil
	}

	//   * allowed to be served stale (see https://httpwg.org/specs/rfc9111.html#rfc.section.4.2.4), or
	if !rescc.NoCache && !rescc.MustRevalidate && rescc.SMaxAge == nil && !rescc.ProxyRevalidate {
		//     > A cache MUST NOT generate a stale response if it is prohibited by an explicit in-protocol directive (e.g., by a no-cache response directive, a must-revalidate response directive, or an applicable s-maxage or proxy-revalidate response directive; see https://httpwg.org/specs/rfc9111.html#rfc.section.5.2.2).
		reqcc := ParseRequestCacheControlHeader(req.Header.Values("Cache-Control"))
		//     > A cache MUST NOT generate a stale response unless it is disconnected or doing so is explicitly permitted by the client or origin server (e.g., by the max-stale request directive in https://httpwg.org/specs/rfc9111.html#rfc.section.5.2.1, extension directives such as those defined in [RFC5861], or configuration in accordance with an out-of-band contract).
		if reqcc.MaxStale == nil {
			// If no value is assigned to max-stale, then the client will accept a stale response of any age (ref https://httpwg.org/specs/rfc9111.html#rfc.section.5.2.1.2).
			return true, cachedRes, nil
		}
		if expires.Add(time.Duration(*reqcc.MaxStale)*time.Second).Sub(now) > 0 {
			return true, cachedRes, nil
		}
	}

	//   * successfully validated (see https://httpwg.org/specs/rfc9111.html#rfc.section.4.3).
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
		if res.StatusCode == http.StatusNotModified {
			return true, cachedRes, nil
		}
		return false, res, nil
	}

	res, err := do(req)
	return false, res, err
}

func CalclateExpires(d *ResponseDirectives, header http.Header, heuristicExpirationRatio float64, now time.Time) time.Time {
	// 	4.2.1. Calculating Freshness Lifetime
	// A cache can calculate the freshness lifetime (denoted as freshness_lifetime) of a response by evaluating the following rules and using the first match:

	// - If the cache is shared and the s-maxage response directive (https://httpwg.org/specs/rfc9111.html#rfc.section.5.2.2.10) is present, use its value, or
	if d.SMaxAge != nil {
		return now.Add(time.Duration(*d.SMaxAge) * time.Second)
	}
	// - If the max-age response directive (https://httpwg.org/specs/rfc9111.html#rfc.section.5.2.2.1) is present, use its value, or
	if d.MaxAge != nil {
		return now.Add(time.Duration(*d.MaxAge) * time.Second)
	}
	if header.Get("Expires") != "" {
		// - If the Expires response header field (https://httpwg.org/specs/rfc9111.html#rfc.section.5.3) is present, use its value minus the value of the Date response header field
		et, err := http.ParseTime(header.Get("Expires"))
		if err == nil {
			if header.Get("Date") != "" {
				dt, err := http.ParseTime(header.Get("Date"))
				if err == nil {
					return now.Add(et.Sub(dt))
				}
			} else {
				// (using the time the message was received if it is not present, as per https://httpwg.org/specs/rfc9110.html#rfc.section.6.6.1 of [HTTP])
				return et // == return now.Add(et.Sub(now))
			}
		}
	}
	// Otherwise, no explicit expiration time is present in the response. A heuristic freshness lifetime might be applicable; see https://httpwg.org/specs/rfc9111.html#rfc.section.4.2.2.
	if header.Get("Last-Modified") != "" {
		lt, err := http.ParseTime(header.Get("Last-Modified"))
		if err == nil {
			// If the response has a Last-Modified header field (https://httpwg.org/specs/rfc9110.html#rfc.section.8.8.2 of [HTTP]), caches are encouraged to use a heuristic expiration value that is no more than some fraction of the interval since that time. A typical setting of this fraction might be 10%.
			if header.Get("Date") != "" {
				dt, err := http.ParseTime(header.Get("Date"))
				if err == nil {
					return dt.Add(time.Duration(float64(dt.Sub(lt)) * heuristicExpirationRatio))
				}
			} else {
				return now.Add(time.Duration(float64(now.Sub(lt)) * heuristicExpirationRatio))
			}
		}
	}

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
