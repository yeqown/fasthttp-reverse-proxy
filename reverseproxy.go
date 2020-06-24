// Copyright 2018 The yeqown Author. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package proxy

import (
	"net"
	"net/http"

	"github.com/imdario/mergo"
	"github.com/valyala/fasthttp"
)

// var _ Proxier = &ReverseProxy{}

// Option . contains option field of ReverseProxy
type Option struct {
	OpenBalance bool
	Ws          []W
	Addrs       []string
}

// WithBalancer .
func WithBalancer(addrWeights map[string]Weight) Option {
	ws := make([]W, 0, len(addrWeights))
	addrs := make([]string, 0, len(addrWeights))
	for k, v := range addrWeights {
		ws = append(ws, v)
		addrs = append(addrs, k)
	}

	return Option{
		OpenBalance: true,
		Ws:          ws,
		Addrs:       addrs,
	}
}

// ReverseProxy reverse handler using fasthttp.HostClient
// TODO: support https config
type ReverseProxy struct {
	oldAddr string                 // old addr to keep old API working as usual
	bla     IBalancer              // balancer
	ws      []W                    // weights of clients, releated by idx
	clients []*fasthttp.HostClient // clients
	opt     *Option                // opt contains finnally option to open reverseProxy
}

// NewReverseProxy create an ReverseProxy with options
func NewReverseProxy(oldAddr string, opts ...Option) *ReverseProxy {
	dstOption := &Option{
		OpenBalance: false,
		Ws:          nil,
		Addrs:       nil,
	}

	// merge opts into dstOption
	for _, opt := range opts {
		if err := mergo.Map(dstOption, opt, mergo.WithOverride); err != nil {
			panic(err)
		}
	}

	// fmt.Printf("dst opt=%+v opts=%+v\n", dstOption, opts)

	// apply an new object of `ReverseProxy`
	proxy := ReverseProxy{
		oldAddr: oldAddr,
		bla:     nil,
		// ws:      make([]W, 0, 1),
		// clients: make([]*fasthttp.HostClient, 0, 1),
		opt: dstOption,
	}

	(&proxy).init()
	return &proxy
}

func (p *ReverseProxy) init() {
	if p.opt.OpenBalance {
		p.oldAddr = ""
		p.clients = make([]*fasthttp.HostClient, len(p.opt.Addrs))
		p.ws = p.opt.Ws
		p.bla = NewBalancer(p.ws)

		for idx, addr := range p.opt.Addrs {
			p.clients[idx] = &fasthttp.HostClient{Addr: addr}
		}

		return
	}

	// not open balancer
	p.ws = append(p.ws, Weight(100))
	p.bla = nil
	p.clients = append(p.clients,
		&fasthttp.HostClient{Addr: p.oldAddr})
}

func (p *ReverseProxy) getClient() *fasthttp.HostClient {
	if p.clients == nil {
		// closed
		panic("ReverseProxy has been closed")
	}

	if p.bla != nil {
		// bla has been opened
		idx := p.bla.Distribute()
		return p.clients[idx]
	}

	return p.clients[0]
}

// ServeHTTP ReverseProxy to serve
// ref to: https://golang.org/src/net/http/httputil/reverseproxy.go#L169
func (p *ReverseProxy) ServeHTTP(ctx *fasthttp.RequestCtx) {
	req := &ctx.Request
	res := &ctx.Response

	// prepare request(replace headers and some URL host)
	if ip, _, err := net.SplitHostPort(ctx.RemoteAddr().String()); err == nil {
		req.Header.Add("X-Forwarded-For", ip)
	}

	// to save all response header
	// resHeaders := make(map[string]string)
	// res.Header.VisitAll(func(k, v []byte) {
	// 	key := string(k)
	// 	value := string(v)
	// 	if val, ok := resHeaders[key]; ok {
	// 		resHeaders[key] = val + "," + value
	// 	}
	// 	resHeaders[key] = value
	// })

	for _, h := range hopHeaders {
		// if h == "Te" && hv == "trailers" {
		// 	continue
		// }
		req.Header.Del(h)
	}

	pc := p.getClient()
	if debug {
		logger.Infof("recv a requets to proxy to: %s", pc.Addr)
	}

	if err := pc.Do(req, res); err != nil {
		logger.Errorf("could not proxy: %v\n", err)
		res.SetStatusCode(http.StatusInternalServerError)
		res.SetBody([]byte(err.Error()))
		return
	}

	logger.Debugf("response headers = %s", res.Header.String())
	// write response headers
	for _, h := range hopHeaders {
		res.Header.Del(h)
	}

	// logger.Debugf("response headers = %s", resHeaders)
	// for k, v := range resHeaders {
	// 	res.Header.Set(k, v)
	// }
}

// SetClient ...
func (p *ReverseProxy) SetClient(addr string) *ReverseProxy {
	for idx := range p.clients {
		p.clients[idx].Addr = addr
	}
	return p
}

// Reset ...
func (p *ReverseProxy) Reset() {
	for idx := range p.clients {
		p.clients[idx].Addr = ""
	}
}

// Close ... clear and release
func (p *ReverseProxy) Close() {
	p.clients = nil
	p.opt = nil
	p.bla = nil
	p.ws = nil
	p = nil
}

//
//func copyResponse(src *fasthttp.Response, dst *fasthttp.Response) {
//	src.CopyTo(dst)
//	logger.Debugf("response header=%v", src.Header)
//}
//
//func copyRequest(src *fasthttp.Request, dst *fasthttp.Request) {
//	src.CopyTo(dst)
//}
//
//func cloneResponse(src *fasthttp.Response) *fasthttp.Response {
//	dst := new(fasthttp.Response)
//	copyResponse(src, dst)
//	return dst
//}
//
//func cloneRequest(src *fasthttp.Request) *fasthttp.Request {
//	dst := new(fasthttp.Request)
//	copyRequest(src, dst)
//	return dst
//}

// Hop-by-hop headers. These are removed when sent to the backend.
// As of RFC 7230, hop-by-hop headers are required to appear in the
// Connection header field. These are the headers defined by the
// obsoleted RFC 2616 (section 13.5.1) and are used for backward
// compatibility.
var hopHeaders = []string{
	"Connection",          // Connection
	"Proxy-Connection",    // non-standard but still sent by libcurl and rejected by e.g. google
	"Keep-Alive",          // Keep-Alive
	"Proxy-Authenticate",  // Proxy-Authenticate
	"Proxy-Authorization", // Proxy-Authorization
	"Te",                  // canonicalized version of "TE"
	"Trailer",             // not Trailers per URL above; https://www.rfc-editor.org/errata_search.php?eid=4522
	"Transfer-Encoding",   // Transfer-Encoding
	"Upgrade",             // Upgrade
}
