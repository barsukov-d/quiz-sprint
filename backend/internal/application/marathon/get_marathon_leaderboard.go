package marathon

import (
	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/solo_marathon"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/user"
)

// GetMarathonLeaderboardUseCase handles retrieving marathon leaderboard
type GetMarathonLeaderboardUseCase struct {
	personalBestRepo solo_marathon.PersonalBestRepository
	categoryRepo     quiz.CategoryRepository
	userRepo         user.UserRepository
}

// NewGetMarathonLeaderboardUseCase creates a new GetMarathonLeaderboardUseCase
func NewGetMarathonLeaderboardUseCase(
	personalBestRepo solo_marathon.PersonalBestRepository,
	categoryRepo quiz.CategoryRepository,
	userRepo user.UserRepository,
) *GetMarathonLeaderboardUseCase {
	return &GetMarathonLeaderboardUseCase{
		personalBestRepo: personalBestRepo,
		categoryRepo:     categoryRepo,
		userRepo:         userRepo,
	}
}

// Execute retrieves the marathon leaderboard for a category
func (uc *GetMarathonLeaderboardUseCase) Execute(input GetMarathonLeaderboardInput) (GetMarathonLeaderboardOutput, error) {
	// 1. Validate limit
	if input.Limit <= 0 {
		input.Limit = 10 // Default limit
	}
	if input.Limit > 100 {
		input.Limit = 100 // Max limit
	}

	// 2. Determine category
	var category solo_marathon.MarathonCategory
	if input.CategoryID == "" || input.CategoryID == "all" {
		// "All categories" mode
		category = solo_marathon.NewMarathonCategoryAll()
	} else {
		// Specific category
		categoryID, err := quiz.NewCategoryIDFromString(input.CategoryID)
		if err != nil {
			return GetMarathonLeaderboardOutput{}, err
		}

		// Validate category exists
		categoryAggregate, err := uc.categoryRepo.FindByID(categoryID)
		if err != nil {
			return GetMarathonLeaderboardOutput{}, err
		}

		category = solo_marathon.NewMarathonCategory(categoryID, categoryAggregate.Name().String())
	}

	// 3. Validate timeFrame
	timeFrame := input.TimeFrame
	if timeFrame == "" {
		timeFrame = "all_time" // Default
	}

	// TODO: Implement time filtering for weekly/daily leaderboards
	// For now, we only support "all_time"
	if timeFrame != "all_time" {
		// Return empty for now
		// TODO: Filter personal bests by achievedAt timestamp
	}

	// 4. Get top personal bests for category
	topRecords, err := uc.personalBestRepo.FindTopByCategory(category, input.Limit)
	if err != nil {
		if err == solo_marathon.ErrPersonalBestNotFound {
			// No records yet - return empty list
			return GetMarathonLeaderboardOutput{
				Category:  ToCategoryDTO(category),
				TimeFrame: timeFrame,
				Entries:   []LeaderboardEntryDTO{},
			}, nil
		}
		return GetMarathonLeaderboardOutput{}, err
	}

	// 5. Build leaderboard entries with usernames
	entries := make([]LeaderboardEntryDTO, 0, len(topRecords))
	for rank, record := range topRecords {
		// Get username from user repository
		username := "Unknown" // Default
		userAggregate, err := uc.userRepo.FindByID(record.PlayerID())
		if err == nil {
			username = userAggregate.Username().String()
		}

		entry := ToLeaderboardEntryDTO(record, username, rank+1)
		entries = append(entries, entry)
	}

	// 6. Build output
	return GetMarathonLeaderboardOutput{
		Category:  ToCategoryDTO(category),
		TimeFrame: timeFrame,
		Entries:   entries,
	}, nil
}
