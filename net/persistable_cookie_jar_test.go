package net_test

import (
	"encoding/json"
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
)

func TestPersistableCookieJar(t *testing.T) {
	spec.Run(t, "PersistableCookieJar", testPersistableCookieJar, spec.Report(report.Terminal{}))
}

func testPersistableCookieJar(t *testing.T, context spec.G, it spec.S) {
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

	it("can export and import cookies", func() {
		jar := net.NewPersistableCookieJar(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
		client := http.Client{Jar: jar}

		resp, err := client.Get(server.URL + "/set-cookies")
		require.NoError(t, err)
		require.NoError(t, resp.Body.Close())

		serverURL, err := url.Parse(server.URL)
		require.NoError(t, err)

		assert.Len(t, jar.Cookies(serverURL), 2)

		cookies := jar.Export()

		serializedCookies, err := json.Marshal(cookies)
		require.NoError(t, err)

		var deserializedCookies map[string]map[string]net.JarEntry
		err = json.Unmarshal(serializedCookies, &deserializedCookies)

		newJar := net.NewPersistableCookieJar(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
		newClient := http.Client{Jar: jar}

		newJar.Import(deserializedCookies)

		resp, err = newClient.Get(server.URL + "/get-cookies")
		require.NoError(t, err)
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.Contains(t, string(body), "some-cookie: some-value")
		assert.Contains(t, string(body), "some-other-cookie: some-other-value")
	})
}
