package solo_marathon

import (
	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
)

// MarathonGame is the aggregate root for Solo Marathon mode
// Represents a single marathon session with lives, bonuses, and adaptive difficulty
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

	// Scoring — score = correct answers count (per docs)
	score          int // Total correct answers
	totalQuestions int // Total questions attempted (answered + skipped)

	// Marathon-specific mechanics
	lives          LivesSystem
	bonusInventory BonusInventory
	difficulty     DifficultyProgression
	shieldActive   bool // Shield currently activated for current question

	// Continue mechanic
	continueCount int // Times continued after game over

	// Personal best reference (for comparison)
	personalBestScore *int

	// Track bonus usage per question
	usedBonuses map[QuestionID][]BonusType

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
	bonuses BonusInventory,
	startedAt int64,
) (*MarathonGameV2, error) {
	// Validate inputs
	if playerID.IsZero() {
		return nil, ErrInvalidGameID
	}

	// Extract PersonalBest score if exists
	var personalBestScore *int
	hasPersonalBest := false
	if personalBest != nil {
		score := personalBest.BestStreak()
		personalBestScore = &score
		hasPersonalBest = true
	}

	// Initialize systems
	lives := NewLivesSystem(startedAt)
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
		score:               0,
		totalQuestions:       0,
		lives:               lives,
		bonusInventory:      bonuses,
		difficulty:          difficulty,
		shieldActive:        false,
		continueCount:       0,
		personalBestScore:   personalBestScore,
		usedBonuses:         make(map[QuestionID][]BonusType),
		events:              make([]Event, 0),
	}

	// Publish MarathonGameStarted event
	game.events = append(game.events, NewMarathonGameStartedEvent(
		game.id,
		playerID,
		category,
		hasPersonalBest,
		personalBestScore,
		startedAt,
	))

	return game, nil
}

// LoadNextQuestion loads the next question using QuestionSelector Domain Service
// This is called:
// 1. After game creation (to get first question)
// 2. After each correct answer (to get next question)
// 3. After skip bonus (to get next question)
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

	// Deactivate shield on new question (does NOT carry to next question)
	mg.shieldActive = false

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
	CorrectAnswerID AnswerID
	TimeTaken       int64
	Score           int             // Total correct answers
	TotalQuestions  int             // Total questions attempted
	DifficultyLevel DifficultyLevel
	LifeLost        bool
	ShieldConsumed  bool
	RemainingLives  int
	IsGameOver      bool
	// Filled when IsGameOver = true
	GameOverData *GameOverData
}

// GameOverData contains data for the game over screen
type GameOverData struct {
	FinalScore     int
	TotalQuestions int
	PersonalBest   int
	IsNewRecord    bool
	ContinueOffer *ContinueOffer
}

// ContinueOffer represents the continue options available
type ContinueOffer struct {
	Available     bool
	CostCoins     int
	HasAd         bool
	ContinueCount int
}

