package quick_duel

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quick_duel"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
	domainUser "github.com/barsukov/quiz-sprint/backend/internal/domain/user"
)

// ========================================
// GetDuelStatus Use Case
// ========================================

type GetDuelStatusUseCase struct {
	playerRatingRepo quick_duel.PlayerRatingRepository
	duelGameRepo     quick_duel.DuelGameRepository
	challengeRepo    quick_duel.ChallengeRepository
	seasonRepo       quick_duel.SeasonRepository
	userRepo         domainUser.UserRepository
	onlineTracker    OnlineTracker // optional, nil if Redis unavailable
}

func NewGetDuelStatusUseCase(
	playerRatingRepo quick_duel.PlayerRatingRepository,
	duelGameRepo quick_duel.DuelGameRepository,
	challengeRepo quick_duel.ChallengeRepository,
	seasonRepo quick_duel.SeasonRepository,
	userRepo domainUser.UserRepository,
	onlineTracker OnlineTracker,
) *GetDuelStatusUseCase {
	return &GetDuelStatusUseCase{
		playerRatingRepo: playerRatingRepo,
		duelGameRepo:     duelGameRepo,
		challengeRepo:    challengeRepo,
		seasonRepo:       seasonRepo,
		userRepo:         userRepo,
		onlineTracker:    onlineTracker,
	}
}

func (uc *GetDuelStatusUseCase) Execute(input GetDuelStatusInput) (GetDuelStatusOutput, error) {
	now := time.Now().UTC().Unix()

	playerID, err := shared.NewUserID(input.PlayerID)
	if err != nil {
		return GetDuelStatusOutput{}, err
	}

	// Mark player as online (TTL=120s; refreshed on every status poll)
	if uc.onlineTracker != nil {
		_ = uc.onlineTracker.SetOnline(input.PlayerID, 120)
	}

	// Get current season
	seasonID, err := uc.seasonRepo.GetCurrentSeason()
	if err != nil {
		seasonID = "2026-02" // Default season
	}

	// Get or create player rating
	rating, err := uc.playerRatingRepo.FindOrCreate(playerID, seasonID, now)
	if err != nil {
		return GetDuelStatusOutput{}, err
	}

	// Check for active game
	var activeGameID *string
	activeGame, err := uc.duelGameRepo.FindActiveByPlayer(playerID)
	if err == nil && activeGame != nil {
		id := activeGame.ID().String()
		activeGameID = &id
	}

	// Get pending challenges
	pendingChallenges, err := uc.challengeRepo.FindPendingForPlayer(playerID)
	if err != nil {
		pendingChallenges = []*quick_duel.DuelChallenge{}
	}

	challengeDTOs := make([]ChallengeDTO, 0, len(pendingChallenges))
	for _, c := range pendingChallenges {
		username := c.ChallengerID().String() // fallback = ID
		if u, err := uc.userRepo.FindByID(c.ChallengerID()); err == nil && u != nil {
			if !u.Username().IsAnonymous() {
				username = u.Username().String()
			} else if u.TelegramUsername().String() != "" {
				username = u.TelegramUsername().String()
			}
		}
		challengeDTOs = append(challengeDTOs, ToChallengeDTO(c, now, username))
	}

	// Get season end time
	_, seasonEndsAt, _ := uc.seasonRepo.GetSeasonInfo(seasonID)

	outgoingChallenges, err := uc.challengeRepo.FindPendingByChallenger(playerID)
	if err != nil {
		outgoingChallenges = []*quick_duel.DuelChallenge{}
	}

	outgoingDTOs := make([]ChallengeDTO, 0, len(outgoingChallenges))
	for _, c := range outgoingChallenges {
		dto := ToChallengeDTO(c, now, "")
		if c.ChallengedID() != nil {
			if u, err := uc.userRepo.FindByID(*c.ChallengedID()); err == nil && u != nil {
				name := u.TelegramUsername().String()
				if name == "" {
					name = u.Username().String()
				}
				dto.InviteeName = name
			}
		}
		outgoingDTOs = append(outgoingDTOs, dto)
	}

	// F1: accepted challenges — invitee is waiting for inviter to start
	acceptedChallenges, err := uc.challengeRepo.FindAcceptedWaitingForPlayer(playerID)
	if err != nil {
		acceptedChallenges = []*quick_duel.DuelChallenge{}
	}
	acceptedDTOs := make([]ChallengeDTO, 0, len(acceptedChallenges))
	for _, c := range acceptedChallenges {
		username := c.ChallengerID().String()
		if u, err := uc.userRepo.FindByID(c.ChallengerID()); err == nil && u != nil {
			if n := u.TelegramUsername().String(); n != "" {
				username = n
			} else if n := u.Username().String(); n != "" {
				username = n
			}
		}
		acceptedDTOs = append(acceptedDTOs, ToChallengeDTO(c, now, username))
	}

	return GetDuelStatusOutput{
		HasActiveDuel:      activeGameID != nil,
		ActiveGameID:       activeGameID,
		Player:             ToPlayerRatingDTO(rating),
		Tickets:            10, // TODO: get from user wallet
		FriendsOnline:      []FriendDTO{}, // TODO: implement friends service
		PendingChallenges:  challengeDTOs,
		OutgoingChallenges: outgoingDTOs,
		AcceptedChallenges: acceptedDTOs,
		SeasonID:           seasonID,
		SeasonEndsAt:       seasonEndsAt,
	}, nil
}

// ========================================
// JoinQueue Use Case
// ========================================

type JoinQueueUseCase struct {
	matchmakingQueue quick_duel.MatchmakingQueue
	playerRatingRepo quick_duel.PlayerRatingRepository
	duelGameRepo     quick_duel.DuelGameRepository
	seasonRepo       quick_duel.SeasonRepository
}

func NewJoinQueueUseCase(
	matchmakingQueue quick_duel.MatchmakingQueue,
	playerRatingRepo quick_duel.PlayerRatingRepository,
	duelGameRepo quick_duel.DuelGameRepository,
	seasonRepo quick_duel.SeasonRepository,
) *JoinQueueUseCase {
	return &JoinQueueUseCase{
		matchmakingQueue: matchmakingQueue,
		playerRatingRepo: playerRatingRepo,
		duelGameRepo:     duelGameRepo,
		seasonRepo:       seasonRepo,
	}
}

func (uc *JoinQueueUseCase) Execute(input JoinQueueInput) (JoinQueueOutput, error) {
	now := time.Now().UTC().Unix()

	playerID, err := shared.NewUserID(input.PlayerID)
	if err != nil {
		return JoinQueueOutput{}, err
	}

	// Check if already in queue
	inQueue, err := uc.matchmakingQueue.IsPlayerInQueue(playerID)
	if err != nil {
		return JoinQueueOutput{}, err
	}
	if inQueue {
		return JoinQueueOutput{}, quick_duel.ErrAlreadyInQueue
	}

	// Check if already in game
	activeGame, err := uc.duelGameRepo.FindActiveByPlayer(playerID)
	if err == nil && activeGame != nil {
		return JoinQueueOutput{}, quick_duel.ErrAlreadyInGame
	}

	// Get player rating
	seasonID, _ := uc.seasonRepo.GetCurrentSeason()
	rating, err := uc.playerRatingRepo.FindOrCreate(playerID, seasonID, now)
	if err != nil {
		return JoinQueueOutput{}, err
	}

	// TODO: Check and consume ticket

	// Add to queue
	err = uc.matchmakingQueue.AddToQueue(playerID, rating.MMR(), now)
	if err != nil {
		return JoinQueueOutput{}, err
	}

	// Get queue info
	queueLength, _ := uc.matchmakingQueue.GetQueueLength()

	return JoinQueueOutput{
		QueueID:       playerID.String(),
		Status:        "searching",
		EstimatedWait: 10, // Estimate based on queue length
		MMRRange:      "±50",
		Position:      queueLength,
	}, nil
}

