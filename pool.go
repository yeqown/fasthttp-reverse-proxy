package proxy

import (
	"errors"

	"github.com/valyala/fasthttp"
)

var (
	// errClosed is the error resulting if the pool is closed via pool.Close().
	errClosed = errors.New("pool is closed")
)

// Proxier can be HTTP or WebSocket proxier
// TODO:
type Proxier interface {
	ServeHTTP(ctx *fasthttp.RequestCtx)
	// ?
	SetClient(addr string) Proxier

	// Reset .
	Reset()

	// Close .
	Close()
}

// Pool interface ...
// this interface ref to: https://github.com/fatih/pool/blob/master/pool.go
type Pool interface {
	// Get returns a new ReverseProxy from the pool.
	Get(string) (*ReverseProxy, error)

	// Put Reseting the ReverseProxy puts it back to the Pool.
	Put(*ReverseProxy) error

	// Close closes the pool and all its connections. After Close() the pool is
	// no longer usable.
	Close()

	// Len returns the current number of connections of the pool.
	Len() int
}
