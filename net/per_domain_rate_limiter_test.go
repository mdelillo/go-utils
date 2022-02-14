package net_test

import (
	"fmt"
	"github.com/mdelillo/go-utils/net"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
	"time"
)

func TestPerDomainRateLimiter(t *testing.T) {
	spec.Run(t, "Per Domain Rate Limiter", testPerDomainRateLimiter, spec.Report(report.Terminal{}))
}

func testPerDomainRateLimiter(t *testing.T, context spec.G, it spec.S) {
	context("AddRequest", func() {
		it("calls AddRequest on the appropriate rate limiter for the domain", func() {
			someDomainRateLimiter := &mockRateLimiter{getBackoffReturn: time.Second}
			someOtherDomainRateLimiter := &mockRateLimiter{getBackoffReturn: 2 * time.Second}
			someDefaultRateLimiter := &mockRateLimiter{getBackoffReturn: 3 * time.Second}

			rateLimiter := net.PerDomainRateLimiter{
				DomainRateLimiters: map[string]net.RateLimiter{
					"some-domain.com":       someDomainRateLimiter,
					"some-other-domain.com": someOtherDomainRateLimiter,
				},
				DefaultRateLimiter: someDefaultRateLimiter,
			}

			rateLimiter.AddRequest(newGetRequest(t, "some-domain.com"), time.Now())
			assert.Equal(t, 1, someDomainRateLimiter.addRequestCallCount)
			assert.Equal(t, 0, someOtherDomainRateLimiter.addRequestCallCount)
			assert.Equal(t, 0, someDefaultRateLimiter.addRequestCallCount)

			rateLimiter.AddRequest(newGetRequest(t, "subdomain.some-domain.com"), time.Now())
			assert.Equal(t, 2, someDomainRateLimiter.addRequestCallCount)
			assert.Equal(t, 0, someOtherDomainRateLimiter.addRequestCallCount)
			assert.Equal(t, 0, someDefaultRateLimiter.addRequestCallCount)

			rateLimiter.AddRequest(newGetRequest(t, "some-other-domain.com"), time.Now())
			assert.Equal(t, 2, someDomainRateLimiter.addRequestCallCount)
			assert.Equal(t, 1, someOtherDomainRateLimiter.addRequestCallCount)
			assert.Equal(t, 0, someDefaultRateLimiter.addRequestCallCount)

			rateLimiter.AddRequest(newGetRequest(t, "some-unknown-domain.com"), time.Now())
			assert.Equal(t, 2, someDomainRateLimiter.addRequestCallCount)
			assert.Equal(t, 1, someOtherDomainRateLimiter.addRequestCallCount)
			assert.Equal(t, 1, someDefaultRateLimiter.addRequestCallCount)

			rateLimiter.AddRequest(newGetRequest(t, "not-some-domain.com"), time.Now())
			assert.Equal(t, 2, someDomainRateLimiter.addRequestCallCount)
			assert.Equal(t, 1, someOtherDomainRateLimiter.addRequestCallCount)
			assert.Equal(t, 2, someDefaultRateLimiter.addRequestCallCount)
		})

		it("does not panic if the default rate limiter is empty", func() {
			rateLimiter := net.PerDomainRateLimiter{}
			rateLimiter.AddRequest(newGetRequest(t, "some-domain.com"), time.Now())
		})

		it("does not panic if the url is empty", func() {
			rateLimiter := net.PerDomainRateLimiter{}
			rateLimiter.AddRequest(&http.Request{}, time.Now())
		})
	})

	context("GetBackoffAt", func() {
		it("calls GetBackoffAt on the appropriate rate limiter for the domain", func() {
			someDomainRateLimiter := &mockRateLimiter{getBackoffReturn: time.Second}
			someOtherDomainRateLimiter := &mockRateLimiter{getBackoffReturn: 2 * time.Second}
			someDefaultRateLimiter := &mockRateLimiter{getBackoffReturn: 3 * time.Second}

			rateLimiter := net.PerDomainRateLimiter{
				DomainRateLimiters: map[string]net.RateLimiter{
					"some-domain.com":       someDomainRateLimiter,
					"some-other-domain.com": someOtherDomainRateLimiter,
				},
				DefaultRateLimiter: someDefaultRateLimiter,
			}

			rateLimiter.GetBackoffAt(newGetRequest(t, "some-domain.com"), time.Now())
			assert.Equal(t, 1, someDomainRateLimiter.getBackoffCallCount)
			assert.Equal(t, 0, someOtherDomainRateLimiter.getBackoffCallCount)
			assert.Equal(t, 0, someDefaultRateLimiter.getBackoffCallCount)

			rateLimiter.GetBackoffAt(newGetRequest(t, "subdomain.some-domain.com"), time.Now())
			assert.Equal(t, 2, someDomainRateLimiter.getBackoffCallCount)
			assert.Equal(t, 0, someOtherDomainRateLimiter.getBackoffCallCount)
			assert.Equal(t, 0, someDefaultRateLimiter.getBackoffCallCount)

			rateLimiter.GetBackoffAt(newGetRequest(t, "some-other-domain.com"), time.Now())
			assert.Equal(t, 2, someDomainRateLimiter.getBackoffCallCount)
			assert.Equal(t, 1, someOtherDomainRateLimiter.getBackoffCallCount)
			assert.Equal(t, 0, someDefaultRateLimiter.getBackoffCallCount)

			rateLimiter.GetBackoffAt(newGetRequest(t, "some-unknown-domain.com"), time.Now())
			assert.Equal(t, 2, someDomainRateLimiter.getBackoffCallCount)
			assert.Equal(t, 1, someOtherDomainRateLimiter.getBackoffCallCount)
			assert.Equal(t, 1, someDefaultRateLimiter.getBackoffCallCount)

			rateLimiter.GetBackoffAt(newGetRequest(t, "not-some-domain.com"), time.Now())
			assert.Equal(t, 2, someDomainRateLimiter.getBackoffCallCount)
			assert.Equal(t, 1, someOtherDomainRateLimiter.getBackoffCallCount)
			assert.Equal(t, 2, someDefaultRateLimiter.getBackoffCallCount)
		})

		it("returns 0 if the default rate limiter is empty", func() {
			rateLimiter := net.PerDomainRateLimiter{}
			assert.Equal(t, 0*time.Second, rateLimiter.GetBackoffAt(newGetRequest(t, "some-domain.com"), time.Now()))
		})

		it("returns 0 if the url is empty", func() {
			rateLimiter := net.PerDomainRateLimiter{}
			assert.Equal(t, 0*time.Second, rateLimiter.GetBackoffAt(&http.Request{}, time.Now()))
		})
	})
}

func newGetRequest(t *testing.T, domain string) *http.Request {
	t.Helper()

	url := fmt.Sprintf("https://%s/some-path", domain)
	someDomainRequest, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)
	return someDomainRequest
}
