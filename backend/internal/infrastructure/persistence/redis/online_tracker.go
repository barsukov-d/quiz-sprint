package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	onlineKeyPrefix = "duel:online:"       // Key: duel:online:{playerID} -> TTL marker
	inMatchKeyPrefix = "duel:inmatch:"     // Key: duel:inmatch:{playerID} -> matchID
	defaultOnlineTTL = 60                  // seconds
)

// OnlineTracker tracks player online status using Redis
type OnlineTracker struct {
	rdb *redis.Client
}

// NewOnlineTracker creates a new Redis-based online tracker
func NewOnlineTracker(client *Client) *OnlineTracker {
	return &OnlineTracker{rdb: client.Redis()}
}

// SetOnline marks a player as online with expiry
func (t *OnlineTracker) SetOnline(playerID string, expiresInSeconds int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := onlineKeyPrefix + playerID
	ttl := time.Duration(expiresInSeconds) * time.Second
	if expiresInSeconds <= 0 {
		ttl = time.Duration(defaultOnlineTTL) * time.Second
	}

	return t.rdb.Set(ctx, key, "1", ttl).Err()
}

// IsOnline checks if a player is online
func (t *OnlineTracker) IsOnline(playerID string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := onlineKeyPrefix + playerID
	exists, err := t.rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

// GetOnlineFriends returns which friends are online
func (t *OnlineTracker) GetOnlineFriends(playerID string, friendIDs []string) ([]string, error) {
	if len(friendIDs) == 0 {
		return []string{}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Build keys to check
	keys := make([]string, len(friendIDs))
	for i, fid := range friendIDs {
		keys[i] = onlineKeyPrefix + fid
	}

	// Check existence of all keys
	pipe := t.rdb.Pipeline()
	cmds := make([]*redis.IntCmd, len(keys))
	for i, key := range keys {
		cmds[i] = pipe.Exists(ctx, key)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}

	// Collect online friends
	online := make([]string, 0)
	for i, cmd := range cmds {
		if cmd.Val() > 0 {
			online = append(online, friendIDs[i])
		}
	}

	return online, nil
}

// SetInMatch marks a player as being in a match
func (t *OnlineTracker) SetInMatch(playerID string, matchID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := inMatchKeyPrefix + playerID
	// Set with long TTL (matches shouldn't last more than 10 minutes)
	return t.rdb.Set(ctx, key, matchID, 10*time.Minute).Err()
}

// ClearInMatch removes the in-match marker
func (t *OnlineTracker) ClearInMatch(playerID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := inMatchKeyPrefix + playerID
	return t.rdb.Del(ctx, key).Err()
}

// GetMatchID returns the match ID a player is in (empty if not in match)
func (t *OnlineTracker) GetMatchID(playerID string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := inMatchKeyPrefix + playerID
	matchID, err := t.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return matchID, nil
}

// Heartbeat refreshes the online status (called periodically by client)
func (t *OnlineTracker) Heartbeat(playerID string) error {
	return t.SetOnline(playerID, defaultOnlineTTL)
}

// GetOnlineCount returns total online players (for stats)
func (t *OnlineTracker) GetOnlineCount() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use SCAN to count keys (not ideal for production, but works for moderate scale)
	var count int64
	var cursor uint64
	pattern := onlineKeyPrefix + "*"

	for {
		keys, nextCursor, err := t.rdb.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return 0, err
		}
		count += int64(len(keys))
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return count, nil
}

// SetOffline explicitly marks a player as offline
func (t *OnlineTracker) SetOffline(playerID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := onlineKeyPrefix + playerID
	return t.rdb.Del(ctx, key).Err()
}

// BulkSetOnline sets multiple players online at once
func (t *OnlineTracker) BulkSetOnline(playerIDs []string, expiresInSeconds int) error {
	if len(playerIDs) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ttl := time.Duration(expiresInSeconds) * time.Second
	if expiresInSeconds <= 0 {
		ttl = time.Duration(defaultOnlineTTL) * time.Second
	}

	pipe := t.rdb.Pipeline()
	for _, playerID := range playerIDs {
		key := onlineKeyPrefix + playerID
		pipe.Set(ctx, key, "1", ttl)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// GetPlayerStatus returns detailed status for a player
func (t *OnlineTracker) GetPlayerStatus(playerID string) (isOnline bool, matchID string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pipe := t.rdb.Pipeline()
	onlineCmd := pipe.Exists(ctx, onlineKeyPrefix+playerID)
	matchCmd := pipe.Get(ctx, inMatchKeyPrefix+playerID)

	_, err = pipe.Exec(ctx)
	// Ignore redis.Nil errors for the match key
	if err != nil && err != redis.Nil {
		return false, "", fmt.Errorf("failed to get player status: %w", err)
	}

	isOnline = onlineCmd.Val() > 0
	matchID, _ = matchCmd.Result() // Ignore error, empty string is fine

	return isOnline, matchID, nil
}
