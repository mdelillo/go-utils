package net

import (
	"net/http"
	"time"
)

var _ RateLimiter = (*RollingWindowRateLimiter)(nil)

type RollingWindowRateLimiter struct {
	Window       time.Duration
	RequestLimit int
	requestTimes []time.Time
}

func (r *RollingWindowRateLimiter) AddRequest(_ *http.Request, t time.Time) {
	r.requestTimes = append(r.requestTimes, t)
}

func (r *RollingWindowRateLimiter) GetBackoffAt(_ *http.Request, t time.Time) time.Duration {
	r.removeOldRequestTimes()

	requestTimes := r.getRelevantRequestTimes(t)

	if len(requestTimes) < r.RequestLimit {
		return 0
	}

	return r.Window - t.Sub(requestTimes[0])
}

func (r *RollingWindowRateLimiter) removeOldRequestTimes() {
	var requestTimes []time.Time
	for _, t := range r.requestTimes {
		if t.Before(time.Now().Add(-1 * r.Window)) {
			continue
		}

		requestTimes = append(requestTimes, t)
	}
	r.requestTimes = requestTimes
}

func (r *RollingWindowRateLimiter) getRelevantRequestTimes(t time.Time) []time.Time {
	if len(r.requestTimes) == 0 {
		return nil
	}

	startIndex := 0
	endIndex := len(r.requestTimes) - 1

	startOfWindow := t.Add(-1 * r.Window)
	for r.requestTimes[startIndex].Before(startOfWindow) && startIndex <= endIndex {
		startIndex++
	}

	for r.requestTimes[endIndex].After(t) && endIndex > 0 {
		endIndex--
	}

	return r.requestTimes[startIndex : endIndex+1]
}
