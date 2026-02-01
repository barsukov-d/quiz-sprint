package handlers

// ========================================
// Swagger Response Models
// ========================================
// These types are ONLY for Swagger documentation.
// They duplicate the application DTOs to make them visible to swag.

// QuizDTO represents basic quiz information
type QuizDTO struct {
	ID             string `json:"id" validate:"required"`
	Title          string `json:"title" validate:"required"`
	Description    string `json:"description"`
	CategoryID     string `json:"categoryId,omitempty"`
	QuestionsCount int    `json:"questionsCount" validate:"required"`
	TimeLimit      int    `json:"timeLimit" validate:"required"`
	PassingScore   int    `json:"passingScore" validate:"required"`
	CreatedAt      int64  `json:"createdAt" validate:"required"`
}

// @name QuizDTO

// QuizDetailDTO represents detailed quiz with questions
type QuizDetailDTO struct {
	ID           string        `json:"id" validate:"required"`
	Title        string        `json:"title" validate:"required"`
	Description  string        `json:"description"`
	CategoryID   string        `json:"categoryId,omitempty"`
	Questions    []QuestionDTO `json:"questions" validate:"required"`
	TimeLimit    int           `json:"timeLimit" validate:"required"`
	PassingScore int           `json:"passingScore" validate:"required"`
	CreatedAt    int64         `json:"createdAt" validate:"required"`
}

// @name QuizDetailDTO

// QuestionDTO represents a quiz question
type QuestionDTO struct {
	ID       string      `json:"id" validate:"required"`
	Text     string      `json:"text" validate:"required"`
	Answers  []AnswerDTO `json:"answers" validate:"required"`
	Points   int         `json:"points" validate:"required"`
	Position int         `json:"position" validate:"required"`
}

// @name QuestionDTO

// AnswerDTO represents an answer option (without correct indicator)
type AnswerDTO struct {
	ID       string `json:"id" validate:"required"`
	Text     string `json:"text" validate:"required"`
	Position int    `json:"position" validate:"required"`
}

// @name AnswerDTO

// SessionDTO represents a quiz session
type SessionDTO struct {
	ID              string `json:"id" validate:"required"`
	QuizID          string `json:"quizId" validate:"required"`
	UserID          string `json:"userId" validate:"required"`
	CurrentQuestion int    `json:"currentQuestion" validate:"required"`
	Score           int    `json:"score" validate:"required"`
	Status          string `json:"status" validate:"required"`
	StartedAt       int64  `json:"startedAt" validate:"required"`
	CompletedAt     int64  `json:"completedAt,omitempty"`
}

// @name SessionDTO

// LeaderboardEntryDTO represents a leaderboard entry
type LeaderboardEntryDTO struct {
	UserID      string `json:"userId" validate:"required"`
	Username    string `json:"username" validate:"required"`
	Score       int    `json:"score" validate:"required"`
	Rank        int    `json:"rank" validate:"required"`
	CompletedAt int64  `json:"completedAt" validate:"required"`
}

// @name LeaderboardEntryDTO

// FinalResultDTO contains final quiz results
type FinalResultDTO struct {
	TotalScore     int  `json:"totalScore" validate:"required"`
	MaxScore       int  `json:"maxScore" validate:"required"`
	Percentage     int  `json:"percentage" validate:"required"`
	Passed         bool `json:"passed" validate:"required"`
	QuestionsCount int  `json:"questionsCount" validate:"required"`
	CorrectCount   int  `json:"correctCount" validate:"required"`
}

// @name FinalResultDTO

// ListQuizzesResponse wraps the quiz list response
type ListQuizzesResponse struct {
	Data []QuizDTO `json:"data" validate:"required"`
}

// @name ListQuizzesResponse

// GetQuizDetailsData wraps quiz details with top scores
type GetQuizDetailsData struct {
	Quiz      QuizDetailDTO         `json:"quiz" validate:"required"`
	TopScores []LeaderboardEntryDTO `json:"topScores" validate:"required"`
}

// @name GetQuizDetailsData

// GetQuizDetailsResponse wraps the quiz details response
type GetQuizDetailsResponse struct {
	Data GetQuizDetailsData `json:"data" validate:"required"`
}

// @name GetQuizDetailsResponse

// StartQuizData contains data for a started quiz
type StartQuizData struct {
	Session              SessionDTO  `json:"session" validate:"required"`
	FirstQuestion        QuestionDTO `json:"firstQuestion" validate:"required"`
	TotalQuestions       int         `json:"totalQuestions" validate:"required"`
	TimeLimit            int         `json:"timeLimit" validate:"required"`
	TimeLimitPerQuestion int         `json:"timeLimitPerQuestion" validate:"required"`
}

// @name StartQuizData

// StartQuizResponse wraps the start quiz response
type StartQuizResponse struct {
	Data StartQuizData `json:"data" validate:"required"`
}

// @name StartQuizResponse

