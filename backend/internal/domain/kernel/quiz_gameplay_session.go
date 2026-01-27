package kernel

import (
	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
)

// Type aliases for better readability
type QuestionID = quiz.QuestionID
type AnswerID = quiz.AnswerID
type Points = quiz.Points

// QuizGameplaySession represents a single player's session of answering questions.
// This is the SHARED KERNEL - pure gameplay logic without mode-specific business rules.
// Used by: Solo Marathon (endless with lives), Daily Challenge (10 questions, no feedback until end).
// NOT used by: Quick Duel (PvP round-by-round), Party Mode (multiplayer synchronized).
type QuizGameplaySession struct {
	id                   SessionID
	quiz                 *quiz.Quiz            // Reference to quiz content
	userAnswers          map[QuestionID]AnswerData // QuestionID -> AnswerData
	currentQuestionIndex int
	baseScore            Points // Base score (without mode-specific bonuses/multipliers)
	startedAt            int64  // Unix timestamp (no time.Time to keep domain pure)
	finishedAt           int64  // Unix timestamp (0 if not finished)
}

// AnswerData stores the answer and metadata for a question
type AnswerData struct {
	answerID   AnswerID
	isCorrect  bool
	timeTaken  int64 // milliseconds
	answeredAt int64 // Unix timestamp
}

// NewAnswerData creates a new AnswerData (used for reconstruction from persistence)
func NewAnswerData(answerID AnswerID, isCorrect bool, timeTaken int64, answeredAt int64) AnswerData {
	return AnswerData{
		answerID:   answerID,
		isCorrect:  isCorrect,
		timeTaken:  timeTaken,
		answeredAt: answeredAt,
	}
}

// Getters for AnswerData
func (a AnswerData) AnswerID() AnswerID   { return a.answerID }
func (a AnswerData) IsCorrect() bool      { return a.isCorrect }
func (a AnswerData) TimeTaken() int64     { return a.timeTaken }
func (a AnswerData) AnsweredAt() int64    { return a.answeredAt }

// NewQuizGameplaySession creates a new gameplay session
func NewQuizGameplaySession(id SessionID, quiz *quiz.Quiz, startedAt int64) (*QuizGameplaySession, error) {
	if id.IsZero() {
		return nil, ErrInvalidSessionID
	}

	if quiz == nil {
		return nil, ErrInvalidQuiz
	}

	if err := quiz.CanStart(); err != nil {
		return nil, err
	}

	return &QuizGameplaySession{
		id:                   id,
		quiz:                 quiz,
		userAnswers:          make(map[QuestionID]AnswerData),
		currentQuestionIndex: 0,
		baseScore:            Points{},
		startedAt:            startedAt,
		finishedAt:           0,
	}, nil
}

// AnswerQuestionResult contains the result of answering a question
type AnswerQuestionResult struct {
	IsCorrect  bool
	BasePoints Points // Base points earned (without mode bonuses)
	TimeTaken  int64  // milliseconds
}

// AnswerQuestion records a user's answer and returns the result.
// This is PURE gameplay logic - no streak, no multiplier, no mode-specific rules.
func (s *QuizGameplaySession) AnswerQuestion(
	questionID QuestionID,
	answerID AnswerID,
	timeTaken int64,
	answeredAt int64,
) (*AnswerQuestionResult, error) {
	// 1. Validate not finished
	if s.IsFinished() {
		return nil, ErrSessionFinished
	}

	// 2. Validate question exists and not already answered
	if _, exists := s.userAnswers[questionID]; exists {
		return nil, quiz.ErrAlreadyAnswered
	}

	// 3. Get current question
	question, err := s.quiz.GetQuestion(questionID)
	if err != nil {
		return nil, err
	}

	// 4. Validate answer belongs to question
	answer, err := question.GetAnswer(answerID)
	if err != nil {
		return nil, err
	}

	// 5. Check correctness
	isCorrect := answer.IsCorrect()

	// 6. Calculate base points (only if correct)
	var basePoints Points
	if isCorrect {
		basePoints = question.Points()
		if basePoints.IsZero() {
			basePoints = s.quiz.BasePoints()
		}
		s.baseScore = s.baseScore.Add(basePoints)
	}

	// 7. Record answer
	s.userAnswers[questionID] = AnswerData{
		answerID:   answerID,
		isCorrect:  isCorrect,
		timeTaken:  timeTaken,
		answeredAt: answeredAt,
	}

	// 8. Move to next question
	s.currentQuestionIndex++

	return &AnswerQuestionResult{
		IsCorrect:  isCorrect,
		BasePoints: basePoints,
		TimeTaken:  timeTaken,
	}, nil
}

