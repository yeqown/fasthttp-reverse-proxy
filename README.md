# fasthttp-reverse-proxy
![](https://img.shields.io/badge/LICENSE-MIT-blue.svg) [![Go Report Card](https://goreportcard.com/badge/github.com/yeqown/fasthttp-reverse-proxy/v2)](https://goreportcard.com/report/github.com/yeqown/fasthttp-reverse-proxy/v2) [![GoReportCard](https://godoc.org/github.com/yeqown/fasthttp-reverse-proxy/v2?status.svg)](https://godoc.org/github.com/yeqown/fasthttp-reverse-proxy/v2)

reverse http proxy hander based on fasthttp.

## features:

* [x] proxy client has `pool` supported

* [x] faster than golang standard `httputil.ReverseProxy`

* [x] simple warpper of `fasthttp.HostClient` 

* [x] websocket proxy

* [x] support balance distibute based `rounddobin`

## quick start

#### HTTP (with balancer option)
```go
import (
	"log"

	"github.com/valyala/fasthttp"
	proxy "github.com/yeqown/fasthttp-reverse-proxy/v2"
)

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

#### Websocket

```go
import (
	"log"
	"text/template"

	"github.com/valyala/fasthttp"
	proxy "github.com/yeqown/fasthttp-reverse-proxy/v2"
)

var (
	proxyServer = proxy.NewWSReverseProxy("localhost:8080", "/echo")
)

// ProxyHandler ... fasthttp.RequestHandler func
func ProxyHandler(ctx *fasthttp.RequestCtx) {
	switch string(ctx.Path()) {
	case "/":
		proxyServer.ServeHTTP(ctx)
	default:
		ctx.Error("Unsupported path", fasthttp.StatusNotFound)
	}
}

func main() {
	if err := fasthttp.ListenAndServe(":8081", ProxyHandler); err != nil {
		log.Fatal(err)
	}
}
```

## usage

* [use it alone](./examples/fasthttp-reverse-proxy/proxy.go)
* [use it with pool](./examples/fasthttp-reverse-proxy-with-pool/pool.go)
* [websocket](./examples/ws-fasthttp-reverse-proxy)

## contrast

* [HTTP benchmark](./docs/http-benchmark.md)
* [Websocket benchmark](./docs/ws-benchmark.md)

## links:

* [fasthttp](https://github.com/valyala/fasthttp)
* [standard httputil.ReverseProxy](https://golang.org/pkg/net/http/httputil/#ReverseProxy)
* [fasthttp/websocket](https://github.com/fasthttp/websocket)
* [koding/websocketproxy](https://github.com/koding/websocketproxy)

## Thanks

<a href="https://www.jetbrains.com/?from=fasthttp-reverse-proxy" _blank="#">
    <img src="https://www.jetbrains.com/company/brand/img/jetbrains_logo.png" width="100" alt="JetBrains"/>
</a>