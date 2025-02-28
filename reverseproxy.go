// Copyright 2018 The yeqown Author. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package proxy

import (
	"errors"
	"net"
	"net/http"

	"github.com/valyala/fasthttp"
)

const (
	_fasthttpHostClientName = "reverse-proxy"
)

// ReverseProxy reverse handler using fasthttp.HostClient
type ReverseProxy struct {
	// bla keeps balancer instance
	bla IBalancer

	// clients
	clients []*fasthttp.HostClient

	// opt contains finally option to open reverseProxy
	opt *buildOption
}

// NewReverseProxyWith create an ReverseProxy with options
func NewReverseProxyWith(options ...Option) (*ReverseProxy, error) {
	option := defaultBuildOption()
	for _, opt := range options {
		opt.apply(option)
	}

	proxy := &ReverseProxy{
		bla:     nil,
		opt:     option,
		clients: make([]*fasthttp.HostClient, 0, 2),
	}

	if err := proxy.init(); err != nil {
		return nil, err
	}

	return proxy, nil
}

// init initialize the ReverseProxy with options,
// if opted.OpenBalance is true then create a balancer to ReverseProxy
// else just create a HostClient to ReverseProxy and use it.
func (p *ReverseProxy) init() error {
	if len(p.opt.addresses) == 0 {
		return errors.New("no upstream server address")
	}

	if p.opt.openBalance {
		// config balancer
		p.clients = make([]*fasthttp.HostClient, 0, len(p.opt.addresses))
		p.bla = NewBalancer(p.opt.weights)

		for _, addr := range p.opt.addresses {
			client := &fasthttp.HostClient{
				Addr:                   addr,
				Name:                   _fasthttpHostClientName,
				IsTLS:                  p.opt.tlsConfig != nil,
				TLSConfig:              p.opt.tlsConfig,
				DisablePathNormalizing: p.opt.disablePathNormalizing,
				MaxResponseBodySize:    p.opt.maxResponseBodySize,
				StreamResponseBody:     p.opt.streamResponseBody,
			}
			p.clients = append(p.clients, client)
		}

		return nil
	}

	// not open balancer
	p.bla = nil
	addr := p.opt.addresses[0]
	client := &fasthttp.HostClient{
		Addr:                   addr,
		Name:                   _fasthttpHostClientName,
		IsTLS:                  p.opt.tlsConfig != nil,
		TLSConfig:              p.opt.tlsConfig,
		DisablePathNormalizing: p.opt.disablePathNormalizing,
		MaxResponseBodySize:    p.opt.maxResponseBodySize,
		StreamResponseBody:     p.opt.streamResponseBody,
		MaxConnDuration:        p.opt.maxConnDuration,
	}
	p.clients = append(p.clients, client)
	return nil
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

	c := p.getClient()
	debugF(p.opt.debug, p.opt.logger, "rev request headers to proxy, addr = %s, headers = %s", c.Addr, req.Header.String())

	// assign the host to support virtual hosting, aka shared web hosting (one IP, multiple domains)
	if !p.opt.disableVirtualHost {
		req.SetHost(c.Addr)
	}

	// execute the request and rev response with timeout
	if err := p.doWithTimeout(c, req, res); err != nil {
		errorF(p.opt.logger, "p.doWithTimeout failed, err = %v, status = %d", err, res.StatusCode())
		res.SetStatusCode(http.StatusInternalServerError)
		if errors.Is(err, fasthttp.ErrTimeout) {
			res.SetStatusCode(http.StatusRequestTimeout)
		}

		res.SetBody([]byte(err.Error()))
		return
	}

	// deal with response headers
	debugF(p.opt.debug, p.opt.logger, "rev response headers from proxy, addr = %s, headers = %s", c.Addr, res.Header.String())

	for _, h := range hopHeaders {
		res.Header.Del(h)
	}
}

// doWithTimeout calls fasthttp.HostClient Do or DoTimeout, this is depends on p.opt.timeout
func (p *ReverseProxy) doWithTimeout(pc *fasthttp.HostClient, req *fasthttp.Request, res *fasthttp.Response) error {
	if p.opt.timeout <= 0 {
		return pc.Do(req, res)
	}

	return pc.DoTimeout(req, res, p.opt.timeout)
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
