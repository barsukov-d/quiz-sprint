package quiz

import (
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
)

// SessionStatus represents the status of a quiz session
type SessionStatus int

const (
	SessionStatusActive SessionStatus = iota
	SessionStatusCompleted
	SessionStatusAbandoned
)

func (s SessionStatus) String() string {
	switch s {
	case SessionStatusActive:
		return "active"
	case SessionStatusCompleted:
		return "completed"
	case SessionStatusAbandoned:
		return "abandoned"
	default:
		return "unknown"
	}
}

// Quiz is the aggregate root for quiz-related entities
type Quiz struct {
	id           QuizID
	title        QuizTitle
	description  string
	categoryID   CategoryID
	questions    []Question
	timeLimit    TimeLimit
	passingScore PassingScore
	createdAt    int64 // Unix timestamp
	updatedAt    int64 // Unix timestamp
	// questionCount is a denormalized count, used when questions are not loaded
	questionCount int

	// Domain events collected during operations
	events []Event
}

// NewQuiz creates a new Quiz aggregate
func NewQuiz(id QuizID, title QuizTitle, description string, categoryID CategoryID, timeLimit TimeLimit, passingScore PassingScore, createdAt int64) (*Quiz, error) {
	if id.IsZero() {
		return nil, ErrInvalidQuizID
	}

	if title.IsEmpty() {
		return nil, ErrInvalidTitle
	}

	return &Quiz{
		id:            id,
		title:         title,
		description:   description,
		categoryID:    categoryID,
		timeLimit:     timeLimit,
		passingScore:  passingScore,
		createdAt:     createdAt,
		updatedAt:     createdAt,
		questions:     make([]Question, 0),
		questionCount: 0,
		events:        make([]Event, 0),
	}, nil
}

// ReconstructQuiz reconstructs a Quiz from persistence (no validation)
// Used by repository when loading from database
func ReconstructQuiz(
	id QuizID,
	title QuizTitle,
	description string,
	categoryID CategoryID,
	timeLimit TimeLimit,
	passingScore PassingScore,
	createdAt int64,
	updatedAt int64,
) *Quiz {
	return &Quiz{
		id:            id,
		title:         title,
		description:   description,
		categoryID:    categoryID,
		timeLimit:     timeLimit,
		passingScore:  passingScore,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
		questions:     make([]Question, 0),
		questionCount: 0,
		events:        make([]Event, 0),
	}
}

// AddQuestion adds a question to the quiz
func (q *Quiz) AddQuestion(question Question) error {
	if len(q.questions) >= 50 {
		return ErrTooManyQuestions
	}

	if !question.HasCorrectAnswer() {
		return ErrInvalidAnswer
	}

	q.questions = append(q.questions, question)
	return nil
}

// CanStart checks if the quiz can be started (business rule)
func (q *Quiz) CanStart() error {
	if len(q.questions) == 0 {
		return ErrNoQuestions
	}

	if q.timeLimit.IsZero() {
		return ErrInvalidTimeLimit
	}

	return nil
}

// HasMinimumQuestions checks if quiz has minimum required questions
func (q *Quiz) HasMinimumQuestions() bool {
	return len(q.questions) >= 5
}

// GetTotalPoints calculates total points available in the quiz
func (q *Quiz) GetTotalPoints() Points {
	total := Points{}
	for _, question := range q.questions {
		total = total.Add(question.Points())
	}
	return total
}

// GetQuestion returns a question by ID
func (q *Quiz) GetQuestion(questionID QuestionID) (*Question, error) {
	for i := range q.questions {
		if q.questions[i].ID().Equals(questionID) {
			return &q.questions[i], nil
		}
	}
	return nil, ErrQuestionNotFound
}

// GetQuestionByIndex returns question by position (0-indexed)
func (q *Quiz) GetQuestionByIndex(index int) (*Question, error) {
	if index < 0 || index >= len(q.questions) {
		return nil, ErrQuestionNotFound
	}
	return &q.questions[index], nil
}

// QuestionsCount returns the number of questions
func (q *Quiz) QuestionsCount() int {
	if len(q.questions) > 0 {
		return len(q.questions)
	}
	return q.questionCount
}

// SetQuestionCount sets the denormalized question count
func (q *Quiz) SetQuestionCount(count int) {
	q.questionCount = count
}

// Getters
func (q *Quiz) ID() QuizID                 { return q.id }
func (q *Quiz) Title() QuizTitle           { return q.title }
func (q *Quiz) Description() string        { return q.description }
func (q *Quiz) CategoryID() CategoryID     { return q.categoryID }
func (q *Quiz) TimeLimit() TimeLimit       { return q.timeLimit }
func (q *Quiz) PassingScore() PassingScore { return q.passingScore }
func (q *Quiz) CreatedAt() int64           { return q.createdAt }
func (q *Quiz) UpdatedAt() int64           { return q.updatedAt }

// Questions returns a copy of questions (protect internal state)
func (q *Quiz) Questions() []Question {
	copies := make([]Question, len(q.questions))
	copy(copies, q.questions)
	return copies
}

// Events returns collected domain events and clears them
func (q *Quiz) Events() []Event {
	events := q.events
	q.events = make([]Event, 0)
	return events
}