// SubmitAnswerData contains answer submission result
type SubmitAnswerData struct {
	IsCorrect       bool            `json:"isCorrect" validate:"required"`
	CorrectAnswerID string          `json:"correctAnswerId" validate:"required"`
	BasePoints      int             `json:"basePoints" validate:"required"`
	TimeBonus       int             `json:"timeBonus" validate:"required"`
	StreakBonus     int             `json:"streakBonus" validate:"required"`
	PointsEarned    int             `json:"pointsEarned" validate:"required"`
	CurrentStreak   int             `json:"currentStreak" validate:"required"`
	TotalScore      int             `json:"totalScore" validate:"required"`
	IsQuizCompleted bool            `json:"isQuizCompleted" validate:"required"`
	NextQuestion    *QuestionDTO    `json:"nextQuestion,omitempty"`
	FinalResult     *FinalResultDTO `json:"finalResult,omitempty"`
}

// @name SubmitAnswerData

// SubmitAnswerResponse wraps the submit answer response
type SubmitAnswerResponse struct {
	Data SubmitAnswerData `json:"data" validate:"required"`
}

// @name SubmitAnswerResponse

// GetLeaderboardResponse wraps the leaderboard response
type GetLeaderboardResponse struct {
	Data []LeaderboardEntryDTO `json:"data" validate:"required"`
}

// @name GetLeaderboardResponse

// ========================================
// Global Leaderboard Models
// ========================================

// GlobalLeaderboardEntryDTO represents a global leaderboard entry
type GlobalLeaderboardEntryDTO struct {
	UserID           string `json:"userId" validate:"required" example:"user123"`
	Username         string `json:"username" validate:"required" example:"JohnDoe"`
	TotalScore       int    `json:"totalScore" validate:"required,min=0" example:"850"`
	QuizzesCompleted int    `json:"quizzesCompleted" validate:"required,min=0" example:"5"`
	Rank             int    `json:"rank" validate:"required,min=1" example:"1"`
	LastActivityAt   int64  `json:"lastActivityAt" validate:"required" example:"1674567890"`
}

// @name GlobalLeaderboardEntryDTO

// GetGlobalLeaderboardResponse wraps the global leaderboard response
type GetGlobalLeaderboardResponse struct {
	Data []GlobalLeaderboardEntryDTO `json:"data" validate:"required"`
}

// @name GetGlobalLeaderboardResponse

// ========================================
// Category Models
// ========================================

// CategoryDTO represents a quiz category
type CategoryDTO struct {
	ID   string `json:"id" validate:"required"`
	Name string `json:"name" validate:"required"`
}

// @name CategoryDTO

// ListCategoriesResponse wraps the category list response
type ListCategoriesResponse struct {
	Data []CategoryDTO `json:"data" validate:"required"`
}

// @name ListCategoriesResponse

// CreateCategoryRequest represents the request body for creating a category
type CreateCategoryRequest struct {
	Name string `json:"name" validate:"required"`
}

// @name CreateCategoryRequest

// CreateCategoryData contains created category data
type CreateCategoryData struct {
	Category CategoryDTO `json:"category" validate:"required"`
}

// @name CreateCategoryData

// CreateCategoryResponse wraps the create category response
type CreateCategoryResponse struct {
	Data CreateCategoryData `json:"data" validate:"required"`
}

// @name CreateCategoryResponse

// ErrorResponse is the standard error response format
type ErrorResponse struct {
	Error ErrorDetail `json:"error" validate:"required"`
}

// @name ErrorResponse

// ErrorDetail contains error details
type ErrorDetail struct {
	Code    int    `json:"code" validate:"required"`
	Message string `json:"message" validate:"required"`
}

// @name ErrorDetail

// ========================================
// Request Models
// ========================================

// StartQuizRequest is the HTTP request body for starting a quiz
type StartQuizRequest struct {
	UserID string `json:"userId" validate:"required"`
}

// SubmitAnswerRequest is the HTTP request body for submitting an answer
type SubmitAnswerRequest struct {
	QuestionID string `json:"questionId" validate:"required"`
	AnswerID   string `json:"answerId" validate:"required"`
	UserID     string `json:"userId" validate:"required"`
	TimeTaken  int64  `json:"timeTaken" validate:"required,min=0"`
}

// GetActiveSessionRequest is the HTTP request for getting an active session
// UserID is passed as a query parameter
type GetActiveSessionRequest struct {
	UserID string `json:"userId" validate:"required"`
}

// GetActiveSessionResponse wraps the active session response
type GetActiveSessionResponse struct {
	Data GetActiveSessionData `json:"data"`
}

// @name GetActiveSessionResponse

// GetActiveSessionData contains the active session details
type GetActiveSessionData struct {
	Session              SessionDTO  `json:"session" validate:"required"`
	CurrentQuestion      QuestionDTO `json:"currentQuestion" validate:"required"`
	TotalQuestions       int         `json:"totalQuestions" validate:"required"`
	TimeLimit            int         `json:"timeLimit" validate:"required"`
	TimeLimitPerQuestion int         `json:"timeLimitPerQuestion" validate:"required"`
}

// @name GetActiveSessionData

