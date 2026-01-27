package solo_marathon

import (
	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
)

// MarathonGame is the aggregate root for Solo Marathon mode
// Represents a single marathon session with lives, hints, and adaptive difficulty
// NOTE: This is V2 - no longer uses kernel.QuizGameplaySession (endless mode needs dynamic questions)
type MarathonGameV2 struct {
	id               GameID
	playerID         UserID
	category         MarathonCategory
	status           GameStatus
	startedAt        int64
	finishedAt       int64

	// Current question being answered
	currentQuestion *quiz.Question

	// Question history
	answeredQuestionIDs []QuestionID // All answered questions (for persistence)
	recentQuestionIDs   []QuestionID // Last 20 questions (for exclusion logic)

	// Scoring
	currentStreak int
	maxStreak     int
	baseScore     int // Total base score earned

	// Marathon-specific mechanics
	lives      LivesSystem
	hints      HintsSystem
	difficulty DifficultyProgression

	// Personal best reference (for comparison)
	personalBestStreak *int

	// Track hint usage per question
	usedHints map[QuestionID][]HintType

	// Domain events
	events []Event
}

// NewMarathonGameV2 creates a new marathon game session
// PersonalBest can be nil if the player has no previous record for this category
// NOTE: Game starts WITHOUT a question - call LoadNextQuestion() after creation
func NewMarathonGameV2(
	playerID UserID,
	category MarathonCategory,
	personalBest *PersonalBest,
	startedAt int64,
) (*MarathonGameV2, error) {
	// Validate inputs
	if playerID.IsZero() {
		return nil, ErrInvalidGameID
	}

	// Extract PersonalBest streak if exists
	var personalBestStreak *int
	hasPersonalBest := false
	if personalBest != nil {
		streak := personalBest.BestStreak()
		personalBestStreak = &streak
		hasPersonalBest = true
	}

	// Initialize systems
	lives := NewLivesSystem(startedAt)
	hints := NewHintsSystem()
	difficulty := NewDifficultyProgression()

	// Create game
	game := &MarathonGameV2{
		id:                  NewGameID(),
		playerID:            playerID,
		category:            category,
		status:              GameStatusInProgress,
		startedAt:           startedAt,
		finishedAt:          0,
		currentQuestion:     nil, // Will be loaded via LoadNextQuestion()
		answeredQuestionIDs: make([]QuestionID, 0),
		recentQuestionIDs:   make([]QuestionID, 0),
		currentStreak:       0,
		maxStreak:           0,
		baseScore:           0,
		lives:               lives,
		hints:               hints,
		difficulty:          difficulty,
		personalBestStreak:  personalBestStreak,
		usedHints:           make(map[QuestionID][]HintType),
		events:              make([]Event, 0),
	}

	// Publish MarathonGameStarted event
	game.events = append(game.events, NewMarathonGameStartedEvent(
		game.id,
		playerID,
		category,
		hasPersonalBest,
		personalBestStreak,
		startedAt,
	))

	return game, nil
}

// LoadNextQuestion loads the next question using QuestionSelector Domain Service
// This is called:
// 1. After game creation (to get first question)
// 2. After each correct answer (to get next question)
func (mg *MarathonGameV2) LoadNextQuestion(questionSelector *QuestionSelector) error {
	if mg.status != GameStatusInProgress {
		return ErrGameNotActive
	}

	// Select next question based on current state
	question, err := questionSelector.SelectNextQuestion(
		mg.category,
		mg.difficulty,
		mg.recentQuestionIDs, // Exclude recent questions
	)
	if err != nil {
		return err
	}

	// Set as current question
	mg.currentQuestion = question

	// Update recent questions list (sliding window of 20)
	mg.recentQuestionIDs = append(mg.recentQuestionIDs, question.ID())
	if len(mg.recentQuestionIDs) > 20 {
		mg.recentQuestionIDs = mg.recentQuestionIDs[1:] // Remove oldest
	}

	return nil
}

// AnswerQuestionResult holds detailed information about a submitted answer
type AnswerQuestionResultV2 struct {
	IsCorrect       bool
	BasePoints      int
	TimeTaken       int64
	CurrentStreak   int
	MaxStreak       int
	DifficultyLevel DifficultyLevel
	LifeLost        bool
	RemainingLives  int
	IsGameOver      bool
}

