package quiz

import (
	domainQuiz "github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
)

// GetSessionResultsInput is the input for GetSessionResultsUseCase
type GetSessionResultsInput struct {
	SessionID string
}

// GetSessionResultsOutput is the output for GetSessionResultsUseCase
type GetSessionResultsOutput struct {
	Session         SessionDTO
	Quiz            QuizDTO
	TotalQuestions  int
	CorrectAnswers  int
	TimeSpent       int64   // seconds
	Passed          bool
	ScorePercentage int
	LongestStreak   int     // Longest streak of correct answers
	AvgAnswerTime   float64 // Average answer time in seconds
}

// GetSessionResultsUseCase retrieves the results of a completed (or active) quiz session
type GetSessionResultsUseCase struct {
	sessionRepo domainQuiz.SessionRepository
	quizRepo    domainQuiz.QuizRepository
}

// NewGetSessionResultsUseCase creates a new GetSessionResultsUseCase
func NewGetSessionResultsUseCase(
	sessionRepo domainQuiz.SessionRepository,
	quizRepo    domainQuiz.QuizRepository,
) *GetSessionResultsUseCase {
	return &GetSessionResultsUseCase{
		sessionRepo: sessionRepo,
		quizRepo:    quizRepo,
	}
}

// Execute retrieves session results with calculated statistics
func (uc *GetSessionResultsUseCase) Execute(input GetSessionResultsInput) (*GetSessionResultsOutput, error) {
	// 1. Parse SessionID
	sessionID, err := domainQuiz.NewSessionIDFromString(input.SessionID)
	if err != nil {
		return nil, err
	}

	// 2. Find session
	session, err := uc.sessionRepo.FindByID(sessionID)
	if err != nil {
		return nil, err
	}

	// 3. Get the quiz
	quiz, err := uc.quizRepo.FindByID(session.QuizID())
	if err != nil {
		return nil, err
	}

	// 4. Calculate statistics
	totalQuestions := quiz.QuestionsCount()
	correctAnswers := 0
	for _, answer := range session.Answers() {
		if answer.IsCorrect() {
			correctAnswers++
		}
	}

	// 5. Calculate time spent (in seconds)
	timeSpent := int64(0)
	if session.CompletedAt() > 0 {
		timeSpent = session.CompletedAt() - session.StartedAt()
	}

	// 6. Calculate score percentage
	totalPoints := quiz.GetTotalPoints().Value()
	scorePercentage := 0
	if totalPoints > 0 {
		scorePercentage = (session.Score().Value() * 100) / totalPoints
	}

	// 7. Check if passed
	passed := session.HasPassed(quiz)

	// 8. Calculate longest streak and average answer time
	longestStreak := calculateLongestStreak(session.Answers())
	avgAnswerTime := calculateAvgAnswerTime(session.Answers())

	// 9. Map to DTOs
	sessionDTO := SessionDTO{
		ID:              session.ID().String(),
		QuizID:          session.QuizID().String(),
		UserID:          session.UserID().String(),
		CurrentQuestion: session.CurrentQuestion(),
		Score:           session.Score().Value(),
		StartedAt:       session.StartedAt(),
		CompletedAt:     session.CompletedAt(),
		Status:          session.Status().String(),
	}

	quizDTO := QuizDTO{
		ID:           quiz.ID().String(),
		Title:        quiz.Title().String(),
		Description:  quiz.Description(),
		TimeLimit:    quiz.TimeLimit().Seconds(),
		PassingScore: quiz.PassingScore().Percentage(),
	}

	return &GetSessionResultsOutput{
		Session:         sessionDTO,
		Quiz:            quizDTO,
		TotalQuestions:  totalQuestions,
		CorrectAnswers:  correctAnswers,
		TimeSpent:       timeSpent,
		Passed:          passed,
		ScorePercentage: scorePercentage,
		LongestStreak:   longestStreak,
		AvgAnswerTime:   avgAnswerTime,
	}, nil
}

// calculateLongestStreak finds the longest streak of correct answers
func calculateLongestStreak(answers []domainQuiz.UserAnswer) int {
	longestStreak := 0
	currentStreak := 0

	for _, answer := range answers {
		if answer.IsCorrect() {
			currentStreak++
			if currentStreak > longestStreak {
				longestStreak = currentStreak
			}
		} else {
			currentStreak = 0
		}
	}

	return longestStreak
}

// calculateAvgAnswerTime calculates average answer time in seconds
func calculateAvgAnswerTime(answers []domainQuiz.UserAnswer) float64 {
	if len(answers) == 0 {
		return 0
	}

	totalTime := int64(0)
	for _, answer := range answers {
		totalTime += answer.TimeSpent() // milliseconds
	}

	avgMs := float64(totalTime) / float64(len(answers))
	return avgMs / 1000.0 // Convert to seconds
}
