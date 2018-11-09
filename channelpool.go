package proxy

import (
	"errors"
	"fmt"
	"sync"
	// "github.com/valyala/fasthttp"
)

type chanPool struct {
	mutex   sync.RWMutex
	proxies chan *ReverseProxy
	factory Factory
}

// Factory the generator to creat ReverseProxy
type Factory func(string) (*ReverseProxy, error)

// NewChanPool to new a pool with some params
func NewChanPool(initialCap, maxCap int, factory Factory) (Pool, error) {
	if initialCap < 0 || maxCap <= 0 || initialCap > maxCap {
		return nil, errors.New("invalid capacity settings")
	}

	pool := &chanPool{
		proxies: make(chan *ReverseProxy, maxCap),
		factory: factory,
	}

	// create initial connections, if something goes wrong,
	// just close the pool error out.
	for i := 0; i < initialCap; i++ {
		proxy, err := factory("")
		if err != nil {
			proxy.Close()
			return nil, fmt.Errorf("factory is not able to fill the pool: %s", err)
		}
		pool.proxies <- proxy
	}

	return pool, nil
}

// getConnsAndFactory ...
func (p *chanPool) getConnsAndFactory() (chan *ReverseProxy, Factory) {
	p.mutex.RLock()
	proxies, factory := p.proxies, p.factory
	p.mutex.RUnlock()
	return proxies, factory
}

// Close close the pool
func (p *chanPool) Close() {
	p.mutex.Lock()
	proxies := p.proxies
	p.proxies = nil
	p.factory = nil
	p.mutex.Unlock()

	if proxies == nil {
		return
	}

	close(proxies)
	for proxy := range proxies {
		proxy.Close()
	}
}

// Get
func (p *chanPool) Get(addr string) (*ReverseProxy, error) {
	proxies, factory := p.getConnsAndFactory()
	if proxies == nil {
		return nil, ErrClosed
	}

	// wrap our connections with out custom net.Conn implementation (wrapConn
	// method) that puts the connection back to the pool if it's closed.
	select {
	case proxy := <-proxies:
		if &proxy == nil {
			return nil, ErrClosed
		}
		return proxy.SetClient(addr), nil
	default:
		proxy, err := factory(addr)
		if err != nil {
			return nil, err
		}
		return proxy, nil
	}
}

// Put ...
func (p *chanPool) Put(proxy *ReverseProxy) error {
	if proxy == nil {
		return errors.New("proxy is nil. rejecting")
	}

	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if p.proxies == nil {
		// pool is closed, close passed connection
		proxy.Close()
		return nil
	}

	// put the resource back into the pool. If the pool is full, this will
	// block and the default case will be executed.
	select {
	case p.proxies <- proxy:
		return nil
	default:
		// pool is full, close passed connection
		proxy.Close()
		return nil
	}
}

// Len ...
func (p *chanPool) Len() int {
	proxies, _ := p.getConnsAndFactory()
	return len(proxies)
}
