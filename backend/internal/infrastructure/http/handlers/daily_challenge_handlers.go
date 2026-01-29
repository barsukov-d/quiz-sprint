package handlers

import (
	"github.com/gofiber/fiber/v3"

	appDaily "github.com/barsukov/quiz-sprint/backend/internal/application/daily_challenge"
	domainDaily "github.com/barsukov/quiz-sprint/backend/internal/domain/daily_challenge"
	domainQuiz "github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
)

// DailyChallengeHandler handles HTTP requests for Daily Challenge game mode
type DailyChallengeHandler struct {
	getOrCreateQuizUC *appDaily.GetOrCreateDailyQuizUseCase
	startChallengeUC  *appDaily.StartDailyChallengeUseCase
	submitAnswerUC    *appDaily.SubmitDailyAnswerUseCase
	getStatusUC       *appDaily.GetDailyGameStatusUseCase
	getLeaderboardUC  *appDaily.GetDailyLeaderboardUseCase
	getStreakUC       *appDaily.GetPlayerStreakUseCase
	openChestUC       *appDaily.OpenChestUseCase
	retryUC           *appDaily.RetryChallengeUseCase
}

func NewDailyChallengeHandler(
	getOrCreateQuizUC *appDaily.GetOrCreateDailyQuizUseCase,
	startChallengeUC *appDaily.StartDailyChallengeUseCase,
	submitAnswerUC *appDaily.SubmitDailyAnswerUseCase,
	getStatusUC *appDaily.GetDailyGameStatusUseCase,
	getLeaderboardUC *appDaily.GetDailyLeaderboardUseCase,
	getStreakUC *appDaily.GetPlayerStreakUseCase,
	openChestUC *appDaily.OpenChestUseCase,
	retryUC *appDaily.RetryChallengeUseCase,
) *DailyChallengeHandler {
	return &DailyChallengeHandler{
		getOrCreateQuizUC: getOrCreateQuizUC,
		startChallengeUC:  startChallengeUC,
		submitAnswerUC:    submitAnswerUC,
		getStatusUC:       getStatusUC,
		getLeaderboardUC:  getLeaderboardUC,
		getStreakUC:       getStreakUC,
		openChestUC:       openChestUC,
		retryUC:           retryUC,
	}
}

// StartDailyChallenge handles POST /api/v1/daily-challenge/start
// @Summary Start daily challenge
// @Description Start today's daily challenge (one attempt per day)
// @Tags daily-challenge
// @Accept json
// @Produce json
// @Param request body StartDailyChallengeRequest true "Start request"
// @Success 201 {object} StartDailyChallengeResponse "Challenge started"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 409 {object} ErrorResponse "Already played today"
// @Failure 500 {object} ErrorResponse "Internal error"
// @Router /daily-challenge/start [post]
func (h *DailyChallengeHandler) StartDailyChallenge(c fiber.Ctx) error {
	var req StartDailyChallengeRequest
	if err := c.Bind().Body(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if req.PlayerID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "playerId is required")
	}

	output, err := h.startChallengeUC.Execute(appDaily.StartDailyChallengeInput{
		PlayerID: req.PlayerID,
		Date:     req.Date,
	})
	if err != nil {
		return mapDailyChallengeError(err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": output})
}

// SubmitDailyAnswer handles POST /api/v1/daily-challenge/:gameId/answer
// @Summary Submit answer
// @Description Submit answer for current question (no immediate feedback)
// @Tags daily-challenge
// @Accept json
// @Produce json
// @Param gameId path string true "Game ID"
// @Param request body SubmitDailyAnswerRequest true "Submit request"
// @Success 200 {object} SubmitDailyAnswerResponse "Answer submitted"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 404 {object} ErrorResponse "Game not found"
// @Failure 500 {object} ErrorResponse "Internal error"
// @Router /daily-challenge/{gameId}/answer [post]
func (h *DailyChallengeHandler) SubmitDailyAnswer(c fiber.Ctx) error {
	var req SubmitDailyAnswerRequest
	if err := c.Bind().Body(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if req.QuestionID == "" || req.AnswerID == "" || req.PlayerID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Missing required fields")
	}

	output, err := h.submitAnswerUC.Execute(appDaily.SubmitDailyAnswerInput{
		GameID:     c.Params("gameId"),
		QuestionID: req.QuestionID,
		AnswerID:   req.AnswerID,
		PlayerID:   req.PlayerID,
		TimeTaken:  req.TimeTaken,
	})
	if err != nil {
		return mapDailyChallengeError(err)
	}

	return c.JSON(fiber.Map{"data": output})
}

// GetDailyStatus handles GET /api/v1/daily-challenge/status
// @Summary Get daily status
// @Description Get player's daily challenge status for today
// @Tags daily-challenge
// @Accept json
// @Produce json
// @Param playerId query string true "Player ID"
// @Param date query string false "Date (YYYY-MM-DD, defaults to today)"
// @Success 200 {object} GetDailyStatusResponse "Status"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 500 {object} ErrorResponse "Internal error"
// @Router /daily-challenge/status [get]
func (h *DailyChallengeHandler) GetDailyStatus(c fiber.Ctx) error {
	playerID := c.Query("playerId")
	if playerID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "playerId is required")
	}

	output, err := h.getStatusUC.Execute(appDaily.GetDailyGameStatusInput{
		PlayerID: playerID,
		Date:     c.Query("date"),
	})
	if err != nil {
		return mapDailyChallengeError(err)
	}

	return c.JSON(fiber.Map{"data": output})
}

