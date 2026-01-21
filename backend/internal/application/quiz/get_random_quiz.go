package quiz

import (
	"math/rand"
	"time"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
)

// GetRandomQuizUseCase handles the business logic for getting a random quiz
type GetRandomQuizUseCase struct {
	quizRepo quiz.QuizRepository
	rng      *rand.Rand
}

// NewGetRandomQuizUseCase creates a new GetRandomQuizUseCase
func NewGetRandomQuizUseCase(quizRepo quiz.QuizRepository) *GetRandomQuizUseCase {
	// Initialize random number generator with current time seed
	source := rand.NewSource(time.Now().UnixNano())
	return &GetRandomQuizUseCase{
		quizRepo: quizRepo,
		rng:      rand.New(source),
	}
}

// Execute retrieves a random quiz, optionally filtered by category
func (uc *GetRandomQuizUseCase) Execute(input GetRandomQuizInput) (GetQuizDetailsOutput, error) {
	var summaries []*quiz.QuizSummary
	var err error

	// 1. Get quizzes (filtered by category if provided)
	if input.CategoryID != "" {
		categoryID, err := quiz.NewCategoryIDFromString(input.CategoryID)
		if err != nil {
			return GetQuizDetailsOutput{}, err
		}
		summaries, err = uc.quizRepo.FindSummariesByCategory(categoryID)
	} else {
		summaries, err = uc.quizRepo.FindAllSummaries()
	}

	if err != nil {
		return GetQuizDetailsOutput{}, err
	}

	if len(summaries) == 0 {
		return GetQuizDetailsOutput{}, quiz.ErrQuizNotFound
	}

	// 2. Select random quiz
	randomIndex := uc.rng.Intn(len(summaries))
	selectedQuizID := summaries[randomIndex].ID()

	// 3. Load full quiz details
	quizAggregate, err := uc.quizRepo.FindByID(selectedQuizID)
	if err != nil {
		return GetQuizDetailsOutput{}, err
	}

	// 4. Get top scores (empty for now, can be enhanced)
	topScores := []LeaderboardEntryDTO{}

	// 5. Return DTO
	return GetQuizDetailsOutput{
		Quiz:      ToQuizDetailDTO(quizAggregate),
		TopScores: topScores,
	}, nil
}
