package net

import (
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

type BrowserOption func(*Browser)

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

func WithRequestDelay(delay time.Duration) func(*Browser) {
	return func(b *Browser) {
		b.RequestDelay = delay
	}
}

func WithCookieJar(jar *cookiejar.Jar) func(*Browser) {
	return func(b *Browser) {
		b.Client.Jar = jar
	}
}

type Browser struct {
	Client          *http.Client
	Headers         map[string]string
	RequestDelay    time.Duration
	lastRequestTime time.Time
}

func NewBrowser(options ...BrowserOption) *Browser {
	browser := &Browser{Client: NewHTTPClient()}

	for _, option := range options {
		option(browser)
	}

	return browser
}

func (b *Browser) CloseIdleConnections() {
	if b.Client != nil {
		b.Client.CloseIdleConnections()
	}
}

func (b *Browser) Do(req *http.Request) (*http.Response, error) {
	b.ensureClient()
	b.setHeaders(req)
	b.waitRequestDelay()
	defer b.setLastRequestTime()

	return b.Client.Do(req)
}

func (b *Browser) Get(url string) (resp *http.Response, err error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	return b.Do(req)
}

func (b *Browser) Head(url string) (resp *http.Response, err error) {
	req, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		return nil, err
	}

	return b.Do(req)
}

func (b *Browser) Post(url, contentType string, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	b.setHeaders(req)
	req.Header.Set("Content-Type", contentType)

	return b.Client.Do(req)
}

func (b *Browser) PostForm(url string, data url.Values) (resp *http.Response, err error) {
	return b.Post(url, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
}

func (b *Browser) ensureClient() {
	if b.Client == nil {
		b.Client = http.DefaultClient
	}
}

func (b *Browser) setHeaders(req *http.Request) {
	for name, value := range b.Headers {
		req.Header.Set(name, value)
	}
}

func (b *Browser) waitRequestDelay() {
	if b.RequestDelay == 0 {
		return
	}

	timeSinceLastRequest := time.Now().Sub(b.lastRequestTime)
	if timeSinceLastRequest < b.RequestDelay {
		time.Sleep(b.RequestDelay - timeSinceLastRequest)
	}
}

func (b *Browser) setLastRequestTime() {
	if b.RequestDelay == 0 {
		return
	}

	b.lastRequestTime = time.Now()
}
