package handlers

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
)

// WebSocketHub manages WebSocket connections and broadcasts
type WebSocketHub struct {
	// Quiz ID -> list of connections
	connections map[uuid.UUID]map[*websocket.Conn]bool
	broadcast   chan BroadcastMessage
	register    chan ConnectionRequest
	unregister  chan ConnectionRequest
	mu          sync.RWMutex
	repo        quiz.QuizRepository
}

// ConnectionRequest represents a WebSocket connection request
type ConnectionRequest struct {
	QuizID uuid.UUID
	Conn   *websocket.Conn
}

// BroadcastMessage represents a message to broadcast
type BroadcastMessage struct {
	QuizID uuid.UUID
	Data   interface{}
}

// NewWebSocketHub creates a new WebSocket hub
func NewWebSocketHub(repo quiz.QuizRepository) *WebSocketHub {
	hub := &WebSocketHub{
		connections: make(map[uuid.UUID]map[*websocket.Conn]bool),
		broadcast:   make(chan BroadcastMessage, 256),
		register:    make(chan ConnectionRequest),
		unregister:  make(chan ConnectionRequest),
		repo:        repo,
	}

	go hub.run()

	return hub
}

func (h *WebSocketHub) run() {
	for {
		select {
		case req := <-h.register:
			h.mu.Lock()
			if _, ok := h.connections[req.QuizID]; !ok {
				h.connections[req.QuizID] = make(map[*websocket.Conn]bool)
			}
			h.connections[req.QuizID][req.Conn] = true
			log.Printf("Client connected to quiz %s. Total: %d", req.QuizID, len(h.connections[req.QuizID]))
			h.mu.Unlock()

		case req := <-h.unregister:
			h.mu.Lock()
			if connections, ok := h.connections[req.QuizID]; ok {
				if _, ok := connections[req.Conn]; ok {
					delete(connections, req.Conn)
					req.Conn.Close()
					log.Printf("Client disconnected from quiz %s. Total: %d", req.QuizID, len(connections))
				}
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.mu.RLock()
			connections := h.connections[msg.QuizID]
			h.mu.RUnlock()

			jsonData, err := json.Marshal(msg.Data)
			if err != nil {
				log.Printf("Failed to marshal broadcast message: %v", err)
				continue
			}

			for conn := range connections {
				if err := conn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
					log.Printf("Failed to write message: %v", err)
					h.unregister <- ConnectionRequest{QuizID: msg.QuizID, Conn: conn}
				}
			}
		}
	}
}

// BroadcastLeaderboardUpdate sends leaderboard update to all connected clients
func (h *WebSocketHub) BroadcastLeaderboardUpdate(quizID uuid.UUID) {
	entries, err := h.repo.GetLeaderboard(nil, quizID, 10)
	if err != nil {
		log.Printf("Failed to get leaderboard: %v", err)
		return
	}

	h.broadcast <- BroadcastMessage{
		QuizID: quizID,
		Data: map[string]interface{}{
			"type": "leaderboard_update",
			"data": entries,
		},
	}
}

// HandleLeaderboardWebSocket handles WebSocket connections for leaderboard
func (h *WebSocketHub) HandleLeaderboardWebSocket(c *websocket.Conn) {
	quizIDParam := c.Params("id")
	quizID, err := uuid.Parse(quizIDParam)
	if err != nil {
		log.Printf("Invalid quiz ID: %v", err)
		c.Close()
		return
	}

	// Register connection
	h.register <- ConnectionRequest{
		QuizID: quizID,
		Conn:   c,
	}

	// Send initial leaderboard
	entries, err := h.repo.GetLeaderboard(nil, quizID, 10)
	if err == nil {
		initialData := map[string]interface{}{
			"type": "leaderboard_update",
			"data": entries,
		}
		if jsonData, err := json.Marshal(initialData); err == nil {
			c.WriteMessage(websocket.TextMessage, jsonData)
		}
	}

	// Clean up on disconnect
	defer func() {
		h.unregister <- ConnectionRequest{
			QuizID: quizID,
			Conn:   c,
		}
	}()

	// Listen for messages (mainly for ping/pong)
	for {
		messageType, _, err := c.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Echo ping/pong
		if messageType == websocket.PingMessage {
			c.WriteMessage(websocket.PongMessage, nil)
		}
	}
}
