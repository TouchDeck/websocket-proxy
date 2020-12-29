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
	agentsByRemoteIp map[string][]*agent
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
		agentsByRemoteIp: map[string][]*agent{},
	}

	p.agentServer.onClientConnected = p.onAgentConnected
	p.remoteServer.onClientConnected = p.onRemoteConnected

	return p
}

func (p *proxy) onAgentConnected(newClient *websocketClient) {
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

	// Store the agent by id and remote IP.
	// TODO: remove agent from lists on disconnect.
	p.agents[newAgent.Id] = newAgent
	p.agentsByRemoteIp[newClient.remoteIp] = append(p.agentsByRemoteIp[newClient.remoteIp], newAgent)

	reply, err := json.Marshal(newAgent)
	if err != nil {
		log.Println("Could not marshal agent response:", err)
		newClient.close()
		return
	}

	newClient.conn.WriteMessage(websocket.TextMessage, reply)
}

func (p *proxy) onRemoteConnected(remote *websocketClient) {
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

func (p *proxy) removeAgent(a *agent) {
	log.Println("Removing agent:", a.Id)

	// Delete the agent from the global agents list.
	delete(p.agents, a.Id)

	// Find and remove the agent from the local agents list.
	localAgents := p.agentsByRemoteIp[a.client.remoteIp]
	localI := -1
	for i, l := range localAgents {
		if l.Id == a.Id {
			localI = i
		}
	}
	if localI >= 0 {
		p.agentsByRemoteIp[a.client.remoteIp] = append(localAgents[:localI], localAgents[localI+1:]...)
	}
}
