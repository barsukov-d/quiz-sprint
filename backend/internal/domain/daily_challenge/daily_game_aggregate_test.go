package daily_challenge

import (
	"testing"
	"time"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
)

// Helper to create test quiz
func createTestQuiz(t *testing.T) *quiz.Quiz {
	t.Helper()

	quizID := quiz.NewQuizID()
	title, _ := quiz.NewQuizTitle("Daily Test Quiz")
	categoryID := quiz.NewCategoryID()
	timeLimit, _ := quiz.NewTimeLimit(600)
	passingScore, _ := quiz.NewPassingScore(70)

	q, err := quiz.NewQuiz(
		quizID,
		title,
		"Daily test quiz description",
		categoryID,
		timeLimit,
		passingScore,
		int64(1000000),
	)
	if err != nil {
		t.Fatalf("Failed to create test quiz: %v", err)
	}

	// Add 10 questions (daily challenge standard)
	for i := 1; i <= 10; i++ {
		question := createTestQuestion(t, i)
		if err := q.AddQuestion(question); err != nil {
			t.Fatalf("Failed to add question %d: %v", i, err)
		}
	}

	return q
}

func createTestQuestion(t *testing.T, position int) quiz.Question {
	t.Helper()

	questionText, _ := quiz.NewQuestionText("Test Question " + string(rune('0'+position)))
	points, _ := quiz.NewPoints(100)

	q, _ := quiz.NewQuestion(
		quiz.NewQuestionID(),
		questionText,
		points,
		position,
	)

	// Add 4 answers
	correctAnswer := createTestAnswer(t, "Correct Answer", true, 1)
	wrongAnswer1 := createTestAnswer(t, "Wrong Answer 1", false, 2)
	wrongAnswer2 := createTestAnswer(t, "Wrong Answer 2", false, 3)
	wrongAnswer3 := createTestAnswer(t, "Wrong Answer 3", false, 4)

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

// TestNewDailyGame_Success tests successful daily game creation
func TestNewDailyGame_Success(t *testing.T) {
	playerID, _ := shared.NewUserID("user123")
	dailyQuizID := NewDailyQuizID()
	date := NewDate(2026, time.January, 25)
	quizAggregate := createTestQuiz(t)
	streak := NewStreakSystem()
	now := int64(1000000)

	game, err := NewDailyGame(playerID, dailyQuizID, date, quizAggregate, streak, now)

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
	if game.IsCompleted() {
		t.Error("Game should not be completed initially")
	}

	// Check event emitted
	events := game.Events()
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}
}

// TestNewDailyGame_InvalidInputs tests validation
func TestNewDailyGame_InvalidInputs(t *testing.T) {
	playerID, _ := shared.NewUserID("user123")
	dailyQuizID := NewDailyQuizID()
	date := NewDate(2026, time.January, 25)
	quizAggregate := createTestQuiz(t)
	streak := NewStreakSystem()
	now := int64(1000000)

	tests := []struct {
		name        string
		playerID    UserID
		dailyQuizID DailyQuizID
		date        Date
		quiz        *quiz.Quiz
		expectedErr error
	}{
		{
			name:        "Invalid player ID",
			playerID:    UserID{},
			dailyQuizID: dailyQuizID,
			date:        date,
			quiz:        quizAggregate,
			expectedErr: ErrInvalidGameID,
		},
		{
			name:        "Invalid daily quiz ID",
			playerID:    playerID,
			dailyQuizID: DailyQuizID{},
			date:        date,
			quiz:        quizAggregate,
			expectedErr: ErrInvalidDailyQuizID,
		},
		{
			name:        "Invalid date",
			playerID:    playerID,
			dailyQuizID: dailyQuizID,
			date:        Date{},
			quiz:        quizAggregate,
			expectedErr: ErrInvalidDate,
		},
		{
			name:        "Nil quiz",
			playerID:    playerID,
			dailyQuizID: dailyQuizID,
			date:        date,
			quiz:        nil,
			expectedErr: quiz.ErrQuizNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game, err := NewDailyGame(tt.playerID, tt.dailyQuizID, tt.date, tt.quiz, streak, now)

			if err == nil {
				t.Error("Expected error, got nil")
			}
			if err != tt.expectedErr {
				t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
			}
			if game != nil {
				t.Error("Expected nil game for invalid input")
			}
		})
	}
}

