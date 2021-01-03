package main

import (
	"encoding/json"
	"github.com/TouchDeck/websocket-proxy/pkg/proxy"
	"github.com/TouchDeck/websocket-proxy/pkg/remoteIp"
	"log"
	"net/http"
)

func discovery(w http.ResponseWriter, r *http.Request) {
	reqIp := remoteIp.FromRequest(r)
	agents := prx.AgentsByRemoteIp[reqIp]
	if agents == nil {
		agents = []*proxy.Agent{}
	}

	j, err := json.Marshal(agents)
	if err != nil {
		log.Println("Could not marshal agents list:", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}
