package daily_challenge

// Event is the base interface for all daily challenge domain events
type Event interface {
	EventType() string
	OccurredAt() int64
}

// DailyQuizCreatedEvent fired when a new daily quiz is created (daily cron job)
type DailyQuizCreatedEvent struct {
	dailyQuizID  DailyQuizID
	date         Date
	questionIDs  []QuestionID
	expiresAt    int64
	occurredAt   int64
}

func NewDailyQuizCreatedEvent(
	dailyQuizID DailyQuizID,
	date Date,
	questionIDs []QuestionID,
	expiresAt int64,
	occurredAt int64,
) DailyQuizCreatedEvent {
	return DailyQuizCreatedEvent{
		dailyQuizID: dailyQuizID,
		date:        date,
		questionIDs: questionIDs,
		expiresAt:   expiresAt,
		occurredAt:  occurredAt,
	}
}

func (e DailyQuizCreatedEvent) EventType() string      { return "daily_quiz_created" }
func (e DailyQuizCreatedEvent) OccurredAt() int64      { return e.occurredAt }
func (e DailyQuizCreatedEvent) DailyQuizID() DailyQuizID { return e.dailyQuizID }
func (e DailyQuizCreatedEvent) Date() Date             { return e.date }
func (e DailyQuizCreatedEvent) QuestionIDs() []QuestionID { return e.questionIDs }
func (e DailyQuizCreatedEvent) ExpiresAt() int64       { return e.expiresAt }

// DailyGameStartedEvent fired when a player starts daily challenge
type DailyGameStartedEvent struct {
	gameID      GameID
	playerID    UserID
	dailyQuizID DailyQuizID
	date        Date
	currentStreak int // Streak before this game
	occurredAt  int64
}

func NewDailyGameStartedEvent(
	gameID GameID,
	playerID UserID,
	dailyQuizID DailyQuizID,
	date Date,
	currentStreak int,
	occurredAt int64,
) DailyGameStartedEvent {
	return DailyGameStartedEvent{
		gameID:        gameID,
		playerID:      playerID,
		dailyQuizID:   dailyQuizID,
		date:          date,
		currentStreak: currentStreak,
		occurredAt:    occurredAt,
	}
}

func (e DailyGameStartedEvent) EventType() string      { return "daily_game_started" }
func (e DailyGameStartedEvent) OccurredAt() int64      { return e.occurredAt }
func (e DailyGameStartedEvent) GameID() GameID         { return e.gameID }
func (e DailyGameStartedEvent) PlayerID() UserID       { return e.playerID }
func (e DailyGameStartedEvent) DailyQuizID() DailyQuizID { return e.dailyQuizID }
func (e DailyGameStartedEvent) Date() Date             { return e.date }
func (e DailyGameStartedEvent) CurrentStreak() int     { return e.currentStreak }

// DailyQuestionAnsweredEvent fired when player answers a question
type DailyQuestionAnsweredEvent struct {
	gameID     GameID
	playerID   UserID
	questionID QuestionID
	answerID   AnswerID
	timeTaken  int64 // milliseconds
	occurredAt int64
}

func NewDailyQuestionAnsweredEvent(
	gameID GameID,
	playerID UserID,
	questionID QuestionID,
	answerID AnswerID,
	timeTaken int64,
	occurredAt int64,
) DailyQuestionAnsweredEvent {
	return DailyQuestionAnsweredEvent{
		gameID:     gameID,
		playerID:   playerID,
		questionID: questionID,
		answerID:   answerID,
		timeTaken:  timeTaken,
		occurredAt: occurredAt,
	}
}

func (e DailyQuestionAnsweredEvent) EventType() string      { return "daily_question_answered" }
func (e DailyQuestionAnsweredEvent) OccurredAt() int64      { return e.occurredAt }
func (e DailyQuestionAnsweredEvent) GameID() GameID         { return e.gameID }
func (e DailyQuestionAnsweredEvent) PlayerID() UserID       { return e.playerID }
func (e DailyQuestionAnsweredEvent) QuestionID() QuestionID { return e.questionID }
func (e DailyQuestionAnsweredEvent) AnswerID() AnswerID     { return e.answerID }
func (e DailyQuestionAnsweredEvent) TimeTaken() int64       { return e.timeTaken }

