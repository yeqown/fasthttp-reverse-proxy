package proxy

import (
	"testing"
)

/*
goos: darwin
goarch: amd64
pkg: github.com/yeqown/fasthttp-reverse-proxy
BenchmarkNewReverseProxy-4   	200000000	         9.16 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	github.com/yeqown/fasthttp-reverse-proxy	2.821s
Success: Benchmarks passed.
*/
func BenchmarkNewReverseProxy(b *testing.B) {
	for i := 0; i < b.N; i++ {
		proxy := NewReverseProxy("localhost:8080")
		if proxy == nil {
			b.Fatalf("could not get from pool, proxy is nil")
		}
		if proxy.client == nil {
			b.Fatalf("could not get from pool, client is nil")
		}
		// fmt.Println(proxy.client.Addr)
	}
}
