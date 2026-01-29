package daily_challenge

// DailyQuiz is the aggregate root for daily quiz content
// One DailyQuiz is created per day (same questions for all players)
type DailyQuiz struct {
	id          DailyQuizID
	date        Date              // 2026-01-25
	questionIDs []QuestionID      // 10 question IDs
	expiresAt   int64             // Unix timestamp when quiz expires (next day 00:00 UTC)
	createdAt   int64             // Unix timestamp when created

	// Domain events collected during operations
	events []Event
}

const (
	QuestionsPerDay = 10 // Fixed: 10 questions per daily challenge
)

// NewDailyQuiz creates a new daily quiz for a specific date
func NewDailyQuiz(
	date Date,
	questionIDs []QuestionID,
	expiresAt int64,
	createdAt int64,
) (*DailyQuiz, error) {
	// Validate
	if date.IsZero() {
		return nil, ErrInvalidDate
	}

	if len(questionIDs) != QuestionsPerDay {
		return nil, ErrInvalidDate // Need exactly 10 questions
	}

	// Create
	dailyQuiz := &DailyQuiz{
		id:          NewDailyQuizID(),
		date:        date,
		questionIDs: questionIDs,
		expiresAt:   expiresAt,
		createdAt:   createdAt,
		events:      make([]Event, 0),
	}

	// Publish DailyQuizCreated event
	dailyQuiz.events = append(dailyQuiz.events, NewDailyQuizCreatedEvent(
		dailyQuiz.id,
		date,
		questionIDs,
		expiresAt,
		createdAt,
	))

	return dailyQuiz, nil
}

// IsExpired checks if quiz has expired
func (dq *DailyQuiz) IsExpired(now int64) bool {
	return now >= dq.expiresAt
}

// ContainsQuestion checks if question is in this daily quiz
func (dq *DailyQuiz) ContainsQuestion(questionID QuestionID) bool {
	for _, qid := range dq.questionIDs {
		if qid.Equals(questionID) {
			return true
		}
	}
	return false
}

// GetQuestionByIndex returns question ID by index (0-9)
func (dq *DailyQuiz) GetQuestionByIndex(index int) (QuestionID, error) {
	if index < 0 || index >= len(dq.questionIDs) {
		return QuestionID{}, ErrQuestionNotInQuiz
	}
	return dq.questionIDs[index], nil
}

// Getters
func (dq *DailyQuiz) ID() DailyQuizID           { return dq.id }
func (dq *DailyQuiz) Date() Date                { return dq.date }
func (dq *DailyQuiz) QuestionIDs() []QuestionID {
	// Return copy to protect internal state
	copy := make([]QuestionID, len(dq.questionIDs))
	for i, qid := range dq.questionIDs {
		copy[i] = qid
	}
	return copy
}
func (dq *DailyQuiz) ExpiresAt() int64   { return dq.expiresAt }
func (dq *DailyQuiz) CreatedAt() int64   { return dq.createdAt }
func (dq *DailyQuiz) QuestionCount() int { return len(dq.questionIDs) }

// Events returns collected domain events and clears them
func (dq *DailyQuiz) Events() []Event {
	events := dq.events
	dq.events = make([]Event, 0)
	return events
}

// ReconstructDailyQuiz reconstructs a DailyQuiz from persistence
// Used by repository when loading from database
func ReconstructDailyQuiz(
	id DailyQuizID,
	date Date,
	questionIDs []QuestionID,
	expiresAt int64,
	createdAt int64,
) *DailyQuiz {
	return &DailyQuiz{
		id:          id,
		date:        date,
		questionIDs: questionIDs,
		expiresAt:   expiresAt,
		createdAt:   createdAt,
		events:      make([]Event, 0), // Don't replay events from DB
	}
}
