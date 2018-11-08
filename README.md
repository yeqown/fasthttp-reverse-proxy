# fasthttp-reverse-proxy
reverse http proxy based on fasthttp

currently, it's so simple ~
```go
// ReverseProxy ...
type ReverseProxy struct {
	client *fasthttp.HostClient
}
```
