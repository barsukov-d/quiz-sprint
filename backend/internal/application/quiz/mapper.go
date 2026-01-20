package quiz

import (
	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
)

// ========================================
// Domain → DTO Mappers
// ========================================

// ToQuizDTO converts a Quiz aggregate to QuizDTO
func ToQuizDTO(q *quiz.Quiz) QuizDTO {
	var categoryID string
	if !q.CategoryID().IsZero() {
		categoryID = q.CategoryID().String()
	}

	return QuizDTO{
		ID:             q.ID().String(),
		Title:          q.Title().String(),
		Description:    q.Description(),
		CategoryID:     categoryID,
		QuestionsCount: q.QuestionsCount(),
		TimeLimit:      q.TimeLimit().Seconds(),
		PassingScore:   q.PassingScore().Percentage(),
		CreatedAt:      q.CreatedAt(),
	}
}

// ToQuizDetailDTO converts a Quiz aggregate to QuizDetailDTO with questions
func ToQuizDetailDTO(q *quiz.Quiz) QuizDetailDTO {
	var categoryID string
	if !q.CategoryID().IsZero() {
		categoryID = q.CategoryID().String()
	}

	questions := make([]QuestionDTO, 0, q.QuestionsCount())
	for _, question := range q.Questions() {
		questions = append(questions, ToQuestionDTO(&question))
	}

	return QuizDetailDTO{
		ID:           q.ID().String(),
		Title:        q.Title().String(),
		Description:  q.Description(),
		CategoryID:   categoryID,
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

// ToCategoryDTO converts a Category aggregate to a DTO.
func ToCategoryDTO(c *quiz.Category) CategoryDTO {
	return CategoryDTO{
		ID:   c.ID().String(),
		Name: c.Name().String(),
	}
}

// ToCategoryDTOs converts a slice of Category aggregates to DTOs.
func ToCategoryDTOs(categories []*quiz.Category) []CategoryDTO {
	dtos := make([]CategoryDTO, 0, len(categories))
	for _, c := range categories {
		dtos = append(dtos, ToCategoryDTO(c))
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

// ToQuizDTOFromSummary converts a QuizSummary to QuizDTO
func ToQuizDTOFromSummary(s *quiz.QuizSummary) QuizDTO {
	var categoryID string
	if !s.CategoryID().IsZero() {
		categoryID = s.CategoryID().String()
	}

	return QuizDTO{
		ID:             s.ID().String(),
		Title:          s.Title().String(),
		Description:    s.Description(),
		CategoryID:     categoryID,
		QuestionsCount: s.QuestionCount(),
		TimeLimit:      s.TimeLimit().Seconds(),
		PassingScore:   s.PassingScore().Percentage(),
		CreatedAt:      s.CreatedAt(),
	}
}

// ToQuizListDTOFromSummaries converts a slice of QuizSummary objects to DTOs
func ToQuizListDTOFromSummaries(summaries []*quiz.QuizSummary) []QuizDTO {
	dtos := make([]QuizDTO, 0, len(summaries))
	for _, s := range summaries {
		dtos = append(dtos, ToQuizDTOFromSummary(s))
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
