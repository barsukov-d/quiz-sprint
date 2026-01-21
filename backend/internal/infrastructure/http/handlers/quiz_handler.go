package handlers

import (
	"github.com/gofiber/fiber/v3"

	appQuiz "github.com/barsukov/quiz-sprint/backend/internal/application/quiz"
	domainQuiz "github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
)

// QuizHandler handles HTTP requests for quizzes
// NOTE: This is a THIN adapter - no business logic here!
type QuizHandler struct {
	listQuizzesUC             *appQuiz.ListQuizzesUseCase
	getQuizUC                 *appQuiz.GetQuizUseCase
	getQuizDetailsUC          *appQuiz.GetQuizDetailsUseCase
	startQuizUC               *appQuiz.StartQuizUseCase
	submitAnswerUC            *appQuiz.SubmitAnswerUseCase
	getLeaderboardUC          *appQuiz.GetLeaderboardUseCase
	getGlobalLeaderboardUC    *appQuiz.GetGlobalLeaderboardUseCase
	getActiveSessionUC        *appQuiz.GetActiveSessionUseCase
	abandonSessionUC          *appQuiz.AbandonSessionUseCase
	getSessionResultsUC       *appQuiz.GetSessionResultsUseCase
}

// NewQuizHandler creates a new QuizHandler
func NewQuizHandler(
	listQuizzesUC *appQuiz.ListQuizzesUseCase,
	getQuizUC *appQuiz.GetQuizUseCase,
	getQuizDetailsUC *appQuiz.GetQuizDetailsUseCase,
	startQuizUC *appQuiz.StartQuizUseCase,
	submitAnswerUC *appQuiz.SubmitAnswerUseCase,
	getLeaderboardUC *appQuiz.GetLeaderboardUseCase,
	getGlobalLeaderboardUC *appQuiz.GetGlobalLeaderboardUseCase,
	getActiveSessionUC *appQuiz.GetActiveSessionUseCase,
	abandonSessionUC *appQuiz.AbandonSessionUseCase,
	getSessionResultsUC *appQuiz.GetSessionResultsUseCase,
) *QuizHandler {
	return &QuizHandler{
		listQuizzesUC:          listQuizzesUC,
		getQuizUC:              getQuizUC,
		getQuizDetailsUC:       getQuizDetailsUC,
		startQuizUC:            startQuizUC,
		submitAnswerUC:         submitAnswerUC,
		getLeaderboardUC:       getLeaderboardUC,
		getGlobalLeaderboardUC: getGlobalLeaderboardUC,
		getActiveSessionUC:     getActiveSessionUC,
		abandonSessionUC:       abandonSessionUC,
		getSessionResultsUC:    getSessionResultsUC,
	}
}

// ========================================
// Handlers (Thin Adapters)
// ========================================
// Note: Request DTOs moved to swagger_models.go

// GetAllQuizzes handles GET /api/v1/quiz
// @Summary List all quizzes
// @Description Get a list of all available quizzes with basic information, optionally filtered by category.
// @Tags quiz
// @Accept json
// @Produce json
// @Param categoryId query string false "Category ID to filter quizzes by"
// @Success 200 {object} ListQuizzesResponse "List of quizzes"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /quiz [get]
func (h *QuizHandler) GetAllQuizzes(c fiber.Ctx) error {
	// 1. Check for categoryId query param
	categoryID := c.Query("categoryId")

	// 2. Execute use case
	input := appQuiz.ListQuizzesInput{
		CategoryID: categoryID,
	}
	output, err := h.listQuizzesUC.Execute(input)
	if err != nil {
		return mapError(err)
	}

	// 3. Return response
	return c.JSON(fiber.Map{
		"data": output.Quizzes,
	})
}

