package solo_marathon

// Event is the base interface for all marathon domain events
type Event interface {
	EventType() string
	OccurredAt() int64
}

// MarathonGameStartedEvent fired when a marathon game starts
type MarathonGameStartedEvent struct {
	gameID     GameID
	playerID   UserID
	category   MarathonCategory
	hasRecord  bool // true if player has existing PersonalBest
	recordScore *int // player's previous best score (nil if no record)
	occurredAt int64
}

func NewMarathonGameStartedEvent(
	gameID GameID,
	playerID UserID,
	category MarathonCategory,
	hasRecord bool,
	recordScore *int,
	occurredAt int64,
) MarathonGameStartedEvent {
	return MarathonGameStartedEvent{
		gameID:      gameID,
		playerID:    playerID,
		category:    category,
		hasRecord:   hasRecord,
		recordScore: recordScore,
		occurredAt:  occurredAt,
	}
}

func (e MarathonGameStartedEvent) EventType() string           { return "marathon_game_started" }
func (e MarathonGameStartedEvent) OccurredAt() int64           { return e.occurredAt }
func (e MarathonGameStartedEvent) GameID() GameID              { return e.gameID }
func (e MarathonGameStartedEvent) PlayerID() UserID            { return e.playerID }
func (e MarathonGameStartedEvent) Category() MarathonCategory  { return e.category }
func (e MarathonGameStartedEvent) HasRecord() bool             { return e.hasRecord }
func (e MarathonGameStartedEvent) RecordScore() *int           { return e.recordScore }

// MarathonQuestionAnsweredEvent fired when player answers a question
type MarathonQuestionAnsweredEvent struct {
	gameID          GameID
	playerID        UserID
	questionID      QuestionID
	answerID        AnswerID
	isCorrect       bool
	timeTaken       int64 // milliseconds
	shieldActive    bool
	shieldConsumed  bool
	currentScore    int
	livesRemaining  int
	difficultyLevel DifficultyLevel
	occurredAt      int64
}

func NewMarathonQuestionAnsweredEvent(
	gameID GameID,
	playerID UserID,
	questionID QuestionID,
	answerID AnswerID,
	isCorrect bool,
	timeTaken int64,
	shieldActive bool,
	shieldConsumed bool,
	currentScore int,
	livesRemaining int,
	difficultyLevel DifficultyLevel,
	occurredAt int64,
) MarathonQuestionAnsweredEvent {
	return MarathonQuestionAnsweredEvent{
		gameID:          gameID,
		playerID:        playerID,
		questionID:      questionID,
		answerID:        answerID,
		isCorrect:       isCorrect,
		timeTaken:       timeTaken,
		shieldActive:    shieldActive,
		shieldConsumed:  shieldConsumed,
		currentScore:    currentScore,
		livesRemaining:  livesRemaining,
		difficultyLevel: difficultyLevel,
		occurredAt:      occurredAt,
	}
}

func (e MarathonQuestionAnsweredEvent) EventType() string            { return "marathon_question_answered" }
func (e MarathonQuestionAnsweredEvent) OccurredAt() int64            { return e.occurredAt }
func (e MarathonQuestionAnsweredEvent) GameID() GameID               { return e.gameID }
func (e MarathonQuestionAnsweredEvent) PlayerID() UserID             { return e.playerID }
func (e MarathonQuestionAnsweredEvent) QuestionID() QuestionID       { return e.questionID }
func (e MarathonQuestionAnsweredEvent) AnswerID() AnswerID           { return e.answerID }
func (e MarathonQuestionAnsweredEvent) IsCorrect() bool              { return e.isCorrect }
func (e MarathonQuestionAnsweredEvent) TimeTaken() int64             { return e.timeTaken }
func (e MarathonQuestionAnsweredEvent) ShieldActive() bool           { return e.shieldActive }
func (e MarathonQuestionAnsweredEvent) ShieldConsumed() bool         { return e.shieldConsumed }
func (e MarathonQuestionAnsweredEvent) CurrentScore() int            { return e.currentScore }
func (e MarathonQuestionAnsweredEvent) LivesRemaining() int          { return e.livesRemaining }
func (e MarathonQuestionAnsweredEvent) DifficultyLevel() DifficultyLevel {
	return e.difficultyLevel
}

// BonusUsedEvent fired when player uses a bonus
type BonusUsedEvent struct {
	gameID          GameID
	playerID        UserID
	questionID      QuestionID
	bonusType       BonusType
	remainingCount  int // Remaining bonuses of this type
	occurredAt      int64
}

func NewBonusUsedEvent(
	gameID GameID,
	playerID UserID,
	questionID QuestionID,
	bonusType BonusType,
	remainingCount int,
	occurredAt int64,
) BonusUsedEvent {
	return BonusUsedEvent{
		gameID:         gameID,
		playerID:       playerID,
		questionID:     questionID,
		bonusType:      bonusType,
		remainingCount: remainingCount,
		occurredAt:     occurredAt,
	}
}

func (e BonusUsedEvent) EventType() string      { return "bonus_used" }
func (e BonusUsedEvent) OccurredAt() int64      { return e.occurredAt }
func (e BonusUsedEvent) GameID() GameID         { return e.gameID }
func (e BonusUsedEvent) PlayerID() UserID       { return e.playerID }
func (e BonusUsedEvent) QuestionID() QuestionID { return e.questionID }
func (e BonusUsedEvent) BonusType() BonusType   { return e.bonusType }
func (e BonusUsedEvent) RemainingCount() int    { return e.remainingCount }

