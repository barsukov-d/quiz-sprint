package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	appDuel "github.com/barsukov/quiz-sprint/backend/internal/application/quick_duel"
	domainDuel "github.com/barsukov/quiz-sprint/backend/internal/domain/quick_duel"
)

// DuelHandler handles HTTP requests for PvP Duel game mode
type DuelHandler struct {
	getStatusUC         *appDuel.GetDuelStatusUseCase
	joinQueueUC         *appDuel.JoinQueueUseCase
	leaveQueueUC        *appDuel.LeaveQueueUseCase
	sendChallengeUC     *appDuel.SendChallengeUseCase
	respondChallengeUC  *appDuel.RespondChallengeUseCase
	createLinkUC        *appDuel.CreateChallengeLinkUseCase
	getHistoryUC        *appDuel.GetMatchHistoryUseCase
	getLeaderboardUC    *appDuel.GetLeaderboardUseCase
	requestRematchUC    *appDuel.RequestRematchUseCase
}

func NewDuelHandler(
	getStatusUC *appDuel.GetDuelStatusUseCase,
	joinQueueUC *appDuel.JoinQueueUseCase,
	leaveQueueUC *appDuel.LeaveQueueUseCase,
	sendChallengeUC *appDuel.SendChallengeUseCase,
	respondChallengeUC *appDuel.RespondChallengeUseCase,
	createLinkUC *appDuel.CreateChallengeLinkUseCase,
	getHistoryUC *appDuel.GetMatchHistoryUseCase,
	getLeaderboardUC *appDuel.GetLeaderboardUseCase,
	requestRematchUC *appDuel.RequestRematchUseCase,
) *DuelHandler {
	return &DuelHandler{
		getStatusUC:         getStatusUC,
		joinQueueUC:         joinQueueUC,
		leaveQueueUC:        leaveQueueUC,
		sendChallengeUC:     sendChallengeUC,
		respondChallengeUC:  respondChallengeUC,
		createLinkUC:        createLinkUC,
		getHistoryUC:        getHistoryUC,
		getLeaderboardUC:    getLeaderboardUC,
		requestRematchUC:    requestRematchUC,
	}
}

// GetDuelStatus handles GET /api/v1/duel/status
// @Summary Get duel status
// @Description Get player's duel status, MMR, pending challenges
// @Tags duel
// @Accept json
// @Produce json
// @Param playerId query string true "Player ID"
// @Success 200 {object} GetDuelStatusResponse "Status"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 500 {object} ErrorResponse "Internal error"
// @Router /duel/status [get]
func (h *DuelHandler) GetDuelStatus(c fiber.Ctx) error {
	playerID := c.Query("playerId")
	if playerID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "playerId is required")
	}

	output, err := h.getStatusUC.Execute(appDuel.GetDuelStatusInput{
		PlayerID: playerID,
	})
	if err != nil {
		return mapDuelError(err)
	}

	return c.JSON(fiber.Map{"data": output})
}

// JoinQueue handles POST /api/v1/duel/queue/join
// @Summary Join matchmaking queue
// @Description Enter random matchmaking queue
// @Tags duel
// @Accept json
// @Produce json
// @Param request body JoinQueueRequest true "Join request"
// @Success 200 {object} JoinQueueResponse "Joined queue"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 409 {object} ErrorResponse "Already in queue or match"
// @Failure 500 {object} ErrorResponse "Internal error"
// @Router /duel/queue/join [post]
func (h *DuelHandler) JoinQueue(c fiber.Ctx) error {
	var req JoinQueueRequest
	if err := c.Bind().Body(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if req.PlayerID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "playerId is required")
	}

	output, err := h.joinQueueUC.Execute(appDuel.JoinQueueInput{
		PlayerID: req.PlayerID,
	})
	if err != nil {
		return mapDuelError(err)
	}

	return c.JSON(fiber.Map{"data": output})
}

// LeaveQueue handles DELETE /api/v1/duel/queue/leave
// @Summary Leave matchmaking queue
// @Description Cancel queue search (ticket refunded)
// @Tags duel
// @Accept json
// @Produce json
// @Param playerId query string true "Player ID"
// @Success 200 {object} LeaveQueueResponse "Left queue"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 500 {object} ErrorResponse "Internal error"
// @Router /duel/queue/leave [delete]
func (h *DuelHandler) LeaveQueue(c fiber.Ctx) error {
	playerID := c.Query("playerId")
	if playerID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "playerId is required")
	}

	output, err := h.leaveQueueUC.Execute(appDuel.LeaveQueueInput{
		PlayerID: playerID,
	})
	if err != nil {
		return mapDuelError(err)
	}

	return c.JSON(fiber.Map{"data": output})
}

