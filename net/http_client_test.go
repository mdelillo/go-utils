package net_test

import (
	"crypto/tls"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/publicsuffix"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/mdelillo/go-utils/net"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"github.com/stretchr/testify/assert"
)

func TestHTTPClient(t *testing.T) {
	spec.Run(t, "HTTPClient", testHTTPClient, spec.Report(report.Terminal{}))
}

func testHTTPClient(t *testing.T, context spec.G, it spec.S) {
	var server *httptest.Server

	it.Before(func() {
		server = httptest.NewServer(testServerHandler)
	})

	it.After(func() {
		server.Close()
	})

	context("NewHTTPClient", func() {
		it("creates an HTTP client with default config", func() {
			client := net.NewHTTPClient()

			resp, err := client.Get(server.URL + "/show-request")
			require.NoError(t, err)
			_ = resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, time.Minute, client.Timeout)
			assert.Equal(t, 5*time.Second, client.Transport.(*http.Transport).TLSHandshakeTimeout)
			assert.NotNil(t, client.Transport.(*http.Transport).DialContext)
		})

		context("WithTimeout", func() {
			it("overwrites the default timeout", func() {
				client := net.NewHTTPClient(net.WithTimeout(1 * time.Second))
				assert.Equal(t, 1*time.Second, client.Timeout)
			})
		})

		context("WithTLSHandshakeTimeout", func() {
			it("overwrites the default timeout", func() {
				client := net.NewHTTPClient(net.WithTLSHandshakeTimeout(1 * time.Second))
				assert.Equal(t, 1*time.Second, client.Transport.(*http.Transport).TLSHandshakeTimeout)
			})
		})

		context("WithTLSClientConfig", func() {
			it("uses the TLS config", func() {
				tlsConfig := &tls.Config{ServerName: "some-server-name"}
				client := net.NewHTTPClient(net.WithTLSClientConfig(tlsConfig))
				assert.Equal(t, tlsConfig, client.Transport.(*http.Transport).TLSClientConfig)
			})
		})

		context("WithCookieJar", func() {
			it("uses the given cookie jar with each request", func() {
				jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
				require.NoError(t, err)

				client := net.NewHTTPClient(net.WithCookieJar(jar))

				resp, err := client.Post(server.URL+"/set-cookies", "", nil)
				require.NoError(t, err)
				require.NoError(t, resp.Body.Close())

				serverURL, err := url.Parse(server.URL)
				require.NoError(t, err)

				cookies := jar.Cookies(serverURL)
				assert.Len(t, cookies, 2)

				resp, err = client.Get(server.URL + "/show-request")
				require.NoError(t, err)
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				require.NoError(t, err)

				assert.Contains(t, string(body), "Cookie: some-cookie=some-value; some-other-cookie=some-other-value")
			})
		})
	})
}
