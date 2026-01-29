package solo_marathon

import (
	"testing"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
)

// Helper function to create a test quiz
func createTestQuiz(t *testing.T) *quiz.Quiz {
	t.Helper()

	// Create quiz aggregate
	quizID := quiz.NewQuizID()
	quizTitle, _ := quiz.NewQuizTitle("Test Quiz")
	categoryID := quiz.NewCategoryID()
	timeLimit, _ := quiz.NewTimeLimit(600) // 10 minutes
	passingScore, _ := quiz.NewPassingScore(70)

	q, err := quiz.NewQuiz(
		quizID,
		quizTitle,
		"Test Description",
		categoryID,
		timeLimit,
		passingScore,
		int64(1000000),
	)
	if err != nil {
		t.Fatalf("Failed to create test quiz: %v", err)
	}

	// Add questions
	q1 := createTestQuestion(t, "Question 1", "A", true, 1)
	q2 := createTestQuestion(t, "Question 2", "B", true, 2)
	q3 := createTestQuestion(t, "Question 3", "C", true, 3)

	if err := q.AddQuestion(q1); err != nil {
		t.Fatalf("Failed to add question 1: %v", err)
	}
	if err := q.AddQuestion(q2); err != nil {
		t.Fatalf("Failed to add question 2: %v", err)
	}
	if err := q.AddQuestion(q3); err != nil {
		t.Fatalf("Failed to add question 3: %v", err)
	}

	return q
}

func createTestQuestion(t *testing.T, text string, correctAnswerText string, isFirst bool, position int) quiz.Question {
	t.Helper()

	questionText, _ := quiz.NewQuestionText(text)
	points, _ := quiz.NewPoints(100)

	q, _ := quiz.NewQuestion(
		quiz.NewQuestionID(),
		questionText,
		points,
		position,
	)

	// Add answers
	correctAnswer := createTestAnswer(t, correctAnswerText, true, 1)
	wrongAnswer1 := createTestAnswer(t, "Wrong 1", false, 2)
	wrongAnswer2 := createTestAnswer(t, "Wrong 2", false, 3)
	wrongAnswer3 := createTestAnswer(t, "Wrong 3", false, 4)

	q.AddAnswer(correctAnswer)
	q.AddAnswer(wrongAnswer1)
	q.AddAnswer(wrongAnswer2)
	q.AddAnswer(wrongAnswer3)

	return *q
}

func createTestAnswer(t *testing.T, text string, isCorrect bool, position int) quiz.Answer {
	t.Helper()

	answerText, _ := quiz.NewAnswerText(text)
	a, _ := quiz.NewAnswer(
		quiz.NewAnswerID(),
		answerText,
		isCorrect,
		position,
	)

	return *a
}

// TestNewMarathonGame_Success tests successful game creation
func TestNewMarathonGame_Success(t *testing.T) {
	playerID, _ := shared.NewUserID("user123")
	category := NewMarathonCategoryAll()
	quizAggregate := createTestQuiz(t)
	now := int64(1000000)

	game, err := NewMarathonGame(playerID, category, quizAggregate, nil, now)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if game == nil {
		t.Fatal("Expected game to be created")
	}
	if game.ID().IsZero() {
		t.Error("Game ID should not be zero")
	}
	if game.PlayerID() != playerID {
		t.Errorf("PlayerID = %v, want %v", game.PlayerID(), playerID)
	}
	if game.Status() != GameStatusInProgress {
		t.Errorf("Status = %v, want %v", game.Status(), GameStatusInProgress)
	}
	if game.CurrentStreak() != 0 {
		t.Errorf("CurrentStreak should be 0, got %d", game.CurrentStreak())
	}
	if game.MaxStreak() != 0 {
		t.Errorf("MaxStreak should be 0, got %d", game.MaxStreak())
	}

	// Check events
	events := game.Events()
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}
}

// TestNewMarathonGame_WithPersonalBest tests game creation with existing record
func TestNewMarathonGame_WithPersonalBest(t *testing.T) {
	playerID, _ := shared.NewUserID("user123")
	category := NewMarathonCategoryAll()
	quizAggregate := createTestQuiz(t)
	now := int64(1000000)

	personalBest, _ := NewPersonalBest(playerID, category, 47, 10000, now)

	game, err := NewMarathonGame(playerID, category, quizAggregate, personalBest, now)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !game.IsNewPersonalBest() {
		// 0 streak is not a new record
		t.Log("Correctly identified as not a new record yet")
	}
}

