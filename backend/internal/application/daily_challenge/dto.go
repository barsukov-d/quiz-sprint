package daily_challenge

// ========================================
// Common DTOs
// ========================================

// DailyQuizDTO represents a daily quiz
type DailyQuizDTO struct {
	ID          string   `json:"id"`
	Date        string   `json:"date"` // "2026-01-25"
	QuestionIDs []string `json:"questionIds"`
	ExpiresAt   int64    `json:"expiresAt"`
	CreatedAt   int64    `json:"createdAt"`
}

// DailyGameDTO represents a player's daily challenge attempt
type DailyGameDTO struct {
	ID               string           `json:"id"` // Game ID (deprecated, use GameID)
	GameID           string           `json:"gameId"` // Game ID (matches Swagger spec)
	PlayerID         string           `json:"playerId"`
	DailyQuizID      string           `json:"dailyQuizId"`
	Date             string           `json:"date"`
	Status           string           `json:"status"` // "in_progress", "completed"
	CurrentQuestion  *QuestionDTO     `json:"currentQuestion,omitempty"`
	QuestionIndex    int              `json:"questionIndex"` // 0-9
	TotalQuestions   int              `json:"totalQuestions"` // Always 10
	BaseScore        int              `json:"baseScore"` // Score before streak bonus
	FinalScore       int              `json:"finalScore"` // Score with streak bonus
	CorrectAnswers   int              `json:"correctAnswers"`
	Streak           StreakDTO        `json:"streak"`
	Rank             *int             `json:"rank,omitempty"` // Player's rank in leaderboard
	TimeRemaining    int64            `json:"timeRemaining"` // Seconds until quiz expires
}

// QuestionDTO represents a quiz question (reused from quiz domain)
type QuestionDTO struct {
	ID       string      `json:"id"`
	Text     string      `json:"text"`
	Answers  []AnswerDTO `json:"answers"`
	Points   int         `json:"points"`
	Position int         `json:"position"`
}

// AnswerDTO represents an answer option (NO IsCorrect field!)
type AnswerDTO struct {
	ID       string `json:"id"`
	Text     string `json:"text"`
	Position int    `json:"position"`
}

// StreakDTO represents daily streak information
type StreakDTO struct {
	CurrentStreak  int    `json:"currentStreak"`
	BestStreak     int    `json:"bestStreak"`
	LastPlayedDate string `json:"lastPlayedDate"` // "2026-01-25"
	BonusPercent   int    `json:"bonusPercent"` // 0, 10, 25, 40, 60, or 100
	IsActive       bool   `json:"isActive"` // Still active for today
}

// LeaderboardEntryDTO represents a leaderboard entry
type LeaderboardEntryDTO struct {
	PlayerID       string `json:"playerId"`
	Username       string `json:"username"`
	Score          int    `json:"score"`
	CorrectAnswers int    `json:"correctAnswers"`
	Rank           int    `json:"rank"`
	StreakDays     int    `json:"streakDays"`
	CompletedAt    int64  `json:"completedAt"`
}

// AnsweredQuestionDTO shows the answer after game completion
type AnsweredQuestionDTO struct {
	QuestionID       string   `json:"questionId"`
	QuestionText     string   `json:"questionText"`
	PlayerAnswerID   string   `json:"playerAnswerId"`
	PlayerAnswerText string   `json:"playerAnswerText"`
	CorrectAnswerID  string   `json:"correctAnswerId"`
	CorrectAnswerText string  `json:"correctAnswerText"`
	IsCorrect        bool     `json:"isCorrect"`
	TimeTaken        int64    `json:"timeTaken"` // Milliseconds
	PointsEarned     int      `json:"pointsEarned"`
}

// ========================================
// GetOrCreateDailyQuiz Use Case
// ========================================

type GetOrCreateDailyQuizInput struct {
	Date string `json:"date"` // "2026-01-25" or empty for today
}

type GetOrCreateDailyQuizOutput struct {
	DailyQuiz       DailyQuizDTO `json:"dailyQuiz"`
	TotalPlayers    int          `json:"totalPlayers"` // Players who completed today
	IsNew           bool         `json:"isNew"` // true if just created
}

// ========================================
// StartDailyChallenge Use Case
// ========================================

type StartDailyChallengeInput struct {
	PlayerID string `json:"playerId"`
	Date     string `json:"date,omitempty"` // Optional, defaults to today UTC
}