// QuizSession is the aggregate root for an active quiz session
type QuizSession struct {
	id              SessionID
	quizID          QuizID
	userID          shared.UserID
	currentQuestion int
	score           Points
	answers         []UserAnswer
	startedAt       int64 // Unix timestamp
	completedAt     int64 // Unix timestamp (0 if not completed)
	status          SessionStatus

	// Domain events collected during operations
	events []Event
}

// NewQuizSession creates a new QuizSession aggregate
func NewQuizSession(id SessionID, quizID QuizID, userID shared.UserID, startedAt int64) (*QuizSession, error) {
	if id.IsZero() {
		return nil, ErrInvalidSessionID
	}

	if quizID.IsZero() {
		return nil, ErrInvalidQuizID
	}

	if userID.IsZero() {
		return nil, shared.ErrInvalidUserID
	}

	session := &QuizSession{
		id:              id,
		quizID:          quizID,
		userID:          userID,
		currentQuestion: 0,
		score:           Points{},
		answers:         make([]UserAnswer, 0),
		startedAt:       startedAt,
		completedAt:     0,
		status:          SessionStatusActive,
		events:          make([]Event, 0),
	}

	// Record domain event
	session.events = append(session.events, NewQuizStartedEvent(quizID, id, userID, startedAt))

	return session, nil
}

// SubmitAnswer processes a user's answer (business logic)
func (qs *QuizSession) SubmitAnswer(question *Question, answerID AnswerID, answeredAt int64) error {
	if qs.status != SessionStatusActive {
		return ErrSessionCompleted
	}

	// Check if already answered this question
	for _, ua := range qs.answers {
		if ua.QuestionID().Equals(question.ID()) {
			return ErrAlreadyAnswered
		}
	}

	// Validate answer belongs to question
	answer, err := question.GetAnswer(answerID)
	if err != nil {
		return err
	}

	// Calculate points
	points := Points{}
	if answer.IsCorrect() {
		points = question.Points()
		qs.score = qs.score.Add(points)
	}

	// Record answer
	userAnswer := NewUserAnswer(question.ID(), answerID, answer.IsCorrect(), points, answeredAt)
	qs.answers = append(qs.answers, userAnswer)
	qs.currentQuestion++

	// Record domain event
	qs.events = append(qs.events, NewAnswerSubmittedEvent(
		qs.id,
		question.ID(),
		answerID,
		answer.IsCorrect(),
		points,
		answeredAt,
	))

	return nil
}

// Complete marks the session as completed
func (qs *QuizSession) Complete(completedAt int64) error {
	if qs.status != SessionStatusActive {
		return ErrSessionCompleted
	}

	qs.status = SessionStatusCompleted
	qs.completedAt = completedAt

	// Record domain event
	qs.events = append(qs.events, NewQuizCompletedEvent(
		qs.quizID,
		qs.id,
		qs.userID,
		qs.score,
		completedAt,
	))

	return nil
}

// Abandon marks the session as abandoned
func (qs *QuizSession) Abandon(abandonedAt int64) error {
	if qs.status != SessionStatusActive {
		return ErrSessionCompleted
	}

	qs.status = SessionStatusAbandoned
	qs.completedAt = abandonedAt
	return nil
}

// HasPassed checks if the user passed the quiz
func (qs *QuizSession) HasPassed(quiz *Quiz) bool {
	if qs.status != SessionStatusCompleted {
		return false
	}

	totalPoints := quiz.GetTotalPoints()
	if totalPoints.IsZero() {
		return false
	}

	scorePercentage := (qs.score.Value() * 100) / totalPoints.Value()
	return scorePercentage >= quiz.PassingScore().Percentage()
}

// IsCompleted checks if the session is completed
func (qs *QuizSession) IsCompleted() bool {
	return qs.status == SessionStatusCompleted
}

// IsActive checks if the session is active
func (qs *QuizSession) IsActive() bool {
	return qs.status == SessionStatusActive
}

// HasAnsweredQuestion checks if a question was already answered
func (qs *QuizSession) HasAnsweredQuestion(questionID QuestionID) bool {
	for _, ua := range qs.answers {
		if ua.QuestionID().Equals(questionID) {
			return true
		}
	}
	return false
}

// Getters
func (qs *QuizSession) ID() SessionID         { return qs.id }
func (qs *QuizSession) QuizID() QuizID        { return qs.quizID }
func (qs *QuizSession) UserID() shared.UserID { return qs.userID }
func (qs *QuizSession) CurrentQuestion() int  { return qs.currentQuestion }
func (qs *QuizSession) Score() Points         { return qs.score }
func (qs *QuizSession) StartedAt() int64      { return qs.startedAt }
func (qs *QuizSession) CompletedAt() int64    { return qs.completedAt }
func (qs *QuizSession) Status() SessionStatus { return qs.status }

// Answers returns a copy of answers (protect internal state)
func (qs *QuizSession) Answers() []UserAnswer {
	copies := make([]UserAnswer, len(qs.answers))
	copy(copies, qs.answers)
	return copies
}

// AnswersCount returns the number of submitted answers
func (qs *QuizSession) AnswersCount() int {
	return len(qs.answers)
}

// Events returns collected domain events and clears them
func (qs *QuizSession) Events() []Event {
	events := qs.events
	qs.events = make([]Event, 0)
	return events
}
