package daily_challenge

import (
	"time"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/daily_challenge"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
	domainUser "github.com/barsukov/quiz-sprint/backend/internal/domain/user"
)

// ========================================
// StartDailyChallenge Use Case
// ========================================

type StartDailyChallengeUseCase struct {
	dailyQuizRepo     daily_challenge.DailyQuizRepository
	dailyGameRepo     daily_challenge.DailyGameRepository
	questionRepo      quiz.QuestionRepository
	quizRepo          quiz.QuizRepository
	eventBus          EventBus
	getOrCreateQuizUC *GetOrCreateDailyQuizUseCase
}

func NewStartDailyChallengeUseCase(
	dailyQuizRepo daily_challenge.DailyQuizRepository,
	dailyGameRepo daily_challenge.DailyGameRepository,
	questionRepo quiz.QuestionRepository,
	quizRepo quiz.QuizRepository,
	eventBus EventBus,
	getOrCreateQuizUC *GetOrCreateDailyQuizUseCase,
) *StartDailyChallengeUseCase {
	return &StartDailyChallengeUseCase{
		dailyQuizRepo:     dailyQuizRepo,
		dailyGameRepo:     dailyGameRepo,
		questionRepo:      questionRepo,
		quizRepo:          quizRepo,
		eventBus:          eventBus,
		getOrCreateQuizUC: getOrCreateQuizUC,
	}
}