// ========================================
// LeaveQueue Use Case
// ========================================

type LeaveQueueUseCase struct {
	matchmakingQueue quick_duel.MatchmakingQueue
}

func NewLeaveQueueUseCase(
	matchmakingQueue quick_duel.MatchmakingQueue,
) *LeaveQueueUseCase {
	return &LeaveQueueUseCase{
		matchmakingQueue: matchmakingQueue,
	}
}

func (uc *LeaveQueueUseCase) Execute(input LeaveQueueInput) (LeaveQueueOutput, error) {
	playerID, err := shared.NewUserID(input.PlayerID)
	if err != nil {
		return LeaveQueueOutput{}, err
	}

	// Check if in queue
	inQueue, err := uc.matchmakingQueue.IsPlayerInQueue(playerID)
	if err != nil {
		return LeaveQueueOutput{}, err
	}

	if !inQueue {
		return LeaveQueueOutput{
			Success:        true,
			TicketRefunded: false,
			NewTicketCount: 10, // TODO: get from wallet
		}, nil
	}

	// Remove from queue
	err = uc.matchmakingQueue.RemoveFromQueue(playerID)
	if err != nil {
		return LeaveQueueOutput{}, err
	}

	// TODO: Refund ticket

	return LeaveQueueOutput{
		Success:        true,
		TicketRefunded: true,
		NewTicketCount: 10, // TODO: get from wallet
	}, nil
}

// ========================================
// SendChallenge Use Case
// ========================================

type SendChallengeUseCase struct {
	challengeRepo    quick_duel.ChallengeRepository
	duelGameRepo     quick_duel.DuelGameRepository
	eventBus         EventBus
}

func NewSendChallengeUseCase(
	challengeRepo quick_duel.ChallengeRepository,
	duelGameRepo quick_duel.DuelGameRepository,
	eventBus EventBus,
) *SendChallengeUseCase {
	return &SendChallengeUseCase{
		challengeRepo: challengeRepo,
		duelGameRepo:  duelGameRepo,
		eventBus:      eventBus,
	}
}

func (uc *SendChallengeUseCase) Execute(input SendChallengeInput) (SendChallengeOutput, error) {
	now := time.Now().UTC().Unix()

	challengerID, err := shared.NewUserID(input.PlayerID)
	if err != nil {
		return SendChallengeOutput{}, err
	}

	friendID, err := shared.NewUserID(input.FriendID)
	if err != nil {
		return SendChallengeOutput{}, err
	}

	// B3: Check if challenger is already in a game
	if activeGame, err := uc.duelGameRepo.FindActiveByPlayer(challengerID); err == nil && activeGame != nil {
		return SendChallengeOutput{}, quick_duel.ErrAlreadyInGame
	}

	// Check if friend is already in a game
	activeGame, err := uc.duelGameRepo.FindActiveByPlayer(friendID)
	if err == nil && activeGame != nil {
		return SendChallengeOutput{}, quick_duel.ErrFriendBusy
	}

	// Check for existing pending challenge to same friend
	existingChallenges, err := uc.challengeRepo.FindPendingByChallenger(challengerID)
	if err == nil {
		for _, c := range existingChallenges {
			if c.ChallengedID() != nil && c.ChallengedID().Equals(friendID) {
				return SendChallengeOutput{}, quick_duel.ErrChallengeAlreadySent
			}
		}
	}

	// Create challenge
	challenge, err := quick_duel.NewDirectChallenge(challengerID, friendID, now)
	if err != nil {
		return SendChallengeOutput{}, err
	}

	// Save challenge
	err = uc.challengeRepo.Save(challenge)
	if err != nil {
		return SendChallengeOutput{}, err
	}

	// Publish events
	for _, event := range challenge.Events() {
		uc.eventBus.Publish(event)
	}

	return SendChallengeOutput{
		ChallengeID:    challenge.ID().String(),
		Status:         "pending",
		ExpiresIn:      quick_duel.DirectChallengeExpirySeconds,
		TicketConsumed: true,
	}, nil
}

// ========================================
// RespondChallenge Use Case
// ========================================

type RespondChallengeUseCase struct {
	challengeRepo    quick_duel.ChallengeRepository
	duelGameRepo     quick_duel.DuelGameRepository
	playerRatingRepo quick_duel.PlayerRatingRepository
	questionRepo     QuestionRepository // nil if questions DB unavailable
	seasonRepo       quick_duel.SeasonRepository
	userRepo         domainUser.UserRepository
	eventBus         EventBus
}

func NewRespondChallengeUseCase(
	challengeRepo quick_duel.ChallengeRepository,
	duelGameRepo quick_duel.DuelGameRepository,
	playerRatingRepo quick_duel.PlayerRatingRepository,
	seasonRepo quick_duel.SeasonRepository,
	questionRepo QuestionRepository,
	userRepo domainUser.UserRepository,
	eventBus EventBus,
) *RespondChallengeUseCase {
	return &RespondChallengeUseCase{
		challengeRepo:    challengeRepo,
		duelGameRepo:     duelGameRepo,
		playerRatingRepo: playerRatingRepo,
		seasonRepo:       seasonRepo,
		questionRepo:     questionRepo,
		userRepo:         userRepo,
		eventBus:         eventBus,
	}
}

func (uc *RespondChallengeUseCase) Execute(input RespondChallengeInput) (RespondChallengeOutput, error) {
	now := time.Now().UTC().Unix()

	playerID, err := shared.NewUserID(input.PlayerID)
	if err != nil {
		return RespondChallengeOutput{}, err
	}

	challengeID := quick_duel.NewChallengeIDFromString(input.ChallengeID)

	// Load challenge
	challenge, err := uc.challengeRepo.FindByID(challengeID)
	if err != nil {
		return RespondChallengeOutput{}, err
	}

	if input.Action == "decline" {
		err = challenge.Decline(playerID, now)
		if err != nil {
			return RespondChallengeOutput{}, err
		}

		err = uc.challengeRepo.Save(challenge)
		if err != nil {
			return RespondChallengeOutput{}, err
		}

		for _, event := range challenge.Events() {
			uc.eventBus.Publish(event)
		}

		return RespondChallengeOutput{
			Success:        true,
			TicketConsumed: false,
		}, nil
	}

	// Accept challenge
	err = challenge.Accept(playerID, now)
	if err != nil {
		return RespondChallengeOutput{}, err
	}

	// Publish events
	for _, event := range challenge.Events() {
		uc.eventBus.Publish(event)
	}

	if uc.questionRepo == nil {
		return RespondChallengeOutput{}, fmt.Errorf("question repository unavailable")
	}

	// Create game immediately for direct challenge
	challengerID := challenge.ChallengerID()
	accepterID := playerID

	challengerName := challengerID.String()
	if u, err := uc.userRepo.FindByID(challengerID); err == nil && u != nil {
		if n := u.TelegramUsername().String(); n != "" {
			challengerName = n
		} else if n := u.Username().String(); n != "" {
			challengerName = n
		}
	}
	accepterName := accepterID.String()
	if u, err := uc.userRepo.FindByID(accepterID); err == nil && u != nil {
		if n := u.TelegramUsername().String(); n != "" {
			accepterName = n
		} else if n := u.Username().String(); n != "" {
			accepterName = n
		}
	}

	seasonID, _ := uc.seasonRepo.GetCurrentSeason()
	rating1, err := uc.playerRatingRepo.FindOrCreate(challengerID, seasonID, now)
	if err != nil {
		return RespondChallengeOutput{}, err
	}
	rating2, err := uc.playerRatingRepo.FindOrCreate(accepterID, seasonID, now)
	if err != nil {
		return RespondChallengeOutput{}, err
	}

	questions, err := uc.questionRepo.FindRandomByDifficulty(quick_duel.QuestionsPerDuel, "medium")
	if err != nil {
		return RespondChallengeOutput{}, err
	}

	questionIDs := make([]quick_duel.QuestionID, 0, len(questions))
	for _, q := range questions {
		qid, _ := quiz.NewQuestionIDFromString(q.ID)
		questionIDs = append(questionIDs, qid)
	}

	player1 := quick_duel.NewDuelPlayer(challengerID, challengerName, quick_duel.ReconstructEloRating(rating1.MMR(), 0))
	player2 := quick_duel.NewDuelPlayer(accepterID, accepterName, quick_duel.ReconstructEloRating(rating2.MMR(), 0))

	game, err := quick_duel.NewDuelGame(player1, player2, questionIDs, now)
	if err != nil {
		return RespondChallengeOutput{}, err
	}
	if err := game.Start(now); err != nil {
		return RespondChallengeOutput{}, err
	}
	if err := uc.duelGameRepo.Save(game); err != nil {
		return RespondChallengeOutput{}, err
	}

	challenge.SetMatchID(game.ID())
	if err := uc.challengeRepo.Save(challenge); err != nil {
		return RespondChallengeOutput{}, err
	}

	gameID := game.ID().String()
	return RespondChallengeOutput{
		Success:        true,
		GameID:         &gameID,
		TicketConsumed: true,
	}, nil
}

