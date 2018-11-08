package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

var (
	URL   *url.URL
	proxy *httputil.ReverseProxy
)

func init() {
	URL, _ = url.Parse("http://localhost:8080")
	proxy = httputil.NewSingleHostReverseProxy(URL)
}

func main() {
	http.HandleFunc("/foo", func(w http.ResponseWriter, req *http.Request) {
		proxy.ServeHTTP(w, req)
	})

	if err := http.ListenAndServe(":8082", nil); err != nil {
		panic(err)
	}
}
