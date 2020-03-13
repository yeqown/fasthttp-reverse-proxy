package main

import (
	"log"

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
		// homeView(ctx)
		fasthttp.ServeFileUncompressed(ctx, "./index.html")
	default:
		ctx.Error("Unsupported path", fasthttp.StatusNotFound)
	}
}

// func homeView(ctx *fasthttp.RequestCtx) {
// 	ctx.SetContentType("text/html")
// 	fd, err := os.Open("./index.html")
// 	if err != nil {
// 		log.Printf("homeView err=%v", err)
// 		ctx.Write(p)
// 	}
// 	// buf := bytes.NewBuffer(nil)
// 	dat, err := ioutil.ReadAll(fd)
// 	if err != nil {
// 		log.Printf("homeView err=%v", err)
// 	}
// 	ctx.Write(dat)
// }

func main() {
	log.Println("serving on: 8081")
	if err := fasthttp.ListenAndServe(":8081", ProxyHandler); err != nil {
		log.Fatal(err)
	}
}