// AnswerQuestion processes a user's answer in marathon mode
// Business Logic:
// - Correct answer: increment score, update difficulty
// - Incorrect answer: if shield active, consume shield; else lose life
// - Game over when lives == 0 (intermediate state, continue offered)
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

	// 5. Find the correct answer ID for response
	var correctAnswerID AnswerID
	for _, a := range mg.currentQuestion.Answers() {
		if a.IsCorrect() {
			correctAnswerID = a.ID()
			break
		}
	}

	// 6. Track shield state for this answer
	wasShieldActive := mg.shieldActive
	shieldConsumed := false

	// 7. Initialize result
	result := &AnswerQuestionResultV2{
		IsCorrect:       isCorrect,
		CorrectAnswerID: correctAnswerID,
		TimeTaken:       timeTaken,
		Score:           mg.score,
		TotalQuestions:  mg.totalQuestions,
		DifficultyLevel: mg.difficulty.Level(),
		LifeLost:        false,
		ShieldConsumed:  false,
		RemainingLives:  mg.lives.CurrentLives(),
		IsGameOver:      false,
	}

	// 8. Process answer based on correctness
	if isCorrect {
		// === CORRECT ANSWER ===

		// a. Increment score (correct answers count)
		mg.score++
		result.Score = mg.score

		// b. Shield deactivates after question (NOT consumed on correct)
		mg.shieldActive = false

		// c. Update difficulty based on question index
		questionIndex := mg.totalQuestions + 1
		previousLevel := mg.difficulty.Level()
		mg.difficulty = mg.difficulty.UpdateFromQuestionIndex(questionIndex)

		// d. Fire DifficultyIncreased event if level changed
		if mg.difficulty.Level() != previousLevel {
			mg.events = append(mg.events, NewDifficultyIncreasedEvent(
				mg.id,
				mg.playerID,
				mg.difficulty.Level(),
				questionIndex,
				answeredAt,
			))
		}

		result.DifficultyLevel = mg.difficulty.Level()

	} else {
		// === INCORRECT ANSWER ===

		if wasShieldActive {
			// Shield protects from life loss
			shieldConsumed = true
			mg.shieldActive = false
			result.ShieldConsumed = true
			// Lives unchanged, game continues
		} else {
			// No shield — lose life
			mg.lives = mg.lives.LoseLife(answeredAt)
			result.LifeLost = true
			result.RemainingLives = mg.lives.CurrentLives()

			// Publish LifeLost event
			mg.events = append(mg.events, NewLifeLostEvent(
				mg.id,
				mg.playerID,
				questionID,
				mg.lives.CurrentLives(),
				answeredAt,
			))

			// Check if game over (no lives remaining)
			if !mg.lives.HasLives() {
				// Transition to game_over (intermediate state — continue offered)
				if !mg.status.CanTransitionTo(GameStatusGameOver) {
					return nil, ErrInvalidGameStatus
				}

				mg.status = GameStatusGameOver
				result.IsGameOver = true

				// Build continue offer
				costCalc := ContinueCostCalculator{}
				continueOffer := &ContinueOffer{
					Available:     true,
					CostCoins:     costCalc.GetCost(mg.continueCount),
					HasAd:         costCalc.HasAdOption(mg.continueCount),
					ContinueCount: mg.continueCount,
				}

				// Determine personal best for display
				personalBest := 0
				if mg.personalBestScore != nil {
					personalBest = *mg.personalBestScore
				}

				result.GameOverData = &GameOverData{
					FinalScore:     mg.score,
					TotalQuestions: mg.totalQuestions + 1,
					PersonalBest:   personalBest,
					IsNewRecord:    mg.IsNewPersonalBest(),
					ContinueOffer: continueOffer,
				}
			}
		}
	}

	// 9. Increment total questions count
	mg.totalQuestions++
	result.TotalQuestions = mg.totalQuestions

	// 10. Record question as answered
	mg.answeredQuestionIDs = append(mg.answeredQuestionIDs, questionID)

	// 11. Clear current question (will be loaded next time)
	mg.currentQuestion = nil

	// 12. Publish MarathonQuestionAnswered event
	mg.events = append(mg.events, NewMarathonQuestionAnsweredEvent(
		mg.id,
		mg.playerID,
		questionID,
		answerID,
		isCorrect,
		timeTaken,
		wasShieldActive,
		shieldConsumed,
		mg.score,
		mg.lives.CurrentLives(),
		mg.difficulty.Level(),
		answeredAt,
	))

	return result, nil
}

