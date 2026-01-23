package classic_mode

import (
	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
)

// PersonalBest tracks a player's best performance for a specific quiz
// Used for ghost comparison during gameplay to show progress vs previous best
type PersonalBest struct {
	playerID       shared.UserID
	quizID         quiz.QuizID
	bestScore      int
	maxStreak      int   // Highest streak achieved in the best run
	achievedAt     int64 // Unix timestamp when this record was set
	scoreByQuestion []int // Cumulative score at each question index (for ghost comparison)

	// Domain events collected during operations
	events []Event
}

// NewPersonalBest creates a new PersonalBest record from a completed game
func NewPersonalBest(
	playerID shared.UserID,
	quizID quiz.QuizID,
	game *ClassicGame,
	achievedAt int64,
) (*PersonalBest, error) {
	if playerID.IsZero() {
		return nil, shared.ErrInvalidUserID
	}

	if quizID.IsZero() {
		return nil, quiz.ErrInvalidQuizID
	}

	if game.Status() != GameStatusFinished {
		return nil, ErrGameNotStarted
	}

	// Extract score and streak from game
	finalScore := game.GetTotalScore()
	maxStreak := game.MaxStreak()

	// Get score-by-question for ghost comparison
	scoreByQuestion := game.Session().GetScoreByQuestion()

	return &PersonalBest{
		playerID:        playerID,
		quizID:          quizID,
		bestScore:       finalScore,
		maxStreak:       maxStreak,
		achievedAt:      achievedAt,
		scoreByQuestion: scoreByQuestion,
		events:          make([]Event, 0),
	}, nil
}

// UpdateIfBetter updates the PersonalBest if the provided game achieved a higher score
// Returns true if updated, false if not
func (pb *PersonalBest) UpdateIfBetter(game *ClassicGame, achievedAt int64) (bool, error) {
	if game.Status() != GameStatusFinished {
		return false, ErrGameNotStarted
	}

	if !game.QuizID().Equals(pb.quizID) {
		return false, quiz.ErrQuizNotFound // Game is for different quiz
	}

	// Get final score and streak from game
	finalScore := game.GetTotalScore()
	maxStreak := game.MaxStreak()

	// Check if new score is better
	if finalScore > pb.bestScore {
		previousBestScore := pb.bestScore

		// Update record
		pb.bestScore = finalScore
		pb.maxStreak = maxStreak
		pb.achievedAt = achievedAt
		pb.scoreByQuestion = game.Session().GetScoreByQuestion()

		// Publish PersonalBestAchieved event
		pb.events = append(pb.events, NewPersonalBestAchieved(
			pb.playerID,
			pb.quizID,
			finalScore,
			&previousBestScore,
			maxStreak,
			achievedAt,
		))

		return true, nil
	}

	return false, nil
}

// GetScoreAtQuestion returns the cumulative score at a specific question index
// Used for ghost comparison during gameplay
// Returns 0 if index is out of bounds
func (pb *PersonalBest) GetScoreAtQuestion(questionIndex int) int {
	if questionIndex < 0 || questionIndex >= len(pb.scoreByQuestion) {
		return 0
	}
	return pb.scoreByQuestion[questionIndex]
}

// Getters
func (pb *PersonalBest) PlayerID() shared.UserID       { return pb.playerID }
func (pb *PersonalBest) QuizID() quiz.QuizID           { return pb.quizID }
func (pb *PersonalBest) BestScore() int                { return pb.bestScore }
func (pb *PersonalBest) MaxStreak() int                { return pb.maxStreak }
func (pb *PersonalBest) AchievedAt() int64             { return pb.achievedAt }
func (pb *PersonalBest) ScoreByQuestion() []int        {
	// Return copy to protect internal state
	scores := make([]int, len(pb.scoreByQuestion))
	copy(scores, pb.scoreByQuestion)
	return scores
}

// Events returns collected domain events and clears them
func (pb *PersonalBest) Events() []Event {
	events := pb.events
	pb.events = make([]Event, 0)
	return events
}

// ReconstructPersonalBest reconstructs a PersonalBest from persistence
// Used by repository when loading from database
func ReconstructPersonalBest(
	playerID shared.UserID,
	quizID quiz.QuizID,
	bestScore int,
	maxStreak int,
	achievedAt int64,
	scoreByQuestion []int,
) *PersonalBest {
	return &PersonalBest{
		playerID:        playerID,
		quizID:          quizID,
		bestScore:       bestScore,
		maxStreak:       maxStreak,
		achievedAt:      achievedAt,
		scoreByQuestion: scoreByQuestion,
		events:          make([]Event, 0), // Don't replay events from DB
	}
}
