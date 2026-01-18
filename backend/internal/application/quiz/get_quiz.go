package quiz

import (
	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
)

// GetQuizUseCase handles the business logic for getting a quiz
type GetQuizUseCase struct {
	quizRepo quiz.QuizRepository
}

// NewGetQuizUseCase creates a new GetQuizUseCase
func NewGetQuizUseCase(quizRepo quiz.QuizRepository) *GetQuizUseCase {
	return &GetQuizUseCase{
		quizRepo: quizRepo,
	}
}

// Execute retrieves a quiz by ID
func (uc *GetQuizUseCase) Execute(input GetQuizInput) (GetQuizOutput, error) {
	// 1. Validate input
	quizID, err := quiz.NewQuizIDFromString(input.QuizID)
	if err != nil {
		return GetQuizOutput{}, err
	}

	// 2. Load quiz aggregate
	quizAggregate, err := uc.quizRepo.FindByID(quizID)
	if err != nil {
		return GetQuizOutput{}, err
	}

	// 3. Return DTO
	return GetQuizOutput{
		Quiz: ToQuizDTO(quizAggregate),
	}, nil
}

// ListQuizzesUseCase handles the business logic for listing quizzes
type ListQuizzesUseCase struct {
	quizRepo quiz.QuizRepository
}

// NewListQuizzesUseCase creates a new ListQuizzesUseCase
func NewListQuizzesUseCase(quizRepo quiz.QuizRepository) *ListQuizzesUseCase {
	return &ListQuizzesUseCase{
		quizRepo: quizRepo,
	}
}

// Execute retrieves all quizzes
func (uc *ListQuizzesUseCase) Execute(input ListQuizzesInput) (ListQuizzesOutput, error) {
	// 1. Load all quiz summaries
	summaries, err := uc.quizRepo.FindAllSummaries()
	if err != nil {
		return ListQuizzesOutput{}, err
	}

	// 2. Return DTOs
	return ListQuizzesOutput{
		Quizzes: ToQuizListDTOFromSummaries(summaries),
	}, nil
}
