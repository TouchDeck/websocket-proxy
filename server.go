package main

import (
	"bufio"
	"log"
	"net"
)

type tcpClient struct {
	conn     net.Conn
	remoteIp string
}

type tcpServer struct {
	onClientConnected func(c *tcpClient)
}

func (c *tcpClient) close() {
	c.conn.Close()
}

func (c *tcpClient) readMessage() (string, error) {
	reader := bufio.NewReader(c.conn)
	msg, err := reader.ReadString('\n')
	if err == nil {
		msg = msg[:len(msg)-1]
	}
	return msg, err
}

func (s *tcpServer) listen(address string) {
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

		newClient := &tcpClient{
			conn:     conn,
			remoteIp: getRemoteIp(conn.RemoteAddr().String()),
		}
		s.onClientConnected(newClient)
	}
}

func newTcpServer() *tcpServer {
	return &tcpServer{
		onClientConnected: func(c *tcpClient) {},
	}
}
