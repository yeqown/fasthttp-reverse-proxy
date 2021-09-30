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
	"github.com/yeqown/log"
)

func BenchmarkNewWSReverseProxy(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := NewWSReverseProxy("localhost", "/path")
		_ = p
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
	entry := logger.WithField("func", "backendProc")
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

func executeAndAssert(t *testing.T, port int) {
	// client
	conn, resp, err := websocket.DefaultDialer.Dial(fmt.Sprintf("ws://localhost:%d",port), nil)
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
		logger.Errorf("websocket proxy server `ListenAndServe` quit, err=%v", err)
	}
}

func (w *wsTestSuite) Test_NewWSReverseProxy() {
	var p *WSReverseProxy
	assert.NotPanics(w.T(), func() {
		p = NewWSReverseProxy("localhost:8080", "/echo")
	}, "compatible old version API failed")
	assert.NotNil(w.T(), p)

	go reverseProxyProc(p, ":8081")
	executeAndAssert(w.T(), 8081)
}

func (w *wsTestSuite)Test_NewWSReverseProxyWith() {
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