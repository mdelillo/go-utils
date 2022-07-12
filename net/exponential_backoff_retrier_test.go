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

func TestExponentialBackoffRetrier(t *testing.T) {
	spec.Run(t, "Exponential Backoff Retrier", testExponentialBackoffRetrier, spec.Report(report.Terminal{}))
}

func testExponentialBackoffRetrier(t *testing.T, when spec.G, it spec.S) {
	it("returns true when the status code is 5XX", func() {
		retrier := net.ExponentialBackoffRetrier{MaxAttempts: 2}

		for _, testCase := range []struct {
			statusCode  int
			shouldRetry bool
		}{
			{statusCode: 200, shouldRetry: false},
			{statusCode: 301, shouldRetry: false},
			{statusCode: 400, shouldRetry: false},
			{statusCode: 404, shouldRetry: false},
			{statusCode: 500, shouldRetry: true},
			{statusCode: 503, shouldRetry: true},
			{statusCode: 599, shouldRetry: true},
		} {
			shouldRetry, _ := retrier.ShouldRetry(nil, &http.Response{StatusCode: testCase.statusCode}, 1)
			assert.Equal(t, testCase.shouldRetry, shouldRetry, "status code: %d", testCase.statusCode)
		}
	})

	it("returns true when attempts is less than MaxAttempts", func() {
		retrier := net.ExponentialBackoffRetrier{MaxAttempts: 10}

		for _, testCase := range []struct {
			attempts    int
			shouldRetry bool
		}{
			{attempts: 0, shouldRetry: true},
			{attempts: 8, shouldRetry: true},
			{attempts: 9, shouldRetry: true},
			{attempts: 10, shouldRetry: false},
			{attempts: 11, shouldRetry: false},
		} {
			shouldRetry, _ := retrier.ShouldRetry(nil, &http.Response{StatusCode: http.StatusInternalServerError}, testCase.attempts)
			assert.Equal(t, testCase.shouldRetry, shouldRetry, "attempts: %d", testCase.attempts)
		}
	})

	it("exponentially increases the backoff with each attempt", func() {
		retrier := net.ExponentialBackoffRetrier{InitialBackoff: 2 * time.Millisecond, MaxAttempts: 5}

		_, backoff := retrier.ShouldRetry(nil, &http.Response{StatusCode: http.StatusInternalServerError}, 1)
		assert.Equal(t, 2*time.Millisecond, backoff)

		_, backoff = retrier.ShouldRetry(nil, &http.Response{StatusCode: http.StatusInternalServerError}, 2)
		assert.Equal(t, 4*time.Millisecond, backoff)

		_, backoff = retrier.ShouldRetry(nil, &http.Response{StatusCode: http.StatusInternalServerError}, 3)
		assert.Equal(t, 8*time.Millisecond, backoff)

		_, backoff = retrier.ShouldRetry(nil, &http.Response{StatusCode: http.StatusInternalServerError}, 4)
		assert.Equal(t, 16*time.Millisecond, backoff)
	})

	it("does not exceed the max backoff", func() {
		retrier := net.ExponentialBackoffRetrier{InitialBackoff: 2 * time.Millisecond, MaxBackoff: 8 * time.Millisecond, MaxAttempts: 5}

		_, backoff := retrier.ShouldRetry(nil, &http.Response{StatusCode: http.StatusInternalServerError}, 1)
		assert.Equal(t, 2*time.Millisecond, backoff)

		_, backoff = retrier.ShouldRetry(nil, &http.Response{StatusCode: http.StatusInternalServerError}, 2)
		assert.Equal(t, 4*time.Millisecond, backoff)

		_, backoff = retrier.ShouldRetry(nil, &http.Response{StatusCode: http.StatusInternalServerError}, 3)
		assert.Equal(t, 8*time.Millisecond, backoff)

		_, backoff = retrier.ShouldRetry(nil, &http.Response{StatusCode: http.StatusInternalServerError}, 4)
		assert.Equal(t, 8*time.Millisecond, backoff)
	})

	it("defaults to a 100ms backoff", func() {
		retrier := net.ExponentialBackoffRetrier{MaxAttempts: 5}

		_, backoff := retrier.ShouldRetry(nil, &http.Response{StatusCode: http.StatusInternalServerError}, 1)
		assert.Equal(t, 100*time.Millisecond, backoff)

		_, backoff = retrier.ShouldRetry(nil, &http.Response{StatusCode: http.StatusInternalServerError}, 2)
		assert.Equal(t, 200*time.Millisecond, backoff)

		_, backoff = retrier.ShouldRetry(nil, &http.Response{StatusCode: http.StatusInternalServerError}, 3)
		assert.Equal(t, 400*time.Millisecond, backoff)

		_, backoff = retrier.ShouldRetry(nil, &http.Response{StatusCode: http.StatusInternalServerError}, 4)
		assert.Equal(t, 800*time.Millisecond, backoff)
	})
}
