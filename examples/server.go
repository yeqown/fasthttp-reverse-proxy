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
	http.HandleFunc("/foo", func(w http.ResponseWriter, req *http.Request) {
		ip := req.RemoteAddr
		w.Header().Add("X-Test", "true")
		fmt.Fprintf(w, "bar: %d, %s", 200, ip)
	})

	addr := fmt.Sprintf(":%d", *port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}
}
