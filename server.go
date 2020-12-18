package main

import (
	"log"
	"net"
	"strings"
)

type client struct {
	conn net.Conn
	serv *server
}

type server struct {
	address           string
	clients           map[string][]*client
	onClientConnected func(c *client)
}

func (s *server) listen() {
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		log.Fatal("Error starting TCP server:", err)
	}
	defer listener.Close()

	for {
		// Accept all incoming connections.
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error while accepting client: ", err)
			continue
		}

		newClient := &client{
			conn: conn,
			serv: s,
		}
		s.onClientConnected(newClient)

		// Store the client by its remote IP address.
		remoteIp := getRemoteIp(conn)
		s.clients[remoteIp] = append(s.clients[remoteIp], newClient)
	}
}

func newServer(address string) *server {
	log.Println("Starting server on:", address)
	return &server{
		address:           address,
		clients:           map[string][]*client{},
		onClientConnected: func(c *client) {},
	}
}

func getRemoteIp(conn net.Conn) string {
	return strings.Split(conn.RemoteAddr().String(), ":")[0]
}
