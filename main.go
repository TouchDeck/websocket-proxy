package main

import (
	"github.com/TouchDeck/websocket-proxy/pkg/proxy"
	"github.com/rs/cors"
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
	handler := cors.AllowAll().Handler(mux)
	err := http.ListenAndServe(":8783", handler)
	if err != nil {
		log.Fatalln("Error starting websocket server:", err)
	}
}
