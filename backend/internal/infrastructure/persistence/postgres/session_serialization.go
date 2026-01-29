package postgres

import (
	"encoding/json"
	"fmt"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/kernel"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
)

// ========================================
// Session Serialization for JSONB Storage
// ========================================
// Used by Daily Challenge and other game modes that store
// kernel.QuizGameplaySession in a single JSONB column

// SerializedSession represents a QuizGameplaySession in JSON format
type SerializedSession struct {
	SessionID            string                       `json:"session_id"`
	QuizID               string                       `json:"quiz_id"`
	UserAnswers          map[string]SerializedAnswer  `json:"user_answers"`
	CurrentQuestionIndex int                          `json:"current_question_index"`
	BaseScore            int                          `json:"base_score"`
	StartedAt            int64                        `json:"started_at"`
	FinishedAt           int64                        `json:"finished_at"`
}

// SerializedAnswer represents AnswerData in JSON format
type SerializedAnswer struct {
	AnswerID   string `json:"answer_id"`
	IsCorrect  bool   `json:"is_correct"`
	TimeTaken  int64  `json:"time_taken"`
	AnsweredAt int64  `json:"answered_at"`
}

// serializeSession converts kernel.QuizGameplaySession to JSON bytes
func serializeSession(session *kernel.QuizGameplaySession) ([]byte, error) {
	if session == nil {
		return nil, fmt.Errorf("session is nil")
	}

	// Convert userAnswers map
	userAnswers := make(map[string]SerializedAnswer)
	for questionID, answerData := range session.GetAllAnswers() {
		userAnswers[questionID.String()] = SerializedAnswer{
			AnswerID:   answerData.AnswerID().String(),
			IsCorrect:  answerData.IsCorrect(),
			TimeTaken:  answerData.TimeTaken(),
			AnsweredAt: answerData.AnsweredAt(),
		}
	}

	serialized := SerializedSession{
		SessionID:            session.ID().String(),
		QuizID:               session.Quiz().ID().String(),
		UserAnswers:          userAnswers,
		CurrentQuestionIndex: session.CurrentQuestionIndex(),
		BaseScore:            session.BaseScore().Value(),
		StartedAt:            session.StartedAt(),
		FinishedAt:           session.FinishedAt(),
	}

	return json.Marshal(serialized)
}

// deserializeSession reconstructs kernel.QuizGameplaySession from JSON bytes
func deserializeSession(
	data []byte,
	quizRepo quiz.QuizRepository,
	questionRepo quiz.QuestionRepository,
) (*kernel.QuizGameplaySession, error) {
	var serialized SerializedSession
	if err := json.Unmarshal(data, &serialized); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	// Load quiz from repository
	quizID, err := quiz.NewQuizIDFromString(serialized.QuizID)
	if err != nil {
		return nil, fmt.Errorf("invalid quiz_id: %w", err)
	}

	quizAggregate, err := quizRepo.FindByID(quizID)
	if err != nil {
		return nil, fmt.Errorf("failed to load quiz: %w", err)
	}

	// Reconstruct userAnswers map
	userAnswers := make(map[kernel.QuestionID]kernel.AnswerData)
	for questionIDStr, serializedAnswer := range serialized.UserAnswers {
		questionID, err := quiz.NewQuestionIDFromString(questionIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid question_id: %w", err)
		}

		answerID, err := quiz.NewAnswerIDFromString(serializedAnswer.AnswerID)
		if err != nil {
			return nil, fmt.Errorf("invalid answer_id: %w", err)
		}

		userAnswers[questionID] = kernel.NewAnswerData(
			answerID,
			serializedAnswer.IsCorrect,
			serializedAnswer.TimeTaken,
			serializedAnswer.AnsweredAt,
		)
	}

	// Parse session ID
	sessionID, err := kernel.NewSessionIDFromString(serialized.SessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session_id: %w", err)
	}

	// Parse base score
	baseScore, err := quiz.NewPoints(serialized.BaseScore)
	if err != nil {
		return nil, fmt.Errorf("invalid base_score: %w", err)
	}

	// Reconstruct session
	return kernel.ReconstructQuizGameplaySession(
		sessionID,
		quizAggregate,
		userAnswers,
		serialized.CurrentQuestionIndex,
		baseScore,
		serialized.StartedAt,
		serialized.FinishedAt,
	), nil
}
