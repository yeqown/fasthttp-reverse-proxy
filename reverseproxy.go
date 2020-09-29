// Copyright 2018 The yeqown Author. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package proxy

import (
	"net"
	"net/http"

	"github.com/valyala/fasthttp"
)

const (
	_fasthttpHostClientName = "reverse-proxy"
)

// ReverseProxy reverse handler using fasthttp.HostClient
type ReverseProxy struct {
	// oldAddr to keep old API working as usual, this field should be removed
	oldAddr string

	// bla keeps balancer instance
	bla IBalancer

	// clients
	clients []*fasthttp.HostClient

	// opt contains finally option to open reverseProxy
	opt *buildOption
}

// NewReverseProxy create an ReverseProxy with options
func NewReverseProxy(oldAddr string, opts ...Option) *ReverseProxy {
	dst := new(buildOption)
	for _, opt := range opts {
		opt.apply(dst)
	}

	logger.Debugf("dst opt=%+v opts=%+v\n", dst, opts)

	// apply an new object of `ReverseProxy`
	proxy := ReverseProxy{
		oldAddr: oldAddr,
		bla:     nil,
		opt:     dst,
		clients: make([]*fasthttp.HostClient, 0, 2),
	}

	(&proxy).init()
	return &proxy
}

// initialize the ReverseProxy with options,
// if opt.OpenBalance is true then create a balancer to ReverseProxy
// else just
func (p *ReverseProxy) init() {
	if p.opt.openBalance {
		// config balancer
		p.oldAddr = ""
		p.clients = make([]*fasthttp.HostClient, len(p.opt.addresses))
		p.bla = NewBalancer(p.opt.weights)

		for idx, addr := range p.opt.addresses {
			p.clients[idx] = &fasthttp.HostClient{
				Addr:      addr,
				Name:      _fasthttpHostClientName,
				IsTLS:     p.opt.tlsConfig != nil,
				TLSConfig: p.opt.tlsConfig,
			}
		}

		return
	}

	// not open balancer
	p.bla = nil
	p.clients = append(p.clients,
		&fasthttp.HostClient{
			Addr:      p.oldAddr,
			Name:      _fasthttpHostClientName,
			IsTLS:     p.opt.tlsConfig != nil,
			TLSConfig: p.opt.tlsConfig,
		})
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

	// assign the host to support virtual hosting, aka shared web hosting (one IP, multiple domains)
	req.SetHost(pc.Addr)

	logger.Debugf("pc with tlsConfig=%+v", pc.TLSConfig)

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
