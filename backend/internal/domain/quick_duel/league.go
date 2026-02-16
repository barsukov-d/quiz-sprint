package quick_duel

// League represents competitive ranking tiers
type League int

const (
	LeagueBronze   League = iota // 0-999 MMR
	LeagueSilver                 // 1000-1499 MMR
	LeagueGold                   // 1500-1999 MMR
	LeaguePlatinum               // 2000-2499 MMR
	LeagueDiamond                // 2500-2999 MMR
	LeagueLegend                 // 3000+ MMR
)

// Division within a league (1-4, except Legend which has 1)
type Division int

const (
	DivisionIV Division = 4 // Lowest division (entry point)
	DivisionIII Division = 3
	DivisionII Division = 2
	DivisionI Division = 1 // Highest division
)

// LeagueInfo holds league and division info with MMR thresholds
type LeagueInfo struct {
	league   League
	division Division
}

// MMR thresholds for each league
const (
	MMRBronzeMin   = 0
	MMRSilverMin   = 1000
	MMRGoldMin     = 1500
	MMRPlatinumMin = 2000
	MMRDiamondMin  = 2500
	MMRLegendMin   = 3000

	// Division span within each league (except Legend)
	DivisionSpan = 125 // 500 MMR range / 4 divisions
)

// GetLeagueFromMMR determines league and division from MMR value
func GetLeagueFromMMR(mmr int) LeagueInfo {
	if mmr < 0 {
		mmr = 0
	}

	var league League
	var division Division

	switch {
	case mmr >= MMRLegendMin:
		league = LeagueLegend
		division = DivisionI // Legend has only one division
	case mmr >= MMRDiamondMin:
		league = LeagueDiamond
		division = getDivisionInRange(mmr, MMRDiamondMin)
	case mmr >= MMRPlatinumMin:
		league = LeaguePlatinum
		division = getDivisionInRange(mmr, MMRPlatinumMin)
	case mmr >= MMRGoldMin:
		league = LeagueGold
		division = getDivisionInRange(mmr, MMRGoldMin)
	case mmr >= MMRSilverMin:
		league = LeagueSilver
		division = getDivisionInRange(mmr, MMRSilverMin)
	default:
		league = LeagueBronze
		division = getDivisionInRange(mmr, MMRBronzeMin)
	}

	return LeagueInfo{
		league:   league,
		division: division,
	}
}

// getDivisionInRange calculates division within a league range
func getDivisionInRange(mmr int, leagueMin int) Division {
	offset := mmr - leagueMin

	switch {
	case offset >= DivisionSpan*3: // 375+
		return DivisionI
	case offset >= DivisionSpan*2: // 250-374
		return DivisionII
	case offset >= DivisionSpan: // 125-249
		return DivisionIII
	default: // 0-124
		return DivisionIV
	}
}

// League methods
func (l League) String() string {
	switch l {
	case LeagueBronze:
		return "bronze"
	case LeagueSilver:
		return "silver"
	case LeagueGold:
		return "gold"
	case LeaguePlatinum:
		return "platinum"
	case LeagueDiamond:
		return "diamond"
	case LeagueLegend:
		return "legend"
	default:
		return "unknown"
	}
}

// Label returns localized label
func (l League) Label() string {
	switch l {
	case LeagueBronze:
		return "Бронза"
	case LeagueSilver:
		return "Серебро"
	case LeagueGold:
		return "Золото"
	case LeaguePlatinum:
		return "Платина"
	case LeagueDiamond:
		return "Алмаз"
	case LeagueLegend:
		return "Легенда"
	default:
		return "Неизвестно"
	}
}

// Icon returns emoji icon
func (l League) Icon() string {
	switch l {
	case LeagueBronze:
		return "🥉"
	case LeagueSilver:
		return "🥈"
	case LeagueGold:
		return "🥇"
	case LeaguePlatinum:
		return "💍"
	case LeagueDiamond:
		return "💎"
	case LeagueLegend:
		return "👑"
	default:
		return "❓"
	}
}

// MinMMR returns minimum MMR for this league
func (l League) MinMMR() int {
	switch l {
	case LeagueBronze:
		return MMRBronzeMin
	case LeagueSilver:
		return MMRSilverMin
	case LeagueGold:
		return MMRGoldMin
	case LeaguePlatinum:
		return MMRPlatinumMin
	case LeagueDiamond:
		return MMRDiamondMin
	case LeagueLegend:
		return MMRLegendMin
	default:
		return 0
	}
}

// LeagueInfo methods
func (li LeagueInfo) League() League     { return li.league }
func (li LeagueInfo) Division() Division { return li.division }

// Label returns formatted label like "Золото II"
func (li LeagueInfo) Label() string {
	if li.league == LeagueLegend {
		return li.league.Label()
	}
	return li.league.Label() + " " + li.division.String()
}

// FullLabel returns label with icon
func (li LeagueInfo) FullLabel() string {
	return li.league.Icon() + " " + li.Label()
}

// Division methods
func (d Division) String() string {
	switch d {
	case DivisionI:
		return "I"
	case DivisionII:
		return "II"
	case DivisionIII:
		return "III"
	case DivisionIV:
		return "IV"
	default:
		return ""
	}
}

// Value returns int value for storage
func (d Division) Value() int {
	return int(d)
}
