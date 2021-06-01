# fasthttp-reverse-proxy
![](https://img.shields.io/badge/LICENSE-MIT-blue.svg) [![Go Report Card](https://goreportcard.com/badge/github.com/yeqown/fasthttp-reverse-proxy/v2)](https://goreportcard.com/report/github.com/yeqown/fasthttp-reverse-proxy/v2) [![GoReportCard](https://godoc.org/github.com/yeqown/fasthttp-reverse-proxy/v2?status.svg)](https://godoc.org/github.com/yeqown/fasthttp-reverse-proxy/v2)

reverse http proxy handler based on fasthttp.

## Features

- [x] `HTTP` reverse proxy based [fasthttp](https://github.com/valyala/fasthttp)
  
	- [x] it's faster than golang standard `httputil.ReverseProxy` library.
	- [x] implemented by `fasthttp.HostClient` 
	- [x] support balance distribute based `rounddobin`
	- [x] `HostClient` object pool with an overlay of fasthttp connection pool.

* [x] `WebSocket` reverse proxy.

## Get started

#### [HTTP (with balancer option)](./examples/fasthttp-reverse-proxy-with-bla/proxy.go)

```go
var (
	proxyServer = proxy.NewReverseProxy("localhost:8080")

	// use with balancer
	// weights = map[string]proxy.Weight{
	// 	"localhost:8080": 20,
	// 	"localhost:8081": 30,
	// 	"localhost:8082": 50,
	// }
	// proxyServer = proxy.NewReverseProxy("", proxy.WithBalancer(weights))

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
```

#### [Websocket](./examples/ws-fasthttp-reverse-proxy/README.md)

```go
var (
	proxyServer *proxy.WSReverseProxy
	once        sync.Once
)

// ProxyHandler ... fasthttp.RequestHandler func
func ProxyHandler(ctx *fasthttp.RequestCtx) {
	once.Do(func() {
		var err error
		proxyServer, err = proxy.NewWSReverseProxyWith(
			proxy.WithURL_OptionWS("ws://localhost:8080/echo"),
		)
		if err != nil {
			panic(err)
		}
	})

	switch string(ctx.Path()) {
	case "/echo":
		proxyServer.ServeHTTP(ctx)
	case "/":
		fasthttp.ServeFileUncompressed(ctx, "./index.html")
	default:
		ctx.Error("Unsupported path", fasthttp.StatusNotFound)
	}
}

func main() {
	log.Println("serving on: 8081")
	if err := fasthttp.ListenAndServe(":8081", ProxyHandler); err != nil {
		log.Fatal(err)
	}
}

```

## Usages

* [HTTP reverse proxy](./examples/fasthttp-reverse-proxy/proxy.go)
* [HTTP reverse proxy with object pool](./examples/fasthttp-reverse-proxy-with-pool/pool.go)
* [Websocket reverse proxy](./examples/ws-fasthttp-reverse-proxy)

## Contrast

* [HTTP benchmark](./docs/http-benchmark.md)
* [Websocket benchmark](./docs/ws-benchmark.md)

## References

* [fasthttp](https://github.com/valyala/fasthttp)
* [standard httputil.ReverseProxy](https://golang.org/pkg/net/http/httputil/#ReverseProxy)
* [fasthttp/websocket](https://github.com/fasthttp/websocket)
* [koding/websocketproxy](https://github.com/koding/websocketproxy)

## Thanks

<a href="https://www.jetbrains.com/?from=fasthttp-reverse-proxy" _blank="#">
    <img src="https://www.jetbrains.com/company/brand/img/jetbrains_logo.png" width="100" alt="JetBrains"/>
</a>