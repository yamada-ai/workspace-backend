package ws

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"

	"github.com/yamada-ai/workspace-backend/usecase/command"
)

// Hub maintains the set of active clients and broadcasts messages to them
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Inbound messages from the clients
	broadcast chan Event

	// Register requests from the clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	mu sync.RWMutex
}

// NewHub creates a new Hub
func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan Event, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("Client connected. Total clients: %d", len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			log.Printf("Client disconnected. Total clients: %d", len(h.clients))

		case event := <-h.broadcast:
			// Marshal event to JSON
			message, err := json.Marshal(event)
			if err != nil {
				log.Printf("Error marshaling event: %v", err)
				continue
			}

			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// Client's send buffer is full, disconnect them
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast sends an event to all connected clients
func (h *Hub) Broadcast(event Event) {
	h.broadcast <- event
}

// BroadcastSessionStart implements command.EventBroadcaster
func (h *Hub) BroadcastSessionStart(event command.SessionStartBroadcast) {
	wsEvent := SessionStartEvent{
		Type:       EventTypeSessionStart,
		ID:         event.SessionID,
		UserID:     event.UserID,
		UserName:   event.UserName,
		WorkName:   event.WorkName,
		Tier:       event.Tier,
		StartTime:  event.StartTime,
		PlannedEnd: event.PlannedEnd,
	}
	h.Broadcast(wsEvent)
}

// BroadcastSessionEnd implements command.EventBroadcaster
func (h *Hub) BroadcastSessionEnd(event command.SessionEndBroadcast) {
	wsEvent := SessionEndEvent{
		Type:      EventTypeSessionEnd,
		ID:        event.SessionID,
		UserID:    event.UserID,
		ActualEnd: event.ActualEnd,
	}
	h.Broadcast(wsEvent)
}

// BroadcastWorkNameChange implements command.WorkNameChangeBroadcaster
func (h *Hub) BroadcastWorkNameChange(event command.WorkNameChangeBroadcast) {
	wsEvent := WorkNameChangeEvent{
		Type:     EventTypeWorkNameChange,
		ID:       event.SessionID,
		UserID:   event.UserID,
		WorkName: event.WorkName,
	}
	h.Broadcast(wsEvent)
}

// BroadcastSessionExtend implements command.EventBroadcaster
func (h *Hub) BroadcastSessionExtend(event command.SessionExtendBroadcast) {
	wsEvent := SessionExtendEvent{
		Type:          EventTypeSessionExtend,
		ID:            event.SessionID,
		UserID:        event.UserID,
		NewPlannedEnd: event.NewPlannedEnd,
	}
	h.Broadcast(wsEvent)
}

// Client represents a WebSocket client
type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

// NewClient creates a new Client
func NewClient(hub *Hub, conn *websocket.Conn) *Client {
	return &Client{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, 256),
	}
}

// ReadPump pumps messages from the websocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
		// Currently we don't process messages from clients
		// In the future, we might handle client-to-server messages here
	}
}

// WritePump pumps messages from the hub to the websocket connection
func (c *Client) WritePump() {
	defer func() {
		c.conn.Close()
	}()

	for {
		message, ok := <-c.send
		if !ok {
			// The hub closed the channel
			_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		err := c.conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			return
		}
	}
}
