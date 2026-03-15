package marathon

import (
	"fmt"
	"time"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/solo_marathon"
)

// WeeklyRewardDistributionRepository tracks which weeks have already been distributed
// to prevent double-distribution.
type WeeklyRewardDistributionRepository interface {
	// HasDistributed returns true if rewards for the given weekID have been distributed.
	HasDistributed(weekID string) (bool, error)
	// MarkDistributed records that rewards for the given weekID have been distributed.
	MarkDistributed(weekID string) error
}

// weeklyRewardTiers defines rewards by leaderboard position
var weeklyRewardTiers = []struct {
	maxRank int
	coins   int
	tickets int
}{
	{3, 5000, 50},
	{10, 3000, 30},
	{25, 2000, 20},
	{50, 1000, 10},
	{100, 500, 5},
}

func getWeeklyReward(rank int) (coins int, tickets int) {
	for _, tier := range weeklyRewardTiers {
		if rank <= tier.maxRank {
			return tier.coins, tier.tickets
		}
	}
	return 0, 0
}

// getLastWeekBounds returns Monday 00:00 UTC and next Monday 00:00 UTC for the PREVIOUS week,
// along with the ISO week ID (e.g. "2026-W11").
func getLastWeekBounds() (from int64, to int64, weekID string) {
	now := time.Now().UTC()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	// This Monday 00:00 UTC
	thisMonday := time.Date(now.Year(), now.Month(), now.Day()-(weekday-1), 0, 0, 0, 0, time.UTC)
	// Last Monday 00:00 UTC
	lastMonday := thisMonday.AddDate(0, 0, -7)

	year, week := lastMonday.ISOWeek()
	weekID = fmt.Sprintf("%d-W%02d", year, week)
	return lastMonday.Unix(), thisMonday.Unix(), weekID
}

// ========================================
// ClaimWeeklyRewardsUseCase (player-pull — kept for backwards compatibility)
// ========================================

type ClaimWeeklyRewardsInput struct {
	PlayerID string
}

type ClaimWeeklyRewardsOutput struct {
	Claimed   bool   `json:"claimed"`
	WeekID    string `json:"weekId"`
	Rank      int    `json:"rank"`
	Coins     int    `json:"coins"`
	Tickets   int    `json:"tickets"`
}

type ClaimWeeklyRewardsUseCase struct {
	personalBestRepo solo_marathon.PersonalBestRepository
	inventoryService InventoryService
}

func NewClaimWeeklyRewardsUseCase(
	personalBestRepo solo_marathon.PersonalBestRepository,
	inventoryService InventoryService,
) *ClaimWeeklyRewardsUseCase {
	return &ClaimWeeklyRewardsUseCase{
		personalBestRepo: personalBestRepo,
		inventoryService: inventoryService,
	}
}

func (uc *ClaimWeeklyRewardsUseCase) Execute(input ClaimWeeklyRewardsInput) (ClaimWeeklyRewardsOutput, error) {
	playerID, err := shared.NewUserID(input.PlayerID)
	if err != nil {
		return ClaimWeeklyRewardsOutput{}, err
	}

	from, to, weekID := getLastWeekBounds()

	// Get last week's leaderboard (all categories, top 100)
	category := solo_marathon.NewMarathonCategoryAll()
	topRecords, err := uc.personalBestRepo.FindTopByCategoryInTimeRange(category, 100, from, to)
	if err != nil {
		return ClaimWeeklyRewardsOutput{Claimed: false, WeekID: weekID}, nil
	}

	// Find player's rank
	rank := 0
	for i, record := range topRecords {
		if record.PlayerID().Equals(playerID) {
			rank = i + 1
			break
		}
	}

	if rank == 0 {
		return ClaimWeeklyRewardsOutput{Claimed: false, WeekID: weekID}, nil
	}

	coins, tickets := getWeeklyReward(rank)
	if coins == 0 && tickets == 0 {
		return ClaimWeeklyRewardsOutput{Claimed: false, WeekID: weekID, Rank: rank}, nil
	}

	if uc.inventoryService != nil {
		rewards := map[string]int{}
		if coins > 0 {
			rewards["coins"] = coins
		}
		if tickets > 0 {
			rewards["pvp_tickets"] = tickets
		}
		_ = uc.inventoryService.Credit(input.PlayerID, "marathon_weekly_reward", rewards)
	}

	return ClaimWeeklyRewardsOutput{
		Claimed: true,
		WeekID:  weekID,
		Rank:    rank,
		Coins:   coins,
		Tickets: tickets,
	}, nil
}

