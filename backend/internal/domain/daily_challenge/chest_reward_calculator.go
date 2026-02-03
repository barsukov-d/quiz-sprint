package daily_challenge

import (
	"math/rand"
)

// ChestRewardCalculator is a domain service that calculates chest rewards
// Per docs/game_modes/daily_challenge/04_rewards.md and 06_domain.md
type ChestRewardCalculator struct {
	rng *rand.Rand
}

// NewChestRewardCalculator creates a new chest reward calculator
func NewChestRewardCalculator(rng *rand.Rand) *ChestRewardCalculator {
	if rng == nil {
		// Fallback to default RNG (for testing or non-critical paths)
		rng = rand.New(rand.NewSource(0))
	}
	return &ChestRewardCalculator{rng: rng}
}

// CalculateRewards calculates chest contents based on type and streak
// Per docs/game_modes/daily_challenge/04_rewards.md:
// - Wooden: 50-100 coins, 1 ticket, no bonuses
// - Silver: 150-250 coins, 2-3 tickets, 30% chance 1 bonus
// - Golden: 300-500 coins, 4-5 tickets, 70% chance 1 bonus OR 30% chance 2 bonuses
// Streak multiplier applies to coins only
func (c *ChestRewardCalculator) CalculateRewards(
	chestType ChestType,
	streakBonus float64,
) ChestReward {
	var baseCoins, pvpTickets int
	var marathonBonuses []MarathonBonus

	switch chestType {
	case ChestWooden:
		baseCoins = c.randomRange(50, 100)
		pvpTickets = 1
		marathonBonuses = c.selectWoodenBonuses()

	case ChestSilver:
		baseCoins = c.randomRange(150, 250)
		pvpTickets = c.randomRange(2, 3)
		marathonBonuses = c.selectSilverBonuses()

	case ChestGolden:
		baseCoins = c.randomRange(300, 500)
		pvpTickets = c.randomRange(4, 5)
		marathonBonuses = c.selectGoldenBonuses()

	default:
		// Fallback to wooden
		baseCoins = 50
		pvpTickets = 1
		marathonBonuses = []MarathonBonus{}
	}

	// Apply streak multiplier to coins only
	finalCoins := int(float64(baseCoins) * streakBonus)

	return ChestReward{
		chestType:       chestType,
		coins:           finalCoins,
		pvpTickets:      pvpTickets,
		marathonBonuses: marathonBonuses,
	}
}

// selectWoodenBonuses selects bonuses for wooden chest
// 50% chance: 1 random bonus
// 50% chance: No bonus
func (c *ChestRewardCalculator) selectWoodenBonuses() []MarathonBonus {
	if c.randomFloat() < 0.5 {
		return []MarathonBonus{c.randomBonus()}
	}
	return []MarathonBonus{}
}

// selectSilverBonuses selects bonuses for silver chest
// 100% chance: 1 random bonus (guaranteed)
// 30% chance: +1 extra bonus (2 total)
func (c *ChestRewardCalculator) selectSilverBonuses() []MarathonBonus {
	bonus1 := c.randomBonus()
	if c.randomFloat() < 0.3 {
		bonus2 := c.randomDifferentBonus(bonus1)
		return []MarathonBonus{bonus1, bonus2}
	}
	return []MarathonBonus{bonus1}
}

// selectGoldenBonuses selects bonuses for golden chest
// 100% chance: 2 random bonuses (guaranteed)
// 40% chance: +1 extra bonus (3 total)
func (c *ChestRewardCalculator) selectGoldenBonuses() []MarathonBonus {
	bonus1 := c.randomBonus()
	bonus2 := c.randomDifferentBonus(bonus1)
	if c.randomFloat() < 0.4 {
		bonus3 := c.randomDifferentBonus2(bonus1, bonus2)
		return []MarathonBonus{bonus1, bonus2, bonus3}
	}
	return []MarathonBonus{bonus1, bonus2}
}

// randomBonus returns a random bonus type with equal probability (25% each)
func (c *ChestRewardCalculator) randomBonus() MarathonBonus {
	allBonuses := []MarathonBonus{
		BonusShield,
		BonusFiftyFifty,
		BonusSkip,
		BonusFreeze,
	}
	return allBonuses[c.rng.Intn(len(allBonuses))]
}

// randomDifferentBonus returns a random bonus different from excluded
func (c *ChestRewardCalculator) randomDifferentBonus(excluded MarathonBonus) MarathonBonus {
	allBonuses := []MarathonBonus{
		BonusShield,
		BonusFiftyFifty,
		BonusSkip,
		BonusFreeze,
	}

	// Filter out excluded
	available := make([]MarathonBonus, 0, 3)
	for _, b := range allBonuses {
		if b != excluded {
			available = append(available, b)
		}
	}

	return available[c.rng.Intn(len(available))]
}

// randomDifferentBonus2 returns a random bonus different from two excluded bonuses
func (c *ChestRewardCalculator) randomDifferentBonus2(excluded1, excluded2 MarathonBonus) MarathonBonus {
	allBonuses := []MarathonBonus{
		BonusShield,
		BonusFiftyFifty,
		BonusSkip,
		BonusFreeze,
	}

	available := make([]MarathonBonus, 0, 2)
	for _, b := range allBonuses {
		if b != excluded1 && b != excluded2 {
			available = append(available, b)
		}
	}

	return available[c.rng.Intn(len(available))]
}

// randomRange returns random int in range [min, max] inclusive
func (c *ChestRewardCalculator) randomRange(min, max int) int {
	if min >= max {
		return min
	}
	return min + c.rng.Intn(max-min+1)
}

// randomFloat returns random float64 in range [0.0, 1.0)
func (c *ChestRewardCalculator) randomFloat() float64 {
	return c.rng.Float64()
}
