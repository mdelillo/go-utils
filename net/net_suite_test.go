package net_test

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

type testServerHandler struct {
	RequestCount int
}

func (t *testServerHandler) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t.RequestCount++

		switch r.URL.Path {
		case "/show-request":
			_ = r.Write(w)
		case "/set-cookies":
			http.SetCookie(w, &http.Cookie{Name: "some-cookie", Value: "some-value"})
			http.SetCookie(w, &http.Cookie{Name: "some-other-cookie", Value: "some-other-value"})
		case "/500":
			w.WriteHeader(http.StatusInternalServerError)
		case "/":
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

func newGetRequest(t *testing.T, domain string) *http.Request {
	t.Helper()

	url := fmt.Sprintf("https://%s/some-path", domain)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)
	return req
}
