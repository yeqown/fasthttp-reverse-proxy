package proxy

import "errors"

var (
	// ErrClosed is the error resulting if the pool is closed via pool.Close().
	ErrClosed = errors.New("pool is closed")
)

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
