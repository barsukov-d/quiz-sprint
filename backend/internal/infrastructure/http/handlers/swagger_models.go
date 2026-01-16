package handlers

// ========================================
// Swagger Response Models
// ========================================
// These types are ONLY for Swagger documentation.
// They duplicate the application DTOs to make them visible to swag.

// QuizDTO represents basic quiz information
type QuizDTO struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	QuestionsCount int    `json:"questionsCount"`
	TimeLimit      int    `json:"timeLimit"`
	PassingScore   int    `json:"passingScore"`
	CreatedAt      int64  `json:"createdAt"`
}

// QuizDetailDTO represents detailed quiz with questions
type QuizDetailDTO struct {
	ID             string        `json:"id"`
	Title          string        `json:"title"`
	Description    string        `json:"description"`
	Questions      []QuestionDTO `json:"questions"`
	TimeLimit      int           `json:"timeLimit"`
	PassingScore   int           `json:"passingScore"`
	CreatedAt      int64         `json:"createdAt"`
}

// QuestionDTO represents a quiz question
type QuestionDTO struct {
	ID       string      `json:"id"`
	Text     string      `json:"text"`
	Answers  []AnswerDTO `json:"answers"`
	Points   int         `json:"points"`
	Position int         `json:"position"`
}

// AnswerDTO represents an answer option (without correct indicator)
type AnswerDTO struct {
	ID       string `json:"id"`
	Text     string `json:"text"`
	Position int    `json:"position"`
}

// SessionDTO represents a quiz session
type SessionDTO struct {
	ID              string `json:"id"`
	QuizID          string `json:"quizId"`
	UserID          string `json:"userId"`
	CurrentQuestion int    `json:"currentQuestion"`
	Score           int    `json:"score"`
	Status          string `json:"status"`
	StartedAt       int64  `json:"startedAt"`
	CompletedAt     int64  `json:"completedAt,omitempty"`
}

// LeaderboardEntryDTO represents a leaderboard entry
type LeaderboardEntryDTO struct {
	UserID      string `json:"userId"`
	Username    string `json:"username"`
	Score       int    `json:"score"`
	Rank        int    `json:"rank"`
	CompletedAt int64  `json:"completedAt"`
}

// FinalResultDTO contains final quiz results
type FinalResultDTO struct {
	TotalScore     int  `json:"totalScore"`
	MaxScore       int  `json:"maxScore"`
	Percentage     int  `json:"percentage"`
	Passed         bool `json:"passed"`
	QuestionsCount int  `json:"questionsCount"`
	CorrectCount   int  `json:"correctCount"`
}

// ListQuizzesResponse wraps the quiz list response
type ListQuizzesResponse struct {
	Data []QuizDTO `json:"data"`
}

// GetQuizDetailsData wraps quiz details with top scores
type GetQuizDetailsData struct {
	Quiz      QuizDetailDTO         `json:"quiz"`
	TopScores []LeaderboardEntryDTO `json:"topScores"`
}

// GetQuizDetailsResponse wraps the quiz details response
type GetQuizDetailsResponse struct {
	Data GetQuizDetailsData `json:"data"`
}

// StartQuizData contains data for a started quiz
type StartQuizData struct {
	Session        SessionDTO  `json:"session"`
	FirstQuestion  QuestionDTO `json:"firstQuestion"`
	TotalQuestions int         `json:"totalQuestions"`
	TimeLimit      int         `json:"timeLimit"`
}

// StartQuizResponse wraps the start quiz response
type StartQuizResponse struct {
	Data StartQuizData `json:"data"`
}

// SubmitAnswerData contains answer submission result
type SubmitAnswerData struct {
	IsCorrect       bool            `json:"isCorrect"`
	CorrectAnswerID string          `json:"correctAnswerId"`
	PointsEarned    int             `json:"pointsEarned"`
	TotalScore      int             `json:"totalScore"`
	IsQuizCompleted bool            `json:"isQuizCompleted"`
	NextQuestion    *QuestionDTO    `json:"nextQuestion,omitempty"`
	FinalResult     *FinalResultDTO `json:"finalResult,omitempty"`
}

// SubmitAnswerResponse wraps the submit answer response
type SubmitAnswerResponse struct {
	Data SubmitAnswerData `json:"data"`
}

// GetLeaderboardResponse wraps the leaderboard response
type GetLeaderboardResponse struct {
	Data []LeaderboardEntryDTO `json:"data"`
}

// ErrorResponse is the standard error response format
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error details
type ErrorDetail struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