// AbandonSessionRequest is the HTTP request body for abandoning a session
type AbandonSessionRequest struct {
	UserID string `json:"userId" validate:"required"`
}

// @name AbandonSessionRequest

// ========================================
// User Models
// ========================================

// UserDTO represents a user
type UserDTO struct {
	ID               string `json:"id" validate:"required"`
	Username         string `json:"username" validate:"required"`
	TelegramUsername string `json:"telegramUsername,omitempty"`
	Email            string `json:"email,omitempty"`
	AvatarURL        string `json:"avatarUrl,omitempty"`
	LanguageCode     string `json:"languageCode" validate:"required"`
	IsBlocked        bool   `json:"isBlocked" validate:"required"`
	CreatedAt        int64  `json:"createdAt" validate:"required"`
	UpdatedAt        int64  `json:"updatedAt" validate:"required"`
}

// @name UserDTO

// UserProfileDTO is a lightweight user profile for leaderboards
type UserProfileDTO struct {
	ID               string `json:"id" validate:"required"`
	Username         string `json:"username" validate:"required"`
	TelegramUsername string `json:"telegramUsername,omitempty"`
	AvatarURL        string `json:"avatarUrl,omitempty"`
}

// @name UserProfileDTO

// RegisterUserRequest is the HTTP request body for registering a user
// NOTE: All fields are optional because actual user data comes from Authorization header (Telegram init data)
// The backend extracts validated user data from the middleware
type RegisterUserRequest struct {
	// All fields optional - kept for backwards compatibility
}

// @name RegisterUserRequest

// RegisterUserData contains user registration result
type RegisterUserData struct {
	User      UserDTO `json:"user" validate:"required"`
	IsNewUser bool    `json:"isNewUser" validate:"required"`
}

// @name RegisterUserData

// RegisterUserResponse wraps the register user response
type RegisterUserResponse struct {
	Data RegisterUserData `json:"data" validate:"required"`
}

// @name RegisterUserResponse

// GetUserResponse wraps the get user response
type GetUserResponse struct {
	Data UserDTO `json:"data" validate:"required"`
}

// @name GetUserResponse

// UpdateUserProfileRequest is the HTTP request body for updating user profile
type UpdateUserProfileRequest struct {
	Username         string `json:"username,omitempty"`
	TelegramUsername string `json:"telegramUsername,omitempty"`
	Email            string `json:"email,omitempty"`
	AvatarURL        string `json:"avatarUrl,omitempty"`
	LanguageCode     string `json:"languageCode,omitempty"`
}

// @name UpdateUserProfileRequest

// UpdateUserProfileResponse wraps the update user profile response
type UpdateUserProfileResponse struct {
	Data UserDTO `json:"data" validate:"required"`
}

// @name UpdateUserProfileResponse

// ListUsersResponse wraps the list users response
type ListUsersResponse struct {
	Data  []UserDTO `json:"data" validate:"required"`
	Total int       `json:"total" validate:"required"`
}

// @name ListUsersResponse

// ========================================
// Session Results
// ========================================

// SessionResultsData contains detailed results of a quiz session
type SessionResultsData struct {
	Session         SessionDTO `json:"session" validate:"required"`
	Quiz            QuizDTO    `json:"quiz" validate:"required"`
	TotalQuestions  int        `json:"totalQuestions" validate:"required"`
	CorrectAnswers  int        `json:"correctAnswers" validate:"required"`
	TimeSpent       int64      `json:"timeSpent" validate:"required"`       // seconds
	Passed          bool       `json:"passed" validate:"required"`
	ScorePercentage int        `json:"scorePercentage" validate:"required"` // 0-100
	LongestStreak   int        `json:"longestStreak" validate:"required"`
	AvgAnswerTime   float64    `json:"avgAnswerTime" validate:"required"` // seconds
}

// @name SessionResultsData

// GetSessionResultsResponse wraps session results
type GetSessionResultsResponse struct {
	Data SessionResultsData `json:"data" validate:"required"`
}

// @name GetSessionResultsResponse

// ========================================
// Daily Quiz & Random Quiz Models
// ========================================

// DailyQuizUserResultDTO contains user's result for daily quiz (if completed)
type DailyQuizUserResultDTO struct {
	Score       int   `json:"score" validate:"required"`
	Rank        int   `json:"rank" validate:"required"`
	CompletedAt int64 `json:"completedAt" validate:"required"`
}

// @name DailyQuizUserResultDTO

// DailyQuizData contains daily quiz details with completion status
type DailyQuizData struct {
	Quiz             QuizDetailDTO           `json:"quiz" validate:"required"`
	CompletionStatus string                  `json:"completionStatus" validate:"required"` // "not_attempted" | "completed"
	UserResult       *DailyQuizUserResultDTO `json:"userResult,omitempty"`
	TopScores        []LeaderboardEntryDTO   `json:"topScores" validate:"required"`
}

// @name DailyQuizData

// GetDailyQuizResponse wraps daily quiz response
type GetDailyQuizResponse struct {
	Data DailyQuizData `json:"data" validate:"required"`
}

// @name GetDailyQuizResponse