// GetDailyLeaderboard handles GET /api/v1/daily-challenge/leaderboard
// @Summary Get leaderboard
// @Description Get daily challenge leaderboard for a specific date
// @Tags daily-challenge
// @Accept json
// @Produce json
// @Param date query string false "Date (YYYY-MM-DD, defaults to today)"
// @Param limit query int false "Limit (default 10, max 100)"
// @Param playerId query string false "Player ID (to get rank)"
// @Success 200 {object} GetDailyLeaderboardResponse "Leaderboard"
// @Failure 500 {object} ErrorResponse "Internal error"
// @Router /daily-challenge/leaderboard [get]
func (h *DailyChallengeHandler) GetDailyLeaderboard(c fiber.Ctx) error {
	limit := fiber.Query[int](c, "limit", 10)
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	output, err := h.getLeaderboardUC.Execute(appDaily.GetDailyLeaderboardInput{
		Date:     c.Query("date"),
		Limit:    limit,
		PlayerID: c.Query("playerId"),
	})
	if err != nil {
		return mapDailyChallengeError(err)
	}

	return c.JSON(fiber.Map{"data": output})
}

// GetPlayerStreak handles GET /api/v1/daily-challenge/streak
// @Summary Get player streak
// @Description Get player's daily streak information
// @Tags daily-challenge
// @Accept json
// @Produce json
// @Param playerId query string true "Player ID"
// @Success 200 {object} GetPlayerStreakResponse "Streak info"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 500 {object} ErrorResponse "Internal error"
// @Router /daily-challenge/streak [get]
func (h *DailyChallengeHandler) GetPlayerStreak(c fiber.Ctx) error {
	playerID := c.Query("playerId")
	if playerID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "playerId is required")
	}

	output, err := h.getStreakUC.Execute(appDaily.GetPlayerStreakInput{
		PlayerID: playerID,
	})
	if err != nil {
		return mapDailyChallengeError(err)
	}

	return c.JSON(fiber.Map{"data": output})
}

// OpenChest handles POST /api/v1/daily-challenge/:gameId/chest/open
// @Summary Open chest
// @Description Get chest rewards (idempotent)
// @Tags daily-challenge
// @Accept json
// @Produce json
// @Param gameId path string true "Game ID"
// @Param request body OpenChestRequest true "Open chest request"
// @Success 200 {object} OpenChestResponse "Chest opened"
// @Failure 400 {object} ErrorResponse "Game not completed"
// @Failure 404 {object} ErrorResponse "Game not found"
// @Failure 500 {object} ErrorResponse "Internal error"
// @Router /daily-challenge/{gameId}/chest/open [post]
func (h *DailyChallengeHandler) OpenChest(c fiber.Ctx) error {
	var req OpenChestRequest
	if err := c.Bind().Body(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if req.PlayerID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "playerId is required")
	}

	output, err := h.openChestUC.Execute(appDaily.OpenChestInput{
		GameID:   c.Params("gameId"),
		PlayerID: req.PlayerID,
	})
	if err != nil {
		return mapDailyChallengeError(err)
	}

	return c.JSON(fiber.Map{"data": output})
}

// RetryChallenge handles POST /api/v1/daily-challenge/:gameId/retry
// @Summary Retry challenge
// @Description Create second attempt (costs 100 coins or ad)
// @Tags daily-challenge
// @Accept json
// @Produce json
// @Param gameId path string true "Original Game ID"
// @Param request body RetryChallengeRequest true "Retry request"
// @Success 201 {object} RetryChallengeResponse "Retry started"
// @Failure 400 {object} ErrorResponse "Invalid request or insufficient coins"
// @Failure 404 {object} ErrorResponse "Game not found"
// @Failure 409 {object} ErrorResponse "Retry limit reached"
// @Failure 500 {object} ErrorResponse "Internal error"
// @Router /daily-challenge/{gameId}/retry [post]
func (h *DailyChallengeHandler) RetryChallenge(c fiber.Ctx) error {
	var req RetryChallengeRequest
	if err := c.Bind().Body(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if req.PlayerID == "" || req.PaymentMethod == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Missing required fields")
	}

	if req.PaymentMethod != "coins" && req.PaymentMethod != "ad" {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid payment method")
	}

	output, err := h.retryUC.Execute(appDaily.RetryChallengeInput{
		GameID:        c.Params("gameId"),
		PlayerID:      req.PlayerID,
		PaymentMethod: req.PaymentMethod,
	})
	if err != nil {
		return mapDailyChallengeError(err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": output})
}

// ========================================
// Error Mapping
// ========================================

func mapDailyChallengeError(err error) error {
	switch err {
	case domainDaily.ErrDailyQuizNotFound:
		return fiber.NewError(fiber.StatusNotFound, "Daily quiz not found")
	case domainDaily.ErrGameNotFound:
		return fiber.NewError(fiber.StatusNotFound, "Game not found")
	case domainDaily.ErrAlreadyPlayedToday:
		return fiber.NewError(fiber.StatusConflict, "Already played today")
	case domainDaily.ErrGameAlreadyCompleted:
		return fiber.NewError(fiber.StatusBadRequest, "Game already completed")
	case domainDaily.ErrGameNotActive:
		return fiber.NewError(fiber.StatusBadRequest, "Game not active")
	case domainQuiz.ErrQuestionNotFound:
		return fiber.NewError(fiber.StatusNotFound, "Question not found")
	case domainQuiz.ErrAnswerNotFound:
		return fiber.NewError(fiber.StatusNotFound, "Answer not found")
	default:
		return fiber.NewError(fiber.StatusInternalServerError, "Internal server error")
	}
}
