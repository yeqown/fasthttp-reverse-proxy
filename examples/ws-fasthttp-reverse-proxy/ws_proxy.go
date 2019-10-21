package main

import (
	"log"
	"text/template"

	"github.com/valyala/fasthttp"
	proxy "github.com/yeqown/fasthttp-reverse-proxy"
)

var (
	proxyServer = proxy.NewWSReverseProxy("localhost:8080", "/echo")
)

// ProxyHandler ... fasthttp.RequestHandler func
func ProxyHandler(ctx *fasthttp.RequestCtx) {
	switch string(ctx.Path()) {
	case "/echo":
		proxyServer.ServeHTTP(ctx)
	case "/":
		homeView(ctx)
	default:
		ctx.Error("Unsupported path", fasthttp.StatusNotFound)
	}
}

func homeView(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("text/html")
	homeTemplate.Execute(ctx, nil)
}

func main() {
	if err := fasthttp.ListenAndServe(":8081", ProxyHandler); err != nil {
		log.Fatal(err)
	}
}

var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
</head>
<body>
	<h1>WS Porxy Body</h1>
</body>
</html>
`))
