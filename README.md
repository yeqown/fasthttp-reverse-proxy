# fasthttp-reverse-proxy
reverse http proxy based on fasthttp

currently, it's so simple ~

### use it alone
```go
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
```

### use with pool 
> deleted in master branch, but I backup in old-with-pool. and here link to the [Code](https://github.com/valyala/fasthttp/blob/caea86794cef49a3c52a535fd7162c17b5b46640/server.go#L1511) of fasthttp pool
```go
// package main

// import (
// 	"log"

// 	"github.com/valyala/fasthttp"
// 	proxy "github.com/yeqown/fasthttp-reverse-proxy"
// )

// var (
// 	pool proxy.Pool
// 	err  error
// )

// func init() {
// 	pool, err = proxy.NewChanPool(10, 100,
// 		func(addr string) (*proxy.ReverseProxy, error) {
// 			p := proxy.NewReverseProxy(addr)
// 			return p, nil
// 		})
// }

// // ProxyPoolHandler ...
// func ProxyPoolHandler(ctx *fasthttp.RequestCtx) {
// 	proxyServer, err := pool.Get("localhost:8080")
// 	if err != nil {
// 		log.Println(err)
// 		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
// 		return
// 	}
// 	defer pool.Put(proxyServer)
// 	// all proxy to localhost
// 	proxyServer.ServeHTTP(ctx)
// }

// func main() {
// 	if err := fasthttp.ListenAndServe(":8083", ProxyPoolHandler); err != nil {
// 		panic(err)
// 	}
// }
```