// SendChallenge handles POST /api/v1/duel/challenge
// @Summary Send friend challenge
// @Description Send a direct challenge to a friend
// @Tags duel
// @Accept json
// @Produce json
// @Param request body SendChallengeRequest true "Challenge request"
// @Success 201 {object} SendChallengeResponse "Challenge sent"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 409 {object} ErrorResponse "Friend busy"
// @Failure 500 {object} ErrorResponse "Internal error"
// @Router /duel/challenge [post]
func (h *DuelHandler) SendChallenge(c fiber.Ctx) error {
	var req SendChallengeRequest
	if err := c.Bind().Body(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if req.PlayerID == "" || req.FriendID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "playerId and friendId are required")
	}

	output, err := h.sendChallengeUC.Execute(appDuel.SendChallengeInput{
		PlayerID: req.PlayerID,
		FriendID: req.FriendID,
	})
	if err != nil {
		return mapDuelError(err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": output})
}

// RespondChallenge handles POST /api/v1/duel/challenge/:challengeId/respond
// @Summary Respond to challenge
// @Description Accept or decline a friend challenge
// @Tags duel
// @Accept json
// @Produce json
// @Param challengeId path string true "Challenge ID"
// @Param request body RespondChallengeRequest true "Response request"
// @Success 200 {object} RespondChallengeResponse "Response processed"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 404 {object} ErrorResponse "Challenge not found"
// @Failure 409 {object} ErrorResponse "Challenge expired"
// @Failure 500 {object} ErrorResponse "Internal error"
// @Router /duel/challenge/{challengeId}/respond [post]
func (h *DuelHandler) RespondChallenge(c fiber.Ctx) error {
	challengeID := c.Params("challengeId")
	if _, err := uuid.Parse(challengeID); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid challenge ID format")
	}

	var req RespondChallengeRequest
	if err := c.Bind().Body(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if req.PlayerID == "" || req.Action == "" {
		return fiber.NewError(fiber.StatusBadRequest, "playerId and action are required")
	}

	if req.Action != "accept" && req.Action != "decline" {
		return fiber.NewError(fiber.StatusBadRequest, "action must be 'accept' or 'decline'")
	}

	output, err := h.respondChallengeUC.Execute(appDuel.RespondChallengeInput{
		PlayerID:    req.PlayerID,
		ChallengeID: challengeID,
		Action:      req.Action,
	})
	if err != nil {
		return mapDuelError(err)
	}

	return c.JSON(fiber.Map{"data": output})
}

// CreateChallengeLink handles POST /api/v1/duel/challenge/link
// @Summary Create challenge link
// @Description Generate a shareable challenge link (24h valid)
// @Tags duel
// @Accept json
// @Produce json
// @Param request body CreateChallengeLinkRequest true "Link request"
// @Success 201 {object} CreateChallengeLinkResponse "Link created"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 500 {object} ErrorResponse "Internal error"
// @Router /duel/challenge/link [post]
func (h *DuelHandler) CreateChallengeLink(c fiber.Ctx) error {
	var req CreateChallengeLinkRequest
	if err := c.Bind().Body(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if req.PlayerID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "playerId is required")
	}

	output, err := h.createLinkUC.Execute(appDuel.CreateChallengeLinkInput{
		PlayerID: req.PlayerID,
	})
	if err != nil {
		return mapDuelError(err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": output})
}

// GetMatchHistory handles GET /api/v1/duel/history
// @Summary Get match history
// @Description Get player's duel match history with pagination
// @Tags duel
// @Accept json
// @Produce json
// @Param playerId query string true "Player ID"
// @Param limit query int false "Limit (default 20, max 100)"
// @Param offset query int false "Offset"
// @Param filter query string false "Filter: all, friends, wins, losses"
// @Success 200 {object} GetMatchHistoryResponse "Match history"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 500 {object} ErrorResponse "Internal error"
// @Router /duel/history [get]
func (h *DuelHandler) GetMatchHistory(c fiber.Ctx) error {
	playerID := c.Query("playerId")
	if playerID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "playerId is required")
	}

	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}
	offset := 0
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil {
			offset = parsed
		}
	}
	filter := c.Query("filter", "all")

	output, err := h.getHistoryUC.Execute(appDuel.GetMatchHistoryInput{
		PlayerID: playerID,
		Limit:    limit,
		Offset:   offset,
		Filter:   filter,
	})
	if err != nil {
		return mapDuelError(err)
	}

	return c.JSON(fiber.Map{"data": output})
}

