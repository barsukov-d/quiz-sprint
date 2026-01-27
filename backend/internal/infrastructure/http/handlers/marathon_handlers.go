package handlers

import (
	"github.com/gofiber/fiber/v3"

	appMarathon "github.com/barsukov/quiz-sprint/backend/internal/application/marathon"
	domainMarathon "github.com/barsukov/quiz-sprint/backend/internal/domain/solo_marathon"
	domainQuiz "github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
)

// MarathonHandler handles HTTP requests for Solo Marathon game mode
// NOTE: This is a THIN adapter - no business logic here!
type MarathonHandler struct {
	startMarathonUC           *appMarathon.StartMarathonUseCase
	submitAnswerUC            *appMarathon.SubmitMarathonAnswerUseCase
	useHintUC                 *appMarathon.UseMarathonHintUseCase
	abandonMarathonUC         *appMarathon.AbandonMarathonUseCase
	getStatusUC               *appMarathon.GetMarathonStatusUseCase
	getPersonalBestsUC        *appMarathon.GetPersonalBestsUseCase
	getLeaderboardUC          *appMarathon.GetMarathonLeaderboardUseCase
}

// NewMarathonHandler creates a new MarathonHandler
func NewMarathonHandler(
	startMarathonUC *appMarathon.StartMarathonUseCase,
	submitAnswerUC *appMarathon.SubmitMarathonAnswerUseCase,
	useHintUC *appMarathon.UseMarathonHintUseCase,
	abandonMarathonUC *appMarathon.AbandonMarathonUseCase,
	getStatusUC *appMarathon.GetMarathonStatusUseCase,
	getPersonalBestsUC *appMarathon.GetPersonalBestsUseCase,
	getLeaderboardUC *appMarathon.GetMarathonLeaderboardUseCase,
) *MarathonHandler {
	return &MarathonHandler{
		startMarathonUC:    startMarathonUC,
		submitAnswerUC:     submitAnswerUC,
		useHintUC:          useHintUC,
		abandonMarathonUC:  abandonMarathonUC,
		getStatusUC:        getStatusUC,
		getPersonalBestsUC: getPersonalBestsUC,
		getLeaderboardUC:   getLeaderboardUC,
	}
}

// ========================================
// Handlers (Thin Adapters)
// ========================================
// Note: Request/Response DTOs are in swagger_models.go

// StartMarathon handles POST /api/v1/marathon/start
// @Summary Start a marathon game
// @Description Start a new marathon game session with adaptive difficulty and dynamic question loading
// @Tags marathon
// @Accept json
// @Produce json
// @Param request body StartMarathonRequest true "Start marathon request"
// @Success 201 {object} StartMarathonResponse "Marathon game started with first question"
// @Failure 400 {object} ErrorResponse "Invalid request or player ID"
// @Failure 409 {object} ErrorResponse "Active game already exists"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /marathon/start [post]
func (h *MarathonHandler) StartMarathon(c fiber.Ctx) error {
	// 1. Parse request body
	var req StartMarathonRequest
	if err := c.Bind().Body(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// 2. Validate required fields
	if req.PlayerID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "playerId is required")
	}

	// 3. Execute use case
	output, err := h.startMarathonUC.Execute(appMarathon.StartMarathonInput{
		PlayerID:   req.PlayerID,
		CategoryID: req.CategoryID,
	})
	if err != nil {
		return mapMarathonError(err)
	}

	// 4. Return response
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": output,
	})
}

// SubmitMarathonAnswer handles POST /api/v1/marathon/:gameId/answer
// @Summary Submit an answer in marathon mode
// @Description Submit an answer for the current question in a marathon game. Game continues if correct, loses life if incorrect.
// @Tags marathon
// @Accept json
// @Produce json
// @Param gameId path string true "Game ID"
// @Param request body SubmitMarathonAnswerRequest true "Submit answer request"
// @Success 200 {object} SubmitMarathonAnswerResponse "Answer result with next question or game over details"
// @Failure 400 {object} ErrorResponse "Invalid request, game not in progress, or wrong question"
// @Failure 401 {object} ErrorResponse "Unauthorized - game belongs to another player"
// @Failure 404 {object} ErrorResponse "Game, question, or answer not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /marathon/{gameId}/answer [post]
func (h *MarathonHandler) SubmitMarathonAnswer(c fiber.Ctx) error {
	// 1. Parse request body
	var req SubmitMarathonAnswerRequest
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
	if req.PlayerID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "playerId is required")
	}
	if req.TimeTaken < 0 {
		return fiber.NewError(fiber.StatusBadRequest, "timeTaken must be non-negative")
	}

	// 3. Execute use case
	output, err := h.submitAnswerUC.Execute(appMarathon.SubmitMarathonAnswerInput{
		GameID:     c.Params("gameId"),
		QuestionID: req.QuestionID,
		AnswerID:   req.AnswerID,
		PlayerID:   req.PlayerID,
		TimeTaken:  req.TimeTaken,
	})
	if err != nil {
		return mapMarathonError(err)
	}

	// 4. Return response
	return c.JSON(fiber.Map{
		"data": output,
	})
}

