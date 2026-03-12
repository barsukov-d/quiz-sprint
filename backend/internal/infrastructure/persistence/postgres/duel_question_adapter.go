package postgres

import (
	appDuel "github.com/barsukov/quiz-sprint/backend/internal/application/quick_duel"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
)

// duelQuestionSource is the subset of quiz.QuestionRepository used by the duel adapter.
type duelQuestionSource interface {
	FindRandomQuestions(filter quiz.QuestionFilter, limit int) ([]*quiz.Question, error)
	FindByID(id quiz.QuestionID) (*quiz.Question, error)
}

// DuelQuestionRepositoryAdapter adapts quiz.QuestionRepository to appDuel.QuestionRepository.
type DuelQuestionRepositoryAdapter struct {
	repo duelQuestionSource
}

func NewDuelQuestionRepositoryAdapter(repo duelQuestionSource) *DuelQuestionRepositoryAdapter {
	return &DuelQuestionRepositoryAdapter{repo: repo}
}

func (a *DuelQuestionRepositoryAdapter) FindRandomByDifficulty(count int, difficulty string) ([]appDuel.QuestionData, error) {
	filter := quiz.NewQuestionFilter().WithDifficulty(difficulty)
	questions, err := a.repo.FindRandomQuestions(filter, count)
	if err != nil {
		return nil, err
	}

	result := make([]appDuel.QuestionData, 0, len(questions))
	for _, q := range questions {
		answers := make([]appDuel.AnswerData, 0, len(q.Answers()))
		for _, ans := range q.Answers() {
			answers = append(answers, appDuel.AnswerData{
				ID:        ans.ID().String(),
				Text:      ans.Text().String(),
				IsCorrect: ans.IsCorrect(),
			})
		}
		result = append(result, appDuel.QuestionData{
			ID:      q.ID().String(),
			Text:    q.Text().String(),
			Answers: answers,
		})
	}
	return result, nil
}

func (a *DuelQuestionRepositoryAdapter) FindByID(questionID quiz.QuestionID) (*quiz.Question, error) {
	return a.repo.FindByID(questionID)
}
