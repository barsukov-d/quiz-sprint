package classic_mode

import (
	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
)

// ClassicGameRepository defines the interface for ClassicGame persistence
// NOTE: No context.Context - domain layer is pure
// Infrastructure implementations add context internally
type ClassicGameRepository interface {
	// FindByID retrieves a game by its ID
	FindByID(id GameID) (*ClassicGame, error)

	// FindActiveByPlayerAndQuiz finds an active game for a player and quiz
	FindActiveByPlayerAndQuiz(playerID shared.UserID, quizID quiz.QuizID) (*ClassicGame, error)

	// FindAllActiveByPlayer retrieves all active games for a player
	FindAllActiveByPlayer(playerID shared.UserID) ([]*ClassicGame, error)

	// Save persists a game (create or update)
	Save(game *ClassicGame) error

	// Delete removes a game by ID
	Delete(id GameID) error
}

// PersonalBestRepository defines the interface for PersonalBest persistence
type PersonalBestRepository interface {
	// FindByPlayerAndQuiz retrieves a PersonalBest record for a player and quiz
	FindByPlayerAndQuiz(playerID shared.UserID, quizID quiz.QuizID) (*PersonalBest, error)

	// FindAllByPlayer retrieves all PersonalBest records for a player
	FindAllByPlayer(playerID shared.UserID) ([]*PersonalBest, error)

	// Save persists a PersonalBest record (create or update)
	Save(personalBest *PersonalBest) error

	// Delete removes a PersonalBest record
	Delete(playerID shared.UserID, quizID quiz.QuizID) error
}
