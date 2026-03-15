package quick_duel

import (
	"fmt"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quick_duel"
)

// DistributeSeasonalRewardsInput holds parameters for reward distribution.
type DistributeSeasonalRewardsInput struct {
	// SeasonID is the season whose players receive rewards.
	// If empty, the current season is used.
	SeasonID string
}

// DistributeSeasonalRewardsOutput describes the result of reward distribution.
type DistributeSeasonalRewardsOutput struct {
	SeasonID       string             `json:"seasonId"`
	RewardsGranted int                `json:"rewardsGranted"`
	Rewards        []SeasonRewardDTO  `json:"rewards"`
}

// DistributeSeasonalRewardsUseCase credits end-of-season rewards to all players
// based on their peak league, then triggers a soft MMR reset.
type DistributeSeasonalRewardsUseCase struct {
	playerRatingRepo quick_duel.PlayerRatingRepository
	seasonRepo       quick_duel.SeasonRepository
	inventoryService InventoryService
	seasonalReset    *SeasonalResetUseCase
}

func NewDistributeSeasonalRewardsUseCase(
	playerRatingRepo quick_duel.PlayerRatingRepository,
	seasonRepo quick_duel.SeasonRepository,
	inventoryService InventoryService,
	seasonalReset *SeasonalResetUseCase,
) *DistributeSeasonalRewardsUseCase {
	return &DistributeSeasonalRewardsUseCase{
		playerRatingRepo: playerRatingRepo,
		seasonRepo:       seasonRepo,
		inventoryService: inventoryService,
		seasonalReset:    seasonalReset,
	}
}

// Execute distributes rewards to all players for the ending season, then runs the soft reset.
func (uc *DistributeSeasonalRewardsUseCase) Execute(input DistributeSeasonalRewardsInput) (DistributeSeasonalRewardsOutput, error) {
	seasonID := input.SeasonID
	if seasonID == "" {
		current, err := uc.seasonRepo.GetCurrentSeason()
		if err != nil {
			return DistributeSeasonalRewardsOutput{}, fmt.Errorf("distribute rewards: get current season: %w", err)
		}
		seasonID = current
	}

	ratings, err := uc.playerRatingRepo.FindAllBySeasonID(seasonID)
	if err != nil {
		return DistributeSeasonalRewardsOutput{}, fmt.Errorf("distribute rewards: find players: %w", err)
	}

	var rewards []SeasonRewardDTO
	rewardsGranted := 0

	for _, rating := range ratings {
		coins := rating.GetSeasonRewardCoins()
		tickets := rating.GetSeasonRewardTickets()

		if uc.inventoryService != nil {
			details := map[string]int{
				"coins":       coins,
				"pvp_tickets": tickets,
			}
			if err := uc.inventoryService.Credit(
				rating.PlayerID().String(),
				"seasonal_reward",
				details,
			); err != nil {
				// log but continue — one player failure must not block the rest
				continue
			}
			rewardsGranted++
			rewards = append(rewards, SeasonRewardDTO{
				PlayerID:   rating.PlayerID().String(),
				PeakLeague: rating.PeakLeague().String(),
				Coins:      coins,
				Tickets:    tickets,
			})
		}
	}

	// Run soft reset after distributing rewards
	if _, err := uc.seasonalReset.Execute(SeasonalResetInput{}); err != nil {
		return DistributeSeasonalRewardsOutput{}, fmt.Errorf("distribute rewards: seasonal reset: %w", err)
	}

	return DistributeSeasonalRewardsOutput{
		SeasonID:       seasonID,
		RewardsGranted: rewardsGranted,
		Rewards:        rewards,
	}, nil
}