// LifeLostEvent fired when player loses a life
type LifeLostEvent struct {
	gameID         GameID
	playerID       UserID
	questionID     QuestionID
	remainingLives int
	occurredAt     int64
}

func NewLifeLostEvent(
	gameID GameID,
	playerID UserID,
	questionID QuestionID,
	remainingLives int,
	occurredAt int64,
) LifeLostEvent {
	return LifeLostEvent{
		gameID:         gameID,
		playerID:       playerID,
		questionID:     questionID,
		remainingLives: remainingLives,
		occurredAt:     occurredAt,
	}
}

func (e LifeLostEvent) EventType() string      { return "life_lost" }
func (e LifeLostEvent) OccurredAt() int64      { return e.occurredAt }
func (e LifeLostEvent) GameID() GameID         { return e.gameID }
func (e LifeLostEvent) PlayerID() UserID       { return e.playerID }
func (e LifeLostEvent) QuestionID() QuestionID { return e.questionID }
func (e LifeLostEvent) RemainingLives() int    { return e.remainingLives }

// MarathonGameOverEvent fired when game ends (no lives, declined continue, or player quit)
type MarathonGameOverEvent struct {
	gameID         GameID
	playerID       UserID
	finalScore     int
	totalQuestions int
	isNewRecord    bool
	previousRecord *int
	continueCount  int
	occurredAt     int64
}

func NewMarathonGameOverEvent(
	gameID GameID,
	playerID UserID,
	finalScore int,
	totalQuestions int,
	isNewRecord bool,
	previousRecord *int,
	continueCount int,
	occurredAt int64,
) MarathonGameOverEvent {
	return MarathonGameOverEvent{
		gameID:         gameID,
		playerID:       playerID,
		finalScore:     finalScore,
		totalQuestions: totalQuestions,
		isNewRecord:    isNewRecord,
		previousRecord: previousRecord,
		continueCount:  continueCount,
		occurredAt:     occurredAt,
	}
}

func (e MarathonGameOverEvent) EventType() string    { return "marathon_game_over" }
func (e MarathonGameOverEvent) OccurredAt() int64    { return e.occurredAt }
func (e MarathonGameOverEvent) GameID() GameID       { return e.gameID }
func (e MarathonGameOverEvent) PlayerID() UserID     { return e.playerID }
func (e MarathonGameOverEvent) FinalScore() int      { return e.finalScore }
func (e MarathonGameOverEvent) TotalQuestions() int   { return e.totalQuestions }
func (e MarathonGameOverEvent) IsNewRecord() bool    { return e.isNewRecord }
func (e MarathonGameOverEvent) PreviousRecord() *int { return e.previousRecord }
func (e MarathonGameOverEvent) ContinueCount() int   { return e.continueCount }

// ContinueUsedEvent fired when player uses continue to resume game
type ContinueUsedEvent struct {
	gameID        GameID
	playerID      UserID
	continueCount int
	paymentMethod PaymentMethod
	costCoins     int
	occurredAt    int64
}

func NewContinueUsedEvent(
	gameID GameID,
	playerID UserID,
	continueCount int,
	paymentMethod PaymentMethod,
	costCoins int,
	occurredAt int64,
) ContinueUsedEvent {
	return ContinueUsedEvent{
		gameID:        gameID,
		playerID:      playerID,
		continueCount: continueCount,
		paymentMethod: paymentMethod,
		costCoins:     costCoins,
		occurredAt:    occurredAt,
	}
}

func (e ContinueUsedEvent) EventType() string         { return "continue_used" }
func (e ContinueUsedEvent) OccurredAt() int64         { return e.occurredAt }
func (e ContinueUsedEvent) GameID() GameID            { return e.gameID }
func (e ContinueUsedEvent) PlayerID() UserID          { return e.playerID }
func (e ContinueUsedEvent) ContinueCount() int        { return e.continueCount }
func (e ContinueUsedEvent) PaymentMethod() PaymentMethod { return e.paymentMethod }
func (e ContinueUsedEvent) CostCoins() int            { return e.costCoins }

// DifficultyIncreasedEvent fired when difficulty level increases
type DifficultyIncreasedEvent struct {
	gameID        GameID
	playerID      UserID
	newLevel      DifficultyLevel
	questionIndex int
	occurredAt    int64
}

func NewDifficultyIncreasedEvent(
	gameID GameID,
	playerID UserID,
	newLevel DifficultyLevel,
	questionIndex int,
	occurredAt int64,
) DifficultyIncreasedEvent {
	return DifficultyIncreasedEvent{
		gameID:        gameID,
		playerID:      playerID,
		newLevel:      newLevel,
		questionIndex: questionIndex,
		occurredAt:    occurredAt,
	}
}

func (e DifficultyIncreasedEvent) EventType() string           { return "difficulty_increased" }
func (e DifficultyIncreasedEvent) OccurredAt() int64           { return e.occurredAt }
func (e DifficultyIncreasedEvent) GameID() GameID              { return e.gameID }
func (e DifficultyIncreasedEvent) PlayerID() UserID            { return e.playerID }
func (e DifficultyIncreasedEvent) NewLevel() DifficultyLevel   { return e.newLevel }
func (e DifficultyIncreasedEvent) QuestionIndex() int          { return e.questionIndex }
