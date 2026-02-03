package marathon

// ========================================
// Common DTOs
// ========================================

// MarathonGameDTO represents a marathon game state
type MarathonGameDTO struct {
	ID              string            `json:"id"`
	PlayerID        string            `json:"playerId"`
	Category        CategoryDTO       `json:"category"`
	Status          string            `json:"status"` // "in_progress", "game_over", "completed", "abandoned"
	Score           int               `json:"score"`  // Total correct answers
	TotalQuestions  int               `json:"totalQuestions"`
	Lives           LivesDTO          `json:"lives"`
	BonusInventory  BonusInventoryDTO `json:"bonusInventory"`
	ShieldActive    bool              `json:"shieldActive"`
	DifficultyLevel string            `json:"difficultyLevel"` // "beginner", "medium", "hard", "master"
	ContinueCount   int               `json:"continueCount"`
	PersonalBest    *int              `json:"personalBest,omitempty"` // Player's best score for comparison
	CurrentQuestion *QuestionDTO      `json:"currentQuestion,omitempty"`
	QuestionNumber  int               `json:"questionNumber"` // 1-based index of next question
	TimeLimit       int               `json:"timeLimit"`      // Seconds for current question
}

// CategoryDTO represents a marathon category
type CategoryDTO struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	IsAllCategories bool   `json:"isAllCategories"` // true for "all categories" mode
}

// LivesDTO represents the lives system state
type LivesDTO struct {
	CurrentLives   int    `json:"currentLives"`
	MaxLives       int    `json:"maxLives"`
	TimeToNextLife int64  `json:"timeToNextLife"` // seconds until next life regenerates
	Label          string `json:"label"`          // "❤️❤️🖤" — UI-ready display
}

// BonusInventoryDTO represents available bonuses
type BonusInventoryDTO struct {
	Shield     int `json:"shield"`     // Protect from one wrong answer
	FiftyFifty int `json:"fiftyFifty"` // Remove 2 incorrect answers
	Skip       int `json:"skip"`       // Skip question without losing life
	Freeze     int `json:"freeze"`     // Add 10 seconds to timer
}

// QuestionDTO represents a quiz question (from quiz aggregate)
type QuestionDTO struct {
	ID       string      `json:"id"`
	Text     string      `json:"text"`
	Answers  []AnswerDTO `json:"answers"`
	Points   int         `json:"points"`
	Position int         `json:"position"`
}

// AnswerDTO represents an answer option
// NOTE: IsCorrect is NOT included - never leak correct answers to client!
type AnswerDTO struct {
	ID       string `json:"id"`
	Text     string `json:"text"`
	Position int    `json:"position"`
}

// PersonalBestDTO represents a personal best record
type PersonalBestDTO struct {
	Category   CategoryDTO `json:"category"`
	BestStreak int         `json:"bestStreak"`
	BestScore  int         `json:"bestScore"`
	AchievedAt int64       `json:"achievedAt"`
}

// LeaderboardEntryDTO represents a leaderboard entry
type LeaderboardEntryDTO struct {
	PlayerID   string `json:"playerId"`
	Username   string `json:"username"`
	BestStreak int    `json:"bestStreak"`
	BestScore  int    `json:"bestScore"`
	Rank       int    `json:"rank"`
	AchievedAt int64  `json:"achievedAt"`
}

// ContinueOfferDTO represents continue options after game over
type ContinueOfferDTO struct {
	Available     bool   `json:"available"`
	CostCoins     int    `json:"costCoins"`
	HasAd         bool   `json:"hasAd"`
	ContinueCount int    `json:"continueCount"`
}

// MilestoneDTO represents the next milestone target
type MilestoneDTO struct {
	Next      int `json:"next"`      // Next milestone target (25, 50, 100, 200, 500)
	Current   int `json:"current"`   // Current score
	Remaining int `json:"remaining"` // Questions remaining to reach milestone
}

// ========================================
// StartMarathon Use Case
// ========================================

// StartMarathonInput is the input for starting a marathon game
type StartMarathonInput struct {
	PlayerID   string  `json:"playerId"`
	CategoryID *string `json:"categoryId,omitempty"` // nil or empty = "all categories"
}

// StartMarathonOutput is the output for starting a marathon game
type StartMarathonOutput struct {
	Game            MarathonGameDTO `json:"game"`
	HasPersonalBest bool            `json:"hasPersonalBest"` // Whether player has previous record
}

// ========================================
// SubmitMarathonAnswer Use Case
// ========================================

// SubmitMarathonAnswerInput is the input for submitting an answer
type SubmitMarathonAnswerInput struct {
	GameID     string `json:"gameId"`
	QuestionID string `json:"questionId"`
	AnswerID   string `json:"answerId"`
	PlayerID   string `json:"playerId"` // For authorization
	TimeTaken  int64  `json:"timeTaken"` // Time taken in milliseconds
}

// SubmitMarathonAnswerOutput is the output for submitting an answer
type SubmitMarathonAnswerOutput struct {
	IsCorrect       bool              `json:"isCorrect"`
	CorrectAnswerID string            `json:"correctAnswerId"`
	TimeTaken       int64             `json:"timeTaken"`
	Score           int               `json:"score"`           // Total correct answers
	TotalQuestions  int               `json:"totalQuestions"`
	DifficultyLevel string            `json:"difficultyLevel"`
	LifeLost        bool              `json:"lifeLost"`
	ShieldConsumed  bool              `json:"shieldConsumed"`
	Lives           LivesDTO          `json:"lives"`           // Full lives state after answer
	BonusInventory  BonusInventoryDTO `json:"bonusInventory"`  // Updated bonuses
	IsGameOver      bool              `json:"isGameOver"`
	NextQuestion    *QuestionDTO      `json:"nextQuestion,omitempty"`
	NextTimeLimit   *int              `json:"nextTimeLimit,omitempty"` // Time limit for next question
	GameOverResult  *GameOverResultDTO `json:"gameOverResult,omitempty"`
	Milestone       *MilestoneDTO     `json:"milestone,omitempty"` // Next milestone progress
}

