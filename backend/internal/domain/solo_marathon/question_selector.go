package solo_marathon

import (
	"math/rand"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
)

// QuestionSelector is a Domain Service that selects questions for Marathon mode
// Business Logic:
// - Adaptive difficulty based on current streak
// - Weighted random selection (e.g., 80% easy, 20% medium at Beginner level)
// - Excludes recently shown questions to avoid repetition
type QuestionSelector struct {
	questionRepo quiz.QuestionRepository
}

// NewQuestionSelector creates a new QuestionSelector
func NewQuestionSelector(questionRepo quiz.QuestionRepository) *QuestionSelector {
	return &QuestionSelector{
		questionRepo: questionRepo,
	}
}

// SelectNextQuestion selects the next question for Marathon game
// This is the CORE business logic for Marathon question selection
func (qs *QuestionSelector) SelectNextQuestion(
	category MarathonCategory,
	difficulty DifficultyProgression,
	recentIDs []QuestionID, // Last N question IDs to exclude (typically 20)
) (*quiz.Question, error) {
	// 1. Get difficulty distribution for current level
	distribution := difficulty.GetDistribution()
	// Example for Beginner: {"easy": 0.8, "medium": 0.2, "hard": 0.0}
	// Example for Master:   {"easy": 0.0, "medium": 0.3, "hard": 0.7}

	// 2. Select difficulty using weighted random
	selectedDifficulty := selectWeightedDifficulty(distribution)

	// 3. Build filter
	filter := quiz.NewQuestionFilter().
		WithDifficulty(selectedDifficulty).
		WithExcludeIDs(recentIDs)

	// Add category filter if not "all categories"
	if !category.IsAllCategories() {
		filter = filter.WithCategory(category.CategoryID())
	}

	// 4. Verify we have questions available
	count, err := qs.questionRepo.CountByFilter(filter)
	if err != nil {
		return nil, err
	}

	if count == 0 {
		// No questions available with this filter
		// Fallback: try without excluding recent questions
		filter = quiz.NewQuestionFilter().
			WithDifficulty(selectedDifficulty)

		if !category.IsAllCategories() {
			filter = filter.WithCategory(category.CategoryID())
		}

		count, err = qs.questionRepo.CountByFilter(filter)
		if err != nil || count == 0 {
			return nil, ErrNoQuestionsAvailable
		}
	}

	// 5. Fetch random question
	questions, err := qs.questionRepo.FindRandomQuestions(filter, 1)
	if err != nil {
		return nil, err
	}

	if len(questions) == 0 {
		return nil, ErrNoQuestionsAvailable
	}

	return questions[0], nil
}

// selectWeightedDifficulty performs weighted random selection
// Input: {"easy": 0.8, "medium": 0.2, "hard": 0.0}
// Output: "easy" 80% of the time, "medium" 20% of the time
func selectWeightedDifficulty(distribution map[string]float64) string {
	// Build cumulative weights
	type weightedOption struct {
		difficulty     string
		cumulativeProb float64
	}

	options := make([]weightedOption, 0, len(distribution))
	cumulative := 0.0

	for difficulty, weight := range distribution {
		if weight > 0 {
			cumulative += weight
			options = append(options, weightedOption{
				difficulty:     difficulty,
				cumulativeProb: cumulative,
			})
		}
	}

	// Generate random number [0, 1)
	randValue := rand.Float64()

	// Select based on cumulative probability
	for _, option := range options {
		if randValue < option.cumulativeProb {
			return option.difficulty
		}
	}

	// Fallback (should never happen if distribution sums to 1.0)
	if len(options) > 0 {
		return options[len(options)-1].difficulty
	}

	return "medium" // ultimate fallback
}

// ErrNoQuestionsAvailable is returned when no questions match the criteria
var ErrNoQuestionsAvailable = ErrInvalidQuestion
