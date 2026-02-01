package daily_challenge

import (
	"github.com/barsukov/quiz-sprint/backend/internal/domain/kernel"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
)

// DailyGame is the aggregate root for a player's daily challenge attempt
type DailyGame struct {
	id                GameID
	playerID          UserID
	dailyQuizID       DailyQuizID
	date              Date
	status            GameStatus
	session           *kernel.QuizGameplaySession // Composition: delegates pure gameplay logic
	streak            StreakSystem                // Daily streak tracking
	rank              *int                        // Player's rank in leaderboard (nil if not yet calculated)
	chestReward       *ChestReward                // Chest earned (nil until game completed)
	questionStartedAt int64                       // Unix timestamp when current question started (for timer persistence)

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
		id:                gameID,
		playerID:          playerID,
		dailyQuizID:       dailyQuizID,
		date:              date,
		status:            GameStatusInProgress,
		session:           session,
		streak:            currentStreak,
		rank:              nil,
		questionStartedAt: startedAt, // First question starts immediately
		events:            make([]Event, 0),
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
	QuestionIndex      int
	TimeTaken          int64
	RemainingQuestions int
	IsGameCompleted    bool
	IsCorrect          bool
	CorrectAnswerID    string
}

// AnswerQuestion processes a user's answer for daily challenge
// Business Logic:
// - Instant feedback (correct/incorrect shown immediately after each answer)
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

	// 2. Find correct answer ID before answering (question is accessible now)
	correctAnswerID := ""
	if question, err := dg.session.Quiz().GetQuestion(questionID); err == nil {
		for _, answer := range question.Answers() {
			if answer.IsCorrect() {
				correctAnswerID = answer.ID().String()
				break
			}
		}
	}

	// 3. Delegate to kernel session for pure gameplay logic
	kernelResult, err := dg.session.AnswerQuestion(questionID, answerID, timeTaken, answeredAt)
	if err != nil {
		return nil, err
	}

	// 4. Build result with instant feedback
	result := &AnswerQuestionResult{
		QuestionIndex:      dg.session.CurrentQuestionIndex() - 1, // -1 because we already moved forward
		TimeTaken:          timeTaken,
		RemainingQuestions: dg.session.Quiz().QuestionsCount() - dg.session.CurrentQuestionIndex(),
		IsGameCompleted:    dg.session.IsFinished(),
		IsCorrect:          kernelResult.IsCorrect,
		CorrectAnswerID:    correctAnswerID,
	}

	// 4. Update questionStartedAt for next question (if not finished)
	if !dg.session.IsFinished() {
		dg.questionStartedAt = answeredAt
	}

	// 5. Publish DailyQuestionAnswered event
	dg.events = append(dg.events, NewDailyQuestionAnsweredEvent(
		dg.id,
		dg.playerID,
		questionID,
		answerID,
		timeTaken,
		answeredAt,
	))

	// 6. Auto-complete if all questions answered
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
	println("🔥 [DailyGame.complete] Before UpdateForDate:")
	println("   - currentStreak:", previousStreak)
	println("   - lastPlayedDate:", dg.streak.LastPlayedDate().String())
	println("   - playedDate:", dg.date.String())

	dg.streak = dg.streak.UpdateForDate(dg.date)

	println("🔥 [DailyGame.complete] After UpdateForDate:")
	println("   - newStreak:", dg.streak.CurrentStreak())
	println("   - lastPlayedDate:", dg.streak.LastPlayedDate().String())

	// 4. Apply streak bonus to score
	baseScore := dg.session.BaseScore().Value()
	streakBonus := dg.streak.GetBonus()
	finalScore := int(float64(baseScore) * streakBonus)

	println("💰 [DailyGame.complete] Score calculation:")
	println("   - baseScore:", baseScore)
	println("   - streakBonus:", streakBonus)
	println("   - finalScore:", finalScore)

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

// isStreakMilestone checks if streak is a milestone (3, 7, 14, 30)
// Per docs/game_modes/daily_challenge/03_rules.md
func isStreakMilestone(streak int) bool {
	milestones := []int{3, 7, 14, 30}
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

// SetChestReward sets the chest reward (called by application layer after calculation)
func (dg *DailyGame) SetChestReward(reward ChestReward) error {
	if !dg.status.IsTerminal() {
		return ErrGameNotActive
	}
	dg.chestReward = &reward
	return nil
}

// EmitChestEarnedEvent emits chest earned event (called by application layer)
func (dg *DailyGame) EmitChestEarnedEvent(reward ChestReward, occurredAt int64) {
	dg.events = append(dg.events, NewChestEarnedEvent(
		dg.id,
		dg.playerID,
		dg.date,
		reward,
		dg.streak.GetBonus(),
		occurredAt,
	))
}

// Getters
func (dg *DailyGame) ID() GameID                           { return dg.id }
func (dg *DailyGame) PlayerID() UserID                     { return dg.playerID }
func (dg *DailyGame) DailyQuizID() DailyQuizID             { return dg.dailyQuizID }
func (dg *DailyGame) Date() Date                           { return dg.date }
func (dg *DailyGame) Status() GameStatus                   { return dg.status }
func (dg *DailyGame) Session() *kernel.QuizGameplaySession { return dg.session }
func (dg *DailyGame) Streak() StreakSystem                 { return dg.streak }
func (dg *DailyGame) Rank() *int                           { return dg.rank }
func (dg *DailyGame) ChestReward() *ChestReward            { return dg.chestReward }
func (dg *DailyGame) QuestionStartedAt() int64             { return dg.questionStartedAt }
func (dg *DailyGame) IsCompleted() bool                    { return dg.status.IsTerminal() }

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
	chestReward *ChestReward,
	questionStartedAt int64,
) *DailyGame {
	return &DailyGame{
		id:                id,
		playerID:          playerID,
		dailyQuizID:       dailyQuizID,
		date:              date,
		status:            status,
		session:           session,
		streak:            streak,
		rank:              rank,
		chestReward:       chestReward,
		questionStartedAt: questionStartedAt,
		events:      make([]Event, 0), // Don't replay events from DB
	}
}
