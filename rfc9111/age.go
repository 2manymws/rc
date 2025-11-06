package rfc9111

import (
	"net/http"
	"strconv"
	"time"
)

func setAgeHeader(useCached bool, resHeader http.Header, now time.Time) {
	// 4.2.3. Calculating Age
	if !useCached {
		// The presence of an Age header field implies that the response was not generated or validated by the origin server for this request. However, lack of an Age header field does not imply the origin was contacted (https://www.rfc-editor.org/rfc/rfc9111#section-5.1).
		return
	}
	// The following is straight code with the expectation that it will be optimized by the compiler
	var (
		// age_value
		ageValue int
		// date_value
		dateValue time.Time
		// now
		// request_time
		requestTime time.Time
		// response_time
		responseTime time.Time
		err          error
	)
	ageValue, err = strconv.Atoi(resHeader.Get("Age"))
	if err != nil {
		ageValue = 0
	}

	dateValue, err = http.ParseTime(resHeader.Get("Date"))
	if err != nil {
		return
	}
	requestTime = dateValue // Approximate value.
	responseTime = now      // Approximate value.
	// apparent_age = max(0, response_time - date_value);
	apparentAge := max(0, int(responseTime.Sub(dateValue)/time.Second))
	// response_delay = response_time - request_time
	responseDelay := int(responseTime.Sub(requestTime) / time.Second)
	// corrected_age_value = age_value + response_delay
	correctedAgeValue := ageValue + responseDelay
	// corrected_initial_age = max(apparent_age, corrected_age_value)
	correctedInitialAge := max(apparentAge, correctedAgeValue)
	// resident_time = now - response_time;
	residentTime := int(now.Sub(responseTime) / time.Second)
	// current_age = corrected_initial_age + resident_time;
	currentAge := correctedInitialAge + residentTime
	resHeader.Set("Age", strconv.Itoa(currentAge))
}
