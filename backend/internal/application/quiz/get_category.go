package quiz

import (
	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
)

// ListCategoriesUseCase handles the business logic for listing categories.
type ListCategoriesUseCase struct {
	categoryRepo quiz.CategoryRepository
}

// NewListCategoriesUseCase creates a new ListCategoriesUseCase.
func NewListCategoriesUseCase(categoryRepo quiz.CategoryRepository) *ListCategoriesUseCase {
	return &ListCategoriesUseCase{
		categoryRepo: categoryRepo,
	}
}

// ListCategoriesInput is the input for the use case.
type ListCategoriesInput struct{}

// ListCategoriesOutput is the output for the use case.
type ListCategoriesOutput struct {
	Categories []CategoryDTO
}

// Execute retrieves all categories.
func (uc *ListCategoriesUseCase) Execute(input ListCategoriesInput) (ListCategoriesOutput, error) {
	categories, err := uc.categoryRepo.FindAll()
	if err != nil {
		return ListCategoriesOutput{}, err
	}

	return ListCategoriesOutput{
		Categories: ToCategoryDTOs(categories),
	}, nil
}
