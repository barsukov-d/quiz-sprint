package marathon

import (
	"testing"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/solo_marathon"
)

// ========================================
// StartMarathon Use Case Tests
// ========================================

func TestStartMarathon_Success_AllCategories(t *testing.T) {
	f := setupFixture(t)
	uc := f.newStartUC()

	output, err := uc.Execute(StartMarathonInput{
		PlayerID: testPlayerID,
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if output.Game.ID == "" {
		t.Error("Expected game ID")
	}
	if output.Game.PlayerID != testPlayerID {
		t.Errorf("PlayerID = %s, want %s", output.Game.PlayerID, testPlayerID)
	}
	if output.Game.Status != "in_progress" {
		t.Errorf("Status = %s, want in_progress", output.Game.Status)
	}
	if !output.Game.Category.IsAllCategories {
		t.Error("Expected all categories mode")
	}
	if output.Game.CurrentQuestion == nil {
		t.Error("Expected current question after start")
	}
	if output.Game.Lives.CurrentLives != 5 {
		t.Errorf("Lives = %d, want 5", output.Game.Lives.CurrentLives)
	}
	if output.Game.Score != 0 {
		t.Errorf("Score = %d, want 0", output.Game.Score)
	}
	if output.HasPersonalBest {
		t.Error("Expected HasPersonalBest=false for first game")
	}

	// Verify events were published
	if len(f.eventBus.events) == 0 {
		t.Error("Expected events to be published")
	}
}

func TestStartMarathon_DefaultBonuses(t *testing.T) {
	f := setupFixture(t)
	output := f.startGameForPlayer(t, testPlayerID)

	// Default bonuses: shield=2, fiftyFifty=1, skip=0, freeze=3
	b := output.Game.BonusInventory
	if b.Shield != 2 {
		t.Errorf("Shield = %d, want 2", b.Shield)
	}
	if b.FiftyFifty != 1 {
		t.Errorf("FiftyFifty = %d, want 1", b.FiftyFifty)
	}
	if b.Skip != 0 {
		t.Errorf("Skip = %d, want 0", b.Skip)
	}
	if b.Freeze != 3 {
		t.Errorf("Freeze = %d, want 3", b.Freeze)
	}
}

func TestStartMarathon_WithWalletBonuses(t *testing.T) {
	f := setupFixture(t)

	// Pre-fill bonus wallet (from daily challenge rewards)
	wallet := solo_marathon.ReconstructBonusWallet(mustUserID(testPlayerID), 1, 2, 3, 1)
	f.bonusWalletRepo.Save(wallet)

	output := f.startGameForPlayer(t, testPlayerID)

	// defaults(2,1,0,3) + wallet(1,2,3,1) = (3,3,3,4)
	b := output.Game.BonusInventory
	if b.Shield != 3 {
		t.Errorf("Shield = %d, want 3 (2+1)", b.Shield)
	}
	if b.FiftyFifty != 3 {
		t.Errorf("FiftyFifty = %d, want 3 (1+2)", b.FiftyFifty)
	}
	if b.Skip != 3 {
		t.Errorf("Skip = %d, want 3 (0+3)", b.Skip)
	}
	if b.Freeze != 4 {
		t.Errorf("Freeze = %d, want 4 (3+1)", b.Freeze)
	}

	// Wallet should be consumed (zeroed)
	savedWallet, _ := f.bonusWalletRepo.FindByPlayer(mustUserID(testPlayerID))
	if savedWallet != nil && !savedWallet.IsEmpty() {
		t.Error("Wallet should be empty after starting marathon")
	}
}

func TestStartMarathon_ActiveGameExists(t *testing.T) {
	f := setupFixture(t)

	// Start first game
	f.startGameForPlayer(t, testPlayerID)

	// Try to start second
	uc := f.newStartUC()
	_, err := uc.Execute(StartMarathonInput{PlayerID: testPlayerID})

	if err != solo_marathon.ErrActiveGameExists {
		t.Errorf("Expected ErrActiveGameExists, got %v", err)
	}
}

func TestStartMarathon_InvalidPlayerID(t *testing.T) {
	f := setupFixture(t)
	uc := f.newStartUC()

	_, err := uc.Execute(StartMarathonInput{PlayerID: ""})

	if err == nil {
		t.Error("Expected error for empty player ID")
	}
}

func TestStartMarathon_WithPersonalBest(t *testing.T) {
	f := setupFixture(t)

	// Create existing personal best
	category := solo_marathon.NewMarathonCategoryAll()
	pb, _ := solo_marathon.NewPersonalBest(mustUserID(testPlayerID), category, 15, 15, 1000000)
	f.personalBestRepo.Save(pb)

	output := f.startGameForPlayer(t, testPlayerID)

	if !output.HasPersonalBest {
		t.Error("Expected HasPersonalBest=true")
	}
	if output.Game.PersonalBest == nil {
		t.Error("Expected PersonalBest to be set")
	} else if *output.Game.PersonalBest != 15 {
		t.Errorf("PersonalBest = %d, want 15", *output.Game.PersonalBest)
	}
}

// ========================================
// SubmitMarathonAnswer Use Case Tests
// ========================================

func TestSubmitAnswer_Correct(t *testing.T) {
	f := setupFixture(t)
	startOutput := f.startGameForPlayer(t, testPlayerID)

	output := f.answerCurrentQuestion(t, startOutput.Game.ID, testPlayerID, true)

	if !output.IsCorrect {
		t.Error("Expected IsCorrect=true")
	}
	if output.Score != 1 {
		t.Errorf("Score = %d, want 1", output.Score)
	}
	if output.IsGameOver {
		t.Error("Game should not be over after 1 correct answer")
	}
	if output.NextQuestion == nil {
		t.Error("Expected next question")
	}
	if output.Lives.CurrentLives != 5 {
		t.Errorf("Lives = %d, want 5 (no life lost)", output.Lives.CurrentLives)
	}
	if output.LifeLost {
		t.Error("No life should be lost on correct answer")
	}
}

func TestSubmitAnswer_Wrong_LosesLife(t *testing.T) {
	f := setupFixture(t)
	startOutput := f.startGameForPlayer(t, testPlayerID)

	output := f.answerCurrentQuestion(t, startOutput.Game.ID, testPlayerID, false)

	if output.IsCorrect {
		t.Error("Expected IsCorrect=false")
	}
	if !output.LifeLost {
		t.Error("Expected life to be lost")
	}
	if output.Lives.CurrentLives != 4 {
		t.Errorf("Lives = %d, want 4", output.Lives.CurrentLives)
	}
	if output.IsGameOver {
		t.Error("Game should not be over with 4 lives remaining")
	}
}

func TestSubmitAnswer_FiveWrong_GameOver(t *testing.T) {
	f := setupFixture(t)
	startOutput := f.startGameForPlayer(t, testPlayerID)
	gameID := startOutput.Game.ID

	// Answer wrong 5 times (lose all 5 lives)
	f.answerCurrentQuestion(t, gameID, testPlayerID, false) // 5→4 lives
	f.answerCurrentQuestion(t, gameID, testPlayerID, false) // 4→3 lives
	f.answerCurrentQuestion(t, gameID, testPlayerID, false) // 3→2 lives
	f.answerCurrentQuestion(t, gameID, testPlayerID, false) // 2→1 lives
	output := f.answerCurrentQuestion(t, gameID, testPlayerID, false) // 1→0 lives

	if !output.IsGameOver {
		t.Error("Expected game over after 5 wrong answers")
	}
	if output.Lives.CurrentLives != 0 {
		t.Errorf("Lives = %d, want 0", output.Lives.CurrentLives)
	}
	if output.GameOverResult == nil {
		t.Fatal("Expected GameOverResult")
	}
	if output.GameOverResult.ContinueOffer == nil {
		t.Fatal("Expected ContinueOffer in game over")
	}
	if !output.GameOverResult.ContinueOffer.Available {
		t.Error("Continue should be available on first game over")
	}
}

func TestSubmitAnswer_PlayerMismatch(t *testing.T) {
	f := setupFixture(t)
	startOutput := f.startGameForPlayer(t, testPlayerID)

	game, _ := f.marathonRepo.FindByID(solo_marathon.NewGameIDFromString(startOutput.Game.ID))
	currentQ, _ := game.GetCurrentQuestion()

	uc := f.newSubmitAnswerUC()
	_, err := uc.Execute(SubmitMarathonAnswerInput{
		GameID:     startOutput.Game.ID,
		QuestionID: currentQ.ID().String(),
		AnswerID:   currentQ.Answers()[0].ID().String(),
		PlayerID:   "wrong-player",
		TimeTaken:  2000,
	})

	if err == nil {
		t.Error("Expected error for player mismatch")
	}
}

func TestSubmitAnswer_GameNotFound(t *testing.T) {
	f := setupFixture(t)
	uc := f.newSubmitAnswerUC()

	// Use valid UUID formats that don't exist in repo
	fakeGameID := solo_marathon.NewGameID().String()
	fakeQuestionID := f.questions[0].ID().String()
	fakeAnswerID := f.questions[0].Answers()[0].ID().String()

	_, err := uc.Execute(SubmitMarathonAnswerInput{
		GameID:     fakeGameID,
		QuestionID: fakeQuestionID,
		AnswerID:   fakeAnswerID,
		PlayerID:   testPlayerID,
		TimeTaken:  2000,
	})

	if err != solo_marathon.ErrGameNotFound {
		t.Errorf("Expected ErrGameNotFound, got %v", err)
	}
}

func TestSubmitAnswer_ShieldAbsorbsWrongAnswer(t *testing.T) {
	f := setupFixture(t)
	startOutput := f.startGameForPlayer(t, testPlayerID)
	gameID := startOutput.Game.ID

	// Activate shield first
	game, _ := f.marathonRepo.FindByID(solo_marathon.NewGameIDFromString(gameID))
	currentQ, _ := game.GetCurrentQuestion()

	bonusUC := f.newUseBonusUC()
	_, err := bonusUC.Execute(UseMarathonBonusInput{
		GameID:     gameID,
		QuestionID: currentQ.ID().String(),
		BonusType:  "shield",
		PlayerID:   testPlayerID,
	})
	if err != nil {
		t.Fatalf("Failed to use shield: %v", err)
	}

	// Answer wrong - shield should absorb
	output := f.answerCurrentQuestion(t, gameID, testPlayerID, false)

	if output.IsCorrect {
		t.Error("Expected IsCorrect=false")
	}
	if !output.ShieldConsumed {
		t.Error("Expected ShieldConsumed=true")
	}
	if output.LifeLost {
		t.Error("Shield should prevent life loss")
	}
	if output.Lives.CurrentLives != 5 {
		t.Errorf("Lives = %d, want 5 (shield absorbed)", output.Lives.CurrentLives)
	}
}

func TestSubmitAnswer_DifficultyProgresses(t *testing.T) {
	f := setupFixture(t)
	startOutput := f.startGameForPlayer(t, testPlayerID)
	gameID := startOutput.Game.ID

	// Difficulty thresholds: 1-10 beginner, 11-30 medium
	// Answer 11 correct → question 12 should be medium
	var lastOutput SubmitMarathonAnswerOutput
	for i := 0; i < 11; i++ {
		lastOutput = f.answerCurrentQuestion(t, gameID, testPlayerID, true)
	}

	if lastOutput.DifficultyLevel != "medium" {
		t.Errorf("DifficultyLevel = %s, want medium after 11 questions", lastOutput.DifficultyLevel)
	}
}

func TestSubmitAnswer_InvalidTimeTaken(t *testing.T) {
	f := setupFixture(t)
	startOutput := f.startGameForPlayer(t, testPlayerID)

	game, _ := f.marathonRepo.FindByID(solo_marathon.NewGameIDFromString(startOutput.Game.ID))
	currentQ, _ := game.GetCurrentQuestion()

	uc := f.newSubmitAnswerUC()
	_, err := uc.Execute(SubmitMarathonAnswerInput{
		GameID:     startOutput.Game.ID,
		QuestionID: currentQ.ID().String(),
		AnswerID:   currentQ.Answers()[0].ID().String(),
		PlayerID:   testPlayerID,
		TimeTaken:  -1,
	})

	if err == nil {
		t.Error("Expected error for negative timeTaken")
	}
}

// ========================================
// UseMarathonBonus Use Case Tests
// ========================================

func TestUseBonus_Shield(t *testing.T) {
	f := setupFixture(t)
	startOutput := f.startGameForPlayer(t, testPlayerID)
	gameID := startOutput.Game.ID

	game, _ := f.marathonRepo.FindByID(solo_marathon.NewGameIDFromString(gameID))
	currentQ, _ := game.GetCurrentQuestion()

	uc := f.newUseBonusUC()
	output, err := uc.Execute(UseMarathonBonusInput{
		GameID:     gameID,
		QuestionID: currentQ.ID().String(),
		BonusType:  "shield",
		PlayerID:   testPlayerID,
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if output.BonusType != "shield" {
		t.Errorf("BonusType = %s, want shield", output.BonusType)
	}
	if output.RemainingCount != 1 {
		t.Errorf("RemainingCount = %d, want 1 (started with 2)", output.RemainingCount)
	}
	if output.BonusResult.ShieldActive == nil || !*output.BonusResult.ShieldActive {
		t.Error("Expected ShieldActive=true")
	}
}

func TestUseBonus_FiftyFifty(t *testing.T) {
	f := setupFixture(t)
	startOutput := f.startGameForPlayer(t, testPlayerID)
	gameID := startOutput.Game.ID

	game, _ := f.marathonRepo.FindByID(solo_marathon.NewGameIDFromString(gameID))
	currentQ, _ := game.GetCurrentQuestion()

	uc := f.newUseBonusUC()
	output, err := uc.Execute(UseMarathonBonusInput{
		GameID:     gameID,
		QuestionID: currentQ.ID().String(),
		BonusType:  "fifty_fifty",
		PlayerID:   testPlayerID,
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(output.BonusResult.HiddenAnswerIDs) != 2 {
		t.Errorf("HiddenAnswerIDs count = %d, want 2", len(output.BonusResult.HiddenAnswerIDs))
	}
	if output.RemainingCount != 0 {
		t.Errorf("RemainingCount = %d, want 0 (started with 1)", output.RemainingCount)
	}
}

func TestUseBonus_Freeze(t *testing.T) {
	f := setupFixture(t)
	startOutput := f.startGameForPlayer(t, testPlayerID)
	gameID := startOutput.Game.ID

	game, _ := f.marathonRepo.FindByID(solo_marathon.NewGameIDFromString(gameID))
	currentQ, _ := game.GetCurrentQuestion()

	uc := f.newUseBonusUC()
	output, err := uc.Execute(UseMarathonBonusInput{
		GameID:     gameID,
		QuestionID: currentQ.ID().String(),
		BonusType:  "freeze",
		PlayerID:   testPlayerID,
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if output.BonusResult.NewTimeLimit == nil {
		t.Fatal("Expected NewTimeLimit for freeze bonus")
	}
	// First question: 15s base + 10s freeze = 25s
	if *output.BonusResult.NewTimeLimit != 25 {
		t.Errorf("NewTimeLimit = %d, want 25", *output.BonusResult.NewTimeLimit)
	}
	if output.RemainingCount != 2 {
		t.Errorf("RemainingCount = %d, want 2 (started with 3)", output.RemainingCount)
	}
}

func TestUseBonus_Skip(t *testing.T) {
	f := setupFixture(t)

	// Pre-fill wallet with skip bonuses (defaults have skip=0)
	wallet := solo_marathon.ReconstructBonusWallet(mustUserID(testPlayerID), 0, 0, 2, 0)
	f.bonusWalletRepo.Save(wallet)

	startOutput := f.startGameForPlayer(t, testPlayerID)
	gameID := startOutput.Game.ID

	game, _ := f.marathonRepo.FindByID(solo_marathon.NewGameIDFromString(gameID))
	currentQ, _ := game.GetCurrentQuestion()
	originalQuestionID := currentQ.ID().String()

	uc := f.newUseBonusUC()
	output, err := uc.Execute(UseMarathonBonusInput{
		GameID:     gameID,
		QuestionID: currentQ.ID().String(),
		BonusType:  "skip",
		PlayerID:   testPlayerID,
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if output.BonusResult.NextQuestion == nil {
		t.Fatal("Expected NextQuestion for skip bonus")
	}
	if output.BonusResult.NextQuestion.ID == originalQuestionID {
		t.Error("Skip should load a different question")
	}
}

func TestUseBonus_InvalidType(t *testing.T) {
	f := setupFixture(t)
	startOutput := f.startGameForPlayer(t, testPlayerID)

	game, _ := f.marathonRepo.FindByID(solo_marathon.NewGameIDFromString(startOutput.Game.ID))
	currentQ, _ := game.GetCurrentQuestion()

	uc := f.newUseBonusUC()
	_, err := uc.Execute(UseMarathonBonusInput{
		GameID:     startOutput.Game.ID,
		QuestionID: currentQ.ID().String(),
		BonusType:  "teleport",
		PlayerID:   testPlayerID,
	})

	if err != solo_marathon.ErrInvalidBonusType {
		t.Errorf("Expected ErrInvalidBonusType, got %v", err)
	}
}

func TestUseBonus_NoBonusesLeft(t *testing.T) {
	f := setupFixture(t)
	startOutput := f.startGameForPlayer(t, testPlayerID)
	gameID := startOutput.Game.ID

	uc := f.newUseBonusUC()

	// Use the only fifty_fifty bonus (default=1)
	game, _ := f.marathonRepo.FindByID(solo_marathon.NewGameIDFromString(gameID))
	currentQ, _ := game.GetCurrentQuestion()

	output, err := uc.Execute(UseMarathonBonusInput{
		GameID:     gameID,
		QuestionID: currentQ.ID().String(),
		BonusType:  "fifty_fifty",
		PlayerID:   testPlayerID,
	})
	if err != nil {
		t.Fatalf("First use should succeed: %v", err)
	}
	if output.RemainingCount != 0 {
		t.Errorf("RemainingCount = %d, want 0 after using only fifty_fifty", output.RemainingCount)
	}

	// Answer to move to next question
	f.answerCurrentQuestion(t, gameID, testPlayerID, true)

	// Try fifty_fifty again on new question — should fail (0 remaining)
	game, _ = f.marathonRepo.FindByID(solo_marathon.NewGameIDFromString(gameID))
	currentQ, _ = game.GetCurrentQuestion()

	_, err = uc.Execute(UseMarathonBonusInput{
		GameID:     gameID,
		QuestionID: currentQ.ID().String(),
		BonusType:  "fifty_fifty",
		PlayerID:   testPlayerID,
	})
	if err == nil {
		t.Error("Expected error when no bonuses left")
	}
}

func TestUseBonus_PlayerMismatch(t *testing.T) {
	f := setupFixture(t)
	startOutput := f.startGameForPlayer(t, testPlayerID)

	game, _ := f.marathonRepo.FindByID(solo_marathon.NewGameIDFromString(startOutput.Game.ID))
	currentQ, _ := game.GetCurrentQuestion()

	uc := f.newUseBonusUC()
	_, err := uc.Execute(UseMarathonBonusInput{
		GameID:     startOutput.Game.ID,
		QuestionID: currentQ.ID().String(),
		BonusType:  "shield",
		PlayerID:   "wrong-player",
	})

	if err == nil {
		t.Error("Expected error for player mismatch")
	}
}

// ========================================
// ContinueMarathon Use Case Tests
// ========================================

func TestContinue_WithCoins(t *testing.T) {
	f := setupFixture(t)
	startOutput := f.startGameForPlayer(t, testPlayerID)
	gameID := startOutput.Game.ID

	// Lose all lives (5 wrong answers with MaxLives=5)
	f.answerCurrentQuestion(t, gameID, testPlayerID, false)
	f.answerCurrentQuestion(t, gameID, testPlayerID, false)
	f.answerCurrentQuestion(t, gameID, testPlayerID, false)
	f.answerCurrentQuestion(t, gameID, testPlayerID, false)
	f.answerCurrentQuestion(t, gameID, testPlayerID, false)

	// Continue
	uc := f.newContinueUC()
	output, err := uc.Execute(ContinueMarathonInput{
		GameID:        gameID,
		PlayerID:      testPlayerID,
		PaymentMethod: "coins",
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if output.Game.Status != "in_progress" {
		t.Errorf("Status = %s, want in_progress after continue", output.Game.Status)
	}
	if output.Game.Lives.CurrentLives != 1 {
		t.Errorf("Lives = %d, want 1 after continue", output.Game.Lives.CurrentLives)
	}
	if output.ContinueCount != 1 {
		t.Errorf("ContinueCount = %d, want 1", output.ContinueCount)
	}
	if output.CoinsDeducted != 200 {
		t.Errorf("CoinsDeducted = %d, want 200", output.CoinsDeducted)
	}
	if output.Game.CurrentQuestion == nil {
		t.Error("Expected current question after continue")
	}
}

func TestContinue_WithAd(t *testing.T) {
	f := setupFixture(t)
	startOutput := f.startGameForPlayer(t, testPlayerID)
	gameID := startOutput.Game.ID

	// Lose all lives (5 wrong answers with MaxLives=5)
	f.answerCurrentQuestion(t, gameID, testPlayerID, false)
	f.answerCurrentQuestion(t, gameID, testPlayerID, false)
	f.answerCurrentQuestion(t, gameID, testPlayerID, false)
	f.answerCurrentQuestion(t, gameID, testPlayerID, false)
	f.answerCurrentQuestion(t, gameID, testPlayerID, false)

	uc := f.newContinueUC()
	output, err := uc.Execute(ContinueMarathonInput{
		GameID:        gameID,
		PlayerID:      testPlayerID,
		PaymentMethod: "ad",
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if output.CoinsDeducted != 0 {
		t.Errorf("CoinsDeducted = %d, want 0 for ad", output.CoinsDeducted)
	}
}

func TestContinue_InvalidPaymentMethod(t *testing.T) {
	f := setupFixture(t)
	startOutput := f.startGameForPlayer(t, testPlayerID)
	gameID := startOutput.Game.ID

	// Lose all lives (5 wrong answers with MaxLives=5)
	f.answerCurrentQuestion(t, gameID, testPlayerID, false)
	f.answerCurrentQuestion(t, gameID, testPlayerID, false)
	f.answerCurrentQuestion(t, gameID, testPlayerID, false)
	f.answerCurrentQuestion(t, gameID, testPlayerID, false)
	f.answerCurrentQuestion(t, gameID, testPlayerID, false)

	uc := f.newContinueUC()
	_, err := uc.Execute(ContinueMarathonInput{
		GameID:        gameID,
		PlayerID:      testPlayerID,
		PaymentMethod: "bitcoin",
	})

	if err == nil {
		t.Error("Expected error for invalid payment method")
	}
}

func TestContinue_GameNotInGameOver(t *testing.T) {
	f := setupFixture(t)
	startOutput := f.startGameForPlayer(t, testPlayerID)

	// Game is in_progress, not game_over
	uc := f.newContinueUC()
	_, err := uc.Execute(ContinueMarathonInput{
		GameID:        startOutput.Game.ID,
		PlayerID:      testPlayerID,
		PaymentMethod: "coins",
	})

	if err == nil {
		t.Error("Expected error when continuing a game not in game_over state")
	}
}

func TestContinue_CostIncreases(t *testing.T) {
	f := setupFixture(t)
	startOutput := f.startGameForPlayer(t, testPlayerID)
	gameID := startOutput.Game.ID

	// First game over + continue (5 wrong answers with MaxLives=5)
	f.answerCurrentQuestion(t, gameID, testPlayerID, false)
	f.answerCurrentQuestion(t, gameID, testPlayerID, false)
	f.answerCurrentQuestion(t, gameID, testPlayerID, false)
	f.answerCurrentQuestion(t, gameID, testPlayerID, false)
	f.answerCurrentQuestion(t, gameID, testPlayerID, false)

	uc := f.newContinueUC()
	output1, _ := uc.Execute(ContinueMarathonInput{
		GameID: gameID, PlayerID: testPlayerID, PaymentMethod: "coins",
	})

	// Second game over + continue
	f.answerCurrentQuestion(t, gameID, testPlayerID, false) // 1→0 (only 1 life after continue)

	output2, _ := uc.Execute(ContinueMarathonInput{
		GameID: gameID, PlayerID: testPlayerID, PaymentMethod: "coins",
	})

	if output2.CoinsDeducted <= output1.CoinsDeducted {
		t.Errorf("Second continue cost (%d) should be higher than first (%d)",
			output2.CoinsDeducted, output1.CoinsDeducted)
	}
}

// ========================================
// AbandonMarathon Use Case Tests
// ========================================

func TestAbandon_Success(t *testing.T) {
	f := setupFixture(t)
	startOutput := f.startGameForPlayer(t, testPlayerID)

	// Answer a few correctly first
	f.answerCurrentQuestion(t, startOutput.Game.ID, testPlayerID, true)
	f.answerCurrentQuestion(t, startOutput.Game.ID, testPlayerID, true)

	uc := f.newAbandonUC()
	output, err := uc.Execute(AbandonMarathonInput{
		GameID:   startOutput.Game.ID,
		PlayerID: testPlayerID,
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if output.GameOverResult.FinalScore != 2 {
		t.Errorf("FinalScore = %d, want 2", output.GameOverResult.FinalScore)
	}
	if output.GameOverResult.TotalQuestions != 2 {
		t.Errorf("TotalQuestions = %d, want 2", output.GameOverResult.TotalQuestions)
	}
}

func TestAbandon_SavesPersonalBest(t *testing.T) {
	f := setupFixture(t)
	startOutput := f.startGameForPlayer(t, testPlayerID)

	// Answer 5 correctly
	for i := 0; i < 5; i++ {
		f.answerCurrentQuestion(t, startOutput.Game.ID, testPlayerID, true)
	}

	uc := f.newAbandonUC()
	output, _ := uc.Execute(AbandonMarathonInput{
		GameID:   startOutput.Game.ID,
		PlayerID: testPlayerID,
	})

	if !output.GameOverResult.IsNewPersonalBest {
		t.Error("Expected IsNewPersonalBest=true for first game")
	}

	// Verify personal best was saved
	category := solo_marathon.NewMarathonCategoryAll()
	pb, err := f.personalBestRepo.FindByPlayerAndCategory(mustUserID(testPlayerID), category)
	if err != nil {
		t.Fatalf("PersonalBest not saved: %v", err)
	}
	if pb.BestScore() != 5 {
		t.Errorf("PersonalBest score = %d, want 5", pb.BestScore())
	}
}

func TestAbandon_PlayerMismatch(t *testing.T) {
	f := setupFixture(t)
	startOutput := f.startGameForPlayer(t, testPlayerID)

	uc := f.newAbandonUC()
	_, err := uc.Execute(AbandonMarathonInput{
		GameID:   startOutput.Game.ID,
		PlayerID: "wrong-player",
	})

	if err == nil {
		t.Error("Expected error for player mismatch")
	}
}

func TestAbandon_FromGameOverState(t *testing.T) {
	f := setupFixture(t)
	startOutput := f.startGameForPlayer(t, testPlayerID)
	gameID := startOutput.Game.ID

	// Get to game_over state (5 wrong answers with MaxLives=5)
	f.answerCurrentQuestion(t, gameID, testPlayerID, false)
	f.answerCurrentQuestion(t, gameID, testPlayerID, false)
	f.answerCurrentQuestion(t, gameID, testPlayerID, false)
	f.answerCurrentQuestion(t, gameID, testPlayerID, false)
	f.answerCurrentQuestion(t, gameID, testPlayerID, false)

	// Abandon from game_over (decline continue)
	uc := f.newAbandonUC()
	_, err := uc.Execute(AbandonMarathonInput{
		GameID:   gameID,
		PlayerID: testPlayerID,
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

// ========================================
// GetMarathonStatus Use Case Tests
// ========================================

func TestGetStatus_NoActiveGame(t *testing.T) {
	f := setupFixture(t)
	uc := f.newGetStatusUC()

	output, err := uc.Execute(GetMarathonStatusInput{PlayerID: testPlayerID})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if output.HasActiveGame {
		t.Error("Expected HasActiveGame=false")
	}
	if output.Game != nil {
		t.Error("Expected Game to be nil")
	}
}

func TestGetStatus_NoActiveGame_ShowsWalletBonuses(t *testing.T) {
	f := setupFixture(t)

	// Pre-fill wallet
	wallet := solo_marathon.ReconstructBonusWallet(mustUserID(testPlayerID), 1, 0, 2, 0)
	f.bonusWalletRepo.Save(wallet)

	uc := f.newGetStatusUC()
	output, err := uc.Execute(GetMarathonStatusInput{PlayerID: testPlayerID})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if output.HasActiveGame {
		t.Error("Expected HasActiveGame=false")
	}
	if output.BonusInventory == nil {
		t.Fatal("Expected BonusInventory even without active game")
	}
	// defaults(2,1,0,3) + wallet(1,0,2,0) = (3,1,2,3)
	if output.BonusInventory.Shield != 3 {
		t.Errorf("Shield = %d, want 3 (2+1)", output.BonusInventory.Shield)
	}
	if output.BonusInventory.Skip != 2 {
		t.Errorf("Skip = %d, want 2 (0+2)", output.BonusInventory.Skip)
	}
}

func TestGetStatus_WithActiveGame(t *testing.T) {
	f := setupFixture(t)
	startOutput := f.startGameForPlayer(t, testPlayerID)

	uc := f.newGetStatusUC()
	output, err := uc.Execute(GetMarathonStatusInput{PlayerID: testPlayerID})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !output.HasActiveGame {
		t.Error("Expected HasActiveGame=true")
	}
	if output.Game == nil {
		t.Fatal("Expected Game to be set")
	}
	if output.Game.ID != startOutput.Game.ID {
		t.Errorf("Game ID = %s, want %s", output.Game.ID, startOutput.Game.ID)
	}
	if output.Game.CurrentQuestion == nil {
		t.Error("Expected CurrentQuestion for in-progress game")
	}
}

func TestGetStatus_InvalidPlayerID(t *testing.T) {
	f := setupFixture(t)
	uc := f.newGetStatusUC()

	_, err := uc.Execute(GetMarathonStatusInput{PlayerID: ""})

	if err == nil {
		t.Error("Expected error for empty player ID")
	}
}

// ========================================
// GetPersonalBests Use Case Tests
// ========================================

func TestGetPersonalBests_NoRecords(t *testing.T) {
	f := setupFixture(t)
	uc := f.newGetPersonalBestsUC()

	output, err := uc.Execute(GetPersonalBestsInput{PlayerID: testPlayerID})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(output.PersonalBests) != 0 {
		t.Errorf("PersonalBests count = %d, want 0", len(output.PersonalBests))
	}
	if output.OverallBest != nil {
		t.Error("Expected OverallBest to be nil")
	}
}

func TestGetPersonalBests_WithRecords(t *testing.T) {
	f := setupFixture(t)

	// Create personal bests for two categories
	catAll := solo_marathon.NewMarathonCategoryAll()
	pb1, _ := solo_marathon.NewPersonalBest(mustUserID(testPlayerID), catAll, 20, 20, 1000000)
	f.personalBestRepo.Save(pb1)

	uc := f.newGetPersonalBestsUC()
	output, err := uc.Execute(GetPersonalBestsInput{PlayerID: testPlayerID})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(output.PersonalBests) != 1 {
		t.Errorf("PersonalBests count = %d, want 1", len(output.PersonalBests))
	}
	if output.OverallBest == nil {
		t.Fatal("Expected OverallBest")
	}
	if output.OverallBest.BestScore != 20 {
		t.Errorf("OverallBest.BestScore = %d, want 20", output.OverallBest.BestScore)
	}
}

func TestGetPersonalBests_InvalidPlayerID(t *testing.T) {
	f := setupFixture(t)
	uc := f.newGetPersonalBestsUC()

	_, err := uc.Execute(GetPersonalBestsInput{PlayerID: ""})

	if err == nil {
		t.Error("Expected error for empty player ID")
	}
}

// ========================================
// GetMarathonLeaderboard Use Case Tests
// ========================================

func TestGetLeaderboard_Empty(t *testing.T) {
	f := setupFixture(t)
	uc := f.newGetLeaderboardUC()

	output, err := uc.Execute(GetMarathonLeaderboardInput{
		CategoryID: "all",
		Limit:      10,
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(output.Entries) != 0 {
		t.Errorf("Entries count = %d, want 0", len(output.Entries))
	}
	if output.Category.IsAllCategories != true {
		t.Error("Expected all categories mode")
	}
}

func TestGetLeaderboard_WithEntries(t *testing.T) {
	f := setupFixture(t)

	// Create personal bests for two players
	catAll := solo_marathon.NewMarathonCategoryAll()
	pb1, _ := solo_marathon.NewPersonalBest(mustUserID(testPlayerID), catAll, 30, 30, 1000000)
	pb2, _ := solo_marathon.NewPersonalBest(mustUserID(testPlayerID2), catAll, 20, 20, 1000000)
	f.personalBestRepo.Save(pb1)
	f.personalBestRepo.Save(pb2)

	uc := f.newGetLeaderboardUC()
	output, err := uc.Execute(GetMarathonLeaderboardInput{
		CategoryID: "all",
		Limit:      10,
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(output.Entries) != 2 {
		t.Errorf("Entries count = %d, want 2", len(output.Entries))
	}
	// First entry should have highest score
	if output.Entries[0].BestScore != 30 {
		t.Errorf("First entry score = %d, want 30", output.Entries[0].BestScore)
	}
	if output.Entries[0].Rank != 1 {
		t.Errorf("First entry rank = %d, want 1", output.Entries[0].Rank)
	}
	if output.Entries[0].Username == "" {
		t.Error("Expected username in leaderboard entry")
	}
}

func TestGetLeaderboard_LimitClamping(t *testing.T) {
	f := setupFixture(t)
	uc := f.newGetLeaderboardUC()

	// Zero limit should default to 10
	output, err := uc.Execute(GetMarathonLeaderboardInput{Limit: 0})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	_ = output

	// Negative limit should default to 10
	output, err = uc.Execute(GetMarathonLeaderboardInput{Limit: -5})
	if err != nil {
		t.Fatalf("Expected no error for negative limit, got %v", err)
	}
}

// ========================================
// Integration: Full Game Flow
// ========================================

func TestFullGameFlow_Start_Answer_GameOver_Continue_Abandon(t *testing.T) {
	f := setupFixture(t)

	// 1. Start marathon
	startOutput := f.startGameForPlayer(t, testPlayerID)
	gameID := startOutput.Game.ID
	if startOutput.Game.Status != "in_progress" {
		t.Fatalf("Expected in_progress, got %s", startOutput.Game.Status)
	}

	// 2. Answer 3 correctly
	for i := 0; i < 3; i++ {
		output := f.answerCurrentQuestion(t, gameID, testPlayerID, true)
		if output.Score != i+1 {
			t.Errorf("Score after %d answers = %d, want %d", i+1, output.Score, i+1)
		}
	}

	// 3. Answer wrong 5 times → game over (MaxLives=5)
	f.answerCurrentQuestion(t, gameID, testPlayerID, false)
	f.answerCurrentQuestion(t, gameID, testPlayerID, false)
	f.answerCurrentQuestion(t, gameID, testPlayerID, false)
	f.answerCurrentQuestion(t, gameID, testPlayerID, false)
	gameOverOutput := f.answerCurrentQuestion(t, gameID, testPlayerID, false)
	if !gameOverOutput.IsGameOver {
		t.Fatal("Expected game over")
	}

	// 4. Continue
	continueUC := f.newContinueUC()
	continueOutput, err := continueUC.Execute(ContinueMarathonInput{
		GameID: gameID, PlayerID: testPlayerID, PaymentMethod: "ad",
	})
	if err != nil {
		t.Fatalf("Continue failed: %v", err)
	}
	if continueOutput.Game.Status != "in_progress" {
		t.Errorf("Status after continue = %s, want in_progress", continueOutput.Game.Status)
	}

	// 5. Answer 2 more correctly
	f.answerCurrentQuestion(t, gameID, testPlayerID, true)
	f.answerCurrentQuestion(t, gameID, testPlayerID, true)

	// 6. Abandon
	abandonUC := f.newAbandonUC()
	abandonOutput, err := abandonUC.Execute(AbandonMarathonInput{
		GameID: gameID, PlayerID: testPlayerID,
	})
	if err != nil {
		t.Fatalf("Abandon failed: %v", err)
	}
	if abandonOutput.GameOverResult.FinalScore != 5 {
		t.Errorf("FinalScore = %d, want 5 (3+2 correct)", abandonOutput.GameOverResult.FinalScore)
	}
	if !abandonOutput.GameOverResult.IsNewPersonalBest {
		t.Error("Expected new personal best for first game")
	}

	// 7. Verify game is no longer active
	statusUC := f.newGetStatusUC()
	statusOutput, _ := statusUC.Execute(GetMarathonStatusInput{PlayerID: testPlayerID})
	if statusOutput.HasActiveGame {
		t.Error("Expected no active game after abandon")
	}

	// 8. Can start new game
	_, err = f.newStartUC().Execute(StartMarathonInput{PlayerID: testPlayerID})
	if err != nil {
		t.Errorf("Expected to start new game after abandon, got %v", err)
	}
}
