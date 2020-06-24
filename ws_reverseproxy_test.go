package proxy

import (
	"bytes"
	"log"
	"net/http"
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

func Test_WSReverseProxy(t *testing.T) {
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
	// backend server initializing
	server := fasthttp.Server{
		Name: "Name",
		Handler: func(ctx *fasthttp.RequestCtx) {
			switch string(ctx.Path()) {
			case "/echo":
				echoHdl(ctx)
			}
		},
	}

	// backend websocket server
	go func() {
		if err := server.ListenAndServe(":8080"); err != nil {
			logger.Errorf("websocket backend server `ListenAndServe` quit, err=%v", err)
		}
	}()

	// start websocket proxy server
	p := NewWSReverseProxy("localhost:8080", "/echo")
	go func() {
		reqHdl := func(ctx *fasthttp.RequestCtx) {
			p.ServeHTTP(ctx)
		}
		if err := fasthttp.ListenAndServe(":8081", reqHdl); err != nil {
			logger.Errorf("websocket proxy server `ListenAndServe` quit, err=%v", err)
		}
	}()

	// client
	conn, resp, err := websocket.DefaultDialer.Dial("ws://localhost:8081", nil)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	t.Log(resp)
	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Error("could not connect to proxy")
		t.FailNow()
	}

	// client send
	sendmsg := []byte("hello")
	err = conn.WriteMessage(websocket.TextMessage, sendmsg)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	// client read echo message
	messageType, recvmsg, err := conn.ReadMessage()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if messageType != websocket.TextMessage || !bytes.Equal(sendmsg, recvmsg) {
		t.Errorf("recv message not wanted: [%v / %v], [%s / %s]",
			messageType, websocket.TextMessage, recvmsg, sendmsg)
		t.FailNow()
	}
}
