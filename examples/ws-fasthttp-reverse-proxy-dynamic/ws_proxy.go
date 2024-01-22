package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/fasthttp/websocket"
	"github.com/valyala/fasthttp"

	proxy "github.com/yeqown/fasthttp-reverse-proxy/v2"
)

var (
	proxyServer *proxy.WSReverseProxy
	once        sync.Once
	mainServer  = fmt.Sprintf("ws://localhost:8080")
)

// ProxyHandler ... fasthttp.RequestHandler func
func ProxyHandler(ctx *fasthttp.RequestCtx) {
	once.Do(func() {
		var err error
		proxyServer, err = proxy.NewWSReverseProxyWith(
			proxy.WithURL_OptionWS(mainServer),
			proxy.WithUpgrader_OptionWS(&websocket.FastHTTPUpgrader{
				CheckOrigin: func(r *fasthttp.RequestCtx) bool { return true },
			}),
		)
		if err != nil {
			panic(err)
		}
	})

	// Delete query arg to not forward it to main server
	ctx.QueryArgs().Del("q")

	// Process
	fruit := ctx.QueryArgs().Peek("fruit")
	if len(fruit) == 0 {
		fruit = []byte("oranges")
	}

	switch string(ctx.Path()) {
	case "/echo":
		ctx.QueryArgs().Set("whoami", "proxy_server")
		ctx.Request.Header.Set("Override-Path", fmt.Sprintf("/talk_about/%s", fruit))
		proxyServer.ServeHTTP(ctx)
	case "/":
		fasthttp.ServeFileUncompressed(ctx, "./index.html")
	default:
		ctx.Error("Unsupported path", fasthttp.StatusNotFound)
	}
}

func main() {
	port := 8081
	log.Println("Starting proxy server on:", port)
	if err := fasthttp.ListenAndServe(fmt.Sprintf(":%d", port), ProxyHandler); err != nil {
		log.Fatal(err)
	}
}
