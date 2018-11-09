package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/foo", func(w http.ResponseWriter, req *http.Request) {
		ip := req.RemoteAddr
		// fmt.Println(ip)
		w.Header().Add("X-Test", "true")
		fmt.Fprintf(w, "bar: %d, %s", 200, ip)
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
