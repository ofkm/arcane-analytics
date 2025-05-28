package main

import (
	"sync"
	"time"
)

const heartbeatLimit = 5 // Maximum heartbeats allowed per IP in a 12-hour period

type heartbeatClient struct {
	heartbeats int
}

var heartbeatClients = make(map[string]*heartbeatClient)
var heartbeatClientsMu sync.Mutex

// IsAllowedToCreateHeartbeat checks if the IP is allowed to create a heartbeat
func IsAllowedToCreateHeartbeat(ip string) bool {
	heartbeatClientsMu.Lock()
	defer heartbeatClientsMu.Unlock()

	client, exists := heartbeatClients[ip]

	if !exists {
		heartbeatClients[ip] = &heartbeatClient{heartbeats: 0}
		return true
	}

	if client.heartbeats >= heartbeatLimit {
		return false
	}

	client.heartbeats++
	return true
}

// init starts a goroutine to periodically clear the heartbeat clients map
func init() {
	go func() {
		ticker := time.NewTicker(12 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			heartbeatClientsMu.Lock()
			clear(heartbeatClients)
			heartbeatClientsMu.Unlock()
		}

	}()
}
