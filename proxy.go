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

type agentId struct {
	Id string `json:"id"`
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

		reply, err := json.Marshal(newAgent)
		if err != nil {
			log.Println("Could not marshal agent response:", err)
			newClient.close()
			return
		}

		// TODO: check for and remove on error.
		newClient.conn.WriteMessage(websocket.TextMessage, reply)
	}

	p.remoteServer.onClientConnected = func(remote *websocketClient) {
		// Listen for one message with the id of the agent to connect to.
		_, msg, err := remote.conn.ReadMessage()
		if err != nil {
			log.Println("Could not read remote message:", err)
			remote.close()
			return
		}

		target := &agentId{}
		err = json.Unmarshal(msg, target)
		if err != nil {
			log.Println("Could not unmarshal remote message:", err)
			remote.close()
			return
		}

		// Find the target agent.
		targetAgent := p.agents[target.Id]
		if targetAgent == nil {
			log.Println("Could not find target agent:", target.Id)
			remote.close()
			return
		}

		log.Println("Remote client connected:", target.Id)

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
