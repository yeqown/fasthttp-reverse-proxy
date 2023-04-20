package main

import (
	"fmt"
	"log"

	"github.com/valyala/fasthttp"

	proxy "github.com/yeqown/fasthttp-reverse-proxy/v2"
)

var (
	weights = map[string]proxy.Weight{
		"localhost:9090": 20,
		"localhost:9091": 30,
		"localhost:9092": 50,
	}

	proxyServer, _ = proxy.NewReverseProxyWith(proxy.WithBalancer(weights), proxy.WithDebug())
)

// ProxyHandler ... fasthttp.RequestHandler func
func ProxyHandler(ctx *fasthttp.RequestCtx) {
	// all proxy to localhost
	proxyServer.ServeHTTP(ctx)
}

func main() {
	fmt.Printf("listening on :8081\n")
	if err := fasthttp.ListenAndServe(":8081", ProxyHandler); err != nil {
		log.Fatal(err)
	}
}
