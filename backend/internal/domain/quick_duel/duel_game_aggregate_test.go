package quick_duel

import (
	"testing"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
)

// TestNewDuelGame_Success tests successful duel creation
func TestNewDuelGame_Success(t *testing.T) {
	player1ID, _ := shared.NewUserID("player1")
	player2ID, _ := shared.NewUserID("player2")

	player1 := NewDuelPlayer(player1ID, "Player1", NewEloRating())
	player2 := NewDuelPlayer(player2ID, "Player2", NewEloRating())

	questionIDs := make([]QuestionID, QuestionsPerDuel)
	for i := 0; i < QuestionsPerDuel; i++ {
		questionIDs[i] = quiz.NewQuestionID()
	}

	now := int64(1000000)

	game, err := NewDuelGame(player1, player2, questionIDs, now)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if game == nil {
		t.Fatal("Expected game to be created")
	}
	if game.ID().IsZero() {
		t.Error("Game ID should not be zero")
	}
	if game.Status() != GameStatusWaitingStart {
		t.Errorf("Status = %v, want %v", game.Status(), GameStatusWaitingStart)
	}
	if game.CurrentRound() != 0 {
		t.Errorf("CurrentRound = %d, want 0", game.CurrentRound())
	}

	// Check players
	if game.Player1().UserID() != player1ID {
		t.Errorf("Player1 ID = %v, want %v", game.Player1().UserID(), player1ID)
	}
	if game.Player2().UserID() != player2ID {
		t.Errorf("Player2 ID = %v, want %v", game.Player2().UserID(), player2ID)
	}

	// Check event emitted
	events := game.Events()
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}
}

