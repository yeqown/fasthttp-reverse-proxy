package proxy

import (
	"testing"
)

/*
goos: darwin
goarch: amd64
pkg: github.com/yeqown/fasthttp-reverse-proxy
BenchmarkNewProxyFromPool-4   	10000000	       152 ns/op	     336 B/op	       2 allocs/op
PASS
ok  	github.com/yeqown/fasthttp-reverse-proxy	1.702s
Success: Benchmarks passed.
*/
func BenchmarkNewProxyFromPool(b *testing.B) {
	b.StopTimer()
	pool := NewPool()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		proxy, putbackFunc := pool.New("locahost:8080")
		if proxy == nil {
			b.Fatalf("could not get from pool, proxy is nil")
		}
		if proxy.client == nil {
			b.Fatalf("could not get from pool, client is nil")
		}
		// fmt.Println(proxy.client.Addr)
		putbackFunc()
	}
}