// TestDailyGame_AnswerQuestion tests answering questions
func TestDailyGame_AnswerQuestion(t *testing.T) {
	playerID, _ := shared.NewUserID("user123")
	dailyQuizID := NewDailyQuizID()
	date := NewDate(2026, time.January, 25)
	quizAggregate := createTestQuiz(t)
	streak := NewStreakSystem()
	now := int64(1000000)

	game, _ := NewDailyGame(playerID, dailyQuizID, date, quizAggregate, streak, now)

	// Clear startup event
	game.Events()

	// Get first question
	questions := quizAggregate.Questions()
	firstQuestion := questions[0]
	correctAnswer := firstQuestion.Answers()[0] // First answer is correct (from helper)

	// Answer first question
	result, err := game.AnswerQuestion(
		firstQuestion.ID(),
		correctAnswer.ID(),
		2000, // 2 seconds
		now+2000,
	)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("Expected result")
	}
	if result.QuestionIndex != 0 {
		t.Errorf("QuestionIndex = %d, want 0", result.QuestionIndex)
	}
	if result.TimeTaken != 2000 {
		t.Errorf("TimeTaken = %d, want 2000", result.TimeTaken)
	}
	if result.RemainingQuestions != 9 {
		t.Errorf("RemainingQuestions = %d, want 9", result.RemainingQuestions)
	}
	if result.IsGameCompleted {
		t.Error("Game should not be completed after 1 question")
	}

	// Check event
	events := game.Events()
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}
}

// TestDailyGame_CompleteGame tests completing all 10 questions
func TestDailyGame_CompleteGame(t *testing.T) {
	playerID, _ := shared.NewUserID("user123")
	dailyQuizID := NewDailyQuizID()
	date := NewDate(2026, time.January, 25)
	quizAggregate := createTestQuiz(t)
	streak := ReconstructStreakSystem(0, 0, Date{}) // No previous streak
	now := int64(1000000)

	game, _ := NewDailyGame(playerID, dailyQuizID, date, quizAggregate, streak, now)
	game.Events() // Clear startup event

	// Answer all 10 questions correctly
	questions := quizAggregate.Questions()
	for i, question := range questions {
		correctAnswer := question.Answers()[0]

		result, err := game.AnswerQuestion(
			question.ID(),
			correctAnswer.ID(),
			2000,
			now+int64((i+1)*2000),
		)

		if err != nil {
			t.Fatalf("Question %d: unexpected error: %v", i+1, err)
		}

		if i < 9 {
			// Not last question
			if result.IsGameCompleted {
				t.Errorf("Question %d: game should not be completed yet", i+1)
			}
		} else {
			// Last question
			if !result.IsGameCompleted {
				t.Error("Game should be completed after 10 questions")
			}
		}
	}

	// Verify game is completed
	if !game.IsCompleted() {
		t.Error("Game should be marked as completed")
	}
	if game.Status() != GameStatusCompleted {
		t.Errorf("Status = %v, want %v", game.Status(), GameStatusCompleted)
	}

	// Verify streak updated
	if game.Streak().CurrentStreak() != 1 {
		t.Errorf("Streak = %d, want 1 (first day)", game.Streak().CurrentStreak())
	}

	// Check final score
	finalScore := game.GetFinalScore()
	if finalScore <= 0 {
		t.Errorf("Final score = %d, should be > 0", finalScore)
	}

	// Check correct answers count
	correctCount := game.GetCorrectAnswersCount()
	if correctCount != 10 {
		t.Errorf("Correct answers = %d, want 10", correctCount)
	}

	// Check events (10 answer events + 1 completed event)
	events := game.Events()
	if len(events) < 11 {
		t.Errorf("Expected at least 11 events, got %d", len(events))
	}
}

// TestDailyGame_StreakBonus tests streak bonus application
func TestDailyGame_StreakBonus(t *testing.T) {
	playerID, _ := shared.NewUserID("user123")
	dailyQuizID := NewDailyQuizID()
	date := NewDate(2026, time.January, 25)
	quizAggregate := createTestQuiz(t)
	now := int64(1000000)

	tests := []struct {
		name             string
		initialStreak    int
		expectedMultiple float64
	}{
		{"No streak", 0, 1.0},
		{"3-day streak", 3, 1.1},
		{"7-day streak", 7, 1.25},
		{"30-day streak", 30, 1.5},
		{"100-day streak", 100, 1.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lastDate := date.Previous()
			streak := ReconstructStreakSystem(tt.initialStreak, tt.initialStreak, lastDate)

			game, _ := NewDailyGame(playerID, dailyQuizID, date, quizAggregate, streak, now)
			game.Events() // Clear startup event

			// Answer all questions correctly
			questions := quizAggregate.Questions()
			for i, question := range questions {
				correctAnswer := question.Answers()[0]
				game.AnswerQuestion(
					question.ID(),
					correctAnswer.ID(),
					2000,
					now+int64((i+1)*2000),
				)
			}

			// Check that streak was incremented
			expectedStreak := tt.initialStreak + 1
			if game.Streak().CurrentStreak() != expectedStreak {
				t.Errorf("Streak = %d, want %d", game.Streak().CurrentStreak(), expectedStreak)
			}

			// Check score has correct multiplier applied
			baseScore := game.Session().BaseScore().Value()
			finalScore := game.GetFinalScore()

			// Expected final score should be approximately baseScore * multiplier
			expectedFinal := int(float64(baseScore) * tt.expectedMultiple)

			// Allow for small rounding differences
			if finalScore < expectedFinal-1 || finalScore > expectedFinal+1 {
				t.Errorf("Final score = %d, expected ~%d (base %d * %.2f)",
					finalScore, expectedFinal, baseScore, tt.expectedMultiple)
			}
		})
	}
}

