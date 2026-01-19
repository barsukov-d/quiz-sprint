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
	Session        SessionDTO  `json:"session" validate:"required"`
	CurrentQuestion QuestionDTO `json:"currentQuestion" validate:"required"`
	TotalQuestions int         `json:"totalQuestions" validate:"required"`
	TimeLimit      int         `json:"timeLimit" validate:"required"`
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
