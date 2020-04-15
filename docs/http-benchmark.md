# HTTP Benchmark

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

``yeqown.fasthttp.ReverseProxy (WithBalancer)`:`
```sh
➜  examples git:(feat-LB-support) ✗ bombardier -c 125 -t 10s localhost:8081/foo
Bombarding http://localhost:8081/foo for 10s using 125 connection(s)

Done!
Statistics        Avg      Stdev        Max
  Reqs/sec     12433.93    4055.02   20192.45
  Latency       10.12ms     8.81ms   242.13ms
  HTTP codes:
    1xx - 0, 2xx - 123328, 3xx - 0, 4xx - 0, 5xx - 139 (too many sockets)
    others - 0
  Throughput:     2.81MB/s%
```

`yeqown.fasthttp.ReverseProxyWithPool(init=100, max=200)`:

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
