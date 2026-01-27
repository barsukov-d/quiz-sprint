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
	dailyGameRepo daily_challenge.DailyGameRepository
	eventBus      EventBus
}

func NewSubmitDailyAnswerUseCase(
	dailyGameRepo daily_challenge.DailyGameRepository,
	eventBus EventBus,
) *SubmitDailyAnswerUseCase {
	return &SubmitDailyAnswerUseCase{
		dailyGameRepo: dailyGameRepo,
		eventBus:      eventBus,
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

	// 6. Build output
	output := SubmitDailyAnswerOutput{
		QuestionIndex:      result.QuestionIndex,
		TotalQuestions:     10,
		RemainingQuestions: result.RemainingQuestions,
		IsGameCompleted:    result.IsGameCompleted,
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
		// Game completed - calculate rank and build results
		rank, _ := uc.dailyGameRepo.GetPlayerRankByDate(game.PlayerID(), game.Date())
		totalPlayers, _ := uc.dailyGameRepo.GetTotalPlayersByDate(game.Date())

		game.SetRank(rank)
		uc.dailyGameRepo.Save(game) // Update with rank

		results := BuildGameResultsDTO(game, rank, totalPlayers)
		output.GameResults = &results
	}

	return output, nil
}

// ========================================
// GetDailyGameStatus Use Case
// ========================================

type GetDailyGameStatusUseCase struct {
	dailyQuizRepo daily_challenge.DailyQuizRepository
	dailyGameRepo daily_challenge.DailyGameRepository
}

func NewGetDailyGameStatusUseCase(
	dailyQuizRepo daily_challenge.DailyQuizRepository,
	dailyGameRepo daily_challenge.DailyGameRepository,
) *GetDailyGameStatusUseCase {
	return &GetDailyGameStatusUseCase{
		dailyQuizRepo: dailyQuizRepo,
		dailyGameRepo: dailyGameRepo,
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
	if game.Status() == daily_challenge.GameStatusInProgress {
		tl := 15
		timeLimit = &tl
	}

	// HasPlayed = true ONLY if game is completed
	hasPlayed := game.Status() == daily_challenge.GameStatusCompleted

	return GetDailyGameStatusOutput{
		HasPlayed:    hasPlayed,
		Game:         &gameDTO,
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

	// Calculate next milestone
	milestones := []int{3, 7, 14, 30, 100}
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
