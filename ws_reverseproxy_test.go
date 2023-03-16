package proxy

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
)

func BenchmarkNewWSReverseProxy(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p, err := NewWSReverseProxyWith(WithURL_OptionWS("ws://localhost:8080/echo"))
		_ = p
		_ = err
	}
}

type wsTestSuite struct {
	suite.Suite

	server fasthttp.Server
}

func (w *wsTestSuite) SetupSuite() {
	go w.backendProc(":8080")

	time.Sleep(3 * time.Second)
}

func (w *wsTestSuite) backendProc(addr string) {
	upgrader := websocket.FastHTTPUpgrader{}

	echoHdl := func(ctx *fasthttp.RequestCtx) {
		fmt.Printf("recv headers: %v\n", string(ctx.Request.Header.Header()))

		err := upgrader.Upgrade(ctx, func(ws *websocket.Conn) {
			defer ws.Close()
			for {
				mt, message, err := ws.ReadMessage()
				if err != nil {
					fmt.Printf("read message failed: %v\n", err)
					break
				}
				fmt.Printf("recv: %s\n", message)
				err = ws.WriteMessage(mt, message)
				if err != nil {
					fmt.Printf("write message failed: %v\n", err)
					break
				}
			}
		})

		if err != nil {
			if _, ok := err.(websocket.HandshakeError); ok {
				fmt.Printf("websocket handshake: %v\n", err)
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
		fmt.Printf("websocket backend server `ListenAndServe` quit, err=%v", err)
	}
}

func executeAndAssert(t *testing.T, port int) {
	// client
	conn, resp, err := websocket.DefaultDialer.Dial(fmt.Sprintf("ws://localhost:%d", port), nil)
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

func reverseProxyProc(p *WSReverseProxy, addr string) {
	reqHdl := func(ctx *fasthttp.RequestCtx) {
		p.ServeHTTP(ctx)
	}

	if err := fasthttp.ListenAndServe(addr, reqHdl); err != nil {
		fmt.Printf("websocket proxy server `ListenAndServe` quit, err=%v\n", err)
	}
}

func (w *wsTestSuite) Test_NewWSReverseProxyWith() {
	p, err := NewWSReverseProxyWith(WithURL_OptionWS("ws://localhost:8080/echo"))
	assert.Nil(w.T(), err)
	assert.NotNil(w.T(), p)

	go reverseProxyProc(p, ":8082")
	executeAndAssert(w.T(), 8082)
}

func (w *wsTestSuite) Test_NewWSReverseProxyWith_WithForwardHeadersHandler() {
	p, err := NewWSReverseProxyWith(
		WithURL_OptionWS("ws://localhost:8080/echo"),
		WithForwardHeadersHandlers_OptionWS(func(ctx *fasthttp.RequestCtx) (forwardHeader http.Header) {
			return http.Header{
				"X-TEST-HEAD": []string{"Test_NewWSReverseProxyWith_WithForwardHeadersHandler"},
			}
		}),
	)
	assert.Nil(w.T(), err)

	go reverseProxyProc(p, ":8083")
	executeAndAssert(w.T(), 8083)
}

func Test_wsTestSuite(t *testing.T) {
	suite.Run(t, new(wsTestSuite))
}
