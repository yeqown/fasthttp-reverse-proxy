package proxy

import (
	"crypto/tls"
	"time"
)

// Option to define all options to reverse http proxy.
type Option interface {
	apply(o *buildOption)
}

// buildOption contains all fields those are used in ReverseProxy.
type buildOption struct {
	// logger to log some info
	logger __Logger
	// debug to open debug mode to log more info to logger
	debug bool

	// openBalance denote whether the balancer is configured or not.
	openBalance bool

	// weights weight of each upstream server. it would be empty if openBalance not true.
	weights []W

	// addresses all upstream server address. if openBalance not true,
	// addresses will keep the only one upstream server address in addresses[0].
	addresses []string

	// tlsConfig is pointer to tls.Config, will be used if the upstream.
	// need TLS handshake
	tlsConfig *tls.Config

	// timeout specify the timeout context with each request.
	timeout time.Duration

	// disablePathNormalizing disable path normalizing.
	disablePathNormalizing bool

	// maxConnDuration of hostClient
	maxConnDuration time.Duration
}

func defaultBuildOption() *buildOption {
	return &buildOption{
		logger:                 &nopLogger{},
		debug:                  false,
		openBalance:            false,
		weights:                nil,
		addresses:              nil,
		tlsConfig:              nil,
		timeout:                0,
		disablePathNormalizing: false,
		maxConnDuration:        0,
	}
}

type funcBuildOption struct {
	f func(o *buildOption)
}

func newFuncBuildOption(f func(o *buildOption)) funcBuildOption { return funcBuildOption{f: f} }
func (fb funcBuildOption) apply(o *buildOption)                 { fb.f(o) }

func WithTLSConfig(config *tls.Config) Option {
	return newFuncBuildOption(func(o *buildOption) {
		o.tlsConfig = config
	})
}

// WithTLS build tls.Config with server certFile and keyFile.
// tlsConfig is nil as default
func WithTLS(certFile, keyFile string) Option {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		panic("" + err.Error())
	}

	return WithTLSConfig(&tls.Config{
		Certificates: []tls.Certificate{cert},
	})
}

// WithAddress generate address options
func WithAddress(addresses ...string) Option {
	return newFuncBuildOption(func(o *buildOption) {
		o.addresses = addresses
	})
}

// WithBalancer generate balancer options
func WithBalancer(addrWeights map[string]Weight) Option {
	weights := make([]W, 0, len(addrWeights))
	addresses := make([]string, 0, len(addrWeights))
	for k, v := range addrWeights {
		weights = append(weights, v)
		addresses = append(addresses, k)
	}

	return newFuncBuildOption(func(o *buildOption) {
		o.addresses = addresses
		o.openBalance = true
		o.weights = weights
	})
}

func WithDebug() Option {
	return newFuncBuildOption(func(o *buildOption) {
		o.debug = true
	})
}

// WithTimeout specify the timeout of each request
func WithTimeout(d time.Duration) Option {
	return newFuncBuildOption(func(o *buildOption) {
		o.timeout = d
	})
}

// WithDisablePathNormalizing sets whether disable path normalizing.
func WithDisablePathNormalizing(isDisablePathNormalizing bool) Option {
	return newFuncBuildOption(func(o *buildOption) {
		o.disablePathNormalizing = isDisablePathNormalizing
	})
}

// WithMaxConnDuration sets maxConnDuration of hostClient, which
// means keep-alive connections are closed after this duration.
func WithMaxConnDuration(d time.Duration) Option {
	return newFuncBuildOption(func(o *buildOption) {
		o.maxConnDuration = d
	})
}
