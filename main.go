package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

// Client represents a connected WebSocket client
type Client struct {
	conn      *websocket.Conn
	name      string
	projectID string
	userID    string
	send      chan []byte
	hub       *Hub
}

// Hub maintains active clients and broadcasts messages
type Hub struct {
	// Registered clients mapped by projectID -> client
	clients map[string]map[*Client]bool

	// Inbound messages from clients
	broadcast chan BroadcastMessage

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Project state manager
	projectManager *ProjectManager

	mu sync.RWMutex
}

// BroadcastMessage contains message and target project
type BroadcastMessage struct {
	projectID string
	message   []byte
	sender    *Client // Optional: to exclude sender from broadcast
}

// NewHub creates a new Hub
func NewHub() *Hub {
	return &Hub{
		clients:        make(map[string]map[*Client]bool),
		broadcast:      make(chan BroadcastMessage, 256),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		projectManager: NewProjectManager(),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.projectID] == nil {
				h.clients[client.projectID] = make(map[*Client]bool)
			}
			h.clients[client.projectID][client] = true
			h.mu.Unlock()
			log.Printf("[Hub] Client registered: %s (user: %s, project: %s)", client.name, client.userID, client.projectID)

			// Send current project state to the new client
			h.sendProjectSync(client)

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.projectID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.send)
					if len(clients) == 0 {
						delete(h.clients, client.projectID)
					}
					log.Printf("[Hub] Client unregistered: %s (user: %s, project: %s)", client.name, client.userID, client.projectID)
				}
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			if clients, ok := h.clients[message.projectID]; ok {
				for client := range clients {
					// Optionally skip sender
					if message.sender != nil && client == message.sender {
						continue
					}
					select {
					case client.send <- message.message:
					default:
						// Client's send channel is full, close and unregister
						close(client.send)
						delete(clients, client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// sendProjectSync sends the current project state to a client
func (h *Hub) sendProjectSync(client *Client) {
	project := h.projectManager.GetProject(client.projectID)
	if project == nil {
		// No project exists yet - first client will send ProjectSync to define initial state
		log.Printf("[Hub] No project found for ID: %s - waiting for client to send initial ProjectSync", client.projectID)
		return
	}

	packet := CreateProjectSync(*project)
	jsonData, err := SerializePacket(packet)
	if err != nil {
		log.Printf("[Hub] Failed to serialize project sync: %v", err)
		return
	}

	select {
	case client.send <- []byte(jsonData):
		log.Printf("[Hub] Sent project sync to client: %s (cells: %d)", client.name, len(project.Cells))
	default:
		log.Printf("[Hub] Failed to send project sync to client: %s", client.name)
	}
}

// readPump reads messages from the WebSocket connection
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[Client] Read error: %v", err)
			}
			break
		}

		log.Printf("[Client] Received message from %s: %s", c.name, string(message))

		// Handle the incoming packet
		response, broadcast := HandlePacket(c.hub.projectManager, c.projectID, message)

		// If broadcast is true, send to all OTHER clients in the project (excluding sender)
		if broadcast && response != nil {
			c.hub.broadcast <- BroadcastMessage{
				projectID: c.projectID,
				message:   response,
				sender:    c, // Exclude sender from broadcast
			}
		}
	}
}

// writePump writes messages to the WebSocket connection
func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()

	for message := range c.send {
		err := c.conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Printf("[Client] Write error: %v", err)
			break
		}
	}
}

// handleWebSocket handles WebSocket connections
func handleWebSocket(hub *Hub, w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	name := r.URL.Query().Get("name")
	projectID := r.URL.Query().Get("projectId")
	userID := r.URL.Query().Get("userId")

	if name == "" || projectID == "" || userID == "" {
		http.Error(w, "Missing required query parameters: name, projectId, userId", http.StatusBadRequest)
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WebSocket] Upgrade error: %v", err)
		return
	}

	// Create new client
	client := &Client{
		conn:      conn,
		name:      name,
		projectID: projectID,
		userID:    userID,
		send:      make(chan []byte, 256),
		hub:       hub,
	}

	// Register client
	hub.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()
}

func main() {
	// Create and start the hub
	hub := NewHub()
	go hub.Run()

	// Setup HTTP routes
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocket(hub, w, r)
	})

	// Start server
	addr := ":80"
	log.Printf("[Server] Starting WebSocket server on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("[Server] ListenAndServe error: ", err)
	}
}
