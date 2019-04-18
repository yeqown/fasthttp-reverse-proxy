# fasthttp-reverse-proxy
[![Go Report Card](https://goreportcard.com/badge/github.com/yeqown/fasthttp-reverse-proxy)](https://goreportcard.com/report/github.com/yeqown/fasthttp-reverse-proxy) [![GoReportCard](https://godoc.org/github.com/yeqown/fasthttp-reverse-proxy?status.svg)](https://godoc.org/github.com/yeqown/fasthttp-reverse-proxy)

reverse http proxy hander based on fasthttp.

## features:

* [x] proxy client has `pool` supported

* [x] faster than golang standard `httputil.ReverseProxy`

* [x] simple warpper of `fasthttp.HostClient` 

## usage

* [use it alone](./examples/fasthttp-reverse-proxy/proxy.go)
* [use it with pool](./examples/fasthttp-reverse-proxy-with-pool/pool.go)

## contrast

`no proxy`:

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

`httputil.ReverseProxy`:

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

`yeqown.fasthttp.ReverseProxy`:

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

`yeqown.fasthttp.ReverseProxyWithPool`:

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

## links:

* [fasthttp](https://github.com/valyala/fasthttp)
* [standard httputil.ReverseProxy](https://golang.org/pkg/net/http/httputil/#ReverseProxy)