// GameOverResultDTO contains game over statistics
type GameOverResultDTO struct {
	FinalScore        int              `json:"finalScore"`       // Total correct answers
	TotalQuestions    int              `json:"totalQuestions"`
	IsNewPersonalBest bool             `json:"isNewPersonalBest"`
	PreviousRecord    *int             `json:"previousRecord,omitempty"`
	ContinueOffer     *ContinueOfferDTO `json:"continueOffer,omitempty"`
}

// ========================================
// UseMarathonBonus Use Case
// ========================================

// UseMarathonBonusInput is the input for using a bonus
type UseMarathonBonusInput struct {
	GameID     string `json:"gameId"`
	QuestionID string `json:"questionId"`
	BonusType  string `json:"bonusType"` // "shield", "fifty_fifty", "skip", "freeze"
	PlayerID   string `json:"playerId"`  // For authorization
}

// UseMarathonBonusOutput is the output for using a bonus
type UseMarathonBonusOutput struct {
	BonusType      string            `json:"bonusType"`
	RemainingCount int               `json:"remainingCount"` // Remaining bonuses of this type
	BonusInventory BonusInventoryDTO `json:"bonusInventory"` // Full updated inventory
	BonusResult    BonusResultDTO    `json:"bonusResult"`
}

// BonusResultDTO contains the result of using a bonus
type BonusResultDTO struct {
	// For fifty_fifty: IDs of answers to hide
	HiddenAnswerIDs []string `json:"hiddenAnswerIds,omitempty"`

	// For freeze: new time limit after adding 10s
	NewTimeLimit *int `json:"newTimeLimit,omitempty"`

	// For skip: next question to display
	NextQuestion  *QuestionDTO `json:"nextQuestion,omitempty"`
	NextTimeLimit *int         `json:"nextTimeLimit,omitempty"`

	// For shield: whether shield is now active
	ShieldActive *bool `json:"shieldActive,omitempty"`
}

// ========================================
// ContinueMarathon Use Case
// ========================================

// ContinueMarathonInput is the input for continuing after game over
type ContinueMarathonInput struct {
	GameID        string `json:"gameId"`
	PlayerID      string `json:"playerId"` // For authorization
	PaymentMethod string `json:"paymentMethod"` // "coins" or "ad"
}

// ContinueMarathonOutput is the output for continuing after game over
type ContinueMarathonOutput struct {
	Game            MarathonGameDTO `json:"game"`       // Updated game state with 1 life
	ContinueCount   int             `json:"continueCount"`
	CoinsDeducted    int             `json:"coinsDeducted"`
	NextContinueCost int            `json:"nextContinueCost"` // Cost for next continue
}

// ========================================
// AbandonMarathon Use Case
// ========================================

// AbandonMarathonInput is the input for abandoning a game
type AbandonMarathonInput struct {
	GameID   string `json:"gameId"`
	PlayerID string `json:"playerId"` // For authorization
}

// AbandonMarathonOutput is the output for abandoning a game
type AbandonMarathonOutput struct {
	GameOverResult GameOverResultDTO `json:"gameOverResult"`
}

// ========================================
// GetMarathonStatus Use Case
// ========================================

// GetMarathonStatusInput is the input for getting marathon status
type GetMarathonStatusInput struct {
	PlayerID string `json:"playerId"` // Get active game for this player
}

// GetMarathonStatusOutput is the output for getting marathon status
type GetMarathonStatusOutput struct {
	HasActiveGame  bool              `json:"hasActiveGame"`
	Game           *MarathonGameDTO  `json:"game,omitempty"`
	BonusInventory *BonusInventoryDTO `json:"bonusInventory,omitempty"` // Available bonuses (shown when idle too)
}

// ========================================
// GetMarathonLeaderboard Use Case
// ========================================

// GetMarathonLeaderboardInput is the input for getting leaderboard
type GetMarathonLeaderboardInput struct {
	CategoryID string `json:"categoryId,omitempty"` // Empty = "all categories"
	TimeFrame  string `json:"timeFrame,omitempty"`  // "all_time" (default), "weekly", "daily"
	Limit      int    `json:"limit"`                // Max entries to return
}

// GetMarathonLeaderboardOutput is the output for getting leaderboard
type GetMarathonLeaderboardOutput struct {
	Category   CategoryDTO           `json:"category"`
	TimeFrame  string                `json:"timeFrame"`
	Entries    []LeaderboardEntryDTO `json:"entries"`
	PlayerRank *int                  `json:"playerRank,omitempty"` // Player's rank (if provided)
}

// ========================================
// GetPersonalBests Use Case
// ========================================

// GetPersonalBestsInput is the input for getting personal bests
type GetPersonalBestsInput struct {
	PlayerID string `json:"playerId"`
}

// GetPersonalBestsOutput is the output for getting personal bests
type GetPersonalBestsOutput struct {
	PersonalBests []PersonalBestDTO `json:"personalBests"` // One per category
	OverallBest   *PersonalBestDTO  `json:"overallBest,omitempty"` // Best across all categories
}
