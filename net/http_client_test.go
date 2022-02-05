package net_test

import (
	"crypto/tls"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
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
		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		}))
	})

	it.After(func() {
		server.Close()
	})

	context("NewHTTPClient", func() {
		it("creates an HTTPClient with default config", func() {
			client := net.NewHTTPClient()

			response, err := client.Get(server.URL)
			require.NoError(t, err)
			defer response.Body.Close()

			assert.Equal(t, http.StatusOK, response.StatusCode)
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

		context("WithTLSCLientConfig", func() {
			it("uses the TLS config", func() {
				tlsConfig := &tls.Config{ServerName: "some-server-name"}
				client := net.NewHTTPClient(net.WithTLSClientConfig(tlsConfig))
				assert.Equal(t, tlsConfig, client.Transport.(*http.Transport).TLSClientConfig)
			})
		})
	})
}