// UseMarathonHint handles POST /api/v1/marathon/:gameId/hint
// @Summary Use a hint in marathon mode
// @Description Use a hint to help with the current question. Available hints: fifty_fifty (remove 2 wrong answers), extra_time (+10 seconds), skip (skip question without losing life)
// @Tags marathon
// @Accept json
// @Produce json
// @Param gameId path string true "Game ID"
// @Param request body UseMarathonHintRequest true "Use hint request"
// @Success 200 {object} UseMarathonHintResponse "Hint result with remaining hints"
// @Failure 400 {object} ErrorResponse "Invalid request, hint not available, or wrong question"
// @Failure 401 {object} ErrorResponse "Unauthorized - game belongs to another player"
// @Failure 404 {object} ErrorResponse "Game or question not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /marathon/{gameId}/hint [post]
func (h *MarathonHandler) UseMarathonHint(c fiber.Ctx) error {
	// 1. Parse request body
	var req UseMarathonHintRequest
	if err := c.Bind().Body(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// 2. Validate required fields
	if req.QuestionID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "questionId is required")
	}
	if req.HintType == "" {
		return fiber.NewError(fiber.StatusBadRequest, "hintType is required")
	}
	if req.PlayerID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "playerId is required")
	}

	// Validate hint type
	validHints := map[string]bool{
		"fifty_fifty": true,
		"extra_time":  true,
		"skip":        true,
	}
	if !validHints[req.HintType] {
		return fiber.NewError(fiber.StatusBadRequest, "hintType must be one of: fifty_fifty, extra_time, skip")
	}

	// 3. Execute use case
	output, err := h.useHintUC.Execute(appMarathon.UseMarathonHintInput{
		GameID:     c.Params("gameId"),
		QuestionID: req.QuestionID,
		HintType:   req.HintType,
		PlayerID:   req.PlayerID,
	})
	if err != nil {
		return mapMarathonError(err)
	}

	// 4. Return response
	return c.JSON(fiber.Map{
		"data": output,
	})
}

// AbandonMarathon handles DELETE /api/v1/marathon/:gameId
// @Summary Abandon a marathon game
// @Description Abandon an active marathon game. Game ends immediately and final statistics are returned.
// @Tags marathon
// @Accept json
// @Produce json
// @Param gameId path string true "Game ID"
// @Param request body AbandonMarathonRequest true "Abandon game request"
// @Success 200 {object} AbandonMarathonResponse "Game abandoned with final statistics"
// @Failure 400 {object} ErrorResponse "Invalid game ID or game not in progress"
// @Failure 401 {object} ErrorResponse "Unauthorized - game belongs to another player"
// @Failure 404 {object} ErrorResponse "Game not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /marathon/{gameId} [delete]
func (h *MarathonHandler) AbandonMarathon(c fiber.Ctx) error {
	// 1. Parse request body
	var req AbandonMarathonRequest
	if err := c.Bind().Body(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// 2. Validate required fields
	if req.PlayerID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "playerId is required")
	}

	// 3. Execute use case
	output, err := h.abandonMarathonUC.Execute(appMarathon.AbandonMarathonInput{
		GameID:   c.Params("gameId"),
		PlayerID: req.PlayerID,
	})
	if err != nil {
		return mapMarathonError(err)
	}

	// 4. Return response
	return c.JSON(fiber.Map{
		"data": output.GameOverResult,
	})
}

// GetMarathonStatus handles GET /api/v1/marathon/status
// @Summary Get marathon game status
// @Description Get the status of the player's active marathon game, if any exists
// @Tags marathon
// @Accept json
// @Produce json
// @Param playerId query string true "Player ID"
// @Success 200 {object} GetMarathonStatusResponse "Marathon game status"
// @Failure 400 {object} ErrorResponse "Invalid player ID"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /marathon/status [get]
func (h *MarathonHandler) GetMarathonStatus(c fiber.Ctx) error {
	// 1. Parse query parameter
	playerID := c.Query("playerId")
	if playerID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "playerId query parameter is required")
	}

	// 2. Execute use case
	output, err := h.getStatusUC.Execute(appMarathon.GetMarathonStatusInput{
		PlayerID: playerID,
	})
	if err != nil {
		return mapMarathonError(err)
	}

	// 3. Return response
	return c.JSON(fiber.Map{
		"data": output,
	})
}

// GetPersonalBests handles GET /api/v1/marathon/personal-bests
// @Summary Get personal best records
// @Description Get all personal best records for a player across all marathon categories
// @Tags marathon
// @Accept json
// @Produce json
// @Param playerId query string true "Player ID"
// @Success 200 {object} GetPersonalBestsResponse "Personal best records"
// @Failure 400 {object} ErrorResponse "Invalid player ID"
// @Failure 404 {object} ErrorResponse "No personal bests found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /marathon/personal-bests [get]
func (h *MarathonHandler) GetPersonalBests(c fiber.Ctx) error {
	// 1. Parse query parameter
	playerID := c.Query("playerId")
	if playerID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "playerId query parameter is required")
	}

	// 2. Execute use case
	output, err := h.getPersonalBestsUC.Execute(appMarathon.GetPersonalBestsInput{
		PlayerID: playerID,
	})
	if err != nil {
		return mapMarathonError(err)
	}

	// 3. Return response
	return c.JSON(fiber.Map{
		"data": output,
	})
}