// GetRandomQuizResponse wraps random quiz response
type GetRandomQuizResponse struct {
	Data GetQuizDetailsData `json:"data" validate:"required"`
}

// @name GetRandomQuizResponse

// ========================================
// User Active Sessions Models
// ========================================

// SessionSummaryDTO represents an active session summary
type SessionSummaryDTO struct {
	SessionID       string `json:"sessionId" validate:"required"`
	QuizID          string `json:"quizId" validate:"required"`
	QuizTitle       string `json:"quizTitle" validate:"required"`
	CurrentQuestion int    `json:"currentQuestion" validate:"required"`
	TotalQuestions  int    `json:"totalQuestions" validate:"required"`
	Score           int    `json:"score" validate:"required"`
	StartedAt       int64  `json:"startedAt" validate:"required"`
}

// @name SessionSummaryDTO

// GetUserActiveSessionsResponse wraps user active sessions
type GetUserActiveSessionsResponse struct {
	Data []SessionSummaryDTO `json:"data" validate:"required"`
}

// @name GetUserActiveSessionsResponse

// ========================================
// Marathon Mode Models
// ========================================

// MarathonCategoryDTO represents a marathon category
type MarathonCategoryDTO struct {
	ID              string `json:"id" validate:"required"`
	Name            string `json:"name" validate:"required"`
	IsAllCategories bool   `json:"isAllCategories" validate:"required"`
}

// @name MarathonCategoryDTO

// MarathonLivesDTO represents the lives system state
type MarathonLivesDTO struct {
	CurrentLives   int   `json:"currentLives" validate:"required"`
	MaxLives       int   `json:"maxLives" validate:"required"`
	TimeToNextLife int64 `json:"timeToNextLife" validate:"required"`
}

// @name MarathonLivesDTO

// MarathonHintsDTO represents available hints
type MarathonHintsDTO struct {
	FiftyFifty int `json:"fiftyFifty" validate:"required"`
	ExtraTime  int `json:"extraTime" validate:"required"`
	Skip       int `json:"skip" validate:"required"`
}

// @name MarathonHintsDTO

// MarathonGameDTO represents a marathon game state
type MarathonGameDTO struct {
	ID                 string               `json:"id" validate:"required"`
	PlayerID           string               `json:"playerId" validate:"required"`
	Category           MarathonCategoryDTO  `json:"category" validate:"required"`
	Status             string               `json:"status" validate:"required"`
	CurrentStreak      int                  `json:"currentStreak" validate:"required"`
	MaxStreak          int                  `json:"maxStreak" validate:"required"`
	Lives              MarathonLivesDTO     `json:"lives" validate:"required"`
	Hints              MarathonHintsDTO     `json:"hints" validate:"required"`
	DifficultyLevel    string               `json:"difficultyLevel" validate:"required"`
	PersonalBestStreak *int                 `json:"personalBestStreak,omitempty"`
	CurrentQuestion    *QuestionDTO         `json:"currentQuestion,omitempty"`
	BaseScore          int                  `json:"baseScore" validate:"required"`
}

// @name MarathonGameDTO

// MarathonPersonalBestDTO represents a personal best record
type MarathonPersonalBestDTO struct {
	Category   MarathonCategoryDTO `json:"category" validate:"required"`
	BestStreak int                 `json:"bestStreak" validate:"required"`
	BestScore  int                 `json:"bestScore" validate:"required"`
	AchievedAt int64               `json:"achievedAt" validate:"required"`
}

// @name MarathonPersonalBestDTO

// MarathonLeaderboardEntryDTO represents a marathon leaderboard entry
type MarathonLeaderboardEntryDTO struct {
	PlayerID   string `json:"playerId" validate:"required"`
	Username   string `json:"username" validate:"required"`
	BestStreak int    `json:"bestStreak" validate:"required"`
	BestScore  int    `json:"bestScore" validate:"required"`
	Rank       int    `json:"rank" validate:"required"`
	AchievedAt int64  `json:"achievedAt" validate:"required"`
}

// @name MarathonLeaderboardEntryDTO

// MarathonGameOverResultDTO contains game over statistics
type MarathonGameOverResultDTO struct {
	FinalStreak       int  `json:"finalStreak" validate:"required"`
	IsNewPersonalBest bool `json:"isNewPersonalBest" validate:"required"`
	PreviousRecord    *int `json:"previousRecord,omitempty"`
	TotalBaseScore    int  `json:"totalBaseScore" validate:"required"`
	GlobalRank        *int `json:"globalRank,omitempty"`
}

// @name MarathonGameOverResultDTO

// MarathonHintResultDTO contains the result of using a hint
type MarathonHintResultDTO struct {
	HiddenAnswerIDs []string     `json:"hiddenAnswerIds,omitempty"`
	NewTimeLimit    *int         `json:"newTimeLimit,omitempty"`
	NextQuestion    *QuestionDTO `json:"nextQuestion,omitempty"`
	NextTimeLimit   *int         `json:"nextTimeLimit,omitempty"`
}

// @name MarathonHintResultDTO

