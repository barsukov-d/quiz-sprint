package daily_challenge

import (
	"time"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/daily_challenge"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
)

type RecoverStreakInput struct {
	PlayerID      string
	PaymentMethod string // "coins" or "ad"
}

type RecoverStreakOutput struct {
	Recovered     bool `json:"recovered"`
	CurrentStreak int  `json:"currentStreak"`
	BonusPercent  int  `json:"bonusPercent"`
	CoinsDeducted int  `json:"coinsDeducted"`
}

const streakRecoveryCostCoins = 50

type RecoverStreakUseCase struct {
	dailyGameRepo    daily_challenge.DailyGameRepository
	inventoryService InventoryService
}

func NewRecoverStreakUseCase(
	dailyGameRepo daily_challenge.DailyGameRepository,
	inventoryService InventoryService,
) *RecoverStreakUseCase {
	return &RecoverStreakUseCase{
		dailyGameRepo:    dailyGameRepo,
		inventoryService: inventoryService,
	}
}

func (uc *RecoverStreakUseCase) Execute(input RecoverStreakInput) (RecoverStreakOutput, error) {
	playerID, err := shared.NewUserID(input.PlayerID)
	if err != nil {
		return RecoverStreakOutput{}, err
	}

	today := daily_challenge.TodayUTC()
	yesterday := today.Previous()
	dayBeforeYesterday := yesterday.Previous()

	// 1. Check player hasn't already played today
	todayGame, _ := uc.dailyGameRepo.FindByPlayerAndDate(playerID, today)
	if todayGame != nil {
		return RecoverStreakOutput{}, daily_challenge.ErrAlreadyPlayedToday
	}

	// 2. Check player did NOT play yesterday (that's the missed day)
	yesterdayGame, _ := uc.dailyGameRepo.FindByPlayerAndDate(playerID, yesterday)
	if yesterdayGame != nil {
		// Played yesterday — streak is not broken, nothing to recover
		return RecoverStreakOutput{}, daily_challenge.ErrStreakNotRecoverable
	}

	// 3. Check player DID play day before yesterday (exactly 1 day gap)
	prevGame, err := uc.dailyGameRepo.FindByPlayerAndDate(playerID, dayBeforeYesterday)
	if err != nil || prevGame == nil {
		// Gap is more than 1 day — not recoverable
		return RecoverStreakOutput{}, daily_challenge.ErrStreakNotRecoverable
	}

	// 4. Verify there's actually a streak worth recovering (streak > 0)
	streak := prevGame.Streak()
	if streak.CurrentStreak() == 0 {
		return RecoverStreakOutput{}, daily_challenge.ErrStreakNotRecoverable
	}

	// 5. Process payment
	coinsDeducted := 0
	if input.PaymentMethod == "coins" {
		coinsDeducted = streakRecoveryCostCoins
		if uc.inventoryService != nil {
			err := uc.inventoryService.Debit(input.PlayerID, "streak_recovery", map[string]int{"coins": coinsDeducted})
			if err != nil {
				return RecoverStreakOutput{}, err
			}
		}
	}
	// "ad" payment: free (ad verification TODO)

	// 6. Recover streak by updating the previous game's lastPlayedDate to yesterday.
	// This way when StartDailyChallenge runs today, UpdateForDate(today) sees
	// lastPlayedDate=yesterday → consecutive → streak continues.
	recoveredStreak := daily_challenge.ReconstructStreakSystem(
		streak.CurrentStreak(),
		streak.BestStreak(),
		yesterday, // pretend they played yesterday
	)

	// Create a minimal "recovery" game for yesterday to bridge the gap
	now := time.Now().UTC().Unix()
	recoveryGame := daily_challenge.ReconstructDailyGame(
		daily_challenge.NewGameID(),
		playerID,
		prevGame.DailyQuizID(),
		yesterday,
		daily_challenge.GameStatusCompleted,
		nil, // no session
		recoveredStreak,
		nil, // no rank
		nil, // no chest reward
		now, // questionStartedAt
	)

	if err := uc.dailyGameRepo.Save(recoveryGame); err != nil {
		return RecoverStreakOutput{}, err
	}

	bonusPercent := int((streak.GetBonus() - 1.0) * 100)

	return RecoverStreakOutput{
		Recovered:     true,
		CurrentStreak: streak.CurrentStreak(),
		BonusPercent:  bonusPercent,
		CoinsDeducted: coinsDeducted,
	}, nil
}
