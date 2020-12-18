package main

import (
	"bufio"
	"io"
	"log"
)

type proxy struct {
	agentServer  *server
	remoteServer *server
	agents       map[string]*client
}

func (p proxy) listen() {
	go p.agentServer.listen()
	p.remoteServer.listen()
}

func newProxy(agentAddress string, remoteAddress string) *proxy {
	p := &proxy{
		agentServer:  newServer(agentAddress),
		remoteServer: newServer(remoteAddress),
		agents:       map[string]*client{},
	}

	p.agentServer.onClientConnected = func(c *client) {
		// TODO: Use json, get more agent information (platform, hostname)
		// Listen for one message with the agent id.
		agentId, err := readOneMessage(c)
		if err != nil {
			c.close()
			return
		}

		// Close the previous client, store the new one.
		if oldClient := p.agents[c.remoteIp+"/"+agentId]; oldClient != nil {
			oldClient.close()
		}
		p.agents[c.remoteIp+"/"+agentId] = c
	}

	p.remoteServer.onClientConnected = func(c *client) {
		// Listen for one message with the id of the agent to connect to.
		agentId, err := readOneMessage(c)
		if err != nil {
			c.close()
			return
		}

		// Find the target agent.
		targetAgent := p.agents[c.remoteIp+"/"+agentId]
		if targetAgent == nil {
			log.Println("Could not find target agent")
			c.close()
			return
		}

		go io.Copy(c.conn, targetAgent.conn)
		go io.Copy(targetAgent.conn, c.conn)
	}

	return p
}

func readOneMessage(c *client) (string, error) {
	reader := bufio.NewReader(c.conn)
	msg, err := reader.ReadString('\n')
	if err != nil {
		log.Println("Could not read message from client:", err)
	}
	return msg, err
}
