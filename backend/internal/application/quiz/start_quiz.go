package quiz

import (
	"time"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
)

// StartQuizUseCase handles the business logic for starting a quiz
type StartQuizUseCase struct {
	quizRepo    quiz.QuizRepository
	sessionRepo quiz.SessionRepository
	eventBus    quiz.EventBus
}

// NewStartQuizUseCase creates a new StartQuizUseCase
func NewStartQuizUseCase(
	quizRepo quiz.QuizRepository,
	sessionRepo quiz.SessionRepository,
	eventBus quiz.EventBus,
) *StartQuizUseCase {
	return &StartQuizUseCase{
		quizRepo:    quizRepo,
		sessionRepo: sessionRepo,
		eventBus:    eventBus,
	}
}

// Execute starts a quiz session
func (uc *StartQuizUseCase) Execute(input StartQuizInput) (StartQuizOutput, error) {
	// 1. Validate and convert input to domain types
	quizID, err := quiz.NewQuizIDFromString(input.QuizID)
	if err != nil {
		return StartQuizOutput{}, err
	}

	userID, err := shared.NewUserID(input.UserID)
	if err != nil {
		return StartQuizOutput{}, err
	}

	// 2. Load quiz aggregate
	quizAggregate, err := uc.quizRepo.FindByID(quizID)
	if err != nil {
		return StartQuizOutput{}, err
	}

	// 3. Apply business rules (delegate to domain)
	if err := quizAggregate.CanStart(); err != nil {
		return StartQuizOutput{}, err
	}

	// 4. Check for existing active session
	existingSession, err := uc.sessionRepo.FindActiveByUserAndQuiz(userID, quizID)
	if err == nil && existingSession != nil {
		return StartQuizOutput{}, quiz.ErrSessionAlreadyExists
	}

	// 5. Create new session aggregate
	sessionID := quiz.NewSessionID()
	now := time.Now().Unix()

	session, err := quiz.NewQuizSession(sessionID, quizID, userID, now)
	if err != nil {
		return StartQuizOutput{}, err
	}

	// 6. Persist session
	if err := uc.sessionRepo.Save(session); err != nil {
		return StartQuizOutput{}, err
	}

	// 7. Publish domain events
	if uc.eventBus != nil {
		uc.eventBus.Publish(session.Events()...)
	}

	// 8. Get first question
	firstQuestion, err := quizAggregate.GetQuestionByIndex(0)
	if err != nil {
		return StartQuizOutput{}, err
	}

	// 9. Return DTO (not domain models!)
	return StartQuizOutput{
		Session:              ToSessionDTO(session),
		FirstQuestion:        ToQuestionDTO(firstQuestion),
		TotalQuestions:       quizAggregate.QuestionsCount(),
		TimeLimit:            quizAggregate.TimeLimit().Seconds(),
		TimeLimitPerQuestion: quizAggregate.TimeLimitPerQuestion(),
	}, nil
}
