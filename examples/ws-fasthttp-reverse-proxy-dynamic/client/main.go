package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/fasthttp/websocket"
)

var addr = flag.String("addr", "localhost:8081", "http service address")

func openChat(url string, wg *sync.WaitGroup, interrupt chan os.Signal) {
	defer wg.Done()
	log.Printf("connecting to %s", url)
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	fruits := []string{
		"apples",
		"oranges",
	}

	var wg sync.WaitGroup

	for i := range fruits {
		wg.Add(1)
		fruitChat := fruits[i]
		u := url.URL{Scheme: "ws", Host: *addr, Path: "/echo", RawQuery: fmt.Sprintf("fruit=%s&q=secret", fruitChat)}
		go openChat(u.String(), &wg, interrupt)
	}
	wg.Wait()
}
