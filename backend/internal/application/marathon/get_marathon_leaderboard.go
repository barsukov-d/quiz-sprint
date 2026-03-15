package marathon

import (
	"time"

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

	// 4. Get top personal bests for category (with optional time filtering)
	var topRecords []*solo_marathon.PersonalBest
	var err error
	if timeFrame == "weekly" {
		// Calculate current week boundaries (Monday 00:00 UTC - Sunday 23:59 UTC)
		now := time.Now().UTC()
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7 // Sunday = 7
		}
		monday := now.AddDate(0, 0, -(weekday - 1))
		monday = time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, time.UTC)
		nextMonday := monday.AddDate(0, 0, 7)

		topRecords, err = uc.personalBestRepo.FindTopByCategoryInTimeRange(
			category, input.Limit, monday.Unix(), nextMonday.Unix(),
		)
	} else {
		topRecords, err = uc.personalBestRepo.FindTopByCategory(category, input.Limit)
	}
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
