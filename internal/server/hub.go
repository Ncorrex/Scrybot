package server

import (
	"sync"

	"github.com/gorilla/websocket"
)

// Hub maintains the set of active WebSocket clients and broadcasts messages.
type Hub struct {
	clients map[*websocket.Conn]struct{}
	mu      sync.Mutex
}

func NewHub() *Hub {
	return &Hub{clients: make(map[*websocket.Conn]struct{})}
}

func (h *Hub) register(conn *websocket.Conn) {
	h.mu.Lock()
	h.clients[conn] = struct{}{}
	h.mu.Unlock()
}

func (h *Hub) unregister(conn *websocket.Conn) {
	h.mu.Lock()
	delete(h.clients, conn)
	h.mu.Unlock()
}

// Broadcast sends msg to all connected WebSocket clients.
func (h *Hub) Broadcast(msg []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for conn := range h.clients {
		_ = conn.WriteMessage(websocket.TextMessage, msg)
	}
}
