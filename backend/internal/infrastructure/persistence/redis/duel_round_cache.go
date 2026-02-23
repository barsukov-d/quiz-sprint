package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	appDuel "github.com/barsukov/quiz-sprint/backend/internal/application/quick_duel"
)

const (
	duelRoundKeyPrefix = "duel:round:" // duel:round:{gameID}:{round}
	duelRoundTTL       = 24 * time.Hour
)

// DuelRoundCache implements appDuel.DuelRoundCache using Redis hashes.
// Key pattern: duel:round:{gameID}:{round}
// Each field in the hash is the playerID; value is a JSON-encoded playerAnswer.
type DuelRoundCache struct {
	rdb *redis.Client
}

// NewDuelRoundCache creates a new Redis-backed duel round cache.
func NewDuelRoundCache(client *Client) *DuelRoundCache {
	return &DuelRoundCache{rdb: client.Redis()}
}

func roundKey(gameID string, round int) string {
	return fmt.Sprintf("%s%s:%d", duelRoundKeyPrefix, gameID, round)
}

// AddAnswer records a player's answer for the given game round.
func (c *DuelRoundCache) AddAnswer(gameID string, round int, answer appDuel.PlayerAnswer) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data, err := json.Marshal(answer)
	if err != nil {
		return fmt.Errorf("duel round cache: marshal answer: %w", err)
	}

	key := roundKey(gameID, round)

	pipe := c.rdb.Pipeline()
	pipe.HSet(ctx, key, answer.PlayerID, data)
	pipe.Expire(ctx, key, duelRoundTTL)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("duel round cache: store answer: %w", err)
	}
	return nil
}

// GetAnswers returns all player answers recorded for the given game round.
func (c *DuelRoundCache) GetAnswers(gameID string, round int) ([]appDuel.PlayerAnswer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := roundKey(gameID, round)

	fields, err := c.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("duel round cache: get answers: %w", err)
	}

	answers := make([]appDuel.PlayerAnswer, 0, len(fields))
	for _, raw := range fields {
		var ans appDuel.PlayerAnswer
		if err := json.Unmarshal([]byte(raw), &ans); err != nil {
			return nil, fmt.Errorf("duel round cache: unmarshal answer: %w", err)
		}
		answers = append(answers, ans)
	}
	return answers, nil
}

// DeleteGame removes all round answer keys for a finished game.
// It uses SCAN to find all keys matching the game prefix.
func (c *DuelRoundCache) DeleteGame(gameID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pattern := duelRoundKeyPrefix + gameID + ":*"

	var cursor uint64
	var keys []string

	for {
		batch, next, err := c.rdb.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return fmt.Errorf("duel round cache: scan keys: %w", err)
		}
		keys = append(keys, batch...)
		cursor = next
		if cursor == 0 {
			break
		}
	}

	if len(keys) == 0 {
		return nil
	}

	if err := c.rdb.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("duel round cache: delete keys: %w", err)
	}
	return nil
}
