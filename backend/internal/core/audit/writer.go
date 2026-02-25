package audit

import (
	"sync"
	"time"
)

// Event is a structured audit record.
type Event struct {
	TenantID   string            `json:"tenant_id"`
	ActorID    string            `json:"actor_id"`
	Action     string            `json:"action"`
	TargetType string            `json:"target_type"`
	TargetID   string            `json:"target_id"`
	Result     string            `json:"result"`
	Reason     string            `json:"reason,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
}

// Writer persists audit events.
type Writer interface {
	Write(event Event) error
	List(tenantID string) []Event
}

// MemoryWriter keeps audit events in memory for MVP wiring/tests.
type MemoryWriter struct {
	mu     sync.RWMutex
	events []Event
}

func NewMemoryWriter() *MemoryWriter {
	return &MemoryWriter{}
}

func (w *MemoryWriter) Write(event Event) error {
	event.CreatedAt = time.Now().UTC()
	w.mu.Lock()
	w.events = append(w.events, event)
	w.mu.Unlock()
	return nil
}

func (w *MemoryWriter) List(tenantID string) []Event {
	w.mu.RLock()
	defer w.mu.RUnlock()
	out := make([]Event, 0)
	for _, event := range w.events {
		if tenantID == "" || event.TenantID == tenantID {
			out = append(out, event)
		}
	}
	return out
}

