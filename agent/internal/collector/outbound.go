package collector

import (
	"runtime"
	"sync"
	"time"
)

const TypeOutbound = "metrics:outbound"

// OutboundConnection represents an outbound TCP connection.
type OutboundConnection struct {
	RemoteAddr string `json:"remoteAddr"`
	RemotePort int    `json:"remotePort"`
	LocalPort  int    `json:"localPort"`
	FirstSeen  int64  `json:"firstSeen"`
	New        bool   `json:"new"` // true if this destination was never seen before
}

// OutboundPayload is the payload for outbound traffic messages.
type OutboundPayload struct {
	Connections    []OutboundConnection `json:"connections"`
	NewDestinations int                 `json:"newDestinations"` // count of never-before-seen remote addrs
	Total          int                  `json:"total"`
	Timestamp      int64                `json:"timestamp"`
}

// OutboundCollector tracks outbound TCP connections and flags new destinations.
func OutboundCollector(interval time.Duration) *Collector {
	tracker := &outboundTracker{
		seen: make(map[string]int64),
	}

	return New(interval, func() (string, any, error) {
		if runtime.GOOS != "linux" {
			return TypeOutbound, OutboundPayload{Timestamp: time.Now().UnixMilli()}, nil
		}

		// Read current TCP connections
		var allConns []NetConnection
		for _, path := range []string{"/proc/net/tcp", "/proc/net/tcp6"} {
			allConns = append(allConns, readProcNet(path)...)
		}

		// Filter to outbound established connections (remote addr != 0.0.0.0 and != 127.*)
		var outbound []NetConnection
		for _, c := range allConns {
			if c.State != "ESTABLISHED" {
				continue
			}
			if c.RemoteAddr == "0.0.0.0" || c.RemoteAddr == "127.0.0.1" {
				continue
			}
			outbound = append(outbound, c)
		}

		result := tracker.process(outbound)
		return TypeOutbound, result, nil
	})
}

type outboundTracker struct {
	mu   sync.Mutex
	seen map[string]int64 // "addr:port" → first seen timestamp
}

func (t *outboundTracker) process(conns []NetConnection) OutboundPayload {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now().UnixMilli()
	var result []OutboundConnection
	newCount := 0

	for _, c := range conns {
		key := c.RemoteAddr
		firstSeen, exists := t.seen[key]
		if !exists {
			firstSeen = now
			t.seen[key] = now
			newCount++
		}

		result = append(result, OutboundConnection{
			RemoteAddr: c.RemoteAddr,
			RemotePort: c.RemotePort,
			LocalPort:  c.LocalPort,
			FirstSeen:  firstSeen,
			New:        !exists,
		})
	}

	return OutboundPayload{
		Connections:     result,
		NewDestinations: newCount,
		Total:           len(result),
		Timestamp:       now,
	}
}
