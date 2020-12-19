package main

import (
	"io"
	"log"
)

type proxy struct {
	agentServer  *server
	remoteServer *websocketServer
	agents       map[string]*client
}

func (p proxy) listen(agentServerAddr string, remoteServerAddr string) {
	go p.agentServer.listen(agentServerAddr)
	p.remoteServer.listen(remoteServerAddr)
}

func newProxy(wsPath string) *proxy {
	p := &proxy{
		agentServer:  newServer(),
		remoteServer: newWebsocketServer(wsPath),
		agents:       map[string]*client{},
	}

	p.agentServer.onClientConnected = func(c *client) {
		// TODO: Use json, get more agent information (platform, hostname)
		// Listen for one message with the agent id.
		agentId, err := c.readMessage()
		if err != nil {
			log.Println("Could not read TCP message:", err)
			c.close()
			return
		}
		log.Println("Agent client connected")
		log.Println(c.remoteIp + "/" + agentId)

		// Close the previous client, store the new one.
		if oldClient := p.agents[c.remoteIp+"/"+agentId]; oldClient != nil {
			oldClient.close()
		}
		p.agents[c.remoteIp+"/"+agentId] = c
	}

	p.remoteServer.onClientConnected = func(c *websocketClient) {
		// Listen for one message with the id of the agent to connect to.
		agentId, err := c.readMessage()
		if err != nil {
			log.Println("Could not read websocket message:", err)
			c.close()
			return
		}
		log.Println("Remote client connected")
		log.Println(c.remoteIp + "/" + agentId)

		// Find the target agent.
		targetAgent := p.agents[c.remoteIp+"/"+agentId]
		if targetAgent == nil {
			log.Println("Could not find target agent")
			c.close()
			return
		}

		// TODO: this is not how this works apparently :')
		go io.Copy(c.conn.UnderlyingConn(), targetAgent.conn)
		go io.Copy(targetAgent.conn, c.conn.UnderlyingConn())
	}

	return p
}
