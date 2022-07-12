package net_test

import (
	"github.com/mdelillo/go-utils/net"
	"net/http"
	"time"
)

var _ net.Retrier = (*mockRetrier)(nil)

type mockRetrier struct {
	shouldRetryCallCount int
	shouldRetryReturn0   bool
	shouldRetryReturn1   time.Duration
}

func (r *mockRetrier) ShouldRetry(_ *http.Request, _ *http.Response, _ int) (bool, time.Duration) {
	r.shouldRetryCallCount++
	return r.shouldRetryReturn0, r.shouldRetryReturn1
}
