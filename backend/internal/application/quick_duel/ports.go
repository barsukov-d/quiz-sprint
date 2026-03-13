package quick_duel

import (
	"context"
	"database/sql"
)

// TxManager provides database transaction support
type TxManager interface {
	RunInTx(ctx context.Context, fn func(tx *sql.Tx) error) error
}

// LobbyEvent is a real-time lobby notification sent to a player via WebSocket.
type LobbyEvent struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// LobbyHub manages WebSocket connections for the duel lobby (pre-game).
// Implementations must be safe for concurrent use.
type LobbyHub interface {
	// IsConnected returns true if the player has an active lobby WS connection.
	IsConnected(playerID string) bool
	// Notify sends an event to a single player. No-ops silently if not connected.
	Notify(playerID string, event LobbyEvent)
	// NotifyBoth sends the same event to two players.
	NotifyBoth(player1ID, player2ID string, event LobbyEvent)
}

// NoOpLobbyHub discards all notifications. Used when WS is not available.
type NoOpLobbyHub struct{}

func (n *NoOpLobbyHub) IsConnected(playerID string) bool           { return false }
func (n *NoOpLobbyHub) Notify(playerID string, event LobbyEvent)   {}
func (n *NoOpLobbyHub) NotifyBoth(p1, p2 string, event LobbyEvent) {}

// DuelRoundCache stores per-round player answers for an ongoing duel game.
// Implementations must be safe for concurrent use and survive process restarts.
type DuelRoundCache interface {
	// AddAnswer records a player's answer for a specific round.
	AddAnswer(gameID string, round int, answer PlayerAnswer) error

	// GetAnswers returns all answers stored for a specific round.
	GetAnswers(gameID string, round int) ([]PlayerAnswer, error)

	// DeleteGame removes all cached answers for a finished game.
	DeleteGame(gameID string) error
}
