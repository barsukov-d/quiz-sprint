package quiz

import (
	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
)

// CreateCategoryUseCase handles the business logic for creating a category.
type CreateCategoryUseCase struct {
	categoryRepo quiz.CategoryRepository
}

// NewCreateCategoryUseCase creates a new CreateCategoryUseCase.
func NewCreateCategoryUseCase(categoryRepo quiz.CategoryRepository) *CreateCategoryUseCase {
	return &CreateCategoryUseCase{
		categoryRepo: categoryRepo,
	}
}

// Execute creates a new category.
func (uc *CreateCategoryUseCase) Execute(input CreateCategoryInput) (CreateCategoryOutput, error) {
	name, err := quiz.NewCategoryName(input.Name)
	if err != nil {
		return CreateCategoryOutput{}, err
	}

	category, err := quiz.NewCategory(quiz.NewCategoryID(), name)
	if err != nil {
		return CreateCategoryOutput{}, err
	}

	if err := uc.categoryRepo.Save(category); err != nil {
		return CreateCategoryOutput{}, err
	}

	return CreateCategoryOutput{
		Category: ToCategoryDTO(category),
	}, nil
}