// AnswerQuestion processes a user's answer in marathon mode
// Business Logic:
// - Correct answer: increment streak, update difficulty, earn points
// - Incorrect answer: lose life, reset streak, check game over
func (mg *MarathonGameV2) AnswerQuestion(
	questionID QuestionID,
	answerID AnswerID,
	timeTaken int64,
	answeredAt int64,
) (*AnswerQuestionResultV2, error) {
	// 1. Validate game state
	if mg.status != GameStatusInProgress {
		return nil, ErrGameNotActive
	}

	if !mg.lives.HasLives() {
		return nil, ErrNoLivesRemaining
	}

	if mg.currentQuestion == nil {
		return nil, ErrInvalidQuestion
	}

	// 2. Validate question ID matches current question
	if !mg.currentQuestion.ID().Equals(questionID) {
		return nil, ErrInvalidQuestion
	}

	// 3. Get the answer
	answer, err := mg.currentQuestion.GetAnswer(answerID)
	if err != nil {
		return nil, err
	}

	// 4. Check correctness
	isCorrect := answer.IsCorrect()

	// 5. Calculate base points (only if correct)
	var basePoints int
	if isCorrect {
		basePoints = mg.currentQuestion.Points().Value()
		mg.baseScore += basePoints
	}

	// 6. Initialize result
	result := &AnswerQuestionResultV2{
		IsCorrect:       isCorrect,
		BasePoints:      basePoints,
		TimeTaken:       timeTaken,
		CurrentStreak:   mg.currentStreak,
		MaxStreak:       mg.maxStreak,
		DifficultyLevel: mg.difficulty.Level(),
		LifeLost:        false,
		RemainingLives:  mg.lives.CurrentLives(),
		IsGameOver:      false,
	}

	// 7. Process answer based on correctness
	if isCorrect {
		// === CORRECT ANSWER ===

		// a. Increment streak
		mg.currentStreak++
		result.CurrentStreak = mg.currentStreak

		// b. Update max streak
		if mg.currentStreak > mg.maxStreak {
			mg.maxStreak = mg.currentStreak
		}
		result.MaxStreak = mg.maxStreak

		// c. Update difficulty based on new streak
		previousLevel := mg.difficulty.Level()
		mg.difficulty = mg.difficulty.UpdateFromStreak(mg.currentStreak)

		// d. Fire DifficultyIncreased event if level changed
		if mg.difficulty.Level() != previousLevel {
			mg.events = append(mg.events, NewDifficultyIncreasedEvent(
				mg.id,
				mg.playerID,
				mg.difficulty.Level(),
				mg.currentStreak,
				answeredAt,
			))
		}

		result.DifficultyLevel = mg.difficulty.Level()

	} else {
		// === INCORRECT ANSWER ===

		// a. Lose life
		mg.lives = mg.lives.LoseLife(answeredAt)
		result.LifeLost = true
		result.RemainingLives = mg.lives.CurrentLives()

		// b. Publish LifeLost event
		mg.events = append(mg.events, NewLifeLostEvent(
			mg.id,
			mg.playerID,
			questionID,
			mg.lives.CurrentLives(),
			answeredAt,
		))

		// c. Reset streak
		mg.currentStreak = 0
		result.CurrentStreak = 0

		// d. Reset difficulty
		mg.difficulty = NewDifficultyProgression()
		result.DifficultyLevel = mg.difficulty.Level()

		// e. Check if game over (no lives remaining)
		if !mg.lives.HasLives() {
			// Validate state transition
			if !mg.status.CanTransitionTo(GameStatusFinished) {
				return nil, ErrInvalidGameStatus
			}

			mg.status = GameStatusFinished
			mg.finishedAt = answeredAt
			result.IsGameOver = true

			// Determine if new record
			isNewRecord := false
			if mg.personalBestStreak == nil || mg.maxStreak > *mg.personalBestStreak {
				isNewRecord = true
			}

			// Publish MarathonGameOver event
			mg.events = append(mg.events, NewMarathonGameOverEvent(
				mg.id,
				mg.playerID,
				mg.maxStreak,
				isNewRecord,
				mg.personalBestStreak,
				answeredAt,
			))
		}
	}

	// 8. Record question as answered
	mg.answeredQuestionIDs = append(mg.answeredQuestionIDs, questionID)

	// 9. Clear current question (will be loaded next time)
	mg.currentQuestion = nil

	// 10. Publish MarathonQuestionAnswered event
	mg.events = append(mg.events, NewMarathonQuestionAnsweredEvent(
		mg.id,
		mg.playerID,
		questionID,
		answerID,
		isCorrect,
		timeTaken,
		mg.currentStreak,
		mg.difficulty.Level(),
		answeredAt,
	))

	return result, nil
}

// UseHint allows player to use a hint for the current question
func (mg *MarathonGameV2) UseHint(questionID QuestionID, hintType HintType, usedAt int64) error {
	// Validate game state
	if mg.status != GameStatusInProgress {
		return ErrGameNotActive
	}

	if mg.currentQuestion == nil {
		return ErrInvalidQuestion
	}

	// Validate question ID matches current question
	if !mg.currentQuestion.ID().Equals(questionID) {
		return ErrInvalidQuestion
	}

	// Check if hint already used for this question
	if usedHints, exists := mg.usedHints[questionID]; exists {
		for _, usedHint := range usedHints {
			if usedHint == hintType {
				return ErrHintAlreadyUsed
			}
		}
	}

	// Use hint (immutable operation)
	newHints, err := mg.hints.UseHint(hintType)
	if err != nil {
		return err
	}

	mg.hints = newHints

	// Record hint usage
	mg.usedHints[questionID] = append(mg.usedHints[questionID], hintType)

	// Get remaining hints of this type
	var remainingHints int
	switch hintType {
	case HintFiftyFifty:
		remainingHints = mg.hints.FiftyFifty()
	case HintExtraTime:
		remainingHints = mg.hints.ExtraTime()
	case HintSkip:
		remainingHints = mg.hints.Skip()
	}

	// Publish HintUsed event
	mg.events = append(mg.events, NewHintUsedEvent(
		mg.id,
		mg.playerID,
		questionID,
		hintType,
		remainingHints,
		usedAt,
	))

	return nil
}