func (uc *StartDailyChallengeUseCase) Execute(input StartDailyChallengeInput) (StartDailyChallengeOutput, error) {
	// 1. Determine date
	var date daily_challenge.Date
	if input.Date != "" {
		date = daily_challenge.NewDateFromString(input.Date)
	} else {
		date = daily_challenge.TodayUTC()
	}

	now := time.Now().UTC().Unix()
	playerID, err := shared.NewUserID(input.PlayerID)
	if err != nil {
		println("❌ [StartDailyChallenge] Invalid player ID:", err.Error())
		return StartDailyChallengeOutput{}, err
	}

	println("🎮 [StartDailyChallenge] Starting for player:", playerID.String(), "date:", date.String())

	// 2. Check if player already played today
	existingGame, err := uc.dailyGameRepo.FindByPlayerAndDate(playerID, date)
	if err == nil && existingGame != nil {
		println("⚠️  [StartDailyChallenge] Player already played today")
		return StartDailyChallengeOutput{}, daily_challenge.ErrAlreadyPlayedToday
	}

	println("✅ [StartDailyChallenge] Player hasn't played today")

	// 3. Get or create daily quiz
	quizOutput, err := uc.getOrCreateQuizUC.Execute(GetOrCreateDailyQuizInput{Date: date.String()})
	if err != nil {
		println("❌ [StartDailyChallenge] Failed to get/create quiz:", err.Error())
		return StartDailyChallengeOutput{}, err
	}

	println("✅ [StartDailyChallenge] Got daily quiz")

	dailyQuizEntity, err := uc.dailyQuizRepo.FindByDate(date)
	if err != nil {
		println("❌ [StartDailyChallenge] Failed to load quiz from repo:", err.Error())
		return StartDailyChallengeOutput{}, err
	}

	println("✅ [StartDailyChallenge] Loaded quiz entity, loading questions...")

	// 4. Load questions and create Quiz aggregate
	questions, err := uc.questionRepo.FindByIDs(dailyQuizEntity.QuestionIDs())
	if err != nil {
		println("❌ [StartDailyChallenge] Failed to load questions:", err.Error())
		return StartDailyChallengeOutput{}, err
	}

	println("✅ [StartDailyChallenge] Loaded", len(questions), "questions")

	// Create Quiz
	println("📋 [StartDailyChallenge] Creating Quiz aggregate...")
	quizID := quiz.NewQuizID()
	quizTitle, _ := quiz.NewQuizTitle("Daily Challenge - " + date.String())
	quizTimeLimit, _ := quiz.NewTimeLimit(15 * 10) // 15 seconds per question
	quizPassingScore, _ := quiz.NewPassingScore(0)  // No passing score requirement

	quizAggregate, err := quiz.NewQuiz(
		quizID,
		quizTitle,
		"Complete 10 questions to see your rank!",
		quiz.CategoryID{}, // No category
		quizTimeLimit,
		quizPassingScore,
		now,
	)
	if err != nil {
		println("❌ [StartDailyChallenge] Failed to create quiz aggregate:", err.Error())
		return StartDailyChallengeOutput{}, err
	}

	// Daily Challenge scoring: basePoints=100, timeBonus=max(0, (15-timeTaken)*5) → max 75
	// Per docs/game_modes/daily_challenge/03_rules.md
	dailyBasePoints, _ := quiz.NewPoints(100)
	dailyMaxTimeBonus, _ := quiz.NewPoints(75)
	quizAggregate.SetBasePoints(dailyBasePoints)
	quizAggregate.SetTimeLimitPerQuestion(15)
	quizAggregate.SetMaxTimeBonus(dailyMaxTimeBonus)

	println("✅ [StartDailyChallenge] Created quiz aggregate, adding questions...")

	// Add questions to quiz
	for i, question := range questions {
		if err := quizAggregate.AddQuestion(*question); err != nil {
			println("❌ [StartDailyChallenge] Failed to add question", i, ":", err.Error())
			return StartDailyChallengeOutput{}, err
		}
	}

	println("✅ [StartDailyChallenge] Added all questions, checking streak...")

	// 5. Get player's current streak
	streak := daily_challenge.NewStreakSystem()
	if lastGame, err := uc.dailyGameRepo.FindByPlayerAndDate(playerID, date.Previous()); err == nil && lastGame != nil {
		streak = lastGame.Streak()
	}

	println("✅ [StartDailyChallenge] Got streak, creating daily game...")

	// 6. Create daily game
	game, err := daily_challenge.NewDailyGame(
		playerID,
		dailyQuizEntity.ID(),
		date,
		quizAggregate,
		streak,
		now,
	)
	if err != nil {
		println("❌ [StartDailyChallenge] Failed to create daily game:", err.Error())
		return StartDailyChallengeOutput{}, err
	}

	println("✅ [StartDailyChallenge] Created daily game, saving...")

	// 7. Save game
	if err := uc.dailyGameRepo.Save(game); err != nil {
		println("❌ [StartDailyChallenge] Failed to save game:", err.Error())
		return StartDailyChallengeOutput{}, err
	}

	println("✅ [StartDailyChallenge] Saved game, publishing events...")

	// 8. Publish events
	for _, event := range game.Events() {
		uc.eventBus.Publish(event)
	}

	println("✅ [StartDailyChallenge] Published events, getting first question...")

	// 9. Get first question
	firstQuestion, err := game.Session().GetCurrentQuestion()
	if err != nil {
		println("❌ [StartDailyChallenge] Failed to get first question:", err.Error())
		return StartDailyChallengeOutput{}, err
	}

	println("✅ [StartDailyChallenge] Got first question, building response...")

	// 10. Calculate time to expire
	timeToExpire := dailyQuizEntity.ExpiresAt() - now

	return StartDailyChallengeOutput{
		Game:          ToDailyGameDTO(game, now),
		FirstQuestion: ToQuestionDTO(firstQuestion),
		TimeLimit:     15, // Fixed: 15 seconds per question
		TotalPlayers:  quizOutput.TotalPlayers,
		TimeToExpire:  timeToExpire,
	}, nil
}

// ========================================
// SubmitDailyAnswer Use Case
// ========================================

type SubmitDailyAnswerUseCase struct {
	dailyGameRepo       daily_challenge.DailyGameRepository
	eventBus            EventBus
	getLeaderboardUC    *GetDailyLeaderboardUseCase
	chestRewardCalc     *daily_challenge.ChestRewardCalculator
}