// GetQuizByID handles GET /api/v1/quiz/:id
// Returns quiz details with questions (but not correct answers) and top scores
// @Summary Get quiz details
// @Description Get detailed quiz information including questions, answers (without correct answer indicators), and top 3 leaderboard scores
// @Tags quiz
// @Accept json
// @Produce json
// @Param id path string true "Quiz ID"
// @Success 200 {object} GetQuizDetailsResponse "Quiz details with questions and top scores"
// @Failure 400 {object} ErrorResponse "Invalid quiz ID"
// @Failure 404 {object} ErrorResponse "Quiz not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /quiz/{id} [get]
func (h *QuizHandler) GetQuizByID(c fiber.Ctx) error {
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
// @Summary Start a quiz session
// @Description Create a new quiz session for a user and return the first question
// @Tags quiz
// @Accept json
// @Produce json
// @Param id path string true "Quiz ID"
// @Param request body StartQuizRequest true "Start quiz request"
// @Success 201 {object} StartQuizResponse "Quiz session started with first question"
// @Failure 400 {object} ErrorResponse "Invalid request or quiz ID"
// @Failure 404 {object} ErrorResponse "Quiz not found"
// @Failure 409 {object} ErrorResponse "Active session already exists"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /quiz/{id}/start [post]
func (h *QuizHandler) StartQuiz(c fiber.Ctx) error {
	// 1. Parse request body
	var req StartQuizRequest
	if err := c.Bind().Body(&req); err != nil {
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
// @Summary Submit an answer
// @Description Submit an answer for a question in an active quiz session
// @Tags quiz
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body SubmitAnswerRequest true "Submit answer request"
// @Success 200 {object} SubmitAnswerResponse "Answer result with correctness, points, and next question or final result"
// @Failure 400 {object} ErrorResponse "Invalid request, session completed, or already answered"
// @Failure 401 {object} ErrorResponse "Unauthorized - user mismatch"
// @Failure 404 {object} ErrorResponse "Session, question, or answer not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /quiz/session/{sessionId}/answer [post]
func (h *QuizHandler) SubmitAnswer(c fiber.Ctx) error {
	// 1. Parse request body
	var req SubmitAnswerRequest
	if err := c.Bind().Body(&req); err != nil {
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
	if req.TimeTaken < 0 {
		return fiber.NewError(fiber.StatusBadRequest, "timeTaken must be non-negative")
	}

	// 3. Execute use case
	output, err := h.submitAnswerUC.Execute(appQuiz.SubmitAnswerInput{
		SessionID:  c.Params("sessionId"),
		QuestionID: req.QuestionID,
		AnswerID:   req.AnswerID,
		UserID:     req.UserID,
		TimeTaken:  req.TimeTaken,
	})
	if err != nil {
		return mapError(err)
	}

	// 4. Return response
	return c.JSON(fiber.Map{
		"data": output,
	})
}

// GetActiveSession handles GET /api/v1/quiz/:id/active-session
// @Summary Get active quiz session
// @Description Retrieve the active quiz session for the current user and quiz
// @Tags quiz
// @Accept json
// @Produce json
// @Param id path string true "Quiz ID"
// @Param request body GetActiveSessionRequest true "User ID for authorization"
// @Success 200 {object} GetActiveSessionResponse "Active session with current question"
// @Failure 400 {object} ErrorResponse "Invalid quiz or user ID"
// @Failure 404 {object} ErrorResponse "No active session found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /quiz/{id}/active-session [get]
func (h *QuizHandler) GetActiveSession(c fiber.Ctx) error {
	// 1. Parse query parameter
	userID := c.Query("userId")
	if userID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "userId query parameter is required")
	}

	// 2. Execute use case
	output, err := h.getActiveSessionUC.Execute(appQuiz.GetActiveSessionInput{
		QuizID: c.Params("id"),
		UserID: userID,
	})
	if err != nil {
		return mapError(err)
	}

	// 3. Return response
	return c.JSON(fiber.Map{
		"data": output,
	})
}

// AbandonSession handles DELETE /api/v1/quiz/session/:sessionId
// @Summary Abandon a quiz session
// @Description Delete an active quiz session, allowing the user to start fresh
// @Tags quiz
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body AbandonSessionRequest true "User ID for authorization"
// @Success 204 "Session abandoned successfully"
// @Failure 400 {object} ErrorResponse "Invalid session ID"
// @Failure 401 {object} ErrorResponse "Unauthorized - session belongs to another user"
// @Failure 404 {object} ErrorResponse "Session not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /quiz/session/{sessionId} [delete]
func (h *QuizHandler) AbandonSession(c fiber.Ctx) error {
	// 1. Parse request body
	var req AbandonSessionRequest
	if err := c.Bind().Body(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// 2. Validate required fields
	if req.UserID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "userId is required")
	}

	// 3. Execute use case
	err := h.abandonSessionUC.Execute(appQuiz.AbandonSessionInput{
		SessionID: c.Params("sessionId"),
		UserID:    req.UserID,
	})
	if err != nil {
		return mapError(err)
	}

	// 4. Return no content (204)
	return c.SendStatus(fiber.StatusNoContent)
}

// GetSessionResults handles GET /api/v1/quiz/session/:sessionId
// @Summary Get session results
// @Description Get detailed results for a quiz session including statistics and quiz info
// @Tags quiz
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} GetSessionResultsResponse "Session results with statistics"
// @Failure 400 {object} ErrorResponse "Invalid session ID"
// @Failure 404 {object} ErrorResponse "Session not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /quiz/session/{sessionId} [get]
func (h *QuizHandler) GetSessionResults(c fiber.Ctx) error {
	// 1. Execute use case
	output, err := h.getSessionResultsUC.Execute(appQuiz.GetSessionResultsInput{
		SessionID: c.Params("sessionId"),
	})
	if err != nil {
		return mapError(err)
	}

	// 2. Map to response DTO
	response := SessionResultsData{
		Session: SessionDTO{
			ID:              output.Session.ID,
			QuizID:          output.Session.QuizID,
			UserID:          output.Session.UserID,
			CurrentQuestion: output.Session.CurrentQuestion,
			Score:           output.Session.Score,
			StartedAt:       output.Session.StartedAt,
			CompletedAt:     output.Session.CompletedAt,
			Status:          output.Session.Status,
		},
		Quiz: QuizDTO{
			ID:           output.Quiz.ID,
			Title:        output.Quiz.Title,
			Description:  output.Quiz.Description,
			TimeLimit:    output.Quiz.TimeLimit,
			PassingScore: output.Quiz.PassingScore,
		},
		TotalQuestions:  output.TotalQuestions,
		CorrectAnswers:  output.CorrectAnswers,
		TimeSpent:       output.TimeSpent,
		Passed:          output.Passed,
		ScorePercentage: output.ScorePercentage,
	}

	// 3. Return wrapped response
	return c.JSON(fiber.Map{
		"data": response,
	})
}

