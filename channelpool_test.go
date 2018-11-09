package proxy

import "testing"

func Test_chanPool(t *testing.T) {
	factoty := func(addr string) (*ReverseProxy, error) {
		p := NewReverseProxy(addr)
		return p, nil
	}

	pool, err := NewChanPool(5, 100, factoty)
	if err != nil {
		t.Fatalf("could not make chan pool: %v", err)
	}

	t.Logf("len of pool is %d", pool.Len())

	p, err := pool.Get("localhost:8080")
	if err != nil {
		t.Fatalf("could not make chan pool: %v", err)
	}

	if p == nil {
		t.Fatalf("could not get one proxy form pool, proxy is nil")
	}

	t.Logf("proxy addr: %v and addr is: %s", p, p.client.Addr)
}
