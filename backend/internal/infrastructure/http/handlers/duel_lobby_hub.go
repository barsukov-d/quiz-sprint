package handlers

import (
	"sync"

	appDuel "github.com/barsukov/quiz-sprint/backend/internal/application/quick_duel"
)

// LobbyConn is the write-side interface of a WebSocket connection.
// Abstracted for testability.
type LobbyConn interface {
	WriteJSON(v interface{}) error
}

// LobbyEventBuffer stores missed events for reconnecting players.
// Can be nil (no buffering).
type LobbyEventBuffer interface {
	Push(playerID string, event appDuel.LobbyEvent) error
	Pop(playerID string) ([]appDuel.LobbyEvent, error)
}

// DuelLobbyHub manages lobby WebSocket connections (pre-game).
// Implements appDuel.LobbyHub.
type DuelLobbyHub struct {
	mu          sync.RWMutex
	connections map[string]LobbyConn
	buffer      LobbyEventBuffer // optional; nil disables event buffering
}

// NewDuelLobbyHub creates a new hub. buffer may be nil.
func NewDuelLobbyHub(buffer LobbyEventBuffer) *DuelLobbyHub {
	return &DuelLobbyHub{
		connections: make(map[string]LobbyConn),
		buffer:      buffer,
	}
}

// Register adds a player's connection. Replaces any existing connection.
func (h *DuelLobbyHub) Register(playerID string, conn LobbyConn) {
	h.mu.Lock()
	h.connections[playerID] = conn
	h.mu.Unlock()
}

// Unregister removes a player's connection.
func (h *DuelLobbyHub) Unregister(playerID string) {
	h.mu.Lock()
	delete(h.connections, playerID)
	h.mu.Unlock()
}

// IsConnected implements appDuel.LobbyHub.
func (h *DuelLobbyHub) IsConnected(playerID string) bool {
	h.mu.RLock()
	_, ok := h.connections[playerID]
	h.mu.RUnlock()
	return ok
}

// Notify implements appDuel.LobbyHub.
func (h *DuelLobbyHub) Notify(playerID string, event appDuel.LobbyEvent) {
	if h.buffer != nil {
		_ = h.buffer.Push(playerID, event)
	}
	h.mu.RLock()
	conn, ok := h.connections[playerID]
	h.mu.RUnlock()
	if ok {
		_ = conn.WriteJSON(event)
	}
}

// NotifyBoth implements appDuel.LobbyHub.
func (h *DuelLobbyHub) NotifyBoth(player1ID, player2ID string, event appDuel.LobbyEvent) {
	h.Notify(player1ID, event)
	h.Notify(player2ID, event)
}

// FlushMissedEvents sends buffered events to a newly connected player and clears them.
func (h *DuelLobbyHub) FlushMissedEvents(playerID string, conn LobbyConn) {
	if h.buffer == nil {
		return
	}
	events, err := h.buffer.Pop(playerID)
	if err != nil || len(events) == 0 {
		return
	}
	for _, ev := range events {
		_ = conn.WriteJSON(ev)
	}
}
