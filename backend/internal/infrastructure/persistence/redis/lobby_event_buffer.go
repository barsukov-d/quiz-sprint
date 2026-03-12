package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	appDuel "github.com/barsukov/quiz-sprint/backend/internal/application/quick_duel"
)

const (
	lobbyBufferMaxLen = 5
	lobbyBufferTTL    = 5 * time.Minute
	lobbyBufferPrefix = "lobby:events:"
)

// LobbyEventBuffer stores missed lobby events per player in Redis.
// Implements handlers.LobbyEventBuffer.
type LobbyEventBuffer struct {
	client *Client
}

func NewLobbyEventBuffer(client *Client) *LobbyEventBuffer {
	return &LobbyEventBuffer{client: client}
}

func (b *LobbyEventBuffer) key(playerID string) string {
	return lobbyBufferPrefix + playerID
}

// Push appends an event to the player's buffer (Redis list, capped at maxLen).
func (b *LobbyEventBuffer) Push(playerID string, event appDuel.LobbyEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("lobby buffer marshal: %w", err)
	}
	ctx := context.Background()
	key := b.key(playerID)
	pipe := b.client.rdb.Pipeline()
	pipe.RPush(ctx, key, data)
	pipe.LTrim(ctx, key, -lobbyBufferMaxLen, -1) // keep last N
	pipe.Expire(ctx, key, lobbyBufferTTL)
	_, err = pipe.Exec(ctx)
	return err
}

// Pop returns all buffered events and deletes the buffer.
func (b *LobbyEventBuffer) Pop(playerID string) ([]appDuel.LobbyEvent, error) {
	ctx := context.Background()
	key := b.key(playerID)
	data, err := b.client.rdb.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, err
	}
	_ = b.client.rdb.Del(ctx, key)

	events := make([]appDuel.LobbyEvent, 0, len(data))
	for _, raw := range data {
		var ev appDuel.LobbyEvent
		if err := json.Unmarshal([]byte(raw), &ev); err == nil {
			events = append(events, ev)
		}
	}
	return events, nil
}
