package quiz

// ========================================
// Common DTOs
// ========================================

// QuizDTO is a data transfer object for Quiz
type QuizDTO struct {
	ID             string        `json:"id"`
	Title          string        `json:"title"`
	Description    string        `json:"description"`
	CategoryID     string        `json:"categoryId,omitempty"`
	QuestionsCount int           `json:"questionsCount"`
	TimeLimit      int           `json:"timeLimit"`
	PassingScore   int           `json:"passingScore"`
	CreatedAt      int64         `json:"createdAt"`
}

// QuizDetailDTO is a detailed quiz DTO with questions
type QuizDetailDTO struct {
	ID             string        `json:"id"`
	Title          string        `json:"title"`
	Description    string        `json:"description"`
	CategoryID     string        `json:"categoryId,omitempty"`
	Questions      []QuestionDTO `json:"questions"`
	TimeLimit      int           `json:"timeLimit"`
	PassingScore   int           `json:"passingScore"`
	CreatedAt      int64         `json:"createdAt"`
}

// QuestionDTO is a data transfer object for Question
type QuestionDTO struct {
	ID       string      `json:"id"`
	Text     string      `json:"text"`
	Answers  []AnswerDTO `json:"answers"`
	Points   int         `json:"points"`
	Position int         `json:"position"`
}

// AnswerDTO is a data transfer object for Answer
// NOTE: IsCorrect is NOT included - never leak correct answers to client!
type AnswerDTO struct {
	ID       string `json:"id"`
	Text     string `json:"text"`
	Position int    `json:"position"`
}

// SessionDTO is a data transfer object for QuizSession
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

// LeaderboardEntryDTO is a data transfer object for LeaderboardEntry
type LeaderboardEntryDTO struct {
	UserID      string `json:"userId"`
	Username    string `json:"username"`
	Score       int    `json:"score"`
	Rank        int    `json:"rank"`
	CompletedAt int64  `json:"completedAt"`
}

// ========================================
// StartQuiz Use Case
// ========================================

// StartQuizInput is the input DTO for StartQuiz use case
type StartQuizInput struct {
	QuizID string `json:"quizId"`
	UserID string `json:"userId"`
}

// StartQuizOutput is the output DTO for StartQuiz use case
type StartQuizOutput struct {
	Session       SessionDTO    `json:"session"`
	FirstQuestion QuestionDTO   `json:"firstQuestion"`
	TotalQuestions int          `json:"totalQuestions"`
	TimeLimit     int           `json:"timeLimit"`
}

// ========================================
// SubmitAnswer Use Case
// ========================================

// SubmitAnswerInput is the input DTO for SubmitAnswer use case
type SubmitAnswerInput struct {
	SessionID  string `json:"sessionId"`
	QuestionID string `json:"questionId"`
	AnswerID   string `json:"answerId"`
	UserID     string `json:"userId"`
}

// SubmitAnswerOutput is the output DTO for SubmitAnswer use case
type SubmitAnswerOutput struct {
	IsCorrect       bool         `json:"isCorrect"`
	CorrectAnswerID string       `json:"correctAnswerId"`
	PointsEarned    int          `json:"pointsEarned"`
	TotalScore      int          `json:"totalScore"`
	IsQuizCompleted bool         `json:"isQuizCompleted"`
	NextQuestion    *QuestionDTO `json:"nextQuestion,omitempty"`
	FinalResult     *FinalResultDTO `json:"finalResult,omitempty"`
}

// FinalResultDTO contains the final quiz result
type FinalResultDTO struct {
	TotalScore     int  `json:"totalScore"`
	MaxScore       int  `json:"maxScore"`
	Percentage     int  `json:"percentage"`
	Passed         bool `json:"passed"`
	QuestionsCount int  `json:"questionsCount"`
	CorrectCount   int  `json:"correctCount"`
}

// ========================================
// GetLeaderboard Use Case
// ========================================

// GetLeaderboardInput is the input DTO for GetLeaderboard use case
type GetLeaderboardInput struct {
	QuizID string `json:"quizId"`
	Limit  int    `json:"limit"`
}

// GetLeaderboardOutput is the output DTO for GetLeaderboard use case
type GetLeaderboardOutput struct {
	QuizID  string                `json:"quizId"`
	Entries []LeaderboardEntryDTO `json:"entries"`
}

// ========================================
// GetQuiz Use Case
// ========================================

// GetQuizInput is the input DTO for GetQuiz use case
type GetQuizInput struct {
	QuizID string `json:"quizId"`
}

// GetQuizOutput is the output DTO for GetQuiz use case
type GetQuizOutput struct {
	Quiz QuizDTO `json:"quiz"`
}

// ========================================
// ListQuizzes Use Case
// ========================================

// ListQuizzesInput is the input DTO for ListQuizzes use case
type ListQuizzesInput struct {
	// Add pagination later: Offset, Limit
}

// ListQuizzesOutput is the output DTO for ListQuizzes use case
type ListQuizzesOutput struct {
	Quizzes []QuizDTO `json:"quizzes"`
}

// ========================================
// GetQuizDetails Use Case
// ========================================

// GetQuizDetailsInput is the input DTO for GetQuizDetails use case
type GetQuizDetailsInput struct {
	QuizID string `json:"quizId"`
}

// GetQuizDetailsOutput is the output DTO for GetQuizDetails use case
// Includes questions with answers (but not which answer is correct!)
type GetQuizDetailsOutput struct {
	Quiz      QuizDetailDTO         `json:"quiz"`
	TopScores []LeaderboardEntryDTO `json:"topScores"`
}
