package proxy

import (
	"net/http"
	"testing"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"github.com/yeqown/log"
)

func BenchmarkNewWSReverseProxy(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := NewWSReverseProxy("localhost", "/path")
		_ = p
	}
}

func runBackend(addr string) {
	upgrader := websocket.FastHTTPUpgrader{}
	entry := logger.WithField("func", "runBackend")
	echoHdl := func(ctx *fasthttp.RequestCtx) {
		entry.
			WithFields(log.Fields{
				"reqHeader": string(ctx.Request.Header.Header()),
			}).
			Debugf("recv headers")

		err := upgrader.Upgrade(ctx, func(ws *websocket.Conn) {
			defer ws.Close()
			for {
				mt, message, err := ws.ReadMessage()
				if err != nil {
					entry.Error("read:", err)
					break
				}
				entry.Infof("recv: %s", message)
				err = ws.WriteMessage(mt, message)
				if err != nil {
					entry.Error("write:", err)
					break
				}
			}
		})

		if err != nil {
			if _, ok := err.(websocket.HandshakeError); ok {
				entry.Info(err)
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
	if err := server.ListenAndServe(addr); err != nil {
		entry.
			Errorf("websocket backend server `ListenAndServe` quit, err=%v", err)
	}
}

func doRequest(t *testing.T) {
	// client
	conn, resp, err := websocket.DefaultDialer.Dial("ws://localhost:8081", nil)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)
	t.Logf("got resp: %+v", resp)

	// client send
	data := []byte("hello")
	err = conn.WriteMessage(websocket.TextMessage, data)
	assert.Nil(t, err)

	// client read echo message
	messageType, p, err := conn.ReadMessage()
	assert.Nil(t, err)
	assert.NotNil(t, p)
	assert.Equal(t, websocket.TextMessage, messageType)
	assert.NotZero(t, p)
	assert.Equal(t, data, p)
}

func runProxy(p *WSReverseProxy, addr string) {
	reqHdl := func(ctx *fasthttp.RequestCtx) {
		p.ServeHTTP(ctx)
	}

	if err := fasthttp.ListenAndServe(addr, reqHdl); err != nil {
		logger.Errorf("websocket proxy server `ListenAndServe` quit, err=%v", err)
	}
}

func Test_NewWSReverseProxy(t *testing.T) {
	go runBackend(":8080")
	time.Sleep(3 * time.Second)

	// constructs a websocket proxy server
	var p *WSReverseProxy
	assert.NotPanics(t, func() {
		p = NewWSReverseProxy("localhost:8080", "/echo")
	}, "compatiable old version API failed")
	assert.NotNil(t, p)

	// star and serve
	go runProxy(p, ":8081")

	doRequest(t)
}

func Test_NewWSReverseProxyWith(t *testing.T) {
	go runBackend(":8080")

	time.Sleep(3 * time.Second)
	// constructs a websocket proxy server
	p, err := NewWSReverseProxyWith(WithURL_OptionWS("ws://localhost:8080/echo"))
	assert.Nil(t, err)
	assert.NotNil(t, p)

	// star and serve
	go runProxy(p, ":8081")

	doRequest(t)
}

func Test_NewWSReverseProxyWith_WithForwardHeadersHandler(t *testing.T) {
	go runBackend(":8080")

	time.Sleep(3 * time.Second)
	// constructs a websocket proxy server
	p, err := NewWSReverseProxyWith(
		WithURL_OptionWS("ws://localhost:8080/echo"),
		WithForwardHeadersHandlers_OptionWS(func(ctx *fasthttp.RequestCtx) (forwardHeader http.Header) {
			return http.Header{
				"X-TEST-HEAD": []string{"Test_NewWSReverseProxyWith_WithForwardHeadersHandler"},
			}
		}),
	)
	assert.Nil(t, err)

	// star and serve
	go runProxy(p, ":8081")

	// doRequest
	doRequest(t)
}
