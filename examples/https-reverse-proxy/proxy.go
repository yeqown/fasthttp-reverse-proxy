package main

import (
	"log"

	"github.com/valyala/fasthttp"

	proxy "github.com/yeqown/fasthttp-reverse-proxy/v2"
)

var (
	proxyServer, _ = proxy.NewReverseProxyWith(
		proxy.WithAddress("localhost:8080"),
		proxy.WithTLS("./selfsigned.crt", "./selfsigned.key"),
	)
)

// ProxyHandler ... fasthttp.RequestHandler func
func ProxyHandler(ctx *fasthttp.RequestCtx) {
	requestURI := string(ctx.RequestURI())
	log.Printf("a request incoming, requestURI=%s\n", requestURI)
	proxyServer.ServeHTTP(ctx)
}

func main() {
	if err := fasthttp.ListenAndServe("0.0.0.0:8081", ProxyHandler); err != nil {
		log.Fatal(err)
	}
}
