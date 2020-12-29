package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

type websocketServer struct {
	address           string
	onClientConnected func(c *websocketClient)
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(_ *http.Request) bool {
		// TODO: Actually check origin
		return true
	},
}

func (s *websocketServer) handleClient(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Could not upgrade HTTP request:", err)
		return
	}

	newClient := &websocketClient{
		conn:     conn,
		remoteIp: remoteIpFromRequest(r),
		send:     make(chan message),
		recv:     make(chan message),
	}

	// The write pump needs to start first, so onClientConnected can send messages.
	// The read pump needs to start last, to prevent race conditions with
	// onClientConnected while reading initialization messages.
	go newClient.writePump()
	s.onClientConnected(newClient)
	go newClient.readPump()
}

func newWebsocketServer(path string) *websocketServer {
	s := &websocketServer{
		onClientConnected: func(c *websocketClient) {},
	}

	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		s.handleClient(w, r)
	})

	return s
}
