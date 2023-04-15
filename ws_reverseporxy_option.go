package proxy

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/fasthttp/websocket"
	"github.com/valyala/fasthttp"
)

// OptionWS to define all options to reverse web socket proxy.
type OptionWS interface {
	apply(o *buildOptionWS)
}

type funcBuildOptionWS struct {
	f func(o *buildOptionWS)
}

func newFuncBuildOptionWS(f func(o *buildOptionWS)) funcBuildOptionWS { return funcBuildOptionWS{f: f} }
func (fb funcBuildOptionWS) apply(o *buildOptionWS)                   { fb.f(o) }

type forwardHeaderHandler func(ctx *fasthttp.RequestCtx) (forwardHeader http.Header)

// buildOptionWS is Option for WS reverse-proxy
type buildOptionWS struct {
	// logger is used to log messages.
	logger __Logger
	// debug is used to enable debug mode.
	debug bool

	// target indicates which backend server to proxy.
	target *url.URL

	// fn is forwardHeaderHandler which allows users customize themselves' forward headers
	// to be proxied to backend server.
	fn forwardHeaderHandler

	// dialer contains options for connecting to the backend WebSocket server.
	// If nil, DefaultDialer is used.
	dialer *websocket.Dialer

	// upgrader specifies the parameters for upgrading a incoming HTTP
	// connection to a WebSocket connection. If nil, DefaultUpgrader is used.
	upgrader *websocket.FastHTTPUpgrader
}

func (o *buildOptionWS) validate() error {
	if o == nil {
		return errors.New("option is nil")
	}

	if o.target == nil {
		return errors.New("target is nil")
	}

	return nil
}

func defaultBuildOptionWS() *buildOptionWS {
	return &buildOptionWS{
		logger:   &nopLogger{},
		debug:    false,
		target:   nil,
		fn:       nil,
		dialer:   nil,
		upgrader: nil,
	}
}

// WithURL_OptionWS specify the url to backend websocket server.
// WithURL_OptionWS("ws://YOUR_WEBSOCKET_HOST:PORT/AND/PATH")
func WithURL_OptionWS(u string) OptionWS {
	return newFuncBuildOptionWS(func(o *buildOptionWS) {
		URL, err := url.Parse(u)
		if err != nil {
			panic(err)
		}

		o.target = URL
	})
}

// WithDebug_OptionWS is used to enable debug mode.
func WithDebug_OptionWS() OptionWS {
	return newFuncBuildOptionWS(func(o *buildOptionWS) {
		o.debug = true
	})
}

// WithDialer_OptionWS use specified dialer
func WithDialer_OptionWS(dialer *websocket.Dialer) OptionWS {
	return newFuncBuildOptionWS(func(o *buildOptionWS) {
		o.dialer = dialer
	})
}

// WithUpgrader_OptionWS use specified upgrader.
func WithUpgrader_OptionWS(upgrader *websocket.FastHTTPUpgrader) OptionWS {
	return newFuncBuildOptionWS(func(o *buildOptionWS) {
		o.upgrader = upgrader
	})
}

// WithForwardHeadersHandlers_OptionWS allows users to customize forward headers.
func WithForwardHeadersHandlers_OptionWS(handler forwardHeaderHandler) OptionWS {
	return newFuncBuildOptionWS(func(o *buildOptionWS) {
		o.fn = handler
	})
}
