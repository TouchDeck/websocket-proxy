package main

import (
	"github.com/TouchDeck/websocket-proxy/pkg/proxy"
	"log"
	"net/http"
)

var prx *proxy.Proxy

func main() {
	log.Println("Starting proxy...")
	mux := http.NewServeMux()
	prx = proxy.NewProxy(mux, "/ws")

	mux.HandleFunc("/agents", discovery)

	// Start the HTTP server.
	log.Println("Starting HTTP server on port 8783")
	err := http.ListenAndServe(":8783", mux)
	if err != nil {
		log.Fatalln("Error starting websocket server:", err)
	}
}
