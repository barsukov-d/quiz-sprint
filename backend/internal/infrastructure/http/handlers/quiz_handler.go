package handlers

import (
	"github.com/gofiber/fiber/v2"

	appQuiz "github.com/barsukov/quiz-sprint/backend/internal/application/quiz"
	domainQuiz "github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
)

// QuizHandler handles HTTP requests for quizzes
// NOTE: This is a THIN adapter - no business logic here!
type QuizHandler struct {
	listQuizzesUC      *appQuiz.ListQuizzesUseCase
	getQuizUC          *appQuiz.GetQuizUseCase
	getQuizDetailsUC   *appQuiz.GetQuizDetailsUseCase
	startQuizUC        *appQuiz.StartQuizUseCase
	submitAnswerUC     *appQuiz.SubmitAnswerUseCase
	getLeaderboardUC   *appQuiz.GetLeaderboardUseCase
}

// NewQuizHandler creates a new QuizHandler
func NewQuizHandler(
	listQuizzesUC *appQuiz.ListQuizzesUseCase,
	getQuizUC *appQuiz.GetQuizUseCase,
	getQuizDetailsUC *appQuiz.GetQuizDetailsUseCase,
	startQuizUC *appQuiz.StartQuizUseCase,
	submitAnswerUC *appQuiz.SubmitAnswerUseCase,
	getLeaderboardUC *appQuiz.GetLeaderboardUseCase,
) *QuizHandler {
	return &QuizHandler{
		listQuizzesUC:    listQuizzesUC,
		getQuizUC:        getQuizUC,
		getQuizDetailsUC: getQuizDetailsUC,
		startQuizUC:      startQuizUC,
		submitAnswerUC:   submitAnswerUC,
		getLeaderboardUC: getLeaderboardUC,
	}
}

// ========================================
// HTTP Request DTOs
// ========================================

// StartQuizRequest is the HTTP request body for starting a quiz
type StartQuizRequest struct {
	UserID string `json:"userId"`
}

// SubmitAnswerRequest is the HTTP request body for submitting an answer
type SubmitAnswerRequest struct {
	QuestionID string `json:"questionId"`
	AnswerID   string `json:"answerId"`
	UserID     string `json:"userId"`
}

// ========================================
// Handlers (Thin Adapters)
// ========================================

// GetAllQuizzes handles GET /api/v1/quiz
func (h *QuizHandler) GetAllQuizzes(c *fiber.Ctx) error {
	// 1. Execute use case
	output, err := h.listQuizzesUC.Execute(appQuiz.ListQuizzesInput{})
	if err != nil {
		return mapError(err)
	}

	// 2. Return response
	return c.JSON(fiber.Map{
		"data": output.Quizzes,
	})
}

// GetQuizByID handles GET /api/v1/quiz/:id
// Returns quiz details with questions (but not correct answers) and top scores
func (h *QuizHandler) GetQuizByID(c *fiber.Ctx) error {
	// 1. Extract path parameter
	quizID := c.Params("id")

	// 2. Execute use case
	output, err := h.getQuizDetailsUC.Execute(appQuiz.GetQuizDetailsInput{
		QuizID: quizID,
	})
	if err != nil {
		return mapError(err)
	}

	// 3. Return response
	return c.JSON(fiber.Map{
		"data": output,
	})
}

// StartQuiz handles POST /api/v1/quiz/:id/start
func (h *QuizHandler) StartQuiz(c *fiber.Ctx) error {
	// 1. Parse request body
	var req StartQuizRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// 2. Validate required fields
	if req.UserID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "userId is required")
	}

	// 3. Execute use case
	output, err := h.startQuizUC.Execute(appQuiz.StartQuizInput{
		QuizID: c.Params("id"),
		UserID: req.UserID,
	})
	if err != nil {
		return mapError(err)
	}

	// 4. Return response
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": output,
	})
}

// SubmitAnswer handles POST /api/v1/quiz/session/:sessionId/answer
func (h *QuizHandler) SubmitAnswer(c *fiber.Ctx) error {
	// 1. Parse request body
	var req SubmitAnswerRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// 2. Validate required fields
	if req.QuestionID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "questionId is required")
	}
	if req.AnswerID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "answerId is required")
	}
	if req.UserID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "userId is required")
	}

	// 3. Execute use case
	output, err := h.submitAnswerUC.Execute(appQuiz.SubmitAnswerInput{
		SessionID:  c.Params("sessionId"),
		QuestionID: req.QuestionID,
		AnswerID:   req.AnswerID,
		UserID:     req.UserID,
	})
	if err != nil {
		return mapError(err)
	}

	// 4. Return response
	return c.JSON(fiber.Map{
		"data": output,
	})
}

// GetLeaderboard handles GET /api/v1/quiz/:id/leaderboard
func (h *QuizHandler) GetLeaderboard(c *fiber.Ctx) error {
	// 1. Extract query parameters
	limit := c.QueryInt("limit", 10)

	// 2. Execute use case
	output, err := h.getLeaderboardUC.Execute(appQuiz.GetLeaderboardInput{
		QuizID: c.Params("id"),
		Limit:  limit,
	})
	if err != nil {
		return mapError(err)
	}

	// 3. Return response
	return c.JSON(fiber.Map{
		"data": output.Entries,
	})
}

// ========================================
// Error Mapping (Domain → HTTP)
// ========================================

func mapError(err error) error {
	switch err {
	// Not Found errors
	case domainQuiz.ErrQuizNotFound:
		return fiber.NewError(fiber.StatusNotFound, "Quiz not found")
	case domainQuiz.ErrSessionNotFound:
		return fiber.NewError(fiber.StatusNotFound, "Session not found")
	case domainQuiz.ErrQuestionNotFound:
		return fiber.NewError(fiber.StatusNotFound, "Question not found")
	case domainQuiz.ErrAnswerNotFound:
		return fiber.NewError(fiber.StatusNotFound, "Answer not found")

	// Bad Request errors (validation)
	case domainQuiz.ErrInvalidQuizID,
		domainQuiz.ErrInvalidSessionID,
		domainQuiz.ErrInvalidQuestionID,
		domainQuiz.ErrInvalidAnswerID,
		domainQuiz.ErrInvalidTitle,
		domainQuiz.ErrInvalidTimeLimit,
		domainQuiz.ErrInvalidPassingScore:
		return fiber.NewError(fiber.StatusBadRequest, err.Error())

	case shared.ErrInvalidUserID:
		return fiber.NewError(fiber.StatusBadRequest, "Invalid user ID")

	// Conflict errors
	case domainQuiz.ErrSessionAlreadyExists:
		return fiber.NewError(fiber.StatusConflict, "Active session already exists")
	case domainQuiz.ErrAlreadyAnswered:
		return fiber.NewError(fiber.StatusConflict, "Question already answered")

	// Authorization errors
	case domainQuiz.ErrUnauthorized:
		return fiber.NewError(fiber.StatusUnauthorized, "Unauthorized")

	// Business rule errors
	case domainQuiz.ErrQuizCannotStart,
		domainQuiz.ErrNoQuestions:
		return fiber.NewError(fiber.StatusBadRequest, err.Error())

	case domainQuiz.ErrSessionCompleted:
		return fiber.NewError(fiber.StatusBadRequest, "Session already completed")

	// Default: Internal Server Error
	default:
		return fiber.NewError(fiber.StatusInternalServerError, "Internal server error")
	}
}
