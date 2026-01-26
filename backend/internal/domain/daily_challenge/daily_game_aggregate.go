package daily_challenge

import (
	"github.com/barsukov/quiz-sprint/backend/internal/domain/kernel"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
)

// DailyGame is the aggregate root for a player's daily challenge attempt
type DailyGame struct {
	id            GameID
	playerID      UserID
	dailyQuizID   DailyQuizID
	date          Date
	status        GameStatus
	session       *kernel.QuizGameplaySession // Composition: delegates pure gameplay logic
	streak        StreakSystem                // Daily streak tracking
	rank          *int                        // Player's rank in leaderboard (nil if not yet calculated)

	// Domain events collected during operations
	events []Event
}

// NewDailyGame creates a new daily game for a player
func NewDailyGame(
	playerID UserID,
	dailyQuizID DailyQuizID,
	date Date,
	quizAggregate *quiz.Quiz,
	currentStreak StreakSystem,
	startedAt int64,
) (*DailyGame, error) {
	// 1. Validate inputs
	if playerID.IsZero() {
		return nil, ErrInvalidGameID
	}

	if dailyQuizID.IsZero() {
		return nil, ErrInvalidDailyQuizID
	}

	if date.IsZero() {
		return nil, ErrInvalidDate
	}

	if quizAggregate == nil {
		return nil, quiz.ErrQuizNotFound
	}

	if err := quizAggregate.CanStart(); err != nil {
		return nil, err
	}

	// 2. Create kernel gameplay session
	gameID := NewGameID()
	sessionID := kernel.NewSessionID()

	session, err := kernel.NewQuizGameplaySession(sessionID, quizAggregate, startedAt)
	if err != nil {
		return nil, err
	}

	// 3. Create game
	game := &DailyGame{
		id:          gameID,
		playerID:    playerID,
		dailyQuizID: dailyQuizID,
		date:        date,
		status:      GameStatusInProgress,
		session:     session,
		streak:      currentStreak,
		rank:        nil,
		events:      make([]Event, 0),
	}

	// 4. Publish DailyGameStarted event
	game.events = append(game.events, NewDailyGameStartedEvent(
		gameID,
		playerID,
		dailyQuizID,
		date,
		currentStreak.CurrentStreak(),
		startedAt,
	))

	return game, nil
}

// AnswerQuestionResult holds result of answering a question
type AnswerQuestionResult struct {
	QuestionIndex  int
	TimeTaken      int64
	RemainingQuestions int
	IsGameCompleted bool
}

// AnswerQuestion processes a user's answer for daily challenge
// Business Logic:
// - No immediate feedback (correct/incorrect not shown until completion)
// - Track answers and continue to next question
// - Complete game after 10 questions
func (dg *DailyGame) AnswerQuestion(
	questionID QuestionID,
	answerID AnswerID,
	timeTaken int64,
	answeredAt int64,
) (*AnswerQuestionResult, error) {
	// 1. Validate game state
	if dg.status != GameStatusInProgress {
		return nil, ErrGameNotActive
	}

	// 2. Delegate to kernel session for pure gameplay logic
	_, err := dg.session.AnswerQuestion(questionID, answerID, timeTaken, answeredAt)
	if err != nil {
		return nil, err
	}

	// 3. Build result (NO correctness feedback until completion)
	result := &AnswerQuestionResult{
		QuestionIndex:      dg.session.CurrentQuestionIndex() - 1, // -1 because we already moved forward
		TimeTaken:          timeTaken,
		RemainingQuestions: dg.session.Quiz().QuestionsCount() - dg.session.CurrentQuestionIndex(),
		IsGameCompleted:    dg.session.IsFinished(),
	}

	// 4. Publish DailyQuestionAnswered event
	dg.events = append(dg.events, NewDailyQuestionAnsweredEvent(
		dg.id,
		dg.playerID,
		questionID,
		answerID,
		timeTaken,
		answeredAt,
	))

	// 5. Auto-complete if all questions answered
	if dg.session.IsFinished() {
		if err := dg.complete(answeredAt); err != nil {
			return nil, err
		}
		result.IsGameCompleted = true
	}

	return result, nil
}

