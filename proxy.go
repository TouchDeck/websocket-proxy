package main

import "log"

type proxy struct {
	agentServer  *server
	remoteServer *server
}

func (p proxy) listen() {
	go p.agentServer.listen()
	p.remoteServer.listen()
}

func newProxy(agentAddress string, remoteAddress string) *proxy {
	agentServer := newServer(agentAddress)

	remoteServer := newServer(remoteAddress)
	remoteServer.onClientConnected = func(c *client) {
		// Listen for one message with local address of agent to connect to.
		targetAgentAddr := readOneMessage(c)
		targetAgent := agentServer.getClient(c.remoteIp, targetAgentAddr)

		if targetAgent == nil {
			log.Println("Could not find target agent")
			c.close()
			return
		}
	}

	return &proxy{
		agentServer:  agentServer,
		remoteServer: remoteServer,
	}
}

func readOneMessage(c *client) string {
	// TODO
	return ""
}
