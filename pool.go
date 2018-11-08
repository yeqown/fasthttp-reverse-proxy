package proxy

import (
	"sync"

	"github.com/valyala/fasthttp"
)

// NewPool ...
func NewPool() *ReverseProxyPool {
	return &ReverseProxyPool{
		pool: &sync.Pool{
			New: func() interface{} { return &ReverseProxy{} },
		},
	}
}

// ReverseProxyPool ...
type ReverseProxyPool struct {
	pool *sync.Pool
}

// New ...
func (p *ReverseProxyPool) New(addr string) (*ReverseProxy, func()) {
	rp := defaultPool.Get().(*ReverseProxy)
	rp.client = new(fasthttp.HostClient)

	putback := func() {
		rp.client = nil
		defaultPool.Put(rp)
	}
	return rp, putback
}

var (
	defaultPool = &sync.Pool{
		New: func() interface{} { return &ReverseProxy{client: new(fasthttp.HostClient)} },
	}
)

// NewProxyFromPool ...
func NewProxyFromPool(addr string) (*ReverseProxy, func()) {
	rp := defaultPool.Get().(*ReverseProxy)
	putback := func() {
		rp.client = nil
		defaultPool.Put(rp)
	}
	return rp, putback
}