// ActivateShield activates shield for the current question
// Shield must be activated BEFORE answering
func (mg *MarathonGameV2) ActivateShield(questionID QuestionID, usedAt int64) error {
	if mg.status != GameStatusInProgress {
		return ErrGameNotActive
	}

	if mg.currentQuestion == nil {
		return ErrInvalidQuestion
	}

	if !mg.currentQuestion.ID().Equals(questionID) {
		return ErrInvalidQuestion
	}

	if mg.shieldActive {
		return ErrShieldAlreadyActive
	}

	// Check if shield bonus is available
	if !mg.bonusInventory.HasBonus(BonusShield) {
		return ErrNoBonusesAvailable
	}

	// Consume shield from inventory
	newInventory, err := mg.bonusInventory.UseBonus(BonusShield)
	if err != nil {
		return err
	}
	mg.bonusInventory = newInventory

	// Activate shield
	mg.shieldActive = true

	// Record bonus usage
	mg.usedBonuses[questionID] = append(mg.usedBonuses[questionID], BonusShield)

	// Publish BonusUsed event
	mg.events = append(mg.events, NewBonusUsedEvent(
		mg.id,
		mg.playerID,
		questionID,
		BonusShield,
		mg.bonusInventory.Shield(),
		usedAt,
	))

	return nil
}

