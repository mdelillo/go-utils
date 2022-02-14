package net_test

import (
	"github.com/mdelillo/go-utils/net"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestBasicRateLimiter(t *testing.T) {
	spec.Run(t, "Basic Rate Limiter", testBasicRateLimiter, spec.Report(report.Terminal{}))
}

func testBasicRateLimiter(t *testing.T, when spec.G, it spec.S) {
	it("does not allow more than one request within the given delay", func() {
		rateLimiter := net.BasicRateLimiter{RequestDelay: 3 * time.Second}
		t0 := time.Now()

		assert.Equal(t, 0*time.Second, rateLimiter.GetBackoffAt(nil, t0))

		rateLimiter.AddRequest(nil, t0.Add(time.Second))
		assert.Equal(t, 3*time.Second, rateLimiter.GetBackoffAt(nil, t0.Add(time.Second)))
		assert.Equal(t, 2*time.Second, rateLimiter.GetBackoffAt(nil, t0.Add(2*time.Second)))
		assert.Equal(t, 1*time.Second, rateLimiter.GetBackoffAt(nil, t0.Add(3*time.Second)))
		assert.Equal(t, 0*time.Second, rateLimiter.GetBackoffAt(nil, t0.Add(4*time.Second)))
	})
}
