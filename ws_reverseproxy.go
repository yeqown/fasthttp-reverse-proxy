package proxy

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"net/url"

	"github.com/fasthttp/websocket"
	"github.com/valyala/fasthttp"
)

var (
	// DefaultUpgrader specifies the parameters for upgrading an HTTP
	// connection to a WebSocket connection.
	DefaultUpgrader = &websocket.FastHTTPUpgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	// DefaultUpgrader = &websocket.Upgrader{
	// 	ReadBufferSize:  1024,
	// 	WriteBufferSize: 1024,
	// }

	// DefaultDialer is a dialer with all fields set to the default zero values.
	DefaultDialer = websocket.DefaultDialer
)

// WSReverseProxy .
// refer to https://github.com/koding/websocketproxy
type WSReverseProxy struct {
	target url.URL

	// Upgrader specifies the parameters for upgrading a incoming HTTP
	// connection to a WebSocket connection. If nil, DefaultUpgrader is used.
	Upgrader *websocket.FastHTTPUpgrader
	// Upgrader *websocket.Upgrader

	//  Dialer contains options for connecting to the backend WebSocket server.
	//  If nil, DefaultDialer is used.
	Dialer *websocket.Dialer
}

// NewWSReverseProxy .
func NewWSReverseProxy(host, path string) *WSReverseProxy {
	return &WSReverseProxy{
		target: url.URL{
			Scheme: "ws",
			Host:   host,
			Path:   path,
		},
	}
}

// ServeHTTP WSReverseProxy to serve
func (w *WSReverseProxy) ServeHTTP(ctx *fasthttp.RequestCtx) {
	if b := websocket.FastHTTPIsWebSocketUpgrade(ctx); b {
		logger.Debugf("Request is upgraded %v", b)
	}

	var (
		req      = &ctx.Request
		resp     = &ctx.Response
		dialer   = DefaultDialer
		upgrader = DefaultUpgrader
	)

	if w.Dialer != nil {
		dialer = w.Dialer
	}

	if w.Upgrader != nil {
		upgrader = w.Upgrader
	}

	// Pass headers from the incoming request to the dialer to forward them to
	// the final destinations.
	requestHeader := http.Header{}
	if origin := req.Header.Peek("Origin"); string(origin) != "" {
		requestHeader.Add("Origin", string(origin))
	}

	if prot := req.Header.Peek("Sec-WebSocket-Protocol"); string(prot) != "" {
		requestHeader.Add("Sec-WebSocket-Protocol", string(prot))
	}

	if cookie := req.Header.Peek("Cookie"); string(cookie) != "" {
		requestHeader.Add("Sec-WebSocket-Protocol", string(cookie))
	}

	if string(req.Host()) != "" {
		requestHeader.Set("Host", string(req.Host()))
	}

	// Pass X-Forwarded-For headers too, code below is a part of
	// httputil.ReverseProxy. See http://en.wikipedia.org/wiki/X-Forwarded-For
	// for more information
	// TODO: use RFC7239 http://tools.ietf.org/html/rfc7239
	if clientIP, _, err := net.SplitHostPort(ctx.RemoteAddr().String()); err == nil {
		// If we aren't the first proxy retain prior
		// X-Forwarded-For information as a comma+space
		// separated list and fold multiple headers into one.
		if prior := req.Header.Peek("X-Forwarded-For"); string(prior) != "" {
			clientIP = string(prior) + ", " + clientIP
		}
		requestHeader.Set("X-Forwarded-For", clientIP)
	}

	// Set the originating protocol of the incoming HTTP request. The SSL might
	// be terminated on our site and because we doing proxy adding this would
	// be helpful for applications on the backend.
	requestHeader.Set("X-Forwarded-Proto", "http")
	if ctx.IsTLS() {
		requestHeader.Set("X-Forwarded-Proto", "https")
	}

	// Connect to the backend URL, also pass the headers we get from the requst
	// together with the Forwarded headers we prepared above.
	// TODO: support multiplexing on the same backend connection instead of
	// opening a new TCP connection time for each request. This should be
	// optional:
	// http://tools.ietf.org/html/draft-ietf-hybi-websocket-multiplexing-01
	connBackend, respBackend, err := dialer.Dial(w.target.String(), requestHeader)
	if err != nil {
		logger.Errorf("websocketproxy: couldn't dial to remote backend host=%s, err=%v", w.target.String(), err)
		logger.Debugf("resp_backent =%v", respBackend)
		if respBackend != nil {
			if err := wsCopyResponse(resp, respBackend); err != nil {
				logger.Errorf("could not finish wsCopyResponse, err=%v", err)
			}
		} else {
			// ctx.SetStatusCode(http.StatusServiceUnavailable)
			// ctx.WriteString(http.StatusText(http.StatusServiceUnavailable))
			ctx.Error(err.Error(), fasthttp.StatusServiceUnavailable)
		}
		return
	}

	// Now upgrade the existing incoming request to a WebSocket connection.
	// Also pass the header that we gathered from the Dial handshake.
	err = upgrader.Upgrade(ctx, func(connPub *websocket.Conn) {
		defer connPub.Close()
		var (
			errClient  = make(chan error, 1)
			errBackend = make(chan error, 1)
			message    string
		)

		logger.Debug("upgrade handler working")
		go replicateWebsocketConn(connPub, connBackend, errClient)  // response
		go replicateWebsocketConn(connBackend, connPub, errBackend) // request

		for {
			select {
			case err = <-errClient:
				message = "websocketproxy: Error when copying response: %v"
			case err = <-errBackend:
				message = "websocketproxy: Error when copying request: %v"
			}

			// log error except '*websocket.CloseError'
			if _, ok := err.(*websocket.CloseError); !ok {
				logger.Errorf(message, err)
			}
		}
	})

	if err != nil {
		logger.Errorf("websocketproxy: couldn't upgrade %s", err)
		return
	}
}

// replicateWebsocketConn to
// copy message from src to dst
func replicateWebsocketConn(dst, src *websocket.Conn, errChan chan error) {
	for {
		msgType, msg, err := src.ReadMessage()
		if err != nil {
			// true: handle websocket close error
			logger.Debugf("src.ReadMessage failed, msgType=%d, msg=%s, err=%v", msgType, msg, err)
			if ce, ok := err.(*websocket.CloseError); ok {
				msg = websocket.FormatCloseMessage(ce.Code, ce.Text)
			} else {
				logger.Errorf("src.ReadMessage failed, err=%v", err)
				msg = websocket.FormatCloseMessage(websocket.CloseAbnormalClosure, err.Error())
			}

			errChan <- err
			if err = dst.WriteMessage(websocket.CloseMessage, msg); err != nil {
				logger.Errorf("write close message failed, err=%v", err)
			}
			break
		}

		err = dst.WriteMessage(msgType, msg)
		if err != nil {
			logger.Errorf("dst.WriteMessage failed, err=%v", err)
			errChan <- err
			break
		}
	}
}

// wsCopyResponse .
// to help copy origin websocket response to client
func wsCopyResponse(dst *fasthttp.Response, src *http.Response) error {
	for k, vv := range src.Header {
		for _, v := range vv {
			dst.Header.Add(k, v)
		}
	}

	dst.SetStatusCode(src.StatusCode)
	defer src.Body.Close()

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, src.Body); err != nil {
		return err
	}
	dst.SetBody(buf.Bytes())
	return nil
}
