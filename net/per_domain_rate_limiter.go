package net

import (
	"net/http"
	"strings"
	"time"
)

var _ RateLimiter = (*PerDomainRateLimiter)(nil)

type PerDomainRateLimiter struct {
	DomainRateLimiters map[string]RateLimiter
	DefaultRateLimiter RateLimiter
}

func (r *PerDomainRateLimiter) AddRequest(req *http.Request, t time.Time) {
	rateLimiter := r.getRateLimiter(req)
	if rateLimiter == nil {
		return
	}

	rateLimiter.AddRequest(req, t)
}

func (r *PerDomainRateLimiter) GetBackoffAt(req *http.Request, t time.Time) time.Duration {
	rateLimiter := r.getRateLimiter(req)
	if rateLimiter == nil {
		return 0
	}

	return rateLimiter.GetBackoffAt(req, t)
}

func (r *PerDomainRateLimiter) getRateLimiter(req *http.Request) RateLimiter {
	if req.URL == nil {
		return nil
	}

	hostname := req.URL.Hostname()
	for domain, rateLimiter := range r.DomainRateLimiters {
		if hostname == domain || strings.HasSuffix(hostname, "."+domain) {
			return rateLimiter
		}
	}
	return r.DefaultRateLimiter
}
