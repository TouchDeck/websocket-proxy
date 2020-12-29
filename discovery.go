package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func discovery(w http.ResponseWriter, r *http.Request) {
	reqIp := remoteIpFromRequest(r)
	agents := prx.agentsByRemoteIp[reqIp]
	if agents == nil {
		agents = []*agent{}
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
