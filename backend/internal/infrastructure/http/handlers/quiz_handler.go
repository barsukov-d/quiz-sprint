package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	appQuiz "github.com/barsukov/quiz-sprint/backend/internal/application/quiz"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
)

// QuizHandler handles HTTP requests for quizzes
type QuizHandler struct {
	repo               quiz.QuizRepository
	startQuizUseCase   *appQuiz.StartQuizUseCase
	submitAnswerUseCase *appQuiz.SubmitAnswerUseCase
	getLeaderboardUseCase *appQuiz.GetLeaderboardUseCase
}

// NewQuizHandler creates a new QuizHandler
func NewQuizHandler(repo quiz.QuizRepository) *QuizHandler {
	return &QuizHandler{
		repo:               repo,
		startQuizUseCase:   appQuiz.NewStartQuizUseCase(repo),
		submitAnswerUseCase: appQuiz.NewSubmitAnswerUseCase(repo),
		getLeaderboardUseCase: appQuiz.NewGetLeaderboardUseCase(repo),
	}
}

// GetAllQuizzes godoc
// @Summary Get all quizzes
// @Description Get a list of all available quizzes
// @Tags quiz
// @Produce json
// @Success 200 {array} quiz.Quiz
// @Router /api/quiz [get]
func (h *QuizHandler) GetAllQuizzes(c *fiber.Ctx) error {
	quizzes, err := h.repo.FindAll(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to fetch quizzes")
	}

	return c.JSON(fiber.Map{
		"data": quizzes,
	})
}

// GetQuizByID godoc
// @Summary Get quiz by ID
// @Description Get a specific quiz by its ID
// @Tags quiz
// @Produce json
// @Param id path string true "Quiz ID"
// @Success 200 {object} quiz.Quiz
// @Router /api/quiz/{id} [get]
func (h *QuizHandler) GetQuizByID(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid quiz ID")
	}

	quizData, err := h.repo.FindByID(c.Context(), id)
	if err != nil {
		if err == quiz.ErrQuizNotFound {
			return fiber.NewError(fiber.StatusNotFound, "Quiz not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to fetch quiz")
	}

	return c.JSON(fiber.Map{
		"data": quizData,
	})
}

// StartQuizRequest represents the request body for starting a quiz
type StartQuizRequest struct {
	UserID string `json:"userId"`
}

// StartQuiz godoc
// @Summary Start a quiz
// @Description Start a new quiz session
// @Tags quiz
// @Accept json
// @Produce json
// @Param id path string true "Quiz ID"
// @Param body body StartQuizRequest true "Start Quiz Request"
// @Success 201 {object} appQuiz.StartQuizResult
// @Router /api/quiz/{id}/start [post]
func (h *QuizHandler) StartQuiz(c *fiber.Ctx) error {
	idParam := c.Params("id")
	quizID, err := uuid.Parse(idParam)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid quiz ID")
	}

	var req StartQuizRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if req.UserID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "User ID is required")
	}

	result, err := h.startQuizUseCase.Execute(c.Context(), appQuiz.StartQuizCommand{
		QuizID: quizID,
		UserID: req.UserID,
	})

	if err != nil {
		switch err {
		case quiz.ErrQuizNotFound:
			return fiber.NewError(fiber.StatusNotFound, "Quiz not found")
		case quiz.ErrQuizCannotStart:
			return fiber.NewError(fiber.StatusBadRequest, "Quiz cannot be started")
		case quiz.ErrSessionAlreadyExists:
			return fiber.NewError(fiber.StatusConflict, "Active session already exists")
		default:
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to start quiz")
		}
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": result,
	})
}

// SubmitAnswerRequest represents the request body for submitting an answer
type SubmitAnswerRequest struct {
	QuestionID string `json:"questionId"`
	AnswerID   string `json:"answerId"`
	UserID     string `json:"userId"`
}

// SubmitAnswer godoc
// @Summary Submit an answer
// @Description Submit an answer for a quiz question
// @Tags quiz
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param body body SubmitAnswerRequest true "Submit Answer Request"
// @Success 200 {object} appQuiz.SubmitAnswerResult
// @Router /api/quiz/session/{sessionId}/answer [post]
func (h *QuizHandler) SubmitAnswer(c *fiber.Ctx) error {
	sessionIDParam := c.Params("sessionId")
	sessionID, err := uuid.Parse(sessionIDParam)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid session ID")
	}

	var req SubmitAnswerRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	questionID, err := uuid.Parse(req.QuestionID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid question ID")
	}

	answerID, err := uuid.Parse(req.AnswerID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid answer ID")
	}

	result, err := h.submitAnswerUseCase.Execute(c.Context(), appQuiz.SubmitAnswerCommand{
		SessionID:  sessionID,
		QuestionID: questionID,
		AnswerID:   answerID,
		UserID:     req.UserID,
	})

	if err != nil {
		switch err {
		case quiz.ErrSessionNotFound:
			return fiber.NewError(fiber.StatusNotFound, "Session not found")
		case quiz.ErrUnauthorized:
			return fiber.NewError(fiber.StatusUnauthorized, "Unauthorized")
		case quiz.ErrSessionCompleted:
			return fiber.NewError(fiber.StatusBadRequest, "Session already completed")
		case quiz.ErrQuestionNotFound:
			return fiber.NewError(fiber.StatusNotFound, "Question not found")
		case quiz.ErrAnswerNotFound:
			return fiber.NewError(fiber.StatusNotFound, "Answer not found")
		case quiz.ErrAlreadyAnswered:
			return fiber.NewError(fiber.StatusConflict, "Question already answered")
		default:
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to submit answer")
		}
	}

	return c.JSON(fiber.Map{
		"data": result,
	})
}

// GetLeaderboard godoc
// @Summary Get leaderboard
// @Description Get the leaderboard for a quiz
// @Tags quiz
// @Produce json
// @Param id path string true "Quiz ID"
// @Param limit query int false "Limit" default(10)
// @Success 200 {array} quiz.LeaderboardEntry
// @Router /api/quiz/{id}/leaderboard [get]
func (h *QuizHandler) GetLeaderboard(c *fiber.Ctx) error {
	idParam := c.Params("id")
	quizID, err := uuid.Parse(idParam)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid quiz ID")
	}

	limit := c.QueryInt("limit", 10)

	entries, err := h.getLeaderboardUseCase.Execute(c.Context(), appQuiz.GetLeaderboardQuery{
		QuizID: quizID,
		Limit:  limit,
	})

	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to fetch leaderboard")
	}

	return c.JSON(fiber.Map{
		"data": entries,
	})
}
