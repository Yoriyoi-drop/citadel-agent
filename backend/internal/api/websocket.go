package api

import (
	"log"

	"github.com/gofiber/contrib/websocket"
)

// WebSocketHandler handles WebSocket connections for real-time updates
func WebSocketHandler(c *websocket.Conn) {
	// Register the client
	client := &Client{
		ID:     c.Locals("id"),
		Conn:   c,
		Rooms:  make(map[string]bool),
		Send:   make(chan []byte, 256), // Buffered channel
	}

	clientManager.Register <- client

	// Send welcome message
	client.Send <- []byte("Connected to Citadel Agent WebSocket")

	// Start listening for messages from the client
	go client.ListenForMessages()
	// Start sending messages to the client
	go client.ListenForSend()

	// Keep the connection alive
	for {
		_, _, err := c.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			clientManager.Unregister <- client
			break
		}
		// In a real implementation, we would process messages here
		// For now, we just ignore them
	}
}

// Client represents a WebSocket client
type Client struct {
	ID    interface{}
	Conn  *websocket.Conn
	Rooms map[string]bool
	Send  chan []byte
}

// ListenForMessages listens for messages from the client
func (c *Client) ListenForMessages() {
	defer func() {
		c.Conn.Close()
	}()

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message from client %v: %v", c.ID, err)
			break
		}
	}
}

// ListenForSend listens for messages to send to the client
func (c *Client) ListenForSend() {
	defer func() {
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				// The channel was closed
				c.Conn.WriteMessage(websocket.TextMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("Error writing message to client %v: %v", c.ID, err)
				return
			}
		}
	}
}

// ClientManager manages WebSocket clients
type ClientManager struct {
	Clients    map[*Client]bool
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan []byte
}

// NewClientManager creates a new client manager
func NewClientManager() *ClientManager {
	return &ClientManager{
		Clients:    make(map[*Client]bool),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan []byte),
	}
}

// Start starts the client manager
func (cm *ClientManager) Start() {
	for {
		select {
		case conn := <-cm.Register:
			cm.Clients[conn] = true
			log.Printf("Client %v connected. Total clients: %d", conn.ID, len(cm.Clients))
			
		case conn := <-cm.Unregister:
			if _, ok := cm.Clients[conn]; ok {
				delete(cm.Clients, conn)
				close(conn.Send)
				log.Printf("Client %v disconnected. Total clients: %d", conn.ID, len(cm.Clients))
			}
			
		case message := <-cm.Broadcast:
			for conn := range cm.Clients {
				select {
				case conn.Send <- message:
				default:
					delete(cm.Clients, conn)
					close(conn.Send)
				}
			}
		}
	}
}

// Initialize the client manager
var clientManager = NewClientManager()