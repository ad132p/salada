package controller

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"
)

type SFURoom struct {
	peers  map[string]*webrtc.PeerConnection
	tracks map[string]*webrtc.TrackLocalStaticRTP
	mu     sync.RWMutex
}

func NewSFURoom() *SFURoom {
	return &SFURoom{
		peers:  make(map[string]*webrtc.PeerConnection),
		tracks: make(map[string]*webrtc.TrackLocalStaticRTP),
	}
}

type Room struct {
	Streamer string
	SFU      *SFURoom
	mu       sync.RWMutex
}

type StreamController struct {
	upgrader websocket.Upgrader
	rooms    map[string]*Room
	mu       sync.RWMutex
}

func NewStreamController() *StreamController {
	return &StreamController{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
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
		SFU:      NewSFURoom(),
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

func (sc *StreamController) createPeer(room *Room, username string) (*webrtc.PeerConnection, error) {

	pc, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return nil, err
	}

	room.SFU.mu.Lock()
	room.SFU.peers[username] = pc
	room.SFU.mu.Unlock()

	room.SFU.mu.RLock()

	for _, track := range room.SFU.tracks {

		_, err := pc.AddTrack(track)
		if err != nil {
			room.SFU.mu.RUnlock()
			return nil, err
		}

	}

	room.SFU.mu.RUnlock()

	return pc, nil
}

func (sc *StreamController) handlePublisher(room *Room, pc *webrtc.PeerConnection) {

	pc.OnTrack(func(remote *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {

		localTrack, err := webrtc.NewTrackLocalStaticRTP(
			remote.Codec().RTPCodecCapability,
			remote.ID(),
			remote.StreamID(),
		)

		if err != nil {
			return
		}

		room.SFU.mu.Lock()
		room.SFU.tracks[remote.ID()] = localTrack
		room.SFU.mu.Unlock()

		buf := make([]byte, 1500)

		for {

			n, _, err := remote.Read(buf)
			if err != nil {
				break
			}

			localTrack.Write(buf[:n])
		}
	})
}

func (sc *StreamController) HandleWebSocket(c *gin.Context) {

	roomParam := c.Query("room")

	username := ""
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

	} else {

		if username == "" {
			conn.WriteJSON(gin.H{"type": "error", "message": "Authentication required"})
			return
		}

		room = sc.GetOrCreateRoom(username)
		isStreamer = true
	}

	pc, err := sc.createPeer(room, username)
	if err != nil {
		return
	}

	if isStreamer {
		sc.handlePublisher(room, pc)
	}

	for {

		var msg map[string]interface{}

		if err := conn.ReadJSON(&msg); err != nil {
			break
		}

		switch msg["type"] {

		case "offer":

			sdp := msg["sdp"].(string)

			offer := webrtc.SessionDescription{
				Type: webrtc.SDPTypeOffer,
				SDP:  sdp,
			}

			pc.SetRemoteDescription(offer)

			answer, err := pc.CreateAnswer(nil)
			if err != nil {
				continue
			}

			pc.SetLocalDescription(answer)

			conn.WriteJSON(gin.H{
				"type": "answer",
				"sdp":  answer.SDP,
			})

		case "ice":

			candidate := webrtc.ICECandidateInit{
				Candidate: msg["candidate"].(string),
			}

			pc.AddICECandidate(candidate)
		}
	}

	if isStreamer {
		sc.DeleteRoom(username)
	}
}

func (sc *StreamController) GetActiveRoomsData() []gin.H {

	sc.mu.RLock()
	defer sc.mu.RUnlock()

	rooms := make([]gin.H, 0, len(sc.rooms))

	for username := range sc.rooms {

		rooms = append(rooms, gin.H{
			"username": username,
			"streamer": username,
			"url":      "/rooms/" + username,
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
	roomCount := len(sc.rooms)
	sc.mu.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"active_rooms": roomCount,
	})
}

func (sc *StreamController) GetStreamPage(c *gin.Context) {

	username, _ := c.Get("username")

	c.HTML(http.StatusOK, "pages/stream.html", gin.H{
		"title":        "Video Stream",
		"is_logged_in": c.GetBool("is_logged_in"),
		"username":     username,
	})
}

func (sc *StreamController) GetRoomsPage(c *gin.Context) {

	c.HTML(http.StatusOK, "pages/rooms.html", gin.H{
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

	c.HTML(http.StatusOK, "pages/watch.html", gin.H{
		"title":        "Watching " + streamer,
		"is_logged_in": c.GetBool("is_logged_in"),
		"streamer":     streamer,
		"viewer_count": len(room.SFU.peers),
	})
}
