package net_test

import (
	"encoding/json"
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
	"testing"
)

func TestPersistableCookieJar(t *testing.T) {
	spec.Run(t, "PersistableCookieJar", testPersistableCookieJar, spec.Report(report.Terminal{}))
}

func testPersistableCookieJar(t *testing.T, context spec.G, it spec.S) {
	var server *httptest.Server

	it.Before(func() {
		server = httptest.NewServer(testServerHandler)
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

		cookies := jar.Export()

		serializedCookies, err := json.Marshal(cookies)
		require.NoError(t, err)

		var deserializedCookies map[string]map[string]net.JarEntry
		err = json.Unmarshal(serializedCookies, &deserializedCookies)

		newJar := net.NewPersistableCookieJar(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
		newClient := http.Client{Jar: jar}

		newJar.Import(deserializedCookies)

		resp, err = newClient.Get(server.URL + "/show-request")
		require.NoError(t, err)
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.Contains(t, string(body), "Cookie: some-cookie=some-value; some-other-cookie=some-other-value")
	})
}