// ========================================
// TelegramNotifier port (implemented in infrastructure/telegram)
// ========================================

// TelegramNotifier sends Telegram notifications to users.
type TelegramNotifier interface {
	NotifyChallengeAccepted(ctx context.Context, inviterTelegramID int64, inviteeName string, lobbyURL string) error
	NotifyInviterWaiting(ctx context.Context, inviteeTelegramID int64, inviterName string, lobbyURL string) error
	NotifyChallengeReceived(ctx context.Context, inviteeTelegramID int64, inviterName string, deepLink string) (int64, error)
	EditChallengeMessage(ctx context.Context, inviteeTelegramID int64, messageID int64, text string) error
}

// ========================================
// AcceptByLinkCode Use Case
// ========================================

type AcceptByLinkCodeUseCase struct {
	challengeRepo quick_duel.ChallengeRepository
	duelGameRepo  quick_duel.DuelGameRepository // B2: added
	userRepo      domainUser.UserRepository
	notifier      TelegramNotifier
	eventBus      EventBus
}

func NewAcceptByLinkCodeUseCase(
	challengeRepo quick_duel.ChallengeRepository,
	duelGameRepo quick_duel.DuelGameRepository, // B2: added
	userRepo domainUser.UserRepository,
	notifier TelegramNotifier,
	eventBus EventBus,
) *AcceptByLinkCodeUseCase {
	return &AcceptByLinkCodeUseCase{
		challengeRepo: challengeRepo,
		duelGameRepo:  duelGameRepo,
		userRepo:      userRepo,
		notifier:      notifier,
		eventBus:      eventBus,
	}
}

func (uc *AcceptByLinkCodeUseCase) Execute(input AcceptByLinkCodeInput) (AcceptByLinkCodeOutput, error) {
	now := time.Now().UTC().Unix()

	accepterID, err := shared.NewUserID(input.PlayerID)
	if err != nil {
		return AcceptByLinkCodeOutput{}, err
	}

	// Find challenge by link code
	challenge, err := uc.challengeRepo.FindByLinkCode(input.LinkCode)
	if err != nil {
		return AcceptByLinkCodeOutput{}, err
	}

	// Idempotency: if already accepted_waiting_inviter by this player, return success
	if challenge.Status() == quick_duel.ChallengeStatusAcceptedWaitingInviter {
		if challenged := challenge.ChallengedID(); challenged != nil && challenged.Equals(accepterID) {
			// F2: Resolve inviter name for idempotent response
			idempotentInviterName := challenge.ChallengerID().String()
			if u, err := uc.userRepo.FindByID(challenge.ChallengerID()); err == nil && u != nil {
				if n := u.TelegramUsername().String(); n != "" {
					idempotentInviterName = n
				} else if n := u.Username().String(); n != "" {
					idempotentInviterName = n
				}
			}
			return AcceptByLinkCodeOutput{
				Success:     true,
				ChallengeID: challenge.ID().String(),
				Status:      string(quick_duel.ChallengeStatusAcceptedWaitingInviter),
				InviterName: idempotentInviterName,
			}, nil
		}
		return AcceptByLinkCodeOutput{}, quick_duel.ErrChallengeNotPending
	}

	// B2: Check if accepter is already in a game
	if activeGame, err := uc.duelGameRepo.FindActiveByPlayer(accepterID); err == nil && activeGame != nil {
		return AcceptByLinkCodeOutput{}, quick_duel.ErrAlreadyInGame
	}

	// Get accepter's display name
	inviteeName := accepterID.String()
	if u, err := uc.userRepo.FindByID(accepterID); err == nil && u != nil {
		if n := u.TelegramUsername().String(); n != "" {
			inviteeName = n
		} else if n := u.Username().String(); n != "" {
			inviteeName = n
		}
	}

	// Set status to accepted_waiting_inviter
	if err := challenge.AcceptWaiting(accepterID, inviteeName, now); err != nil {
		return AcceptByLinkCodeOutput{}, err
	}

	if err := uc.challengeRepo.Save(challenge); err != nil {
		return AcceptByLinkCodeOutput{}, err
	}

	for _, event := range challenge.Events() {
		uc.eventBus.Publish(event)
	}

	// Notify inviter via Telegram (best-effort — do not fail if notification errors)
	challengerID := challenge.ChallengerID()
	if tgID, err := strconv.ParseInt(challengerID.String(), 10, 64); err == nil && tgID > 0 {
		lobbyURL := "https://t.me/quiz_sprint_dev_bot?startapp=lobby"
		_ = uc.notifier.NotifyChallengeAccepted(context.Background(), tgID, inviteeName, lobbyURL)
	}

	// F2: Resolve inviter's display name
	inviterName := challengerID.String()
	if u, err := uc.userRepo.FindByID(challengerID); err == nil && u != nil {
		if n := u.TelegramUsername().String(); n != "" {
			inviterName = n
		} else if n := u.Username().String(); n != "" {
			inviterName = n
		}
	}

	return AcceptByLinkCodeOutput{
		Success:     true,
		ChallengeID: challenge.ID().String(),
		Status:      string(quick_duel.ChallengeStatusAcceptedWaitingInviter),
		InviterName: inviterName,
	}, nil
}

// ========================================
// CreateChallengeLink Use Case
// ========================================

type CreateChallengeLinkUseCase struct {
	challengeRepo quick_duel.ChallengeRepository
	eventBus      EventBus
}

func NewCreateChallengeLinkUseCase(
	challengeRepo quick_duel.ChallengeRepository,
	eventBus EventBus,
) *CreateChallengeLinkUseCase {
	return &CreateChallengeLinkUseCase{
		challengeRepo: challengeRepo,
		eventBus:      eventBus,
	}
}

func (uc *CreateChallengeLinkUseCase) Execute(input CreateChallengeLinkInput) (CreateChallengeLinkOutput, error) {
	now := time.Now().UTC().Unix()

	playerID, err := shared.NewUserID(input.PlayerID)
	if err != nil {
		return CreateChallengeLinkOutput{}, err
	}

	// Create link challenge
	challenge, err := quick_duel.NewLinkChallenge(playerID, now)
	if err != nil {
		return CreateChallengeLinkOutput{}, err
	}

	// Save challenge
	err = uc.challengeRepo.Save(challenge)
	if err != nil {
		return CreateChallengeLinkOutput{}, err
	}

	// Publish events
	for _, event := range challenge.Events() {
		uc.eventBus.Publish(event)
	}

	return CreateChallengeLinkOutput{
		ChallengeLink: challenge.ChallengeLink(),
		ExpiresAt:     challenge.ExpiresAt(),
		ShareText:     "Вызываю тебя на дуэль! 🎯 Проверим, кто умнее?",
	}, nil
}

