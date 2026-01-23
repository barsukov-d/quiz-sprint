package classic_mode

import (
	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
)

// Event is the interface for all domain events
type Event interface {
	EventName() string
	OccurredAt() int64
}

// ClassicGameStarted is published when a new classic game is created
type ClassicGameStarted struct {
	gameID           GameID
	playerID         shared.UserID
	quizID           quiz.QuizID
	hasPersonalBest  bool
	personalBestScore *int
	occurredAt       int64
}

func NewClassicGameStarted(
	gameID GameID,
	playerID shared.UserID,
	quizID quiz.QuizID,
	hasPersonalBest bool,
	personalBestScore *int,
	occurredAt int64,
) ClassicGameStarted {
	return ClassicGameStarted{
		gameID:            gameID,
		playerID:          playerID,
		quizID:            quizID,
		hasPersonalBest:   hasPersonalBest,
		personalBestScore: personalBestScore,
		occurredAt:        occurredAt,
	}
}

func (e ClassicGameStarted) EventName() string           { return "classic_mode.game_started" }
func (e ClassicGameStarted) OccurredAt() int64           { return e.occurredAt }
func (e ClassicGameStarted) GameID() GameID              { return e.gameID }
func (e ClassicGameStarted) PlayerID() shared.UserID     { return e.playerID }
func (e ClassicGameStarted) QuizID() quiz.QuizID         { return e.quizID }
func (e ClassicGameStarted) HasPersonalBest() bool       { return e.hasPersonalBest }
func (e ClassicGameStarted) PersonalBestScore() *int     { return e.personalBestScore }

// ClassicGameFinished is published when a game is completed
type ClassicGameFinished struct {
	gameID             GameID
	playerID           shared.UserID
	quizID             quiz.QuizID
	finalScore         int
	maxStreak          int
	isNewPersonalBest  bool
	occurredAt         int64
}

func NewClassicGameFinished(
	gameID GameID,
	playerID shared.UserID,
	quizID quiz.QuizID,
	finalScore int,
	maxStreak int,
	isNewPersonalBest bool,
	occurredAt int64,
) ClassicGameFinished {
	return ClassicGameFinished{
		gameID:            gameID,
		playerID:          playerID,
		quizID:            quizID,
		finalScore:        finalScore,
		maxStreak:         maxStreak,
		isNewPersonalBest: isNewPersonalBest,
		occurredAt:        occurredAt,
	}
}

func (e ClassicGameFinished) EventName() string          { return "classic_mode.game_finished" }
func (e ClassicGameFinished) OccurredAt() int64          { return e.occurredAt }
func (e ClassicGameFinished) GameID() GameID             { return e.gameID }
func (e ClassicGameFinished) PlayerID() shared.UserID    { return e.playerID }
func (e ClassicGameFinished) QuizID() quiz.QuizID        { return e.quizID }
func (e ClassicGameFinished) FinalScore() int            { return e.finalScore }
func (e ClassicGameFinished) MaxStreak() int             { return e.maxStreak }
func (e ClassicGameFinished) IsNewPersonalBest() bool    { return e.isNewPersonalBest }

// StreakMilestoneReached is published when player reaches significant streak milestones (3 or 6)
// Used for dramatic UI feedback (On Fire / Godlike transitions)
type StreakMilestoneReached struct {
	gameID       GameID
	playerID     shared.UserID
	streakCount  int
	visualState  VisualState
	multiplier   Multiplier
	occurredAt   int64
}

func NewStreakMilestoneReached(
	gameID GameID,
	playerID shared.UserID,
	streakCount int,
	visualState VisualState,
	multiplier Multiplier,
	occurredAt int64,
) StreakMilestoneReached {
	return StreakMilestoneReached{
		gameID:      gameID,
		playerID:    playerID,
		streakCount: streakCount,
		visualState: visualState,
		multiplier:  multiplier,
		occurredAt:  occurredAt,
	}
}

func (e StreakMilestoneReached) EventName() string       { return "classic_mode.streak_milestone_reached" }
func (e StreakMilestoneReached) OccurredAt() int64       { return e.occurredAt }
func (e StreakMilestoneReached) GameID() GameID          { return e.gameID }
func (e StreakMilestoneReached) PlayerID() shared.UserID { return e.playerID }
func (e StreakMilestoneReached) StreakCount() int        { return e.streakCount }
func (e StreakMilestoneReached) VisualState() VisualState { return e.visualState }
func (e StreakMilestoneReached) Multiplier() Multiplier  { return e.multiplier }

// StreakBroken is published when a player's in-game streak resets due to incorrect answer or timeout
// Triggers dramatic shake effect in UI
type StreakBroken struct {
	gameID         GameID
	playerID       shared.UserID
	previousStreak int
	questionID     quiz.QuestionID
	occurredAt     int64
}

func NewStreakBroken(
	gameID GameID,
	playerID shared.UserID,
	previousStreak int,
	questionID quiz.QuestionID,
	occurredAt int64,
) StreakBroken {
	return StreakBroken{
		gameID:         gameID,
		playerID:       playerID,
		previousStreak: previousStreak,
		questionID:     questionID,
		occurredAt:     occurredAt,
	}
}

func (e StreakBroken) EventName() string            { return "classic_mode.streak_broken" }
func (e StreakBroken) OccurredAt() int64            { return e.occurredAt }
func (e StreakBroken) GameID() GameID               { return e.gameID }
func (e StreakBroken) PlayerID() shared.UserID      { return e.playerID }
func (e StreakBroken) PreviousStreak() int          { return e.previousStreak }
func (e StreakBroken) QuestionID() quiz.QuestionID  { return e.questionID }

// PersonalBestAchieved is published when a player sets a new PersonalBest for a quiz
type PersonalBestAchieved struct {
	playerID          shared.UserID
	quizID            quiz.QuizID
	newBestScore      int
	previousBestScore *int
	maxStreak         int
	occurredAt        int64
}

func NewPersonalBestAchieved(
	playerID shared.UserID,
	quizID quiz.QuizID,
	newBestScore int,
	previousBestScore *int,
	maxStreak int,
	occurredAt int64,
) PersonalBestAchieved {
	return PersonalBestAchieved{
		playerID:          playerID,
		quizID:            quizID,
		newBestScore:      newBestScore,
		previousBestScore: previousBestScore,
		maxStreak:         maxStreak,
		occurredAt:        occurredAt,
	}
}

func (e PersonalBestAchieved) EventName() string         { return "classic_mode.personal_best_achieved" }
func (e PersonalBestAchieved) OccurredAt() int64         { return e.occurredAt }
func (e PersonalBestAchieved) PlayerID() shared.UserID   { return e.playerID }
func (e PersonalBestAchieved) QuizID() quiz.QuizID       { return e.quizID }
func (e PersonalBestAchieved) NewBestScore() int         { return e.newBestScore }
func (e PersonalBestAchieved) PreviousBestScore() *int   { return e.previousBestScore }
func (e PersonalBestAchieved) MaxStreak() int            { return e.maxStreak }

// EventBus is the interface for publishing domain events
// Defined in domain, implemented in infrastructure
type EventBus interface {
	// Publish publishes events asynchronously
	Publish(events ...Event)

	// Subscribe registers a handler for a specific event type
	Subscribe(eventName string, handler EventHandler)
}

// EventHandler is a function that handles a domain event
type EventHandler func(event Event)
