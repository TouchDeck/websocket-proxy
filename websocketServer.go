package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strings"
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

func (c *websocketClient) readMessage() (string, error) {
	_, msg, err := c.conn.ReadMessage()
	return string(msg), err
}

func (s *websocketServer) handleClient(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Could not upgrade HTTP request:", err)
		return
	}

	newClient := &websocketClient{
		conn:     conn,
		remoteIp: getRemoteIp(r.RemoteAddr),
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

func getRemoteIp(address string) string {
	var ip string

	if address[0] == '[' {
		// IPv6: [<ip>]:<port>
		ip = address[1:strings.IndexRune(address, ']')]
	} else {
		// IPv4: <ip>:<port>
		ip = strings.Split(address, ":")[0]
	}

	// Make sure the IP address is always the same when coming from localhost.
	if ip == "127.0.0.1" || ip == "::ffff:127.0.0.1" {
		ip = "::1"
	}

	return ip
}
