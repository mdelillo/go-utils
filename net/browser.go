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

type Browser struct {
	Client      *http.Client
	Headers     map[string]string
	RateLimiter RateLimiter
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

func WithRateLimiter(rateLimiter RateLimiter) func(*Browser) {
	return func(b *Browser) {
		b.RateLimiter = rateLimiter
	}
}

type RequestOption func(r *http.Request)

func WithHeader(name, value string) func(r *http.Request) {
	return func(r *http.Request) {
		r.Header.Set(name, value)
	}
}

func WithHeaders(headers map[string]string) func(r *http.Request) {
	return func(r *http.Request) {
		for name, value := range headers {
			r.Header.Set(name, value)
		}
	}
}

func WithContentType(contentType string) func(r *http.Request) {
	return func(r *http.Request) {
		r.Header.Set("Content-Type", contentType)
	}
}

func WithAccept(accept string) func(r *http.Request) {
	return func(r *http.Request) {
		r.Header.Set("Accept", accept)
	}
}

func WithReferer(referer string) func(r *http.Request) {
	return func(r *http.Request) {
		r.Header.Set("Referer", referer)
	}
}

func WithBasicAuth(username, password string) func(r *http.Request) {
	return func(r *http.Request) {
		r.SetBasicAuth(username, password)
	}
}

func WithCookie(cookie *http.Cookie) func(r *http.Request) {
	return func(r *http.Request) {
		r.AddCookie(cookie)
	}
}

func (b *Browser) Do(req *http.Request, options ...RequestOption) (*http.Response, error) {
	b.ensureClient()
	b.setHeaders(req)
	b.rateLimit(req)
	defer b.notifyRateLimiter(req)

	for _, option := range options {
		option(req)
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

func (b *Browser) rateLimit(req *http.Request) {
	if b.RateLimiter != nil {
		time.Sleep(b.RateLimiter.GetBackoffAt(req, time.Now()))
	}
}

func (b *Browser) notifyRateLimiter(req *http.Request) {
	if b.RateLimiter != nil {
		b.RateLimiter.AddRequest(req, time.Now())
	}
}
