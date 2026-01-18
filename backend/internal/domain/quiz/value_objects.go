package quiz

import (
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
)

// QuizID is a value object for quiz identifier
type QuizID struct {
	id shared.ID
}

// NewQuizID creates a new QuizID
func NewQuizID() QuizID {
	return QuizID{id: shared.NewID()}
}

// NewQuizIDFromString creates QuizID from string
func NewQuizIDFromString(value string) (QuizID, error) {
	id, err := shared.NewIDFromString(value)
	if err != nil {
		return QuizID{}, ErrInvalidQuizID
	}
	return QuizID{id: id}, nil
}

func (id QuizID) String() string {
	return id.id.String()
}

func (id QuizID) Equals(other QuizID) bool {
	return id.id.Equals(other.id)
}

func (id QuizID) IsZero() bool {
	return id.id.IsZero()
}

// QuestionID is a value object for question identifier
type QuestionID struct {
	id shared.ID
}

func NewQuestionID() QuestionID {
	return QuestionID{id: shared.NewID()}
}

func NewQuestionIDFromString(value string) (QuestionID, error) {
	id, err := shared.NewIDFromString(value)
	if err != nil {
		return QuestionID{}, ErrInvalidQuestionID
	}
	return QuestionID{id: id}, nil
}

func (id QuestionID) String() string {
	return id.id.String()
}

func (id QuestionID) Equals(other QuestionID) bool {
	return id.id.Equals(other.id)
}

func (id QuestionID) IsZero() bool {
	return id.id.IsZero()
}

// AnswerID is a value object for answer identifier
type AnswerID struct {
	id shared.ID
}

func NewAnswerID() AnswerID {
	return AnswerID{id: shared.NewID()}
}

func NewAnswerIDFromString(value string) (AnswerID, error) {
	id, err := shared.NewIDFromString(value)
	if err != nil {
		return AnswerID{}, ErrInvalidAnswerID
	}
	return AnswerID{id: id}, nil
}

func (id AnswerID) String() string {
	return id.id.String()
}

func (id AnswerID) Equals(other AnswerID) bool {
	return id.id.Equals(other.id)
}

func (id AnswerID) IsZero() bool {
	return id.id.IsZero()
}

// SessionID is a value object for quiz session identifier
type SessionID struct {
	id shared.ID
}

func NewSessionID() SessionID {
	return SessionID{id: shared.NewID()}
}

func NewSessionIDFromString(value string) (SessionID, error) {
	id, err := shared.NewIDFromString(value)
	if err != nil {
		return SessionID{}, ErrInvalidSessionID
	}
	return SessionID{id: id}, nil
}

func (id SessionID) String() string {
	return id.id.String()
}

func (id SessionID) Equals(other SessionID) bool {
	return id.id.Equals(other.id)
}

func (id SessionID) IsZero() bool {
	return id.id.IsZero()
}

// QuizTitle is a value object for quiz title
type QuizTitle struct {
	value string
}

func NewQuizTitle(value string) (QuizTitle, error) {
	if value == "" {
		return QuizTitle{}, ErrInvalidTitle
	}
	if len(value) > 200 {
		return QuizTitle{}, ErrTitleTooLong
	}
	return QuizTitle{value: value}, nil
}

func (t QuizTitle) String() string {
	return t.value
}

func (t QuizTitle) IsEmpty() bool {
	return t.value == ""
}

// QuestionText is a value object for question text
type QuestionText struct {
	value string
}

func NewQuestionText(value string) (QuestionText, error) {
	if value == "" {
		return QuestionText{}, ErrInvalidQuestionText
	}
	if len(value) > 500 {
		return QuestionText{}, ErrQuestionTextTooLong
	}
	return QuestionText{value: value}, nil
}

func (t QuestionText) String() string {
	return t.value
}

func (t QuestionText) IsEmpty() bool {
	return t.value == ""
}

// AnswerText is a value object for answer text
type AnswerText struct {
	value string
}

func NewAnswerText(value string) (AnswerText, error) {
	if value == "" {
		return AnswerText{}, ErrInvalidAnswerText
	}
	if len(value) > 200 {
		return AnswerText{}, ErrAnswerTextTooLong
	}
	return AnswerText{value: value}, nil
}

func (t AnswerText) String() string {
	return t.value
}

// Points is a value object for points/score
type Points struct {
	value int
}

func NewPoints(value int) (Points, error) {
	if value < 0 {
		return Points{}, ErrNegativePoints
	}
	if value > 1000 {
		return Points{}, ErrPointsTooHigh
	}
	return Points{value: value}, nil
}

func (p Points) Value() int {
	return p.value
}

func (p Points) Add(other Points) Points {
	return Points{value: p.value + other.value}
}

func (p Points) IsZero() bool {
	return p.value == 0
}

// TimeLimit is a value object for time limit in seconds
type TimeLimit struct {
	seconds int
}

func NewTimeLimit(seconds int) (TimeLimit, error) {
	if seconds <= 0 {
		return TimeLimit{}, ErrInvalidTimeLimit
	}
	if seconds > 3600 { // Max 1 hour
		return TimeLimit{}, ErrTimeLimitTooHigh
	}
	return TimeLimit{seconds: seconds}, nil
}

func (t TimeLimit) Seconds() int {
	return t.seconds
}

func (t TimeLimit) IsZero() bool {
	return t.seconds == 0
}

// PassingScore is a value object for passing score percentage
type PassingScore struct {
	percentage int
}

func NewPassingScore(percentage int) (PassingScore, error) {
	if percentage < 0 || percentage > 100 {
		return PassingScore{}, ErrInvalidPassingScore
	}
	return PassingScore{percentage: percentage}, nil
}

func (ps PassingScore) Percentage() int {
	return ps.percentage
}

func (ps PassingScore) IsZero() bool {
	return ps.percentage == 0
}

// CategoryID is a value object for category identifier
type CategoryID struct {
	id shared.ID
}

func NewCategoryID() CategoryID {
	return CategoryID{id: shared.NewID()}
}

func NewCategoryIDFromString(value string) (CategoryID, error) {
	id, err := shared.NewIDFromString(value)
	if err != nil {
		return CategoryID{}, ErrInvalidCategoryID
	}
	return CategoryID{id: id}, nil
}

func (id CategoryID) String() string {
	return id.id.String()
}

func (id CategoryID) Equals(other CategoryID) bool {
	return id.id.Equals(other.id)
}

func (id CategoryID) IsZero() bool {
	return id.id.IsZero()
}

// CategoryName is a value object for category name
type CategoryName struct {
	value string
}

// NewCategoryName creates a new CategoryName value object after validation.
func NewCategoryName(value string) (CategoryName, error) {
	if value == "" {
		return CategoryName{}, ErrInvalidCategoryName
	}
	if len(value) > 100 {
		return CategoryName{}, ErrCategoryNameTooLong
	}
	return CategoryName{value: value}, nil
}

// String returns the primitive string value.
func (n CategoryName) String() string {
	return n.value
}
