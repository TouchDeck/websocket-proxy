package main

import (
	"github.com/gorilla/websocket"
	"log"
)

type proxy struct {
	agentServer  *tcpServer
	remoteServer *websocketServer
	agents       map[string]*tcpClient
}

func (p proxy) listen(agentServerAddr string, remoteServerAddr string) {
	go p.agentServer.listen(agentServerAddr)
	p.remoteServer.listen(remoteServerAddr)
}

func newProxy(wsPath string) *proxy {
	p := &proxy{
		agentServer:  newTcpServer(),
		remoteServer: newWebsocketServer(wsPath),
		agents:       map[string]*tcpClient{},
	}

	p.agentServer.onClientConnected = func(c *tcpClient) {
		// TODO: Use json, get more agent information (platform, hostname)
		// Listen for one message with the agent id.
		agentId, err := c.readMessage()
		if err != nil {
			log.Println("Could not read TCP message:", err)
			c.close()
			return
		}
		log.Println("Agent client connected")

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

		// Find the target agent.
		targetAgent := p.agents[c.remoteIp+"/"+agentId]
		if targetAgent == nil {
			log.Println("Could not find target agent")
			c.close()
			return
		}

		// Pipe all data both ways.
		// TODO Stop piping if either client disconnects.
		go func() {
			for {
				_, msg, err := c.conn.ReadMessage()
				if err != nil {
					log.Println("Error while reading websocket message:", err)
					return
				}
				_, err = targetAgent.conn.Write(msg)
				if err != nil {
					log.Println("Error while writing TCP message:", err)
					return
				}
			}
		}()
		go func() {
			for {
				msg, err := targetAgent.readMessage()
				if err != nil {
					log.Println("Error while reading TCP message:", err)
					return
				}
				err = c.conn.WriteMessage(websocket.TextMessage, []byte(msg))
				if err != nil {
					log.Println("Error while writing websocket message:", err)
					return
				}
			}
		}()
	}

	return p
}