// ========================================
// Marathon Request Models
// ========================================

// StartMarathonRequest is the HTTP request body for starting a marathon
type StartMarathonRequest struct {
	PlayerID   string  `json:"playerId" validate:"required"`
	CategoryID *string `json:"categoryId,omitempty"`
}

// @name StartMarathonRequest

// SubmitMarathonAnswerRequest is the HTTP request body for submitting an answer
type SubmitMarathonAnswerRequest struct {
	QuestionID string `json:"questionId" validate:"required"`
	AnswerID   string `json:"answerId" validate:"required"`
	PlayerID   string `json:"playerId" validate:"required"`
	TimeTaken  int64  `json:"timeTaken" validate:"required,min=0"`
}

// @name SubmitMarathonAnswerRequest

// UseMarathonHintRequest is the HTTP request body for using a hint
type UseMarathonHintRequest struct {
	QuestionID string `json:"questionId" validate:"required"`
	HintType   string `json:"hintType" validate:"required"`
	PlayerID   string `json:"playerId" validate:"required"`
}

// @name UseMarathonHintRequest

// AbandonMarathonRequest is the HTTP request body for abandoning a marathon
type AbandonMarathonRequest struct {
	PlayerID string `json:"playerId" validate:"required"`
}

// @name AbandonMarathonRequest

// ========================================
// Marathon Response Models
// ========================================

// StartMarathonData contains data for a started marathon
type StartMarathonData struct {
	Game            MarathonGameDTO `json:"game" validate:"required"`
	FirstQuestion   QuestionDTO     `json:"firstQuestion" validate:"required"`
	TimeLimit       int             `json:"timeLimit" validate:"required"`
	HasPersonalBest bool            `json:"hasPersonalBest" validate:"required"`
}

// @name StartMarathonData

// StartMarathonResponse wraps the start marathon response
type StartMarathonResponse struct {
	Data StartMarathonData `json:"data" validate:"required"`
}

// @name StartMarathonResponse

// SubmitMarathonAnswerData contains answer submission result
type SubmitMarathonAnswerData struct {
	IsCorrect       bool                       `json:"isCorrect" validate:"required"`
	CorrectAnswerID string                     `json:"correctAnswerId" validate:"required"`
	BasePoints      int                        `json:"basePoints" validate:"required"`
	TimeTaken       int64                      `json:"timeTaken" validate:"required"`
	CurrentStreak   int                        `json:"currentStreak" validate:"required"`
	MaxStreak       int                        `json:"maxStreak" validate:"required"`
	DifficultyLevel string                     `json:"difficultyLevel" validate:"required"`
	LifeLost        bool                       `json:"lifeLost" validate:"required"`
	RemainingLives  int                        `json:"remainingLives" validate:"required"`
	IsGameOver      bool                       `json:"isGameOver" validate:"required"`
	NextQuestion    *QuestionDTO               `json:"nextQuestion,omitempty"`
	NextTimeLimit   *int                       `json:"nextTimeLimit,omitempty"`
	GameOverResult  *MarathonGameOverResultDTO `json:"gameOverResult,omitempty"`
}

// @name SubmitMarathonAnswerData

// SubmitMarathonAnswerResponse wraps the submit answer response
type SubmitMarathonAnswerResponse struct {
	Data SubmitMarathonAnswerData `json:"data" validate:"required"`
}

// @name SubmitMarathonAnswerResponse

// UseMarathonHintData contains hint usage result
type UseMarathonHintData struct {
	HintType       string                `json:"hintType" validate:"required"`
	RemainingHints int                   `json:"remainingHints" validate:"required"`
	HintResult     MarathonHintResultDTO `json:"hintResult" validate:"required"`
}

// @name UseMarathonHintData

// UseMarathonHintResponse wraps the use hint response
type UseMarathonHintResponse struct {
	Data UseMarathonHintData `json:"data" validate:"required"`
}

// @name UseMarathonHintResponse

// AbandonMarathonResponse wraps the abandon marathon response
type AbandonMarathonResponse struct {
	Data MarathonGameOverResultDTO `json:"data" validate:"required"`
}

// @name AbandonMarathonResponse

// GetMarathonStatusData contains marathon status information
type GetMarathonStatusData struct {
	HasActiveGame bool             `json:"hasActiveGame" validate:"required"`
	Game          *MarathonGameDTO `json:"game,omitempty"`
	TimeLimit     *int             `json:"timeLimit,omitempty"`
}

// @name GetMarathonStatusData

// GetMarathonStatusResponse wraps the marathon status response
type GetMarathonStatusResponse struct {
	Data GetMarathonStatusData `json:"data" validate:"required"`
}

// @name GetMarathonStatusResponse

// GetPersonalBestsData contains personal best records
type GetPersonalBestsData struct {
	PersonalBests []MarathonPersonalBestDTO `json:"personalBests" validate:"required"`
	OverallBest   *MarathonPersonalBestDTO  `json:"overallBest,omitempty"`
}

// @name GetPersonalBestsData

