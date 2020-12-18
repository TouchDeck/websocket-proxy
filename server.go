package main

import (
	"log"
	"net"
	"strings"
)

type client struct {
	conn      net.Conn
	serv      *server
	remoteIp  string
	localAddr string
}

type server struct {
	address              string
	clients              map[string][]*client
	onClientConnected    func(c *client)
	onClientDisconnected func(c *client)
}

func (c *client) close() {
	c.conn.Close()
	c.serv.removeClient(c)
}

func (s *server) removeClient(remove *client) {
	var newClients []*client
	for _, c := range s.clients[remove.remoteIp] {
		if c.localAddr != remove.localAddr {
			newClients = append(newClients, c)
		}
	}
	s.clients[remove.remoteIp] = newClients
}

func (s *server) getClient(remoteIp string, localAddr string) *client {
	for _, c := range s.clients[remoteIp] {
		if c.localAddr == localAddr {
			return c
		}
	}
	return nil
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
			conn:      conn,
			serv:      s,
			remoteIp:  getRemoteIp(conn),
			localAddr: conn.LocalAddr().String(),
		}

		// Store the client by its remote IP address.
		remoteIp := getRemoteIp(conn)
		s.clients[remoteIp] = append(s.clients[remoteIp], newClient)

		s.onClientConnected(newClient)
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
