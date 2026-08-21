package observability

import (
	"sync"
	"time"
)

type Counter struct {
	mu    sync.Mutex
	value uint64
}

func (c *Counter) Inc()          { c.mu.Lock(); c.value++; c.mu.Unlock() }
func (c *Counter) Value() uint64 { c.mu.Lock(); defer c.mu.Unlock(); return c.value }

type Histogram struct {
	mu      sync.Mutex
	samples []time.Duration
}

func (h *Histogram) Observe(d time.Duration) {
	h.mu.Lock()
	h.samples = append(h.samples, d)
	h.mu.Unlock()
}
func (h *Histogram) Count() int { h.mu.Lock(); defer h.mu.Unlock(); return len(h.samples) }
