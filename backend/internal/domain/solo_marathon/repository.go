package solo_marathon

// Repository defines the interface for marathon game persistence
type Repository interface {
	// Save persists a marathon game
	Save(game *MarathonGame) error

	// FindByID retrieves a marathon game by ID
	FindByID(id GameID) (*MarathonGame, error)

	// FindActiveByPlayer retrieves the active marathon game for a player
	// Returns nil if no active game found
	FindActiveByPlayer(playerID UserID) (*MarathonGame, error)

	// Delete removes a marathon game
	Delete(id GameID) error
}

// PersonalBestRepository defines the interface for personal best persistence
type PersonalBestRepository interface {
	// Save persists a personal best record
	Save(pb *PersonalBest) error

	// FindByPlayerAndCategory retrieves personal best for a player in specific category
	// Returns nil if no record found
	FindByPlayerAndCategory(playerID UserID, category MarathonCategory) (*PersonalBest, error)

	// FindTopByCategory retrieves top N players in a category
	FindTopByCategory(category MarathonCategory, limit int) ([]*PersonalBest, error)

	// FindAllByPlayer retrieves all personal bests for a player (across all categories)
	FindAllByPlayer(playerID UserID) ([]*PersonalBest, error)
}
