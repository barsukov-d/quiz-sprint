package quiz

import (
	"context"
	"time"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/google/uuid"
)

// SubmitAnswerCommand contains the data needed to submit an answer
type SubmitAnswerCommand struct {
	SessionID  uuid.UUID
	QuestionID uuid.UUID
	AnswerID   uuid.UUID
	UserID     string
}

// SubmitAnswerResult contains the result of submitting an answer
type SubmitAnswerResult struct {
	IsCorrect       bool
	Points          int
	TotalScore      int
	IsQuizCompleted bool
	NextQuestion    *quiz.Question
}

// SubmitAnswerUseCase handles the business logic for submitting an answer
type SubmitAnswerUseCase struct {
	repo quiz.QuizRepository
}

// NewSubmitAnswerUseCase creates a new SubmitAnswerUseCase
func NewSubmitAnswerUseCase(repo quiz.QuizRepository) *SubmitAnswerUseCase {
	return &SubmitAnswerUseCase{repo: repo}
}

// Execute submits an answer for a quiz question
func (uc *SubmitAnswerUseCase) Execute(ctx context.Context, cmd SubmitAnswerCommand) (*SubmitAnswerResult, error) {
	// Get session
	session, err := uc.repo.FindSessionByID(ctx, cmd.SessionID)
	if err != nil {
		return nil, err
	}

	// Validate session belongs to user
	if session.UserID != cmd.UserID {
		return nil, quiz.ErrUnauthorized
	}

	// Validate session is active
	if session.Status != quiz.SessionStatusActive {
		return nil, quiz.ErrSessionCompleted
	}

	// Get quiz
	quizData, err := uc.repo.FindByID(ctx, session.QuizID)
	if err != nil {
		return nil, err
	}

	// Find the question
	var question *quiz.Question
	for i := range quizData.Questions {
		if quizData.Questions[i].ID == cmd.QuestionID {
			question = &quizData.Questions[i]
			break
		}
	}

	if question == nil {
		return nil, quiz.ErrQuestionNotFound
	}

	// Check if already answered
	for _, answer := range session.Answers {
		if answer.QuestionID == cmd.QuestionID {
			return nil, quiz.ErrAlreadyAnswered
		}
	}

	// Find the selected answer
	var selectedAnswer *quiz.Answer
	for i := range question.Answers {
		if question.Answers[i].ID == cmd.AnswerID {
			selectedAnswer = &question.Answers[i]
			break
		}
	}

	if selectedAnswer == nil {
		return nil, quiz.ErrAnswerNotFound
	}

	// Calculate points
	points := 0
	if selectedAnswer.IsCorrect {
		points = question.Points
		session.Score += points
	}

	// Record answer
	userAnswer := quiz.UserAnswer{
		QuestionID: cmd.QuestionID,
		AnswerID:   cmd.AnswerID,
		IsCorrect:  selectedAnswer.IsCorrect,
		Points:     points,
		AnsweredAt: time.Now(),
	}

	session.Answers = append(session.Answers, userAnswer)
	session.CurrentQuestion++

	// Check if quiz is completed
	isCompleted := session.CurrentQuestion >= len(quizData.Questions)
	var nextQuestion *quiz.Question

	if isCompleted {
		session.Status = quiz.SessionStatusCompleted
		now := time.Now()
		session.CompletedAt = &now
	} else {
		nextQuestion = session.GetNextQuestion(quizData)
	}

	// Update session
	if err := uc.repo.UpdateSession(ctx, session); err != nil {
		return nil, err
	}

	return &SubmitAnswerResult{
		IsCorrect:       selectedAnswer.IsCorrect,
		Points:          points,
		TotalScore:      session.Score,
		IsQuizCompleted: isCompleted,
		NextQuestion:    nextQuestion,
	}, nil
}
