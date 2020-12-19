package main

import (
	"bufio"
	"log"
	"net"
)

type client struct {
	conn     net.Conn
	remoteIp string
}

type server struct {
	onClientConnected func(c *client)
}

func (c *client) close() {
	c.conn.Close()
}

func (c *client) readMessage() (string, error) {
	reader := bufio.NewReader(c.conn)
	msg, err := reader.ReadString('\n')
	if err == nil {
		msg = msg[:len(msg)-1]
	}
	return msg, err
}

func (s *server) listen(address string) {
	log.Println("Starting TCP server on:", address)
	listener, err := net.Listen("tcp", address)
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
			conn:     conn,
			remoteIp: getRemoteIp(conn.RemoteAddr().String()),
		}
		s.onClientConnected(newClient)
	}
}

func newServer() *server {
	return &server{
		onClientConnected: func(c *client) {},
	}
}