// GetPersonalBestsResponse wraps the personal bests response
type GetPersonalBestsResponse struct {
	Data GetPersonalBestsData `json:"data" validate:"required"`
}

// @name GetPersonalBestsResponse

// GetMarathonLeaderboardData contains leaderboard information
type GetMarathonLeaderboardData struct {
	Category   MarathonCategoryDTO           `json:"category" validate:"required"`
	TimeFrame  string                        `json:"timeFrame" validate:"required"`
	Entries    []MarathonLeaderboardEntryDTO `json:"entries" validate:"required"`
	PlayerRank *int                          `json:"playerRank,omitempty"`
}

// @name GetMarathonLeaderboardData

// GetMarathonLeaderboardResponse wraps the leaderboard response
type GetMarathonLeaderboardResponse struct {
	Data GetMarathonLeaderboardData `json:"data" validate:"required"`
}

// @name GetMarathonLeaderboardResponse

// ========================================
// Daily Challenge Models
// ========================================

// StartDailyChallengeRequest is the HTTP request for starting daily challenge
type StartDailyChallengeRequest struct {
	PlayerID string `json:"playerId" validate:"required"`
	Date     string `json:"date,omitempty"` // YYYY-MM-DD, defaults to today
}

// @name StartDailyChallengeRequest

// SubmitDailyAnswerRequest is the HTTP request for submitting answer
type SubmitDailyAnswerRequest struct {
	QuestionID string `json:"questionId" validate:"required"`
	AnswerID   string `json:"answerId" validate:"required"`
	PlayerID   string `json:"playerId" validate:"required"`
	TimeTaken  int64  `json:"timeTaken" validate:"required,min=0"`
}

// @name SubmitDailyAnswerRequest

// DailyGameDTO represents a daily challenge game
type DailyGameDTO struct {
	GameID                string       `json:"gameId" validate:"required"`
	PlayerID              string       `json:"playerId" validate:"required"`
	DailyQuizID           string       `json:"dailyQuizId" validate:"required"`
	Date                  string       `json:"date" validate:"required"` // YYYY-MM-DD
	Status                string       `json:"status" validate:"required"` // "in_progress" | "completed"
	CurrentQuestion       *QuestionDTO `json:"currentQuestion,omitempty"`
	QuestionIndex         int          `json:"questionIndex" validate:"required"`
	TotalQuestions        int          `json:"totalQuestions" validate:"required"`
	BaseScore             int          `json:"baseScore" validate:"required"`
	FinalScore            int          `json:"finalScore" validate:"required"`
	CorrectAnswers        int          `json:"correctAnswers" validate:"required"`
	Streak                StreakDTO    `json:"streak" validate:"required"`
	Rank                  *int         `json:"rank,omitempty"`
	TimeRemaining         int64        `json:"timeRemaining" validate:"required"` // Seconds until daily quiz expires
	QuestionTimeRemaining *int         `json:"questionTimeRemaining,omitempty"` // Seconds remaining for current question
}

// @name DailyGameDTO

// GameResultsDTO contains final results after completing all questions
type GameResultsDTO struct {
	BaseScore          int                     `json:"baseScore" validate:"required"`
	FinalScore         int                     `json:"finalScore" validate:"required"`
	CorrectAnswers     int                     `json:"correctAnswers" validate:"required"`
	TotalQuestions     int                     `json:"totalQuestions" validate:"required"`
	StreakBonus        int                     `json:"streakBonus" validate:"required"`
	CurrentStreak      int                     `json:"currentStreak" validate:"required"`
	Rank               int                     `json:"rank" validate:"required"`
	TotalPlayers       int                     `json:"totalPlayers" validate:"required"`
	Percentile         int                     `json:"percentile" validate:"required"`
	ChestReward        ChestRewardDTO          `json:"chestReward" validate:"required"`
	AnsweredQuestions  []AnsweredQuestionDTO   `json:"answeredQuestions" validate:"required"`
	Leaderboard        []LeaderboardEntryDTO   `json:"leaderboard" validate:"required"`
}

// @name GameResultsDTO

// ReviewAnswerDTO contains answer review info (DEPRECATED - use AnsweredQuestionDTO)
type ReviewAnswerDTO struct {
	QuestionID     string `json:"questionId" validate:"required"`
	QuestionText   string `json:"questionText" validate:"required"`
	SelectedAnswer string `json:"selectedAnswer" validate:"required"`
	CorrectAnswer  string `json:"correctAnswer" validate:"required"`
	IsCorrect      bool   `json:"isCorrect" validate:"required"`
}

// @name ReviewAnswerDTO

// AnsweredQuestionDTO shows the answer after game completion
type AnsweredQuestionDTO struct {
	QuestionID        string `json:"questionId" validate:"required"`
	QuestionText      string `json:"questionText" validate:"required"`
	PlayerAnswerID    string `json:"playerAnswerId" validate:"required"`
	PlayerAnswerText  string `json:"playerAnswerText" validate:"required"`
	CorrectAnswerID   string `json:"correctAnswerId" validate:"required"`
	CorrectAnswerText string `json:"correctAnswerText" validate:"required"`
	IsCorrect         bool   `json:"isCorrect" validate:"required"`
	TimeTaken         int64  `json:"timeTaken" validate:"required"`
	PointsEarned      int    `json:"pointsEarned" validate:"required"`
}

