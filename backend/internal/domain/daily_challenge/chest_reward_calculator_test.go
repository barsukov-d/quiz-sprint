package daily_challenge

import (
	"math/rand"
	"testing"
)

// TestChestRewardCalculator_WoodenChest tests wooden chest rewards
func TestChestRewardCalculator_WoodenChest(t *testing.T) {
	seed := int64(12345)
	rng := rand.New(rand.NewSource(seed))
	calculator := NewChestRewardCalculator(rng)

	reward := calculator.CalculateRewards(ChestWooden, 1.0)

	// Wooden chest: 50-100 coins, 1 ticket, 0 or 1 bonus
	if reward.ChestType() != ChestWooden {
		t.Errorf("ChestType = %v, want %v", reward.ChestType(), ChestWooden)
	}

	if reward.Coins() < 50 || reward.Coins() > 100 {
		t.Errorf("Coins = %d, want 50-100", reward.Coins())
	}

	if reward.PvpTickets() != 1 {
		t.Errorf("PvpTickets = %d, want 1", reward.PvpTickets())
	}

	// 50% chance for 1 bonus - can be 0 or 1
	bonusCount := len(reward.MarathonBonuses())
	if bonusCount > 1 {
		t.Errorf("MarathonBonuses count = %d, want 0 or 1", bonusCount)
	}
}

// TestChestRewardCalculator_SilverChest tests silver chest rewards
func TestChestRewardCalculator_SilverChest(t *testing.T) {
	seed := int64(12345)
	rng := rand.New(rand.NewSource(seed))
	calculator := NewChestRewardCalculator(rng)

	reward := calculator.CalculateRewards(ChestSilver, 1.0)

	// Silver chest: 150-250 coins, 2-3 tickets, 1-2 bonuses
	if reward.ChestType() != ChestSilver {
		t.Errorf("ChestType = %v, want %v", reward.ChestType(), ChestSilver)
	}

	if reward.Coins() < 150 || reward.Coins() > 250 {
		t.Errorf("Coins = %d, want 150-250", reward.Coins())
	}

	if reward.PvpTickets() < 2 || reward.PvpTickets() > 3 {
		t.Errorf("PvpTickets = %d, want 2-3", reward.PvpTickets())
	}

	// 100% chance for 1 bonus + 30% chance for 2nd = 1 or 2
	bonusCount := len(reward.MarathonBonuses())
	if bonusCount < 1 || bonusCount > 2 {
		t.Errorf("MarathonBonuses count = %d, want 1 or 2", bonusCount)
	}
}

// TestChestRewardCalculator_GoldenChest tests golden chest rewards
func TestChestRewardCalculator_GoldenChest(t *testing.T) {
	seed := int64(12345)
	rng := rand.New(rand.NewSource(seed))
	calculator := NewChestRewardCalculator(rng)

	reward := calculator.CalculateRewards(ChestGolden, 1.0)

	// Golden chest: 300-500 coins, 4-5 tickets, 2-3 bonuses
	if reward.ChestType() != ChestGolden {
		t.Errorf("ChestType = %v, want %v", reward.ChestType(), ChestGolden)
	}

	if reward.Coins() < 300 || reward.Coins() > 500 {
		t.Errorf("Coins = %d, want 300-500", reward.Coins())
	}

	if reward.PvpTickets() < 4 || reward.PvpTickets() > 5 {
		t.Errorf("PvpTickets = %d, want 4-5", reward.PvpTickets())
	}

	// 100% chance for 2 bonuses + 40% chance for 3rd = 2 or 3
	bonusCount := len(reward.MarathonBonuses())
	if bonusCount < 2 || bonusCount > 3 {
		t.Errorf("MarathonBonuses count = %d, want 2 or 3", bonusCount)
	}
}

// TestChestRewardCalculator_StreakBonus tests streak multiplier application
func TestChestRewardCalculator_StreakBonus(t *testing.T) {
	tests := []struct {
		name        string
		chestType   ChestType
		streakBonus float64
		minCoins    int
		maxCoins    int
	}{
		{
			name:        "Wooden with 1.5x streak",
			chestType:   ChestWooden,
			streakBonus: 1.5,
			minCoins:    75,  // 50 * 1.5
			maxCoins:    150, // 100 * 1.5
		},
		{
			name:        "Silver with 1.25x streak",
			chestType:   ChestSilver,
			streakBonus: 1.25,
			minCoins:    187, // 150 * 1.25
			maxCoins:    312, // 250 * 1.25
		},
		{
			name:        "Golden with 1.1x streak",
			chestType:   ChestGolden,
			streakBonus: 1.1,
			minCoins:    330, // 300 * 1.1
			maxCoins:    550, // 500 * 1.1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seed := int64(12345)
			rng := rand.New(rand.NewSource(seed))
			calculator := NewChestRewardCalculator(rng)

			reward := calculator.CalculateRewards(tt.chestType, tt.streakBonus)

			if reward.Coins() < tt.minCoins || reward.Coins() > tt.maxCoins {
				t.Errorf("Coins = %d, want %d-%d", reward.Coins(), tt.minCoins, tt.maxCoins)
			}
		})
	}
}

