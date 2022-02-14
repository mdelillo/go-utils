package net

import (
	"net/http"
	"time"
)

var _ RateLimiter = (*MultiRateLimiter)(nil)

type MultiRateLimiter struct {
	RateLimiters []RateLimiter
}

func (r *MultiRateLimiter) AddRequest(req *http.Request, t time.Time) {
	for _, rateLimiter := range r.RateLimiters {
		rateLimiter.AddRequest(req, t)
	}
}

func (r *MultiRateLimiter) GetBackoffAt(req *http.Request, t time.Time) time.Duration {
	var largestBackoff time.Duration
	for _, rateLimiter := range r.RateLimiters {
		backoff := rateLimiter.GetBackoffAt(req, t)
		if largestBackoff < backoff {
			largestBackoff = backoff
		}
	}
	return largestBackoff
}
