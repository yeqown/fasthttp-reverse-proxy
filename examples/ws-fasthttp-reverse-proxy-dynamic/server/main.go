package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/fasthttp/websocket"
	"github.com/valyala/fasthttp"
)

var upgrader = websocket.FastHTTPUpgrader{
	CheckOrigin: func(ctx *fasthttp.RequestCtx) bool { return true },
}

func applesView(ctx *fasthttp.RequestCtx) {
	err := upgrader.Upgrade(ctx, func(ws *websocket.Conn) {
		defer ws.Close()
		for {
			mt, message, err := ws.ReadMessage()
			if err != nil {
				log.Println("read error:", err)
				break
			}
			log.Printf("'applesView' recv: %s", message)
			log.Printf("'applesView' query args: %v", ctx.QueryArgs())
			err = ws.WriteMessage(mt, []byte("You are talking about apples"))
			if err != nil {
				log.Println("'applesView' write error:", err)
				break
			}
		}
	})
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); ok {
			log.Println(err)
		}
		return
	}
	log.Println("'applesView' connection done")
}

func orangesView(ctx *fasthttp.RequestCtx) {
	err := upgrader.Upgrade(ctx, func(ws *websocket.Conn) {
		defer ws.Close()
		for {
			mt, message, err := ws.ReadMessage()
			if err != nil {
				log.Println("read error:", err)
				break
			}
			log.Printf("'orangesView' recv: %s", message)
			log.Printf("'orangesView' query args: %v", ctx.QueryArgs())
			err = ws.WriteMessage(mt, []byte("You are talking about oranges"))
			if err != nil {
				log.Println("'orangesView' write error:", err)
				break
			}
		}
	})
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); ok {
			log.Println(err)
		}
		return
	}
	log.Println("'orangesView' connection done")
}

func main() {
	flag.Parse()
	log.SetFlags(0)

	requestHandler := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/talk_about/apples":
			applesView(ctx)
		case "/talk_about/oranges":
			orangesView(ctx)
		default:
			ctx.Error("Unsupported path", fasthttp.StatusNotFound)
		}
	}

	server := fasthttp.Server{
		Name:    "EchoExample with dynamic routing",
		Handler: requestHandler,
	}

	port := 8080
	log.Println("Starting main server on:", port)
	log.Fatal(server.ListenAndServe(fmt.Sprintf(":%d", port)))
}