// TestNewDuelGame_InvalidInputs tests validation
func TestNewDuelGame_InvalidInputs(t *testing.T) {
	player1ID, _ := shared.NewUserID("player1")
	player2ID, _ := shared.NewUserID("player2")

	player1 := NewDuelPlayer(player1ID, "Player1", NewEloRating())
	player2 := NewDuelPlayer(player2ID, "Player2", NewEloRating())

	validQuestions := make([]QuestionID, QuestionsPerDuel)
	for i := 0; i < QuestionsPerDuel; i++ {
		validQuestions[i] = quiz.NewQuestionID()
	}

	now := int64(1000000)

	tests := []struct {
		name        string
		player1     DuelPlayer
		player2     DuelPlayer
		questionIDs []QuestionID
		expectedErr error
	}{
		{
			name:        "Invalid player1 ID",
			player1:     NewDuelPlayer(UserID{}, "Invalid", NewEloRating()),
			player2:     player2,
			questionIDs: validQuestions,
			expectedErr: ErrInvalidGameID,
		},
		{
			name:        "Invalid player2 ID",
			player1:     player1,
			player2:     NewDuelPlayer(UserID{}, "Invalid", NewEloRating()),
			questionIDs: validQuestions,
			expectedErr: ErrInvalidGameID,
		},
		{
			name:        "Too few questions",
			player1:     player1,
			player2:     player2,
			questionIDs: make([]QuestionID, 5),
			expectedErr: ErrInvalidRound,
		},
		{
			name:        "Too many questions",
			player1:     player1,
			player2:     player2,
			questionIDs: make([]QuestionID, 10),
			expectedErr: ErrInvalidRound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game, err := NewDuelGame(tt.player1, tt.player2, tt.questionIDs, now)

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

// TestDuelGame_Start tests starting the game
func TestDuelGame_Start(t *testing.T) {
	player1ID, _ := shared.NewUserID("player1")
	player2ID, _ := shared.NewUserID("player2")

	player1 := NewDuelPlayer(player1ID, "Player1", NewEloRating())
	player2 := NewDuelPlayer(player2ID, "Player2", NewEloRating())

	questionIDs := make([]QuestionID, QuestionsPerDuel)
	for i := 0; i < QuestionsPerDuel; i++ {
		questionIDs[i] = quiz.NewQuestionID()
	}

	now := int64(1000000)

	game, _ := NewDuelGame(player1, player2, questionIDs, now)
	game.Events() // Clear creation event

	// Start game
	err := game.Start(now + 1000)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if game.Status() != GameStatusInProgress {
		t.Errorf("Status = %v, want %v", game.Status(), GameStatusInProgress)
	}
	if game.CurrentRound() != 1 {
		t.Errorf("CurrentRound = %d, want 1", game.CurrentRound())
	}

	// Check events (started + round started)
	events := game.Events()
	if len(events) < 2 {
		t.Errorf("Expected at least 2 events, got %d", len(events))
	}
}

// TestDuelGame_Start_AlreadyStarted tests error when starting twice
func TestDuelGame_Start_AlreadyStarted(t *testing.T) {
	player1ID, _ := shared.NewUserID("player1")
	player2ID, _ := shared.NewUserID("player2")

	player1 := NewDuelPlayer(player1ID, "Player1", NewEloRating())
	player2 := NewDuelPlayer(player2ID, "Player2", NewEloRating())

	questionIDs := make([]QuestionID, QuestionsPerDuel)
	for i := 0; i < QuestionsPerDuel; i++ {
		questionIDs[i] = quiz.NewQuestionID()
	}

	now := int64(1000000)

	game, _ := NewDuelGame(player1, player2, questionIDs, now)
	game.Start(now + 1000)

	// Try to start again
	err := game.Start(now + 2000)

	if err == nil {
		t.Error("Expected error when starting already started game")
	}
	if err != ErrGameNotActive {
		t.Errorf("Expected ErrGameNotActive, got %v", err)
	}
}

// TestDuelGame_CurrentRound tests current round tracking
func TestDuelGame_CurrentRound(t *testing.T) {
	player1ID, _ := shared.NewUserID("player1")
	player2ID, _ := shared.NewUserID("player2")

	player1 := NewDuelPlayer(player1ID, "Player1", NewEloRating())
	player2 := NewDuelPlayer(player2ID, "Player2", NewEloRating())

	questionIDs := make([]QuestionID, QuestionsPerDuel)
	for i := 0; i < QuestionsPerDuel; i++ {
		questionIDs[i] = quiz.NewQuestionID()
	}

	now := int64(1000000)

	game, _ := NewDuelGame(player1, player2, questionIDs, now)

	// Before start
	if game.CurrentRound() != 0 {
		t.Errorf("CurrentRound before start = %d, want 0", game.CurrentRound())
	}

	game.Start(now + 1000)

	// After start
	if game.CurrentRound() != 1 {
		t.Errorf("CurrentRound after start = %d, want 1", game.CurrentRound())
	}
}

// TestDuelGame_IsFinished tests game completion detection
func TestDuelGame_IsFinished(t *testing.T) {
	player1ID, _ := shared.NewUserID("player1")
	player2ID, _ := shared.NewUserID("player2")

	player1 := NewDuelPlayer(player1ID, "Player1", NewEloRating())
	player2 := NewDuelPlayer(player2ID, "Player2", NewEloRating())

	questionIDs := make([]QuestionID, QuestionsPerDuel)
	for i := 0; i < QuestionsPerDuel; i++ {
		questionIDs[i] = quiz.NewQuestionID()
	}

	now := int64(1000000)

	game, _ := NewDuelGame(player1, player2, questionIDs, now)

	// Not started - not finished
	if game.IsFinished() {
		t.Error("Game should not be finished before start")
	}

	game.Start(now + 1000)

	// In progress - not finished
	if game.IsFinished() {
		t.Error("Game should not be finished during progress")
	}
}

// TestDuelGame_PlayerScores tests player score tracking
func TestDuelGame_PlayerScores(t *testing.T) {
	player1ID, _ := shared.NewUserID("player1")
	player2ID, _ := shared.NewUserID("player2")

	// Create players with initial scores
	player1 := NewDuelPlayer(player1ID, "Player1", NewEloRating()).AddScore(500)
	player2 := NewDuelPlayer(player2ID, "Player2", NewEloRating()).AddScore(300)

	questionIDs := make([]QuestionID, QuestionsPerDuel)
	for i := 0; i < QuestionsPerDuel; i++ {
		questionIDs[i] = quiz.NewQuestionID()
	}

	now := int64(1000000)

	// Create game with pre-scored players
	game := &DuelGame{
		id:           NewGameID(),
		player1:      player1,
		player2:      player2,
		questionIDs:  questionIDs,
		currentRound: 3,
		status:       GameStatusInProgress,
		roundAnswers: make(map[int][]RoundAnswer),
		startedAt:    now,
		finishedAt:   0,
		events:       make([]Event, 0),
	}

	// Verify players have correct scores
	if game.Player1().Score() != 500 {
		t.Errorf("Player1 score = %d, want 500", game.Player1().Score())
	}
	if game.Player2().Score() != 300 {
		t.Errorf("Player2 score = %d, want 300", game.Player2().Score())
	}
}

// TestDuelGame_ReconstructDuelGame tests reconstruction from persistence
func TestDuelGame_ReconstructDuelGame(t *testing.T) {
	player1ID, _ := shared.NewUserID("player1")
	player2ID, _ := shared.NewUserID("player2")

	player1 := NewDuelPlayer(player1ID, "Player1", ReconstructEloRating(1200, 50))
	player2 := NewDuelPlayer(player2ID, "Player2", ReconstructEloRating(1100, 30))

	gameID := NewGameID()
	questionIDs := make([]QuestionID, QuestionsPerDuel)
	for i := 0; i < QuestionsPerDuel; i++ {
		questionIDs[i] = quiz.NewQuestionID()
	}

	now := int64(1000000)
	roundAnswers := make(map[int][]RoundAnswer)

	game := ReconstructDuelGame(
		gameID,
		player1,
		player2,
		questionIDs,
		5, // Round 5
		GameStatusInProgress,
		roundAnswers,
		now,
		0, // Not finished
	)

	if game == nil {
		t.Fatal("Expected game to be reconstructed")
	}
	if game.ID() != gameID {
		t.Errorf("ID = %v, want %v", game.ID(), gameID)
	}
	if game.CurrentRound() != 5 {
		t.Errorf("CurrentRound = %d, want 5", game.CurrentRound())
	}
	if game.Status() != GameStatusInProgress {
		t.Errorf("Status = %v, want %v", game.Status(), GameStatusInProgress)
	}

	// Events should be empty after reconstruction
	events := game.Events()
	if len(events) != 0 {
		t.Errorf("Reconstructed game should have no events, got %d", len(events))
	}
}

