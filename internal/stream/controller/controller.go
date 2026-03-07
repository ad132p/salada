package controller

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// Room represents a streaming room with a streamer and viewers
type Room struct {
	Streamer string
	Clients  map[*websocket.Conn]bool
	mu       sync.RWMutex
}

// StreamController handles video streaming requests
type StreamController struct {
	upgrader websocket.Upgrader
	rooms    map[string]*Room // map[username]*Room
	mu       sync.RWMutex
}

// NewStreamController creates a new StreamController instance
func NewStreamController() *StreamController {
	return &StreamController{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		rooms: make(map[string]*Room),
	}
}

// GetOrCreateRoom gets an existing room or creates a new one
func (sc *StreamController) GetOrCreateRoom(username string) *Room {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if room, exists := sc.rooms[username]; exists {
		return room
	}

	room := &Room{
		Streamer: username,
		Clients:  make(map[*websocket.Conn]bool),
	}
	sc.rooms[username] = room
	return room
}

// GetRoom gets an existing room by username
func (sc *StreamController) GetRoom(username string) (*Room, bool) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	room, exists := sc.rooms[username]
	return room, exists
}

// DeleteRoom removes a room
func (sc *StreamController) DeleteRoom(username string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	delete(sc.rooms, username)
}

// GetActiveRoomsData returns room data for the API
func (sc *StreamController) GetActiveRoomsData() []gin.H {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	rooms := make([]gin.H, 0, len(sc.rooms))
	for username, room := range sc.rooms {
		room.mu.RLock()
		viewerCount := len(room.Clients)
		room.mu.RUnlock()

		rooms = append(rooms, gin.H{
			"username":     username,
			"streamer":     username,
			"viewer_count": viewerCount,
			"url":          "/rooms/" + username,
		})
	}
	return rooms
}

// GetStreamPage renders the stream page for streamers (protected)
func (sc *StreamController) GetStreamPage(c *gin.Context) {
	username, _ := c.Get("username")
	c.HTML(http.StatusOK, "pages/stream.html", gin.H{
		"title":        "Video Stream",
		"is_logged_in": c.GetBool("is_logged_in"),
		"username":     username,
	})
}

// GetRoomsPage renders the rooms list page (public)
func (sc *StreamController) GetRoomsPage(c *gin.Context) {
	c.HTML(http.StatusOK, "pages/rooms.html", gin.H{
		"title":        "Rooms",
		"is_logged_in": c.GetBool("is_logged_in"),
	})
}

// GetWatchRoomPage renders the watch page for a specific room (public)
func (sc *StreamController) GetWatchRoomPage(c *gin.Context) {
	streamer := c.Param("username")
	room, exists := sc.GetRoom(streamer)
	if !exists {
		c.HTML(http.StatusNotFound, "pages/404.html", gin.H{
			"title":        "Room Not Found",
			"is_logged_in": c.GetBool("is_logged_in"),
		})
		return
	}

	room.mu.RLock()
	viewerCount := len(room.Clients)
	room.mu.RUnlock()

	c.HTML(http.StatusOK, "pages/watch.html", gin.H{
		"title":        "Watching " + streamer,
		"is_logged_in": c.GetBool("is_logged_in"),
		"streamer":     streamer,
		"viewer_count": viewerCount,
	})
}

// HandleWebSocket handles WebSocket connections for streaming
func (sc *StreamController) HandleWebSocket(c *gin.Context) {
	// Get the room parameter (which room to join)
	roomParam := c.Query("room")

	// Get username from context if authenticated
	username, isAuthenticated := c.Get("username")
	var userNameStr string
	if isAuthenticated {
		userNameStr = username.(string)
	}

	conn, err := sc.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	var room *Room
	isStreamer := false

	if roomParam != "" {
		// Viewer joining an existing room
		var exists bool
		room, exists = sc.GetRoom(roomParam)
		if !exists {
			conn.WriteJSON(gin.H{"type": "error", "message": "Room not found"})
			return
		}
	} else if isAuthenticated {
		// Streamer creating their own room
		room = sc.GetOrCreateRoom(userNameStr)
		isStreamer = true
	} else {
		// Not authenticated and no room specified
		conn.WriteJSON(gin.H{"type": "error", "message": "Authentication required"})
		return
	}

	// Register client in the room
	room.mu.Lock()
	room.Clients[conn] = true
	room.mu.Unlock()

	defer func() {
		room.mu.Lock()
		delete(room.Clients, conn)
		clientCount := len(room.Clients)
		room.mu.Unlock()

		// If streamer disconnects and no more clients, delete the room
		if isStreamer && clientCount == 0 {
			sc.DeleteRoom(userNameStr)
		}
	}()

	// Message handling loop
	for {
		var msg map[string]interface{}
		if err := conn.ReadJSON(&msg); err != nil {
			break
		}

		// Add sender info
		if userNameStr != "" {
			msg["sender"] = userNameStr
		}

		// Broadcast to all other clients in the room
		sc.broadcastToRoom(room, conn, msg)
	}
}

// broadcastToRoom sends a message to all clients in a room except the sender
func (sc *StreamController) broadcastToRoom(room *Room, sender *websocket.Conn, msg map[string]interface{}) {
	room.mu.RLock()
	defer room.mu.RUnlock()

	for client := range room.Clients {
		if client != sender {
			client.WriteJSON(msg)
		}
	}
}

// GetConnectedClients returns client counts
func (sc *StreamController) GetConnectedClients(c *gin.Context) {
	sc.mu.RLock()
	totalClients := 0
	for _, room := range sc.rooms {
		room.mu.RLock()
		totalClients += len(room.Clients)
		room.mu.RUnlock()
	}
	roomCount := len(sc.rooms)
	sc.mu.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"connected_clients": totalClients,
		"active_rooms":      roomCount,
	})
}

// GetActiveRooms returns active rooms as JSON
func (sc *StreamController) GetActiveRooms(c *gin.Context) {
	rooms := sc.GetActiveRoomsData()
	c.JSON(http.StatusOK, gin.H{
		"rooms": rooms,
	})
}
