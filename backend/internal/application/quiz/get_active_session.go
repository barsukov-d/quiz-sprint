package quiz

import (
	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
)

// GetActiveSessionUseCase handles retrieving an active quiz session
type GetActiveSessionUseCase struct {
	quizRepo    quiz.QuizRepository
	sessionRepo quiz.SessionRepository
}

// NewGetActiveSessionUseCase creates a new GetActiveSessionUseCase
func NewGetActiveSessionUseCase(
	quizRepo quiz.QuizRepository,
	sessionRepo quiz.SessionRepository,
) *GetActiveSessionUseCase {
	return &GetActiveSessionUseCase{
		quizRepo:    quizRepo,
		sessionRepo: sessionRepo,
	}
}

// GetActiveSessionInput is the input DTO
type GetActiveSessionInput struct {
	QuizID string
	UserID string
}

// GetActiveSessionOutput is the output DTO
type GetActiveSessionOutput struct {
	Session              SessionDTO  `json:"session"`
	CurrentQuestion      QuestionDTO `json:"currentQuestion"`
	TotalQuestions       int         `json:"totalQuestions"`
	TimeLimit            int         `json:"timeLimit"`
	TimeLimitPerQuestion int         `json:"timeLimitPerQuestion"`
}

// Execute retrieves the active session for a user and quiz
func (uc *GetActiveSessionUseCase) Execute(input GetActiveSessionInput) (GetActiveSessionOutput, error) {
	// 1. Validate and convert input
	quizID, err := quiz.NewQuizIDFromString(input.QuizID)
	if err != nil {
		return GetActiveSessionOutput{}, err
	}

	userID, err := shared.NewUserID(input.UserID)
	if err != nil {
		return GetActiveSessionOutput{}, err
	}

	// 2. Find active session
	session, err := uc.sessionRepo.FindActiveByUserAndQuiz(userID, quizID)
	if err != nil {
		return GetActiveSessionOutput{}, quiz.ErrSessionNotFound
	}

	// 3. Load quiz to get current question
	quizAggregate, err := uc.quizRepo.FindByID(quizID)
	if err != nil {
		return GetActiveSessionOutput{}, err
	}

	// 4. Get current question (session tracks which question we're on)
	currentQuestion, err := quizAggregate.GetQuestionByIndex(session.CurrentQuestion())
	if err != nil {
		return GetActiveSessionOutput{}, err
	}

	// 5. Return DTO
	return GetActiveSessionOutput{
		Session:              ToSessionDTO(session),
		CurrentQuestion:      ToQuestionDTO(currentQuestion),
		TotalQuestions:       quizAggregate.QuestionsCount(),
		TimeLimit:            quizAggregate.TimeLimit().Seconds(),
		TimeLimitPerQuestion: quizAggregate.TimeLimitPerQuestion(),
	}, nil
}
