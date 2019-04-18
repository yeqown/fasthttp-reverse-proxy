package main

import (
	"log"

	"github.com/valyala/fasthttp"
	proxy "github.com/yeqown/fasthttp-reverse-proxy"
)

var (
	proxyServer = proxy.NewReverseProxy("localhost:8080")
)

// ProxyHandler ... fasthttp.RequestHandler func
func ProxyHandler(ctx *fasthttp.RequestCtx) {
	// all proxy to localhost
	proxyServer.ServeHTTP(ctx)
}

func main() {
	if err := fasthttp.ListenAndServe(":8081", ProxyHandler); err != nil {
		log.Fatal(err)
	}
}
