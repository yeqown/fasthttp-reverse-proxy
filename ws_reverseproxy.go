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
	// DefaultUpgrader = &websocket.upgrader{
	// 	ReadBufferSize:  1024,
	// 	WriteBufferSize: 1024,
	// }

	// DefaultDialer is a dialer with all fields set to the default zero values.
	DefaultDialer = websocket.DefaultDialer
)

// WSReverseProxy .
// refer to https://github.com/koding/websocketproxy
type WSReverseProxy struct {
	option *buildOptionWS
}

// NewWSReverseProxyWith constructs a new WSReverseProxy with options.
func NewWSReverseProxyWith(options ...OptionWS) (*WSReverseProxy, error) {
	option := defaultBuildOptionWS()
	for _, opt := range options {
		opt.apply(option)
	}

	if err := option.validate(); err != nil {
		return nil, err
	}

	return &WSReverseProxy{
		option: option,
	}, nil
}

// ServeHTTP WSReverseProxy to serve
func (w *WSReverseProxy) ServeHTTP(ctx *fasthttp.RequestCtx) {
	if websocket.FastHTTPIsWebSocketUpgrade(ctx) {
		debugF(w.option.debug, w.option.logger, "websocketproxy: got websocket request")
	}

	var (
		// req      = &ctx.Request
		resp     = &ctx.Response
		dialer   = DefaultDialer
		upgrader = DefaultUpgrader
	)

	if w.option.dialer != nil {
		dialer = w.option.dialer
	}

	if w.option.upgrader != nil {
		upgrader = w.option.upgrader
	}

	// handle request header
	forwardHeader := builtinForwardHeaderHandler(ctx)

	// customize headers to forward, this may override headers from builtinForwardHeaderHandler
	// so be careful to set header only when you do need it.
	if w.option.fn != nil {
		appendHeaders := w.option.fn(ctx)
		for k, vs := range appendHeaders {
			for _, v := range vs {
				forwardHeader.Set(k, v)
			}
		}
	}

	// Connect to the backend URL, also pass the headers we get from the request
	// together with the Forwarded headers we prepared above.
	// TODO: support multiplexing on the same backend connection instead of
	// opening a new TCP connection time for each request. This should be
	// optional:
	// http://tools.ietf.org/html/draft-ietf-hybi-websocket-multiplexing-01
	overridePath := ctx.Request.Header.Peek("Override-Path")
	if len(overridePath) == 0 {
		overridePath = []byte(w.option.target.Path)
	}
	ref := &url.URL{Path: string(overridePath), RawQuery: string(ctx.QueryArgs().QueryString())}
	newURL := w.option.target.ResolveReference(ref)
	connBackend, respBackend, err := dialer.Dial(newURL.String(), forwardHeader)
	if err != nil {
		errorF(w.option.logger, "websocketproxy: couldn't dial to remote backend(%s): %v", w.option.target.String(), err)

		if respBackend != nil {
			if err = wsCopyResponse(resp, respBackend); err != nil {
				errorF(w.option.logger, "websocketproxy: couldn't copy response: %v", err)
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

		debugF(w.option.debug, w.option.logger, "websocketproxy: upgrade handler working")

		go replicateWebsocketConn(w.option.logger, connPub, connBackend, errClient)  // response
		go replicateWebsocketConn(w.option.logger, connBackend, connPub, errBackend) // request

		for {
			select {
			case err = <-errClient:
				message = "websocketproxy: Error when copying response: %v"
			case err = <-errBackend:
				message = "websocketproxy: Error when copying request: %v"
			}

			// log error except '*websocket.CloseError'
			if _, ok := err.(*websocket.CloseError); !ok {
				errorF(w.option.logger, "websocketproxy: error when copying %s: %v", message, err)
			}
		}
	})

	if err != nil {
		errorF(w.option.logger, "websocketproxy: couldn't upgrade %s", err)
	}

	return
}

// builtinForwardHeaderHandler built in handler for dealing forward request headers.
func builtinForwardHeaderHandler(ctx *fasthttp.RequestCtx) (forwardHeader http.Header) {
	forwardHeader = make(http.Header, 4)

	// Pass headers from the incoming request to the dialer to forward them to
	// the final destinations.
	if origin := ctx.Request.Header.Peek("Origin"); string(origin) != "" {
		forwardHeader.Add("Origin", string(origin))
	}

	if prot := ctx.Request.Header.Peek("Sec-WebSocket-Protocol"); string(prot) != "" {
		forwardHeader.Add("Sec-WebSocket-Protocol", string(prot))
	}

	if cookie := ctx.Request.Header.Peek("Cookie"); string(cookie) != "" {
		forwardHeader.Add("Cookie", string(cookie))
	}

	if string(ctx.Request.Host()) != "" {
		forwardHeader.Set("Host", string(ctx.Request.Host()))
	}

	// Pass X-Forwarded-For headers too, code below is a part of
	// httputil.ReverseProxy. See http://en.wikipedia.org/wiki/X-Forwarded-For
	// for more information
	// TODO: use RFC7239 http://tools.ietf.org/html/rfc7239
	if clientIP, _, err := net.SplitHostPort(ctx.RemoteAddr().String()); err == nil {
		// If we aren't the first proxy retain prior
		// X-Forwarded-For information as a comma+space
		// separated list and fold multiple headers into one.
		if prior := ctx.Request.Header.Peek("X-Forwarded-For"); string(prior) != "" {
			clientIP = string(prior) + ", " + clientIP
		}
		forwardHeader.Set("X-Forwarded-For", clientIP)
	}

	// Set the originating protocol of the incoming HTTP request. The SSL might
	// be terminated on our site and because we're doing proxy adding this would
	// be helpful for applications on the backend.
	forwardHeader.Set("X-Forwarded-Proto", "http")
	if ctx.IsTLS() {
		forwardHeader.Set("X-Forwarded-Proto", "https")
	}

	return
}

// replicateWebsocketConn to
// copy message from src to dst
func replicateWebsocketConn(logger __Logger, dst, src *websocket.Conn, errChan chan error) {
	for {
		msgType, msg, err := src.ReadMessage()
		if err != nil {
			// true: handle websocket close error
			errorF(logger, "replicateWebsocketConn: src.ReadMessage failed, msgType=%d, msg=%s, err=%v", msgType, msg, err)
			if ce, ok := err.(*websocket.CloseError); ok {
				msg = websocket.FormatCloseMessage(ce.Code, ce.Text)
			} else {
				errorF(logger, "replicateWebsocketConn: src.ReadMessage failed, err=%v", err)
				msg = websocket.FormatCloseMessage(websocket.CloseAbnormalClosure, err.Error())
			}

			errChan <- err
			if err = dst.WriteMessage(websocket.CloseMessage, msg); err != nil {
				errorF(logger, "replicateWebsocketConn: dst.WriteMessage failed, err=%v", err)
			}
			break
		}

		err = dst.WriteMessage(msgType, msg)
		if err != nil {
			errorF(logger, "replicateWebsocketConn: dst.WriteMessage failed, msgType=%d, msg=%s, err=%v", msgType, msg, err)
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