func NewSubmitDailyAnswerUseCase(
	dailyGameRepo daily_challenge.DailyGameRepository,
	eventBus EventBus,
	getLeaderboardUC *GetDailyLeaderboardUseCase,
	chestRewardCalc *daily_challenge.ChestRewardCalculator,
) *SubmitDailyAnswerUseCase {
	return &SubmitDailyAnswerUseCase{
		dailyGameRepo:       dailyGameRepo,
		eventBus:            eventBus,
		getLeaderboardUC:    getLeaderboardUC,
		chestRewardCalc:     chestRewardCalc,
	}
}

func (uc *SubmitDailyAnswerUseCase) Execute(input SubmitDailyAnswerInput) (SubmitDailyAnswerOutput, error) {
	now := time.Now().UTC().Unix()

	println("📝 [SubmitDailyAnswer] Received request for game:", input.GameID, "question:", input.QuestionID)

	// 1. Load game
	gameID := daily_challenge.NewGameIDFromString(input.GameID)
	game, err := uc.dailyGameRepo.FindByID(gameID)
	if err != nil {
		println("❌ [SubmitDailyAnswer] Failed to find game:", err.Error())
		return SubmitDailyAnswerOutput{}, err
	}

	println("✅ [SubmitDailyAnswer] Game found, checking authorization...")

	// 2. Authorization check
	if game.PlayerID().String() != input.PlayerID {
		println("❌ [SubmitDailyAnswer] Player ID mismatch")
		return SubmitDailyAnswerOutput{}, daily_challenge.ErrGameNotFound
	}

	println("✅ [SubmitDailyAnswer] Authorization OK, answering question...")

	// 3. Answer question
	questionID, _ := quiz.NewQuestionIDFromString(input.QuestionID)
	answerID, _ := quiz.NewAnswerIDFromString(input.AnswerID)

	result, err := game.AnswerQuestion(questionID, answerID, input.TimeTaken, now)
	if err != nil {
		println("❌ [SubmitDailyAnswer] Failed to answer question:", err.Error())
		return SubmitDailyAnswerOutput{}, err
	}

	println("✅ [SubmitDailyAnswer] Question answered")
	println("   - Question index:", result.QuestionIndex)
	println("   - Remaining questions:", result.RemainingQuestions)
	println("   - Is game completed:", result.IsGameCompleted)
	println("   - Game status:", game.Status())

	println("📝 [SubmitDailyAnswer] Saving game...")

	// 4. Save game
	if err := uc.dailyGameRepo.Save(game); err != nil {
		println("❌ [SubmitDailyAnswer] Failed to save game:", err.Error())
		return SubmitDailyAnswerOutput{}, err
	}

	println("✅ [SubmitDailyAnswer] Game saved, publishing events...")

	// 5. Publish events
	for _, event := range game.Events() {
		uc.eventBus.Publish(event)
	}

	// 6. Build output (with instant feedback)
	output := SubmitDailyAnswerOutput{
		QuestionIndex:      result.QuestionIndex,
		TotalQuestions:     10,
		RemainingQuestions: result.RemainingQuestions,
		IsGameCompleted:    result.IsGameCompleted,
		IsCorrect:          result.IsCorrect,
		CorrectAnswerID:    result.CorrectAnswerID,
	}

	// 7. If game continues, return next question
	if !result.IsGameCompleted {
		if nextQ, err := game.Session().GetCurrentQuestion(); err == nil {
			questionDTO := ToQuestionDTO(nextQ)
			timeLimit := 15
			output.NextQuestion = &questionDTO
			output.NextTimeLimit = &timeLimit
		}
	} else {
		// Game completed - calculate rank and chest rewards
		rank, _ := uc.dailyGameRepo.GetPlayerRankByDate(game.PlayerID(), game.Date())
		totalPlayers, _ := uc.dailyGameRepo.GetTotalPlayersByDate(game.Date())

		// Calculate chest rewards per docs/game_modes/daily_challenge/04_rewards.md
		correctAnswers := game.GetCorrectAnswersCount()
		totalQuestions := game.Session().Quiz().QuestionsCount()
		chestType := daily_challenge.CalculateChestType(correctAnswers, totalQuestions)
		streakBonus := game.Streak().GetBonus()

		println("🎁 [SubmitDailyAnswer] Calculating chest rewards:")
		println("   - correctAnswers:", correctAnswers)
		println("   - chestType:", chestType)
		println("   - streakBonus:", streakBonus)

		chestReward := uc.chestRewardCalc.CalculateRewards(chestType, streakBonus)

		println("   - coins:", chestReward.Coins())
		println("   - pvpTickets:", chestReward.PvpTickets())
		println("   - bonuses:", len(chestReward.MarathonBonuses()))

		// Set chest reward and emit event
		game.SetChestReward(chestReward)
		game.EmitChestEarnedEvent(chestReward, now)

		game.SetRank(rank)
		uc.dailyGameRepo.Save(game) // Update with rank and chest reward

		// Publish events (including ChestEarnedEvent)
		for _, event := range game.Events() {
			uc.eventBus.Publish(event)
		}

		// Fetch leaderboard
		leaderboardEntries := make([]LeaderboardEntryDTO, 0)
		if leaderboard, err := uc.getLeaderboardUC.Execute(GetDailyLeaderboardInput{
			Date:  game.Date().String(),
			Limit: 10,
		}); err == nil {
			leaderboardEntries = leaderboard.Entries
		}

		results := BuildGameResultsDTO(game, rank, totalPlayers, leaderboardEntries)
		output.GameResults = &results
	}

	return output, nil
}