type StartDailyChallengeOutput struct {
	Game            DailyGameDTO `json:"game"`
	FirstQuestion   QuestionDTO  `json:"firstQuestion"`
	TimeLimit       int          `json:"timeLimit"` // Seconds per question (always 15)
	TotalPlayers    int          `json:"totalPlayers"` // Players who completed today
	TimeToExpire    int64        `json:"timeToExpire"` // Seconds until quiz expires
}

// ========================================
// SubmitDailyAnswer Use Case
// ========================================

type SubmitDailyAnswerInput struct {
	GameID     string `json:"gameId"`
	QuestionID string `json:"questionId"`
	AnswerID   string `json:"answerId"`
	PlayerID   string `json:"playerId"` // For authorization
	TimeTaken  int64  `json:"timeTaken"` // Milliseconds
}

type SubmitDailyAnswerOutput struct {
	QuestionIndex      int               `json:"questionIndex"` // Question just answered (0-9)
	TotalQuestions     int               `json:"totalQuestions"` // Always 10
	RemainingQuestions int               `json:"remainingQuestions"`
	IsGameCompleted    bool              `json:"isGameCompleted"`
	NextQuestion       *QuestionDTO      `json:"nextQuestion,omitempty"` // If game continues
	NextTimeLimit      *int              `json:"nextTimeLimit,omitempty"` // Always 15 if next question exists
	GameResults        *GameResultsDTO   `json:"gameResults,omitempty"` // If game completed
}

// GameResultsDTO contains final results after completing all 10 questions
type GameResultsDTO struct {
	BaseScore         int                     `json:"baseScore"`
	FinalScore        int                     `json:"finalScore"` // With streak bonus
	CorrectAnswers    int                     `json:"correctAnswers"`
	TotalQuestions    int                     `json:"totalQuestions"` // Always 10
	StreakBonus       int                     `json:"streakBonus"` // Bonus percentage (0-100)
	CurrentStreak     int                     `json:"currentStreak"`
	Rank              int                     `json:"rank"` // Player's rank
	TotalPlayers      int                     `json:"totalPlayers"`
	Percentile        int                     `json:"percentile"` // Top X%
	AnsweredQuestions []AnsweredQuestionDTO   `json:"answeredQuestions"` // Full breakdown
	Leaderboard       []LeaderboardEntryDTO   `json:"leaderboard"` // Top players
}

// ========================================
// GetDailyGameStatus Use Case
// ========================================

type GetDailyGameStatusInput struct {
	PlayerID string `json:"playerId"`
	Date     string `json:"date,omitempty"` // Optional, defaults to today
}

type GetDailyGameStatusOutput struct {
	HasPlayed       bool              `json:"hasPlayed"`
	Game            *DailyGameDTO     `json:"game,omitempty"`
	Results         *GameResultsDTO   `json:"results,omitempty"` // Results if game completed
	TimeLimit       *int              `json:"timeLimit,omitempty"` // 15 seconds if in progress
	TimeToExpire    int64             `json:"timeToExpire"` // Seconds until quiz resets
	TotalPlayers    int               `json:"totalPlayers"` // Players who completed today
}

// ========================================
// GetDailyLeaderboard Use Case
// ========================================

type GetDailyLeaderboardInput struct {
	Date     string `json:"date,omitempty"` // Optional, defaults to today
	Limit    int    `json:"limit"` // Max entries
	PlayerID string `json:"playerId,omitempty"` // Optional, to find player's rank
}

type GetDailyLeaderboardOutput struct {
	Date         string                `json:"date"`
	Entries      []LeaderboardEntryDTO `json:"entries"`
	TotalPlayers int                   `json:"totalPlayers"`
	PlayerRank   *int                  `json:"playerRank,omitempty"` // If playerID provided
}

// ========================================
// GetPlayerStreak Use Case
// ========================================

type GetPlayerStreakInput struct {
	PlayerID string `json:"playerId"`
}

type GetPlayerStreakOutput struct {
	Streak         StreakDTO `json:"streak"`
	NextMilestone  int       `json:"nextMilestone"` // Next milestone (3, 7, 14, 30, 100)
	DaysToNext     int       `json:"daysToNext"` // Days until next milestone
	CanRestore     bool      `json:"canRestore"` // If streak was broken yesterday
}
