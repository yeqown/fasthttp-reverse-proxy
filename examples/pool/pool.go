package main

import (
	"log"

	"github.com/valyala/fasthttp"
	proxy "github.com/yeqown/fasthttp-reverse-proxy"
)

var (
	pool proxy.Pool
	err  error
)

func init() {
	pool, err = proxy.NewChanPool(100, 200,
		func(addr string) (*proxy.ReverseProxy, error) {
			p := proxy.NewReverseProxy(addr)
			return p, nil
		})
}

// ProxyPoolHandler ...
func ProxyPoolHandler(ctx *fasthttp.RequestCtx) {
	proxyServer, err := pool.Get("localhost:8080")
	if err != nil {
		log.Println(err)
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}
	defer pool.Put(proxyServer)
	// all proxy to localhost
	proxyServer.ServeHTTP(ctx)
}

func main() {
	if err := fasthttp.ListenAndServe(":8083", ProxyPoolHandler); err != nil {
		panic(err)
	}
}