// complete marks the game as completed and updates streak
func (dg *DailyGame) complete(completedAt int64) error {
	if dg.status == GameStatusCompleted {
		return ErrGameAlreadyCompleted
	}

	// 1. Mark session as finished
	if err := dg.session.Finish(completedAt); err != nil {
		return err
	}

	// 2. Validate state transition
	if !dg.status.CanTransitionTo(GameStatusCompleted) {
		return ErrInvalidGameStatus
	}

	// 3. Mark game as completed
	dg.status = GameStatusCompleted

	// 3. Update streak for this date
	previousStreak := dg.streak.CurrentStreak()
	dg.streak = dg.streak.UpdateForDate(dg.date)

	// 4. Apply streak bonus to score
	baseScore := dg.session.BaseScore().Value()
	streakBonus := dg.streak.GetBonus()
	finalScore := int(float64(baseScore) * streakBonus)

	// 5. Check for streak milestone
	if isStreakMilestone(dg.streak.CurrentStreak()) && dg.streak.CurrentStreak() > previousStreak {
		bonusPercent := int((streakBonus - 1.0) * 100)
		dg.events = append(dg.events, NewStreakMilestoneReachedEvent(
			dg.id,
			dg.playerID,
			dg.streak.CurrentStreak(),
			bonusPercent,
			completedAt,
		))
	}

	// 6. Get correct answers count
	correctAnswers := dg.session.CountCorrectAnswers()
	totalQuestions := dg.session.Quiz().QuestionsCount()

	// 7. Publish DailyGameCompleted event
	dg.events = append(dg.events, NewDailyGameCompletedEvent(
		dg.id,
		dg.playerID,
		dg.dailyQuizID,
		dg.date,
		finalScore,
		correctAnswers,
		totalQuestions,
		dg.streak.CurrentStreak(),
		streakBonus,
		dg.rank, // Rank will be calculated by application layer later
		completedAt,
	))

	return nil
}

// isStreakMilestone checks if streak is a milestone (3, 7, 14, 30, 100)
func isStreakMilestone(streak int) bool {
	milestones := []int{3, 7, 14, 30, 100}
	for _, m := range milestones {
		if streak == m {
			return true
		}
	}
	return false
}

// GetFinalScore returns final score with streak bonus applied
func (dg *DailyGame) GetFinalScore() int {
	baseScore := dg.session.BaseScore().Value()
	streakBonus := dg.streak.GetBonus()
	return int(float64(baseScore) * streakBonus)
}

// GetCorrectAnswersCount returns number of correct answers
func (dg *DailyGame) GetCorrectAnswersCount() int {
	return dg.session.CountCorrectAnswers()
}

// SetRank sets the player's rank in leaderboard (called by application layer)
func (dg *DailyGame) SetRank(rank int) {
	dg.rank = &rank
}

// Getters
func (dg *DailyGame) ID() GameID                     { return dg.id }
func (dg *DailyGame) PlayerID() UserID               { return dg.playerID }
func (dg *DailyGame) DailyQuizID() DailyQuizID       { return dg.dailyQuizID }
func (dg *DailyGame) Date() Date                     { return dg.date }
func (dg *DailyGame) Status() GameStatus             { return dg.status }
func (dg *DailyGame) Session() *kernel.QuizGameplaySession { return dg.session }
func (dg *DailyGame) Streak() StreakSystem           { return dg.streak }
func (dg *DailyGame) Rank() *int                     { return dg.rank }
func (dg *DailyGame) IsCompleted() bool              { return dg.status.IsTerminal() }

// Events returns collected domain events and clears them
func (dg *DailyGame) Events() []Event {
	events := dg.events
	dg.events = make([]Event, 0)
	return events
}

// ReconstructDailyGame reconstructs a DailyGame from persistence
// Used by repository when loading from database
func ReconstructDailyGame(
	id GameID,
	playerID UserID,
	dailyQuizID DailyQuizID,
	date Date,
	status GameStatus,
	session *kernel.QuizGameplaySession,
	streak StreakSystem,
	rank *int,
) *DailyGame {
	return &DailyGame{
		id:          id,
		playerID:    playerID,
		dailyQuizID: dailyQuizID,
		date:        date,
		status:      status,
		session:     session,
		streak:      streak,
		rank:        rank,
		events:      make([]Event, 0), // Don't replay events from DB
	}
}