// TestDailyGame_AnswerAfterCompletion tests error when answering completed game
func TestDailyGame_AnswerAfterCompletion(t *testing.T) {
	playerID, _ := shared.NewUserID("user123")
	dailyQuizID := NewDailyQuizID()
	date := NewDate(2026, time.January, 25)
	quizAggregate := createTestQuiz(t)
	streak := NewStreakSystem()
	now := int64(1000000)

	game, _ := NewDailyGame(playerID, dailyQuizID, date, quizAggregate, streak, now)

	// Complete all questions
	questions := quizAggregate.Questions()
	for i, question := range questions {
		correctAnswer := question.Answers()[0]
		game.AnswerQuestion(
			question.ID(),
			correctAnswer.ID(),
			2000,
			now+int64((i+1)*2000),
		)
	}

	// Try to answer again
	firstQuestion := questions[0]
	correctAnswer := firstQuestion.Answers()[0]

	result, err := game.AnswerQuestion(
		firstQuestion.ID(),
		correctAnswer.ID(),
		2000,
		now+100000,
	)

	if err == nil {
		t.Error("Expected error when answering completed game")
	}
	if err != ErrGameNotActive {
		t.Errorf("Expected ErrGameNotActive, got %v", err)
	}
	if result != nil {
		t.Error("Expected nil result")
	}
}

// TestDailyGame_SetRank tests setting leaderboard rank
func TestDailyGame_SetRank(t *testing.T) {
	playerID, _ := shared.NewUserID("user123")
	dailyQuizID := NewDailyQuizID()
	date := NewDate(2026, time.January, 25)
	quizAggregate := createTestQuiz(t)
	streak := NewStreakSystem()
	now := int64(1000000)

	game, _ := NewDailyGame(playerID, dailyQuizID, date, quizAggregate, streak, now)

	// Initially no rank
	if game.Rank() != nil {
		t.Error("Initial rank should be nil")
	}

	// Set rank
	game.SetRank(42)

	if game.Rank() == nil {
		t.Fatal("Rank should not be nil after setting")
	}
	if *game.Rank() != 42 {
		t.Errorf("Rank = %d, want 42", *game.Rank())
	}
}

// TestDailyGame_ReconstructDailyGame tests reconstruction from persistence
func TestDailyGame_ReconstructDailyGame(t *testing.T) {
	playerID, _ := shared.NewUserID("user123")
	gameID := NewGameID()
	dailyQuizID := NewDailyQuizID()
	date := NewDate(2026, time.January, 25)
	streak := ReconstructStreakSystem(5, 10, date)
	rank := 3

	// Create minimal session for reconstruction (would come from DB)
	// In real scenario, this would be fully reconstructed with answers
	game := ReconstructDailyGame(
		gameID,
		playerID,
		dailyQuizID,
		date,
		GameStatusCompleted,
		nil, // Session would be reconstructed separately
		streak,
		&rank,
		nil, // ChestReward would be reconstructed if present
	)

	if game == nil {
		t.Fatal("Expected game to be reconstructed")
	}
	if game.ID() != gameID {
		t.Errorf("ID = %v, want %v", game.ID(), gameID)
	}
	if game.PlayerID() != playerID {
		t.Errorf("PlayerID = %v, want %v", game.PlayerID(), playerID)
	}
	if game.Status() != GameStatusCompleted {
		t.Errorf("Status = %v, want %v", game.Status(), GameStatusCompleted)
	}
	if *game.Rank() != rank {
		t.Errorf("Rank = %d, want %d", *game.Rank(), rank)
	}

	// Events should be empty after reconstruction
	events := game.Events()
	if len(events) != 0 {
		t.Errorf("Reconstructed game should have no events, got %d", len(events))
	}
}
