package main

import (
	"github.com/TouchDeck/websocket-proxy/pkg/proxy"
	"log"
	"net/http"
)

var prx *proxy.Proxy

func main() {
	log.Println("Starting proxy...")
	prx = proxy.NewProxy("/ws")

	http.HandleFunc("/agents", discovery)

	// Start the HTTP server.
	log.Println("Starting HTTP server on port 8783")
	err := http.ListenAndServe(":8783", nil)
	if err != nil {
		log.Fatalln("Error starting websocket server:", err)
	}
}
