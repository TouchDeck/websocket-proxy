package main

import (
	"log"
	"net/http"
)

var prx *proxy

func main() {
	log.Println("Starting proxy...")
	prx = newProxy("/ws")

	http.HandleFunc("/agents", discovery)

	// Start the HTTP server.
	log.Println("Starting HTTP server on port 8783")
	err := http.ListenAndServe(":8783", nil)
	if err != nil {
		log.Fatalln("Error starting websocket server:", err)
	}
}
