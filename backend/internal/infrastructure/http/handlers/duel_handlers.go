package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	appDuel "github.com/barsukov/quiz-sprint/backend/internal/application/quick_duel"
	domainDuel "github.com/barsukov/quiz-sprint/backend/internal/domain/quick_duel"
	"github.com/barsukov/quiz-sprint/backend/internal/infrastructure/http/middleware"
)

// DuelHandler handles HTTP requests for PvP Duel game mode
type DuelHandler struct {
	getStatusUC          *appDuel.GetDuelStatusUseCase
	joinQueueUC          *appDuel.JoinQueueUseCase
	leaveQueueUC         *appDuel.LeaveQueueUseCase
	sendChallengeUC      *appDuel.SendChallengeUseCase
	respondChallengeUC   *appDuel.RespondChallengeUseCase
	acceptByLinkCodeUC   *appDuel.AcceptByLinkCodeUseCase
	startChallengeUC     *appDuel.StartChallengeUseCase
	createLinkUC         *appDuel.CreateChallengeLinkUseCase
	getHistoryUC         *appDuel.GetGameHistoryUseCase
	getLeaderboardUC     *appDuel.GetLeaderboardUseCase
	requestRematchUC     *appDuel.RequestRematchUseCase
	getGameResultUC      *appDuel.GetGameResultUseCase
	getRivalsUC          *appDuel.GetRivalsUseCase
	prepareShareUC       *appDuel.PrepareShareUseCase
	getReferralsUC       *appDuel.GetReferralsUseCase
	claimReferralUC      *appDuel.ClaimReferralRewardUseCase
	surrenderGameUC      *appDuel.SurrenderGameUseCase
}

func NewDuelHandler(
	getStatusUC *appDuel.GetDuelStatusUseCase,
	joinQueueUC *appDuel.JoinQueueUseCase,
	leaveQueueUC *appDuel.LeaveQueueUseCase,
	sendChallengeUC *appDuel.SendChallengeUseCase,
	respondChallengeUC *appDuel.RespondChallengeUseCase,
	acceptByLinkCodeUC *appDuel.AcceptByLinkCodeUseCase,
	startChallengeUC *appDuel.StartChallengeUseCase,
	createLinkUC *appDuel.CreateChallengeLinkUseCase,
	getHistoryUC *appDuel.GetGameHistoryUseCase,
	getLeaderboardUC *appDuel.GetLeaderboardUseCase,
	requestRematchUC *appDuel.RequestRematchUseCase,
	getGameResultUC *appDuel.GetGameResultUseCase,
	getRivalsUC *appDuel.GetRivalsUseCase,
	prepareShareUC *appDuel.PrepareShareUseCase,
	getReferralsUC *appDuel.GetReferralsUseCase,
	claimReferralUC *appDuel.ClaimReferralRewardUseCase,
	surrenderGameUC *appDuel.SurrenderGameUseCase,
) *DuelHandler {
	return &DuelHandler{
		getStatusUC:          getStatusUC,
		joinQueueUC:          joinQueueUC,
		leaveQueueUC:         leaveQueueUC,
		sendChallengeUC:      sendChallengeUC,
		respondChallengeUC:   respondChallengeUC,
		acceptByLinkCodeUC:   acceptByLinkCodeUC,
		startChallengeUC:     startChallengeUC,
		createLinkUC:         createLinkUC,
		getHistoryUC:         getHistoryUC,
		getLeaderboardUC:     getLeaderboardUC,
		requestRematchUC:     requestRematchUC,
		getGameResultUC:      getGameResultUC,
		getRivalsUC:          getRivalsUC,
		prepareShareUC:       prepareShareUC,
		getReferralsUC:       getReferralsUC,
		claimReferralUC:      claimReferralUC,
		surrenderGameUC:      surrenderGameUC,
	}
}