// ========================================
// GetGameHistory Use Case
// ========================================

type GetGameHistoryUseCase struct {
	duelGameRepo quick_duel.DuelGameRepository
	userRepo     domainUser.UserRepository
}

func NewGetGameHistoryUseCase(
	duelGameRepo quick_duel.DuelGameRepository,
	userRepo domainUser.UserRepository,
) *GetGameHistoryUseCase {
	return &GetGameHistoryUseCase{
		duelGameRepo: duelGameRepo,
		userRepo:     userRepo,
	}
}

func (uc *GetGameHistoryUseCase) Execute(input GetGameHistoryInput) (GetGameHistoryOutput, error) {
	playerID, err := shared.NewUserID(input.PlayerID)
	if err != nil {
		return GetGameHistoryOutput{}, err
	}

	limit := input.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	games, total, err := uc.duelGameRepo.FindByPlayerPaginated(playerID, limit, input.Offset, input.Filter)
	if err != nil {
		return GetGameHistoryOutput{}, err
	}

	entries := make([]GameHistoryEntryDTO, 0, len(games))
	for _, game := range games {
		// Determine opponent
		opponentID := game.Player2().UserID()
		if game.Player1().UserID().String() != input.PlayerID {
			opponentID = game.Player1().UserID()
		}

		opponentUsername := "Player"
		if user, err := uc.userRepo.FindByID(opponentID); err == nil && user != nil {
			opponentUsername = user.Username().String()
		}

		entries = append(entries, ToGameHistoryEntryDTO(game, input.PlayerID, opponentUsername))
	}

	return GetGameHistoryOutput{
		Games:   entries,
		Total:   total,
		HasMore: input.Offset+len(games) < total,
	}, nil
}

// ========================================
// GetLeaderboard Use Case
// ========================================

type GetLeaderboardUseCase struct {
	playerRatingRepo quick_duel.PlayerRatingRepository
	referralRepo     quick_duel.ReferralRepository
	seasonRepo       quick_duel.SeasonRepository
	userRepo         domainUser.UserRepository
}

func NewGetLeaderboardUseCase(
	playerRatingRepo quick_duel.PlayerRatingRepository,
	referralRepo quick_duel.ReferralRepository,
	seasonRepo quick_duel.SeasonRepository,
	userRepo domainUser.UserRepository,
) *GetLeaderboardUseCase {
	return &GetLeaderboardUseCase{
		playerRatingRepo: playerRatingRepo,
		referralRepo:     referralRepo,
		seasonRepo:       seasonRepo,
		userRepo:         userRepo,
	}
}

func (uc *GetLeaderboardUseCase) Execute(input GetLeaderboardInput) (GetLeaderboardOutput, error) {
	playerID, err := shared.NewUserID(input.PlayerID)
	if err != nil {
		return GetLeaderboardOutput{}, err
	}

	limit := input.Limit
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	seasonID, _ := uc.seasonRepo.GetCurrentSeason()
	_, seasonEndsAt, _ := uc.seasonRepo.GetSeasonInfo(seasonID)

	var entries []LeaderboardEntryDTO
	var playerRank int

	switch input.Type {
	case "seasonal":
		ratings, err := uc.playerRatingRepo.GetLeaderboard(seasonID, limit, 0)
		if err != nil {
			return GetLeaderboardOutput{}, err
		}

		entries = make([]LeaderboardEntryDTO, 0, len(ratings))
		for i, rating := range ratings {
			username := "Player"
			if user, err := uc.userRepo.FindByID(rating.PlayerID()); err == nil && user != nil {
				username = user.Username().String()
			}
			entries = append(entries, ToLeaderboardEntryDTO(rating, i+1, username))
		}

		playerRank, _ = uc.playerRatingRepo.GetPlayerRank(playerID, seasonID)

	case "referrals":
		refEntries, err := uc.referralRepo.GetReferralLeaderboard(limit)
		if err != nil {
			return GetLeaderboardOutput{}, err
		}

		entries = make([]LeaderboardEntryDTO, 0, len(refEntries))
		for i, entry := range refEntries {
			entries = append(entries, LeaderboardEntryDTO{
				Rank:     i + 1,
				PlayerID: entry.PlayerID.String(),
				Username: entry.Username,
				Wins:     entry.TotalReferrals,
			})
		}

		playerRank, _ = uc.referralRepo.GetPlayerReferralRank(playerID)

	default: // "friends"
		// TODO: implement friends leaderboard
		entries = []LeaderboardEntryDTO{}
		playerRank = 0
	}

	return GetLeaderboardOutput{
		Type:       input.Type,
		SeasonID:   seasonID,
		EndsAt:     seasonEndsAt,
		Entries:    entries,
		PlayerRank: playerRank,
	}, nil
}

// ========================================
// StartGame Use Case
// ========================================

type StartGameUseCase struct {
	duelGameRepo     quick_duel.DuelGameRepository
	playerRatingRepo quick_duel.PlayerRatingRepository
	questionRepo     QuestionRepository
	seasonRepo       quick_duel.SeasonRepository
	eventBus         EventBus
}

// QuestionRepository interface for getting questions
type QuestionRepository interface {
	FindRandomByDifficulty(count int, difficulty string) ([]QuestionData, error)
	// FindByID retrieves a single question with all answers by its ID.
	// Used by SubmitDuelAnswerUseCase to validate answer correctness.
	FindByID(questionID quiz.QuestionID) (*quiz.Question, error)
}

// QuestionData represents question data for duels
type QuestionData struct {
	ID      string
	Text    string
	Answers []AnswerData
}

// AnswerData represents answer data
type AnswerData struct {
	ID        string
	Text      string
	IsCorrect bool
}

func NewStartGameUseCase(
	duelGameRepo quick_duel.DuelGameRepository,
	playerRatingRepo quick_duel.PlayerRatingRepository,
	questionRepo QuestionRepository,
	seasonRepo quick_duel.SeasonRepository,
	eventBus EventBus,
) *StartGameUseCase {
	return &StartGameUseCase{
		duelGameRepo:     duelGameRepo,
		playerRatingRepo: playerRatingRepo,
		questionRepo:     questionRepo,
		seasonRepo:       seasonRepo,
		eventBus:         eventBus,
	}
}

func (uc *StartGameUseCase) Execute(input StartGameInput) (StartGameOutput, error) {
	now := time.Now().UTC().Unix()

	player1ID, err := shared.NewUserID(input.Player1ID)
	if err != nil {
		return StartGameOutput{}, err
	}

	player2ID, err := shared.NewUserID(input.Player2ID)
	if err != nil {
		return StartGameOutput{}, err
	}

	// Get player ratings
	seasonID, _ := uc.seasonRepo.GetCurrentSeason()
	rating1, err := uc.playerRatingRepo.FindOrCreate(player1ID, seasonID, now)
	if err != nil {
		return StartGameOutput{}, err
	}

	rating2, err := uc.playerRatingRepo.FindOrCreate(player2ID, seasonID, now)
	if err != nil {
		return StartGameOutput{}, err
	}

	// Get random questions for the duel
	questions, err := uc.questionRepo.FindRandomByDifficulty(quick_duel.QuestionsPerDuel, "medium")
	if err != nil {
		return StartGameOutput{}, err
	}

	// Create players
	player1 := quick_duel.NewDuelPlayer(
		player1ID,
		input.Player1Username,
		quick_duel.ReconstructEloRating(rating1.MMR(), 0),
	)

	player2 := quick_duel.NewDuelPlayer(
		player2ID,
		input.Player2Username,
		quick_duel.ReconstructEloRating(rating2.MMR(), 0),
	)

	// Convert question IDs
	questionIDs := make([]quick_duel.QuestionID, 0, len(questions))
	for _, q := range questions {
		qid, _ := quiz.NewQuestionIDFromString(q.ID)
		questionIDs = append(questionIDs, qid)
	}

	// Create duel game
	game, err := quick_duel.NewDuelGame(player1, player2, questionIDs, now)
	if err != nil {
		return StartGameOutput{}, err
	}
	game.Start(now)

	// Save game
	err = uc.duelGameRepo.Save(game)
	if err != nil {
		return StartGameOutput{}, err
	}

	return StartGameOutput{
		GameID:    game.ID().String(),
		Player1ID: player1ID.String(),
		Player2ID: player2ID.String(),
		StartsAt:  now + 3, // 3 second countdown
	}, nil
}

