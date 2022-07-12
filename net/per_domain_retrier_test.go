package net_test

import (
	"github.com/mdelillo/go-utils/net"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

func TestPerDomainRetrier(t *testing.T) {
	spec.Run(t, "Per Domain Retrier", testPerDomainRetrier, spec.Report(report.Terminal{}))
}

func testPerDomainRetrier(t *testing.T, when spec.G, it spec.S) {
	it("calls ShouldRetry on the appropriate retrier for the domain", func() {
		someDomainRetrier := &mockRetrier{shouldRetryReturn1: time.Second}
		someOtherDomainRetrier := &mockRetrier{shouldRetryReturn1: 2 * time.Second}
		someDefaultRetrier := &mockRetrier{shouldRetryReturn1: 3 * time.Second}

		retrier := net.PerDomainRetrier{
			DomainRetriers: map[string]net.Retrier{
				"some-domain.com":       someDomainRetrier,
				"some-other-domain.com": someOtherDomainRetrier,
			},
			DefaultRetrier: someDefaultRetrier,
		}

		_, backoff := retrier.ShouldRetry(newGetRequest(t, "some-domain.com"), nil, 1)
		assert.Equal(t, time.Second, backoff)
		assert.Equal(t, 1, someDomainRetrier.shouldRetryCallCount)
		assert.Equal(t, 0, someOtherDomainRetrier.shouldRetryCallCount)
		assert.Equal(t, 0, someDefaultRetrier.shouldRetryCallCount)

		_, backoff = retrier.ShouldRetry(newGetRequest(t, "subdomain.some-domain.com"), nil, 1)
		assert.Equal(t, time.Second, backoff)
		assert.Equal(t, 2, someDomainRetrier.shouldRetryCallCount)
		assert.Equal(t, 0, someOtherDomainRetrier.shouldRetryCallCount)
		assert.Equal(t, 0, someDefaultRetrier.shouldRetryCallCount)

		_, backoff = retrier.ShouldRetry(newGetRequest(t, "some-other-domain.com"), nil, 1)
		assert.Equal(t, 2*time.Second, backoff)
		assert.Equal(t, 2, someDomainRetrier.shouldRetryCallCount)
		assert.Equal(t, 1, someOtherDomainRetrier.shouldRetryCallCount)
		assert.Equal(t, 0, someDefaultRetrier.shouldRetryCallCount)

		_, backoff = retrier.ShouldRetry(newGetRequest(t, "some-unknown-domain.com"), nil, 1)
		assert.Equal(t, 3*time.Second, backoff)
		assert.Equal(t, 2, someDomainRetrier.shouldRetryCallCount)
		assert.Equal(t, 1, someOtherDomainRetrier.shouldRetryCallCount)
		assert.Equal(t, 1, someDefaultRetrier.shouldRetryCallCount)

		_, backoff = retrier.ShouldRetry(newGetRequest(t, "not-some-domain.com"), nil, 1)
		assert.Equal(t, 3*time.Second, backoff)
		assert.Equal(t, 2, someDomainRetrier.shouldRetryCallCount)
		assert.Equal(t, 1, someOtherDomainRetrier.shouldRetryCallCount)
		assert.Equal(t, 2, someDefaultRetrier.shouldRetryCallCount)
	})

	it("returns false if the default retrier is empty", func() {
		retrier := net.PerDomainRetrier{}
		shouldRetry, _ := retrier.ShouldRetry(newGetRequest(t, "some-domain.com"), nil, 1)
		assert.False(t, shouldRetry)
	})

	it("returns false if the url is empty", func() {
		retrier := net.PerDomainRetrier{}
		shouldRetry, _ := retrier.ShouldRetry(&http.Request{}, nil, 1)
		assert.False(t, shouldRetry)
	})
}
