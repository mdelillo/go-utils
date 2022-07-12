package net_test

import (
	"github.com/mdelillo/go-utils/net"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	assertpkg "github.com/stretchr/testify/assert"
	requirepkg "github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBrowser(t *testing.T) {
	spec.Run(t, "Browser", testBrowser, spec.Report(report.Terminal{}))
}

func testBrowser(t *testing.T, context spec.G, it spec.S) {
	var (
		server  *httptest.Server
		handler *testServerHandler

		assert  = assertpkg.New(t)
		require = requirepkg.New(t)
	)

	it.Before(func() {
		handler = &testServerHandler{}
		server = httptest.NewServer(handler.Handler())
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

				req, err := http.NewRequest(http.MethodGet, server.URL+"/show-request", nil)
				require.NoError(err)

				doResp, err := browser.Do(req)
				require.NoError(err)
				defer doResp.Body.Close()

				doBody, err := ioutil.ReadAll(doResp.Body)
				require.NoError(err)

				assert.Contains(string(doBody), "Some-Header: some-value")
				assert.Contains(string(doBody), "Some-Other-Header: some-other-value")

				getResp, err := browser.Get(server.URL + "/show-request")
				require.NoError(err)
				defer getResp.Body.Close()

				getBody, err := ioutil.ReadAll(getResp.Body)
				require.NoError(err)

				assert.Contains(string(getBody), "Some-Header: some-value")
				assert.Contains(string(getBody), "Some-Other-Header: some-other-value")

				postResp, err := browser.Post(server.URL+"/show-request", "", nil)
				require.NoError(err)
				defer postResp.Body.Close()

				postBody, err := ioutil.ReadAll(postResp.Body)
				require.NoError(err)

				assert.Contains(string(postBody), "Some-Header: some-value")
				assert.Contains(string(postBody), "Some-Other-Header: some-other-value")

				postFormResp, err := browser.PostForm(server.URL+"/show-request", nil)
				require.NoError(err)
				defer postFormResp.Body.Close()

				postFormBody, err := ioutil.ReadAll(postFormResp.Body)
				require.NoError(err)

				assert.Contains(string(postFormBody), "Some-Header: some-value")
				assert.Contains(string(postFormBody), "Some-Other-Header: some-other-value")
			})
		})

		context("WithDefaultHeaders", func() {
			it("sets the headers on every request", func() {
				browser := net.NewBrowser(net.WithDefaultHeaders(map[string]string{
					"some-header":       "some-value",
					"some-other-header": "some-other-value",
				}))

				req, err := http.NewRequest(http.MethodGet, server.URL+"/show-request", nil)
				require.NoError(err)

				doResp, err := browser.Do(req)
				require.NoError(err)
				defer doResp.Body.Close()

				doBody, err := ioutil.ReadAll(doResp.Body)
				require.NoError(err)

				assert.Contains(string(doBody), "Some-Header: some-value")
				assert.Contains(string(doBody), "Some-Other-Header: some-other-value")

				getResp, err := browser.Get(server.URL + "/show-request")
				require.NoError(err)
				defer getResp.Body.Close()

				getBody, err := ioutil.ReadAll(getResp.Body)
				require.NoError(err)

				assert.Contains(string(getBody), "Some-Header: some-value")
				assert.Contains(string(getBody), "Some-Other-Header: some-other-value")

				postResp, err := browser.Post(server.URL+"/show-request", "", nil)
				require.NoError(err)
				defer postResp.Body.Close()

				postBody, err := ioutil.ReadAll(postResp.Body)
				require.NoError(err)

				assert.Contains(string(postBody), "Some-Header: some-value")
				assert.Contains(string(postBody), "Some-Other-Header: some-other-value")

				postFormResp, err := browser.PostForm(server.URL+"/show-request", nil)
				require.NoError(err)
				defer postFormResp.Body.Close()

				postFormBody, err := ioutil.ReadAll(postFormResp.Body)
				require.NoError(err)

				assert.Contains(string(postFormBody), "Some-Header: some-value")
				assert.Contains(string(postFormBody), "Some-Other-Header: some-other-value")
			})
		})

		context("WithDefaultUserAgent", func() {
			it("sets the user-agent header on every request", func() {
				browser := net.NewBrowser(net.WithDefaultUserAgent("some-user-agent"))

				req, err := http.NewRequest(http.MethodGet, server.URL+"/show-request", nil)
				require.NoError(err)

				doResp, err := browser.Do(req)
				require.NoError(err)
				defer doResp.Body.Close()

				doBody, err := ioutil.ReadAll(doResp.Body)
				require.NoError(err)

				assert.Contains(string(doBody), "User-Agent: some-user-agent")

				getResp, err := browser.Get(server.URL + "/show-request")
				require.NoError(err)
				defer getResp.Body.Close()

				getBody, err := ioutil.ReadAll(getResp.Body)
				require.NoError(err)

				assert.Contains(string(getBody), "User-Agent: some-user-agent")

				postResp, err := browser.Post(server.URL+"/show-request", "", nil)
				require.NoError(err)
				defer postResp.Body.Close()

				postBody, err := ioutil.ReadAll(postResp.Body)
				require.NoError(err)

				assert.Contains(string(postBody), "User-Agent: some-user-agent")

				postFormResp, err := browser.PostForm(server.URL+"/show-request", nil)
				require.NoError(err)
				defer postFormResp.Body.Close()

				postFormBody, err := ioutil.ReadAll(postFormResp.Body)
				require.NoError(err)

				assert.Contains(string(postFormBody), "User-Agent: some-user-agent")
			})
		})

		context("WithDefaultRateLimiter", func() {
			it("uses the rate limiter to wait between requests", func() {
				rateLimiter := &mockRateLimiter{}
				browser := net.NewBrowser(net.WithDefaultRateLimiter(rateLimiter))

				_, err := browser.Get(server.URL)
				require.NoError(err)
				assert.Equal(1, rateLimiter.getBackoffCallCount)
				assert.Equal(1, rateLimiter.addRequestCallCount)

				_, err = browser.Head(server.URL)
				require.NoError(err)
				assert.Equal(2, rateLimiter.getBackoffCallCount)
				assert.Equal(2, rateLimiter.addRequestCallCount)

				_, err = browser.Post(server.URL, "", nil)
				require.NoError(err)
				assert.Equal(3, rateLimiter.getBackoffCallCount)
				assert.Equal(3, rateLimiter.addRequestCallCount)

				_, err = browser.PostForm(server.URL, nil)
				require.NoError(err)
				assert.Equal(4, rateLimiter.getBackoffCallCount)
				assert.Equal(4, rateLimiter.addRequestCallCount)

				req, err := http.NewRequest(http.MethodGet, server.URL, nil)
				require.NoError(err)
				_, err = browser.Do(req)
				require.NoError(err)
				assert.Equal(5, rateLimiter.getBackoffCallCount)
				assert.Equal(5, rateLimiter.addRequestCallCount)
			})
		})

		context("WithDefaultRetrier", func() {
			it("uses the retrier to determine whether to retry the request", func() {
				browser := net.NewBrowser(net.WithDefaultRetrier(net.ExponentialBackoffRetrier{MaxAttempts: 3}))

				_, err := browser.Get(server.URL + "/500")
				require.NoError(err)
				assert.Equal(3, handler.RequestCount)

				_, err = browser.Head(server.URL + "/500")
				require.NoError(err)
				assert.Equal(6, handler.RequestCount)

				_, err = browser.Post(server.URL+"/500", "", nil)
				require.NoError(err)
				assert.Equal(9, handler.RequestCount)

				_, err = browser.PostForm(server.URL+"/500", nil)
				require.NoError(err)
				assert.Equal(12, handler.RequestCount)

				req, err := http.NewRequest(http.MethodGet, server.URL+"/500", nil)
				require.NoError(err)
				_, err = browser.Do(req)
				require.NoError(err)
				assert.Equal(15, handler.RequestCount)
			})
		})
	})

	context("Do", func() {
		context("WithHeader", func() {
			it("sets the header, overwriting any default headers", func() {
				browser := net.NewBrowser(
					net.WithDefaultHeader("some-header", "some-value"),
					net.WithDefaultHeader("some-other-header", "some-other-value"),
				)

				req, err := http.NewRequest(http.MethodGet, server.URL+"/show-request", nil)
				require.NoError(err)

				resp, err := browser.Do(
					req,
					net.WithHeader("some-other-header", "some-new-value"),
					net.WithHeader("some-third-header", "some-third-value"),
				)
				require.NoError(err)
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				require.NoError(err)

				assert.Contains(string(body), "Some-Header: some-value")
				assert.Contains(string(body), "Some-Other-Header: some-new-value")
				assert.Contains(string(body), "Some-Third-Header: some-third-value")
			})
		})

		context("WithHeaders", func() {
			it("sets the header, overwriting any default headers", func() {
				browser := net.NewBrowser(net.WithDefaultHeaders(map[string]string{
					"some-header":       "some-value",
					"some-other-header": "some-other-value",
				}))

				req, err := http.NewRequest(http.MethodGet, server.URL+"/show-request", nil)
				require.NoError(err)

				resp, err := browser.Do(req, net.WithHeaders(map[string]string{
					"some-other-header": "some-new-value",
					"some-third-header": "some-third-value",
				}))
				require.NoError(err)
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				require.NoError(err)

				assert.Contains(string(body), "Some-Header: some-value")
				assert.Contains(string(body), "Some-Other-Header: some-new-value")
				assert.Contains(string(body), "Some-Third-Header: some-third-value")
			})
		})

		context("WithContentType", func() {
			it("sets the content-type header", func() {
				browser := net.NewBrowser()

				req, err := http.NewRequest(http.MethodGet, server.URL+"/show-request", nil)
				require.NoError(err)

				resp, err := browser.Do(req, net.WithContentType("some-content-type"))
				require.NoError(err)
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				require.NoError(err)

				assert.Contains(string(body), "Content-Type: some-content-type")
			})
		})

		context("WithAccept", func() {
			it("sets the accept header", func() {
				browser := net.NewBrowser()

				req, err := http.NewRequest(http.MethodGet, server.URL+"/show-request", nil)
				require.NoError(err)

				resp, err := browser.Do(req, net.WithAccept("some-accept"))
				require.NoError(err)
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				require.NoError(err)

				assert.Contains(string(body), "Accept: some-accept")
			})
		})

		context("WithReferer", func() {
			it("sets the referer header", func() {
				browser := net.NewBrowser()

				req, err := http.NewRequest(http.MethodGet, server.URL+"/show-request", nil)
				require.NoError(err)

				resp, err := browser.Do(req, net.WithReferer("some-referer"))
				require.NoError(err)
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				require.NoError(err)

				assert.Contains(string(body), "Referer: some-referer")
			})
		})

		context("WithBasicAuth", func() {
			it("sets basic auth on the request", func() {
				browser := net.NewBrowser()

				req, err := http.NewRequest(http.MethodGet, server.URL+"/show-request", nil)
				require.NoError(err)

				resp, err := browser.Do(req, net.WithBasicAuth("some-username", "some-password"))
				require.NoError(err)
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				require.NoError(err)

				assert.Contains(string(body), "Authorization: Basic c29tZS11c2VybmFtZTpzb21lLXBhc3N3b3Jk")
			})
		})

		context("WithCookie", func() {
			it("sets the cookie", func() {
				browser := net.NewBrowser()

				req, err := http.NewRequest(http.MethodGet, server.URL+"/show-request", nil)
				require.NoError(err)

				cookie := &http.Cookie{Name: "some-name", Value: "some-value"}

				resp, err := browser.Do(req, net.WithCookie(cookie))
				require.NoError(err)
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				require.NoError(err)

				assert.Contains(string(body), "Cookie: some-name=some-value")
			})
		})

		context("WithRateLimiter", func() {
			it("uses the rate limiter to wait between requests", func() {
				defaultRateLimiter := &mockRateLimiter{}
				browser := net.NewBrowser(net.WithDefaultRateLimiter(defaultRateLimiter))

				rateLimiter := &mockRateLimiter{}

				req, err := http.NewRequest(http.MethodGet, server.URL, nil)
				require.NoError(err)
				_, err = browser.Do(req, net.WithRateLimiter(rateLimiter))
				require.NoError(err)
				assert.Equal(1, rateLimiter.getBackoffCallCount)
				assert.Equal(1, rateLimiter.addRequestCallCount)

				assert.Equal(0, defaultRateLimiter.getBackoffCallCount)
				assert.Equal(0, defaultRateLimiter.addRequestCallCount)
			})
		})

		context("WithRetrier", func() {
			it("uses the retrier for the request", func() {
				browser := net.NewBrowser(net.WithDefaultRetrier(net.ExponentialBackoffRetrier{MaxAttempts: 2}))

				req, err := http.NewRequest(http.MethodGet, server.URL+"/500", nil)
				require.NoError(err)
				_, err = browser.Do(req, net.WithRetrier(net.ExponentialBackoffRetrier{MaxAttempts: 3}))
				require.NoError(err)
				assert.Equal(3, handler.RequestCount)

				req, err = http.NewRequest(http.MethodGet, server.URL+"/500", nil)
				require.NoError(err)
				_, err = browser.Do(req)
				require.NoError(err)
				assert.Equal(5, handler.RequestCount)

				req, err = http.NewRequest(http.MethodGet, server.URL+"/500", nil)
				require.NoError(err)
				_, err = browser.Do(req, net.WithRetrier(nil))
				require.NoError(err)
				assert.Equal(6, handler.RequestCount)
			})
		})
	})

	context("Get", func() {
		context("WithHeader", func() {
			it("sets the header, overwriting any default headers", func() {
				browser := net.NewBrowser(
					net.WithDefaultHeader("some-header", "some-value"),
					net.WithDefaultHeader("some-other-header", "some-other-value"),
				)

				resp, err := browser.Get(
					server.URL+"/show-request",
					net.WithHeader("some-other-header", "some-new-value"),
					net.WithHeader("some-third-header", "some-third-value"),
				)
				require.NoError(err)
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				require.NoError(err)

				assert.Contains(string(body), "Some-Header: some-value")
				assert.Contains(string(body), "Some-Other-Header: some-new-value")
				assert.Contains(string(body), "Some-Third-Header: some-third-value")
			})
		})

		context("WithHeaders", func() {
			it("sets the header, overwriting any default headers", func() {
				browser := net.NewBrowser(net.WithDefaultHeaders(map[string]string{
					"some-header":       "some-value",
					"some-other-header": "some-other-value",
				}))

				resp, err := browser.Get(server.URL+"/show-request", net.WithHeaders(map[string]string{
					"some-other-header": "some-new-value",
					"some-third-header": "some-third-value",
				}))
				require.NoError(err)
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				require.NoError(err)

				assert.Contains(string(body), "Some-Header: some-value")
				assert.Contains(string(body), "Some-Other-Header: some-new-value")
				assert.Contains(string(body), "Some-Third-Header: some-third-value")
			})
		})

		context("WithContentType", func() {
			it("sets the content-type header", func() {
				browser := net.NewBrowser()

				resp, err := browser.Get(server.URL+"/show-request", net.WithContentType("some-content-type"))
				require.NoError(err)
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				require.NoError(err)

				assert.Contains(string(body), "Content-Type: some-content-type")
			})
		})

		context("WithAccept", func() {
			it("sets the accept header", func() {
				browser := net.NewBrowser()

				resp, err := browser.Get(server.URL+"/show-request", net.WithAccept("some-accept"))
				require.NoError(err)
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				require.NoError(err)

				assert.Contains(string(body), "Accept: some-accept")
			})
		})

		context("WithReferer", func() {
			it("sets the referer header", func() {
				browser := net.NewBrowser()

				resp, err := browser.Get(server.URL+"/show-request", net.WithReferer("some-referer"))
				require.NoError(err)
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				require.NoError(err)

				assert.Contains(string(body), "Referer: some-referer")
			})
		})

		context("WithBasicAuth", func() {
			it("sets basic auth on the request", func() {
				browser := net.NewBrowser()

				resp, err := browser.Get(server.URL+"/show-request", net.WithBasicAuth("some-username", "some-password"))
				require.NoError(err)
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				require.NoError(err)

				assert.Contains(string(body), "Authorization: Basic c29tZS11c2VybmFtZTpzb21lLXBhc3N3b3Jk")
			})
		})

		context("WithCookie", func() {
			it("sets the cookie", func() {
				browser := net.NewBrowser()

				cookie := &http.Cookie{Name: "some-name", Value: "some-value"}

				resp, err := browser.Get(server.URL+"/show-request", net.WithCookie(cookie))
				require.NoError(err)
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				require.NoError(err)

				assert.Contains(string(body), "Cookie: some-name=some-value")
			})
		})

		context("WithRateLimiter", func() {
			it("uses the rate limiter to wait between requests", func() {
				defaultRateLimiter := &mockRateLimiter{}
				browser := net.NewBrowser(net.WithDefaultRateLimiter(defaultRateLimiter))

				rateLimiter := &mockRateLimiter{}

				_, err := browser.Get(server.URL, net.WithRateLimiter(rateLimiter))
				require.NoError(err)
				assert.Equal(1, rateLimiter.getBackoffCallCount)
				assert.Equal(1, rateLimiter.addRequestCallCount)

				assert.Equal(0, defaultRateLimiter.getBackoffCallCount)
				assert.Equal(0, defaultRateLimiter.addRequestCallCount)
			})
		})

		context("WithRetrier", func() {
			it("uses the retrier for the request", func() {
				browser := net.NewBrowser(net.WithDefaultRetrier(net.ExponentialBackoffRetrier{MaxAttempts: 2}))

				_, err := browser.Get(server.URL+"/500", net.WithRetrier(net.ExponentialBackoffRetrier{MaxAttempts: 3}))
				require.NoError(err)
				assert.Equal(3, handler.RequestCount)

				_, err = browser.Get(server.URL + "/500")
				require.NoError(err)
				assert.Equal(5, handler.RequestCount)

				_, err = browser.Get(server.URL+"/500", net.WithRetrier(nil))
				require.NoError(err)
				assert.Equal(6, handler.RequestCount)
			})
		})
	})

	context("Head", func() {
		context("WithRateLimiter", func() {
			it("uses the rate limiter to wait between requests", func() {
				defaultRateLimiter := &mockRateLimiter{}
				browser := net.NewBrowser(net.WithDefaultRateLimiter(defaultRateLimiter))

				rateLimiter := &mockRateLimiter{}

				_, err := browser.Head(server.URL, net.WithRateLimiter(rateLimiter))
				require.NoError(err)
				assert.Equal(1, rateLimiter.getBackoffCallCount)
				assert.Equal(1, rateLimiter.addRequestCallCount)

				assert.Equal(0, defaultRateLimiter.getBackoffCallCount)
				assert.Equal(0, defaultRateLimiter.addRequestCallCount)
			})
		})

		context("WithRetrier", func() {
			it("uses the retrier for the request", func() {
				browser := net.NewBrowser(net.WithDefaultRetrier(net.ExponentialBackoffRetrier{MaxAttempts: 2}))

				_, err := browser.Head(server.URL+"/500", net.WithRetrier(net.ExponentialBackoffRetrier{MaxAttempts: 3}))
				require.NoError(err)
				assert.Equal(3, handler.RequestCount)

				_, err = browser.Head(server.URL + "/500")
				require.NoError(err)
				assert.Equal(5, handler.RequestCount)

				_, err = browser.Head(server.URL+"/500", net.WithRetrier(nil))
				require.NoError(err)
				assert.Equal(6, handler.RequestCount)
			})
		})
	})

	context("Post", func() {
		context("WithHeader", func() {
			it("sets the header, overwriting any default headers", func() {
				browser := net.NewBrowser(
					net.WithDefaultHeader("some-header", "some-value"),
					net.WithDefaultHeader("some-other-header", "some-other-value"),
				)

				resp, err := browser.Post(
					server.URL+"/show-request", "", nil,
					net.WithHeader("some-other-header", "some-new-value"),
					net.WithHeader("some-third-header", "some-third-value"),
				)
				require.NoError(err)
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				require.NoError(err)

				assert.Contains(string(body), "Some-Header: some-value")
				assert.Contains(string(body), "Some-Other-Header: some-new-value")
				assert.Contains(string(body), "Some-Third-Header: some-third-value")
			})
		})

		context("WithHeaders", func() {
			it("sets the header, overwriting any default headers", func() {
				browser := net.NewBrowser(net.WithDefaultHeaders(map[string]string{
					"some-header":       "some-value",
					"some-other-header": "some-other-value",
				}))

				resp, err := browser.Post(server.URL+"/show-request", "", nil, net.WithHeaders(map[string]string{
					"some-other-header": "some-new-value",
					"some-third-header": "some-third-value",
				}))
				require.NoError(err)
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				require.NoError(err)

				assert.Contains(string(body), "Some-Header: some-value")
				assert.Contains(string(body), "Some-Other-Header: some-new-value")
				assert.Contains(string(body), "Some-Third-Header: some-third-value")
			})
		})

		context("WithContentType", func() {
			it("sets the content-type header", func() {
				browser := net.NewBrowser()

				resp, err := browser.Post(server.URL+"/show-request", "some-other-content-type", nil, net.WithContentType("some-content-type"))
				require.NoError(err)
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				require.NoError(err)

				assert.Contains(string(body), "Content-Type: some-content-type")
			})
		})

		context("WithAccept", func() {
			it("sets the accept header", func() {
				browser := net.NewBrowser()

				resp, err := browser.Post(server.URL+"/show-request", "", nil, net.WithAccept("some-accept"))
				require.NoError(err)
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				require.NoError(err)

				assert.Contains(string(body), "Accept: some-accept")
			})
		})

		context("WithReferer", func() {
			it("sets the referer header", func() {
				browser := net.NewBrowser()

				resp, err := browser.Post(server.URL+"/show-request", "", nil, net.WithReferer("some-referer"))
				require.NoError(err)
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				require.NoError(err)

				assert.Contains(string(body), "Referer: some-referer")
			})
		})

		context("WithBasicAuth", func() {
			it("sets basic auth on the request", func() {
				browser := net.NewBrowser()

				resp, err := browser.Post(server.URL+"/show-request", "", nil, net.WithBasicAuth("some-username", "some-password"))
				require.NoError(err)
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				require.NoError(err)

				assert.Contains(string(body), "Authorization: Basic c29tZS11c2VybmFtZTpzb21lLXBhc3N3b3Jk")
			})
		})

		context("WithCookie", func() {
			it("sets the cookie", func() {
				browser := net.NewBrowser()

				cookie := &http.Cookie{Name: "some-name", Value: "some-value"}

				resp, err := browser.Post(server.URL+"/show-request", "", nil, net.WithCookie(cookie))
				require.NoError(err)
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				require.NoError(err)

				assert.Contains(string(body), "Cookie: some-name=some-value")
			})
		})

		context("WithRateLimiter", func() {
			it("uses the rate limiter to wait between requests", func() {
				defaultRateLimiter := &mockRateLimiter{}
				browser := net.NewBrowser(net.WithDefaultRateLimiter(defaultRateLimiter))

				rateLimiter := &mockRateLimiter{}

				_, err := browser.Post(server.URL, "", nil, net.WithRateLimiter(rateLimiter))
				require.NoError(err)
				assert.Equal(1, rateLimiter.getBackoffCallCount)
				assert.Equal(1, rateLimiter.addRequestCallCount)

				assert.Equal(0, defaultRateLimiter.getBackoffCallCount)
				assert.Equal(0, defaultRateLimiter.addRequestCallCount)
			})
		})

		context("WithRetrier", func() {
			it("uses the retrier for the request", func() {
				browser := net.NewBrowser(net.WithDefaultRetrier(net.ExponentialBackoffRetrier{MaxAttempts: 2}))

				_, err := browser.Post(server.URL+"/500", "", nil, net.WithRetrier(net.ExponentialBackoffRetrier{MaxAttempts: 3}))
				require.NoError(err)
				assert.Equal(3, handler.RequestCount)

				_, err = browser.Post(server.URL+"/500", "", nil)
				require.NoError(err)
				assert.Equal(5, handler.RequestCount)

				_, err = browser.Post(server.URL+"/500", "", nil, net.WithRetrier(nil))
				require.NoError(err)
				assert.Equal(6, handler.RequestCount)
			})
		})
	})

	context("PostForm", func() {
		context("WithHeader", func() {
			it("sets the header, overwriting any default headers", func() {
				browser := net.NewBrowser(
					net.WithDefaultHeader("some-header", "some-value"),
					net.WithDefaultHeader("some-other-header", "some-other-value"),
				)

				resp, err := browser.PostForm(
					server.URL+"/show-request", nil,
					net.WithHeader("some-other-header", "some-new-value"),
					net.WithHeader("some-third-header", "some-third-value"),
				)
				require.NoError(err)
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				require.NoError(err)

				assert.Contains(string(body), "Some-Header: some-value")
				assert.Contains(string(body), "Some-Other-Header: some-new-value")
				assert.Contains(string(body), "Some-Third-Header: some-third-value")
			})
		})

		context("WithHeaders", func() {
			it("sets the header, overwriting any default headers", func() {
				browser := net.NewBrowser(net.WithDefaultHeaders(map[string]string{
					"some-header":       "some-value",
					"some-other-header": "some-other-value",
				}))

				resp, err := browser.PostForm(server.URL+"/show-request", nil, net.WithHeaders(map[string]string{
					"some-other-header": "some-new-value",
					"some-third-header": "some-third-value",
				}))
				require.NoError(err)
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				require.NoError(err)

				assert.Contains(string(body), "Some-Header: some-value")
				assert.Contains(string(body), "Some-Other-Header: some-new-value")
				assert.Contains(string(body), "Some-Third-Header: some-third-value")
			})
		})

		context("WithContentType", func() {
			it("sets the content-type header", func() {
				browser := net.NewBrowser()

				resp, err := browser.PostForm(server.URL+"/show-request", nil, net.WithContentType("some-content-type"))
				require.NoError(err)
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				require.NoError(err)

				assert.Contains(string(body), "Content-Type: some-content-type")
			})
		})

		context("WithAccept", func() {
			it("sets the accept header", func() {
				browser := net.NewBrowser()

				resp, err := browser.PostForm(server.URL+"/show-request", nil, net.WithAccept("some-accept"))
				require.NoError(err)
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				require.NoError(err)

				assert.Contains(string(body), "Accept: some-accept")
			})
		})

		context("WithReferer", func() {
			it("sets the referer header", func() {
				browser := net.NewBrowser()

				resp, err := browser.PostForm(server.URL+"/show-request", nil, net.WithReferer("some-referer"))
				require.NoError(err)
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				require.NoError(err)

				assert.Contains(string(body), "Referer: some-referer")
			})
		})

		context("WithBasicAuth", func() {
			it("sets basic auth on the request", func() {
				browser := net.NewBrowser()

				resp, err := browser.PostForm(server.URL+"/show-request", nil, net.WithBasicAuth("some-username", "some-password"))
				require.NoError(err)
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				require.NoError(err)

				assert.Contains(string(body), "Authorization: Basic c29tZS11c2VybmFtZTpzb21lLXBhc3N3b3Jk")
			})
		})

		context("WithCookie", func() {
			it("sets the cookie", func() {
				browser := net.NewBrowser()

				cookie := &http.Cookie{Name: "some-name", Value: "some-value"}

				resp, err := browser.PostForm(server.URL+"/show-request", nil, net.WithCookie(cookie))
				require.NoError(err)
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				require.NoError(err)

				assert.Contains(string(body), "Cookie: some-name=some-value")
			})
		})

		context("WithRateLimiter", func() {
			it("uses the rate limiter to wait between requests", func() {
				defaultRateLimiter := &mockRateLimiter{}
				browser := net.NewBrowser(net.WithDefaultRateLimiter(defaultRateLimiter))

				rateLimiter := &mockRateLimiter{}

				_, err := browser.PostForm(server.URL, nil, net.WithRateLimiter(rateLimiter))
				require.NoError(err)
				assert.Equal(1, rateLimiter.getBackoffCallCount)
				assert.Equal(1, rateLimiter.addRequestCallCount)

				assert.Equal(0, defaultRateLimiter.getBackoffCallCount)
				assert.Equal(0, defaultRateLimiter.addRequestCallCount)
			})
		})

		context("WithRetrier", func() {
			it("uses the retrier for the request", func() {
				browser := net.NewBrowser(net.WithDefaultRetrier(net.ExponentialBackoffRetrier{MaxAttempts: 2}))

				_, err := browser.PostForm(server.URL+"/500", nil, net.WithRetrier(net.ExponentialBackoffRetrier{MaxAttempts: 3}))
				require.NoError(err)
				assert.Equal(3, handler.RequestCount)

				_, err = browser.PostForm(server.URL+"/500", nil)
				require.NoError(err)
				assert.Equal(5, handler.RequestCount)

				_, err = browser.PostForm(server.URL+"/500", nil, net.WithRetrier(nil))
				require.NoError(err)
				assert.Equal(6, handler.RequestCount)
			})
		})
	})
}
