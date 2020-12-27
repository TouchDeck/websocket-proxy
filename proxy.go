package main

import (
	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
	"io"
	"log"
)

type proxy struct {
	agentServer  *websocketServer
	remoteServer *websocketServer
	agents       map[string]*websocketClient
}

func newProxy(basePath string) *proxy {
	p := &proxy{
		agentServer:  newWebsocketServer(basePath + "/agent"),
		remoteServer: newWebsocketServer(basePath + "/remote"),
		agents:       map[string]*websocketClient{},
	}

	p.agentServer.onClientConnected = func(agent *websocketClient) {
		// TODO: Use json, get more agent information (platform, hostname)
		// Listen for one message with the agent information.
		_, err := agent.readMessage()
		if err != nil {
			log.Println("Could not read agent message:", err)
			agent.close()
			return
		}
		log.Println("Agent client connected")

		// TODO: Use V5 to make sure the same client info results in the same id.
		agentId := uuid.Must(uuid.NewV4()).String()

		// Close the previous client, store the new one.
		if oldClient := p.agents[agentId]; oldClient != nil {
			oldClient.close()
		}
		p.agents[agentId] = agent

		agent.conn.WriteMessage(websocket.TextMessage, []byte(agentId))
	}

	p.remoteServer.onClientConnected = func(remote *websocketClient) {
		// Listen for one message with the id of the agent to connect to.
		agentId, err := remote.readMessage()
		if err != nil {
			log.Println("Could not read remote message:", err)
			remote.close()
			return
		}
		log.Println("Remote client connected")

		// Find the target agent.
		targetAgent := p.agents[agentId]
		if targetAgent == nil {
			log.Println("Could not find target agent")
			remote.close()
			return
		}

		// Pipe all data both ways.
		// TODO Stop piping if either client disconnects.

		go io.Copy(remote.conn.UnderlyingConn(), targetAgent.conn.UnderlyingConn())
		go io.Copy(targetAgent.conn.UnderlyingConn(), remote.conn.UnderlyingConn())
	}

	return p
}
