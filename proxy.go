package main

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
		// Listen for one message with local ip, pipe to agent.
	}

	return &proxy{
		agentServer:  agentServer,
		remoteServer: remoteServer,
	}
}