// ========================================
// GetDailyGameStatus Use Case
// ========================================

type GetDailyGameStatusUseCase struct {
	dailyQuizRepo    daily_challenge.DailyQuizRepository
	dailyGameRepo    daily_challenge.DailyGameRepository
	getLeaderboardUC *GetDailyLeaderboardUseCase
}

func NewGetDailyGameStatusUseCase(
	dailyQuizRepo daily_challenge.DailyQuizRepository,
	dailyGameRepo daily_challenge.DailyGameRepository,
	getLeaderboardUC *GetDailyLeaderboardUseCase,
) *GetDailyGameStatusUseCase {
	return &GetDailyGameStatusUseCase{
		dailyQuizRepo:    dailyQuizRepo,
		dailyGameRepo:    dailyGameRepo,
		getLeaderboardUC: getLeaderboardUC,
	}
}

func (uc *GetDailyGameStatusUseCase) Execute(input GetDailyGameStatusInput) (GetDailyGameStatusOutput, error) {
	// 1. Determine date
	var date daily_challenge.Date
	if input.Date != "" {
		date = daily_challenge.NewDateFromString(input.Date)
	} else {
		date = daily_challenge.TodayUTC()
	}

	now := time.Now().UTC().Unix()
	playerID, err := shared.NewUserID(input.PlayerID)
	if err != nil {
		return GetDailyGameStatusOutput{}, err
	}

	// 2. Find player's game for date
	game, err := uc.dailyGameRepo.FindByPlayerAndDate(playerID, date)

	// 3. Get daily quiz info (for expiry time)
	dailyQuiz, _ := uc.dailyQuizRepo.FindByDate(date)
	timeToExpire := int64(0)
	if dailyQuiz != nil {
		timeToExpire = dailyQuiz.ExpiresAt() - now
		if timeToExpire < 0 {
			timeToExpire = 0
		}
	}

	// 4. Get total players
	totalPlayers, _ := uc.dailyGameRepo.GetTotalPlayersByDate(date)

	// 5. Build output
	if err != nil || game == nil {
		return GetDailyGameStatusOutput{
			HasPlayed:    false,
			TimeToExpire: timeToExpire,
			TotalPlayers: totalPlayers,
		}, nil
	}

	gameDTO := ToDailyGameDTO(game, now)
	var timeLimit *int
	var results *GameResultsDTO

	if game.Status() == daily_challenge.GameStatusInProgress {
		tl := 15
		timeLimit = &tl
	} else if game.Status() == daily_challenge.GameStatusCompleted {
		// Build results for completed games
		rank, _ := uc.dailyGameRepo.GetPlayerRankByDate(game.PlayerID(), game.Date())

		// Fetch leaderboard
		leaderboardEntries := make([]LeaderboardEntryDTO, 0)
		if leaderboard, err := uc.getLeaderboardUC.Execute(GetDailyLeaderboardInput{
			Date:  game.Date().String(),
			Limit: 10,
		}); err == nil {
			leaderboardEntries = leaderboard.Entries
		}

		gameResults := BuildGameResultsDTO(game, rank, totalPlayers, leaderboardEntries)
		results = &gameResults
	}

	// HasPlayed = true ONLY if game is completed
	hasPlayed := game.Status() == daily_challenge.GameStatusCompleted

	return GetDailyGameStatusOutput{
		HasPlayed:    hasPlayed,
		Game:         &gameDTO,
		Results:      results,
		TimeLimit:    timeLimit,
		TimeToExpire: timeToExpire,
		TotalPlayers: totalPlayers,
	}, nil
}