// @name AnsweredQuestionDTO

// StreakDTO represents player's streak info
type StreakDTO struct {
	CurrentStreak  int    `json:"currentStreak" validate:"required"`
	BestStreak     int    `json:"bestStreak" validate:"required"`
	LastPlayedDate string `json:"lastPlayedDate" validate:"required"` // YYYY-MM-DD
	BonusPercent   int    `json:"bonusPercent" validate:"required"`
	IsActive       bool   `json:"isActive" validate:"required"`
}

// @name StreakDTO

// StartDailyChallengeData contains start response data
type StartDailyChallengeData struct {
	Game           DailyGameDTO `json:"game" validate:"required"`
	FirstQuestion  QuestionDTO  `json:"firstQuestion" validate:"required"`
	TimeLimit      int          `json:"timeLimit" validate:"required"`
	TotalPlayers   int          `json:"totalPlayers" validate:"required"`
	TimeToExpire   int64        `json:"timeToExpire" validate:"required"`
}

// @name StartDailyChallengeData

// StartDailyChallengeResponse wraps start response
type StartDailyChallengeResponse struct {
	Data StartDailyChallengeData `json:"data" validate:"required"`
}

// @name StartDailyChallengeResponse

// SubmitDailyAnswerData contains submit answer response data
type SubmitDailyAnswerData struct {
	QuestionIndex      int              `json:"questionIndex" validate:"required"`
	TotalQuestions     int              `json:"totalQuestions" validate:"required"`
	RemainingQuestions int              `json:"remainingQuestions" validate:"required"`
	IsGameCompleted    bool             `json:"isGameCompleted" validate:"required"`
	IsCorrect          bool             `json:"isCorrect" validate:"required"`
	CorrectAnswerID    string           `json:"correctAnswerId" validate:"required"`
	NextQuestion       *QuestionDTO     `json:"nextQuestion,omitempty"`
	NextTimeLimit      *int             `json:"nextTimeLimit,omitempty"`
	GameResults        *GameResultsDTO  `json:"gameResults,omitempty"`
}

// @name SubmitDailyAnswerData

// SubmitDailyAnswerResponse wraps submit response
type SubmitDailyAnswerResponse struct {
	Data SubmitDailyAnswerData `json:"data" validate:"required"`
}

// @name SubmitDailyAnswerResponse

// GetDailyStatusData contains status response data
type GetDailyStatusData struct {
	HasPlayed     bool             `json:"hasPlayed" validate:"required"`
	Game          *DailyGameDTO    `json:"game,omitempty"`
	Results       *GameResultsDTO  `json:"results,omitempty"`
	TimeLimit     *int             `json:"timeLimit,omitempty"` // Fixed 15 seconds per question
	TimeRemaining *int             `json:"timeRemaining,omitempty"` // Seconds remaining for current question (server-side timer)
	TimeToExpire  int64            `json:"timeToExpire" validate:"required"` // Seconds until daily quiz expires
	TotalPlayers  int              `json:"totalPlayers" validate:"required"`
}

// @name GetDailyStatusData

// GetDailyStatusResponse wraps status response
type GetDailyStatusResponse struct {
	Data GetDailyStatusData `json:"data" validate:"required"`
}

// @name GetDailyStatusResponse

// GetDailyLeaderboardData contains leaderboard response data
type GetDailyLeaderboardData struct {
	Date         string                `json:"date" validate:"required"`
	Entries      []LeaderboardEntryDTO `json:"entries" validate:"required"`
	TotalPlayers int                   `json:"totalPlayers" validate:"required"`
	PlayerRank   *int                  `json:"playerRank,omitempty"`
}

// @name GetDailyLeaderboardData

// GetDailyLeaderboardResponse wraps leaderboard response
type GetDailyLeaderboardResponse struct {
	Data GetDailyLeaderboardData `json:"data" validate:"required"`
}

// @name GetDailyLeaderboardResponse

// GetPlayerStreakData contains streak response data
type GetPlayerStreakData struct {
	Streak        StreakDTO `json:"streak" validate:"required"`
	NextMilestone int       `json:"nextMilestone" validate:"required"`
	DaysToNext    int       `json:"daysToNext" validate:"required"`
	CanRestore    bool      `json:"canRestore" validate:"required"`
}

// @name GetPlayerStreakData

// GetPlayerStreakResponse wraps streak response
type GetPlayerStreakResponse struct {
	Data GetPlayerStreakData `json:"data" validate:"required"`
}

// @name GetPlayerStreakResponse

// ChestRewardDTO represents chest rewards
type ChestRewardDTO struct {
	ChestType       string   `json:"chestType" validate:"required"`
	Coins           int      `json:"coins" validate:"required"`
	PvpTickets      int      `json:"pvpTickets" validate:"required"`
	MarathonBonuses []string `json:"marathonBonuses" validate:"required"`
}