// IsFinished checks if all questions have been answered
func (s *QuizGameplaySession) IsFinished() bool {
	return s.currentQuestionIndex >= s.quiz.QuestionsCount()
}

// Finish marks the session as finished
func (s *QuizGameplaySession) Finish(finishedAt int64) error {
	if s.finishedAt > 0 {
		return ErrSessionFinished
	}

	s.finishedAt = finishedAt
	return nil
}

// GetCurrentQuestion returns the current question to be answered
func (s *QuizGameplaySession) GetCurrentQuestion() (*quiz.Question, error) {
	if s.IsFinished() {
		return nil, ErrSessionFinished
	}

	return s.quiz.GetQuestionByIndex(s.currentQuestionIndex)
}

// GetAnswer returns the user's answer for a specific question
func (s *QuizGameplaySession) GetAnswer(questionID QuestionID) (AnswerData, bool) {
	data, exists := s.userAnswers[questionID]
	return data, exists
}

// GetScoreByQuestion returns cumulative score at each question index
// Used for Ghost Comparison in Classic Mode
func (s *QuizGameplaySession) GetScoreByQuestion() []int {
	scores := make([]int, 0, s.quiz.QuestionsCount())
	cumulativeScore := 0

	questions := s.quiz.Questions()
	for _, question := range questions {
		if answerData, exists := s.userAnswers[question.ID()]; exists && answerData.isCorrect {
			basePoints := question.Points()
			if basePoints.IsZero() {
				basePoints = s.quiz.BasePoints()
			}
			cumulativeScore += basePoints.Value()
		}
		scores = append(scores, cumulativeScore)
	}

	return scores
}

// CountCorrectAnswers returns the number of correct answers
func (s *QuizGameplaySession) CountCorrectAnswers() int {
	count := 0
	for _, answerData := range s.userAnswers {
		if answerData.isCorrect {
			count++
		}
	}
	return count
}

// GetAccuracy returns accuracy percentage (0-100)
func (s *QuizGameplaySession) GetAccuracy() int {
	if len(s.userAnswers) == 0 {
		return 0
	}
	correct := s.CountCorrectAnswers()
	return (correct * 100) / len(s.userAnswers)
}

// Getters
func (s *QuizGameplaySession) ID() SessionID               { return s.id }
func (s *QuizGameplaySession) Quiz() *quiz.Quiz            { return s.quiz }
func (s *QuizGameplaySession) CurrentQuestionIndex() int   { return s.currentQuestionIndex }
func (s *QuizGameplaySession) BaseScore() Points           { return s.baseScore }
func (s *QuizGameplaySession) StartedAt() int64            { return s.startedAt }
func (s *QuizGameplaySession) FinishedAt() int64           { return s.finishedAt }
func (s *QuizGameplaySession) AnswersCount() int           { return len(s.userAnswers) }

// GetAllAnswers returns all recorded answers
func (s *QuizGameplaySession) GetAllAnswers() map[QuestionID]AnswerData {
	// Return copy to protect internal state
	answers := make(map[QuestionID]AnswerData, len(s.userAnswers))
	for k, v := range s.userAnswers {
		answers[k] = v
	}
	return answers
}

// ReconstructQuizGameplaySession reconstructs a session from persistence
// Used by repository when loading from database
func ReconstructQuizGameplaySession(
	id SessionID,
	quiz *quiz.Quiz,
	userAnswers map[QuestionID]AnswerData,
	currentQuestionIndex int,
	baseScore Points,
	startedAt int64,
	finishedAt int64,
) *QuizGameplaySession {
	return &QuizGameplaySession{
		id:                   id,
		quiz:                 quiz,
		userAnswers:          userAnswers,
		currentQuestionIndex: currentQuestionIndex,
		baseScore:            baseScore,
		startedAt:            startedAt,
		finishedAt:           finishedAt,
	}
}
