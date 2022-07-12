package net

import (
	"math"
	"net/http"
	"time"
)

var _ Retrier = (*ExponentialBackoffRetrier)(nil)

type ExponentialBackoffRetrier struct {
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	MaxAttempts    int
}

func (r ExponentialBackoffRetrier) ShouldRetry(_ *http.Request, resp *http.Response, attempts int) (bool, time.Duration) {
	if resp.StatusCode < 500 || attempts >= r.MaxAttempts {
		return false, 0
	}

	backoff := r.InitialBackoff
	if backoff == 0 {
		backoff = 100 * time.Millisecond
	}

	backoff *= time.Duration(int(math.Pow(2, float64(attempts-1))))

	if r.MaxBackoff != 0 {
		backoff = time.Duration(math.Min(float64(backoff), float64(r.MaxBackoff)))
	}

	return true, backoff
}