// @name ChestRewardDTO

// OpenChestRequest is the HTTP request for opening chest
type OpenChestRequest struct {
	PlayerID string `json:"playerId" validate:"required"`
}

// @name OpenChestRequest

// OpenChestData contains chest opening response data
type OpenChestData struct {
	ChestType      string         `json:"chestType" validate:"required"`
	Rewards        ChestRewardDTO `json:"rewards" validate:"required"`
	StreakBonus    float64        `json:"streakBonus" validate:"required"`
	PremiumApplied bool           `json:"premiumApplied" validate:"required"`
}

// @name OpenChestData

// OpenChestResponse wraps chest opening response
type OpenChestResponse struct {
	Data OpenChestData `json:"data" validate:"required"`
}

// @name OpenChestResponse

// RetryChallengeRequest is the HTTP request for retrying challenge
type RetryChallengeRequest struct {
	PlayerID      string `json:"playerId" validate:"required"`
	PaymentMethod string `json:"paymentMethod" validate:"required"` // "coins" or "ad"
}

// @name RetryChallengeRequest

// RetryChallengeData contains retry response data
type RetryChallengeData struct {
	NewGameID      string      `json:"newGameId" validate:"required"`
	FirstQuestion  QuestionDTO `json:"firstQuestion" validate:"required"`
	CoinsDeducted  int         `json:"coinsDeducted" validate:"required"`
	RemainingCoins int         `json:"remainingCoins" validate:"required"`
	TimeLimit      int         `json:"timeLimit" validate:"required"`
}

// @name RetryChallengeData

// RetryChallengeResponse wraps retry response
type RetryChallengeResponse struct {
	Data RetryChallengeData `json:"data" validate:"required"`
}

// @name RetryChallengeResponse

// ========================================
// Admin DTOs (for testing/debug endpoints)
// ========================================

// AdminUpdateStreakRequest is the request body for updating player streak
type AdminUpdateStreakRequest struct {
	PlayerID       string `json:"playerId" validate:"required"`
	CurrentStreak  *int   `json:"currentStreak,omitempty"`
	BestStreak     *int   `json:"bestStreak,omitempty"`
	LastPlayedDate *string `json:"lastPlayedDate,omitempty"` // YYYY-MM-DD
}

// @name AdminUpdateStreakRequest

// AdminUpdateStreakResponse wraps the update streak response
type AdminUpdateStreakResponse struct {
	Data struct {
		Updated  int64  `json:"updated"`
		PlayerID string `json:"playerId"`
	} `json:"data"`
}

// @name AdminUpdateStreakResponse

// AdminDeleteGamesResponse wraps the delete games response
type AdminDeleteGamesResponse struct {
	Data struct {
		Deleted  int64  `json:"deleted"`
		PlayerID string `json:"playerId"`
		Date     string `json:"date"`
	} `json:"data"`
}

// @name AdminDeleteGamesResponse

// AdminGameInfo represents a single game in the list response
type AdminGameInfo struct {
	ID             string `json:"id"`
	Date           string `json:"date"`
	Status         string `json:"status"`
	AttemptNumber  int    `json:"attemptNumber"`
	CurrentStreak  int    `json:"currentStreak"`
	BestStreak     int    `json:"bestStreak"`
	LastPlayedDate string `json:"lastPlayedDate"`
	BaseScore      int    `json:"baseScore"`
	ChestType      string `json:"chestType,omitempty"`
	ChestCoins     int64  `json:"chestCoins,omitempty"`
	Rank           int64  `json:"rank,omitempty"`
}

// @name AdminGameInfo

// AdminListGamesResponse wraps the list games response
type AdminListGamesResponse struct {
	Data struct {
		PlayerID string          `json:"playerId"`
		Games    []AdminGameInfo `json:"games"`
		Count    int             `json:"count"`
	} `json:"data"`
}

// @name AdminListGamesResponse

// AdminSimulateStreakRequest is the request body for simulating a streak
type AdminSimulateStreakRequest struct {
	PlayerID  string `json:"playerId" validate:"required"`
	Days      int    `json:"days" validate:"required"` // 1-365
	BaseScore int    `json:"baseScore,omitempty"`       // default 40
}

// @name AdminSimulateStreakRequest

// AdminSimulateStreakResponse wraps the simulate streak response
type AdminSimulateStreakResponse struct {
	Data struct {
		PlayerID    string `json:"playerId"`
		DaysCreated int    `json:"daysCreated"`
		StreakBuilt int    `json:"streakBuilt"`
		DateRange   struct {
			From string `json:"from"`
			To   string `json:"to"`
		} `json:"dateRange"`
	} `json:"data"`
}

// @name AdminSimulateStreakResponse

// AdminResetPlayerResponse wraps the full player reset response
type AdminResetPlayerResponse struct {
	Data struct {
		PlayerID     string         `json:"playerId"`
		TotalDeleted int64          `json:"totalDeleted"`
		Deleted      map[string]int64 `json:"deleted"`
	} `json:"data"`
}

// @name AdminResetPlayerResponse