func getAuthPlayerID(c fiber.Ctx) (string, error) {
	initData := middleware.GetTelegramInitData(c)
	if initData == nil {
		return "", fiber.NewError(fiber.StatusUnauthorized, "Authentication required")
	}
	return strconv.FormatInt(initData.User.ID, 10), nil
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
	playerID, err := getAuthPlayerID(c)
	if err != nil {
		return err
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
// @Failure 409 {object} ErrorResponse "Already in queue or game"
// @Failure 500 {object} ErrorResponse "Internal error"
// @Router /duel/queue/join [post]
func (h *DuelHandler) JoinQueue(c fiber.Ctx) error {
	playerID, err := getAuthPlayerID(c)
	if err != nil {
		return err
	}

	output, err := h.joinQueueUC.Execute(appDuel.JoinQueueInput{
		PlayerID: playerID,
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
	playerID, err := getAuthPlayerID(c)
	if err != nil {
		return err
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
	playerID, err := getAuthPlayerID(c)
	if err != nil {
		return err
	}

	var req SendChallengeRequest
	if err := c.Bind().Body(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if req.FriendID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "friendId is required")
	}

	output, err := h.sendChallengeUC.Execute(appDuel.SendChallengeInput{
		PlayerID: playerID,
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
	playerID, err := getAuthPlayerID(c)
	if err != nil {
		return err
	}

	challengeID := c.Params("challengeId")
	if _, err := uuid.Parse(challengeID); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid challenge ID format")
	}

	var req RespondChallengeRequest
	if err := c.Bind().Body(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if req.Action == "" {
		return fiber.NewError(fiber.StatusBadRequest, "action is required")
	}

	if req.Action != "accept" && req.Action != "decline" {
		return fiber.NewError(fiber.StatusBadRequest, "action must be 'accept' or 'decline'")
	}

	output, err := h.respondChallengeUC.Execute(appDuel.RespondChallengeInput{
		PlayerID:    playerID,
		ChallengeID: challengeID,
		Action:      req.Action,
	})
	if err != nil {
		return mapDuelError(err)
	}

	return c.JSON(fiber.Map{"data": output})
}

// AcceptByLinkCode handles POST /api/v1/duel/challenge/accept-by-code
// @Summary Accept challenge by link code
// @Description Accept a challenge using the link code from deep link (e.g., "duel_abc12345")
// @Tags duel
// @Accept json
// @Produce json
// @Param request body AcceptByLinkCodeRequest true "Accept request"
// @Success 200 {object} AcceptByLinkCodeResponse "Challenge accepted"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 404 {object} ErrorResponse "Challenge not found"
// @Failure 409 {object} ErrorResponse "Challenge expired or already accepted"
// @Failure 500 {object} ErrorResponse "Internal error"
// @Router /duel/challenge/accept-by-code [post]
func (h *DuelHandler) AcceptByLinkCode(c fiber.Ctx) error {
	playerID, err := getAuthPlayerID(c)
	if err != nil {
		return err
	}

	var req AcceptByLinkCodeRequest
	if err := c.Bind().Body(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if req.LinkCode == "" {
		return fiber.NewError(fiber.StatusBadRequest, "linkCode is required")
	}

	output, err := h.acceptByLinkCodeUC.Execute(appDuel.AcceptByLinkCodeInput{
		PlayerID: playerID,
		LinkCode: req.LinkCode,
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
	playerID, err := getAuthPlayerID(c)
	if err != nil {
		return err
	}

	output, err := h.createLinkUC.Execute(appDuel.CreateChallengeLinkInput{
		PlayerID: playerID,
	})
	if err != nil {
		return mapDuelError(err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": output})
}

// GetGameHistory handles GET /api/v1/duel/history
// @Summary Get game history
// @Description Get player's duel game history with pagination
// @Tags duel
// @Accept json
// @Produce json
// @Param playerId query string true "Player ID"
// @Param limit query int false "Limit (default 20, max 100)"
// @Param offset query int false "Offset"
// @Param filter query string false "Filter: all, friends, wins, losses"
// @Success 200 {object} GetGameHistoryResponse "Game history"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 500 {object} ErrorResponse "Internal error"
// @Router /duel/history [get]
func (h *DuelHandler) GetGameHistory(c fiber.Ctx) error {
	playerID, err := getAuthPlayerID(c)
	if err != nil {
		return err
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

	output, err := h.getHistoryUC.Execute(appDuel.GetGameHistoryInput{
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
// @Success 200 {object} GetDuelLeaderboardResponse "Leaderboard"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 500 {object} ErrorResponse "Internal error"
// @Router /duel/leaderboard [get]
func (h *DuelHandler) GetDuelLeaderboard(c fiber.Ctx) error {
	playerID, err := getAuthPlayerID(c)
	if err != nil {
		return err
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

// RequestRematch handles POST /api/v1/duel/game/:gameId/rematch
// @Summary Request rematch
// @Description Request a rematch after a completed duel
// @Tags duel
// @Accept json
// @Produce json
// @Param gameId path string true "Game ID"
// @Param request body RequestRematchRequest true "Rematch request"
// @Success 200 {object} RequestRematchResponse "Rematch requested"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 404 {object} ErrorResponse "Game not found"
// @Failure 409 {object} ErrorResponse "Cannot rematch"
// @Failure 500 {object} ErrorResponse "Internal error"
// @Router /duel/game/{gameId}/rematch [post]
func (h *DuelHandler) RequestRematch(c fiber.Ctx) error {
	if h.requestRematchUC == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "Rematch not available")
	}

	gameID := c.Params("gameId")
	if _, err := uuid.Parse(gameID); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid game ID format")
	}

	playerID, err := getAuthPlayerID(c)
	if err != nil {
		return err
	}

	output, err := h.requestRematchUC.Execute(appDuel.RequestRematchInput{
		PlayerID: playerID,
		GameID:   gameID,
	})
	if err != nil {
		return mapDuelError(err)
	}

	return c.JSON(fiber.Map{"data": output})
}

// GetGameResult handles GET /api/v1/duel/game/:gameId
// @Summary Get game result
// @Description Get full result of a finished duel game
// @Tags duel
// @Accept json
// @Produce json
// @Param gameId path string true "Game ID"
// @Param playerId query string true "Player ID"
// @Success 200 {object} GetGameResultResponse "Game result"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 404 {object} ErrorResponse "Game not found"
// @Failure 500 {object} ErrorResponse "Internal error"
// @Router /duel/game/{gameId} [get]
func (h *DuelHandler) GetGameResult(c fiber.Ctx) error {
	playerID, err := getAuthPlayerID(c)
	if err != nil {
		return err
	}

	gameID := c.Params("gameId")
	if _, err := uuid.Parse(gameID); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid game ID format")
	}

	output, err := h.getGameResultUC.Execute(appDuel.GetGameResultInput{
		PlayerID: playerID,
		GameID:   gameID,
	})
	if err != nil {
		return mapDuelError(err)
	}

	return c.JSON(fiber.Map{"data": output})
}

// StartChallenge handles POST /api/v1/duel/challenge/:challengeId/start
// @Summary Start the duel after invitee accepted
// @Description Inviter confirms game start after invitee accepted via link
// @Tags duel
// @Accept json
// @Produce json
// @Param challengeId path string true "Challenge ID"
// @Param request body StartChallengeRequest true "Start request"
// @Success 200 {object} StartChallengeResponse "Game started"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 404 {object} ErrorResponse "Challenge not found"
// @Failure 409 {object} ErrorResponse "Challenge not in accepted_waiting_inviter state"
// @Router /duel/challenge/{challengeId}/start [post]
func (h *DuelHandler) StartChallenge(c fiber.Ctx) error {
	playerID, err := getAuthPlayerID(c)
	if err != nil {
		return err
	}

	challengeID := c.Params("challengeId")
	if _, err := uuid.Parse(challengeID); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid challenge ID format")
	}

	if h.startChallengeUC == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "Service not available")
	}

	output, err := h.startChallengeUC.Execute(appDuel.StartChallengeInput{
		PlayerID:    playerID,
		ChallengeID: challengeID,
	})
	if err != nil {
		return mapDuelError(err)
	}

	return c.JSON(fiber.Map{"data": output})
}

// PrepareShare handles POST /api/v1/duel/challenge/prepare-share
// @Summary Prepare inline message for sharing challenge
// @Description Creates a Telegram prepared inline message with an "Accept Challenge" button. Use the returned preparedMessageId with the TMA SDK shareMessage() call (Mini Apps v8.0+).
// @Tags duel
// @Accept json
// @Produce json
// @Param request body PrepareShareRequest true "Prepare share request"
// @Success 200 {object} PrepareShareResponse "Prepared message ID"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 500 {object} ErrorResponse "Internal error"
// @Router /duel/challenge/prepare-share [post]
func (h *DuelHandler) PrepareShare(c fiber.Ctx) error {
	if h.prepareShareUC == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "Prepare share not available")
	}

	playerID, err := getAuthPlayerID(c)
	if err != nil {
		return err
	}

	var req PrepareShareRequest
	if err := c.Bind().Body(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if req.ChallengeLink == "" {
		return fiber.NewError(fiber.StatusBadRequest, "challengeLink is required")
	}

	output, err := h.prepareShareUC.Execute(c.Context(), appDuel.PrepareShareInput{
		PlayerID:      playerID,
		ChallengeLink: req.ChallengeLink,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{"data": output})
}

// mapDuelError maps domain errors to HTTP errors with structured error codes.
// The errorCode field allows frontend to handle errors programmatically.
func mapDuelError(err error) error {
	switch err {
	case domainDuel.ErrGameNotFound:
		return NewAppError(fiber.StatusNotFound, string(domainDuel.CodeGameNotFound), "Game not found")
	case domainDuel.ErrGameNotActive:
		return NewAppError(fiber.StatusConflict, string(domainDuel.CodeGameNotActive), "Game is not active")
	case domainDuel.ErrGameAlreadyFinished:
		return NewAppError(fiber.StatusConflict, string(domainDuel.CodeGameAlreadyFinished), "Game is already finished")
	case domainDuel.ErrChallengeNotFound:
		return NewAppError(fiber.StatusNotFound, string(domainDuel.CodeChallengeNotFound), "Challenge not found")
	case domainDuel.ErrChallengeExpired:
		return NewAppError(fiber.StatusConflict, string(domainDuel.CodeChallengeExpired), "Challenge has expired")
	case domainDuel.ErrChallengeNotPending:
		return NewAppError(fiber.StatusConflict, string(domainDuel.CodeChallengeNotPending), "Challenge is no longer pending")
	case domainDuel.ErrAlreadyInQueue:
		return NewAppError(fiber.StatusConflict, string(domainDuel.CodeAlreadyInQueue), "Already in matchmaking queue")
	case domainDuel.ErrAlreadyInGame:
		return NewAppError(fiber.StatusConflict, string(domainDuel.CodeAlreadyInGame), "Already in an active game")
	case domainDuel.ErrFriendBusy:
		return NewAppError(fiber.StatusConflict, string(domainDuel.CodeFriendBusy), "Friend is already in a game")
	case domainDuel.ErrChallengeAlreadySent:
		return NewAppError(fiber.StatusConflict, string(domainDuel.CodeChallengeAlreadySent), "Challenge already sent to this player")
	case domainDuel.ErrInsufficientTickets:
		return NewAppError(fiber.StatusPaymentRequired, string(domainDuel.CodeInsufficientTickets), "Insufficient tickets")
	case domainDuel.ErrCannotChallengeSelf:
		return NewAppError(fiber.StatusBadRequest, string(domainDuel.CodeCannotChallengeSelf), "Cannot challenge yourself")
	case domainDuel.ErrNotChallengedPlayer:
		return NewAppError(fiber.StatusForbidden, string(domainDuel.CodeNotChallengedPlayer), "Not the challenged player")
	case domainDuel.ErrPlayerNotInGame:
		return NewAppError(fiber.StatusForbidden, string(domainDuel.CodePlayerNotInGame), "Player not in this game")
	case domainDuel.ErrTooEarlyToSurrender:
		return NewAppError(fiber.StatusConflict, string(domainDuel.CodeTooEarlyToSurrender), "Cannot surrender before answering 3 questions")
	case domainDuel.ErrReferralNotFound:
		return NewAppError(fiber.StatusNotFound, string(domainDuel.CodeReferralNotFound), "Referral not found")
	case domainDuel.ErrMilestoneNotReached:
		return NewAppError(fiber.StatusBadRequest, string(domainDuel.CodeMilestoneNotReached), "Milestone not reached")
	case domainDuel.ErrRewardAlreadyClaimed:
		return NewAppError(fiber.StatusConflict, string(domainDuel.CodeRewardAlreadyClaimed), "Reward already claimed")
	default:
		return fiber.NewError(fiber.StatusInternalServerError, "Internal server error")
	}
}

// GetReferrals handles GET /api/v1/duel/referrals
// @Summary Get player referrals
// @Description List player's referrals with milestone progress and pending rewards
// @Tags duel
// @Accept json
// @Produce json
// @Success 200 {object} GetReferralsResponse "Referrals list"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 500 {object} ErrorResponse "Internal error"
// @Router /duel/referrals [get]
func (h *DuelHandler) GetReferrals(c fiber.Ctx) error {
	playerID, err := getAuthPlayerID(c)
	if err != nil {
		return err
	}

	output, err := h.getReferralsUC.Execute(appDuel.GetReferralsInput{
		PlayerID: playerID,
	})
	if err != nil {
		return mapDuelError(err)
	}

	return c.JSON(fiber.Map{"data": output})
}

// ClaimReferralReward handles POST /api/v1/duel/referrals/:friendId/claim
// @Summary Claim referral milestone reward
// @Description Claim the reward for a reached milestone with a referred friend
// @Tags duel
// @Accept json
// @Produce json
// @Param friendId path string true "Friend (invitee) player ID"
// @Param request body ClaimReferralRewardRequest true "Claim request"
// @Success 200 {object} ClaimReferralRewardResponse "Reward claimed"
// @Failure 400 {object} ErrorResponse "Invalid request or milestone not reached"
// @Failure 404 {object} ErrorResponse "Referral not found"
// @Failure 409 {object} ErrorResponse "Reward already claimed"
// @Failure 500 {object} ErrorResponse "Internal error"
// @Router /duel/referrals/{friendId}/claim [post]
func (h *DuelHandler) ClaimReferralReward(c fiber.Ctx) error {
	playerID, err := getAuthPlayerID(c)
	if err != nil {
		return err
	}

	friendID := c.Params("friendId")
	if friendID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "friendId is required")
	}

	var req ClaimReferralRewardRequest
	if err := c.Bind().Body(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if req.Milestone == "" {
		return fiber.NewError(fiber.StatusBadRequest, "milestone is required")
	}

	output, err := h.claimReferralUC.Execute(appDuel.ClaimReferralRewardInput{
		PlayerID:  playerID,
		FriendID:  friendID,
		Milestone: req.Milestone,
	})
	if err != nil {
		return mapDuelError(err)
	}

	return c.JSON(fiber.Map{"data": output})
}

// GetRivals handles GET /api/v1/duel/rivals
// @Summary Get recent rivals
// @Description Get list of recent unique opponents the player has faced
// @Tags duel
// @Accept json
// @Produce json
// @Param playerId query string true "Player ID"
// @Success 200 {object} GetRivalsResponse "Rivals list"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 500 {object} ErrorResponse "Internal error"
// @Router /duel/rivals [get]
func (h *DuelHandler) GetRivals(c fiber.Ctx) error {
	playerID, err := getAuthPlayerID(c)
	if err != nil {
		return err
	}

	output, err := h.getRivalsUC.Execute(appDuel.GetRivalsInput{
		PlayerID: playerID,
	})
	if err != nil {
		return mapDuelError(err)
	}

	return c.JSON(fiber.Map{"data": output})
}

// SurrenderGame handles POST /api/v1/duel/game/:gameId/surrender
// @Summary Surrender a duel game
// @Description Forfeit an active duel game. The surrendering player takes an ELO loss; the opponent wins.
// @Tags duel
// @Accept json
// @Produce json
// @Param gameId path string true "Game ID"
// @Success 200 {object} SurrenderGameResponse "Surrender processed"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 403 {object} ErrorResponse "Player not in game"
// @Failure 404 {object} ErrorResponse "Game not found"
// @Failure 409 {object} ErrorResponse "Game is not active"
// @Failure 500 {object} ErrorResponse "Internal error"
// @Router /duel/game/{gameId}/surrender [post]
func (h *DuelHandler) SurrenderGame(c fiber.Ctx) error {
	playerID, err := getAuthPlayerID(c)
	if err != nil {
		return err
	}

	gameID := c.Params("gameId")
	if gameID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "gameId is required")
	}

	if h.surrenderGameUC == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "Surrender not available")
	}

	output, err := h.surrenderGameUC.Execute(appDuel.SurrenderGameInput{
		PlayerID: playerID,
		GameID:   gameID,
	})
	if err != nil {
		return mapDuelError(err)
	}

	return c.JSON(fiber.Map{"data": output})
}