// Abandon marks the game as abandoned (player quit voluntarily)
func (mg *MarathonGameV2) Abandon(abandonedAt int64) error {
	if mg.status != GameStatusInProgress {
		return ErrGameNotActive
	}

	// Validate state transition
	if !mg.status.CanTransitionTo(GameStatusAbandoned) {
		return ErrInvalidGameStatus
	}

	mg.status = GameStatusAbandoned
	mg.finishedAt = abandonedAt

	// Determine if new record
	isNewRecord := false
	if mg.personalBestStreak == nil || mg.maxStreak > *mg.personalBestStreak {
		isNewRecord = true
	}

	// Publish MarathonGameOver event
	mg.events = append(mg.events, NewMarathonGameOverEvent(
		mg.id,
		mg.playerID,
		mg.maxStreak,
		isNewRecord,
		mg.personalBestStreak,
		abandonedAt,
	))

	return nil
}

// IsGameOver checks if game is finished or abandoned
func (mg *MarathonGameV2) IsGameOver() bool {
	return mg.status.IsTerminal()
}

// GetCurrentQuestion returns the current question to be answered
func (mg *MarathonGameV2) GetCurrentQuestion() (*quiz.Question, error) {
	if mg.currentQuestion == nil {
		return nil, ErrInvalidQuestion
	}
	return mg.currentQuestion, nil
}

// IsNewPersonalBest checks if this game's maxStreak is a new personal best
func (mg *MarathonGameV2) IsNewPersonalBest() bool {
	if mg.personalBestStreak == nil {
		return mg.maxStreak > 0
	}
	return mg.maxStreak > *mg.personalBestStreak
}

// Getters
func (mg *MarathonGameV2) ID() GameID                                  { return mg.id }
func (mg *MarathonGameV2) PlayerID() UserID                            { return mg.playerID }
func (mg *MarathonGameV2) Category() MarathonCategory                  { return mg.category }
func (mg *MarathonGameV2) Status() GameStatus                          { return mg.status }
func (mg *MarathonGameV2) StartedAt() int64                            { return mg.startedAt }
func (mg *MarathonGameV2) FinishedAt() int64                           { return mg.finishedAt }
func (mg *MarathonGameV2) CurrentQuestion() *quiz.Question             { return mg.currentQuestion }
func (mg *MarathonGameV2) AnsweredQuestionIDs() []QuestionID           { return mg.answeredQuestionIDs }
func (mg *MarathonGameV2) RecentQuestionIDs() []QuestionID             { return mg.recentQuestionIDs }
func (mg *MarathonGameV2) CurrentStreak() int                          { return mg.currentStreak }
func (mg *MarathonGameV2) MaxStreak() int                              { return mg.maxStreak }
func (mg *MarathonGameV2) BaseScore() int                              { return mg.baseScore }
func (mg *MarathonGameV2) Lives() LivesSystem                          { return mg.lives }
func (mg *MarathonGameV2) Hints() HintsSystem                          { return mg.hints }
func (mg *MarathonGameV2) Difficulty() DifficultyProgression           { return mg.difficulty }
func (mg *MarathonGameV2) PersonalBestStreak() *int                    { return mg.personalBestStreak }

// Events returns collected domain events and clears them
func (mg *MarathonGameV2) Events() []Event {
	events := mg.events
	mg.events = make([]Event, 0)
	return events
}

// ReconstructMarathonGameV2 reconstructs a MarathonGame from persistence
// Used by repository when loading from database
func ReconstructMarathonGameV2(
	id GameID,
	playerID UserID,
	category MarathonCategory,
	status GameStatus,
	startedAt int64,
	finishedAt int64,
	currentQuestion *quiz.Question,
	answeredQuestionIDs []QuestionID,
	recentQuestionIDs []QuestionID,
	currentStreak int,
	maxStreak int,
	baseScore int,
	lives LivesSystem,
	hints HintsSystem,
	difficulty DifficultyProgression,
	personalBestStreak *int,
	usedHints map[QuestionID][]HintType,
) *MarathonGameV2 {
	return &MarathonGameV2{
		id:                  id,
		playerID:            playerID,
		category:            category,
		status:              status,
		startedAt:           startedAt,
		finishedAt:          finishedAt,
		currentQuestion:     currentQuestion,
		answeredQuestionIDs: answeredQuestionIDs,
		recentQuestionIDs:   recentQuestionIDs,
		currentStreak:       currentStreak,
		maxStreak:           maxStreak,
		baseScore:           baseScore,
		lives:               lives,
		hints:               hints,
		difficulty:          difficulty,
		personalBestStreak:  personalBestStreak,
		usedHints:           usedHints,
		events:              make([]Event, 0), // Don't replay events from DB
	}
}
