package main

import (
	"sync"
	"time"
)

const heartbeatLimit = 20  // Maximum heartbeats allowed per IP in a 12-hour period
const newInstanceLimit = 3 // Maximum new instances allowed per IP in a 12-hour period

type heartbeatClient struct {
	heartbeats   int
	newInstances int
}

var heartbeatClients = make(map[string]*heartbeatClient)
var heartbeatClientsMu sync.Mutex

// IsAllowedToCreateHeartbeat checks if the IP is allowed to create a heartbeat
func IsAllowedToCreateHeartbeat(ip string, isExistingInstance bool) bool {
	heartbeatClientsMu.Lock()
	defer heartbeatClientsMu.Unlock()

	client, exists := heartbeatClients[ip]

	if !exists {
		heartbeatClients[ip] = &heartbeatClient{heartbeats: 0, newInstances: 0}
		client = heartbeatClients[ip]
	}

	if !isExistingInstance {
		if client.newInstances < newInstanceLimit {
			client.newInstances++
		} else {
			return false
		}

	}

	if client.heartbeats < heartbeatLimit {
		client.heartbeats++
		return true
	}
	return false
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
