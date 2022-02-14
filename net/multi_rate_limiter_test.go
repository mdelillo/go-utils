package net_test

import (
	"github.com/mdelillo/go-utils/net"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestMultiRateLimiter(t *testing.T) {
	spec.Run(t, "Multi Rate Limiter", testMultiRateLimiter, spec.Report(report.Terminal{}))
}

func testMultiRateLimiter(t *testing.T, context spec.G, it spec.S) {
	context("AddRequest", func() {
		it("calls AddRequest on every rate limiter", func() {
			rateLimiter1 := &mockRateLimiter{}
			rateLimiter2 := &mockRateLimiter{}
			rateLimiter3 := &mockRateLimiter{}

			rateLimiter := net.MultiRateLimiter{
				RateLimiters: []net.RateLimiter{rateLimiter1, rateLimiter2, rateLimiter3},
			}

			rateLimiter.AddRequest(nil, time.Now())
			assert.Equal(t, 1, rateLimiter1.addRequestCallCount)
			assert.Equal(t, 1, rateLimiter2.addRequestCallCount)
			assert.Equal(t, 1, rateLimiter3.addRequestCallCount)

			rateLimiter.AddRequest(nil, time.Now())
			assert.Equal(t, 2, rateLimiter1.addRequestCallCount)
			assert.Equal(t, 2, rateLimiter2.addRequestCallCount)
			assert.Equal(t, 2, rateLimiter3.addRequestCallCount)
		})
	})

	context("GetBackoffAt", func() {
		it("gets the largest backoff of all the rate limiters", func() {
			rateLimiter := net.MultiRateLimiter{
				RateLimiters: []net.RateLimiter{
					&mockRateLimiter{getBackoffReturn: time.Second},
					&mockRateLimiter{getBackoffReturn: time.Hour},
					&mockRateLimiter{getBackoffReturn: 0},
					&mockRateLimiter{getBackoffReturn: time.Minute},
				},
			}

			assert.Equal(t, time.Hour, rateLimiter.GetBackoffAt(nil, time.Now()))
		})

		it("returns 0 if there are no rate limiters", func() {
			rateLimiter := net.MultiRateLimiter{}
			assert.Equal(t, 0*time.Second, rateLimiter.GetBackoffAt(nil, time.Now()))
		})
	})
}
