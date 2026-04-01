package collector

import (
	"context"
	"log"
	"time"
)

// BroadcastFunc sends a message to all connected desktops.
type BroadcastFunc = func(msgType string, payload any) error

// Collector gathers data and broadcasts it on an interval.
type Collector struct {
	interval time.Duration
	gather   func() (msgType string, payload any, err error)
}

// New creates a collector with a gather function and interval.
func New(interval time.Duration, gather func() (string, any, error)) *Collector {
	return &Collector{
		interval: interval,
		gather:   gather,
	}
}

// Run starts the collector. It sends immediately, then on each interval.
// Blocks until ctx is canceled.
func (c *Collector) Run(ctx context.Context, broadcast BroadcastFunc) {
	c.tick(broadcast)

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.tick(broadcast)
		}
	}
}

func (c *Collector) tick(broadcast BroadcastFunc) {
	msgType, payload, err := c.gather()
	if err != nil {
		log.Printf("collect %s: %v", msgType, err)
		return
	}
	if err := broadcast(msgType, payload); err != nil {
		log.Printf("broadcast %s: %v", msgType, err)
	}
}

// Manager runs multiple collectors. Starts when the first desktop connects,
// stops when the last desktop disconnects.
type Manager struct {
	collectors []*Collector
}

// NewManager creates a collector manager.
func NewManager(collectors ...*Collector) *Manager {
	return &Manager{collectors: collectors}
}

// HandleActive is meant to be passed to transport.Server.OnActive.
// It starts all collectors and blocks until ctx is canceled (all desktops disconnected).
func (m *Manager) HandleActive(ctx context.Context, broadcast BroadcastFunc) {
	log.Printf("starting %d collectors", len(m.collectors))

	for _, c := range m.collectors {
		go c.Run(ctx, broadcast)
	}

	<-ctx.Done()
	log.Println("all desktops disconnected — stopping collectors")
}