// GetMarathonLeaderboard handles GET /api/v1/marathon/leaderboard
// @Summary Get marathon leaderboard
// @Description Get the leaderboard for a specific category or all categories, with optional time frame filter
// @Tags marathon
// @Accept json
// @Produce json
// @Param categoryId query string false "Category ID (empty or 'all' for all categories)"
// @Param timeFrame query string false "Time frame: all_time (default), weekly, daily"
// @Param limit query int false "Number of entries to return (default 10, max 100)"
// @Success 200 {object} GetMarathonLeaderboardResponse "Leaderboard entries"
// @Failure 400 {object} ErrorResponse "Invalid parameters"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /marathon/leaderboard [get]
func (h *MarathonHandler) GetMarathonLeaderboard(c fiber.Ctx) error {
	// 1. Extract query parameters
	categoryID := c.Query("categoryId")
	timeFrame := c.Query("timeFrame", "all_time")
	limit := fiber.Query[int](c, "limit", 10)

	// Validate limit
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	// Validate timeFrame
	validTimeFrames := map[string]bool{
		"all_time": true,
		"weekly":   true,
		"daily":    true,
	}
	if !validTimeFrames[timeFrame] {
		return fiber.NewError(fiber.StatusBadRequest, "timeFrame must be one of: all_time, weekly, daily")
	}

	// 2. Execute use case
	output, err := h.getLeaderboardUC.Execute(appMarathon.GetMarathonLeaderboardInput{
		CategoryID: categoryID,
		TimeFrame:  timeFrame,
		Limit:      limit,
	})
	if err != nil {
		return mapMarathonError(err)
	}

	// 3. Return response
	return c.JSON(fiber.Map{
		"data": output,
	})
}

// ========================================
// Error Mapping (Domain → HTTP)
// ========================================

func mapMarathonError(err error) error {
	switch err {
	// Not Found errors
	case domainMarathon.ErrGameNotFound:
		return fiber.NewError(fiber.StatusNotFound, "Marathon game not found")
	case domainMarathon.ErrPersonalBestNotFound:
		return fiber.NewError(fiber.StatusNotFound, "Personal best not found")
	case domainQuiz.ErrQuestionNotFound:
		return fiber.NewError(fiber.StatusNotFound, "Question not found")
	case domainQuiz.ErrAnswerNotFound:
		return fiber.NewError(fiber.StatusNotFound, "Answer not found")
	case domainQuiz.ErrCategoryNotFound:
		return fiber.NewError(fiber.StatusNotFound, "Category not found")

	// Bad Request errors (validation)
	case domainMarathon.ErrInvalidGameID,
		domainMarathon.ErrInvalidPersonalBestID:
		return fiber.NewError(fiber.StatusBadRequest, err.Error())

	case domainQuiz.ErrInvalidQuestionID,
		domainQuiz.ErrInvalidAnswerID,
		domainQuiz.ErrInvalidCategoryID:
		return fiber.NewError(fiber.StatusBadRequest, err.Error())

	// Conflict errors
	case domainMarathon.ErrActiveGameExists:
		return fiber.NewError(fiber.StatusConflict, "Active marathon game already exists")
	case domainMarathon.ErrGameAlreadyFinished:
		return fiber.NewError(fiber.StatusConflict, "Marathon game already finished")

	// Business rule errors
	case domainMarathon.ErrGameNotActive:
		return fiber.NewError(fiber.StatusBadRequest, "Game is not active")
	case domainMarathon.ErrInvalidQuestion:
		return fiber.NewError(fiber.StatusBadRequest, "Question does not match current question or is invalid")
	case domainMarathon.ErrNoHintsAvailable:
		return fiber.NewError(fiber.StatusBadRequest, "Hint not available")
	case domainMarathon.ErrInvalidHintType:
		return fiber.NewError(fiber.StatusBadRequest, "Invalid hint type")
	case domainMarathon.ErrHintAlreadyUsed:
		return fiber.NewError(fiber.StatusBadRequest, "Hint already used for this question")
	case domainMarathon.ErrNoLivesRemaining:
		return fiber.NewError(fiber.StatusBadRequest, "No lives remaining")
	case domainMarathon.ErrNoQuestionsAvailable:
		return fiber.NewError(fiber.StatusBadRequest, "Insufficient questions available for marathon mode")
	case domainMarathon.ErrInvalidGameStatus:
		return fiber.NewError(fiber.StatusBadRequest, "Invalid game status transition")

	// Default: Internal Server Error
	default:
		return fiber.NewError(fiber.StatusInternalServerError, "Internal server error")
	}
}
