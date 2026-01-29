package quiz

import (
	"time"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
)

// SubmitAnswerUseCase handles the business logic for submitting an answer
type SubmitAnswerUseCase struct {
	quizRepo    quiz.QuizRepository
	sessionRepo quiz.SessionRepository
	eventBus    quiz.EventBus
}

// NewSubmitAnswerUseCase creates a new SubmitAnswerUseCase
func NewSubmitAnswerUseCase(
	quizRepo quiz.QuizRepository,
	sessionRepo quiz.SessionRepository,
	eventBus quiz.EventBus,
) *SubmitAnswerUseCase {
	return &SubmitAnswerUseCase{
		quizRepo:    quizRepo,
		sessionRepo: sessionRepo,
		eventBus:    eventBus,
	}
}

// Execute submits an answer for a quiz question
func (uc *SubmitAnswerUseCase) Execute(input SubmitAnswerInput) (SubmitAnswerOutput, error) {
	// 1. Validate and convert input to domain types
	sessionID, err := quiz.NewSessionIDFromString(input.SessionID)
	if err != nil {
		return SubmitAnswerOutput{}, err
	}

	questionID, err := quiz.NewQuestionIDFromString(input.QuestionID)
	if err != nil {
		return SubmitAnswerOutput{}, err
	}

	answerID, err := quiz.NewAnswerIDFromString(input.AnswerID)
	if err != nil {
		return SubmitAnswerOutput{}, err
	}

	userID, err := shared.NewUserID(input.UserID)
	if err != nil {
		return SubmitAnswerOutput{}, err
	}

	// 1a. Validate timeTaken
	if input.TimeTaken < 0 {
		return SubmitAnswerOutput{}, quiz.ErrInvalidTimeTaken
	}
	if input.TimeTaken > 3600000 { // Max 1 hour in milliseconds
		return SubmitAnswerOutput{}, quiz.ErrTimeTakenTooLong
	}

	// 2. Load session aggregate
	session, err := uc.sessionRepo.FindByID(sessionID)
	if err != nil {
		return SubmitAnswerOutput{}, err
	}

	// 3. Validate session belongs to user
	if !session.UserID().Equals(userID) {
		return SubmitAnswerOutput{}, quiz.ErrUnauthorized
	}

	// 4. Load quiz aggregate
	quizAggregate, err := uc.quizRepo.FindByID(session.QuizID())
	if err != nil {
		return SubmitAnswerOutput{}, err
	}

	// 5. Get the question
	question, err := quizAggregate.GetQuestion(questionID)
	if err != nil {
		return SubmitAnswerOutput{}, err
	}

	// 6. Submit answer (domain business logic with new scoring system)
	now := time.Now().Unix()
	result, err := session.SubmitAnswer(question, answerID, now, input.TimeTaken, quizAggregate)
	if err != nil {
		return SubmitAnswerOutput{}, err
	}

	// 7. Check if quiz is completed
	isCompleted := session.CurrentQuestion() >= quizAggregate.QuestionsCount()
	if isCompleted {
		if err := session.Complete(now); err != nil {
			return SubmitAnswerOutput{}, err
		}
	}

	// 8. Persist session
	if err := uc.sessionRepo.Save(session); err != nil {
		return SubmitAnswerOutput{}, err
	}

	// 9. Publish domain events
	if uc.eventBus != nil {
		uc.eventBus.Publish(session.Events()...)
	}

	// 10. Build output with detailed points breakdown
	output := SubmitAnswerOutput{
		IsCorrect:       result.IsCorrect,
		CorrectAnswerID: FindCorrectAnswerID(question),
		BasePoints:      result.BasePoints.Value(),
		TimeBonus:       result.TimeBonus.Value(),
		StreakBonus:     result.StreakBonus.Value(),
		PointsEarned:    result.TotalPoints.Value(),
		CurrentStreak:   result.CurrentStreak,
		TotalScore:      session.Score().Value(),
		IsQuizCompleted: isCompleted,
	}

	// 12. Include next question or final result
	if isCompleted {
		finalResult := BuildFinalResult(session, quizAggregate)
		output.FinalResult = &finalResult
	} else {
		nextQuestion, err := quizAggregate.GetQuestionByIndex(session.CurrentQuestion())
		if err == nil {
			dto := ToQuestionDTO(nextQuestion)
			output.NextQuestion = &dto
		}
	}

	return output, nil
}
