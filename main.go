package main

import (
	"log"
	"strings"
)

func main() {
	log.Println("Starting proxy...")
	newProxy("/ws").listen(":6061", ":6062")
}

func getRemoteIp(address string) string {
	var ip string

	if address[0] == '[' {
		// IPv6: [<ip>]:<port>
		ip = address[1:strings.IndexRune(address, ']')]
	} else {
		// IPv4: <ip>:<port>
		ip = strings.Split(address, ":")[0]
	}

	// Make sure the IP address is always the same when coming from localhost.
	if ip == "127.0.0.1" || ip == "::ffff:127.0.0.1" {
		ip = "::1"
	}

	return ip
}
