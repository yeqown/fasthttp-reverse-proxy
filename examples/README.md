### fasthttp proxy and std http proxy perform

files content are following

> server.go
```go
package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/foo", func(w http.ResponseWriter, req *http.Request) {
		ip := req.Header.Get("X-Real-Ip")
		// fmt.Println(ip)
		w.Header().Add("X-Test", "true")
		fmt.Fprintf(w, "bar: %d, %s", 200, ip)
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

```

> httpReverseProxy.go
```go
package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

var (
	URL   *url.URL
	proxy *httputil.ReverseProxy
)

func init() {
	URL, _ = url.Parse("http://localhost:8080")
	proxy = httputil.NewSingleHostReverseProxy(URL)
}

func main() {
	http.HandleFunc("/foo", func(w http.ResponseWriter, req *http.Request) {
		proxy.ServeHTTP(w, req)
	})

	if err := http.ListenAndServe(":8082", nil); err != nil {
		panic(err)
	}
}

```

> fasthttpProxy.go
```go
package main

import (
	"github.com/valyala/fasthttp"
	proxy "github.com/yeqown/fasthttp-reverse-proxy"
)

var (
	proxyServer = proxy.NewReverseProxy("localhost:8080")
)

// ProxyHandler ...
func ProxyHandler(ctx *fasthttp.RequestCtx) {
	// all proxy to localhost
	proxyServer.ServeHTTP(ctx)
}

func main() {
	fasthttp.ListenAndServe(":8081", ProxyHandler)
}

```

### 1. Server alone (Without any proxy)

```sh
➜  examples git:(master) ✗ bombardier -c 125 -t 10s localhost:8080/foo
Bombarding http://localhost:8080/foo for 10s using 125 connection(s)

Done!
Statistics        Avg      Stdev        Max
  Reqs/sec     44720.43   10528.44   59074.06
  Latency        2.80ms     2.91ms   171.62ms
  HTTP codes:
    1xx - 0, 2xx - 446323, 3xx - 0, 4xx - 0, 5xx - 0
    others - 0
  Throughput:     8.77MB/s%
```

### 2. Server with http.Proxy

```sh
➜  examples git:(master) ✗ bombardier -c 125 -t 10s localhost:8082/foo
Bombarding http://localhost:8082/foo for 10s using 125 connection(s)

Done!
Statistics        Avg      Stdev        Max
  Reqs/sec      5772.12    1370.75    9986.58
  Latency       21.66ms     9.32ms   173.96ms
  HTTP codes:
    1xx - 0, 2xx - 57731, 3xx - 0, 4xx - 0, 5xx - 0
    others - 0
  Throughput:     1.13MB/s%
```

### 3. Server with fasthttp.Proxy (global proxy varible)

```sh
➜  examples git:(master) ✗ bombardier -c 125 -t 10s localhost:8081/foo
Bombarding http://localhost:8081/foo for 10s using 125 connection(s)

Done!
Statistics        Avg      Stdev        Max
  Reqs/sec     29587.28    3392.08   37173.86
  Latency        4.22ms     2.45ms   149.31ms
  HTTP codes:
    1xx - 0, 2xx - 295646, 3xx - 0, 4xx - 0, 5xx - 0
    others - 0
  Throughput:     6.31MB/s%
```

### 4. Server with fasthttp ProxyPool (poolConfig, initialCap:10, maxCap:100)

link to [Code](pool/pool.go)

```sh
➜  ~ bombardier -c 125 -t 10s localhost:8083/foo
Bombarding http://localhost:8083/foo for 10s using 125 connection(s)

Done!
Statistics        Avg      Stdev        Max
  Reqs/sec     11914.93    1935.58   16369.31
  Latency       10.48ms     2.23ms    72.94ms
  HTTP codes:
    1xx - 0, 2xx - 118995, 3xx - 0, 4xx - 0, 5xx - 0
    others - 0
  Throughput:     2.71MB/s%
```

> Notice: some close work maybe cause bug, but result here in testing DOSE NOT do this.