// ========================================
// DistributeWeeklyMarathonRewardsUseCase (cron/background job)
// ========================================

// DistributeWeeklyMarathonRewardsInput is the input for distributing weekly rewards.
// Typically called by a cron job after the week ends (Monday 00:00 UTC).
type DistributeWeeklyMarathonRewardsInput struct{}

// DistributeWeeklyMarathonRewardsOutput summarises the distribution run.
type DistributeWeeklyMarathonRewardsOutput struct {
	WeekID      string `json:"weekId"`
	Distributed int    `json:"distributed"` // number of players credited
	Skipped     bool   `json:"skipped"`     // true if week already distributed
}

// DistributeWeeklyMarathonRewardsUseCase distributes leaderboard rewards
// to the top 100 players for the previous week.
type DistributeWeeklyMarathonRewardsUseCase struct {
	personalBestRepo solo_marathon.PersonalBestRepository
	inventoryService InventoryService
	distributionRepo WeeklyRewardDistributionRepository
}

// NewDistributeWeeklyMarathonRewardsUseCase creates the use case.
func NewDistributeWeeklyMarathonRewardsUseCase(
	personalBestRepo solo_marathon.PersonalBestRepository,
	inventoryService InventoryService,
	distributionRepo WeeklyRewardDistributionRepository,
) *DistributeWeeklyMarathonRewardsUseCase {
	return &DistributeWeeklyMarathonRewardsUseCase{
		personalBestRepo: personalBestRepo,
		inventoryService: inventoryService,
		distributionRepo: distributionRepo,
	}
}

// Execute distributes weekly marathon leaderboard rewards.
// It is idempotent: if rewards for the week have already been distributed, it returns early.
func (uc *DistributeWeeklyMarathonRewardsUseCase) Execute(_ DistributeWeeklyMarathonRewardsInput) (DistributeWeeklyMarathonRewardsOutput, error) {
	from, to, weekID := getLastWeekBounds()

	// Guard: prevent double-distribution
	if uc.distributionRepo != nil {
		already, err := uc.distributionRepo.HasDistributed(weekID)
		if err != nil {
			return DistributeWeeklyMarathonRewardsOutput{}, fmt.Errorf("check distribution status: %w", err)
		}
		if already {
			return DistributeWeeklyMarathonRewardsOutput{WeekID: weekID, Skipped: true}, nil
		}
	}

	// Fetch top 100 for the previous week (all categories)
	category := solo_marathon.NewMarathonCategoryAll()
	topRecords, err := uc.personalBestRepo.FindTopByCategoryInTimeRange(category, 100, from, to)
	if err != nil {
		// No records for last week — nothing to distribute, still mark as done
		if markErr := uc.markDistributed(weekID); markErr != nil {
			return DistributeWeeklyMarathonRewardsOutput{}, markErr
		}
		return DistributeWeeklyMarathonRewardsOutput{WeekID: weekID, Distributed: 0}, nil
	}

	// Credit each eligible player
	distributed := 0
	for i, record := range topRecords {
		rank := i + 1
		coins, tickets := getWeeklyReward(rank)
		if coins == 0 && tickets == 0 {
			continue
		}

		if uc.inventoryService == nil {
			continue
		}

		rewards := map[string]int{}
		if coins > 0 {
			rewards["coins"] = coins
		}
		if tickets > 0 {
			rewards["pvp_tickets"] = tickets
		}

		// Source includes weekID for audit trail
		source := fmt.Sprintf("marathon_weekly_reward_%s", weekID)
		if err := uc.inventoryService.Credit(record.PlayerID().String(), source, rewards); err != nil {
			// Log and continue — partial distribution is better than no distribution
			continue
		}
		distributed++
	}

	// Mark week as distributed (idempotency guard for next run)
	if err := uc.markDistributed(weekID); err != nil {
		return DistributeWeeklyMarathonRewardsOutput{}, err
	}

	return DistributeWeeklyMarathonRewardsOutput{
		WeekID:      weekID,
		Distributed: distributed,
	}, nil
}

func (uc *DistributeWeeklyMarathonRewardsUseCase) markDistributed(weekID string) error {
	if uc.distributionRepo == nil {
		return nil
	}
	if err := uc.distributionRepo.MarkDistributed(weekID); err != nil {
		return fmt.Errorf("mark distribution complete: %w", err)
	}
	return nil
}
