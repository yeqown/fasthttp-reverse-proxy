package main

import (
	"flag"
	"fmt"
	"net/http"
)

var (
	port = flag.Int("port", 8080, "assign the port of server listen")
)

func main() {
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/foo", func(w http.ResponseWriter, req *http.Request) {
		fmt.Println("a request incoming")
		ip := req.RemoteAddr
		w.Header().Add("X-Test", "true")
		_, _ = fmt.Fprintf(w, "bar: %d, %s", 200, ip)
	})

	addr := fmt.Sprintf(":%d", *port)
	svr := http.Server{
		Addr:    addr,
		Handler: mux,
		//TLSConfig: &tls.Config{
		//	InsecureSkipVerify: true, // 忽略对客户端的认证
		//},
	}

	if err := svr.
		ListenAndServeTLS("../selfsigned.crt", "../selfsigned.key"); err != nil {
		panic(err)
	}
}