// GetRoundQuestion returns the question for a specific round
func (uc *StartGameUseCase) GetRoundQuestion(gameIDStr string, roundNum int) (*RoundQuestionOutput, error) {
	gameID := quick_duel.NewGameIDFromString(gameIDStr)
	game, err := uc.duelGameRepo.FindByID(gameID)
	if err != nil {
		return nil, err
	}

	questionIDs := game.QuestionIDs()
	if roundNum < 1 || roundNum > len(questionIDs) {
		return nil, quick_duel.ErrGameNotFound
	}

	questionID := questionIDs[roundNum-1]
	question, err := uc.questionRepo.FindByID(questionID)
	if err != nil {
		return nil, fmt.Errorf("get round question: load question %s: %w", questionID, err)
	}

	answers := make([]map[string]string, 0, len(question.Answers()))
	for _, a := range question.Answers() {
		answers = append(answers, map[string]string{
			"id":   a.ID().String(),
			"text": a.Text().String(),
		})
	}

	return &RoundQuestionOutput{
		QuestionID:   questionID.String(),
		QuestionText: question.Text().String(),
		Answers:      answers,
	}, nil
}

// GetDomainPlayerOrder returns the domain's canonical player1ID and player2ID
// for a given game. This is the authoritative order (set at game creation):
// Player1 = challenger, Player2 = accepter.
// Used by the WS hub to send consistent player IDs in game_ready messages.
func (uc *StartGameUseCase) GetDomainPlayerOrder(gameIDStr string) (player1ID, player2ID string, err error) {
	gameID := quick_duel.NewGameIDFromString(gameIDStr)
	game, err := uc.duelGameRepo.FindByID(gameID)
	if err != nil {
		return "", "", err
	}
	return game.Player1().UserID().String(), game.Player2().UserID().String(), nil
}

// ========================================
// SubmitDuelAnswer Use Case
// ========================================

type SubmitDuelAnswerUseCase struct {
	duelGameRepo     quick_duel.DuelGameRepository
	playerRatingRepo quick_duel.PlayerRatingRepository
	questionRepo     QuestionRepository
	seasonRepo       quick_duel.SeasonRepository
	eventBus         EventBus
	roundCache       DuelRoundCache
}

// PlayerAnswer holds a single player's answer for one duel round.
// Exported so it can be stored and retrieved by DuelRoundCache implementations.
type PlayerAnswer struct {
	PlayerID  string `json:"player_id"`
	AnswerID  string `json:"answer_id"`
	IsCorrect bool   `json:"is_correct"`
	TimeTaken int    `json:"time_taken"`
	Points    int    `json:"points"`
}

func NewSubmitDuelAnswerUseCase(
	duelGameRepo quick_duel.DuelGameRepository,
	playerRatingRepo quick_duel.PlayerRatingRepository,
	questionRepo QuestionRepository,
	seasonRepo quick_duel.SeasonRepository,
	eventBus EventBus,
	roundCache DuelRoundCache,
) *SubmitDuelAnswerUseCase {
	return &SubmitDuelAnswerUseCase{
		duelGameRepo:     duelGameRepo,
		playerRatingRepo: playerRatingRepo,
		questionRepo:     questionRepo,
		seasonRepo:       seasonRepo,
		eventBus:         eventBus,
		roundCache:       roundCache,
	}
}

func (uc *SubmitDuelAnswerUseCase) Execute(input SubmitDuelAnswerInput) (*SubmitDuelAnswerOutput, error) {
	if uc.questionRepo == nil {
		return nil, fmt.Errorf("submit duel answer: question repository not configured")
	}
	if uc.roundCache == nil {
		return nil, fmt.Errorf("submit duel answer: round cache not configured")
	}

	now := time.Now().UTC().Unix()

	gameID := quick_duel.NewGameIDFromString(input.GameID)
	game, err := uc.duelGameRepo.FindByID(gameID)
	if err != nil {
		return nil, err
	}

	// Check if game is in progress
	if game.Status() != quick_duel.GameStatusInProgress {
		return nil, quick_duel.ErrGameNotFound
	}

	// Validate player is in the game
	isPlayer1 := game.Player1().UserID().String() == input.PlayerID
	isPlayer2 := game.Player2().UserID().String() == input.PlayerID
	if !isPlayer1 && !isPlayer2 {
		return nil, quick_duel.ErrGameNotFound
	}

	// Determine the current round's question ID and look it up
	currentRound := game.CurrentRound()
	questionIDs := game.QuestionIDs()
	if currentRound < 1 || currentRound > len(questionIDs) {
		return nil, fmt.Errorf("submit duel answer: invalid round %d", currentRound)
	}
	currentQuestionID := questionIDs[currentRound-1]

	question, err := uc.questionRepo.FindByID(currentQuestionID)
	if err != nil {
		return nil, fmt.Errorf("submit duel answer: load question: %w", err)
	}

	// Parse player and answer IDs
	playerID, err := shared.NewUserID(input.PlayerID)
	if err != nil {
		return nil, err
	}
	answerID, err := quiz.NewAnswerIDFromString(input.AnswerID)
	if err != nil {
		return nil, err
	}

	// Delegate correctness check and scoring to the domain aggregate
	result, err := game.SubmitAnswer(playerID, answerID, int64(input.TimeTaken), question, now)
	if err != nil {
		return nil, err
	}

	isCorrect := result.IsCorrect
	points := result.PointsEarned

	// Determine the correct answer ID to return to the client
	var correctAnswerID string
	for _, a := range question.Answers() {
		if a.IsCorrect() {
			correctAnswerID = a.ID().String()
			break
		}
	}

	// Track this answer in the distributed cache
	if err := uc.roundCache.AddAnswer(input.GameID, currentRound, PlayerAnswer{
		PlayerID:  input.PlayerID,
		AnswerID:  input.AnswerID,
		IsCorrect: isCorrect,
		TimeTaken: input.TimeTaken,
		Points:    points,
	}); err != nil {
		return nil, fmt.Errorf("submit duel answer: cache answer: %w", err)
	}

	// Check if round is complete (both players answered)
	roundAnswers, err := uc.roundCache.GetAnswers(input.GameID, currentRound)
	if err != nil {
		return nil, fmt.Errorf("submit duel answer: get round answers: %w", err)
	}
	roundComplete := len(roundAnswers) >= 2

	// roundAnswers are not persisted to the DB (see TODO in reconstructGame), so the
	// domain's bothAnswered check is always false after reconstruction. When the cache
	// confirms both players answered, force completeRound() via RecordTimeoutAnswer for
	// the opponent — this advances currentRound in the DB so the next answer hits the
	// correct question.
	if roundComplete && !result.BothAnswered {
		var opponentID quick_duel.UserID
		if game.Player1().UserID().String() == input.PlayerID {
			opponentID = game.Player2().UserID()
		} else {
			opponentID = game.Player1().UserID()
		}
		if timeoutResult, _ := game.RecordTimeoutAnswer(opponentID, now); timeoutResult != nil && timeoutResult.IsGameFinished {
			result.IsGameFinished = true
		}
	}

	// Scores are already maintained by the domain aggregate
	player1Score := game.Player1().Score()
	player2Score := game.Player2().Score()

	// Game is complete when the domain aggregate reports so (last round, both answered)
	gameComplete := result.IsGameFinished

	output := &SubmitDuelAnswerOutput{
		IsCorrect:       isCorrect,
		CorrectAnswerID: correctAnswerID,
		PointsEarned:    points,
		Player1Score:    player1Score,
		Player2Score:    player2Score,
		RoundComplete:   roundComplete,
		GameComplete:    gameComplete,
	}

	// Save updated game state (covers both mid-game and final-round persistence).
	// finalizeGame will save again after applying MMR changes.
	if !gameComplete {
		if err := uc.duelGameRepo.Save(game); err != nil {
			return nil, fmt.Errorf("submit duel answer: save game: %w", err)
		}
	}

	// If game is complete, apply MMR changes and save the finalised game
	if gameComplete {
		uc.finalizeGame(game, player1Score, player2Score, now, output)
	}

	return output, nil
}

