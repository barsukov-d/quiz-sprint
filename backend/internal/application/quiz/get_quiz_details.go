package quiz

import (
	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
)

// GetQuizDetailsUseCase handles the business logic for getting quiz details with questions
// NOTE: This returns questions and answers but NOT the correct answers (security)
type GetQuizDetailsUseCase struct {
	quizRepo            quiz.QuizRepository
	leaderboardRepo     quiz.LeaderboardRepository
}

// NewGetQuizDetailsUseCase creates a new GetQuizDetailsUseCase
func NewGetQuizDetailsUseCase(
	quizRepo quiz.QuizRepository,
	leaderboardRepo quiz.LeaderboardRepository,
) *GetQuizDetailsUseCase {
	return &GetQuizDetailsUseCase{
		quizRepo:        quizRepo,
		leaderboardRepo: leaderboardRepo,
	}
}

// Execute retrieves quiz details including questions (but not correct answers)
// and top leaderboard entries
func (uc *GetQuizDetailsUseCase) Execute(input GetQuizDetailsInput) (GetQuizDetailsOutput, error) {
	// 1. Validate input
	quizID, err := quiz.NewQuizIDFromString(input.QuizID)
	if err != nil {
		return GetQuizDetailsOutput{}, err
	}

	// 2. Load quiz aggregate
	quizAggregate, err := uc.quizRepo.FindByID(quizID)
	if err != nil {
		return GetQuizDetailsOutput{}, err
	}

	// 3. Get top leaderboard entries (top 3 for preview)
	topScores, err := uc.leaderboardRepo.GetLeaderboard(quizID, 3)
	if err != nil {
		// Leaderboard might be empty for new quizzes, that's ok
		topScores = []quiz.LeaderboardEntry{}
	}

	// 4. Convert to DTO (with questions but WITHOUT correct answers)
	return GetQuizDetailsOutput{
		Quiz:      ToQuizDetailDTO(quizAggregate),
		TopScores: ToLeaderboardEntriesDTO(topScores),
	}, nil
}
