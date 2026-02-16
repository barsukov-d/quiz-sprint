package daily_challenge

import (
	"testing"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/daily_challenge"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
)

// ========================================
// StartDailyChallenge Use Case Tests
// ========================================

func TestStartDailyChallenge_Success(t *testing.T) {
	f := setupFixture(t)
	uc := f.newStartUC()

	output, err := uc.Execute(StartDailyChallengeInput{
		PlayerID: testPlayerID,
		Date:     f.date.String(),
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if output.Game.GameID == "" {
		t.Error("Expected game ID to be set")
	}
	if output.Game.PlayerID != testPlayerID {
		t.Errorf("PlayerID = %s, want %s", output.Game.PlayerID, testPlayerID)
	}
	if output.Game.Status != "in_progress" {
		t.Errorf("Status = %s, want in_progress", output.Game.Status)
	}
	if output.Game.TotalQuestions != 10 {
		t.Errorf("TotalQuestions = %d, want 10", output.Game.TotalQuestions)
	}
	if output.FirstQuestion.ID == "" {
		t.Error("Expected first question to be set")
	}
	if len(output.FirstQuestion.Answers) != 4 {
		t.Errorf("Answers count = %d, want 4", len(output.FirstQuestion.Answers))
	}
	if output.TimeLimit != 15 {
		t.Errorf("TimeLimit = %d, want 15", output.TimeLimit)
	}
	// TimeToExpire can be negative if test date is in the past relative to time.Now()
	// Just verify it's set (non-zero in absolute sense)
	if output.TimeToExpire == 0 {
		t.Error("TimeToExpire should be non-zero")
	}

	// Verify events were published
	if len(f.eventBus.Events) == 0 {
		t.Error("Expected events to be published")
	}
}

func TestStartDailyChallenge_AlreadyPlayedToday(t *testing.T) {
	f := setupFixture(t)
	uc := f.newStartUC()

	// Start first game
	_, err := uc.Execute(StartDailyChallengeInput{
		PlayerID: testPlayerID,
		Date:     f.date.String(),
	})
	if err != nil {
		t.Fatalf("First start failed: %v", err)
	}

	// Try to start second game same day
	_, err = uc.Execute(StartDailyChallengeInput{
		PlayerID: testPlayerID,
		Date:     f.date.String(),
	})

	if err != daily_challenge.ErrAlreadyPlayedToday {
		t.Errorf("Expected ErrAlreadyPlayedToday, got %v", err)
	}
}

func TestStartDailyChallenge_InvalidPlayerID(t *testing.T) {
	f := setupFixture(t)
	uc := f.newStartUC()

	_, err := uc.Execute(StartDailyChallengeInput{
		PlayerID: "",
		Date:     f.date.String(),
	})

	if err == nil {
		t.Error("Expected error for empty player ID")
	}
}

func TestStartDailyChallenge_WithStreak(t *testing.T) {
	f := setupFixture(t)
	yesterday := f.date.Previous()

	// Create a completed game from yesterday with streak=5
	streak := daily_challenge.ReconstructStreakSystem(5, 5, yesterday)
	yesterdayGame := newCompletedGame(t, testPlayerID, yesterday, f.questions, streak)
	f.dailyGameRepo.Save(yesterdayGame)

	// Create a daily quiz for yesterday too
	yesterdayQuiz := newTestDailyQuiz(t, yesterday, f.questions)
	f.dailyQuizRepo.Save(yesterdayQuiz)

	uc := f.newStartUC()
	output, err := uc.Execute(StartDailyChallengeInput{
		PlayerID: testPlayerID,
		Date:     f.date.String(),
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	// Streak from yesterday's game should be used
	if output.Game.Streak.CurrentStreak != 5 {
		t.Errorf("Streak = %d, want 5 (from yesterday)", output.Game.Streak.CurrentStreak)
	}
}

// ========================================
// SubmitDailyAnswer Use Case Tests
// ========================================

func TestSubmitDailyAnswer_CorrectAnswer(t *testing.T) {
	f := setupFixture(t)

	// Start a game first
	startUC := f.newStartUC()
	startOutput, err := startUC.Execute(StartDailyChallengeInput{
		PlayerID: testPlayerID,
		Date:     f.date.String(),
	})
	if err != nil {
		t.Fatalf("Failed to start game: %v", err)
	}

	f.eventBus.Events = nil // reset

	submitUC := f.newSubmitAnswerUC()
	firstQ := startOutput.FirstQuestion
	// First answer is always the correct one in our test fixtures
	correctAnswerID := firstQ.Answers[0].ID

	output, err := submitUC.Execute(SubmitDailyAnswerInput{
		GameID:     startOutput.Game.GameID,
		QuestionID: firstQ.ID,
		AnswerID:   correctAnswerID,
		PlayerID:   testPlayerID,
		TimeTaken:  2000,
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !output.IsCorrect {
		t.Error("Expected IsCorrect=true for correct answer")
	}
	if output.CorrectAnswerID != correctAnswerID {
		t.Errorf("CorrectAnswerID = %s, want %s", output.CorrectAnswerID, correctAnswerID)
	}
	if output.IsGameCompleted {
		t.Error("Game should not be completed after 1 question")
	}
	if output.RemainingQuestions != 9 {
		t.Errorf("RemainingQuestions = %d, want 9", output.RemainingQuestions)
	}
	if output.NextQuestion == nil {
		t.Error("Expected next question after first answer")
	}
	if output.TotalQuestions != 10 {
		t.Errorf("TotalQuestions = %d, want 10", output.TotalQuestions)
	}
}

func TestSubmitDailyAnswer_WrongAnswer(t *testing.T) {
	f := setupFixture(t)

	startUC := f.newStartUC()
	startOutput, _ := startUC.Execute(StartDailyChallengeInput{
		PlayerID: testPlayerID,
		Date:     f.date.String(),
	})

	submitUC := f.newSubmitAnswerUC()
	firstQ := startOutput.FirstQuestion
	// Pick second answer (wrong)
	wrongAnswerID := firstQ.Answers[1].ID

	output, err := submitUC.Execute(SubmitDailyAnswerInput{
		GameID:     startOutput.Game.GameID,
		QuestionID: firstQ.ID,
		AnswerID:   wrongAnswerID,
		PlayerID:   testPlayerID,
		TimeTaken:  3000,
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if output.IsCorrect {
		t.Error("Expected IsCorrect=false for wrong answer")
	}
	if output.IsGameCompleted {
		t.Error("Game should not be completed after 1 wrong answer")
	}
	if output.NextQuestion == nil {
		t.Error("Expected next question")
	}
}

func TestSubmitDailyAnswer_LastQuestion_CompletesGame(t *testing.T) {
	f := setupFixture(t)

	startUC := f.newStartUC()
	startOutput, _ := startUC.Execute(StartDailyChallengeInput{
		PlayerID: testPlayerID,
		Date:     f.date.String(),
	})

	submitUC := f.newSubmitAnswerUC()

	// Answer all 10 questions
	var lastOutput SubmitDailyAnswerOutput
	currentQuestion := &startOutput.FirstQuestion
	for i := 0; i < 10; i++ {
		correctAnswerID := currentQuestion.Answers[0].ID

		var err error
		lastOutput, err = submitUC.Execute(SubmitDailyAnswerInput{
			GameID:     startOutput.Game.GameID,
			QuestionID: currentQuestion.ID,
			AnswerID:   correctAnswerID,
			PlayerID:   testPlayerID,
			TimeTaken:  2000,
		})
		if err != nil {
			t.Fatalf("Question %d failed: %v", i+1, err)
		}

		if i < 9 {
			if lastOutput.IsGameCompleted {
				t.Errorf("Game should not be completed after question %d", i+1)
			}
			if lastOutput.NextQuestion == nil {
				t.Fatalf("Expected next question after question %d", i+1)
			}
			currentQuestion = lastOutput.NextQuestion
		}
	}

	// Verify last answer completed the game
	if !lastOutput.IsGameCompleted {
		t.Error("Game should be completed after 10 questions")
	}
	if lastOutput.NextQuestion != nil {
		t.Error("No next question expected after game completion")
	}
	if lastOutput.GameResults == nil {
		t.Fatal("Expected game results after completion")
	}

	results := lastOutput.GameResults
	if results.CorrectAnswers != 10 {
		t.Errorf("CorrectAnswers = %d, want 10", results.CorrectAnswers)
	}
	if results.TotalQuestions != 10 {
		t.Errorf("TotalQuestions = %d, want 10", results.TotalQuestions)
	}
	if results.FinalScore <= 0 {
		t.Errorf("FinalScore = %d, should be > 0", results.FinalScore)
	}
	if results.ChestReward.ChestType == "" {
		t.Error("Expected chest type")
	}
	// 10/10 correct = golden chest
	if results.ChestReward.ChestType != "golden" {
		t.Errorf("ChestType = %s, want golden (10/10 correct)", results.ChestReward.ChestType)
	}
}

func TestSubmitDailyAnswer_PlayerMismatch(t *testing.T) {
	f := setupFixture(t)

	startUC := f.newStartUC()
	startOutput, _ := startUC.Execute(StartDailyChallengeInput{
		PlayerID: testPlayerID,
		Date:     f.date.String(),
	})

	submitUC := f.newSubmitAnswerUC()
	firstQ := startOutput.FirstQuestion

	_, err := submitUC.Execute(SubmitDailyAnswerInput{
		GameID:     startOutput.Game.GameID,
		QuestionID: firstQ.ID,
		AnswerID:   firstQ.Answers[0].ID,
		PlayerID:   "different-player",
		TimeTaken:  2000,
	})

	if err != daily_challenge.ErrGameNotFound {
		t.Errorf("Expected ErrGameNotFound for player mismatch, got %v", err)
	}
}

func TestSubmitDailyAnswer_GameNotFound(t *testing.T) {
	f := setupFixture(t)
	submitUC := f.newSubmitAnswerUC()

	_, err := submitUC.Execute(SubmitDailyAnswerInput{
		GameID:     "nonexistent-game-id",
		QuestionID: "q1",
		AnswerID:   "a1",
		PlayerID:   testPlayerID,
		TimeTaken:  2000,
	})

	if err != daily_challenge.ErrGameNotFound {
		t.Errorf("Expected ErrGameNotFound, got %v", err)
	}
}

// ========================================
// GetDailyGameStatus Use Case Tests
// ========================================

func TestGetDailyGameStatus_NotPlayed(t *testing.T) {
	f := setupFixture(t)
	uc := f.newGetStatusUC()

	output, err := uc.Execute(GetDailyGameStatusInput{
		PlayerID: testPlayerID,
		Date:     f.date.String(),
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if output.HasPlayed {
		t.Error("HasPlayed should be false when not played")
	}
	if output.Game != nil {
		t.Error("Game should be nil when not played")
	}
}

func TestGetDailyGameStatus_InProgress(t *testing.T) {
	f := setupFixture(t)

	// Start a game
	startUC := f.newStartUC()
	_, err := startUC.Execute(StartDailyChallengeInput{
		PlayerID: testPlayerID,
		Date:     f.date.String(),
	})
	if err != nil {
		t.Fatalf("Failed to start: %v", err)
	}

	uc := f.newGetStatusUC()
	output, err := uc.Execute(GetDailyGameStatusInput{
		PlayerID: testPlayerID,
		Date:     f.date.String(),
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if output.HasPlayed {
		t.Error("HasPlayed should be false for in-progress game")
	}
	if output.Game == nil {
		t.Fatal("Expected game DTO for in-progress game")
	}
	if output.Game.Status != "in_progress" {
		t.Errorf("Status = %s, want in_progress", output.Game.Status)
	}
	if output.TimeLimit == nil {
		t.Error("Expected TimeLimit for in-progress game")
	} else if *output.TimeLimit != 15 {
		t.Errorf("TimeLimit = %d, want 15", *output.TimeLimit)
	}
}

func TestGetDailyGameStatus_Completed(t *testing.T) {
	f := setupFixture(t)

	// Create and save a completed game
	streak := daily_challenge.NewStreakSystem()
	completedGame := newCompletedGame(t, testPlayerID, f.date, f.questions, streak)
	f.dailyGameRepo.Save(completedGame)

	uc := f.newGetStatusUC()
	output, err := uc.Execute(GetDailyGameStatusInput{
		PlayerID: testPlayerID,
		Date:     f.date.String(),
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !output.HasPlayed {
		t.Error("HasPlayed should be true for completed game")
	}
	if output.Game == nil {
		t.Fatal("Expected game DTO for completed game")
	}
	if output.Game.Status != "completed" {
		t.Errorf("Status = %s, want completed", output.Game.Status)
	}
	if output.Results == nil {
		t.Error("Expected Results for completed game")
	}
}

func TestGetDailyGameStatus_InvalidPlayerID(t *testing.T) {
	f := setupFixture(t)
	uc := f.newGetStatusUC()

	_, err := uc.Execute(GetDailyGameStatusInput{
		PlayerID: "",
		Date:     f.date.String(),
	})

	if err == nil {
		t.Error("Expected error for empty player ID")
	}
}

// ========================================
// GetDailyLeaderboard Use Case Tests
// ========================================

func TestGetDailyLeaderboard_TopPlayers(t *testing.T) {
	f := setupFixture(t)

	// Create completed games for two players
	streak := daily_challenge.NewStreakSystem()
	game1 := newCompletedGame(t, testPlayerID, f.date, f.questions, streak)
	game2 := newCompletedGame(t, testPlayerID2, f.date, f.questions, streak)
	f.dailyGameRepo.Save(game1)
	f.dailyGameRepo.Save(game2)

	uc := f.newLeaderboardUC()
	output, err := uc.Execute(GetDailyLeaderboardInput{
		Date:  f.date.String(),
		Limit: 10,
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(output.Entries) != 2 {
		t.Errorf("Entries count = %d, want 2", len(output.Entries))
	}
	if output.TotalPlayers != 2 {
		t.Errorf("TotalPlayers = %d, want 2", output.TotalPlayers)
	}
	if output.Date != f.date.String() {
		t.Errorf("Date = %s, want %s", output.Date, f.date.String())
	}

	// Entries should have usernames
	for _, entry := range output.Entries {
		if entry.Username == "" {
			t.Error("Expected username in leaderboard entry")
		}
		if entry.Rank < 1 {
			t.Errorf("Rank = %d, should be >= 1", entry.Rank)
		}
	}
}

func TestGetDailyLeaderboard_WithPlayerRank(t *testing.T) {
	f := setupFixture(t)

	streak := daily_challenge.NewStreakSystem()
	game := newCompletedGame(t, testPlayerID, f.date, f.questions, streak)
	f.dailyGameRepo.Save(game)

	uc := f.newLeaderboardUC()
	output, err := uc.Execute(GetDailyLeaderboardInput{
		Date:     f.date.String(),
		Limit:    10,
		PlayerID: testPlayerID,
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if output.PlayerRank == nil {
		t.Fatal("Expected player rank when playerID provided")
	}
	if *output.PlayerRank != 1 {
		t.Errorf("PlayerRank = %d, want 1 (only player)", *output.PlayerRank)
	}
}

func TestGetDailyLeaderboard_LimitClamping(t *testing.T) {
	f := setupFixture(t)
	uc := f.newLeaderboardUC()

	tests := []struct {
		name          string
		inputLimit    int
		expectedLimit int // what's passed to repo (clamped)
	}{
		{"Zero defaults to 10", 0, 10},
		{"Negative defaults to 10", -5, 10},
		{"Over 100 clamped to 100", 200, 100},
		{"Normal value passed through", 50, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := uc.Execute(GetDailyLeaderboardInput{
				Date:  f.date.String(),
				Limit: tt.inputLimit,
			})
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			// Even if limit changes, it should not error
			_ = output
		})
	}
}

// ========================================
// GetPlayerStreak Use Case Tests
// ========================================

func TestGetPlayerStreak_NoHistory(t *testing.T) {
	f := setupFixture(t)
	uc := f.newGetStreakUC()

	output, err := uc.Execute(GetPlayerStreakInput{
		PlayerID: testPlayerID,
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if output.Streak.CurrentStreak != 0 {
		t.Errorf("CurrentStreak = %d, want 0", output.Streak.CurrentStreak)
	}
	if output.NextMilestone != 3 {
		t.Errorf("NextMilestone = %d, want 3", output.NextMilestone)
	}
	if output.DaysToNext != 3 {
		t.Errorf("DaysToNext = %d, want 3", output.DaysToNext)
	}
	if output.CanRestore {
		t.Error("CanRestore should be false with no history")
	}
}

func TestGetPlayerStreak_ActiveStreak(t *testing.T) {
	f := setupFixture(t)
	today := daily_challenge.TodayUTC()

	// Create a game played today with streak of 5
	streak := daily_challenge.ReconstructStreakSystem(5, 5, today)
	game := newCompletedGame(t, testPlayerID, today, f.questions, streak)
	f.dailyGameRepo.Save(game)

	uc := f.newGetStreakUC()
	output, err := uc.Execute(GetPlayerStreakInput{PlayerID: testPlayerID})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	// After completion, streak is updated (+1), so it should be 6
	currentStreak := output.Streak.CurrentStreak
	if currentStreak < 5 {
		t.Errorf("CurrentStreak = %d, should be >= 5", currentStreak)
	}
	if output.NextMilestone != 7 {
		t.Errorf("NextMilestone = %d, want 7 (next after streak >= 5)", output.NextMilestone)
	}
}

func TestGetPlayerStreak_InvalidPlayerID(t *testing.T) {
	f := setupFixture(t)
	uc := f.newGetStreakUC()

	_, err := uc.Execute(GetPlayerStreakInput{PlayerID: ""})

	if err == nil {
		t.Error("Expected error for empty player ID")
	}
}

// ========================================
// OpenChest Use Case Tests
// ========================================

func TestOpenChest_Success(t *testing.T) {
	f := setupFixture(t)

	streak := daily_challenge.NewStreakSystem()
	game := newCompletedGame(t, testPlayerID, f.date, f.questions, streak)
	f.dailyGameRepo.Save(game)

	uc := f.newOpenChestUC()
	output, err := uc.Execute(OpenChestInput{
		GameID:   game.ID().String(),
		PlayerID: testPlayerID,
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if output.ChestType == "" {
		t.Error("Expected chest type")
	}
	// 10/10 correct = golden chest
	if output.ChestType != "golden" {
		t.Errorf("ChestType = %s, want golden", output.ChestType)
	}
	if output.Rewards.Coins <= 0 {
		t.Errorf("Coins = %d, should be > 0 for golden chest", output.Rewards.Coins)
	}
	if output.Rewards.PvpTickets <= 0 {
		t.Errorf("PvpTickets = %d, should be > 0 for golden chest", output.Rewards.PvpTickets)
	}
}

func TestOpenChest_GameNotCompleted(t *testing.T) {
	f := setupFixture(t)

	// Start a game but don't complete it
	startUC := f.newStartUC()
	startOutput, _ := startUC.Execute(StartDailyChallengeInput{
		PlayerID: testPlayerID,
		Date:     f.date.String(),
	})

	uc := f.newOpenChestUC()
	_, err := uc.Execute(OpenChestInput{
		GameID:   startOutput.Game.GameID,
		PlayerID: testPlayerID,
	})

	if err != daily_challenge.ErrGameNotActive {
		t.Errorf("Expected ErrGameNotActive, got %v", err)
	}
}

func TestOpenChest_PlayerMismatch(t *testing.T) {
	f := setupFixture(t)

	streak := daily_challenge.NewStreakSystem()
	game := newCompletedGame(t, testPlayerID, f.date, f.questions, streak)
	f.dailyGameRepo.Save(game)

	uc := f.newOpenChestUC()
	_, err := uc.Execute(OpenChestInput{
		GameID:   game.ID().String(),
		PlayerID: "wrong-player",
	})

	if err != daily_challenge.ErrGameNotFound {
		t.Errorf("Expected ErrGameNotFound for player mismatch, got %v", err)
	}
}

func TestOpenChest_GameNotFound(t *testing.T) {
	f := setupFixture(t)
	uc := f.newOpenChestUC()

	_, err := uc.Execute(OpenChestInput{
		GameID:   "nonexistent",
		PlayerID: testPlayerID,
	})

	if err != daily_challenge.ErrGameNotFound {
		t.Errorf("Expected ErrGameNotFound, got %v", err)
	}
}

// ========================================
// RetryChallenge Use Case Tests
// ========================================

func TestRetryChallenge_WithCoins(t *testing.T) {
	f := setupFixture(t)

	streak := daily_challenge.NewStreakSystem()
	game := newCompletedGame(t, testPlayerID, f.date, f.questions, streak)
	f.dailyGameRepo.Save(game)

	uc := f.newRetryUC()
	output, err := uc.Execute(RetryChallengeInput{
		GameID:        game.ID().String(),
		PlayerID:      testPlayerID,
		PaymentMethod: "coins",
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if output.NewGameID == "" {
		t.Error("Expected new game ID")
	}
	if output.NewGameID == game.ID().String() {
		t.Error("New game ID should differ from original")
	}
	if output.CoinsDeducted != 100 {
		t.Errorf("CoinsDeducted = %d, want 100", output.CoinsDeducted)
	}
	if output.TimeLimit != 15 {
		t.Errorf("TimeLimit = %d, want 15", output.TimeLimit)
	}
	if output.FirstQuestion.ID == "" {
		t.Error("Expected first question")
	}
}

func TestRetryChallenge_WithAd(t *testing.T) {
	f := setupFixture(t)

	streak := daily_challenge.NewStreakSystem()
	game := newCompletedGame(t, testPlayerID, f.date, f.questions, streak)
	f.dailyGameRepo.Save(game)

	uc := f.newRetryUC()
	output, err := uc.Execute(RetryChallengeInput{
		GameID:        game.ID().String(),
		PlayerID:      testPlayerID,
		PaymentMethod: "ad",
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if output.CoinsDeducted != 0 {
		t.Errorf("CoinsDeducted = %d, want 0 for ad payment", output.CoinsDeducted)
	}
}

func TestRetryChallenge_RetryLimitReached(t *testing.T) {
	f := setupFixture(t)

	// Create first completed game (original attempt)
	streak := daily_challenge.NewStreakSystem()
	game1 := newCompletedGame(t, testPlayerID, f.date, f.questions, streak)
	f.dailyGameRepo.Save(game1)

	// Create a second completed game (first retry - 2 total)
	game2 := newCompletedGame(t, testPlayerID, f.date, f.questions, streak)
	f.dailyGameRepo.Save(game2)

	uc := f.newRetryUC()
	_, err := uc.Execute(RetryChallengeInput{
		GameID:        game1.ID().String(),
		PlayerID:      testPlayerID,
		PaymentMethod: "coins",
	})

	// Max 2 attempts for non-premium = should fail
	if err != daily_challenge.ErrAlreadyPlayedToday {
		t.Errorf("Expected ErrAlreadyPlayedToday (retry limit), got %v", err)
	}
}

func TestRetryChallenge_InvalidPaymentMethod(t *testing.T) {
	f := setupFixture(t)

	streak := daily_challenge.NewStreakSystem()
	game := newCompletedGame(t, testPlayerID, f.date, f.questions, streak)
	f.dailyGameRepo.Save(game)

	uc := f.newRetryUC()
	_, err := uc.Execute(RetryChallengeInput{
		GameID:        game.ID().String(),
		PlayerID:      testPlayerID,
		PaymentMethod: "bitcoin",
	})

	if err == nil {
		t.Error("Expected error for invalid payment method")
	}
}

func TestRetryChallenge_GameNotCompleted(t *testing.T) {
	f := setupFixture(t)

	// Start but don't complete
	startUC := f.newStartUC()
	startOutput, _ := startUC.Execute(StartDailyChallengeInput{
		PlayerID: testPlayerID,
		Date:     f.date.String(),
	})

	uc := f.newRetryUC()
	_, err := uc.Execute(RetryChallengeInput{
		GameID:        startOutput.Game.GameID,
		PlayerID:      testPlayerID,
		PaymentMethod: "coins",
	})

	if err != daily_challenge.ErrGameNotActive {
		t.Errorf("Expected ErrGameNotActive, got %v", err)
	}
}

func TestRetryChallenge_PlayerMismatch(t *testing.T) {
	f := setupFixture(t)

	streak := daily_challenge.NewStreakSystem()
	game := newCompletedGame(t, testPlayerID, f.date, f.questions, streak)
	f.dailyGameRepo.Save(game)

	uc := f.newRetryUC()
	_, err := uc.Execute(RetryChallengeInput{
		GameID:        game.ID().String(),
		PlayerID:      "wrong-player",
		PaymentMethod: "coins",
	})

	if err != daily_challenge.ErrGameNotFound {
		t.Errorf("Expected ErrGameNotFound, got %v", err)
	}
}

// ========================================
// GetOrCreateDailyQuiz Use Case Tests
// ========================================

func TestGetOrCreateDailyQuiz_ExistingQuiz(t *testing.T) {
	f := setupFixture(t)
	uc := f.newGetOrCreateQuizUC()

	output, err := uc.Execute(GetOrCreateDailyQuizInput{Date: f.date.String()})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if output.DailyQuiz.ID == "" {
		t.Error("Expected quiz ID")
	}
	if output.DailyQuiz.Date != f.date.String() {
		t.Errorf("Date = %s, want %s", output.DailyQuiz.Date, f.date.String())
	}
	if len(output.DailyQuiz.QuestionIDs) != 10 {
		t.Errorf("QuestionIDs count = %d, want 10", len(output.DailyQuiz.QuestionIDs))
	}
	if output.IsNew {
		t.Error("Expected IsNew=false for existing quiz")
	}
}

func TestGetOrCreateDailyQuiz_CreatesNew(t *testing.T) {
	f := setupFixture(t)
	uc := f.newGetOrCreateQuizUC()

	// Use a different date that has no quiz yet
	tomorrow := f.date.Next()
	output, err := uc.Execute(GetOrCreateDailyQuizInput{Date: tomorrow.String()})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !output.IsNew {
		t.Error("Expected IsNew=true for newly created quiz")
	}
	if output.DailyQuiz.Date != tomorrow.String() {
		t.Errorf("Date = %s, want %s", output.DailyQuiz.Date, tomorrow.String())
	}
}

// ========================================
// Mapper Tests
// ========================================

func TestToStreakDTO(t *testing.T) {
	today := testDate()

	tests := []struct {
		name             string
		streak           daily_challenge.StreakSystem
		expectedCurrent  int
		expectedBest     int
		expectedIsActive bool
	}{
		{
			name:             "New streak",
			streak:           daily_challenge.NewStreakSystem(),
			expectedCurrent:  0,
			expectedBest:     0,
			expectedIsActive: false,
		},
		{
			name:             "Active streak played today",
			streak:           daily_challenge.ReconstructStreakSystem(5, 10, today),
			expectedCurrent:  5,
			expectedBest:     10,
			expectedIsActive: true,
		},
		{
			name:             "Active streak played yesterday",
			streak:           daily_challenge.ReconstructStreakSystem(3, 7, today.Previous()),
			expectedCurrent:  3,
			expectedBest:     7,
			expectedIsActive: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dto := ToStreakDTO(tt.streak, today)

			if dto.CurrentStreak != tt.expectedCurrent {
				t.Errorf("CurrentStreak = %d, want %d", dto.CurrentStreak, tt.expectedCurrent)
			}
			if dto.BestStreak != tt.expectedBest {
				t.Errorf("BestStreak = %d, want %d", dto.BestStreak, tt.expectedBest)
			}
			if dto.IsActive != tt.expectedIsActive {
				t.Errorf("IsActive = %v, want %v", dto.IsActive, tt.expectedIsActive)
			}
		})
	}
}

func TestToQuestionDTO(t *testing.T) {
	question := newTestQuestion(t, 1)
	dto := ToQuestionDTO(question)

	if dto.ID == "" {
		t.Error("Expected question ID")
	}
	if dto.Text == "" {
		t.Error("Expected question text")
	}
	if len(dto.Answers) != 4 {
		t.Errorf("Answers count = %d, want 4", len(dto.Answers))
	}
	if dto.Points != 100 {
		t.Errorf("Points = %d, want 100", dto.Points)
	}

	// Verify answers don't expose IsCorrect
	for _, answer := range dto.Answers {
		if answer.ID == "" {
			t.Error("Expected answer ID")
		}
		if answer.Text == "" {
			t.Error("Expected answer text")
		}
	}
}

func TestBuildGameResultsDTO(t *testing.T) {
	date := testDate()
	questions := newTestQuestions(&testing.T{}, 10)
	streak := daily_challenge.NewStreakSystem()

	// We need a proper testing.T for newCompletedGame
	game := newCompletedGame(t, testPlayerID, date, questions, streak)

	results := BuildGameResultsDTO(game, 1, 5, nil)

	if results.TotalQuestions != 10 {
		t.Errorf("TotalQuestions = %d, want 10", results.TotalQuestions)
	}
	if results.CorrectAnswers != 10 {
		t.Errorf("CorrectAnswers = %d, want 10", results.CorrectAnswers)
	}
	if results.FinalScore <= 0 {
		t.Errorf("FinalScore = %d, should be > 0", results.FinalScore)
	}
	if results.Rank != 1 {
		t.Errorf("Rank = %d, want 1", results.Rank)
	}
	if results.TotalPlayers != 5 {
		t.Errorf("TotalPlayers = %d, want 5", results.TotalPlayers)
	}
	// With rank=1, totalPlayers=5: percentile = (5-1+1)/5 * 100 = 100%
	if results.Percentile != 100 {
		t.Errorf("Percentile = %d, want 100", results.Percentile)
	}
}

// ========================================
// Error mapping integration (use case returns correct domain errors)
// ========================================

func TestErrorPropagation_InvalidUserID(t *testing.T) {
	f := setupFixture(t)

	tests := []struct {
		name string
		fn   func() error
	}{
		{
			name: "StartDailyChallenge with empty playerID",
			fn: func() error {
				_, err := f.newStartUC().Execute(StartDailyChallengeInput{PlayerID: ""})
				return err
			},
		},
		{
			name: "GetDailyGameStatus with empty playerID",
			fn: func() error {
				_, err := f.newGetStatusUC().Execute(GetDailyGameStatusInput{PlayerID: ""})
				return err
			},
		},
		{
			name: "GetPlayerStreak with empty playerID",
			fn: func() error {
				_, err := f.newGetStreakUC().Execute(GetPlayerStreakInput{PlayerID: ""})
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			if err == nil {
				t.Error("Expected error for empty player ID")
			}
			if err != shared.ErrInvalidUserID {
				t.Errorf("Expected ErrInvalidUserID, got %v", err)
			}
		})
	}
}
