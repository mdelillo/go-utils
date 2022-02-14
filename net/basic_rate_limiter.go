package net

import (
	"net/http"
	"time"
)

var _ RateLimiter = (*BasicRateLimiter)(nil)

type BasicRateLimiter struct {
	RequestDelay time.Duration
	lastRequest  time.Time
}

func (r *BasicRateLimiter) AddRequest(_ *http.Request, t time.Time) {
	r.lastRequest = t
}

func (r *BasicRateLimiter) GetBackoffAt(_ *http.Request, t time.Time) time.Duration {
	timeSinceLastRequest := t.Sub(r.lastRequest)
	if timeSinceLastRequest >= r.RequestDelay {
		return 0
	}

	return r.RequestDelay - timeSinceLastRequest
}
