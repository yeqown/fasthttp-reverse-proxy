package proxy

import "crypto/tls"

// Option to define all options to reverse http proxy
type Option interface {
	apply(o *buildOption)
}

// buildOption .
type buildOption struct {
	openBalance bool     // denote whether the balancer is configured or not
	weights     []W      // weight of each upstream server
	addresses   []string // all upstream server address
	tlsConfig   *tls.Config
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