// ========================================
// GetDailyLeaderboard Use Case
// ========================================

type GetDailyLeaderboardUseCase struct {
	dailyGameRepo daily_challenge.DailyGameRepository
	userRepo      domainUser.UserRepository
}

func NewGetDailyLeaderboardUseCase(
	dailyGameRepo daily_challenge.DailyGameRepository,
	userRepo domainUser.UserRepository,
) *GetDailyLeaderboardUseCase {
	return &GetDailyLeaderboardUseCase{
		dailyGameRepo: dailyGameRepo,
		userRepo:      userRepo,
	}
}

func (uc *GetDailyLeaderboardUseCase) Execute(input GetDailyLeaderboardInput) (GetDailyLeaderboardOutput, error) {
	// 1. Determine date
	var date daily_challenge.Date
	if input.Date != "" {
		date = daily_challenge.NewDateFromString(input.Date)
	} else {
		date = daily_challenge.TodayUTC()
	}

	// 2. Validate limit
	limit := input.Limit
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	// 3. Get top games
	topGames, err := uc.dailyGameRepo.FindTopByDate(date, limit)
	if err != nil {
		return GetDailyLeaderboardOutput{}, err
	}

	// 4. Build leaderboard entries
	entries := make([]LeaderboardEntryDTO, 0, len(topGames))
	for rank, game := range topGames {
		username := "Player"
		if user, err := uc.userRepo.FindByID(game.PlayerID()); err == nil && user != nil {
			username = user.Username().String()
		}

		entries = append(entries, ToLeaderboardEntryDTO(game, username, rank+1))
	}

	// 5. Get player rank if requested
	var playerRank *int
	if input.PlayerID != "" {
		playerID, err := shared.NewUserID(input.PlayerID)
		if err != nil {
			return GetDailyLeaderboardOutput{}, err
		}
		rank, err := uc.dailyGameRepo.GetPlayerRankByDate(playerID, date)
		if err == nil && rank > 0 {
			playerRank = &rank
		}
	}

	// 6. Get total players
	totalPlayers, _ := uc.dailyGameRepo.GetTotalPlayersByDate(date)

	return GetDailyLeaderboardOutput{
		Date:         date.String(),
		Entries:      entries,
		TotalPlayers: totalPlayers,
		PlayerRank:   playerRank,
	}, nil
}

// ========================================
// GetPlayerStreak Use Case
// ========================================

type GetPlayerStreakUseCase struct {
	dailyGameRepo daily_challenge.DailyGameRepository
}

func NewGetPlayerStreakUseCase(
	dailyGameRepo daily_challenge.DailyGameRepository,
) *GetPlayerStreakUseCase {
	return &GetPlayerStreakUseCase{
		dailyGameRepo: dailyGameRepo,
	}
}

