package rfc9111

import (
	"math"
	"strconv"
	"strings"
)

type RequestDirectives struct {
	// max-age https://httpwg.org/specs/rfc9111.html#rfc.section.5.2.1.1.
	MaxAge *uint32
	// max-stale https://httpwg.org/specs/rfc9111.html#rfc.section.5.2.1.2.
	MaxStale *uint32
	// min-fresh https://httpwg.org/specs/rfc9111.html#rfc.section.5.2.1.3.
	MinFresh *uint32
	// no-cache https://httpwg.org/specs/rfc9111.html#rfc.section.5.2.1.4.
	NoCache bool
	// no-store https://httpwg.org/specs/rfc9111.html#rfc.section.5.2.1.5.
	NoStore bool
	// no-transform https://httpwg.org/specs/rfc9111.html#rfc.section.5.2.1.6.
	NoTransform bool
	// only-if-cached https://httpwg.org/specs/rfc9111.html#rfc.section.5.2.1.7.
	OnlyIfCached bool
}

type ResponseDirectives struct {
	// max-age https://httpwg.org/specs/rfc9111.html#rfc.section.5.2.2.1.
	MaxAge *uint32
	// must-revalidate https://httpwg.org/specs/rfc9111.html#rfc.section.5.2.2.2.
	MustRevalidate bool
	// must-understand https://httpwg.org/specs/rfc9111.html#rfc.section.5.2.2.3.
	MustUnderstand bool
	// no-cache https://httpwg.org/specs/rfc9111.html#rfc.section.5.2.2.4.
	NoCache bool
	// no-store https://httpwg.org/specs/rfc9111.html#rfc.section.5.2.2.5.
	NoStore bool
	// no-transform https://httpwg.org/specs/rfc9111.html#rfc.section.5.2.2.6.
	NoTransform bool
	// private https://httpwg.org/specs/rfc9111.html#rfc.section.5.2.2.7.
	Private bool
	// proxy-revalidate https://httpwg.org/specs/rfc9111.html#rfc.section.5.2.2.8.
	ProxyRevalidate bool
	// public https://httpwg.org/specs/rfc9111.html#rfc.section.5.2.2.9.
	Public bool
	// s-maxag https://httpwg.org/specs/rfc9111.html#rfc.section.5.2.2.10.
	SMaxAge *uint32
}

// ParseRequestCacheControlHeader parses the Cache-Control header of a request.
func ParseRequestCacheControlHeader(headers []string) *RequestDirectives {
	d := &RequestDirectives{}
	for _, h := range headers {
		tokens := strings.Split(h, ",")
		for _, t := range tokens {
			t = strings.TrimSpace(t)
			switch {
			// When there is more than one value present for a given directive (e.g., two Expires header field lines or multiple Cache-Control: max-age directives), either the first occurrence should be used or the response should be considered stale.
			case strings.HasPrefix(t, "max-age=") && d.MaxAge == nil:
				sec := strings.TrimPrefix(t, "max-age=")
				u64, err := strconv.ParseUint(sec, 10, 32)
				if err != nil {
					continue
				}
				u32 := uint32(u64)
				d.MaxAge = &u32
			case strings.HasPrefix(t, "max-stale=") && d.MaxStale == nil:
				sec := strings.TrimPrefix(t, "max-stale=")
				u64, err := strconv.ParseUint(sec, 10, 32)
				if err != nil {
					continue
				}
				u32 := uint32(u64)
				d.MaxStale = &u32
			case strings.HasPrefix(t, "max-stale") && d.MaxStale == nil:
				// If no value is assigned to max-stale, then the client will accept a stale response of any age (ref https://httpwg.org/specs/rfc9111.html#rfc.section.5.2.1.2).
				max := uint32(math.MaxUint32)
				d.MaxStale = &max
			case strings.HasPrefix(t, "min-fresh=") && d.MinFresh == nil:
				sec := strings.TrimPrefix(t, "min-fresh=")
				u64, err := strconv.ParseUint(sec, 10, 32)
				if err != nil {
					continue
				}
				u32 := uint32(u64)
				d.MinFresh = &u32
			case t == "no-cache":
				d.NoCache = true
			case t == "no-store":
				d.NoStore = true
			case t == "no-transform":
				d.NoTransform = true
			case t == "only-if-cached":
				d.OnlyIfCached = true
			default:
				// A cache MUST ignore unrecognized cache directives. (https://httpwg.org/specs/rfc9111.html#rfc.section.5.2.3)
			}
		}
	}
	return d
}

// ParseResponseCacheControlHeader parses the Cache-Control header of a response.
func ParseResponseCacheControlHeader(headers []string) *ResponseDirectives {
	d := &ResponseDirectives{}
	for _, h := range headers {
		tokens := strings.Split(h, ",")
		for _, t := range tokens {
			t = strings.TrimSpace(t)
			switch {
			// When there is more than one value present for a given directive (e.g., two Expires header field lines or multiple Cache-Control: max-age directives), either the first occurrence should be used or the response should be considered stale.
			case strings.HasPrefix(t, "max-age=") && d.MaxAge == nil:
				sec := strings.TrimPrefix(t, "max-age=")
				u64, err := strconv.ParseUint(sec, 10, 32)
				if err != nil {
					continue
				}
				u32 := uint32(u64)
				d.MaxAge = &u32
			case t == "must-revalidate":
				d.MustRevalidate = true
			case t == "must-understand":
				d.MustUnderstand = true
			case t == "no-cache":
				d.NoCache = true
			case t == "no-store":
				d.NoStore = true
			case t == "no-transform":
				d.NoTransform = true
			case t == "private":
				d.Private = true
			case t == "proxy-revalidate":
				d.ProxyRevalidate = true
			case t == "public":
				d.Public = true
			case strings.HasPrefix(t, "s-maxage=") && d.SMaxAge == nil:
				sec := strings.TrimPrefix(t, "s-maxage=")
				u64, err := strconv.ParseUint(sec, 10, 32)
				if err != nil {
					continue
				}
				u32 := uint32(u64)
				d.SMaxAge = &u32
			default:
				// A cache MUST ignore unrecognized cache directives. (https://httpwg.org/specs/rfc9111.html#rfc.section.5.2.3)
			}
		}
	}
	return d
}
