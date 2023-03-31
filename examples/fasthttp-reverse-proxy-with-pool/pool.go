package main

import (
	"log"

	"github.com/valyala/fasthttp"

	proxy "github.com/yeqown/fasthttp-reverse-proxy/v2"
)

var (
	pool proxy.Pool
)

// ProxyPoolHandler ...
func ProxyPoolHandler(ctx *fasthttp.RequestCtx) {
	proxyServer, err := pool.Get("localhost:9090")
	if err != nil {
		log.Println("ProxyPoolHandler got an error: ", err)
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}
	defer pool.Put(proxyServer)
	proxyServer.ServeHTTP(ctx)
}

func factory(hostAddr string) (*proxy.ReverseProxy, error) {
	return proxy.NewReverseProxyWith(proxy.WithAddress(hostAddr))
}

func main() {
	initialCap, maxCap := 100, 1000
	var err error
	pool, err = proxy.NewChanPool(initialCap, maxCap, factory)
	if err != nil {
		panic(err)
	}
	if err = fasthttp.ListenAndServe(":8083", ProxyPoolHandler); err != nil {
		panic(err)
	}
}
