package proxy

// Copyright 2018 The yeqown Author. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

import (
	"errors"
	"sync"
)

var (
	errFactoryNotHelp         = errors.New("factory is not able to fill the pool")
	errInvalidCapacitySetting = errors.New("invalid capacity settings")
)

// Pool interface impelement based on channel
// there is a channel to contain ReverseProxy object,
// and provide Get and Put method to handle with RevsereProxy
type chanPool struct {
	// mutex makes the chanPool woking with goroutine safely
	mutex sync.RWMutex

	// reverseProxyChan chan of getting the *ReverseProxy and putting it back
	reverseProxyChan chan *ReverseProxy

	// factory is factory method to generate ReverseProxy
	// this can be customized
	factory Factory
}

// Factory the generator to creat ReverseProxy
type Factory func(string) (*ReverseProxy, error)

// NewChanPool to new a pool with some params
func NewChanPool(initialCap, maxCap int, factory Factory) (Pool, error) {
	if initialCap < 0 || maxCap <= 0 || initialCap > maxCap {
		return nil, errInvalidCapacitySetting
	}

	// initialize the chanPool
	pool := &chanPool{
		mutex:            sync.RWMutex{},
		reverseProxyChan: make(chan *ReverseProxy, maxCap),
		factory:          factory,
	}

	// create initial connections, if something goes wrong,
	// just close the pool error out.
	for i := 0; i < initialCap; i++ {
		proxy, err := factory("")
		if err != nil {
			proxy.Close()
			return nil, errFactoryNotHelp
		}
		pool.reverseProxyChan <- proxy
	}

	return pool, nil
}

// getConnsAndFactory ... get a copy of chanPool's reverseProxyChan and factory
func (p *chanPool) getConnsAndFactory() (chan *ReverseProxy, Factory) {
	p.mutex.RLock()
	reverseProxyChan, factory := p.reverseProxyChan, p.factory
	p.mutex.RUnlock()
	return reverseProxyChan, factory
}

// Close close the pool
func (p *chanPool) Close() {
	p.mutex.Lock()
	reverseProxyChan := p.reverseProxyChan
	p.reverseProxyChan = nil
	p.factory = nil
	p.mutex.Unlock()

	if reverseProxyChan == nil {
		return
	}

	close(reverseProxyChan)
	for proxy := range reverseProxyChan {
		proxy.Close()
	}
}

// Get a *ReverseProxy from pool, it will get an error while
// reverseProxyChan is nil or pool has been closed
func (p *chanPool) Get(addr string) (*ReverseProxy, error) {
	// reverseProxyChan, factory := p.getConnsAndFactory()
	// if reverseProxyChan == nil {
	// return nil, ErrClosed
	// }

	if p.reverseProxyChan == nil {
		return nil, errClosed
	}

	// wrap our connections with out custom net.Conn implementation (wrapConn
	// method) that puts the connection back to the pool if it's closed.
	select {
	case proxy := <-p.reverseProxyChan:
		// FIXME: judge empty proxy correctly
		if &proxy == nil {
			return nil, errClosed
		}
		return proxy.SetClient(addr), nil
	default:
		proxy, err := p.factory(addr)
		if err != nil {
			return nil, err
		}
		return proxy, nil
	}
}

// Put ... put a *ReverseProxy object back into chanPool
func (p *chanPool) Put(proxy *ReverseProxy) error {
	if proxy == nil {
		return errors.New("proxy is nil. rejecting")
	}

	// p.mutex.RLock()
	// defer p.mutex.RUnlock()

	if p.reverseProxyChan == nil {
		// pool is closed, close passed connection
		proxy.Close()
		return nil
	}

	// put the resource back into the pool. If the pool is full, this will
	// block and the default case will be executed.
	select {
	case p.reverseProxyChan <- proxy:
		return nil
	default:
		// pool is full, close passed connection
		proxy.Close()
		return nil
	}
}

// Len get chanPool channel length
func (p *chanPool) Len() int {
	reverseProxyChan, _ := p.getConnsAndFactory()
	return len(reverseProxyChan)
}
