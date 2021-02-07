package proxy

import (
	"encoding/json"
	"github.com/TouchDeck/websocket-proxy/pkg/ws"
	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
	"log"
	"net/http"
	"sync"
)

// TODO: Use channels to register/remove agents to prevent race conditions
// See: https://github.com/gorilla/websocket/blob/master/examples/chat/hub.go
type Proxy struct {
	agentServer      *ws.Server
	remoteServer     *ws.Server
	agents           map[string]*Agent
	AgentsByRemoteIp map[string][]*Agent
}

type Agent struct {
	Id          string                 `json:"id"`
	Meta        map[string]interface{} `json:"meta"`
	client      *ws.Client
	remotes     map[*remote]bool
	remotesLock *sync.RWMutex
}

type remote struct {
	client *ws.Client
}

type agentId struct {
	Id string `json:"id"`
}

func NewProxy(mux *http.ServeMux, basePath string) *Proxy {
	p := &Proxy{
		agentServer:      ws.NewServer(mux, basePath+"/agent"),
		remoteServer:     ws.NewServer(mux, basePath+"/remote"),
		agents:           map[string]*Agent{},
		AgentsByRemoteIp: map[string][]*Agent{},
	}

	p.agentServer.SetOnClientConnected(p.onAgentConnected)
	p.remoteServer.SetOnClientConnected(p.onRemoteConnected)

	return p
}

func (p *Proxy) onAgentConnected(newClient *ws.Client) {
	// Listen for one message with the agent information.
	msg, ok := <-newClient.Recv
	if !ok {
		log.Println("Could not read agent message")
		return
	}

	// Unmarshal the agent information.
	newAgent := &Agent{
		Meta:        map[string]interface{}{},
		client:      newClient,
		remotes:     map[*remote]bool{},
		remotesLock: &sync.RWMutex{},
	}
	if err := json.Unmarshal(msg.Data, newAgent); err != nil {
		log.Println("Could not unmarshal agent message:", err)
		newClient.Close()
		return
	}

	newAgent.Id = uuid.NewV4().String()
	log.Println("Agent client connected:", newAgent.Id)

	// Store the agent by id and remote IP.
	p.agents[newAgent.Id] = newAgent
	p.AgentsByRemoteIp[newClient.RemoteIp] = append(p.AgentsByRemoteIp[newClient.RemoteIp], newAgent)

	// Send the agent info to the agent.
	reply, err := json.Marshal(newAgent)
	if err != nil {
		log.Println("Could not marshal agent response:", err)
		newClient.Close()
		return
	}
	newAgent.client.Send <- ws.Message{
		MessageType: websocket.TextMessage,
		Data:        reply,
	}

	// Pipe received messages to all remotes.
	go func() {
		for msg := range newAgent.client.Recv {
			newAgent.remotesLock.RLock()
			for r := range newAgent.remotes {
				r.client.Send <- msg
			}
			newAgent.remotesLock.RUnlock()
		}

		log.Println("Closed agent:", newAgent.Id)
		p.removeAgent(newAgent)
	}()
}

func (p *Proxy) onRemoteConnected(newClient *ws.Client) {
	// Listen for one message with the id of the agent to connect to.
	msg, ok := <-newClient.Recv
	if !ok {
		log.Println("Could not read remote message")
		return
	}

	// Unmarshal the remote information.
	target := &agentId{}
	if err := json.Unmarshal(msg.Data, target); err != nil {
		log.Println("Could not unmarshal remote message:", err)
		newClient.Close()
		return
	}

	// Find the target agent.
	targetAgent := p.agents[target.Id]
	if targetAgent == nil {
		log.Println("Could not find target agent:", target.Id)
		newClient.Close()
		return
	}

	log.Println("Remote client connected to:", target.Id)
	newRemote := &remote{
		client: newClient,
	}
	targetAgent.remotesLock.Lock()
	targetAgent.remotes[newRemote] = true
	targetAgent.remotesLock.Unlock()

	// Pipe received messages to the agent.
	go func() {
		for msg := range newRemote.client.Recv {
			targetAgent.client.Send <- msg
		}

		log.Println("Closed remote:", target.Id)
		targetAgent.remotesLock.Lock()
		delete(targetAgent.remotes, newRemote)
		targetAgent.remotesLock.Unlock()
	}()
}

func (p *Proxy) removeAgent(a *Agent) {
	// Close all remotes.
	a.remotesLock.Lock()
	for r := range a.remotes {
		r.client.Close()
		delete(a.remotes, r)
	}
	a.remotesLock.Unlock()

	// Delete the agent from the global agents list.
	delete(p.agents, a.Id)

	// Find and remove the agent from the local agents list.
	localAgents := p.AgentsByRemoteIp[a.client.RemoteIp]
	localI := -1
	for i, l := range localAgents {
		if l.Id == a.Id {
			localI = i
		}
	}
	if localI >= 0 {
		p.AgentsByRemoteIp[a.client.RemoteIp] = append(localAgents[:localI], localAgents[localI+1:]...)
	}

	// If the last agent from a remote IP disconnected, clear the entry.
	if len(p.AgentsByRemoteIp[a.client.RemoteIp]) == 0 {
		delete(p.AgentsByRemoteIp, a.client.RemoteIp)
	}
}
