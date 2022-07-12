package net_test

import (
	"github.com/mdelillo/go-utils/net"
	"net/http"
	"time"
)

var _ net.RateLimiter = (*mockRateLimiter)(nil)

type mockRateLimiter struct {
	getBackoffCallCount int
	getBackoffReturn    time.Duration
	addRequestCallCount int
}

func (r *mockRateLimiter) AddRequest(_ *http.Request, _ time.Time) {
	r.addRequestCallCount++
}

func (r *mockRateLimiter) GetBackoffAt(_ *http.Request, _ time.Time) time.Duration {
	r.getBackoffCallCount++
	return r.getBackoffReturn
}
