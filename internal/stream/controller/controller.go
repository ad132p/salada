package controller

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Room struct {
	Streamer     string
	StreamerConn *websocket.Conn
	Viewers      map[*websocket.Conn]bool
	mu           sync.RWMutex
}

type StreamController struct {
	upgrader websocket.Upgrader
	rooms    map[string]*Room
	mu       sync.RWMutex
}

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

func (sc *StreamController) GetOrCreateRoom(username string) *Room {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if room := sc.rooms[username]; room != nil {
		return room
	}

	room := &Room{
		Streamer: username,
		Viewers:  make(map[*websocket.Conn]bool),
	}

	sc.rooms[username] = room
	return room
}

func (sc *StreamController) GetRoom(username string) (*Room, bool) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	room, ok := sc.rooms[username]
	return room, ok
}

func (sc *StreamController) DeleteRoom(username string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	delete(sc.rooms, username)
}

func (sc *StreamController) closeRoom(room *Room) {
	room.mu.Lock()
	defer room.mu.Unlock()

	for viewer := range room.Viewers {
		viewer.WriteJSON(gin.H{"type": "stream-ended"})
		viewer.Close()
	}

	room.Viewers = map[*websocket.Conn]bool{}
}

func (sc *StreamController) broadcastToViewers(room *Room, sender *websocket.Conn, msg map[string]interface{}) {

	room.mu.RLock()

	viewers := make([]*websocket.Conn, 0, len(room.Viewers))
	for v := range room.Viewers {
		if v != sender {
			viewers = append(viewers, v)
		}
	}

	room.mu.RUnlock()

	var failed []*websocket.Conn

	for _, v := range viewers {
		v.SetWriteDeadline(time.Now().Add(5 * time.Second))

		if err := v.WriteJSON(msg); err != nil {
			failed = append(failed, v)
		}
	}

	if len(failed) == 0 {
		return
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	for _, f := range failed {
		f.Close()
		delete(room.Viewers, f)
	}
}

func (sc *StreamController) HandleWebSocket(c *gin.Context) {

	roomParam := c.Query("room")

	var username string
	if u, ok := c.Get("username"); ok {
		if s, ok := u.(string); ok {
			username = s
		}
	}

	conn, err := sc.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	var room *Room
	isStreamer := false

	if roomParam != "" {

		var exists bool
		room, exists = sc.GetRoom(roomParam)
		if !exists {
			conn.WriteJSON(gin.H{"type": "error", "message": "Room not found"})
			return
		}

		room.mu.Lock()
		room.Viewers[conn] = true
		streamerConn := room.StreamerConn
		room.mu.Unlock()

		if streamerConn != nil {
			streamerConn.WriteJSON(gin.H{"type": "viewer-joined"})
		}

	} else if username != "" {

		room = sc.GetOrCreateRoom(username)
		isStreamer = true

		room.mu.Lock()
		room.StreamerConn = conn
		room.mu.Unlock()

	} else {

		conn.WriteJSON(gin.H{"type": "error", "message": "Authentication required"})
		return
	}

	defer func() {

		if isStreamer {

			sc.closeRoom(room)
			sc.DeleteRoom(username)
			return

		}

		room.mu.Lock()
		delete(room.Viewers, conn)
		room.mu.Unlock()

	}()

	for {

		var msg map[string]interface{}

		if err := conn.ReadJSON(&msg); err != nil {
			break
		}

		if username != "" {
			msg["sender"] = username
		}

		if isStreamer {
			sc.broadcastToViewers(room, conn, msg)
		}
	}
}

func (sc *StreamController) GetActiveRoomsData() []gin.H {

	sc.mu.RLock()
	defer sc.mu.RUnlock()

	rooms := make([]gin.H, 0, len(sc.rooms))

	for username, room := range sc.rooms {

		room.mu.RLock()
		viewerCount := len(room.Viewers)
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

func (sc *StreamController) GetActiveRooms(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"rooms": sc.GetActiveRoomsData(),
	})
}

func (sc *StreamController) GetConnectedClients(c *gin.Context) {

	sc.mu.RLock()

	total := 0
	for _, room := range sc.rooms {

		room.mu.RLock()
		total += len(room.Viewers)
		room.mu.RUnlock()

	}

	roomCount := len(sc.rooms)

	sc.mu.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"connected_clients": total,
		"active_rooms":      roomCount,
	})
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

func (sc *StreamController) GetRoomsPage(c *gin.Context) {
	c.HTML(http.StatusOK, "pages/rooms.html",
		gin.H{
			"title":        "Rooms",
			"is_logged_in": c.GetBool("is_logged_in"),
		})
}

func (sc *StreamController) GetWatchRoomPage(c *gin.Context) {
	streamer := c.Param("username")

	room, exists := sc.GetRoom(streamer)
	if !exists {
		c.HTML(http.StatusNotFound, "pages/404.html", gin.H{
			"title":        "Stream Not Found",
			"is_logged_in": c.GetBool("is_logged_in"),
		})
		return
	}

	room.mu.RLock()
	viewerCount := len(room.Viewers)
	room.mu.RUnlock()

	c.HTML(http.StatusOK, "pages/watch.html", gin.H{
		"title":        "Watching " + streamer,
		"is_logged_in": c.GetBool("is_logged_in"),
		"streamer":     streamer,
		"viewer_count": viewerCount,
	})
}
