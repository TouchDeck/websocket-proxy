package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

type websocketClient struct {
	conn     *websocket.Conn
	remoteIp string
}

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

func (c *websocketClient) close() {
	c.conn.Close()
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
	}
	s.onClientConnected(newClient)
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
