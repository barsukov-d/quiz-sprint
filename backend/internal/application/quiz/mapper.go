package quiz

import (
	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
)

// ========================================
// Domain → DTO Mappers
// ========================================

// ToQuizDTO converts a Quiz aggregate to QuizDTO
func ToQuizDTO(q *quiz.Quiz) QuizDTO {
	return QuizDTO{
		ID:             q.ID().String(),
		Title:          q.Title().String(),
		Description:    q.Description(),
		QuestionsCount: q.QuestionsCount(),
		TimeLimit:      q.TimeLimit().Seconds(),
		PassingScore:   q.PassingScore().Percentage(),
		CreatedAt:      q.CreatedAt(),
	}
}

// ToQuizDetailDTO converts a Quiz aggregate to QuizDetailDTO with questions
func ToQuizDetailDTO(q *quiz.Quiz) QuizDetailDTO {
	questions := make([]QuestionDTO, 0, q.QuestionsCount())
	for _, question := range q.Questions() {
		questions = append(questions, ToQuestionDTO(&question))
	}

	return QuizDetailDTO{
		ID:           q.ID().String(),
		Title:        q.Title().String(),
		Description:  q.Description(),
		Questions:    questions,
		TimeLimit:    q.TimeLimit().Seconds(),
		PassingScore: q.PassingScore().Percentage(),
		CreatedAt:    q.CreatedAt(),
	}
}

// ToQuestionDTO converts a Question entity to QuestionDTO
// NOTE: Does NOT include IsCorrect - never leak correct answers!
func ToQuestionDTO(q *quiz.Question) QuestionDTO {
	answers := make([]AnswerDTO, 0, len(q.Answers()))
	for _, answer := range q.Answers() {
		answers = append(answers, ToAnswerDTO(&answer))
	}

	return QuestionDTO{
		ID:       q.ID().String(),
		Text:     q.Text().String(),
		Answers:  answers,
		Points:   q.Points().Value(),
		Position: q.Position(),
	}
}

// ToAnswerDTO converts an Answer entity to AnswerDTO
// NOTE: IsCorrect is intentionally NOT included!
func ToAnswerDTO(a *quiz.Answer) AnswerDTO {
	return AnswerDTO{
		ID:       a.ID().String(),
		Text:     a.Text().String(),
		Position: a.Position(),
	}
}

// ToSessionDTO converts a QuizSession aggregate to SessionDTO
func ToSessionDTO(s *quiz.QuizSession) SessionDTO {
	return SessionDTO{
		ID:              s.ID().String(),
		QuizID:          s.QuizID().String(),
		UserID:          s.UserID().String(),
		CurrentQuestion: s.CurrentQuestion(),
		Score:           s.Score().Value(),
		Status:          s.Status().String(),
		StartedAt:       s.StartedAt(),
		CompletedAt:     s.CompletedAt(),
	}
}

// ToLeaderboardEntryDTO converts a LeaderboardEntry to DTO
func ToLeaderboardEntryDTO(e quiz.LeaderboardEntry) LeaderboardEntryDTO {
	return LeaderboardEntryDTO{
		UserID:      e.UserID().String(),
		Username:    e.Username(),
		Score:       e.Score().Value(),
		Rank:        e.Rank(),
		CompletedAt: e.CompletedAt(),
	}
}

// ToLeaderboardEntriesDTO converts a slice of LeaderboardEntry to DTOs
func ToLeaderboardEntriesDTO(entries []quiz.LeaderboardEntry) []LeaderboardEntryDTO {
	dtos := make([]LeaderboardEntryDTO, 0, len(entries))
	for _, entry := range entries {
		dtos = append(dtos, ToLeaderboardEntryDTO(entry))
	}
	return dtos
}

// ToQuizListDTO converts a slice of Quiz aggregates to DTOs
func ToQuizListDTO(quizzes []quiz.Quiz) []QuizDTO {
	dtos := make([]QuizDTO, 0, len(quizzes))
	for _, q := range quizzes {
		dtos = append(dtos, ToQuizDTO(&q))
	}
	return dtos
}

// ========================================
// Result Builders
// ========================================

// BuildFinalResult creates a FinalResultDTO from session and quiz
func BuildFinalResult(session *quiz.QuizSession, quizAggregate *quiz.Quiz) FinalResultDTO {
	totalScore := session.Score().Value()
	maxScore := quizAggregate.GetTotalPoints().Value()

	percentage := 0
	if maxScore > 0 {
		percentage = (totalScore * 100) / maxScore
	}

	correctCount := 0
	for _, answer := range session.Answers() {
		if answer.IsCorrect() {
			correctCount++
		}
	}

	return FinalResultDTO{
		TotalScore:     totalScore,
		MaxScore:       maxScore,
		Percentage:     percentage,
		Passed:         session.HasPassed(quizAggregate),
		QuestionsCount: quizAggregate.QuestionsCount(),
		CorrectCount:   correctCount,
	}
}

// FindCorrectAnswerID finds the correct answer ID for a question
func FindCorrectAnswerID(q *quiz.Question) string {
	for _, answer := range q.Answers() {
		if answer.IsCorrect() {
			return answer.ID().String()
		}
	}
	return ""
}
