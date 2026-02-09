package quick_duel

import (
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

// ========================================
// StartMatch Use Case
// ========================================

type StartMatchUseCase struct {
	duelGameRepo     quick_duel.DuelGameRepository
	playerRatingRepo quick_duel.PlayerRatingRepository
	questionRepo     QuestionRepository
	seasonRepo       quick_duel.SeasonRepository
	eventBus         EventBus
}

// QuestionRepository interface for getting questions
type QuestionRepository interface {
	FindRandomByDifficulty(count int, difficulty string) ([]QuestionData, error)
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

func NewStartMatchUseCase(
	duelGameRepo quick_duel.DuelGameRepository,
	playerRatingRepo quick_duel.PlayerRatingRepository,
	questionRepo QuestionRepository,
	seasonRepo quick_duel.SeasonRepository,
	eventBus EventBus,
) *StartMatchUseCase {
	return &StartMatchUseCase{
		duelGameRepo:     duelGameRepo,
		playerRatingRepo: playerRatingRepo,
		questionRepo:     questionRepo,
		seasonRepo:       seasonRepo,
		eventBus:         eventBus,
	}
}

func (uc *StartMatchUseCase) Execute(input StartMatchInput) (StartMatchOutput, error) {
	now := time.Now().UTC().Unix()

	player1ID, err := shared.NewUserID(input.Player1ID)
	if err != nil {
		return StartMatchOutput{}, err
	}

	player2ID, err := shared.NewUserID(input.Player2ID)
	if err != nil {
		return StartMatchOutput{}, err
	}

	// Get player ratings
	seasonID, _ := uc.seasonRepo.GetCurrentSeason()
	rating1, err := uc.playerRatingRepo.FindOrCreate(player1ID, seasonID, now)
	if err != nil {
		return StartMatchOutput{}, err
	}

	rating2, err := uc.playerRatingRepo.FindOrCreate(player2ID, seasonID, now)
	if err != nil {
		return StartMatchOutput{}, err
	}

	// Get random questions for the duel
	questions, err := uc.questionRepo.FindRandomByDifficulty(quick_duel.QuestionsPerDuel, "medium")
	if err != nil {
		return StartMatchOutput{}, err
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
		return StartMatchOutput{}, err
	}
	game.Start(now)

	// Save game
	err = uc.duelGameRepo.Save(game)
	if err != nil {
		return StartMatchOutput{}, err
	}

	return StartMatchOutput{
		MatchID:   game.ID().String(),
		Player1ID: player1ID.String(),
		Player2ID: player2ID.String(),
		StartsAt:  now + 3, // 3 second countdown
	}, nil
}

// GetRoundQuestion returns the question for a specific round
func (uc *StartMatchUseCase) GetRoundQuestion(matchID string, roundNum int) (*RoundQuestionOutput, error) {
	gameID := quick_duel.NewGameIDFromString(matchID)
	game, err := uc.duelGameRepo.FindByID(gameID)
	if err != nil {
		return nil, err
	}

	if roundNum < 1 || roundNum > len(game.QuestionIDs()) {
		return nil, quick_duel.ErrGameNotFound
	}

	// Get question data
	questionID := game.QuestionIDs()[roundNum-1]
	questions, err := uc.questionRepo.FindRandomByDifficulty(1, "medium")
	if err != nil || len(questions) == 0 {
		return nil, err
	}

	// In real implementation, we'd look up the specific question by ID
	// For now, just return the first random question as placeholder
	q := questions[0]

	answers := make([]map[string]string, 0, len(q.Answers))
	for _, a := range q.Answers {
		answers = append(answers, map[string]string{
			"id":   a.ID,
			"text": a.Text,
		})
	}

	return &RoundQuestionOutput{
		QuestionID:   questionID.String(),
		QuestionText: q.Text,
		Answers:      answers,
	}, nil
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
	// Track round answers temporarily (in production, store in Redis)
	roundAnswers map[string]map[int][]playerAnswer
}

type playerAnswer struct {
	PlayerID   string
	AnswerID   string
	IsCorrect  bool
	TimeTaken  int
	Points     int
}

func NewSubmitDuelAnswerUseCase(
	duelGameRepo quick_duel.DuelGameRepository,
	playerRatingRepo quick_duel.PlayerRatingRepository,
	questionRepo QuestionRepository,
	seasonRepo quick_duel.SeasonRepository,
	eventBus EventBus,
) *SubmitDuelAnswerUseCase {
	return &SubmitDuelAnswerUseCase{
		duelGameRepo:     duelGameRepo,
		playerRatingRepo: playerRatingRepo,
		questionRepo:     questionRepo,
		seasonRepo:       seasonRepo,
		eventBus:         eventBus,
		roundAnswers:     make(map[string]map[int][]playerAnswer),
	}
}

func (uc *SubmitDuelAnswerUseCase) Execute(input SubmitDuelAnswerInput) (*SubmitDuelAnswerOutput, error) {
	now := time.Now().UTC().Unix()

	gameID := quick_duel.NewGameIDFromString(input.MatchID)
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

	// Get the correct answer (in real implementation, look up from question repo)
	// For now, simulate answer validation
	isCorrect := true // Placeholder - would check against actual correct answer
	correctAnswerID := input.AnswerID // Placeholder

	// Calculate points based on correctness and time
	points := 0
	if isCorrect {
		// Base 100 points + time bonus (max 100 for instant answer)
		timeBonus := max(0, 100-(input.TimeTaken/100)) // 100ms per point deduction
		points = 100 + timeBonus
	}

	// Track this answer
	currentRound := game.CurrentRound()
	if uc.roundAnswers[input.MatchID] == nil {
		uc.roundAnswers[input.MatchID] = make(map[int][]playerAnswer)
	}

	uc.roundAnswers[input.MatchID][currentRound] = append(
		uc.roundAnswers[input.MatchID][currentRound],
		playerAnswer{
			PlayerID:  input.PlayerID,
			AnswerID:  input.AnswerID,
			IsCorrect: isCorrect,
			TimeTaken: input.TimeTaken,
			Points:    points,
		},
	)

	// Check if round is complete (both players answered)
	roundAnswers := uc.roundAnswers[input.MatchID][currentRound]
	roundComplete := len(roundAnswers) >= 2

	// Calculate current scores
	player1Score := game.Player1().Score()
	player2Score := game.Player2().Score()

	for _, ans := range roundAnswers {
		if ans.PlayerID == game.Player1().UserID().String() {
			player1Score += ans.Points
		} else {
			player2Score += ans.Points
		}
	}

	// Check if match is complete
	matchComplete := roundComplete && currentRound >= quick_duel.QuestionsPerDuel

	output := &SubmitDuelAnswerOutput{
		IsCorrect:       isCorrect,
		CorrectAnswerID: correctAnswerID,
		PointsEarned:    points,
		Player1Score:    player1Score,
		Player2Score:    player2Score,
		RoundComplete:   roundComplete,
		MatchComplete:   matchComplete,
	}

	// If match is complete, calculate MMR changes
	if matchComplete {
		uc.finalizeMatch(game, player1Score, player2Score, now, output)
	}

	return output, nil
}

func (uc *SubmitDuelAnswerUseCase) finalizeMatch(
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
		rating1.ApplyMatchResult(quick_duel.MatchResult{
			Won:         player1Won,
			OpponentMMR: game.Player2().Elo().Rating(),
			MatchTime:   now,
		})
		uc.playerRatingRepo.Save(rating1)
		output.Player1MMRChange = rating1.MMR() - oldMMR1
		output.Player1NewMMR = rating1.MMR()
	}

	// Update player 2 rating
	rating2, err := uc.playerRatingRepo.FindOrCreate(game.Player2().UserID(), seasonID, now)
	if err == nil {
		oldMMR2 := rating2.MMR()
		rating2.ApplyMatchResult(quick_duel.MatchResult{
			Won:         player2Won,
			OpponentMMR: game.Player1().Elo().Rating(),
			MatchTime:   now,
		})
		uc.playerRatingRepo.Save(rating2)
		output.Player2MMRChange = rating2.MMR() - oldMMR2
		output.Player2NewMMR = rating2.MMR()
	}

	// Save game (domain already updated status internally)
	uc.duelGameRepo.Save(game)

	// Clean up round answers
	delete(uc.roundAnswers, game.ID().String())
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

	// Get the original match
	matchID := quick_duel.NewGameIDFromString(input.MatchID)
	match, err := uc.duelGameRepo.FindByID(matchID)
	if err != nil {
		return RequestRematchOutput{}, err
	}

	// Verify player was in the match
	isPlayer1 := match.Player1().UserID().String() == input.PlayerID
	isPlayer2 := match.Player2().UserID().String() == input.PlayerID
	if !isPlayer1 && !isPlayer2 {
		return RequestRematchOutput{}, quick_duel.ErrGameNotFound
	}

	// Verify match is finished
	if match.Status() != quick_duel.GameStatusFinished {
		return RequestRematchOutput{}, quick_duel.ErrGameNotActive
	}

	// Determine opponent
	var opponentID quick_duel.UserID
	if isPlayer1 {
		opponentID = match.Player2().UserID()
	} else {
		opponentID = match.Player1().UserID()
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
					MatchID:   nil, // Would be set when match starts
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
	SetInMatch(playerID string, matchID string) error
	ClearInMatch(playerID string) error
	GetMatchID(playerID string) (string, error)
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

		// Check if in match
		matchID, _ := uc.onlineTracker.GetMatchID(friendID)
		inMatch := matchID != ""

		friends = append(friends, FriendDTO{
			ID:       friendID,
			Username: user.Username().String(),
			IsOnline: true,
			InMatch:  inMatch,
		})
	}

	return GetOnlineFriendsOutput{
		OnlineFriends: friends,
	}, nil
}
