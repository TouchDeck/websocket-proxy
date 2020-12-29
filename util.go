package main

import (
	"net/http"
	"strings"
)

func remoteIpFromRequest(r *http.Request) string {
	// X-Real-IP contains a single IP address.
	if realIp := r.Header.Get("X-Real-IP"); realIp != "" {
		return realIp
	}

	// X-Forwarded-For contains one or more IP addresses, comma-separated.
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		return strings.Split(forwardedFor, ",")[0]
	}

	// Last resort, and probably wrong when deployed.
	// This is used during local development though, and is an IP with port.
	return ipWithoutPort(r.RemoteAddr)
}

func ipWithoutPort(address string) string {
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