func (uc *GetPlayerStreakUseCase) Execute(input GetPlayerStreakInput) (GetPlayerStreakOutput, error) {
	playerID, err := shared.NewUserID(input.PlayerID)
	if err != nil {
		return GetPlayerStreakOutput{}, err
	}
	today := daily_challenge.TodayUTC()

	// Get player's most recent game
	streak := daily_challenge.NewStreakSystem()
	if game, err := uc.dailyGameRepo.FindByPlayerAndDate(playerID, today); err == nil && game != nil {
		streak = game.Streak()
	} else if game, err := uc.dailyGameRepo.FindByPlayerAndDate(playerID, today.Previous()); err == nil && game != nil {
		streak = game.Streak()
	}

	// Calculate next milestone (per docs/GLOSSARY.md: 3, 7, 14, 30)
	milestones := []int{3, 7, 14, 30}
	nextMilestone := 3
	for _, m := range milestones {
		if streak.CurrentStreak() < m {
			nextMilestone = m
			break
		}
	}
	daysToNext := nextMilestone - streak.CurrentStreak()
	if daysToNext < 0 {
		daysToNext = 0
	}

	// Check if can restore (missed yesterday)
	canRestore := false
	if !streak.LastPlayedDate().IsZero() {
		twoDaysAgo := today.Previous().Previous()
		canRestore = streak.LastPlayedDate().Equals(twoDaysAgo)
	}

	return GetPlayerStreakOutput{
		Streak:        ToStreakDTO(streak, today),
		NextMilestone: nextMilestone,
		DaysToNext:    daysToNext,
		CanRestore:    canRestore,
	}, nil
}

// ========================================
// OpenChest Use Case
// ========================================

type OpenChestUseCase struct {
	dailyGameRepo daily_challenge.DailyGameRepository
}

func NewOpenChestUseCase(
	dailyGameRepo daily_challenge.DailyGameRepository,
) *OpenChestUseCase {
	return &OpenChestUseCase{
		dailyGameRepo: dailyGameRepo,
	}
}

func (uc *OpenChestUseCase) Execute(input OpenChestInput) (OpenChestOutput, error) {
	// 1. Load game
	gameID := daily_challenge.NewGameIDFromString(input.GameID)
	game, err := uc.dailyGameRepo.FindByID(gameID)
	if err != nil {
		return OpenChestOutput{}, err
	}

	// 2. Authorization check
	if game.PlayerID().String() != input.PlayerID {
		return OpenChestOutput{}, daily_challenge.ErrGameNotFound
	}

	// 3. Verify game is completed
	if !game.IsCompleted() {
		return OpenChestOutput{}, daily_challenge.ErrGameNotActive
	}

	// 4. Get chest reward (should already be calculated and stored)
	chestReward := game.ChestReward()
	if chestReward == nil {
		// Shouldn't happen if game completed properly
		return OpenChestOutput{}, daily_challenge.ErrGameNotFound
	}

	// 5. Build output (idempotent - just returns stored data)
	return OpenChestOutput{
		ChestType:      chestReward.ChestType().String(),
		Rewards:        ToChestRewardDTO(*chestReward),
		StreakBonus:    game.Streak().GetBonus(),
		PremiumApplied: false, // TODO: check if player has premium
	}, nil
}

// ========================================
// RetryChallenge Use Case
// ========================================

type RetryChallengeUseCase struct {
	dailyGameRepo     daily_challenge.DailyGameRepository
	dailyQuizRepo     daily_challenge.DailyQuizRepository
	questionRepo      quiz.QuestionRepository
	eventBus          EventBus
}

func NewRetryChallengeUseCase(
	dailyGameRepo daily_challenge.DailyGameRepository,
	dailyQuizRepo daily_challenge.DailyQuizRepository,
	questionRepo quiz.QuestionRepository,
	eventBus EventBus,
) *RetryChallengeUseCase {
	return &RetryChallengeUseCase{
		dailyGameRepo: dailyGameRepo,
		dailyQuizRepo: dailyQuizRepo,
		questionRepo:  questionRepo,
		eventBus:      eventBus,
	}
}

