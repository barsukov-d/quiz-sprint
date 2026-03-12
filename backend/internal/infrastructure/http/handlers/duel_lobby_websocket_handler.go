package handlers

import (
	"log"

	"github.com/gofiber/contrib/v3/websocket"

	appDuel "github.com/barsukov/quiz-sprint/backend/internal/application/quick_duel"
)

// DuelLobbyWebSocketHandler handles /ws/duel/lobby connections.
type DuelLobbyWebSocketHandler struct {
	hub *DuelLobbyHub
}

func NewDuelLobbyWebSocketHandler(hub *DuelLobbyHub) *DuelLobbyWebSocketHandler {
	return &DuelLobbyWebSocketHandler{hub: hub}
}

// HandleLobbyWebSocket registers the player, flushes missed events,
// then parks until disconnect. Client-to-server messages are ignored
// (lobby is server-push only).
func (h *DuelLobbyWebSocketHandler) HandleLobbyWebSocket(c *websocket.Conn) {
	playerID := c.Query("playerId")
	if playerID == "" {
		_ = c.WriteJSON(map[string]string{"type": "error", "error": "playerId required"})
		c.Close()
		return
	}

	log.Printf("[LobbyWS] %s connected", playerID)

	// Flush any missed events before registering to avoid race:
	// flush → register ensures no events are lost between reconnect and live delivery.
	h.hub.FlushMissedEvents(playerID, c)
	h.hub.Register(playerID, c)

	_ = c.WriteJSON(appDuel.LobbyEvent{Type: "connected", Data: map[string]string{"playerId": playerID}})

	defer func() {
		h.hub.Unregister(playerID)
		log.Printf("[LobbyWS] %s disconnected", playerID)
	}()

	// Park: read loop to detect disconnect (client doesn't send messages to lobby WS).
	for {
		if _, _, err := c.ReadMessage(); err != nil {
			break
		}
	}
}
