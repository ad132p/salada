package controller

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// StreamController handles video streaming requests
type StreamController struct {
	upgrader websocket.Upgrader
	clients  map[*websocket.Conn]bool
	mu       sync.RWMutex
}

// NewStreamController creates a new StreamController instance
func NewStreamController() *StreamController {
	return &StreamController{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for development
			},
		},
		clients: make(map[*websocket.Conn]bool),
	}
}

// GetStreamPage renders the stream page with the React UI
func (sc *StreamController) GetStreamPage(c *gin.Context) {
	c.HTML(http.StatusOK, "pages/stream.html", gin.H{
		"title":        "Video Stream",
		"is_logged_in": c.GetBool("is_logged_in"),
	})
}

// HandleWebSocket upgrades HTTP connection to WebSocket for video streaming
func (sc *StreamController) HandleWebSocket(c *gin.Context) {
	conn, err := sc.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to upgrade connection"})
		return
	}
	defer conn.Close()

	// Register client
	sc.mu.Lock()
	sc.clients[conn] = true
	sc.mu.Unlock()

	defer func() {
		sc.mu.Lock()
		delete(sc.clients, conn)
		sc.mu.Unlock()
	}()

	// Handle WebSocket messages (client -> server signaling)
	for {
		var msg map[string]interface{}
		if err := conn.ReadJSON(&msg); err != nil {
			break
		}

		// Handle signaling messages (offer, answer, ICE candidates)
		sc.handleSignalingMessage(conn, msg)
	}
}

// handleSignalingMessage processes WebRTC signaling messages
func (sc *StreamController) handleSignalingMessage(sender *websocket.Conn, msg map[string]interface{}) {
	// Broadcast the message to all other connected clients
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	msgType, ok := msg["type"].(string)
	if !ok {
		return
	}

	// Broadcast signaling messages to all connected clients
	for client := range sc.clients {
		if client != sender {
			if err := client.WriteJSON(msg); err != nil {
				// Client disconnected, will be cleaned up on next iteration
				continue
			}
		}
	}

	// Log for debugging
	_ = msgType
}

// GetConnectedClients returns the number of connected WebSocket clients
func (sc *StreamController) GetConnectedClients(c *gin.Context) {
	sc.mu.RLock()
	count := len(sc.clients)
	sc.mu.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"connected_clients": count,
	})
}
