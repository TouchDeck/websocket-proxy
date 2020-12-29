package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
	"log"
)

// TODO: Use channels to register/remove agents to prevent race conditions
// See: https://github.com/gorilla/websocket/blob/master/examples/chat/hub.go
type proxy struct {
	agentServer      *websocketServer
	remoteServer     *websocketServer
	agents           map[string]*agent
	agentsByRemoteIp map[string][]*agent
}

type agent struct {
	Id      string                 `json:"id"`
	Meta    map[string]interface{} `json:"meta"`
	client  *websocketClient
	remotes map[*remote]bool
}

type remote struct {
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
	// TODO: read message from newClient.recv instead of reading from newClient.conn

	// Listen for one message with the agent information.
	newAgent := &agent{
		Meta:    map[string]interface{}{},
		client:  newClient,
		remotes: map[*remote]bool{},
	}
	err := newClient.conn.ReadJSON(newAgent)
	if err != nil {
		log.Println("Could not read agent message:", err)
		newClient.close()
		return
	}

	newAgent.Id = uuid.Must(uuid.NewV4()).String()
	log.Println("Agent client connected:", newAgent.Id)

	// Store the agent by id and remote IP.
	p.agents[newAgent.Id] = newAgent
	p.agentsByRemoteIp[newClient.remoteIp] = append(p.agentsByRemoteIp[newClient.remoteIp], newAgent)

	// Send the agent info to the agent.
	reply, err := json.Marshal(newAgent)
	if err != nil {
		log.Println("Could not marshal agent response:", err)
		newClient.close()
		return
	}
	newAgent.client.send <- message{
		messageType: websocket.TextMessage,
		data:        reply,
	}

	// Pipe received messages to all remotes.
	go func() {
		for msg := range newAgent.client.recv {
			for r := range newAgent.remotes {
				r.client.send <- msg
			}
		}

		log.Println("Closed agent:", newAgent.Id)
		p.removeAgent(newAgent)
	}()
}

func (p *proxy) onRemoteConnected(newClient *websocketClient) {
	// TODO: read message from newClient.recv instead of reading from newClient.conn

	// Listen for one message with the id of the agent to connect to.
	target := &agentId{}
	err := newClient.conn.ReadJSON(target)
	if err != nil {
		log.Println("Could not read remote message:", err)
		newClient.close()
		return
	}

	// Find the target agent.
	targetAgent := p.agents[target.Id]
	if targetAgent == nil {
		log.Println("Could not find target agent:", target.Id)
		newClient.close()
		return
	}

	log.Println("Remote client connected to:", target.Id)
	newRemote := &remote{
		client: newClient,
	}
	targetAgent.remotes[newRemote] = true

	// Pipe received messages to the agent.
	go func() {
		for msg := range newRemote.client.recv {
			targetAgent.client.send <- msg
		}

		log.Println("Closed remote:", target.Id)
		delete(targetAgent.remotes, newRemote)
	}()
}

func (p *proxy) removeAgent(a *agent) {
	a.client.close()

	// Close all remotes.
	for r := range a.remotes {
		r.client.close()
		delete(a.remotes, r)
	}

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
