package quiz

// Answer is an entity representing a possible answer to a question
type Answer struct {
	id        AnswerID
	text      AnswerText
	isCorrect bool
	position  int
}

// NewAnswer creates a new Answer entity
func NewAnswer(id AnswerID, text AnswerText, isCorrect bool, position int) (*Answer, error) {
	if id.IsZero() {
		return nil, ErrInvalidAnswerID
	}

	return &Answer{
		id:        id,
		text:      text,
		isCorrect: isCorrect,
		position:  position,
	}, nil
}

// Getters (no setters - immutable after creation)
func (a *Answer) ID() AnswerID       { return a.id }
func (a *Answer) Text() AnswerText   { return a.text }
func (a *Answer) IsCorrect() bool    { return a.isCorrect }
func (a *Answer) Position() int      { return a.position }

// Question is an entity representing a quiz question
type Question struct {
	id       QuestionID
	text     QuestionText
	answers  []Answer
	points   Points
	position int
}

// NewQuestion creates a new Question entity
func NewQuestion(id QuestionID, text QuestionText, points Points, position int) (*Question, error) {
	if id.IsZero() {
		return nil, ErrInvalidQuestionID
	}

	return &Question{
		id:       id,
		text:     text,
		points:   points,
		position: position,
		answers:  make([]Answer, 0),
	}, nil
}

// AddAnswer adds an answer to the question
func (q *Question) AddAnswer(answer Answer) error {
	if len(q.answers) >= 4 {
		return ErrTooManyAnswers
	}

	q.answers = append(q.answers, answer)
	return nil
}

// HasCorrectAnswer checks if question has at least one correct answer
func (q *Question) HasCorrectAnswer() bool {
	for _, answer := range q.answers {
		if answer.IsCorrect() {
			return true
		}
	}
	return false
}

// IsValidAnswer checks if the given answer ID belongs to this question
func (q *Question) IsValidAnswer(answerID AnswerID) bool {
	for _, answer := range q.answers {
		if answer.ID().Equals(answerID) {
			return true
		}
	}
	return false
}

// GetAnswer returns answer by ID
func (q *Question) GetAnswer(answerID AnswerID) (*Answer, error) {
	for i := range q.answers {
		if q.answers[i].ID().Equals(answerID) {
			return &q.answers[i], nil
		}
	}
	return nil, ErrAnswerNotFound
}

// Getters
func (q *Question) ID() QuestionID     { return q.id }
func (q *Question) Text() QuestionText { return q.text }
func (q *Question) Points() Points     { return q.points }
func (q *Question) Position() int      { return q.position }

// Answers returns a copy of answers (protect internal state)
func (q *Question) Answers() []Answer {
	copies := make([]Answer, len(q.answers))
	copy(copies, q.answers)
	return copies
}

// UserAnswer represents a user's answer to a question (entity within QuizSession aggregate)
type UserAnswer struct {
	questionID QuestionID
	answerID   AnswerID
	isCorrect  bool
	points     Points
	answeredAt int64 // Unix timestamp (no time.Time to keep domain pure)
}

// NewUserAnswer creates a new UserAnswer
func NewUserAnswer(questionID QuestionID, answerID AnswerID, isCorrect bool, points Points, answeredAt int64) UserAnswer {
	return UserAnswer{
		questionID: questionID,
		answerID:   answerID,
		isCorrect:  isCorrect,
		points:     points,
		answeredAt: answeredAt,
	}
}

// Getters
func (ua UserAnswer) QuestionID() QuestionID { return ua.questionID }
func (ua UserAnswer) AnswerID() AnswerID     { return ua.answerID }
func (ua UserAnswer) IsCorrect() bool        { return ua.isCorrect }
func (ua UserAnswer) Points() Points         { return ua.points }
func (ua UserAnswer) AnsweredAt() int64      { return ua.answeredAt }