func (uc *RetryChallengeUseCase) Execute(input RetryChallengeInput) (RetryChallengeOutput, error) {
	now := time.Now().UTC().Unix()

	// 1. Load original game
	gameID := daily_challenge.NewGameIDFromString(input.GameID)
	originalGame, err := uc.dailyGameRepo.FindByID(gameID)
	if err != nil {
		return RetryChallengeOutput{}, err
	}

	// 2. Authorization check
	playerID, err := shared.NewUserID(input.PlayerID)
	if err != nil {
		return RetryChallengeOutput{}, err
	}

	if originalGame.PlayerID() != playerID {
		return RetryChallengeOutput{}, daily_challenge.ErrGameNotFound
	}

	// 3. Verify original game is completed
	if !originalGame.IsCompleted() {
		return RetryChallengeOutput{}, daily_challenge.ErrGameNotActive
	}

	// 4. Check retry limit
	attemptCount, err := uc.dailyGameRepo.CountAttemptsByPlayerAndDate(playerID, originalGame.Date())
	if err != nil {
		return RetryChallengeOutput{}, err
	}

	// TODO: check premium status - for now limit to 2 attempts (1 free + 1 retry)
	isPremium := false
	maxAttempts := 2
	if isPremium {
		maxAttempts = 999 // Unlimited for premium
	}

	if attemptCount >= maxAttempts {
		return RetryChallengeOutput{}, daily_challenge.ErrAlreadyPlayedToday // Reuse error (TODO: specific error)
	}

	// 5. Process payment
	coinsDeducted := 0
	if input.PaymentMethod == "coins" {
		coinsDeducted = 100
		// TODO: deduct coins from user account
		// For now just return the cost
	} else if input.PaymentMethod == "ad" {
		// TODO: verify ad was watched (via ad network callback)
		coinsDeducted = 0
	} else {
		return RetryChallengeOutput{}, daily_challenge.ErrInvalidGameID // Invalid payment method
	}

	// 6. Load daily quiz
	dailyQuiz, err := uc.dailyQuizRepo.FindByDate(originalGame.Date())
	if err != nil {
		return RetryChallengeOutput{}, err
	}

	// 7. Load questions and create Quiz aggregate
	questions, err := uc.questionRepo.FindByIDs(dailyQuiz.QuestionIDs())
	if err != nil {
		return RetryChallengeOutput{}, err
	}

	quizID := quiz.NewQuizID()
	quizTitle, _ := quiz.NewQuizTitle("Daily Challenge - Retry")
	quizTimeLimit, _ := quiz.NewTimeLimit(15 * 10)
	quizPassingScore, _ := quiz.NewPassingScore(0)

	quizAggregate, err := quiz.NewQuiz(
		quizID,
		quizTitle,
		"Retry attempt",
		quiz.CategoryID{},
		quizTimeLimit,
		quizPassingScore,
		now,
	)
	if err != nil {
		return RetryChallengeOutput{}, err
	}

	// Daily Challenge scoring overrides (same as start)
	retryBasePoints, _ := quiz.NewPoints(100)
	retryMaxTimeBonus, _ := quiz.NewPoints(75)
	quizAggregate.SetBasePoints(retryBasePoints)
	quizAggregate.SetTimeLimitPerQuestion(15)
	quizAggregate.SetMaxTimeBonus(retryMaxTimeBonus)

	for _, question := range questions {
		if err := quizAggregate.AddQuestion(*question); err != nil {
			return RetryChallengeOutput{}, err
		}
	}

	// 8. Create new game (IMPORTANT: streak from ORIGINAL game, not updated)
	newGame, err := daily_challenge.NewDailyGame(
		playerID,
		dailyQuiz.ID(),
		originalGame.Date(),
		quizAggregate,
		originalGame.Streak(), // Use streak from first attempt
		now,
	)
	if err != nil {
		return RetryChallengeOutput{}, err
	}

	// 9. Save new game
	if err := uc.dailyGameRepo.Save(newGame); err != nil {
		return RetryChallengeOutput{}, err
	}

	// 10. Publish events
	for _, event := range newGame.Events() {
		uc.eventBus.Publish(event)
	}

	// 11. Get first question
	firstQuestion, err := newGame.Session().GetCurrentQuestion()
	if err != nil {
		return RetryChallengeOutput{}, err
	}

	return RetryChallengeOutput{
		NewGameID:      newGame.ID().String(),
		FirstQuestion:  ToQuestionDTO(firstQuestion),
		CoinsDeducted:  coinsDeducted,
		RemainingCoins: 0, // TODO: get from user account
		TimeLimit:      15,
	}, nil
}
