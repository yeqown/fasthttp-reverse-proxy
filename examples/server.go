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
	addr := fmt.Sprintf(":%d", *port)

	http.HandleFunc("/foo", func(w http.ResponseWriter, req *http.Request) {
		clientIP := req.Header.Get("X-Forwarded-For")
		fmt.Printf("got request from %s\n", clientIP)

		_ = req.ParseForm()
		timeout, err := strconv.Atoi(req.FormValue("timeout"))
		if err == nil && timeout > 0 {
			fmt.Println("timeout=", timeout)
			time.Sleep(time.Duration(timeout) * time.Second)
		}

		w.Header().Add("X-Test", "true")
		_, _ = fmt.Fprintf(w, "response from %s", addr)
	})

	fmt.Printf("listening on %s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}
}