// TestNewMarathonGame_InvalidInputs tests validation
func TestNewMarathonGame_InvalidInputs(t *testing.T) {
	playerID, _ := shared.NewUserID("user123")
	category := NewMarathonCategoryAll()
	quizAggregate := createTestQuiz(t)
	now := int64(1000000)

	tests := []struct {
		name      string
		playerID  shared.UserID
		quiz      *quiz.Quiz
		expectErr bool
	}{
		{
			name:      "Invalid player ID",
			playerID:  shared.UserID{},
			quiz:      quizAggregate,
			expectErr: true,
		},
		{
			name:      "Nil quiz",
			playerID:  playerID,
			quiz:      nil,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewMarathonGame(tt.playerID, category, tt.quiz, nil, now)

			if tt.expectErr && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

// TestMarathonGame_AnswerQuestion_Correct tests answering correctly
func TestMarathonGame_AnswerQuestion_Correct(t *testing.T) {
	playerID, _ := shared.NewUserID("user123")
	category := NewMarathonCategoryAll()
	quizAggregate := createTestQuiz(t)
	now := int64(1000000)

	game, _ := NewMarathonGame(playerID, category, quizAggregate, nil, now)
	game.Events() // Clear initial events

	// Get first question
	currentQuestion, _ := game.GetCurrentQuestion()
	correctAnswer := currentQuestion.Answers()[0] // First answer is correct

	// Answer correctly
	result, err := game.AnswerQuestion(
		currentQuestion.ID(),
		correctAnswer.ID(),
		2000, // 2 seconds
		now+2000,
	)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !result.IsCorrect {
		t.Error("Expected correct answer")
	}
	if result.CurrentStreak != 1 {
		t.Errorf("Expected streak 1, got %d", result.CurrentStreak)
	}
	if result.MaxStreak != 1 {
		t.Errorf("Expected max streak 1, got %d", result.MaxStreak)
	}
	if result.LifeLost {
		t.Error("Should not lose life on correct answer")
	}
	if result.RemainingLives != 3 {
		t.Errorf("Expected 3 lives, got %d", result.RemainingLives)
	}
	if result.IsGameOver {
		t.Error("Game should not be over")
	}

	// Check game state
	if game.CurrentStreak() != 1 {
		t.Errorf("Game streak = %d, want 1", game.CurrentStreak())
	}
	if game.MaxStreak() != 1 {
		t.Errorf("Game max streak = %d, want 1", game.MaxStreak())
	}

	// Check difficulty updated
	if game.Difficulty().Level() != DifficultyBeginner {
		t.Errorf("Difficulty should still be beginner at streak 1")
	}

	// Check events
	events := game.Events()
	if len(events) != 1 {
		t.Errorf("Expected 1 event (QuestionAnswered), got %d", len(events))
	}
}

// TestMarathonGame_AnswerQuestion_Incorrect tests answering incorrectly
func TestMarathonGame_AnswerQuestion_Incorrect(t *testing.T) {
	playerID, _ := shared.NewUserID("user123")
	category := NewMarathonCategoryAll()
	quizAggregate := createTestQuiz(t)
	now := int64(1000000)

	game, _ := NewMarathonGame(playerID, category, quizAggregate, nil, now)
	game.Events() // Clear initial events

	// Get first question
	currentQuestion, _ := game.GetCurrentQuestion()
	wrongAnswer := currentQuestion.Answers()[1] // Second answer is wrong

	// Answer incorrectly
	result, err := game.AnswerQuestion(
		currentQuestion.ID(),
		wrongAnswer.ID(),
		2000,
		now+2000,
	)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result.IsCorrect {
		t.Error("Expected incorrect answer")
	}
	if result.CurrentStreak != 0 {
		t.Errorf("Expected streak 0 after wrong answer, got %d", result.CurrentStreak)
	}
	if !result.LifeLost {
		t.Error("Should lose life on incorrect answer")
	}
	if result.RemainingLives != 2 {
		t.Errorf("Expected 2 lives remaining, got %d", result.RemainingLives)
	}
	if result.IsGameOver {
		t.Error("Game should not be over with 2 lives")
	}

	// Check events (QuestionAnswered + LifeLost)
	events := game.Events()
	if len(events) != 2 {
		t.Errorf("Expected 2 events, got %d", len(events))
	}
}

// TestMarathonGame_AnswerQuestion_GameOver tests game over when no lives
func TestMarathonGame_AnswerQuestion_GameOver(t *testing.T) {
	playerID, _ := shared.NewUserID("user123")
	category := NewMarathonCategoryAll()
	quizAggregate := createTestQuiz(t)
	now := int64(1000000)

	game, _ := NewMarathonGame(playerID, category, quizAggregate, nil, now)

	// Manually set lives to 1
	game.lives = ReconstructLivesSystem(1, now)
	game.Events() // Clear events

	// Get first question
	currentQuestion, _ := game.GetCurrentQuestion()
	wrongAnswer := currentQuestion.Answers()[1]

	// Answer incorrectly (lose last life)
	result, err := game.AnswerQuestion(
		currentQuestion.ID(),
		wrongAnswer.ID(),
		2000,
		now+2000,
	)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result.RemainingLives != 0 {
		t.Errorf("Expected 0 lives, got %d", result.RemainingLives)
	}
	if !result.IsGameOver {
		t.Error("Game should be over with 0 lives")
	}
	if game.Status() != GameStatusFinished {
		t.Errorf("Game status should be finished, got %v", game.Status())
	}
	if !game.IsGameOver() {
		t.Error("IsGameOver() should return true")
	}

	// Check events (QuestionAnswered + LifeLost + GameOver)
	events := game.Events()
	if len(events) != 3 {
		t.Errorf("Expected 3 events, got %d", len(events))
	}
}

// TestMarathonGame_StreakIncrement tests streak building
func TestMarathonGame_StreakIncrement(t *testing.T) {
	playerID, _ := shared.NewUserID("user123")
	category := NewMarathonCategoryAll()
	quizAggregate := createTestQuiz(t)
	now := int64(1000000)

	game, _ := NewMarathonGame(playerID, category, quizAggregate, nil, now)

	// Answer 3 questions correctly
	for i := 0; i < 3; i++ {
		currentQuestion, _ := game.GetCurrentQuestion()
		correctAnswer := currentQuestion.Answers()[0]

		result, err := game.AnswerQuestion(
			currentQuestion.ID(),
			correctAnswer.ID(),
			2000,
			now+int64(i*2000),
		)

		if err != nil {
			t.Fatalf("Question %d: unexpected error %v", i+1, err)
		}
		if result.CurrentStreak != i+1 {
			t.Errorf("Question %d: expected streak %d, got %d", i+1, i+1, result.CurrentStreak)
		}
		if result.MaxStreak != i+1 {
			t.Errorf("Question %d: expected max streak %d, got %d", i+1, i+1, result.MaxStreak)
		}
	}

	if game.CurrentStreak() != 3 {
		t.Errorf("Final streak = %d, want 3", game.CurrentStreak())
	}
	if game.MaxStreak() != 3 {
		t.Errorf("Final max streak = %d, want 3", game.MaxStreak())
	}
}

// TestMarathonGame_DifficultyProgression tests difficulty changes
func TestMarathonGame_DifficultyProgression(t *testing.T) {
	playerID, _ := shared.NewUserID("user123")
	category := NewMarathonCategoryAll()
	quizAggregate := createTestQuiz(t)
	now := int64(1000000)

	game, _ := NewMarathonGame(playerID, category, quizAggregate, nil, now)

	// Start at Beginner
	if game.Difficulty().Level() != DifficultyBeginner {
		t.Errorf("Should start at Beginner, got %v", game.Difficulty().Level())
	}

	// Manually set streak to 6 (triggers Medium)
	game.currentStreak = 6
	game.difficulty = game.difficulty.UpdateFromStreak(6)

	if game.Difficulty().Level() != DifficultyMedium {
		t.Errorf("At streak 6, should be Medium, got %v", game.Difficulty().Level())
	}
}

// TestMarathonGame_UseHint tests using hints
func TestMarathonGame_UseHint(t *testing.T) {
	playerID, _ := shared.NewUserID("user123")
	category := NewMarathonCategoryAll()
	quizAggregate := createTestQuiz(t)
	now := int64(1000000)

	game, _ := NewMarathonGame(playerID, category, quizAggregate, nil, now)
	game.Events() // Clear events

	currentQuestion, _ := game.GetCurrentQuestion()

	// Use 50/50 hint
	err := game.UseHint(currentQuestion.ID(), HintFiftyFifty, now+1000)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check hint was decremented
	if game.Hints().FiftyFifty() != 2 {
		t.Errorf("Expected 2 50/50 hints remaining, got %d", game.Hints().FiftyFifty())
	}

	// Check event
	events := game.Events()
	if len(events) != 1 {
		t.Errorf("Expected 1 HintUsed event, got %d", len(events))
	}

	// Try to use same hint twice on same question
	err = game.UseHint(currentQuestion.ID(), HintFiftyFifty, now+2000)
	if err != ErrHintAlreadyUsed {
		t.Errorf("Expected ErrHintAlreadyUsed, got %v", err)
	}
}

// TestMarathonGame_UseHint_NoHintsAvailable tests using hint when none available
func TestMarathonGame_UseHint_NoHintsAvailable(t *testing.T) {
	playerID, _ := shared.NewUserID("user123")
	category := NewMarathonCategoryAll()
	quizAggregate := createTestQuiz(t)
	now := int64(1000000)

	game, _ := NewMarathonGame(playerID, category, quizAggregate, nil, now)

	// Manually set hints to 0
	game.hints = ReconstructHintsSystem(0, 0, 0)

	currentQuestion, _ := game.GetCurrentQuestion()

	err := game.UseHint(currentQuestion.ID(), HintFiftyFifty, now+1000)

	if err != ErrNoHintsAvailable {
		t.Errorf("Expected ErrNoHintsAvailable, got %v", err)
	}
}

// TestMarathonGame_Abandon tests abandoning game
func TestMarathonGame_Abandon(t *testing.T) {
	playerID, _ := shared.NewUserID("user123")
	category := NewMarathonCategoryAll()
	quizAggregate := createTestQuiz(t)
	now := int64(1000000)

	game, _ := NewMarathonGame(playerID, category, quizAggregate, nil, now)
	game.Events() // Clear events

	// Set some streak
	game.currentStreak = 5
	game.maxStreak = 5

	err := game.Abandon(now + 10000)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if game.Status() != GameStatusAbandoned {
		t.Errorf("Status should be Abandoned, got %v", game.Status())
	}
	if !game.IsGameOver() {
		t.Error("Game should be over after abandon")
	}

	// Check event
	events := game.Events()
	if len(events) != 1 {
		t.Errorf("Expected 1 GameOver event, got %d", len(events))
	}
}

// TestMarathonGame_Abandon_AlreadyFinished tests abandoning finished game
func TestMarathonGame_Abandon_AlreadyFinished(t *testing.T) {
	playerID, _ := shared.NewUserID("user123")
	category := NewMarathonCategoryAll()
	quizAggregate := createTestQuiz(t)
	now := int64(1000000)

	game, _ := NewMarathonGame(playerID, category, quizAggregate, nil, now)

	// Manually finish game
	game.status = GameStatusFinished

	err := game.Abandon(now + 10000)

	if err != ErrGameNotActive {
		t.Errorf("Expected ErrGameNotActive, got %v", err)
	}
}

// TestMarathonGame_IsNewPersonalBest tests record detection
func TestMarathonGame_IsNewPersonalBest(t *testing.T) {
	playerID, _ := shared.NewUserID("user123")
	category := NewMarathonCategoryAll()
	quizAggregate := createTestQuiz(t)
	now := int64(1000000)

	tests := []struct {
		name              string
		personalBest      *PersonalBest
		currentMaxStreak  int
		expectedNewRecord bool
	}{
		{
			name:              "No previous record",
			personalBest:      nil,
			currentMaxStreak:  5,
			expectedNewRecord: true,
		},
		{
			name: "Beat previous record",
			personalBest: func() *PersonalBest {
				pb, _ := NewPersonalBest(playerID, category, 10, 5000, now)
				return pb
			}(),
			currentMaxStreak:  15,
			expectedNewRecord: true,
		},
		{
			name: "Did not beat record",
			personalBest: func() *PersonalBest {
				pb, _ := NewPersonalBest(playerID, category, 20, 10000, now)
				return pb
			}(),
			currentMaxStreak:  15,
			expectedNewRecord: false,
		},
		{
			name: "Tied record",
			personalBest: func() *PersonalBest {
				pb, _ := NewPersonalBest(playerID, category, 15, 7500, now)
				return pb
			}(),
			currentMaxStreak:  15,
			expectedNewRecord: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game, _ := NewMarathonGame(playerID, category, quizAggregate, tt.personalBest, now)
			game.maxStreak = tt.currentMaxStreak

			result := game.IsNewPersonalBest()

			if result != tt.expectedNewRecord {
				t.Errorf("IsNewPersonalBest() = %v, want %v", result, tt.expectedNewRecord)
			}
		})
	}
}
