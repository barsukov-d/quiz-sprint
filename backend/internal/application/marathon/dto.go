package marathon

// ========================================
// Common DTOs
// ========================================

// MarathonGameDTO represents a marathon game state
type MarathonGameDTO struct {
	ID                 string              `json:"id"`
	PlayerID           string              `json:"playerId"`
	Category           CategoryDTO         `json:"category"`
	Status             string              `json:"status"` // "in_progress", "finished", "abandoned"
	CurrentStreak      int                 `json:"currentStreak"`
	MaxStreak          int                 `json:"maxStreak"`
	Lives              LivesDTO            `json:"lives"`
	Hints              HintsDTO            `json:"hints"`
	DifficultyLevel    string              `json:"difficultyLevel"` // "beginner", "medium", "hard", "expert", "master"
	PersonalBestStreak *int                `json:"personalBestStreak,omitempty"`
	CurrentQuestion    *QuestionDTO        `json:"currentQuestion,omitempty"`
	BaseScore          int                 `json:"baseScore"` // Total base score from QuizGameplaySession
}

// CategoryDTO represents a marathon category
type CategoryDTO struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	IsAllCategories bool   `json:"isAllCategories"` // true for "all categories" mode
}

// LivesDTO represents the lives system state
type LivesDTO struct {
	CurrentLives   int   `json:"currentLives"`
	MaxLives       int   `json:"maxLives"`
	TimeToNextLife int64 `json:"timeToNextLife"` // seconds until next life regenerates
}

// HintsDTO represents available hints
type HintsDTO struct {
	FiftyFifty int `json:"fiftyFifty"` // Remove 2 incorrect answers
	ExtraTime  int `json:"extraTime"`  // Add 10 seconds
	Skip       int `json:"skip"`       // Skip question without losing life
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
	FirstQuestion   QuestionDTO     `json:"firstQuestion"`
	TimeLimit       int             `json:"timeLimit"`       // Time limit in seconds (adaptive)
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
	IsCorrect       bool         `json:"isCorrect"`
	CorrectAnswerID string       `json:"correctAnswerId"`
	BasePoints      int          `json:"basePoints"`
	TimeTaken       int64        `json:"timeTaken"`
	CurrentStreak   int          `json:"currentStreak"`
	MaxStreak       int          `json:"maxStreak"`
	DifficultyLevel string       `json:"difficultyLevel"`
	LifeLost        bool         `json:"lifeLost"`
	RemainingLives  int          `json:"remainingLives"`
	IsGameOver      bool         `json:"isGameOver"`
	NextQuestion    *QuestionDTO `json:"nextQuestion,omitempty"`
	NextTimeLimit   *int         `json:"nextTimeLimit,omitempty"` // Time limit for next question (adaptive)
	GameOverResult  *GameOverResultDTO `json:"gameOverResult,omitempty"`
}

// GameOverResultDTO contains game over statistics
type GameOverResultDTO struct {
	FinalStreak        int  `json:"finalStreak"`
	IsNewPersonalBest  bool `json:"isNewPersonalBest"`
	PreviousRecord     *int `json:"previousRecord,omitempty"`
	TotalBaseScore     int  `json:"totalBaseScore"`
	GlobalRank         *int `json:"globalRank,omitempty"` // Player's rank in leaderboard
}

// ========================================
// UseMarathonHint Use Case
// ========================================

// UseMarathonHintInput is the input for using a hint
type UseMarathonHintInput struct {
	GameID     string `json:"gameId"`
	QuestionID string `json:"questionId"`
	HintType   string `json:"hintType"` // "fifty_fifty", "extra_time", "skip"
	PlayerID   string `json:"playerId"` // For authorization
}

// UseMarathonHintOutput is the output for using a hint
type UseMarathonHintOutput struct {
	HintType       string   `json:"hintType"`
	RemainingHints int      `json:"remainingHints"` // Remaining hints of this type
	HintResult     HintResultDTO `json:"hintResult"`
}

// HintResultDTO contains the result of using a hint
type HintResultDTO struct {
	// For fifty_fifty: IDs of answers to hide
	HiddenAnswerIDs []string `json:"hiddenAnswerIds,omitempty"`

	// For extra_time: new time limit
	NewTimeLimit *int `json:"newTimeLimit,omitempty"`

	// For skip: next question
	NextQuestion *QuestionDTO `json:"nextQuestion,omitempty"`
	NextTimeLimit *int `json:"nextTimeLimit,omitempty"`
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
	HasActiveGame bool             `json:"hasActiveGame"`
	Game          *MarathonGameDTO `json:"game,omitempty"`
	TimeLimit     *int             `json:"timeLimit,omitempty"` // Current question time limit
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
	Category CategoryDTO           `json:"category"`
	TimeFrame string               `json:"timeFrame"`
	Entries  []LeaderboardEntryDTO `json:"entries"`
	PlayerRank *int                `json:"playerRank,omitempty"` // Player's rank (if provided)
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
