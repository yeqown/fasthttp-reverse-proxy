// Copyright yeqown The yeqown Author. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
package proxy

import (
	"testing"
)

func BenchmarkNewReverseProxy(b *testing.B) {
	proxy, _ := NewReverseProxyWith(WithAddress("localhost:8080"))
	if proxy == nil {
		b.Fatalf("could not get from pool, proxy is nil")
	}
	for i := 0; i < b.N; i++ {
		if proxy.getClient() == nil {
			b.Fatalf("could not get from pool, client is nil")
		}
		// fmt.Println(proxy.client.Addr)
	}
}

func BenchmarkNewReverseProxyWithBla(b *testing.B) {
	weigths := map[string]Weight{
		"localhost:8080": 10,
		"localhost:8081": 30,
		"localhost:8082": 60,
	}
	proxy, _ := NewReverseProxyWith(WithBalancer(weigths))
	if proxy == nil {
		b.Fatalf("could not get from pool, proxy is nil")
	}
	for i := 0; i < b.N; i++ {
		if proxy.getClient() == nil {
			b.Fatalf("could not get from pool, client is nil")
		}
	}
}

func Test_NewReverseProxy(t *testing.T) {
	proxy, _ := NewReverseProxyWith(WithAddress("https://www.baidu.com"))
	if proxy == nil {
		t.Error("failed create NewReverseProxyWith")
		t.FailNow()
	}
	client := proxy.getClient()
	if client == nil {
		t.Error("failed getClient")
		t.FailNow()
	}

	if client.Addr != "https://www.baidu.com" {
		t.Error("wrong init hostclient addr")
		t.FailNow()
	}
}

func Test_NewReversePorxyWithBalancer(t *testing.T) {
	weights := map[string]Weight{
		"http://localhost:9090": 20,
		"http://localhost:9091": 30,
		"http://localhost:9092": 50,
	}

	proxy, _ := NewReverseProxyWith(WithBalancer(weights))
	if proxy == nil {
		t.Error("failed create NewReverseProxyWith")
		t.FailNow()
	}
	client := proxy.getClient()
	if client == nil {
		t.Error("failed getClient")
		t.FailNow()
	}

	if client.Addr == "" {
		t.Error("wrong init hostclient addr")
		t.FailNow()
	}
}
