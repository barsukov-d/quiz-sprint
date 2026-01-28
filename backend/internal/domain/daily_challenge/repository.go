package daily_challenge

// DailyQuizRepository defines the interface for daily quiz persistence
type DailyQuizRepository interface {
	// Save persists a daily quiz
	Save(dailyQuiz *DailyQuiz) error

	// FindByID retrieves a daily quiz by ID
	FindByID(id DailyQuizID) (*DailyQuiz, error)

	// FindByDate retrieves the daily quiz for a specific date
	// Returns nil if no quiz found for that date
	FindByDate(date Date) (*DailyQuiz, error)

	// Delete removes a daily quiz
	Delete(id DailyQuizID) error
}

// DailyGameRepository defines the interface for daily game persistence
type DailyGameRepository interface {
	// Save persists a daily game
	Save(game *DailyGame) error

	// FindByID retrieves a daily game by ID
	FindByID(id GameID) (*DailyGame, error)

	// FindByPlayerAndDate retrieves a player's best game for a specific date
	// Returns best attempt if multiple exist, nil if player hasn't played
	FindByPlayerAndDate(playerID UserID, date Date) (*DailyGame, error)

	// FindAllAttemptsByPlayerAndDate retrieves all player's attempts for a date
	FindAllAttemptsByPlayerAndDate(playerID UserID, date Date) ([]*DailyGame, error)

	// CountAttemptsByPlayerAndDate returns number of attempts player made for date
	CountAttemptsByPlayerAndDate(playerID UserID, date Date) (int, error)

	// FindTopByDate retrieves top N players for a specific date (leaderboard)
	// Sorted by score descending, best attempt per player
	FindTopByDate(date Date, limit int) ([]*DailyGame, error)

	// GetPlayerRankByDate calculates player's rank for a specific date
	// Returns 0 if player hasn't played
	GetPlayerRankByDate(playerID UserID, date Date) (int, error)

	// GetTotalPlayersByDate returns total number of players who played on date
	GetTotalPlayersByDate(date Date) (int, error)

	// Delete removes a daily game
	Delete(id GameID) error
}
