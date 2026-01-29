package marathon

import (
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/solo_marathon"
)

// GetPersonalBestsUseCase handles retrieving player's personal best records
type GetPersonalBestsUseCase struct {
	personalBestRepo solo_marathon.PersonalBestRepository
}

// NewGetPersonalBestsUseCase creates a new GetPersonalBestsUseCase
func NewGetPersonalBestsUseCase(
	personalBestRepo solo_marathon.PersonalBestRepository,
) *GetPersonalBestsUseCase {
	return &GetPersonalBestsUseCase{
		personalBestRepo: personalBestRepo,
	}
}

// Execute retrieves all personal best records for a player
func (uc *GetPersonalBestsUseCase) Execute(input GetPersonalBestsInput) (GetPersonalBestsOutput, error) {
	// 1. Validate and convert input to domain types
	playerID, err := shared.NewUserID(input.PlayerID)
	if err != nil {
		return GetPersonalBestsOutput{}, err
	}

	// 2. Find all personal bests for player (across all categories)
	personalBests, err := uc.personalBestRepo.FindAllByPlayer(playerID)
	if err != nil {
		if err == solo_marathon.ErrPersonalBestNotFound {
			// No records yet - return empty list
			return GetPersonalBestsOutput{
				PersonalBests: []PersonalBestDTO{},
			}, nil
		}
		return GetPersonalBestsOutput{}, err
	}

	// 3. Convert to DTOs
	dtos := ToPersonalBestDTOs(personalBests)

	// 4. Find overall best (highest streak across all categories)
	var overallBest *PersonalBestDTO
	if len(personalBests) > 0 {
		best := personalBests[0]
		for _, pb := range personalBests[1:] {
			if pb.BestStreak() > best.BestStreak() {
				best = pb
			}
		}
		dto := ToPersonalBestDTO(best)
		overallBest = &dto
	}

	// 5. Build output
	return GetPersonalBestsOutput{
		PersonalBests: dtos,
		OverallBest:   overallBest,
	}, nil
}
