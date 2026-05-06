package chat

import "log"

// Hub maintains the set of active WebSocket clients and delivers
// messages to specific users. It supports multiple connections per user
// (multiple browser tabs / windows).
//
// All map operations happen exclusively inside the Run goroutine,
// so no mutex is needed — thread safety is achieved via channels.
type Hub struct {
	// clients maps user IDs to their active connections.
	clients map[string]map[*Client]bool

	// register adds a client to the hub.
	register chan *Client

	// unregister removes a client from the hub.
	unregister chan *Client

	// send delivers a message to a specific user's connections.
	send chan *targetedMessage
}

// targetedMessage carries a payload addressed to a specific user.
type targetedMessage struct {
	userID string
	data   []byte
}

// NewHub creates a new Hub instance.
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		send:       make(chan *targetedMessage, 256),
	}
}

// Run starts the hub's main event loop. It must be launched as a goroutine
// and runs for the lifetime of the application.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			if h.clients[client.UserID] == nil {
				h.clients[client.UserID] = make(map[*Client]bool)
			}
			h.clients[client.UserID][client] = true
			log.Printf("[WS Hub] User %s connected (connections: %d)",
				client.UserID, len(h.clients[client.UserID]))

		case client := <-h.unregister:
			if conns, ok := h.clients[client.UserID]; ok {
				if _, exists := conns[client]; exists {
					delete(conns, client)
					close(client.send)
					if len(conns) == 0 {
						delete(h.clients, client.UserID)
					}
				}
			}
			log.Printf("[WS Hub] User %s disconnected", client.UserID)

		case msg := <-h.send:
			conns, ok := h.clients[msg.userID]
			if !ok {
				continue
			}
			for client := range conns {
				select {
				case client.send <- msg.data:
				default:
					// Client's send buffer is full — they are too slow.
					// Disconnect them to prevent memory buildup.
					close(client.send)
					delete(conns, client)
					if len(conns) == 0 {
						delete(h.clients, msg.userID)
					}
				}
			}
		}
	}
}

// SendToUser delivers data to every active connection for the given user.
// It is safe to call from any goroutine.
func (h *Hub) SendToUser(userID string, data []byte) {
	h.send <- &targetedMessage{userID: userID, data: data}
}
