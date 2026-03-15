package marathon

import (
	"time"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/solo_marathon"
)

type ClaimWeeklyRewardsInput struct {
	PlayerID string
}

type ClaimWeeklyRewardsOutput struct {
	Claimed   bool   `json:"claimed"`
	WeekLabel string `json:"weekLabel"`
	Rank      int    `json:"rank"`
	Coins     int    `json:"coins"`
	Tickets   int    `json:"tickets"`
}

// weeklyRewardTiers defines rewards by leaderboard position
var weeklyRewardTiers = []struct {
	maxRank int
	coins   int
	tickets int
}{
	{1, 500, 5},
	{3, 300, 3},
	{10, 200, 2},
	{50, 100, 0},
	{100, 50, 0},
}

func getWeeklyReward(rank int) (coins int, tickets int) {
	for _, tier := range weeklyRewardTiers {
		if rank <= tier.maxRank {
			return tier.coins, tier.tickets
		}
	}
	return 0, 0
}

// getLastWeekBounds returns Monday 00:00 UTC and next Monday 00:00 UTC for the PREVIOUS week
func getLastWeekBounds() (from int64, to int64, label string) {
	now := time.Now().UTC()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	// This Monday
	thisMonday := time.Date(now.Year(), now.Month(), now.Day()-(weekday-1), 0, 0, 0, 0, time.UTC)
	// Last Monday
	lastMonday := thisMonday.AddDate(0, 0, -7)

	label = lastMonday.Format("2006-01-02") + " — " + thisMonday.AddDate(0, 0, -1).Format("2006-01-02")
	return lastMonday.Unix(), thisMonday.Unix(), label
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

	from, to, weekLabel := getLastWeekBounds()

	// Get last week's leaderboard (all categories, top 100)
	category := solo_marathon.NewMarathonCategoryAll()
	topRecords, err := uc.personalBestRepo.FindTopByCategoryInTimeRange(category, 100, from, to)
	if err != nil {
		// No records for last week — nothing to claim
		return ClaimWeeklyRewardsOutput{Claimed: false, WeekLabel: weekLabel}, nil
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
		// Player not in top 100 — no reward
		return ClaimWeeklyRewardsOutput{Claimed: false, WeekLabel: weekLabel}, nil
	}

	coins, tickets := getWeeklyReward(rank)
	if coins == 0 && tickets == 0 {
		return ClaimWeeklyRewardsOutput{Claimed: false, WeekLabel: weekLabel, Rank: rank}, nil
	}

	// Credit rewards via inventory (idempotent — InventoryService handles duplicates via transaction log)
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
		Claimed:   true,
		WeekLabel: weekLabel,
		Rank:      rank,
		Coins:     coins,
		Tickets:   tickets,
	}, nil
}
