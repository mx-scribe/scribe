package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/mx-scribe/scribe/internal/domain/entities"
)

// SSEHub manages Server-Sent Events connections.
type SSEHub struct {
	clients    map[chan SSEEvent]bool
	register   chan chan SSEEvent
	unregister chan chan SSEEvent
	broadcast  chan SSEEvent
	mu         sync.RWMutex
}

// SSEEvent represents an event sent to clients.
type SSEEvent struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

// NewSSEHub creates a new SSE hub.
func NewSSEHub() *SSEHub {
	hub := &SSEHub{
		clients:    make(map[chan SSEEvent]bool),
		register:   make(chan chan SSEEvent),
		unregister: make(chan chan SSEEvent),
		broadcast:  make(chan SSEEvent, 100),
	}
	go hub.run()
	return hub
}

// run processes hub events.
func (h *SSEHub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client)
			}
			h.mu.Unlock()

		case event := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client <- event:
				default:
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastLogCreated sends a log created event to all clients.
func (h *SSEHub) BroadcastLogCreated(log *entities.Log) {
	h.broadcast <- SSEEvent{
		Type: "log_created",
		Data: logToSSEResponse(log),
	}
}

// BroadcastLogDeleted sends a log deleted event to all clients.
func (h *SSEHub) BroadcastLogDeleted(id int64) {
	h.broadcast <- SSEEvent{
		Type: "log_deleted",
		Data: map[string]int64{"id": id},
	}
}

// BroadcastStatsUpdated sends a stats updated event to all clients.
func (h *SSEHub) BroadcastStatsUpdated(stats any) {
	h.broadcast <- SSEEvent{
		Type: "stats_updated",
		Data: stats,
	}
}

// ClientCount returns the number of connected clients.
func (h *SSEHub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// SSEHandler handles GET /api/events for SSE connections.
func SSEHandler(hub *SSEHub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		client := make(chan SSEEvent, 10)
		hub.register <- client

		sendSSEEvent(w, flusher, SSEEvent{
			Type: "connected",
			Data: map[string]any{
				"message":   "Connected to SCRIBE event stream",
				"timestamp": time.Now().Format(time.RFC3339),
			},
		})

		notify := r.Context().Done()
		go func() {
			<-notify
			hub.unregister <- client
		}()

		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case event, ok := <-client:
				if !ok {
					return
				}
				sendSSEEvent(w, flusher, event)

			case <-ticker.C:
				sendSSEEvent(w, flusher, SSEEvent{
					Type: "ping",
					Data: map[string]string{"timestamp": time.Now().Format(time.RFC3339)},
				})

			case <-notify:
				return
			}
		}
	}
}

// sendSSEEvent sends a single SSE event.
func sendSSEEvent(w http.ResponseWriter, flusher http.Flusher, event SSEEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	fmt.Fprintf(w, "event: %s\n", event.Type)
	fmt.Fprintf(w, "data: %s\n\n", data)
	flusher.Flush()
}

// logToSSEResponse converts a Log to SSE response format.
func logToSSEResponse(log *entities.Log) map[string]any {
	return map[string]any{
		"id": log.ID,
		"header": map[string]any{
			"title":       log.Header.Title,
			"severity":    string(log.EffectiveSeverity()),
			"source":      log.Header.Source,
			"color":       string(log.EffectiveColor()),
			"description": log.Header.Description,
		},
		"body": log.Body,
		"metadata": map[string]any{
			"derived_severity": log.Metadata.DerivedSeverity,
			"derived_source":   log.Metadata.DerivedSource,
			"derived_category": log.Metadata.DerivedCategory,
		},
		"created_at": log.CreatedAt.Format(time.RFC3339),
	}
}
