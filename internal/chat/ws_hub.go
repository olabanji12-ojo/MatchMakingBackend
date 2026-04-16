package chat

import (
	"encoding/json"
	"log"
	"sync"
)

type Hub struct {
	clients    map[string]*Client // userID -> Client
	broadcast  chan WSMessage
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan WSMessage),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[string]*Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.userID] = client
			h.mu.Unlock()
			log.Printf("User %s connected to WebSocket", client.userID)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.userID]; ok {
				delete(h.clients, client.userID)
				close(client.send)
				log.Printf("User %s disconnected from WebSocket", client.userID)
			}
			h.mu.Unlock()

		case _ = <-h.broadcast:
			h.mu.RLock()
			// For a direct message, we find the receiver
			// In a more complex hub, we might check participation in specific chat rooms
			// But for now, we just check if the other participant is online
			// The sender is handled by the initial POST response or readPump loop
			h.mu.RUnlock()
			// Broadcast logic is usually handled by SendMessage service or specific event triggers
		}
	}
}

func (h *Hub) BroadcastToUser(userID string, message WSMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if client, ok := h.clients[userID]; ok {
		msgBytes, _ := json.Marshal(message)
		select {
		case client.send <- msgBytes:
		default:
			log.Printf("Failed to send message to user %s, buffer full", userID)
		}
	}
}
