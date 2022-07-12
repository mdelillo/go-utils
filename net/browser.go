package net

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type BrowserOption func(*Browser)

type RateLimiter interface {
	AddRequest(req *http.Request, t time.Time)
	GetBackoffAt(req *http.Request, t time.Time) time.Duration
}

type Retrier interface {
	ShouldRetry(req *http.Request, resp *http.Response, attempt int) (retry bool, backoff time.Duration)
}

type Browser struct {
	Client      *http.Client
	Headers     map[string]string
	RateLimiter RateLimiter
	Retrier     Retrier
}

func NewBrowser(options ...BrowserOption) *Browser {
	browser := &Browser{Client: NewHTTPClient()}

	for _, option := range options {
		option(browser)
	}

	return browser
}

func WithClient(client *http.Client) func(*Browser) {
	return func(b *Browser) {
		b.Client = client
	}
}

func WithDefaultHeader(name, value string) func(*Browser) {
	return func(b *Browser) {
		if b.Headers == nil {
			b.Headers = map[string]string{}
		}
		b.Headers[name] = value
	}
}

func WithDefaultHeaders(headers map[string]string) func(*Browser) {
	return func(b *Browser) {
		if b.Headers == nil {
			b.Headers = map[string]string{}
		}
		for name, value := range headers {
			b.Headers[name] = value
		}
	}
}

func WithDefaultUserAgent(userAgent string) func(*Browser) {
	return func(b *Browser) {
		if b.Headers == nil {
			b.Headers = map[string]string{}
		}
		b.Headers["User-Agent"] = userAgent
	}
}

func WithDefaultRetrier(retrier Retrier) func(*Browser) {
	return func(b *Browser) {
		b.Retrier = retrier
	}
}

func WithDefaultRateLimiter(rateLimiter RateLimiter) func(*Browser) {
	return func(b *Browser) {
		b.RateLimiter = rateLimiter
	}
}

type requestOptions struct {
	rateLimiter RateLimiter
	retrier     Retrier
}

type RequestOption func(r *http.Request, opts *requestOptions)

func WithHeader(name, value string) func(r *http.Request, _ *requestOptions) {
	return func(r *http.Request, _ *requestOptions) {
		r.Header.Set(name, value)
	}
}

func WithHeaders(headers map[string]string) func(r *http.Request, _ *requestOptions) {
	return func(r *http.Request, _ *requestOptions) {
		for name, value := range headers {
			r.Header.Set(name, value)
		}
	}
}

func WithContentType(contentType string) func(r *http.Request, _ *requestOptions) {
	return func(r *http.Request, _ *requestOptions) {
		r.Header.Set("Content-Type", contentType)
	}
}

func WithAccept(accept string) func(r *http.Request, _ *requestOptions) {
	return func(r *http.Request, _ *requestOptions) {
		r.Header.Set("Accept", accept)
	}
}

func WithReferer(referer string) func(r *http.Request, _ *requestOptions) {
	return func(r *http.Request, _ *requestOptions) {
		r.Header.Set("Referer", referer)
	}
}

func WithBasicAuth(username, password string) func(r *http.Request, _ *requestOptions) {
	return func(r *http.Request, _ *requestOptions) {
		r.SetBasicAuth(username, password)
	}
}

func WithCookie(cookie *http.Cookie) func(r *http.Request, _ *requestOptions) {
	return func(r *http.Request, _ *requestOptions) {
		r.AddCookie(cookie)
	}
}

func WithRateLimiter(rateLimiter RateLimiter) func(_ *http.Request, opts *requestOptions) {
	return func(_ *http.Request, opts *requestOptions) {
		opts.rateLimiter = rateLimiter
	}
}

func WithRetrier(retrier Retrier) func(_ *http.Request, opts *requestOptions) {
	return func(_ *http.Request, opts *requestOptions) {
		opts.retrier = retrier
	}
}

func (b *Browser) Do(req *http.Request, options ...RequestOption) (*http.Response, error) {
	b.ensureClient()
	b.setHeaders(req)

	opts := &requestOptions{
		rateLimiter: b.RateLimiter,
		retrier:     b.Retrier,
	}

	for _, option := range options {
		option(req, opts)
	}

	var attempt int
	for {
		attempt++

		resp, err := b.doWithRateLimiter(req, opts.rateLimiter)
		if err != nil {
			return resp, err
		}

		if opts.retrier != nil {
			shouldRetry, backoff := opts.retrier.ShouldRetry(req, resp, attempt)
			if shouldRetry {
				time.Sleep(backoff)
				continue
			}
		}

		return resp, nil
	}
}

func (b *Browser) doWithRateLimiter(req *http.Request, rateLimiter RateLimiter) (*http.Response, error) {
	if rateLimiter != nil {
		time.Sleep(rateLimiter.GetBackoffAt(req, time.Now()))
		defer rateLimiter.AddRequest(req, time.Now())
	}

	return b.Client.Do(req)
}

func (b *Browser) Get(url string, options ...RequestOption) (resp *http.Response, err error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	return b.Do(req, options...)
}

func (b *Browser) Head(url string, options ...RequestOption) (resp *http.Response, err error) {
	req, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		return nil, err
	}

	return b.Do(req, options...)
}

func (b *Browser) Post(url, contentType string, body io.Reader, options ...RequestOption) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	b.setHeaders(req)
	req.Header.Set("Content-Type", contentType)

	return b.Do(req, options...)
}

func (b *Browser) PostForm(url string, data url.Values, options ...RequestOption) (resp *http.Response, err error) {
	return b.Post(url, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()), options...)
}

func (b *Browser) ensureClient() {
	if b.Client == nil {
		b.Client = NewHTTPClient()
	}
}

func (b *Browser) setHeaders(req *http.Request) {
	for name, value := range b.Headers {
		req.Header.Set(name, value)
	}
}
