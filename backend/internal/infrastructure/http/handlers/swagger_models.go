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
	Session        SessionDTO  `json:"session" validate:"required"`
	FirstQuestion  QuestionDTO `json:"firstQuestion" validate:"required"`
	TotalQuestions int         `json:"totalQuestions" validate:"required"`
	TimeLimit      int         `json:"timeLimit" validate:"required"`
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
	PointsEarned    int             `json:"pointsEarned" validate:"required"`
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
}
