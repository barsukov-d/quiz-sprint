package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quick_duel"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
)

const (
	queueKey       = "duel:matchmaking:queue"    // ZSET: playerID -> MMR score
	queueInfoKey   = "duel:matchmaking:info"     // HASH: playerID -> joinedAt
	initialMMRRange = 100                        // Initial MMR search range
	maxMMRRange     = 500                        // Maximum MMR range after expansion
	rangeExpansion  = 50                         // Expand range by this much per second
)

// MatchmakingQueue implements quick_duel.MatchmakingQueue using Redis
type MatchmakingQueue struct {
	rdb *redis.Client
}

// NewMatchmakingQueue creates a new Redis-based matchmaking queue
func NewMatchmakingQueue(client *Client) *MatchmakingQueue {
	return &MatchmakingQueue{rdb: client.Redis()}
}

// AddToQueue adds a player to matchmaking queue
func (q *MatchmakingQueue) AddToQueue(playerID quick_duel.UserID, mmr int, joinedAt int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pipe := q.rdb.Pipeline()

	// Add to sorted set with MMR as score
	pipe.ZAdd(ctx, queueKey, redis.Z{
		Score:  float64(mmr),
		Member: playerID.String(),
	})

	// Store join time in hash
	pipe.HSet(ctx, queueInfoKey, playerID.String(), strconv.FormatInt(joinedAt, 10))

	_, err := pipe.Exec(ctx)
	return err
}

// RemoveFromQueue removes a player from matchmaking queue
func (q *MatchmakingQueue) RemoveFromQueue(playerID quick_duel.UserID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pipe := q.rdb.Pipeline()
	pipe.ZRem(ctx, queueKey, playerID.String())
	pipe.HDel(ctx, queueInfoKey, playerID.String())

	_, err := pipe.Exec(ctx)
	return err
}

// FindMatch finds a suitable opponent for a player
// Expands search range based on how long the player has been waiting
func (q *MatchmakingQueue) FindMatch(playerID quick_duel.UserID, mmr int, searchSeconds int) (*quick_duel.UserID, *int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Calculate MMR range based on search duration
	mmrRange := initialMMRRange + (searchSeconds * rangeExpansion)
	if mmrRange > maxMMRRange {
		mmrRange = maxMMRRange
	}

	minMMR := mmr - mmrRange
	maxMMR := mmr + mmrRange

	// Find players in MMR range using ZRANGEBYSCORE
	results, err := q.rdb.ZRangeByScoreWithScores(ctx, queueKey, &redis.ZRangeBy{
		Min: strconv.Itoa(minMMR),
		Max: strconv.Itoa(maxMMR),
	}).Result()
	if err != nil {
		return nil, nil, err
	}

	// Find best match (closest MMR, not self)
	var bestMatch *redis.Z
	var bestDiff int = mmrRange + 1

	for _, z := range results {
		opponentID := z.Member.(string)
		if opponentID == playerID.String() {
			continue // Skip self
		}

		opponentMMR := int(z.Score)
		diff := abs(opponentMMR - mmr)
		if diff < bestDiff {
			bestDiff = diff
			bestMatch = &z
		}
	}

	if bestMatch == nil {
		return nil, nil, nil // No match found
	}

	opponentIDStr := bestMatch.Member.(string)
	opponentID, err := shared.NewUserID(opponentIDStr)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid opponent ID in queue: %w", err)
	}

	opponentMMR := int(bestMatch.Score)

	return &opponentID, &opponentMMR, nil
}

// GetQueueLength returns number of players in queue
func (q *MatchmakingQueue) GetQueueLength() (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := q.rdb.ZCard(ctx, queueKey).Result()
	return int(count), err
}

// IsPlayerInQueue checks if player is already in queue
func (q *MatchmakingQueue) IsPlayerInQueue(playerID quick_duel.UserID) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if player exists in sorted set
	_, err := q.rdb.ZScore(ctx, queueKey, playerID.String()).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// GetPlayerQueueInfo returns player's queue info (joinedAt, mmr)
func (q *MatchmakingQueue) GetPlayerQueueInfo(playerID quick_duel.UserID) (int64, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get MMR from sorted set
	mmr, err := q.rdb.ZScore(ctx, queueKey, playerID.String()).Result()
	if err == redis.Nil {
		return 0, 0, fmt.Errorf("player not in queue")
	}
	if err != nil {
		return 0, 0, err
	}

	// Get join time from hash
	joinedAtStr, err := q.rdb.HGet(ctx, queueInfoKey, playerID.String()).Result()
	if err == redis.Nil {
		return 0, int(mmr), nil // No join time stored
	}
	if err != nil {
		return 0, 0, err
	}

	joinedAt, err := strconv.ParseInt(joinedAtStr, 10, 64)
	if err != nil {
		return 0, int(mmr), nil
	}

	return joinedAt, int(mmr), nil
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
