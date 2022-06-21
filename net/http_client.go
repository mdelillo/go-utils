package net

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

const (
	defaultTimeout             = time.Minute
	defaultDialTimeout         = 5 * time.Second
	defaultTLSHandshakeTimeout = 5 * time.Second
)

type ClientOption func(h *http.Client)

func WithTimeout(timeout time.Duration) func(h *http.Client) {
	return func(h *http.Client) {
		h.Timeout = timeout
	}
}

func WithTLSHandshakeTimeout(timeout time.Duration) func(h *http.Client) {
	return func(h *http.Client) {
		h.Transport.(*http.Transport).TLSHandshakeTimeout = timeout
	}
}

func WithDialTimeout(timeout time.Duration) func(h *http.Client) {
	return func(h *http.Client) {
		h.Transport.(*http.Transport).DialContext = (&net.Dialer{
			Timeout: timeout,
		}).DialContext
	}
}

func WithTLSClientConfig(tlsConfig *tls.Config) func(h *http.Client) {
	return func(h *http.Client) {
		h.Transport.(*http.Transport).TLSClientConfig = tlsConfig
	}
}

func WithCookieJar(jar http.CookieJar) func(h *http.Client) {
	return func(h *http.Client) {
		h.Jar = jar
	}
}

func NewHTTPClient(options ...ClientOption) *http.Client {
	client := &http.Client{
		Timeout: defaultTimeout,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: defaultDialTimeout,
			}).DialContext,
			TLSHandshakeTimeout: defaultTLSHandshakeTimeout,
		},
	}

	for _, option := range options {
		option(client)
	}

	return client
}