func (uc *SubmitDuelAnswerUseCase) finalizeGame(
	game *quick_duel.DuelGame,
	player1Score, player2Score int,
	now int64,
	output *SubmitDuelAnswerOutput,
) {
	// Determine winner
	var winnerID string
	var player1Won, player2Won bool

	if player1Score > player2Score {
		winnerID = game.Player1().UserID().String()
		player1Won = true
		player2Won = false
	} else if player2Score > player1Score {
		winnerID = game.Player2().UserID().String()
		player1Won = false
		player2Won = true
	} else {
		// Tie - treat as draw (neither won)
		player1Won = false
		player2Won = false
	}

	output.WinnerID = winnerID

	// Get current season
	seasonID, _ := uc.seasonRepo.GetCurrentSeason()

	// Update player 1 rating
	rating1, err := uc.playerRatingRepo.FindOrCreate(game.Player1().UserID(), seasonID, now)
	if err == nil {
		oldMMR1 := rating1.MMR()
		rating1.ApplyGameResult(quick_duel.GameResult{
			Won:         player1Won,
			OpponentMMR: game.Player2().Elo().Rating(),
			GameTime:    now,
		})
		uc.playerRatingRepo.Save(rating1)
		output.Player1MMRChange = rating1.MMR() - oldMMR1
		output.Player1NewMMR = rating1.MMR()
	}

	// Update player 2 rating
	rating2, err := uc.playerRatingRepo.FindOrCreate(game.Player2().UserID(), seasonID, now)
	if err == nil {
		oldMMR2 := rating2.MMR()
		rating2.ApplyGameResult(quick_duel.GameResult{
			Won:         player2Won,
			OpponentMMR: game.Player1().Elo().Rating(),
			GameTime:    now,
		})
		uc.playerRatingRepo.Save(rating2)
		output.Player2MMRChange = rating2.MMR() - oldMMR2
		output.Player2NewMMR = rating2.MMR()
	}

	// Save game (domain already updated status internally)
	uc.duelGameRepo.Save(game)

	// Clean up cached round answers — best-effort, ignore errors
	_ = uc.roundCache.DeleteGame(game.ID().String())
}

// TimeoutRound submits timeout answers for any players who have not yet answered the given round.
// Advances the domain's currentRound so subsequent answers are validated against the right question.
// Returns nil output (no error) when the round already advanced or the game is not in progress.
func (uc *SubmitDuelAnswerUseCase) TimeoutRound(gameIDStr string, roundNum int) (*SubmitDuelAnswerOutput, error) {
	now := time.Now().UTC().Unix()

	gameID := quick_duel.NewGameIDFromString(gameIDStr)
	game, err := uc.duelGameRepo.FindByID(gameID)
	if err != nil {
		return nil, err
	}

	if game.Status() != quick_duel.GameStatusInProgress {
		return nil, nil // Game already finished — nothing to do
	}
	if game.CurrentRound() != roundNum {
		return nil, nil // Round already advanced — nothing to do
	}

	// Record timeout for each player who hasn't answered yet
	var lastResult *quick_duel.SubmitAnswerResult
	for _, playerID := range []quick_duel.UserID{game.Player1().UserID(), game.Player2().UserID()} {
		result, err := game.RecordTimeoutAnswer(playerID, now)
		if err != nil {
			if isErr(err, quick_duel.ErrPlayerAlreadyAnswered) {
				continue
			}
			return nil, fmt.Errorf("timeout round %d player %s: %w", roundNum, playerID, err)
		}

		_ = uc.roundCache.AddAnswer(gameIDStr, roundNum, PlayerAnswer{
			PlayerID:  playerID.String(),
			AnswerID:  "",
			IsCorrect: false,
			TimeTaken: quick_duel.TimePerQuestionSec * 1000,
			Points:    0,
		})

		lastResult = result
	}

	if lastResult == nil {
		return nil, nil // Both players already answered before timeout fired
	}

	player1Score := game.Player1().Score()
	player2Score := game.Player2().Score()

	output := &SubmitDuelAnswerOutput{
		IsCorrect:     false,
		Player1Score:  player1Score,
		Player2Score:  player2Score,
		RoundComplete: lastResult.BothAnswered,
		GameComplete:  lastResult.IsGameFinished,
	}

	if lastResult.IsGameFinished {
		uc.finalizeGame(game, player1Score, player2Score, now, output)
	} else if err := uc.duelGameRepo.Save(game); err != nil {
		return nil, fmt.Errorf("timeout round: save game: %w", err)
	}

	return output, nil
}

// ========================================
// GetGameResult Use Case
// ========================================

type GetGameResultUseCase struct {
	duelGameRepo     quick_duel.DuelGameRepository
	playerRatingRepo quick_duel.PlayerRatingRepository
	seasonRepo       quick_duel.SeasonRepository
	userRepo         domainUser.UserRepository
}

func NewGetGameResultUseCase(
	duelGameRepo quick_duel.DuelGameRepository,
	playerRatingRepo quick_duel.PlayerRatingRepository,
	seasonRepo quick_duel.SeasonRepository,
	userRepo domainUser.UserRepository,
) *GetGameResultUseCase {
	return &GetGameResultUseCase{
		duelGameRepo:     duelGameRepo,
		playerRatingRepo: playerRatingRepo,
		seasonRepo:       seasonRepo,
		userRepo:         userRepo,
	}
}

