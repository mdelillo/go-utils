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

func (m *mockRateLimiter) AddRequest(_ *http.Request, _ time.Time) {
	m.addRequestCallCount++
}

func (m *mockRateLimiter) GetBackoffAt(_ *http.Request, _ time.Time) time.Duration {
	m.getBackoffCallCount++
	return m.getBackoffReturn
}
