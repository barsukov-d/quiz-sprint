package quick_duel

// Repository defines the interface for duel game persistence
type Repository interface {
	// Save persists a duel game
	Save(game *DuelGame) error

	// FindByID retrieves a duel game by ID
	FindByID(id GameID) (*DuelGame, error)

	// FindActiveByPlayer retrieves the active duel game for a player
	// Returns nil if no active game found
	FindActiveByPlayer(playerID UserID) (*DuelGame, error)

	// Delete removes a duel game
	Delete(id GameID) error
}

// MatchmakingQueue defines the interface for matchmaking queue operations
// (Usually implemented with Redis sorted sets)
type MatchmakingQueue interface {
	// AddToQueue adds a player to matchmaking queue
	// Priority is typically based on ELO rating
	AddToQueue(playerID UserID, elo EloRating, joinedAt int64) error

	// RemoveFromQueue removes a player from matchmaking queue
	RemoveFromQueue(playerID UserID) error

	// FindMatch finds a suitable opponent for a player
	// Returns opponent's UserID and ELO, or nil if no match found
	FindMatch(playerID UserID, elo EloRating, searchSeconds int) (*UserID, *EloRating, error)

	// GetQueueLength returns number of players in queue
	GetQueueLength() (int, error)
}
