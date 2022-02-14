package net_test

import (
	"github.com/mdelillo/go-utils/net"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRollingWindowRateLimiter(t *testing.T) {
	spec.Run(t, "Rolling Window Rate Limiter", testRollingWindowRateLimiter, spec.Report(report.Terminal{}))
}

func testRollingWindowRateLimiter(t *testing.T, when spec.G, it spec.S) {
	it("uses a configurable rolling window to limit requests", func() {
		rateLimiter := net.RollingWindowRateLimiter{Window: 5 * time.Second, RequestLimit: 3}
		t0 := time.Now()

		assert.Equal(t, 0*time.Second, rateLimiter.GetBackoffAt(nil, t0))

		rateLimiter.AddRequest(nil, t0)
		assert.Equal(t, 0*time.Second, rateLimiter.GetBackoffAt(nil, t0))

		rateLimiter.AddRequest(nil, t0)
		rateLimiter.AddRequest(nil, t0)
		assert.Equal(t, 5*time.Second, rateLimiter.GetBackoffAt(nil, t0))
		assert.Equal(t, 4*time.Second, rateLimiter.GetBackoffAt(nil, t0.Add(time.Second)))
		assert.Equal(t, 1*time.Second, rateLimiter.GetBackoffAt(nil, t0.Add(4*time.Second)))
		assert.Equal(t, 0*time.Second, rateLimiter.GetBackoffAt(nil, t0.Add(5*time.Second)))

		rateLimiter.AddRequest(nil, t0.Add(5*time.Second))
		rateLimiter.AddRequest(nil, t0.Add(6*time.Second))
		assert.Equal(t, 0*time.Second, rateLimiter.GetBackoffAt(nil, t0.Add(7*time.Second)))

		rateLimiter.AddRequest(nil, t0.Add(7*time.Second))
		assert.Equal(t, 3*time.Second, rateLimiter.GetBackoffAt(nil, t0.Add(7*time.Second)))
		assert.Equal(t, 2*time.Second, rateLimiter.GetBackoffAt(nil, t0.Add(8*time.Second)))
		assert.Equal(t, 1*time.Second, rateLimiter.GetBackoffAt(nil, t0.Add(9*time.Second)))
		assert.Equal(t, 0*time.Second, rateLimiter.GetBackoffAt(nil, t0.Add(10*time.Second)))
	})
}
