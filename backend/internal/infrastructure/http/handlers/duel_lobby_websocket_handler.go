package handlers

import (
	"log"

	"github.com/gofiber/contrib/v3/websocket"

	appDuel "github.com/barsukov/quiz-sprint/backend/internal/application/quick_duel"
)

// DuelLobbyWebSocketHandler handles /ws/duel/lobby connections.
type DuelLobbyWebSocketHandler struct {
	hub           *DuelLobbyHub
	onlineTracker appDuel.OnlineTracker
}

func NewDuelLobbyWebSocketHandler(hub *DuelLobbyHub, onlineTracker appDuel.OnlineTracker) *DuelLobbyWebSocketHandler {
	return &DuelLobbyWebSocketHandler{hub: hub, onlineTracker: onlineTracker}
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

	if h.onlineTracker != nil {
		_ = h.onlineTracker.SetOnline(playerID, 90)
	}

	// Flush any missed events before registering to avoid race:
	// flush → register ensures no events are lost between reconnect and live delivery.
	h.hub.FlushMissedEvents(playerID, c)
	h.hub.Register(playerID, c)

	_ = c.WriteJSON(appDuel.LobbyEvent{Type: "connected", Data: map[string]string{"playerId": playerID}})

	defer func() {
		h.hub.Unregister(playerID)
		if h.onlineTracker != nil {
			_ = h.onlineTracker.SetOffline(playerID)
		}
		log.Printf("[LobbyWS] %s disconnected", playerID)
	}()

	// Park: read loop to detect disconnect (client doesn't send messages to lobby WS).
	for {
		if _, _, err := c.ReadMessage(); err != nil {
			break
		}
	}
}
