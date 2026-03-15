package quick_duel

import (
	"fmt"
	"time"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quick_duel"
)

type ResetSeasonInput struct {
	NewSeasonID string
}

type ResetSeasonOutput struct {
	PlayersReset   int              `json:"playersReset"`
	RewardsGranted int              `json:"rewardsGranted"`
	OldSeasonID    string           `json:"oldSeasonId"`
	NewSeasonID    string           `json:"newSeasonId"`
	RewardsSummary []SeasonRewardDTO `json:"rewardsSummary"`
}

type SeasonRewardDTO struct {
	PlayerID string `json:"playerId"`
	PeakLeague string `json:"peakLeague"`
	Coins    int    `json:"coins"`
	Tickets  int    `json:"tickets"`
}

type ResetSeasonUseCase struct {
	playerRatingRepo quick_duel.PlayerRatingRepository
	seasonRepo       quick_duel.SeasonRepository
	inventoryService InventoryService
}

func NewResetSeasonUseCase(
	playerRatingRepo quick_duel.PlayerRatingRepository,
	seasonRepo quick_duel.SeasonRepository,
	inventoryService InventoryService,
) *ResetSeasonUseCase {
	return &ResetSeasonUseCase{
		playerRatingRepo: playerRatingRepo,
		seasonRepo:       seasonRepo,
		inventoryService: inventoryService,
	}
}

func (uc *ResetSeasonUseCase) Execute(input ResetSeasonInput) (ResetSeasonOutput, error) {
	now := time.Now().UTC().Unix()

	// 1. Get current season
	oldSeasonID, err := uc.seasonRepo.GetCurrentSeason()
	if err != nil {
		return ResetSeasonOutput{}, fmt.Errorf("reset season: get current season: %w", err)
	}

	newSeasonID := input.NewSeasonID
	if newSeasonID == "" {
		// Auto-generate: YYYY-MM format
		newSeasonID = time.Now().UTC().Format("2006-01")
	}

	// 2. Fetch all players in current season (batch of 1000)
	allRatings, err := uc.playerRatingRepo.GetLeaderboard(oldSeasonID, 10000, 0)
	if err != nil {
		return ResetSeasonOutput{}, fmt.Errorf("reset season: get leaderboard: %w", err)
	}

	// 3. Grant rewards and reset each player
	var rewards []SeasonRewardDTO
	rewardsGranted := 0

	for _, rating := range allRatings {
		// Credit season rewards based on peak league
		coins := rating.GetSeasonRewardCoins()
		tickets := rating.GetSeasonRewardTickets()

		if uc.inventoryService != nil && (coins > 0 || tickets > 0) {
			rewardDetails := map[string]int{}
			if coins > 0 {
				rewardDetails["coins"] = coins
			}
			if tickets > 0 {
				rewardDetails["pvp_tickets"] = tickets
			}
			_ = uc.inventoryService.Credit(
				rating.PlayerID().String(),
				"pvp_season_reward",
				rewardDetails,
			)
			rewardsGranted++

			rewards = append(rewards, SeasonRewardDTO{
				PlayerID:   rating.PlayerID().String(),
				PeakLeague: rating.PeakLeague().String(),
				Coins:      coins,
				Tickets:    tickets,
			})
		}

		// Apply soft reset
		rating.SeasonReset(newSeasonID, now)

		// Save updated rating
		if err := uc.playerRatingRepo.Save(rating); err != nil {
			return ResetSeasonOutput{}, fmt.Errorf("reset season: save rating for %s: %w", rating.PlayerID().String(), err)
		}
	}

	// 4. Create new season record
	monthStart := time.Date(time.Now().UTC().Year(), time.Now().UTC().Month(), 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Second)
	_ = uc.seasonRepo.CreateSeason(newSeasonID, monthStart.Unix(), monthEnd.Unix())

	return ResetSeasonOutput{
		PlayersReset:   len(allRatings),
		RewardsGranted: rewardsGranted,
		OldSeasonID:    oldSeasonID,
		NewSeasonID:    newSeasonID,
		RewardsSummary: rewards,
	}, nil
}
