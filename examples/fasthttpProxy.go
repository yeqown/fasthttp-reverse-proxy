package main

import (
	"github.com/valyala/fasthttp"
	proxy "github.com/yeqown/fasthttp-reverse-proxy"
)

var (
	proxyServer = proxy.NewReverseProxy("localhost:8080")
)

// ProxyHandler ...
func ProxyHandler(ctx *fasthttp.RequestCtx) {
	// all proxy to localhost
	proxyServer.ServeHTTP(ctx)
}

func main() {
	fasthttp.ListenAndServe(":8081", ProxyHandler)
}
