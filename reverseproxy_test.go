// Copyright yeqown The yeqown Author. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
package proxy

import (
	"testing"
)

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
