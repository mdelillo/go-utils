package net

import (
	"net/http"
	"strings"
	"time"
)

var _ Retrier = (*PerDomainRetrier)(nil)

type PerDomainRetrier struct {
	DomainRetriers map[string]Retrier
	DefaultRetrier Retrier
}

func (r PerDomainRetrier) ShouldRetry(req *http.Request, resp *http.Response, attempts int) (bool, time.Duration) {
	retrier := r.getRetrier(req)
	if retrier == nil {
		return false, 0
	}

	return retrier.ShouldRetry(req, resp, attempts)
}

func (r *PerDomainRetrier) getRetrier(req *http.Request) Retrier {
	if req.URL == nil {
		return nil
	}

	hostname := req.URL.Hostname()
	for domain, retrier := range r.DomainRetriers {
		if hostname == domain || strings.HasSuffix(hostname, "."+domain) {
			return retrier
		}
	}
	return r.DefaultRetrier
}
