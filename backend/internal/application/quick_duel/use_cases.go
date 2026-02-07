package quick_duel

import (
	"time"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quick_duel"
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
}

func NewGetDuelStatusUseCase(
	playerRatingRepo quick_duel.PlayerRatingRepository,
	duelGameRepo quick_duel.DuelGameRepository,
	challengeRepo quick_duel.ChallengeRepository,
	seasonRepo quick_duel.SeasonRepository,
	userRepo domainUser.UserRepository,
) *GetDuelStatusUseCase {
	return &GetDuelStatusUseCase{
		playerRatingRepo: playerRatingRepo,
		duelGameRepo:     duelGameRepo,
		challengeRepo:    challengeRepo,
		seasonRepo:       seasonRepo,
		userRepo:         userRepo,
	}
}

func (uc *GetDuelStatusUseCase) Execute(input GetDuelStatusInput) (GetDuelStatusOutput, error) {
	now := time.Now().UTC().Unix()

	playerID, err := shared.NewUserID(input.PlayerID)
	if err != nil {
		return GetDuelStatusOutput{}, err
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

	// Check for active match
	var activeMatchID *string
	activeMatch, err := uc.duelGameRepo.FindActiveByPlayer(playerID)
	if err == nil && activeMatch != nil {
		id := activeMatch.ID().String()
		activeMatchID = &id
	}

	// Get pending challenges
	pendingChallenges, err := uc.challengeRepo.FindPendingForPlayer(playerID)
	if err != nil {
		pendingChallenges = []*quick_duel.DuelChallenge{}
	}

	challengeDTOs := make([]ChallengeDTO, 0, len(pendingChallenges))
	for _, c := range pendingChallenges {
		challengeDTOs = append(challengeDTOs, ToChallengeDTO(c, now))
	}

	// Get season end time
	_, seasonEndsAt, _ := uc.seasonRepo.GetSeasonInfo(seasonID)

	return GetDuelStatusOutput{
		HasActiveDuel:     activeMatchID != nil,
		ActiveMatchID:     activeMatchID,
		Player:            ToPlayerRatingDTO(rating),
		Tickets:           10, // TODO: get from user wallet
		FriendsOnline:     []FriendDTO{}, // TODO: implement friends service
		PendingChallenges: challengeDTOs,
		SeasonID:          seasonID,
		SeasonEndsAt:      seasonEndsAt,
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

	// Check if already in match
	activeMatch, err := uc.duelGameRepo.FindActiveByPlayer(playerID)
	if err == nil && activeMatch != nil {
		return JoinQueueOutput{}, quick_duel.ErrAlreadyInMatch
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

	// Check if friend is already in a match
	activeMatch, err := uc.duelGameRepo.FindActiveByPlayer(friendID)
	if err == nil && activeMatch != nil {
		return SendChallengeOutput{}, quick_duel.ErrFriendBusy
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
	questionRepo     interface{ FindRandomQuestions(count int) ([]*interface{}, error) } // Simplified
	seasonRepo       quick_duel.SeasonRepository
	eventBus         EventBus
}

func NewRespondChallengeUseCase(
	challengeRepo quick_duel.ChallengeRepository,
	duelGameRepo quick_duel.DuelGameRepository,
	playerRatingRepo quick_duel.PlayerRatingRepository,
	seasonRepo quick_duel.SeasonRepository,
	eventBus EventBus,
) *RespondChallengeUseCase {
	return &RespondChallengeUseCase{
		challengeRepo:    challengeRepo,
		duelGameRepo:     duelGameRepo,
		playerRatingRepo: playerRatingRepo,
		seasonRepo:       seasonRepo,
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

	// Save challenge
	err = uc.challengeRepo.Save(challenge)
	if err != nil {
		return RespondChallengeOutput{}, err
	}

	// Publish events
	for _, event := range challenge.Events() {
		uc.eventBus.Publish(event)
	}

	// TODO: Create match between challenger and challenged
	// For now, return placeholder
	startsIn := 3

	return RespondChallengeOutput{
		Success:        true,
		MatchID:        nil, // Will be set when match is created
		TicketConsumed: true,
		StartsIn:       &startsIn,
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
// GetMatchHistory Use Case
// ========================================

type GetMatchHistoryUseCase struct {
	duelGameRepo quick_duel.DuelGameRepository
	userRepo     domainUser.UserRepository
}

func NewGetMatchHistoryUseCase(
	duelGameRepo quick_duel.DuelGameRepository,
	userRepo domainUser.UserRepository,
) *GetMatchHistoryUseCase {
	return &GetMatchHistoryUseCase{
		duelGameRepo: duelGameRepo,
		userRepo:     userRepo,
	}
}

func (uc *GetMatchHistoryUseCase) Execute(input GetMatchHistoryInput) (GetMatchHistoryOutput, error) {
	playerID, err := shared.NewUserID(input.PlayerID)
	if err != nil {
		return GetMatchHistoryOutput{}, err
	}

	limit := input.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	matches, total, err := uc.duelGameRepo.FindByPlayerPaginated(playerID, limit, input.Offset, input.Filter)
	if err != nil {
		return GetMatchHistoryOutput{}, err
	}

	entries := make([]MatchHistoryEntryDTO, 0, len(matches))
	for _, match := range matches {
		// Determine opponent
		opponentID := match.Player2().UserID()
		if match.Player1().UserID().String() != input.PlayerID {
			opponentID = match.Player1().UserID()
		}

		opponentUsername := "Player"
		if user, err := uc.userRepo.FindByID(opponentID); err == nil && user != nil {
			opponentUsername = user.Username().String()
		}

		entries = append(entries, ToMatchHistoryEntryDTO(match, input.PlayerID, opponentUsername))
	}

	return GetMatchHistoryOutput{
		Matches: entries,
		Total:   total,
		HasMore: input.Offset+len(matches) < total,
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
