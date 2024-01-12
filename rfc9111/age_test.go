package rfc9111

import (
	"net/http"
	"testing"
	"time"
)

func TestSetAgeHeader(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name       string
		useCached  bool
		resHeader  http.Header
		now        time.Time
		wantAge    string
		wantHeader http.Header
	}{
		{
			name:       "No cache and no Age header",
			useCached:  false,
			resHeader:  http.Header{},
			now:        now,
			wantAge:    "",
			wantHeader: http.Header{},
		},
		{
			name:      "No cache and Age header +5sec",
			useCached: false,
			resHeader: http.Header{
				"Age": []string{"5"},
			},
			now:     now,
			wantAge: "5",
			wantHeader: http.Header{
				"Age": []string{"5"},
			},
		},
		{
			name:      "Cached +10sec",
			useCached: true,
			resHeader: http.Header{
				"Date": []string{now.Add(-10 * time.Second).Format(http.TimeFormat)},
			},
			now:     now,
			wantAge: "10",
			wantHeader: http.Header{
				"Age":  []string{"10"},
				"Date": []string{now.Add(-10 * time.Second).Format(http.TimeFormat)},
			},
		},
		{
			name:      "Cached +10sec with Age header +5sec",
			useCached: true,
			resHeader: http.Header{
				"Age":  []string{"5"},
				"Date": []string{now.Add(-10 * time.Second).Format(http.TimeFormat)},
			},
			now:     now,
			wantAge: "15",
			wantHeader: http.Header{
				"Age":  []string{"15"},
				"Date": []string{now.Add(-10 * time.Second).Format(http.TimeFormat)},
			},
		},
		{
			name:      "invalid Age header",
			useCached: true,
			resHeader: http.Header{
				"Age":  []string{"invalid"},
				"Date": []string{now.Add(-10 * time.Second).Format(http.TimeFormat)},
			},
			now:     now,
			wantAge: "10",
			wantHeader: http.Header{
				"Age":  []string{"10"},
				"Date": []string{now.Add(-10 * time.Second).Format(http.TimeFormat)},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setAgeHeader(tt.useCached, tt.resHeader, tt.now)
			gotAge := tt.resHeader.Get("Age")
			if gotAge != tt.wantAge {
				t.Errorf("Age header got = %v, want %v", gotAge, tt.wantAge)
			}
			if !headersEqual(tt.resHeader, tt.wantHeader) {
				t.Errorf("Headers got = %v, want %v", tt.resHeader, tt.wantHeader)
			}
		})
	}
}

// headersEqual compares two http.Header objects for equality.
func headersEqual(a, b http.Header) bool {
	if len(a) != len(b) {
		return false
	}
	for key, av := range a {
		bv, ok := b[key]
		if !ok {
			return false
		}
		for i, v := range av {
			if v != bv[i] {
				return false
			}
		}
	}
	return true
}
