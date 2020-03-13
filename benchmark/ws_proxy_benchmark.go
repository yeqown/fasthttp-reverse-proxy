package main

import (
	"log"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"

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

// TODO: pprof capture
func main() {
	fd, err := os.Create("./cpu.prof")
	defer fd.Close()
	if err != nil {
		panic(err)
	}
	if err = pprof.StartCPUProfile(fd); err != nil {
		panic(err)
	}

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGHUP)

	go func() {
		for s := range ch {
			log.Printf("got signal = %v", s)
			if s == syscall.SIGINT || s == syscall.SIGQUIT {
				pprof.StopCPUProfile()
				os.Exit(2)
			}
		}
	}()

	log.Println("serving on: 8081")
	if err := fasthttp.ListenAndServe(":8081", ProxyHandler); err != nil {
		log.Fatal(err)
	}
}
