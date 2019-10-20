package proxy

import (
	"log"
	"testing"

	"github.com/fasthttp/websocket"
	"github.com/valyala/fasthttp"
)

func BenchmarkNewWSReverseProxy(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := NewWSReverseProxy("localhost", "/path")
		_ = p
	}
}

// TODO: generate fasthttp websocket RequestCtx
func BenchmarkWSReverseProxy(b *testing.B) {
	upgrader := websocket.FastHTTPUpgrader{}
	echoHdl := func(ctx *fasthttp.RequestCtx) {
		err := upgrader.Upgrade(ctx, func(ws *websocket.Conn) {
			defer ws.Close()
			for {
				mt, message, err := ws.ReadMessage()
				if err != nil {
					log.Println("read:", err)
					break
				}
				log.Printf("recv: %s", message)
				err = ws.WriteMessage(mt, message)
				if err != nil {
					log.Println("write:", err)
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
	}

	hdl := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/echo":
			echoHdl(ctx)
		}
	}
	server := fasthttp.Server{
		Name:    "Name",
		Handler: hdl,
	}

	go server.ListenAndServe(":8080")

	ctx := &fasthttp.RequestCtx{}
	p := NewWSReverseProxy("localhost:8080", "/echo")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.ServeHTTP(ctx)
	}
}
