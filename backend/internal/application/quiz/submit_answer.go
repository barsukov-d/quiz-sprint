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

	// 6. Submit answer (domain business logic)
	now := time.Now().Unix()
	if err := session.SubmitAnswer(question, answerID, now); err != nil {
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

	// 10. Get the submitted answer details
	answer, _ := question.GetAnswer(answerID)

	// 11. Build output
	output := SubmitAnswerOutput{
		IsCorrect:       answer.IsCorrect(),
		CorrectAnswerID: FindCorrectAnswerID(question),
		PointsEarned:    0,
		TotalScore:      session.Score().Value(),
		IsQuizCompleted: isCompleted,
	}

	if answer.IsCorrect() {
		output.PointsEarned = question.Points().Value()
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
