package main

import (
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

var (
	port = flag.Int("port", 8080, "assign the port of server listen")
)

func main() {
	flag.Parse()
	http.HandleFunc("/foo", func(w http.ResponseWriter, req *http.Request) {
		_ = req.ParseForm()
		timeout, err := strconv.Atoi(req.FormValue("timeout"))
		if err == nil && timeout > 0 {
			fmt.Println("timeout=", timeout)
			time.Sleep(time.Duration(timeout) * time.Second)
		}

		ip := req.RemoteAddr
		w.Header().Add("X-Test", "true")
		_, _ = fmt.Fprintf(w, "bar: %d, %s", 200, ip)
	})

	addr := fmt.Sprintf(":%d", *port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}
}