func (uc *GetGameResultUseCase) Execute(input GetGameResultInput) (*GetGameResultOutput, error) {
	now := time.Now().UTC().Unix()

	playerID, err := shared.NewUserID(input.PlayerID)
	if err != nil {
		return nil, err
	}

	gameID := quick_duel.NewGameIDFromString(input.GameID)
	game, err := uc.duelGameRepo.FindByID(gameID)
	if err != nil {
		return nil, err
	}

	isPlayer1 := game.Player1().UserID().String() == input.PlayerID
	isPlayer2 := game.Player2().UserID().String() == input.PlayerID
	if !isPlayer1 && !isPlayer2 {
		return nil, quick_duel.ErrGameNotFound
	}

	var playerScore, opponentScore, mmrBefore int
	var opponentPlayer quick_duel.DuelPlayer
	if isPlayer1 {
		playerScore = game.Player1().Score()
		opponentScore = game.Player2().Score()
		mmrBefore = game.Player1().Elo().Rating()
		opponentPlayer = game.Player2()
	} else {
		playerScore = game.Player2().Score()
		opponentScore = game.Player1().Score()
		mmrBefore = game.Player2().Elo().Rating()
		opponentPlayer = game.Player1()
	}

	var result string
	switch {
	case playerScore > opponentScore:
		result = "win"
	case playerScore < opponentScore:
		result = "loss"
	default:
		result = "draw"
	}

	seasonID, _ := uc.seasonRepo.GetCurrentSeason()
	rating, err := uc.playerRatingRepo.FindOrCreate(playerID, seasonID, now)
	if err != nil {
		return nil, err
	}

	mmrChange := rating.MMR() - mmrBefore

	// Get opponent info
	opponentUsername := opponentPlayer.Username()
	if user, err := uc.userRepo.FindByID(opponentPlayer.UserID()); err == nil && user != nil {
		opponentUsername = user.Username().String()
	}

	opponentRating, _ := uc.playerRatingRepo.FindOrCreate(opponentPlayer.UserID(), seasonID, now)
	opponentDTO := ToDuelPlayerDTO(opponentPlayer, opponentRating)
	opponentDTO.Username = opponentUsername

	return &GetGameResultOutput{
		GameID:           input.GameID,
		Result:           result,
		PlayerScore:      playerScore,
		OpponentScore:    opponentScore,
		MMRChange:        mmrChange,
		NewMMR:           rating.MMR(),
		NewLeague:        rating.League().String(),
		NewDivision:      rating.Division().Value(),
		Opponent:         opponentDTO,
		Questions:        []GameQuestionResultDTO{},
		CanRematch:       true,
		RematchExpiresIn: nil,
	}, nil
}