// DailyGameCompletedEvent fired when player completes all 10 questions
type DailyGameCompletedEvent struct {
	gameID         GameID
	playerID       UserID
	dailyQuizID    DailyQuizID
	date           Date
	finalScore     int
	correctAnswers int
	totalQuestions int
	newStreak      int   // Streak after completing
	streakBonus    float64 // Streak multiplier applied
	rank           *int  // Player's rank (nil if not yet calculated)
	occurredAt     int64
}

func NewDailyGameCompletedEvent(
	gameID GameID,
	playerID UserID,
	dailyQuizID DailyQuizID,
	date Date,
	finalScore int,
	correctAnswers int,
	totalQuestions int,
	newStreak int,
	streakBonus float64,
	rank *int,
	occurredAt int64,
) DailyGameCompletedEvent {
	return DailyGameCompletedEvent{
		gameID:         gameID,
		playerID:       playerID,
		dailyQuizID:    dailyQuizID,
		date:           date,
		finalScore:     finalScore,
		correctAnswers: correctAnswers,
		totalQuestions: totalQuestions,
		newStreak:      newStreak,
		streakBonus:    streakBonus,
		rank:           rank,
		occurredAt:     occurredAt,
	}
}

func (e DailyGameCompletedEvent) EventType() string      { return "daily_game_completed" }
func (e DailyGameCompletedEvent) OccurredAt() int64      { return e.occurredAt }
func (e DailyGameCompletedEvent) GameID() GameID         { return e.gameID }
func (e DailyGameCompletedEvent) PlayerID() UserID       { return e.playerID }
func (e DailyGameCompletedEvent) DailyQuizID() DailyQuizID { return e.dailyQuizID }
func (e DailyGameCompletedEvent) Date() Date             { return e.date }
func (e DailyGameCompletedEvent) FinalScore() int        { return e.finalScore }
func (e DailyGameCompletedEvent) CorrectAnswers() int    { return e.correctAnswers }
func (e DailyGameCompletedEvent) TotalQuestions() int    { return e.totalQuestions }
func (e DailyGameCompletedEvent) NewStreak() int         { return e.newStreak }
func (e DailyGameCompletedEvent) StreakBonus() float64   { return e.streakBonus }
func (e DailyGameCompletedEvent) Rank() *int             { return e.rank }

// StreakMilestoneReachedEvent fired when player reaches streak milestone (3, 7, 14, 30, 100 days)
type StreakMilestoneReachedEvent struct {
	gameID       GameID
	playerID     UserID
	streakDays   int
	bonusPercent int // 10, 25, 40, 60, 100
	occurredAt   int64
}

func NewStreakMilestoneReachedEvent(
	gameID GameID,
	playerID UserID,
	streakDays int,
	bonusPercent int,
	occurredAt int64,
) StreakMilestoneReachedEvent {
	return StreakMilestoneReachedEvent{
		gameID:       gameID,
		playerID:     playerID,
		streakDays:   streakDays,
		bonusPercent: bonusPercent,
		occurredAt:   occurredAt,
	}
}

func (e StreakMilestoneReachedEvent) EventType() string { return "streak_milestone_reached" }
func (e StreakMilestoneReachedEvent) OccurredAt() int64 { return e.occurredAt }
func (e StreakMilestoneReachedEvent) GameID() GameID    { return e.gameID }
func (e StreakMilestoneReachedEvent) PlayerID() UserID  { return e.playerID }
func (e StreakMilestoneReachedEvent) StreakDays() int   { return e.streakDays }
func (e StreakMilestoneReachedEvent) BonusPercent() int { return e.bonusPercent }
