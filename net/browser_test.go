package net_test

import (
	"fmt"
	"github.com/mdelillo/go-utils/net"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/publicsuffix"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestBrowser(t *testing.T) {
	spec.Run(t, "Browser", testBrowser, spec.Report(report.Terminal{}))
}

func testBrowser(t *testing.T, context spec.G, it spec.S) {
	var server *httptest.Server

	it.Before(func() {
		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/200":
				w.WriteHeader(http.StatusOK)
			case "/headers":
				for key, value := range r.Header {
					_, _ = fmt.Fprintf(w, "%s: %s\n", key, value)
				}
			case "/set-cookies":
				http.SetCookie(w, &http.Cookie{Name: "some-cookie", Value: "some-value"})
				http.SetCookie(w, &http.Cookie{Name: "some-other-cookie", Value: "some-other-value"})
			case "/get-cookies":
				for _, cookie := range r.Cookies() {
					_, _ = fmt.Fprintf(w, "%s: %s\n", cookie.Name, cookie.Value)
				}
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
	})

	it.After(func() {
		server.Close()
	})

	context("NewBrowser", func() {
		context("WithDefaultHeader", func() {
			it("sets the header on every request", func() {
				browser := net.NewBrowser(
					net.WithDefaultHeader("some-header", "some-value"),
					net.WithDefaultHeader("some-other-header", "some-other-value"),
				)

				getResp, err := browser.Get(server.URL + "/headers")
				require.NoError(t, err)
				defer getResp.Body.Close()

				getBody, err := ioutil.ReadAll(getResp.Body)
				require.NoError(t, err)

				assert.Contains(t, string(getBody), "Some-Header: [some-value]")
				assert.Contains(t, string(getBody), "Some-Other-Header: [some-other-value]")

				postResp, err := browser.Post(server.URL+"/headers", "", nil)
				require.NoError(t, err)
				defer postResp.Body.Close()

				postBody, err := ioutil.ReadAll(postResp.Body)
				require.NoError(t, err)

				assert.Contains(t, string(postBody), "Some-Header: [some-value]")
				assert.Contains(t, string(postBody), "Some-Other-Header: [some-other-value]")
			})

			it("overwrites previously set headers", func() {
				browser := net.NewBrowser(
					net.WithDefaultHeader("some-header", "some-value"),
					net.WithDefaultHeader("some-header", "some-new-value"),
				)

				resp, err := browser.Get(server.URL + "/headers")
				require.NoError(t, err)
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				require.NoError(t, err)

				assert.Contains(t, string(body), "Some-Header: [some-new-value]")
			})
		})

		context("WithDefaultHeaders", func() {
			it("sets the headers on every request", func() {
				browser := net.NewBrowser(net.WithDefaultHeaders(map[string]string{
					"some-header":       "some-value",
					"some-other-header": "some-other-value",
				}))

				resp, err := browser.Get(server.URL + "/headers")
				require.NoError(t, err)
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				require.NoError(t, err)

				assert.Contains(t, string(body), "Some-Header: [some-value]")
				assert.Contains(t, string(body), "Some-Other-Header: [some-other-value]")
			})
		})

		context("WithRequestDelay", func() {
			it("waits for delay between requests", func() {
				normalRequestTime := 2 * time.Millisecond
				delay := 20 * time.Millisecond
				browser := net.NewBrowser(net.WithRequestDelay(delay))

				t1 := time.Now()
				_, _ = browser.Get(server.URL + "/headers")
				t2 := time.Now()
				_, _ = browser.Get(server.URL + "/headers")
				t3 := time.Now()

				request1Time := t2.Sub(t1)
				request2Time := t3.Sub(t2)

				assert.Less(t, request1Time, normalRequestTime)
				assert.Greater(t, request2Time, delay)

				_, _ = browser.Get(server.URL + "/headers")
				time.Sleep(delay)
				t4 := time.Now()
				_, _ = browser.Get(server.URL + "/headers")
				t5 := time.Now()

				request3Time := t5.Sub(t4)
				assert.Less(t, request3Time, normalRequestTime)
			})
		})

		context("WithCookieJar", func() {
			it("uses the given cookie jar with each request", func() {
				jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
				require.NoError(t, err)

				browser := net.NewBrowser(net.WithCookieJar(jar))

				resp, err := browser.Post(server.URL+"/set-cookies", "", nil)
				require.NoError(t, err)
				require.NoError(t, resp.Body.Close())

				serverURL, err := url.Parse(server.URL)
				require.NoError(t, err)

				cookies := jar.Cookies(serverURL)
				assert.Len(t, cookies, 2)

				resp, err = browser.Get(server.URL + "/get-cookies")
				require.NoError(t, err)
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				require.NoError(t, err)

				assert.Contains(t, string(body), "some-cookie: some-value")
				assert.Contains(t, string(body), "some-other-cookie: some-other-value")
			})
		})
	})
}
