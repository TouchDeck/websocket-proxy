package main

import (
	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
	"io"
	"log"
)

type proxy struct {
	agentServer      *websocketServer
	remoteServer     *websocketServer
	agents           map[string]*agent
	agentsByPublicIp map[string][]*agent
}

type agent struct {
	Id           string `json:"id"`
	Version      string `json:"version"`
	LocalAddress string `json:"localAddress"`
	Platform     string `json:"platform"`
	Hostname     string `json:"hostname"`
	client       *websocketClient
}

func newProxy(basePath string) *proxy {
	p := &proxy{
		agentServer:      newWebsocketServer(basePath + "/agent"),
		remoteServer:     newWebsocketServer(basePath + "/remote"),
		agents:           map[string]*agent{},
		agentsByPublicIp: map[string][]*agent{},
	}

	p.agentServer.onClientConnected = func(newClient *websocketClient) {
		// TODO: Use json, get more agent information (platform, hostname)
		// Listen for one message with the agent information.
		_, err := newClient.readMessage()
		if err != nil {
			log.Println("Could not read agent message:", err)
			newClient.close()
			return
		}
		log.Println("Agent client connected")

		// TODO: Use V5 to make sure the same client info results in the same id.
		agentId := uuid.Must(uuid.NewV4()).String()

		// Close the previous client, store the new one.
		if oldClient := p.agents[agentId]; oldClient != nil {
			oldClient.client.close()
		}

		// TODO
		newAgent := &agent{
			Id:           agentId,
			Version:      "1.0.0",
			LocalAddress: "192.168.0.10",
			Platform:     "windows",
			Hostname:     "host",
			client:       newClient,
		}

		// Store the agent by id and public ip.
		p.agents[agentId] = newAgent
		p.agentsByPublicIp[newClient.remoteIp] = append(p.agentsByPublicIp[newClient.remoteIp], newAgent)

		newClient.conn.WriteMessage(websocket.TextMessage, []byte(agentId))
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
			log.Println("Could not find target agent:", agentId)
			remote.close()
			return
		}

		// Pipe all data both ways.
		go func() {
			_, err := io.Copy(remote.conn.UnderlyingConn(), targetAgent.client.conn.UnderlyingConn())
			if err != nil {
				log.Println("Error piping remote -> agent:", err)
			}
		}()
		go func() {
			_, err := io.Copy(targetAgent.client.conn.UnderlyingConn(), remote.conn.UnderlyingConn())
			if err != nil {
				log.Println("Error piping agent -> remote:", err)
			}
		}()
	}

	return p
}
