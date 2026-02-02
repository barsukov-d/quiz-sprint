package solo_marathon

import (
	"github.com/barsukov/quiz-sprint/backend/internal/domain/kernel"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
)

// MarathonGame is the V1 aggregate root for Solo Marathon mode (uses kernel session)
// DEPRECATED: Use MarathonGameV2 for new features (V2 uses dynamic question loading)
type MarathonGame struct {
	id               GameID
	playerID         UserID
	category         MarathonCategory
	status           GameStatus
	session          *kernel.QuizGameplaySession
	currentStreak    int
	maxStreak        int
	lives            LivesSystem
	bonusInventory   BonusInventory
	difficulty       DifficultyProgression
	personalBestStreak *int
	usedBonuses      map[QuestionID][]BonusType

	events []Event
}

// NewMarathonGame creates a new marathon game session (V1 — uses kernel session)
func NewMarathonGame(
	playerID UserID,
	category MarathonCategory,
	quizAggregate *quiz.Quiz,
	personalBest *PersonalBest,
	startedAt int64,
) (*MarathonGame, error) {
	if playerID.IsZero() {
		return nil, ErrInvalidGameID
	}

	if quizAggregate == nil {
		return nil, quiz.ErrQuizNotFound
	}

	if err := quizAggregate.CanStart(); err != nil {
		return nil, err
	}

	gameID := NewGameID()
	sessionID := kernel.NewSessionID()

	session, err := kernel.NewQuizGameplaySession(sessionID, quizAggregate, startedAt)
	if err != nil {
		return nil, err
	}

	var personalBestStreak *int
	hasPersonalBest := false
	if personalBest != nil {
		streak := personalBest.BestStreak()
		personalBestStreak = &streak
		hasPersonalBest = true
	}

	lives := NewLivesSystem(startedAt)
	bonuses := NewBonusInventory()
	difficulty := NewDifficultyProgression()

	game := &MarathonGame{
		id:                 gameID,
		playerID:           playerID,
		category:           category,
		status:             GameStatusInProgress,
		session:            session,
		currentStreak:      0,
		maxStreak:          0,
		lives:              lives,
		bonusInventory:     bonuses,
		difficulty:         difficulty,
		personalBestStreak: personalBestStreak,
		usedBonuses:        make(map[QuestionID][]BonusType),
		events:             make([]Event, 0),
	}

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

// AnswerQuestionResult holds detailed information about a submitted answer (V1)
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

// AnswerQuestion processes a user's answer in marathon mode (V1)
func (mg *MarathonGame) AnswerQuestion(
	questionID QuestionID,
	answerID AnswerID,
	timeTaken int64,
	answeredAt int64,
) (*AnswerQuestionResult, error) {
	if mg.status != GameStatusInProgress {
		return nil, ErrGameNotActive
	}

	if !mg.lives.HasLives() {
		return nil, ErrNoLivesRemaining
	}

	kernelResult, err := mg.session.AnswerQuestion(questionID, answerID, timeTaken, answeredAt)
	if err != nil {
		return nil, err
	}

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

	if kernelResult.IsCorrect {
		mg.currentStreak++
		result.CurrentStreak = mg.currentStreak

		if mg.currentStreak > mg.maxStreak {
			mg.maxStreak = mg.currentStreak
		}
		result.MaxStreak = mg.maxStreak

		previousLevel := mg.difficulty.Level()
		mg.difficulty = mg.difficulty.UpdateFromQuestionIndex(mg.currentStreak)

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
		mg.lives = mg.lives.LoseLife(answeredAt)
		result.LifeLost = true
		result.RemainingLives = mg.lives.CurrentLives()

		mg.events = append(mg.events, NewLifeLostEvent(
			mg.id,
			mg.playerID,
			questionID,
			mg.lives.CurrentLives(),
			answeredAt,
		))

		mg.currentStreak = 0
		result.CurrentStreak = 0

		mg.difficulty = NewDifficultyProgression()
		result.DifficultyLevel = mg.difficulty.Level()

		if !mg.lives.HasLives() {
			if !mg.status.CanTransitionTo(GameStatusGameOver) {
				return nil, ErrInvalidGameStatus
			}

			mg.status = GameStatusGameOver
			result.IsGameOver = true

			isNewRecord := false
			if mg.personalBestStreak == nil || mg.maxStreak > *mg.personalBestStreak {
				isNewRecord = true
			}

			mg.events = append(mg.events, NewMarathonGameOverEvent(
				mg.id,
				mg.playerID,
				mg.maxStreak,
				0, // totalQuestions not tracked in V1
				isNewRecord,
				mg.personalBestStreak,
				0, // continueCount not supported in V1
				answeredAt,
			))
		}
	}

	mg.events = append(mg.events, NewMarathonQuestionAnsweredEvent(
		mg.id,
		mg.playerID,
		questionID,
		answerID,
		kernelResult.IsCorrect,
		timeTaken,
		false, // shieldActive not supported in V1
		false, // shieldConsumed not supported in V1
		mg.maxStreak,
		mg.lives.CurrentLives(),
		mg.difficulty.Level(),
		answeredAt,
	))

	return result, nil
}

// UseBonus allows player to use a bonus for the current question (V1)
func (mg *MarathonGame) UseBonus(questionID QuestionID, bonusType BonusType, usedAt int64) error {
	if mg.status != GameStatusInProgress {
		return ErrGameNotActive
	}

	if usedBonuses, exists := mg.usedBonuses[questionID]; exists {
		for _, usedBonus := range usedBonuses {
			if usedBonus == bonusType {
				return ErrBonusAlreadyUsed
			}
		}
	}

	newInventory, err := mg.bonusInventory.UseBonus(bonusType)
	if err != nil {
		return err
	}

	mg.bonusInventory = newInventory
	mg.usedBonuses[questionID] = append(mg.usedBonuses[questionID], bonusType)

	remainingCount := mg.bonusInventory.Count(bonusType)

	mg.events = append(mg.events, NewBonusUsedEvent(
		mg.id,
		mg.playerID,
		questionID,
		bonusType,
		remainingCount,
		usedAt,
	))

	return nil
}

// Abandon marks the game as abandoned (player quit voluntarily)
func (mg *MarathonGame) Abandon(abandonedAt int64) error {
	if mg.status != GameStatusInProgress {
		return ErrGameNotActive
	}

	if !mg.status.CanTransitionTo(GameStatusAbandoned) {
		return ErrInvalidGameStatus
	}

	mg.status = GameStatusAbandoned

	isNewRecord := false
	if mg.personalBestStreak == nil || mg.maxStreak > *mg.personalBestStreak {
		isNewRecord = true
	}

	mg.events = append(mg.events, NewMarathonGameOverEvent(
		mg.id,
		mg.playerID,
		mg.maxStreak,
		0, // totalQuestions not tracked in V1
		isNewRecord,
		mg.personalBestStreak,
		0, // continueCount not supported in V1
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
func (mg *MarathonGame) ID() GameID                            { return mg.id }
func (mg *MarathonGame) PlayerID() UserID                      { return mg.playerID }
func (mg *MarathonGame) Category() MarathonCategory            { return mg.category }
func (mg *MarathonGame) Status() GameStatus                    { return mg.status }
func (mg *MarathonGame) Session() *kernel.QuizGameplaySession  { return mg.session }
func (mg *MarathonGame) CurrentStreak() int                    { return mg.currentStreak }
func (mg *MarathonGame) MaxStreak() int                        { return mg.maxStreak }
func (mg *MarathonGame) Lives() LivesSystem                    { return mg.lives }
func (mg *MarathonGame) Hints() BonusInventory                 { return mg.bonusInventory }
func (mg *MarathonGame) Difficulty() DifficultyProgression     { return mg.difficulty }
func (mg *MarathonGame) PersonalBestStreak() *int              { return mg.personalBestStreak }

// Events returns collected domain events and clears them
func (mg *MarathonGame) Events() []Event {
	events := mg.events
	mg.events = make([]Event, 0)
	return events
}

// ReconstructMarathonGame reconstructs a MarathonGame from persistence
func ReconstructMarathonGame(
	id GameID,
	playerID UserID,
	category MarathonCategory,
	status GameStatus,
	session *kernel.QuizGameplaySession,
	currentStreak int,
	maxStreak int,
	lives LivesSystem,
	bonuses BonusInventory,
	difficulty DifficultyProgression,
	personalBestStreak *int,
	usedBonuses map[QuestionID][]BonusType,
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
		bonusInventory:     bonuses,
		difficulty:         difficulty,
		personalBestStreak: personalBestStreak,
		usedBonuses:        usedBonuses,
		events:             make([]Event, 0),
	}
}
