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
	recordStreak *int // player's previous best streak (nil if no record)
	occurredAt int64
}

func NewMarathonGameStartedEvent(
	gameID GameID,
	playerID UserID,
	category MarathonCategory,
	hasRecord bool,
	recordStreak *int,
	occurredAt int64,
) MarathonGameStartedEvent {
	return MarathonGameStartedEvent{
		gameID:       gameID,
		playerID:     playerID,
		category:     category,
		hasRecord:    hasRecord,
		recordStreak: recordStreak,
		occurredAt:   occurredAt,
	}
}

func (e MarathonGameStartedEvent) EventType() string { return "marathon_game_started" }
func (e MarathonGameStartedEvent) OccurredAt() int64 { return e.occurredAt }
func (e MarathonGameStartedEvent) GameID() GameID    { return e.gameID }
func (e MarathonGameStartedEvent) PlayerID() UserID  { return e.playerID }
func (e MarathonGameStartedEvent) Category() MarathonCategory { return e.category }
func (e MarathonGameStartedEvent) HasRecord() bool            { return e.hasRecord }
func (e MarathonGameStartedEvent) RecordStreak() *int         { return e.recordStreak }

// MarathonQuestionAnsweredEvent fired when player answers a question
type MarathonQuestionAnsweredEvent struct {
	gameID       GameID
	playerID     UserID
	questionID   QuestionID
	answerID     AnswerID
	isCorrect    bool
	timeTaken    int64 // milliseconds
	currentStreak int
	difficultyLevel DifficultyLevel
	occurredAt   int64
}

func NewMarathonQuestionAnsweredEvent(
	gameID GameID,
	playerID UserID,
	questionID QuestionID,
	answerID AnswerID,
	isCorrect bool,
	timeTaken int64,
	currentStreak int,
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
		currentStreak:   currentStreak,
		difficultyLevel: difficultyLevel,
		occurredAt:      occurredAt,
	}
}

func (e MarathonQuestionAnsweredEvent) EventType() string       { return "marathon_question_answered" }
func (e MarathonQuestionAnsweredEvent) OccurredAt() int64       { return e.occurredAt }
func (e MarathonQuestionAnsweredEvent) GameID() GameID          { return e.gameID }
func (e MarathonQuestionAnsweredEvent) PlayerID() UserID        { return e.playerID }
func (e MarathonQuestionAnsweredEvent) QuestionID() QuestionID  { return e.questionID }
func (e MarathonQuestionAnsweredEvent) AnswerID() AnswerID      { return e.answerID }
func (e MarathonQuestionAnsweredEvent) IsCorrect() bool         { return e.isCorrect }
func (e MarathonQuestionAnsweredEvent) TimeTaken() int64        { return e.timeTaken }
func (e MarathonQuestionAnsweredEvent) CurrentStreak() int      { return e.currentStreak }
func (e MarathonQuestionAnsweredEvent) DifficultyLevel() DifficultyLevel {
	return e.difficultyLevel
}

// HintUsedEvent fired when player uses a hint
type HintUsedEvent struct {
	gameID      GameID
	playerID    UserID
	questionID  QuestionID
	hintType    HintType
	remainingHints int // Remaining hints of this type
	occurredAt  int64
}

func NewHintUsedEvent(
	gameID GameID,
	playerID UserID,
	questionID QuestionID,
	hintType HintType,
	remainingHints int,
	occurredAt int64,
) HintUsedEvent {
	return HintUsedEvent{
		gameID:         gameID,
		playerID:       playerID,
		questionID:     questionID,
		hintType:       hintType,
		remainingHints: remainingHints,
		occurredAt:     occurredAt,
	}
}

func (e HintUsedEvent) EventType() string      { return "hint_used" }
func (e HintUsedEvent) OccurredAt() int64      { return e.occurredAt }
func (e HintUsedEvent) GameID() GameID         { return e.gameID }
func (e HintUsedEvent) PlayerID() UserID       { return e.playerID }
func (e HintUsedEvent) QuestionID() QuestionID { return e.questionID }
func (e HintUsedEvent) HintType() HintType     { return e.hintType }
func (e HintUsedEvent) RemainingHints() int    { return e.remainingHints }

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

// MarathonGameOverEvent fired when game ends (no lives or player quit)
type MarathonGameOverEvent struct {
	gameID        GameID
	playerID      UserID
	finalStreak   int
	isNewRecord   bool // true if finalStreak > previous PersonalBest
	previousRecord *int // previous best streak (nil if no previous record)
	occurredAt    int64
}

func NewMarathonGameOverEvent(
	gameID GameID,
	playerID UserID,
	finalStreak int,
	isNewRecord bool,
	previousRecord *int,
	occurredAt int64,
) MarathonGameOverEvent {
	return MarathonGameOverEvent{
		gameID:         gameID,
		playerID:       playerID,
		finalStreak:    finalStreak,
		isNewRecord:    isNewRecord,
		previousRecord: previousRecord,
		occurredAt:     occurredAt,
	}
}

func (e MarathonGameOverEvent) EventType() string   { return "marathon_game_over" }
func (e MarathonGameOverEvent) OccurredAt() int64   { return e.occurredAt }
func (e MarathonGameOverEvent) GameID() GameID      { return e.gameID }
func (e MarathonGameOverEvent) PlayerID() UserID    { return e.playerID }
func (e MarathonGameOverEvent) FinalStreak() int    { return e.finalStreak }
func (e MarathonGameOverEvent) IsNewRecord() bool   { return e.isNewRecord }
func (e MarathonGameOverEvent) PreviousRecord() *int { return e.previousRecord }

// DifficultyIncreasedEvent fired when difficulty level increases
type DifficultyIncreasedEvent struct {
	gameID       GameID
	playerID     UserID
	newLevel     DifficultyLevel
	streakReached int
	occurredAt   int64
}

func NewDifficultyIncreasedEvent(
	gameID GameID,
	playerID UserID,
	newLevel DifficultyLevel,
	streakReached int,
	occurredAt int64,
) DifficultyIncreasedEvent {
	return DifficultyIncreasedEvent{
		gameID:        gameID,
		playerID:      playerID,
		newLevel:      newLevel,
		streakReached: streakReached,
		occurredAt:    occurredAt,
	}
}

func (e DifficultyIncreasedEvent) EventType() string           { return "difficulty_increased" }
func (e DifficultyIncreasedEvent) OccurredAt() int64           { return e.occurredAt }
func (e DifficultyIncreasedEvent) GameID() GameID              { return e.gameID }
func (e DifficultyIncreasedEvent) PlayerID() UserID            { return e.playerID }
func (e DifficultyIncreasedEvent) NewLevel() DifficultyLevel   { return e.newLevel }
func (e DifficultyIncreasedEvent) StreakReached() int          { return e.streakReached }