// GetLeaderboard handles GET /api/v1/quiz/:id/leaderboard
// @Summary Get quiz leaderboard
// @Description Get the leaderboard for a specific quiz with top scores
// @Tags quiz
// @Accept json
// @Produce json
// @Param id path string true "Quiz ID"
// @Param limit query int false "Number of entries to return (default 10)"
// @Success 200 {object} GetLeaderboardResponse "Leaderboard entries"
// @Failure 400 {object} ErrorResponse "Invalid quiz ID"
// @Failure 404 {object} ErrorResponse "Quiz not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /quiz/{id}/leaderboard [get]
func (h *QuizHandler) GetLeaderboard(c fiber.Ctx) error {
	// 1. Extract query parameters
	limit := fiber.Query[int](c, "limit", 10)

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

// GetGlobalLeaderboard handles GET /api/v1/leaderboard
// @Summary Get global leaderboard
// @Description Get the global leaderboard across all quizzes (sum of best scores per quiz)
// @Tags leaderboard
// @Accept json
// @Produce json
// @Param limit query int false "Number of entries to return (default 10, max 100)"
// @Success 200 {object} GetGlobalLeaderboardResponse "Global leaderboard entries"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /leaderboard [get]
func (h *QuizHandler) GetGlobalLeaderboard(c fiber.Ctx) error {
	// 1. Extract query parameters
	limit := fiber.Query[int](c, "limit", 10)

	// 2. Execute use case
	output, err := h.getGlobalLeaderboardUC.Execute(appQuiz.GetGlobalLeaderboardInput{
		Limit: limit,
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
	case domainQuiz.ErrCategoryNotFound:
		return fiber.NewError(fiber.StatusNotFound, "Category not found")

	// Bad Request errors (validation)
	case domainQuiz.ErrInvalidQuizID,
		domainQuiz.ErrInvalidSessionID,
		domainQuiz.ErrInvalidQuestionID,
		domainQuiz.ErrInvalidAnswerID,
		domainQuiz.ErrInvalidTitle,
		domainQuiz.ErrInvalidTimeLimit,
		domainQuiz.ErrInvalidPassingScore,
		domainQuiz.ErrInvalidCategoryName,
		domainQuiz.ErrCategoryNameTooLong:
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
