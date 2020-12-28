package main

import (
	"encoding/json"
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
	Id     string                 `json:"id"`
	Meta   map[string]interface{} `json:"meta"`
	client *websocketClient
}

func newProxy(basePath string) *proxy {
	p := &proxy{
		agentServer:      newWebsocketServer(basePath + "/agent"),
		remoteServer:     newWebsocketServer(basePath + "/remote"),
		agents:           map[string]*agent{},
		agentsByPublicIp: map[string][]*agent{},
	}

	p.agentServer.onClientConnected = func(newClient *websocketClient) {
		// Listen for one message with the agent information.
		_, msg, err := newClient.conn.ReadMessage()
		if err != nil {
			log.Println("Could not read agent message:", err)
			newClient.close()
			return
		}

		newAgent := &agent{
			client: newClient,
			Meta:   map[string]interface{}{},
		}
		err = json.Unmarshal(msg, newAgent)
		if err != nil {
			log.Println("Could not unmarshal agent message:", err)
			newClient.close()
			return
		}

		newAgent.Id = uuid.Must(uuid.NewV4()).String()
		log.Println("Agent client connected:", newAgent.Id)

		// Store the agent by id and public ip.
		p.agents[newAgent.Id] = newAgent
		p.agentsByPublicIp[newClient.remoteIp] = append(p.agentsByPublicIp[newClient.remoteIp], newAgent)

		newClient.conn.WriteMessage(websocket.TextMessage, []byte(newAgent.Id))
	}

	p.remoteServer.onClientConnected = func(remote *websocketClient) {
		// Listen for one message with the id of the agent to connect to.
		_, msg, err := remote.conn.ReadMessage()
		if err != nil {
			log.Println("Could not read remote message:", err)
			remote.close()
			return
		}
		agentId := string(msg)

		// Find the target agent.
		targetAgent := p.agents[agentId]
		if targetAgent == nil {
			log.Println("Could not find target agent:", agentId)
			remote.close()
			return
		}

		log.Println("Remote client connected:", agentId)

		// TODO: remove agent from lists on disconnect.

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
