package proxy

import (
	"bytes"
	"fmt"
	"io"
	"log"
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
		log.Println("Request is upgrade: ", b)
	}

	var (
		req      = &ctx.Request
		res      = &ctx.Response
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

	// log.Printf("requestHeader: %v", requestHeader)
	// Connect to the backend URL, also pass the headers we get from the requst
	// together with the Forwarded headers we prepared above.
	// TODO: support multiplexing on the same backend connection instead of
	// opening a new TCP connection time for each request. This should be
	// optional:
	// http://tools.ietf.org/html/draft-ietf-hybi-websocket-multiplexing-01
	connBackend, resp, err := dialer.Dial(w.target.String(), requestHeader)
	if err != nil {
		log.Printf("websocketproxy: couldn't dial to remote backend url %s", err)
		if resp != nil {
			wsCopyResponse(res, resp)
			// send http.Response to fasthttp.Response
		} else {
			ctx.SetStatusCode(http.StatusServiceUnavailable)
			ctx.WriteString(http.StatusText(http.StatusServiceUnavailable))
		}
		return
	}

	errClient := make(chan error, 1)
	errBackend := make(chan error, 1)

	// Now upgrade the existing incoming request to a WebSocket connection.
	// Also pass the header that we gathered from the Dial handshake.
	err = upgrader.Upgrade(ctx, func(connPub *websocket.Conn) {
		defer connPub.Close()

		log.Println("upgrade handler worked")

		go replicateWebsocketConn(connPub, connBackend, errClient)
		go replicateWebsocketConn(connBackend, connPub, errBackend)

		var message string
		for {
			select {
			case err = <-errClient:
				message = "websocketproxy: Error when copying from backend to client: %v"
			case err = <-errBackend:
				message = "websocketproxy: Error when copying from client to backend: %v"
			}
			if e, ok := err.(*websocket.CloseError); !ok || e.Code == websocket.CloseAbnormalClosure {
				log.Printf(message, err)
			}
		}
	})

	if err != nil {
		log.Printf("websocketproxy: couldn't upgrade %s", err)
		return
	}
}

// replicateWebsocketConn to
// copy message from src to dst
func replicateWebsocketConn(dst, src *websocket.Conn, errc chan error) {
	for {
		msgType, msg, err := src.ReadMessage()
		if err != nil {
			m := websocket.FormatCloseMessage(websocket.CloseNormalClosure, fmt.Sprintf("%v", err))
			if e, ok := err.(*websocket.CloseError); ok {
				if e.Code != websocket.CloseNoStatusReceived {
					m = websocket.FormatCloseMessage(e.Code, e.Text)
				}
			}
			errc <- err
			dst.WriteMessage(websocket.CloseMessage, m)
			break
		}
		err = dst.WriteMessage(msgType, msg)
		if err != nil {
			errc <- err
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
