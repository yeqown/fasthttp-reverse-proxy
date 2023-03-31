package main

import (
	"log"
	"strings"
	"time"

	"github.com/valyala/fasthttp"

	proxy "github.com/yeqown/fasthttp-reverse-proxy/v2"
)

var (
	proxyServer, _  = proxy.NewReverseProxyWith(proxy.WithAddress("localhost:8080"), proxy.WithTimeout(5*time.Second))
	proxyServer2, _ = proxy.NewReverseProxyWith(proxy.WithAddress("api-js.mixpanel.com"))
	proxyServer3, _ = proxy.NewReverseProxyWith(proxy.WithAddress("baidu.com"))
)

// ProxyHandler ... fasthttp.RequestHandler func
func ProxyHandler(ctx *fasthttp.RequestCtx) {
	requestURI := string(ctx.RequestURI())
	log.Printf("requestURI=%s\n", requestURI)

	if strings.HasPrefix(requestURI, "/local") {
		// "/local" path proxy to localhost
		arr := strings.Split(requestURI, "?")
		if len(arr) > 1 {
			arr = append([]string{"/foo"}, arr[1:]...)
			requestURI = strings.Join(arr, "?")
		}

		ctx.Request.SetRequestURI(requestURI)
		proxyServer.ServeHTTP(ctx)
	} else if strings.HasPrefix(requestURI, "/baidu") {
		proxyServer3.ServeHTTP(ctx)
	} else {
		proxyServer2.ServeHTTP(ctx)
	}
}

func main() {
	if err := fasthttp.ListenAndServe(":8081", ProxyHandler); err != nil {
		log.Fatal(err)
	}
}
