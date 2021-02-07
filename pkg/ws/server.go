package ws

import (
	"github.com/TouchDeck/websocket-proxy/pkg/remoteIp"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

type Server struct {
	address           string
	onClientConnected func(c *Client)
}

var upgrader = websocket.Upgrader{
	// Allow all origins.
	CheckOrigin: func(_ *http.Request) bool {
		return true
	},
}

func (s *Server) handleClient(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Could not upgrade HTTP request:", err)
		return
	}

	newClient := &Client{
		conn:     conn,
		RemoteIp: remoteIp.FromRequest(r),
		Send:     make(chan Message),
		Recv:     make(chan Message),
	}

	// Start the read and write pump.
	go newClient.readPump()
	go newClient.writePump()

	s.onClientConnected(newClient)
}

func (s *Server) SetOnClientConnected(f func(newClient *Client)) {
	s.onClientConnected = f
}

func NewServer(mux *http.ServeMux, path string) *Server {
	s := &Server{
		onClientConnected: func(c *Client) {},
	}

	mux.HandleFunc(path, s.handleClient)

	return s
}