// isErr reports whether err or any of its unwrapped causes equals target.
func isErr(err, target error) bool {
	if err == target {
		return true
	}
	if u, ok := err.(interface{ Unwrap() error }); ok {
		return isErr(u.Unwrap(), target)
	}
	return false
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ========================================
// RequestRematch Use Case
// ========================================

type RequestRematchUseCase struct {
	duelGameRepo  quick_duel.DuelGameRepository
	challengeRepo quick_duel.ChallengeRepository
	eventBus      EventBus
}

func NewRequestRematchUseCase(
	duelGameRepo quick_duel.DuelGameRepository,
	challengeRepo quick_duel.ChallengeRepository,
	eventBus EventBus,
) *RequestRematchUseCase {
	return &RequestRematchUseCase{
		duelGameRepo:  duelGameRepo,
		challengeRepo: challengeRepo,
		eventBus:      eventBus,
	}
}

func (uc *RequestRematchUseCase) Execute(input RequestRematchInput) (RequestRematchOutput, error) {
	now := time.Now().UTC().Unix()

	playerID, err := shared.NewUserID(input.PlayerID)
	if err != nil {
		return RequestRematchOutput{}, err
	}

	// Get the original game
	gameID := quick_duel.NewGameIDFromString(input.GameID)
	game, err := uc.duelGameRepo.FindByID(gameID)
	if err != nil {
		return RequestRematchOutput{}, err
	}

	// Verify player was in the game
	isPlayer1 := game.Player1().UserID().String() == input.PlayerID
	isPlayer2 := game.Player2().UserID().String() == input.PlayerID
	if !isPlayer1 && !isPlayer2 {
		return RequestRematchOutput{}, quick_duel.ErrGameNotFound
	}

	// Verify game is finished
	if game.Status() != quick_duel.GameStatusFinished {
		return RequestRematchOutput{}, quick_duel.ErrGameNotActive
	}

	// Determine opponent
	var opponentID quick_duel.UserID
	if isPlayer1 {
		opponentID = game.Player2().UserID()
	} else {
		opponentID = game.Player1().UserID()
	}

	// Check if opponent already requested rematch (auto-accept)
	existingChallenges, err := uc.challengeRepo.FindPendingByChallenger(opponentID)
	if err == nil {
		for _, c := range existingChallenges {
			// If opponent sent a rematch challenge to this player
			if c.ChallengedID() != nil && c.ChallengedID().String() == input.PlayerID {
				// Auto-accept the rematch
				c.Accept(playerID, now)
				uc.challengeRepo.Save(c)

				for _, event := range c.Events() {
					uc.eventBus.Publish(event)
				}

				return RequestRematchOutput{
					RematchID: c.ID().String(),
					Status:    "accepted",
					ExpiresIn: 0,
					GameID:    nil, // Would be set when game starts
				}, nil
			}
		}
	}

	// Create new rematch challenge
	challenge, err := quick_duel.NewDirectChallenge(playerID, opponentID, now)
	if err != nil {
		return RequestRematchOutput{}, err
	}

	// Save challenge
	err = uc.challengeRepo.Save(challenge)
	if err != nil {
		return RequestRematchOutput{}, err
	}

	// Publish events
	for _, event := range challenge.Events() {
		uc.eventBus.Publish(event)
	}

	return RequestRematchOutput{
		RematchID: challenge.ID().String(),
		Status:    "pending",
		ExpiresIn: quick_duel.DirectChallengeExpirySeconds,
	}, nil
}

// ========================================
// GetOnlineFriends Use Case
// ========================================

type GetOnlineFriendsUseCase struct {
	onlineTracker OnlineTracker
	userRepo      domainUser.UserRepository
}

// OnlineTracker interface for tracking online status
type OnlineTracker interface {
	SetOnline(playerID string, expiresInSeconds int) error
	IsOnline(playerID string) (bool, error)
	GetOnlineFriends(playerID string, friendIDs []string) ([]string, error)
	SetInGame(playerID string, gameID string) error
	ClearInGame(playerID string) error
	GetGameID(playerID string) (string, error)
}

func NewGetOnlineFriendsUseCase(
	onlineTracker OnlineTracker,
	userRepo domainUser.UserRepository,
) *GetOnlineFriendsUseCase {
	return &GetOnlineFriendsUseCase{
		onlineTracker: onlineTracker,
		userRepo:      userRepo,
	}
}

type GetOnlineFriendsInput struct {
	PlayerID  string   `json:"playerId"`
	FriendIDs []string `json:"friendIds"`
}

type GetOnlineFriendsOutput struct {
	OnlineFriends []FriendDTO `json:"onlineFriends"`
}

func (uc *GetOnlineFriendsUseCase) Execute(input GetOnlineFriendsInput) (GetOnlineFriendsOutput, error) {
	// Get online status for all friends
	onlineIDs, err := uc.onlineTracker.GetOnlineFriends(input.PlayerID, input.FriendIDs)
	if err != nil {
		return GetOnlineFriendsOutput{}, err
	}

	friends := make([]FriendDTO, 0, len(onlineIDs))
	for _, friendID := range onlineIDs {
		uid, err := shared.NewUserID(friendID)
		if err != nil {
			continue
		}

		user, err := uc.userRepo.FindByID(uid)
		if err != nil {
			continue
		}

		// Check if in game
		gameID, _ := uc.onlineTracker.GetGameID(friendID)
		inGame := gameID != ""

		friends = append(friends, FriendDTO{
			ID:       friendID,
			Username: user.Username().String(),
			IsOnline: true,
			InGame:   inGame,
		})
	}

	return GetOnlineFriendsOutput{
		OnlineFriends: friends,
	}, nil
}

// ========================================
// StartChallenge Use Case
// ========================================

type StartChallengeUseCase struct {
	challengeRepo    quick_duel.ChallengeRepository
	duelGameRepo     quick_duel.DuelGameRepository
	playerRatingRepo quick_duel.PlayerRatingRepository
	seasonRepo       quick_duel.SeasonRepository
	questionRepo     QuestionRepository
	userRepo         domainUser.UserRepository
	eventBus         EventBus
}

func NewStartChallengeUseCase(
	challengeRepo quick_duel.ChallengeRepository,
	duelGameRepo quick_duel.DuelGameRepository,
	playerRatingRepo quick_duel.PlayerRatingRepository,
	seasonRepo quick_duel.SeasonRepository,
	questionRepo QuestionRepository,
	userRepo domainUser.UserRepository,
	eventBus EventBus,
) *StartChallengeUseCase {
	return &StartChallengeUseCase{
		challengeRepo:    challengeRepo,
		duelGameRepo:     duelGameRepo,
		playerRatingRepo: playerRatingRepo,
		seasonRepo:       seasonRepo,
		questionRepo:     questionRepo,
		userRepo:         userRepo,
		eventBus:         eventBus,
	}
}

func (uc *StartChallengeUseCase) Execute(input StartChallengeInput) (StartChallengeOutput, error) {
	now := time.Now().UTC().Unix()

	inviterID, err := shared.NewUserID(input.PlayerID)
	if err != nil {
		return StartChallengeOutput{}, err
	}

	challengeID := quick_duel.NewChallengeIDFromString(input.ChallengeID)
	challenge, err := uc.challengeRepo.FindByID(challengeID)
	if err != nil {
		return StartChallengeOutput{}, err
	}

	// Validate inviter is the challenger
	if !challenge.ChallengerID().Equals(inviterID) {
		return StartChallengeOutput{}, quick_duel.ErrNotChallengedPlayer
	}

	// Validate status
	if challenge.Status() != quick_duel.ChallengeStatusAcceptedWaitingInviter {
		return StartChallengeOutput{}, quick_duel.ErrChallengeNotPending
	}

	if challenge.ChallengedID() == nil {
		return StartChallengeOutput{}, quick_duel.ErrChallengeNotFound
	}

	accepterID := *challenge.ChallengedID()

	// B1: Guard — inviter must not be in an active game
	if active, err := uc.duelGameRepo.FindActiveByPlayer(inviterID); err == nil && active != nil {
		return StartChallengeOutput{}, quick_duel.ErrAlreadyInGame
	}
	// B1: Guard — invitee must not be in an active game
	if active, err := uc.duelGameRepo.FindActiveByPlayer(accepterID); err == nil && active != nil {
		return StartChallengeOutput{}, quick_duel.ErrAlreadyInGame
	}

	seasonID, _ := uc.seasonRepo.GetCurrentSeason()

	rating1, err := uc.playerRatingRepo.FindOrCreate(inviterID, seasonID, now)
	if err != nil {
		return StartChallengeOutput{}, err
	}
	rating2, err := uc.playerRatingRepo.FindOrCreate(accepterID, seasonID, now)
	if err != nil {
		return StartChallengeOutput{}, err
	}

	// Get usernames
	inviterName := inviterID.String()
	if u, err := uc.userRepo.FindByID(inviterID); err == nil && u != nil {
		if n := u.TelegramUsername().String(); n != "" {
			inviterName = n
		} else if n := u.Username().String(); n != "" {
			inviterName = n
		}
	}
	accepterName := accepterID.String()
	if u, err := uc.userRepo.FindByID(accepterID); err == nil && u != nil {
		if n := u.TelegramUsername().String(); n != "" {
			accepterName = n
		} else if n := u.Username().String(); n != "" {
			accepterName = n
		}
	}

	// Select random questions
	questions, err := uc.questionRepo.FindRandomByDifficulty(quick_duel.QuestionsPerDuel, "medium")
	if err != nil {
		return StartChallengeOutput{}, err
	}

	questionIDs := make([]quick_duel.QuestionID, 0, len(questions))
	for _, q := range questions {
		qid, _ := quiz.NewQuestionIDFromString(q.ID)
		questionIDs = append(questionIDs, qid)
	}

	player1 := quick_duel.NewDuelPlayer(inviterID, inviterName, quick_duel.ReconstructEloRating(rating1.MMR(), 0))
	player2 := quick_duel.NewDuelPlayer(accepterID, accepterName, quick_duel.ReconstructEloRating(rating2.MMR(), 0))

	game, err := quick_duel.NewDuelGame(player1, player2, questionIDs, now)
	if err != nil {
		return StartChallengeOutput{}, err
	}
	if err := game.Start(now); err != nil {
		return StartChallengeOutput{}, err
	}
	if err := uc.duelGameRepo.Save(game); err != nil {
		return StartChallengeOutput{}, err
	}

	// Transition challenge to accepted — removes it from lobby outgoing cards
	if err := challenge.MarkStarted(game.ID()); err != nil {
		return StartChallengeOutput{}, err
	}
	if err := uc.challengeRepo.Save(challenge); err != nil {
		return StartChallengeOutput{}, err
	}

	return StartChallengeOutput{GameID: game.ID().String()}, nil
}

// ========================================
// GetRivals Use Case
// ========================================

type GetRivalsUseCase struct {
	duelGameRepo     quick_duel.DuelGameRepository
	playerRatingRepo quick_duel.PlayerRatingRepository
	userRepo         domainUser.UserRepository
	onlineTracker    OnlineTracker
	challengeRepo    quick_duel.ChallengeRepository
}

func NewGetRivalsUseCase(
	duelGameRepo quick_duel.DuelGameRepository,
	playerRatingRepo quick_duel.PlayerRatingRepository,
	userRepo domainUser.UserRepository,
	onlineTracker OnlineTracker,
	challengeRepo quick_duel.ChallengeRepository,
) *GetRivalsUseCase {
	return &GetRivalsUseCase{
		duelGameRepo:     duelGameRepo,
		playerRatingRepo: playerRatingRepo,
		userRepo:         userRepo,
		onlineTracker:    onlineTracker,
		challengeRepo:    challengeRepo,
	}
}

func (uc *GetRivalsUseCase) Execute(input GetRivalsInput) (GetRivalsOutput, error) {
	playerID, err := shared.NewUserID(input.PlayerID)
	if err != nil {
		return GetRivalsOutput{}, err
	}

	const limit = 20
	opponents, err := uc.duelGameRepo.FindRecentOpponents(playerID, limit)
	if err != nil {
		return GetRivalsOutput{Rivals: []RivalDTO{}}, nil
	}

	// Build set of rival IDs that already have a pending challenge from this player
	pendingChallengedIDs := map[string]bool{}
	if pending, err := uc.challengeRepo.FindPendingByChallenger(playerID); err == nil {
		for _, c := range pending {
			if c.ChallengedID() != nil {
				pendingChallengedIDs[c.ChallengedID().String()] = true
			}
		}
	}

	rivals := make([]RivalDTO, 0, len(opponents))
	for _, opp := range opponents {
		user, err := uc.userRepo.FindByID(opp.OpponentID)
		if err != nil {
			continue
		}

		mmr := 1000
		leagueStr := "bronze"
		leagueIcon := "🥉"
		if rating, err := uc.playerRatingRepo.FindByPlayerID(opp.OpponentID); err == nil && rating != nil {
			mmr = rating.MMR()
			leagueStr = rating.League().String()
			leagueIcon = rating.League().Icon()
		}

		isOnline := false
		if uc.onlineTracker != nil {
			isOnline, _ = uc.onlineTracker.IsOnline(opp.OpponentID.String())
		}

		rivals = append(rivals, RivalDTO{
			ID:                  opp.OpponentID.String(),
			Username:            user.Username().String(),
			MMR:                 mmr,
			League:              leagueStr,
			LeagueIcon:          leagueIcon,
			IsOnline:            isOnline,
			GamesCount:          opp.GamesCount,
			HasPendingChallenge: pendingChallengedIDs[opp.OpponentID.String()],
		})
	}

	return GetRivalsOutput{Rivals: rivals}, nil
}
