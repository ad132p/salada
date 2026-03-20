package controller

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"
)

type Peer struct {
	pc   *webrtc.PeerConnection
	conn *websocket.Conn
	mu   sync.Mutex
}

func (p *Peer) WriteJSON(v interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.conn.WriteJSON(v)
}

type SFURoom struct {
	peers  map[string]*Peer
	tracks map[string]*webrtc.TrackLocalStaticRTP
	mu     sync.RWMutex
}

func NewSFURoom() *SFURoom {
	return &SFURoom{
		peers:  make(map[string]*Peer),
		tracks: make(map[string]*webrtc.TrackLocalStaticRTP),
	}
}

type Room struct {
	Streamer string
	SFU      *SFURoom
	mu       sync.RWMutex
}

type StreamController struct {
	upgrader   websocket.Upgrader
	rooms      map[string]*Room
	mu         sync.RWMutex
	iceServers []webrtc.ICEServer
}

func NewStreamController() *StreamController {

	sc := &StreamController{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
				if allowedOrigins == "" {
					return true // Development only
				}
				for _, allowed := range strings.Split(allowedOrigins, ",") {
					if origin == strings.TrimSpace(allowed) {
						return true
					}
				}
				return false
			},
		},
		rooms: make(map[string]*Room),
	}

	sc.iceServers = []webrtc.ICEServer{
		{URLs: []string{"stun:stun.l.google.com:19302"}},
		{URLs: []string{"stun:stun1.l.google.com:19302"}},
	}

	if iceJSON := os.Getenv("ICE_SERVERS_JSON"); iceJSON != "" {

		var servers []webrtc.ICEServer

		if err := json.Unmarshal([]byte(iceJSON), &servers); err == nil {
			sc.iceServers = servers
		}
	}

	return sc
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

func (sc *StreamController) createPeer(room *Room, username string, conn *websocket.Conn) (*Peer, error) {

	pc, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: sc.iceServers,
	})

	if err != nil {
		return nil, err
	}

	peer := &Peer{
		pc:   pc,
		conn: conn,
	}

	pc.OnICECandidate(func(c *webrtc.ICECandidate) {

		if c == nil {
			return
		}

		peer.WriteJSON(gin.H{
			"type":      "ice-candidate",
			"candidate": c.ToJSON(),
		})
	})

	pc.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {

		if s == webrtc.PeerConnectionStateClosed ||
			s == webrtc.PeerConnectionStateFailed ||
			s == webrtc.PeerConnectionStateDisconnected {

			room.SFU.mu.Lock()
			delete(room.SFU.peers, username)
			room.SFU.mu.Unlock()
		}
	})

	room.SFU.mu.Lock()
	room.SFU.peers[username] = peer
	room.SFU.mu.Unlock()

	room.SFU.mu.RLock()

	for _, track := range room.SFU.tracks {

		pc.AddTrack(track)

	}

	room.SFU.mu.RUnlock()

	return peer, nil
}

func (sc *StreamController) renegotiatePeer(peer *Peer) {
	offer, err := peer.pc.CreateOffer(nil)

	if err != nil {
		return
	}

	err = peer.pc.SetLocalDescription(offer)
	if err != nil {
		return
	}

	peer.WriteJSON(gin.H{
		"type": "offer",
		"sdp":  offer.SDP,
	})
}

func (sc *StreamController) renegotiatePeers(room *Room) {

	room.SFU.mu.RLock()
	defer room.SFU.mu.RUnlock()

	for _, peer := range room.SFU.peers {
		sc.renegotiatePeer(peer)
	}
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

		for _, peer := range room.SFU.peers {
			peer.pc.AddTrack(localTrack)
		}

		room.SFU.mu.Unlock()

		sc.renegotiatePeers(room)

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

			conn.WriteJSON(gin.H{
				"type":    "error",
				"message": "Room not found",
			})

			return
		}

	} else {

		if username == "" {

			conn.WriteJSON(gin.H{
				"type":    "error",
				"message": "Authentication required",
			})

			return
		}

		room = sc.GetOrCreateRoom(username)

		isStreamer = true
	}

	peer, err := sc.createPeer(room, username, conn)

	if err != nil {
		return
	}

	if isStreamer {
		sc.handlePublisher(room, peer.pc)
	} else {
		// If viewer joins and tracks already exist, trigger renegotiation for this peer
		room.SFU.mu.RLock()
		hasTracks := len(room.SFU.tracks) > 0
		room.SFU.mu.RUnlock()

		if hasTracks {
			sc.renegotiatePeer(peer)
		}
	}

	for {

		var msg map[string]interface{}

		if err := conn.ReadJSON(&msg); err != nil {
			break
		}

		msgType, ok := msg["type"].(string)

		if !ok {
			continue
		}

		switch msgType {

		case "offer":

			sdp, ok := msg["sdp"].(string)

			if !ok {
				continue
			}

			offer := webrtc.SessionDescription{
				Type: webrtc.SDPTypeOffer,
				SDP:  sdp,
			}

			peer.pc.SetRemoteDescription(offer)

			answer, err := peer.pc.CreateAnswer(nil)

			if err != nil {
				continue
			}

			peer.pc.SetLocalDescription(answer)

			peer.WriteJSON(gin.H{
				"type": "answer",
				"sdp":  answer.SDP,
			})

		case "answer":

			sdp, ok := msg["sdp"].(string)

			if !ok {
				continue
			}

			answer := webrtc.SessionDescription{
				Type: webrtc.SDPTypeAnswer,
				SDP:  sdp,
			}

			peer.pc.SetRemoteDescription(answer)

		case "ice-candidate":

			candidateMsg, ok := msg["candidate"]

			if !ok {
				continue
			}

			candidateBytes, err := json.Marshal(candidateMsg)
			if err != nil {
				continue
			}

			var candidate webrtc.ICECandidateInit
			if err := json.Unmarshal(candidateBytes, &candidate); err != nil {
				continue
			}

			peer.pc.AddICECandidate(candidate)
		}
	}

	peer.pc.Close()

	room.SFU.mu.Lock()
	delete(room.SFU.peers, username)
	room.SFU.mu.Unlock()

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
	defer sc.mu.RUnlock()

	total := 0

	for _, room := range sc.rooms {

		room.SFU.mu.RLock()
		total += len(room.SFU.peers)
		room.SFU.mu.RUnlock()

	}

	c.JSON(http.StatusOK, gin.H{
		"connected_clients": total,
		"active_rooms":      len(sc.rooms),
	})
}

func (sc *StreamController) GetStreamPage(c *gin.Context) {

	username, _ := c.Get("username")

	iceServersJSON, _ := json.Marshal(sc.iceServers)

	c.HTML(http.StatusOK, "pages/stream.html", gin.H{
		"title":        "Video Stream",
		"is_logged_in": c.GetBool("is_logged_in"),
		"username":     username,
		"ice_servers":  string(iceServersJSON),
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

	room.SFU.mu.RLock()
	viewerCount := len(room.SFU.peers)
	room.SFU.mu.RUnlock()

	iceServersJSON, _ := json.Marshal(sc.iceServers)

	c.HTML(http.StatusOK, "pages/watch.html", gin.H{
		"title":        "Watching " + streamer,
		"is_logged_in": c.GetBool("is_logged_in"),
		"streamer":     streamer,
		"viewer_count": viewerCount,
		"ice_servers":  string(iceServersJSON),
	})
}