// UseBonus allows player to use a bonus for the current question
// Shield is handled separately via ActivateShield()
// Skip immediately moves to next question (caller must call LoadNextQuestion)
func (mg *MarathonGameV2) UseBonus(questionID QuestionID, bonusType BonusType, usedAt int64) error {
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

	// Shield is handled via ActivateShield, not UseBonus
	if bonusType == BonusShield {
		return mg.ActivateShield(questionID, usedAt)
	}

	// Check if this specific bonus type already used for this question
	// (Shield + 50/50 is allowed, but not same bonus twice)
	if usedBonuses, exists := mg.usedBonuses[questionID]; exists {
		for _, usedBonus := range usedBonuses {
			if usedBonus == bonusType {
				return ErrBonusAlreadyUsed
			}
		}
	}

	// Use bonus (immutable operation)
	newInventory, err := mg.bonusInventory.UseBonus(bonusType)
	if err != nil {
		return err
	}

	mg.bonusInventory = newInventory

	// Record bonus usage
	mg.usedBonuses[questionID] = append(mg.usedBonuses[questionID], bonusType)

	// Get remaining count of this type
	remainingCount := mg.bonusInventory.Count(bonusType)

	// Handle skip: clear current question, deactivate shield
	if bonusType == BonusSkip {
		mg.totalQuestions++ // Skipped questions count toward total
		mg.answeredQuestionIDs = append(mg.answeredQuestionIDs, questionID)
		mg.currentQuestion = nil
		mg.shieldActive = false // Shield does NOT carry to next question
	}

	// Publish BonusUsed event
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

// Continue allows player to resume game after game over
// Resets lives to 1, increments continueCount, sets status back to in_progress
func (mg *MarathonGameV2) Continue(paymentMethod PaymentMethod, costCoins int, continuedAt int64) error {
	// Can only continue from game_over state
	if mg.status != GameStatusGameOver {
		return ErrContinueNotAvailable
	}

	// Validate state transition
	if !mg.status.CanTransitionTo(GameStatusInProgress) {
		return ErrInvalidGameStatus
	}

	// Reset lives to 1
	mg.lives = mg.lives.ResetForContinue(continuedAt)

	// Increment continue count
	mg.continueCount++

	// Set status back to in_progress
	mg.status = GameStatusInProgress

	// Publish ContinueUsed event
	mg.events = append(mg.events, NewContinueUsedEvent(
		mg.id,
		mg.playerID,
		mg.continueCount,
		paymentMethod,
		costCoins,
		continuedAt,
	))

	return nil
}

// CompleteGame finalizes a game that is in game_over state (player declined continue)
func (mg *MarathonGameV2) CompleteGame(completedAt int64) error {
	if mg.status != GameStatusGameOver {
		return ErrContinueNotAvailable
	}

	if !mg.status.CanTransitionTo(GameStatusCompleted) {
		return ErrInvalidGameStatus
	}

	mg.status = GameStatusCompleted
	mg.finishedAt = completedAt

	// Determine if new record
	isNewRecord := mg.IsNewPersonalBest()

	// Publish MarathonGameOver event
	mg.events = append(mg.events, NewMarathonGameOverEvent(
		mg.id,
		mg.playerID,
		mg.score,
		mg.totalQuestions,
		isNewRecord,
		mg.personalBestScore,
		mg.continueCount,
		completedAt,
	))

	return nil
}

// Abandon marks the game as abandoned (player quit voluntarily)
func (mg *MarathonGameV2) Abandon(abandonedAt int64) error {
	if mg.status != GameStatusInProgress && mg.status != GameStatusGameOver {
		return ErrGameNotActive
	}

	// Validate state transition
	targetStatus := GameStatusAbandoned
	if mg.status == GameStatusGameOver {
		// From game_over, abandoning is completing
		targetStatus = GameStatusCompleted
	}

	if !mg.status.CanTransitionTo(targetStatus) {
		return ErrInvalidGameStatus
	}

	mg.status = targetStatus
	mg.finishedAt = abandonedAt

	// Determine if new record
	isNewRecord := mg.IsNewPersonalBest()

	// Publish MarathonGameOver event
	mg.events = append(mg.events, NewMarathonGameOverEvent(
		mg.id,
		mg.playerID,
		mg.score,
		mg.totalQuestions,
		isNewRecord,
		mg.personalBestScore,
		mg.continueCount,
		abandonedAt,
	))

	return nil
}

// IsGameOver checks if game is finished or abandoned
func (mg *MarathonGameV2) IsGameOver() bool {
	return mg.status.IsTerminal()
}

// IsWaitingForContinue checks if game is in game_over state (waiting for continue decision)
func (mg *MarathonGameV2) IsWaitingForContinue() bool {
	return mg.status == GameStatusGameOver
}

// GetCurrentQuestion returns the current question to be answered
func (mg *MarathonGameV2) GetCurrentQuestion() (*quiz.Question, error) {
	if mg.currentQuestion == nil {
		return nil, ErrInvalidQuestion
	}
	return mg.currentQuestion, nil
}

// IsNewPersonalBest checks if this game's score is a new personal best
func (mg *MarathonGameV2) IsNewPersonalBest() bool {
	if mg.personalBestScore == nil {
		return mg.score > 0
	}
	return mg.score > *mg.personalBestScore
}

// QuestionNumber returns the 1-based index of the next question to answer
func (mg *MarathonGameV2) QuestionNumber() int {
	return mg.totalQuestions + 1
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
func (mg *MarathonGameV2) Score() int                                  { return mg.score }
func (mg *MarathonGameV2) TotalQuestions() int                         { return mg.totalQuestions }
func (mg *MarathonGameV2) Lives() LivesSystem                          { return mg.lives }
func (mg *MarathonGameV2) BonusInventory() BonusInventory              { return mg.bonusInventory }
func (mg *MarathonGameV2) Difficulty() DifficultyProgression           { return mg.difficulty }
func (mg *MarathonGameV2) ShieldActive() bool                          { return mg.shieldActive }
func (mg *MarathonGameV2) ContinueCount() int                         { return mg.continueCount }
func (mg *MarathonGameV2) PersonalBestScore() *int                     { return mg.personalBestScore }

// Backwards-compatible getters (transitional — some callers may still use streak/hints)
func (mg *MarathonGameV2) CurrentStreak() int    { return mg.score }
func (mg *MarathonGameV2) MaxStreak() int        { return mg.score }
func (mg *MarathonGameV2) BaseScore() int        { return mg.score }

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
	score int,
	totalQuestions int,
	lives LivesSystem,
	bonusInventory BonusInventory,
	difficulty DifficultyProgression,
	shieldActive bool,
	continueCount int,
	personalBestScore *int,
	usedBonuses map[QuestionID][]BonusType,
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
		score:               score,
		totalQuestions:       totalQuestions,
		lives:               lives,
		bonusInventory:      bonusInventory,
		difficulty:          difficulty,
		shieldActive:        shieldActive,
		continueCount:       continueCount,
		personalBestScore:   personalBestScore,
		usedBonuses:         usedBonuses,
		events:              make([]Event, 0), // Don't replay events from DB
	}
}
