package net_test

import "net/http"

var testServerHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/show-request":
		_ = r.Write(w)
	case "/set-cookies":
		http.SetCookie(w, &http.Cookie{Name: "some-cookie", Value: "some-value"})
		http.SetCookie(w, &http.Cookie{Name: "some-other-cookie", Value: "some-other-value"})
	case "/":
	default:
		w.WriteHeader(http.StatusNotFound)
	}
})