// TestChestRewardCalculator_WoodenBonusProbability tests wooden chest bonus distribution
func TestChestRewardCalculator_WoodenBonusProbability(t *testing.T) {
	iterations := 1000
	withBonus := 0

	for i := 0; i < iterations; i++ {
		rng := rand.New(rand.NewSource(int64(i)))
		calculator := NewChestRewardCalculator(rng)
		reward := calculator.CalculateRewards(ChestWooden, 1.0)

		if len(reward.MarathonBonuses()) > 0 {
			withBonus++
		}
	}

	// Should be around 50% (±10% tolerance for randomness)
	percentage := float64(withBonus) / float64(iterations)
	if percentage < 0.40 || percentage > 0.60 {
		t.Errorf("Bonus probability = %.2f%%, want ~50%%", percentage*100)
	}
}

// TestChestRewardCalculator_SilverBonusProbability tests silver chest bonus distribution
func TestChestRewardCalculator_SilverBonusProbability(t *testing.T) {
	iterations := 1000
	withOneBonus := 0
	withTwoBonuses := 0

	for i := 0; i < iterations; i++ {
		rng := rand.New(rand.NewSource(int64(i)))
		calculator := NewChestRewardCalculator(rng)
		reward := calculator.CalculateRewards(ChestSilver, 1.0)

		bonusCount := len(reward.MarathonBonuses())
		if bonusCount == 1 {
			withOneBonus++
		} else if bonusCount == 2 {
			withTwoBonuses++
		}
	}

	// 100% get at least 1 bonus
	totalWithBonus := withOneBonus + withTwoBonuses
	if totalWithBonus != iterations {
		t.Errorf("Total with bonus = %d, want %d (100%%)", totalWithBonus, iterations)
	}

	// Should be around 30% for 2 bonuses (±10% tolerance)
	twoPercentage := float64(withTwoBonuses) / float64(iterations)
	if twoPercentage < 0.20 || twoPercentage > 0.40 {
		t.Errorf("2 bonuses probability = %.2f%%, want ~30%%", twoPercentage*100)
	}
}

// TestChestRewardCalculator_GoldenBonusProbability tests golden chest bonus distribution
func TestChestRewardCalculator_GoldenBonusProbability(t *testing.T) {
	iterations := 1000
	withTwoBonuses := 0
	withThreeBonuses := 0

	for i := 0; i < iterations; i++ {
		rng := rand.New(rand.NewSource(int64(i)))
		calculator := NewChestRewardCalculator(rng)
		reward := calculator.CalculateRewards(ChestGolden, 1.0)

		bonusCount := len(reward.MarathonBonuses())
		if bonusCount == 2 {
			withTwoBonuses++
		} else if bonusCount == 3 {
			withThreeBonuses++
		}
	}

	// 100% get at least 2 bonuses
	totalWithBonus := withTwoBonuses + withThreeBonuses
	if totalWithBonus != iterations {
		t.Errorf("Total with 2+ bonuses = %d, want %d (100%%)", totalWithBonus, iterations)
	}

	// Should be around 60% for 2 bonuses (±10% tolerance)
	twoPercentage := float64(withTwoBonuses) / float64(iterations)
	if twoPercentage < 0.50 || twoPercentage > 0.70 {
		t.Errorf("2 bonuses probability = %.2f%%, want ~60%%", twoPercentage*100)
	}

	// Should be around 40% for 3 bonuses (±10% tolerance)
	threePercentage := float64(withThreeBonuses) / float64(iterations)
	if threePercentage < 0.30 || threePercentage > 0.50 {
		t.Errorf("3 bonuses probability = %.2f%%, want ~40%%", threePercentage*100)
	}
}

// TestChestRewardCalculator_BonusTypes tests all bonus types can be selected
func TestChestRewardCalculator_BonusTypes(t *testing.T) {
	bonusSeen := make(map[MarathonBonus]bool)
	iterations := 100

	for i := 0; i < iterations; i++ {
		rng := rand.New(rand.NewSource(int64(i)))
		calculator := NewChestRewardCalculator(rng)
		reward := calculator.CalculateRewards(ChestGolden, 1.0)

		for _, bonus := range reward.MarathonBonuses() {
			bonusSeen[bonus] = true
		}
	}

	// Should have seen all 4 bonus types
	expectedBonuses := []MarathonBonus{BonusShield, BonusFiftyFifty, BonusSkip, BonusFreeze}
	for _, expected := range expectedBonuses {
		if !bonusSeen[expected] {
			t.Errorf("Bonus %v never appeared in %d iterations", expected, iterations)
		}
	}
}

// TestChestRewardCalculator_GoldenBonusesDifferent tests that bonuses are always different
func TestChestRewardCalculator_GoldenBonusesDifferent(t *testing.T) {
	iterations := 100

	for i := 0; i < iterations; i++ {
		rng := rand.New(rand.NewSource(int64(i)))
		calculator := NewChestRewardCalculator(rng)
		reward := calculator.CalculateRewards(ChestGolden, 1.0)

		bonuses := reward.MarathonBonuses()
		seen := make(map[MarathonBonus]bool)
		for _, b := range bonuses {
			if seen[b] {
				t.Errorf("Iteration %d: Duplicate bonus: %v in %v", i, b, bonuses)
			}
			seen[b] = true
		}
	}
}