// GetDuelLeaderboard handles GET /api/v1/duel/leaderboard
// @Summary Get duel leaderboard
// @Description Get leaderboard by type (seasonal, friends, referrals)
// @Tags duel
// @Accept json
// @Produce json
// @Param playerId query string true "Player ID"
// @Param type query string false "Type: seasonal, friends, referrals (default seasonal)"
// @Param limit query int false "Limit (default 10, max 100)"
// @Success 200 {object} GetLeaderboardResponse "Leaderboard"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 500 {object} ErrorResponse "Internal error"
// @Router /duel/leaderboard [get]
func (h *DuelHandler) GetDuelLeaderboard(c fiber.Ctx) error {
	playerID := c.Query("playerId")
	if playerID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "playerId is required")
	}

	leaderboardType := c.Query("type", "seasonal")
	limit := 10
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	output, err := h.getLeaderboardUC.Execute(appDuel.GetLeaderboardInput{
		PlayerID: playerID,
		Type:     leaderboardType,
		Limit:    limit,
	})
	if err != nil {
		return mapDuelError(err)
	}

	return c.JSON(fiber.Map{"data": output})
}

// RequestRematch handles POST /api/v1/duel/match/:matchId/rematch
// @Summary Request rematch
// @Description Request a rematch after a completed duel
// @Tags duel
// @Accept json
// @Produce json
// @Param matchId path string true "Match ID"
// @Param request body RequestRematchRequest true "Rematch request"
// @Success 200 {object} RequestRematchResponse "Rematch requested"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 404 {object} ErrorResponse "Match not found"
// @Failure 409 {object} ErrorResponse "Cannot rematch"
// @Failure 500 {object} ErrorResponse "Internal error"
// @Router /duel/match/{matchId}/rematch [post]
func (h *DuelHandler) RequestRematch(c fiber.Ctx) error {
	if h.requestRematchUC == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "Rematch not available")
	}

	matchID := c.Params("matchId")
	if _, err := uuid.Parse(matchID); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid match ID format")
	}

	var req RequestRematchRequest
	if err := c.Bind().Body(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if req.PlayerID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "playerId is required")
	}

	output, err := h.requestRematchUC.Execute(appDuel.RequestRematchInput{
		PlayerID: req.PlayerID,
		MatchID:  matchID,
	})
	if err != nil {
		return mapDuelError(err)
	}

	return c.JSON(fiber.Map{"data": output})
}

// mapDuelError maps domain errors to HTTP errors
func mapDuelError(err error) error {
	switch err {
	case domainDuel.ErrGameNotFound:
		return fiber.NewError(fiber.StatusNotFound, "Match not found")
	case domainDuel.ErrGameNotActive:
		return fiber.NewError(fiber.StatusConflict, "Match is not active")
	case domainDuel.ErrChallengeNotFound:
		return fiber.NewError(fiber.StatusNotFound, "Challenge not found")
	case domainDuel.ErrChallengeExpired:
		return fiber.NewError(fiber.StatusConflict, "Challenge has expired")
	case domainDuel.ErrAlreadyInQueue:
		return fiber.NewError(fiber.StatusConflict, "Already in matchmaking queue")
	case domainDuel.ErrAlreadyInMatch:
		return fiber.NewError(fiber.StatusConflict, "Already in an active match")
	case domainDuel.ErrFriendBusy:
		return fiber.NewError(fiber.StatusConflict, "Friend is already in a match")
	case domainDuel.ErrInsufficientTickets:
		return fiber.NewError(fiber.StatusBadRequest, "Insufficient tickets")
	case domainDuel.ErrCannotChallengeSelf:
		return fiber.NewError(fiber.StatusBadRequest, "Cannot challenge yourself")
	default:
		return fiber.NewError(fiber.StatusInternalServerError, "Internal server error")
	}
}
