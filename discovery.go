package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func discovery(w http.ResponseWriter, r *http.Request) {
	reqIp := getRemoteIp(r.RemoteAddr)
	j, err := json.Marshal(prx.agentsByPublicIp[reqIp])

	if err != nil {
		log.Println("Could not marshal agents list:", err)
		w.WriteHeader(500)
		return
	}

	// TODO: j is nil instead of empty list.
	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}
