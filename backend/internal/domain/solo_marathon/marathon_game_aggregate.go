package solo_marathon

import (
	"github.com/barsukov/quiz-sprint/backend/internal/domain/kernel"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
)

// MarathonGame is the aggregate root for Solo Marathon mode
// Represents a single marathon session with lives, hints, and adaptive difficulty
type MarathonGame struct {
	id               GameID
	playerID         UserID
	category         MarathonCategory
	status           GameStatus
	session          *kernel.QuizGameplaySession // Composition: delegates pure gameplay logic
	currentStreak    int                         // Current streak of correct answers
	maxStreak        int                         // Maximum streak achieved during this game
	lives            LivesSystem                 // Lives management
	hints            HintsSystem                 // Hints management
	difficulty       DifficultyProgression       // Adaptive difficulty
	personalBestStreak *int                      // Player's PersonalBest streak (nil if no previous record)
	usedHints        map[QuestionID][]HintType   // Track which hints were used on which questions

	// Domain events collected during operations
	events []Event
}

// NewMarathonGame creates a new marathon game session
// PersonalBest can be nil if the player has no previous record for this category
func NewMarathonGame(
	playerID UserID,
	category MarathonCategory,
	quizAggregate *quiz.Quiz,
	personalBest *PersonalBest,
	startedAt int64,
) (*MarathonGame, error) {
	// 1. Validate inputs
	if playerID.IsZero() {
		return nil, ErrInvalidGameID
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

	// 3. Extract PersonalBest streak if exists
	var personalBestStreak *int
	hasPersonalBest := false
	if personalBest != nil {
		streak := personalBest.BestStreak()
		personalBestStreak = &streak
		hasPersonalBest = true
	}

	// 4. Initialize lives and hints
	lives := NewLivesSystem(startedAt)
	hints := NewHintsSystem()
	difficulty := NewDifficultyProgression()

	// 5. Create game
	game := &MarathonGame{
		id:                 gameID,
		playerID:           playerID,
		category:           category,
		status:             GameStatusInProgress,
		session:            session,
		currentStreak:      0,
		maxStreak:          0,
		lives:              lives,
		hints:              hints,
		difficulty:         difficulty,
		personalBestStreak: personalBestStreak,
		usedHints:          make(map[QuestionID][]HintType),
		events:             make([]Event, 0),
	}

	// 6. Publish MarathonGameStarted event
	game.events = append(game.events, NewMarathonGameStartedEvent(
		gameID,
		playerID,
		category,
		hasPersonalBest,
		personalBestStreak,
		startedAt,
	))

	return game, nil
}

// AnswerQuestionResult holds detailed information about a submitted answer
type AnswerQuestionResult struct {
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
// - Correct answer: increment streak, update difficulty
// - Incorrect answer: lose life, reset streak, check game over
func (mg *MarathonGame) AnswerQuestion(
	questionID QuestionID,
	answerID AnswerID,
	timeTaken int64,
	answeredAt int64,
) (*AnswerQuestionResult, error) {
	// 1. Validate game state
	if mg.status != GameStatusInProgress {
		return nil, ErrGameNotActive
	}

	if !mg.lives.HasLives() {
		return nil, ErrNoLivesRemaining
	}

	// 2. Delegate to kernel session for pure gameplay logic
	kernelResult, err := mg.session.AnswerQuestion(questionID, answerID, timeTaken, answeredAt)
	if err != nil {
		return nil, err
	}

	// 3. Initialize result
	result := &AnswerQuestionResult{
		IsCorrect:       kernelResult.IsCorrect,
		BasePoints:      kernelResult.BasePoints.Value(),
		TimeTaken:       timeTaken,
		CurrentStreak:   mg.currentStreak,
		MaxStreak:       mg.maxStreak,
		DifficultyLevel: mg.difficulty.Level(),
		LifeLost:        false,
		RemainingLives:  mg.lives.CurrentLives(),
		IsGameOver:      false,
	}

	// 4. Process answer based on correctness
	if kernelResult.IsCorrect {
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

	// 5. Publish MarathonQuestionAnswered event
	mg.events = append(mg.events, NewMarathonQuestionAnsweredEvent(
		mg.id,
		mg.playerID,
		questionID,
		answerID,
		kernelResult.IsCorrect,
		timeTaken,
		mg.currentStreak,
		mg.difficulty.Level(),
		answeredAt,
	))

	return result, nil
}

// UseHint allows player to use a hint for the current question
func (mg *MarathonGame) UseHint(questionID QuestionID, hintType HintType, usedAt int64) error {
	// 1. Validate game state
	if mg.status != GameStatusInProgress {
		return ErrGameNotActive
	}

	// 2. Check if hint already used for this question
	if usedHints, exists := mg.usedHints[questionID]; exists {
		for _, usedHint := range usedHints {
			if usedHint == hintType {
				return ErrHintAlreadyUsed
			}
		}
	}

	// 3. Use hint (immutable operation)
	newHints, err := mg.hints.UseHint(hintType)
	if err != nil {
		return err
	}

	mg.hints = newHints

	// 4. Record hint usage
	mg.usedHints[questionID] = append(mg.usedHints[questionID], hintType)

	// 5. Get remaining hints of this type
	var remainingHints int
	switch hintType {
	case HintFiftyFifty:
		remainingHints = mg.hints.FiftyFifty()
	case HintExtraTime:
		remainingHints = mg.hints.ExtraTime()
	case HintSkip:
		remainingHints = mg.hints.Skip()
	}

	// 6. Publish HintUsed event
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
func (mg *MarathonGame) Abandon(abandonedAt int64) error {
	if mg.status != GameStatusInProgress {
		return ErrGameNotActive
	}

	// Validate state transition
	if !mg.status.CanTransitionTo(GameStatusAbandoned) {
		return ErrInvalidGameStatus
	}

	mg.status = GameStatusAbandoned

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
func (mg *MarathonGame) IsGameOver() bool {
	return mg.status.IsTerminal()
}

// GetCurrentQuestion returns the current question to be answered
func (mg *MarathonGame) GetCurrentQuestion() (*quiz.Question, error) {
	return mg.session.GetCurrentQuestion()
}

// IsNewPersonalBest checks if this game's maxStreak is a new personal best
func (mg *MarathonGame) IsNewPersonalBest() bool {
	if mg.personalBestStreak == nil {
		return mg.maxStreak > 0
	}
	return mg.maxStreak > *mg.personalBestStreak
}

// Getters
func (mg *MarathonGame) ID() GameID                     { return mg.id }
func (mg *MarathonGame) PlayerID() UserID               { return mg.playerID }
func (mg *MarathonGame) Category() MarathonCategory     { return mg.category }
func (mg *MarathonGame) Status() GameStatus             { return mg.status }
func (mg *MarathonGame) Session() *kernel.QuizGameplaySession { return mg.session }
func (mg *MarathonGame) CurrentStreak() int             { return mg.currentStreak }
func (mg *MarathonGame) MaxStreak() int                 { return mg.maxStreak }
func (mg *MarathonGame) Lives() LivesSystem             { return mg.lives }
func (mg *MarathonGame) Hints() HintsSystem             { return mg.hints }
func (mg *MarathonGame) Difficulty() DifficultyProgression { return mg.difficulty }
func (mg *MarathonGame) PersonalBestStreak() *int       { return mg.personalBestStreak }

// Events returns collected domain events and clears them
func (mg *MarathonGame) Events() []Event {
	events := mg.events
	mg.events = make([]Event, 0)
	return events
}

// ReconstructMarathonGame reconstructs a MarathonGame from persistence
// Used by repository when loading from database
func ReconstructMarathonGame(
	id GameID,
	playerID UserID,
	category MarathonCategory,
	status GameStatus,
	session *kernel.QuizGameplaySession,
	currentStreak int,
	maxStreak int,
	lives LivesSystem,
	hints HintsSystem,
	difficulty DifficultyProgression,
	personalBestStreak *int,
	usedHints map[QuestionID][]HintType,
) *MarathonGame {
	return &MarathonGame{
		id:                 id,
		playerID:           playerID,
		category:           category,
		status:             status,
		session:            session,
		currentStreak:      currentStreak,
		maxStreak:          maxStreak,
		lives:              lives,
		hints:              hints,
		difficulty:         difficulty,
		personalBestStreak: personalBestStreak,
		usedHints:          usedHints,
		events:             make([]Event, 0), // Don't replay events from DB
	}
}
